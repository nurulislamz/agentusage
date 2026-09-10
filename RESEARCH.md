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

### Update 2026-09-09

Fresh Wednesday scrape (Europe/London). Twin-risk with OpenUsage.sh unchanged; **juliantanx/aiusage** enters as a new local multi-tool dashboard peer (name collision with agentUsage); Splitrail and ccusage both pushed overnight; LiteLLM keeps shipping gateway spend features.

| Product | Delta | Vs agentUsage |
|---|---|---|
| [OpenUsage.sh](https://openusage.sh/) ([GitHub](https://github.com/janekbaraniewski/openusage), **~188★**, +1 vs yesterday) | Homepage still markets **34 providers** + local SQLite daemon + tmux/Claude statusline + auto-detect. Docs still say **36 providers** + Antigravity hooks. Last push **2026-09-08 ~20:03 BST**. | **Nearest twin** — category definition still ~1:1. Star creep is slow; product surface stable day-over-day. |
| [ccusage](https://github.com/ccusage/ccusage) / [ccusage.com](https://ccusage.com/) (**~18.4k★** / 18,444) | Still the category CLI for historical daily/weekly/monthly/session/`blocks` (+ Antigravity / Grok Build / ZCode). **Pushed today** (2026-09-09 ~07:38 BST). | Best **log-history / cost-report** OSS peer; thinner live quota/rate-limit TUI + key autodetection. |
| [Splitrail](https://github.com/Piebald-AI/splitrail) (**~219★**, flat) | Real-time cross-agent monitor + optional Cloud + MCP. **Pushed overnight** (2026-09-09 ~01:33 BST) after weekend quiet — activity resumed, stars unchanged. | #2 local “live monitor” peer. |
| [AIUsage](https://github.com/juliantanx/aiusage) (**~120★**) — **new-to-Competitors dump** | Local-first web dashboard (`aiusage serve` → localhost:3847) for tokens/cost/sessions/models/projects/tool calls/quotas across **20+** coding tools; optional GitHub/S3/R2 sync + public leaderboard + desktop widget. Pushed 2026-09-08. | **Rising local peer** closer to OpenUsage/agentUsage than menu-bar wrappers. Also a **product-name collision** risk (AIUsage vs agentUsage) — keep branding distinct. |
| [openusage.ai](https://github.com/robinebers/openusage) (**~4.1k★** / 4,076) | macOS menu-bar tracker (different product). No push since 2026-09-06; stars +~76 vs yesterday’s ~4.0k note. | Naming collision only — do not conflate with openusage.sh. |
| Vendor UIs | Cursor Spending / Claude `/usage`/`/cost` / OpenRouter Activity still single-vendor panes. | Fragmentation OpenUsage/agentUsage claim to fix. |
| Helicone / Langfuse / LiteLLM | **LiteLLM** (~58.3k★, pushed today): spend-by-tool / `LiteLLM_DailyToolSpend` rollups, credential usage tags, spend-log retention — gateway FinOps still moving. **Langfuse** (~34.4k★, pushed today) active OTel observability. **Helicone** (~6.1k★) quiet since ~2026-08-31. | Adjacent layers, not desktop twins. |
| Also watch (local TUI / wrappers) | [SophanaSok/ai-usage-tui](https://github.com/SophanaSok/ai-usage-tui) (0★, **pushed today**) — Rust TUI for OpenCode/Claude/Codex/Ollama costs + budgets; [paperwave/codeburn](https://github.com/paperwave/codeburn) (0★, stale since Apr) — Claude/Codex/Cursor TUI; [Halloweedev/usagepal](https://github.com/Halloweedev/usagepal) (76★, pushed today); [ItsJazii/pane](https://github.com/ItsJazii/pane) (41★, pushed today) Windows tray OpenUsage port. | Early/niche — track, don’t treat as category leaders yet. |

**Positioning refresh**
- **OpenUsage.sh** = primary collision (34 homepage / 36 docs; +1★ overnight).
- **juliantanx/aiusage** = new local multi-tool dashboard peer + **name collision** to watch in copy/SEO.
- **Splitrail** = #2 live local peer (push activity resumed; stars flat).
- **ccusage** = report/parser standard (still shipping hard today).
- Keep Helicone / Langfuse / LiteLLM / AgentCost as complementary observability/gateway spend — LiteLLM’s tool-spend rollups are the notable adjacent move this week.

Sources: [openusage.sh](https://openusage.sh/); [openusage.sh/docs](https://openusage.sh/docs/); [github.com/janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage); [github.com/ccusage/ccusage](https://github.com/ccusage/ccusage); [ccusage.com](https://ccusage.com/); [github.com/Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail); [github.com/juliantanx/aiusage](https://github.com/juliantanx/aiusage); [github.com/robinebers/openusage](https://github.com/robinebers/openusage); [github.com/SophanaSok/ai-usage-tui](https://github.com/SophanaSok/ai-usage-tui); LiteLLM docs (spend logs / credential usage); WebSearch 2026-09-09; GitHub API star/push metadata 2026-09-09 (Europe/London).



### Update 2026-09-10

Fresh Thursday scrape (Europe/London). Twin-risk with OpenUsage.sh unchanged; star/push creep on the local peers, and openusage.ai (menu-bar) keeps shipping overnight:

| Product | Delta | Vs agentUsage |
|---|---|---|
| [OpenUsage.sh](https://openusage.sh/) ([GitHub](https://github.com/janekbaraniewski/openusage), **~190★**, +2 vs yesterday) | Homepage still **34 providers** + local SQLite daemon + tmux/Claude statusline + auto-detect. Docs still **36 providers** + Antigravity hooks. Last pushes overnight were docs-site dep bumps (2026-09-09 ~23:31–23:39 UTC) — product surface stable. | **Nearest twin** — category definition still ~1:1. |
| [ccusage](https://github.com/ccusage/ccusage) / [ccusage.com](https://ccusage.com/) (**~18.5k★** / 18,469; +~25 vs yesterday) | Still the category CLI for historical daily/weekly/monthly/session/`blocks`. **Pushed today** (2026-09-10 ~01:26 UTC): LiteLLM pricing snapshot chore. | Best **log-history / cost-report** OSS peer; thinner live quota/rate-limit TUI + key autodetection. |
| [Splitrail](https://github.com/Piebald-AI/splitrail) (**~220★**, +1) | Real-time cross-agent monitor + optional Cloud + MCP. No new push since 2026-09-09 ~00:33 UTC; slow star creep. | #2 local “live monitor” peer. |
| [AIUsage](https://github.com/juliantanx/aiusage) (**~122★**, +2) | Local web dashboard (`aiusage serve` → localhost:3847) across 20+ tools. Quiet since 2026-09-08 push. | Rising local peer + **name collision** (AIUsage vs agentUsage) — keep branding distinct. |
| [openusage.ai](https://github.com/robinebers/openusage) (**~4.1k★** / 4,090; +~14) | macOS menu-bar tracker (different product). **Pushed today** (2026-09-10 ~05:32 UTC). | Naming collision only — do not conflate with openusage.sh. |
| Vendor UIs | Cursor Spending / Claude `/usage`/`/cost` / OpenRouter Activity still single-vendor panes. | Fragmentation OpenUsage/agentUsage claim to fix. |
| Helicone / Langfuse / LiteLLM | **LiteLLM** (~58.4k★ / 58,426, pushed today) and **Langfuse** (~34.4k★ / 34,427, pushed today) still active gateway/OTel FinOps. **Helicone** (~6.1k★) still quiet since ~2026-08-31. | Adjacent layers, not desktop twins. |
| Also watch | [ItsJazii/pane](https://github.com/ItsJazii/pane) (42★, +1) Windows tray OpenUsage port; [Halloweedev/usagepal](https://github.com/Halloweedev/usagepal) (76★); [SophanaSok/ai-usage-tui](https://github.com/SophanaSok/ai-usage-tui) (0★, last push 2026-09-09). | Early/niche — track, don’t treat as category leaders. |

**Positioning refresh**
- **OpenUsage.sh** = primary collision (34 homepage / 36 docs; +2★ overnight, docs-only commits).
- **ccusage** = report/parser standard (still shipping hard — pricing snapshot today; stars climbing faster than OpenUsage).
- **juliantanx/aiusage** = local multi-tool dashboard peer + name collision.
- **Splitrail** = #2 live local peer (slow star creep).
- Keep Helicone / Langfuse / LiteLLM / AgentCost as complementary observability/gateway spend.

Sources: [openusage.sh](https://openusage.sh/); [openusage.sh/docs](https://openusage.sh/docs/); [github.com/janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage); [github.com/ccusage/ccusage](https://github.com/ccusage/ccusage); [ccusage.com](https://ccusage.com/); [github.com/Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail); [github.com/juliantanx/aiusage](https://github.com/juliantanx/aiusage); [github.com/robinebers/openusage](https://github.com/robinebers/openusage); denshub macOS monitors roundup; GitHub API star/push metadata 2026-09-10 (Europe/London).

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

### Open source — 8 Sep 2026

Fresh weekday GitHub/gh pass (Europe/London). Substantive landscape across ccusage/OpenUsage peers, LLM cost CLIs, local SQLite/TUI dashboards, OTel cost stacks, and Cursor/Claude/Copilot/Codex parsers. Stars/licenses from GitHub API 2026-09-08.

| Repo | Stars | License | Why it matters |
|---|---:|---|---|
| [ccusage/ccusage](https://github.com/ccusage/ccusage) | ~18.4k | NOASSERTION | Category CLI daily/session/blocks cost from local agent JSONL. Pushed today. |
| [janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage) | ~187 | MIT | openusage.sh local multi-tool quota/spend TUI + SQLite. Nearest product twin. |
| [robinebers/openusage](https://github.com/robinebers/openusage) | ~4.0k | MIT | openusage.ai macOS menu-bar tracker. Naming collision only. |
| [Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail) | ~219 | MIT | Real-time cross-agent token/cost monitor (+ Cloud/MCP). Number-two local peer. |
| [zhnd/lumo](https://github.com/zhnd/lumo) | ~147 | MIT | Local-first Claude Code usage/cost/session dashboard. SQLite/TUI UX reference. |
| [sculptdotfun/viberank](https://github.com/sculptdotfun/viberank) | ~115 | MIT | Public leaderboard on ccusage data. |
| [cobra91/better-ccusage](https://github.com/cobra91/better-ccusage) | ~84 | MIT | Faster multi-provider ccusage-style JSONL analyzer. |
| [Halloweedev/usagepal](https://github.com/Halloweedev/usagepal) | ~76 | MIT | Open-source OpenUsage menu-bar fork; Rust + Tauri. |
| [851-labs/tokenmaxxing](https://github.com/851-labs/tokenmaxxing) | ~69 | MIT | Local CLI on ccusage that syncs personal usage with peers. |
| [Nihondo/AgentLimits](https://github.com/Nihondo/AgentLimits) | ~56 | MIT | macOS widgets for Codex/Claude limits + ccusage heatmap. |
| [kenn-io/vibepulse](https://github.com/kenn-io/vibepulse) | ~53 | MIT | macOS menubar for Claude Code + Codex via ccusage. |
| [ItsJazii/pane](https://github.com/ItsJazii/pane) | ~40 | NOASSERTION | Windows tray port of OpenUsage menu-bar line (Claude/Codex/Cursor/Copilot). |
| [ofershap/cursor-usage-tracker](https://github.com/ofershap/cursor-usage-tracker) | ~33 | MIT | Self-hosted Cursor Enterprise spend + anomaly alerts. Team FinOps. |
| [agiwhitelist/tokdiet](https://github.com/agiwhitelist/tokdiet) | ~33 | MIT | Local reverse-proxy meter between agents and APIs + live dashboard. |
| [itvincent-git/codex-usage-desktop](https://github.com/itvincent-git/codex-usage-desktop) | ~28 | MIT | Local-first Codex quota/desktop analytics (macOS + Windows). |
| [BerriAI/litellm](https://github.com/BerriAI/litellm) | ~58.3k | NOASSERTION | AI gateway spend tracking / virtual keys / pricing tables. Adjacent layer. |
| [langfuse/langfuse](https://github.com/langfuse/langfuse) | ~34.3k | NOASSERTION | OSS LLM observability + OTel. Complementary if exporting traces. |
| [Portkey-AI/gateway](https://github.com/Portkey-AI/gateway) | ~12.9k | MIT | AI gateway with cost/guardrails. Adjacent FinOps layer. |
| [Helicone/helicone](https://github.com/Helicone/helicone) | ~6.1k | Apache-2.0 | OSS LLM observability proxy. Adjacent, not a desktop twin. |

**Also watch (wrappers / OTel / niche parsers)**
- [atomchung/ccstory](https://github.com/atomchung/ccstory) (~43★, MIT), [goniszewski/cctray](https://github.com/goniszewski/cctray) (~42★, MIT), [sivchari/ccowl](https://github.com/sivchari/ccowl) (~40★, MIT) — narrative / menu-bar wrappers on Claude usage.
- [saurabhhgi/claude-code-otel-dashboard](https://github.com/saurabhhgi/claude-code-otel-dashboard) — OTel Collector to Prometheus to Grafana for Claude Code cost/tokens/sessions.
- [hestonhamilton/claude-code-usage-stats](https://github.com/hestonhamilton/claude-code-usage-stats) (GPL-3.0) — FastAPI + SQLite multi-machine Claude usage + Prometheus/Grafana.
- [chpock/openusage-cli](https://github.com/chpock/openusage-cli) (MIT) — daemon/CLI collecting provider usage via OpenUsage plugins to local REST API.
- [ryoppippi/ccusage](https://github.com/ryoppippi/ccusage) mirrors ccusage/ccusage; AgentCost (https://agentcost.in/docs/guides/agentcost-integration-into-existing-projects/) — instrumented gateway FinOps (adjacent).

**Build takeaways (8 Sep)**
1. **openusage.sh** (janekbaraniewski) remains the collision twin; **openusage.ai** (robinebers) is the larger menu-bar brand — keep the naming split sharp in copy.
2. Stay compatible with **ccusage** report formats (still shipping hard today); differentiate on live quotas + multi-provider auto-detect + dashboard UX.
3. Splitrail = number-two live local peer; tokdiet = proxy-meter architecture alternative; lumo = local SQLite/TUI UX reference for Claude-centric dashboards.
4. LiteLLM / Langfuse / Helicone / Portkey / AgentCost stay gateway/OTel-adjacent — useful for pricing/export ideas, not category twins.
5. Cursor Admin API trackers and single-vendor Codex/Claude desktop apps stay single-pane — agentUsage job is unification across every agent on the machine.

Sources: gh search + GitHub API star/license/push metadata 2026-09-08 (Europe/London).

### Open source — 9 Sep 2026

Fresh weekday GitHub pass (Europe/London). ccusage/OpenUsage ecosystem still defining the category; several peers pushed today (ccusage, Splitrail, UsagePal, Pane, Codex desktop).

| Repo | Stars | License | Why it matters |
|---|---:|---|---|
| [ccusage/ccusage](https://github.com/ccusage/ccusage) | ~18.4k | NOASSERTION | Category CLI; pushed today — stay format-compatible. |
| [janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage) | ~188 | MIT | openusage.sh local multi-tool TUI + SQLite. Nearest product twin. |
| [robinebers/openusage](https://github.com/robinebers/openusage) | ~4.1k | MIT | openusage.ai menu-bar tracker — naming collision only. |
| [Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail) | ~219 | MIT | Real-time cross-agent monitor; pushed today — number-two local peer. |
| [zhnd/lumo](https://github.com/zhnd/lumo) | ~147 | MIT | Local-first Claude Code usage/cost/session dashboard. |
| [Halloweedev/usagepal](https://github.com/Halloweedev/usagepal) | ~76 | MIT | OpenUsage menu-bar fork (Rust + Tauri); pushed today. |
| [ItsJazii/pane](https://github.com/ItsJazii/pane) | ~41 | NOASSERTION | Windows tray OpenUsage port; pushed today. |
| [itvincent-git/codex-usage-desktop](https://github.com/itvincent-git/codex-usage-desktop) | ~29 | MIT | Local-first Codex quota desktop app; pushed today. |
| [Baek-Seunghyun/ai-coding-usage-card](https://github.com/Baek-Seunghyun/ai-coding-usage-card) | ~24 | MIT | **New fold-in.** Self-hosted GitHub profile SVG card from ccusage data. |
| [sculptdotfun/viberank](https://github.com/sculptdotfun/viberank) | ~115 | MIT | Public leaderboard on ccusage data. |
| [cobra91/better-ccusage](https://github.com/cobra91/better-ccusage) | ~84 | MIT | Faster multi-provider ccusage-style analyzer. |
| [Nihondo/AgentLimits](https://github.com/Nihondo/AgentLimits) | ~56 | MIT | macOS widgets for Codex/Claude limits + ccusage heatmap. |
| [agiwhitelist/tokdiet](https://github.com/agiwhitelist/tokdiet) | ~33 | MIT | Local reverse-proxy meter between agents and APIs. |
| [ofershap/cursor-usage-tracker](https://github.com/ofershap/cursor-usage-tracker) | ~33 | MIT | Self-hosted Cursor Enterprise spend + anomaly alerts. |
| [BerriAI/litellm](https://github.com/BerriAI/litellm) | ~58.3k | NOASSERTION | AI gateway spend/pricing; active today. Adjacent layer. |
| [langfuse/langfuse](https://github.com/langfuse/langfuse) | ~34.4k | NOASSERTION | OSS LLM observability + OTel; active today. |
| [Helicone/helicone](https://github.com/Helicone/helicone) | ~6.1k | Apache-2.0 | OSS LLM observability proxy. Adjacent. |

**Build takeaways:** openusage.sh is the collision twin; openusage.ai is menu-bar brand. Stay ccusage-compatible; differentiate on live quotas + auto-detect + UX. Splitrail/UsagePal/Pane moved today.

Sources: GitHub API 2026-09-09 (Europe/London).

### Open source — 10 Sep 2026

Fresh weekday GitHub pass (Europe/London). ccusage / OpenUsage ecosystem still defines the category; **live pushes** on ccusage, openusage.ai, Codex desktop, LiteLLM, Langfuse. New budget-cap / proxy peers worth watching.

| Repo | Stars | License | Why it matters |
|---|---:|---|---|
| [ccusage/ccusage](https://github.com/ccusage/ccusage) | ~18.5k | NOASSERTION | **Pushed today.** Category CLI — stay format-compatible. |
| [janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage) | ~190 | MIT | openusage.sh local multi-tool TUI + SQLite. Nearest product twin. |
| [robinebers/openusage](https://github.com/robinebers/openusage) | ~4.1k | MIT | **Pushed today.** openusage.ai menu-bar tracker — naming collision only. |
| [Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail) | ~220 | MIT | Real-time cross-agent monitor — number-two local peer. |
| [zhnd/lumo](https://github.com/zhnd/lumo) | ~147 | MIT | Local-first Claude Code usage/cost/session dashboard. |
| [Halloweedev/usagepal](https://github.com/Halloweedev/usagepal) | ~76 | MIT | OpenUsage menu-bar fork (Rust + Tauri). |
| [ItsJazii/pane](https://github.com/ItsJazii/pane) | ~42 | NOASSERTION | Windows tray OpenUsage port. |
| [itvincent-git/codex-usage-desktop](https://github.com/itvincent-git/codex-usage-desktop) | ~30 | MIT | **Pushed today.** Local-first Codex quota desktop app. |
| [Baek-Seunghyun/ai-coding-usage-card](https://github.com/Baek-Seunghyun/ai-coding-usage-card) | ~24 | MIT | Self-hosted GitHub profile SVG card from ccusage data. |
| [sculptdotfun/viberank](https://github.com/sculptdotfun/viberank) | ~116 | MIT | Public leaderboard on ccusage data. |
| [cobra91/better-ccusage](https://github.com/cobra91/better-ccusage) | ~85 | MIT | Faster multi-provider ccusage-style analyzer. |
| [Nihondo/AgentLimits](https://github.com/Nihondo/AgentLimits) | ~56 | MIT | macOS widgets for Codex/Claude limits + ccusage heatmap. |
| [agiwhitelist/tokdiet](https://github.com/agiwhitelist/tokdiet) | ~33 | MIT | Local reverse-proxy meter between agents and APIs. |
| [ofershap/cursor-usage-tracker](https://github.com/ofershap/cursor-usage-tracker) | ~33 | MIT | Self-hosted Cursor Enterprise spend + anomaly alerts. |
| [RoninForge/budgetclaw](https://github.com/RoninForge/budgetclaw) | ~8 | MIT | **New fold-in.** Local Claude Code spend monitor with hard budget caps (project/branch). |
| [DataGrout/lumen](https://github.com/DataGrout/lumen) | ~11 | MIT | **New fold-in.** Real-time LLM token/cost monitor via TLS proxy or HTTP relay. |
| [sergey-homenko/llm_cost_tracker](https://github.com/sergey-homenko/llm_cost_tracker) | ~44 | MIT | **New fold-in.** Rails-native LLM cost ledger with budget guardrails. |
| [BerriAI/litellm](https://github.com/BerriAI/litellm) | ~58.4k | NOASSERTION | **Active today.** AI gateway spend/pricing — adjacent layer. |
| [langfuse/langfuse](https://github.com/langfuse/langfuse) | ~34.4k | NOASSERTION | **Active today.** OSS LLM observability + OTel. |
| [maximhq/bifrost](https://github.com/maximhq/bifrost) | ~7.9k | Apache-2.0 | **New fold-in / active today.** Enterprise AI gateway (LiteLLM-adjacent). |
| [Helicone/helicone](https://github.com/Helicone/helicone) | ~6.1k | Apache-2.0 | OSS LLM observability proxy. Adjacent. |
| [ferro-labs/ai-gateway](https://github.com/ferro-labs/ai-gateway) | ~255 | Apache-2.0 | **New fold-in.** Go AI gateway with cost controls (LiteLLM alternative class). |

**Build takeaways:** openusage.sh is the collision twin; openusage.ai is menu-bar brand. Stay ccusage-compatible; differentiate on live quotas + auto-detect + UX. Budget-cap peers (budgetclaw) and proxy meters (lumen/tokdiet) are rising adjacent patterns.

Sources: GitHub API 2026-09-10 (Europe/London).
