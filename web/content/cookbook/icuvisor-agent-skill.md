---
title: "icuvisor Agent Skill"
description: "Reusable Agent Skills instructions for using icuvisor well in OpenAI, Anthropic, and other skills-compatible clients."
weight: 16
---

The icuvisor Agent Skill is a portable `SKILL.md` file for AI clients that support the Agent Skills format. It gives the assistant concise standing instructions for using icuvisor well: call MCP tools instead of guessing, use athlete-local dates, preserve wellness scales, handle stale data, and treat writes carefully.

Use it after connecting icuvisor to your client.

## Get the skill

Download or copy:

```text
https://icuvisor.app/skills/icuvisor-training/SKILL.md
```

The skill name is `icuvisor-training`.

## Install in a skills-compatible client

Create a folder named `icuvisor-training`, put the downloaded `SKILL.md` inside it, and install that folder using your client's skill flow.

Examples:

- Claude Code or Codex-style local skills: place it at `~/.claude/skills/icuvisor-training/SKILL.md`, `.claude/skills/icuvisor-training/SKILL.md`, or the equivalent skills directory for your OpenAI client.
- Claude.ai custom skills: zip the `icuvisor-training` folder, open <https://claude.ai/customize/skills>, and upload it from Claude's Skills settings if your plan exposes custom skills.
- Clients without Agent Skills support: paste the `SKILL.md` body into the client's reusable instructions or project instructions.

## What it changes

The skill does not store credentials and does not connect to intervals.icu by itself. It only teaches the assistant how to behave once icuvisor is already connected.

It tells the assistant to:

- use icuvisor MCP tools or prompts for training-data claims;
- cite the tool or prompt behind key numbers;
- interpret dates in the athlete-local timezone;
- use `resolve_calendar_dates` for relative planning dates;
- keep sleep quality on 1-4, feel on 1-5, and RPE on 1-10;
- flag missing, stale, unavailable, paginated, or truncated data;
- preview writes before changing events, workouts, or wellness rows; and
- avoid asking for API keys, OAuth tokens, cookies, raw authorization headers, or local config files in chat.

For Claude Projects specifically, the longer copy-paste version remains available in [Claude Project instructions]({{< relref "../guides/claude-project-instructions" >}}).
