package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

func TestGetFitnessProjectionStandardRampGolden(t *testing.T) {
	t.Parallel()

	client := newFakeFitnessMetricsClient(t)
	tool := newGetFitnessProjectionTool(client, client, "test", "UTC", false)
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"start_date":"2026-05-01","horizon_days":14,"weekly_ramp_pct":7,"recovery_week_cadence":0}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	assertAnalyzerGolden(t, "analyzer/fitness_projection_standard_ramp.golden.json", result.StructuredContent)
	root := analyzerMap(t, result.StructuredContent)
	if _, ok := root["series"]; ok {
		t.Fatal("terse projection unexpectedly included series")
	}
	meta := analyzerMap(t, root["_meta"])
	for _, key := range []string{"method", "source_tools", "n", "missing_days", "missing_action", "insufficient_sample", "assumptions", "boundaries"} {
		if _, ok := meta[key]; !ok {
			t.Fatalf("mandatory projection _meta.%s missing from %#v", key, meta)
		}
	}
	assumptions := analyzerMap(t, meta["assumptions"])
	if assumptions["model"] != fitnessProjectionModel || assumptions["weekly_ramp_pct"] != float64(7) {
		t.Fatalf("assumptions = %#v", assumptions)
	}
}

func TestGetFitnessProjectionRecoveryWeekFullGolden(t *testing.T) {
	t.Parallel()

	client := newFakeFitnessMetricsClient(t)
	tool := newGetFitnessProjectionTool(client, client, "test", "UTC", false)
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"start_date":"2026-05-01","horizon_days":28,"weekly_ramp_pct":5,"recovery_week_cadence":4,"recovery_week_load_pct":50,"include_full":true}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	assertAnalyzerGolden(t, "analyzer/fitness_projection_recovery_week_full.golden.json", result.StructuredContent)
	series := analyzerMap(t, result.StructuredContent)["series"].([]any)
	if len(series) != 29 {
		t.Fatalf("series length = %d, want start point plus 28 projected days", len(series))
	}
	foundRecovery := false
	for _, rawPoint := range series {
		point := rawPoint.(map[string]any)
		if point["training_load_source"] == "modeled_recovery_week" {
			foundRecovery = true
			break
		}
	}
	if !foundRecovery {
		t.Fatal("full projection series did not include a modeled_recovery_week point")
	}
}

func TestGetFitnessProjectionRejectsInvalidInputs(t *testing.T) {
	t.Parallel()

	client := newFakeFitnessMetricsClient(t)
	tool := newGetFitnessProjectionTool(client, client, "test", "UTC", false)
	cases := []struct {
		name string
		args string
	}{
		{name: "free form model", args: `{"start_date":"2026-05-01","horizon_days":7,"model":"coach_magic"}`},
		{name: "both horizon fields", args: `{"start_date":"2026-05-01","horizon_days":7,"horizon_date":"2026-05-08"}`},
		{name: "ramp out of bounds", args: `{"start_date":"2026-05-01","horizon_days":7,"weekly_ramp_pct":75}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(tc.args)})
			if err == nil {
				t.Fatal("Handler() error = nil, want validation error")
			}
			message, ok := PublicErrorMessage(err)
			if !ok || message != invalidFitnessProjectionMessage {
				t.Fatalf("public error = %q (ok=%v), want %q", message, ok, invalidFitnessProjectionMessage)
			}
		})
	}
}

func TestGetFitnessProjectionInsufficientCurrentFitnessData(t *testing.T) {
	t.Parallel()

	client := &fakeFitnessMetricsClient{
		fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
		summaries:         decodeSummaries(t, `[{"date":"2026-05-01","fitness":null,"fatigue":80,"form":-8}]`),
	}
	tool := newGetFitnessProjectionTool(client, client, "test", "UTC", false)
	_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"start_date":"2026-05-01","horizon_days":7}`)})
	if err == nil {
		t.Fatal("Handler() error = nil, want insufficient-data error")
	}
	message, ok := PublicErrorMessage(err)
	if !ok || !strings.Contains(message, "insufficient current fitness data") {
		t.Fatalf("public error = %q (ok=%v), want insufficient current fitness data", message, ok)
	}
}
