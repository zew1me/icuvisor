package jsonenc

import (
	"encoding"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
)

// Encode converts value into the JSON-shaped tree used by response shaping.
func Encode(value any) (any, error) {
	out, err := toJSONValue(reflect.ValueOf(value), map[jsonVisit]bool{})
	if err != nil {
		return nil, err
	}
	return out, nil
}

type jsonVisit struct {
	typ reflect.Type
	ptr uintptr
}

func toJSONValue(value reflect.Value, visits map[jsonVisit]bool) (any, error) {
	if !value.IsValid() {
		return nil, nil
	}
	if value.Kind() == reflect.Interface {
		if value.IsNil() {
			return nil, nil
		}
		return toJSONValue(value.Elem(), visits)
	}
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil, nil
		}
		if canInterface(value) {
			if marshaled, ok, err := marshalSpecialValue(value.Interface()); ok || err != nil {
				return marshaled, err
			}
		}
		cleanup, err := enterJSONVisit(value, visits)
		if err != nil {
			return nil, err
		}
		defer cleanup()
		return toJSONValue(value.Elem(), visits)
	}
	if canInterface(value) {
		if marshaled, ok, err := marshalSpecialValue(value.Interface()); ok || err != nil {
			return marshaled, err
		}
	}
	if value.Kind() == reflect.Slice && value.Type().Elem().Kind() == reflect.Uint8 {
		if canInterface(value) {
			return marshalJSONValue(value.Interface())
		}
	}
	return reflectJSONValue(value, visits)
}

func reflectJSONValue(value reflect.Value, visits map[jsonVisit]bool) (any, error) {
	switch value.Kind() {
	case reflect.Bool:
		return value.Bool(), nil
	case reflect.String:
		return value.String(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(value.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return float64(value.Uint()), nil
	case reflect.Float32:
		floatValue := value.Float()
		if math.IsNaN(floatValue) || math.IsInf(floatValue, 0) {
			return nil, unsupportedFloatError(floatValue, 32)
		}
		if canInterface(value) {
			return marshalJSONValue(value.Interface())
		}
		return float64(float32(floatValue)), nil
	case reflect.Float64:
		floatValue := value.Float()
		if math.IsNaN(floatValue) || math.IsInf(floatValue, 0) {
			return nil, unsupportedFloatError(floatValue, 64)
		}
		return floatValue, nil
	case reflect.Map:
		return mapToJSONValue(value, visits)
	case reflect.Slice, reflect.Array:
		return sliceToJSONValue(value, visits)
	case reflect.Struct:
		return structToJSONValue(value, visits)
	case reflect.Invalid:
		return nil, nil
	default:
		return nil, fmt.Errorf("marshaling response value: unsupported JSON value %s", value.Kind())
	}
}

func unsupportedFloatError(value float64, bitSize int) error {
	return fmt.Errorf("marshaling response value: json: unsupported value: %s", strconv.FormatFloat(value, 'g', -1, bitSize))
}

func mapToJSONValue(value reflect.Value, visits map[jsonVisit]bool) (any, error) {
	if value.IsNil() {
		return nil, nil
	}
	cleanup, err := enterJSONVisit(value, visits)
	if err != nil {
		return nil, err
	}
	defer cleanup()
	if value.Type().Key().Kind() != reflect.String {
		if canInterface(value) {
			return marshalJSONValue(value.Interface())
		}
		return nil, fmt.Errorf("marshaling response value: unsupported map key type %s", value.Type().Key())
	}
	out := make(map[string]any, value.Len())
	iter := value.MapRange()
	for iter.Next() {
		item, err := toJSONValue(iter.Value(), visits)
		if err != nil {
			return nil, err
		}
		out[iter.Key().String()] = item
	}
	return out, nil
}

func sliceToJSONValue(value reflect.Value, visits map[jsonVisit]bool) (any, error) {
	if value.Kind() == reflect.Slice && value.IsNil() {
		return nil, nil
	}
	if value.Kind() == reflect.Slice {
		cleanup, err := enterJSONVisit(value, visits)
		if err != nil {
			return nil, err
		}
		defer cleanup()
	}
	out := make([]any, value.Len())
	for i := range value.Len() {
		item, err := toJSONValue(value.Index(i), visits)
		if err != nil {
			return nil, err
		}
		out[i] = item
	}
	return out, nil
}

func structToJSONValue(value reflect.Value, visits map[jsonVisit]bool) (any, error) {
	out := make(map[string]any, value.NumField())
	seenFields := map[string]struct{}{}
	valueType := value.Type()
	for i := range value.NumField() {
		field := valueType.Field(i)
		if field.PkgPath != "" {
			continue
		}
		name, omitEmpty, skip, fallback := jsonField(field)
		if skip {
			continue
		}
		if fallback {
			return marshalJSONValue(value.Interface())
		}
		if _, exists := seenFields[name]; exists {
			return marshalJSONValue(value.Interface())
		}
		seenFields[name] = struct{}{}
		fieldValue := value.Field(i)
		if omitEmpty && isEmptyJSONValue(fieldValue) {
			continue
		}
		item, err := toJSONValue(fieldValue, visits)
		if err != nil {
			return nil, err
		}
		out[name] = item
	}
	return out, nil
}

func enterJSONVisit(value reflect.Value, visits map[jsonVisit]bool) (func(), error) {
	if !canInterface(value) {
		return func() {}, nil
	}
	ptr := value.Pointer()
	if ptr == 0 {
		return func() {}, nil
	}
	visit := jsonVisit{typ: value.Type(), ptr: ptr}
	if visits[visit] {
		_, err := marshalJSONValue(value.Interface())
		if err != nil {
			return func() {}, err
		}
		return func() {}, nil
	}
	visits[visit] = true
	return func() { delete(visits, visit) }, nil
}

func jsonField(field reflect.StructField) (name string, omitEmpty bool, skip bool, fallback bool) {
	name = field.Name
	tag := field.Tag.Get("json")
	if tag == "-" {
		return "", false, true, false
	}
	parts := strings.Split(tag, ",")
	if parts[0] != "" {
		name = parts[0]
	} else if field.Anonymous {
		return "", false, false, true
	}
	for _, option := range parts[1:] {
		switch option {
		case "omitempty":
			omitEmpty = true
		case "string":
			fallback = true
		}
	}
	return name, omitEmpty, false, fallback
}

func isEmptyJSONValue(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return value.Len() == 0
	case reflect.Bool:
		return !value.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return value.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return value.Float() == 0
	case reflect.Interface, reflect.Pointer:
		return value.IsNil()
	default:
		return false
	}
}

func marshalSpecialValue(value any) (any, bool, error) {
	if number, ok := value.(json.Number); ok {
		out, err := marshalJSONValue(number)
		return out, true, err
	}
	if marshaler, ok := value.(json.Marshaler); ok {
		out, err := marshalJSONValue(marshaler)
		return out, true, err
	}
	if marshaler, ok := value.(encoding.TextMarshaler); ok {
		text, err := marshaler.MarshalText()
		if err != nil {
			return nil, true, fmt.Errorf("marshaling response value: %w", err)
		}
		return string(text), true, nil
	}
	return nil, false, nil
}

func marshalJSONValue(value any) (any, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("marshaling response value: %w", err)
	}
	var out any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("unmarshaling response value: %w", err)
	}
	return out, nil
}

func canInterface(value reflect.Value) bool {
	return value.IsValid() && value.CanInterface()
}
