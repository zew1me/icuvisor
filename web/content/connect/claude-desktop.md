---
title: "Connect Claude Desktop"
description: "Configure Claude Desktop to start icuvisor over MCP stdio on macOS."
---

Use this guide after installing `icuvisor.app` from the macOS DMG and storing your intervals.icu API key with `icuvisor setup`.

## Before you start

You need:

- Claude Desktop installed on macOS.
- `icuvisor.app` installed in `/Applications`.
- Your intervals.icu athlete ID, written as `i12345` or `12345`.
- Your API key stored in the macOS Keychain under service `icuvisor` and account `intervals-icu-api-key`.

Do not put your intervals.icu API key in `claude_desktop_config.json`. The config should contain only non-secret values such as athlete ID, timezone, and transport.

If you have not run setup yet:

```bash
/Applications/icuvisor.app/Contents/MacOS/icuvisor setup
```

## Edit Claude Desktop config

Claude Desktop reads MCP server definitions from:

```text
~/Library/Application Support/Claude/claude_desktop_config.json
```

Create the file if it does not exist. Add or merge this `mcpServers.icuvisor` block, replacing only the non-secret placeholders:

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

Notes:

- `ICUVISOR_TRANSPORT=stdio` is optional because stdio is the default, but keeping it explicit makes the config easier to audit.
- Use a real IANA timezone such as `UTC`, `America/Sao_Paulo`, or `Europe/London`.
- If you installed the app somewhere else, update `command` to the absolute path to `icuvisor.app/Contents/MacOS/icuvisor`.

After editing the file, fully quit and reopen Claude Desktop.

<details>
<summary>Smoke checklist</summary>

1. Open Claude Desktop after restarting it.
2. Start a new chat so Claude refreshes the MCP tool catalog.
3. Ask: `What's my FTP?`
4. Expected result: Claude calls icuvisor, uses [`get_athlete_profile`]({{< relref "../reference/tools#get_athlete_profile" >}}), and answers with your configured FTP/threshold data from intervals.icu.

If the answer says the tool is missing or cannot start, run:

```bash
/Applications/icuvisor.app/Contents/MacOS/icuvisor version
plutil -lint "$HOME/Library/Application Support/Claude/claude_desktop_config.json"
```

If the answer reports missing credentials, confirm the Keychain item exists and the athlete ID is set in the JSON:

```bash
security find-generic-password -s icuvisor -a intervals-icu-api-key >/dev/null
```

</details>

## Updating the app

Download the newer signed DMG from the GitHub release, replace `/Applications/icuvisor.app`, fully quit Claude Desktop, and start a new chat. Do not move the API key into the JSON during upgrades; it remains in Keychain.
