---
title: "Safety modes and toolset tiers"
description: "Reference for icuvisor's write/delete gate and compact/core/full tool catalog tiers."
---

icuvisor decides which MCP tools exist at server startup. The AI client cannot enable hidden tools with a tool argument. Change these values in your MCP client configuration, then restart icuvisor and start a fresh conversation so the client reloads the catalog.

For the complete flag and environment-variable list, see the [CLI reference]({{< relref "cli" >}}).

## Delete/write registration mode

`ICUVISOR_DELETE_MODE` controls write and delete tool registration.

| Value            | Registration effect                                                                                  | Typical use                                                                                 |
| ---------------- | ---------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------- |
| `safe`           | Registers read tools and non-destructive write tools. Delete tools are omitted. This is the default. | Normal athlete and coach use.                                                               |
| `full`           | Registers read tools, write tools, and delete tools.                                                 | Deliberate destructive maintenance where you are comfortable allowing delete-capable tools. |
| `none`           | Registers read tools only. Write and delete tools are omitted.                                       | Read-only analysis, demos, and cautious coach review.                                       |
| empty or unknown | Falls back to `safe`.                                                                                | Misconfiguration does not unlock deletes.                                                   |

## Toolset tier

`ICUVISOR_TOOLSET` controls how much of the catalog is registered. This is a startup-time routing aid: icuvisor does not expose a large generated OpenAPI surface and ask the model to sort through it. The default `core` catalog is curated for common training questions, `compact` trims further for smaller/local models, and `full` is an explicit opt-in for advanced or heavier workflows.

| Value            | Registration effect                                                                                                                                                                                                                  | Typical use                                                                                       |
| ---------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------- |
| `compact`        | Registers a smaller read-focused compatibility catalog for status, activity reads, streams, wellness, calendar reads, and safe planning context. Write/delete and specialist analyzer tools are omitted even when delete mode permits them. | Local/Ollama/OpenRouter-style clients, smaller models, and clients that choose nonexistent tools or arguments from larger catalogs. |
| `core`           | Registers the daily-use catalog. This is the default.                                                                                                                                                                                | Routine questions about activities, wellness, fitness, events, and common non-destructive writes. |
| `full`           | Registers `core` plus advanced or heavier workflows such as workout-library management, custom-item management, sport settings, training-plan application, and delete-capable tools when delete mode also permits them. | Power-user and coach workflows that need the full surface.                                        |
| empty or unknown | Falls back to `core`.                                                                                                                                                                                                                | Misconfiguration preserves the token-saving default.                                              |

The [`icuvisor_list_advanced_capabilities`]({{< relref "tools#icuvisor_list_advanced_capabilities" >}}) tool remains available in `compact` and `core` so an AI client can explain which hidden tools are available and how to enable `core` or `full`. Use that discovery tool before switching a whole client to `full`; it is designed to answer "what am I missing?" without loading every advanced tool into every conversation.

### Which tier should you pick?

- **Claude, ChatGPT, Cursor, Cline, and most daily-use clients**: keep `core`. The curated catalog balances capability with tool-selection clarity and keeps per-session tool-description load smaller than exposing every tool by default.
- **Local/Ollama/OpenRouter-style clients, Haiku-class, older, or smaller models**: start with `compact`. The reduced catalog lowers tool-selection load and per-conversation tokens while preserving common read workflows.
- **Current frontier models and expert workflows**: use `full` only when you need specialist analyzers, library management, custom items, sport settings, or other broad maintenance tools.
- **Unsure or shared setups**: keep `core`; an AI client can still call `icuvisor_list_advanced_capabilities` to discover hidden compact/core/full differences.

To switch, set `ICUVISOR_TOOLSET=compact`, `core`, or `full` in your MCP client's icuvisor entry, restart icuvisor, and start a fresh conversation.

## How the gates combine

A tool is visible only when every relevant gate allows it:

1. The active toolset must include the tool (`compact` exposes the reduced compatibility allow-list, `core` exposes the daily-use catalog, and `full` exposes all registered tools allowed by the other gates).
2. The delete/write mode must include the tool (`read` is always allowed, `write` requires `safe` or `full`, `delete` requires `full`).
3. In coach mode, the selected athlete's ACL must allow the tool and must not deny it.

Examples:

| Configuration                                           | Result                                                                           |
| ------------------------------------------------------- | -------------------------------------------------------------------------------- |
| `ICUVISOR_TOOLSET=compact`, `ICUVISOR_DELETE_MODE=safe` | Compact read-focused tools are registered. Write/delete and specialist tools are hidden. |
| `ICUVISOR_TOOLSET=core`, `ICUVISOR_DELETE_MODE=safe`    | Core read and write tools are registered. Full-tier and delete tools are hidden. |
| `ICUVISOR_TOOLSET=full`, `ICUVISOR_DELETE_MODE=safe`    | Core and full read/write tools are registered. Delete tools are hidden.          |
| `ICUVISOR_TOOLSET=full`, `ICUVISOR_DELETE_MODE=full`    | Core, full, write, and delete tools are registered.                              |
| `ICUVISOR_TOOLSET=full`, `ICUVISOR_DELETE_MODE=none`    | Core and full read tools are registered. Write and delete tools are hidden.      |

## Response metadata

Every shaped tool response includes the active gate values in `_meta`:

```json
{
  "_meta": {
    "delete_mode": "safe",
    "toolset": "core"
  }
}
```

Use these fields to confirm which server process answered a call. They are echoes of startup configuration; changing the environment after icuvisor starts does not change a running process.

## No `confirm: true` override

Destructive tools are protected at registration time. They do not expose a model-controlled `confirm` argument. If a delete tool is not registered, the AI client cannot call it until a human changes the server environment and restarts icuvisor.
