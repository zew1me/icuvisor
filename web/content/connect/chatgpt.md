---
title: "Connect ChatGPT"
description: "Minimal ChatGPT MCP connection notes for icuvisor."
---

ChatGPT MCP support is evolving across local and remote connector surfaces. Use local mode when ChatGPT can run a local MCP server by stdio or connect to loopback Streamable HTTP. Use hosted mode when ChatGPT asks for a remote HTTPS connector URL.

## Before you start

- Install icuvisor and run setup.
- Confirm the binary starts with `icuvisor version`.
- Know your non-secret athlete ID and timezone.

## Stdio configuration shape

Use the same stdio server definition as the Claude clients when ChatGPT asks for a local MCP server command:

macOS:

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

Windows:

```json
{
  "mcpServers": {
    "icuvisor": {
      "command": "C:\\Users\\<you>\\AppData\\Local\\Programs\\icuvisor\\icuvisor.exe",
      "env": {
        "INTERVALS_ICU_ATHLETE_ID": "i12345",
        "ICUVISOR_TIMEZONE": "Europe/Brussels",
        "ICUVISOR_TRANSPORT": "stdio"
      }
    }
  }
}
```

If your ChatGPT MCP surface expects a single server object rather than a full `mcpServers` map, use the `icuvisor` object from the example as that server definition.

## HTTP alternative

If your ChatGPT MCP surface expects an HTTP URL, start icuvisor with Streamable HTTP on loopback:

macOS:

```bash
ICUVISOR_TRANSPORT=http /Applications/icuvisor.app/Contents/MacOS/icuvisor
```

Windows PowerShell:

```powershell
$env:ICUVISOR_TRANSPORT = "http"
& "$env:LOCALAPPDATA\Programs\icuvisor\icuvisor.exe"
```

Then point the client at:

```text
http://127.0.0.1:8765/mcp
```

Do not bind icuvisor to a LAN address unless you intentionally want other machines to reach the unauthenticated local MCP server. The HTTP transport guide covers the security tradeoff in more detail.

## Remote connector UI

ChatGPT-style remote custom connector UIs are different from local MCP client configuration. They run from the provider's infrastructure and require an HTTPS MCP endpoint that is reachable from that infrastructure. They cannot call `http://127.0.0.1:8765/mcp` on your laptop.

Use the hosted connector URL:

```text
https://connect.icuvisor.app/mcp
```

The hosted flow signs in with Intervals.icu OAuth and lets you choose the hosted tool preferences before the client receives a grant. Do not tunnel the local server with cloudflared, ngrok, or a similar public tunnel; that would expose an unauthenticated MCP endpoint using the intervals.icu credentials configured for the local process.

See [Hosted connector]({{< relref "hosted" >}}) for the full hosted path.

## Verify

Start a fresh ChatGPT conversation after saving the MCP configuration, then ask a simple profile question such as `What's my FTP?` The expected tool call is [`get_athlete_profile`]({{< relref "../reference/tools#get_athlete_profile" >}}). If ChatGPT keeps using old tool names, old schemas, or stale timezone/zone assumptions, follow the [stale conversation troubleshooting guide]({{< relref "../guides/troubleshooting#stale-conversations-and-cached-tool-catalogs" >}}).
