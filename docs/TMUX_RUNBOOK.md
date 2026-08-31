# agentUsage / OpenUsage — tmux Integration Runbook

Comprehensive operational guide for installing, configuring, troubleshooting, and maintaining the agentUsage tmux status bar integration.

---

## 1. Overview & Architecture

The `agentusage tmux` integration injects a lightweight, real-time AI tool usage monitor into the tmux status line (`status-left` or `status-right`). It tracks active models, costs, context window usage, quota thresholds (e.g., 5-hour rolling blocks), and burn rate ($/hr).

```
┌────────────────────────────────────────────────────────────────────────┐
│ tmux Server (reads status-right / status-left at status-interval)       │
└───────────────────────────────────┬────────────────────────────────────┘
                                    │ executes every N seconds
                                    ▼
┌────────────────────────────────────────────────────────────────────────┐
│ CLI: `agentusage tmux` (800ms hard budget + anti-flicker cache)        │
└───────────┬────────────────────────────────────────────────┬───────────┘
            │                                                │
            ▼ (via IPC socket)                               ▼ (direct probe)
┌───────────────────────────────┐               ┌────────────────────────┐
│ Local Telemetry Daemon        │               │ Provider CLI/State     │
│ (SQLite at telemetry.db)      │               │ (Claude, Codex, etc.)  │
└───────────────────────────────┘               └────────────────────────┘
```

### Key Components
- **CLI Renderer (`agentusage tmux`)**: Formats metrics according to selected preset/template. Operates under an 800ms timeout budget so tmux is never blocked.
- **Anti-Flicker Cache (`~/.cache/agentusage/tmux-laststatus`)**: Caches last valid render (TTL 10m) to prevent empty status segments during transient daemon restarts.
- **Managed tmux.conf Block**: Cleanly bracketed between `# >>> agentusage tmux >>>` and `# <<< agentusage tmux <<<` sentinels.
- **Provider Icon Font (`openusage-icons.ttf`)**: Private Use Area (PUA) glyphs providing crisp, brand-colored vector logos in supported terminals.

---

## 2. Installation & Setup

### Standard Interactive Setup (Recommended)
Launches the interactive TUI wizard to configure position, presets, colors, icons, and keybindings:

```bash
agentusage tmux install
```

After the wizard finishes, reload your tmux configuration:
```bash
# For modern XDG tmux configs:
tmux source-file ~/.config/tmux/tmux.conf

# For legacy home directory configs:
tmux source-file ~/.tmux.conf
```

---

### Non-Interactive / Scripted Installation
For automated provisioning, dotfiles scripts, or CI/CD environments:

```bash
# Basic installation with 'compact' preset to status-right
agentusage tmux install --write

# Specific preset with custom refresh rate and popup binding
agentusage tmux install \
  --write \
  --preset claude-focused \
  --position right \
  --interval 5 \
  --bind-popup u \
  --bind-refresh r

# Multi-tool pinned installation (displays Claude Code and Cursor side-by-side)
agentusage tmux install --write --preset compact
```

#### CLI Installation Flags
| Flag | Default | Description |
|---|---|---|
| `--write` | `false` | Writes directly to detected `tmux.conf` (creates `.bak` backup) |
| `--position` | `right` | Placement in status line: `left`, `right`, or `both` |
| `--preset` | `compact` | Named visual template (see catalog below) |
| `--interval` | `5` | Status refresh interval in seconds (`status-interval`) |
| `--right-length` | `200` | Max character width for `status-right` |
| `--left-length` | `80` | Max character width for `status-left` |
| `--bind-popup <key>` | `""` | Binds `Prefix + <key>` to open full dashboard popup (`tmux 3.2+`) |
| `--bind-refresh <key>`| `""` | Binds `Prefix + <key>` to trigger instant status recalculation |
| `--with-font` | `false` | Auto-installs icon font without interactive prompts |
| `--no-font` | `false` | Skips icon font prompt entirely |

---

## 3. Presets & Custom Formatting

### Built-in Presets Catalog (`agentusage tmux presets`)

| Preset | Glyph Tier | Output Example | Best Suited For |
|---|---|---|---|
| `compact` *(default)* | Unicode | `🤖 5h 15% $6.79/today` | General minimal usage |
| `claude-focused` | Unicode | `🤖 Opus 4.7 $3.40 block (2h17m) 🔥 $1.20/hr 🧠 42%` | Heavy Claude Code workflows |
| `burn` | Unicode | `🔥 $1.20/hr → $9.40 EOB` | Cost-aware & spike monitoring |
| `emoji-rich` | Unicode | `🤖 CLAUDE_CODE │ 💰 $4.21 │ 📅 42 req │ 🔥 $1.20/hr │ 🧠 42%` | High-detail segmented bar |
| `minimal` | ASCII | `claude_code $4.21` | Distraction-free / text-only |
| `cost-only` | ASCII | `$4.21` | Ultra-compact status bars |
| `ascii-safe` | ASCII | `[CLAUDE_CODE] $4.21 block:47% burn:$1.20/hr ctx:42%` | Legacy terminals / NO unicode |
| `verbose` | Unicode | `🤖 Opus 4.7 │ 💰 $4.21 today / $3.40 block │ 🔥 $1.20/hr │ 🧠 84k (42%)` | Wide displays & full telemetry |
| `themed` | Unicode | `🤖 $4.21 $1.20/hr` | Styled brand colors |
| `multi-tool` | Unicode | `claude_code │ cursor │ codex` | Pinned multi-provider view |
| `nerdfont` | NerdFont | `claude_code  $4.21  $1.20` | Terminals using Nerd Fonts |
| `powerline` | NerdFont | `🤖  $4.21  $1.20/hr` | Powerline theme integrations |

---

### Custom Format Templates

You can pass custom format strings via the `--format` flag:

```bash
agentusage tmux --format "{tool:icon:brand} {model} #[fg=green]{today_cost:money}#[default] ({burn_rate:money}/hr)"
```

#### Supported Template Variables
- `{tool}` / `{provider}`: Active provider name (e.g., `claude_code`, `cursor`, `codex`).
- `{tool:icon:brand}`: Provider icon tinted with its brand hex color.
- `{model}`: Active LLM model name (e.g., `claude-3-7-sonnet`, `gpt-4o`).
- `{today_cost}`: Accumulated cost for the current calendar day.
- `{block_cost}`: Accumulated cost in the current billing / rate-limit window.
- `{block_pct}`: Percentage of quota used in current block (e.g., 5-hour limit).
- `{block_remaining}`: Human-readable duration remaining until block reset (e.g., `2h14m`).
- `{context_pct}`: Context window usage percentage.
- `{context_tokens}`: Context tokens currently active in session.
- `{burn_rate}`: Current expenditure rate formatted as `$/hr`.

#### Variable Format Modifiers
- `:money` — Formats as currency (e.g., `$4.20`).
- `:pct` / `:pct:color` — Formats as percentage, optionally colored dynamically (`green` < 70%, `yellow` 70-90%, `red` > 90%).
- `:duration` — Formats time spans (e.g., `1h45m`).
- `:trunc:<N>` — Truncates string to N characters.

#### Conditionals
Conditional blocks display sections only when data is present:
```text
{?block_pct: 5h {block_pct:pct:color}}
{?today_cost: {today_cost:money}/today}
```

---

## 4. Provider Icon Font Setup

agentUsage ships with custom vector icons for AI providers (Anthropic, OpenAI, Cursor, Copilot, Gemini, Mistral, Ollama, etc.).

### Automatic Terminal Configuration
For terminals supporting per-range font fallback (Kitty, Ghostty, WezTerm):

```bash
# 1. Install font file into user fonts (~/.local/share/fonts or ~/Library/Fonts)
agentusage tmux font install

# 2. Auto-configure terminal fallback config
agentusage tmux font setup
```

### Non-Fallback Terminals (iTerm2, Terminal.app)
Terminals without unicode range fallback require font augmentation (patching):

```bash
agentusage tmux font patch
```
*Note: This creates an augmented copy (e.g., `JetBrains Mono +agentUsage`) without touching the original font.*

### Verify Font Status
```bash
agentusage tmux font status
```

---

## 5. Operations & Diagnostics

### Run Health Checks (`doctor`)
Runs end-to-end environment validation:

```bash
agentusage tmux doctor
```
Checks verified:
- [x] tmux binary existence and version (requires >= 3.0, popup features >= 3.2).
- [x] Terminal truecolor capability (`$COLORTERM` / 24-bit support).
- [x] Background telemetry daemon connection and socket health.
- [x] Active provider detection and data feed status.
- [x] `tmux.conf` configuration snippet integrity.

---

### Real-Time Live Preview
Preview how the status line will render in ANSI terminal output without opening tmux:

```bash
# Preview current active configuration
agentusage tmux preview

# Preview a specific preset
agentusage tmux preview --preset claude-focused
```

---

### JSON Metric Export
Dump the current rendered state and raw metrics for programmatic consumption:

```bash
agentusage tmux --json
```

---

### Background Usage Watcher (`agentusage tmux watch`)
Monitors spending spikes and fires tmux native message bar warnings or terminal bells when thresholds are breached:

```bash
# Start background watcher
agentusage tmux watch --interval 5s --alert-mode message
```
*Alert modes: `message` (tmux display-message), `bell` (terminal bell), `both`, or `none`.*

---

## 6. Troubleshooting & Incident Runbook

### Issue 1: Status Line Shows `?` Placeholder
**Symptom**: The tmux status bar displays `?` instead of metrics.

**Root Causes**:
1. Daemon is stopped or unresponsive.
2. Direct provider lookups timed out (> 800ms budget) and no anti-flicker cache exists.

**Resolution Steps**:
```bash
# Step 1: Check daemon status
agentusage telemetry daemon status

# Step 2: Restart telemetry daemon if down
agentusage telemetry daemon start

# Step 3: Run render manually with debug output to pinpoint the error
agentusage tmux --source direct
```

---

### Issue 2: Broken Characters / Question Mark Boxes for Logos
**Symptom**: Provider logos appear as `` or rectangle boxes.

**Root Causes**:
1. Bundled icon font is not installed.
2. Terminal does not have font fallback configured.

**Resolution Steps**:
```bash
# Step 1: Check font installation
agentusage tmux font status

# Step 2: Install and configure font fallback
agentusage tmux font install
agentusage tmux font setup

# Step 3: If running iTerm2 or non-fallback terminal, run patcher:
agentusage tmux font patch

# Step 4: Fallback to standard emoji or ASCII if fonts are restricted:
agentusage tmux --glyphs unicode
# or
agentusage tmux --glyphs ascii
```

---

### Issue 3: Status Segment Not Appearing in tmux
**Symptom**: `agentusage tmux install` completed, but status bar does not show the segment.

**Resolution Steps**:
```bash
# Step 1: Verify tmux.conf location and snippet presence
agentusage tmux doctor

# Step 2: Force tmux to reload configuration
tmux source-file ~/.config/tmux/tmux.conf
# or
tmux source-file ~/.tmux.conf

# Step 3: Check if status-right-length is too short (truncating the segment)
tmux set-option -g status-right-length 250
```

---

### Issue 4: Popup Dashboard Keybinding Fails
**Symptom**: Pressing `Prefix + u` produces an error or does nothing.

**Root Causes**:
1. tmux version is older than `3.2` (`display-popup` is unsupported).

**Resolution Steps**:
```bash
# Check tmux version
tmux -V

# If < 3.2, upgrade tmux via package manager:
# Ubuntu/Debian: sudo apt install tmux
# macOS: brew install tmux
```

---

## 7. Rollback & Uninstallation

To cleanly remove agentUsage from your tmux environment without affecting the rest of your tmux configuration:

```bash
# 1. Remove managed snippet from tmux.conf (creates .bak backup)
agentusage tmux uninstall

# 2. (Optional) Remove bundled icon font
agentusage tmux font uninstall

# 3. Reload tmux configuration
tmux source-file ~/.config/tmux/tmux.conf
```

### Manual Rollback via Backup
If you need to restore your previous configuration directly:
```bash
cp ~/.config/tmux/tmux.conf.bak ~/.config/tmux/tmux.conf
tmux source-file ~/.config/tmux/tmux.conf
```
