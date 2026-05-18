---
title: "Windows install status"
description: "Windows installer status for icuvisor."
---

A signed Windows installer is planned for v1.0. Until then, Windows users can build icuvisor from source if they are comfortable with Go tooling.

## Current recommendation

- If you need the easiest path today, use the macOS beta installer on a Mac.
- If you are a developer or power user, follow the repository build instructions on GitHub: [Build from source](https://github.com/ricardocabral/icuvisor#build-from-source).

The future Windows installer will keep the same security model: the intervals.icu API key belongs in Windows Credential Manager or the setup flow, not in MCP client JSON.
