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

## 2. CLI Quota & Discovery Queries

```bash
# Query remaining quota and reset times for an account or box (defaults to 5h window)
agentusage get <account-id>

# Output format options for scripting
agentusage get <account-id> --format json
agentusage get <account-id> --format plain
agentusage get <account-id> --format table

# Query specific quota horizons
agentusage get <account-id> --window weekly
agentusage get <account-id> --window all

# List all configured accounts, providers, and container statuses
agentusage list
agentusage list --format table
agentusage list --format json
agentusage list -q                       # Output bare IDs only for shell pipelines
```

---

## 3. Background Telemetry Daemon

```bash
# Check daemon health, uptime, and event stats
agentusage daemon status
agentusage daemon status --details

# Run daemon in foreground (for debugging/containers)
agentusage daemon run
agentusage daemon run --verbose

# Install daemon as autostart system service (systemd / launchd)
agentusage daemon install

# Uninstall autostart system service
agentusage daemon uninstall

# Ingest live coding tool hook events (Claude Code, Codex, OpenCode)
agentusage daemon hook claude_code < /tmp/turn.json
agentusage daemon hook codex < /tmp/codex-event.json
agentusage daemon hook opencode < /tmp/opencode-hook.json
```

### Multi-account boxes

`make box` installs the matching CLI (`agent-box`, `agy-box`, `opencode-box`)
into `~/.local/bin` (and appends that dir to `PATH` in `~/.bashrc` / `~/.zshrc`
if needed), then creates the profile. From the agentusage repo:

```bash
# Create a profile (installs agent-box, then same as `agent-box add physics`)
make box agent-box NAME=physics
make box agent-box physics
make box agy-box NAME=chaos
make box opencode-box NAME=work

# List / remove
make box-list
make box-list agent-box
make box-rm agent-box NAME=physics

# Launch (after the box exists)
agent-box physics
```

Aliases: `agent`/`cursor-box` → `agent-box`; `agy` → `agy-box`; `opencode` → `opencode-box`.
Launching a box needs `bwrap` (`sudo apt install bubblewrap`) and the tool CLI
(`agent`, `agy`, or `opencode`) on `PATH`.

**Ubuntu 24.04+ — `bwrap: setting up uid map: Permission denied`**

AppArmor blocks unprivileged user namespaces unless `/usr/bin/bwrap` has a
profile. Use the distro package (not Homebrew `bwrap`) and load the profile:

```bash
sudo apt install bubblewrap apparmor-profiles apparmor-utils
sudo install -m 0644 \
  /usr/share/apparmor/extra-profiles/bwrap-userns-restrict \
  /etc/apparmor.d/bwrap-userns-restrict
sudo apparmor_parser -r /etc/apparmor.d/bwrap-userns-restrict

# sanity check
/usr/bin/bwrap --ro-bind / / true

# refresh installed box CLIs (prefer /usr/bin/bwrap)
cd ~/agentusage && git pull && make box agy-box NAME=nurulz
agy-box nurulz
```

If the extra profile is missing, create a minimal one:

```bash
sudo tee /etc/apparmor.d/bwrap <<'EOF'
abi <abi/4.0>,
include <tunables/global>

profile bwrap /usr/bin/bwrap flags=(unconfined) {
  userns,
  include if exists <local/bwrap>
}
EOF
sudo apparmor_parser -r /etc/apparmor.d/bwrap
```

Check for conflicting profiles: `sudo aa-status | grep -i bwrap`.

---

## 5. Diagnostics, Maintenance & Migration

```bash
# Run 4-point system, config, daemon, and hook diagnostics
agentusage doctor
agentusage doctor --verbose

# Run detailed workstation tool and credential discovery (read-only)
agentusage doctor --detect
agentusage doctor --detect --all

# Clean obsolete agentUsage/openusage statusline settings from configs (with backup)
agentusage doctor --fix-legacy-statuslines

# Check SQLite database integrity
sqlite3 ~/.local/state/agentusage/telemetry.db "PRAGMA integrity_check;"

# Clean reset (wipes cache and restarts cleanly)
rm -rf ~/.cache/agentusage ~/.local/state/agentusage/daemon.sock
```

---

## 6. Simulation & Testing

```bash
# Run standalone interactive simulation dashboard with synthetic workloads
make demo

# Run with custom loop interval
go run ./cmd/demo -interval 2s -loop
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
