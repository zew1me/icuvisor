# Code Review — TP-229 Step 3

## Verdict: APPROVE

The write path now converts explicit pace durations to canonical m/s before transport, while sending a valid presentation `pace_units` value and preserving or deriving the `RUN`/`SWIM` load type. The write echo correctly converts returned or fallback m/s values back to the selected display-distance field and uses the unambiguous m/s fallback for `NONE` and unknown display units.

Pace-zone replacement now treats boundaries as finite, strictly increasing percentages in `(0, 200]`, forwards them unchanged, and retains the pre-client full delete-mode gate. The input schema and snapshot no longer describe pace-zone values as durations; tests cover wire payloads, display/load fallback, echo conversion, invalid percentage boundaries, safe-mode no-client behavior, and full-mode percentage round trips.

## Verification

- `git diff --check fc596db..HEAD` — pass
- `go test -count=1 ./internal/tools -run 'UpdateSportSettings|Pace|Zone'` — pass
- `go test -count=1 ./internal/tools ./internal/intervals ./cmd/gendocs` — pass
- `go test -count=1 ./...` — pass
- `make docs-tools` — no generated `web/data` drift
- `go run ./scripts/check_schema_stability.go -baseline-dir /tmp/schema-baseline-r015/internal/tools/schema_snapshot -require-baseline` (baseline `fc596db`) — pass
