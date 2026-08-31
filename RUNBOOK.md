# agentUsage — Command Runbook

A concise quick-reference runbook of commands to run for **agentUsage (OpenUsage)**.

---

## 1. Daily Usage & Dashboards

```bash
# Launch interactive terminal TUI dashboard
agentusage

# Launch local browser web dashboard (http://127.0.0.1:8080)
agentusage serve

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
