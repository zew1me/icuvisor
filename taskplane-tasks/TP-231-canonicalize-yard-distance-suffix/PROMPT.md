# Task: TP-231 - Validate and canonicalize the yard distance suffix

**Created:** 2026-07-10
**Size:** M

## Review Level: 2 (Plan and Code)

**Assessment:** This changes the canonical serialized workout DSL across parser, serializer, MCP resource, fixtures, and user documentation. Existing `yd` input should remain reversible, but uploaded workouts may depend on exact token behavior.
**Score:** 4/8 — Blast radius: 1, Pattern novelty: 1, Security: 0, Reversibility: 2

## Canonical Task Folder

```
taskplane-tasks/TP-231-canonicalize-yard-distance-suffix/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Align yard-based pool-swim distance serialization with the publicly documented intervals.icu workout DSL. Public syntax evidence uses `yrd`, `yards`, or `y`, while icuvisor currently emits `yd`. Validate that contract from public upstream material, accept legacy/user-friendly aliases including `yd`, and serialize one canonical `yrd` suffix. Update parser, serializer, resource text, golden fixtures, PRD, and website examples together so no surface continues teaching the old token.

## Dependencies

- **None**

## Context to Read First

**Tier 2 (area context):**

- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**

- `docs/prd/PRD-icuvisor.md` — WorkoutDoc round-trip and canonical DSL rules
- `https://forum.intervals.icu/t/workout-builder/1163?page=53` — public workout-builder unit syntax
- `https://forum.intervals.icu/t/distanced-based-workouts-supported/9973` — public distance-unit behavior

## Environment

- **Workspace:** repository root
- **Services required:** None; do not upload workouts to a real account

## File Scope

- `internal/workoutdoc/parse.go`
- `internal/workoutdoc/serialize.go`
- `internal/workoutdoc/syntax.go`
- `internal/workoutdoc/workoutdoc_test.go`
- `internal/workoutdoc/validate_test.go`
- `internal/workoutdoc/yard_suffix_test.go`
- `internal/resources/testdata/workout_syntax.md`
- `internal/resources/workout_syntax_test.go`
- `internal/tools/validate_workout_test.go`
- `web/content/cookbook/build-workouts.md`
- `docs/prd/PRD-icuvisor.md`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm public upstream syntax supports `yrd` and does not document `yd` as canonical

### Step 1: Update parser and canonical serializer

**Plan-review checkpoint**

- [ ] Accept `yrd` as a distance token and preserve `yd`, `yard`, and `yards` as backward-compatible input aliases
- [ ] Serialize all yard distances as canonical `yrd`
- [ ] Preserve `mtr`, `km`, and `mi` behavior exactly
- [ ] Ensure structured description validation does not confuse yard tokens with duration text
- [ ] Run targeted tests: `go test ./internal/workoutdoc`

**Artifacts:**

- `internal/workoutdoc/parse.go` (modified)
- `internal/workoutdoc/serialize.go` (modified if canonicalization lives there)
- `internal/workoutdoc/syntax.go` (modified)
- `internal/workoutdoc/yard_suffix_test.go` (new)

### Step 2: Update resources, examples, and round-trip fixtures

- [ ] Replace canonical `yd` output expectations with `yrd` across workout syntax resources and golden fixtures
- [ ] Add round-trip tests proving legacy `100yd` input becomes canonical `100yrd` without changing distance meaning
- [ ] Add validation coverage for `yrd`, `yards`, and malformed yard tokens
- [ ] Update website examples and PRD canonical suffix text
- [ ] Run targeted tests: `go test ./internal/workoutdoc ./internal/resources ./internal/tools -run 'Workout|Yard|Syntax'`

**Artifacts:**

- `internal/workoutdoc/workoutdoc_test.go` (modified)
- `internal/workoutdoc/validate_test.go` (modified)
- `internal/resources/testdata/workout_syntax.md` (modified)
- `internal/resources/workout_syntax_test.go` (modified)
- `internal/tools/validate_workout_test.go` (modified)
- `web/content/cookbook/build-workouts.md` (modified)
- `docs/prd/PRD-icuvisor.md` (modified)

### Step 3: Testing & Verification

**Code review checkpoint**

- [ ] Run FULL test suite: `make test`
- [ ] Run race suite: `make test-race`
- [ ] Run lint: `make lint`
- [ ] Fix all failures
- [ ] Build passes: `make build`
- [ ] Verify generated resource/golden text and clean diff: `git diff --check`

### Step 4: Documentation & Delivery

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] Discoveries logged in STATUS.md

## Documentation Requirements

**Must Update:**

- `docs/prd/PRD-icuvisor.md` — change canonical yard suffix from `yd` to `yrd`
- `web/content/cookbook/build-workouts.md` — teach `yrd` while noting accepted structured aliases
- `CHANGELOG.md` — record canonical yard serialization correction

**Check If Affected:**

- `internal/resources/testdata/workout_syntax.md` — generated resource must match code
- Other website workout recipes and schema examples found by structural/search audit

## Completion Criteria

- [ ] Parser accepts canonical `yrd` and backward-compatible `yd` aliases
- [ ] Serializer emits only `yrd` for yard distance
- [ ] Meter/minute disambiguation remains unchanged
- [ ] New yard-specific regression test file exists
- [ ] Resources, PRD, website examples, and fixtures agree
- [ ] Full tests, race, lint, and build pass

## Git Commit Convention

Commits happen at step boundaries. All commits MUST include TP-231:

- **Step completion:** `fix(TP-231): complete Step N — description`
- **Bug fixes:** `fix(TP-231): description`
- **Tests:** `test(TP-231): description`
- **Hydration:** `hydrate: TP-231 expand Step N checkboxes`

## Do NOT

- Remove support for existing structured inputs using `yd`
- Change `mtr`, `km`, or `mi` canonical behavior
- Treat bare `m` as meters
- Upload a workout or use athlete credentials
- Copy GPL implementation code
- Expand into unrelated WorkoutDoc grammar changes
- Commit without TP-231 in the message

---

## Amendments (Added During Execution)

<!-- Workers add amendments here if prerequisites or instructions are contradictory. -->
