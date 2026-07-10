# Review R006 — Code Review: Step 1

**Task:** TP-229 — Treat threshold pace as m/s and pace zones as percentages
**Step:** Step 1: Define canonical pace conversions
**Verdict:** REVISE

## Blocking finding

### 1. Workout pace previews still ignore `pace_units` for run and newly recognized display distances

`paceTargetPreview` passes `setting.PaceUnits` to `preferredPacePreviewUnit` (`internal/tools/workout_target_previews.go:202-207`), but that function only honors `SECS_100M`, `SECS_100Y`, and `SECS_500M` (`:230-237`). `MINS_KM`, `MINS_MILE`, `SECS_400M`, and `SECS_250M` instead fall through to the profile-wide metric/imperial preference (`:239-242`). This conflicts with the new response contract, where `pace_units` selects the display distance regardless of the broader preferred-units setting (`internal/athleteprofile/profile.go:311-370`; see the miles-profile / `MINS_KM` assertion in `internal/tools/get_athlete_profile_test.go:590-608`).

For example, a metric athlete whose sport setting has `pace_units: "MINS_MILE"` and `threshold_pace: 3.5714285` receives a `4:40/km` workout basis/preview, even though that setting declares mile pace and the profile response correctly exposes approximately `7:31/mi`. Likewise, `SECS_400M` and `SECS_250M`, which this change adds to the recognized display enums, are rendered as `/km` or `/mi` rather than `/400m` or `/250m`.

Make preview-distance selection cover every recognized `pace_units` value (preferably through the shared pace-distance conversion), using the global unit system only as the fallback for `NONE` or unknown units. Add regression coverage where profile-wide units conflict with `MINS_KM`/`MINS_MILE`, plus 400 m and 250 m displays.

## Verification

- `go test -count=1 ./internal/units ./internal/response ./internal/athleteprofile ./internal/analysis ./internal/tools` — pass
- `go test -race -count=1 ./internal/units ./internal/response ./internal/athleteprofile ./internal/analysis ./internal/tools` — pass
- `make test` — pass
- `make lint` — pass
- `make build` — pass
- `make docs-tools` — pass
- `git diff --check a077f6616fa8cad40619afa11a616c7b010d56f8..HEAD` — pass
