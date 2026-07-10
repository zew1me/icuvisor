# Task: TP-235 - Add plan-filler constraint model and validator

**Created:** 2026-07-10
**Size:** L

## Review Level: 2 (Plan and Code)

**Assessment:** This introduces a new planning-domain contract that future calendar-filling tools will rely on to prevent over-scheduling. It is isolated from live writes but establishes novel invariants across availability, duration, session count, and weekly load.
**Score:** 4/8 — Blast radius: 1, Pattern novelty: 2, Security: 0, Reversibility: 1

## Canonical Task Folder

```
taskplane-tasks/TP-235-plan-filler-constraints/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Create a deterministic, pure planning constraint model and candidate-schedule validator for the roadmap v2.0 Plan Filler. Model requested sessions separately from available days, athlete-local daily slots, per-day session counts, per-session and per-mode duration caps, fixed existing commitments, and full-week versus remaining-week load accounting. The validator must report explicit violations and reconciliation totals without selecting workouts or writing calendar events, so the later `fill_calendar_from_library` implementation cannot silently exceed time/load limits or combine separate slots into one oversized session.

## Dependencies

- **None**

## Context to Read First

**Tier 2 (area context):**

- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**

- `docs/prd/PRD-icuvisor.md` — planning, safety, timezone, and response-transparency rules
- `ROADMAP.md` — v2.0 Plan Filler and v2.2 plan-validation scope
- `internal/tools/propose_annual_training_plan.go` — existing weekly load/hour terminology and boundaries
- `internal/tools/apply_training_plan.go` — existing protected-event and calendar conflict behavior

## Environment

- **Workspace:** repository root
- **Services required:** None

## File Scope

- `internal/planning/constraints.go`
- `internal/planning/constraints_test.go`
- `docs/design/plan-filler-constraints.md`
- `docs/prd/PRD-icuvisor.md`
- `ROADMAP.md`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Existing annual-plan and training-plan load/time semantics reviewed

### Step 1: Define the constraint contract

**Plan-review checkpoint**

- [ ] Define athlete-local weekly targets, completed/fixed load, requested session count, daily availability slots, and per-slot limits without using free text for hard constraints
- [ ] Define candidate-session inputs and deterministic violation/warning/result codes suitable for a future MCP response
- [ ] Specify full-week versus in-progress remaining-load reconciliation and protected fixed-event accounting
- [ ] Document that availability means where sessions may be placed, while requested sessions means how many should be placed
- [ ] Run targeted tests: `go test ./internal/planning`

**Artifacts:**

- `internal/planning/constraints.go` (new)
- `docs/design/plan-filler-constraints.md` (new)

### Step 2: Implement validation and reconciliation

- [ ] Validate daily session count, individual duration, combined daily duration, indoor/outdoor caps, allowed sport/mode, and requested weekly session count
- [ ] Compute completed, fixed, candidate, remaining, and projected weekly time/load totals without silently redistributing deficits
- [ ] Return deterministic violations when a candidate exceeds a cap and warnings when requested load is infeasible within available slots
- [ ] Keep the package pure: no intervals.icu client calls, calendar writes, model inference, or physiology classification
- [ ] Run targeted tests: `go test ./internal/planning`

**Artifacts:**

- `internal/planning/constraints.go` (new)
- `internal/planning/constraints_test.go` (new)

### Step 3: Add boundary-focused regression coverage

- [ ] Cover an in-progress week where only the remaining target may be scheduled, preventing full-target overshoot
- [ ] Cover two separate 45-minute slots that cannot become one 95-minute session
- [ ] Cover a 60-minute indoor cap that does not constrain a longer allowed outdoor slot
- [ ] Cover fixed events, zero remaining load, unavailable days, and infeasible requested-session counts
- [ ] Run targeted tests: `go test ./internal/planning -run 'Constraint|Reconciliation'`

**Artifacts:**

- `internal/planning/constraints_test.go` (new)

### Step 4: Align roadmap and product contract

- [ ] Add the structured constraint and reconciliation acceptance criteria to v2.0 without claiming Plan Filler is already shipped
- [ ] Clarify the boundary between constraint validation and v2.2 evidence-based coaching guardrails
- [ ] Update the PRD only where needed to preserve authoritative product behavior and add an Unreleased changelog note only if a public contract becomes user-visible
- [ ] Run formatting/document checks: `gofmt -w internal/planning/*.go && git diff --check`

**Artifacts:**

- `docs/design/plan-filler-constraints.md` (new)
- `docs/prd/PRD-icuvisor.md` (modified if needed)
- `ROADMAP.md` (modified)
- `CHANGELOG.md` (modified only if user-visible)

### Step 5: Testing & Verification

**Code review checkpoint**

- [ ] Run FULL test suite: `make test`
- [ ] Run race suite: `make test-race`
- [ ] Run lint: `make lint`
- [ ] Fix all failures
- [ ] Build passes: `make build`
- [ ] Verify clean formatting and docs diff: `git diff --check`

### Step 6: Documentation & Delivery

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] Discoveries logged in STATUS.md

## Documentation Requirements

**Must Update:**

- `docs/design/plan-filler-constraints.md` — record the field semantics, invariants, result codes, and examples
- `ROADMAP.md` — add the constraint-validator acceptance criteria under v2.0

**Check If Affected:**

- `docs/prd/PRD-icuvisor.md` — update only if the product contract needs the new hard-constraint semantics
- `CHANGELOG.md` — update only if the work exposes a user-visible capability

## Completion Criteria

- [ ] A new pure planning package validates candidate schedules against structured constraints
- [ ] Availability and requested-session count are represented separately
- [ ] Remaining-week load cannot be mistaken for the original full-week target
- [ ] Daily/session/indoor caps produce deterministic violations
- [ ] New boundary-focused test file exists and passes
- [ ] Roadmap and design documentation do not claim the Plan Filler is shipped
- [ ] Full tests, race, lint, build, and formatting checks pass

## Git Commit Convention

Commits happen at step boundaries. All commits MUST include TP-235:

- **Step completion:** `feat(TP-235): complete Step N — description`
- **Bug fixes:** `fix(TP-235): description`
- **Tests:** `test(TP-235): description`
- **Hydration:** `hydrate: TP-235 expand Step N checkboxes`

## Do NOT

- Implement workout selection or calendar writes
- Treat free-text instructions as enforceable hard constraints
- Infer indoor/outdoor, sport, age, or recovery limits that the caller did not supply
- Redistribute excess or missing load silently
- Add a model-controlled confirmation override
- Change current annual-plan or apply-training-plan behavior unless required by an explicitly reviewed contract update
- Commit without TP-235 in the message

---

## Amendments (Added During Execution)

<!-- Workers add amendments here if prerequisites or instructions are contradictory. -->
