package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestGetActivitiesAroundRegistrationMetadata(t *testing.T) {
	t.Parallel()

	client := newFakeActivitiesClient(t, nil, "metric")
	tool := newGetActivitiesAroundTool(client, client, nil, nil, "test", "UTC", false)
	if tool.Name != getActivitiesAroundName {
		t.Fatalf("Name = %q, want %q", tool.Name, getActivitiesAroundName)
	}
	if !strings.Contains(tool.Description, "known reference activity_id") || !strings.Contains(tool.Description, "use get_activities instead") {
		t.Fatalf("description = %q, want known-reference routing guidance", tool.Description)
	}
	props := tool.InputSchema.(map[string]any)["properties"].(map[string]any)
	limit := props["limit"].(map[string]any)
	if limit["default"] != defaultActivitiesAroundLimit || limit["minimum"] != 1 || limit["maximum"] != maxActivitiesAroundLimit {
		t.Fatalf("limit schema = %#v, want pinned default/range", limit)
	}
	outputDescription := tool.OutputSchema.(map[string]any)["description"].(string)
	for _, want := range []string{"reference_activity_id", "get_activities terse activity shaping", "Strava unavailable markers", "Empty results"} {
		if !strings.Contains(outputDescription, want) {
			t.Fatalf("output description = %q, missing %q", outputDescription, want)
		}
	}
}

func TestGetActivitiesAroundSuccessReusesActivityRows(t *testing.T) {
	t.Parallel()

	client := newFakeActivitiesClient(t, nil, "metric")
	client.profile.Timezone = "America/Sao_Paulo"
	client.aroundActivities = decodeActivityPage(t,
		`{"id":"near-1","name":"Tempo","type":"Run","start_date_local":"2026-05-01T07:00:00","start_date":"2026-05-01T10:00:00Z","distance":5000,"moving_time":1500,"calories":300,"carbs_ingested":40,"carbs_used":80,"stream_types":["time"],"has_weather":true,"average_weather_temp":22.5}`,
	)
	tool := newGetActivitiesAroundTool(client, client, nil, nil, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":" ref-1 ","limit":3}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	if len(client.aroundCalls) != 1 || client.aroundCalls[0].ActivityID != "ref-1" || client.aroundCalls[0].Limit != 3 {
		t.Fatalf("around calls = %#v, want trimmed activity_id and limit", client.aroundCalls)
	}
	payload := resultMap(t, result)
	if payload["reference_activity_id"] != "ref-1" {
		t.Fatalf("reference_activity_id = %#v, want ref-1", payload["reference_activity_id"])
	}
	activities := payload["activities"].([]any)
	if len(activities) != 1 {
		t.Fatalf("activities length = %d, want 1", len(activities))
	}
	row := activities[0].(map[string]any)
	if row["activity_id"] != "near-1" || row["distance_km"] != float64(5) || row["pace_seconds_per_km"] != float64(300) || row["calories_burned"] != float64(300) {
		t.Fatalf("row = %#v, want compact activity row fields", row)
	}
	if _, ok := row["weather"].(map[string]any); !ok {
		t.Fatalf("row weather = %#v, want historical weather summary", row["weather"])
	}
	meta := payload["_meta"].(map[string]any)
	if meta["limit"] != float64(3) || meta["returned_count"] != float64(1) || meta["timezone"] != "America/Sao_Paulo" {
		t.Fatalf("_meta = %#v, want limit/count/timezone", meta)
	}
	semantics := meta["field_semantics"].(map[string]any)
	if !strings.Contains(semantics["calories_burned"].(string), "Active/exercise calories") {
		t.Fatalf("field semantics = %#v, want activity calorie guidance", semantics)
	}
}

func TestGetActivitiesAroundEmptyResponseGuidance(t *testing.T) {
	t.Parallel()

	client := newFakeActivitiesClient(t, nil, "metric")
	tool := newGetActivitiesAroundTool(client, client, nil, nil, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"ref-1"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	payload := resultMap(t, result)
	activities := payload["activities"].([]any)
	if len(activities) != 0 {
		t.Fatalf("activities length = %d, want empty", len(activities))
	}
	meta := payload["_meta"].(map[string]any)
	if meta["empty_reason"] != "no_activities_returned_for_reference_activity" || !strings.Contains(meta["guidance"].(string), "Use get_activities") {
		t.Fatalf("_meta = %#v, want empty reason and routing guidance", meta)
	}
	if len(client.aroundCalls) != 1 || client.aroundCalls[0].Limit != defaultActivitiesAroundLimit {
		t.Fatalf("around calls = %#v, want default limit", client.aroundCalls)
	}
}

func TestGetActivitiesAroundValidationFailures(t *testing.T) {
	t.Parallel()

	client := newFakeActivitiesClient(t, nil, "metric")
	tool := newGetActivitiesAroundTool(client, client, nil, nil, "test", "UTC", false)
	for _, raw := range []string{
		`{}`,
		`{"activity_id":"   "}`,
		`{"activity_id":"ref","limit":0}`,
		`{"activity_id":"ref","limit":-1}`,
		`{"activity_id":"ref","limit":51}`,
		`{"activity_id":"ref","count":5}`,
	} {
		_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(raw)})
		if err == nil {
			t.Fatalf("Handler(%s) error = nil, want validation error", raw)
		}
		message, ok := PublicErrorMessage(err)
		if !ok || !strings.Contains(message, "invalid get_activities_around arguments") {
			t.Fatalf("PublicErrorMessage = %q, %v; want invalid arguments for %s", message, ok, raw)
		}
	}
}

func TestGetActivitiesAroundStravaCaveatAndIncludeFull(t *testing.T) {
	t.Parallel()

	client := newFakeActivitiesClient(t, nil, "metric")
	client.aroundActivities = decodeActivityPage(t, `{"id":"hidden","source":"strava","_note":"Strava activity hidden","name":null}`)
	tool := newGetActivitiesAroundTool(client, client, nil, nil, "test", "UTC", false)

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"ref","include_full":true}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	payload := resultMap(t, result)
	row := payload["activities"].([]any)[0].(map[string]any)
	if row["strava_imported"] != true {
		t.Fatalf("row = %#v, want strava_imported", row)
	}
	unavailable := row["unavailable"].(map[string]any)
	if unavailable["reason"] != "strava_blocked" || unavailable["workaround"] == "" {
		t.Fatalf("unavailable = %#v, want Strava caveat", unavailable)
	}
	full := row["full"].(map[string]any)
	if _, ok := full["name"]; !ok || full["name"] != nil {
		t.Fatalf("full = %#v, want raw null preserved", full)
	}
}

func TestGetActivitiesAroundClientErrors(t *testing.T) {
	t.Parallel()

	client := newFakeActivitiesClient(t, nil, "metric")
	client.aroundErr = errors.New("upstream unavailable")
	tool := newGetActivitiesAroundTool(client, client, nil, nil, "test", "UTC", false)

	_, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"activity_id":"ref"}`)})
	if err == nil {
		t.Fatal("Handler() error = nil, want client error")
	}
	message, ok := PublicErrorMessage(err)
	if !ok || !strings.Contains(message, "could not fetch activities around activity") {
		t.Fatalf("PublicErrorMessage = %q, %v; want fetch error", message, ok)
	}
}
