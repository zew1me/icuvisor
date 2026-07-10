# Code Review: TP-235 Step 1 — Define the constraint contract

**Verdict:** REVISE

**Reviewed:** 2026-07-10
**Step:** Step 1: Define the constraint contract
**Artifacts reviewed:** `internal/planning/constraints.go` (new), `docs/design/plan-filler-constraints.md` (new)

---

## Summary

The type definitions and design document are structurally sound and correctly implement the separation of concerns required by the PROMPT. All R001 structural requirements are satisfied: the availability-vs-requested-count distinction is clean, `Reconciliation` carries parallel load and duration fields, and `ViolationCode`/`WarningCode` are separate typed enums. The package compiles cleanly, `gofmt` is clean, and `go vet` passes.

However, the commit overshoots Step 1's scope: the full validation logic (all five public functions plus four private helpers) is present in `constraints.go` even though Steps 2 and 3 are marked "Not Started". Reviewing that logic reveals three defects that need to be fixed before proceeding:

1. A code/spec inconsistency on when `WarnZeroRemainingLoad` fires.
2. Asymmetric zero-budget handling: load gets a warning when the budget is exhausted; time silently gets neither a warning nor a violation.
3. The R001 required action to document the duration-unit rationale was not completed.

---

## Issues Requiring Changes Before Proceeding

### 1. Internal inconsistency in `WarnZeroRemainingLoad` trigger condition

The design doc contains two contradictory specifications:

- **Validation Logic section, step 5:** "If remaining ≤ 0, emit `zero_remaining_load` warning."
- **Boundary Conditions table:** "`FixedLoad + CompletedLoad > WeeklyTargetLoad` | RemainingLoad is negative; `zero_remaining_load` warning fires for any candidate with Load > 0."

The code implements the Validation Logic version (fires unconditionally when `remainingLoad <= 0`). The Boundary Conditions row adds a `Load > 0` filter that the code does not apply. These cannot both be correct. A zero-load session (e.g., mobility work with `Load: 0`) submitted against an exhausted budget would trigger the warning under the code but not under the boundary conditions table.

**Required action:** decide the intended trigger — either remove "with Load > 0" from the boundary conditions table and document why zero-load sessions still produce the warning, or add `candidate.Load > 0` to the code guard and align the Validation Logic prose accordingly. Both choices are defensible; the inconsistency is not.

### 2. Asymmetric zero-budget handling: time has no warning when budget is exhausted

For the load dimension, when `remainingLoad <= 0` the code emits `WarnZeroRemainingLoad`. For the time dimension, the check is:

```go
if wc.WeeklyTargetMinutes > 0 && remainingMin > 0 && candidate.DurationMinutes > remainingMin {
    // ViolationWeeklyTimeOvershoot
}
```

When `remainingMin <= 0` (the week is already over the time target), the condition is false and no feedback is produced — no warning, no violation. A caller who submits a 60-minute session against a week where `CompletedMinutes` already exceeds `WeeklyTargetMinutes` gets silence. The Boundary Conditions table does not document this case at all.

This is a real signal loss. A batch-validation caller relying on the time dimension for budget-awareness will silently get `Valid: true` on sessions that push the projected total further over target.

**Required action:** choose one of:
- (a) Emit `WarnZeroRemainingLoad` for the time case in parallel with the load case (add a separate code or extend the existing one to cover both dimensions).
- (b) Emit `ViolationWeeklyTimeOvershoot` even when `remainingMin <= 0` (treating any addition to an already-exceeded budget as a violation), matching the stricter load-overshoot semantics.
- (c) Document the intentional asymmetry explicitly in the Boundary Conditions table with the rationale (e.g., "time target is advisory; load target is hard").

Do not leave this as an undocumented gap.

### 3. R001 required action not completed: duration-unit rationale absent from design doc

R001 explicitly required: "pick a concrete representation, **document it in the design doc under 'Field semantics'**, and keep it consistent across all structs." The recommendation was `time.Duration`; the implementation chose `float64` minutes. That choice is reasonable (simpler JSON serialization, no import coupling) but the design doc has no "Field semantics" section and no rationale for the `float64`/minutes decision. The type column in the data-model tables states `float64` but that is notation, not rationale.

**Required action:** add a "Field semantics" or "Units" subsection to the design doc that names the unit convention (`float64` minutes for all duration fields), explains why `time.Duration` was not used, and confirms the convention is uniform across `SlotConstraint`, `DayConstraints`, `WeekConstraints`, and `CandidateSession`.

---

## Non-blocking Observations (address in Step 2 or document as intentional)

### A. Invalid candidates accumulate weekly budget in batch validation

In `ValidateCandidates`, `priorLoad` and `priorMinutes` are incremented for every candidate regardless of validity:

```go
// after validateWithState ...
priorLoad += candidate.Load
priorMinutes += candidate.DurationMinutes
```

The comment documents this only for the day counter ("valid or not"). If the third candidate in a batch is on an unavailable day (invalid, `Load: 80`), that 80 points still reduces the remaining budget for the fourth candidate. Whether this is the intended semantic is unclear; an "optimistic" accumulator (only valid candidates consume budget) and a "pessimistic" one (all do) are both defensible.

**What to do:** make the behaviour explicit in the design doc's "Batch validation" paragraph — "accumulated load from all candidates including invalid ones" or "accumulated load from valid candidates only" — and update the code to match.

### B. `WarnInfeasibleLoad` uses sum of all candidates including invalid ones

In `ValidateCandidates`:

```go
var candMin, candLoad float64
for _, c := range candidates {
    candMin += c.DurationMinutes
    candLoad += c.Load
}
recon := buildReconciliation(wc, candMin, candLoad)

if recon.RemainingLoad > 0 && candLoad < recon.RemainingLoad {
    // WarnInfeasibleLoad
}
```

`candLoad` includes invalid candidates. If the batch contains 3 invalid candidates each with `Load: 50` (total 150) and one valid candidate with `Load: 20`, and `RemainingLoad: 200`, the warning is suppressed because `170 < 200` is false. But the 150 units from invalid candidates will never actually be placed. The warning fires correctly only when the caller provides the right candidates, not when the batch itself is misconfigured.

This follows naturally from observation A above — both stem from the same unconditional accumulator. Resolve them together.

### C. `availableSlotCount` does not account for sport/mode filtering

`availableSlotCount` counts structural slot capacity without evaluating `AllowedSports` or `AllowedModes`. A day with two slots where one allows "Ride" only and the other "Run" only counts as 2 slots for a "Swim" candidate, potentially suppressing `WarnInfeasibleSessionCount`. For Step 1 scope this is acceptable, but the design doc's "Validation Logic → Batch validation" description should note that the slot count is structural capacity, not candidate-specific capacity.

### D. STATUS.md is inaccurate

`constraints.go` contains the full implementation of Steps 1 and 2. STATUS.md marks Step 2 as "Not Started". Update STATUS.md to reflect reality before the next step runs, so the review chain is traceable.

---

## What is Correct and Must Not Change

- Struct decomposition and field naming are clean. The `WeekConstraints` / `DayConstraints` / `SlotConstraint` / `CandidateSession` hierarchy is correct.
- `ViolationCode` and `WarningCode` as separate typed constants with separate `Violation`/`Warning` structs satisfies the R001 severity-separation requirement.
- Parallel `...Minutes` and `...Load` fields throughout `Reconciliation` correctly addresses R001 issue 2.
- Early return on `ViolationDayUnavailable` (short-circuit before other checks) is the right behaviour.
- `checkSlotConstraints` returning `nil` on any-fit and deduplicating violation codes across slots is correct.
- `availableSlotCount` capping at `min(MaxSessionsPerDay, len(Slots))` is correct.
- JSON tags only on output types (`Violation`, `Warning`, `Reconciliation`, `CandidateResult`, `BatchResult`) and absent from input structs is the right boundary.
- `slices.Contains` from stdlib is appropriate for Go 1.25.

---

## Required Changes Summary

| # | Severity | Location | Description |
|---|----------|----------|-------------|
| 1 | **Blocking** | `constraints.go` L367 and design doc boundary table | Resolve `WarnZeroRemainingLoad` trigger: code fires unconditionally when remaining ≤ 0, but boundary conditions table says "Load > 0". Pick one and align both. |
| 2 | **Blocking** | `constraints.go` L384 and design doc boundary table | Add explicit handling when `remainingMin <= 0` (warning or violation); document the chosen semantics in the boundary conditions table. |
| 3 | **Blocking** | `docs/design/plan-filler-constraints.md` | Add "Field semantics" / "Units" section documenting the `float64` minutes choice and why `time.Duration` was not used (R001 required action). |
| A | Non-blocking | `constraints.go` + design doc | Document whether invalid candidates' load/minutes count in the weekly budget accumulator for subsequent candidates. |
| B | Non-blocking | `constraints.go` + design doc | Document that `WarnInfeasibleLoad` includes invalid candidates in `candLoad`. |
| C | Non-blocking | Design doc | Note that `availableSlotCount` is structural capacity, not candidate-type-filtered. |
| D | Non-blocking | `STATUS.md` | Mark Step 2 as completed. |
