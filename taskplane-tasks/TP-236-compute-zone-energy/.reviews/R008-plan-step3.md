# Plan Review — TP-236 Step 3

## Verdict: REVISE

The orchestration/response contract carried forward from R001–R003 is detailed enough, but the Step 3 change plan does not match the repository's actual registration and protocol surfaces.

1. **The planned registration file is not the catalog source of truth.** Default tools and `tools.Catalog()` are both built from `registryBaseTools` in `internal/tools/catalog.go`, not `internal/tools/registry.go`. Adding the tool only through `Registry.Register` would omit it from generated catalog metadata, and the current file scope/artifact list does not permit the required `catalog.go` edit. Amend the plan/scope to add `newComputeZoneEnergyTool(...)` in `registryBaseTools`, classify it in `toolCatalogGroup`, and update `internal/tools/catalog_test.go`'s analyzer-family activation list as well as the tier/registry tests. This also ensures `icuvisor_list_advanced_capabilities` sees the full-only analyzer.

2. **“Read-only annotations” currently has no implementation path.** `RequirementRead` controls safety filtering, but `internal/mcp/registrar_tools.go` does not populate SDK `Tool.Annotations`; tools/list therefore will not expose `annotations.readOnlyHint: true`. The plan must say whether Step 3 will add an annotation field to the internal tool model or map `RequirementRead` to SDK annotations in the registrar, include the necessary production file(s) in scope, and add a protocol assertion against tools/list. Merely checking `Requirement == RequirementRead` is not the requested MCP annotation.

3. **Make the cross-activity numerical aggregation explicit.** `analysis.ComputeZoneEnergy` exposes already rounded per-activity zone totals. State whether Step 3 sums those displayed values (so aggregate rows, headline totals, and full audit rows reconcile exactly) or introduces an unrounded accumulator, then test multiple activities/configuration groups and the response-wide share remainder rule. The plan already fixes ordering and denominators, but not this rounding boundary.

Retain the planned error-classification matrix, terse/full skip-reason checks, exact insufficient-state table, no-raw-stream assertion, 201/200 cap ordering, serialized `_meta.analysis_units`, coach-scoped catalog membership, schema snapshot count/update, safety matrix, and full-only protocol visibility coverage.
