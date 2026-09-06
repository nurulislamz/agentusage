# Telemetry Integrations

This repository supports three native coding-agent telemetry streams:

1. OpenCode plugin hooks
2. Codex `notify` hook
3. Claude Code command hooks

Antigravity quota is polled directly by the Antigravity provider (OAuth token +
`retrieveUserQuotaSummary`) and does not use a status-line integration hook.

All streams emit normalized telemetry events into the same SQLite store:

- `~/.local/state/agentusage/telemetry.db`

When the agentUsage app is running, background collection and canonical telemetry read-model updates are automatic.
You do not need to run manual collection commands for normal operation.
agentUsage does not auto-create synthetic providers from telemetry. Unmapped telemetry provider IDs are flagged for explicit user action.

## Managing Integrations

All integration hook/plugin definitions are embedded in the `agentusage` binary.
You can view status, install, and update integrations directly from the terminal UI:

1. Open `agentusage`.
2. Press `,` (or `Shift+S`) to open the Settings modal.
3. Navigate to the **Integrations** tab to view detected tools and install/update hooks.

Hooks can also ingest events directly via the background daemon:

```bash
agentusage daemon hook claude_code < /tmp/turn.json
agentusage daemon hook codex '{"type":"agent-turn-complete"}'
agentusage daemon hook opencode < /tmp/opencode-hook.json
```

The daemon also prints a hint at startup when it detects tools with missing integrations.

## What Gets Installed

### OpenCode (Plugin)

- `~/.config/opencode/plugins/agentusage-telemetry.ts`
- plugin entry in `~/.config/opencode/opencode.json`

### Codex (Notify Hook)

- `~/.config/agentusage/hooks/codex-notify.sh`
- `notify = ["~/.config/agentusage/hooks/codex-notify.sh"]` in `~/.codex/config.toml`

### Claude Code (Command Hooks)

- `~/.config/agentusage/hooks/claude-hook.sh`
- command hooks in `~/.claude/settings.json` for:
  - `Stop`
  - `SubagentStop`
  - `PostToolUse`

### Antigravity CLI (API poll)

- agentUsage reads `antigravity-oauth-token` under each box config dir
  (`~/.agy-containers/<box>/.gemini/antigravity-cli/` or `~/.gemini/antigravity-cli/`).
- The daemon polls `https://daily-cloudcode-pa.googleapis.com/v1internal:retrieveUserQuotaSummary`
  with `User-Agent: antigravity`.
- Expired access tokens renew directly over HTTP when refresh credentials and client configuration are present.
  agentUsage never executes `agy`, `agy-box`, or any background prompt to refresh credentials. When refresh
  is unavailable, an actionable sign-in notice is presented for the affected account while other accounts continue updating.

## Provider Linking (Explicit Control)

Telemetry events are tagged with whatever `provider_id` the source tool uses. When that id doesn't match any configured account, agentUsage attempts a link via `telemetry.provider_links`, then falls back to flagging the source as unmapped.

### Built-in defaults

The following links are applied automatically and cover known rename mismatches between source-tool vocabulary and openusage's internal provider ids:

| Source provider id | Mapped to    | Why                                                    |
|--------------------|--------------|--------------------------------------------------------|
| `anthropic`        | `claude_code`| OpenCode/Codex/Claude Code emit `anthropic`            |
| `google`           | `gemini_api` | OpenCode emits `google` for the Gemini API             |
| `github-copilot`   | `copilot`    | OpenCode emits `github-copilot` for GitHub Copilot     |

Identity links (e.g. `openai` → `openai`) are intentionally not enumerated — direct id matches are handled by the matcher without a link.

### User overrides

Add custom or override entries in `~/.config/agentusage/settings.json`:

```json
{
  "telemetry": {
    "provider_links": {
      "google": "my-personal-gemini-account",
      "moonshot": "kimi"
    }
  }
}
```

User entries take precedence over defaults. The daemon picks up changes on the next poll cycle (no restart needed).

### Interactive remap

Open the TUI Settings modal (`s`), navigate to **6 TELEM**. Unmapped telemetry sources are listed below the time-window picker, each with a category badge:

- `[no account configured]` — no agentUsage account exists for this source.
- `[suggested: <id>]` — a configured provider id whose name overlaps with the source. Press `m` to open a picker pre-selecting the suggestion.
- `[mapped → <id>, target not configured]` — a link points to an id that has no account. Resolve by changing the link target or creating the missing account.

Keybindings on each unmapped row:

- `m` (or Enter) — open a target picker showing all configured provider ids; Enter to apply, Esc to cancel.
- `x` — clear an existing user-defined link for this source (built-in defaults can't be cleared this way; override them with a different target instead).

### Diagnostics emitted on snapshots

When at least one source is unmapped, every snapshot picks up two diagnostic keys:

- `telemetry_unmapped_providers` — comma-separated list of unmapped source ids.
- `telemetry_unmapped_meta` — comma-separated `<source>=<category>[:<suggestion-or-target>]` entries. Categories: `unconfigured`, `mapped_target_missing`. The optional suffix is a configured provider id suggestion (for `unconfigured`) or the link's target id (for `mapped_target_missing`).

### Behavior summary

1. No automatic telemetry-only providers are created — sources without a configured account stay flagged.
2. Canonical telemetry usage metrics are applied only to configured providers or explicitly linked providers.
3. Built-in defaults can be overridden but not erased; setting `provider_links.<source>` replaces the default for that source.

## Optional runtime env vars (all integrations)

- `AGENTUSAGE_TELEMETRY_ENABLED=true|false`
- `AGENTUSAGE_BIN=/absolute/path/to/agentusage`
- `AGENTUSAGE_TELEMETRY_ACCOUNT_ID=<logical account override>`
- `AGENTUSAGE_TELEMETRY_DB_PATH=/path/to/telemetry.db`
- `AGENTUSAGE_TELEMETRY_SPOOL_DIR=/path/to/spool`
- `AGENTUSAGE_TELEMETRY_SPOOL_ONLY=true|false`
- `AGENTUSAGE_TELEMETRY_VERBOSE=true|false`

## Verify Ingestion

OpenCode:

```bash
sqlite3 ~/.local/state/agentusage/telemetry.db "select r.source_system, r.source_channel, e.event_type, count(*) from usage_events e join usage_raw_events r on r.raw_event_id=e.raw_event_id where r.source_system='opencode' group by 1,2,3 order by 1,2,3;"
```

Codex:

```bash
sqlite3 ~/.local/state/agentusage/telemetry.db "select r.source_system, r.source_channel, e.event_type, count(*) from usage_events e join usage_raw_events r on r.raw_event_id=e.raw_event_id where r.source_system='codex' group by 1,2,3 order by 1,2,3;"
```

Claude Code:

```bash
sqlite3 ~/.local/state/agentusage/telemetry.db "select r.source_system, r.source_channel, e.event_type, count(*) from usage_events e join usage_raw_events r on r.raw_event_id=e.raw_event_id where r.source_system='claude_code' group by 1,2,3 order by 1,2,3;"
```

Antigravity:

```bash
sqlite3 ~/.local/state/agentusage/telemetry.db "select r.source_system, r.source_channel, e.event_type, count(*) from usage_events e join usage_raw_events r on r.raw_event_id=e.raw_event_id where r.source_system='antigravity' group by 1,2,3 order by 1,2,3;"
```

Inspect latest canonical metrics:

```bash
sqlite3 ~/.local/state/agentusage/telemetry.db <<'SQL'
select
  e.occurred_at,
  r.source_system,
  r.source_channel,
  e.event_type,
  e.provider_id,
  e.account_id,
  e.model_raw,
  e.input_tokens,
  e.output_tokens,
  e.reasoning_tokens,
  e.cache_read_tokens,
  e.cache_write_tokens,
  e.total_tokens,
  e.cost_usd,
  e.requests,
  e.session_id,
  e.turn_id,
  e.message_id,
  e.tool_call_id,
  e.tool_name
from usage_events e
join usage_raw_events r on r.raw_event_id = e.raw_event_id
order by e.occurred_at desc
limit 100;
SQL
```
