# Plan Review — TP-236 Step 1

## Verdict: REVISE

The R002 revision closes the stream-error, grouping, audit-enum, and pure mismatch contracts, but a few contract points still cannot be implemented deterministically as written.

### 1. The proposed `_meta.units` path still cannot preserve the promised units

A tool-specific payload/encoder “following `get_activity_histogram`” is not enough. `response.Shape` calls `addScaleMeta`, which removes an existing `_meta.units`, and then `addCommonMeta`, which treats `units` as response-owned and replaces it with the preferred metric/imperial metadata (`internal/response/meta.go`). Consequently, `zoneEnergyMeta.Units = {power,time,integration_work,work}` will not survive the stated encoder path.

Choose a viable contract and implementation path now: extend response shaping to preserve a tool-specific analysis-units object (and add the affected response files/tests to scope), use a non-conflicting exact key such as `_meta.analysis_units`, or define a genuinely custom shaping path that retains common catalog metadata without dropping these units. The Step 3 schema/tests must assert the final serialized JSON, not merely the pre-shaping Go struct.

### 2. Embedding `analysis.AnalyzerMeta` exposes fields whose semantics remain undefined

The selected metadata type necessarily also emits `n`, `missing_days`, `missing_action`, and `insufficient_sample`. The plan does not define whether `n` means usable activities, usable intervals, or samples; how `missing_days` is computed for a range with rest days and/or skipped activities; or how `insufficient_sample` relates to the tool's `insufficient` status. Lock those values (and the minimum-sample input), or do not embed the shared type. Also pin the exact `_meta.boundaries` strings rather than saying they “repeat” several policies.

### 3. Complete the insufficient-status decision table

The three insufficient reasons are named but their precedence is not. A no-usable-data window can contain a mixture of `no_matching_power_zone_config`, `invalid_power_zone_config`, missing streams, and unusable intervals, and different implementations can currently select either `missing_power_zones` or `no_usable_power_streams`. Define the exact reduction/priority, including:

- whether zero configured boundaries are “missing” or an invalid configuration;
- whether `no_activities` means no retained candidates or no retained candidates after the sport filter;
- mixed missing/invalid-zone and stream-failure skip reasons when no activity is usable.

### 4. Pin the remaining model-visible fixed text

`interpretation` is described as fixed text, but no exact value is supplied, and the assumptions only give exact values for some metadata. Since this step exists to lock the response contract and golden/resource text will rely on the same interpretation boundary, state the literal `interpretation` value and exact ordered `_meta.boundaries` values. This avoids schema/golden tests blessing wording invented during Step 3 or Step 4.

The numerical integration, diagnostic-counter equations, mixed-zone fingerprint/order/share rules, operational error classification, and formula-resource deferral are otherwise sufficiently concrete.
