package tools

import (
	"context"
	"encoding/json"
	"math"
	"strings"
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

type fakeSportSettingsCreatorClient struct {
	fakeProfileClient
	setting intervals.SportSettings
	calls   []intervals.CreateSportSettingsParams
	err     error
}

func (f *fakeSportSettingsCreatorClient) CreateSportSettings(_ context.Context, params intervals.CreateSportSettingsParams) (intervals.SportSettings, error) {
	f.calls = append(f.calls, params)
	return f.setting, f.err
}

func TestCreateSportSettingsSchemaDocumentsThresholdOnlyCreation(t *testing.T) {
	t.Parallel()

	tool := newCreateSportSettingsTool(&fakeSportSettingsCreatorClient{}, &fakeProfileClient{}, "test", "UTC", false)
	if !strings.Contains(tool.Description, "cannot replace zones") || !strings.Contains(tool.Description, "does not yet have settings") {
		t.Fatalf("description = %q, want threshold-only missing-setting guidance", tool.Description)
	}
	if tool.Requirement != RequirementWrite || tool.EffectiveToolset().String() != "full" {
		t.Fatalf("tool = %#v, want full write tool", tool)
	}
	schema := tool.InputSchema.(map[string]any)
	if required := schema["required"].([]string); len(required) != 1 || required[0] != "sport" {
		t.Fatalf("required = %#v, want only sport", required)
	}
	props := schema["properties"].(map[string]any)
	for _, field := range []string{"sport", "ftp", "indoor_ftp", "threshold_hr", "threshold_pace"} {
		if _, ok := props[field]; !ok {
			t.Fatalf("schema missing field %q", field)
		}
	}
}

func TestCreateSportSettingsCreatesThresholdOnlySettings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		arguments     string
		created       intervals.SportSettings
		wantFTP       *int
		wantIndoorFTP *int
		wantHR        *int
		wantPace      float64
		wantUnits     string
		wantLoadType  string
		wantEchoKey   string
		wantEchoValue float64
		wantFields    []string
	}{
		{
			name:          "ride FTP and indoor FTP",
			arguments:     `{"sport":"Ride","ftp":285,"indoor_ftp":265}`,
			created:       intervals.SportSettings{ID: 12, Type: "Ride", FTP: 285, IndoorFTP: 265},
			wantFTP:       intPtr(285),
			wantIndoorFTP: intPtr(265),
			wantFields:    []string{"ftp", "indoor_ftp"},
		},
		{
			name:       "ride threshold heart rate",
			arguments:  `{"sport":"Ride","threshold_hr":171}`,
			created:    intervals.SportSettings{ID: 13, Type: "Ride", LTHR: 171},
			wantHR:     intPtr(171),
			wantFields: []string{"threshold_hr"},
		},
		{
			name:          "run threshold pace",
			arguments:     `{"sport":"Run","threshold_pace":{"value":300,"unit":"seconds_per_km"}}`,
			created:       intervals.SportSettings{ID: 14, Type: "Run", ThresholdPace: 1000.0 / 280, PaceUnits: "MINS_KM", PaceLoadType: "RUN"},
			wantPace:      1000.0 / 300,
			wantUnits:     "MINS_KM",
			wantLoadType:  "RUN",
			wantEchoKey:   "threshold_pace_seconds_per_km",
			wantEchoValue: 280,
			wantFields:    []string{"threshold_pace"},
		},
		{
			name:          "swim threshold pace",
			arguments:     `{"sport":"Swim","threshold_pace":{"value":90,"unit":"seconds_per_100y"}}`,
			created:       intervals.SportSettings{ID: 15, Type: "Swim", ThresholdPace: 100.0 / 85, PaceUnits: "SECS_100M", PaceLoadType: "SWIM"},
			wantPace:      metersPer100Yards / 90,
			wantUnits:     "SECS_100Y",
			wantLoadType:  "SWIM",
			wantEchoKey:   "threshold_pace_seconds_per_100m",
			wantEchoValue: 85,
			wantFields:    []string{"threshold_pace"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := &fakeSportSettingsCreatorClient{
				fakeProfileClient: fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}},
				setting:           tc.created,
			}
			tool := newCreateSportSettingsTool(client, client, "test", "UTC", false)

			result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(tc.arguments)})
			if err != nil {
				t.Fatalf("Handler() error = %v", err)
			}
			if len(client.calls) != 1 {
				t.Fatalf("create calls = %#v, want one", client.calls)
			}
			call := client.calls[0]
			if call.Sport != tc.created.Type || !sameIntPtr(call.FTP, tc.wantFTP) || !sameIntPtr(call.IndoorFTP, tc.wantIndoorFTP) || !sameIntPtr(call.ThresholdHR, tc.wantHR) {
				t.Fatalf("create call = %+v, want sport and requested threshold values", call)
			}
			if tc.wantPace == 0 {
				if call.ThresholdPace != nil {
					t.Fatalf("threshold pace = %+v, want nil", call.ThresholdPace)
				}
			} else if call.ThresholdPace == nil || math.Abs(call.ThresholdPace.Value-tc.wantPace) > 0.000001 || call.ThresholdPace.PaceUnits != tc.wantUnits || call.ThresholdPace.PaceLoadType != tc.wantLoadType {
				t.Fatalf("threshold pace = %+v, want %.6f with %s/%s", call.ThresholdPace, tc.wantPace, tc.wantUnits, tc.wantLoadType)
			}

			out := resultMap(t, result)
			settings := out["sport_settings"].(map[string]any)
			if settings["sport"] != tc.created.Type || settings["sport_setting_id"] != float64(tc.created.ID) {
				t.Fatalf("sport settings = %#v, want created sport and ID", settings)
			}
			if tc.wantEchoKey != "" {
				if got, ok := settings[tc.wantEchoKey].(float64); !ok || math.Abs(got-tc.wantEchoValue) > 0.000001 || settings["pace_units_source"] != tc.created.PaceUnits || settings["pace_load_type"] != tc.wantLoadType {
					t.Fatalf("sport settings = %#v, want returned threshold pace rendered in selected units", settings)
				}
			}
			meta := out["_meta"].(map[string]any)
			if meta["operation"] != "create" || !sameAnyStringSlice(meta["fields_created"].([]any), tc.wantFields) {
				t.Fatalf("meta = %#v, want create operation and fields %v", meta, tc.wantFields)
			}
		})
	}
}

func sameAnyStringSlice(got []any, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for index, value := range want {
		if got[index] != value {
			return false
		}
	}
	return true
}
