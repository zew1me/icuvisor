package prompts

import (
	"strings"
	"testing"
)

func TestMastersPlanReviewEvidenceBoundaries(t *testing.T) {
	t.Parallel()

	prompt := MastersPlanReviewPrompt()
	text := renderPromptText(t, prompt, map[string]string{
		"planned_start":          "2026-05-18",
		"planned_end":            "2026-06-01",
		"history_lookback_days":  "28",
		"baseline_lookback_days": "56",
		"race_date":              "2026-06-07",
		"race_name":              "A Race",
	})

	for _, tc := range []struct {
		name string
		want string
	}{
		{name: "athlete-local sequence", want: "Establish the athlete-local timezone"},
		{name: "non-overlapping windows", want: "Partition non-overlapping personal-baseline/history, completed, planned, and race windows"},
		{name: "partial coverage", want: "label the coverage partial and do not treat it as complete"},
		{name: "race scenario", want: "scenario anchor rather than observed race evidence"},
		{name: "separate observed evidence", want: "Observed tool evidence (tool, athlete-local window, freshness/coverage)"},
		{name: "separate preferences", want: "Athlete-stated preferences (availability and requested duration only)"},
		{name: "separate interpretation", want: "Cautious interpretation"},
		{name: "separate questions", want: "Insufficient evidence and focused questions"},
		{name: "separate proposals", want: "Reviewable proposals"},
		{name: "baseline metadata", want: "status, n_baseline, n_current, min_samples, missing-day counts, freshness_status, caveats, `_meta.method`, and `_meta.formula_ref`"},
		{name: "hard session fallback", want: "Titles, aggregate load, calendar proximity, zones that are absent or invalid, and age cannot classify a session as hard"},
		{name: "projection default boundary", want: "never present default weekly_ramp_pct or recovery_week_cadence values as plan evidence or a masters recommendation"},
		{name: "no inferred constraints", want: "not inferred hard constraints or an implied session count"},
		{name: "ambiguous hard sessions", want: "ambiguous or unavailable hard-session or plan detail"},
		{name: "absent zones", want: "absent or invalid zones"},
		{name: "incomplete history", want: "short, partial, truncated, or missing historical coverage"},
		{name: "wellness freshness", want: "missing, stale, or partial wellness"},
		{name: "provider-native readiness", want: "missing or provider-native readiness"},
		{name: "race gap", want: "missing race context"},
		{name: "projection gap", want: "insufficient explicit projection targets"},
		{name: "gap behavior", want: "name the missing evidence, make no comparison or conclusion for that affected dimension, and ask one focused question"},
		{name: "absolute read-only", want: "This workflow is absolutely read-only: never call write or delete tools, including after approval"},
		{name: "unapplied proposal", want: "conditional, unapplied proposal"},
		{name: "no age policy", want: "Masters is an audience label only, never an age-derived policy or universal cutoff"},
		{name: "no medical claims", want: "Do not make medical, diagnostic, treatment, or injury-risk claims"},
		{name: "no score", want: "do not invent a black-box readiness or risk score"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if !strings.Contains(text, tc.want) {
				t.Fatalf("masters plan review missing %q:\n%s", tc.want, text)
			}
		})
	}

	for _, arg := range prompt.Arguments {
		if arg.Name == "age" || arg.Name == "date_of_birth" {
			t.Fatalf("masters plan review exposes prohibited argument %q", arg.Name)
		}
	}
	for _, tool := range renderedPromptTools(t, prompt, nil) {
		if strings.Contains(tool, "write") || strings.Contains(tool, "delete") {
			t.Fatalf("masters plan review exposes prohibited tool %q", tool)
		}
	}
}
