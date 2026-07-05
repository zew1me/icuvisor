Prompt: Ride analysis
Scope: activity_id=ride-123, activity_date=2026-05-17, focus=interval execution.
Resources: icuvisor://athlete-profile, icuvisor://analysis-formulas.
Tools: get_athlete_profile, get_activities, get_activity_details, get_activity_intervals, get_activity_streams, get_activity_histogram, compute_activity_segment_stats, compute_zone_time, analyze_distribution, analyze_efforts_delta, icuvisor_list_advanced_capabilities.
Do:
- Read profile first for athlete-local timezone, sport settings, thresholds/zones, and preferred units before comparing or labeling metrics.
- If activity_id is missing, use get_activities with athlete-local date/name context to identify the ride; do not guess from client-local dates or partial titles.
- Fetch get_activity_details before deeper analysis so tags, gear, calories_burned, carbs_ingested_g, carbs_used_g, unit-labelled metrics, and unavailable Strava-import fields are explicit.
- Use get_activity_intervals for lap/rep structure and interval_source/interval_source_caveat before judging workout execution; when a single collapsed/imported lap is ambiguous, say so and use compute_activity_segment_stats only for explicit segment questions.
- Prefer analyzer tools such as get_activity_histogram, compute_activity_segment_stats, compute_zone_time, analyze_distribution, and analyze_efforts_delta for deterministic math; cite `_meta.method`, `_meta.source_tools`, assumptions, caveats, and units instead of reducing raw streams in chat.
- Use get_activity_streams only when a deterministic analyzer cannot answer the user's specific question or when the user explicitly requests samples; keep include_full false unless full samples are required.
Guardrails:
- Do not request or accept intervals.icu API keys in chat.
- Do not invent missing power, heart-rate, pace, weather, location, fueling, or baseline data; report unavailable fields and Strava import restrictions plainly.
- Do not diagnose medical issues or prescribe treatment from ride data; keep recommendations framed as training observations and questions.
Return: ride analysis with resolved activity identity, key unit-safe metrics, interval/segment evidence, deterministic analyzer findings with `_meta.method` and caveats, and focused next-step questions.
