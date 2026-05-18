---
title: "Connect an AI client"
description: "Connect icuvisor to Claude, ChatGPT, Cursor, Continue, Zed, Pi, or another MCP client."
---

After installing icuvisor, configure your AI client to start the local MCP server binary. Keep the intervals.icu API key out of client JSON; icuvisor reads it from the OS keychain or server environment.

## Choose your client

| Client                                           | Start here                                                |
| ------------------------------------------------ | --------------------------------------------------------- |
| Claude Desktop                                   | [Connect Claude Desktop]({{< relref "claude-desktop" >}}) |
| Claude Code                                      | [Connect Claude Code]({{< relref "claude-code" >}})       |
| ChatGPT                                          | [Connect ChatGPT]({{< relref "chatgpt" >}})               |
| Cursor, Continue, Zed, Pi, or another MCP client | [Other MCP clients]({{< relref "other-clients" >}})       |

## macOS binary path

If you installed the macOS DMG, use this command in MCP client configuration:

```text
/Applications/icuvisor.app/Contents/MacOS/icuvisor
```

If you built from source, use the absolute path to `bin/icuvisor` in your clone.

## Non-secret configuration

Most stdio MCP clients use the same shape:

```json
{
  "mcpServers": {
    "icuvisor": {
      "command": "/Applications/icuvisor.app/Contents/MacOS/icuvisor",
      "env": {
        "INTERVALS_ICU_ATHLETE_ID": "i12345",
        "ICUVISOR_TIMEZONE": "America/Sao_Paulo",
        "ICUVISOR_TRANSPORT": "stdio"
      }
    }
  }
}
```

The full list of flags and environment variables is in the [CLI reference]({{< relref "../reference/cli" >}}). For Streamable HTTP instead of stdio, see the [HTTP transport guide]({{< relref "../guides/http-transport" >}}).
