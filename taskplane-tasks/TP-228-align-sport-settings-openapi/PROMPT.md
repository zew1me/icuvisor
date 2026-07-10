# Task: TP-228 - Align sport-settings update and apply requests with live OpenAPI

**Created:** 2026-07-10
**Size:** L

## Review Level: 3 (Full)

**Assessment:** This changes an MCP write contract and removes a misleading date boundary from a call that can trigger broad historical recomputation. A mistake can either break all sport-settings writes or mutate historical activity analysis unexpectedly.
**Score:** 6/8 — Blast radius: 2, Pattern novelty: 1, Security: 1, Reversibility: 2

## Canonical Task Folder

```
taskplane-tasks/TP-228-align-sport-settings-openapi/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Bring the sport-settings update/apply transport and MCP surface into exact agreement with the live intervals.icu OpenAPI contract. `PUT /athlete/{athleteId}/sport-settings/{id}` requires the `recalcHrZones` query parameter; `PUT .../{id}/apply` accepts no date parameter and no request body. Remove the false `effective_date` promise and prevent `update_sport_settings` from silently starting an unbounded historical apply. Preserve terse actionable errors, additive schema discipline where defensible, and the existing delete-mode protection for zone replacement.

## Dependencies

- **None**

## Context to Read First

**Tier 2 (area context):**

- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**

- `docs/prd/PRD-icuvisor.md` — sport-settings behavior and destructive-operation policy
- `ROADMAP.md` — current stable-line scope
- `docs/dogfood/v0.3-findings.md` — prior live write evidence that may now be stale
- `https://intervals.icu/api/v1/docs` — authoritative live OpenAPI; verify required parameters before implementation

## Environment

- **Workspace:** repository root
- **Services required:** None; tests must use local HTTP fixtures and must not call a real athlete account

## File Scope

- `internal/intervals/client.go`
- `internal/intervals/sport_settings.go`
- `internal/intervals/sport_settings_test.go`
- `internal/intervals/sport_settings_openapi_contract_test.go`
- `internal/tools/update_sport_settings.go`
- `internal/tools/update_sport_settings_test.go`
- `internal/tools/schema_snapshot/update_sport_settings.json`
- `internal/toolchecks/schema_stability_test.go`
- `docs/upstream-gaps/sport-settings-write-contract.md`
- `web/data/tools.json`
- `web/data/tool_schemas.json`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Reconfirm the update and apply operations in the live public OpenAPI without using credentials

### Step 1: Define the corrected write boundary

**Plan-review checkpoint**

- [ ] Document the exact update/apply request shapes and decide how the public MCP schema handles removal of the non-functional `effective_date`
- [ ] Ensure `update_sport_settings` performs only the settings update and cannot implicitly trigger an all-history apply
- [ ] Define response metadata that reports HR-zone recalculation accurately without claiming a date-scoped historical recompute
- [ ] Run targeted tests: `go test ./internal/intervals ./internal/tools`

**Artifacts:**

- `docs/upstream-gaps/sport-settings-write-contract.md` (new)
- `internal/tools/update_sport_settings.go` (modified)

### Step 2: Align the intervals.icu client

- [ ] Send required `recalcHrZones=true|false` on every sport-settings update request, defaulting the tool input to true
- [ ] Make `ApplySportSettings` issue the documented PUT with no unsupported `oldest` query or JSON body
- [ ] Remove automatic apply behavior from `UpdateSportSettings`
- [ ] Preserve cancellation, wrapped errors, retry behavior, and sparse update bodies
- [ ] Run targeted tests: `go test ./internal/intervals -run 'SportSettings'`

**Artifacts:**

- `internal/intervals/client.go` (modified if a body-plus-query helper is required)
- `internal/intervals/sport_settings.go` (modified)
- `internal/intervals/sport_settings_openapi_contract_test.go` (new)

### Step 3: Align the MCP schema and response

- [ ] Add documented `recalc_hr_zones` input with default true and forward both true and false exactly
- [ ] Remove the unsupported effective-date behavior from decoding, examples, metadata, and descriptions; do not retain a field that merely pretends to scope upstream work
- [ ] Replace `recompute_pending` or equivalent claims with truthful update/recalculation metadata
- [ ] Regenerate schema snapshots and website catalog data
- [ ] Run targeted tests: `go test ./internal/tools ./internal/toolchecks`

**Artifacts:**

- `internal/tools/update_sport_settings.go` (modified)
- `internal/tools/update_sport_settings_test.go` (modified)
- `internal/tools/schema_snapshot/update_sport_settings.json` (modified)
- `internal/toolchecks/schema_stability_test.go` (modified if contract approval requires it)
- `web/data/tools.json` (modified)
- `web/data/tool_schemas.json` (modified)

### Step 4: Regression coverage

- [ ] Assert exact update method, path, query, and sparse body for true and false recalculation modes
- [ ] Assert apply uses PUT with no query parameters and no semantic request payload
- [ ] Assert `update_sport_settings` never calls apply implicitly
- [ ] Assert invalid/extra legacy arguments return the terse public validation error rather than reaching upstream
- [ ] Run targeted tests: `go test ./internal/intervals ./internal/tools`

**Artifacts:**

- `internal/intervals/sport_settings_openapi_contract_test.go` (new)
- `internal/intervals/sport_settings_test.go` (modified)
- `internal/tools/update_sport_settings_test.go` (modified)

### Step 5: Testing & Verification

**Code review checkpoint**

- [ ] Run FULL test suite: `make test`
- [ ] Run race suite: `make test-race`
- [ ] Run lint: `make lint`
- [ ] Fix all failures
- [ ] Build passes: `make build`
- [ ] Generated tool docs are clean: `make docs-tools && git diff --check`

### Step 6: Documentation & Delivery

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] Discoveries logged in STATUS.md

## Documentation Requirements

**Must Update:**

- `docs/upstream-gaps/sport-settings-write-contract.md` — record the observed OpenAPI contract and explicitly state that apply is not date-scoped
- `CHANGELOG.md` — describe the corrected update contract and removal of implicit historical apply under Unreleased

**Check If Affected:**

- `docs/prd/PRD-icuvisor.md` — remove or clarify effective-date/history claims if present
- `web/content/reference/tools.md` and generated tool data — ensure public schema matches the registry
- `docs/dogfood/v0.3-findings.md` — annotate stale evidence rather than rewriting historical results

## Completion Criteria

- [ ] Every update request includes `recalcHrZones`
- [ ] Apply sends no unsupported date or payload
- [ ] Updating settings cannot silently trigger broad historical recomputation
- [ ] MCP schema and metadata make no false date-scope claim
- [ ] New exact-wire regression test file exists
- [ ] All tests, lint, generated docs, and build pass

## Git Commit Convention

Commits happen at step boundaries. All commits MUST include TP-228:

- **Step completion:** `fix(TP-228): complete Step N — description`
- **Bug fixes:** `fix(TP-228): description`
- **Tests:** `test(TP-228): description`
- **Hydration:** `hydrate: TP-228 expand Step N checkboxes`

## Do NOT

- Call the live write/apply endpoints or use a real API key
- Preserve `effective_date` by silently ignoring it
- Add a model-controlled confirmation flag as a safety boundary
- Automatically apply settings to historical activities
- Copy implementation code from GPL competitors
- Expand task scope into pace-value semantics; TP-229 owns that work
- Skip tests or generated catalog updates
- Commit without the task ID prefix

---

## Amendments (Added During Execution)

<!-- Workers add amendments here if prerequisites or instructions are contradictory. -->
