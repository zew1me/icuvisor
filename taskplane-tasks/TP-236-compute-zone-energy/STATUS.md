# TP-236: Add deterministic power-zone energy analysis — Status

**Current Step:** Step 1: Lock the numerical and response contract
**Status:** 🟡 In Progress
**Last Updated:** 2026-07-10
**Review Level:** 2
**Review Counter:** 1
**Iteration:** 1
**Size:** L

---

### Step 0: Preflight

**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Stream, time-base, and zone semantics reviewed

---

### Step 1: Lock the numerical and response contract

**Status:** 🟨 In Progress

- [ ] Integration and boundary rules defined
- [ ] Range limits and statuses defined
- [ ] Terse/full response contract defined
- [ ] Mechanical versus metabolic boundary explicit
- [ ] R001 pure input, diagnostics, invalid-sample, and rounding contract locked
- [ ] R001 configured-zone validation, selection, and mixed-configuration policy locked
- [ ] R001 deterministic range cap and exact status state machine locked
- [ ] R001 concrete JSON/meta contract and coherent formula/test checkpoints locked

---

### Step 2: Implement and test the pure calculation

**Status:** ⬜ Not Started

- [ ] Timestamp-weighted zone integration implemented
- [ ] Invalid and missing samples reported explicitly
- [ ] Stable totals, shares, and rounding implemented
- [ ] Numerical boundary tests added

---

### Step 3: Add and register the MCP analyzer

**Status:** ⬜ Not Started

- [ ] Range, profile, and stream orchestration implemented
- [ ] Partial activity coverage reported
- [ ] Full-toolset registration and schemas added
- [ ] Catalog, safety, protocol, and tier coverage added

---

### Step 4: Formula resource, generated data, and docs

**Status:** ⬜ Not Started

- [ ] Formula resource updated
- [ ] PRD and roadmap aligned
- [ ] Generated website data and cookbook updated
- [ ] Changelog updated

---

### Step 5: Testing & Verification

**Status:** ⬜ Not Started

- [ ] FULL test suite passing
- [ ] Race suite passing
- [ ] Lint passing
- [ ] All failures fixed
- [ ] Build passes
- [ ] Generated docs clean

---

### Step 6: Documentation & Delivery

**Status:** ⬜ Not Started

- [ ] Must Update docs modified
- [ ] Check If Affected docs reviewed
- [ ] Discoveries logged

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|------|
| R001 | Plan | 1 | REVISE | `.reviews/R001-plan-step1.md` |

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| Tier-2 `taskplane-tasks/CONTEXT.md` is absent; required implementation paths and authoritative PRD/roadmap context are available. | Proceed using task packet and repository guidance. | Preflight |

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-07-10 | Task staged | PROMPT.md and STATUS.md created |
| 2026-07-10 16:11 | Task started | Runtime V2 lane-runner execution |
| 2026-07-10 16:11 | Step 0 started | Preflight |

## Blockers

*None*

## Notes

### Step 1 revised contract plan (R001)

- **Pure integration contract:** `ZoneEnergyInput` carries elapsed `TimestampsSeconds`, `PowerWatts`, and one validated `PowerZoneConfig`; `ComputeZoneEnergy` applies the left-endpoint rule `work_j += power[i] * (time[i+1]-time[i])`. The final sample contributes nothing without a following timestamp. A length mismatch rejects the whole activity (no aligned-prefix guess) while reporting `misaligned_samples = abs(len(power)-len(time))`. Intervals are evaluated in deterministic precedence: non-finite endpoint timestamp; duplicate timestamp; reversed timestamp; gap over 60 seconds; non-finite left power; negative left power. Invalid timestamps therefore invalidate each adjacent interval that uses them; invalid power invalidates only its own left interval. Zero watts is valid coasting time and is zoned, contributing zero joules. Diagnostics are exactly `input_samples`, `aligned_samples`, `usable_intervals`, `skipped_intervals`, `misaligned_samples`, `skipped_non_finite_timestamp`, `skipped_duplicate_timestamp`, `skipped_reversed_timestamp`, `skipped_large_gap`, `skipped_non_finite_power`, and `skipped_negative_power`.
- **Numerical presentation:** accumulate unrounded float64 seconds and joules, convert with `1 kJ = 1000 W*s`, then emit seconds/kJ to 3 decimals and shares to 4 decimals. Headline totals are sums of displayed zone totals; shares use displayed totals and deterministically adjust the last nonzero zone so nonzero shares sum to exactly `1.0000`. All-zero work emits `energy_share: 0` for every zone while retaining valid coasting seconds/time shares.
- **Configured zones:** boundaries are ordered lower bounds in watts and are never sorted. The first may be zero; all others must be finite, strictly increasing, and positive. Duplicate, descending, negative, or non-finite boundaries reject the configuration. Names pair by boundary index; missing/blank names become `Zone N`, and extras are ignored. Inclusion is `lower <= watts < upper`, final zone open-ended. If the first boundary is above zero, an explicit ordinal-0 `Below <first zone>` bucket `[0, first)` captures coasting/sub-boundary power. Sport settings are selected in activity `Type`, then `SubType`, order, matching either `SportSettings.Type` or any `Types` value case-insensitively. Distinct setting IDs/boundary/name vectors remain distinct configuration groups; rows are never silently merged by ordinal across configurations.
- **Bounded orchestration/state machine:** accept an inclusive athlete-local range of 1..366 days. Request 201 activity candidates, sort by local start (fallback UTC start) then activity ID, retain 200, and mark truncation only when candidate 201 exists. Optional sport filter is exact case-insensitive activity type/subtype. Operational profile/list/cancellation errors return tool errors. Per-activity stream fetch errors are data-coverage skips. Status is exactly `ok` (one or more usable activities, no skip/invalid/truncation), `partial` (usable data plus any skipped activity, skipped invalid interval, incompatible/missing config, stream-fetch failure, or truncation), or `insufficient` (no usable activity): reason `no_activities`, `missing_power_zones`, or `no_usable_power_streams`. Zero-power-only usable data is `ok` unless another partial condition applies.
- **Concrete response:** terse `result` always has `status`, optional `insufficient_reason`, `start_date`, `end_date`, optional `sport`, `activity_count`, `usable_activity_count`, `skipped_activity_count`, `invalid_interval_count`, `truncated_activity_candidates`, `total_seconds`, `total_kj`, `zones`, and `interpretation`. Each zone row has `zone_key`, `sport_setting_id`, `sport`, `zone`, `name`, `lower_watts`, optional `upper_watts`, `seconds`, `kj`, `time_share`, and `energy_share`. `interpretation` is fixed text stating this is power-derived mechanical work, not metabolic energy, calorie expenditure, or food calories. With `include_full`, `series[]` adds activity audit rows with `activity_id`, local `date`, `sport`, `status`, optional `reason`, `sport_setting_id`, totals, usable/skipped interval counts, and diagnostics; IDs/reasons are omitted from terse output. Raw samples never appear.
- **Pinned metadata:** `_meta.method = "left_endpoint_power_timestamp_integration"`; `_meta.source_tools = ["get_activities","get_activity_streams","get_athlete_profile"]`; `_meta.formula_ref = "icuvisor://analysis-formulas#power_zone_mechanical_work"`; `_meta.units = {"power":"W","time":"s","integration_work":"J","work":"kJ"}`; `_meta.assumptions` includes 60-second maximum interval, left-endpoint/final-sample rules, activity cap, and disclosed configured-zone sources; `_meta.boundaries` repeats the mechanical/metabolic distinction and no interpolation/raw-sample policy.
- **Step coherence:** Step 1 adds contract constants/types plus a real `TestZoneEnergyContract` compile/definition test and runs the filtered checkpoint. The formula ref constant, rendered paragraph, BIPM SI citation, golden resource, and resource tests move together in Step 4, avoiding intentional golden drift. The PRD update will replace the stale claim that only `compute_activity_segment_stats` reads streams.
