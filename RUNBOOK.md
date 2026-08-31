# agentUsage — Command Runbook

A concise quick-reference runbook of commands to run for **agentUsage (OpenUsage)**.

---

## 1. Daily Usage & Dashboards

```bash
# Launch interactive terminal TUI dashboard
agentusage

# Launch local browser web dashboard (http://127.0.0.1:8080)
agentusage serve

# Run the dashboard in the background (survives the terminal)
agentusage serve --detach
agentusage serve --listen 127.0.0.1:8088 --base-path /agentusage --detach

# Stop a detached dashboard
agentusage serve --stop

# Launch web dashboard with synthetic demo data
agentusage serve --demo

# Run with debug logging to stderr
AGENTUSAGE_DEBUG=1 agentusage
```

---

## 2. CLI Reports & Cost Exports

```bash
# Daily token and cost breakdown per provider & model
agentusage daily

# Weekly usage summary (Monday-Sunday)
agentusage weekly

# Monthly aggregate spend
agentusage monthly

# Active billing block tracking (e.g. Claude Code 5-hour rolling blocks)
agentusage blocks

# Top spending models across all agents
agentusage top

# Export telemetry to JSON or CSV
agentusage export --format json --since 7d --output ./export.json
agentusage export --format csv --since 30d --output ./export.csv
```

---

## 3. Background Telemetry Daemon

```bash
# Check daemon health, uptime, and event stats
agentusage telemetry daemon status

# Start daemon in background
agentusage telemetry daemon start

# Stop daemon
agentusage telemetry daemon stop

# Restart daemon
agentusage telemetry daemon restart

# Install daemon as autostart system service (systemd / launchd)
agentusage telemetry daemon install

# Uninstall autostart system service
agentusage telemetry daemon uninstall

# Run daemon in foreground (for debugging/containers)
agentusage telemetry daemon
```

---

## 4. Agent Integrations & Hooks

```bash
# List detected coding agents and integration hook status
agentusage integrations list

# Install integration hooks
agentusage integrations install claude_code
agentusage integrations install codex
agentusage integrations install opencode

# Upgrade all installed hooks to current binary version
agentusage integrations upgrade --all

# Uninstall an integration hook
agentusage integrations uninstall claude_code
```

### Multi-account boxes

Box CLIs (`agent-box`, `agy-box`, `opencode-box`) live on `PATH`. From the
agentusage repo:

```bash
# Create a profile (same as `agent-box add physics`)
make box agent-box NAME=physics
make box agent-box physics
make box agy-box NAME=chaos
make box opencode-box NAME=work

# List / remove
make box-list
make box-list agent-box
make box-rm agent-box NAME=physics
```

Aliases: `agent`/`cursor-box` → `agent-box`; `agy` → `agy-box`; `opencode` → `opencode-box`.

---

## 5. tmux & Statusline Setup

```bash
# Run interactive tmux setup wizard
agentusage tmux install

# Scripted install (write directly to tmux.conf)
agentusage tmux install --write --preset compact

# Reload tmux configuration immediately
tmux source-file ~/.config/tmux/tmux.conf
# (or if using ~/.tmux.conf): tmux source-file ~/.tmux.conf

# Setup provider icon fonts (Kitty, Ghostty, WezTerm)
agentusage tmux font setup

# View all 12 built-in tmux status bar presets
agentusage tmux presets

# Diagnose tmux integration setup
agentusage tmux doctor

# Install Claude Code terminal footer statusline
agentusage statusline install
```

---

## 6. Token Pricing & Rate Calculations

```bash
# Look up token pricing rates for a model
agentusage pricing lookup gpt-4o
agentusage pricing lookup claude-3-7-sonnet

# Calculate estimated cost for input/output token counts
agentusage pricing calculate --model claude-3-7-sonnet --input 50000 --output 10000

# Update pricing registry cache from upstream feeds
agentusage pricing update
```

---

## 7. Diagnostics, Maintenance & Reset

```bash
# Run full system and environment health checks
agentusage doctor

# Scan system for installed AI tools & API key env vars
agentusage detect

# Check SQLite database integrity
sqlite3 ~/.local/state/openusage/telemetry.db "PRAGMA integrity_check;"

# Compact database and prune old records
agentusage telemetry vacuum

# Remove stale socket file if daemon crashed
rm -f ~/.local/state/openusage/daemon.sock

# Kill all running agentusage instances
killall agentusage 2>/dev/null

# Clean reset (wipes cache and restarts cleanly)
rm -rf ~/.cache/agentusage ~/.local/state/openusage/daemon.sock
```

---

## 8. Tailscale Serve (`/agentusage` on a shared hostname)

On `jobby-dev-use`, HTTPS Serve already uses `/` (OpenCode) and `/app2` (Jobagami).
Mount the web dashboard at **`/agentusage`** — do not steal `/`.

Never run `tailscale serve reset`, `tailscale serve clear`, or `tailscale serve <port>`
without `--set-path`. Check `tailscale serve status` before and after.

```bash
# 1. Build if needed, then serve on loopback :8088 under /agentusage
cd ~/agentusage && make build
./bin/agentusage serve --listen 127.0.0.1:8088 --base-path /agentusage --no-open

# Installed binary:
# agentusage serve --listen 127.0.0.1:8088 --base-path /agentusage --no-open

# 2. Add Serve path (does not replace / or /app2)
tailscale serve --bg --https=443 --set-path=/agentusage http://127.0.0.1:8088/agentusage

# 3. Confirm all three rows
tailscale serve status
curl -fsS http://127.0.0.1:8088/agentusage/healthz
```

Open `https://jobby-dev-use.tail95afc9.ts.net/agentusage` on a tailnet device
(use the hostname from `tailscale serve status` if it differs).

Remove only this path:

```bash
tailscale serve --https=443 --set-path=/agentusage off
```
