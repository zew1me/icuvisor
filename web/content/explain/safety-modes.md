---
title: "Why safety modes exist"
description: "The reasoning behind safe, full, and none delete/write modes."
---

An AI assistant can be helpful with training data, but it should not be able to talk itself into destructive permissions. icuvisor's safety modes make write/delete capability a server startup choice controlled by the human running icuvisor, not a per-call choice controlled by the model.

## Why there is no `confirm: true`

A model-controlled confirmation flag is not a strong safety boundary. If a tool error says "set `confirm: true` to override," the model can simply send that argument.

icuvisor handles destructive capability before tools are registered. If delete mode hides a tool, the client cannot see it in `tools/list` and cannot call it by inventing an override.

## Why `safe` is the default

Most useful workflows are not destructive: reading activities, reading wellness, checking fitness, adding planned events, updating wellness fields, or adding comments. `safe` keeps those workflows available while omitting delete tools.

This gives the assistant enough capability for normal use without making irreversible operations part of the default catalog.

## When to use `none`

Use `none` when you want analysis only: demos, read-only coach review, or any session where you do not want the assistant to write upstream data.

## When to use `full`

Use `full` only when you intentionally want delete-capable workflows available. Pair it with a fresh conversation after restart so the client sees the updated catalog clearly.

In coach mode, `full` is still not enough by itself: the selected athlete's ACL must also allow the tool.

## Exact values

For the complete table of modes, toolset tiers, and `_meta.delete_mode` / `_meta.toolset` echoes, see [Safety modes and toolset tiers]({{< relref "../reference/safety-modes" >}}).
