---
title: "Linux install status"
description: "Linux package status for icuvisor."
---

Linux packages are planned for v1.0. Until `.deb`, `.rpm`, or another packaged install path is available, Linux users can build icuvisor from source if they are comfortable with Go tooling.

## Current recommendation

- Follow the repository build instructions on GitHub: [Build from source](https://github.com/ricardocabral/icuvisor#build-from-source).
- Store API keys with libsecret/Secret Service where available. On systems without libsecret, use process environment only for deliberate headless or emergency fallback.

The future Linux packages will keep the same local-first security model: the intervals.icu API key belongs in the OS keychain or setup flow, not in MCP client JSON.
