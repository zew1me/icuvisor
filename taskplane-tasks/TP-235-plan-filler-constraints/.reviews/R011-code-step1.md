# Code Review: TP-235 Step 1 — Define the constraint contract

**Verdict:** REVISE

## Blocking findings

### 1. Requested session count is only an upper bound, so underfilled schedules are reported as successful

`RequestedSessionCount` is specified as how many sessions the caller *wants placed* (`docs/design/plan-filler-constraints.md:36,65`), but `ValidateCandidates` only rejects a candidate after `validCount` has reached that number (`internal/planning/constraints.go:502-515`). Its only final count warning compares the request with structural capacity (`:519-530`). Thus, with a request for three sessions, three compatible slots, and a batch containing one valid candidate, the batch has no violation or warning that two requested sessions are missing.

This lets a future filler treat an underfilled batch as satisfying the requested weekly session count, contrary to the availability-vs-request distinction and the task requirement to validate requested weekly session count. Define and return a deterministic underfill result (normally a batch warning such as `requested_session_count_unmet`, including requested and accepted counts), document whether candidates with other violations count, and add the three-request/one-valid-candidate regression test. Alternatively, explicitly redefine the field as a maximum rather than the requested placement count; that would not meet the task contract as written.

### 2. Lint fails

`make lint` fails on the changed validator:

```
internal/planning/constraints.go:802:6: QF1001: could apply De Morgan's law
internal/planning/constraints.go:805:6: QF1001: could apply De Morgan's law
```

Fix the two expressions (or otherwise satisfy the configured linter) before completion. Project guidance requires `make lint` to pass.

## Verification

- `go test ./...` passed.
- `make test-race` passed.
- `make build` passed.
- `git diff --check 6a30598..HEAD` passed.
- `make lint` failed as above.
