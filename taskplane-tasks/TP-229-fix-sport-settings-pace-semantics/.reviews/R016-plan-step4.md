# Plan Review — TP-229 Step 4

## Verdict: REVISE

The proposed fixture replacement and regression matrix cover the primary conversion cases, but the plan does not yet lock the checked-in fixture or remove the remaining duration-encoded synthetic sport settings.

1. **Make the fixture migration observable through its real consumer.** The fixture is read by `TestGetAthleteProfileDecodesFixture` (`internal/intervals/client_test.go`), but that test presently asserts only `PaceUnits`; it would still pass if `threshold_pace` remained `255.5`, zones remained `[360,330,300]`, or `pace_load_type` were absent. Include that test (or an equivalent fixture-decoding test) in the step and assert the decoded Run setting/type, `ThresholdPace == 3.5714285` (with an appropriate float tolerance), `PaceUnits == "MINS_KM"`, `PaceLoadType == "RUN"`, and unchanged percentage boundaries/names `[77.5,90,100]`. The JSON change must also replace the current Ride/VirtualRide types with a Run setting so the fixture itself is internally consistent.

2. **Audit and correct all remaining synthetic duration fixtures, not just `testdata/athlete_profile.json`.** `internal/tools/get_athlete_profile_test.go:373-374` still creates `PaceThreshold: 300/90` with duration-shaped zones `[360,330]` and `[100,90]`; `internal/tools/get_data_quality_report_test.go:153,217` similarly uses `PaceThreshold: 240` and `[300,270,240]`. These are misleading upstream-shaped records even if their current assertions only exercise readiness/data-quality behavior. The plan must replace them with semantically valid m/s `ThresholdPace` values, explicit applicable `PaceUnits`, and ascending percentage boundaries (for example `[77.5,90,100]`), and avoid retaining `PaceThreshold` as a duration compatibility fixture. Add the affected tool tests to the targeted verification command.

For the new table-driven tool regression, use a returned upstream `ThresholdPace` m/s value as well as asserting the outbound `SportSettingsPace.Value`, so each run/swim/rowing case proves both transport and response shaping rather than only the params-fallback echo. Use tolerance-based comparisons for the `280 s/km` reciprocal.
