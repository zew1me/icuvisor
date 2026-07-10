# Review R002 — Code Review: Step 1

**Task:** TP-231 — Validate and canonicalize the yard distance suffix  
**Step:** Step 1: Update parser and canonical serializer  
**Verdict:** REVISE

---

## Summary

The implementation correctly addresses every requirement raised in R001 and the step plan. The logic, test coverage, and multi-file coordination are all sound. One blocking issue remains: `yard_suffix_test.go` has a `gofmt` struct-field alignment violation that will fail CI.

---

## Blocking Issue

### 1. `gofmt` violation in `yard_suffix_test.go`

Running `gofmt -l internal/workoutdoc/yard_suffix_test.go` reports the file as needing formatting. The struct in `TestYardSuffixDSLRoundTrip` declares `wantReserial string` as the longest field, but the three preceding fields (`name`, `input`, `wantUnit`) are not padded to align:

```go
// current (fails gofmt)
for _, tc := range []struct {
    name        string
    input       string
    wantUnit    string
    wantReserial string   // <-- needs one more space on name/input/wantUnit
}{
```

`gofmt` wants to add a trailing space to `name`, `input`, and `wantUnit` so the type column aligns. The same misalignment propagates to all four struct literal assignments (`name:`, `input:`, `wantUnit:` all need an extra space to match `wantReserial:`). Per CLAUDE.md, CI fails on dirty `gofmt` diffs.

**Fix:** Run `gofmt -w internal/workoutdoc/yard_suffix_test.go` and commit the result.

---

## What Was Done Well

**All R001 issues addressed:**

- **(R001 §1)** `workoutdoc_test.go` updated in Step 1 alongside the serializer change, so the targeted test checkpoint (`go test ./internal/workoutdoc`) passes cleanly. Both affected assertions updated to `100yrd`.

- **(R001 §2)** All five `syntax.go` touch points updated:
  - `workoutDistanceUnits[3].Canonical` → `"yrd"` ✓
  - `workoutDistanceUnits[3].Aliases` → `["yrd", "yd", "yard", "yards"]` — canonical `yrd` is present so round-tripping already-canonical DSL parses back correctly ✓
  - `workoutDistanceUnits[3].Description` → `yrd` ✓
  - `CheatSheet.Examples[2].DSL` → `"- Swim 100yrd 95% Pace"` ✓
  - `SyntaxExample distance_yd` description → `"Yard distance canonicalizes to yrd."` ✓
  - The `distance_steps` feature prose was also updated (`"mtr, km, mi, or yrd"`) — good catch beyond R001's explicit list.

- **(R001 §3)** Regex alternation order `(mtr|km|mi|yrd|yards?|yd)` places `yrd` and `yards?` before `yd`. No ambiguity risk — the regex is fully anchored and no alternative overlaps with another for any valid input.

**`yard_suffix_test.go` covers all five scenarios recommended in R001:**
1. All four input aliases (`yrd`, `yd`, `yard`, `yards`) serialize to `yrd` — `TestYardSuffixCanonicalSerialization` ✓
2. Round-trip of legacy `100yd` DSL input produces `100yrd` — `TestYardSuffixDSLRoundTrip` ✓  
3. Canonical `100yrd` input is idempotent — `TestYardSuffixDSLRoundTrip` ✓
4. `100yrd` in a step description triggers `STRUCTURAL_TOKEN_IN_STEP_DESCRIPTION` — `TestYardSuffixDescriptionTokenError` ✓
5. `mtr`, `km`, `mi` canonical serialization unchanged — `TestYardSuffixOtherDistanceUnitsUnchanged` ✓
6. `TestYardSuffixValidateDescriptionLegacyYdInput` is a bonus: confirms `ValidateDescription` accepts legacy `yd` DSL and canonicalizes on re-serialization.

**No unintended scope creep:** `internal/response/units.go` contains a separate `"yd"` return for API response field suffixes (swim pool distance display in activity data). This is correctly left alone — it is not a workout DSL serialization path.

**No testdata fixture changes needed:** Confirmed no `testdata/` fixture files contain yard distance tokens; the step correctly does not touch them.

**Serialization mechanism is clean:** `formatDistance` in `serialize.go` required no code change. It already iterates `workoutDistanceUnits` aliases at lookup time and emits `syntax.Canonical`, so updating the table entry in `syntax.go` is sufficient for canonicalization. This is the right approach.

---

## Required Fix

```
gofmt -w internal/workoutdoc/yard_suffix_test.go
```

Verify with `gofmt -l internal/workoutdoc/yard_suffix_test.go` producing no output, then commit. No other changes needed for this step.
