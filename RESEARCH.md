# Research report — agentUsage

> **DO NOT MERGE.** Living research dump. Routines append here.

Project: terminal-first local dashboard for AI coding-agent / IDE / LLM API usage, spend, quotas, and rate limits. Auto-detects tools and keys. Public repo: https://github.com/nurulislamz/agentusage

## Competitors

### Seeded 2026-09-04

agentUsage's wedge is **one local TUI/dashboard across many coding agents + API keys**, zero config. Competitors are either single-vendor (Claude-only, Cursor-only) or cloud observability for apps you instrument yourself.

| Product | What they do | Vs agentUsage |
|---|---|---|
| [ccusage](https://github.com/ryoppippi/ccusage) | Popular local CLI: daily/weekly/monthly/session reports from Claude Code, Codex, OpenCode, Goose, Copilot CLI, Gemini CLI, and more. Blocks report for Claude 5h windows. | Direct OSS peer. agentUsage differentiates on broader auto-detect (Cursor SQLite, IDE agents), richer dashboard UX, quotas/rate-limit probing. |
| Claude Code `/usage` | Built-in session spend estimate from local tokens × list rates. | Free but session-scoped; resets on `/clear`; no multi-tool view. |
| Anthropic Console / admin analytics | Org billing, CSV, Enterprise analytics API. | Cloud/org only; not a local multi-agent TUI. |
| Cursor usage UI + Admin API | In-product spend; Enterprise `/teams/spend`, filtered events, analytics. | First-party Cursor only. |
| [Cursor Spend Tracker](https://marketplace.visualstudio.com/items?itemName=helper2424.cursor-spend-tracker) | VS Code/Cursor status-bar spend via Admin API. | Cursor-team only; needs admin key. |
| [ofershap/cursor-usage-tracker](https://github.com/ofershap/cursor-usage-tracker) | Self-hosted Cursor Enterprise spend + anomaly Slack alerts. | Team/finance tool, not local multi-provider autodetection. |
| [Vantage](https://www.vantage.sh/) Cursor cost reports | FinOps cloud spend including Cursor. | Enterprise FinOps; not terminal-first local. |
| [Helicone](https://www.helicone.ai/), [Langfuse](https://langfuse.com/) | LLM observability / tracing for apps you instrument. | Wrong layer for CLI agent logs (Langfuse wants spans; Claude Code emits metrics). Useful if agentUsage later exports OTel. |
| LiteLLM proxy spend / OpenRouter dashboard | Gateway usage and credits for routed API traffic. | Covers API platforms agentUsage already tracks; not coding-agent local logs. |
| Lineman.io (sponsors ccusage) | Team Claude Code spend visibility / optimization. | Team SaaS for Claude; not open multi-agent local. |

**Positioning notes**
- Own the narrative: **one pane for every agent on the machine** (ccusage is the main OSS rival to beat on coverage + UX).
- Cursor + Claude first-party UIs are incomplete by design — use that in marketing.
- Don't pretend to replace Helicone/Langfuse; complementary for app builders.

Sources: github.com/ryoppippi/ccusage; ssdnodes Claude spend tools; vantage.sh Cursor costs; marketplace Cursor Spend Tracker; github.com/ofershap/cursor-usage-tracker.

## Open source

_No OSS notes yet. The OSS-watch routine fills this._
