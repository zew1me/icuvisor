# Review R011 — Code: Step 4 (Testing & Verification)

**Task:** TP-230 — Distinguish ATP-generated notes from personal calendar notes
**Step:** 4 — Testing & Verification
**Type:** Code review
**Verdict:** APPROVE

---

## What this step produced

Step 4 is a pure verification gate. The only committed changes are:

- `STATUS.md` — all five Step 4 checkboxes marked complete
- `.reviews/R010-plan-step4.md` — plan-review artifact confirming the gates passed

No implementation code was modified in this step. The correctness claim rests entirely on whether the underlying implementation (Steps 1–3) holds up under the full quality suite.

---

## Independent verification

All gates re-run with `--count=1` to bypass the cache:

| Gate | Command | Result |
|------|---------|--------|
| Full test suite | `go test ./... -count=1` | All 27 packages `ok` |
| Race suite | `make test-race` (`-count=1` is forced in the Makefile target) | All 27 packages `ok`, no races |
| Lint | `make lint` | `0 issues.` |
| Build | `make build` | Binary emitted to `bin/icuvisor` |
| Docs regen | `make docs-tools && git diff --check` | No output — generated artifacts are current, no trailing whitespace |

Working tree is clean; only the untracked `.reviewer-state.json` is present.

---

## Implementation review (Steps 1–3 gate)

Since Step 4 gates the full implementation, I reviewed the key paths in `internal/tools/get_annual_training_plan.go` and the two test files.

**Provenance classification is correct and locale-neutral.**  
`strings.TrimSpace(stringValue(event.PlanApplied)) != ""` is the single, field-driven split point in `annualTrainingPlanPeriodizationEvents`. Null, empty, and whitespace-only `plan_applied` values all route to `contextNotes`; non-empty values route to `notes`. No English keywords, no `for_week`, no name matching.

**`recovery_hint` and `annualTrainingPlanRecoveryHint` are gone.**  
Existing tests that previously asserted `notes[0]["recovery_hint"] == true` now assert `notes[0]["status"] == "atp_generated"` and explicitly check that `"recovery_hint"` is absent. The removal is complete and tested.

**Personal notes never contribute to ATP counts.**  
- `periodizationCount = len(classified.plans) + len(classified.targets) + len(classified.notes)` — excludes `contextNotes`.
- `_meta.PeriodizationEventCount` and `_meta.ATPNoteEventCount` exclude context notes.
- Week loops for `atp_note_ids`/`atp_note_count` and `context_note_ids`/`context_note_count` are separate and correct.
- Summary fields renamed to `atp_note_count`/`context_note_count`.

**`plan_applied` field is correct for both collection types.**  
The `annualTrainingPlanNote` struct carries `PlanApplied string \`json:"plan_applied,omitempty"\``. For context notes the trimmed value is always `""` (they classified as context because `plan_applied` was null/empty), so the field is omitted from JSON output via `omitempty`. For ATP notes the non-empty timestamp is present. The provenance test asserts both (`note.PlanApplied != planApplied` and `note.PlanApplied != ""`).

**`unavailable` is set correctly for context-only responses.**  
When `periodizationCount == 0` but `len(classified.contextNotes) > 0`, `Unavailable.Reason` is `"no_periodization_events"` and `Detail` mentions `context_notes` so the LLM knows why the plan is missing and where neutral context lives. Covered by `TestAnnualTrainingPlanPersonalContextOnlyRetainsContextAndUnavailable`.

**Projection bridge is unaffected.**  
`annualTrainingPlanBridge` iterates `weeks` and looks only at `TargetEventCount`, `PartialWeek`, `MissingLoadTargetCount`, and `LoadTarget`. Note classification does not touch it. `TestAnnualTrainingPlanOneDayTargetKeepsISOWeekAndProjectionBridge` covers the one-day TARGET + personal-context-note combination and confirms the bridge row is unchanged.

**New test file covers the regression matrix.**  
`get_annual_training_plan_provenance_test.go` has four focused tests:
1. Mixed ATP / personal / localized fixture — verifies counts, ordering, week associations, terse output shape.
2. Context-only — verifies `unavailable` and that context notes are retained.
3. `include_full` terse/full table — verifies `Full` is present only when opted in, for both collections.
4. One-day TARGET + personal context — verifies ISO-week boundary and projection bridge correctness.

**Existing tests correctly updated.**  
Fixtures that previously had `NOTE` rows without `plan_applied` are updated to include `"plan_applied":"2025-12-15T09:30:00Z"` so they remain ATP notes in the new model. Summary and week-level field assertions updated from `note_count`/`recovery_note_count` to `atp_note_count`/`context_note_count`.

---

## No issues found

The step commit is minimal and accurate. The quality gates pass cleanly under independent forced re-execution. The implementation is coherent, the tests are specific, and the provenance-only classification rule is applied consistently through the entire response path.

**APPROVE** — the worker may proceed to Step 5 (Documentation & Delivery).
