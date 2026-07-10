# Code Review — TP-229 Step 4

## Verdict: REVISE

1. **Lock all fixture zone values and names.** `internal/intervals/client_test.go:353` checks only the first and last pace-zone values and only the final name. It would accept a corrupted middle `90` boundary or `Moderate` name, despite Step 4 requiring the fixture's complete `[77.5, 90, 100]` percentage boundaries and names to be locked. Assert the complete slices (or every element).

2. **Make the returned-upstream echo distinguishable from the request.** `internal/tools/sport_settings_pace_semantics_test.go:26-30` gives every `returnedMPS` the same value as the outbound conversion. A regression that ignores the upstream response and echoes `params.ThresholdPace` would therefore pass all five rows. Use a deliberately different finite returned m/s value (and its expected selected-display seconds) for each row, while retaining a separate exact `3.5714285 m/s` / `280 s/km` reciprocal assertion.

Verification run: `go test -race ./internal/response ./internal/athleteprofile ./internal/intervals ./internal/tools` passed.
