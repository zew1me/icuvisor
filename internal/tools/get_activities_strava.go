package tools

import (
	"fmt"
	"strings"

	"github.com/ricardocabral/icuvisor/internal/intervals"
)

func isStravaBlocked(activity intervals.Activity) bool {
	source := strings.ToLower(stringValue(activity.Source))
	note := strings.ToLower(stringValue(activity.Note))
	if strings.Contains(source, "strava") || strings.Contains(note, "strava") {
		return true
	}
	if source != "" {
		return false
	}
	meaningful := 0
	for _, key := range []string{"name", "type", "distance", "icu_distance", "moving_time", "elapsed_time"} {
		if value, ok := activity.Raw[key]; ok && value != nil && strings.TrimSpace(fmt.Sprint(value)) != "" {
			meaningful++
		}
	}
	if meaningful > 0 {
		return false
	}
	if len(activity.Raw) == 0 {
		return true
	}
	_, hasHiddenNote := activity.Raw["_note"]
	if hasHiddenNote {
		return true
	}
	stubKeys := map[string]bool{"id": true, "icu_athlete_id": true, "athlete_id": true, "start_date_local": true, "start_date": true, "external_id": true}
	nullableStubKeys := map[string]bool{"name": true, "type": true, "sub_type": true, "distance": true, "icu_distance": true, "moving_time": true, "elapsed_time": true, "source": true, "_note": true}
	for key, value := range activity.Raw {
		if stubKeys[key] {
			continue
		}
		if value == nil && nullableStubKeys[key] {
			continue
		}
		return false
	}
	return true
}
