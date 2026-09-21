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

### Update 2026-09-11

Fresh Friday scrape (Europe/London). Twin-risk with OpenUsage.sh unchanged; **ccusage still shipping hardest** (push + stars today); OpenUsage homepage still flubs 34 vs 35 provider counts:

| Product | Delta | Vs agentUsage |
|---|---|---|
| [OpenUsage.sh](https://openusage.sh/) ([GitHub](https://github.com/janekbaraniewski/openusage), **~192★**, +2) | Hero still **“35”** providers; lower page still **“34 providers, and counting.”** Listed agents now explicitly include Hermes, Crush, Mux, Codebuff, Droid, Pi beside Claude/Codex/Cursor/Copilot/Gemini/OpenCode/… Last push still docs-site deps (~2026-09-09 23:39 UTC) — no product surface move overnight. | **Nearest twin** — category definition still ~1:1. Count inconsistency is a copy smell, not a wedge. |
| [ccusage](https://github.com/ccusage/ccusage) / [ccusage.com](https://ccusage.com/) (**~18.5k★** / 18,491; +~22) | Still the category CLI for historical daily/weekly/monthly/session/`blocks` across many agents (**no Cursor** in official source list). **Pushed today** (2026-09-11 ~05:39 UTC). | Best **log-history / cost-report** OSS peer; thinner live quota/rate-limit TUI + key autodetection. |
| [Splitrail](https://github.com/Piebald-AI/splitrail) (**~220★**, flat) | Real-time cross-agent monitor + optional Cloud + MCP. No new push since 2026-09-09. | #2 local “live monitor” peer — quiet day. |
| [AIUsage](https://github.com/juliantanx/aiusage) (**~123★**, +1) | Local web dashboard across 20+ tools. Quiet since 2026-09-08. | Rising local peer + **name collision** (AIUsage vs agentUsage). |
| [openusage.ai](https://github.com/robinebers/openusage) (**~4.1k★** / 4,106; +~16) | macOS menu-bar tracker (different product). Last push ~2026-09-10 19:03 UTC. | Naming collision only — do not conflate with openusage.sh. |
| Vendor UIs | Cursor Spending / Claude `/usage`/`/cost` / OpenRouter Activity still single-vendor panes. | Fragmentation OpenUsage/agentUsage claim to fix. |
| Helicone / Langfuse / LiteLLM | **LiteLLM** (~58.5k★ / 58,500, **pushed today**) and **Langfuse** (~34.5k★ / 34,470, **pushed today**) still active gateway/OTel FinOps. **Helicone** (~6.1k★ / 6,144) still quiet since ~2026-08-31. | Adjacent layers, not desktop twins. |

**Positioning refresh**
- **OpenUsage.sh** = primary collision (34/35 copy split; +2★; docs-only since Wed).
- **ccusage** = report/parser standard (still shipping today; stars climbing faster than OpenUsage).
- **juliantanx/aiusage** = local multi-tool dashboard peer + name collision.
- **Splitrail** = #2 live local peer (flat today).
- Keep Helicone / Langfuse / LiteLLM / AgentCost as complementary observability/gateway spend.

Sources: [openusage.sh](https://openusage.sh/); [github.com/janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage); [github.com/ccusage/ccusage](https://github.com/ccusage/ccusage); [ccusage.com](https://ccusage.com/); [github.com/Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail); [github.com/juliantanx/aiusage](https://github.com/juliantanx/aiusage); [github.com/robinebers/openusage](https://github.com/robinebers/openusage); GitHub API star/push metadata 2026-09-11 (Europe/London).


### Update 2026-09-12

Fresh Saturday scrape (Europe/London). Twin-risk with OpenUsage.sh unchanged; **material overnight moves** on Splitrail (pricing push), Helicone (broke Aug quiet), and openusage.ai (menu-bar still shipping + star climb). Homepage/README/docs now disagree three ways on OpenUsage provider count:

| Product | Delta | Vs agentUsage |
|---|---|---|
| [OpenUsage.sh](https://openusage.sh/) ([GitHub](https://github.com/janekbaraniewski/openusage), **~193★**, +1) | Hero still **“35”** providers; lower page still **“34 providers, and counting.”** README + docs still claim **“36 providers”** (incl. Antigravity). Listed agents still include Hermes/Crush/Mux/Codebuff/Droid/Pi beside Claude/Codex/Cursor/Copilot/Gemini/OpenCode/…. Last product-adjacent push still docs-site deps (~2026-09-09 23:39 UTC) — no surface move overnight. | **Nearest twin** — category definition still ~1:1. 34/35/36 count split is a copy smell, not a wedge. |
| [ccusage](https://github.com/ccusage/ccusage) / [ccusage.com](https://ccusage.com/) (**~18.5k★** / 18,506; +~15) | Still the category CLI for historical daily/weekly/monthly/session/`blocks` across many agents (**no Cursor** in official source list; Grok Build CLI now listed on npm/site). **Pushed today** (2026-09-12 ~05:07 UTC) — LiteLLM pricing snapshot chores. | Best **log-history / cost-report** OSS peer; thinner live quota/rate-limit TUI + key autodetection. |
| [Splitrail](https://github.com/Piebald-AI/splitrail) (**~221★**, +1) | Real-time cross-agent monitor + optional Cloud + MCP. **Pushed yesterday** (2026-09-11 ~16:40 UTC): Grok 4.6 pricing + Gemini/MiniMax/DeepSeek rate sync (#257) — first product commit since 2026-09-09. | #2 local “live monitor” peer — quiet stars, active pricing. |
| [AIUsage](https://github.com/juliantanx/aiusage) (**~125★**, +2) | Local web dashboard (`aiusage serve` → :3847) across 20+ tools. Quiet since 2026-09-08 release. | Rising local peer + **name collision** (AIUsage vs agentUsage). |
| [openusage.ai](https://github.com/robinebers/openusage) (**~4.1k★** / 4,135; +~29) | macOS menu-bar tracker (different product). **Pushed today** (2026-09-12 ~06:43 UTC): v0.7.12-beta.1 changelog + Cursor Grok Bot / Codex Luna pricing fixes. | Naming collision only — do not conflate with openusage.sh. |
| Vendor UIs | Cursor Spending / Claude `/usage`/`/cost` / OpenRouter Activity still single-vendor panes. | Fragmentation OpenUsage/agentUsage claim to fix. |
| Helicone / Langfuse / LiteLLM | **LiteLLM** (~58.5k★ / 58,549, **pushed today**, +~49) and **Langfuse** (~34.5k★ / 34,498, pushed ~2026-09-11 22:55 UTC, +~28) still active gateway/OTel FinOps. **Helicone** (~6.1k★ / 6,149) **broke quiet** — push 2026-09-11 ~18:07 UTC (jawn/tsoa routes fix) after ~2026-08-31 silence. | Adjacent layers, not desktop twins. |
| Also watch (local peers) | [getagentseal/codeburn](https://github.com/getagentseal/codeburn) (~11.0k★, pushed today) — local multi-tool cost CLI (`npx codeburn`, 37 tools); [junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) (~5.4k★, pushed today) — terminal token tracker + leaderboard; [deviffyy/OpenQuota](https://github.com/deviffyy/OpenQuota) (~204★, pushed today) — cross-platform limits/spend; [Nanako0129/TokenBar](https://github.com/Nanako0129/TokenBar) (~345★) / [tddworks/ClaudeBar](https://github.com/tddworks/ClaudeBar) (~1.5k★) — macOS menu-bar quota peers beside openusage.ai. | High-star adjacent local trackers — watch for category overlap, not yet twins on live multi-provider quota TUI + key autodetection. |

**Positioning refresh**
- **OpenUsage.sh** = primary collision (34/35/36 copy split; +1★; still docs-only since Wed).
- **ccusage** = report/parser standard (still shipping today; +~15★ overnight).
- **Splitrail** = #2 live local peer (pricing push overnight; stars nearly flat).
- **juliantanx/aiusage** = local multi-tool dashboard peer + name collision (+2★, quiet code).
- **openusage.ai** = menu-bar naming collision; fastest star climb among “OpenUsage*” names today.
- Keep Helicone / Langfuse / LiteLLM / AgentCost as complementary observability/gateway spend — Helicone no longer fully dormant.

Sources: [openusage.sh](https://openusage.sh/); [github.com/janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage); [github.com/ccusage/ccusage](https://github.com/ccusage/ccusage); [ccusage.com](https://ccusage.com/); [github.com/Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail); [github.com/juliantanx/aiusage](https://github.com/juliantanx/aiusage); [github.com/robinebers/openusage](https://github.com/robinebers/openusage); [github.com/getagentseal/codeburn](https://github.com/getagentseal/codeburn); [github.com/junhoyeo/tokscale](https://github.com/junhoyeo/tokscale); [github.com/deviffyy/OpenQuota](https://github.com/deviffyy/OpenQuota); GitHub API star/push metadata 2026-09-12 (Europe/London).


### Update 2026-09-13

Fresh Sunday scrape (Europe/London). Twin-risk with OpenUsage.sh unchanged; **34/35 provider copy split still live** on the homepage (meta/JSON-LD say 35; hero eyebrow + providers H2 say 34). Material overnight moves: **ccusage** still shipping, **TokenBar** engine work, **codeburn** test merges; Splitrail/AIUsage quiet on code.

| Product | Delta | Vs agentUsage |
|---|---|---|
| [OpenUsage.sh](https://openusage.sh/) ([GitHub](https://github.com/janekbaraniewski/openusage), **~194★**, +1) | Hero meta / FAQ / schema still **“35”** providers; hero eyebrow + “Supported providers” H2 still **“34 providers, and counting.”** README/docs “36” claim from prior passes not contradicted by a new ship — last product-adjacent push still **2026-09-09** (~23:39 UTC). | **Nearest twin** — category definition still ~1:1. Copy smell persists; no surface move overnight. |
| [ccusage](https://github.com/ccusage/ccusage) / [ccusage.com](https://ccusage.com/) (**~18.5k★** / 18,520; +~14) | Still the category CLI for historical daily/weekly/monthly/session/`blocks` across many agents (**no Cursor** in official source list). **Pushed today** (2026-09-13 ~05:16 UTC) — LiteLLM pricing snapshots + issue-gate triage policy (#1717). Statusline beta still Claude-focused. | Best **log-history / cost-report** OSS peer; thinner live quota/rate-limit TUI + key autodetection. |
| [Splitrail](https://github.com/Piebald-AI/splitrail) (**~223★**, +2) | Real-time cross-agent monitor + optional Cloud + MCP. Last product commit still **2026-09-11** (Grok 4.6 pricing + Gemini/MiniMax/DeepSeek rates). Stars nudged; no Sunday code. | #2 local “live monitor” peer — quiet since Fri pricing push. |
| [AIUsage](https://github.com/juliantanx/aiusage) (**~125★**, flat) | Local web dashboard (`aiusage serve` → :3847) across 20+ tools. Quiet since 2026-09-08 release. | Rising local peer + **name collision** (AIUsage vs agentUsage). |
| [openusage.ai](https://github.com/robinebers/openusage) (**~4.1k★** / 4,149; +~14) | macOS menu-bar tracker (different product). Last product commits still v0.7.12-beta.1 / Cursor Grok Bot pricing (2026-09-11–12). | Naming collision only — do not conflate with openusage.sh. |
| Vendor UIs | Cursor Spending / Claude `/usage`/`/cost` / OpenRouter Activity still single-vendor panes. | Fragmentation OpenUsage/agentUsage claim to fix. |
| Helicone / Langfuse / LiteLLM | **LiteLLM** (~58.6k★ / 58,610, **pushed today**) and **Langfuse** (~34.5k★ / 34,524, pushed ~2026-09-13 04:46 UTC) still active gateway/OTel FinOps. **Helicone** (~6.2k★ / 6,152) still quiet since 2026-09-11 jawn/tsoa fix. | Adjacent layers, not desktop twins. |
| Also watch (local peers) | [getagentseal/codeburn](https://github.com/getagentseal/codeburn) (~11.0k★ / 10,985, **pushed today**) — local multi-tool cost CLI; [junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) (~5.4k★); [deviffyy/OpenQuota](https://github.com/deviffyy/OpenQuota) (~203★); [Nanako0129/TokenBar](https://github.com/Nanako0129/TokenBar) (~347★, +2; overnight engine-pin-window / pending-scan docs) / [tddworks/ClaudeBar](https://github.com/tddworks/ClaudeBar) (~1.5k★) — macOS menu-bar quota peers. | High-star adjacent local trackers — watch for category overlap, not yet twins on live multi-provider quota TUI + key autodetection. |

**Positioning refresh**
- **OpenUsage.sh** = primary collision (34/35 copy split still live; +1★; still no code since Wed).
- **ccusage** = report/parser standard (still shipping today; +~14★ overnight).
- **Splitrail** = #2 live local peer (stars +2; quiet code since Fri).
- **juliantanx/aiusage** = local multi-tool dashboard peer + name collision (flat).
- **openusage.ai** = menu-bar naming collision; still climbing stars without conflating brands.
- Keep Helicone / Langfuse / LiteLLM as complementary observability/gateway spend.

Sources: [openusage.sh](https://openusage.sh/); [github.com/janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage); [github.com/ccusage/ccusage](https://github.com/ccusage/ccusage); [ccusage.com](https://ccusage.com/); [github.com/Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail); [github.com/juliantanx/aiusage](https://github.com/juliantanx/aiusage); [github.com/robinebers/openusage](https://github.com/robinebers/openusage); [github.com/getagentseal/codeburn](https://github.com/getagentseal/codeburn); [github.com/Nanako0129/TokenBar](https://github.com/Nanako0129/TokenBar); GitHub API star/push metadata 2026-09-13 (Europe/London).


### Update 2026-09-14

Fresh Monday scrape (Europe/London). Twin-risk with OpenUsage.sh unchanged; **34/35 homepage copy split still live**, and GitHub README now claims **36** (three-way count drift). Material overnight/day moves: **OpenUsage.sh** shipped Codex 0.153+ + Crush timestamp fixes (first product code since Wed); **AIUsage** released **v1.5.16**; **codeburn** menubar quota-crossing notifications; **tokscale** / **LiteLLM** / **Langfuse** still shipping; Splitrail quiet; ccusage pricing-snapshot cadence continues.

| Product | Delta | Vs agentUsage |
|---|---|---|
| [OpenUsage.sh](https://openusage.sh/) ([GitHub](https://github.com/janekbaraniewski/openusage), **~195★**, +1) | Homepage meta / “Supported providers (35)” / “across 35” still say **35**; hero eyebrow + “Supported providers” H2 still **“34 providers, and counting.”**; README / docs still claim **36 providers**. **Product code overnight:** `fix(codex)` CLI 0.153+ model detection + cached-token double-billing (#365) and `fix(crush)` Unix-seconds timestamps (#357) — pushed **2026-09-13 ~15:58 UTC** (first product commits since 2026-09-09). | **Nearest twin** — category definition still ~1:1. Copy smell worsened (34/35/36); twin shipped real provider fixes while agentUsage research PR stays draft. |
| [ccusage](https://github.com/ccusage/ccusage) / [ccusage.com](https://ccusage.com/) (**~18.5k★** / 18,543; +~23) | Still the category CLI for historical daily/weekly/monthly/session/`blocks` across many agents (**no Cursor** in official source list). Last main commits still **2026-09-13** LiteLLM pricing snapshots + issue-gate triage (#1717); stars climbing; no new Mon product commit visible on default branch tip. | Best **log-history / cost-report** OSS peer; thinner live quota/rate-limit TUI + key autodetection. |
| [Splitrail](https://github.com/Piebald-AI/splitrail) (**~223★**, flat) | Real-time cross-agent monitor + optional Cloud + MCP. Last product commit still **2026-09-11** (Grok 4.6 pricing + Gemini/MiniMax/DeepSeek rates). No Mon code/star move. | #2 local “live monitor” peer — still quiet since Fri pricing push. |
| [AIUsage](https://github.com/juliantanx/aiusage) (**~126★**, +1) | Local web dashboard (`aiusage serve` → :3847) across 20+ tools. **Broke quiet period:** released **v1.5.16** (2026-09-14 ~03:21 UTC) after Antigravity varint + Codex dashboard trust-boundary merges (#53/#54). | Rising local peer + **name collision** (AIUsage vs agentUsage) — now shipping again. |
| [openusage.ai](https://github.com/robinebers/openusage) (**~4.2k★** / 4,158; +~9) | macOS menu-bar tracker (different product). Last product commits still v0.7.12-beta.1 / Cursor Grok Bot pricing (2026-09-11). Stars still climbing without new Mon code. | Naming collision only — do not conflate with openusage.sh. |
| Vendor UIs | **Cursor:** Spending (`cursor.com/dashboard/spending`) + Usage (`…/usage`) still single-vendor; self-serve Usage remains token/“Included”-centric after mid-2026 dollar-column removal. **Claude Code:** `/usage` (plan windows + local breakdown) and `/cost` (session API spend) remain in-tool panes; Console for authoritative billing. **OpenRouter:** Activity / credits dashboard still API-platform-only (fetch often 403 to bots). | Fragmentation OpenUsage/agentUsage claim to fix. |
| Helicone / Langfuse / LiteLLM | **LiteLLM** (~58.7k★ / 58,680, **pushed today**) — gateway spend UI + budgets still the reference proxy FinOps layer. **Langfuse** (~34.6k★ / 34,574, **pushed today**) — OTel traces/evals still very active. **Helicone** (~6.2k★ / 6,153) still quiet since 2026-09-11 jawn/tsoa fix. | Adjacent layers, not desktop twins. |
| Also watch (local peers) | [getagentseal/codeburn](https://github.com/getagentseal/codeburn) (~11.0k★ / 10,999, +~14; **shipping**) — now **37 tools** in description; overnight menubar quota 80%/100% crossing notifications + window-rollover refresh (#1360/#1361). [junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) (~5.4k★ / 5,425; **pushed today**) — Codex thread-kind grouping + Antigravity CLI timestamp fix. [deviffyy/OpenQuota](https://github.com/deviffyy/OpenQuota) (~203★, flat; quiet since 2026-09-12). [Nanako0129/TokenBar](https://github.com/Nanako0129/TokenBar) (~348★, +1; last engine-pin-window work 2026-09-12/13) / [tddworks/ClaudeBar](https://github.com/tddworks/ClaudeBar) (~1.5k★ / 1,483; v0.4.92 changelog 2026-09-12) — macOS menu-bar quota peers. | High-star adjacent local trackers — codeburn + tokscale closest “also watch” movers; not yet twins on live multi-provider quota TUI + key autodetection. |

**Positioning refresh**
- **OpenUsage.sh** = primary collision (34/35/36 count drift; +1★; **shipped Codex/Crush fixes** after multi-day quiet).
- **ccusage** = report/parser standard (stars +~23; pricing snapshots still the cadence).
- **Splitrail** = #2 live local peer (flat; quiet code since Fri).
- **juliantanx/aiusage** = local multi-tool dashboard peer + name collision (**v1.5.16 today**).
- **openusage.ai** = menu-bar naming collision; stars still climbing without conflating brands.
- **codeburn / tokscale** = highest-velocity adjacent local CLIs (menubar quota alerts; Codex/Antigravity fixes).
- Keep Helicone / Langfuse / LiteLLM as complementary observability/gateway spend.

Sources: [openusage.sh](https://openusage.sh/); [github.com/janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage); [github.com/ccusage/ccusage](https://github.com/ccusage/ccusage); [ccusage.com](https://ccusage.com/); [github.com/Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail); [github.com/juliantanx/aiusage](https://github.com/juliantanx/aiusage); [github.com/robinebers/openusage](https://github.com/robinebers/openusage); [github.com/getagentseal/codeburn](https://github.com/getagentseal/codeburn); [github.com/junhoyeo/tokscale](https://github.com/junhoyeo/tokscale); [github.com/Nanako0129/TokenBar](https://github.com/Nanako0129/TokenBar); [github.com/tddworks/ClaudeBar](https://github.com/tddworks/ClaudeBar); [github.com/deviffyy/OpenQuota](https://github.com/deviffyy/OpenQuota); [docs.litellm.ai spend tracking](https://docs.litellm.ai/docs/proxy/cost_tracking); [code.claude.com costs](https://code.claude.com/docs/en/costs); Cursor forum Spending/Usage threads; GitHub API star/push metadata 2026-09-14 (Europe/London).


### Update 2026-09-15

Fresh Tuesday scrape (Europe/London). Twin-risk with OpenUsage.sh unchanged; **34/35/36 count drift still live** (homepage meta/H2 vs README). Material day moves: **AIUsage v1.5.17**; **openusage.ai** Claude Swap / Devin / plan-badge fixes; **tokscale 4.17.0**; ccusage pricing-snapshot cadence; OpenUsage.sh deps-only since yesterday’s Codex/Crush product fix.

| Product | Delta | Vs agentUsage |
|---|---|---|
| [OpenUsage.sh](https://openusage.sh/) ([GitHub](https://github.com/janekbaraniewski/openusage), **~195★**, flat) | Homepage still **“Supported providers (35)”** + hero **“34 providers, and counting.”** README/docs now firmly market **36 providers** and list **Antigravity CLI** (`agy` + `~/.gemini/antigravity-cli`) with status-line/session/quota hooks. Last pushes Mon **deps/CI only** after Sun Codex 0.153+ + Crush timestamp product fixes. | **Nearest twin** — category definition still ~1:1. Copy smell persists (34/35/36); Antigravity called out in README while site count lags. |
| [ccusage](https://github.com/ccusage/ccusage) / [ccusage.com](https://ccusage.com/) (**~18.6k★** / 18,561; +~18) | Still the category CLI for historical daily/weekly/monthly/session/`blocks` (**no Cursor** in official source list). **Shipping today** — repeated LiteLLM pricing snapshots (through ~2026-09-15 03:21 UTC). | Best **log-history / cost-report** OSS peer; thinner live quota/rate-limit TUI + key autodetection. |
| [Splitrail](https://github.com/Piebald-AI/splitrail) (**~223★**, flat) | Real-time cross-agent monitor + optional Cloud + MCP. Last product commit still **2026-09-11** (Grok 4.6 pricing). Quiet again. | #2 local “live monitor” peer — multi-day quiet. |
| [AIUsage](https://github.com/juliantanx/aiusage) (**~127★**, +1) | Local web dashboard (`aiusage serve` → :3847) across 20+ tools. **v1.5.17 today** (2026-09-15 ~03:10 UTC): responsive startup (listen before parse; yield during history backfill; cache pricing); custom date ranges use **local calendar** not UTC (#62); hide Support nav item. | Rising local peer + **name collision** (AIUsage vs agentUsage) — second release in two days. |
| [openusage.ai](https://github.com/robinebers/openusage) (**~4.2k★** / 4,162; +~4) | macOS menu-bar tracker (different product). **Shipped overnight:** Claude Swap account discovery + isolated usage (#1226); Devin weekly quota when % omitted (#1251); Claude plan badge from live Anthropic profile (#1262); null nested-model usage fix (#1261). Changelog **v0.7.12-beta.2**. | Naming collision only — do not conflate with openusage.sh; still the louder star graph. |
| Vendor UIs | Cursor Spending/Usage, Claude `/usage`/`/cost`, OpenRouter Activity still single-vendor panes. | Fragmentation OpenUsage/agentUsage claim to fix. |
| Helicone / Langfuse / LiteLLM | **LiteLLM** (~58.8k★ / 58,767, **pushed today**) and **Langfuse** (~34.6k★ / 34,627, **pushed today**) still active gateway/OTel FinOps. **Helicone** (~6.2k★ / 6,158) quieter since 2026-09-13. | Adjacent layers, not desktop twins. |
| Also watch (local peers) | [getagentseal/codeburn](https://github.com/getagentseal/codeburn) (~11.0k★ / 11,024; Mon menubar quota alerts still the last product story; Tue = deps/tests/Tauri). [junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) (~5.4k★ / 5,434; **v4.17.0 today**) — Cline `modelInfo` identity, Antigravity CLI log-port fix, OpenRouter version-separator pricing. [Nanako0129/TokenBar](https://github.com/Nanako0129/TokenBar) (~348★; last engine commits 2026-09-12). [planetf1/otelite](https://github.com/planetf1/otelite) (~89★, **pushed today**). | tokscale = sharpest adjacent CLI mover today; codeburn still high-star multi-tool gravity. |

**Positioning refresh**
- **OpenUsage.sh** = primary collision (34/35/36 + Antigravity README; no new product code today).
- **ccusage** = report/parser standard (pricing snapshots still the cadence; +~18★).
- **Splitrail** = #2 live local peer (flat/quiet).
- **juliantanx/aiusage** = local multi-tool dashboard peer + name collision (**v1.5.17**).
- **openusage.ai** = menu-bar naming collision still shipping hard (Claude Swap / Devin / plan badge).
- **tokscale / codeburn** = highest-velocity adjacent local CLIs.
- Keep Helicone / Langfuse / LiteLLM as complementary observability/gateway spend.

Sources: [openusage.sh](https://openusage.sh/); [github.com/janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage) README (36 / Antigravity); [github.com/ccusage/ccusage](https://github.com/ccusage/ccusage); [github.com/juliantanx/aiusage](https://github.com/juliantanx/aiusage) v1.5.17; [github.com/robinebers/openusage](https://github.com/robinebers/openusage); [github.com/junhoyeo/tokscale](https://github.com/junhoyeo/tokscale); [github.com/getagentseal/codeburn](https://github.com/getagentseal/codeburn); [github.com/Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail); GitHub API star/push metadata 2026-09-15 (Europe/London).



### Update 2026-09-16

Fresh Wednesday scrape (Europe/London). Twin-risk with OpenUsage.sh unchanged; **34/35 homepage vs 36 README** count drift still live. Material day moves: **ccusage** shipping hard (OpenClaw SQLite, Copilot resume, Codex GPT-6 Astra, CLI date/timezone guards); **openusage.ai Codex Swap**; **codeburn** Compare-periods / session-drawer UX; **AIUsage CodeBuddy** undercount fix; adjacent layer — **Helicone confirmed maintenance mode** (Mintlify acq.; Langfuse migration guide).

| Product | Delta | Vs agentUsage |
|---|---|---|
| [OpenUsage.sh](https://openusage.sh/) ([GitHub](https://github.com/janekbaraniewski/openusage), **~196★**, +1) | Homepage still **“Supported providers (35)”** + hero **“34 providers, and counting.”**; docs/README still **36**. Last pushes still **deps/CI only** (2026-09-14); last product code still Sun Codex 0.153+ + Crush timestamp fixes. | **Nearest twin** — category definition still ~1:1. Copy smell persists; no product surface move since Mon. |
| [ccusage](https://github.com/ccusage/ccusage) / [ccusage.com](https://ccusage.com/) (**~18.6k★** / 18,576; +~15) | Still the category CLI for historical daily/weekly/monthly/session/`blocks` (**no Cursor** in official source list; Grok Build listed on site). **Shipping today:** OpenClaw per-agent SQLite transcripts (#1727); Copilot resume history (#1724); Codex GPT-6 Astra fast multiplier (#1726); `--breakdown` in Codex/shared tables (#1725); reject bad `--timezone` / `--since`>`--until` (#1733/#1735); LiteLLM pricing snapshots. | Best **log-history / cost-report** OSS peer; thinner live quota/rate-limit TUI + key autodetection. |
| [Splitrail](https://github.com/Piebald-AI/splitrail) (**~223★**, flat) | Real-time cross-agent monitor + optional Cloud + MCP. Last product commit still **2026-09-11** (Grok 4.6 pricing). Quiet fifth day. | #2 local “live monitor” peer — multi-day quiet. |
| [AIUsage](https://github.com/juliantanx/aiusage) (**~128★**, +1) | Local web dashboard (`aiusage serve` → :3847) across 20+ tools. **Merged today:** CodeBuddy CLI undercount fix — count usage on `function_call` lines (was missing ~95% of DeepSeek agent requests) (#65). Still on **v1.5.17** (Tue). | Rising local peer + **name collision** (AIUsage vs agentUsage) — parser correctness shipping. |
| [openusage.ai](https://github.com/robinebers/openusage) (**~4.2k★** / 4,173; +~11) | macOS menu-bar tracker (different product). **Shipped today:** Codex Swap account support (#1264) after Tue’s Claude Swap / Devin / plan-badge wave (v0.7.12-beta.2). | Naming collision only — do not conflate with openusage.sh; still the louder star graph. |
| Vendor UIs | **Cursor:** Spending (`cursor.com/dashboard/spending`) + Usage still split — self-serve Usage token/“Included”-centric; USD on Spending / invoices / Admin API (forum threads still active through Aug 2026). **Claude Code:** `/usage` + `/cost` unchanged in-tool panes. **OpenRouter:** Activity / credits still API-platform-only. | Fragmentation OpenUsage/agentUsage claim to fix. |
| Helicone / Langfuse / LiteLLM | **Helicone** (~6.2k★ / 6,159): quiet since 2026-09-11; **Langfuse’s migrate-from-Helicone guide** now states Helicone is in **maintenance mode** after **Mintlify acquisition (Mar 2026)** — services live, feature roadmap stopped; export early. **Langfuse** (~34.7k★ / 34,675, **pushed today**) — v4 Cloud cutover still **2026-11-16**. **LiteLLM** (~58.9k★ / 58,863, **pushed today**) — admin UI routing/forecast polish + active gateway spend FinOps. | Adjacent layers, not desktop twins — Helicone now a **declining** observability peer; prefer Langfuse/LiteLLM for new gateway/OTel work. |
| Also watch (local peers) | [getagentseal/codeburn](https://github.com/getagentseal/codeburn) (~11.0k★ / 11,043; **shipping**) — Compare periods UX (#1446), session drawer (#1444), refresh without layout shift (#1445), Codex distinct-request fix (#1264). [junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) (~5.4k★ / 5,447; quiet since **v4.17.0** Tue). [Nanako0129/TokenBar](https://github.com/Nanako0129/TokenBar) (~350★, +2; **Grok Bot** credential/path support merged #316). [tddworks/ClaudeBar](https://github.com/tddworks/ClaudeBar) (~1.5k★); [deviffyy/OpenQuota](https://github.com/deviffyy/OpenQuota) (~208★, quiet). | codeburn = sharpest adjacent multi-tool UI mover today; TokenBar adds Grok Bot; not yet twins on live multi-provider quota TUI + key autodetection. |

**Positioning refresh**
- **OpenUsage.sh** = primary collision (34/35/36 + Antigravity README; deps-only since Sun product fixes; +1★).
- **ccusage** = report/parser standard (hardest ship cadence today — OpenClaw/Copilot/Codex + CLI guards).
- **Splitrail** = #2 live local peer (flat/quiet since Fri).
- **juliantanx/aiusage** = local multi-tool dashboard peer + name collision (CodeBuddy parser fix).
- **openusage.ai** = menu-bar naming collision still shipping (Codex Swap after Claude Swap).
- **codeburn / TokenBar** = highest-velocity adjacent local apps today.
- Keep Langfuse / LiteLLM as complementary observability/gateway spend; treat **Helicone as maintenance-mode / migrate-away** adjacent, not a growing peer.

Sources: [openusage.sh](https://openusage.sh/); [github.com/janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage); [github.com/ccusage/ccusage](https://github.com/ccusage/ccusage); [ccusage.com](https://ccusage.com/); [github.com/juliantanx/aiusage](https://github.com/juliantanx/aiusage) #65; [github.com/robinebers/openusage](https://github.com/robinebers/openusage) #1264; [github.com/getagentseal/codeburn](https://github.com/getagentseal/codeburn); [github.com/Nanako0129/TokenBar](https://github.com/Nanako0129/TokenBar); [github.com/Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail); [langfuse.com migrate-from-helicone](https://langfuse.com/resources/engineering/migrate-from-helicone); [helicone.ai/blog/joining-mintlify](https://www.helicone.ai/blog/joining-mintlify); [docs.litellm.ai](https://docs.litellm.ai/); Cursor Spending/Usage help + forum threads; GitHub API star/push metadata 2026-09-16 (Europe/London).



### Update 2026-09-17

Fresh Thursday scrape (Europe/London). Twin-risk with OpenUsage.sh unchanged; **34/35 homepage vs 36 README** count drift still live. Material day moves: **ccusage** Codex **GPT Reserve → GPT-5.6 Luna** pricing + LiteLLM snapshot churn; **codeburn** Antigravity tool/bash/MCP/skills SQLite parse + overnight **Grok Bot** sessions/estimated spend; **TokenBar v1.18.0** Codex turn-count fix; **CodeZeno Usage-Monitor v2.11.29**.

| Product | Delta | Vs agentUsage |
|---|---|---|
| [OpenUsage.sh](https://openusage.sh/) ([GitHub](https://github.com/janekbaraniewski/openusage), **~197★**, +1) | Homepage still **“Supported providers (35)”** + hero **“34 providers, and counting.”**; docs/README still **36**. Last pushes still **deps/CI only** (2026-09-14); last product code still Sun Codex 0.153+ + Crush timestamp fixes. | **Nearest twin** — category definition still ~1:1. Copy smell persists; no product surface move since Mon. |
| [ccusage](https://github.com/ccusage/ccusage) / [ccusage.com](https://ccusage.com/) (**~18.6k★** / 18,593; +~17) | Still the category CLI for historical daily/weekly/monthly/session/`blocks` (**no Cursor** in official source list). **Shipped since Wed:** price Codex **GPT Reserve** as **GPT-5.6 Luna** (#1739 — clears “Missing pricing for gpt-reserve”); continuous LiteLLM pricing snapshots overnight/today; Nix CLI features (#1741). | Best **log-history / cost-report** OSS peer; thinner live quota/rate-limit TUI + key autodetection. |
| [Splitrail](https://github.com/Piebald-AI/splitrail) (**~223★**, flat) | Real-time cross-agent monitor + optional Cloud + MCP. Last product commit still **2026-09-11** (Grok 4.6 pricing). Quiet sixth day. | #2 local “live monitor” peer — multi-day quiet. |
| [AIUsage](https://github.com/juliantanx/aiusage) (**~128★**, flat) | Local web dashboard (`aiusage serve` → :3847) across 20+ tools. Still on **v1.5.17** + Wed’s CodeBuddy `function_call` undercount fix (#65). Quiet Thursday. | Rising local peer + **name collision** (AIUsage vs agentUsage). |
| [openusage.ai](https://github.com/robinebers/openusage) (**~4.2k★** / 4,184; +~11) | macOS menu-bar tracker (different product). Latest product still Wed’s **Codex Swap** (#1264) after Tue’s Claude Swap / Devin wave (**v0.7.12-beta.2**). Dep churn only since. | Naming collision only — do not conflate with openusage.sh; still the louder star graph. |
| Vendor UIs | **Cursor:** Spending (`cursor.com/dashboard/spending`) + Usage still split — self-serve Usage token/“Included”-centric; USD on Spending / invoices / Admin API. **Claude Code:** `/usage` + `/cost` unchanged in-tool panes. **OpenRouter:** Activity / credits still API-platform-only. | Fragmentation OpenUsage/agentUsage claim to fix. |
| Helicone / Langfuse / LiteLLM | **Helicone** (~6.2k★ / 6,161): push overnight (2026-09-16) but **Langfuse migrate-from-Helicone** still frames Helicone as **maintenance mode** post-Mintlify (Mar 2026). **Langfuse** (~34.7k★ / 34,712, **pushed today**). **LiteLLM** (~59.0k★ / 58,964, **pushed today**) — gateway spend FinOps + pricing feed that ccusage snapshots. | Adjacent layers, not desktop twins — prefer Langfuse/LiteLLM for new gateway/OTel work; Helicone still migrate-away adjacent. |
| Also watch (local peers) | [getagentseal/codeburn](https://github.com/getagentseal/codeburn) (~11.1k★ / 11,055; **shipping today**) — Antigravity SQLite **steps** parse for tools/bash/MCP/skills (#1456); overnight **Grok Bot** sessions + estimated spend + weekly allowance on all surfaces (#1462); Desktop polish (#1459). [Nanako0129/TokenBar](https://github.com/Nanako0129/TokenBar) (~352★, **v1.18.0** 16 Sep) — Codex turn-count fix for CLI **0.145+** (`item_completed`/`UserMessage` vs old `user_message`). [CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) (~489★, **v2.11.29** today) after Wed’s multi-account **v2.11.28**. [ItsJazii/pane](https://github.com/ItsJazii/pane) (~50★, +2) Windows tray coverage. [junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) (~5.4k★; quiet since **v4.17.0**). | codeburn = sharpest adjacent multi-tool mover today (Antigravity depth + Grok Bot); TokenBar / CodeZeno = Codex/Claude account UX; not yet twins on live multi-provider quota TUI + key autodetection. |

**Positioning refresh**
- **OpenUsage.sh** = primary collision (34/35/36 + Antigravity README; deps-only since Sun product fixes; +1★).
- **ccusage** = report/parser standard (GPT Reserve/Luna pricing + LiteLLM snapshot cadence).
- **Splitrail** = #2 live local peer (flat/quiet since Fri).
- **juliantanx/aiusage** = local multi-tool dashboard peer + name collision (quiet after CodeBuddy fix).
- **openusage.ai** = menu-bar naming collision (Codex Swap still latest feature).
- **codeburn / TokenBar / CodeZeno** = highest-velocity adjacent local apps today.
- Keep Langfuse / LiteLLM as complementary observability/gateway spend; treat **Helicone as maintenance-mode / migrate-away** adjacent, not a growing peer.

Sources: [openusage.sh](https://openusage.sh/); [github.com/janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage); [github.com/ccusage/ccusage](https://github.com/ccusage/ccusage) #1739; [ccusage.com](https://ccusage.com/); [github.com/juliantanx/aiusage](https://github.com/juliantanx/aiusage); [github.com/robinebers/openusage](https://github.com/robinebers/openusage); [github.com/getagentseal/codeburn](https://github.com/getagentseal/codeburn) #1456 #1462; [github.com/Nanako0129/TokenBar](https://github.com/Nanako0129/TokenBar) v1.18.0; [github.com/CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor); [github.com/Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail); [langfuse.com migrate-from-helicone](https://langfuse.com/resources/engineering/migrate-from-helicone); [docs.litellm.ai](https://docs.litellm.ai/); Cursor Spending/Usage help; GitHub API star/push metadata 2026-09-17 (Europe/London).


### Update 2026-09-18

Fresh Friday scrape (Europe/London). Twin-risk with OpenUsage.sh unchanged; **34/35 homepage vs 36 README** count drift still live. Material day moves: **ccusage v20.0.22** maps Codex **auto-review → GPT-5.6 Luna** (+ Antigravity SQLite / Copilot session-state in **v20.0.21**); **tokscale** ships Copilot CLI **session-store.db** parser overnight; **codeburn** Devin token pricing + OpenCode 2.x + desktop $ breakdown popover; **openusage.ai v0.7.12 / v0.7.13-beta.1** (Codex Swap).

| Product | Delta | Vs agentUsage |
|---|---|---|
| [OpenUsage.sh](https://openusage.sh/) ([GitHub](https://github.com/janekbaraniewski/openusage), **~197★**, flat) | Homepage still **“Supported providers (35)”** + hero **“34 providers, and counting.”**; README still **36**. Last pushes still **deps/CI only** (2026-09-14); last product code still Sun Codex 0.153+ fixes. | **Nearest twin** — category definition still ~1:1. Copy smell persists; no product surface move. |
| [ccusage](https://github.com/ccusage/ccusage) / [ccusage.com](https://ccusage.com/) (**~18.6k★** / 18,614; +~21) | **v20.0.22** (17 Sep ~23:10 BST): map Codex **`codex-auto-review` → `gpt-5.6-luna`** from Jul 30 transition (#1753). **v20.0.21** (17 Sep midday): Antigravity SQLite conversation usage (#1677), Copilot session-state events (#1676), ZCode adapter, Codex originator breakdowns + missing-pricing reports. Still no Cursor in official source list. | Best **log-history / cost-report** OSS peer; thinner live quota/rate-limit TUI + key autodetection. |
| [Splitrail](https://github.com/Piebald-AI/splitrail) (**~223★**, flat) | Last product commit still **2026-09-11** (Grok 4.6 pricing). Quiet seventh day. | #2 local “live monitor” peer — multi-day quiet. |
| [AIUsage](https://github.com/juliantanx/aiusage) (**~129★**, +1) | Still **v1.5.17** + Wed’s CodeBuddy `function_call` undercount fix (#65). Quiet since. | Rising local peer + **name collision** (AIUsage vs agentUsage). |
| [openusage.ai](https://github.com/robinebers/openusage) (**~4.2k★** / 4,196; +~12) | **v0.7.12** + **v0.7.13-beta.1** shipped Thu: Codex Swap (#1264), Muse Spark 1.3 effort pricing, Claude Swap discover, Devin weekly-quota edge, Cursor **Grok Bot** mode rate split. | Naming collision only — do not conflate with openusage.sh; still the louder star graph. |
| Vendor UIs | **Cursor:** Spending + Usage still split. **Claude Code:** `/usage` + `/cost` unchanged. **OpenRouter:** Activity / credits still API-platform-only. | Fragmentation OpenUsage/agentUsage claim to fix. |
| Helicone / Langfuse / LiteLLM | **Helicone** (~6.2k★ / 6,163): quiet since 16 Sep; Langfuse migrate-from-Helicone still frames **maintenance mode**. **Langfuse** (~34.8k★ / 34,763, **pushed today**). **LiteLLM** (~59.1k★ / 59,054, **pushed today**). | Adjacent layers — prefer Langfuse/LiteLLM for new gateway/OTel work; Helicone still migrate-away adjacent. |
| Also watch (local peers) | [getagentseal/codeburn](https://github.com/getagentseal/codeburn) (~11.1k★ / 11,069; **shipping today**) — Devin sessions priced from tokens not ACU (#1469); OpenCode **2.x** `session_v2` / `session_message` (#1436); Codex early-reset menubar (#1339); desktop **token breakdown popover** behind $ amounts (#1472, today). [junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) (~5.5k★ / 5,475; **pushed overnight**) — Copilot CLI **`session-store.db` / `assistant_usage_events`** parser (#1346); Kiro input/cache estimate (#1342). [Nanako0129/TokenBar](https://github.com/Nanako0129/TokenBar) (~357★, +5; still **v1.18.0** 16 Sep). [CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) (~494★, +5; still **v2.11.29** — OpenCode Go console polling). [ItsJazii/pane](https://github.com/ItsJazii/pane) (~51★, +1). | codeburn + tokscale = highest-velocity adjacent movers today; TokenBar/CodeZeno flat on version after Thu ships. |

**Positioning refresh**
- **OpenUsage.sh** = primary collision (34/35/36 + Antigravity README; deps-only since Sun product fixes).
- **ccusage** = report/parser standard (auto-review→Luna + Antigravity/Copilot/ZCode wave in v20.0.21–22).
- **Splitrail** = #2 live local peer (flat/quiet since Fri 11 Sep).
- **juliantanx/aiusage** = local multi-tool dashboard peer + name collision.
- **openusage.ai** = menu-bar naming collision (Codex Swap now in tagged releases).
- **codeburn / tokscale** = highest-velocity adjacent local apps today (Devin/OpenCode depth; Copilot CLI DB).
- Keep Langfuse / LiteLLM as complementary observability/gateway spend; treat **Helicone as maintenance-mode / migrate-away** adjacent.

Sources: [openusage.sh](https://openusage.sh/); [github.com/janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage); [github.com/ccusage/ccusage](https://github.com/ccusage/ccusage) v20.0.21–22 / #1753 #1677; [github.com/robinebers/openusage](https://github.com/robinebers/openusage) v0.7.12 / v0.7.13-beta.1; [github.com/getagentseal/codeburn](https://github.com/getagentseal/codeburn) #1469 #1436 #1472; [github.com/junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) #1346; [github.com/Nanako0129/TokenBar](https://github.com/Nanako0129/TokenBar); [github.com/CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor); [github.com/Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail); [langfuse.com migrate-from-helicone](https://langfuse.com/resources/engineering/migrate-from-helicone); GitHub API star/push metadata 2026-09-18 (Europe/London).

### Update 2026-09-19

Fresh Saturday scrape (Europe/London). Twin-risk with OpenUsage.sh unchanged; **34/35 homepage vs 36 README** count drift still live. Material day moves: **ccusage v20.0.23** (Claude copied-request dedupe); **TokenBar v1.19.0 / v1.19.1** (hide unset provider tabs + merge Antigravity IDE/CLI + quota-lens history keep); **codeburn** ships **ZCode (z.ai) live quota**; **CodeZeno** jumped **v2.11.29 → v2.12.37** (widget drag/dock + Claude desktop usage restore).

| Product | Delta | Vs agentUsage |
|---|---|---|
| [OpenUsage.sh](https://openusage.sh/) ([GitHub](https://github.com/janekbaraniewski/openusage), **~199★**, +2) | Homepage still **“34 providers, and counting.”**; README still **36**. Last pushes still **deps/CI only** (2026-09-14); last product code still Sun Codex 0.153+ fixes. | **Nearest twin** — category definition still ~1:1. Copy smell persists; no product surface move. |
| [ccusage](https://github.com/ccusage/ccusage) / [ccusage.com](https://ccusage.com/) (**~18.6k★** / 18,624; +~10) | **v20.0.23** (18 Sep ~13:40 BST): **claude** dedupe copied requests across sessions (#1765). Still no Cursor in official source list. | Best **log-history / cost-report** OSS peer; thinner live quota/rate-limit TUI + key autodetection. |
| [Splitrail](https://github.com/Piebald-AI/splitrail) (**~223★**, flat) | Last product commit still **2026-09-11** (Grok 4.6 pricing). Quiet eighth day. | #2 local “live monitor” peer — multi-day quiet. |
| [AIUsage](https://github.com/juliantanx/aiusage) (**~129★**, flat) | Still **v1.5.17** + Wed’s CodeBuddy `function_call` undercount fix (#65). Quiet since. | Rising local peer + **name collision** (AIUsage vs agentUsage). |
| [openusage.ai](https://github.com/robinebers/openusage) (**~4.2k★** / 4,198; +~2) | Still on **v0.7.12** + **v0.7.13-beta.1** (Thu Codex Swap). Saturday pushes are quiet / non-product. | Naming collision only — do not conflate with openusage.sh; still the louder star graph. |
| Vendor UIs | **Cursor:** Spending + Usage still split. **Claude Code:** `/usage` + `/cost` unchanged. **OpenRouter:** Activity / credits still API-platform-only. | Fragmentation OpenUsage/agentUsage claim to fix. |
| Helicone / Langfuse / LiteLLM | **Helicone** (~6.2k★ / 6,164): quiet since 16 Sep; Langfuse migrate-from-Helicone still frames **maintenance mode**. **Langfuse** (~34.8k★ / 34,800, pushed overnight). **LiteLLM** (~59.1k★ / 59,129, **pushed today**). | Adjacent layers — prefer Langfuse/LiteLLM for new gateway/OTel work; Helicone still migrate-away adjacent. |
| Also watch (local peers) | [getagentseal/codeburn](https://github.com/getagentseal/codeburn) (~11.1k★ / 11,094; **shipping overnight**) — **ZCode (z.ai coding plan) live quota** in Plans sidebar + `codeburn quota` (#1347); plus $0-priced usage naming / Warp FDA freeze / Grok Bot quota harden (#1487–#1491). [Nanako0129/TokenBar](https://github.com/Nanako0129/TokenBar) (~359★; **v1.19.0** 18 Sep night + **v1.19.1** 19 Sep ~03:29 BST) — hide tabs for unset providers; merge Antigravity IDE+CLI into one tab; Quota lens keeps on-screen history when a refresh returns empty (#348/#349/#356). [CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) (~496★; **v2.12.37**, was **v2.11.29**) — free-drag widget + magnetic dock + Claude desktop-login usage restore (#104) across ~22 commits. [junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) (~5.5k★ / 5,486) — quiet since overnight Copilot CLI `session-store.db` parser (#1346). [ItsJazii/pane](https://github.com/ItsJazii/pane) (~52★, +1). | TokenBar + CodeZeno + codeburn ZCode = highest-velocity adjacent movers overnight; tokscale flat after Fri Copilot CLI ship. |

**Positioning refresh**
- **OpenUsage.sh** = primary collision (34/35/36 + Antigravity README; deps-only since Sun product fixes).
- **ccusage** = report/parser standard (v20.0.23 Claude copied-request dedupe on top of Thu’s Luna / Antigravity / Copilot wave).
- **Splitrail** = #2 live local peer (flat/quiet since Fri 11 Sep).
- **juliantanx/aiusage** = local multi-tool dashboard peer + name collision.
- **openusage.ai** = menu-bar naming collision (Codex Swap still latest tagged).
- **TokenBar / CodeZeno / codeburn** = highest-velocity adjacent local apps overnight (provider-tab UX; widget dock + Claude desktop; ZCode live quota).
- Keep Langfuse / LiteLLM as complementary observability/gateway spend; treat **Helicone as maintenance-mode / migrate-away** adjacent.

Sources: [openusage.sh](https://openusage.sh/); [github.com/janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage); [github.com/ccusage/ccusage](https://github.com/ccusage/ccusage) v20.0.23 / #1765; [github.com/robinebers/openusage](https://github.com/robinebers/openusage); [github.com/getagentseal/codeburn](https://github.com/getagentseal/codeburn) #1347; [github.com/Nanako0129/TokenBar](https://github.com/Nanako0129/TokenBar) v1.19.0 / v1.19.1; [github.com/CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) v2.12.37; [github.com/junhoyeo/tokscale](https://github.com/junhoyeo/tokscale); [github.com/Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail); [langfuse.com migrate-from-helicone](https://langfuse.com/resources/engineering/migrate-from-helicone); GitHub API star/push metadata 2026-09-19 (Europe/London).

### Update 2026-09-20

Fresh Sunday scrape (Europe/London). Twin-risk with OpenUsage.sh unchanged; **34/35 homepage vs README** count drift still live (homepage “34 providers, and counting” + “Supported providers (35)”). Material day moves: **CodeZeno v2.12.37 → v2.12.39**; **AIUsage v1.5.17 → v1.5.18** (multi-device sync correctness); **codeburn** ships **Windows menu bar / Capacity Dock Plugins card** (#1507); **TokenBar** still tagged **v1.19.1** but **+8 commits** ahead (quota-lens restore / reopen strip); **ccusage** still **v20.0.23** (pricing-snapshot churn only today).

| Product | Delta | Vs agentUsage |
|---|---|---|
| [OpenUsage.sh](https://openusage.sh/) ([GitHub](https://github.com/janekbaraniewski/openusage), **~200★**, +1) | Homepage still **“34 providers, and counting.”** / analytics “34…” while body lists **Supported providers (35)**. Last pushes still **deps/CI only** (2026-09-14). | **Nearest twin** — category definition still ~1:1. Copy smell persists; no product surface move. |
| [ccusage](https://github.com/ccusage/ccusage) / [ccusage.com](https://ccusage.com/) (**~18.6k★** / 18,642; +~18) | Still **v20.0.23** (18 Sep). Sunday commits = **models.dev / LiteLLM pricing snapshots** only — no new tagged release. Still no Cursor in official source list. | Best **log-history / cost-report** OSS peer; thinner live quota/rate-limit TUI + key autodetection. |
| [Splitrail](https://github.com/Piebald-AI/splitrail) (**~223★**, flat) | Last product commit still **2026-09-11** (Grok 4.6 pricing). Quiet ninth day. | #2 local “live monitor” peer — multi-day quiet. |
| [AIUsage](https://github.com/juliantanx/aiusage) (**~129★**, flat) | **v1.5.18** (20 Sep ~07:20 BST): **multi-device sync** correctness across GitHub / S3/R2 / Cloud — Antigravity+Trae wire-id collisions, upsert-only namespaces that duplicated/lost records, authoritative per-device snapshots. | Rising local peer + **name collision** (AIUsage vs agentUsage); sync fix is the day’s product ship. |
| [openusage.ai](https://github.com/robinebers/openusage) (**~4.2k★** / 4,205; +~7) | Still on **v0.7.12** + **v0.7.13-beta.1** (Thu Codex Swap). Quiet overnight. | Naming collision only — do not conflate with openusage.sh. |
| Vendor UIs | **Cursor:** Spending + Usage still split. **Claude Code:** `/usage` + `/cost` unchanged. **OpenRouter:** Activity / credits still API-platform-only. | Fragmentation OpenUsage/agentUsage claim to fix. |
| Helicone / Langfuse / LiteLLM | **Helicone** (~6.2k★ / 6,166): last push still **16 Sep**. **Langfuse** (~34.8k★ / 34,834). **LiteLLM** (~59.2k★ / 59,200, **pushed today**). | Adjacent layers — prefer Langfuse/LiteLLM for new gateway/OTel work; Helicone still migrate-away adjacent. |
| Also watch (local peers) | [CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) (~**500★**, +4; **v2.12.39**, was **v2.12.37**) — **v2.12.38** theme/settings user guide + changelog; **v2.12.39** dashboard update controls + prevent duplicate WinGet update launches (published ~01:36 BST). [getagentseal/codeburn](https://github.com/getagentseal/codeburn) (~11.1k★ / 11,115) — **Windows menu bar + Capacity Dock as Plugins card** parity with macOS (#1507, ~02:43 BST) plus overnight WSL provider-root discovery / billing-route CLI filter / Claude Team-vs-Max label fix (#1062/#1486/#1492). [Nanako0129/TokenBar](https://github.com/Nanako0129/TokenBar) (~**360★**, +1; tagged **v1.19.1**) — **8 commits ahead of tag**: quota-lens restore-on-reopen + strip partial-provider-failure retention (#360/#361) — unreleased. [junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) (~5.5k★ / 5,490) — quiet since Fri Copilot CLI parser. [ItsJazii/pane](https://github.com/ItsJazii/pane) (~52★, flat). | CodeZeno + codeburn Windows tray + AIUsage sync = highest-velocity adjacent movers overnight; TokenBar unreleased quota-lens harden. |

**Positioning refresh**
- **OpenUsage.sh** = primary collision (34/35 copy drift; deps-only since mid-Sep product fixes).
- **ccusage** = report/parser standard (still v20.0.23; pricing snapshots only today).
- **Splitrail** = #2 live local peer (flat/quiet since Fri 11 Sep).
- **juliantanx/aiusage** = local multi-tool dashboard peer + name collision (**v1.5.18** sync correctness).
- **openusage.ai** = menu-bar naming collision (Codex Swap still latest tagged).
- **CodeZeno / codeburn / TokenBar** = highest-velocity adjacent local apps overnight (WinGet update UX; Windows tray Plugins card; unreleased quota-lens restore).
- Keep Langfuse / LiteLLM as complementary observability/gateway spend; treat **Helicone as maintenance-mode / migrate-away** adjacent.

Sources: [openusage.sh](https://openusage.sh/); [github.com/janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage) (~200★); [github.com/ccusage/ccusage](https://github.com/ccusage/ccusage) v20.0.23; [github.com/robinebers/openusage](https://github.com/robinebers/openusage); [github.com/juliantanx/aiusage](https://github.com/juliantanx/aiusage) v1.5.18; [github.com/getagentseal/codeburn](https://github.com/getagentseal/codeburn) #1507; [github.com/Nanako0129/TokenBar](https://github.com/Nanako0129/TokenBar) v1.19.1 + HEAD; [github.com/CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) v2.12.39; [github.com/junhoyeo/tokscale](https://github.com/junhoyeo/tokscale); [github.com/Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail); [langfuse.com migrate-from-helicone](https://langfuse.com/resources/engineering/migrate-from-helicone); GitHub API star/push metadata 2026-09-20 (Europe/London).

### Update 2026-09-21

Fresh Monday scrape (Europe/London). Twin-risk with OpenUsage.sh unchanged; **34/35 homepage vs 36 README** count drift still live. Material overnight moves: **TokenBar v1.19.1 → v1.20.0** (ships Sunday’s unreleased quota-lens harden + history hover breakdown + repair-instead-of-refuse); **CodeZeno v2.12.39 → v2.12.42** (taskbar jitter debounce); **codeburn 0.9.25** (provider/cache correctness + menu-bar/installer); **pane v0.4.53** (weekly-limit reset notify). **ccusage** still **v20.0.23** (pricing-snapshot churn only today).

| Product | Delta | Vs agentUsage |
|---|---|---|
| [OpenUsage.sh](https://openusage.sh/) ([GitHub](https://github.com/janekbaraniewski/openusage), **~200★**, flat) | Homepage still **“Supported providers (35)”** + footer **“34 providers, and counting.”**; README still **36** (incl. Antigravity). Last pushes still **deps/CI only** (2026-09-14). | **Nearest twin** — category definition still ~1:1. Copy smell persists; no product surface move (7th day of deps-only). |
| [ccusage](https://github.com/ccusage/ccusage) / [ccusage.com](https://ccusage.com/) (**~18.7k★** / 18,658; +~16) | Still **v20.0.23** (18 Sep). Monday commits = **models.dev / LiteLLM pricing snapshots** only (latest ~08:31 BST) — no new tagged release. Still no Cursor in official source list. | Best **log-history / cost-report** OSS peer; thinner live quota/rate-limit TUI + key autodetection. |
| [Splitrail](https://github.com/Piebald-AI/splitrail) (**~223★**, flat) | Last product commit still **2026-09-11** (Grok 4.6 pricing). Quiet tenth day. | #2 local “live monitor” peer — multi-day quiet. |
| [AIUsage](https://github.com/juliantanx/aiusage) (**~129★**, flat) | Still **v1.5.18** (Sun multi-device sync correctness). Quiet overnight. | Rising local peer + **name collision** (AIUsage vs agentUsage). |
| [openusage.ai](https://github.com/robinebers/openusage) (**~4.2k★** / 4,213; +~8) | Still on **v0.7.12** + **v0.7.13-beta.1** (Thu Codex Swap). Sunday pushes = dependency bumps (KeyboardShortcuts / PostHog / Sparkle). | Naming collision only — do not conflate with openusage.sh. |
| Vendor UIs | **Cursor:** Spending tab still covers pool usage, on-demand, monthly spend limits (incl. Teams admin-only toggle + Enterprise member/group overrides / Admin API) — Usage pane remains separate. **Claude Code:** `/usage` primary; `/cost` (and `/stats`) remain aliases per current docs. **OpenRouter:** Activity / usage-accounting still API-platform credits & analytics. | Fragmentation OpenUsage/agentUsage claim to fix. |
| Helicone / Langfuse / LiteLLM | **Helicone** (~6.2k★ / 6,168): last push still **16 Sep** (platform-admin / HQL cross-tenant fix). **Langfuse** (~34.9k★ / 34,880, **pushed today**, +~46). **LiteLLM** (~59.3k★ / 59,292, **pushed today**, +~92). | Adjacent layers — prefer Langfuse/LiteLLM for new gateway/OTel work; Helicone still migrate-away adjacent. |
| Also watch (local peers) | [Nanako0129/TokenBar](https://github.com/Nanako0129/TokenBar) (~**360★**, flat; **v1.20.0**, was **v1.19.1**) — history-row hover shows input/output/cache breakdown (#368); repair unwritable history instead of permanent “unavailable” (#371/#370); per-provider read failure isolation (#360); reset-time drift grouping (#369); reopen cache for quota cards / daily chart (#361/#362) — ships what was unreleased ahead-of-tag yesterday (~07:30 BST). [CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) (~**501★**, +1; **v2.12.42**, was **v2.12.39**) — trailing debounce on tray events to kill taskbar jitter (#107, ~00:51 BST). [getagentseal/codeburn](https://github.com/getagentseal/codeburn) (~11.1k★ / 11,134; +~19) — **0.9.25** (~23:16 BST Sun): ~15 provider/cache correctness fixes (locked DB ≠ “no usage”, ZCode reasoning tokens, GLM/Warp pricing, Codex credit tiers) + menu-bar/installer; then **never hold back a save for durable-source providers** (#1514). [ItsJazii/pane](https://github.com/ItsJazii/pane) (~52★; **v0.4.53**, ~09:10 BST) — notify when a weekly limit resets (#234). [junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) (~5.5k★ / 5,497) — quiet since Fri Copilot CLI parser. Also name-collision adjacent: [sylearn/AIUsage](https://github.com/sylearn/AIUsage) (~679★) — macOS subscription dashboard + coding proxies (different layer than juliantanx local web dash; docs-only today). | TokenBar v1.20.0 + CodeZeno jitter fix + codeburn 0.9.25 = highest-velocity adjacent movers overnight. |

**Positioning refresh**
- **OpenUsage.sh** = primary collision (34/35/36 copy drift; deps-only since mid-Sep product fixes — week-long quiet on product surface).
- **ccusage** = report/parser standard (still v20.0.23; pricing snapshots only today; stars still climbing).
- **Splitrail** = #2 live local peer (flat/quiet since Fri 11 Sep — tenth day).
- **juliantanx/aiusage** = local multi-tool dashboard peer + name collision (still v1.5.18).
- **openusage.ai** = menu-bar naming collision (Codex Swap still latest tagged; deps only).
- **TokenBar / CodeZeno / codeburn** = highest-velocity adjacent local apps overnight (quota-lens repair+hover; taskbar jitter; 0.9.25 correctness).
- Keep Langfuse / LiteLLM as complementary observability/gateway spend; treat **Helicone as maintenance-mode / migrate-away** adjacent.

Sources: [openusage.sh](https://openusage.sh/); [github.com/janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage) (~200★); [github.com/ccusage/ccusage](https://github.com/ccusage/ccusage) v20.0.23; [github.com/robinebers/openusage](https://github.com/robinebers/openusage); [github.com/juliantanx/aiusage](https://github.com/juliantanx/aiusage) v1.5.18; [github.com/Nanako0129/TokenBar](https://github.com/Nanako0129/TokenBar) v1.20.0; [github.com/CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) v2.12.42 / #107; [github.com/getagentseal/codeburn](https://github.com/getagentseal/codeburn) 0.9.25 / #1514; [github.com/ItsJazii/pane](https://github.com/ItsJazii/pane) v0.4.53; [cursor.com spend limits](https://cursor.com/help/account-and-billing/spend-limits); [code.claude.com costs](https://code.claude.com/docs/en/costs); [openrouter activity](https://openrouter.ai/blog/announcements/activity-dashboard/); [langfuse.com migrate-from-helicone](https://langfuse.com/resources/engineering/migrate-from-helicone); GitHub API star/push metadata 2026-09-21 (Europe/London).

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

### Open source — 11 Sep 2026

Fresh weekday GitHub pass (Europe/London). ccusage / OpenUsage ecosystem still defines the category; **live pushes** on ccusage, Codex desktop, budgetclaw, llm_cost_tracker, LiteLLM, Langfuse, Bifrost. New menu-bar / narrative / multi-tool peers (vibebuddy, brink, tokenmaxxing, ccstory).

| Repo | Stars | License | Why it matters |
|---|---:|---|---|
| [ccusage/ccusage](https://github.com/ccusage/ccusage) | ~18.5k | NOASSERTION | **Pushed today.** Category CLI — stay format-compatible. |
| [janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage) | ~192 | MIT | openusage.sh local multi-tool TUI + SQLite. Nearest product twin. |
| [robinebers/openusage](https://github.com/robinebers/openusage) | ~4.1k | MIT | openusage.ai menu-bar tracker — naming collision only. |
| [Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail) | ~220 | MIT | Real-time cross-agent monitor — number-two local peer. |
| [zhnd/lumo](https://github.com/zhnd/lumo) | ~147 | MIT | Local-first Claude Code usage/cost/session dashboard. |
| [Halloweedev/usagepal](https://github.com/Halloweedev/usagepal) | ~77 | MIT | OpenUsage menu-bar fork (Rust + Tauri). |
| [ItsJazii/pane](https://github.com/ItsJazii/pane) | ~44 | NOASSERTION | Windows tray OpenUsage port. |
| [itvincent-git/codex-usage-desktop](https://github.com/itvincent-git/codex-usage-desktop) | ~32 | MIT | **Pushed today.** Local-first Codex quota desktop app. |
| [semantic-craft/iOS-vibebuddy](https://github.com/semantic-craft/iOS-vibebuddy) | ~86 | MIT | **New fold-in / pushed today.** Tracks Claude/Codex/Grok + Cursor usage (closest multi-surface peer). |
| [semihtalii/brink](https://github.com/semihtalii/brink) | ~53 | MIT | **New fold-in.** Claude/Codex/Cursor limit "on the brink" alerts. |
| [851-labs/tokenmaxxing](https://github.com/851-labs/tokenmaxxing) | ~69 | MIT | **New fold-in.** Local CLI on ccusage that syncs usage socially. |
| [atomchung/ccstory](https://github.com/atomchung/ccstory) | ~43 | MIT | **New fold-in.** Narrative recap on top of ccusage ("story", not just bill). |
| [kenn-io/vibepulse](https://github.com/kenn-io/vibepulse) | ~53 | MIT | **New fold-in.** macOS menubar Claude Code + Codex via ccusage. |
| [ansonliam/AIUsageMonitor](https://github.com/ansonliam/AIUsageMonitor) | ~28 | MIT | **New fold-in.** Compact Windows widget (Codex/Claude/Antigravity…). |
| [Baek-Seunghyun/ai-coding-usage-card](https://github.com/Baek-Seunghyun/ai-coding-usage-card) | ~24 | MIT | Self-hosted GitHub profile SVG card from ccusage data. |
| [sculptdotfun/viberank](https://github.com/sculptdotfun/viberank) | ~116 | MIT | Public leaderboard on ccusage data. |
| [cobra91/better-ccusage](https://github.com/cobra91/better-ccusage) | ~85 | MIT | Faster multi-provider ccusage-style analyzer. |
| [Nihondo/AgentLimits](https://github.com/Nihondo/AgentLimits) | ~56 | MIT | macOS widgets for Codex/Claude limits + ccusage heatmap. |
| [agiwhitelist/tokdiet](https://github.com/agiwhitelist/tokdiet) | ~33 | MIT | Local reverse-proxy meter between agents and APIs. |
| [ofershap/cursor-usage-tracker](https://github.com/ofershap/cursor-usage-tracker) | ~33 | MIT | Self-hosted Cursor Enterprise spend + anomaly alerts. |
| [RoninForge/budgetclaw](https://github.com/RoninForge/budgetclaw) | ~8 | MIT | **Pushed today.** Local Claude Code spend monitor with hard budget caps. |
| [DataGrout/lumen](https://github.com/DataGrout/lumen) | ~11 | MIT | Real-time LLM token/cost monitor via TLS proxy or HTTP relay. |
| [sergey-homenko/llm_cost_tracker](https://github.com/sergey-homenko/llm_cost_tracker) | ~44 | MIT | **Pushed today.** Rails-native LLM cost ledger with budget guardrails. |
| [BerriAI/litellm](https://github.com/BerriAI/litellm) | ~58.5k | NOASSERTION | **Active today.** AI gateway spend/pricing — adjacent layer. |
| [langfuse/langfuse](https://github.com/langfuse/langfuse) | ~34.5k | NOASSERTION | **Active today.** OSS LLM observability + OTel. |
| [maximhq/bifrost](https://github.com/maximhq/bifrost) | ~8.0k | Apache-2.0 | **Active today.** Enterprise AI gateway (LiteLLM-adjacent). |
| [Helicone/helicone](https://github.com/Helicone/helicone) | ~6.1k | Apache-2.0 | OSS LLM observability proxy. Adjacent. |
| [ferro-labs/ai-gateway](https://github.com/ferro-labs/ai-gateway) | ~256 | Apache-2.0 | Go AI gateway with cost controls (LiteLLM alternative class). |

**Build takeaways:** openusage.sh is the collision twin; openusage.ai is menu-bar brand. Stay ccusage-compatible; differentiate on live quotas + auto-detect + UX. Multi-surface peers (vibebuddy/brink) and narrative layers (ccstory) are the new product angles beyond raw spend.

Sources: GitHub API 2026-09-11 (Europe/London).

### Open source — 12 Sep 2026

Fresh weekend GitHub API pass (Europe/London). Category still owned by **ccusage** + the two OpenUsage brands; **codeburn** remains the loud multi-tool local tracker. **Live today:** ccusage, robinebers/openusage, codeburn, otelite, AgentHarbor, claude-usage-mac, vibepulse (hardware). New fold-ins: single-CLI multi-provider trackers, OTel local dashboards, quota menubars, and hardware status surfaces.

| Repo | Stars | License | Why it matters |
|---|---:|---|---|
| [ccusage/ccusage](https://github.com/ccusage/ccusage) | ~18.5k | NOASSERTION | **Pushed today.** Category CLI — stay format-compatible. |
| [getagentseal/codeburn](https://github.com/getagentseal/codeburn) | ~11.0k | MIT | **Pushed today.** Local tracker across 37 tools/agents — star-gravity peer. |
| [robinebers/openusage](https://github.com/robinebers/openusage) | ~4.1k | MIT | **Pushed today.** openusage.ai menu-bar brand (naming collision only). |
| [janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage) | ~193 | MIT | openusage.sh local multi-tool TUI + SQLite — nearest product twin. |
| [Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail) | ~222 | MIT | Real-time cross-agent cost monitor. |
| [Halloweedev/usagepal](https://github.com/Halloweedev/usagepal) | ~77 | MIT | OpenUsage menu-bar fork (Rust + Tauri). |
| [Dicklesworthstone/coding_agent_usage_tracker](https://github.com/Dicklesworthstone/coding_agent_usage_tracker) | ~85 | NOASSERTION | **New fold-in.** Single CLI for remaining quotas across Codex/Claude/Gemini/Cursor/Copilot. |
| [Rodiun/frugon](https://github.com/Rodiun/frugon) | ~212 | MIT | **New fold-in.** Local LLM bill-leak analyzer (where spend escapes). |
| [mag123c/toktrack](https://github.com/mag123c/toktrack) | ~189 | MIT | **New fold-in.** Ultra-fast Claude Code token/cost tracker. |
| [planetf1/otelite](https://github.com/planetf1/otelite) | ~88 | NOASSERTION | **New fold-in / pushed today.** Single-binary OTel receiver + local LLM dashboard. |
| [majiayu000/quotabar](https://github.com/majiayu000/quotabar) | ~51 | MIT | **New fold-in.** Tauri v2 menubar for Claude/Codex/Cursor/Antigravity quotas. |
| [stevemcqueenz/claude-notch-tracker](https://github.com/stevemcqueenz/claude-notch-tracker) | ~34 | NOASSERTION | **New fold-in.** Mac notch / Dynamic Island live Claude+Codex usage. |
| [saeedkolivand/claude-usage-mac](https://github.com/saeedkolivand/claude-usage-mac) | ~15 | NOASSERTION | **New fold-in / pushed today.** Claude Code menu-bar + desktop widget. |
| [abhiunix/AgentHarbor](https://github.com/abhiunix/AgentHarbor) | ~12 | MIT | **New fold-in / pushed today.** Native multi-agent rate-limit + session usage app. |
| [fagemx/edda](https://github.com/fagemx/edda) | ~36 | Apache-2.0 | **New fold-in / pushed today.** Tamper-evident local ledger for agent decisions (adjacent audit layer). |
| [niclasvestlund-YT/vibepulse](https://github.com/niclasvestlund-YT/vibepulse) | ~194 | MIT | **New fold-in / pushed today.** ESP32 AMOLED hardware surface for Claude/Codex usage + "needs you" alerts (distinct from kenn-io/vibepulse menubar). |
| [ticpu/ccusage-statusline-rs](https://github.com/ticpu/ccusage-statusline-rs) | ~16 | MIT | **New fold-in.** Rust ccusage statusline (perf niche). |
| [semantic-craft/iOS-vibebuddy](https://github.com/semantic-craft/iOS-vibebuddy) | ~86 | MIT | Multi-surface Claude/Codex/Grok/Cursor tracker. |
| [semihtalii/brink](https://github.com/semihtalii/brink) | ~53 | MIT | Limit "on the brink" alerts. |
| [BerriAI/litellm](https://github.com/BerriAI/litellm) | ~58.5k | NOASSERTION | Gateway spend/pricing — adjacent. |
| [langfuse/langfuse](https://github.com/langfuse/langfuse) | ~34.5k | NOASSERTION | OSS LLM observability + OTel. |

**Build takeaways:** Stay ccusage-compatible; differentiate on live quotas + auto-detect + UX. Watch `coding_agent_usage_tracker` / `frugon` / `toktrack` as CLI peers, `otelite` for local OTel dashboards, and menubar/notch/hardware surfaces (quotabar, claude-notch-tracker, ESP32 vibepulse) as distribution angles — not core parsers.

Sources: GitHub API 2026-09-12 (Europe/London).

### Open source — 13 Sep 2026

Fresh Sunday GitHub API pass (Europe/London). Category still owned by **ccusage** + the two OpenUsage brands; **codeburn** remains the loud multi-tool local tracker. **Live today:** ccusage, codeburn, AgentHarbor, edda, iOS-vibebuddy, litellm, langfuse (plus yesterday's openusage / otelite / vibepulse / claude-usage-mac). New fold-ins: **ccusage leaderboards**, **better-ccusage parsers**, **macOS quota widgets**, **Codex desktop trackers**, and narrative / social usage surfaces.

| Repo | Stars | License | Why it matters |
|---|---:|---|---|
| [ccusage/ccusage](https://github.com/ccusage/ccusage) | ~18.5k | NOASSERTION | **Pushed today.** Category CLI — stay format-compatible. |
| [getagentseal/codeburn](https://github.com/getagentseal/codeburn) | ~11.0k | MIT | **Pushed today.** Local tracker across 37 tools/agents — star-gravity peer. |
| [robinebers/openusage](https://github.com/robinebers/openusage) | ~4.1k | MIT | openusage.ai menu-bar brand (naming collision only). |
| [janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage) | ~194 | MIT | openusage.sh local multi-tool TUI + SQLite — nearest product twin. |
| [Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail) | ~223 | MIT | Real-time cross-agent cost monitor. |
| [sculptdotfun/viberank](https://github.com/sculptdotfun/viberank) | ~116 | MIT | **New fold-in.** Public AI-coding usage leaderboard fed by ccusage data (`npx viberank-cli`). |
| [cobra91/better-ccusage](https://github.com/cobra91/better-ccusage) | ~85 | MIT | **New fold-in.** Faster multi-provider analyzer over local JSONL (Claude/Droid/OpenCode/Codex/…). |
| [Dicklesworthstone/coding_agent_usage_tracker](https://github.com/Dicklesworthstone/coding_agent_usage_tracker) | ~85 | NOASSERTION | Single CLI for remaining quotas across Codex/Claude/Gemini/Cursor/Copilot. |
| [Rodiun/frugon](https://github.com/Rodiun/frugon) | ~212 | MIT | Local LLM bill-leak analyzer (where spend escapes). |
| [mag123c/toktrack](https://github.com/mag123c/toktrack) | ~189 | MIT | Ultra-fast Claude Code token/cost tracker. |
| [851-labs/tokenmaxxing](https://github.com/851-labs/tokenmaxxing) | ~70 | MIT | **New fold-in.** Local CLI on ccusage that syncs token usage socially. |
| [Nihondo/AgentLimits](https://github.com/Nihondo/AgentLimits) | ~58 | MIT | **New fold-in.** macOS widgets for Codex/Claude limits + ccusage tokens/cost/heatmap. |
| [Halloweedev/usagepal](https://github.com/Halloweedev/usagepal) | ~77 | MIT | OpenUsage menu-bar fork (Rust + Tauri). |
| [planetf1/otelite](https://github.com/planetf1/otelite) | ~88 | NOASSERTION | Single-binary OTel receiver + local LLM dashboard. |
| [majiayu000/quotabar](https://github.com/majiayu000/quotabar) | ~51 | MIT | Tauri v2 menubar for Claude/Codex/Cursor/Antigravity quotas. |
| [atomchung/ccstory](https://github.com/atomchung/ccstory) | ~43 | MIT | **New fold-in.** Narrative Claude Code usage recap — "ccusage tells the bill, ccstory tells the story." |
| [itvincent-git/codex-usage-desktop](https://github.com/itvincent-git/codex-usage-desktop) | ~35 | MIT | **New fold-in / pushed recently.** Local-first Codex quota + session desktop app (macOS/Windows). |
| [stevemcqueenz/claude-notch-tracker](https://github.com/stevemcqueenz/claude-notch-tracker) | ~34 | NOASSERTION | Mac notch / Dynamic Island live Claude+Codex usage. |
| [saeedkolivand/claude-usage-mac](https://github.com/saeedkolivand/claude-usage-mac) | ~15 | NOASSERTION | Claude Code menu-bar + desktop widget. |
| [abhiunix/AgentHarbor](https://github.com/abhiunix/AgentHarbor) | ~12 | MIT | **Pushed today.** Native multi-agent rate-limit + session usage app. |
| [fagemx/edda](https://github.com/fagemx/edda) | ~36 | Apache-2.0 | **Pushed today.** Tamper-evident local ledger for agent decisions (adjacent audit). |
| [niclasvestlund-YT/vibepulse](https://github.com/niclasvestlund-YT/vibepulse) | ~195 | MIT | ESP32 AMOLED hardware surface for Claude/Codex usage + "needs you" alerts. |
| [semantic-craft/iOS-vibebuddy](https://github.com/semantic-craft/iOS-vibebuddy) | ~87 | MIT | **Pushed today.** Multi-surface Claude/Codex/Grok/Cursor tracker. |
| [semihtalii/brink](https://github.com/semihtalii/brink) | ~53 | MIT | Limit "on the brink" alerts. |
| [PioneerSquareLabs/metergraph](https://github.com/PioneerSquareLabs/metergraph) | ~4 | Apache-2.0 | **New fold-in / watch.** Content-blind self-hosted cost tracking by function + price catalog. |
| [AgentOps-AI/agentops](https://github.com/AgentOps-AI/agentops) | ~5.8k | MIT | **New fold-in (adjacent).** Agent monitoring / cost SDK — not local CLI twin, but popular instrumented cost layer. |
| [BerriAI/litellm](https://github.com/BerriAI/litellm) | ~58.6k | NOASSERTION | **Pushed today.** Gateway spend/pricing — adjacent. |
| [langfuse/langfuse](https://github.com/langfuse/langfuse) | ~34.5k | NOASSERTION | **Pushed today.** OSS LLM observability + OTel. |

**Build takeaways:** Stay ccusage-compatible; differentiate on live quotas + auto-detect + UX. Watch `better-ccusage` / `viberank` / `tokenmaxxing` as ecosystem gravity around ccusage formats; menubar/widget/desktop peers (AgentLimits, codex-usage-desktop, quotabar) for distribution — not core parsers. Keep Langfuse/LiteLLM/AgentOps as adjacent observability, not product twins.

Sources: GitHub API 2026-09-13 (Europe/London).


### Open source — 14 Sep 2026

Fresh Monday GitHub API pass (Europe/London). Category still owned by **ccusage** + the two OpenUsage brands; **codeburn** remains the loud multi-tool local tracker. **Live today:** ccusage, openusage.sh (deps + yesterday's Codex/Crush fixes), codeburn, otelite (**productivity + tool-mix views**), AIUsage **v1.5.16**, TokenBar, usagepal, frugon, AgentHarbor, iOS-vibebuddy, LiteLLM, Langfuse. **New fold-ins:** high-star Claude Code monitors (`Claude-Code-Usage-Monitor`, Clawdmeter, ccseva, claude-lens, claude-code-otel).

| Repo | Stars | License | Why it matters |
|---|---:|---|---|
| [ccusage/ccusage](https://github.com/ccusage/ccusage) | ~18.5k | NOASSERTION | **Pushed today.** Category CLI — stay format-compatible. |
| [getagentseal/codeburn](https://github.com/getagentseal/codeburn) | ~11.0k | MIT | **Pushed today.** Local tracker across 37 tools — star-gravity peer (CI/docs churn). |
| [robinebers/openusage](https://github.com/robinebers/openusage) | ~4.2k | MIT | openusage.ai menu-bar brand (naming collision only). |
| [janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage) | ~195 | MIT | **Active today** (deps). openusage.sh local multi-tool TUI + SQLite — nearest product twin. Product fixes yesterday: Codex 0.153+ + Crush timestamps. |
| [Maciek-roboblog/Claude-Code-Usage-Monitor](https://github.com/Maciek-roboblog/Claude-Code-Usage-Monitor) | ~8.7k | MIT | **New fold-in.** Real-time Claude Code usage monitor with predictions/warnings — highest-star Claude-only live monitor. |
| [HermannBjorgvin/Clawdmeter](https://github.com/HermannBjorgvin/Clawdmeter) | ~2.2k | (check) | **New fold-in.** ESP32 desk dashboard for Claude Code usage — hardware distribution angle. |
| [Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail) | ~223 | MIT | Real-time cross-agent cost monitor. |
| [juliantanx/aiusage](https://github.com/juliantanx/aiusage) | ~126 | MIT | **v1.5.16 today** — local web dashboard across 20+ tools + name collision. |
| [junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) | ~5.4k | MIT | Terminal token tracker + leaderboard (Codex/Antigravity fixes overnight). |
| [Iamshankhadeep/ccseva](https://github.com/Iamshankhadeep/ccseva) | ~805 | MIT | **New fold-in.** macOS menu bar for live Claude Code usage. |
| [ColeMurray/claude-code-otel](https://github.com/ColeMurray/claude-code-otel) | ~498 | MIT | **New fold-in.** Observability stack for Claude Code usage/perf/cost (OTel path). |
| [foyzulkarim/claude-lens](https://github.com/foyzulkarim/claude-lens) | ~246 | MIT | **New fold-in.** Local dashboard: sessions, token costs, cache, tool calls, daily breakdowns. |
| [frankchiu-dev/claude-codex-usage-dashboard](https://github.com/frankchiu-dev/claude-codex-usage-dashboard) | ~173 | MIT | **New fold-in.** Local Windows dashboard for Claude Code + Codex limits. |
| [planetf1/otelite](https://github.com/planetf1/otelite) | ~88 | NOASSERTION | **Pushed today.** Productivity view (commits/PRs/LOC per tool) + daily tool-mix token/cost views — OTel local dashboard deepening. |
| [Rodiun/frugon](https://github.com/Rodiun/frugon) | ~212 | MIT | **Pushed today.** Local LLM bill-leak analyzer. |
| [mag123c/toktrack](https://github.com/mag123c/toktrack) | ~189 | MIT | Ultra-fast Claude Code token/cost tracker. |
| [Halloweedev/usagepal](https://github.com/Halloweedev/usagepal) | ~77 | MIT | **Pushed today.** OpenUsage menu-bar fork (Rust + Tauri). |
| [Nanako0129/TokenBar](https://github.com/Nanako0129/TokenBar) | ~348 | MIT | **Pushed today.** macOS menu-bar quota peer. |
| [tddworks/ClaudeBar](https://github.com/tddworks/ClaudeBar) | ~1.5k | (check) | macOS menu-bar Claude quota peer. |
| [deviffyy/OpenQuota](https://github.com/deviffyy/OpenQuota) | ~204 | MIT | Cross-platform limits/spend tracker. |
| [Dicklesworthstone/coding_agent_usage_tracker](https://github.com/Dicklesworthstone/coding_agent_usage_tracker) | ~85 | NOASSERTION | Single CLI for remaining quotas across Codex/Claude/Gemini/Cursor/Copilot. |
| [sculptdotfun/viberank](https://github.com/sculptdotfun/viberank) | ~116 | MIT | Public AI-coding usage leaderboard on ccusage data. |
| [cobra91/better-ccusage](https://github.com/cobra91/better-ccusage) | ~85 | MIT | Faster multi-provider JSONL analyzer. |
| [851-labs/tokenmaxxing](https://github.com/851-labs/tokenmaxxing) | ~71 | MIT | Local CLI on ccusage that syncs usage socially. |
| [Nihondo/AgentLimits](https://github.com/Nihondo/AgentLimits) | ~58 | MIT | macOS widgets for Codex/Claude limits + ccusage heatmap. |
| [majiayu000/quotabar](https://github.com/majiayu000/quotabar) | ~52 | MIT | Tauri v2 menubar for Claude/Codex/Cursor/Antigravity quotas. |
| [abhiunix/AgentHarbor](https://github.com/abhiunix/AgentHarbor) | ~12 | MIT | **Pushed today.** Native multi-agent rate-limit + session usage app. |
| [semantic-craft/iOS-vibebuddy](https://github.com/semantic-craft/iOS-vibebuddy) | ~87 | MIT | **Pushed today.** Multi-surface Claude/Codex/Grok/Cursor tracker. |
| [zhnd/lumo](https://github.com/zhnd/lumo) | ~147 | MIT | Local-first Claude Code usage/cost/session dashboard. |
| [BerriAI/litellm](https://github.com/BerriAI/litellm) | ~58.7k | NOASSERTION | **Pushed today.** Gateway spend/pricing — adjacent. |
| [langfuse/langfuse](https://github.com/langfuse/langfuse) | ~34.6k | NOASSERTION | **Pushed today.** OSS LLM observability + OTel. |
| [Helicone/helicone](https://github.com/Helicone/helicone) | ~6.2k | Apache-2.0 | OSS LLM observability proxy. Adjacent. |

**Build takeaways:** Stay ccusage-compatible; differentiate on live quotas + auto-detect + UX vs openusage.sh. Fold high-star Claude monitors (Usage-Monitor / Clawdmeter / ccseva / claude-lens) as UX references for predictions, hardware, and local dashboards — not multi-provider twins. `otelite`'s new productivity + tool-mix views are the sharpest adjacent OTel dashboard move today. Keep Langfuse/LiteLLM/Helicone as complementary observability, not product twins.

Sources: GitHub API 2026-09-14 (Europe/London).

### Open source — 15 Sep 2026

Fresh Tuesday GitHub API pass (Europe/London). Category still owned by **ccusage** + the two OpenUsage brands; **codeburn** remains the loud multi-tool local tracker. **Live today:** ccusage (LiteLLM pricing snapshots), openusage.ai (**Claude Swap account discovery** + Devin quota fix → **v0.7.12-beta.2**), tokscale **4.17.0** (Cline modelInfo + scan paths), otelite (**human response latency** per tool/hour + **v0.1.136**), AIUsage **v1.5.17**, usagepal (**per-provider account switcher** + OpenCode-Go names → **v0.7.76-beta.1**), TokenBar, AgentHarbor, iOS-vibebuddy, toktrack, LiteLLM, Langfuse. **New fold-in:** `ItsJazii/pane` (~48★) — Windows tray OpenUsage port (Claude/Codex/Cursor/Copilot/Kimi/Grok +15).

| Repo | Stars | License | Why it matters |
|---|---:|---|---|
| [ccusage/ccusage](https://github.com/ccusage/ccusage) | ~18.6k | NOASSERTION | **Pushed today.** Category CLI (LiteLLM pricing churn) — stay format-compatible. |
| [getagentseal/codeburn](https://github.com/getagentseal/codeburn) | ~11.0k | MIT | Local tracker across 37 tools — star-gravity peer. |
| [robinebers/openusage](https://github.com/robinebers/openusage) | ~4.2k | MIT | **Pushed today.** openusage.ai menu-bar — Claude Swap multi-account + Devin weekly-quota fix (**v0.7.12-beta.2**). |
| [janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage) | ~195 | MIT | openusage.sh local multi-tool TUI + SQLite — nearest product twin. |
| [ItsJazii/pane](https://github.com/ItsJazii/pane) | ~48 | NOASSERTION | **New fold-in.** Windows tray OpenUsage port across 15+ plans — platform-coverage peer. |
| [Maciek-roboblog/Claude-Code-Usage-Monitor](https://github.com/Maciek-roboblog/Claude-Code-Usage-Monitor) | ~8.7k | MIT | Real-time Claude Code usage monitor with predictions/warnings. |
| [HermannBjorgvin/Clawdmeter](https://github.com/HermannBjorgvin/Clawdmeter) | ~2.2k | (check) | ESP32 desk dashboard for Claude Code usage — hardware distribution angle. |
| [Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail) | ~223 | MIT | Real-time cross-agent cost monitor. |
| [juliantanx/aiusage](https://github.com/juliantanx/aiusage) | ~127 | MIT | **v1.5.17 today** — local web dashboard across 20+ tools. |
| [junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) | ~5.4k | MIT | **4.17.0 today** — Cline per-message modelInfo + documented scan paths. |
| [Iamshankhadeep/ccseva](https://github.com/Iamshankhadeep/ccseva) | ~807 | MIT | macOS menu bar for live Claude Code usage. |
| [ColeMurray/claude-code-otel](https://github.com/ColeMurray/claude-code-otel) | ~499 | MIT | Observability stack for Claude Code usage/perf/cost (OTel path). |
| [CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) | ~481 | MIT | Windows taskbar Claude Code usage monitor. |
| [leeguooooo/claude-code-usage-bar](https://github.com/leeguooooo/claude-code-usage-bar) | ~373 | MIT | Lightweight Claude Code statusLine (5h/7d limits, cache age). |
| [foyzulkarim/claude-lens](https://github.com/foyzulkarim/claude-lens) | ~247 | MIT | Local dashboard: sessions, token costs, cache, tool calls, daily breakdowns. |
| [frankchiu-dev/claude-codex-usage-dashboard](https://github.com/frankchiu-dev/claude-codex-usage-dashboard) | ~173 | MIT | Local Windows dashboard for Claude Code + Codex limits. |
| [planetf1/otelite](https://github.com/planetf1/otelite) | ~89 | NOASSERTION | **Pushed today.** Human response-latency gaps per tool/hour + **v0.1.136** — OTel local dashboard deepening. |
| [Rodiun/frugon](https://github.com/Rodiun/frugon) | ~212 | MIT | Local LLM bill-leak analyzer. |
| [mag123c/toktrack](https://github.com/mag123c/toktrack) | ~189 | MIT | **Pushed today.** Ultra-fast Claude Code token/cost tracker. |
| [Halloweedev/usagepal](https://github.com/Halloweedev/usagepal) | ~79 | MIT | **Pushed today.** OpenUsage menu-bar fork — per-provider account switcher + OpenCode-Go names (**v0.7.76-beta.1**). |
| [Nanako0129/TokenBar](https://github.com/Nanako0129/TokenBar) | ~348 | MIT | **Pushed today.** macOS menu-bar quota peer. |
| [tddworks/ClaudeBar](https://github.com/tddworks/ClaudeBar) | ~1.5k | (check) | macOS menu-bar Claude quota peer. |
| [deviffyy/OpenQuota](https://github.com/deviffyy/OpenQuota) | ~204 | MIT | Cross-platform limits/spend tracker. |
| [Dicklesworthstone/coding_agent_usage_tracker](https://github.com/Dicklesworthstone/coding_agent_usage_tracker) | ~85 | NOASSERTION | Single CLI for remaining quotas across Codex/Claude/Gemini/Cursor/Copilot. |
| [sculptdotfun/viberank](https://github.com/sculptdotfun/viberank) | ~117 | MIT | Public AI-coding usage leaderboard on ccusage data. |
| [cobra91/better-ccusage](https://github.com/cobra91/better-ccusage) | ~85 | MIT | Faster multi-provider JSONL analyzer. |
| [851-labs/tokenmaxxing](https://github.com/851-labs/tokenmaxxing) | ~71 | MIT | Local CLI on ccusage that syncs usage socially. |
| [Nihondo/AgentLimits](https://github.com/Nihondo/AgentLimits) | ~59 | MIT | macOS widgets for Codex/Claude limits + ccusage heatmap. |
| [majiayu000/quotabar](https://github.com/majiayu000/quotabar) | ~52 | MIT | Tauri v2 menubar for Claude/Codex/Cursor/Antigravity quotas. |
| [abhiunix/AgentHarbor](https://github.com/abhiunix/AgentHarbor) | ~12 | MIT | **Pushed today.** Native multi-agent rate-limit + session usage app. |
| [semantic-craft/iOS-vibebuddy](https://github.com/semantic-craft/iOS-vibebuddy) | ~87 | MIT | **Pushed today.** Multi-surface Claude/Codex/Grok/Cursor tracker (retired-UI cleanup). |
| [zhnd/lumo](https://github.com/zhnd/lumo) | ~147 | MIT | Local-first Claude Code usage/cost/session dashboard. |
| [BerriAI/litellm](https://github.com/BerriAI/litellm) | ~58.8k | NOASSERTION | **Pushed today.** Gateway spend/pricing — adjacent. |
| [langfuse/langfuse](https://github.com/langfuse/langfuse) | ~34.6k | NOASSERTION | **Pushed today.** OSS LLM observability + OTel. |
| [Helicone/helicone](https://github.com/Helicone/helicone) | ~6.2k | Apache-2.0 | OSS LLM observability proxy. Adjacent. |

**Build takeaways:** Stay ccusage-compatible; differentiate on live quotas + auto-detect + UX vs openusage.sh. Today's openusage.ai multi-account (Claude Swap) + usagepal account switcher + `pane` Windows tray coverage are the sharpest cross-platform quota UX signals. `otelite`'s human-latency view is the best adjacent OTel dashboard move. Keep Langfuse/LiteLLM/Helicone as complementary observability, not product twins.

Sources: GitHub API 2026-09-15 (Europe/London).

### Open source — 16 Sep 2026

Fresh Wednesday GitHub API pass (Europe/London). Category still owned by **ccusage** + the two OpenUsage brands; **codeburn** remains the loud multi-tool local tracker. **Live today:** ccusage (timezone/`--since`/`--until` CLI guards + session-ID dedupe tests), openusage.ai (**Codex Swap** multi-account → follows yesterday's Claude Swap), codeburn (menubar early-reset dock band removed; Codex cumulative-usage request preservation), otelite (**Codex sub-agent analytics** + context-composition panels → **v0.1.142**), CodeZeno Usage-Monitor (**v2.11.28** — multi Claude/Codex accounts + custom refresh), claude-code-usage-bar (predict profile config-dir keys → **3.43.2** path), AIUsage (CodeBuddy function-call usage), TokenBar (legend width clamp), quotabar (Codex weekly capacity model samples), AgentHarbor, iOS-vibebuddy, LiteLLM, Langfuse. `pane` Windows tray still the coverage peer (~48★).

| Repo | Stars | License | Why it matters |
|---|---:|---|---|
| [ccusage/ccusage](https://github.com/ccusage/ccusage) | ~18.6k | NOASSERTION | **Pushed today.** Category CLI — reject bad `--timezone` / inverted `--since`--`--until`; session-ID dedupe tests. Stay format-compatible. |
| [getagentseal/codeburn](https://github.com/getagentseal/codeburn) | ~11.0k | MIT | **Pushed today.** Local tracker across 37 tools — menubar early-reset band dropped; Codex request identity when cumulative usage missing. |
| [robinebers/openusage](https://github.com/robinebers/openusage) | ~4.2k | MIT | **Pushed today.** openusage.ai menu-bar — **Codex Swap** account support (multi-account parity with Claude Swap). |
| [janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage) | ~196 | MIT | openusage.sh local multi-tool TUI + SQLite — nearest product twin. |
| [ItsJazii/pane](https://github.com/ItsJazii/pane) | ~48 | NOASSERTION | Windows tray OpenUsage port across 15+ plans — platform-coverage peer. |
| [Maciek-roboblog/Claude-Code-Usage-Monitor](https://github.com/Maciek-roboblog/Claude-Code-Usage-Monitor) | ~8.7k | MIT | Real-time Claude Code usage monitor with predictions/warnings. |
| [HermannBjorgvin/Clawdmeter](https://github.com/HermannBjorgvin/Clawdmeter) | ~2.2k | (check) | ESP32 desk dashboard for Claude Code usage — hardware distribution angle. |
| [Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail) | ~223 | MIT | Real-time cross-agent cost monitor. |
| [juliantanx/aiusage](https://github.com/juliantanx/aiusage) | ~128 | MIT | **Pushed today.** CodeBuddy CLI per-request usage on function_call lines (post **v1.5.17**). |
| [junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) | ~5.4k | MIT | Still on **4.17.0** — Cline modelInfo + scan paths. |
| [Iamshankhadeep/ccseva](https://github.com/Iamshankhadeep/ccseva) | ~807 | MIT | macOS menu bar for live Claude Code usage. |
| [ColeMurray/claude-code-otel](https://github.com/ColeMurray/claude-code-otel) | ~500 | MIT | Observability stack for Claude Code usage/perf/cost (OTel path). |
| [CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) | ~487 | MIT | **v2.11.28 today** — multi Claude Code + Codex accounts; custom refresh intervals. |
| [leeguooooo/claude-code-usage-bar](https://github.com/leeguooooo/claude-code-usage-bar) | ~374 | MIT | **Pushed today.** Predict profiles keyed by session config dir — statusLine quota UX. |
| [foyzulkarim/claude-lens](https://github.com/foyzulkarim/claude-lens) | ~247 | MIT | Local dashboard: sessions, token costs, cache, tool calls, daily breakdowns. |
| [frankchiu-dev/claude-codex-usage-dashboard](https://github.com/frankchiu-dev/claude-codex-usage-dashboard) | ~173 | MIT | Local Windows dashboard for Claude Code + Codex limits. |
| [planetf1/otelite](https://github.com/planetf1/otelite) | ~90 | NOASSERTION | **Pushed today.** Codex sub-agent analytics per session + context-composition panels (**v0.1.142**) — sharpest adjacent OTel dashboard move. |
| [Rodiun/frugon](https://github.com/Rodiun/frugon) | ~212 | MIT | Local LLM bill-leak analyzer. |
| [mag123c/toktrack](https://github.com/mag123c/toktrack) | ~189 | MIT | Ultra-fast Claude Code token/cost tracker. |
| [Halloweedev/usagepal](https://github.com/Halloweedev/usagepal) | ~80 | MIT | OpenUsage menu-bar fork — per-provider account switcher. |
| [Nanako0129/TokenBar](https://github.com/Nanako0129/TokenBar) | ~350 | MIT | **Pushed today.** macOS menu-bar quota peer — legend width clamp for long model names. |
| [tddworks/ClaudeBar](https://github.com/tddworks/ClaudeBar) | ~1.5k | (check) | macOS menu-bar Claude quota peer. |
| [deviffyy/OpenQuota](https://github.com/deviffyy/OpenQuota) | ~208 | MIT | Cross-platform limits/spend tracker. |
| [Dicklesworthstone/coding_agent_usage_tracker](https://github.com/Dicklesworthstone/coding_agent_usage_tracker) | ~85 | NOASSERTION | Single CLI for remaining quotas across Codex/Claude/Gemini/Cursor/Copilot. |
| [sculptdotfun/viberank](https://github.com/sculptdotfun/viberank) | ~117 | MIT | Public AI-coding usage leaderboard on ccusage data. |
| [cobra91/better-ccusage](https://github.com/cobra91/better-ccusage) | ~85 | MIT | Faster multi-provider JSONL analyzer. |
| [851-labs/tokenmaxxing](https://github.com/851-labs/tokenmaxxing) | ~72 | MIT | Local CLI on ccusage that syncs usage socially. |
| [Nihondo/AgentLimits](https://github.com/Nihondo/AgentLimits) | ~59 | MIT | macOS widgets for Codex/Claude limits + ccusage heatmap. |
| [majiayu000/quotabar](https://github.com/majiayu000/quotabar) | ~52 | MIT | **Pushed today.** Tauri v2 menubar — Codex weekly capacity from per-model samples (Astra/Sol focus). |
| [abhiunix/AgentHarbor](https://github.com/abhiunix/AgentHarbor) | ~12 | MIT | **Pushed today.** Native multi-agent rate-limit + session usage app. |
| [semantic-craft/iOS-vibebuddy](https://github.com/semantic-craft/iOS-vibebuddy) | ~88 | MIT | **Pushed today.** Multi-surface Claude/Codex/Grok/Cursor tracker (agent-CLI settings / build 36). |
| [zhnd/lumo](https://github.com/zhnd/lumo) | ~147 | MIT | Local-first Claude Code usage/cost/session dashboard. |
| [BerriAI/litellm](https://github.com/BerriAI/litellm) | ~58.9k | NOASSERTION | **Pushed today.** Gateway spend/pricing — adjacent. |
| [langfuse/langfuse](https://github.com/langfuse/langfuse) | ~34.7k | NOASSERTION | **Pushed today.** OSS LLM observability + OTel. |
| [Helicone/helicone](https://github.com/Helicone/helicone) | ~6.2k | Apache-2.0 | OSS LLM observability proxy. Adjacent. |

**Build takeaways:** Stay ccusage-compatible; differentiate on live quotas + auto-detect + UX vs openusage.sh. Today's openusage.ai **Codex Swap**, CodeZeno multi-account **v2.11.28**, and `otelite` Codex sub-agent analytics are the sharpest quota/OTel UX signals. Keep Langfuse/LiteLLM/Helicone as complementary observability, not product twins.

Sources: GitHub API 2026-09-16 (Europe/London).

### Open source — 17 Sep 2026

Fresh Thursday GitHub API pass (Europe/London). Category still owned by **ccusage** + the two OpenUsage brands; **codeburn** remains the loud multi-tool local tracker. **Live today:** ccusage (LiteLLM pricing snapshots + Renovate/Nix/Bun gate fixes), codeburn (**Antigravity** — parse tool calls / bash / MCP / skills from SQLite steps), CodeZeno Usage-Monitor (**v2.11.29**), quotabar (Codex local-capacity empty-state + cost-tile UX; Developer ID / hardened runtime release chore), AgentHarbor (traffic snapshot), iOS-vibebuddy (speech-pause resume fixes), LiteLLM, Langfuse (transcript-from-observations). **New fold-ins:** `alibaba/loongsuite-pilot` (~187★, Apache-2.0) — local-first OTel collector for Claude Code / Codex / Cursor (token/cost/traces); `seakee/CPA-Manager-Plus` (~3.5k★, MIT) — self-hosted CLIProxyAPI / gateway usage-cost-quota panel (adjacent gateway ops, not desktop twin).

| Repo | Stars | License | Why it matters |
|---|---:|---|---|
| [ccusage/ccusage](https://github.com/ccusage/ccusage) | ~18.6k | NOASSERTION | **Pushed today.** Category CLI — LiteLLM pricing snapshots; Renovate Nix/Bun gates. Stay format-compatible. |
| [getagentseal/codeburn](https://github.com/getagentseal/codeburn) | ~11.1k | MIT | **Pushed today.** Local tracker across many tools — **Antigravity** SQLite step parsing (tool/bash/MCP/skills). |
| [robinebers/openusage](https://github.com/robinebers/openusage) | ~4.2k | MIT | openusage.ai menu-bar — Codex Swap still the freshest multi-account UX (yesterday). |
| [janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage) | ~197 | MIT | openusage.sh local multi-tool TUI + SQLite — nearest product twin. |
| [ItsJazii/pane](https://github.com/ItsJazii/pane) | ~50 | NOASSERTION | Windows tray OpenUsage port across 15+ plans — platform-coverage peer. |
| [Maciek-roboblog/Claude-Code-Usage-Monitor](https://github.com/Maciek-roboblog/Claude-Code-Usage-Monitor) | ~8.7k | MIT | Real-time Claude Code usage monitor with predictions/warnings. |
| [HermannBjorgvin/Clawdmeter](https://github.com/HermannBjorgvin/Clawdmeter) | ~2.2k | (check) | ESP32 desk dashboard for Claude Code usage — hardware distribution angle. |
| [Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail) | ~223 | MIT | Real-time cross-agent cost monitor. |
| [juliantanx/aiusage](https://github.com/juliantanx/aiusage) | ~128 | MIT | CodeBuddy CLI per-request usage (post **v1.5.17**). |
| [junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) | ~5.5k | MIT | Still on recent 4.x — Cline modelInfo + scan paths. |
| [Iamshankhadeep/ccseva](https://github.com/Iamshankhadeep/ccseva) | ~808 | MIT | macOS menu bar for live Claude Code usage. |
| [ColeMurray/claude-code-otel](https://github.com/ColeMurray/claude-code-otel) | ~501 | MIT | Observability stack for Claude Code usage/perf/cost (OTel path). |
| [CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) | ~489 | MIT | **v2.11.29 today** — multi Claude Code + Codex accounts; custom refresh. |
| [leeguooooo/claude-code-usage-bar](https://github.com/leeguooooo/claude-code-usage-bar) | ~374 | MIT | Predict profiles keyed by session config dir — statusLine quota UX. |
| [foyzulkarim/claude-lens](https://github.com/foyzulkarim/claude-lens) | ~247 | MIT | Local dashboard: sessions, token costs, cache, tool calls, daily breakdowns. |
| [frankchiu-dev/claude-codex-usage-dashboard](https://github.com/frankchiu-dev/claude-codex-usage-dashboard) | ~173 | MIT | Local Windows dashboard for Claude Code + Codex limits. |
| [planetf1/otelite](https://github.com/planetf1/otelite) | ~91 | NOASSERTION | Codex sub-agent analytics + context-composition panels (**v0.1.142** path). |
| [alibaba/loongsuite-pilot](https://github.com/alibaba/loongsuite-pilot) | ~187 | Apache-2.0 | **New fold-in / pushed today.** Local-first OTel for Claude Code/Codex/Cursor — token/cost/traces/security; gateway span nesting. |
| [seakee/CPA-Manager-Plus](https://github.com/seakee/CPA-Manager-Plus) | ~3.5k | MIT | **New fold-in.** Self-hosted CLIProxyAPI / gateway usage-cost-quota + account health panel — adjacent ops, not desktop twin. |
| [Rodiun/frugon](https://github.com/Rodiun/frugon) | ~212 | MIT | Local LLM bill-leak analyzer. |
| [mag123c/toktrack](https://github.com/mag123c/toktrack) | ~189 | MIT | Ultra-fast Claude Code token/cost tracker. |
| [Halloweedev/usagepal](https://github.com/Halloweedev/usagepal) | ~82 | MIT | OpenUsage menu-bar fork — per-provider account switcher. |
| [Nanako0129/TokenBar](https://github.com/Nanako0129/TokenBar) | ~354 | MIT | macOS menu-bar quota peer — legend width clamp still recent. |
| [tddworks/ClaudeBar](https://github.com/tddworks/ClaudeBar) | ~1.5k | (check) | macOS menu-bar Claude quota peer. |
| [deviffyy/OpenQuota](https://github.com/deviffyy/OpenQuota) | ~211 | MIT | Cross-platform limits/spend tracker. |
| [Dicklesworthstone/coding_agent_usage_tracker](https://github.com/Dicklesworthstone/coding_agent_usage_tracker) | ~85 | NOASSERTION | Single CLI for remaining quotas across Codex/Claude/Gemini/Cursor/Copilot. |
| [sculptdotfun/viberank](https://github.com/sculptdotfun/viberank) | ~117 | MIT | Public AI-coding usage leaderboard on ccusage data. |
| [cobra91/better-ccusage](https://github.com/cobra91/better-ccusage) | ~85 | MIT | Faster multi-provider JSONL analyzer. |
| [851-labs/tokenmaxxing](https://github.com/851-labs/tokenmaxxing) | ~72 | MIT | Local CLI on ccusage that syncs usage socially. |
| [Nihondo/AgentLimits](https://github.com/Nihondo/AgentLimits) | ~60 | MIT | macOS widgets for Codex/Claude limits + ccusage heatmap. |
| [majiayu000/quotabar](https://github.com/majiayu000/quotabar) | ~52 | MIT | **Pushed today.** Tauri v2 menubar — Codex capacity empty-state + cost tiles; hardened-runtime release chore. |
| [abhiunix/AgentHarbor](https://github.com/abhiunix/AgentHarbor) | ~12 | MIT | **Pushed today.** Native multi-agent rate-limit + session usage app. |
| [semantic-craft/iOS-vibebuddy](https://github.com/semantic-craft/iOS-vibebuddy) | ~89 | MIT | **Pushed today.** Multi-surface Claude/Codex/Grok/Cursor tracker — speech-pause resume fixes. |
| [zhnd/lumo](https://github.com/zhnd/lumo) | ~147 | MIT | Local-first Claude Code usage/cost/session dashboard. |
| [BerriAI/litellm](https://github.com/BerriAI/litellm) | ~59.0k | NOASSERTION | **Pushed today.** Gateway spend/pricing — adjacent. |
| [langfuse/langfuse](https://github.com/langfuse/langfuse) | ~34.7k | NOASSERTION | **Pushed today.** OSS LLM observability + OTel; transcript-from-observations. |
| [Helicone/helicone](https://github.com/Helicone/helicone) | ~6.2k | Apache-2.0 | OSS LLM observability proxy. Adjacent. |

**Build takeaways:** Stay ccusage-compatible; differentiate on live quotas + auto-detect + UX vs openusage.sh. Today's codeburn **Antigravity** tool/MCP/skills SQLite parse, CodeZeno **v2.11.29**, quotabar Codex capacity empty-state, and new **loongsuite-pilot** OTel collector are the sharpest signals. Treat CPA-Manager-Plus as gateway-ops adjacency, not a product twin. Keep Langfuse/LiteLLM/Helicone as complementary observability.

Sources: GitHub API 2026-09-17 (Europe/London).

### Open source — 18 Sep 2026

Fresh Friday GitHub API pass (Europe/London). Category still owned by **ccusage** + the two OpenUsage brands; **codeburn** remains the loud multi-tool local tracker. **Live today:** ccusage (**v20.0.22** overnight — pnpm/deps), codeburn (token-breakdown popover behind dollar amounts; never wedge on Full Disk Access–blocked Warp DB), otelite (**v0.1.149** — SQL trace-list summaries + instant all-time metrics via `metric_latest`), tokscale (**Copilot** `session-store.db` usage-event parse + webpki SSL roots), loongsuite-pilot (Codex user-isolated transcript discovery; Claude hooks into session config dirs; masking previews), quotabar (six-digit tray cost tiles), Clawdmeter (C6 AMOLED IMU auto-rotation), CPA-Manager-Plus, LiteLLM, Langfuse. **Also watch (new low-★):** `fschmutz/claude-usage-panel` — GNOME/macOS/statusLine + MCP `get_usage` from official Claude usage API.

| Repo | Stars | License | Why it matters |
|---|---:|---|---|
| [ccusage/ccusage](https://github.com/ccusage/ccusage) | ~18.6k | NOASSERTION | **v20.0.22 overnight.** Category CLI — stay format-compatible. |
| [getagentseal/codeburn](https://github.com/getagentseal/codeburn) | ~11.1k | MIT | **Pushed today.** Multi-tool local tracker — token-breakdown popover; Warp FDA wedge fix. |
| [robinebers/openusage](https://github.com/robinebers/openusage) | ~4.2k | MIT | openusage.ai menu-bar — Codex Swap still the freshest multi-account UX. |
| [janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage) | ~197 | MIT | openusage.sh local multi-tool TUI + SQLite — nearest product twin. |
| [ItsJazii/pane](https://github.com/ItsJazii/pane) | ~51 | NOASSERTION | Windows tray OpenUsage port across 15+ plans — platform-coverage peer. |
| [Maciek-roboblog/Claude-Code-Usage-Monitor](https://github.com/Maciek-roboblog/Claude-Code-Usage-Monitor) | ~8.7k | MIT | Real-time Claude Code usage monitor with predictions/warnings. |
| [HermannBjorgvin/Clawdmeter](https://github.com/HermannBjorgvin/Clawdmeter) | ~2.2k | (check) | **Pushed today.** ESP32 desk dashboard — C6 AMOLED IMU auto-rotation. |
| [Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail) | ~223 | MIT | Real-time cross-agent cost monitor. |
| [juliantanx/aiusage](https://github.com/juliantanx/aiusage) | ~129 | MIT | CodeBuddy CLI per-request usage. |
| [junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) | ~5.5k | MIT | **Pushed overnight.** **Copilot** session-store.db usage events + SSL root trust fix. |
| [Iamshankhadeep/ccseva](https://github.com/Iamshankhadeep/ccseva) | ~808 | MIT | macOS menu bar for live Claude Code usage. |
| [ColeMurray/claude-code-otel](https://github.com/ColeMurray/claude-code-otel) | ~502 | MIT | Observability stack for Claude Code usage/perf/cost (OTel path). |
| [CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) | ~495 | MIT | Multi Claude Code + Codex accounts; custom refresh (**v2.11.29** path). |
| [leeguooooo/claude-code-usage-bar](https://github.com/leeguooooo/claude-code-usage-bar) | ~375 | MIT | Predict profiles keyed by session config dir — statusLine quota UX. |
| [foyzulkarim/claude-lens](https://github.com/foyzulkarim/claude-lens) | ~247 | MIT | Local dashboard: sessions, token costs, cache, tool calls, daily breakdowns. |
| [frankchiu-dev/claude-codex-usage-dashboard](https://github.com/frankchiu-dev/claude-codex-usage-dashboard) | ~173 | MIT | Local Windows dashboard for Claude Code + Codex limits. |
| [planetf1/otelite](https://github.com/planetf1/otelite) | ~92 | NOASSERTION | **v0.1.149 today.** SQL trace-list summaries + instant all-time metrics — sharpest adjacent OTel dashboard move. |
| [alibaba/loongsuite-pilot](https://github.com/alibaba/loongsuite-pilot) | ~187 | Apache-2.0 | **Pushed today.** Local-first OTel for Claude/Codex/Cursor — Codex transcript discovery + Claude session hooks. |
| [seakee/CPA-Manager-Plus](https://github.com/seakee/CPA-Manager-Plus) | ~3.5k | MIT | Self-hosted CLIProxyAPI / gateway usage-cost-quota panel — adjacent ops, not desktop twin. |
| [Rodiun/frugon](https://github.com/Rodiun/frugon) | ~212 | MIT | Local LLM bill-leak analyzer. |
| [mag123c/toktrack](https://github.com/mag123c/toktrack) | ~189 | MIT | Ultra-fast Claude Code token/cost tracker. |
| [Halloweedev/usagepal](https://github.com/Halloweedev/usagepal) | ~82 | MIT | OpenUsage menu-bar fork — per-provider account switcher. |
| [Nanako0129/TokenBar](https://github.com/Nanako0129/TokenBar) | ~357 | MIT | macOS menu-bar quota peer across Claude/Codex/Cursor/OpenCode + 25 agents. |
| [tddworks/ClaudeBar](https://github.com/tddworks/ClaudeBar) | ~1.5k | (check) | macOS menu-bar Claude quota peer. |
| [deviffyy/OpenQuota](https://github.com/deviffyy/OpenQuota) | ~214 | MIT | Cross-platform limits/spend tracker. |
| [Dicklesworthstone/coding_agent_usage_tracker](https://github.com/Dicklesworthstone/coding_agent_usage_tracker) | ~85 | NOASSERTION | Single CLI for remaining quotas across Codex/Claude/Gemini/Cursor/Copilot. |
| [sculptdotfun/viberank](https://github.com/sculptdotfun/viberank) | ~117 | MIT | Public AI-coding usage leaderboard on ccusage data. |
| [cobra91/better-ccusage](https://github.com/cobra91/better-ccusage) | ~85 | MIT | Faster multi-provider JSONL analyzer. |
| [851-labs/tokenmaxxing](https://github.com/851-labs/tokenmaxxing) | ~72 | MIT | Local CLI on ccusage that syncs usage socially. |
| [Nihondo/AgentLimits](https://github.com/Nihondo/AgentLimits) | ~60 | MIT | macOS widgets for Codex/Claude limits + ccusage heatmap. |
| [majiayu000/quotabar](https://github.com/majiayu000/quotabar) | ~52 | MIT | **Pushed today.** Tauri v2 menubar — six-digit tray cost tiles readable. |
| [abhiunix/AgentHarbor](https://github.com/abhiunix/AgentHarbor) | ~12 | MIT | Native multi-agent rate-limit + session usage app. |
| [semantic-craft/iOS-vibebuddy](https://github.com/semantic-craft/iOS-vibebuddy) | ~89 | MIT | Multi-surface Claude/Codex/Grok/Cursor tracker. |
| [zhnd/lumo](https://github.com/zhnd/lumo) | ~147 | MIT | Local-first Claude Code usage/cost/session dashboard. |
| [fschmutz/claude-usage-panel](https://github.com/fschmutz/claude-usage-panel) | ~6 | MIT | **New watch.** Official usage API → GNOME/macOS/statusLine + MCP get_usage (Claude + Cursor). |
| [BerriAI/litellm](https://github.com/BerriAI/litellm) | ~59.1k | NOASSERTION | **Pushed today.** Gateway spend/pricing — adjacent. |
| [langfuse/langfuse](https://github.com/langfuse/langfuse) | ~34.8k | NOASSERTION | **Pushed today.** OSS LLM observability + OTel. |
| [Helicone/helicone](https://github.com/Helicone/helicone) | ~6.2k | Apache-2.0 | OSS LLM observability proxy. Adjacent. |

**Build takeaways:** Stay ccusage-compatible (**v20.0.22**); differentiate on live quotas + auto-detect + UX vs openusage.sh. Today's **otelite v0.1.149** storage perf, **tokscale Copilot session-store parse**, codeburn token-breakdown/Warp FDA fix, and loongsuite-pilot Codex/Claude hooks are the sharpest signals. Treat CPA-Manager-Plus as gateway-ops adjacency. Keep Langfuse/LiteLLM/Helicone as complementary observability.

Sources: GitHub API 2026-09-18 (Europe/London).

### Open source — 19 Sep 2026

Fresh Saturday GitHub API pass (Europe/London). Category still owned by **ccusage** + the two OpenUsage brands; **codeburn** remains the loud multi-tool local tracker. **Live today:** ccusage (**v20.0.23** overnight + continuous LiteLLM / models.dev pricing snapshots today), codeburn (**ZCode / z.ai coding-plan live quota** provider; yield rescue for sessions whose branch shipped past the window; name $0-priced usage instead of hiding it), TokenBar (**v1.19.1** — quota-lens restore/cache + failed-window retention), CodeZeno Usage-Monitor (**v2.12.37** overnight), otelite (**v0.1.153** — stats ANALYZE on always-active DBs + ingest drop reporting), CPA-Manager-Plus (**v1.13.1**), pane (star-prompt + reset toasts), LiteLLM (Gemini cache-control TTL clamp). openusage.ai quiet after **v0.7.13-beta.1**; tokscale Copilot `session-store.db` parse still the freshest provider win from Friday.

| Repo | Stars | License | Why it matters |
|---|---:|---|---|
| [ccusage/ccusage](https://github.com/ccusage/ccusage) | ~18.6k | NOASSERTION | **v20.0.23** + live pricing snapshots today — stay format-compatible. |
| [getagentseal/codeburn](https://github.com/getagentseal/codeburn) | ~11.1k | MIT | **Pushed today.** Multi-tool tracker — **ZCode live quota**; yield/session-window rescue; surface $0-priced usage. |
| [robinebers/openusage](https://github.com/robinebers/openusage) | ~4.2k | MIT | openusage.ai menu-bar — Codex Swap / **v0.7.13-beta.1** still freshest multi-account UX. |
| [janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage) | ~199 | MIT | openusage.sh local multi-tool TUI + SQLite — nearest product twin. |
| [ItsJazii/pane](https://github.com/ItsJazii/pane) | ~52 | NOASSERTION | Windows tray OpenUsage port — star-prompt + reset toasts overnight. |
| [Maciek-roboblog/Claude-Code-Usage-Monitor](https://github.com/Maciek-roboblog/Claude-Code-Usage-Monitor) | ~8.7k | MIT | Real-time Claude Code usage monitor with predictions/warnings. |
| [HermannBjorgvin/Clawdmeter](https://github.com/HermannBjorgvin/Clawdmeter) | ~2.2k | (check) | ESP32 desk dashboard — C6 AMOLED IMU auto-rotation still the recent hardware win. |
| [Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail) | ~223 | MIT | Real-time cross-agent cost monitor. |
| [juliantanx/aiusage](https://github.com/juliantanx/aiusage) | ~129 | MIT | CodeBuddy CLI per-request usage. |
| [junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) | ~5.5k | MIT | Copilot `session-store.db` usage-event parse + webpki SSL roots (Friday) still the sharpest provider parse. |
| [Iamshankhadeep/ccseva](https://github.com/Iamshankhadeep/ccseva) | ~808 | MIT | macOS menu bar for live Claude Code usage. |
| [ColeMurray/claude-code-otel](https://github.com/ColeMurray/claude-code-otel) | ~502 | MIT | Observability stack for Claude Code usage/perf/cost (OTel path). |
| [CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) | ~496 | MIT | **v2.12.37 overnight.** Multi Claude/Codex accounts + custom refresh. |
| [leeguooooo/claude-code-usage-bar](https://github.com/leeguooooo/claude-code-usage-bar) | ~377 | MIT | Predict profiles keyed by session config dir — statusLine quota UX (**3.43.2**). |
| [foyzulkarim/claude-lens](https://github.com/foyzulkarim/claude-lens) | ~247 | MIT | Local dashboard: sessions, token costs, cache, tool calls, daily breakdowns. |
| [frankchiu-dev/claude-codex-usage-dashboard](https://github.com/frankchiu-dev/claude-codex-usage-dashboard) | ~173 | MIT | Local Windows dashboard for Claude Code + Codex limits. |
| [planetf1/otelite](https://github.com/planetf1/otelite) | ~92 | NOASSERTION | **v0.1.153.** Storage ANALYZE + bounded ingest / drop reporting — sharpest adjacent OTel dashboard. |
| [alibaba/loongsuite-pilot](https://github.com/alibaba/loongsuite-pilot) | ~187 | Apache-2.0 | Local-first OTel for Claude/Codex/Cursor — Codex transcript discovery + Claude session hooks. |
| [seakee/CPA-Manager-Plus](https://github.com/seakee/CPA-Manager-Plus) | ~3.5k | MIT | **v1.13.1 today.** Self-hosted CLIProxyAPI / gateway usage-cost-quota panel — adjacent ops. |
| [Rodiun/frugon](https://github.com/Rodiun/frugon) | ~212 | MIT | Local LLM bill-leak analyzer. |
| [mag123c/toktrack](https://github.com/mag123c/toktrack) | ~189 | MIT | Ultra-fast Claude Code token/cost tracker. |
| [Halloweedev/usagepal](https://github.com/Halloweedev/usagepal) | ~82 | MIT | OpenUsage menu-bar fork — per-provider account switcher. |
| [Nanako0129/TokenBar](https://github.com/Nanako0129/TokenBar) | ~359 | MIT | **v1.19.1 today.** macOS menu-bar quota peer — quota-lens restore/cache across 25 agents. |
| [tddworks/ClaudeBar](https://github.com/tddworks/ClaudeBar) | ~1.5k | (check) | macOS menu-bar Claude quota peer. |
| [deviffyy/OpenQuota](https://github.com/deviffyy/OpenQuota) | ~214 | MIT | Cross-platform limits/spend tracker. |
| [Dicklesworthstone/coding_agent_usage_tracker](https://github.com/Dicklesworthstone/coding_agent_usage_tracker) | ~85 | NOASSERTION | Single CLI for remaining quotas across Codex/Claude/Gemini/Cursor/Copilot. |
| [sculptdotfun/viberank](https://github.com/sculptdotfun/viberank) | ~117 | MIT | Public AI-coding usage leaderboard on ccusage data. |
| [cobra91/better-ccusage](https://github.com/cobra91/better-ccusage) | ~85 | MIT | Faster multi-provider JSONL analyzer. |
| [851-labs/tokenmaxxing](https://github.com/851-labs/tokenmaxxing) | ~72 | MIT | Local CLI on ccusage that syncs usage socially. |
| [Nihondo/AgentLimits](https://github.com/Nihondo/AgentLimits) | ~60 | MIT | macOS widgets for Codex/Claude limits + ccusage heatmap. |
| [majiayu000/quotabar](https://github.com/majiayu000/quotabar) | ~52 | MIT | Tauri v2 menubar — six-digit tray cost tiles. |
| [abhiunix/AgentHarbor](https://github.com/abhiunix/AgentHarbor) | ~12 | MIT | Native multi-agent rate-limit + session usage app. |
| [semantic-craft/iOS-vibebuddy](https://github.com/semantic-craft/iOS-vibebuddy) | ~89 | MIT | Multi-surface Claude/Codex/Grok/Cursor tracker. |
| [zhnd/lumo](https://github.com/zhnd/lumo) | ~147 | MIT | Local-first Claude Code usage/cost/session dashboard. |
| [fschmutz/claude-usage-panel](https://github.com/fschmutz/claude-usage-panel) | ~6 | MIT | Official usage API → GNOME/macOS/statusLine + MCP get_usage (Claude + Cursor). |
| [sylearn/AIUsage](https://github.com/sylearn/AIUsage) | ~675 | Apache-2.0 | Desktop multi-tool usage peer (CodeBuddy function-call path). |
| [BerriAI/litellm](https://github.com/BerriAI/litellm) | ~59.1k | NOASSERTION | **Pushed today.** Gateway spend/pricing — Gemini cache TTL clamp. |
| [langfuse/langfuse](https://github.com/langfuse/langfuse) | ~34.8k | NOASSERTION | OSS LLM observability + OTel. |
| [Helicone/helicone](https://github.com/Helicone/helicone) | ~6.2k | Apache-2.0 | OSS LLM observability proxy. Adjacent. |

**Build takeaways:** Stay ccusage-compatible (**v20.0.23**); differentiate on live quotas + auto-detect + UX vs openusage.sh. Today's sharpest signals: **codeburn ZCode quota**, **TokenBar v1.19.1** restore/cache, **CodeZeno v2.12.37**, **otelite v0.1.153**. Treat CPA-Manager-Plus as gateway-ops adjacency. Keep Langfuse/LiteLLM/Helicone complementary.

Sources: GitHub API 2026-09-19 (Europe/London).

### Open source — 20 Sep 2026

Fresh Sunday GitHub API pass (Europe/London). Category still owned by **ccusage** + the two OpenUsage brands; **codeburn** remains the loud multi-tool local tracker. **Live today:** `CodeZeno/Claude-Code-Usage-Monitor` (**v2.12.39** overnight, **~500★**), `getagentseal/codeburn` (**Windows menu bar + Capacity Dock as Plugins card**; WSL provider-root discovery; CLI filter by billing route/mode; Claude Team seat label via `subscriptionType`; six-language localization), `ccusage/ccusage` (continuous **models.dev** pricing snapshots — still on **v20.0.23**, ~18.6k★), `mag123c/toktrack` (**archived Codex sessions** sync + data-dir path escape), `alibaba/loongsuite-pilot` (`PI_CODING_AGENT_DIR` / `GROK_HOME` / `DSH_HOME` honor + `--multimodal-mode` installer), `juliantanx/aiusage` (**1.5.18**), `851-labs/tokenmaxxing` (web header/weekday polish, **~75★**). TokenBar still **v1.19.1**; openusage.ai quiet after **v0.7.13-beta.1**; otelite still **v0.1.153**.

| Repo | Stars | License | Why it matters |
|---|---:|---|---|
| [ccusage/ccusage](https://github.com/ccusage/ccusage) | ~18.6k | NOASSERTION | **v20.0.23** + live pricing snapshots today — stay format-compatible. |
| [getagentseal/codeburn](https://github.com/getagentseal/codeburn) | ~11.1k | MIT | **Hottest push today** — Windows menu bar / Capacity Dock plugins; WSL roots; billing-route filter; Team seat label. |
| [robinebers/openusage](https://github.com/robinebers/openusage) | ~4.2k | MIT | openusage.ai menu-bar — Codex Swap / **v0.7.13-beta.1** still freshest multi-account UX. |
| [janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage) | ~200 | MIT | openusage.sh local multi-tool TUI + SQLite — nearest product twin. |
| [ItsJazii/pane](https://github.com/ItsJazii/pane) | ~52 | NOASSERTION | Windows tray OpenUsage port — star-prompt + reset toasts. |
| [Maciek-roboblog/Claude-Code-Usage-Monitor](https://github.com/Maciek-roboblog/Claude-Code-Usage-Monitor) | ~8.7k | MIT | Real-time Claude Code usage monitor with predictions/warnings. |
| [HermannBjorgvin/Clawdmeter](https://github.com/HermannBjorgvin/Clawdmeter) | ~2.2k | (check) | ESP32 desk dashboard — C6 AMOLED IMU auto-rotation still the recent hardware win. |
| [Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail) | ~223 | MIT | Real-time cross-agent cost monitor. |
| [juliantanx/aiusage](https://github.com/juliantanx/aiusage) | ~129 | MIT | **1.5.18 today** — CodeBuddy CLI per-request usage. |
| [junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) | ~5.5k | MIT | Copilot `session-store.db` usage-event parse still the sharpest provider parse (Friday). |
| [Iamshankhadeep/ccseva](https://github.com/Iamshankhadeep/ccseva) | ~808 | MIT | macOS menu bar for live Claude Code usage. |
| [ColeMurray/claude-code-otel](https://github.com/ColeMurray/claude-code-otel) | ~502 | MIT | Observability stack for Claude Code usage/perf/cost (OTel path). |
| [CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) | ~500 | MIT | **v2.12.39 overnight** — multi Claude/Codex accounts + custom refresh. |
| [leeguooooo/claude-code-usage-bar](https://github.com/leeguooooo/claude-code-usage-bar) | ~377 | MIT | Predict profiles keyed by session config dir — statusLine quota UX. |
| [foyzulkarim/claude-lens](https://github.com/foyzulkarim/claude-lens) | ~247 | MIT | Local dashboard: sessions, token costs, cache, tool calls, daily breakdowns. |
| [frankchiu-dev/claude-codex-usage-dashboard](https://github.com/frankchiu-dev/claude-codex-usage-dashboard) | ~173 | MIT | Local Windows dashboard for Claude Code + Codex limits. |
| [planetf1/otelite](https://github.com/planetf1/otelite) | ~92 | NOASSERTION | Still **v0.1.153** — storage ANALYZE + bounded ingest / drop reporting. |
| [alibaba/loongsuite-pilot](https://github.com/alibaba/loongsuite-pilot) | ~187 | Apache-2.0 | **Pushed today** — agent-home env honor + multimodal installer — local-first OTel for Claude/Codex/Cursor. |
| [seakee/CPA-Manager-Plus](https://github.com/seakee/CPA-Manager-Plus) | ~3.5k | MIT | Still **v1.13.1** — self-hosted CLIProxyAPI / gateway usage-cost-quota panel. |
| [Rodiun/frugon](https://github.com/Rodiun/frugon) | ~212 | MIT | Local LLM bill-leak analyzer. |
| [mag123c/toktrack](https://github.com/mag123c/toktrack) | ~190 | MIT | **Pushed today** — archived Codex session sync + path escape — Claude Code token/cost tracker. |
| [Halloweedev/usagepal](https://github.com/Halloweedev/usagepal) | ~82 | MIT | OpenUsage menu-bar fork — per-provider account switcher. |
| [Nanako0129/TokenBar](https://github.com/Nanako0129/TokenBar) | ~360 | MIT | Still **v1.19.1** — macOS menu-bar quota peer — quota-lens restore/cache across 25 agents. |
| [tddworks/ClaudeBar](https://github.com/tddworks/ClaudeBar) | ~1.5k | (check) | macOS menu-bar Claude quota peer. |
| [deviffyy/OpenQuota](https://github.com/deviffyy/OpenQuota) | ~216 | MIT | Cross-platform limits/spend tracker (**+2★**). |
| [Dicklesworthstone/coding_agent_usage_tracker](https://github.com/Dicklesworthstone/coding_agent_usage_tracker) | ~87 | NOASSERTION | Single CLI for remaining quotas across Codex/Claude/Gemini/Cursor/Copilot. |
| [sculptdotfun/viberank](https://github.com/sculptdotfun/viberank) | ~117 | MIT | Public AI-coding usage leaderboard on ccusage data. |
| [cobra91/better-ccusage](https://github.com/cobra91/better-ccusage) | ~88 | MIT | Faster multi-provider JSONL analyzer. |
| [851-labs/tokenmaxxing](https://github.com/851-labs/tokenmaxxing) | ~75 | MIT | **Pushed today** — local CLI on ccusage that syncs usage socially (**+3★**). |
| [Nihondo/AgentLimits](https://github.com/Nihondo/AgentLimits) | ~61 | MIT | macOS widgets for Codex/Claude limits + ccusage heatmap. |
| [majiayu000/quotabar](https://github.com/majiayu000/quotabar) | ~52 | MIT | Tauri v2 menubar — six-digit tray cost tiles. |
| [abhiunix/AgentHarbor](https://github.com/abhiunix/AgentHarbor) | ~12 | MIT | Native multi-agent rate-limit + session usage app. |
| [semantic-craft/iOS-vibebuddy](https://github.com/semantic-craft/iOS-vibebuddy) | ~89 | MIT | Multi-surface Claude/Codex/Grok/Cursor tracker. |
| [zhnd/lumo](https://github.com/zhnd/lumo) | ~147 | MIT | Local-first Claude Code usage/cost/session dashboard. |
| [fschmutz/claude-usage-panel](https://github.com/fschmutz/claude-usage-panel) | ~6 | MIT | Official usage API → GNOME/macOS/statusLine + MCP get_usage (Claude + Cursor); GNOME header icon buttons Saturday. |
| [sylearn/AIUsage](https://github.com/sylearn/AIUsage) | ~678 | Apache-2.0 | Desktop multi-tool usage peer (CodeBuddy function-call path). |
| [BerriAI/litellm](https://github.com/BerriAI/litellm) | ~59.2k | NOASSERTION | **Pushed today.** Gateway spend/pricing — adjacent. |
| [langfuse/langfuse](https://github.com/langfuse/langfuse) | ~34.8k | NOASSERTION | OSS LLM observability + OTel. |
| [Helicone/helicone](https://github.com/Helicone/helicone) | ~6.2k | Apache-2.0 | OSS LLM observability proxy. Adjacent. |

**Build takeaways:** Stay ccusage-compatible (**v20.0.23**); differentiate on live quotas + auto-detect + UX vs openusage.sh. Today's sharpest signals: **codeburn Windows menu bar / Capacity Dock + WSL roots**, **CodeZeno v2.12.39**, **toktrack archived Codex sessions**, **loongsuite-pilot multimodal/env homes**. Treat CPA-Manager-Plus as gateway-ops adjacency. Keep Langfuse/LiteLLM/Helicone complementary.

Sources: GitHub API 2026-09-20 (Europe/London).
