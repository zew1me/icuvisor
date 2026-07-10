# Task: TP-238 - Add grounded fueling review prompt pack

**Created:** 2026-07-10
**Size:** M

## Review Level: 1 (Plan Only)

**Assessment:** This adds a read-only MCP prompt, portable prompt pack, docs, and evals using already exposed nutrition fields. It follows established patterns, but unit, missing-data, and health-claim boundaries require explicit review.
**Score:** 3/8 — Blast radius: 1, Pattern novelty: 1, Security: 0, Reversibility: 1

## Canonical Task Folder

```
taskplane-tasks/TP-238-fueling-review-prompt/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Add a read-only `fueling_review` MCP prompt and portable prompt pack that uses Icuvisor's existing activity fueling fields, wellness nutrition fields, session duration/load, and race/calendar context to review what the athlete actually logged. The workflow must calculate only transparent quantities such as logged grams per hour, distinguish ingested carbs from upstream estimated carbs used and wellness intake, state missing logs instead of estimating them, and keep general planning guidance separate from medical or individualized dietary claims.

## Dependencies

- **None**

## Context to Read First

**Tier 2 (area context):**

- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**

- `docs/prd/PRD-icuvisor.md` — prompt catalog, response units, and analyzer rules
- `internal/tools/get_activity_details.go` — activity fueling field semantics
- `web/content/cookbook/prompt-library.md` — existing short nutrition prompts
- `internal/prompts/catalog.go` — registered prompt patterns

## Environment

- **Workspace:** repository root
- **Services required:** None

## File Scope

- `internal/prompts/catalog.go`
- `internal/prompts/registry.go`
- `internal/prompts/catalog_test.go`
- `internal/prompts/fueling_review_test.go`
- `internal/prompts/testdata/fueling_review.md`
- `docs/prompts/client-prompt-packs/fueling-review.md`
- `web/content/cookbook/fueling-review.md`
- `web/content/cookbook/prompt-library.md`
- `web/content/cookbook/_index.md`
- `web/content/reference/resources-prompts.md`
- `scripts/eval/scenarios/cookbook_scenarios.json`
- `docs/prd/PRD-icuvisor.md`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Current fueling, nutrition, duration, load, calendar, and custom-field semantics reviewed

### Step 1: Define the fueling evidence contract

**Plan-review checkpoint**

- [ ] Define activity/session and date-range review modes using only currently exposed fields and tools
- [ ] Distinguish `carbs_ingested_g`, `carbs_used_g`, wellness intake/macros, custom fields, and missing logs in every instruction and example
- [ ] Define transparent grams-per-hour calculations with valid-duration and sample-coverage requirements
- [ ] Add health boundaries: no diagnosis, no invented sweat/sodium needs, no product claims, and no personalized medical nutrition prescription
- [ ] Run targeted tests: `go test ./internal/prompts -run 'FuelingReview'`

**Artifacts:**

- `internal/prompts/catalog.go` (modified)
- `docs/prompts/client-prompt-packs/fueling-review.md` (new)

### Step 2: Register the prompt and add regression coverage

- [ ] Add and register `fueling_review` with bounded optional activity/date/race context and the minimum necessary read tools
- [ ] Require athlete-local date anchoring, terse reads, unit-labelled calculations, and explicit missing-data counts
- [ ] Add a deterministic golden fixture and new focused tests for field distinctions, invalid/missing duration, no-log behavior, and read-only operation
- [ ] Ensure the prompt never asks the model to fabricate a food/product library or write wellness/activity data
- [ ] Run targeted tests: `go test ./internal/prompts -run 'Prompt|FuelingReview'`

**Artifacts:**

- `internal/prompts/catalog.go` (modified)
- `internal/prompts/registry.go` (modified)
- `internal/prompts/fueling_review_test.go` (new)
- `internal/prompts/testdata/fueling_review.md` (new)

### Step 3: Publish cookbook, portable pack, and evals

- [ ] Add a cookbook workflow for recent-session review and race/session planning that labels sourced facts versus general guidance
- [ ] Expand the prompt library with concise, field-correct examples and add a copyable client prompt pack
- [ ] Add eval scenarios for logged grams/hour, missing logs, carbs-used versus carbs-ingested, and refusal to invent sodium/calorie targets
- [ ] Update prompt reference, cookbook index, PRD catalog, and Unreleased changelog
- [ ] Run targeted checks: `go test ./internal/prompts && python3 scripts/eval/run_eval.py --validate`

**Artifacts:**

- `docs/prompts/client-prompt-packs/fueling-review.md` (new)
- `web/content/cookbook/fueling-review.md` (new)
- `web/content/cookbook/prompt-library.md` (modified)
- `web/content/cookbook/_index.md` (modified)
- `web/content/reference/resources-prompts.md` (modified)
- `scripts/eval/scenarios/cookbook_scenarios.json` (modified)
- `docs/prd/PRD-icuvisor.md` (modified)
- `CHANGELOG.md` (modified)

### Step 4: Testing & Verification

- [ ] Run FULL test suite: `make test`
- [ ] Run prompt eval validation: `make eval-validate`
- [ ] Run lint: `make lint`
- [ ] Fix all failures
- [ ] Build passes: `make build`
- [ ] Verify clean Markdown and diff: `git diff --check`

### Step 5: Documentation & Delivery

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] Discoveries logged in STATUS.md

## Documentation Requirements

**Must Update:**

- `web/content/cookbook/fueling-review.md` — document evidence fields, calculations, limits, and examples
- `web/content/reference/resources-prompts.md` — document the registered prompt
- `docs/prd/PRD-icuvisor.md` — add the prompt to the current catalog
- `CHANGELOG.md` — record the new prompt and prompt pack

**Check If Affected:**

- `web/content/cookbook/prompt-library.md` — align existing fueling examples
- `web/content/cookbook/activity-retrospective.md` — link the dedicated workflow if useful
- `README.md` — update only if it lists every registered prompt

## Completion Criteria

- [ ] `fueling_review` is registered and read-only
- [ ] Activity carbs ingested, carbs used, and wellness nutrition remain semantically distinct
- [ ] Grams/hour calculations require valid logged grams and duration
- [ ] Missing data is reported rather than estimated
- [ ] No medical, product, sodium, sweat-rate, or calorie claims are invented
- [ ] New focused test, golden fixture, portable pack, cookbook page, and eval scenarios exist
- [ ] Full tests, eval validation, lint, build, and diff checks pass

## Git Commit Convention

Commits happen at step boundaries. All commits MUST include TP-238:

- **Step completion:** `feat(TP-238): complete Step N — description`
- **Bug fixes:** `fix(TP-238): description`
- **Tests:** `test(TP-238): description`
- **Hydration:** `hydrate: TP-238 expand Step N checkboxes`

## Do NOT

- Add a proprietary food, gel, drink, or product database
- Invent missing intake, sodium, hydration, sweat-rate, calorie, or medical values
- Collapse carbs used and carbs ingested into one field
- Present general guidance as individualized medical nutrition advice
- Call write or delete tools
- Add direct wearable or nutrition-service integrations
- Commit without TP-238 in the message

---

## Amendments (Added During Execution)

<!-- Workers add amendments here if prerequisites or instructions are contradictory. -->
