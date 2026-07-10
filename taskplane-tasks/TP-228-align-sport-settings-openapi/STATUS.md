# TP-228: Align sport-settings update and apply requests with live OpenAPI — Status

**Current Step:** Step 0: Preflight
**Status:** 🟡 In Progress
**Last Updated:** 2026-07-10
**Review Level:** 3
**Review Counter:** 0
**Iteration:** 1
**Size:** L

---

### Step 0: Preflight

**Status:** ✅ Complete

- [x] Required files and paths exist
- [x] Dependencies satisfied
- [x] Live public OpenAPI contract reconfirmed

---

### Step 1: Define the corrected write boundary

**Status:** ⬜ Not Started

- [ ] Exact update/apply request contract documented
- [ ] Public schema migration for effective_date decided
- [ ] Implicit historical apply removed from design
- [ ] Truthful response metadata defined

---

### Step 2: Align the intervals.icu client

**Status:** ⬜ Not Started

- [ ] Required recalcHrZones query implemented
- [ ] Apply PUT sends no date or semantic body
- [ ] Update no longer invokes apply
- [ ] Client contract tests pass

---

### Step 3: Align the MCP schema and response

**Status:** ⬜ Not Started

- [ ] recalc_hr_zones schema and forwarding implemented
- [ ] Unsupported effective-date behavior removed
- [ ] Response metadata corrected
- [ ] Schema snapshots and generated data updated

---

### Step 4: Regression coverage

**Status:** ⬜ Not Started

- [ ] Exact update/apply wire tests added
- [ ] No-implicit-apply regression added
- [ ] Legacy invalid argument behavior covered

---

### Step 5: Testing & Verification

**Status:** ⬜ Not Started

- [ ] FULL test suite passing
- [ ] Race suite passing
- [ ] Lint passing
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

## Discoveries

| Discovery | Disposition | Location |
|-----------|-------------|----------|

## Execution Log

| Timestamp | Action | Outcome |
|-----------|--------|---------|
| 2026-07-10 | Task staged | PROMPT.md and STATUS.md created |
| 2026-07-10 11:36 | Task started | Runtime V2 lane-runner execution |
| 2026-07-10 11:36 | Step 0 started | Preflight |

## Blockers

*None*

## Notes

*Reserved for execution notes*
