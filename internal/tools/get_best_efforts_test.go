package tools

import (
	"context"
	"encoding/json"
	"testing"
)

func TestGetBestEffortsToolShapes(t *testing.T) {
	t.Parallel()

	client := newFakeFitnessMetricsClient(t)
	tool := newGetBestEffortsTool(client, "test", false)
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"sports":["Ride","Run"],"duration_seconds":[60],"distance_meters":[1000]}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	got := resultMap(t, result)
	sports := got["sports"].([]any)
	if len(sports) != 2 {
		t.Fatalf("sports count = %d, want 2", len(sports))
	}
	efforts := sports[0].(map[string]any)["efforts"].([]any)
	if len(efforts) != 3 || efforts[0].(map[string]any)["duration_seconds"] != float64(60) {
		t.Fatalf("efforts = %#v", efforts)
	}
}
