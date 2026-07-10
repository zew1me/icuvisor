# Plan Review — TP-236 Step 5

## Verdict: REVISE

The test, race, lint, and build commands match the required verification surface, but the generated-docs check does not prove that generation is reproducible. `git diff --check` only detects whitespace errors; it still exits successfully when `make docs-tools` changes `web/data/tools.json` or `web/data/tool_schemas.json`. That allows stale committed generated data to pass the stated “Generated docs clean” checkpoint.

Revise the plan to fail on generator-induced changes to both generated files, while retaining the repository-wide whitespace check. Because prior implementation steps are committed, a suitable command is:

```sh
make docs-tools && \
  git diff --exit-code -- web/data/tools.json web/data/tool_schemas.json && \
  git diff --check
```

If Step 5 begins with intentional uncommitted changes to those files, capture their pre-generation state and compare after generation instead. Record each command and its exit status in `STATUS.md`, and rerun any affected verification command after fixes.
