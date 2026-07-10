---
title: "Fueling review"
description: "Review logged activity and daily nutrition evidence without estimating intake or prescribing targets."
weight: 45
---

This recipe reviews **what you logged**. It does not turn load, calories, or
an upstream estimate into a personal fueling prescription. Every fact names its
source; any general education is separate from the evidence.

## When to use this

- To see which recent sessions have usable logged carbohydrate-per-hour evidence.
- To identify activity records whose intake or duration is missing or invalid.
- To place a confirmed same-day race calendar entry next to the review without
  inventing race details or nutrition targets.

## Recent sessions or a date range

Use this copyable workflow for one activity or a bounded review window. The
default is the last 14 **completed athlete-local** days.

```text
Review my logged fueling evidence using icuvisor and my intervals.icu data.

Start by calling get_athlete_profile for my athlete-local timezone and units.
For a relative or default window, call resolve_calendar_dates with offsets -14
and -1 and use its returned dates. For a date-range review, call terse
get_activities with include_unnamed: true and fetch every page needed before
calling the sample complete. Do not expose next_page_token. If the range is
partial, say exactly which athlete-local window, pages, and rows are covered.
Count unnamed, unavailable, and Strava-blocked rows separately. For one known
activity ID, use terse get_activity_details once; do not fetch details for every
row in a range.

Use moving_time_seconds as the only duration denominator. Calculate logged
carbs/hour only as carbs_ingested_g / (moving_time_seconds / 3600), label every
result g/h, and only when carbs_ingested_g is a returned non-negative numeric
value and moving_time_seconds is positive. A logged zero is valid and is 0 g/h.
An absent intake log, a negative intake value, a missing/zero/non-positive
duration, or an unavailable row gets no rate and is counted as an exclusion.
For a range total, sum only eligible logged grams and those same eligible moving
durations; state eligible sessions / total sessions and each exclusion category.

Keep carbs_ingested_g (athlete-logged during-activity intake), carbs_used_g
(upstream used/burned estimate), calories_burned, training load, and daily
wellness nutrition separate. Never use carbs_used_g, calories, load, daily
wellness totals, or an invented target to fill a missing intake value or either
side of the calculation. If daily nutrition is useful, call get_wellness_data
only with kcalConsumed, carbohydrates, protein, and fatTotal plus any explicit
custom code; label returned calories_intake, carbs_g, protein_g, and fat_g as
daily fields, not session intake. Treat a custom field as unknown unless I give
its exact meaning.

Return sections in this order: Sourced activity evidence; Sourced daily-wellness
evidence; Labelled calculations; Coverage and data gaps; General educational
guidance. Give each fact its tool and athlete-local date/window. Surface
_meta.as_of partial-day context, _meta.stale, _meta.missing_fields,
field-semantics/provenance, and availability warnings. Missing is not zero or
evidence of inadequate fueling.

Do not call write/delete tools, get_activity_streams, include_full, or request
raw payloads. Do not diagnose, estimate intake, label a session underfueled, infer
a deficit, prescribe nutrition, recommend carbohydrate/calorie/sodium/fluid/
sweat-rate targets, claim a product effect, or invent a product library. Keep any
general material conditional and educational; for individualized or medical
nutrition questions, suggest a qualified sports dietitian or clinician.
```

## Optional race or session-planning context

A calendar entry can give context, but it does not confirm a nutrition plan or
supply a target. Ask for a date before looking up an event by name.

```text
For the fueling review, also check my calendar context for race_date
[YYYY-MM-DD] and optional race_name [name]. Use get_events with oldest and
newest both equal to that athlete-local race_date and limit: 100. Race_name only
disambiguates entries on that date; never scan the calendar by name alone. If
_meta.truncated is true, call the calendar evidence partial. If a complete
same-day lookup has no matching event, say there is no confirmed calendar event
rather than inventing race context.

Render Sourced race/calendar context separately from the logged activity and
daily-wellness evidence. You may offer clearly labelled conditional general
education or questions for a qualified sports dietitian/clinician, but do not
recommend individual carbohydrate, calorie, sodium, fluid, or sweat-rate targets,
claim product effects, or write anything to my calendar.
```

## What a careful answer says

A careful answer can say that three of eight covered sessions had eligible
logged g/h calculations, two had missing intake logs, one had an invalid duration,
and two were unavailable. It cannot say that the missing sessions were poorly
fueled, use `carbs_used_g` as ingested carbohydrate, or combine a wellness day's
macros with a session's load to estimate a deficit.

See the downloadable [Fueling review prompt pack](https://github.com/ricardocabral/icuvisor/blob/main/docs/prompts/client-prompt-packs/fueling-review.md)
for client-mode instructions, or invoke the `fueling_review` MCP prompt when
your client exposes prompts.
