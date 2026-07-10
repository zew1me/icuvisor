# Code Review: TP-235 Step 3 — Add boundary-focused regression coverage

**Verdict:** APPROVE

The renamed regression groups are selected by the required targeted command. Coverage now checks remaining versus full-week reconciliation totals, fixed commitments, slot non-combination, indoor/outdoor caps, unavailable/exhausted cases, and ordered batch requested-count/infeasibility outcomes.

**Verification:**
- `go test ./internal/planning -run 'Constraint|Reconciliation' -count=1`
- `go test ./internal/planning -count=1`
- `git diff --check acfcf07..HEAD`
