package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/ricardocabral/icuvisor/internal/intervals"
	"github.com/ricardocabral/icuvisor/internal/safety"
)

func fixedSeasonPlanClock() func() time.Time {
	return func() time.Time { return time.Date(2026, 7, 6, 12, 0, 0, 0, time.UTC) }
}

func TestProposeAnnualTrainingPlanRegistrationMetadata(t *testing.T) {
	t.Parallel()

	tool := newProposeAnnualTrainingPlanToolWithClock(&fakeProfileClient{}, "test", "UTC", false, fixedSeasonPlanClock())
	if tool.Name != proposeAnnualTrainingPlanName || !strings.Contains(tool.Description, "season plan") || !strings.Contains(tool.Description, "do not roll ATP/projection math") || !strings.Contains(tool.Description, "do not write calendar data") {
		t.Fatalf("tool metadata = %#v, want read-only proposal activation hint", tool)
	}
	if tool.EffectiveToolset() != safety.ToolsetFull || tool.RequiresWrite() {
		t.Fatalf("toolset/write = %q/%t, want full read-only", tool.EffectiveToolset(), tool.RequiresWrite())
	}
	props := tool.InputSchema.(map[string]any)["properties"].(map[string]any)
	for _, name := range []string{"goal_date", "start_date", "target_weekly_load", "current_weekly_load", "max_hours_per_week", "sports", "strength_context", "include_full"} {
		if _, ok := props[name]; !ok {
			t.Fatalf("schema missing %s: %#v", name, props)
		}
	}
}

func TestProposeAnnualTrainingPlanStartsNextMondayAfterToday(t *testing.T) {
	t.Parallel()

	client := &fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}}
	tool := newProposeAnnualTrainingPlanToolWithClock(client, "test", "UTC", false, fixedSeasonPlanClock())

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"goal_date":"2026-08-03","current_weekly_load":300,"target_weekly_load":420}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	out := resultMap(t, result)
	summary := out["summary"].(map[string]any)
	if summary["start_date"] != "2026-07-13" {
		t.Fatalf("start_date = %v, want following Monday after Monday today", summary["start_date"])
	}
	weeks := out["weekly_targets"].([]any)
	if weeks[0].(map[string]any)["week_start_date"] != "2026-07-13" {
		t.Fatalf("first week = %#v, want no current partial week", weeks[0])
	}

	_, err = tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"start_date":"2026-07-08","goal_date":"2026-08-03"}`)})
	if err == nil || !strings.Contains(err.Error(), invalidSeasonPlanProposalMessage) {
		t.Fatalf("non-Monday start err = %v, want user error", err)
	}
}

func TestProposeAnnualTrainingPlanBoundaryWeeksBelongToNewPhase(t *testing.T) {
	t.Parallel()

	client := &fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}}
	tool := newProposeAnnualTrainingPlanToolWithClock(client, "test", "UTC", false, fixedSeasonPlanClock())

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"start_date":"2026-07-13","goal_date":"2026-08-03","current_weekly_load":300,"target_weekly_load":500}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	out := resultMap(t, result)
	phases := out["phases"].([]any)
	if len(phases) != 3 {
		t.Fatalf("phases = %#v, want base/build/race_taper for 4-week horizon with zero peak omitted", phases)
	}
	peak := phases[1].(map[string]any)
	if peak["phase_id"] != "phase_02_peak" || peak["start_week_index"] != float64(2) || peak["start_date"] != "2026-07-20" {
		t.Fatalf("peak phase = %#v, want week-2 boundary owned by new phase", peak)
	}
	weeks := out["weekly_targets"].([]any)
	week2 := weeks[1].(map[string]any)
	if week2["week_start_date"] != "2026-07-20" || week2["phase_id"] != "phase_02_peak" || week2["phase_type"] != "peak" {
		t.Fatalf("week 2 = %#v, want new phase ownership on boundary", week2)
	}
}

func TestProposeAnnualTrainingPlanMissingDataFallbacksAndWarnings(t *testing.T) {
	t.Parallel()

	client := &fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}}
	tool := newProposeAnnualTrainingPlanToolWithClock(client, "test", "UTC", false, fixedSeasonPlanClock())

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"goal_date":"2026-07-27"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	out := resultMap(t, result)
	assumptions := noticeCodes(out["assumptions"].([]any))
	warnings := noticeCodes(out["warnings"].([]any))
	for _, code := range []string{"default_current_weekly_load", "default_target_weekly_load", "current_hours_from_load", "target_hours_from_load"} {
		if !assumptions[code] {
			t.Fatalf("assumption codes = %#v, missing %s", assumptions, code)
		}
	}
	for _, code := range []string{"missing_current_load", "missing_power_curve_profile"} {
		if !warnings[code] {
			t.Fatalf("warning codes = %#v, missing %s", warnings, code)
		}
	}
}

func TestProposeAnnualTrainingPlanCapsHoursAndEchoesContextAsAssumptions(t *testing.T) {
	t.Parallel()

	client := &fakeProfileClient{profile: intervals.AthleteWithSportSettings{ID: "i12345", PreferredUnits: "metric", Timezone: "UTC"}}
	tool := newProposeAnnualTrainingPlanToolWithClock(client, "test", "UTC", false, fixedSeasonPlanClock())

	result, err := tool.Handler(context.Background(), Request{Name: tool.Name, Arguments: json.RawMessage(`{"goal_date":"2026-08-10","current_weekly_load":500,"target_weekly_load":900,"target_hours_per_week":25,"max_hours_per_week":8,"sports":["Ride","Run"],"strength_context":"two gym sessions, maintenance only"}`)})
	if err != nil {
		t.Fatalf("Handler() error = %v", err)
	}
	out := resultMap(t, result)
	weeks := out["weekly_targets"].([]any)
	for _, item := range weeks {
		week := item.(map[string]any)
		if week["target_hours"].(float64) > 8 {
			t.Fatalf("week target_hours = %#v, want capped at 8", week)
		}
	}
	assumptions := noticeCodes(out["assumptions"].([]any))
	warnings := noticeCodes(out["warnings"].([]any))
	for _, code := range []string{"sports_context_input_only", "strength_context_input_only"} {
		if !assumptions[code] {
			t.Fatalf("assumption codes = %#v, missing %s", assumptions, code)
		}
	}
	for _, code := range []string{"infeasible_hours_cap", "multi_sport_not_allocated", "strength_not_first_class"} {
		if !warnings[code] {
			t.Fatalf("warning codes = %#v, missing %s", warnings, code)
		}
	}
}

func noticeCodes(rows []any) map[string]bool {
	out := make(map[string]bool, len(rows))
	for _, row := range rows {
		item := row.(map[string]any)
		out[item["code"].(string)] = true
	}
	return out
}
