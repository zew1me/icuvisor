---
title: "Coaching conversation handoff"
description: "Carry durable coaching context into a fresh chat without mixing user decisions with live training evidence."
weight: 85
---

Long coaching chats eventually become difficult to navigate and may retain stale assumptions. This read-only recipe creates a compact Markdown handoff that you review before manually copying it into a fresh Claude, ChatGPT, Cursor, or other client conversation.

## When to use this

- A coaching conversation is long, slow, or approaching a context limit.
- You want to change AI clients without losing durable goals and decisions.
- Training evidence in the current chat may be stale and should be refreshed.

## Start the handoff

If your client exposes [MCP prompts]({{< relref "/reference/resources-prompts" >}}), invoke `coaching_handoff`. You may optionally set `lookback_days` (default 28, range 1-90) and `race_context_days` (default 90, range 1-365).

If the client does not expose MCP prompts, copy the canonical [coaching handoff prompt pack](https://github.com/ricardocabral/icuvisor/blob/main/docs/prompts/client-prompt-packs/coaching-handoff.md) into the current chat. Both entry points use the same contract and read-only tool route.

## The six-section contract

The handoff must contain these sections in order:

1. **Handoff scope** — generated-on date, timezone, and covered windows, all anchored to the athlete-local calendar.
2. **Conversation-stated context** — separate **Goals**, **Constraints**, and **Accepted decisions** lists. A decision belongs here only when you explicitly stated or accepted it. Assistant suggestions, summaries, and calendar rows are not user decisions.
3. **Icuvisor evidence** — compact rows with `Claim | Source tool | Athlete-local evidence date/window | Freshness/as-of`.
4. **Current plan state** — only calendar and training-plan facts retrieved through Icuvisor.
5. **Data gaps and unresolved questions** — missing, stale, partial, unavailable, or failed reads and anything the next chat still needs to resolve.
6. **Next actions** — concise follow-ups for the fresh conversation.

A record date says when evidence applies; freshness says how current the source is. Preserve a returned `as_of` or provider freshness marker. If none is returned, write `not provided` rather than inventing a retrieval timestamp.

## What icuvisor reads

| Need | Read-only route |
| --- | --- |
| Timezone and date anchors | [`get_athlete_profile`]({{< relref "/reference/tools#get_athlete_profile" >}}), then [`resolve_calendar_dates`]({{< relref "/reference/tools#resolve_calendar_dates" >}}) |
| Race and plan state | Terse [`get_events`]({{< relref "/reference/tools#get_events" >}}) and [`get_training_plan`]({{< relref "/reference/tools#get_training_plan" >}}) |
| Durable recent evidence | Terse `get_fitness`, `get_training_summary`, `get_activities`, and `get_wellness_data` only as needed |
| Missing advanced analyzer | `icuvisor_list_advanced_capabilities`; record the gap instead of calculating a substitute in chat |

The workflow never calls write or delete tools and never requests `include_full`, raw streams, raw payloads, or full histories. If a response is paginated, the assistant either reads the pages required for a completeness claim or labels the evidence partial with its covered window or count.

## Review and move to a fresh chat

1. Review every goal, constraint, and accepted decision. Remove anything you did not explicitly state or accept.
2. Check that every Icuvisor claim names its source, athlete-local evidence date or window, and freshness (`as_of` or `not provided`).
3. Remove private details you do not need in the next chat.
4. Start a fresh conversation with Icuvisor connected and manually paste only the reviewed Markdown.
5. Ask the new chat to treat conversation-stated context as user-provided and refresh time-sensitive Icuvisor evidence before relying on it.

No client automatically imports, persists, or remembers this handoff.

## Privacy and data-quality boundaries

The draft excludes credentials, API or OAuth tokens, secrets, raw athlete identifiers, local or config paths, pagination tokens, raw payloads or streams, and transport or debug metadata. Health details, precise locations, and private free-text notes are omitted by default and should appear only when you approve the minimum needed.

Missing values are never zero. Chat memory never fills a tool gap. The handoff must label `_meta.stale`, `_meta.missing_fields`, unavailable or Strava-blocked data, current-day partial data, pagination limits, and unresolved tool failures. It must not add unsupported physiological conclusions.

## Why this prompt works

- **Source-separated.** Your words remain distinct from live Icuvisor facts.
- **Date-anchored.** Athlete-local dates prevent stale-chat, client-time, and UTC drift.
- **Freshness-aware.** The next chat can tell what should be refreshed instead of treating old evidence as current.
- **Portable and private.** You control the reviewed Markdown and manually decide where it goes.
