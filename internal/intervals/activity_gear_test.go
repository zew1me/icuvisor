package intervals

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestActivityGearIDDecodesFromListAndDetailFixtures(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fixtures := map[string]string{
			"/athlete/i12345/activities": "testdata/activity_list_with_gear.json",
			"/activity/a-bike":           "testdata/activity_detail_with_gear.json",
		}
		fixture, ok := fixtures[r.URL.Path]
		if !ok {
			t.Fatalf("path = %q, want activity list or detail path", r.URL.Path)
		}
		data, err := os.ReadFile(fixture)
		if err != nil {
			t.Fatalf("read fixture %s: %v", fixture, err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	activities, err := client.ListActivities(context.Background(), ListActivitiesParams{Oldest: "2026-05-20", Fields: []string{"id", "gear_id"}})
	if err != nil {
		t.Fatalf("ListActivities() error = %v", err)
	}
	if len(activities) != 2 || activities[0].GearID != "123" || activities[1].GearID != "shoe-7" {
		t.Fatalf("activities = %#v, want gear_id decoded from list fixture", activities)
	}

	activity, err := client.GetActivity(context.Background(), "a-bike")
	if err != nil {
		t.Fatalf("GetActivity() error = %v", err)
	}
	if activity.GearID != "123" {
		t.Fatalf("activity.GearID = %q, want 123 from detail fixture", activity.GearID)
	}
}
