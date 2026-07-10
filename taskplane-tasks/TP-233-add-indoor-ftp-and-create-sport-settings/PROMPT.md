# Task: TP-233 - Support indoor FTP and missing sport-setting creation

**Created:** 2026-07-10
**Size:** L

## Review Level: 2 (Plan and Code)

**Assessment:** This adds a new MCP write tool, client operation, catalog entry, and an additive field to an existing write tool. It touches remote athlete configuration but follows established sparse-write and registration patterns.
**Score:** 4/8 — Blast radius: 2, Pattern novelty: 1, Security: 0, Reversibility: 1

## Canonical Task Folder

```
taskplane-tasks/TP-233-add-indoor-ftp-and-create-sport-settings/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Close two sport-settings capability gaps after the transport and pace contracts are corrected. Add `indoor_ftp` as a safe threshold field on `update_sport_settings`, and add a full-toolset `create_sport_settings` tool for athletes who have no existing setting for a sport. Use the official POST `/athlete/{athleteId}/sport-settings` contract with `types:[sport]`, the corrected m/s threshold semantics from TP-229, terse confirmations, typed schemas, input examples, and no credential parameters. Keep initial creation threshold-only; do not smuggle destructive zone replacement into the create path.

## Dependencies

- **Task:** TP-228 (correct update/apply transport contract)
- **Task:** TP-229 (correct threshold pace and pace-zone semantics)

## Context to Read First

**Tier 2 (area context):**

- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**

- `docs/prd/PRD-icuvisor.md` — tool catalog and sport-settings behavior
- `ROADMAP.md` — current feature phasing
- `https://intervals.icu/api/v1/docs` — create/update SportSettings operations and schema

## Environment

- **Workspace:** repository root
- **Services required:** None; tests use local HTTP fixtures only

## File Scope

- `internal/intervals/sport_settings.go`
- `internal/intervals/sport_settings_test.go`
- `internal/tools/update_sport_settings.go`
- `internal/tools/update_sport_settings_test.go`
- `internal/tools/create_sport_settings.go`
- `internal/tools/create_sport_settings_test.go`
- `internal/tools/registry.go`
- `internal/tools/registry_test.go`
- `internal/tools/catalog_tiers_test.go`
- `internal/tools/schema_snapshot/create_sport_settings.json`
- `internal/tools/schema_snapshot/update_sport_settings.json`
- `internal/toolcatalog/catalog.go`
- `internal/toolchecks/schema_stability_test.go`
- `internal/mcp/protocol_test.go`
- `web/data/tools.json`
- `web/data/tool_schemas.json`
- `docs/prd/PRD-icuvisor.md`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] TP-228 and TP-229 are complete
- [ ] Confirm POST create fields and response in the current public OpenAPI

### Step 1: Extend the typed client

**Plan-review checkpoint**

- [ ] Add sparse indoor FTP support to the corrected update request and response echo
- [ ] Add typed `CreateSportSettings` client operation using POST and `types:[sport]`
- [ ] Reuse corrected threshold pace conversion and explicit `pace_load_type` behavior
- [ ] Validate positive FTP/indoor FTP/HR values and reject an indoor FTP greater than an explicitly supplied outdoor FTP only if product rules justify that restriction; do not invent an upstream constraint
- [ ] Run targeted tests: `go test ./internal/intervals -run 'SportSettings'`

**Artifacts:**

- `internal/intervals/sport_settings.go` (modified)
- `internal/intervals/sport_settings_test.go` (modified)

### Step 2: Add and register MCP surfaces

- [ ] Add `indoor_ftp` to `update_sport_settings` schema, examples, echo, and fields-updated metadata
- [ ] Implement `create_sport_settings` with sport plus optional outdoor FTP, indoor FTP, threshold HR, and threshold pace; require at least one meaningful threshold field in addition to sport unless the upstream default-create behavior is explicitly chosen and documented
- [ ] Register the new tool in the full toolset with write/create annotations and no model-controlled delete override
- [ ] Add output schema, input examples, catalog constants, tier expectations, and stable snapshots
- [ ] Run targeted tests: `go test ./internal/tools ./internal/toolcatalog ./internal/toolchecks ./internal/mcp`

**Artifacts:**

- `internal/tools/update_sport_settings.go` (modified)
- `internal/tools/create_sport_settings.go` (new)
- `internal/tools/registry.go` (modified)
- `internal/toolcatalog/catalog.go` (modified)
- `internal/tools/schema_snapshot/create_sport_settings.json` (new)
- `internal/tools/schema_snapshot/update_sport_settings.json` (modified)

### Step 3: Regression and safety coverage

- [ ] Add exact-wire tests for create bodies across Ride, Run, and Swim, including indoor FTP and m/s threshold pace
- [ ] Assert duplicate/missing sport validation is actionable and no network call occurs on invalid input
- [ ] Assert tool schemas expose no API key, credential, confirm, or zone-replacement arguments
- [ ] Assert core/default tool counts and full-toolset catalog registration are updated deliberately
- [ ] Run targeted tests: `go test ./internal/intervals ./internal/tools ./internal/mcp`

**Artifacts:**

- `internal/tools/create_sport_settings_test.go` (new)
- `internal/tools/update_sport_settings_test.go` (modified)
- `internal/tools/registry_test.go` (modified)
- `internal/tools/catalog_tiers_test.go` (modified)
- `internal/mcp/protocol_test.go` (modified)

### Step 4: Generated docs and public contract

- [ ] Add the new tool and indoor FTP write semantics to the PRD catalog
- [ ] Regenerate website tool/schema data and confirm tool count/hash guards
- [ ] Add an Unreleased changelog entry
- [ ] Run targeted tests: `make docs-tools && go test ./internal/toolcatalog ./internal/toolchecks`

**Artifacts:**

- `docs/prd/PRD-icuvisor.md` (modified)
- `web/data/tools.json` (modified)
- `web/data/tool_schemas.json` (modified)
- `CHANGELOG.md` (modified)

### Step 5: Testing & Verification

**Code review checkpoint**

- [ ] Run FULL test suite: `make test`
- [ ] Run race suite: `make test-race`
- [ ] Run lint: `make lint`
- [ ] Fix all failures
- [ ] Build passes: `make build`
- [ ] Regenerate docs and verify clean diff: `make docs-tools && git diff --check`

### Step 6: Documentation & Delivery

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] Discoveries logged in STATUS.md

## Documentation Requirements

**Must Update:**

- `docs/prd/PRD-icuvisor.md` — add create_sport_settings and indoor FTP write behavior
- `CHANGELOG.md` — record both new capabilities under Unreleased

**Check If Affected:**

- Generated tool reference/schema data
- `web/content/cookbook/ftp-and-zones.md` — add a safe indoor FTP or missing-sport example if useful
- `README.md` — update only if it enumerates tool counts or sport-setting writes

## Completion Criteria

- [ ] update_sport_settings can write indoor FTP through the corrected client contract
- [ ] create_sport_settings can create a threshold-only setting for a missing sport
- [ ] New tool is full-toolset, typed, documented, and credential-free
- [ ] Create path cannot replace zones
- [ ] New create tool and test files exist
- [ ] Catalog/hash/count guards, full tests, race, lint, build, and generated docs pass

## Git Commit Convention

Commits happen at step boundaries. All commits MUST include TP-233:

- **Step completion:** `feat(TP-233): complete Step N — description`
- **Bug fixes:** `fix(TP-233): description`
- **Tests:** `test(TP-233): description`
- **Hydration:** `hydrate: TP-233 expand Step N checkboxes`

## Do NOT

- Start before TP-228 and TP-229 are complete
- Add API keys or athlete credentials to tool arguments
- Add zone definitions to create_sport_settings in this task
- Automatically apply settings to historical activities
- Guess pace storage units; reuse TP-229's m/s boundary
- Add a delete/confirm override
- Call the live write endpoint
- Commit without TP-233 in the message

---

## Amendments (Added During Execution)

<!-- Workers add amendments here if prerequisites or instructions are contradictory. -->
