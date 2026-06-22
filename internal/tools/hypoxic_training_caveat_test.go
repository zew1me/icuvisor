package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

func TestActivityRowsEmitHypoxicTrainingCaveatForExplicitEvidence(t *testing.T) {
	t.Parallel()

	client := newFakeActivitiesClient(t, []string{
		`{"id":"a1","name":"Hypoxic endurance ride","type":"Ride","start_date_local":"2026-01-03T07:00:00","icu_training_load":55,"power_load":55,"total_elevation_gain":120,"tags":["altitude tent"],"hypoxia_note":"reduced oxygen block"}`,
	}, "metric")
	client.customItems = decodeCustomItems(t, `{"id":"c1","type":"ACTIVITY_FIELD","content":{"field":"hypoxia_note"}}`)
	tool := newGetActivitiesToolWithGear(client, client, nil, nil, client, newCustomFieldCache(), "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-01-01","custom_fields":["hypoxia_note"]}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	row := resultMap(t, result)["activities"].([]any)[0].(map[string]any)
	caveat := row["hypoxic_training_caveat"].(map[string]any)
	if caveat["reason"] != "explicit_hypoxia_evidence" || caveat["load_basis"] != "power_load_available" {
		t.Fatalf("caveat = %#v, want explicit power-load caveat", caveat)
	}
	message := caveat["message"].(string)
	for _, want := range []string{"power-based", "under-represent", "Do not change logged training_load", "hypoxia multiplier"} {
		if !strings.Contains(message, want) {
			t.Fatalf("caveat message %q missing %q", message, want)
		}
	}
	provenance := joinedStrings(caveat["provenance"].([]any))
	for _, want := range []string{"activity.name", "activity.tags", "activity.custom_fields.hypoxia_note"} {
		if !strings.Contains(provenance, want) {
			t.Fatalf("provenance %q missing %q", provenance, want)
		}
	}
	if row["training_load"] != float64(55) {
		t.Fatalf("training_load = %#v, want unchanged 55", row["training_load"])
	}
}

func TestExtendedMetricsHypoxicCaveatUsesHRLoadWording(t *testing.T) {
	t.Parallel()

	client := newFakeExtendedMetricsClient(t)
	client.activity = decodeExtendedMetricsActivity(t, `{"id":"activity-hr","name":"Low oxygen treadmill","type":"Run","icu_training_load":48,"hr_load":48,"feel":2,"session_rpe":7}`)
	client.intervalsErr = intervals.ErrNotFound
	client.powerErr = intervals.ErrNotFound
	tool := newGetExtendedMetricsTool(client, client, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"activity-hr"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	meta := resultMap(t, result)["_meta"].(map[string]any)
	caveat := meta["hypoxic_training_caveat"].(map[string]any)
	if caveat["load_basis"] != "hr_load_available" {
		t.Fatalf("caveat = %#v, want HR load basis", caveat)
	}
	message := caveat["message"].(string)
	for _, want := range []string{"HR-based load", "acute cardiovascular response", "supporting context only", "hypoxia multiplier"} {
		if !strings.Contains(message, want) {
			t.Fatalf("caveat message %q missing %q", message, want)
		}
	}
}

func TestHypoxicCaveatDoesNotInferFromAltitudeOrSpO2Alone(t *testing.T) {
	t.Parallel()

	raw := map[string]any{
		"name":                 "High altitude climb",
		"total_elevation_gain": float64(1800),
		"stream_types":         []any{"time", "altitude", "heartrate"},
		"spO2":                 float64(88),
		"tags":                 []any{"mountain", "climb"},
	}
	if caveat := hypoxicTrainingCaveatForActivity(raw, nil); caveat != nil {
		t.Fatalf("hypoxicTrainingCaveatForActivity() = %#v, want nil for altitude/SpO2 without explicit provenance", caveat)
	}
}
