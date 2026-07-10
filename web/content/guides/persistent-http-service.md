---
title: "Keep local HTTP running"
description: "Run icuvisor's loopback-only Streamable HTTP endpoint as a persistent per-user service."
---

Use this guide when a client connects to an existing local HTTP URL and you want icuvisor to start at login and recover from a crash. It creates a service only for your current OS account; it does not install a system daemon or make the endpoint reachable from another machine.

The client URL is always:

```text
http://127.0.0.1:8765/mcp
```

That URL belongs in the MCP client. The server process receives `--transport http --http-bind 127.0.0.1:8765`; do not append `/mcp` to its arguments.

## Choose local or hosted mode

These recipes are for a client running on the same computer as icuvisor. A provider-hosted connector UI runs on provider infrastructure and cannot reach your laptop's loopback address. For that case, use the [hosted ICU Visor connector]({{< relref "../connect/hosted" >}}) at `https://connect.icuvisor.app/mcp`, which uses hosted OAuth. Do not publish a local endpoint through a generic public tunnel: a tunnel URL is not authentication.

If a client asks for a connector key or short name, use `icuvisor`. Avoid punctuation-heavy names because some clients reject them.

## Before you create a service

1. Install icuvisor and run `icuvisor setup` **interactively as the same OS account that will own the service**. Setup stores the API key in that account's credential store and writes only non-secret per-user configuration. See [API key setup]({{< relref "api-key" >}}) if setup has not been completed.
2. Keep the service per-user and logged-in. Do not enable Linux lingering and do not use a background or service account: Keychain, Credential Manager, and Secret Service can be locked or unavailable outside the desktop user's session.
3. Do not put an API key in a service definition, config JSON, environment directive, wrapper, or task argument. These recipes deliberately do not use a dotenv file or an explicit environment-file argument. Their working directories are system-owned locations so an accidental user-owned dotenv file cannot become a credential source.
4. Before starting any recipe, ensure another program is not already listening on port 8765.

The service definitions below use the normal non-secret config lookup and the OS credential store from setup. Removing a service never removes that configuration or the credential-store entry.

## macOS: LaunchAgent

This LaunchAgent runs only in your logged-in graphical session. It uses the signed app-bundle binary documented in [Install on macOS]({{< relref "../install/macos" >}}), launches at login, and restarts after an unsuccessful exit.

Create the plist and load it into your current graphical-user domain. The unquoted heredoc intentionally expands `$HOME` before launchd reads the plist, so the two log paths are real absolute paths rather than unexpanded shell notation.

```bash
mkdir -p "$HOME/Library/LaunchAgents" "$HOME/Library/Logs/icuvisor"
cat > "$HOME/Library/LaunchAgents/app.icuvisor.http.plist" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>app.icuvisor.http</string>
  <key>ProgramArguments</key>
  <array>
    <string>/Applications/icuvisor.app/Contents/MacOS/icuvisor</string>
    <string>--transport</string>
    <string>http</string>
    <string>--http-bind</string>
    <string>127.0.0.1:8765</string>
  </array>
  <key>WorkingDirectory</key>
  <string>/</string>
  <key>RunAtLoad</key>
  <true/>
  <key>KeepAlive</key>
  <dict>
    <key>SuccessfulExit</key>
    <false/>
  </dict>
  <key>ThrottleInterval</key>
  <integer>10</integer>
  <key>StandardOutPath</key>
  <string>$HOME/Library/Logs/icuvisor/http-service.log</string>
  <key>StandardErrorPath</key>
  <string>$HOME/Library/Logs/icuvisor/http-service.log</string>
</dict>
</plist>
EOF
plutil -lint "$HOME/Library/LaunchAgents/app.icuvisor.http.plist"
launchctl bootstrap "gui/$(id -u)" "$HOME/Library/LaunchAgents/app.icuvisor.http.plist"
```

Inspect status and logs:

```bash
launchctl print "gui/$(id -u)/app.icuvisor.http"
tail -n 100 "$HOME/Library/Logs/icuvisor/http-service.log"
lsof -nP -iTCP:8765 -sTCP:LISTEN
```

After correcting a config, credential-store, or port-conflict problem, restart the loaded service:

```bash
launchctl kickstart -k "gui/$(id -u)/app.icuvisor.http"
```

To stop it for this login session and remove it permanently, unload it first, then delete only the LaunchAgent and its non-secret log:

```bash
launchctl bootout "gui/$(id -u)" "$HOME/Library/LaunchAgents/app.icuvisor.http.plist"
rm -f "$HOME/Library/LaunchAgents/app.icuvisor.http.plist" "$HOME/Library/Logs/icuvisor/http-service.log"
```

If you installed the app in `~/Applications` rather than `/Applications`, replace only the binary path in `ProgramArguments` with that absolute app-bundle path before loading the plist.

## Linux: systemd user service

This is a `systemd --user` service, not a system service. It starts when your user manager starts during a normal login and restarts after failure. Do not run `loginctl enable-linger` for this recipe because it would outlive the desktop session that can unlock the credential store.

First resolve the actual installed binary. The shell installer can use an existing path, `/usr/local/bin`, or `~/.local/bin`, so do not guess a Linux path. The check below requires an absolute executable path and expands that path into the unit; it does not create a systemd environment variable.

```bash
ICUVISOR_BINARY="$(command -v icuvisor)"
case "$ICUVISOR_BINARY" in
  /*) ;;
  *) echo "icuvisor is not installed at an absolute path" >&2; exit 1 ;;
esac
test -x "$ICUVISOR_BINARY" || { echo "icuvisor is not executable" >&2; exit 1; }

mkdir -p "$HOME/.config/systemd/user"
cat > "$HOME/.config/systemd/user/icuvisor-http.service" <<EOF
[Unit]
Description=icuvisor loopback HTTP MCP service

[Service]
Type=simple
ExecStart="$ICUVISOR_BINARY" --transport http --http-bind 127.0.0.1:8765
WorkingDirectory=/
Restart=on-failure
RestartSec=5

[Install]
WantedBy=default.target
EOF

systemctl --user daemon-reload
systemctl --user enable --now icuvisor-http.service
```

Check the service and journal, including whether port 8765 is already occupied:

```bash
systemctl --user status icuvisor-http.service
journalctl --user --unit icuvisor-http.service -n 100 --no-pager
ss -ltn 'sport = :8765'
```

After fixing setup, Secret Service, config, or a port conflict, restart the service:

```bash
systemctl --user restart icuvisor-http.service
```

Stop, disable, and remove the user-service definition without touching your config or credential store:

```bash
systemctl --user disable --now icuvisor-http.service
rm -f "$HOME/.config/systemd/user/icuvisor-http.service"
systemctl --user daemon-reload
```

## Windows: Task Scheduler

This recipe creates a task for the current interactive user only. It runs at that user's logon, has no finite execution limit, and retries a failed server process once per minute. It does not use the Task Scheduler choices for a service account or for running whether the user is logged on or not.

Run `icuvisor setup` interactively first, then open PowerShell as the same normal user. This command verifies the normal per-user installer location and runs setup without placing the credential in the task:

```powershell
$icuvisorBinary = Join-Path $env:LOCALAPPDATA "Programs\icuvisor\icuvisor.exe"
if (-not (Test-Path -LiteralPath $icuvisorBinary -PathType Leaf)) {
  throw "icuvisor.exe was not found at $icuvisorBinary"
}
& $icuvisorBinary setup
```

Create a non-secret wrapper and register the task. The wrapper quotes the absolute binary and log paths, redirects every output stream to the application log, and returns the child process exit code so Task Scheduler can apply its retry policy. The task itself starts PowerShell by its resolved absolute path and uses `C:\Windows\System32` as its working directory.

```powershell
$taskName = "icuvisor-http"
$icuvisorBinary = Join-Path $env:LOCALAPPDATA "Programs\icuvisor\icuvisor.exe"
$serviceDirectory = Join-Path $env:LOCALAPPDATA "icuvisor\http-service"
$wrapper = Join-Path $serviceDirectory "icuvisor-http.ps1"
$logPath = Join-Path $serviceDirectory "icuvisor-http.log"
$currentUser = [System.Security.Principal.WindowsIdentity]::GetCurrent().Name

if (-not (Test-Path -LiteralPath $icuvisorBinary -PathType Leaf)) {
  throw "icuvisor.exe was not found at $icuvisorBinary"
}
New-Item -ItemType Directory -Force -Path $serviceDirectory | Out-Null
$wrapperContent = @"
`$ErrorActionPreference = 'Stop'
& '$icuvisorBinary' --transport http --http-bind 127.0.0.1:8765 *>> '$logPath'
exit `$LASTEXITCODE
"@
Set-Content -LiteralPath $wrapper -Value $wrapperContent -Encoding utf8

$action = New-ScheduledTaskAction `
  -Execute "$PSHOME\powershell.exe" `
  -Argument "-NoLogo -NoProfile -NonInteractive -File `"$wrapper`"" `
  -WorkingDirectory "$env:SystemRoot\System32"
$principal = New-ScheduledTaskPrincipal -UserId $currentUser -LogonType Interactive -RunLevel Limited
$trigger = New-ScheduledTaskTrigger -AtLogOn -User $currentUser
$settings = New-ScheduledTaskSettingsSet `
  -ExecutionTimeLimit ([TimeSpan]::Zero) `
  -RestartCount 999 `
  -RestartInterval (New-TimeSpan -Minutes 1) `
  -MultipleInstances IgnoreNew
$task = New-ScheduledTask -Action $action -Principal $principal -Trigger $trigger -Settings $settings
Register-ScheduledTask -TaskName $taskName -InputObject $task -Force | Out-Null
Start-ScheduledTask -TaskName $taskName
```

Inspect Task Scheduler state, the application log written by the wrapper, and recent task events:

```powershell
Get-ScheduledTask -TaskName "icuvisor-http"
Get-ScheduledTaskInfo -TaskName "icuvisor-http"
Get-Content -LiteralPath "$env:LOCALAPPDATA\icuvisor\http-service\icuvisor-http.log" -Tail 100
Get-WinEvent -LogName "Microsoft-Windows-TaskScheduler/Operational" -MaxEvents 50 |
  Where-Object { $_.Message -match "icuvisor-http" } |
  Select-Object TimeCreated, Id, LevelDisplayName, Message
Get-NetTCPConnection -LocalPort 8765 -State Listen -ErrorAction SilentlyContinue
```

After correcting a setup, Credential Manager, config, or port-conflict problem, stop and start the registered task:

```powershell
Stop-ScheduledTask -TaskName "icuvisor-http"
Start-ScheduledTask -TaskName "icuvisor-http"
```

Remove the task, wrapper, and non-secret log without touching config or Credential Manager:

```powershell
Stop-ScheduledTask -TaskName "icuvisor-http" -ErrorAction SilentlyContinue
Unregister-ScheduledTask -TaskName "icuvisor-http" -Confirm:$false
Remove-Item -LiteralPath "$env:LOCALAPPDATA\icuvisor\http-service" -Recurse -Force
```

## Connect and recover safely

Configure the same-machine client with `http://127.0.0.1:8765/mcp`. If it cannot connect, inspect the service status and log for your platform, then check whether another process owns port 8765. Correct the credential-store or non-secret config problem with `icuvisor setup` under the same logged-in account, then use the platform restart command above.

Do not change the service bind address to make a remote client work. A LAN bind exposes an unauthenticated MCP server to anyone who can reach it. For a client whose connector runs remotely, use [hosted mode]({{< relref "../connect/hosted" >}}) instead.
