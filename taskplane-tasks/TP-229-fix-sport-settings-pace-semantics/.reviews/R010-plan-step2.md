# Plan Review — TP-229 Step 2

## Verdict: REVISE

No Step 2 implementation plan was submitted: the unchecked checklist repeats the outcome but does not say how the remaining model/response contract will change. This is material because the current decoder has no `pace_load_type` field at all, while the Step 1 response changes already occupy much of the stated Step 2 scope.

Revise the plan to specify:

1. Add the exact `json:"pace_load_type"` typed field to `intervals.SportSettings`, alongside the canonical `threshold_pace`, `pace_units`, `pace_zones`, and names fields. State the preservation path into the sport-setting data consumed by Step 3 (and, if exposed by `get_athlete_profile`, the exact unambiguous response field). It must preserve the upstream `RUN`/`SWIM` value without deriving it from the sport name or `pace_units`.
2. Treat `ThresholdPace` exclusively as m/s in profile shaping. For every recognized display enum, call the Step 1 canonical converter and populate exactly the matching `threshold_pace_seconds_per_{km,mile,100m,100y,500m}` field; `pace_units` chooses only that display distance. Do not reintroduce `PaceThreshold` as a duration fallback or convert according to the athlete's general preferred-unit setting.
3. Copy `pace_zones` unchanged to `pace_zones_percent_of_threshold`, retain matching names, and assert serialized output contains no `pace_zones_seconds_per_*` field. Include the percentage legend in response metadata, rather than making a percentage array appear to be seconds.
4. Define the failure/fallback matrix: unknown non-empty `pace_units` retains the trimmed raw token and `_meta.unknown_unit`, emits a finite positive `threshold_pace_meters_per_second` when available, and still returns unchanged percentage zones; recognized `NONE` and a known-unit display-conversion overflow take the same unambiguous m/s fallback without false unknown-unit metadata. None may fail the profile request.
5. Add table-driven `internal/athleteprofile/profile_test.go` coverage for JSON decoding plus shaped run/km, run/mile, swim/100m, swim/100yd, and row/500m values, the `pace_load_type` preservation case, percentage-zone/name preservation, legacy duration-zone omission, unknown-unit fallback, `NONE`, and overflow. Keep the synthetic fixture change for Step 4, but update `internal/tools/get_athlete_profile_test.go` only where the shared public serialized response needs a contract assertion. Then run `go test ./internal/athleteprofile ./internal/tools -run 'Profile|Pace'`.
