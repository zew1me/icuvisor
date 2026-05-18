package tools

import (
	"context"
	"encoding/json"
	"testing"
)

func TestGetTrainingSummaryToolShapes(t *testing.T) {
	t.Parallel()

	client := newFakeFitnessMetricsClient(t)
	tool := newGetTrainingSummaryTool(client, client, "test", "UTC", false)
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"start_date":"2026-05-01","end_date":"2026-05-02"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	got := resultMap(t, result)
	summary := got["summary"].(map[string]any)
	if summary["training_load"] != float64(130) || summary["distance_km"] != float64(45) {
		t.Fatalf("summary = %#v", summary)
	}
	if _, ok := summary["tss"]; ok {
		t.Fatal("summary should not expose tss")
	}
}
