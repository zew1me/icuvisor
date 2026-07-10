# TP-238: Add grounded fueling review prompt pack — Status

**Current Step:** Step 1: Define the fueling evidence contract
**Status:** 🟡 In Progress
**Last Updated:** 2026-07-10
**Review Level:** 1
**Review Counter:** 1
**Iteration:** 1
**Size:** M

---

### Step 0: Preflight

**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Fueling and nutrition field semantics reviewed

---

### Step 1: Define the fueling evidence contract

**Status:** 🟨 In Progress

- [ ] Define bounded activity-ID and athlete-local date-range modes, source-tool route, pagination, and optional race context
- [ ] Define closed nutrition vocabulary, source-labelled return layout, and missing/freshness/availability reporting
- [ ] Define grams-per-hour denominator, valid-row eligibility, range aggregation, and coverage/exclusion rules
- [ ] Define read-only health, product, target, and custom-field boundaries
- [ ] Resolve Step 1/2 ownership for contract verification, prompt function/registry, golden fixture, and portable-pack discoverability

---

### Step 2: Register the prompt and add regression coverage

**Status:** ⬜ Not Started

- [ ] Prompt implemented and registered
- [ ] Date, unit, and missing-data discipline encoded
- [ ] Focused tests and golden fixture added
- [ ] Read-only and no-product-library behavior covered

---

### Step 3: Publish cookbook, portable pack, and evals

**Status:** ⬜ Not Started

- [ ] Cookbook and client pack added
- [ ] Prompt library aligned
- [ ] Eval scenarios added
- [ ] References, PRD, and changelog updated

---

### Step 4: Testing & Verification

**Status:** ⬜ Not Started

- [ ] FULL test suite passing
- [ ] Prompt eval validation passing
- [ ] Lint passing
- [ ] All failures fixed
- [ ] Build passes
- [ ] Markdown and diff clean

---

### Step 5: Documentation & Delivery

**Status:** ⬜ Not Started

- [ ] Must Update docs modified
- [ ] Check If Affected docs reviewed
- [ ] Discoveries logged

---

## Reviews

| # | Type | Step | Verdict | File |
|---|------|------|---------|
| R001 | Plan | 1 | REVISE | `.reviews/R001-plan-step1.md` |

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|
| Referenced area-context file is absent; all task implementation paths are present | Proceeded using the listed Tier 3 source/docs instead | `taskplane-tasks/CONTEXT.md` |

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-07-10 | Task staged | PROMPT.md and STATUS.md created |
| 2026-07-10 18:24 | Task started | Runtime V2 lane-runner execution |
| 2026-07-10 18:24 | Step 0 started | Preflight |

## Blockers

*None*

## Notes

### Step 1 revised evidence-contract plan

- **Modes and source route:** Activity-ID mode uses `get_athlete_profile`, then `get_activity_details` for the selected activity. Date-range mode defaults to the last 14 completed athlete-local days, validates a supplied bounded window (1–90 days), uses `resolve_calendar_dates` for relative dates, then `get_activities` in terse pages as the index/source for duration, load, and activity carbs; it uses `get_activity_details` only for a selected session needing more context, never for every row. Both modes call `get_wellness_data` only when daily nutrition evidence is requested or useful, `get_training_summary` only for aggregate load context, and `get_events` only when a supplied `race_date` or `race_name` requests calendar context. The prompt must fetch all pages needed before saying a range is complete; otherwise it reports count/window as partial and omits the opaque token.
- **Evidence and output:** `carbs_ingested_g` is a numeric athlete-logged during-activity intake; an absent key is a missing log and a numeric zero remains logged zero. `carbs_used_g` is an upstream used/burned estimate, never intake or an intake substitute. Wellness `carbs_g`, `calories_intake`, `protein_g`, and `fat_g` are daily dietary fields, separate from activity records and not summed/subtracted with them. `calories_burned` and training load are context only. Requested custom fields retain their exact code and unknown meaning. Return separate sections for sourced activity evidence, daily wellness evidence, race/calendar context, labelled calculations, coverage/data gaps, and separately labelled non-personalized general guidance.
- **Calculation:** Use `moving_time_seconds` only. Per-session `logged carbs/hour = carbs_ingested_g / (moving_time_seconds / 3600)` with units `g/h`; calculate only for a returned numeric ingested value and positive moving time, never unavailable/Strava-blocked rows. A range rate, if shown, is the sum of valid logged ingested grams divided by the sum of the same valid moving durations and states eligible/total sessions and every exclusion. It never uses carbs used, calories, load, wellness totals, or targets as either operand.
- **Dates and availability:** Anchor windows to athlete-local dates, preserve each activity timezone, label current-day `_meta.as_of` evidence partial, and surface stale wellness, provenance/field-semantics warnings, missing fields, unavailable/Strava-blocked rows, and missing/invalid durations. Missing is neither zero nor inadequate fueling. An absent requested race is reported as unconfirmed calendar context, never invented.
- **Boundaries and ownership:** The prompt is read-only: no write/delete tools, `include_full`, streams, or raw payloads. It does not diagnose, assess eating disorders, prescribe individual nutrition, calculate/recommend carbohydrate/calorie/sodium/fluid/sweat targets, infer deficits, claim product/performance effects, or invent food/product libraries. General material is visibly educational/conditional and refers individual or medical requests to a qualified sports dietitian/clinician. Step 1 owns the contract and portable pack draft plus a contract-focused test; Step 2 owns the prompt function/registry and deterministic golden fixture; `docs/prompts/README.md` will list the new portable pack.
| 2026-07-10 18:29 | Review R001 | plan Step 1: REVISE |
