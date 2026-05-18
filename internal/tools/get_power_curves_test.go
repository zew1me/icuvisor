package tools

import (
	"context"
	"encoding/json"
	"testing"
)

func TestGetPowerCurvesToolShapes(t *testing.T) {
	t.Parallel()

	client := newFakeFitnessMetricsClient(t)
	tool := newGetPowerCurvesTool(client, "test", false)
	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"oldest":"2026-05-01","newest":"2026-05-07","duration_seconds":[60,300]}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	got := resultMap(t, result)
	points := got["points"].([]any)
	if len(points) != 2 || points[1].(map[string]any)["watts"] != float64(310) {
		t.Fatalf("points = %#v", points)
	}
	if _, ok := got["full"]; ok {
		t.Fatal("full payload present in terse power curve response")
	}
}
