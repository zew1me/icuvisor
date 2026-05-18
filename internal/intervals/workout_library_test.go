package intervals

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sync/atomic"
	"testing"
)

func TestWorkoutLibraryClientListsFoldersAndWorkouts(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/athlete/i12345/folders":
			_, _ = w.Write([]byte(`[{"id":10,"type":"FOLDER","name":"Threshold","children":[{"id":2,"name":"FTP","type":"Ride","workout_doc":{"steps":[{"duration":600}]}}]}]`))
		case "/athlete/i12345/workouts":
			_, _ = w.Write([]byte(`[{"id":2,"name":"FTP","type":"Ride","folder_id":10,"workout_doc":{"steps":[{"duration":600}],"name":"raw"}}]`))
		default:
			t.Fatalf("path = %q, want folders or workouts", r.URL.Path)
		}
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	folders, err := client.ListWorkoutFolders(context.Background())
	if err != nil {
		t.Fatalf("ListWorkoutFolders() error = %v", err)
	}
	if len(folders) != 1 || folders[0].ID != "10" || len(folders[0].Children) != 1 || folders[0].Children[0].WorkoutDoc == nil {
		t.Fatalf("folders = %+v, want raw children and workout_doc", folders)
	}
	workouts, err := client.ListLibraryWorkouts(context.Background())
	if err != nil {
		t.Fatalf("ListLibraryWorkouts() error = %v", err)
	}
	if len(workouts) != 1 || workouts[0].ID != "2" || rawIDString(workouts[0].Raw["folder_id"]) != "10" {
		t.Fatalf("workouts = %+v, want folder_id preserved", workouts)
	}
	doc, ok := workouts[0].WorkoutDoc.(map[string]any)
	if !ok || doc["name"] != "raw" {
		t.Fatalf("workout_doc = %#v, want verbatim map", workouts[0].WorkoutDoc)
	}
}

func TestCreateLibraryWorkoutSendsWritableFieldsOnly(t *testing.T) {
	t.Parallel()

	wantBodyBytes, err := os.ReadFile("testdata/workout_library/create_request.json")
	if err != nil {
		t.Fatalf("read create request fixture: %v", err)
	}
	var wantBody map[string]any
	if err := json.Unmarshal(wantBodyBytes, &wantBody); err != nil {
		t.Fatalf("decode create request fixture: %v", err)
	}
	responseBody, err := os.ReadFile("testdata/workout_library/create_response.json")
	if err != nil {
		t.Fatalf("read create response fixture: %v", err)
	}

	var request struct {
		method  string
		path    string
		rawBody string
		body    map[string]any
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		var decoded map[string]any
		if err := json.Unmarshal(body, &decoded); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		request = struct {
			method  string
			path    string
			rawBody string
			body    map[string]any
		}{method: r.Method, path: r.URL.Path, rawBody: string(body), body: decoded}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(responseBody)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	description := "- 10m Z2"
	workout, err := client.CreateLibraryWorkout(context.Background(), WriteWorkoutParams{Name: " tp-077-fixture-create ", FolderID: " f-test-folder ", Sport: " Ride ", Description: &description})
	if err != nil {
		t.Fatalf("CreateLibraryWorkout() error = %v", err)
	}
	if workout.ID != "w-test-created" || workout.WorkoutDoc == nil {
		t.Fatalf("workout = %+v, want decoded workout with workout_doc", workout)
	}
	if request.method != http.MethodPost || request.path != "/athlete/i12345/workouts" {
		t.Fatalf("request = %#v, want POST athlete workouts", request)
	}
	if !reflect.DeepEqual(request.body, wantBody) {
		t.Fatalf("body = %#v, want fixture %#v", request.body, wantBody)
	}
	if _, ok := request.body["workout_doc"]; ok {
		t.Fatalf("body includes workout_doc: %#v", request.body)
	}
}

func TestCreateLibraryWorkoutRequiresWritableBasics(t *testing.T) {
	t.Parallel()

	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.WriteHeader(http.StatusTeapot)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	if _, err := client.CreateLibraryWorkout(context.Background(), WriteWorkoutParams{Sport: "Ride"}); err == nil {
		t.Fatal("CreateLibraryWorkout() error = nil, want required name error")
	}
	if _, err := client.CreateLibraryWorkout(context.Background(), WriteWorkoutParams{Name: "Tempo"}); err == nil {
		t.Fatal("CreateLibraryWorkout() error = nil, want required sport error")
	}
	if got := requests.Load(); got != 0 {
		t.Fatalf("requests = %d after basic validation, want no network I/O", got)
	}
	for _, params := range []WriteWorkoutParams{
		{Name: "Tempo", Sport: "Ride"},
		{Name: "Tempo", Sport: "Ride", FolderID: "   "},
	} {
		requests.Store(0)
		if _, err := client.CreateLibraryWorkout(context.Background(), params); err == nil {
			t.Fatalf("CreateLibraryWorkout(%+v) error = nil, want required folder ID error", params)
		}
		if got := requests.Load(); got != 0 {
			t.Fatalf("CreateLibraryWorkout(%+v) made %d request(s), want local validation before I/O", params, got)
		}
	}
}

func TestUpdateLibraryWorkoutSendsSparseWritableFieldsOnly(t *testing.T) {
	t.Parallel()

	var request struct {
		method  string
		path    string
		rawBody string
		body    map[string]any
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		var decoded map[string]any
		if err := json.Unmarshal(body, &decoded); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		request = struct {
			method  string
			path    string
			rawBody string
			body    map[string]any
		}{method: r.Method, path: r.URL.Path, rawBody: string(body), body: decoded}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"w-2","name":"Renamed","type":"Ride"}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	workout, err := client.UpdateLibraryWorkout(context.Background(), WriteWorkoutParams{WorkoutID: " w-2 ", Name: " Renamed ", NameSet: true})
	if err != nil {
		t.Fatalf("UpdateLibraryWorkout() error = %v", err)
	}
	if workout.ID != "w-2" {
		t.Fatalf("workout ID = %q, want decoded update response", workout.ID)
	}
	if request.method != http.MethodPut || request.path != "/athlete/i12345/workouts/w-2" {
		t.Fatalf("request = %#v, want PUT athlete workouts/{id}", request)
	}
	if request.rawBody != `{"name":"Renamed"}` {
		t.Fatalf("raw body = %s, want sparse name only", request.rawBody)
	}
	if len(request.body) != 1 || request.body["name"] != "Renamed" {
		t.Fatalf("body = %#v, want sparse name only", request.body)
	}
}

func TestUpdateLibraryWorkoutCanSendDescriptionTagsAndTopLevelFolder(t *testing.T) {
	t.Parallel()

	var body map[string]any
	var rawBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		decoded, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		rawBody = string(decoded)
		if err := json.Unmarshal(decoded, &body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"w-3","name":"Tempo","type":"Ride"}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	description := "- 15m 70%"
	_, err := client.UpdateLibraryWorkout(context.Background(), WriteWorkoutParams{WorkoutID: "w-3", FolderIDSet: true, Description: &description, DescriptionSet: true, Tags: []string{"base", "new"}, TagsSet: true})
	if err != nil {
		t.Fatalf("UpdateLibraryWorkout() error = %v", err)
	}
	if rawBody != `{"description":"- 15m 70%","folder_id":"","tags":["base","new"]}` {
		t.Fatalf("raw body = %s, want byte-identical sparse update", rawBody)
	}
	if body["folder_id"] != "" || body["description"] != description {
		t.Fatalf("body = %#v, want explicit top-level folder and description", body)
	}
	tags := body["tags"].([]any)
	if len(tags) != 2 || tags[0] != "base" || tags[1] != "new" {
		t.Fatalf("tags = %#v, want replacement tag list", tags)
	}
	if _, ok := body["type"]; ok {
		t.Fatalf("body = %#v, want omitted sport untouched", body)
	}
}

func TestUpdateLibraryWorkoutCanClearTags(t *testing.T) {
	t.Parallel()

	var rawBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		decoded, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		rawBody = string(decoded)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"w-4","name":"Tempo","type":"Ride"}`))
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	_, err := client.UpdateLibraryWorkout(context.Background(), WriteWorkoutParams{WorkoutID: "w-4", TagsSet: true})
	if err != nil {
		t.Fatalf("UpdateLibraryWorkout() error = %v", err)
	}
	if rawBody != `{"tags":[]}` {
		t.Fatalf("raw body = %s, want explicit empty tags", rawBody)
	}
}

func TestUpdateLibraryWorkoutRequiresIDAndField(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, "https://example.invalid", http.DefaultClient, RetryConfig{})
	if _, err := client.UpdateLibraryWorkout(context.Background(), WriteWorkoutParams{Name: "Renamed", NameSet: true}); err == nil {
		t.Fatal("UpdateLibraryWorkout() error = nil, want required workout ID error")
	}
	if _, err := client.UpdateLibraryWorkout(context.Background(), WriteWorkoutParams{WorkoutID: "w-1"}); err == nil {
		t.Fatal("UpdateLibraryWorkout() error = nil, want required sparse field error")
	}
}

func TestDeleteLibraryWorkoutSendsDeletePath(t *testing.T) {
	t.Parallel()

	var method, path string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		path = r.URL.Path
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client(), RetryConfig{})
	if err := client.DeleteLibraryWorkout(context.Background(), " w-4 "); err != nil {
		t.Fatalf("DeleteLibraryWorkout() error = %v", err)
	}
	if method != http.MethodDelete || path != "/athlete/i12345/workouts/w-4" {
		t.Fatalf("request = %s %s, want DELETE athlete workouts/{id}", method, path)
	}
}

func TestDeleteLibraryWorkoutRequiresID(t *testing.T) {
	t.Parallel()

	client := newTestClient(t, "https://example.invalid", http.DefaultClient, RetryConfig{})
	if err := client.DeleteLibraryWorkout(context.Background(), " "); err == nil {
		t.Fatal("DeleteLibraryWorkout() error = nil, want required workout ID error")
	}
}
