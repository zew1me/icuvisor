---
title: "Terse by default"
description: "Why icuvisor keeps tool responses small unless you ask for full detail."
---

AI assistants have limited context windows. If every tool call returns raw streams, null-heavy wellness rows, and debug metadata, the assistant spends its attention on plumbing instead of coaching decisions.

icuvisor is terse by default so routine questions fit comfortably in a conversation:

- Read tools return the smallest useful shape first.
- Null values are stripped from terse responses, while meaningful zero, empty string, and `false` values remain.
- Response metadata explains scales and units when the assistant needs them.
- Activity custom fields are opt-in by field code on activity reads and analyzers, so default activity rows stay small. For example, request `custom_fields: ["vo2max_est"]` and use `custom:vo2max_est` in `analyze_correlation` when you need that history.
- Small scalar interval custom fields that are already part of interval rows stay in terse mode when they are safe to name directly. For example, `get_activity_intervals` returns manually-entered interval fields such as lactate under each row's `custom_fields` map when intervals.icu includes them.
- Heavy payloads, such as raw samples and full raw upstream objects, require an explicit `include_full: true` argument on tools that support it.
- Workout-library reads are shaped to find the right folder first and inspect folder-scoped examples instead of sending every template into the conversation.

## What `include_full` means

`include_full` is an opt-in for detail. Use it when you are debugging, when the user explicitly asks for raw fields, or when the terse response is missing evidence needed for the answer.

It is not a "better answer" switch. Full payloads can be much larger and can make weaker or smaller-context models more likely to lose the thread. Because icuvisor runs locally against your intervals.icu account, it can fetch the next precise slice on demand instead of preloading broad libraries or raw streams before the assistant knows which details matter.

## Toolset tiers are part of the same idea

icuvisor also keeps the default tool catalog small. The `core` tier covers daily-use tools, `compact` narrows the catalog further for smaller/local models, and `full` adds advanced or heavier workflows. This helps the AI client pick the right tool and reduces per-conversation tool-description load without hiding data permanently.

The exact `ICUVISOR_TOOLSET` values live in [Safety modes and toolset tiers]({{< relref "../reference/safety-modes#toolset-tier" >}}). Before enabling `full` just to browse, ask the assistant to call [`icuvisor_list_advanced_capabilities`]({{< relref "../reference/tools#icuvisor_list_advanced_capabilities" >}}). It lists hidden advanced capabilities and explains whether the next step is to keep using `core`, switch from `compact` to `core`, or opt in to `full`.

## The goal

Terse-by-default does not hide data from you. It makes the common path fast, predictable, and token-efficient, while still allowing full detail when a prompt genuinely needs it.
