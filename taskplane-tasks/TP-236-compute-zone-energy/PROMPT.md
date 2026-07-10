# Task: TP-236 - Add deterministic power-zone energy analysis

**Created:** 2026-07-10
**Size:** L

## Review Level: 2 (Plan and Code)

**Assessment:** This adds a stream-backed analyzer, MCP tool, catalog entry, formula documentation, and generated schemas. Numerical integration and missing-stream behavior affect model-visible training analysis across multiple packages.
**Score:** 5/8 — Blast radius: 2, Pattern novelty: 2, Security: 0, Reversibility: 1

## Canonical Task Folder

```
taskplane-tasks/TP-236-compute-zone-energy/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Add a full-toolset, read-only `compute_zone_energy` analyzer that reports mechanical work in kJ by configured power zone over a bounded athlete-local date range. Integrate canonical power streams using sample timestamps, expose time and energy per zone plus exact method/source metadata, and degrade explicitly when power, timestamps, activities, or athlete power zones are unavailable. Keep responses terse and never present mechanical work as metabolic calories or energy expenditure.

## Dependencies

- **None**

## Context to Read First

**Tier 2 (area context):**

- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**

- `docs/prd/PRD-icuvisor.md` — analyzer rules, unit discipline, terse responses, and tool catalog
- `internal/tools/compute_zone_time.go` — existing range aggregation and analyzer metadata pattern
- `internal/tools/get_activity_histogram.go` — canonical stream sampling and configured-zone selection patterns
- `internal/resources/analysis_formulas.go` — formula registry pattern

## Environment

- **Workspace:** repository root
- **Services required:** None; tests use fixtures only

## File Scope

- `internal/analysis/zone_energy.go`
- `internal/analysis/zone_energy_test.go`
- `internal/tools/compute_zone_energy.go`
- `internal/tools/compute_zone_energy_test.go`
- `internal/tools/registry.go`
- `internal/tools/registry_test.go`
- `internal/tools/catalog_tiers_test.go`
- `internal/tools/schema_snapshot/compute_zone_energy.json`
- `internal/resources/analysis_formulas.go`
- `internal/resources/analysis_formulas_test.go`
- `internal/resources/testdata/analysis_formulas.md`
- `internal/toolcatalog/catalog.go`
- `internal/toolchecks/schema_stability_test.go`
- `internal/safety/adversarial_test.go`
- `internal/mcp/protocol_test.go`
- `web/data/tools.json`
- `web/data/tool_schemas.json`
- `web/content/cookbook/weekly-review.md`
- `docs/prd/PRD-icuvisor.md`
- `ROADMAP.md`
- `CHANGELOG.md`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Existing stream keys, time-base semantics, power-zone boundaries, and analyzer metadata reviewed

### Step 1: Lock the numerical and response contract

**Plan-review checkpoint**

- [ ] Define timestamp-weighted work integration, final-sample handling, zone-boundary inclusion, zero-power handling, and J-to-kJ conversion
- [ ] Define bounded date-range/activity limits and explicit partial/insufficient statuses
- [ ] Define terse zone rows and optional per-activity audit rows with `_meta.method`, source tools, units, missing activities, and assumptions
- [ ] State clearly that power-derived mechanical work is not metabolic energy or food calories
- [ ] Run targeted tests: `go test ./internal/analysis -run 'ZoneEnergy'`

**Artifacts:**

- `internal/analysis/zone_energy.go` (new)
- `internal/resources/testdata/analysis_formulas.md` (modified)

### Step 2: Implement and test the pure calculation

- [ ] Integrate watts over valid time deltas and classify each interval into configured power zones deterministically
- [ ] Reject or skip non-finite, negative-duration, missing, and misaligned samples with explicit counters rather than guessed values
- [ ] Return total time, total kJ, zone time, zone kJ, and energy share with stable rounding
- [ ] Add table-driven tests for irregular timestamps, coasting, exact boundaries, missing samples, and unit conversion
- [ ] Run targeted tests: `go test ./internal/analysis -run 'ZoneEnergy'`

**Artifacts:**

- `internal/analysis/zone_energy.go` (new)
- `internal/analysis/zone_energy_test.go` (new)

### Step 3: Add and register the MCP analyzer

- [ ] Implement range-based activity discovery, configured power-zone lookup, canonical stream fetches, sport filtering, and deterministic caps
- [ ] Emit partial results when only some activities have usable power streams, including skipped activity IDs/reasons only behind appropriate terse/full shaping
- [ ] Register `compute_zone_energy` in the full toolset with read-only annotations, input schema, output schema, activation hint, and no raw streams in the response
- [ ] Add catalog, safety, protocol, tier, and schema-stability coverage
- [ ] Run targeted tests: `go test ./internal/tools ./internal/toolcatalog ./internal/toolchecks ./internal/safety ./internal/mcp -run 'ZoneEnergy|Catalog|Schema|Registry'`

**Artifacts:**

- `internal/tools/compute_zone_energy.go` (new)
- `internal/tools/compute_zone_energy_test.go` (new)
- `internal/tools/registry.go` (modified)
- `internal/tools/schema_snapshot/compute_zone_energy.json` (new)
- `internal/toolcatalog/catalog.go` (modified)

### Step 4: Formula resource, generated data, and docs

- [ ] Add the pinned integration formula and interpretation boundary to `icuvisor://analysis-formulas`
- [ ] Update PRD analyzer catalog and roadmap disposition without implying metabolic-energy analysis
- [ ] Regenerate website tool/schema data and add a concise cookbook example
- [ ] Record the new capability under Unreleased
- [ ] Run targeted checks: `make docs-tools && go test ./internal/resources ./internal/toolcatalog ./internal/toolchecks`

**Artifacts:**

- `internal/resources/analysis_formulas.go` (modified)
- `internal/resources/analysis_formulas_test.go` (modified)
- `internal/resources/testdata/analysis_formulas.md` (modified)
- `web/data/tools.json` (modified)
- `web/data/tool_schemas.json` (modified)
- `web/content/cookbook/weekly-review.md` (modified)
- `docs/prd/PRD-icuvisor.md` (modified)
- `ROADMAP.md` (modified)
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

- `docs/prd/PRD-icuvisor.md` — add the analyzer and its interpretation boundary
- `CHANGELOG.md` — record the new read-only capability
- `internal/resources/testdata/analysis_formulas.md` — publish the pinned formula

**Check If Affected:**

- `ROADMAP.md` — mark the energy-by-zone opportunity appropriately
- `web/content/cookbook/weekly-review.md` — include only a grounded example
- Generated tool reference/schema data

## Completion Criteria

- [ ] `compute_zone_energy` is registered only in the full toolset and is read-only
- [ ] Results are timestamp-weighted and expressed in seconds and kJ with exact metadata
- [ ] Missing streams/zones produce partial or insufficient results, never zeros masquerading as data
- [ ] Mechanical work is never labelled as calories or metabolic energy
- [ ] New analysis and tool test files exist
- [ ] Catalog, schema, resource, full tests, race, lint, build, and generated docs pass

## Git Commit Convention

Commits happen at step boundaries. All commits MUST include TP-236:

- **Step completion:** `feat(TP-236): complete Step N — description`
- **Bug fixes:** `fix(TP-236): description`
- **Tests:** `test(TP-236): description`
- **Hydration:** `hydrate: TP-236 expand Step N checkboxes`

## Do NOT

- Treat kJ of mechanical work as dietary calories or metabolic expenditure
- Return raw streams or unbounded per-sample payloads
- Interpolate missing power silently
- Invent power zones when athlete settings are absent
- Add the stream-heavy tool to compact or core without benchmark evidence
- Hit live intervals.icu endpoints in tests
- Commit without TP-236 in the message

---

## Amendments (Added During Execution)

<!-- Workers add amendments here if prerequisites or instructions are contradictory. -->

- Step 3 also permits the actual catalog/registration surfaces `internal/tools/catalog.go`, `internal/tools/catalog_test.go`, and MCP annotation wiring in `internal/mcp/registrar_tools.go` plus its protocol tests. Plan review R008 identified these as required because `registryBaseTools` is the catalog source of truth and read-only SDK annotations are not currently derived from safety requirements.
