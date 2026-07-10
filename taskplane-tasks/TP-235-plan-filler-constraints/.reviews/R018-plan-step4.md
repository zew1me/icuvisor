# Plan Review: TP-235 Step 4 — Align roadmap and product contract

**Verdict:** REVISE

## Required revisions

1. **Specify the v2.0 roadmap acceptance criteria to add, rather than only saying they will be added.** The v2.0 bullet should explicitly require that Plan Filler validates a proposed schedule before its separate commit; keeps requested session count separate from athlete-local availability/independent slots; enforces per-session, daily, indoor/mode/sport, and remaining-week time/load caps; accounts for completed and protected fixed commitments; and reports deterministic violations/reconciliation without silently redistributing deficits. State that this is a prerequisite/acceptance criterion for the future `fill_calendar_from_library` tool, not a claim that the tool or write path exists today.

2. **Make the PRD and changelog disposition an explicit planned decision.** The PRD has no `fill_calendar_from_library` catalog entry, while the roadmap is the authoritative forward-looking phasing document. The plan should state that this internal, unregistered validator does not change the current public MCP contract, so no PRD catalog or Unreleased changelog entry will be made unless the review identifies a conflicting existing promise. Record the PRD and changelog review (and the no-change rationale) in `STATUS.md`; do not leave “if needed” unresolved.

3. **Plan a content-level documentation verification in addition to formatting.** After editing, inspect the v2.0/v2.2 and design-document status wording together to confirm: (a) the roadmap says the validator is an acceptance criterion for future Plan Filler rather than shipped functionality, and (b) v2.2 remains limited to evidence-based ramp/recovery/taper/intensity guardrails and does not duplicate deterministic placement/budget validation. `gofmt` and `git diff --check` cannot validate either condition. Include the exact review command(s), such as `git diff -- ROADMAP.md docs/prd/PRD-icuvisor.md docs/design/plan-filler-constraints.md CHANGELOG.md`, in the step verification plan.
