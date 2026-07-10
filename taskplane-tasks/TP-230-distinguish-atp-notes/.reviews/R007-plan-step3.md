# R007 — Plan Review: Step 3 (Update schema and generated surfaces)

**Verdict: APPROVE**

---

## What Step 3 Is Doing

Three mechanical tasks follow from the Step 1-2 implementation:

1. **Confirm/update the schema snapshot** — run the snapshot script to verify the committed file in `internal/tools/schema_snapshot/` still matches the live registry.
2. **Regenerate website tool data** — run `make docs-tools` to update `web/data/tools.json` and `web/data/tool_schemas.json`.
3. **Update the tool description** — change `getAnnualTrainingPlanDescription` and `getAnnualTrainingPlanOutputSchema()` to use provenance-aware terminology and explicitly tell models that `context_notes` are personal calendar notes, not ATP instructions.

The plan is sound. One nuance below to avoid wasted effort.

---

## Key Finding: The Schema Snapshot File Will Not Change

The snapshot in `internal/tools/schema_snapshot/get_annual_training_plan.json` captures only the **input** schema (confirmed in `toolchecks/schema_stability.go`: `CanonicalJSON(tool.InputSchema)`). Steps 1-2 changed only output shaping — the five input parameters (`oldest`, `newest`, `calendar_id`, `limit`, `include_full`) are unchanged.

Running `go run ./scripts/snapshot_tool_schemas.go` is the correct verification step, but it will produce **no diff**. The PROMPT artifact list marks this file as `(modified)` — that is incorrect. Do not force-edit the snapshot file; just confirm the script exits cleanly.

The toolchecks and schema-stability tests already pass against the current snapshot, confirming this.

---

## What Will Actually Change

| File | Change |
|------|--------|
| `internal/tools/get_annual_training_plan.go` | Update `getAnnualTrainingPlanDescription` and `getAnnualTrainingPlanOutputSchema()` description string |
| `web/data/tools.json` | Regenerated from updated description (via `make docs-tools`) |
| `web/data/tool_schemas.json` | Regenerated from updated description (via `make docs-tools`) |
| `internal/tools/schema_snapshot/get_annual_training_plan.json` | No change expected |

Note that `get_annual_training_plan.go` is not listed in the Step 3 PROMPT artifacts, but it must be edited here — it is the source of both the tool description and the output schema description. The prior steps already modified this file; Step 3 adds description-only changes.

---

## Required Description Changes

### `getAnnualTrainingPlanDescription` (the model-visible tool description)

Current text includes "recovery weeks" (stale — recovery classification was removed) and "recovery/context notes" (old framing). It must:
- Remove "recovery weeks" from the trigger list.
- Use "ATP-generated notes" for `notes` (provenance-verified).
- Explicitly state that `context_notes` are **personal calendar notes, not ATP instructions**.

### `getAnnualTrainingPlanOutputSchema()` description

Currently says "recovery/context notes" — should be updated to match the new `notes`/`context_notes` naming and clarify what each collection represents. This string is not captured by the snapshot or web-data generator, but it is part of the MCP tool contract exposed to models via the output schema field.

---

## Checklist Correctness

| Checklist item | Assessment |
|----------------|------------|
| Update output schema snapshot for the provenance-aware note shape | Run snapshot script to confirm freshness; file will not change |
| Regenerate website tool data | Required after description update; correct |
| Ensure the tool description tells models that personal context notes are not ATP instructions | Required; must cover both `getAnnualTrainingPlanDescription` and `getAnnualTrainingPlanOutputSchema()` |

---

## Test Command

`go test ./internal/tools ./internal/toolchecks` is correct. The `TestCheckSnapshotFreshness` path in toolchecks will confirm the snapshot is still in sync after the description-only edit.

---

## No Other Risks

- No input schema changes → no stability-check concerns.
- No new tool arguments → no breaking-change review needed.
- Description-only edits to `web/data/` are safe and self-contained.
- Step 4 will run the full suite; Step 3 only needs the targeted test command.
