# Plan Review — TP-228 Step 4

## Verdict: REVISE

No Step 4 implementation plan was submitted; STATUS only marks the step in progress and repeats its outcome checklist. Much of the requested coverage already exists from Steps 2–3, so the revised plan should identify what will be retained versus strengthened rather than duplicating tests.

Revise the plan to specify:

1. In `internal/intervals/sport_settings_openapi_contract_test.go`, retain the true/false table and assert the **entire** update query, not only `Query().Get("recalcHrZones")`. The current assertion would still pass if an unsupported key such as `oldest` were added. Assert the sole encoded query is `recalcHrZones=true` or `recalcHrZones=false`, alongside exact PUT/path and the existing one-field sparse body assertion.
2. Retain the apply contract assertion for exact PUT/path, empty `RawQuery`, zero-byte body, and absent JSON content type. State that this is a direct `ApplySportSettings` test and does not call a live service.
3. In `internal/intervals/sport_settings_test.go`, make the no-implicit-apply regression explicit: count update requests, fail on `/apply` or any other path, and assert exactly one update request after `UpdateSportSettings` returns. This protects against both an implicit apply and accidental duplicate writes.
4. In `internal/tools/update_sport_settings_test.go`, retain a table-driven strict-decoding test covering legacy `effective_date`, an unrelated unknown property, and invalid `recalc_hr_zones` (including explicit null). For every case assert the exact terse public validation message and zero profile and writer calls, proving rejection occurs before upstream selection or mutation.
5. Limit Step 4 to test changes unless a regression exposes a defect, then run `go test ./internal/intervals ./internal/tools` as required.
