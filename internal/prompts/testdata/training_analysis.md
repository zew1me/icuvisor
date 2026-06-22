Prompt: Training analysis
Scope: start_date=2026-04-01, end_date=2026-04-30.
Resources: icuvisor://athlete-profile.
Tools: get_athlete_profile, get_fitness, get_training_summary, get_best_efforts, get_activities.
Do:
- Read profile first for timezone, sport settings, and units.
- Use fitness and summary rows for CTL/ATL/TSB, ramp, volume, load, and intensity mix.
- Use best efforts and recent activities only for context; keep raw rows terse unless the user asks for detail.
- If the user explicitly mentions hypoxic training, altitude tents/chambers, or reduced oxygen exposure, state that CTL/ATL/Form use logged training_load: power-based load may under-represent extra hypoxic strain, HR/RPE/feel/recovery can be supporting context, and you must not apply a hypoxia multiplier without evidence.
Guardrails:
- Do not request or accept intervals.icu API keys in chat.
- Prefer terse default tool responses; use include_full only when the user asks or evidence is missing.
Return: load/trend readout with notable changes, likely drivers, missing-data caveats, and 2-3 next-step questions or actions.
