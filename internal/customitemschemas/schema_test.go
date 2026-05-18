package customitemschemas

import (
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

func TestValidateContentAgainstReadSchemaRejectsUnknownKeysAndWrongKinds(t *testing.T) {
	t.Parallel()

	itemType := "FITNESS_CHART"
	items := []intervals.CustomItem{{Type: &itemType, Content: map[string]any{"series": []any{map[string]any{"field": "ctl"}}, "layout": map[string]any{"height": float64(240)}}}}
	count, source, err := ValidateContentAgainstReadSchema(items, itemType, map[string]any{"series": []any{map[string]any{"field": "atl"}}, "layout": map[string]any{"height": float64(260)}}, true)
	if err != nil {
		t.Fatalf("ValidateContentAgainstReadSchema() valid content error = %v", err)
	}
	if source != SchemaSourceRead || count != 1 {
		t.Fatalf("ValidateContentAgainstReadSchema() source=%q count=%d, want read/1", source, count)
	}
	_, _, err = ValidateContentAgainstReadSchema(items, itemType, map[string]any{"series": []any{map[string]any{"field": "atl"}}, "layout": map[string]any{"height": "tall"}}, true)
	if err == nil || !strings.Contains(err.Error(), "content.layout.height must be number") {
		t.Fatalf("wrong kind error = %v, want layout height kind error", err)
	}
	_, _, err = ValidateContentAgainstReadSchema(items, itemType, map[string]any{"series": []any{map[string]any{"field": "atl"}}, "layout": map[string]any{"height": float64(260)}, "extra": true}, true)
	if err == nil || !strings.Contains(err.Error(), "content.extra is not in the readable schema") {
		t.Fatalf("unknown key error = %v, want readable schema error", err)
	}
}

func TestValidateContentAgainstReadSchemaFallsBackToDescriptor(t *testing.T) {
	t.Parallel()

	content := map[string]any{
		"field":      "travel_fatigue",
		"label":      "Travel fatigue",
		"type":       "number",
		"units":      "score",
		"format":     "0.0",
		"script":     "return input",
		"visibility": "PRIVATE",
	}
	count, source, err := ValidateContentAgainstReadSchema(nil, "INPUT_FIELD", content, true)
	if err != nil {
		t.Fatalf("descriptor fallback rejected valid INPUT_FIELD: %v", err)
	}
	if source != SchemaSourceDescriptor || count != 0 {
		t.Fatalf("source=%q count=%d, want descriptor/0", source, count)
	}

	partial := map[string]any{"field": "only_field", "type": "number"}
	if _, _, err := ValidateContentAgainstReadSchema(nil, "INPUT_FIELD", partial, true); err != nil {
		t.Fatalf("descriptor fallback should skip requireComplete check, got %v", err)
	}

	_, _, err = ValidateContentAgainstReadSchema(nil, "INPUT_FIELD", map[string]any{"unknown_key": "x"}, true)
	if err == nil || !strings.Contains(err.Error(), "content.unknown_key is not in the readable schema") {
		t.Fatalf("descriptor fallback should still reject unknown keys, got %v", err)
	}

	_, _, err = ValidateContentAgainstReadSchema(nil, "NOT_A_REAL_ITEM_TYPE", map[string]any{}, true)
	if err == nil || !strings.Contains(err.Error(), "no readable custom item schema found") {
		t.Fatalf("unknown item_type should still fail, got %v", err)
	}
}

func TestSampleForItemType(t *testing.T) {
	t.Parallel()

	sample, ok := SampleForItemType("INPUT_FIELD")
	if !ok {
		t.Fatalf("SampleForItemType(INPUT_FIELD) missing")
	}
	if _, hasField := sample["field"]; !hasField {
		t.Fatalf("INPUT_FIELD sample lacks field key: %#v", sample)
	}
	if _, ok := SampleForItemType("NOT_A_REAL_TYPE"); ok {
		t.Fatalf("SampleForItemType(NOT_A_REAL_TYPE) should be missing")
	}
}

func TestFamiliesHaveSamplesAndInferredPaths(t *testing.T) {
	t.Parallel()

	for _, family := range Families() {
		if family.Key == "" || family.Title == "" || family.Description == "" || len(family.Items) == 0 {
			t.Fatalf("family is incomplete: %#v", family)
		}
		for _, item := range family.Items {
			if item.ItemType == "" || item.Description == "" {
				t.Fatalf("item descriptor is incomplete: %#v", item)
			}
			if item.Sample == nil && item.SharesSchemaWith == "" {
				t.Fatalf("item descriptor %s has no sample or alias", item.ItemType)
			}
			if item.Sample == nil {
				continue
			}
			schema, err := InferContentSchema([]map[string]any{item.Sample})
			if err != nil {
				t.Fatalf("InferContentSchema(%s) error = %v", item.ItemType, err)
			}
			if len(SchemaPaths(schema)) < 2 {
				t.Fatalf("SchemaPaths(%s) too short: %#v", item.ItemType, SchemaPaths(schema))
			}
		}
	}
}
