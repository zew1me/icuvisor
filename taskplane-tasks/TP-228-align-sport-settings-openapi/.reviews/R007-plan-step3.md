# Plan Review — TP-228 Step 3

## Verdict: REVISE

No Step 3 implementation plan was submitted. The unchecked STATUS checklist repeats outcomes, but does not specify the migration through the current request/response types, the required presence-aware default, or how the deliberate breaking schema removal will be approved without weakening the schema-stability guard.

Revise the plan to specify:

1. In `internal/tools/update_sport_settings.go`, replace `EffectiveDate` with an optional/presence-aware `recalc_hr_zones` input (for example `*bool`), resolve omission to `true` in decoding, and copy the resolved value into `intervals.WriteSportSettingsParams.RecalcHRZones`. An explicit `false` must survive unchanged. Remove `EffectiveDate` from the internal write params as well; it is now inert but retaining it leaves a date-capable API surface contrary to R004.
2. Remove `effective_date` from the strict request type, required schema list, validation, public error strings, descriptions, examples, response metadata, and every corresponding assertion. `additionalProperties: false`/`DecodeStrict` must reject legacy `effective_date` before profile lookup or any writer call; it must not be accepted and ignored.
3. Replace `recompute_pending` and date metadata with an always-emitted `hr_zone_recalculation_requested` boolean whose value is exactly the resolved query option. Do not claim that recalculation ran, is pending, or is scoped to historical dates. Retain unrelated fields such as zone-delete-mode and unit metadata.
4. Update the input schema to make `sport` the only required existing selector, define optional boolean `recalc_hr_zones` with JSON-Schema `default: true` and an LLM-readable description, and update examples. Update the committed snapshot and run `make docs-tools` to refresh both generated web JSON files.
5. Add table-driven tool coverage for omitted/default-true and explicit-false forwarding and response metadata; update the existing normal-write/pace cases to omit the removed argument. Add a legacy `effective_date`/unknown-argument case that asserts the terse public validation error and zero profile/writer calls.
6. State the explicit, narrow schema-stability approval mechanism for this safety correction. The current `CheckSchemaStability` intentionally reports `effective_date` removal as `property-removed`; do not weaken its generic removal check. Add a TP-228-specific, documented exception/approval test or equivalent narrowly scoped policy, with coverage that the exception permits only this property on `update_sport_settings` and continues to reject all unrelated removals.

End the plan with `make docs-tools` and `go test ./internal/tools ./internal/toolchecks`.
