# Review R012 — Code Review: Step 2

**Task:** TP-229 — Treat threshold pace as m/s and pace zones as percentages
**Step:** Step 2: Correct read shaping and typed models
**Verdict:** APPROVE

The upstream `pace_load_type` is decoded with an explicit typed field and preserved in the profile response without deriving it from sport or display units. The read-shaping contract continues to use `threshold_pace` as m/s, emits only the display-unit-selected seconds-per-distance field, retains percentage pace zones and names, and preserves the established unknown-unit m/s fallback.

The added coverage verifies JSON decoding, run/swim/rowing conversions, zone percentage serialization without legacy duration fields, `pace_load_type`, and unknown/`NONE`/overflow fallback behavior.

## Verification

- `go test -count=1 ./internal/athleteprofile ./internal/tools -run 'Profile|Pace'` — pass
- `go test -count=1 ./...` — pass
- `gofmt -l internal/intervals/types.go internal/athleteprofile/profile.go internal/athleteprofile/profile_test.go` — clean
- `git diff --check 4ac6442..HEAD` — pass
