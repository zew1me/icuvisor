# Task: TP-230 - Distinguish ATP-generated notes from personal calendar notes

**Created:** 2026-07-10
**Size:** M

## Review Level: 2 (Plan and Code)

**Assessment:** The change is localized to annual-plan shaping but alters model-visible planning context and recovery conclusions. Incorrect classification can make an assistant confidently misstate a training plan.
**Score:** 4/8 — Blast radius: 2, Pattern novelty: 1, Security: 0, Reversibility: 1

## Canonical Task Folder

```
taskplane-tasks/TP-230-distinguish-atp-notes/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Use upstream provenance to distinguish Annual Training Plan notes from ordinary personal calendar notes. Live calendar evidence shows ATP-generated PLAN, TARGET, and NOTE events share a non-null `plan_applied` timestamp, while personal notes have `plan_applied: null`; `for_week` is false for both and note text may be localized. Keep useful personal notes as neutral context if the response contract benefits from them, but never count or label them as ATP recovery/week notes. Replace English keyword dependence with provenance-visible, locale-independent shaping.

## Dependencies

- **None**

## Context to Read First

**Tier 2 (area context):**

- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**

- `docs/prd/PRD-icuvisor.md` — annual training plan response contract
- `ROADMAP.md` — planning scope
- `docs/upstream-gaps/periodization-parameters.md` — boundary between explicit upstream data and inference
- `https://github.com/hhopke/intervals-icu-mcp/issues/79` — public behavior report and live `plan_applied` evidence; use behavior only

## Environment

- **Workspace:** repository root
- **Services required:** None

## File Scope

- `internal/intervals/events.go`
- `internal/tools/get_annual_training_plan.go`
- `internal/tools/get_annual_training_plan_test.go`
- `internal/tools/get_annual_training_plan_provenance_test.go`
- `internal/tools/schema_snapshot/get_annual_training_plan.json`
- `web/data/tools.json`
- `web/data/tool_schemas.json`
- `docs/prd/PRD-icuvisor.md`
- `docs/upstream-gaps/periodization-parameters.md`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm `plan_applied` is already decoded and identify every current keyword-based recovery path

### Step 1: Design provenance-aware note shaping

**Plan-review checkpoint**

- [ ] Define separate model-visible status for ATP-generated week notes and personal/context notes
- [ ] Use non-empty `plan_applied` as the ATP provenance signal; do not use `for_week` or English names
- [ ] Ensure summaries and week counts cannot classify a personal note as a recovery or ATP note
- [ ] Define deterministic ordering and terse/full behavior for both note classes
- [ ] Run targeted tests: `go test ./internal/tools -run 'AnnualTrainingPlan'`

**Artifacts:**

- `internal/tools/get_annual_training_plan.go` (modified)

### Step 2: Implement and cover classification

- [ ] Add provenance fields or neutral response structures without dropping useful context silently
- [ ] Remove or constrain English keyword recovery classification so localized plan notes and personal Rest/Travel notes behave correctly
- [ ] Add fixtures for a personal `Travel — Rest` note with null plan_applied and localized ATP notes with a shared plan_applied timestamp
- [ ] Verify real one-day TARGET spans still produce Monday-through-Sunday week boundaries
- [ ] Preserve projection bridge behavior based only on explicit TARGET load rows
- [ ] Run targeted tests: `go test ./internal/tools -run 'AnnualTrainingPlan'`

**Artifacts:**

- `internal/tools/get_annual_training_plan.go` (modified)
- `internal/tools/get_annual_training_plan_provenance_test.go` (new)
- `internal/tools/get_annual_training_plan_test.go` (modified)

### Step 3: Update schema and generated surfaces

- [ ] Update output schema snapshot for the provenance-aware note shape
- [ ] Regenerate website tool data
- [ ] Ensure the tool description tells models that personal context notes are not ATP instructions
- [ ] Run targeted tests: `go test ./internal/tools ./internal/toolchecks`

**Artifacts:**

- `internal/tools/schema_snapshot/get_annual_training_plan.json` (modified)
- `web/data/tools.json` (modified)
- `web/data/tool_schemas.json` (modified)

### Step 4: Testing & Verification

**Code review checkpoint**

- [ ] Run FULL test suite: `make test`
- [ ] Run race suite: `make test-race`
- [ ] Run lint: `make lint`
- [ ] Fix all failures
- [ ] Build passes: `make build`
- [ ] Regenerate docs and verify clean diff: `make docs-tools && git diff --check`

### Step 5: Documentation & Delivery

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] Discoveries logged in STATUS.md

## Documentation Requirements

**Must Update:**

- `docs/prd/PRD-icuvisor.md` — clarify ATP-generated week-note provenance versus personal context notes
- `CHANGELOG.md` — record the corrected note classification under Unreleased

**Check If Affected:**

- `docs/upstream-gaps/periodization-parameters.md` — record `plan_applied` as event provenance, not an athlete-level planning parameter
- Generated tool reference/schema data

## Completion Criteria

- [ ] Personal notes never increment ATP/recovery week-note counts
- [ ] Localized ATP notes are recognized by provenance without English keywords
- [ ] Response exposes enough provenance for the LLM to explain the distinction
- [ ] Projection bridge remains deterministic and unchanged for explicit targets
- [ ] New provenance regression test file exists
- [ ] Full tests, race, lint, build, and generated docs pass

## Git Commit Convention

Commits happen at step boundaries. All commits MUST include TP-230:

- **Step completion:** `fix(TP-230): complete Step N — description`
- **Bug fixes:** `fix(TP-230): description`
- **Tests:** `test(TP-230): description`
- **Hydration:** `hydrate: TP-230 expand Step N checkboxes`

## Do NOT

- Filter ATP notes using `for_week`
- Infer provenance from localized names or English keywords
- Drop personal notes without documenting whether they remain as neutral context
- Invent recovery cadence or other athlete-level parameters
- Copy GPL implementation code
- Change annual-plan proposal/apply behavior
- Commit without TP-230 in the message

---

## Amendments (Added During Execution)

<!-- Workers add amendments here if prerequisites or instructions are contradictory. -->
