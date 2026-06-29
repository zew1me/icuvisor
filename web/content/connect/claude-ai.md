---
title: "Connect Claude.ai"
description: "Add hosted icuvisor to Claude.ai as a custom remote MCP connector."
weight: 2
---

Use this path for Claude in the browser at `claude.ai`. Claude.ai cannot launch the local `icuvisor` binary on your computer, so it needs the hosted HTTPS MCP endpoint.

Claude custom connectors are currently a beta feature. Anthropic supports them on Free, Pro, Max, Team, and Enterprise plans, with Free users limited to one custom connector.

## Add the connector

1. Open [Claude.ai connector settings](https://claude.ai/customize/connectors).
2. If you are on an individual Free, Pro, or Max plan, click **+**, then **Add custom connector**.
3. If you are on a Team or Enterprise plan, an Owner may need to add the connector first from **Organization settings > Connectors**. Members then return to **Customize > Connectors** and click **Connect**.
4. Paste this remote MCP server URL:

   ```text
   https://connect.icuvisor.app/mcp
   ```

5. Leave **OAuth Client ID** and **OAuth Client Secret** empty unless the hosted icuvisor operator gave you explicit client credentials.
6. Click **Add**.

Claude opens the hosted icuvisor authorization flow. Choose your hosted preferences, continue to Intervals.icu, approve the requested OAuth scopes, and return to Claude.

Provider reference: [Claude custom connectors using remote MCP](https://support.claude.com/en/articles/11175166-get-started-with-custom-connectors-using-remote-mcp).

## Enable it in a chat

1. Start a new Claude chat.
2. Use the **+** button near the composer, open **Connectors**, and enable icuvisor for the conversation.
3. Ask a simple check:

   ```text
   Use icuvisor to tell me my current FTP and timezone. Do not estimate.
   ```

A working setup should call [`get_athlete_profile`]({{< relref "../reference/tools#get_athlete_profile" >}}) or a hosted setup/status tool before answering. Then use [What can I ask icuvisor?]({{< relref "../cookbook/what-can-i-ask" >}}) for starter prompts that do not require knowing tool names.

## Notes

- Hosted mode uses Intervals.icu OAuth. Do not paste an Intervals API key into Claude.ai, the connector settings, or chat.
- Claude connects from Anthropic's infrastructure, so the MCP URL must be public HTTPS. Do not use `http://127.0.0.1:8765/mcp` here.
- Start a new chat after changing hosted preferences or reconnecting the connector so Claude reloads the current tool catalog.

For the shared hosted flow and settings page, see [Hosted connector]({{< relref "hosted" >}}).
