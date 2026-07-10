# Code Review: TP-235 Step 1 — Define the constraint contract

**Verdict:** REVISE

## Blocking findings

### 1. An invalid numeric candidate poisons the rest of a batch and its response

`ValidateCandidates` recognizes an invalid `NaN`/infinite/negative candidate at lines 404–406, but unconditionally adds that same untrusted value to `priorLoad`/`priorMinutes` (lines 453–455) and reconciliation totals (lines 473–476). A `NaN` first candidate therefore makes all subsequent budget comparisons false and makes `Reconciliation` contain `NaN`; a negative invalid candidate can manufacture headroom for later work. The echoed `Candidate` and `Violation.Value` also retain `NaN`, so JSON-marshaling the result fails.

This contradicts the design document's promise that invalid values do not propagate into budget accumulations or reconciliation (line 289), and prevents the result from being a future MCP response. Do not accumulate malformed candidates or expose non-finite values in serializable output (or fail the batch before producing a result), and add regression coverage for an invalid candidate followed by an otherwise-over-budget candidate.

### 2. The zero-budget documentation still describes a warning-only outcome, while the code blocks the candidate

The implementation correctly adds `weekly_load_overshoot`/`weekly_time_overshoot` for positive work against an exhausted budget. However, the key invariant says the validator emits a warning and “lets the caller decide” (docs line 46), validation steps 5–6 describe only warnings (lines 187–188), and the zero-load example shows only `zero_remaining_load` (lines 254–262). All contradict the boundary-condition table and current code.

Align the invariant, validation logic, and example with the hard-block semantics (including the zero-time equivalent). This is a public planning contract and callers must not infer that a warning-only result is placeable.

### 3. `ValidateWeekConstraints` is not deterministic when multiple top-level numeric fields are malformed

Lines 268–280 put the six fields in a map and return the first invalid entry encountered. Go deliberately randomizes map iteration, so the returned error (and field reported to the caller) changes between otherwise identical calls containing multiple invalid values. That violates the stated deterministic-validator contract.

Validate a fixed ordered slice (or explicit fields) and add a multiple-invalid-fields test that asserts the stable first error.

## Verification

- `go test ./internal/planning` and `go vet ./internal/planning` pass, but the package has no committed tests.
- A temporary regression test with a `NaN` first candidate and a 200-load second candidate against a 100-load target failed: the second candidate was returned `Valid: true`, and reconciliation contained `CandidateLoad: NaN`.
- `git diff --check 6a30598..HEAD` still fails due to trailing whitespace and a final blank line in committed `R001-plan-step1.md`; clean that artifact before the required formatting check.
