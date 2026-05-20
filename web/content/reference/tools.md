---
title: "Tool reference"
description: "Generated MCP tool catalog for icuvisor."
---

This reference is generated from the MCP tool registry. It shows the tools currently registered by icuvisor, grouped by domain, with their toolset tier and safety gate.

## NOTE calendar events

Use [`add_or_update_event`](#add_or_update_event) with `category: "NOTE"` to create free-text calendar notes. NOTE is an event category handled by the event writer, not a separate `add_note` tool. For new NOTE events, include a non-empty `name`; put the note body in `description`. Do not use `workout_doc` for NOTE descriptions because `workout_doc` is only for structured workouts.

Common NOTE use cases:

```json
{
  "date": "2026-06-12",
  "category": "NOTE",
  "name": "Race-week nutrition plan",
  "description": "Breakfast: oatmeal and banana. Lunch: rice bowl. Carry 90 g carbs/hour for the long ride."
}
```

```json
{
  "date": "2026-06-15",
  "category": "NOTE",
  "name": "Travel logistics",
  "description": "Flight lands at 14:20. Pack pedals, charger, spare cleats, bottles, and race license."
}
```

```json
{
  "date": "2026-06-17",
  "category": "NOTE",
  "name": "Daily reminder",
  "description": "Take resting HR after waking, do 10 minutes mobility, and log sleep quality before training."
}
```

```json
{
  "date": "2026-06-18",
  "category": "NOTE",
  "name": "Coach annotation",
  "description": "Athlete reported tight calves; keep Thursday aerobic and reassess before intensity."
}
```

## Activity interval source metadata

`get_activity_intervals` includes additive response metadata to help clients distinguish structured workout segments from generic device laps:

- `_meta.interval_source`: `structured_workout`, `device_laps`, or `unknown`.
- `_meta.auto_lap_suspected`: `true` when generic near-uniform 1 km / 1 mi (or supported duration) rows look like device auto-laps.

When auto-laps are suspected, analyzer-style clients should avoid claiming the athlete hit or missed individual structured workout steps from those rows; they are device splits, not necessarily planned workout segments.

{{< tool-catalog >}}
