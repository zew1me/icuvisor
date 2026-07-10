# Plan Review — TP-236 Step 1

## Verdict: REVISE

The revision resolves most numerical choices, but several model-visible and failure contracts are still left for implementation to invent.

### 1. Do not turn operational stream failures into data-coverage skips

The proposed rule that every per-activity stream-fetch error is a coverage skip does not satisfy the requested operational/data-insufficiency distinction. The intervals client exposes stable `ErrUnauthorized`, `ErrRateLimited`, `ErrUpstream`, and context errors; treating a credentials failure or exhausted 429/5xx response as `partial`/`insufficient` would misleadingly report missing athlete data and could repeat the same failing request up to 200 times.

Lock an exact classification. At minimum, cancellation/deadline, unauthorized, rate-limit, transient upstream/transport, and malformed-response failures should abort with a short tool error. Only explicitly recognized activity-level absence/restriction cases should become a skipped audit row, with their reason enums stated. Add tests for both classes to the Step 3 plan.

### 2. The promised `_meta.units` cannot be emitted by the stated analyzer path

`analysis.AnalyzerMeta` has no `units` field, and `encodeAnalyzerResponse` always constructs that fixed type from `AnalyzerMetaInput`. The plan promises an exact `_meta.units` object but does not say whether Step 3 will:

- extend the shared analyzer metadata contract (which broadens file/test/schema scope), or
- use a tool-specific metadata type/encoder, as `get_activity_histogram` does.

Choose and name the implementation path now. Also replace “assumptions includes … disclosed configured-zone sources” with exact keys and JSON types. The configured-zone source is especially important when multiple settings produce separate groups; it cannot remain unspecified prose after this checkpoint.

### 3. Mixed-zone grouping and aggregate response semantics remain underspecified

The plan correctly refuses to merge distinct configurations, but it does not define enough to reproduce the output deterministically. Lock:

- how `zone_key` is derived and how rows are ordered;
- whether row `sport` is the activity type, matched setting `Type`, or another canonical value;
- whether `time_share`/`energy_share` denominators span the whole response or one configuration group, and which row receives the rounding remainder across groups;
- what makes two configurations identical when setting IDs are absent/zero or duplicated;
- whether `activity_count` is fetched, sport-matched, retained-after-cap, or usable-plus-skipped, and how all three counts reconcile when candidate 201 triggers truncation;
- the closed per-activity `status` and `reason` enums in `series[]`.

Without these rules, two conforming implementations can return different zone identities, shares, counts, and audit rows for the same athlete window.

### 4. Finish the pure function’s rejection/result contract

`ZoneEnergyInput` is described as carrying a “validated” config while the same plan says invalid configurations are rejected. State where validation occurs and whether `ComputeZoneEnergy` returns `(result, error)` or a diagnostic result with an unusable status. For stream mismatch/short input, define `input_samples`, `aligned_samples`, `skipped_intervals`, and per-reason counter values; `input_samples` is currently ambiguous because there are two stream lengths. The Step 1 contract test should assert these definitions rather than only compile the types.

The formula-resource deferral to Step 4 and the planned correction of the stale PRD raw-stream statement are coherent.
