# Plan Review — TP-238 Step 1

## Verdict: REVISE

The revised contract addresses the prior mode, pagination, negative-intake, safety, and ownership findings. One source-route detail is still needed for it to be a genuinely closed, nutrition-only evidence contract:

- **Specify the wellness field projection and test it.** The plan says to call `get_wellness_data` when daily nutrition evidence is useful, but does not name its `fields` argument. With no `fields`, that tool requests the full wellness row and shapes unrelated sensitive fields (for example injury, menstrual, glucose, readiness, and comments) in addition to nutrition. Require the nutrition-only upstream projection—`kcalConsumed`, `carbohydrates`, `protein`, and `fatTotal`—with only explicitly user-requested custom codes appended; make clear that their returned aliases remain `calories_intake`, `carbs_g`, `protein_g`, and `fat_g`. If nutrition freshness/provenance is unavailable under that projection, report it as unavailable rather than broadening the read to health fields. Add these exact `fields` rules (and no unrelated wellness interpretation) to the portable pack and `TestFuelingReviewPortablePackContract`.

This closes the gap between the stated closed vocabulary/read-only health boundary and the actual `get_wellness_data` request semantics.
