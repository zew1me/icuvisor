# Task: TP-232 - Correct athlete-ID normalization and hosted HTTP guidance

**Created:** 2026-07-10
**Size:** S

## Review Level: 1 (Plan Only)

**Assessment:** This is a small documentation correction, but both statements affect athlete routing or whether users choose a secure hosted endpoint. A lightweight plan review should confirm wording against current code and product behavior.
**Score:** 2/8 — Blast radius: 1, Pattern novelty: 0, Security: 1, Reversibility: 0

## Canonical Task Folder

```
taskplane-tasks/TP-232-correct-docs-guidance/
├── PROMPT.md
├── STATUS.md
├── .reviews/
└── .DONE
```

## Mission

Correct two misleading public documentation claims. Bare-numeric intervals.icu athlete IDs must remain bare and must never be rewritten with an `i` prefix. The Streamable HTTP troubleshooting guide must direct remote-only clients to the already-available hosted HTTPS connector rather than saying hosted support is future work. Add a small automated content contract so these trust- and routing-sensitive statements do not regress.

## Dependencies

- **None**

## Context to Read First

**Tier 2 (area context):**

- `taskplane-tasks/CONTEXT.md`

**Tier 3 (load only if needed):**

- `internal/config/athlete.go` — authoritative normalization behavior
- `web/content/explain/local-first.md` — current local/hosted trust model

## Environment

- **Workspace:** repository root and `web/`
- **Services required:** None

## File Scope

- `web/content/reference/config-file.md`
- `web/content/guides/http-transport.md`
- `scripts/tests/test_docs_guidance.py`

## Steps

### Step 0: Preflight

- [ ] Required files and paths exist
- [ ] Dependencies satisfied
- [ ] Confirm the hosted connector URL and current athlete-ID code behavior

### Step 1: Correct and lock the guidance

**Plan-review checkpoint**

- [ ] Replace the claim that numeric IDs normalize to `i12345` with exact trim/lowercase/validation behavior that never adds or removes the prefix
- [ ] Replace the future-hosted-relay HTTP troubleshooting text with a direct hosted-mode link and retain the warning against exposing local loopback through a generic tunnel
- [ ] Add an automated content test that fails on the old false claims and requires the corrected hosted/ID guidance
- [ ] Run targeted tests: `python3 scripts/tests/test_docs_guidance.py`

**Artifacts:**

- `web/content/reference/config-file.md` (modified)
- `web/content/guides/http-transport.md` (modified)
- `scripts/tests/test_docs_guidance.py` (new)

### Step 2: Testing & Verification

- [ ] Run FULL test suite: `make test`
- [ ] Run documentation content test: `python3 scripts/tests/test_docs_guidance.py`
- [ ] Build website: `make web-build`
- [ ] Fix all failures
- [ ] Build passes: `make build`
- [ ] Verify clean Markdown/diff: `git diff --check`

### Step 3: Documentation & Delivery

- [ ] "Must Update" docs modified
- [ ] "Check If Affected" docs reviewed
- [ ] Discoveries logged in STATUS.md

## Documentation Requirements

**Must Update:**

- `web/content/reference/config-file.md` — document exact athlete-ID normalization
- `web/content/guides/http-transport.md` — point remote-only clients to hosted mode

**Check If Affected:**

- `web/content/guides/api-key.md` — verify it already preserves both athlete-ID shapes
- `web/content/connect/chatgpt.md` and `web/content/connect/claude-ai.md` — verify hosted URL and security wording agree

## Completion Criteria

- [ ] No public doc claims a bare ID gains an `i` prefix
- [ ] HTTP troubleshooting acknowledges the current hosted connector
- [ ] Generic tunnel warning remains explicit
- [ ] New documentation contract test exists and passes
- [ ] Full tests, website build, and binary build pass

## Git Commit Convention

Commits happen at step boundaries. All commits MUST include TP-232:

- **Step completion:** `docs(TP-232): complete Step N — description`
- **Bug fixes:** `docs(TP-232): description`
- **Tests:** `test(TP-232): description`
- **Hydration:** `hydrate: TP-232 expand Step N checkboxes`

## Do NOT

- Change athlete-ID code behavior
- Add or remove an athlete-ID prefix in examples that represent bare-numeric accounts
- Recommend ngrok, cloudflared, or another unauthenticated tunnel for local HTTP
- Change hosted authentication or deployment code
- Expand into persistent-service recipes; TP-234 owns that work
- Commit without TP-232 in the message

---

## Amendments (Added During Execution)

<!-- Workers add amendments here if prerequisites or instructions are contradictory. -->
