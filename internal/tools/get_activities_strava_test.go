package tools

import (
	"testing"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

func TestIsStravaBlockedGoldenBranches(t *testing.T) {
	t.Parallel()

	stringPtr := func(value string) *string { return &value }
	tests := []struct {
		name     string
		activity intervals.Activity
		want     bool
	}{
		{
			name: "Strava upstream marker present",
			activity: intervals.Activity{
				Source: stringPtr("Strava"),
				Raw:    map[string]any{"id": "strava-marker", "source": "Strava", "start_date_local": "2026-01-01T07:00:00"},
			},
			want: true,
		},
		{
			name: "N/A HR and cadence on power-only file is not a Strava stub",
			activity: intervals.Activity{
				Raw: map[string]any{
					"id":                 "power-only",
					"start_date_local":   "2026-01-01T07:00:00",
					"average_heartrate":  "N/A",
					"average_cadence":    "N/A",
					"weighted_average_w": 247,
				},
			},
			want: false,
		},
		{
			name: "manual entry with meaningful fields is not blocked",
			activity: intervals.Activity{
				Source: stringPtr("manual"),
				Raw:    map[string]any{"id": "manual", "source": "manual", "name": "Mobility", "type": "Other", "start_date_local": "2026-01-01T07:00:00"},
			},
			want: false,
		},
		{
			name:     "empty upstream object is treated as blocked stub",
			activity: intervals.Activity{Raw: map[string]any{}},
			want:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := isStravaBlocked(tc.activity); got != tc.want {
				t.Fatalf("isStravaBlocked() = %v, want %v", got, tc.want)
			}
		})
	}
}
