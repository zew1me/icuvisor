---
title: "Local-first design"
description: "Why icuvisor keeps credentials and data flow on your computer."
---

icuvisor is local-first: the server binary runs on your computer, reads your intervals.icu data with your credentials, and returns only the requested tool response to the AI client you chose.

## Why the API key lives in the OS keychain

An intervals.icu API key can read and change training data according to the permissions intervals.icu gives that key. It should not be pasted into chat or saved in MCP client JSON.

`icuvisor setup` stores the key in the OS keychain instead:

- macOS Keychain on macOS.
- Windows Credential Manager on Windows.
- libsecret/Secret Service on Linux desktops.

The non-secret config file can hold athlete ID, timezone, transport, and coach roster settings. The key stays separate.

## What leaves your machine

The local icuvisor process calls intervals.icu's public API and sends tool responses to the local MCP client process. icuvisor does not host athlete data on an icuvisor server, and there is no icuvisor SaaS account in the normal local setup.

Your chosen AI client may send conversation content to its model provider according to that client's own terms and settings. icuvisor's job is to keep the intervals.icu credential out of the conversation and provide the smallest useful data response for each tool call.

## Why local-first matters

Local-first keeps the trust boundary small:

- You can inspect the open-source binary's behavior.
- You do not hand an API key to a third-party training SaaS.
- You can revoke or replace the intervals.icu key yourself.
- You can choose which MCP clients to connect.

For setup steps, see [API key setup]({{< relref "../guides/api-key" >}}).
