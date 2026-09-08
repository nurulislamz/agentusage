# Research report — agentUsage

> **DO NOT MERGE.** Living research dump. Routines append here.

Project: terminal-first local dashboard for AI coding-agent / IDE / LLM API usage, spend, quotas, and rate limits. Auto-detects tools and keys. Public repo: https://github.com/nurulislamz/agentusage

## Competitors

### Seeded 2026-09-04 (v2)

**Primary peer risk:** [OpenUsage.sh](https://openusage.sh) markets nearly the same thesis (local terminal multi-tool quotas/spend/auto-detect). Differentiate on coverage, UX, or packaging — or collide.

| Product | Layer | What they do | Vs agentUsage |
|---|---|---|---|
| [OpenUsage.sh](https://openusage.sh) ([GitHub](https://github.com/janekbaraniewski/openusage)) | Local multi-tool TUI | Auto-detect agents/keys, quotas, spend, rate limits, burn rate, statusline; 35+ tools. | **Nearest twin.** Treat as #1 peer. |
| [ccusage](https://github.com/ryoppippi/ccusage) | Local CLI reports | Daily/weekly/monthly/session/blocks from agent logs (Claude, Codex, OpenCode, Goose, Copilot CLI, Gemini, …). | Strong OSS peer on history/reports; thinner live quota dashboard. |
| ccusage UI dashboards / claude-monitor | Local UI/TUI | HTML or live burn-rate for Claude Code. | Claude-centric niche. |
| Cursor Spending tab + Admin API | Vendor | Plan pools, spend limits; Enterprise team spend APIs. | Cursor only. |
| [Cursor Spend Tracker](https://marketplace.visualstudio.com/items?itemName=helper2424.cursor-spend-tracker) | Extension | Status-bar spend via Admin API. | Cursor team only. |
| [cursor-usage-tracker](https://github.com/ofershap/cursor-usage-tracker) | Self-host team | Cursor Enterprise spend + anomaly Slack alerts. | FinOps for Cursor orgs, not local multi-provider autodetection. |
| Claude Code `/usage` `/cost`, claude.ai Usage | Vendor | Session/plan bars, 5h blocks. | Anthropic only. |
| OpenRouter Activity | Vendor / gateway | Credits and model cost analytics. | Hosted gateway billing. |
| [Helicone](https://www.helicone.ai), [Langfuse](https://langfuse.com) | Cloud/proxy observability | Traces, sessions, cost for instrumented apps. | Wrong layer for raw CLI agent logs; complementary if exporting OTel later. |
| [LiteLLM](https://docs.litellm.ai) proxy spend | Gateway | Virtual keys, budgets, admin UI. | Org gateway control, not personal agent detector. |
| Vantage Cursor reports / Lineman.io | FinOps / team SaaS | Cloud spend or Claude team visibility. | Enterprise; not terminal-first local. |

**Positioning**
- Narrative: **one pane for every agent on the machine**.
- Beat ccusage on live quotas/rate-limit probing + dashboard UX.
- Explicitly differentiate from OpenUsage.sh (providers, UX, install path) — same category.

Sources: openusage.sh; github.com/ryoppippi/ccusage; Cursor/Anthropic docs; Helicone/Langfuse/LiteLLM docs.

### Update 2026-09-04 (weekday scrape)

OpenUsage and ccusage both moved; twin-risk is sharper.

| Product | Delta | Vs agentUsage |
|---|---|---|
| [OpenUsage.sh](https://openusage.sh) | Now markets **34–36 providers**, local SQLite daemon, tmux status-bar + Claude Code statusline installers, headless daily/weekly/monthly/session/billing-block reports, Prometheus hub, explicit “vs Langfuse/Helicone” docs, brew/curl/go install. Coverage spans agents (Claude, Codex, Cursor, Copilot, Gemini, OpenCode, Ollama, Amp, Goose, Roo, Kilo, Kiro, Zed, …) + API platforms (OpenRouter, Groq, Mistral, DeepSeek, xAI, …). | Category definition is almost 1:1 with agentUsage. Must pick a wedge: deeper auto-detect reliability, better UX, niche providers, or packaging — not “we also do local quotas.” |
| [ccusage](https://github.com/ryoppippi/ccusage) ([ccusage org mirror](https://github.com/ccusage/ccusage)) | Unified multi-source CLI; agents now include Amp, Droid, Codebuff, Hermes, pi, Goose, OpenClaw, Kilo, Kimi, Qwen, Copilot CLI, Gemini, plus newer **Antigravity / Grok Build / ZCode**. `blocks` for Claude 5h windows; LiteLLM-backed pricing; Nix sandbox pricing lock. | Remains the best **log-history / cost-report** OSS peer. Still thinner on live quota/rate-limit TUI + key autodetection. |
| OpenUsage docs vs observability | OpenUsage publishes a comparison framing Langfuse/Helicone as *hosted app observability*, itself as *local quota tracker*. | Keep Helicone/Langfuse/LiteLLM in “adjacent layer” — don’t pretend they’re desktop twins. |
| Vendor UIs | Cursor Spending / Claude `/usage`/`/cost` / OpenRouter Activity unchanged as single-vendor panes. | Still the fragmentation OpenUsage/agentUsage claim to fix. |

**Positioning refresh**
- Treat **OpenUsage.sh as the primary collision risk** (same install story, same auto-detect thesis).
- Beat **ccusage** on live quotas + multi-provider dashboard, not on historical report breadth (they’re expanding fast).
- Keep cloud observability (Helicone/Langfuse) and gateway spend (LiteLLM) as complementary, not competitors for the terminal autodetection use case.

Sources: [openusage.sh](https://openusage.sh/); [openusage.sh/llms.txt](https://openusage.sh/llms.txt); [github.com/janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage); [github.com/ryoppippi/ccusage](https://github.com/ryoppippi/ccusage); [github.com/ccusage/ccusage](https://github.com/ccusage/ccusage); [openusage local-quota guide](https://openusage.sh/local-quota-tracker-for-claude-code-codex-cursor/).


### Update 2026-09-07

Fresh weekend scrape. Twin-risk with OpenUsage remains the headline; Splitrail is the clearest secondary local peer moving.

| Product | Delta | Vs agentUsage |
|---|---|---|
| [OpenUsage.sh](https://openusage.sh/) ([GitHub](https://github.com/janekbaraniewski/openusage), ~187★) | Still markets **35 providers**, local SQLite daemon, tmux + Claude Code statusline, headless reports, Prometheus hub, auto-detect tools/keys. Newer docs emphasize **cost-attribution hooks** (`openusage integrations install` for Claude Code / Codex / OpenCode) to get per-turn/project rows beyond poll totals. Updated ~2026-09-05. | **Nearest twin** — category definition still ~1:1. Wedge must be reliability, UX, niche providers, or packaging — not “also local quotas.” |
| [ccusage](https://github.com/ryoppippi/ccusage) (~18.4k★) | Still the category CLI for historical daily/weekly/monthly/session/`blocks` reports across many agents (Claude, Codex, OpenCode, Amp, Droid, Hermes, pi, Goose, OpenClaw, Kilo, Kimi, Qwen, Copilot CLI, Gemini, …). Active today (updated 2026-09-07). | Best **log-history / cost-report** OSS peer; thinner live quota/rate-limit TUI + key autodetection. |
| [Splitrail](https://github.com/Piebald-AI/splitrail) (~219★) | Real-time cross-agent token/cost monitor; agent list now includes Antigravity, Zoo Code, Cline/Roo/Kilo, Copilot VS Code + CLI, OpenCode, Pi, Piebald. Optional **Splitrail Cloud** sync + **MCP server** (`get_daily_stats`, `get_cost_breakdown`, …). Updated ~2026-09-06. | Stronger secondary local peer than menu-bar wrappers; more “live monitor” than OpenUsage’s full quota dashboard thesis, but closing the gap. |
| Vendor UIs | Cursor Spending / Claude `/usage`/`/cost` / OpenRouter Activity unchanged as single-vendor panes. | Fragmentation OpenUsage/agentUsage claim to fix. |
| Helicone / Langfuse / LiteLLM | Still hosted observability / gateway spend layers. | Adjacent, not desktop twins — keep that framing. |

**Positioning refresh**
- **OpenUsage.sh** = primary collision (same install story + auto-detect + live quotas).
- **Splitrail** = rising real-time multi-agent monitor (+ MCP/cloud option) — watch as #2 local peer.
- **ccusage** = report/parser standard to stay compatible with; don’t try to out-breadth their historical CLI alone.
- Differentiate on live quota/rate-limit probing + auto-detect reliability + dashboard UX across *every* agent on the machine.

Sources: [openusage.sh](https://openusage.sh/); [openusage cost attribution](https://openusage.sh/docs/guides/cost-attribution/); [github.com/janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage); [github.com/ryoppippi/ccusage](https://github.com/ryoppippi/ccusage); [github.com/Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail); GitHub API star counts 2026-09-07.

### Update 2026-09-08

Fresh weekday scrape. Twin-risk with OpenUsage.sh remains the headline; docs now claim a wider provider set + Antigravity hooks, and a new gateway FinOps name enters the adjacent layer.

| Product | Delta | Vs agentUsage |
|---|---|---|
| [OpenUsage.sh](https://openusage.sh/) ([GitHub](https://github.com/janekbaraniewski/openusage), ~187★) | Homepage still markets ~34–35 providers + local SQLite daemon + tmux/Claude statusline + auto-detect. Docs now say **36 providers** and list opt-in hooks for Claude Code / Codex / OpenCode / **Antigravity**. Last push ~2026-09-07. | **Nearest twin** — category definition still ~1:1. Wedge must be reliability, UX, niche providers, or packaging. |
| [ccusage](https://github.com/ryoppippi/ccusage) / [ccusage.com](https://ccusage.com/) (~18.4k★) | Still the category CLI for historical daily/weekly/monthly/session/`blocks` across many agents. **Pushed today** (2026-09-08). | Best **log-history / cost-report** OSS peer; thinner live quota/rate-limit TUI + key autodetection. |
| [Splitrail](https://github.com/Piebald-AI/splitrail) (~219★) | Real-time cross-agent monitor + optional Cloud + MCP (`get_daily_stats`, `get_cost_breakdown`, …). No material star/push move since weekend note (~pushed 2026-09-06). | #2 local “live monitor” peer. |
| [openusage.ai](https://github.com/robinebers/openusage) (~4.0k★) | Different product (macOS menu-bar). Active Antigravity local spend/history work (reads `~/.gemini/antigravity-cli/conversations/*.db`). | Naming collision only — do not conflate with openusage.sh. |
| [AgentCost](https://agentcost.in/docs/guides/agentcost-integration-into-existing-projects/) | New-to-dump **instrumented** FinOps: OpenAI-compatible gateway proxy, SDK decorators, LangChain/CrewAI callbacks, OTel/Prometheus export, budgets/caching. | Adjacent **app/gateway** layer (like Helicone/Langfuse/LiteLLM) — not a local multi-agent auto-detect TUI. |
| Vendor UIs | Cursor Spending / Claude `/usage`/`/cost` / OpenRouter Activity still single-vendor panes. | Fragmentation OpenUsage/agentUsage claim to fix. |

**Positioning refresh**
- **OpenUsage.sh** = primary collision (now explicitly Antigravity + 36-provider docs).
- **Splitrail** = rising real-time multi-agent monitor (+ MCP/cloud).
- **ccusage** = report/parser standard (still shipping hard).
- Keep Helicone / Langfuse / LiteLLM / **AgentCost** as complementary observability/gateway spend — not desktop twins.

Sources: [openusage.sh](https://openusage.sh/); [openusage.sh/docs](https://openusage.sh/docs/); [github.com/janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage); [github.com/ryoppippi/ccusage](https://github.com/ryoppippi/ccusage); [ccusage.com](https://ccusage.com/); [github.com/Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail); [github.com/robinebers/openusage](https://github.com/robinebers/openusage); [agentcost.in integration guide](https://agentcost.in/docs/guides/agentcost-integration-into-existing-projects/); GitHub API star counts 2026-09-08.

## Open source

### Seeded 2026-09-04

Peers and libraries for a local multi-agent usage/spend/quota dashboard: **log parsers**, **live TUIs**, **gateway cost**, **OTel observability**, **Cursor/Claude/Codex parsers**.

| Repo | Stars | License | Why it matters |
|---|---:|---|---|
| [ccusage/ccusage](https://github.com/ccusage/ccusage) | ~18k | (check / npm) | Category-defining CLI: daily/session/blocks cost from local agent JSONL (Claude, Codex, Copilot CLI, Gemini, …). Primary report-format peer. |
| [janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage) | ~186 | MIT | Local multi-tool quota/spend TUI + SQLite — **nearest product twin** (same thesis as agentUsage). |
| [Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail) | ~217 | MIT | Cross-platform real-time token/cost monitor across many coding agents. |
| [cobra91/better-ccusage](https://github.com/cobra91/better-ccusage) | ~81 | MIT | Faster multi-provider ccusage-style analyzer — watch for format forks. |
| [ofershap/cursor-usage-tracker](https://github.com/ofershap/cursor-usage-tracker) | ~33 | MIT | Self-hosted Cursor Enterprise spend + anomaly alerts — team FinOps, not local autodetection. |
| [BerriAI/litellm](https://github.com/BerriAI/litellm) | ~58k | (check) | AI gateway with spend tracking / virtual keys — pricing tables useful; wrong layer for personal agent logs. |
| [langfuse/langfuse](https://github.com/langfuse/langfuse) | ~34k | (check) | OSS LLM observability + OTel — complementary if exporting traces later. |
| [Helicone/helicone](https://github.com/Helicone/helicone) | ~6.1k | Apache-2.0 | OSS LLM observability proxy — same adjacent layer as Langfuse. |

**Also watch (ccusage ecosystem)**
- [sculptdotfun/viberank](https://github.com/sculptdotfun/viberank) (~115★, MIT) — public leaderboard on ccusage data.
- [851-labs/tokenmaxxing](https://github.com/851-labs/tokenmaxxing) (~68★, MIT) — sync personal ccusage with peers.
- [Nihondo/AgentLimits](https://github.com/Nihondo/AgentLimits), [kenn-io/vibepulse](https://github.com/kenn-io/vibepulse), [goniszewski/cctray](https://github.com/goniszewski/cctray), [sivchari/ccowl](https://github.com/sivchari/ccowl) — macOS menu-bar / widget wrappers around Claude/Codex usage.
- [atomchung/ccstory](https://github.com/atomchung/ccstory) (~43★) — narrative recap on top of ccusage bills.

**Build takeaways**
1. Treat **OpenUsage** as the collision product; **ccusage** as the report/parser standard to stay compatible with.
2. Differentiate on live quota probing + auto-detect reliability + dashboard UX, not raw historical report breadth.
3. Reuse LiteLLM pricing ideas; don't become a hosted OTel stack (Langfuse/Helicone stay adjacent).
4. Cursor Admin API / vendor `/usage` endpoints remain single-vendor — agentUsage's job is unification.

Sources: GitHub search/API 2026-09-04 (stars approximate).

### Update 2026-09-07

Fresh GitHub pass. Important naming split: **openusage.ai** (macOS menu bar) ≠ **openusage.sh** (local multi-tool TUI twin). Ecosystem around ccusage keep spawning wrappers.

| Repo | Stars | License | Delta / why it matters |
|---|---:|---|---|
| [ccusage/ccusage](https://github.com/ccusage/ccusage) | ~18.4k | (npm / check) | Still the category CLI for historical daily/session/`blocks` reports. Pushed **today** (2026-09-07). Primary report-format peer. |
| [janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage) | ~187 | MIT | **openusage.sh** — local multi-tool quota/spend + SQLite. Nearest *product* twin to agentUsage. Pushed 2026-09-07. |
| [robinebers/openusage](https://github.com/robinebers/openusage) | ~4.0k | MIT | **openusage.ai** — popular macOS menu-bar subscription/usage tracker (different product). Do not conflate with openusage.sh. |
| [Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail) | ~219 | MIT | Real-time cross-agent token/cost monitor (+ optional cloud/MCP). Strong #2 local peer. |
| [Halloweedev/usagepal](https://github.com/Halloweedev/usagepal) | ~76 | MIT | Open-source fork of OpenUsage (menu-bar lineage); Rust + Tauri; active 2026-09-07. |
| [cobra91/better-ccusage](https://github.com/cobra91/better-ccusage) | ~84 | MIT | Faster multi-provider ccusage-style analyzer. |
| [kenn-io/vibepulse](https://github.com/kenn-io/vibepulse) | ~53 | MIT | macOS menubar on ccusage (Claude + Codex). |
| [Nihondo/AgentLimits](https://github.com/Nihondo/AgentLimits) | ~55 | MIT | macOS widgets for Codex/Claude limits + ccusage heatmap; pushed 2026-09-05. |
| [itvincent-git/codex-usage-desktop](https://github.com/itvincent-git/codex-usage-desktop) | ~27 | MIT | Local-first Codex quota/desktop analytics (macOS + Windows). |
| [agiwhitelist/tokdiet](https://github.com/agiwhitelist/tokdiet) | ~33 | MIT | Local reverse-proxy meter between agents and APIs + live dashboard — interesting *instrumentation* approach vs log parsing. |
| [sculptdotfun/viberank](https://github.com/sculptdotfun/viberank) / [851-labs/tokenmaxxing](https://github.com/851-labs/tokenmaxxing) | ~115 / ~69 | MIT | Social/leaderboard layers on ccusage data. |
| [BerriAI/litellm](https://github.com/BerriAI/litellm) | ~58k | (check) | Gateway spend/pricing tables — adjacent layer. |
| [langfuse/langfuse](https://github.com/langfuse/langfuse) / [Helicone/helicone](https://github.com/Helicone/helicone) | ~34k / ~6.1k | (check) / Apache-2.0 | Hosted OTel observability — complementary, not desktop twins. |
| [Portkey-AI/gateway](https://github.com/Portkey-AI/gateway) | ~13k | MIT | Another AI gateway with cost/guardrails — same adjacent layer. |

**Also watch**
- [ofershap/cursor-usage-tracker](https://github.com/ofershap/cursor-usage-tracker) (~33★) — Cursor Enterprise FinOps, not personal autodetection.
- [atomchung/ccstory](https://github.com/atomchung/ccstory), [goniszewski/cctray](https://github.com/goniszewski/cctray), [sivchari/ccowl](https://github.com/sivchari/ccowl) — narrative / menu-bar wrappers around Claude usage.
- Windows ports of the menu-bar OpenUsage line (`ItsJazii/pane`, `mesomya/openusage-windows`) — packaging forks, not category innovators.

**Build takeaways refresh**
1. Name carefully in copy: **openusage.sh** (janekbaraniewski) is the collision twin; **openusage.ai** (robinebers) is the larger menu-bar brand.
2. Stay compatible with **ccusage** report formats; differentiate on live quotas + multi-provider auto-detect + dashboard UX.
3. Watch Splitrail as rising real-time peer; tokdiet as an alternate metering architecture (proxy vs parse).
4. LiteLLM/Langfuse/Helicone/Portkey stay gateway/observability-adjacent.

Sources: GitHub search/API 2026-09-07 (stars approximate).

### Update 2026-09-08

| Repo | Stars | Delta / why it matters |
|---|---:|---|
| [ccusage/ccusage](https://github.com/ccusage/ccusage) | ~18.4k | Pushed **today** (2026-09-08) — report-format peer still racing. |
| [janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage) | ~187 | openusage.sh twin; last push ~2026-09-07. |
| [Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail) | ~219 | Real-time #2 local peer. |
| [robinebers/openusage](https://github.com/robinebers/openusage) | ~4.0k | openusage.ai menu-bar — naming collision only. |
| [Halloweedev/usagepal](https://github.com/Halloweedev/usagepal) | ~76 | OpenUsage menu-bar fork lineage; quiet since 2026-09-07. |

**Also watch:** AgentCost (hosted/gateway FinOps docs at agentcost.in) — adjacent layer, not a GitHub twin to agentUsage’s local autodetection thesis.

Sources: GitHub API 2026-09-08.
