# Plan Review — TP-238 Step 2

## Verdict: REVISE

The approved evidence contract makes the prompt implementation largely executable, but two registration/validation cases need to be added to this step's plan:

1. **Cover every registry consumer, not just `internal/prompts`.** Update `catalog_test.go`'s count/order, golden table, portable-pack linkage table, and terse-resource list for `fueling_review`. Also add `internal/mcp/protocol_test.go` to Step 2 scope and change its `prompts/list` expectation from 11 to 12 with the sorted `fueling_review` name (and retrieve the new prompt or otherwise exercise its MCP registration). The current focused command cannot detect that stale protocol expectation; the full suite will fail after registration otherwise.

2. **Test name-only race context at the actual handler.** The contract says `race_name` without `race_date` must ask for a date rather than trigger an unbounded event scan, but the enumerated handler cases omit it. Make it a short pre-render user error such as `missing race_date; provide YYYY-MM-DD`, and add it to the table-driven `FuelingReviewPrompt` handler tests alongside malformed `race_date`. Keep the rendered prompt/golden explicit that a valid race lookup uses same-day `get_events` bounds and never scans by name alone.
