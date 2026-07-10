# Review R001 — Plan Review: Step 1

**Task:** TP-231 — Validate and canonicalize the yard distance suffix  
**Step:** Step 1: Update parser and canonical serializer  
**Verdict:** REVISE

---

## Summary

The plan identifies the right three source files and the right mechanical changes. The logic of the implementation is sound. However, there is one blocking structural problem with the step boundary: the plan's own test checkpoint (`go test ./internal/workoutdoc`) will fail at the end of Step 1 because existing tests that assert `yd` as canonical live in `workoutdoc_test.go`, which is deferred to Step 2. Two secondary issues need addressing before implementation begins.

---

## Required Changes to the Plan

### 1. Step boundary leaves the test suite broken (blocking)

The following tests in `workoutdoc_test.go` (a Step 2 artifact) hard-assert the old canonical:

```go
// TestWorkoutDocYardDistanceSerializeParseValidate
if got != "- Swim 100yd 95% Pace" {  // will fail once canonical is yrd

// TestWorkoutDocDistanceAliasesRemainCanonical
{name: "yards alias emits yd", unit: "yards", want: "- Stride 25yd 120%"},  // will fail
```

After Step 1 changes `syntax.go` to `Canonical: "yrd"`, serialization immediately emits `yrd` for all yard inputs. These two tests fail before Step 2 touches `workoutdoc_test.go`. The plan's checkpoint `go test ./internal/workoutdoc` cannot pass at the end of Step 1 as written.

**Fix:** Move the update of these two tests into Step 1. They are small, mechanical changes (two string literals in `workoutdoc_test.go`). Alternatively, if keeping the step boundary is important, the plan must explicitly note which tests will fail and instruct the implementer not to run the full package test until Step 2. The first option is strongly preferred — broken checkpoints mask real regressions.

### 2. `syntax.go` has inline DSL string literals that must also change in Step 1

Beyond the `workoutDistanceUnits` table entry, `WorkoutSyntaxSpec()` contains hardcoded `yd` strings in two places:

```go
// CheatSheet.Examples
{Label: "Yard swim step", DSL: "- Swim 100yd 95% Pace"},

// Features > distance_steps > SyntaxExamples
{Key: "distance_yd", Description: "Yard distance canonicalizes to yd.", ...}
```

And the `workoutDistanceUnits` description string itself:
```go
Description: "Yards serialize with the canonical `yd` suffix for pool-swim distances."
```

The `workout_syntax.md` golden (a Step 2 artifact) is generated from these struct literals. If Step 1 updates the table `Canonical` but leaves these strings as `yd`, the generated resource and the serializer will be inconsistent until Step 2. The plan should explicitly call out all four touch points in `syntax.go`:

- `workoutDistanceUnits[3].Canonical` → `"yrd"`
- `workoutDistanceUnits[3].Aliases` → add `"yrd"` (so round-tripping canonical output also parses)
- `workoutDistanceUnits[3].Description` → update to `yrd`
- `CheatSheet.Examples[3].DSL` → `"- Swim 100yrd 95% Pace"`
- `SyntaxExample` for `distance_yd` description → `"Yard distance canonicalizes to yrd."`

### 3. `distanceTokenRE` alternation order should be specified

The current regex is:

```go
distanceTokenRE = regexp.MustCompile(`^([0-9]+(?:\.[0-9]+)?)(mtr|km|mi|yd)$`)
```

The plan must add `yrd`, `yard`, and `yards`. Because the regex is anchored (`^...$`), correctness holds for any ordering, but the plan should note that `yrd` and `yards?` must precede `yd` in the alternation to make the intent readable. A concrete updated pattern to specify:

```
(mtr|km|mi|yrd|yards?|yd)
```

This is a minor but concrete gap — the plan leaves the exact regex change to inference.

---

## What the Plan Gets Right

- **`serialize.go` is correctly identified as a conditional change.** `formatDistance` iterates `workoutDistanceUnits.Aliases` and emits `syntax.Canonical`, so updating `syntax.go` is sufficient for serialization — no code change to `serialize.go` is needed.
- **The new `yard_suffix_test.go` is the right structural choice.** A dedicated regression file keeps the invariant visible and avoids growing the already-large `workoutdoc_test.go`.
- **The plan correctly identifies that `description_tokens.go` needs no code change.** `structuralTokenInDescription` delegates to `parseDistanceToken`, which reads `distanceTokenRE`. Updating the regex in `parse.go` is sufficient to make `100yrd` in a step description trigger the correct structural token error.
- **No golden fixtures in `testdata/` contain `yd`.** The golden fixture audit shows that no existing `.txt` fixture uses a yard distance, so Step 1 will not require touching `testdata/`.

---

## Suggestions for `yard_suffix_test.go` Content

The plan creates the file but does not describe what it must contain. At minimum:

1. All input aliases (`yd`, `yard`, `yards`, `yrd`) serialize to `yrd`.
2. `Parse("- Swim 100yd 95% Pace")` succeeds and `Distance.Unit` is `"yd"` (raw parse token); re-serializing produces `"100yrd"`.
3. `Parse("- Swim 100yrd 95% Pace")` succeeds and `Distance.Unit` is `"yrd"`; re-serializing produces `"100yrd"` (idempotent).
4. `"100yrd"` in a step description field triggers `*StructuralTokenInDescriptionError` via `ValidateDoc`.
5. `mtr`, `km`, and `mi` serialization is unchanged (regression guard, even if brief).

---

## Revised Step 1 Checklist

```
- [ ] Add yrd, yard, yards to distanceTokenRE (before yd in alternation)
- [ ] Update workoutDistanceUnits: Canonical → yrd, add yrd to Aliases, update Description
- [ ] Update CheatSheet.Examples yard DSL string → 100yrd
- [ ] Update SyntaxExample distance_yd description → "canonicalizes to yrd"
- [ ] Update TestWorkoutDocYardDistanceSerializeParseValidate expectation → 100yrd
- [ ] Update TestWorkoutDocDistanceAliasesRemainCanonical yards row → 100yrd
- [ ] Add yard_suffix_test.go with alias round-trips and description-token checks
- [ ] go test ./internal/workoutdoc — must pass cleanly
```
