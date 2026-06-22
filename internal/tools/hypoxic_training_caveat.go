package tools

import (
	"fmt"
	"sort"
	"strings"
)

type hypoxicTrainingCaveat struct {
	Reason     string   `json:"reason"`
	Provenance []string `json:"provenance"`
	LoadBasis  string   `json:"load_basis"`
	Message    string   `json:"message"`
}

func hypoxicTrainingCaveatForActivity(raw map[string]any, customFieldCodes []string) *hypoxicTrainingCaveat {
	provenance := hypoxicTrainingProvenance(raw, customFieldCodes)
	if len(provenance) == 0 {
		return nil
	}
	sort.Strings(provenance)
	basis := hypoxicLoadBasis(raw)
	return &hypoxicTrainingCaveat{
		Reason:     "explicit_hypoxia_evidence",
		Provenance: provenance,
		LoadBasis:  basis,
		Message:    hypoxicLoadCaveatMessage(basis),
	}
}

func hypoxicTrainingProvenance(raw map[string]any, customFieldCodes []string) []string {
	if len(raw) == 0 {
		return nil
	}
	var provenance []string
	for _, key := range []string{"name", "_note", "description", "notes"} {
		if explicitHypoxiaText(anyText(raw[key])) {
			provenance = append(provenance, "activity."+key)
		}
	}
	for _, tag := range anyStringSlice(raw["tags"]) {
		if explicitHypoxiaText(tag) {
			provenance = append(provenance, "activity.tags")
			break
		}
	}
	for _, code := range customFieldCodes {
		trimmed := strings.TrimSpace(code)
		if trimmed == "" {
			continue
		}
		value, ok := raw[trimmed]
		if !ok || value == nil {
			continue
		}
		if explicitHypoxiaText(anyText(value)) || customFieldKeyExplicitHypoxia(trimmed, value) {
			provenance = append(provenance, "activity.custom_fields."+trimmed)
		}
	}
	return sortedUniqueStrings(provenance)
}

func customFieldKeyExplicitHypoxia(key string, value any) bool {
	if !explicitHypoxiaText(key) {
		return false
	}
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		lower := strings.ToLower(strings.TrimSpace(typed))
		return lower != "" && lower != "false" && lower != "no" && lower != "none" && lower != "0"
	default:
		return true
	}
}

func explicitHypoxiaText(value string) bool {
	lower := strings.ToLower(strings.TrimSpace(value))
	if lower == "" {
		return false
	}
	for _, negated := range []string{"no hypox", "not hypox", "without hypox", "no altitude tent", "not altitude tent", "without altitude tent"} {
		if strings.Contains(lower, negated) {
			return false
		}
	}
	for _, term := range []string{
		"hypoxic", "hypoxia", "altitude tent", "altitude chamber", "altitude room",
		"simulated altitude", "normobaric hypoxia", "reduced oxygen", "low oxygen",
		"oxygen restricted", "oxygen restriction",
	} {
		if strings.Contains(lower, term) {
			return true
		}
	}
	return false
}

func hypoxicLoadBasis(raw map[string]any) string {
	icuLoad, hasICULoad := rawNumber(raw, "icu_training_load")
	powerLoad, hasPowerLoad := rawNumber(raw, "power_load")
	hrLoad, hasHRLoad := rawNumber(raw, "hr_load")
	_, hasPaceLoad := rawNumber(raw, "pace_load")
	if hasPowerLoad && (!hasHRLoad || (hasICULoad && nearlyEqual(powerLoad, icuLoad))) {
		return "power_load_available"
	}
	if hasHRLoad && (!hasPowerLoad || (hasICULoad && nearlyEqual(hrLoad, icuLoad))) {
		return "hr_load_available"
	}
	if hasPowerLoad && hasHRLoad {
		return "multiple_load_variants_available"
	}
	if hasPaceLoad {
		return "pace_load_available"
	}
	return "unknown"
}

func hypoxicLoadCaveatMessage(basis string) string {
	switch basis {
	case "power_load_available":
		return "Explicit hypoxia evidence is present. Do not change logged training_load or apply a hypoxia multiplier. If this session's TSS/load is power-based, it may under-represent extra hypoxic physiological strain; use HR, RPE/feel, and recovery trends as supporting context only."
	case "hr_load_available":
		return "Explicit hypoxia evidence is present. Do not change logged training_load or apply a hypoxia multiplier. HR-based load may capture some acute cardiovascular response to reduced oxygen, but do not assume it fully models hypoxic stress; use RPE/feel and recovery trends as supporting context only."
	default:
		return "Explicit hypoxia evidence is present. CTL, ATL, and Form use logged training_load; do not apply a hypoxia multiplier. If the load was power-based it may under-represent extra strain, while HR, RPE/feel, and recovery trends can provide supporting context only."
	}
}

func anyText(value any) string {
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return typed
	case []string:
		return strings.Join(typed, " ")
	case []any:
		parts := make([]string, 0, len(typed))
		for _, item := range typed {
			parts = append(parts, anyText(item))
		}
		return strings.Join(parts, " ")
	default:
		return fmt.Sprint(value)
	}
}

func anyStringSlice(value any) []string {
	switch typed := value.(type) {
	case []string:
		return typed
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			text := strings.TrimSpace(anyText(item))
			if text != "" {
				out = append(out, text)
			}
		}
		return out
	default:
		return nil
	}
}

func nearlyEqual(a float64, b float64) bool {
	if a > b {
		return a-b < 0.001
	}
	return b-a < 0.001
}

func sortedUniqueStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}
