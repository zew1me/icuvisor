---
title: "Local-first design"
description: "Why icuvisor keeps credentials and data flow on your computer."
---

icuvisor is local-first by default: the server binary can run on your computer, read your intervals.icu data with your credentials, and return only the requested tool response to the AI client you chose. Hosted mode is available when a client needs a public HTTPS MCP endpoint.

The goal for local mode is to keep the connector simple and inspectable. You install one signed binary, connect it to intervals.icu with your own API key, and point an MCP-compatible client at that local process. Local mode does not use an icuvisor-hosted account, relay, onboarding credit, OAuth consent screen, or SaaS quota.

## Why the API key lives in the OS keychain

An intervals.icu API key can read and change training data according to the permissions intervals.icu gives that key. It should not be pasted into chat or saved in MCP client JSON.

`icuvisor setup` stores the key in the OS keychain instead:

- macOS Keychain on macOS.
- Windows Credential Manager on Windows.
- libsecret/Secret Service on Linux desktops.

The non-secret config file can hold athlete ID, timezone, transport, and coach roster settings. The key stays separate.

## Local binary versus hosted connector flows

Some MCP clients cannot launch a local server and require a public HTTPS endpoint. For those clients, use hosted ICU Visor at `https://connect.icuvisor.app/mcp`; it signs in with Intervals.icu OAuth and stores encrypted OAuth credentials so it can relay authorized tool calls.

Choose the path that matches the client:

| Concern | Local icuvisor binary | Hosted ICU Visor connector |
| --- | --- | --- |
| Connector account | No icuvisor account to create. | Hosted account state is created after Intervals OAuth so you can manage grants and deletion. |
| intervals.icu credential | API key stored in the OS keychain on your machine. | Intervals OAuth credentials encrypted in hosted storage. |
| Tool-call path | MCP client ↔ local icuvisor ↔ intervals.icu. | MCP client ↔ hosted connector ↔ upstream service. |
| Operational dependency | Your machine, your client, intervals.icu, and your model provider. | Those services plus the hosted connector's availability and quota model. |

This is not a promise that nothing ever leaves your machine. icuvisor calls intervals.icu's public API, and your chosen AI client may send conversation content and tool results to its model provider according to that client's own terms and settings. The local-first promise is narrower and concrete: local mode does not custody your intervals.icu API key or run an icuvisor SaaS middle layer.

## What leaves your machine

The local icuvisor process calls intervals.icu's public API and sends tool responses to the local MCP client process. Local mode does not host athlete data on an icuvisor server or create an icuvisor SaaS account.

Your chosen AI client may send conversation content to its model provider according to that client's own terms and settings. icuvisor's job is to keep the intervals.icu credential out of the conversation and provide the smallest useful data response for each tool call.

## Shareable reports are user-controlled

Workflows such as `shareable_training_report` help draft Markdown from your own training data, but icuvisor does not publish, host, upload, or auto-share the result. Review and redact private health details, locations, notes, identifiers, and race logistics before you manually copy, export, or post anything.

Because the local binary is not an icuvisor SaaS product, there is no icuvisor app-side credit quota for these report drafts. You still use your chosen AI client or model subscription, and that provider's billing, privacy, and retention terms still apply.

## Why local-first matters

Local-first keeps the trust boundary small:

- You can inspect the open-source binary's behavior.
- You do not hand an API key to an icuvisor-hosted service.
- You can revoke or replace the intervals.icu key yourself.
- You can choose which MCP clients to connect.
- You have fewer moving parts to check when a client caches old tools or a local config changes.

For local setup steps, see [API key setup]({{< relref "../guides/api-key" >}}). For hosted setup, see [Hosted connector]({{< relref "../connect/hosted" >}}). For the broader privacy model, see [Privacy posture]({{< relref "privacy" >}}).
