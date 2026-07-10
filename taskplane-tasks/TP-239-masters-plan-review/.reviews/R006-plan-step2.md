# Plan Review — Step 2

## Verdict: REVISE

1. **The validation/default-window contract is still incomplete.** R005 validates only the paired planned dates and name-only race. Specify that non-empty `race_date` is also trimmed and strict `YYYY-MM-DD`, rendered normalized, and has a short `UserError` for malformed input. Define how the 28-day history and 56-day baseline are resolved relative to default *and supplied* planned windows, including which windows are exclusive, so the default cannot violate the Step 1 non-overlap requirement. Add cases for malformed race dates, whitespace normalization, and `errors.As(err, *UserError)` with the exact public message.

2. **The planned tests do not yet cover all affected catalog contracts.** Alongside R005's registry/protocol/golden updates, add `MastersPlanReviewPrompt()` to `TestPromptResourceCitationsStayTerse`. Have the focused test assert the exact six-argument allowlist (and no credential, age-policy, write, or delete arguments), the deterministic analyzer/`icuvisor_list_advanced_capabilities` fallback route, and the portable pack's bounded/default scope contract—not only its registry link and tool names. This keeps the newly validating MCP prompt and the already-created portable pack from drifting.
