# Code Review: TP-235 Step 1 — Define the constraint contract

**Verdict:** REVISE

## Blocking findings

### 1. Batch validation misclassifies a zero-capacity day

`ValidateCandidate` explicitly returns `day_unavailable` when a matching day has `MaxSessionsPerDay == 0` (`internal/planning/constraints.go:373-385`). `ValidateCandidates` only treats a missing `dayState` as unavailable (`:451-466`), so the same one-candidate input instead reaches `validateAgainstDay` and returns `daily_session_count_exceeded` (and can include unrelated slot/weekly findings). This contradicts the documented `day_unavailable` trigger (`docs/design/plan-filler-constraints.md:156`) and makes the two public APIs disagree.

Treat `ds.day.MaxSessionsPerDay == 0` as unavailable on the batch path before calling `validateAgainstDay`, matching the single-candidate path. Add a table test asserting both APIs return `day_unavailable` for a zero-capacity day.

### 2. `infeasible_load` documentation and exported contract still contradict the implementation

The implementation intentionally excludes invalid numeric candidates from `candLoad` before producing `WarnInfeasibleLoad` (`internal/planning/constraints.go:501-516`), and the batch-validation prose later correctly says this. But the exported `WarnInfeasibleLoad` comment (`:178-181`) and the result-code table (`docs/design/plan-filler-constraints.md:174`) still state that invalid candidates are included. Thus callers cannot reliably determine whether malformed candidates can suppress the warning—the exact contract discrepancy R006 was meant to resolve.

Update both stale statements to say valid-input candidates only, and add coverage with an invalid candidate whose nominal load would otherwise satisfy the remaining target.

## Verification

- `git diff --check 6a30598..HEAD` passes.
- `go test ./...` passes; `internal/planning` currently reports `[no test files]`.
- A focused temporary regression test reproduced finding 1: single validation returned `day_unavailable`, while batch validation returned `daily_session_count_exceeded`.
