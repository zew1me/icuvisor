<p align="center">
  <img src="docs/brand/logo-wordmark.png" alt="icuvisor" width="540" />
</p>

[![CI](https://github.com/ricardocabral/icuvisor/actions/workflows/ci.yml/badge.svg)](https://github.com/ricardocabral/icuvisor/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/ricardocabral/icuvisor?sort=semver)](https://github.com/ricardocabral/icuvisor/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

> icuvisor is an open-source, local-first [Model Context Protocol](https://modelcontextprotocol.io) server for [intervals.icu](https://intervals.icu). Run the signed binary on your computer to connect an MCP-compatible AI client to your training data without putting your intervals.icu API key in chat. If your client needs a public HTTPS endpoint, use the optional hosted connector instead.

## Get started

1. [Install icuvisor](https://icuvisor.app/install/).
2. [Connect your AI client](https://icuvisor.app/connect/).
3. Try a prompt from the [Cookbook](https://icuvisor.app/cookbook/).

The documentation covers local and hosted modes, privacy, safety modes, supported clients, troubleshooting, and the complete tool reference.

## For contributors

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and contribution guidelines. Product scope and planned work live in the [PRD](docs/prd/PRD-icuvisor.md) and [roadmap](ROADMAP.md).

## Security

```bash
curl -fsSL https://icuvisor.app/install.sh | sh
```

On native Windows / PowerShell:

```powershell
iwr -useb https://icuvisor.app/install.ps1 | iex
```

See [SECURITY.md](SECURITY.md#installer-integrity) for installer signature verification and in-place update behaviour.

Prefer a package manager? Use `brew install ricardocabral/tap/icuvisor` on macOS, or install from Winget on Windows:

```powershell
winget install --id RicardoCabral.icuvisor --exact
```

Open a new PowerShell or Command Prompt window after installation, then run `icuvisor version`. Windows users can also use the PowerShell installer above or download the `.msi` from the [releases page](https://github.com/ricardocabral/icuvisor/releases).

### Connect from Cursor

Run `icuvisor setup` first so the local server can read your intervals.icu API key from the OS keychain or your explicit headless config. Do not paste API keys into Cursor or into an MCP tool call.

Use the Cursor install link for a local stdio server with no secrets embedded: [Install icuvisor in Cursor](https://cursor.com/install-mcp?name=icuvisor&config=%7B%22command%22%3A%22icuvisor%22%2C%22args%22%3A%5B%5D%2C%22env%22%3A%7B%22ICUVISOR_CONFIG%22%3A%22%22%7D%7D).

If you prefer manual configuration, add this to `~/.cursor/mcp.json` and leave `ICUVISOR_CONFIG` empty unless you intentionally use a non-default config path:

```json
{
  "mcpServers": {
    "icuvisor": {
      "command": "icuvisor",
      "args": [],
      "env": {
        "ICUVISOR_CONFIG": ""
      }
    }
  }
}
```

### MCP discovery

`server.json` in this repository describes icuvisor for MCP Registry-style discovery and points to the signed MCPB release artifact. No Glama badge is shown until Glama exposes a stable public icuvisor page; speculative discovery links are intentionally omitted.

Learn how to connect your AI assistant, try beginner prompt examples, read the tool catalog, and troubleshoot stale conversations or cached tool catalogs at <https://icuvisor.app>.

### Downloadable prompt packs

Copyable client prompt packs are available in [`docs/prompts/`](docs/prompts/README.md) for Weekly review, Coaching conversation handoff, Race-week taper, Ride analysis, Fueling review, Masters plan review, and Coach roster triage modes. Use them as custom mode/profile instructions in clients that support saved modes, or paste the prompt block at the start of a fresh chat in clients that do not. The packs are designed to route assistants through Icuvisor's deterministic tools, `_meta.method`/unit metadata, and explicit caveats instead of generic prompt-only coaching or invented baselines.

### Fitness projection with ATP/periodization targets

`get_annual_training_plan` summarizes existing PLAN, TARGET, and NOTE calendar events into season phases, weekly load/TSS targets, recovery/context notes, and `_meta.projection_bridge.weekly_plan_targets`. `propose_annual_training_plan` creates a read-only deterministic season proposal from caller-provided goals and constraints when no calendar writes should occur. `apply_annual_training_plan` is the separate full-toolset write path for those proposals: it defaults to dry-run, returns a preview token, and only writes deterministic Icuvisor-owned ATP phase notes after an explicit commit call with that token. Copy existing or proposed bridge rows into `get_fitness_projection.weekly_plan_targets` to model future CTL/ATL/TSB scenarios without asking the assistant to invent daily loads. `get_fitness_projection` distributes each ISO-Monday weekly `training_load` evenly as `training_load/7` across projected future dates in that week. Explicit `planned_daily_loads` win for matching dates and are not redistributed.

```json
{
  "start_date": "2026-06-03",
  "horizon_days": 14,
  "weekly_plan_targets": [
    { "week_start_date": "2026-06-01", "training_load": 700 },
    { "week_start_date": "2026-06-08", "training_load": 840 }
  ],
  "planned_daily_loads": [
    { "date": "2026-06-10", "training_load": 60 }
  ]
}
```

The ATP `_meta.projection_bridge` reports which weekly TARGET rows are safe to copy and which partial or missing-load weeks were excluded. The projection `_meta.assumptions` reports target counts, filled days, override counts, the ISO-Monday anchor convention, and `source_tools` adds planning sources when weekly targets are supplied.

Example calendar write prompt: "Mark me sick from 2026-08-10 through 2026-08-12" maps to `add_unavailable_date_range` with `category: "SICK"`, `start_date`, and `end_date`; use `include_full: true` only when raw upstream event payloads are needed.

## For developers

### Build from source

```bash
git clone https://github.com/ricardocabral/icuvisor.git
cd icuvisor
make build
./bin/icuvisor version
```

### Direct CLI

The release also includes `icuvisor-cli`, a standalone view over the same registered tool catalog used by MCP. It is useful for scripts or local agents that load a concise skill/command contract instead of an MCP tool schema. Run `icuvisor setup` once to provision credentials; never pass an API key as a tool argument.

```bash
icuvisor-cli capabilities
icuvisor-cli doctor
icuvisor-cli tools list
icuvisor-cli tools describe get_today
icuvisor-cli tools call get_today --args '{}'
echo '{}' | icuvisor-cli tools call get_today --args-file -
```

This PR's standalone contract is tools-only; Resources and Prompts are required follow-ups. `tools list`, `tools describe`, and successful `tools call` results are JSON on stdout. `tools call` accepts exactly one JSON object through `--args <json>`, `--args-file <path>`, or `--args-file -` for stdin (and defaults to `{}`). Failures leave stdout empty and write one JSON error to stderr (exit `2` for usage errors, `1` for runtime errors). The standalone view uses the configured local athlete only; coach workflows remain MCP-only.

### Project layout

```
cmd/icuvisor/             MCP server binary entrypoint
cmd/icuvisor-cli/         Standalone direct-tool CLI binary entrypoint
internal/app/             MCP process dispatch, startup wiring, `setup` / `diagnostics` commands
internal/cli/             Standalone CLI view and setup prompting
internal/cli/prompt/      Terminal prompting (masked input) for first-run setup
internal/config/          Config load/validate/write, athlete-ID/timezone normalization, HTTP bind, dotenv, redaction
internal/credstore/       OS keychain wrapper (macOS Keychain, Windows Credential Manager, libsecret)
internal/diagnostics/     Redacted runtime/config snapshot for `icuvisor diagnostics`
internal/intervals/       intervals.icu API client (Basic Auth, retries, structured errors)
internal/clients/         Shared typed client interfaces (athlete profile, etc.)
internal/mcp/             MCP server + stdio/Streamable HTTP transports, schema, recovery
internal/tools/           Tool implementations (registered via `tools.Catalog()`)
internal/toolcatalog/     Catalog hashing and stale-catalog CI guard surface
internal/toolchecks/      Cross-tool invariants (delete-mode gating, examples, schema snapshots)
internal/coach/           Coach-mode roster and per-athlete tool ACLs
internal/safety/          Delete-mode resolution and registration-time gating
internal/response/        Terse/full response shaping and `_meta` plumbing
internal/analysis/        Deterministic analyzer math + interval-source / auto-lap classifier
internal/prompts/         Curated MCP prompt registry
internal/resources/       MCP Resources (workout syntax, event categories, schemas, analysis formulas, athlete profile)
internal/athleteprofile/  Athlete profile read shaping shared by tool + resource
internal/workoutdoc/      WorkoutDoc Parse/Serialize for the upstream description DSL
internal/customitemschemas/ Custom-item content schema samples used by write validation
internal/units/           Unit enum parsing + preferred-unit conversion
internal/streams/         Canonical stream key normalization
docs/                     PRD, roadmap, design notes
```

### Development

Requires Go 1.25+ and (optionally) [`golangci-lint`](https://golangci-lint.run) and [`goreleaser`](https://goreleaser.com).

```bash
make build       # build ./bin/icuvisor
make test        # unit tests
make test-race   # tests with the race detector
make lint        # golangci-lint
make check       # fmt-check + vet + lint + test-race (run before pushing)
make snapshot    # local goreleaser snapshot
make docs-tools  # regenerate website tool catalog/schema data
make help        # list all targets
```

See [CONTRIBUTING.md](CONTRIBUTING.md), [SECURITY.md](SECURITY.md), and the [PRD](docs/prd/PRD-icuvisor.md).
See [SECURITY.md](SECURITY.md) for supported versions, vulnerability reporting, and installer integrity.

## License

[MIT](LICENSE).
