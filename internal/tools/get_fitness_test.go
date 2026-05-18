package tools

import (
	"context"
	"encoding/json"
	"testing"
)

func TestGetFitnessToolShapes(t *testing.T) {
	t.Parallel()

	client := newFakeFitnessMetricsClient(t)
	tool := newGetFitnessTool(client, client, "test", "UTC", false)
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"start_date":"2026-05-01","end_date":"2026-05-02"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	got := resultMap(t, result)
	rows := got["fitness"].([]any)
	if rows[0].(map[string]any)["date"] != "2026-05-01" || rows[1].(map[string]any)["date"] != "2026-05-02" {
		t.Fatalf("fitness row order = %#v", rows)
	}
	meta := got["_meta"].(map[string]any)
	if meta["timezone"] != "America/Sao_Paulo" || meta["server_version"] != "test" {
		t.Fatalf("meta = %#v", meta)
	}
}
