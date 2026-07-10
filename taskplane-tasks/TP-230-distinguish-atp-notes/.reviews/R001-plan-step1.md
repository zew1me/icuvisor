# Plan Review — TP-230 Step 1

## Verdict: REVISE

No Step 1 implementation plan was submitted: `STATUS.md` only repeats the unchecked checkpoint checklist. It does not make the response-contract decisions needed to safely change a model-visible ATP summary.

A revised plan must specify:

1. The exact terse response shape and names for the two classes. State whether ATP-generated notes remain in `notes` and personal `NOTE` rows move to a separate `context_notes` collection (or the alternative), and include a terse, model-visible provenance/status value for every returned note. `include_full` alone is insufficient because it hides `plan_applied` in the default response.
2. The exact count/association migration. Define summary and per-week fields for ATP notes versus personal context, which existing `note_count`, `recovery_note_count`, and `note_ids` fields are removed/renamed, and whether neutral context gets separately labelled week associations. Personal rows must not contribute to any ATP count or ID list.
3. The recovery conclusion policy. `annualTrainingPlanRecoveryHint` currently derives `recovery_hint` and `recovery_note_count` from English text. The plan must explicitly remove that locale-dependent inference (or identify an independent upstream, locale-neutral recovery field); `plan_applied` establishes ATP provenance, not recovery semantics. It must not replace the heuristic with another name/tag keyword list.
4. The classification boundary and predicate: use `strings.TrimSpace(event.PlanApplied) != ""` for ATP provenance, without `for_week`, names, descriptions, or tags. Also resolve whether the provenance gate applies only to `NOTE` rows (the stated note-shaping scope) or to PLAN/TARGET inclusion too, since the mission calls `plan_applied` provenance for all three categories and the existing classifier admits all category-matching rows.
5. Deterministic ordering for both output collections and all week association ID lists, including the tie-breakers (the current source ordering is start date, updated, ID, while week IDs are sorted lexically), plus the exact terse/full behavior for the new provenance fields and raw event payloads.
6. The Step 2 regression matrix: a null/whitespace `plan_applied` personal `Travel — Rest` note that cannot produce ATP/recovery counts; localized ATP NOTE rows with non-empty `plan_applied`; multi-day notes and week boundaries; terse versus `include_full`; stable ordering; and unchanged TARGET-only projection-bridge output. Update existing assertions that presently require `recovery_hint` and `recovery_note_count`.

The current implementation counts every category `NOTE` as periodization context and applies the prohibited English keyword heuristic in `internal/tools/get_annual_training_plan.go`; these contract decisions are necessary before implementation and schema regeneration can be reviewed.
