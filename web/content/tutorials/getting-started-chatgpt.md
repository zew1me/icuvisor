---
title: Getting started with ChatGPT
description: Build icuvisor on macOS, connect it to ChatGPT over local MCP stdio, and ask your first training-data question.
---

By the end you'll have asked ChatGPT about real activity from your intervals.icu account and gotten a sourced answer from icuvisor.

## What you'll need

- macOS 13 or newer.
- Git and Apple's Command Line Tools already installed.
- Go 1.25.10 or newer already installed.
- A ChatGPT account with custom MCP connectors enabled.
- An intervals.icu account.
- About 10 minutes after those tools are installed.

Until the v1.0 installer ships, you'll build the binary first.

## Step 1 — Install icuvisor

Open Terminal and check that your Mac is ready:

```bash
git --version
go version
make --version
```

Each command prints a version.

Now paste this block:

```bash
rm -rf /Users/Shared/icuvisor-src
git clone https://github.com/ricardocabral/icuvisor.git /Users/Shared/icuvisor-src
cd /Users/Shared/icuvisor-src
make build
mkdir -p /Users/Shared/icuvisor/bin
cp ./bin/icuvisor /Users/Shared/icuvisor/bin/icuvisor
/Users/Shared/icuvisor/bin/icuvisor version
```

The last command prints the icuvisor version. Leave this Terminal window open.

![Illustrative Terminal view showing the copied /Users/Shared/icuvisor/bin/icuvisor binary printing its version.](/img/tutorials/chatgpt/01-install.png)

## Step 2 — Get your intervals.icu API key

Open [intervals.icu settings](https://intervals.icu/settings) in your browser.

Scroll to **Developer Settings** or **API Key**. Create a key if you do not already have one. Copy the key. You will paste it once into icuvisor, not into ChatGPT.

![Illustrative intervals.icu settings view focused on the API Key section, with the key and account details redacted.](/img/tutorials/chatgpt/02-api-key.png)

## Step 3 — Run `icuvisor setup`

Return to Terminal and run:

```bash
/Users/Shared/icuvisor/bin/icuvisor setup
```

Paste the API key when icuvisor asks for it. The prompt is masked, so the key does not appear in Terminal.

Enter your athlete ID and timezone when prompted. Your athlete ID is the number from your intervals.icu profile URL, with or without the `i` prefix. Use an IANA timezone such as `America/New_York`, `Europe/London`, or `America/Sao_Paulo`. Setup verifies the key, stores it in the macOS Keychain, and writes only non-secret settings to the icuvisor config file.

![Illustrative Terminal setup view showing a masked API key prompt and non-secret setup completion messages.](/img/tutorials/chatgpt/03-setup.png)

## Step 4 — Connect ChatGPT

Open ChatGPT.

Go to **Settings** → **Connectors** → **Add custom MCP**.

Name the connector `icuvisor` and paste this configuration:

```text
{
  "name": "icuvisor",
  "command": "/Users/Shared/icuvisor/bin/icuvisor",
  "transport": "stdio"
}
```

Save the connector. ChatGPT starts icuvisor when it needs your training data.

![Simulator view of ChatGPT connector settings with the icuvisor local stdio configuration filled in.](/img/tutorials/chatgpt/04-connector.png)

When ChatGPT shows the connector as connected, start a new chat.

![Simulator view showing the icuvisor connector changed to a connected state.](/img/tutorials/chatgpt/05-connected.png)

## Step 5 — Ask your first question

Paste this prompt into the new chat:

```text
Use the icuvisor connector only. Summarize my training load over the last 14 days using my intervals.icu data. Do not answer from memory or estimates.
```

ChatGPT asks icuvisor for the data it needs and then answers in plain language.

A good first answer looks like this:

> Over the last 14 days you completed 8 activities for a synthetic training load of 420. Most of the load came from riding, with smaller run and swim contributions. Your total time was 7 hours 35 minutes and your total distance was 84.2 km. I used icuvisor's training summary tool, so I did not need your API key, activity titles, or location data in the chat.

![Simulator view of a first ChatGPT answer showing icuvisor tool use and synthetic aggregate training-load values.](/img/tutorials/chatgpt/06-first-answer.png)

## What just happened

ChatGPT asked icuvisor for your data through MCP. icuvisor talked to intervals.icu using your API key, which stayed on your Mac in the Keychain.

Learn more in [What is MCP?](/explain/what-is-mcp/) and [Local-first by design](/explain/local-first/).

## Where to next

- [Configure another AI client](/connect/)
- [Explore the full tool catalog](/reference/tools/)
- [Set up coach mode for a roster](/guides/coach-mode/)

---

If a macOS security prompt, Keychain prompt, or connector error interrupts the flow, use the [troubleshooting guide](/guides/troubleshooting/) after this tutorial.
