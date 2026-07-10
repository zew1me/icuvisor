# Plan Review — Step 2

## Verdict: REVISE

1. **The paired planned-window contract lacks the tests that prove it.** R006-1 requires strict paired `planned_start`/`planned_end`, but R006-2 lists malformed, reversed, and overlong dates—not start-only and end-only requests. Add table cases for each one-sided input that assert the exact `UserError`; otherwise the handler could silently default one endpoint. Also state whether a same-day (one-day inclusive) window is valid, and cover that decision plus the accepted 90-day boundary.

2. **The portable pack needs an explicit update action, not only coverage.** The current `docs/prompts/client-prompt-packs/masters-plan-review.md` contains none of the handler's defaults, lookback ranges, paired-date requirement, maximum planned window, or race-name/date dependency. R006-3 says to test a bounded/default scope contract, but does not explicitly schedule adding that contract to the pack. Add the pack modification to Step 2 and specify that it mirrors the handler's normalized scope/defaults and validation limits, with focused assertions for those terms so the two cannot drift.
