# Code Review — TP-236 Step 4

## Verdict: REVISE

### Finding

1. **The pinned formula does not describe the implemented below-first-boundary bucket** (`internal/resources/analysis_formulas.go:68`, mirrored in `internal/resources/testdata/analysis_formulas.md`). The resource says every eligible interval is assigned to the configured lower-inclusive/upper-exclusive zone containing `power_i`. However, the approved contract and `ComputeZoneEnergy` synthesize a `Below <first zone>` bucket when the first configured boundary is greater than zero, and assign zero/sub-boundary power to it. In that case no configured zone contains `power_i`, so the public formula is incomplete and potentially misleading precisely around the locked coasting/boundary behavior. Amend the pinned paragraph to state that the final zone is open-ended and that `[0, first_boundary)` is emitted as an explicit below-zone bucket when needed; update the golden hash and assertions accordingly.

### Verification

The following otherwise passed:

- `make docs-tools` (generated web data remained clean)
- `go test ./internal/resources ./internal/toolcatalog ./internal/toolchecks -count=1`
- `go test ./internal/resources -run 'AnalysisFormulas' -count=1`
- `git diff --check 1c6876a..HEAD`
- Generated catalog count matches the PRD: 69 total, 30 core and 39 full.
