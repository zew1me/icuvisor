# Plan Review — TP-238 Step 1

## Verdict: REVISE

The evidence, calculation, source-separation, and safety contract is otherwise implementation-ready. One executable validation/testing detail remains open:

- **Make the prompt-date contract date-only and test the actual handler.** The plan currently permits generic “ISO endpoints,” but `get_wellness_data`, `get_training_summary`, and `get_events` all require athlete-local `YYYY-MM-DD` values (whereas `get_activities` also accepts date-times). Lock `start_date`, `end_date`, and `race_date` to strict athlete-local `YYYY-MM-DD` dates before rendering, including the inclusive 1–90-day check. In Step 2, add table-driven handler tests—not only the Step 1 portable-pack text assertions—for the default range, valid activity/range modes, activity/date conflict, one-sided range, malformed date (including date-time), reversed and over-90-day range, and malformed `race_date`. This makes the claimed pre-render rejection behavior enforceable and ensures all downstream read calls receive compatible arguments.
