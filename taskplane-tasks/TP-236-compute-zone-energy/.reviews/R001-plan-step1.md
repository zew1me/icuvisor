# Plan Review — TP-236 Step 1

## Verdict: REVISE

No Step 1 implementation plan or contract has been submitted. `STATUS.md` only repeats the unchecked checkpoint headings, while neither planned artifact exists. For this step, “define” is the work: the implementation cannot safely infer model-visible numerical and response behavior from the mission statement.

A revised plan must lock the following decisions.

### 1. Exact integration and invalid-sample rules

Specify the pure calculator input/output types and the precise formula, including:

- timestamp units and whether integration is the left-endpoint rule `work_j += watts[i] * (time[i+1]-time[i])`;
- that the final sample contributes no duration/work without a following timestamp;
- whether stream-length mismatch rejects the activity or processes an aligned prefix (the task says not to guess, so this cannot remain implicit);
- interval behavior for non-finite power/timestamps, negative power, duplicate/reversed timestamps, and any large timestamp gap;
- whether a bad sample invalidates only its own left-hand interval or both adjacent intervals;
- zero watts as valid coasting time, its zone assignment, and the all-zero-energy share result;
- accumulation before rounding, `1 kJ = 1000 W·s`, output precision for seconds/kJ/shares, and how rounded zone totals reconcile with rounded headline totals.

The contract must name every diagnostic counter returned by the pure calculation (usable intervals plus each skipped reason), rather than leaving Step 2 to invent them.

### 2. Configured-zone semantics

Define how `SportSettings.PowerZones` is interpreted and validated: boundary order, duplicate/non-finite/non-positive boundaries, name-to-boundary pairing, and whether sorting is allowed. State exact inclusion (`lower <= watts < upper`, with an open final zone) and what happens below the first configured boundary, especially at zero watts. Also define sport-setting selection using `Type`/`Types` and activity type/subtype.

The range tool needs an explicit policy for a window containing activities that resolve to different power-zone configurations. It must either require/select one sport setting, aggregate by ordinal with all distinct configurations disclosed, or decline incompatible rows; silently combining kJ under one set of labels/bounds would be incorrect.

### 3. Range limits and status state machine

Give concrete inclusive athlete-local range and activity caps, deterministic ordering/tie-breakers, and truncation detection (prefer fetching cap+1 so exactly-cap results are not falsely marked truncated). Define the exact status enum and conditions for each state, covering at least:

- no activities in range;
- no matching configured power zones;
- all activities missing/misaligned streams;
- some usable and some skipped activities;
- usable streams with skipped invalid intervals;
- activity-cap truncation;
- zero-power-only usable data.

Distinguish operational upstream failures from data insufficiency, and specify terse counts versus `include_full` activity IDs/reasons. “Partial” and “insufficient” cannot be left as prose because they are part of the MCP contract.

### 4. Concrete terse/full JSON contract

Provide a sample or field table with exact names and types. It should settle headline totals and zone-row fields (zone identity/name, bounds in watts, seconds, kJ, time/energy share), where status and insufficient reason live, and what appears in optional per-activity `series[]` rows. Terse output must still expose coverage/missing counts and the mechanical-work interpretation boundary; skipped IDs and audit detail may remain full-only. Raw samples must never appear in either shape.

Also pin the exact `_meta.method`, `source_tools` (`get_activities`, `get_activity_streams`, `get_athlete_profile` as actually used), formula ref/anchor, units for W/s/J/kJ, assumptions, zone-source metadata, and the visible statement that this is mechanical work—not metabolic energy, calorie expenditure, or food calories. The generic metric/imperial `_meta.units` injected by response shaping is not sufficient to communicate W and kJ.

### 5. Step artifacts and test checkpoint must remain coherent

The planned Step 1 edit to `internal/resources/testdata/analysis_formulas.md` without the corresponding `analysis_formulas.go`, resource assertions, and pinned SHA update would intentionally break `TestAnalysisFormulasMarkdownGolden`; either move the complete formula-resource change into Step 1 or defer the golden to Step 4. Pin the new formula ref and paragraph together with its interpretation boundary and citation.

Likewise, `go test ./internal/analysis -run 'ZoneEnergy'` currently passes with `[no tests to run]`. If Step 1 creates only contract types, call this a compile checkpoint or add contract tests in this step; do not present a vacuous filtered run as numerical verification.

### Cross-step note

The later PRD update must amend the current statement that `compute_activity_segment_stats` is the only analyzer that reads raw streams by default. This task deliberately introduces a second, full-toolset stream-backed analyzer, so leaving that sentence unchanged would make the authoritative product contract contradictory.
