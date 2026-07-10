# Code Review: TP-235 Step 1 — Define the constraint contract

**Verdict:** REVISE

## Blocking findings

### 1. Exhausted weekly budgets still accept additional positive work

`validateAgainstDay` emits only `zero_remaining_load` / `zero_remaining_time` warnings when the remaining budget is zero or negative (`constraints.go:459-505`). Consequently, with a 100-load target and `CompletedLoad: 100`, a compatible candidate with `Load: 10` is returned as `Valid: true` (and the reconciliation projects 110). The same is true for a positive-duration candidate when the time budget is exhausted.

This contradicts the documented hard-code triggers (`weekly_load_overshoot` / `weekly_time_overshoot`, design doc:164-165) and, more importantly, fails the task's purpose of preventing the later writer from treating over-budget work as placeable. Retain the zero-budget warning if useful, but add the corresponding overshoot violation whenever the candidate adds positive load/time beyond a non-positive remaining budget. Add zero-remaining boundary tests for both dimensions.

### 2. The public float-based contract accepts invalid numeric inputs as valid sessions

The contract specifies `float64` inputs but defines neither a valid domain nor validation for them. All cap checks rely on `>` comparisons (`constraints.go:429-438`, `457-505`, `519-533`), so a candidate with negative duration/load or `math.NaN()` passes the slot, daily, and weekly checks. Negative candidates are also accumulated at `constraints.go:364-366`, allowing them to create artificial headroom for later candidates; NaN propagates into `Reconciliation`, which cannot be JSON-marshaled as a normal numeric result.

Define the input domain (at least finite, non-negative duration/load/totals/caps and non-negative session counts), add deterministic invalid-input result codes or a constraint-validation error path, and test negative and non-finite candidates/constraints. The current behavior turns malformed hard constraints into silently uncapped or budget-increasing values, contrary to the model's safety purpose.

### 3. Duplicate day dates yield contradictory single- and batch-validation results

`ValidateCandidate` uses `findDay`, which selects the first matching `AvailableDays` entry (`constraints.go:668-675`), while `ValidateCandidates` builds a map that silently keeps the last entry (`constraints.go:304-311`). For a duplicated date whose first entry has `MaxSessionsPerDay: 0` and second has one compatible slot, the single-candidate API returns `day_unavailable`, while a one-item batch returns valid. Duplicate dates also inflate `availableSlotCount` even though only one map entry is subsequently used.

Reject duplicate/out-of-week day dates during constraint validation, or establish and apply one canonical resolution policy in every public API. Add a test that `ValidateCandidate` and `ValidateCandidates` agree for a one-candidate input.

## Coverage and verification

- The committed change includes the full Step 2 validator, but `internal/planning` has no `_test.go` file (`go test ./internal/planning` passes with `[no test files]`). Add the focused tests for the public reconciliation and validation contract before treating the implementation as complete.
- `go vet ./internal/planning` passes and `gofmt -d internal/planning/constraints.go` is clean.
- `git diff --check 6a30598..HEAD` still fails on whitespace in the committed `R001-plan-step1.md` review artifact.
