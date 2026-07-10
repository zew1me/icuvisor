# Plan Review — TP-228 Step 3

## Verdict: REVISE

No revised Step 3 implementation plan was submitted. `STATUS.md` remains an outcome checklist; its two R007 bullets identify constraints but do not say what code, schema, tests, and narrow stability approval will be changed. The prior R007 blocking concerns are therefore unaddressed.

Submit a concrete plan that specifies:

1. The request migration in `internal/tools/update_sport_settings.go`: remove `EffectiveDate`, add a presence-aware `recalc_hr_zones` (`*bool` or equivalent), resolve omission to `true`, and copy the resolved true/false value to `intervals.WriteSportSettingsParams.RecalcHRZones`.
2. Removal of `effective_date` from strict decoding, validation, schemas, examples, error/description text, internal params, response metadata, and tests. Confirm strict decoding rejects it before profile or writer calls.
3. Replacement of recompute/date claims with always-emitted `hr_zone_recalculation_requested`, equal exactly to the resolved request option, without claiming execution, pending work, or date scope.
4. The schema changes: only `sport` remains required; `recalc_hr_zones` is optional boolean with `default: true` and an LLM-readable description; regenerate snapshot and web catalogs with `make docs-tools`.
5. Table-driven coverage for omitted/default-true and explicit-false forwarding and metadata, plus legacy `effective_date`/unknown-field rejection with zero profile and writer calls.
6. A documented TP-228-only `effective_date` removal approval in schema stability checks. It must permit only removal of that property from `update_sport_settings` while retaining generic removal detection and proving unrelated removals still fail.

End with `go test ./internal/tools ./internal/toolchecks` and `make docs-tools`.
