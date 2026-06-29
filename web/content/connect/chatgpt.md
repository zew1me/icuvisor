---
title: "Connect ChatGPT"
description: "Connect ChatGPT to icuvisor with a hosted custom connector or a local MCP surface."
weight: 5
---

ChatGPT has two different setup paths:

- **Hosted connector**: use this for ChatGPT web at `chatgpt.com` when it asks for a remote HTTPS connector URL.
- **Local MCP surface**: use this only when your ChatGPT client explicitly supports launching a local MCP server by stdio or connecting to loopback HTTP.

## Hosted connector for ChatGPT web

Use hosted mode for ChatGPT's custom connector/app flow. ChatGPT connects from OpenAI's infrastructure, so it cannot reach `http://127.0.0.1:8765/mcp` on your laptop.

1. In ChatGPT, open your profile menu and go to **Settings > Apps & Connectors**.
2. Open **Advanced settings** at the bottom of the page and enable **Developer mode** if your account or workspace allows it.
3. Go to **Settings > Connectors > Create**.
4. Fill in the connector metadata:

   | Field | Value |
   | --- | --- |
   | Connector name | `icuvisor` |
   | Description | `Connects ChatGPT to my intervals.icu training data through hosted icuvisor. Use for athlete profile, fitness, wellness, activities, events, training plans, workouts, and safe write workflows. Do not invent unavailable data.` |
   | Connector URL | `https://connect.icuvisor.app/mcp` |

5. Click **Create**.
6. Complete the hosted icuvisor authorization flow, choose hosted preferences, continue to Intervals.icu, and approve the requested OAuth scopes. Choose `core` for normal ChatGPT use; use `compact` only if a smaller/local-compatible ChatGPT surface struggles with the tool catalog, or `full` for expert workflows that need every tool.
7. Start a new ChatGPT conversation, click **+**, choose **More**, and add the icuvisor connector to the chat.

Verify with:

```text
Use icuvisor to tell me my current FTP and timezone. Do not estimate.
```

Hosted mode uses Intervals.icu OAuth. Do not paste an Intervals API key into ChatGPT, the connector metadata, or chat.

Provider reference: [OpenAI: Connect from ChatGPT](https://developers.openai.com/apps-sdk/deploy/connect-chatgpt).

## Local MCP surfaces

Use this section only when the ChatGPT surface you are using explicitly supports local MCP servers. ChatGPT web custom connectors should use the hosted flow above.

### Before you start

- Install icuvisor and run setup.
- Confirm the binary starts with `icuvisor version`.
- Know your non-secret athlete ID and timezone.

### Stdio configuration shape

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
        "ICUVISOR_TRANSPORT": "stdio",
        "ICUVISOR_TOOLSET": "core"
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
        "ICUVISOR_TRANSPORT": "stdio",
        "ICUVISOR_TOOLSET": "core"
      }
    }
  }
}
```

If your ChatGPT MCP surface expects a single server object rather than a full `mcpServers` map, use the `icuvisor` object from the example as that server definition. `ICUVISOR_TOOLSET=core` is the recommended default for ChatGPT; switch to `compact` for reduced-catalog compatibility only after starting a fresh conversation.

### HTTP alternative

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

If a ChatGPT UI asks only for a remote HTTPS connector URL, use the hosted connector instead. Do not tunnel the local server with cloudflared, ngrok, or a similar public tunnel; that would expose an unauthenticated MCP endpoint using the intervals.icu credentials configured for the local process.

## Verify

Start a fresh ChatGPT conversation after saving the MCP configuration, then ask a simple profile question such as `What's my FTP?` The expected tool call is [`get_athlete_profile`]({{< relref "../reference/tools#get_athlete_profile" >}}). After that, try [What can I ask icuvisor?]({{< relref "../cookbook/what-can-i-ask" >}}) for concrete starter prompts. If ChatGPT keeps using old tool names, old schemas, or stale timezone/zone assumptions, follow the [stale conversation troubleshooting guide]({{< relref "../guides/troubleshooting#stale-conversations-and-cached-tool-catalogs" >}}).
