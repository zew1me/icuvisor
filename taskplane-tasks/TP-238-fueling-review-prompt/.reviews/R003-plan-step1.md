# Plan Review — TP-238 Step 1

## Verdict: REVISE

The revised contract now makes the modes, bounded date resolution, route/pagination coverage, source separation, safety boundaries, ownership, and Step 1 contract test concrete. One calculation-validity rule remains necessary for the stated “valid logged grams” criterion:

- Define `carbs_ingested_g` eligibility as a returned **non-negative** numeric value, not merely any numeric value. A logged `0` remains valid and must yield `0 g/h` when moving time is positive, but a negative upstream/custom value is invalid evidence: do not calculate or aggregate it, label/count it as an invalid intake value, and include it among the stated range exclusions. Add this case to `TestFuelingReviewPortablePackContract` alongside the absent-value, zero-value, and invalid/non-positive-duration clauses.

Without this distinction, the current “returned numeric ingested value” rule permits a nonsensical negative grams-per-hour result while claiming the calculation uses valid logged grams.
