# Plan Review: TP-235 Step 3 — Add boundary-focused regression coverage

**Verdict:** REVISE

## Required revisions

1. **Make the planned verification command run the planned regressions.** The stated command, `go test ./internal/planning -run 'Constraint|Reconciliation'`, currently selects only the `TestValidateWeekConstraints_*` tests. It does **not** select the relevant `TestValidateCandidate_*`, `TestValidateCandidates_*`, or `TestReconcile_*` boundary tests. Rename/group the Step 3 tests to match that expression, or replace the command with an expression that explicitly selects all boundary test groups (and verify with `go test -list`). Otherwise Step 3 can report a passing targeted run without executing its required coverage.

2. **Plan assertions for reconciliation values, not only pass/fail codes.** The in-progress-week and fixed-event cases must assert `RemainingLoad`/`RemainingMinutes` and `ProjectedLoad`/`ProjectedMinutes` as applicable, alongside the overshoot result. This directly protects the core contract: candidates are compared with `target - completed - fixed`, never the full-week target, and reconciliation reports the same totals.

3. **State which API each regression exercises.** Include batch (`ValidateCandidates`) coverage for the slot and requested-count/infeasibility cases, including the expected per-candidate result order and batch warnings. Single-candidate tests alone cannot protect batch slot consumption/matching, count caps, or reconciliation of a candidate schedule.

## Notes

- The existing `constraints_test.go` already contains tests resembling all four listed Step 3 cases, and STATUS R009 says they were written, while Step 3 remains unticked. The implementation plan should begin with an audit of those tests and add only missing assertions/coverage rather than duplicating them; then update the Step 3 checkboxes to accurately reflect the audited result.
- Use a table-driven matrix for the related fixed/exhausted/unavailable/infeasible boundaries, per project test conventions, while keeping separate tests where a clearer batch setup is warranted.
