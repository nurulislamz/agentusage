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

### Update 2026-09-22

Fresh Tuesday scrape (Europe/London). Twin-risk with OpenUsage.sh unchanged; **34/35 homepage vs 36 README** count drift still live (8th day of product-surface quiet — deps/CI only). Material overnight moves: **ccusage v20.0.23 → v20.0.24** (models.dev pricing resilience); **CodeZeno v2.12.42 → v2.13.43** (Claude quota bindings + Theme Studio + usage-error diagnostics); **Splitrail** broke its multi-day quiet with peak/off-peak pricing (#258); **toktrack** OpenCode v2 parser fixes today; **tokscale** Muse Code CLI + MiMo split (unreleased ahead of v4.17.0). TokenBar / pane / codeburn hold yesterday’s tags with packaging/star creep only.

| Product | Delta | Vs agentUsage |
|---|---|---|
| [OpenUsage.sh](https://openusage.sh/) ([GitHub](https://github.com/janekbaraniewski/openusage), **~200★**, flat) | Homepage still **“Supported providers (35)”** + footer **“34 providers, and counting.”**; README still **36** (incl. Antigravity). Latest pushes still **deps/CI only** (2026-09-21 actions/website/docs bumps #375–#377). Last product tag still **v0.25.0** (31 Aug). | **Nearest twin** — category definition still ~1:1. Copy smell persists; product surface quiet now **8 days** (deps-only since mid-Sep fixes). |
| [ccusage](https://github.com/ccusage/ccusage) / [ccusage.com](https://ccusage.com/) (**~18.7k★** / 18,682; +~24) | **v20.0.24** published ~12:23 BST Mon (was **v20.0.23**) — bugfix: keep models.dev updates resilient to rate changes (#1767). Tuesday commits = **models.dev / LiteLLM pricing snapshots** only (latest ~09:26 BST). Still no Cursor in official source list. | Best **log-history / cost-report** OSS peer; thinner live quota/rate-limit TUI + key autodetection. |
| [Splitrail](https://github.com/Piebald-AI/splitrail) (**~223★**, flat) | Broke quiet: **feat(models): peak/off-peak pricing + cite an official source per model** (#258, ~18:50 BST Mon). Still tagged **v3.9.1** (6 Sep). | #2 local “live monitor” peer — pricing-model refresh after ~10 quiet days. |
| [AIUsage](https://github.com/juliantanx/aiusage) (**~129★**, flat) | Still **v1.5.18** (Sun multi-device sync). Quiet since 20 Sep. | Rising local peer + **name collision** (AIUsage vs agentUsage). |
| [openusage.ai](https://github.com/robinebers/openusage) (**~4.2k★** / 4,217; +~4) | Still on **v0.7.12** + **v0.7.13-beta.1** (Thu Codex Swap). No new tagged release overnight. | Naming collision only — do not conflate with openusage.sh. |
| Vendor UIs | **Cursor:** Spending tab still caps on-demand usage with team/member spend limits (Teams admin + Enterprise member/group overrides / Admin API) — Usage pane remains separate ([docs](https://cursor.com/help/account-and-billing/spend-limits)). **Claude Code:** `/usage` primary in current costs docs (session cost, plan bars, attribution, prompt-cache stats, usage-credits row); `/cost` historically aliased. **OpenRouter:** Activity / [usage-accounting](https://openrouter.ai/docs/cookbook/administration/usage-accounting) still API-platform credits & analytics. | Fragmentation OpenUsage/agentUsage claim to fix. |
| Helicone / Langfuse / LiteLLM | **Helicone** (~6.2k★ / 6,172; +~4): last push still **16 Sep** (platform-admin / HQL cross-tenant fix) — **6th quiet day**. **Langfuse** (~34.9k★ / 34,921, **pushed today**, +~41) — AI-gateway Anthropic connections in resolution contract (#17754). **LiteLLM** (~59.4k★ / 59,380, **pushed today**, +~88) — Presidio-masked output in spend logs (#42441); native compact-to-fit router (#42074). | Adjacent layers — prefer Langfuse/LiteLLM for new gateway/OTel work; Helicone still migrate-away adjacent. |
| Also watch (local peers) | [Nanako0129/TokenBar](https://github.com/Nanako0129/TokenBar) (~**362★**, +2; still **v1.20.0**) — quiet overnight (appcast-only ahead of tag). [CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) (~**509★**, +8; **v2.13.43**, was **v2.12.42**) — Claude quota bindings + Theme Studio values + improved usage-error diagnostics (single-commit minor bump, ~07:13 BST Mon). [getagentseal/codeburn](https://github.com/getagentseal/codeburn) (~11.2k★ / 11,156; +~22) — still **0.9.25**; packaging follow-up: plain DMG window + Flathub hash for desktop-v0.9.25 (#1519, ~22:44 BST Mon). [ItsJazii/pane](https://github.com/ItsJazii/pane) (~52★; still **v0.4.53**). [junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) (~5.5k★ / 5,506; +~9) — **Muse Code CLI** session tracking + Xiaomi **MiMo desktop vs MiMo Code** split + MiMo/MiniMax/Muse accounting fix (#1355, ~01:55 BST Tue); still tagged **v4.17.0**. [mag123c/toktrack](https://github.com/mag123c/toktrack) (~**190★**; **v2.17.3** Mon) — today’s OpenCode v2 parser work: track usage without migration duplicates (#259) + count compaction usage / stream SQLite rows (#261, ~10:01 BST). Also name-collision adjacent: [sylearn/AIUsage](https://github.com/sylearn/AIUsage) (~681★, +2) — macOS Sonoma icon fit (#75) today (different layer than juliantanx local web dash). | CodeZeno v2.13.43 + toktrack OpenCode v2 + tokscale Muse/MiMo = highest-velocity adjacent movers overnight. |

**Positioning refresh**
- **OpenUsage.sh** = primary collision (34/35/36 copy drift; deps-only product quiet now into a second week).
- **ccusage** = report/parser standard (**v20.0.24**; pricing resilience + snapshot churn; stars still climbing).
- **Splitrail** = #2 live local peer (peak/off-peak pricing refresh after ~10 quiet days).
- **juliantanx/aiusage** = local multi-tool dashboard peer + name collision (still v1.5.18).
- **openusage.ai** = menu-bar naming collision (Codex Swap still latest tagged).
- **CodeZeno / toktrack / tokscale** = highest-velocity adjacent local apps overnight (quota bindings; OpenCode v2; Muse/MiMo).
- Keep Langfuse / LiteLLM as complementary observability/gateway spend; treat **Helicone as maintenance-mode / migrate-away** adjacent (6 quiet days).

Sources: [openusage.sh](https://openusage.sh/); [github.com/janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage) (~200★); [github.com/ccusage/ccusage](https://github.com/ccusage/ccusage) v20.0.24 / #1767; [github.com/robinebers/openusage](https://github.com/robinebers/openusage); [github.com/juliantanx/aiusage](https://github.com/juliantanx/aiusage) v1.5.18; [github.com/Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail) #258; [github.com/Nanako0129/TokenBar](https://github.com/Nanako0129/TokenBar) v1.20.0; [github.com/CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) v2.13.43; [github.com/getagentseal/codeburn](https://github.com/getagentseal/codeburn) 0.9.25 / #1519; [github.com/ItsJazii/pane](https://github.com/ItsJazii/pane) v0.4.53; [github.com/junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) #1355; [github.com/mag123c/toktrack](https://github.com/mag123c/toktrack) v2.17.3 / #259 / #261; [cursor.com spend limits](https://cursor.com/help/account-and-billing/spend-limits); [code.claude.com costs](https://code.claude.com/docs/en/costs); [openrouter usage-accounting](https://openrouter.ai/docs/cookbook/administration/usage-accounting); [langfuse.com migrate-from-helicone](https://langfuse.com/resources/engineering/migrate-from-helicone); GitHub API star/push metadata 2026-09-22 (Europe/London).

### Update 2026-09-23

Fresh Wednesday scrape (Europe/London). Twin-risk with OpenUsage.sh unchanged; homepage now consistently **“34 providers and counting”** (the old “Supported providers (35)” hero wording is gone) while README still claims **36** (incl. Antigravity) — **9th day** of product-surface quiet (deps/CI only; last push still 21 Sep). Material overnight moves: **Splitrail v3.9.1 → v3.10.0** (ships Mon’s peak/off-peak + Grok 4.7 / Opus 5.5); **CodeZeno v2.13.43 → v2.14.55** (Grok Build monitoring + updater SemVer integrity); **TokenBar v1.20.0 → v1.20.1** (unpriced-model `$0.00` → `—`); **toktrack v2.17.3 → v2.17.4** (tags Tue’s OpenCode v2 parser fixes); **openusage.ai v0.7.13-beta.2** (Claude rate-limit resets + Cursor Grok 4.7 pricing). ccusage holds **v20.0.24** with pricing-snapshot churn only.

| Product | Delta | Vs agentUsage |
|---|---|---|
| [OpenUsage.sh](https://openusage.sh/) ([GitHub](https://github.com/janekbaraniewski/openusage), **~203★**, +3) | Homepage hero + section now both **“34 providers and counting”** (no more “35” claim); README still **36** (incl. Antigravity). Latest pushes still **deps/CI only** (2026-09-21 actions/website/docs bumps #375–#377). Last product tag still **v0.25.0** (31 Aug). | **Nearest twin** — category definition still ~1:1. Copy smell persists (34 vs 36); product surface quiet now **9 days** (deps-only since mid-Sep). |
| [ccusage](https://github.com/ccusage/ccusage) / [ccusage.com](https://ccusage.com/) (**~18.7k★** / 18,696; +~14) | Still **v20.0.24** (Mon). Wednesday commits = **models.dev / LiteLLM pricing snapshots** only (latest ~09:26 BST). Still no Cursor in official source list. | Best **log-history / cost-report** OSS peer; thinner live quota/rate-limit TUI + key autodetection. |
| [Splitrail](https://github.com/Piebald-AI/splitrail) (**~222★**, −1) | **v3.10.0** published ~18:33 BST Tue (was **v3.9.1**) — packages peak/off-peak pricing (#258), Claude Opus 5.5 + Opus fast-mode (#260), Grok 4.7 + 200K-token xAI long rates (#261), TUI model-share/wrap polish (#253/#254). | #2 local “live monitor” peer — first tagged release since 6 Sep. |
| [AIUsage](https://github.com/juliantanx/aiusage) (**~130★**, +1) | Still **v1.5.18** (Sun multi-device sync). Quiet since 20 Sep (**3rd quiet day**). | Rising local peer + **name collision** (AIUsage vs agentUsage). |
| [openusage.ai](https://github.com/robinebers/openusage) (**~4.2k★** / 4,230; +~13) | Stable tag still **v0.7.12**; **v0.7.13-beta.2** published ~06:28 BST Wed — Claude usage-limit reset grants (#1290), Cursor Grok 4.7 supplement pricing (#1286), OpenCode 2 ChatGPT OAuth attribution (#1284), Ollama monthly cloud limit (#1270). | Naming collision only — do not conflate with openusage.sh. |
| Vendor UIs | **Cursor:** Spending tab still caps on-demand usage with team/member spend limits (Teams admin + Enterprise member/group overrides / Admin API) — Usage pane remains separate ([docs](https://cursor.com/help/account-and-billing/spend-limits)). **Claude Code:** `/usage` primary in current costs docs (session cost, plan bars, attribution, prompt-cache stats, usage-credits row); `/cost` historically aliased. **OpenRouter:** Activity / [usage-accounting](https://openrouter.ai/docs/cookbook/administration/usage-accounting) still API-platform credits & analytics (usage always included; deprecated `usage.include` params). | Fragmentation OpenUsage/agentUsage claim to fix. |
| Helicone / Langfuse / LiteLLM | **Helicone** (~6.2k★ / 6,176; +~4): last push still **16 Sep** (platform-admin / HQL cross-tenant fix) — **7th quiet day**. **Langfuse** (~35.0k★ / 34,956, **pushed today**, +~35) — **v4.42.0** (~08:19 BST); member role filters (#17810). **LiteLLM** (~59.5k★ / 59,455, **pushed today**, +~75) — **v1.102.1** (~07:12 BST); OTel `gen_ai.conversation.id` from session (#42486); UI session-token revoke on logout (#42463). | Adjacent layers — prefer Langfuse/LiteLLM for new gateway/OTel work; Helicone still migrate-away adjacent. |
| Also watch (local peers) | [Nanako0129/TokenBar](https://github.com/Nanako0129/TokenBar) (~**364★**, +2; **v1.20.1**, was **v1.20.0**) — unpriced models show `—` not `$0.00`; positive sub-cent amounts read `<$0.01` (#375, ~21:17 BST Tue). [CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) (~**517★**, +8; **v2.14.55**, was **v2.13.43**) — Grok Build usage monitoring (#109); updater integrity + SemVer enforce (#113); Haiku alias for usage probes (#115); ~20 commits overnight (v2.14.47→55). [getagentseal/codeburn](https://github.com/getagentseal/codeburn) (~11.2k★ / 11,179; +~23) — still **0.9.25**; README story rewrite (#1527) + Bronze/Gold company sponsor tiers (#1533, today). [ItsJazii/pane](https://github.com/ItsJazii/pane) (~53★, +1; still **v0.4.53**). [junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) (~5.5k★ / 5,514; +~8) — still tagged **v4.17.0**; last product commit still Tue Muse/MiMo #1355. [mag123c/toktrack](https://github.com/mag123c/toktrack) (~**191★**, +1; **v2.17.4**, was **v2.17.3**) — tags Tue’s OpenCode v2 parser fixes (#259/#261). Also name-collision adjacent: [sylearn/AIUsage](https://github.com/sylearn/AIUsage) (~682★, +1; **v0.15.21** today) — OpenCode v2 session/usage/credential adapt (#77) + macOS icon fit (#75) (different layer than juliantanx local web dash). | CodeZeno v2.14.55 + Splitrail v3.10.0 + TokenBar v1.20.1 = highest-velocity adjacent movers overnight. |

**Positioning refresh**
- **OpenUsage.sh** = primary collision (homepage **34** vs README **36**; deps-only product quiet into **day 9**).
- **ccusage** = report/parser standard (**v20.0.24**; snapshot churn only; stars still climbing).
- **Splitrail** = #2 live local peer (**v3.10.0** — peak/off-peak + Grok 4.7 / Opus 5.5 now tagged).
- **juliantanx/aiusage** = local multi-tool dashboard peer + name collision (still v1.5.18; 3 quiet days).
- **openusage.ai** = menu-bar naming collision (**v0.7.13-beta.2** ahead of stable).
- **CodeZeno / TokenBar / toktrack** = highest-velocity adjacent local apps overnight (Grok Build; unpriced `$0.00` fix; OpenCode v2 tag).
- Keep Langfuse / LiteLLM as complementary observability/gateway spend; treat **Helicone as maintenance-mode / migrate-away** adjacent (7 quiet days).

Sources: [openusage.sh](https://openusage.sh/); [github.com/janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage) (~203★); [github.com/ccusage/ccusage](https://github.com/ccusage/ccusage) v20.0.24; [github.com/robinebers/openusage](https://github.com/robinebers/openusage) v0.7.13-beta.2 / #1290 / #1286 / #1284; [github.com/juliantanx/aiusage](https://github.com/juliantanx/aiusage) v1.5.18; [github.com/Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail) v3.10.0 / #258 / #260 / #261; [github.com/Nanako0129/TokenBar](https://github.com/Nanako0129/TokenBar) v1.20.1 / #375; [github.com/CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) v2.14.55 / #109 / #113; [github.com/getagentseal/codeburn](https://github.com/getagentseal/codeburn) 0.9.25 / #1527 / #1533; [github.com/ItsJazii/pane](https://github.com/ItsJazii/pane) v0.4.53; [github.com/junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) #1355; [github.com/mag123c/toktrack](https://github.com/mag123c/toktrack) v2.17.4 / #259 / #261; [github.com/sylearn/AIUsage](https://github.com/sylearn/AIUsage) v0.15.21 / #77; [cursor.com spend limits](https://cursor.com/help/account-and-billing/spend-limits); [code.claude.com costs](https://code.claude.com/docs/en/costs); [openrouter usage-accounting](https://openrouter.ai/docs/cookbook/administration/usage-accounting); [langfuse.com migrate-from-helicone](https://langfuse.com/resources/engineering/migrate-from-helicone); GitHub API star/push metadata 2026-09-23 (Europe/London).

### Update 2026-09-24

Fresh Thursday scrape (Europe/London). Twin-risk with OpenUsage.sh unchanged; homepage copy drift is **three-way again** — meta + “Supported providers (**35**)” returned, while hero/footer still say “**34** providers and counting” and README still claims **36** (incl. Antigravity) — **10th day** of product-surface quiet (no new push; last still deps/CI on 21 Sep; tag still **v0.25.0** 31 Aug). Material overnight moves: **CodeZeno v2.14.55 → v2.15.14** (Builder unification + Antigravity OAuth refresh); **Splitrail v3.10.0 → v3.10.1** (GPT-6 Sol/Luna API pricing); **juliantanx/aiusage v1.5.18 → v1.5.19** (broke 3 quiet days — Antigravity readable-model attribution + Opus 5.5 / GPT-6 prices); **TokenBar v1.20.1 → v1.20.2** (signed-out `agy` no longer opens Google login every refresh); **pane v0.4.53 → v0.4.54** (Claude banked resets + weekly capacity); **Langfuse v4.42.0 → v4.44.0**. ccusage holds published **v20.0.24** (git tag **v20.0.25** since Mon, no GitHub Release) with pricing-snapshot churn only.

| Product | Delta | Vs agentUsage |
|---|---|---|
| [OpenUsage.sh](https://openusage.sh/) ([GitHub](https://github.com/janekbaraniewski/openusage), **~206★**, +3) | Homepage again **34 / 35 / 36** drift: hero + footer “**34** providers and counting”; meta + “Supported providers (**35**)”; README still **36** (incl. Antigravity). **No new commits** since 21 Sep deps/CI (#375–#377). Last product tag still **v0.25.0** (31 Aug). | **Nearest twin** — category definition still ~1:1. Copy smell back to three-way; product surface quiet now **10 days**. |
| [ccusage](https://github.com/ccusage/ccusage) / [ccusage.com](https://ccusage.com/) (**~18.7k★** / 18,726; +~30) | Published release still **v20.0.24**; git tag **v20.0.25** (21 Sep, Antigravity zero-byte conversation DB skip #1774) has **no** GitHub Release. Thursday commits = **models.dev / LiteLLM pricing snapshots** only (latest ~09:28 BST). Still no Cursor in official source list. | Best **log-history / cost-report** OSS peer; thinner live quota/rate-limit TUI + key autodetection. |
| [Splitrail](https://github.com/Piebald-AI/splitrail) (**~222★**, flat) | **v3.10.1** published ~04:11 BST Thu (was **v3.10.0**) — GPT-6 Sol and Luna API pricing (#264 / #265). | #2 local “live monitor” peer — second tagged release this week. |
| [AIUsage](https://github.com/juliantanx/aiusage) (**~132★**, +2) | **v1.5.19** published ~08:39 BST Thu (was **v1.5.18**) — Antigravity readable-model attribution + Gemini 3.x Pro pricing (#70); curated Opus 5.5 / GPT-6 Sol / Luna prices ahead of LiteLLM sync (#71); cloud-push insert/update/no-op counts (#67). Broke the quiet streak. | Rising local peer + **name collision** (AIUsage vs agentUsage). |
| [openusage.ai](https://github.com/robinebers/openusage) (**~4.3k★** / 4,250; +~20) | Stable tag still **v0.7.12**; latest pre still **v0.7.13-beta.2**. Post-tag Wed fix: Claude terminal sessions counted under the signed-in account when multiple accounts are known (#1299, ~11:55 BST Wed). **No Thu commits.** | Naming collision only — do not conflate with openusage.sh. |
| Vendor UIs | **Cursor:** Spending tab still caps on-demand usage with team/member spend limits (Teams admin + Enterprise member/group overrides / Admin API; dynamic seat-proportional limits) — Usage pane remains separate ([docs](https://cursor.com/help/account-and-billing/spend-limits)). **Claude Code:** `/usage` primary in current costs docs (session cost, plan bars, attribution, prompt-cache stats, usage-credits row); `/cost` historically aliased. **OpenRouter:** Activity / [usage-accounting](https://openrouter.ai/docs/cookbook/administration/usage-accounting) still API-platform credits & analytics (usage always included; deprecated `usage.include` params). | Fragmentation OpenUsage/agentUsage claim to fix. |
| Helicone / Langfuse / LiteLLM | **Helicone** (~6.2k★ / 6,176; flat): last push still **16 Sep** — **8th quiet day**. **Langfuse** (~35.0k★ / 34,994, **pushed today**, +~38) — **v4.44.0** (~10:07 BST; was **v4.42.0** Wed AM via **v4.43.0** Wed PM); Vertex Gemini extra headers, evals gallery, trace-batching (#17831/#17845/#17839). **LiteLLM** (~59.5k★ / 59,529, **pushed today**, +~74) — latest non-prerelease still **v1.102.1**; Thu = cost-map churn (wandb DeepSeek-V4.1-Flash, Bedrock mantle GPT-6 Luna/Sol/Terra + grok-4.6, Vertex gemini-3.8-live, OpenRouter/Fireworks sync). | Adjacent layers — prefer Langfuse/LiteLLM for new gateway/OTel work; Helicone still migrate-away adjacent. |
| Also watch (local peers) | [Nanako0129/TokenBar](https://github.com/Nanako0129/TokenBar) (~**367★**, +3; **v1.20.2**, was **v1.20.1**) — skip `agy` when Antigravity signed out so refresh no longer opens Google login (#378, ~03:27 BST Thu). [CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) (~**517★**, flat; **v2.15.14**, was **v2.14.55**) — **v2.15.0** (~00:12 BST Thu): Builder unification (#122), Retry-After cooldowns (#117), Windows Hyper-V harness (#124); then rapid **v2.15.5→14** overnight incl. Antigravity OAuth refresh (#139), Claude probe freshness during cooldowns (#133). [getagentseal/codeburn](https://github.com/getagentseal/codeburn) (~11.2k★ / 11,208; +~29) — still **0.9.25**; Wed/Thu product: Opus 5.5 pricing snapshot (#1537), low prompt-cache-hit session flag (#1536), Crush projects registry (#1510), Claude credential recover (#1516). [ItsJazii/pane](https://github.com/ItsJazii/pane) (~**55★**, +2; **v0.4.54**, was **v0.4.53**) — Claude banked limit resets (#241) + Codex/Claude weekly capacity (#242). [junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) (~5.5k★ / 5,525; +~11) — still tagged **v4.17.0**; Wed night: Antigravity IDE extension sessions + `import --submit` labelled backfill. [mag123c/toktrack](https://github.com/mag123c/toktrack) (~**192★**, +1; still **v2.17.4**; quiet since Tue tag). Also name-collision adjacent: [sylearn/AIUsage](https://github.com/sylearn/AIUsage) (~682★, flat; still **v0.15.21**) — no Thu code (OpenCode v2 adapt still Wed). | CodeZeno v2.15.x + Splitrail v3.10.1 + AIUsage v1.5.19 + TokenBar v1.20.2 = highest-velocity adjacent movers overnight. |

**Positioning refresh**
- **OpenUsage.sh** = primary collision (homepage **34/35** vs README **36**; deps-only product quiet into **day 10**).
- **ccusage** = report/parser standard (published **v20.0.24** / tagged **v20.0.25**; snapshot churn only; stars still climbing).
- **Splitrail** = #2 live local peer (**v3.10.1** — GPT-6 Sol/Luna on top of Tue’s peak/off-peak + Grok 4.7 / Opus 5.5).
- **juliantanx/aiusage** = local multi-tool dashboard peer + name collision (**v1.5.19** — Antigravity attribution + new-model prices).
- **openusage.ai** = menu-bar naming collision (**v0.7.13-beta.2** ahead of stable; multi-account Claude terminal fix Wed).
- **CodeZeno / TokenBar / pane** = highest-velocity adjacent local apps overnight (v2.15.x Builder+OAuth; signed-out `agy` guard; banked resets + weekly capacity).
- Keep Langfuse / LiteLLM as complementary observability/gateway spend; treat **Helicone as maintenance-mode / migrate-away** adjacent (8 quiet days).

Sources: [openusage.sh](https://openusage.sh/); [github.com/janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage) (~206★); [github.com/ccusage/ccusage](https://github.com/ccusage/ccusage) v20.0.24 / tag v20.0.25 / #1774; [github.com/robinebers/openusage](https://github.com/robinebers/openusage) v0.7.13-beta.2 / #1299; [github.com/juliantanx/aiusage](https://github.com/juliantanx/aiusage) v1.5.19 / #70 / #71 / #67; [github.com/Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail) v3.10.1 / #264; [github.com/Nanako0129/TokenBar](https://github.com/Nanako0129/TokenBar) v1.20.2 / #378; [github.com/CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) v2.15.14 / #122 / #139 / #133; [github.com/getagentseal/codeburn](https://github.com/getagentseal/codeburn) 0.9.25 / #1537 / #1536; [github.com/ItsJazii/pane](https://github.com/ItsJazii/pane) v0.4.54 / #241 / #242; [github.com/junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) Antigravity IDE + import --submit; [github.com/mag123c/toktrack](https://github.com/mag123c/toktrack) v2.17.4; [github.com/sylearn/AIUsage](https://github.com/sylearn/AIUsage) v0.15.21; [cursor.com spend limits](https://cursor.com/help/account-and-billing/spend-limits); [code.claude.com costs](https://code.claude.com/docs/en/costs); [openrouter usage-accounting](https://openrouter.ai/docs/cookbook/administration/usage-accounting); [langfuse.com migrate-from-helicone](https://langfuse.com/resources/engineering/migrate-from-helicone); GitHub API star/push metadata 2026-09-24 (Europe/London).

### Update 2026-09-25

Fresh Friday scrape (Europe/London). Twin-risk with OpenUsage.sh unchanged; homepage copy drift still **three-way** — meta + “Supported providers (**35**)” vs hero/footer “**34** providers and counting” vs README **36** (incl. Antigravity) — **11th day** of product-surface quiet (overnight was deps-only #380 on docs/site; tag still **v0.25.0** 31 Aug). Material overnight/Fri moves: **Langfuse v4.44.0 → v4.45.2** (via **v4.45.0** / **v4.45.1** Thu PM); **CodeZeno** git tags **v2.15.15→v2.15.17** ahead of GitHub Release (latest Release still **v2.15.14**) — Direct3D dashboard fallback + fractional-scale segments; **codeburn** Cursor usage-CSV import (#1558) to replace local estimates; **pane** unreleased main: Claude Cloud session credits (#246) + StepFun provider (#245); **tokscale** TUI i18n (en/ko/ja/zh-CN/fr) + OpenClaw compressed SQLite + Codex tier / Antigravity ledger fixes (#1365). ccusage still published **v20.0.24** (git tag **v20.0.25** no GitHub Release) with Fri pricing-snapshot churn only. Splitrail / juliantanx/aiusage / TokenBar tagged releases hold at Thu levels.

| Product | Delta | Vs agentUsage |
|---|---|---|
| [OpenUsage.sh](https://openusage.sh/) ([GitHub](https://github.com/janekbaraniewski/openusage), **~207★**, +1) | Homepage still **34 / 35 / 36** drift: hero + footer “**34** providers and counting”; meta + “Supported providers (**35**)”; README still **36** (incl. Antigravity). One overnight commit: deps(docs) bump image-size in `/docs/site` (#380, ~22:23 BST Thu). Last product tag still **v0.25.0** (31 Aug). | **Nearest twin** — category definition still ~1:1. Copy smell persists; product surface quiet now **11 days** (deps-only pulses). |
| [ccusage](https://github.com/ccusage/ccusage) / [ccusage.com](https://ccusage.com/) (**~18.7k★** / 18,734; +~8) | Published release still **v20.0.24**; git tag **v20.0.25** (21 Sep) still has **no** GitHub Release. Fri commits = **models.dev / LiteLLM pricing snapshots** only (through ~09:26 BST). Still no Cursor in official source list. | Best **log-history / cost-report** OSS peer; thinner live quota/rate-limit TUI + key autodetection. |
| [Splitrail](https://github.com/Piebald-AI/splitrail) (**~222★**, flat) | Still **v3.10.1** (Thu GPT-6 Sol/Luna). **No Fri commits.** | #2 local “live monitor” peer — quiet after Thu tag. |
| [AIUsage](https://github.com/juliantanx/aiusage) (**~132★**, flat) | Still **v1.5.19** (Thu Antigravity attribution + Opus 5.5 / GPT-6 prices). **No Fri commits.** | Rising local peer + **name collision** (AIUsage vs agentUsage); quiet after Thu release. |
| [openusage.ai](https://github.com/robinebers/openusage) (**~4.3k★** / 4,257; +~7) | Stable tag still **v0.7.12**; latest pre still **v0.7.13-beta.2**. Last code still Wed (#1299 multi-account Claude terminal sessions). **No Thu/Fri commits.** | Naming collision only — do not conflate with openusage.sh. |
| Vendor UIs | **Cursor:** Spending tab still caps on-demand usage with team/member spend limits (Teams admin + Enterprise member/group overrides / Admin API; dynamic seat-proportional limits) — Usage pane remains separate ([docs](https://cursor.com/help/account-and-billing/spend-limits)). **Claude Code:** `/usage` primary in current costs docs (session cost, plan bars, attribution, prompt-cache stats, usage-credits row); `/cost` historically aliased. **OpenRouter:** Activity / [usage-accounting](https://openrouter.ai/docs/cookbook/administration/usage-accounting) still API-platform credits & analytics (usage always included; deprecated `usage.include` params). | Fragmentation OpenUsage/agentUsage claim to fix. |
| Helicone / Langfuse / LiteLLM | **Helicone** (~6.2k★ / 6,175; −1): last push still **16 Sep** — **9th quiet day**. **Langfuse** (~35.0k★ / 35,034, **pushed today**, +~40) — **v4.45.2** (~20:16 BST Thu; was **v4.44.0** Thu AM via **v4.45.0** ~13:15 / **v4.45.1**); AI-gateway batch inference telemetry (#17855), design-system charts, billing checkout org name. **LiteLLM** (~59.6k★ / 59,592, **pushed today**, +~63) — latest non-prerelease still **v1.102.1**; Fri = Langfuse SDK callback → v4 (#36741), cost-map OpenAI cached image input (#43143), Presidio stream mask / key MCP grant fixes. | Adjacent layers — prefer Langfuse/LiteLLM for new gateway/OTel work; Helicone still migrate-away adjacent. |
| Also watch (local peers) | [Nanako0129/TokenBar](https://github.com/Nanako0129/TokenBar) (~**368★**, +1; still **v1.20.2**) — Fri merges ahead of tag: Codex OAuth refresh before expiry (#344), Local Network usage description (#374), cold-load timeline / concurrent graph compute (#384). [CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) (~**524★**, +7; GitHub Release still **v2.15.14**; git tags **v2.15.15→v2.15.17** ~03:46 BST Fri) — Direct3D dashboard startup fallback (#141); uniform segments at fractional display scales (#142). [getagentseal/codeburn](https://github.com/getagentseal/codeburn) (~11.2k★ / 11,228; +~20) — still **0.9.25**; Fri product: `codeburn import cursor <csv>` replaces local Cursor estimates with dashboard CSV (#1558 — cache-read gap); Cursor Agent `~/.cursor/chats` store.db sessions (#1556); Copilot OTel workspace attribution (#1555); WSL Claude quota credential (#1554). [ItsJazii/pane](https://github.com/ItsJazii/pane) (~**57★**, +2; Release still **v0.4.54**; unreleased main ~21:24 BST Thu) — Claude Cloud session credits via `iguana_necktie` (#246); StepFun balance / Step Plan credits + spend from Claude Code/Codex/OpenCode (#245). [junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) (~5.5k★ / 5,537; +~12) — still tagged **v4.17.0**; Fri: TUI language selection (en/ko/ja/zh-CN/fr); OpenClaw compressed SQLite transcripts; Codex tier source + Antigravity ledger reconciliation (#1365). [mag123c/toktrack](https://github.com/mag123c/toktrack) (~**192★**, flat; still **v2.17.4**; quiet since Tue). Also name-collision adjacent: [sylearn/AIUsage](https://github.com/sylearn/AIUsage) (~683★, +1; still **v0.15.21**) — no Fri code. | Highest-velocity Fri: Langfuse v4.45.x + CodeZeno tag climb + codeburn Cursor CSV + pane StepFun/Cloud credits + tokscale i18n. |

**Positioning refresh**
- **OpenUsage.sh** = primary collision (homepage **34/35** vs README **36**; deps-only into **day 11**).
- **ccusage** = report/parser standard (published **v20.0.24** / tagged **v20.0.25**; snapshot churn only; stars still climbing).
- **Splitrail** = #2 live local peer (**v3.10.1** hold — GPT-6 Sol/Luna still latest tag).
- **juliantanx/aiusage** = local multi-tool dashboard peer + name collision (**v1.5.19** hold).
- **openusage.ai** = menu-bar naming collision (**v0.7.13-beta.2** ahead of stable; quiet since Wed).
- **CodeZeno / codeburn / pane / tokscale** = highest-velocity adjacent local apps overnight (tags→v2.15.17 no Release; Cursor CSV truth; Cloud credits + StepFun; TUI i18n).
- Keep Langfuse / LiteLLM as complementary observability/gateway spend (**Langfuse v4.45.2** notable); treat **Helicone as maintenance-mode / migrate-away** adjacent (9 quiet days).

Sources: [openusage.sh](https://openusage.sh/); [github.com/janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage) (~207★) / #380; [github.com/ccusage/ccusage](https://github.com/ccusage/ccusage) v20.0.24 / tag v20.0.25; [github.com/robinebers/openusage](https://github.com/robinebers/openusage) v0.7.13-beta.2; [github.com/juliantanx/aiusage](https://github.com/juliantanx/aiusage) v1.5.19; [github.com/Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail) v3.10.1; [github.com/Nanako0129/TokenBar](https://github.com/Nanako0129/TokenBar) v1.20.2 / #344 / #374 / #384; [github.com/CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) Release v2.15.14 / tags v2.15.17 / #141 / #142; [github.com/getagentseal/codeburn](https://github.com/getagentseal/codeburn) 0.9.25 / #1558 / #1556; [github.com/ItsJazii/pane](https://github.com/ItsJazii/pane) v0.4.54 / #246 / #245; [github.com/junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) TUI i18n / #1365; [github.com/mag123c/toktrack](https://github.com/mag123c/toktrack) v2.17.4; [github.com/sylearn/AIUsage](https://github.com/sylearn/AIUsage) v0.15.21; [cursor.com spend limits](https://cursor.com/help/account-and-billing/spend-limits); [code.claude.com costs](https://code.claude.com/docs/en/costs); [openrouter usage-accounting](https://openrouter.ai/docs/cookbook/administration/usage-accounting); [langfuse/langfuse](https://github.com/langfuse/langfuse) v4.45.2 / #17855; [BerriAI/litellm](https://github.com/BerriAI/litellm) v1.102.1 / #36741; [langfuse.com migrate-from-helicone](https://langfuse.com/resources/engineering/migrate-from-helicone); GitHub API star/push metadata 2026-09-25 (Europe/London).


### Update 2026-09-26

Fresh Saturday scrape (Europe/London). Twin-risk with OpenUsage.sh unchanged; homepage copy drift still **three-way** — meta + “Supported providers (**35**)” vs hero/footer “**34** providers and counting” vs README **36** (incl. Antigravity) — **12th day** of product-surface quiet (last code still deps-only #380 Thu; tag still **v0.25.0** 31 Aug; **~211★**). Biggest Sat move: **TokenBar → Syrtis v2.0.0** (repo rename [Nanako0129/syrtis](https://github.com/Nanako0129/syrtis); app renames on disk; new brew cask; Liquid Glass panel on macOS 27). Also: **CodeZeno** GitHub Release caught up **v2.15.14 → v2.15.19** (Fri Direct3D + fractional-scale tags now shipped); **pane v0.4.55** ships Claude Cloud session credits (#246) + StepFun (#245); **Langfuse v4.45.2 → v4.46.0** (experiments side-by-side, skills management). ccusage still published **v20.0.24** (git tag **v20.0.25** no Release) with Sat pricing-snapshot churn only. Splitrail / juliantanx/aiusage / toktrack / sylearn hold.

| Product | Delta | Vs agentUsage |
|---|---|---|
| [OpenUsage.sh](https://openusage.sh/) ([GitHub](https://github.com/janekbaraniewski/openusage), **~211★**, +4) | Homepage still **34 / 35 / 36** drift: hero + “34 providers, and counting”; meta + “Supported providers (**35**)”; README still **36** (incl. Antigravity). **No Fri/Sat commits** after Thu deps(docs) #380. Last product tag still **v0.25.0** (31 Aug); last product-surface code still **13 Sep** (Codex/Crush fixes). | **Nearest twin** — category definition still ~1:1. Copy smell persists; product surface quiet now **12 days** (deps-only pulses). |
| [ccusage](https://github.com/ccusage/ccusage) / [ccusage.com](https://ccusage.com/) (**~18.7k★** / 18,744; +~10) | Published release still **v20.0.24**; git tag **v20.0.25** (21 Sep) still has **no** GitHub Release. Sat commits = **models.dev / LiteLLM pricing snapshots** only (through ~09:23 BST). Still no Cursor in official source list. | Best **log-history / cost-report** OSS peer; thinner live quota/rate-limit TUI + key autodetection. |
| [Splitrail](https://github.com/Piebald-AI/splitrail) (**~222★**, flat) | Still **v3.10.1** (Thu GPT-6 Sol/Luna). **No Fri/Sat commits.** | #2 local “live monitor” peer — quiet after Thu tag. |
| [AIUsage](https://github.com/juliantanx/aiusage) (**~133★**, +1) | Still **v1.5.19** (Thu Antigravity attribution + Opus 5.5 / GPT-6 prices). **No Fri/Sat commits.** | Rising local peer + **name collision** (AIUsage vs agentUsage); quiet after Thu release. |
| [openusage.ai](https://github.com/robinebers/openusage) (**~4.3k★** / 4,267; +~10) | Stable tag still **v0.7.12**; latest pre still **v0.7.13-beta.2**. Last code still Wed (#1299 multi-account Claude terminal sessions). Repo `pushed_at` bumped Sat (events/PRs); **no new main commits** since Wed. | Naming collision only — do not conflate with openusage.sh. |
| Vendor UIs | **Cursor:** Spending tab still caps on-demand usage with team/member spend limits (Teams admin + Enterprise member/group overrides / Admin API; dynamic seat-proportional limits) — Usage pane remains separate ([docs](https://cursor.com/help/account-and-billing/spend-limits)). **Claude Code:** `/usage` primary in current costs docs (session cost, plan bars, attribution, prompt-cache stats, usage-credits row); `/cost` historically aliased ([docs](https://code.claude.com/docs/en/costs)). **OpenRouter:** Activity / [usage-accounting](https://openrouter.ai/docs/cookbook/administration/usage-accounting) still API-platform credits & analytics (usage always included; deprecated `usage.include` params). | Fragmentation OpenUsage/agentUsage claim to fix. |
| Helicone / Langfuse / LiteLLM | **Helicone** (~6.2k★ / 6,177; +2): last push still **16 Sep** — **10th quiet day**. **Langfuse** (~35.1k★ / 35,059, **pushed today**, +~25) — **v4.46.0** (~15:20 BST Fri; was **v4.45.2**) via **v4.45.3** / **v4.45.4**; experiments side-by-side comparison (#17929), basic skill management (#17803), formatted JSON on dataset items, 100-row tracing page size. **LiteLLM** (~59.6k★ / 59,647, **pushed today**, +~55) — latest non-prerelease still **v1.102.1** (Latest); Sat = cost-map duplicate OpenRouter Perceptron cleanup (#43273), Langfuse DB-callback test isolation (#43288), rust gateway/config crates (#43289). | Adjacent layers — prefer Langfuse/LiteLLM for new gateway/OTel work; Helicone still migrate-away adjacent. |
| Also watch (local peers) | [Nanako0129/TokenBar](https://github.com/Nanako0129/TokenBar) → **[Syrtis](https://github.com/Nanako0129/syrtis)** (~**369★**, +1; **v2.0.0** “Syrtis” ~04:39 BST Sat) — full rebrand: app renames `TokenBar.app`→`Syrtis.app` on first launch; brew `nanako0129/tap/syrtis`; Liquid Glass dashboard panel on macOS 27 (#380/#391); client cache re-scan (Pi fork dedupe, OpenClaw `.jsonl.zst`, Cursor cache-write); site [syrtis.nyanako.com](https://syrtis.nyanako.com). [CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) (~**538★**, +14; GitHub Release now **v2.15.19**, was stuck at **v2.15.14**) — Fri tags **v2.15.15–19** now Released (Direct3D fallback #141; fractional-scale segments #142). [getagentseal/codeburn](https://github.com/getagentseal/codeburn) (~11.2k★ / 11,248; +~20) — still **0.9.25**; Fri PM: mouse tracking off by default so text selection works (#1562); Cursor CSV import (#1558) still latest product surface. [ItsJazii/pane](https://github.com/ItsJazii/pane) (~**58★**, +1; **v0.4.55** ~15:21 BST Fri) — ships Claude Cloud session credits + expiry (#246); StepFun balance / Step Plan credits + spend from Claude Code/Codex/OpenCode (#245); updater proxy/retry (#244). [junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) (~5.5k★ / 5,548; +~11) — still tagged **v4.17.0**; Fri: OpenClaw compressed SQLite decode fix; prior TUI i18n + Codex/Antigravity ledger (#1365) still untagged. [mag123c/toktrack](https://github.com/mag123c/toktrack) (~**192★**, flat; still **v2.17.4**; quiet since Tue). Also name-collision adjacent: [sylearn/AIUsage](https://github.com/sylearn/AIUsage) (~684★, +1; still **v0.15.21**) — no Fri/Sat code. | Highest-velocity Sat: **Syrtis v2.0.0 rebrand** + CodeZeno Release catch-up + pane v0.4.55 + Langfuse v4.46.0. |

**Positioning refresh**
- **OpenUsage.sh** = primary collision (homepage **34/35** vs README **36**; deps-only into **day 12**).
- **ccusage** = report/parser standard (published **v20.0.24** / tagged **v20.0.25**; snapshot churn only; stars still climbing).
- **Splitrail** = #2 live local peer (**v3.10.1** hold — GPT-6 Sol/Luna still latest tag).
- **juliantanx/aiusage** = local multi-tool dashboard peer + name collision (**v1.5.19** hold).
- **openusage.ai** = menu-bar naming collision (**v0.7.13-beta.2** ahead of stable; quiet since Wed).
- **Syrtis (ex-TokenBar) / CodeZeno / pane / tokscale / codeburn** = highest-velocity adjacent local apps (full rebrand v2.0.0; Release→v2.15.19; Cloud credits + StepFun shipped; TUI i18n still untagged; Cursor CSV truth).
- Keep Langfuse / LiteLLM as complementary observability/gateway spend (**Langfuse v4.46.0** notable); treat **Helicone as maintenance-mode / migrate-away** adjacent (10 quiet days).

Sources: [openusage.sh](https://openusage.sh/); [github.com/janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage) (~211★) / #380; [github.com/ccusage/ccusage](https://github.com/ccusage/ccusage) v20.0.24 / tag v20.0.25; [github.com/robinebers/openusage](https://github.com/robinebers/openusage) v0.7.13-beta.2; [github.com/juliantanx/aiusage](https://github.com/juliantanx/aiusage) v1.5.19; [github.com/Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail) v3.10.1; [github.com/Nanako0129/syrtis](https://github.com/Nanako0129/syrtis) (ex-TokenBar) v2.0.0; [github.com/CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) Release v2.15.19 / #141 / #142; [github.com/getagentseal/codeburn](https://github.com/getagentseal/codeburn) 0.9.25 / #1562 / #1558; [github.com/ItsJazii/pane](https://github.com/ItsJazii/pane) v0.4.55 / #246 / #245; [github.com/junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) v4.17.0 / OpenClaw decode; [github.com/mag123c/toktrack](https://github.com/mag123c/toktrack) v2.17.4; [github.com/sylearn/AIUsage](https://github.com/sylearn/AIUsage) v0.15.21; [cursor.com spend limits](https://cursor.com/help/account-and-billing/spend-limits); [code.claude.com costs](https://code.claude.com/docs/en/costs); [openrouter usage-accounting](https://openrouter.ai/docs/cookbook/administration/usage-accounting); [langfuse/langfuse](https://github.com/langfuse/langfuse) v4.46.0 / #17929 / #17803; [BerriAI/litellm](https://github.com/BerriAI/litellm) v1.102.1 / #43273; [langfuse.com migrate-from-helicone](https://langfuse.com/resources/engineering/migrate-from-helicone); GitHub API star/push metadata 2026-09-26 (Europe/London).

### Update 2026-09-27

Fresh Sunday scrape (Europe/London). Twin-risk with OpenUsage.sh unchanged; homepage copy drift still **three-way** — meta + “Supported providers (**35**)” vs hero/footer “**34** providers and counting” vs README **36** (incl. Antigravity) — **13th day** of product-surface quiet (last code still deps-only #380 Thu; tag still **v0.25.0** 31 Aug; **~212★**). Biggest overnight/Sun move: **Syrtis v2.0.1** (frosted-glass hover tooltips on macOS 27 Liquid Glass panel) after Sat’s TokenBar→Syrtis **v2.0.0** rebrand; main now **~58 commits ahead** of the tag with unreleased product fixes (hourly model-report refresh #413, Claude Keychain hex decode #411, quota Past-windows unread signal #405, Antigravity `agy` discovery backoff #402) plus Sun CI/selftest speed. Also: **LiteLLM** Sun ships opt-in **prompt-cache cost routing** (#43232) + SigNoz OTel v2 preset (#43296) while Latest non-prerelease stays **v1.102.1**. ccusage still published **v20.0.24** (git tag **v20.0.25** no Release) with Sun pricing-snapshot churn only. CodeZeno / pane / Splitrail / juliantanx/aiusage / toktrack / sylearn hold.

| Product | Delta | Vs agentUsage |
|---|---|---|
| [OpenUsage.sh](https://openusage.sh/) ([GitHub](https://github.com/janekbaraniewski/openusage), **~212★**, +1) | Homepage still **34 / 35 / 36** drift: hero + “34 providers, and counting”; meta + “Supported providers (**35**)”; README still **36** (incl. Antigravity). **No Sat/Sun commits** after Thu deps(docs) #380. Last product tag still **v0.25.0** (31 Aug); last product-surface code still **13 Sep** (Codex/Crush fixes). | **Nearest twin** — category definition still ~1:1. Copy smell persists; product surface quiet now **13 days** (deps-only pulses). |
| [ccusage](https://github.com/ccusage/ccusage) / [ccusage.com](https://ccusage.com/) (**~18.8k★** / 18,758; +~14) | Published release still **v20.0.24**; git tag **v20.0.25** (21 Sep) still has **no** GitHub Release. Sun commits = **models.dev / LiteLLM pricing snapshots** only (through ~08:20 BST). Still no Cursor in official source list. | Best **log-history / cost-report** OSS peer; thinner live quota/rate-limit TUI + key autodetection. |
| [Splitrail](https://github.com/Piebald-AI/splitrail) (**~222★**, flat) | Still **v3.10.1** (Thu GPT-6 Sol/Luna). **No Fri–Sun commits.** | #2 local “live monitor” peer — quiet after Thu tag. |
| [AIUsage](https://github.com/juliantanx/aiusage) (**~133★**, flat) | Still **v1.5.19** (Thu Antigravity attribution + Opus 5.5 / GPT-6 prices). **No Fri–Sun commits.** | Rising local peer + **name collision** (AIUsage vs agentUsage); quiet after Thu release. |
| [openusage.ai](https://github.com/robinebers/openusage) (**~4.3k★** / 4,276; +~9) | Stable tag still **v0.7.12**; latest pre still **v0.7.13-beta.2**. Last code still Wed (#1299 multi-account Claude terminal sessions). Repo `pushed_at` bumped Sat (events/PRs); **no new main commits** since Wed / no Sun commits. | Naming collision only — do not conflate with openusage.sh. |
| Vendor UIs | **Cursor:** Spending tab still caps on-demand usage with team/member spend limits (Teams admin + Enterprise member/group overrides / Admin API; dynamic seat-proportional limits) — Usage pane remains separate ([docs](https://cursor.com/help/account-and-billing/spend-limits)). **Claude Code:** `/usage` primary in current costs docs (session cost, plan bars, attribution, prompt-cache stats, usage-credits row); `/cost` historically aliased ([docs](https://code.claude.com/docs/en/costs)). **OpenRouter:** Activity / [usage-accounting](https://openrouter.ai/docs/cookbook/administration/usage-accounting) still API-platform credits & analytics (usage always included; deprecated `usage.include` params). | Fragmentation OpenUsage/agentUsage claim to fix. |
| Helicone / Langfuse / LiteLLM | **Helicone** (~6.2k★ / 6,179; +2): last push still **16 Sep** — **11th quiet day**. **Langfuse** (~35.1k★ / 35,091, **pushed today**, +~32) — still **v4.46.0** (Fri); Sat PM: dashboard large query params as multipart body (#17970); no new Sun release. **LiteLLM** (~59.7k★ / 59,696, **pushed today**, +~49) — latest non-prerelease still **v1.102.1** (Latest); Sun = opt-in prompt-cache **cost routing** (#43232 ~02:13 BST), SigNoz OTel v2 preset (#43296), keep `litellm` Usage on text-completion stream chunks (#43047), Gemini `function_declarations` token count (#43417). | Adjacent layers — prefer Langfuse/LiteLLM for new gateway/OTel work; Helicone still migrate-away adjacent. |
| Also watch (local peers) | **[Syrtis](https://github.com/Nanako0129/syrtis)** (ex-[TokenBar](https://github.com/Nanako0129/TokenBar); ~**374★**, +5; **v2.0.1** ~21:16 BST Sat after **v2.0.0** Sat AM) — frosted-glass hover tooltips over Liquid Glass panel (#399); main **~58 ahead** of tag with unreleased: hourly model-report refresh so prices don’t freeze (#413), Claude Keychain hex decode + named credential failures (#411), quota Past-windows “unread” signal (#405), Antigravity login-shell discovery rate-limit (#402); Sun = CI/selftest parallelization (#422/#412/#414). Site [syrtis.nyanako.com](https://syrtis.nyanako.com). [CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) (~**542★**, +4; still Release **v2.15.19**) — quiet since Fri Direct3D / fractional-scale catch-up. [getagentseal/codeburn](https://github.com/getagentseal/codeburn) (~11.3k★ / 11,255; +~7) — still **0.9.25**; last product still Fri Cursor CSV import (#1558) + mouse-tracking default (#1562). [ItsJazii/pane](https://github.com/ItsJazii/pane) (~**60★**, +2; still **v0.4.55**) — Claude Cloud credits + StepFun hold; no Sat/Sun commits. [junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) (~5.6k★ / 5,553; +~5) — still tagged **v4.17.0**; Fri TUI i18n + OpenClaw decode + Codex/Antigravity ledger (#1365) still untagged. [mag123c/toktrack](https://github.com/mag123c/toktrack) (~**192★**, flat; still **v2.17.4**; quiet since Tue). Also name-collision adjacent: [sylearn/AIUsage](https://github.com/sylearn/AIUsage) (~687★, +3; still **v0.15.21**) — no Sat/Sun code. | Highest-velocity Sun: **Syrtis v2.0.1 + unreleased quota/pricing fixes** + LiteLLM cache-aware cost routing; OpenUsage twin still idle day 13. |

**Positioning refresh**
- **OpenUsage.sh** = primary collision (homepage **34/35** vs README **36**; deps-only into **day 13**).
- **ccusage** = report/parser standard (published **v20.0.24** / tagged **v20.0.25**; snapshot churn only; stars still climbing).
- **Splitrail** = #2 live local peer (**v3.10.1** hold — GPT-6 Sol/Luna still latest tag).
- **juliantanx/aiusage** = local multi-tool dashboard peer + name collision (**v1.5.19** hold).
- **openusage.ai** = menu-bar naming collision (**v0.7.13-beta.2** ahead of stable; quiet since Wed).
- **Syrtis (ex-TokenBar) / CodeZeno / pane / tokscale / codeburn** = highest-velocity adjacent local apps (v2.0.1 + unreleased main; Release hold v2.15.19; Cloud credits hold; TUI i18n still untagged; Cursor CSV truth).
- Keep Langfuse / LiteLLM as complementary observability/gateway spend (**LiteLLM prompt-cache cost routing** notable Sun; Langfuse still **v4.46.0**); treat **Helicone as maintenance-mode / migrate-away** adjacent (11 quiet days).

Sources: [openusage.sh](https://openusage.sh/); [github.com/janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage) (~212★) / #380; [github.com/ccusage/ccusage](https://github.com/ccusage/ccusage) v20.0.24 / tag v20.0.25; [github.com/robinebers/openusage](https://github.com/robinebers/openusage) v0.7.13-beta.2; [github.com/juliantanx/aiusage](https://github.com/juliantanx/aiusage) v1.5.19; [github.com/Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail) v3.10.1; [github.com/Nanako0129/syrtis](https://github.com/Nanako0129/syrtis) (ex-TokenBar) v2.0.1 / #399 / #413 / #411 / #405; [github.com/CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) Release v2.15.19; [github.com/getagentseal/codeburn](https://github.com/getagentseal/codeburn) 0.9.25 / #1558 / #1562; [github.com/ItsJazii/pane](https://github.com/ItsJazii/pane) v0.4.55; [github.com/junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) v4.17.0 / #1365; [github.com/mag123c/toktrack](https://github.com/mag123c/toktrack) v2.17.4; [github.com/sylearn/AIUsage](https://github.com/sylearn/AIUsage) v0.15.21; [cursor.com spend limits](https://cursor.com/help/account-and-billing/spend-limits); [code.claude.com costs](https://code.claude.com/docs/en/costs); [openrouter usage-accounting](https://openrouter.ai/docs/cookbook/administration/usage-accounting); [langfuse/langfuse](https://github.com/langfuse/langfuse) v4.46.0 / #17970; [BerriAI/litellm](https://github.com/BerriAI/litellm) v1.102.1 / #43232 / #43296 / #43047; [langfuse.com migrate-from-helicone](https://langfuse.com/resources/engineering/migrate-from-helicone); GitHub API star/push metadata 2026-09-27 (Europe/London).


### Update 2026-09-28

Fresh Monday scrape (Europe/London / BST). Twin-risk with OpenUsage.sh unchanged; homepage copy drift still **three-way** — meta + “Supported providers (**35**)” vs hero/footer “**34** providers and counting” vs README **36** (incl. Antigravity) — **14th day** of product-surface quiet (last code still deps-only #380 Thu; tag still **v0.25.0** 31 Aug; **~213★**). Biggest overnight/Mon move: **Syrtis v2.1.0** (~20:43 BST Sun) — one-click usage→subscription attribution cards (#437), Grok Build model unify (#417), stale menu-bar gauge greys after 30m (#418), clock-back quota-history repair (#415), plus Sun’s unreleased pricing/quota fixes now cut into the tag; main now **~27 commits ahead** (tray sand animation #441, first-run onboarding cards #442, Settings plain-language rewrite #443, popover scroll lenses #444). Also: **ccusage** published **v20.0.26** (Sun ~17:26 BST; npm skipped **20.0.25**) with Codex session-id filter + Claude/OpenCode report fixes; Mon = models.dev/LiteLLM pricing snapshots only. **LiteLLM** Latest jumps **v1.102.1 → v1.103.0** (~06:43 BST) + rust MCP gateway groundwork (#43470) while cost-map syncs OpenRouter prices (#43506). CodeZeno / pane / Splitrail / juliantanx/aiusage / toktrack / sylearn hold.

| Product | Delta | Vs agentUsage |
|---|---|---|
| [OpenUsage.sh](https://openusage.sh/) ([GitHub](https://github.com/janekbaraniewski/openusage), **~213★**, +1) | Homepage still **34 / 35 / 36** drift: hero + “34 providers, and counting”; meta + “Supported providers (**35**)”; README still **36** (incl. Antigravity). **No Fri–Mon commits** after Thu deps(docs) #380. Last product tag still **v0.25.0** (31 Aug); last product-surface code still **13 Sep** (Codex/Crush fixes). | **Nearest twin** — category definition still ~1:1. Copy smell persists; product surface quiet now **14 days** (deps-only pulses). |
| [ccusage](https://github.com/ccusage/ccusage) / [ccusage.com](https://ccusage.com/) (**~18.8k★** / 18,772; +~14) | Published release now **v20.0.26** (Sun; Codex session-id filter #1777; Claude requestless dedupe by timestamp #1799; OpenCode exclude fork-copied history #1782). Git tag **v20.0.25** still has **no** GitHub Release and is **absent from npm** (publish jumped 20.0.24 → 20.0.26). Mon commits = **models.dev / LiteLLM pricing snapshots** only (through ~09:33 BST). Still no Cursor in official source list. | Best **log-history / cost-report** OSS peer; thinner live quota/rate-limit TUI + key autodetection. |
| [Splitrail](https://github.com/Piebald-AI/splitrail) (**~222★**, flat) | Still **v3.10.1** (Thu GPT-6 Sol/Luna). **No Fri–Mon commits.** | #2 local “live monitor” peer — quiet after Thu tag. |
| [AIUsage](https://github.com/juliantanx/aiusage) (**~133★**, flat) | Still **v1.5.19** (Thu Antigravity attribution + Opus 5.5 / GPT-6 prices). **No Fri–Mon commits.** | Rising local peer + **name collision** (AIUsage vs agentUsage); quiet after Thu release. |
| [openusage.ai](https://github.com/robinebers/openusage) (**~4.3k★** / 4,284; +~8) | Stable tag still **v0.7.12**; latest pre still **v0.7.13-beta.2**. Last code still Wed (#1299 multi-account Claude terminal sessions). Repo `pushed_at` still Sat bump; **no new main commits** since Wed / no Sun–Mon commits. | Naming collision only — do not conflate with openusage.sh. |
| Vendor UIs | **Cursor:** Spending tab still caps on-demand usage with team/member spend limits (Teams admin + Enterprise member/group overrides / Admin API; dynamic seat-proportional limits) — Usage pane remains separate ([docs](https://cursor.com/help/account-and-billing/spend-limits)). **Claude Code:** `/usage` primary in current costs docs (session cost, plan bars, attribution, prompt-cache stats, usage-credits row); `/cost` historically aliased ([docs](https://code.claude.com/docs/en/costs)). **OpenRouter:** Activity / [usage-accounting](https://openrouter.ai/docs/cookbook/administration/usage-accounting) still API-platform credits & analytics (usage always included; deprecated `usage.include` params). | Fragmentation OpenUsage/agentUsage claim to fix. |
| Helicone / Langfuse / LiteLLM | **Helicone** (~6.2k★ / 6,183; +4): last push still **16 Sep** — **12th quiet day**. **Langfuse** (~35.1k★ / 35,129, **pushed today**, +~38) — still **v4.46.0** (Fri); Mon AM: trace/session header actions (#17765), model-definitions settings table design-system migrate (#17943); no new Mon release. **LiteLLM** (~59.8k★ / 59,763, **pushed today**, +~67) — Latest non-prerelease now **v1.103.0** (~06:43 BST; was **v1.102.1**); also **v1.104.0-rc.1**; Mon = OpenRouter cost-map sync (#43506), Azure Mistral OCR pricing (#43530), rust MCP gateway + UI sessions (#43470/#43469). | Adjacent layers — prefer Langfuse/LiteLLM for new gateway/OTel work; Helicone still migrate-away adjacent. |
| Also watch (local peers) | **[Syrtis](https://github.com/Nanako0129/syrtis)** (ex-[TokenBar](https://github.com/Nanako0129/TokenBar); ~**377★**, +3; **v2.1.0** ~20:43 BST Sun after **v2.0.1**) — attribution suggestion cards (#437), Grok Build single-name models (#417), stale gauge greys (#418), clock-back history keep (#415/#433/#434); main **~27 ahead** of tag with Mon: tray sand animation (#441), first-run onboarding cards (#442), Settings copy rewrite (#443), per-lens popover scroll (#444). Site [syrtis.nyanako.com](https://syrtis.nyanako.com). [CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) (~**543★**, +1; still Release **v2.15.19**) — quiet since Fri Direct3D / fractional-scale catch-up. [getagentseal/codeburn](https://github.com/getagentseal/codeburn) (~11.3k★ / 11,265; +~10) — still **0.9.25**; Sun product: Codex fork-replay preserve (#1569), GNOME paired-device combined usage (#1566), menubar package-manager symlink (#1563) + dock-at-rest rings (#1560). [ItsJazii/pane](https://github.com/ItsJazii/pane) (~**61★**, +1; still **v0.4.55**) — Claude Cloud credits + StepFun hold; no Sat–Mon commits. [junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) (~5.6k★ / 5,563; +~10) — still tagged **v4.17.0**; Fri TUI i18n + OpenClaw decode + Codex/Antigravity ledger (#1365) still untagged. [mag123c/toktrack](https://github.com/mag123c/toktrack) (~**192★**, flat; still **v2.17.4**; quiet since Tue). Also name-collision adjacent: [sylearn/AIUsage](https://github.com/sylearn/AIUsage) (~688★, +1; still **v0.15.21**) — no Sat–Mon code. | Highest-velocity Mon: **Syrtis v2.1.0 + onboarding/tray unreleased** + **ccusage v20.0.26** + LiteLLM **v1.103.0**; OpenUsage twin still idle day 14. |

**Positioning refresh**
- **OpenUsage.sh** = primary collision (homepage **34/35** vs README **36**; deps-only into **day 14**).
- **ccusage** = report/parser standard (published **v20.0.26**; **v20.0.25** tag-only / not on npm; snapshot churn Mon; stars still climbing).
- **Splitrail** = #2 live local peer (**v3.10.1** hold — GPT-6 Sol/Luna still latest tag).
- **juliantanx/aiusage** = local multi-tool dashboard peer + name collision (**v1.5.19** hold).
- **openusage.ai** = menu-bar naming collision (**v0.7.13-beta.2** ahead of stable; quiet since Wed).
- **Syrtis (ex-TokenBar) / CodeZeno / pane / tokscale / codeburn** = highest-velocity adjacent local apps (v2.1.0 + unreleased main; Release hold v2.15.19; Cloud credits hold; TUI i18n still untagged; Codex fork + GNOME paired usage).
- Keep Langfuse / LiteLLM as complementary observability/gateway spend (**LiteLLM Latest v1.103.0** + OpenRouter cost-map sync notable Mon; Langfuse still **v4.46.0**); treat **Helicone as maintenance-mode / migrate-away** adjacent (12 quiet days).

Sources: [openusage.sh](https://openusage.sh/); [github.com/janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage) (~213★) / #380; [github.com/ccusage/ccusage](https://github.com/ccusage/ccusage) v20.0.26 / tag v20.0.25; [github.com/robinebers/openusage](https://github.com/robinebers/openusage) v0.7.13-beta.2; [github.com/juliantanx/aiusage](https://github.com/juliantanx/aiusage) v1.5.19; [github.com/Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail) v3.10.1; [github.com/Nanako0129/syrtis](https://github.com/Nanako0129/syrtis) (ex-TokenBar) v2.1.0 / #437 / #417 / #418 / #441 / #442; [github.com/CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) Release v2.15.19; [github.com/getagentseal/codeburn](https://github.com/getagentseal/codeburn) 0.9.25 / #1569 / #1566 / #1563; [github.com/ItsJazii/pane](https://github.com/ItsJazii/pane) v0.4.55; [github.com/junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) v4.17.0 / #1365; [github.com/mag123c/toktrack](https://github.com/mag123c/toktrack) v2.17.4; [github.com/sylearn/AIUsage](https://github.com/sylearn/AIUsage) v0.15.21; [cursor.com spend limits](https://cursor.com/help/account-and-billing/spend-limits); [code.claude.com costs](https://code.claude.com/docs/en/costs); [openrouter usage-accounting](https://openrouter.ai/docs/cookbook/administration/usage-accounting); [langfuse/langfuse](https://github.com/langfuse/langfuse) v4.46.0 / #17765 / #17943; [BerriAI/litellm](https://github.com/BerriAI/litellm) v1.103.0 / #43506 / #43470 / #43530; [langfuse.com migrate-from-helicone](https://langfuse.com/resources/engineering/migrate-from-helicone); GitHub API star/push metadata 2026-09-28 (Europe/London).

### Update 2026-09-29

Fresh Tuesday scrape (Europe/London / BST). Twin-risk with OpenUsage.sh unchanged; homepage copy drift still **three-way** — meta + “Supported providers (**35**)” vs hero/footer “**34** providers and counting” vs README **36** (incl. Antigravity) — **15th day** of product-surface quiet (Mon–Tue still deps/ci only: #383–#387; tag still **v0.25.0** 31 Aug; **~213★** flat). Biggest overnight/Tue move: **Syrtis v2.2.0** (~16:02 BST Mon) — Developer-ID signing + notarized installer DMG (#445/#454), Sand shoal animated tray icon + animation-pace curve (#441), first-run Overview setup cards (#442), plus Mon’s unreleased tray/onboarding/settings work now cut into the tag; main now only **~3 commits ahead** (landing notarized-DMG copy #456). Also: **Langfuse v4.46.0 → v4.47.0** (~09:25 BST Tue) with dedicated trace transcripts + AI-gateway capture/cache-pricing fixes; **Splitrail v3.10.2** (Mon Claude Sonnet 5.5 pricing); **CodeZeno v2.15.22** (~00:11 BST Tue Grok billing-period / violet Compact Fluent); **pane** merges unreleased widget mode #248. **ccusage** still **v20.0.26** (Tue = models.dev/LiteLLM pricing snapshots only). **LiteLLM** Latest still **v1.103.0** (+ **v1.104.0-rc.1**); Helicone quiet day **13**.

| Product | Delta | Vs agentUsage |
|---|---|---|
| [OpenUsage.sh](https://openusage.sh/) ([GitHub](https://github.com/janekbaraniewski/openusage), **~213★**, flat) | Homepage still **34 / 35 / 36** drift: hero + “34 providers, and counting”; meta + “Supported providers (**35**)”; README still **36** (incl. Antigravity). Mon–Tue = deps/ci only (#383–#387 undici/posthog/website/go/actions). Last product tag still **v0.25.0** (31 Aug); last product-surface code still **13 Sep**. | **Nearest twin** — category definition still ~1:1. Copy smell persists; product surface quiet now **15 days** (deps-only pulses). |
| [ccusage](https://github.com/ccusage/ccusage) / [ccusage.com](https://ccusage.com/) (**~18.8k★** / 18,790; +~18) | Published release still **v20.0.26** (Sun). Git tag **v20.0.25** still has **no** GitHub Release and is **absent from npm**. Tue commits = **models.dev / LiteLLM pricing snapshots** only (through ~09:27 BST). Still no Cursor in official source list. | Best **log-history / cost-report** OSS peer; thinner live quota/rate-limit TUI + key autodetection. |
| [Splitrail](https://github.com/Piebald-AI/splitrail) (**~222★**, flat) | New **v3.10.2** (~19:07 BST Mon) — Claude Sonnet 5.5 pricing (#266/#267). Was **v3.10.1** (Thu GPT-6 Sol/Luna). No Tue commits yet. | #2 local “live monitor” peer — pricing patch after Thu hold. |
| [AIUsage](https://github.com/juliantanx/aiusage) (**~134★**, +1) | Still **v1.5.19** (Thu Antigravity attribution + Opus 5.5 / GPT-6 prices). **No Fri–Tue commits.** | Rising local peer + **name collision** (AIUsage vs agentUsage); quiet after Thu release. |
| [openusage.ai](https://github.com/robinebers/openusage) (**~4.3k★** / 4,286; +~2) | Stable tag still **v0.7.12**; latest pre still **v0.7.13-beta.2**. Tue: Claude Sonnet 5.5 pricing commit (~05:31 BST). Last product code before that still Wed (#1299 multi-account Claude terminal sessions). | Naming collision only — do not conflate with openusage.sh. |
| Vendor UIs | **Cursor:** Spending tab still caps on-demand usage with team/member spend limits (Teams admin + Enterprise member/group overrides / Admin API) — Usage pane remains separate ([docs](https://cursor.com/help/account-and-billing/spend-limits)). **Claude Code:** `/usage` primary in current costs docs (session cost, plan bars, attribution, prompt-cache stats, usage-credits row); `/cost` historically aliased ([docs](https://code.claude.com/docs/en/costs)). **OpenRouter:** Activity / [usage-accounting](https://openrouter.ai/docs/cookbook/administration/usage-accounting) still API-platform credits & analytics. | Fragmentation OpenUsage/agentUsage claim to fix. |
| Helicone / Langfuse / LiteLLM | **Helicone** (~6.2k★ / 6,187; +4): last push still **16 Sep** — **13th quiet day**. **Langfuse** (~35.2k★ / 35,182, **pushed today**, +~53) — Latest now **v4.47.0** (~09:25 BST; was **v4.46.0**): dedicated trace transcripts (#17925), AI-gateway full-mode capture + env header (#17956/#17960), Anthropic/OpenAI 1-hour cache-write pricing (#17958/#17957). **LiteLLM** (~59.8k★ / 59,835, **pushed today**, +~72) — Latest non-prerelease still **v1.103.0**; also **v1.104.0-rc.1**; Tue = OTel Langfuse cache/reasoning usage_details (#43553), Bedrock Mantle Claude Opus/Sonnet 5.5 cost-map (#43647), model leaderboard page (#43649), Fireworks routers (#43641). | Adjacent layers — prefer Langfuse/LiteLLM for new gateway/OTel work; Helicone still migrate-away adjacent. |
| Also watch (local peers) | **[Syrtis](https://github.com/Nanako0129/syrtis)** (ex-[TokenBar](https://github.com/Nanako0129/TokenBar); ~**385★**, +8; **v2.2.0** ~16:02 BST Mon after **v2.1.0**) — Developer-ID signing + notarized DMG (#445/#454), Sand shoal tray animation + pace curve (#441), first-run setup cards (#442), empty-quota gauge slash (#439), stale-title grey (#440); main **~3 ahead** (landing DMG copy #456). Site [syrtis.nyanako.com](https://syrtis.nyanako.com). [CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) (~**545★**, +2; Release **v2.15.22** ~00:11 BST Tue) — Grok billing-period zero-usage + violet Compact Fluent (#149); was **v2.15.19**. [getagentseal/codeburn](https://github.com/getagentseal/codeburn) (~11.3k★ / 11,273; +~8) — still **0.9.25**; Mon: per-config Claude dock rings (#1570), OpenClaw per-agent sqlite (#1526), DeepSeek/Z.ai peak/off-peak (#1561). [ItsJazii/pane](https://github.com/ItsJazii/pane) (~**64★**, +3; still **v0.4.55**) — Mon merge **widget mode** #248 (pinned/draggable/collapsible usage bar; **5 commits ahead** of tag). [junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) (~5.6k★ / 5,570; +~7) — still tagged **v4.17.0**; Tue: import recovery-only days + TUI i18n finish (#1371), Claude Code 1-hour cache-write rate (#1374) still untagged. [mag123c/toktrack](https://github.com/mag123c/toktrack) (~**192★**, flat; still **v2.17.4**; quiet since Tue 22 Sep). Also name-collision adjacent: [sylearn/AIUsage](https://github.com/sylearn/AIUsage) (~689★, +1; still **v0.15.21**) — no Sat–Tue code. [Halloweedev/usagepal](https://github.com/Halloweedev/usagepal) (~82★; OpenUsage macOS fork) — quiet since Sat. | Highest-velocity Tue: **Syrtis v2.2.0 (notarized DMG)** + **Langfuse v4.47.0** + Splitrail/CodeZeno patches + pane widget unreleased; OpenUsage twin still idle day 15. |

**Positioning refresh**
- **OpenUsage.sh** = primary collision (homepage **34/35** vs README **36**; deps-only into **day 15**).
- **ccusage** = report/parser standard (published **v20.0.26** hold; **v20.0.25** tag-only / not on npm; snapshot churn Tue; stars still climbing).
- **Splitrail** = #2 live local peer (**v3.10.2** — Sonnet 5.5 pricing).
- **juliantanx/aiusage** = local multi-tool dashboard peer + name collision (**v1.5.19** hold).
- **openusage.ai** = menu-bar naming collision (**v0.7.13-beta.2** ahead of stable; Sonnet 5.5 pricing Tue).
- **Syrtis (ex-TokenBar) / CodeZeno / pane / tokscale / codeburn** = highest-velocity adjacent local apps (v2.2.0 notarized; v2.15.22 Grok fix; widget mode unreleased; TUI/cache-write still untagged; dock rings + OpenClaw sqlite).
- Keep Langfuse / LiteLLM as complementary observability/gateway spend (**Langfuse Latest v4.47.0** notable Tue; LiteLLM still **v1.103.0** + OTel/cost-map Tue); treat **Helicone as maintenance-mode / migrate-away** adjacent (13 quiet days).

Sources: [openusage.sh](https://openusage.sh/); [github.com/janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage) (~213★) / #387; [github.com/ccusage/ccusage](https://github.com/ccusage/ccusage) v20.0.26 / tag v20.0.25; [github.com/robinebers/openusage](https://github.com/robinebers/openusage) v0.7.13-beta.2 / Sonnet 5.5; [github.com/juliantanx/aiusage](https://github.com/juliantanx/aiusage) v1.5.19; [github.com/Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail) v3.10.2 / #266; [github.com/Nanako0129/syrtis](https://github.com/Nanako0129/syrtis) (ex-TokenBar) v2.2.0 / #445 / #454 / #441 / #442 / #456; [github.com/CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) Release v2.15.22 / #149; [github.com/getagentseal/codeburn](https://github.com/getagentseal/codeburn) 0.9.25 / #1570 / #1526 / #1561; [github.com/ItsJazii/pane](https://github.com/ItsJazii/pane) v0.4.55 / #248; [github.com/junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) v4.17.0 / #1371 / #1374; [github.com/mag123c/toktrack](https://github.com/mag123c/toktrack) v2.17.4; [github.com/sylearn/AIUsage](https://github.com/sylearn/AIUsage) v0.15.21; [github.com/Halloweedev/usagepal](https://github.com/Halloweedev/usagepal); [cursor.com spend limits](https://cursor.com/help/account-and-billing/spend-limits); [code.claude.com costs](https://code.claude.com/docs/en/costs); [openrouter usage-accounting](https://openrouter.ai/docs/cookbook/administration/usage-accounting); [langfuse/langfuse](https://github.com/langfuse/langfuse) v4.47.0 / #17925 / #17956 / #17958; [BerriAI/litellm](https://github.com/BerriAI/litellm) v1.103.0 / #43553 / #43647 / #43649; [langfuse.com migrate-from-helicone](https://langfuse.com/resources/engineering/migrate-from-helicone); GitHub API star/push metadata 2026-09-29 (Europe/London).

### Update 2026-09-30

Fresh Wednesday scrape (Europe/London / BST). Twin-risk with OpenUsage.sh unchanged on homepage copy, but the **15-day product-surface quiet broke Tue evening** — **not** quiet day 16. Material overnight/Wed: **OpenUsage** landed four product fixes (~20:00–22:11 BST Tue) — Antigravity frozen quota snapshot (#378), Z.AI credit-based Coding Plan quotas (#388), Copilot GHE hostnames (#373), Codex latest-session-by-filename (#372) — plus a hermes test isolation (#391); tag still **v0.25.0** (31 Aug); stars **~213→215**; homepage drift still **34 / 35 / 36** (hero/footer “34 providers” vs meta/h2 “Supported providers (**35**)” vs README **36** incl. Antigravity). Also: **Splitrail v3.10.2→v3.10.3** (~20:25 BST Tue) GPT-6.1 Sol pricing (#268/#269); **LiteLLM** Latest **v1.103.0→v1.103.1** (~01:59 BST Wed) plus **v1.104.0-rc.2** and **v1.105.0-dev.1**. **Syrtis** still **v2.2.0** / **~387★** (+2); main still **~3 ahead** (landing notarized-DMG copy). **ccusage** still **v20.0.26** (Wed = models.dev/LiteLLM pricing snapshots only; ~**18.8k★** / 18,812). **Langfuse** Latest still **v4.47.0** (~35.2k★ / 35,227) with GPT-6.1 Sol / Claude Sonnet 5.5 pricing commits. **pane** still **v0.4.55** / widget #248 unreleased (**5** ahead; ~**65★**). **Helicone** quiet day **14**.

| Product | Delta | Vs agentUsage |
|---|---|---|
| [OpenUsage.sh](https://openusage.sh/) ([GitHub](https://github.com/janekbaraniewski/openusage), **~215★**, +2) | Homepage still **34 / 35 / 36** drift: hero + “34 providers, and counting”; meta + “Supported providers (**35**)”; README still **36** (incl. Antigravity). **Product-surface quiet broken** Tue evening: #378 antigravity frozen snapshot, #388 zai credit Coding Plan quotas, #373 copilot GHE, #372 codex session-by-filename (+ #391 hermes test). Last product tag still **v0.25.0** (31 Aug). | **Nearest twin** — category definition still ~1:1. Copy smell persists; **product code resumed** after 15 quiet days (deps-only Mon–Tue morning). |
| [ccusage](https://github.com/ccusage/ccusage) / [ccusage.com](https://ccusage.com/) (**~18.8k★** / 18,812; +~22) | Published release still **v20.0.26** (Sun). Git tag **v20.0.25** still has **no** GitHub Release and is **absent from npm**. Wed commits = **models.dev / LiteLLM pricing snapshots** only. Still no Cursor in official source list. | Best **log-history / cost-report** OSS peer; thinner live quota/rate-limit TUI + key autodetection. |
| [Splitrail](https://github.com/Piebald-AI/splitrail) (**~222★**, flat) | New **v3.10.3** (~20:25 BST Tue) — GPT-6.1 Sol pricing (#268/#269). Was **v3.10.2** (Mon Claude Sonnet 5.5). | #2 local “live monitor” peer — pricing patch after Mon Sonnet 5.5. |
| [AIUsage](https://github.com/juliantanx/aiusage) (**~135★**, +1) | Still **v1.5.19** (Thu Antigravity attribution + Opus 5.5 / GPT-6 prices). **No Fri–Wed commits.** | Rising local peer + **name collision** (AIUsage vs agentUsage); quiet after Thu release. |
| [openusage.ai](https://github.com/robinebers/openusage) (**~4.3k★** / 4,289; +~3) | Stable tag still **v0.7.12**; latest pre still **v0.7.13-beta.2**. Pushed Wed (pricing/maintenance). | Naming collision only — do not conflate with openusage.sh. |
| Vendor UIs | **Cursor:** Spending tab still caps on-demand usage with team/member spend limits (Teams admin + Enterprise member/group overrides / Admin API) — Usage pane remains separate ([docs](https://cursor.com/help/account-and-billing/spend-limits)). **Claude Code:** `/usage` primary in current costs docs (session cost, plan bars, attribution, prompt-cache stats, usage-credits row); `/cost` historically aliased ([docs](https://code.claude.com/docs/en/costs)). **OpenRouter:** Activity / [usage-accounting](https://openrouter.ai/docs/cookbook/administration/usage-accounting) still API-platform credits & analytics. | Fragmentation OpenUsage/agentUsage claim to fix. |
| Helicone / Langfuse / LiteLLM | **Helicone** (~6.2k★ / 6,189; +2): last push still **16 Sep** — **14th quiet day**. **Langfuse** (~35.2k★ / 35,227, **pushed today**, +~45) — Latest still **v4.47.0** (Tue); Wed/Tue pricing commits GPT-6.1 Sol + Claude Sonnet 5.5 defaults (#18055/#18023). **LiteLLM** (~59.9k★ / 59,905, **pushed today**, +~70) — Latest non-prerelease now **v1.103.1** (~01:59 BST Wed; was **v1.103.0**); also **v1.104.0-rc.2** + **v1.105.0-dev.1**; Wed = router cooldown/prefetch, Bedrock beta header, MCP permissions UI. | Adjacent layers — prefer Langfuse/LiteLLM for new gateway/OTel work; Helicone still migrate-away adjacent. |
| Also watch (local peers) | **[Syrtis](https://github.com/Nanako0129/syrtis)** (ex-[TokenBar](https://github.com/Nanako0129/TokenBar); ~**387★**, +2; still **v2.2.0**) — main still **~3 ahead** (landing notarized-DMG copy #456); no Tue/Wed product commits after Mon tag. Site [syrtis.nyanako.com](https://syrtis.nyanako.com). [CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) (~**546★**, +1; still **v2.15.22**) — quiet since Tue Grok billing-period fix. [getagentseal/codeburn](https://github.com/getagentseal/codeburn) (~11.3k★ / 11,285; +~12) — still **0.9.25**; Tue: Codex response usage (#1571), DeepSeek/gpt-5.6-codex pricing invariants (#1577). [ItsJazii/pane](https://github.com/ItsJazii/pane) (~**65★**, +1; still **v0.4.55**) — widget mode #248 still **5 commits ahead** of tag (unreleased). [junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) (~5.6k★ / 5,580; +~10) — still tagged **v4.17.0**; Tue import recovery-only days + TUI i18n (#1371) / Claude Code 1-hour cache-write (#1374) still untagged. [mag123c/toktrack](https://github.com/mag123c/toktrack) (~**192★**, flat; still **v2.17.4**; Tue deps-only thiserror bump). Also name-collision adjacent: [sylearn/AIUsage](https://github.com/sylearn/AIUsage) (~692★, +3; still **v0.15.21**) — quiet. [Halloweedev/usagepal](https://github.com/Halloweedev/usagepal) (~82★; OpenUsage macOS fork) — quiet since Sat. | Highest-velocity Wed: **OpenUsage product resume** + **Splitrail v3.10.3** + **LiteLLM v1.103.1**; Syrtis/CodeZeno/pane hold; Helicone day 14. |

**Positioning refresh**
- **OpenUsage.sh** = primary collision (homepage **34/35** vs README **36**; **product surface resumed** Tue evening after 15 quiet days — still no new tag).
- **ccusage** = report/parser standard (published **v20.0.26** hold; **v20.0.25** tag-only / not on npm; snapshot churn Wed; stars still climbing).
- **Splitrail** = #2 live local peer (**v3.10.3** — GPT-6.1 Sol pricing).
- **juliantanx/aiusage** = local multi-tool dashboard peer + name collision (**v1.5.19** hold).
- **openusage.ai** = menu-bar naming collision (**v0.7.13-beta.2** ahead of stable).
- **Syrtis (ex-TokenBar) / CodeZeno / pane / tokscale / codeburn** = adjacent local apps (v2.2.0 hold ~3 ahead; v2.15.22 hold; widget mode unreleased; TUI/cache-write still untagged; Codex usage + pricing invariants).
- Keep Langfuse / LiteLLM as complementary observability/gateway spend (**LiteLLM Latest v1.103.1** notable Wed; Langfuse still **v4.47.0** + pricing commits); treat **Helicone as maintenance-mode / migrate-away** adjacent (14 quiet days).

Sources: [openusage.sh](https://openusage.sh/); [github.com/janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage) (~215★) / #378 / #388 / #373 / #372 / #391; [github.com/ccusage/ccusage](https://github.com/ccusage/ccusage) v20.0.26 / tag v20.0.25; [github.com/robinebers/openusage](https://github.com/robinebers/openusage) v0.7.13-beta.2; [github.com/juliantanx/aiusage](https://github.com/juliantanx/aiusage) v1.5.19; [github.com/Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail) v3.10.3 / #268; [github.com/Nanako0129/syrtis](https://github.com/Nanako0129/syrtis) (ex-TokenBar) v2.2.0 / #456; [github.com/CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) Release v2.15.22; [github.com/getagentseal/codeburn](https://github.com/getagentseal/codeburn) 0.9.25 / #1571 / #1577; [github.com/ItsJazii/pane](https://github.com/ItsJazii/pane) v0.4.55 / #248; [github.com/junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) v4.17.0 / #1371 / #1374; [github.com/mag123c/toktrack](https://github.com/mag123c/toktrack) v2.17.4; [github.com/sylearn/AIUsage](https://github.com/sylearn/AIUsage) v0.15.21; [github.com/Halloweedev/usagepal](https://github.com/Halloweedev/usagepal); [cursor.com spend limits](https://cursor.com/help/account-and-billing/spend-limits); [code.claude.com costs](https://code.claude.com/docs/en/costs); [openrouter usage-accounting](https://openrouter.ai/docs/cookbook/administration/usage-accounting); [langfuse/langfuse](https://github.com/langfuse/langfuse) v4.47.0 / #18055 / #18023; [BerriAI/litellm](https://github.com/BerriAI/litellm) v1.103.1 / v1.104.0-rc.2 / v1.105.0-dev.1; [langfuse.com migrate-from-helicone](https://langfuse.com/resources/engineering/migrate-from-helicone); GitHub API star/push metadata 2026-09-30 (Europe/London).

### Update 2026-10-01

Fresh Thursday scrape (Europe/London / BST). Twin-risk with OpenUsage.sh unchanged on homepage copy; **product-surface quiet resumed** after Tue evening burst — **no Wed/Thu commits** on [janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage); tag still **v0.25.0** (31 Aug); stars still **~215**; homepage drift still **34 / 35 / 36** (hero/footer “34 providers” vs meta/h2 “Supported providers (**35**)” vs README **36** incl. Antigravity). Material overnight/Thu: **LiteLLM** Latest **v1.103.1→v1.103.2** (~07:37 BST Thu; also still shows **v1.104.0-rc.2** / **v1.105.0-dev.1**); **Langfuse** Latest **v4.47.0→v4.48.0** (~11:12 BST Wed) — seeder incident-session + skills draft/hash features (#18061/#18062); **openusage.ai** latest pre now **v0.7.13-beta.3** (Wed ~10:28 BST; was beta.2 on Wed competitors note) — Codex multi-account card + Ultrafast/GPT-6 Sol + Cursor grok-bot-cua pricing; **sylearn/AIUsage** **v0.15.21→v0.15.22** (~08:15 BST Thu) Claude Science local skills/MCP + reasoning-effort fixes; **[majiayu000/quotabar](https://github.com/majiayu000/quotabar)** **v0.5.7→v0.5.9** (Wed) menubar quota peer (~**53★**). **Syrtis** still **v2.2.0** / **~387★**; main now **~5 ahead** (+ #457 grouped-tab window history). **ccusage** still **v20.0.26** (Thu = models.dev/LiteLLM pricing snapshots only; ~**18.8k★** / 18,826). **pane** still **v0.4.55** / now **11** ahead (README i18n #249/#250; widget #248 still unreleased; ~**65★**). **Helicone** quiet day **15** (last push still **16 Sep**).

| Product | Delta | Vs agentUsage |
|---|---|---|
| [OpenUsage.sh](https://openusage.sh/) ([GitHub](https://github.com/janekbaraniewski/openusage), **~215★**, flat) | Homepage still **34 / 35 / 36** drift. **Quiet resumed** after Tue evening product fixes — last commit still Tue ~22:11 BST (#391 hermes test). Last product tag still **v0.25.0** (31 Aug). | **Nearest twin** — category definition still ~1:1. Copy smell persists; product code paused again after Tue burst. |
| [ccusage](https://github.com/ccusage/ccusage) / [ccusage.com](https://ccusage.com/) (**~18.8k★** / 18,826; +~14) | Published release still **v20.0.26** (Sun). Git tag **v20.0.25** still has **no** GitHub Release and is **absent from npm**. Thu commits = **models.dev / LiteLLM pricing snapshots** only. Still no Cursor in official source list. | Best **log-history / cost-report** OSS peer; thinner live quota/rate-limit TUI + key autodetection. |
| [Splitrail](https://github.com/Piebald-AI/splitrail) (**~222★**, flat) | Still **v3.10.3** (Tue GPT-6.1 Sol pricing). **No Wed/Thu commits.** | #2 local “live monitor” peer — hold after Tue pricing patch. |
| [AIUsage](https://github.com/juliantanx/aiusage) (**~135★**, flat) | Still **v1.5.19** (Thu Antigravity attribution). **No Fri–Thu commits.** | Rising local peer + **name collision** (AIUsage vs agentUsage); quiet after last week’s release. |
| [openusage.ai](https://github.com/robinebers/openusage) (**~4.3k★** / 4,292; +~3) | Stable tag still **v0.7.12**; latest pre now **v0.7.13-beta.3** (Wed; was beta.2). Codex one-card-per-account + Ultrafast/GPT-6 Sol + Cursor grok-bot-cua as Grok 4.7 pricing. | Naming collision only — do not conflate with openusage.sh. |
| Vendor UIs | **Cursor:** Spending tab still caps on-demand usage with team/member spend limits ([docs](https://cursor.com/help/account-and-billing/spend-limits)). **Claude Code:** `/usage` primary in current costs docs; `/cost` historically aliased ([docs](https://code.claude.com/docs/en/costs)). **OpenRouter:** Activity / [usage-accounting](https://openrouter.ai/docs/cookbook/administration/usage-accounting) still API-platform credits & analytics. | Fragmentation OpenUsage/agentUsage claim to fix. |
| Helicone / Langfuse / LiteLLM | **Helicone** (~6.2k★ / 6,190; +1): last push still **16 Sep** — **15th quiet day**. **Langfuse** (~35.3k★ / 35,255, **pushed today**, +~28) — Latest now **v4.48.0** (Wed; was **v4.47.0**); seeder incident-session + skills draft/hash (#18061/#18062); Thu deps/API rate-limit chore. **LiteLLM** (~60.0k★ / 59,969, **pushed today**, +~64) — Latest non-prerelease now **v1.103.2** (~07:37 BST Thu; was **v1.103.1**); still **v1.104.0-rc.2** + **v1.105.0-dev.1**; Thu = guardrails Responses API, Lens traces, cost-map. | Adjacent layers — prefer Langfuse/LiteLLM for new gateway/OTel work; Helicone still migrate-away adjacent. |
| Also watch (local peers) | **[Syrtis](https://github.com/Nanako0129/syrtis)** (ex-TokenBar; ~**387★**, flat; still **v2.2.0**) — main now **~5 ahead** (landing notarized-DMG #456 + **#457** grouped-tab window history). Site [syrtis.nyanako.com](https://syrtis.nyanako.com). [CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) (~**552★**, +6; still **v2.15.22**) — quiet since Tue. [getagentseal/codeburn](https://github.com/getagentseal/codeburn) (~11.3k★ / 11,293; +~8) — still **0.9.25**; Wed: release “What’s new” multi-language stats copy. [ItsJazii/pane](https://github.com/ItsJazii/pane) (~**65★**, flat; still **v0.4.55**) — now **11** ahead (Chinese/Russian README #249/#250; widget mode #248 still unreleased). [junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) (~5.6k★ / 5,593; +~13) — still **v4.17.0**. [mag123c/toktrack](https://github.com/mag123c/toktrack) (~**192★**, flat; still **v2.17.4**). [majiayu000/quotabar](https://github.com/majiayu000/quotabar) (~**53★**; **v0.5.9**, was v0.5.7 Wed OSS note) — menubar Claude/Codex/Cursor/Grok quotas. Also name-collision adjacent: [sylearn/AIUsage](https://github.com/sylearn/AIUsage) (~692★, flat; now **v0.15.22**) — Claude Science local skills/MCP + effort fix. [Halloweedev/usagepal](https://github.com/Halloweedev/usagepal) (~82★; OpenUsage macOS fork) — quiet since Sat. [851-labs/tokenmaxxing](https://github.com/851-labs/tokenmaxxing) (~78★) — still watch-list local CLI. | Highest-velocity Thu: **LiteLLM v1.103.2** + **Langfuse v4.48.0** + **openusage.ai beta.3** + **sylearn v0.15.22** + **QuotaBar v0.5.9**; OpenUsage.sh quiet again; Helicone day 15. |

**Positioning refresh**
- **OpenUsage.sh** = primary collision (homepage **34/35** vs README **36**; product surface quiet again after Tue evening burst — still no new tag).
- **ccusage** = report/parser standard (published **v20.0.26** hold; **v20.0.25** tag-only / not on npm; snapshot churn Thu; stars still climbing).
- **Splitrail** = #2 live local peer (**v3.10.3** hold).
- **juliantanx/aiusage** = local multi-tool dashboard peer + name collision (**v1.5.19** hold).
- **openusage.ai** = menu-bar naming collision (**v0.7.13-beta.3** ahead of stable).
- **Syrtis (ex-TokenBar) / CodeZeno / pane / tokscale / codeburn / QuotaBar** = adjacent local apps (v2.2.0 ~5 ahead; v2.15.22 hold; pane 11 ahead; QuotaBar **v0.5.9**).
- Keep Langfuse / LiteLLM as complementary observability/gateway spend (**LiteLLM Latest v1.103.2** + **Langfuse v4.48.0** notable); treat **Helicone as maintenance-mode / migrate-away** adjacent (15 quiet days).

Sources: [openusage.sh](https://openusage.sh/); [github.com/janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage) (~215★); [github.com/ccusage/ccusage](https://github.com/ccusage/ccusage) v20.0.26 / 18,826★; [github.com/robinebers/openusage](https://github.com/robinebers/openusage) v0.7.13-beta.3; [github.com/juliantanx/aiusage](https://github.com/juliantanx/aiusage) v1.5.19; [github.com/Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail) v3.10.3; [github.com/Nanako0129/syrtis](https://github.com/Nanako0129/syrtis) v2.2.0 / #457; [github.com/CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) v2.15.22; [github.com/getagentseal/codeburn](https://github.com/getagentseal/codeburn) 0.9.25; [github.com/ItsJazii/pane](https://github.com/ItsJazii/pane) v0.4.55 / #249/#250; [github.com/junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) v4.17.0; [github.com/mag123c/toktrack](https://github.com/mag123c/toktrack) v2.17.4; [github.com/majiayu000/quotabar](https://github.com/majiayu000/quotabar) v0.5.9; [github.com/sylearn/AIUsage](https://github.com/sylearn/AIUsage) v0.15.22; [github.com/Halloweedev/usagepal](https://github.com/Halloweedev/usagepal); [851-labs/tokenmaxxing](https://github.com/851-labs/tokenmaxxing); [cursor.com spend limits](https://cursor.com/help/account-and-billing/spend-limits); [code.claude.com costs](https://code.claude.com/docs/en/costs); [openrouter usage-accounting](https://openrouter.ai/docs/cookbook/administration/usage-accounting); [langfuse/langfuse](https://github.com/langfuse/langfuse) v4.48.0 / #18061 / #18062; [BerriAI/litellm](https://github.com/BerriAI/litellm) v1.103.2 / v1.104.0-rc.2 / v1.105.0-dev.1; [langfuse.com migrate-from-helicone](https://langfuse.com/resources/engineering/migrate-from-helicone); GitHub API star/push metadata 2026-10-01 (Europe/London).

### Update 2026-10-02

Fresh Friday scrape (Europe/London / BST). Twin-risk with OpenUsage.sh still **product-surface quiet** — **no Wed/Thu/Fri commits** on [janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage); tag still **v0.25.0** (31 Aug); stars still **~215**; homepage drift still **34 / 35 / 36** (hero/footer “34 providers” vs meta/h2 “Supported providers (**35**)” vs README **36** incl. Antigravity). Material overnight/Fri: **openusage.ai** stable **v0.7.12→v0.7.13** (~03:48 BST Fri) — Codex multi-account card + Ultrafast/GPT-6 Sol + Cursor grok-bot-cua/Grok 4.7 + Claude Sonnet 5.5 pricing + Rate Limit Resets row; **[CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor)** **v2.15.22→v2.17.0** (~04:31 BST Fri) — GitHub Copilot usage monitoring (#151) + lock-in-taskbar (#150); **[ItsJazii/pane](https://github.com/ItsJazii/pane)** **v0.4.55→v0.4.56** (~11:14 BST Thu) — opt-in widget mode #248 shipped (pinned/draggable/collapsible glass bar) + in-card row drag; **Langfuse** Latest **v4.48.0→v4.49.0** (~10:58 BST Thu); **[851-labs/tokenmaxxing](https://github.com/851-labs/tokenmaxxing)** **cli-v0.7.1→cli-v0.7.6** (Fri burst — Windows bun/npx PATH, scheduled-sync config dirs); **Syrtis** still **v2.2.0** / now **~396★** (+~9); main **~7 ahead** (+ **#458** Antigravity captured Google accounts). **ccusage** still **v20.0.26** (Fri = models.dev/LiteLLM pricing snapshots; ~**18.8k★** / 18,835). **LiteLLM** Latest still **v1.103.2**; also **v1.104.0-rc.2** / now **v1.105.0-dev.2**. **Helicone** quiet day **16** (last push still **16 Sep**).

| Product | Delta | Vs agentUsage |
|---|---|---|
| [OpenUsage.sh](https://openusage.sh/) ([GitHub](https://github.com/janekbaraniewski/openusage), **~215★**, flat) | Homepage still **34 / 35 / 36** drift. **Quiet continues** — last commit still Tue ~22:11 BST (#391 hermes test); **no Wed–Fri commits**. Last product tag still **v0.25.0** (31 Aug). | **Nearest twin** — category definition still ~1:1. Copy smell persists; product code still paused after Tue burst. |
| [ccusage](https://github.com/ccusage/ccusage) / [ccusage.com](https://ccusage.com/) (**~18.8k★** / 18,835; +~9) | Published release still **v20.0.26** (Sun). Git tag **v20.0.25** still has **no** GitHub Release and is **absent from npm**. Fri commits = **models.dev / LiteLLM pricing snapshots** only (~12 since Fri 00:00 UTC). Still no Cursor in official source list. | Best **log-history / cost-report** OSS peer; thinner live quota/rate-limit TUI + key autodetection. |
| [Splitrail](https://github.com/Piebald-AI/splitrail) (**~222★**, flat) | Still **v3.10.3** (Tue GPT-6.1 Sol pricing). **No Wed–Fri commits.** | #2 local “live monitor” peer — hold. |
| [AIUsage](https://github.com/juliantanx/aiusage) (**~136★**, +1) | Still **v1.5.19** (Thu 24 Sep Antigravity attribution). **No commits since 24 Sep.** | Rising local peer + **name collision** (AIUsage vs agentUsage); quiet hold. |
| [openusage.ai](https://github.com/robinebers/openusage) (**~4.3k★** / 4,302; +~10) | Stable now **v0.7.13** (~03:48 BST Fri; was **v0.7.12** / pre **v0.7.13-beta.3**). Codex one-card-per-account + Ultrafast/GPT-6 Sol + Cursor grok-bot-cua/Grok 4.7 + Claude Sonnet 5.5 pricing + Rate Limit Resets; post-tag fixes prefer Cursor structured team usage pools (#1337) + Codex slow-history non-block (#1338). | Naming collision only — do not conflate with openusage.sh. |
| Vendor UIs | **Cursor:** Spending tab still caps on-demand usage with team/member spend limits; docs still describe Cursor Models vs Other Models pools ([docs](https://cursor.com/help/account-and-billing/spend-limits)). **Claude Code:** `/usage` primary in current costs docs; `/cost` historically aliased ([docs](https://code.claude.com/docs/en/costs)). **OpenRouter:** Activity / [usage-accounting](https://openrouter.ai/docs/cookbook/administration/usage-accounting) still API-platform credits & analytics (usage object always-on; `usage.include` deprecated). | Fragmentation OpenUsage/agentUsage claim to fix. |
| Helicone / Langfuse / LiteLLM | **Helicone** (~6.2k★ / 6,193; +3): last push still **16 Sep** — **16th quiet day**. **Langfuse** (~35.3k★ / 35,295, **pushed today**, +~40) — Latest now **v4.49.0** (Thu; was **v4.48.0**); Fri = evals/web polish. **LiteLLM** (~60.0k★ / 60,029, **pushed today**, +~60) — Latest non-prerelease still **v1.103.2**; still **v1.104.0-rc.2** + now **v1.105.0-dev.2** (was dev.1); Fri = proxy spend-log unmask, guardrails, cost-map, tracing SQL. | Adjacent layers — prefer Langfuse/LiteLLM for new gateway/OTel work; Helicone still migrate-away adjacent. |
| Also watch (local peers) | **[Syrtis](https://github.com/Nanako0129/syrtis)** (ex-TokenBar; ~**396★**, +~9; still **v2.2.0**) — main now **~7 ahead** (#456/#457 + **#458** Antigravity captured Google accounts as extra cards). Site [syrtis.nyanako.com](https://syrtis.nyanako.com). [CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) (~**558★**, +6; now **v2.17.0**) — Copilot tray + lock-in-taskbar. [getagentseal/codeburn](https://github.com/getagentseal/codeburn) (~11.3k★ / 11,301; +~8) — still **0.9.25**; Fri: Amp local thread-mirror support (#1578) + GLM-5.3/Flash spend tiers (#1598). [ItsJazii/pane](https://github.com/ItsJazii/pane) (~**66★**, +1; now **v0.4.56**) — widget #248 **released**; main **0** ahead. [junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) (~5.6k★ / 5,601; +~8) — still **v4.17.0**. [mag123c/toktrack](https://github.com/mag123c/toktrack) (~**193★**, +1; still **v2.17.4**) — Thu pricing-snapshot chore (#263). [majiayu000/quotabar](https://github.com/majiayu000/quotabar) (~**53★**; still **v0.5.9**) — Thu landing-page quota docs. Also name-collision adjacent: [sylearn/AIUsage](https://github.com/sylearn/AIUsage) (~694★, +2; still **v0.15.22**). [Halloweedev/usagepal](https://github.com/Halloweedev/usagepal) (~82★; OpenUsage macOS fork) — quiet since Sat. [851-labs/tokenmaxxing](https://github.com/851-labs/tokenmaxxing) (~78★; now **cli-v0.7.6**) — Fri Windows/sync release train. | Highest-velocity Fri: **openusage.ai v0.7.13** + **CodeZeno v2.17.0** + **pane v0.4.56** + **Langfuse v4.49.0** + **tokenmaxxing cli-v0.7.6** + **Syrtis #458**; OpenUsage.sh quiet again; Helicone day 16. |

**Positioning refresh**
- **OpenUsage.sh** = primary collision (homepage **34/35** vs README **36**; product surface still quiet after Tue evening burst — still no new tag).
- **ccusage** = report/parser standard (published **v20.0.26** hold; **v20.0.25** tag-only / not on npm; snapshot churn Fri; stars still climbing).
- **Splitrail** = #2 live local peer (**v3.10.3** hold).
- **juliantanx/aiusage** = local multi-tool dashboard peer + name collision (**v1.5.19** hold).
- **openusage.ai** = menu-bar naming collision (**stable v0.7.13** shipped).
- **Syrtis (ex-TokenBar) / CodeZeno / pane / tokscale / codeburn / QuotaBar** = adjacent local apps (v2.2.0 ~7 ahead; **v2.17.0**; **pane v0.4.56** widget shipped; QuotaBar **v0.5.9**).
- Keep Langfuse / LiteLLM as complementary observability/gateway spend (**LiteLLM Latest v1.103.2** + **Langfuse v4.49.0** notable); treat **Helicone as maintenance-mode / migrate-away** adjacent (16 quiet days).

Sources: [openusage.sh](https://openusage.sh/); [github.com/janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage) (~215★); [github.com/ccusage/ccusage](https://github.com/ccusage/ccusage) v20.0.26 / 18,835★; [github.com/robinebers/openusage](https://github.com/robinebers/openusage) v0.7.13; [github.com/juliantanx/aiusage](https://github.com/juliantanx/aiusage) v1.5.19; [github.com/Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail) v3.10.3; [github.com/Nanako0129/syrtis](https://github.com/Nanako0129/syrtis) v2.2.0 / #458; [github.com/CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) v2.17.0; [github.com/getagentseal/codeburn](https://github.com/getagentseal/codeburn) 0.9.25; [github.com/ItsJazii/pane](https://github.com/ItsJazii/pane) v0.4.56 / #248; [github.com/junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) v4.17.0; [github.com/mag123c/toktrack](https://github.com/mag123c/toktrack) v2.17.4; [github.com/majiayu000/quotabar](https://github.com/majiayu000/quotabar) v0.5.9; [github.com/sylearn/AIUsage](https://github.com/sylearn/AIUsage) v0.15.22; [github.com/Halloweedev/usagepal](https://github.com/Halloweedev/usagepal); [851-labs/tokenmaxxing](https://github.com/851-labs/tokenmaxxing) cli-v0.7.6; [cursor.com spend limits](https://cursor.com/help/account-and-billing/spend-limits); [code.claude.com costs](https://code.claude.com/docs/en/costs); [openrouter usage-accounting](https://openrouter.ai/docs/cookbook/administration/usage-accounting); [langfuse/langfuse](https://github.com/langfuse/langfuse) v4.49.0; [BerriAI/litellm](https://github.com/BerriAI/litellm) v1.103.2 / v1.104.0-rc.2 / v1.105.0-dev.2; [langfuse.com migrate-from-helicone](https://langfuse.com/resources/engineering/migrate-from-helicone); GitHub API star/push metadata 2026-10-02 (Europe/London).


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

### Open source — 21 Sep 2026

Fresh Monday GitHub API pass (Europe/London). Category still owned by **ccusage** + the two OpenUsage brands; **codeburn** remains the loud multi-tool local tracker. **Live today:** `Nanako0129/TokenBar` (**v1.20.0** — history-row token hover + unwritable-history repair), `CodeZeno/Claude-Code-Usage-Monitor` (**v2.12.42**, climbed from Sunday’s v2.12.39, **~501★**), `ItsJazii/pane` (**v0.4.53** — weekly-limit reset notify + star prompt), `mag123c/toktrack` (**v2.17.3** — collapse byte-identical `token_count` re-emissions), `getagentseal/codeburn` (**0.9.25** correctness/energy/menu-bar/installer fixes on top of Windows Capacity Dock), `ccusage/ccusage` (continuous **models.dev** pricing snapshots — still **v20.0.23**, ~18.7k★), `alibaba/loongsuite-pilot` (**~189★**, **+2★** — prefer Codex provider response id), `semantic-craft/iOS-vibebuddy` (**v1.3.25**). openusage.ai quiet after beta; otelite still **v0.1.153**; Helicone quiet since 16 Sep.

| Repo | Stars | License | Why it matters |
|---|---:|---|---|
| [ccusage/ccusage](https://github.com/ccusage/ccusage) | ~18.7k | NOASSERTION | **v20.0.23** + live pricing snapshots today — stay format-compatible. |
| [getagentseal/codeburn](https://github.com/getagentseal/codeburn) | ~11.1k | MIT | **0.9.25** — correctness/energy/menu-bar/installer fixes after Windows Capacity Dock plugins. |
| [robinebers/openusage](https://github.com/robinebers/openusage) | ~4.2k | MIT | openusage.ai menu-bar — Codex Swap / **v0.7.13-beta.1** still freshest multi-account UX. |
| [janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage) | ~200 | MIT | openusage.sh local multi-tool TUI + SQLite — nearest product twin (deps bumps today). |
| [ItsJazii/pane](https://github.com/ItsJazii/pane) | ~52 | NOASSERTION | **v0.4.53 today** — weekly limit reset notify + star prompt — Windows tray OpenUsage port. |
| [Maciek-roboblog/Claude-Code-Usage-Monitor](https://github.com/Maciek-roboblog/Claude-Code-Usage-Monitor) | ~8.7k | MIT | Real-time Claude Code usage monitor with predictions/warnings. |
| [HermannBjorgvin/Clawdmeter](https://github.com/HermannBjorgvin/Clawdmeter) | ~2.2k | (check) | ESP32 desk dashboard — C6 AMOLED IMU auto-rotation still the recent hardware win. |
| [Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail) | ~223 | MIT | Real-time cross-agent cost monitor. |
| [juliantanx/aiusage](https://github.com/juliantanx/aiusage) | ~129 | MIT | Still **1.5.18** — CodeBuddy CLI per-request usage. |
| [junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) | ~5.5k | MIT | **Pushed today** — Xiaomi MiMo desktop sessions split from MiMo Code + stream-json parse. |
| [Iamshankhadeep/ccseva](https://github.com/Iamshankhadeep/ccseva) | ~808 | MIT | macOS menu bar for live Claude Code usage. |
| [ColeMurray/claude-code-otel](https://github.com/ColeMurray/claude-code-otel) | ~503 | MIT | Observability stack for Claude Code usage/perf/cost (OTel path). |
| [CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) | ~501 | MIT | **v2.12.42 overnight** (from v2.12.39) — multi Claude/Codex accounts + custom refresh. |
| [leeguooooo/claude-code-usage-bar](https://github.com/leeguooooo/claude-code-usage-bar) | ~378 | MIT | Predict profiles keyed by session config dir — statusLine quota UX. |
| [foyzulkarim/claude-lens](https://github.com/foyzulkarim/claude-lens) | ~248 | MIT | Local dashboard: sessions, token costs, cache, tool calls, daily breakdowns. |
| [frankchiu-dev/claude-codex-usage-dashboard](https://github.com/frankchiu-dev/claude-codex-usage-dashboard) | ~173 | MIT | Local Windows dashboard for Claude Code + Codex limits. |
| [planetf1/otelite](https://github.com/planetf1/otelite) | ~92 | NOASSERTION | Still **v0.1.153** — storage ANALYZE + bounded ingest / drop reporting. |
| [alibaba/loongsuite-pilot](https://github.com/alibaba/loongsuite-pilot) | ~189 | Apache-2.0 | **Pushed today** — prefer Codex provider response id; multimodal/env homes still freest (**+2★**). |
| [seakee/CPA-Manager-Plus](https://github.com/seakee/CPA-Manager-Plus) | ~3.6k | MIT | Still **v1.13.1** — self-hosted CLIProxyAPI / gateway usage-cost-quota panel. |
| [Rodiun/frugon](https://github.com/Rodiun/frugon) | ~212 | MIT | **Pushed today** — curated seed pricing refresh — local LLM bill-leak analyzer. |
| [mag123c/toktrack](https://github.com/mag123c/toktrack) | ~190 | MIT | **v2.17.3 today** — collapse byte-identical token_count re-emissions + archived Codex sync. |
| [Halloweedev/usagepal](https://github.com/Halloweedev/usagepal) | ~82 | MIT | OpenUsage menu-bar fork — per-provider account switcher. |
| [Nanako0129/TokenBar](https://github.com/Nanako0129/TokenBar) | ~360 | MIT | **v1.20.0 today** — history-row token hover + unwritable-history repair — macOS quota peer. |
| [tddworks/ClaudeBar](https://github.com/tddworks/ClaudeBar) | ~1.5k | (check) | macOS menu-bar Claude quota peer. |
| [deviffyy/OpenQuota](https://github.com/deviffyy/OpenQuota) | ~216 | MIT | Cross-platform limits/spend tracker. |
| [Dicklesworthstone/coding_agent_usage_tracker](https://github.com/Dicklesworthstone/coding_agent_usage_tracker) | ~87 | NOASSERTION | Single CLI for remaining quotas across Codex/Claude/Gemini/Cursor/Copilot. |
| [sculptdotfun/viberank](https://github.com/sculptdotfun/viberank) | ~117 | MIT | Public AI-coding usage leaderboard on ccusage data. |
| [cobra91/better-ccusage](https://github.com/cobra91/better-ccusage) | ~88 | MIT | Faster multi-provider JSONL analyzer. |
| [851-labs/tokenmaxxing](https://github.com/851-labs/tokenmaxxing) | ~76 | MIT | Local CLI on ccusage that syncs usage socially (**+1★**). |
| [Nihondo/AgentLimits](https://github.com/Nihondo/AgentLimits) | ~61 | MIT | macOS widgets for Codex/Claude limits + ccusage heatmap. |
| [majiayu000/quotabar](https://github.com/majiayu000/quotabar) | ~52 | MIT | Tauri v2 menubar — six-digit tray cost tiles. |
| [abhiunix/AgentHarbor](https://github.com/abhiunix/AgentHarbor) | ~12 | MIT | Native multi-agent rate-limit + session usage app. |
| [semantic-craft/iOS-vibebuddy](https://github.com/semantic-craft/iOS-vibebuddy) | ~90 | MIT | **v1.3.25 today** — managed task instructions / phone stop — multi-surface tracker (**+1★**). |
| [zhnd/lumo](https://github.com/zhnd/lumo) | ~147 | MIT | Local-first Claude Code usage/cost/session dashboard. |
| [fschmutz/claude-usage-panel](https://github.com/fschmutz/claude-usage-panel) | ~6 | MIT | Official usage API → GNOME/macOS/statusLine + MCP get_usage (Claude + Cursor). |
| [sylearn/AIUsage](https://github.com/sylearn/AIUsage) | ~679 | Apache-2.0 | Desktop multi-tool usage peer (docs/sponsor polish today). |
| [BerriAI/litellm](https://github.com/BerriAI/litellm) | ~59.3k | NOASSERTION | **Pushed today.** Gateway spend/pricing — adjacent. |
| [langfuse/langfuse](https://github.com/langfuse/langfuse) | ~34.9k | NOASSERTION | **Pushed today.** OSS LLM observability + OTel. |
| [Helicone/helicone](https://github.com/Helicone/helicone) | ~6.2k | Apache-2.0 | Quiet since 16 Sep. OSS LLM observability proxy. Adjacent. |

**Build takeaways:** Stay ccusage-compatible (**v20.0.23**); differentiate on live quotas + auto-detect + UX vs openusage.sh. Today's sharpest signals: **TokenBar v1.20.0**, **CodeZeno v2.12.42**, **pane v0.4.53**, **toktrack v2.17.3**, **codeburn 0.9.25**. Treat CPA-Manager-Plus as gateway-ops adjacency. Keep Langfuse/LiteLLM/Helicone complementary.

Sources: GitHub API 2026-09-21 (Europe/London).

### Open source — 22 Sep 2026

Fresh Tuesday GitHub API pass (Europe/London). Category still owned by **ccusage** + the two OpenUsage brands; **codeburn** remains the loud multi-tool local tracker. **Live today / overnight:** `ccusage/ccusage` (**v20.0.24** — models.dev pricing resilience [#1767](https://github.com/ccusage/ccusage/issues/1767) + continuous LiteLLM/models.dev snapshots, **~18,683★**), `CodeZeno/Claude-Code-Usage-Monitor` (**v2.13.43** overnight from Monday’s v2.12.42, **~509★**), `Piebald-AI/splitrail` (**#258** peak/off-peak pricing + official per-model cite — Monday evening after ~10 quiet days), `mag123c/toktrack` (still **v2.17.3**; today **OpenCode v2** compaction/usage + stream SQLite rows [#259](https://github.com/mag123c/toktrack/pull/259)/[#261](https://github.com/mag123c/toktrack/pull/261)), `junhoyeo/tokscale` (**#1355** MiMo / MiniMax / Muse usage accounting overnight, **~5,507★**), `seakee/CPA-Manager-Plus` (**v1.13.2 today** — Muse/Meta end-to-end + Codex reset-credit / xAI weekly / Devin cache fixes), `tddworks/ClaudeBar` (status colors + current Claude generation pricing), `semantic-craft/iOS-vibebuddy` (**Mac 1.3.30 / iOS 1.3.26**), `alibaba/loongsuite-pilot` (qwen-code subagent traces + Codex/hooks perf). TokenBar still **v1.20.0** (~362★); codeburn still **0.9.25**; pane still **v0.4.53**; openusage.sh still **~200★** deps-only (product quiet continues into a second week); openusage.ai quiet after **v0.7.13-beta.1**; otelite still **v0.1.153**. Helicone quiet since **16 Sep** (6th day); LiteLLM / Langfuse active.

| Repo | Stars | License | Why it matters |
|---|---:|---|---|
| [ccusage/ccusage](https://github.com/ccusage/ccusage) | ~18.7k | NOASSERTION | **v20.0.24** + live pricing snapshots today — stay format-compatible. |
| [getagentseal/codeburn](https://github.com/getagentseal/codeburn) | ~11.2k | MIT | Still **0.9.25** (Monday) — correctness/energy/menu-bar/installer + Flathub hash. |
| [robinebers/openusage](https://github.com/robinebers/openusage) | ~4.2k | MIT | openusage.ai menu-bar — Codex Swap / **v0.7.13-beta.1** still freshest multi-account UX. |
| [janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage) | ~200 | MIT | openusage.sh local multi-tool TUI + SQLite — nearest product twin (deps-only Monday; product quiet continues). |
| [ItsJazii/pane](https://github.com/ItsJazii/pane) | ~52 | NOASSERTION | Still **v0.4.53** — weekly limit reset notify + star prompt — Windows tray OpenUsage port. |
| [Maciek-roboblog/Claude-Code-Usage-Monitor](https://github.com/Maciek-roboblog/Claude-Code-Usage-Monitor) | ~8.7k | MIT | Real-time Claude Code usage monitor with predictions/warnings. |
| [HermannBjorgvin/Clawdmeter](https://github.com/HermannBjorgvin/Clawdmeter) | ~2.2k | (check) | ESP32 desk dashboard — C6 AMOLED IMU auto-rotation still the recent hardware win. |
| [Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail) | ~223 | MIT | **#258 Monday** — peak/off-peak pricing + official cite per model — number-two local peer. |
| [juliantanx/aiusage](https://github.com/juliantanx/aiusage) | ~129 | MIT | Still **1.5.18** — CodeBuddy CLI per-request usage. |
| [junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) | ~5.5k | MIT | **#1355 overnight** — correct MiMo / MiniMax / Muse usage accounting (+ Muse Code CLI tracking Monday). |
| [Iamshankhadeep/ccseva](https://github.com/Iamshankhadeep/ccseva) | ~807 | MIT | macOS menu bar for live Claude Code usage. |
| [ColeMurray/claude-code-otel](https://github.com/ColeMurray/claude-code-otel) | ~503 | MIT | Observability stack for Claude Code usage/perf/cost (OTel path). |
| [CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) | ~509 | MIT | **v2.13.43 overnight** (from v2.12.42) — multi Claude/Codex accounts + custom refresh. |
| [leeguooooo/claude-code-usage-bar](https://github.com/leeguooooo/claude-code-usage-bar) | ~377 | MIT | Predict profiles keyed by session config dir — statusLine quota UX. |
| [foyzulkarim/claude-lens](https://github.com/foyzulkarim/claude-lens) | ~248 | MIT | Local dashboard: sessions, token costs, cache, tool calls, daily breakdowns. |
| [frankchiu-dev/claude-codex-usage-dashboard](https://github.com/frankchiu-dev/claude-codex-usage-dashboard) | ~173 | MIT | Local Windows dashboard for Claude Code + Codex limits. |
| [planetf1/otelite](https://github.com/planetf1/otelite) | ~92 | NOASSERTION | Still **v0.1.153** — storage ANALYZE + bounded ingest / drop reporting. |
| [alibaba/loongsuite-pilot](https://github.com/alibaba/loongsuite-pilot) | ~191 | Apache-2.0 | **Pushed today** — qwen-code recursive subagent traces + Codex/hooks perf; multimodal/env homes still freest. |
| [seakee/CPA-Manager-Plus](https://github.com/seakee/CPA-Manager-Plus) | ~3.6k | MIT | **v1.13.2 today** — Muse/Meta end-to-end + Codex/xAI/Devin quota correctness — gateway-ops adjacency. |
| [Rodiun/frugon](https://github.com/Rodiun/frugon) | ~212 | MIT | Local LLM bill-leak analyzer (seed pricing refresh Monday). |
| [mag123c/toktrack](https://github.com/mag123c/toktrack) | ~190 | MIT | **v2.17.3** + **today** OpenCode v2 compaction/usage + stream SQLite (#259/#261). |
| [Halloweedev/usagepal](https://github.com/Halloweedev/usagepal) | ~82 | MIT | OpenUsage menu-bar fork — per-provider account switcher. |
| [Nanako0129/TokenBar](https://github.com/Nanako0129/TokenBar) | ~362 | MIT | Still **v1.20.0** — history-row token hover + unwritable-history repair — macOS quota peer (**+2★**). |
| [tddworks/ClaudeBar](https://github.com/tddworks/ClaudeBar) | ~1.5k | (check) | **Pushed today** — status colors / High Contrast + current Claude generation pricing. |
| [deviffyy/OpenQuota](https://github.com/deviffyy/OpenQuota) | ~216 | MIT | Cross-platform limits/spend tracker. |
| [Dicklesworthstone/coding_agent_usage_tracker](https://github.com/Dicklesworthstone/coding_agent_usage_tracker) | ~87 | NOASSERTION | Single CLI for remaining quotas across Codex/Claude/Gemini/Cursor/Copilot. |
| [sculptdotfun/viberank](https://github.com/sculptdotfun/viberank) | ~117 | MIT | Public AI-coding usage leaderboard on ccusage data. |
| [cobra91/better-ccusage](https://github.com/cobra91/better-ccusage) | ~87 | MIT | Faster multi-provider JSONL analyzer. |
| [851-labs/tokenmaxxing](https://github.com/851-labs/tokenmaxxing) | ~76 | MIT | **Pushed today** — local CLI on ccusage that syncs usage socially (www time-range experiment reverted). |
| [Nihondo/AgentLimits](https://github.com/Nihondo/AgentLimits) | ~61 | MIT | macOS widgets for Codex/Claude limits + ccusage heatmap. |
| [majiayu000/quotabar](https://github.com/majiayu000/quotabar) | ~52 | MIT | Tauri v2 menubar — six-digit tray cost tiles. |
| [abhiunix/AgentHarbor](https://github.com/abhiunix/AgentHarbor) | ~12 | MIT | Native multi-agent rate-limit + session usage — Windows Codex/Cursor git-spawn console fix Monday. |
| [semantic-craft/iOS-vibebuddy](https://github.com/semantic-craft/iOS-vibebuddy) | ~90 | MIT | **Mac 1.3.30 / iOS 1.3.26 today** — multi-surface Claude/Codex/Grok/Cursor tracker. |
| [zhnd/lumo](https://github.com/zhnd/lumo) | ~147 | MIT | Local-first Claude Code usage/cost/session dashboard. |
| [fschmutz/claude-usage-panel](https://github.com/fschmutz/claude-usage-panel) | ~6 | MIT | Official usage API → GNOME/macOS/statusLine + MCP get_usage (Claude + Cursor). |
| [sylearn/AIUsage](https://github.com/sylearn/AIUsage) | ~681 | Apache-2.0 | Desktop multi-tool usage peer — Sonoma icon polish today. |
| [BerriAI/litellm](https://github.com/BerriAI/litellm) | ~59.4k | NOASSERTION | **Pushed today.** Gateway spend/pricing — JWT team alias / Presidio spend logs / compact-to-fit. |
| [langfuse/langfuse](https://github.com/langfuse/langfuse) | ~34.9k | NOASSERTION | **Pushed today.** OSS LLM observability + OTel — Anthropic AI-gateway connections. |
| [Helicone/helicone](https://github.com/Helicone/helicone) | ~6.2k | Apache-2.0 | Quiet since 16 Sep (6th day). OSS LLM observability proxy. Adjacent. |

**Also watch**
- [ClaudeCodeUsage/ClaudeCodeUsage](https://github.com/ClaudeCodeUsage/ClaudeCodeUsage) (~64★, MIT, **v2.3.3**) — VS Code status-bar Claude Code usage/cost extension; active today.
- [itvincent-git/codex-usage-desktop](https://github.com/itvincent-git/codex-usage-desktop) (~38★, MIT, **app-v3.5.1 today**) — local-first Codex quota desktop (macOS + Windows).
- [saeedkolivand/claude-usage-mac](https://github.com/saeedkolivand/claude-usage-mac) (~15★) — Claude Code macOS menu bar + desktop widget; active today.
- [tyuan511/vibe-usage](https://github.com/tyuan511/vibe-usage) (~4★) — ccusage-inspired macOS menu bar (Claude/Codex/Gemini/Qwen).
- [ofershap/cursor-usage-tracker](https://github.com/ofershap/cursor-usage-tracker) (~33★, MIT) — Cursor Enterprise FinOps, not personal autodetection.
- Micro/zero-★ Claude wrappers keep spawning daily (pace-guard, statusline, team dashboards) — noise floor, not category peers.

**Build takeaways:** Stay ccusage-compatible (**v20.0.24**); differentiate on live quotas + auto-detect + UX vs openusage.sh (still product-quiet). Today's sharpest signals: **ccusage v20.0.24**, **CodeZeno v2.13.43**, **Splitrail peak/off-peak #258**, **toktrack OpenCode v2 #259/#261**, **tokscale Muse/MiMo #1355**, **CPA v1.13.2 Muse/Meta**. Treat CPA-Manager-Plus as gateway-ops adjacency. Keep Langfuse/LiteLLM complementary; Helicone remains maintenance-mode adjacent.

Sources: GitHub API 2026-09-22 (Europe/London).

### Open source — 23 Sep 2026

Fresh Wednesday GitHub API pass (Europe/London). Category still owned by **ccusage** + the two OpenUsage brands; **codeburn** remains the loud multi-tool local tracker. **Live today / overnight:** `CodeZeno/Claude-Code-Usage-Monitor` (**v2.14.55 today** from Tue's v2.13.43 — **Grok Build usage monitoring** + system-theme floating card + updater SemVer integrity + Haiku alias probes; **~517★**, **+8★**), `Piebald-AI/splitrail` (**v3.10.0** Tue — ships Mon **#258** peak/off-peak + **Opus 5.5** / Opus fast-mode + **Grok 4.7** / 200K xAI + Grok 4.6 + Gemini/MiniMax/DeepSeek rate sync; **~222★**), `mag123c/toktrack` (**v2.17.4** Tue — tagged OpenCode v2 compaction/usage + stream SQLite [#259](https://github.com/mag123c/toktrack/pull/259)/[#261](https://github.com/mag123c/toktrack/pull/261); **~191★**), `Nanako0129/TokenBar` (**v1.20.1** Tue — unpriced models show `—` not `$0.00` + `<$0.01` for sub-cent; **~364★**), `robinebers/openusage` (**v0.7.13-beta.2 today** — Claude Rate Limit Resets row + Cursor Grok 4.7 pricing + OpenCode 2 ChatGPT OAuth attribution; **~4.2k★**), `ccusage/ccusage` (still **v20.0.24**; continuous models.dev/LiteLLM pricing snapshots today; **~18,698★**), `getagentseal/codeburn` (still **0.9.25**; today README sponsor-tier docs only; **~11.2k★**), `janekbaraniewski/openusage` (**~203★**, **+3★** — still deps-only; product quiet continues into **day 9**), `semantic-craft/iOS-vibebuddy` (today `claude agents --json` background sessions), `ClaudeCodeUsage/ClaudeCodeUsage` (today incremental window/perf), `alibaba/loongsuite-pilot` (pushed today). pane still **v0.4.53**; otelite still **v0.1.153**; CPA still **v1.13.2**; Helicone quiet since **16 Sep** (7th day).

| Repo | Stars | License | Why it matters |
|---|---:|---|---|
| [ccusage/ccusage](https://github.com/ccusage/ccusage) | ~18.7k | NOASSERTION | Still **v20.0.24** + live pricing snapshots today — stay format-compatible (**~18,698★**). |
| [getagentseal/codeburn](https://github.com/getagentseal/codeburn) | ~11.2k | MIT | Still **0.9.25** — today README sponsor tiers only; correctness/energy/menu-bar still the Mon story. |
| [robinebers/openusage](https://github.com/robinebers/openusage) | ~4.2k | MIT | openusage.ai — **v0.7.13-beta.2 today** Rate Limit Resets + Cursor Grok 4.7 + OpenCode 2 OAuth. |
| [janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage) | ~203 | MIT | openusage.sh local multi-tool TUI + SQLite — nearest product twin (**+3★**; product quiet day 9). |
| [ItsJazii/pane](https://github.com/ItsJazii/pane) | ~53 | NOASSERTION | Still **v0.4.53** — weekly limit reset notify — Windows tray OpenUsage port (**+1★**). |
| [Maciek-roboblog/Claude-Code-Usage-Monitor](https://github.com/Maciek-roboblog/Claude-Code-Usage-Monitor) | ~8.7k | MIT | Real-time Claude Code usage monitor with predictions/warnings. |
| [HermannBjorgvin/Clawdmeter](https://github.com/HermannBjorgvin/Clawdmeter) | ~2.2k | (check) | ESP32 desk dashboard — C6 AMOLED IMU auto-rotation still the recent hardware win. |
| [Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail) | ~222 | MIT | **v3.10.0** — peak/off-peak + Opus 5.5 + Grok 4.7 — number-two local peer. |
| [juliantanx/aiusage](https://github.com/juliantanx/aiusage) | ~131 | MIT | Still **1.5.18** — CodeBuddy CLI per-request usage (**+2★**). |
| [junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) | ~5.5k | MIT | Quiet after Mon **#1355** MiMo / MiniMax / Muse accounting (**~5,514★**). |
| [Iamshankhadeep/ccseva](https://github.com/Iamshankhadeep/ccseva) | ~807 | MIT | macOS menu bar for live Claude Code usage. |
| [ColeMurray/claude-code-otel](https://github.com/ColeMurray/claude-code-otel) | ~504 | MIT | Observability stack for Claude Code usage/perf/cost (OTel path). |
| [CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) | ~517 | MIT | **v2.14.55 today** — Grok Build monitoring + theme/updater/Haiku fixes (**+8★**). |
| [leeguooooo/claude-code-usage-bar](https://github.com/leeguooooo/claude-code-usage-bar) | ~377 | MIT | Predict profiles keyed by session config dir — statusLine quota UX. |
| [foyzulkarim/claude-lens](https://github.com/foyzulkarim/claude-lens) | ~248 | MIT | Local dashboard: sessions, token costs, cache, tool calls, daily breakdowns. |
| [frankchiu-dev/claude-codex-usage-dashboard](https://github.com/frankchiu-dev/claude-codex-usage-dashboard) | ~173 | MIT | Local Windows dashboard for Claude Code + Codex limits. |
| [planetf1/otelite](https://github.com/planetf1/otelite) | ~94 | NOASSERTION | Still **v0.1.153** — storage ANALYZE + bounded ingest (**+2★**). |
| [alibaba/loongsuite-pilot](https://github.com/alibaba/loongsuite-pilot) | ~191 | Apache-2.0 | **Pushed today** — qwen-code / Codex/hooks path; multimodal/env homes still freest. |
| [seakee/CPA-Manager-Plus](https://github.com/seakee/CPA-Manager-Plus) | ~3.6k | MIT | Still **v1.13.2** — Muse/Meta + Codex/xAI/Devin quota correctness — gateway-ops adjacency. |
| [Rodiun/frugon](https://github.com/Rodiun/frugon) | ~212 | MIT | Local LLM bill-leak analyzer. |
| [mag123c/toktrack](https://github.com/mag123c/toktrack) | ~191 | MIT | **v2.17.4** — OpenCode v2 compaction/usage + stream SQLite tagged (**+1★**). |
| [Halloweedev/usagepal](https://github.com/Halloweedev/usagepal) | ~82 | MIT | OpenUsage menu-bar fork — per-provider account switcher. |
| [Nanako0129/TokenBar](https://github.com/Nanako0129/TokenBar) | ~364 | MIT | **v1.20.1** — unpriced→`—` + sub-cent `<$0.01` — macOS quota peer (**+2★**). |
| [tddworks/ClaudeBar](https://github.com/tddworks/ClaudeBar) | ~1.5k | (check) | Quiet after Tue status colors / High Contrast + current Claude generation pricing. |
| [deviffyy/OpenQuota](https://github.com/deviffyy/OpenQuota) | ~216 | MIT | Cross-platform limits/spend tracker. |
| [Dicklesworthstone/coding_agent_usage_tracker](https://github.com/Dicklesworthstone/coding_agent_usage_tracker) | ~87 | NOASSERTION | Single CLI for remaining quotas across Codex/Claude/Gemini/Cursor/Copilot. |
| [sculptdotfun/viberank](https://github.com/sculptdotfun/viberank) | ~117 | MIT | Public AI-coding usage leaderboard on ccusage data. |
| [cobra91/better-ccusage](https://github.com/cobra91/better-ccusage) | ~87 | MIT | Faster multi-provider JSONL analyzer. |
| [851-labs/tokenmaxxing](https://github.com/851-labs/tokenmaxxing) | ~76 | MIT | Local CLI on ccusage that syncs usage socially. |
| [Nihondo/AgentLimits](https://github.com/Nihondo/AgentLimits) | ~63 | MIT | macOS widgets for Codex/Claude limits + ccusage heatmap (**+2★**). |
| [majiayu000/quotabar](https://github.com/majiayu000/quotabar) | ~53 | MIT | Tauri v2 menubar — six-digit tray cost tiles (**+1★**). |
| [abhiunix/AgentHarbor](https://github.com/abhiunix/AgentHarbor) | ~12 | MIT | Native multi-agent rate-limit + session usage app. |
| [semantic-craft/iOS-vibebuddy](https://github.com/semantic-craft/iOS-vibebuddy) | ~90 | MIT | **Today** `claude agents --json` background sessions — multi-surface tracker. |
| [zhnd/lumo](https://github.com/zhnd/lumo) | ~147 | MIT | Local-first Claude Code usage/cost/session dashboard. |
| [fschmutz/claude-usage-panel](https://github.com/fschmutz/claude-usage-panel) | ~6 | MIT | Official usage API → GNOME/macOS/statusLine + MCP get_usage (Claude + Cursor). |
| [sylearn/AIUsage](https://github.com/sylearn/AIUsage) | ~681 | Apache-2.0 | Desktop multi-tool usage peer — pushed today. |
| [BerriAI/litellm](https://github.com/BerriAI/litellm) | ~59.4k | NOASSERTION | Gateway spend/pricing — adjacent. |
| [langfuse/langfuse](https://github.com/langfuse/langfuse) | ~34.9k | NOASSERTION | OSS LLM observability + OTel — adjacent. |
| [Helicone/helicone](https://github.com/Helicone/helicone) | ~6.2k | Apache-2.0 | Quiet since 16 Sep (7th day). OSS LLM observability proxy. Adjacent. |

**Also watch**
- [ClaudeCodeUsage/ClaudeCodeUsage](https://github.com/ClaudeCodeUsage/ClaudeCodeUsage) (~64★, MIT) — **today** incremental window/perf; VS Code status-bar Claude Code usage/cost.
- [itvincent-git/codex-usage-desktop](https://github.com/itvincent-git/codex-usage-desktop) (~41★, MIT, **+3★**) — local-first Codex quota desktop (macOS + Windows).
- [saeedkolivand/claude-usage-mac](https://github.com/saeedkolivand/claude-usage-mac) (~15★) — Claude Code macOS menu bar + desktop widget.
- [ofershap/cursor-usage-tracker](https://github.com/ofershap/cursor-usage-tracker) (~33★, MIT) — Cursor Enterprise FinOps, not personal autodetection.
- Micro/zero-★ Claude wrappers keep spawning daily — noise floor, not category peers.

**Build takeaways:** Stay ccusage-compatible (**v20.0.24**); differentiate on live quotas + auto-detect + UX vs openusage.sh (still product-quiet day 9). Today's sharpest signals: **CodeZeno v2.14.55 Grok Build**, **Splitrail v3.10.0 Opus 5.5 / Grok 4.7**, **toktrack v2.17.4**, **TokenBar v1.20.1**, **openusage.ai v0.7.13-beta.2**. Treat CPA-Manager-Plus as gateway-ops adjacency. Keep Langfuse/LiteLLM complementary; Helicone remains maintenance-mode adjacent.

Sources: GitHub API 2026-09-23 (Europe/London).

### Open source — 24 Sep 2026

Fresh Thursday GitHub API pass (Europe/London). Category still owned by **ccusage** + the two OpenUsage brands; **codeburn** remains the loud multi-tool local tracker. **Live today / overnight:** `CodeZeno/Claude-Code-Usage-Monitor` (**v2.14.55→v2.15.0→v2.15.14 overnight** — Antigravity OAuth refresh-on-401 + Retry-After cooldown caps + tray DPI snap + builder focus/catalogue fixes; still **~517★**), `Piebald-AI/splitrail` (**v3.10.1 today** — **GPT-6 Sol / Luna** API pricing on top of Wed's peak/off-peak + Opus 5.5 / Grok 4.7; **~222★**), `Nanako0129/TokenBar` (**v1.20.2 today** — skip Antigravity `agy` CLI when Keychain login item missing so signed-out refresh no longer spam-opens Google sign-in; **~367★**, **+3★**), `ItsJazii/pane` (**v0.4.54** Wed — Claude banked limit-reset redeem + weekly Codex/Claude capacity at 100%; **~55★**, **+2★**), `juliantanx/aiusage` (**v1.5.19 today** — Antigravity model-id attribution fix + cloud sync/pricing; **~132★**, **+1★**), `langfuse/langfuse` (**v4.44.0 today** — billing org name + design-system exports + confusion-matrix key fix; **~35.0k★**), `ccusage/ccusage` (still published **v20.0.24**; tag **v20.0.25** exists with **no Release**; continuous models.dev/LiteLLM pricing snapshots today; **~18,727★**), `getagentseal/codeburn` (still **0.9.25**; **today** Opus 5.5 pricing snapshot + low prompt-cache-hit session flag; **~11.2k★**), `janekbaraniewski/openusage` (**~206★**, **+3★** — still deps-only; product quiet continues into **day 10**), `junhoyeo/tokscale` (Wed Antigravity IDE extension sessions + `import --submit` labelled backfill; still **v4.17.0**; **~5.5k★**), `robinebers/openusage` (still **v0.7.13-beta.2**; **~4.3k★**), `seakee/CPA-Manager-Plus` (**v1.13.2→v1.14.0 today** — Usage Maintenance workspace: archive/validate/cleanup/recoverable import-export + offline compress), `semantic-craft/iOS-vibebuddy` (**v1.3.32 today** — status-line model-switch hold fix / AI-10), `ClaudeCodeUsage/ClaudeCodeUsage` (today contributor-credits docs; **~66★**, **+2★**), `alibaba/loongsuite-pilot` (today IP/card masking align; **~194★**, **+3★**). otelite still **v0.1.153**; Helicone quiet since **16 Sep** (**8th day**).

| Repo | Stars | License | Why it matters |
|---|---:|---|---|
| [ccusage/ccusage](https://github.com/ccusage/ccusage) | ~18.7k | NOASSERTION | Still published **v20.0.24** (tag **v20.0.25** no Release) + live pricing snapshots — stay format-compatible (**~18,727★**). |
| [getagentseal/codeburn](https://github.com/getagentseal/codeburn) | ~11.2k | MIT | Still **0.9.25** — **today** Opus 5.5 pricing + low cache-hit session flag. |
| [robinebers/openusage](https://github.com/robinebers/openusage) | ~4.3k | MIT | openusage.ai — still **v0.7.13-beta.2** Rate Limit Resets + Cursor Grok 4.7 + OpenCode 2 OAuth. |
| [janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage) | ~206 | MIT | openusage.sh local multi-tool TUI + SQLite — nearest product twin (**+3★**; product quiet day 10). |
| [ItsJazii/pane](https://github.com/ItsJazii/pane) | ~55 | NOASSERTION | **v0.4.54** — banked limit-reset redeem + weekly capacity — Windows tray OpenUsage port (**+2★**). |
| [Maciek-roboblog/Claude-Code-Usage-Monitor](https://github.com/Maciek-roboblog/Claude-Code-Usage-Monitor) | ~8.7k | MIT | Real-time Claude Code usage monitor with predictions/warnings. |
| [HermannBjorgvin/Clawdmeter](https://github.com/HermannBjorgvin/Clawdmeter) | ~2.2k | (check) | ESP32 desk dashboard — C6 AMOLED IMU auto-rotation still the recent hardware win. |
| [Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail) | ~222 | MIT | **v3.10.1 today** — GPT-6 Sol/Luna pricing on peak/off-peak + Opus 5.5 + Grok 4.7. |
| [juliantanx/aiusage](https://github.com/juliantanx/aiusage) | ~132 | MIT | **v1.5.19 today** — Antigravity model attribution + cloud sync (**+1★**). |
| [junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) | ~5.5k | MIT | Wed Antigravity IDE sessions + `import --submit` backfill — still **v4.17.0** (**~5,526★**). |
| [Iamshankhadeep/ccseva](https://github.com/Iamshankhadeep/ccseva) | ~807 | MIT | macOS menu bar for live Claude Code usage. |
| [ColeMurray/claude-code-otel](https://github.com/ColeMurray/claude-code-otel) | ~504 | MIT | Observability stack for Claude Code usage/perf/cost (OTel path). |
| [CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) | ~517 | MIT | **v2.15.14 today** — Antigravity OAuth refresh + poller/window/builder fixes (Grok Build still in v2.14.55 lineage). |
| [leeguooooo/claude-code-usage-bar](https://github.com/leeguooooo/claude-code-usage-bar) | ~377 | MIT | Predict profiles keyed by session config dir — statusLine quota UX. |
| [foyzulkarim/claude-lens](https://github.com/foyzulkarim/claude-lens) | ~249 | MIT | Local dashboard: sessions, token costs, cache, tool calls, daily breakdowns (**+1★**). |
| [frankchiu-dev/claude-codex-usage-dashboard](https://github.com/frankchiu-dev/claude-codex-usage-dashboard) | ~173 | MIT | Local Windows dashboard for Claude Code + Codex limits. |
| [planetf1/otelite](https://github.com/planetf1/otelite) | ~94 | NOASSERTION | Still **v0.1.153** — storage ANALYZE + bounded ingest. |
| [alibaba/loongsuite-pilot](https://github.com/alibaba/loongsuite-pilot) | ~194 | Apache-2.0 | **Today** masking align — qwen-code / Codex/hooks path (**+3★**). |
| [seakee/CPA-Manager-Plus](https://github.com/seakee/CPA-Manager-Plus) | ~3.6k | MIT | **v1.14.0 today** — Usage Maintenance workspace (archive/validate/cleanup/import-export) — gateway-ops adjacency. |
| [Rodiun/frugon](https://github.com/Rodiun/frugon) | ~212 | MIT | Local LLM bill-leak analyzer. |
| [mag123c/toktrack](https://github.com/mag123c/toktrack) | ~192 | MIT | Still **v2.17.4** — OpenCode v2 compaction/usage + stream SQLite (**+1★**). |
| [Halloweedev/usagepal](https://github.com/Halloweedev/usagepal) | ~82 | MIT | OpenUsage menu-bar fork — per-provider account switcher. |
| [Nanako0129/TokenBar](https://github.com/Nanako0129/TokenBar) | ~367 | MIT | **v1.20.2 today** — Antigravity signed-out sign-in spam fix — macOS quota peer (**+3★**). |
| [tddworks/ClaudeBar](https://github.com/tddworks/ClaudeBar) | ~1.5k | (check) | Quiet after Tue status colors / High Contrast + current Claude generation pricing. |
| [deviffyy/OpenQuota](https://github.com/deviffyy/OpenQuota) | ~219 | MIT | Cross-platform limits/spend tracker (**+3★**). |
| [Dicklesworthstone/coding_agent_usage_tracker](https://github.com/Dicklesworthstone/coding_agent_usage_tracker) | ~87 | NOASSERTION | Single CLI for remaining quotas across Codex/Claude/Gemini/Cursor/Copilot. |
| [sculptdotfun/viberank](https://github.com/sculptdotfun/viberank) | ~117 | MIT | Public AI-coding usage leaderboard on ccusage data. |
| [cobra91/better-ccusage](https://github.com/cobra91/better-ccusage) | ~87 | MIT | Faster multi-provider JSONL analyzer. |
| [851-labs/tokenmaxxing](https://github.com/851-labs/tokenmaxxing) | ~76 | MIT | Local CLI on ccusage that syncs usage socially. |
| [Nihondo/AgentLimits](https://github.com/Nihondo/AgentLimits) | ~62 | MIT | macOS widgets for Codex/Claude limits + ccusage heatmap. |
| [majiayu000/quotabar](https://github.com/majiayu000/quotabar) | ~53 | MIT | Tauri v2 menubar — six-digit tray cost tiles. |
| [abhiunix/AgentHarbor](https://github.com/abhiunix/AgentHarbor) | ~12 | MIT | Native multi-agent rate-limit + session usage app. |
| [semantic-craft/iOS-vibebuddy](https://github.com/semantic-craft/iOS-vibebuddy) | ~90 | MIT | **v1.3.32 today** — status-line model-switch hold / AI-10 — multi-surface tracker. |
| [zhnd/lumo](https://github.com/zhnd/lumo) | ~147 | MIT | Local-first Claude Code usage/cost/session dashboard. |
| [fschmutz/claude-usage-panel](https://github.com/fschmutz/claude-usage-panel) | ~6 | MIT | Official usage API → GNOME/macOS/statusLine + MCP get_usage (Claude + Cursor). |
| [sylearn/AIUsage](https://github.com/sylearn/AIUsage) | ~683 | Apache-2.0 | Desktop multi-tool usage peer (**+2★**). |
| [BerriAI/litellm](https://github.com/BerriAI/litellm) | ~59.5k | NOASSERTION | Gateway spend/pricing — adjacent (**v1.102.1**). |
| [langfuse/langfuse](https://github.com/langfuse/langfuse) | ~35.0k | NOASSERTION | **v4.44.0 today** — OSS LLM observability + OTel — adjacent. |
| [Helicone/helicone](https://github.com/Helicone/helicone) | ~6.2k | Apache-2.0 | Quiet since 16 Sep (**8th day**). OSS LLM observability proxy. Adjacent. |

**Also watch**
- [ClaudeCodeUsage/ClaudeCodeUsage](https://github.com/ClaudeCodeUsage/ClaudeCodeUsage) (~66★, MIT, **+2★**) — today contributor-credits docs; VS Code status-bar Claude Code usage/cost.
- [itvincent-git/codex-usage-desktop](https://github.com/itvincent-git/codex-usage-desktop) (~43★, MIT, **+2★**) — local-first Codex quota desktop (macOS + Windows).
- [saeedkolivand/claude-usage-mac](https://github.com/saeedkolivand/claude-usage-mac) (~15★) — Claude Code macOS menu bar + desktop widget (pushed today).
- [ofershap/cursor-usage-tracker](https://github.com/ofershap/cursor-usage-tracker) (~33★, MIT) — Cursor Enterprise FinOps, not personal autodetection.
- Micro/zero-★ Claude wrappers keep spawning daily — noise floor, not category peers.

**Build takeaways:** Stay ccusage-compatible (**v20.0.24**; watch tag **v20.0.25**). Differentiate on live quotas + auto-detect + UX vs openusage.sh (still product-quiet day 10). Today's sharpest signals: **CodeZeno v2.15.14 Antigravity OAuth**, **Splitrail v3.10.1 GPT-6 Sol/Luna**, **TokenBar v1.20.2**, **pane v0.4.54**, **aiusage v1.5.19**, **CPA v1.14.0**, **Langfuse v4.44.0**. Keep Langfuse/LiteLLM complementary; Helicone remains maintenance-mode adjacent (day 8).

Sources: GitHub API 2026-09-24 (Europe/London).

### Open source — 25 Sep 2026

Fresh Friday GitHub API pass (Europe/London). Category still owned by **ccusage** + the two OpenUsage brands; **codeburn** remains the loud multi-tool local tracker. **Live today / overnight:** `CodeZeno/Claude-Code-Usage-Monitor` (**v2.15.14→v2.15.15→…→v2.15.18 today** — Direct3D fallback when dashboard OpenGL fails + uniform segments at fractional display scales; tags **v2.15.15–17** exist without separate Release notes; **~524★**, **+7★**), `langfuse/langfuse` (**v4.44.0→v4.45.0→…→v4.45.4 today** — rapid 4.45.x train after Thu billing/design-system cut; eval decision-model dedupe + embedded skills refresh; **~35.0k★**), `getagentseal/codeburn` (still **0.9.25**; **today** `codeburn import cursor <file.csv>` to replace local Cursor estimates with dashboard Export CSV (#1558) + Cursor Agent `~/.cursor/chats` store.db sessions + WSL Claude quota credential + Copilot OTel workspace attribution; **~11.2k★**), `ItsJazii/pane` (still published **v0.4.54**; **Thu night** Claude **Cloud session credits** bar from `iguana_necktie` + **StepFun** wallet/plan-credits card merged, unreleased; **~57★**, **+2★**), `junhoyeo/tokscale` (**today** TUI i18n en/ko/ja/zh-CN/fr + OpenClaw compressed SQLite transcript decode + Codex tier source / Antigravity ledger reconciliation; still **v4.17.0**; **~5.5k★**, **+11★**), `Nanako0129/TokenBar` (still **v1.20.2**; **today** Codex OAuth refresh replacement + Liquid Glass panel + cold-load / quota-window perf + local-network usage description; **~368★**, **+1★**), `ccusage/ccusage` (still published **v20.0.24**; tag **v20.0.25** still no Release; continuous models.dev/LiteLLM pricing snapshots; **~18,735★**), `janekbaraniewski/openusage` (**~209★**, **+3★** — still deps-only; product quiet continues into **day 11**), `Piebald-AI/splitrail` (still **v3.10.1** GPT-6 Sol/Luna; **~222★**), `juliantanx/aiusage` (still **v1.5.19**; **~132★**), `tddworks/ClaudeBar` (**v0.4.93** Thu — docs/design split + troubleshooting; **~1.5k★**), `itvincent-git/codex-usage-desktop` (**app-v3.6.1** Thu — macOS relaunch-after-update + Windows duplicate limit-fetch fix; **~43★**), `robinebers/openusage` (still **v0.7.13-beta.2**; **~4.3k★**), `seakee/CPA-Manager-Plus` (still **v1.14.0**; **~3.6k★**). otelite still **v0.1.153**; Helicone quiet since **16 Sep** (**9th day**).

| Repo | Stars | License | Why it matters |
|---|---:|---|---|
| [ccusage/ccusage](https://github.com/ccusage/ccusage) | ~18.7k | NOASSERTION | Still published **v20.0.24** (tag **v20.0.25** no Release) + live pricing snapshots — stay format-compatible (**~18,735★**). |
| [getagentseal/codeburn](https://github.com/getagentseal/codeburn) | ~11.2k | MIT | Still **0.9.25** — **today** Cursor CSV import (#1558) + store.db Cursor Agent sessions + WSL Claude quota. |
| [robinebers/openusage](https://github.com/robinebers/openusage) | ~4.3k | MIT | openusage.ai — still **v0.7.13-beta.2** Rate Limit Resets + Cursor Grok 4.7 + OpenCode 2 OAuth. |
| [janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage) | ~209 | MIT | openusage.sh local multi-tool TUI + SQLite — nearest product twin (**+3★**; product quiet day 11). |
| [ItsJazii/pane](https://github.com/ItsJazii/pane) | ~57 | NOASSERTION | Still **v0.4.54** — Claude Cloud credits + StepFun merged unreleased — Windows tray OpenUsage port (**+2★**). |
| [Maciek-roboblog/Claude-Code-Usage-Monitor](https://github.com/Maciek-roboblog/Claude-Code-Usage-Monitor) | ~8.7k | MIT | Real-time Claude Code usage monitor with predictions/warnings. |
| [HermannBjorgvin/Clawdmeter](https://github.com/HermannBjorgvin/Clawdmeter) | ~2.2k | (check) | ESP32 desk dashboard — C6 AMOLED IMU auto-rotation still the recent hardware win. |
| [Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail) | ~222 | MIT | Still **v3.10.1** — GPT-6 Sol/Luna pricing on peak/off-peak + Opus 5.5 + Grok 4.7. |
| [juliantanx/aiusage](https://github.com/juliantanx/aiusage) | ~132 | MIT | Still **v1.5.19** — Antigravity model attribution + cloud sync. |
| [junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) | ~5.5k | MIT | **Today** TUI i18n + OpenClaw SQLite decode + Codex/Antigravity ledger — still **v4.17.0** (**~5,537★**). |
| [Iamshankhadeep/ccseva](https://github.com/Iamshankhadeep/ccseva) | ~807 | MIT | macOS menu bar for live Claude Code usage. |
| [ColeMurray/claude-code-otel](https://github.com/ColeMurray/claude-code-otel) | ~504 | MIT | Observability stack for Claude Code usage/perf/cost (OTel path). |
| [CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) | ~524 | MIT | **v2.15.18 today** — Direct3D OpenGL fallback + fractional DPI segments (**+7★**). |
| [leeguooooo/claude-code-usage-bar](https://github.com/leeguooooo/claude-code-usage-bar) | ~377 | MIT | Predict profiles keyed by session config dir — statusLine quota UX. |
| [foyzulkarim/claude-lens](https://github.com/foyzulkarim/claude-lens) | ~249 | MIT | Local dashboard: sessions, token costs, cache, tool calls, daily breakdowns. |
| [frankchiu-dev/claude-codex-usage-dashboard](https://github.com/frankchiu-dev/claude-codex-usage-dashboard) | ~173 | MIT | Local Windows dashboard for Claude Code + Codex limits. |
| [planetf1/otelite](https://github.com/planetf1/otelite) | ~94 | NOASSERTION | Still **v0.1.153** — storage ANALYZE + bounded ingest. |
| [alibaba/loongsuite-pilot](https://github.com/alibaba/loongsuite-pilot) | ~195 | Apache-2.0 | Quiet after Thu masking align — qwen-code / Codex/hooks path (**+1★**). |
| [seakee/CPA-Manager-Plus](https://github.com/seakee/CPA-Manager-Plus) | ~3.6k | MIT | Still **v1.14.0** — Usage Maintenance workspace — gateway-ops adjacency. |
| [Rodiun/frugon](https://github.com/Rodiun/frugon) | ~212 | MIT | Local LLM bill-leak analyzer. |
| [mag123c/toktrack](https://github.com/mag123c/toktrack) | ~192 | MIT | Still **v2.17.4** — OpenCode v2 compaction/usage + stream SQLite. |
| [Halloweedev/usagepal](https://github.com/Halloweedev/usagepal) | ~82 | MIT | OpenUsage menu-bar fork — per-provider account switcher. |
| [Nanako0129/TokenBar](https://github.com/Nanako0129/TokenBar) | ~368 | MIT | Still **v1.20.2** — today Codex OAuth refresh + Liquid Glass + cold-load perf (**+1★**). |
| [tddworks/ClaudeBar](https://github.com/tddworks/ClaudeBar) | ~1.5k | (check) | **v0.4.93** Thu — docs/design split + troubleshooting. |
| [deviffyy/OpenQuota](https://github.com/deviffyy/OpenQuota) | ~219 | MIT | Cross-platform limits/spend tracker. |
| [Dicklesworthstone/coding_agent_usage_tracker](https://github.com/Dicklesworthstone/coding_agent_usage_tracker) | ~87 | NOASSERTION | Single CLI for remaining quotas across Codex/Claude/Gemini/Cursor/Copilot. |
| [sculptdotfun/viberank](https://github.com/sculptdotfun/viberank) | ~117 | MIT | Public AI-coding usage leaderboard on ccusage data. |
| [cobra91/better-ccusage](https://github.com/cobra91/better-ccusage) | ~87 | MIT | Faster multi-provider JSONL analyzer. |
| [851-labs/tokenmaxxing](https://github.com/851-labs/tokenmaxxing) | ~76 | MIT | Local CLI on ccusage that syncs usage socially. |
| [Nihondo/AgentLimits](https://github.com/Nihondo/AgentLimits) | ~62 | MIT | macOS widgets for Codex/Claude limits + ccusage heatmap. |
| [majiayu000/quotabar](https://github.com/majiayu000/quotabar) | ~53 | MIT | Tauri v2 menubar — six-digit tray cost tiles. |
| [abhiunix/AgentHarbor](https://github.com/abhiunix/AgentHarbor) | ~12 | MIT | Native multi-agent rate-limit + session usage app. |
| [semantic-craft/iOS-vibebuddy](https://github.com/semantic-craft/iOS-vibebuddy) | ~90 | MIT | Still **v1.3.32** — today Watch WR-12 active-refresh / quiet-context polish. |
| [zhnd/lumo](https://github.com/zhnd/lumo) | ~147 | MIT | Local-first Claude Code usage/cost/session dashboard. |
| [fschmutz/claude-usage-panel](https://github.com/fschmutz/claude-usage-panel) | ~6 | MIT | Official usage API → GNOME/macOS/statusLine + MCP get_usage (Claude + Cursor). |
| [sylearn/AIUsage](https://github.com/sylearn/AIUsage) | ~683 | Apache-2.0 | Desktop multi-tool usage peer (still **v0.15.21**). |
| [BerriAI/litellm](https://github.com/BerriAI/litellm) | ~59.6k | NOASSERTION | Gateway spend/pricing — adjacent (**v1.102.1**; today Langfuse v4 callback migrate + cost-map). |
| [langfuse/langfuse](https://github.com/langfuse/langfuse) | ~35.0k | NOASSERTION | **v4.45.4 today** — OSS LLM observability + OTel — adjacent (4.45.x train). |
| [Helicone/helicone](https://github.com/Helicone/helicone) | ~6.2k | Apache-2.0 | Quiet since 16 Sep (**9th day**). OSS LLM observability proxy. Adjacent. |

**Also watch**
- [ClaudeCodeUsage/ClaudeCodeUsage](https://github.com/ClaudeCodeUsage/ClaudeCodeUsage) (~66★, MIT) — quiet after Thu contributor-credits docs; VS Code status-bar Claude Code usage/cost.
- [itvincent-git/codex-usage-desktop](https://github.com/itvincent-git/codex-usage-desktop) (~43★, MIT) — **app-v3.6.1** macOS relaunch + Windows dup-fetch fix.
- [saeedkolivand/claude-usage-mac](https://github.com/saeedkolivand/claude-usage-mac) (~15★) — Claude Code macOS menu bar + desktop widget (stats bump).
- [ofershap/cursor-usage-tracker](https://github.com/ofershap/cursor-usage-tracker) (~33★, MIT) — Cursor Enterprise FinOps, not personal autodetection.
- [DeepAgentLabs/agenticlens](https://github.com/DeepAgentLabs/agenticlens) (~59★, MIT) — agent token/cost/latency profiler (quiet since Thu).
- Micro/zero-★ Claude wrappers keep spawning daily — noise floor, not category peers.

**Build takeaways:** Stay ccusage-compatible (**v20.0.24**; watch tag **v20.0.25**). Differentiate on live quotas + auto-detect + UX vs openusage.sh (still product-quiet day 11). Today's sharpest signals: **CodeZeno v2.15.18 DPI/Direct3D**, **Langfuse v4.45.4**, **codeburn Cursor CSV import**, **pane Claude Cloud credits + StepFun (unreleased)**, **tokscale TUI i18n / OpenClaw**. Keep Langfuse/LiteLLM complementary; Helicone remains maintenance-mode adjacent (day 9).

Sources: GitHub API 2026-09-25 (Europe/London).

### Open source — 26 Sep 2026

Fresh Saturday GitHub API pass (Europe/London). Category still owned by **ccusage** + the two OpenUsage brands; **codeburn** remains the loud multi-tool local tracker. **Live today / overnight:** `Nanako0129/syrtis` (**TokenBar → Syrtis v2.0.0 today** — first launch renames `TokenBar.app`→`Syrtis.app` in place; Homebrew tap `nanako0129/tap/syrtis`; Liquid Glass panel on macOS 27; Pi fork-session double-count fix + OpenClaw `.jsonl.zst` archive read / compaction de-count; Codex refresh by token expiry; cold-load + quota-window perf; **~369★**, **+1★**), `ItsJazii/pane` (**v0.4.54→v0.4.55** — Claude **Cloud session credits** with expiry + **StepFun** balance/plan-credits/spend from Claude Code/Codex/OpenCode/oh-my-pi/Step Code + updater proxy/retry; **~58★**, **+1★**), `CodeZeno/Claude-Code-Usage-Monitor` (**v2.15.18→v2.15.19** — closes Friday's Direct3D + fractional-DPI train; **~539★**, **+15★** since Fri open), `langfuse/langfuse` (**v4.45.4→v4.46.0** — basic skill management + experiment side-by-side + formatted JSON on dataset items + 100-row tracing page size; **~35.1k★**), `seakee/CPA-Manager-Plus` (**v1.14.0→v1.14.1** — reverse-proxy HEAD→200 falsely marking Usage Maintenance unsupported + update-index fetch retry; **~3.6k★**), `semantic-craft/iOS-vibebuddy` (**v1.3.35** macOS — iCloud private-DB push when no APNs key + Doubao/Qwen read-aloud personas; Recap removed; **~90★**), `getagentseal/codeburn` (still **0.9.25**; Fri night mouse-tracking off by default so text selection works, `m` toggles; Cursor CSV import still the Fri headline; **~11.2k★**), `tddworks/ClaudeBar` (still **v0.4.93**; **today** Touch Bar gauges colour by quota status; **~1.5k★**), `ccusage/ccusage` (still published **v20.0.24**; tag **v20.0.25** still no Release; continuous models.dev/LiteLLM pricing snapshots; **~18,744★**), `janekbaraniewski/openusage` (**~211★**, **+2★** — still deps-only; product quiet continues into **day 12**), `robinebers/openusage` (still **v0.7.13-beta.2**; **~4.3k★**), `junhoyeo/tokscale` (still **v4.17.0** after Fri TUI i18n / OpenClaw decode; **~5.5k★**), `Piebald-AI/splitrail` (still **v3.10.1**; **~222★**), `juliantanx/aiusage` (still **v1.5.19**; **~133★**), otelite still **v0.1.153**. Helicone quiet since **16 Sep** (**10th day**).

| Repo | Stars | License | Why it matters |
|---|---:|---|---|
| [ccusage/ccusage](https://github.com/ccusage/ccusage) | ~18.7k | NOASSERTION | Still published **v20.0.24** (tag **v20.0.25** no Release) + live pricing snapshots — stay format-compatible (**~18,744★**). |
| [getagentseal/codeburn](https://github.com/getagentseal/codeburn) | ~11.2k | MIT | Still **0.9.25** — Fri Cursor CSV import (#1558) + store.db Agent sessions; Fri night mouse-track off-by-default. |
| [robinebers/openusage](https://github.com/robinebers/openusage) | ~4.3k | MIT | openusage.ai — still **v0.7.13-beta.2** Rate Limit Resets + Cursor Grok 4.7 + OpenCode 2 OAuth. |
| [janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage) | ~211 | MIT | openusage.sh local multi-tool TUI + SQLite — nearest product twin (**+2★**; product quiet day 12; tag **v0.25.0**). |
| [ItsJazii/pane](https://github.com/ItsJazii/pane) | ~58 | NOASSERTION | **v0.4.55** — Claude Cloud credits + StepFun wallet/plan — Windows tray OpenUsage port (**+1★**). |
| [Maciek-roboblog/Claude-Code-Usage-Monitor](https://github.com/Maciek-roboblog/Claude-Code-Usage-Monitor) | ~8.7k | MIT | Real-time Claude Code usage monitor with predictions/warnings. |
| [HermannBjorgvin/Clawdmeter](https://github.com/HermannBjorgvin/Clawdmeter) | ~2.2k | (check) | ESP32 desk dashboard — C6 AMOLED IMU auto-rotation still the recent hardware win. |
| [Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail) | ~222 | MIT | Still **v3.10.1** — GPT-6 Sol/Luna pricing on peak/off-peak + Opus 5.5 + Grok 4.7. |
| [juliantanx/aiusage](https://github.com/juliantanx/aiusage) | ~133 | MIT | Still **v1.5.19** — Antigravity model attribution + cloud sync. |
| [junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) | ~5.5k | MIT | Still **v4.17.0** after Fri TUI i18n + OpenClaw SQLite decode + Codex/Antigravity ledger (**~5,548★**). |
| [Iamshankhadeep/ccseva](https://github.com/Iamshankhadeep/ccseva) | ~807 | MIT | macOS menu bar for live Claude Code usage. |
| [ColeMurray/claude-code-otel](https://github.com/ColeMurray/claude-code-otel) | ~506 | MIT | Observability stack for Claude Code usage/perf/cost (OTel path). |
| [CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) | ~539 | MIT | **v2.15.19** — closes Fri Direct3D OpenGL fallback + fractional DPI segments (**+15★**). |
| [leeguooooo/claude-code-usage-bar](https://github.com/leeguooooo/claude-code-usage-bar) | ~376 | MIT | Predict profiles keyed by session config dir — statusLine quota UX. |
| [foyzulkarim/claude-lens](https://github.com/foyzulkarim/claude-lens) | ~249 | MIT | Local dashboard: sessions, token costs, cache, tool calls, daily breakdowns. |
| [frankchiu-dev/claude-codex-usage-dashboard](https://github.com/frankchiu-dev/claude-codex-usage-dashboard) | ~173 | MIT | Local Windows dashboard for Claude Code + Codex limits. |
| [planetf1/otelite](https://github.com/planetf1/otelite) | ~94 | NOASSERTION | Still **v0.1.153** — storage ANALYZE + bounded ingest. |
| [alibaba/loongsuite-pilot](https://github.com/alibaba/loongsuite-pilot) | ~195 | Apache-2.0 | Quiet after Thu masking align — qwen-code / Codex/hooks path. |
| [seakee/CPA-Manager-Plus](https://github.com/seakee/CPA-Manager-Plus) | ~3.6k | MIT | **v1.14.1 today** — Usage Maintenance reverse-proxy HEAD probe fix. |
| [Rodiun/frugon](https://github.com/Rodiun/frugon) | ~213 | MIT | Local LLM bill-leak analyzer. |
| [mag123c/toktrack](https://github.com/mag123c/toktrack) | ~192 | MIT | Still **v2.17.4** — OpenCode v2 compaction/usage + stream SQLite. |
| [Halloweedev/usagepal](https://github.com/Halloweedev/usagepal) | ~82 | MIT | OpenUsage menu-bar fork — per-provider account switcher. |
| [Nanako0129/syrtis](https://github.com/Nanako0129/syrtis) | ~369 | MIT | **Was TokenBar** — **Syrtis v2.0.0 today** rebrand + Liquid Glass + Pi/OpenClaw count fixes (**+1★**). |
| [tddworks/ClaudeBar](https://github.com/tddworks/ClaudeBar) | ~1.5k | (check) | Still **v0.4.93** — **today** Touch Bar gauges colour by quota status. |
| [deviffyy/OpenQuota](https://github.com/deviffyy/OpenQuota) | ~221 | MIT | Cross-platform limits/spend tracker. |
| [Dicklesworthstone/coding_agent_usage_tracker](https://github.com/Dicklesworthstone/coding_agent_usage_tracker) | ~87 | NOASSERTION | Single CLI for remaining quotas across Codex/Claude/Gemini/Cursor/Copilot. |
| [sculptdotfun/viberank](https://github.com/sculptdotfun/viberank) | ~117 | MIT | Public AI-coding usage leaderboard on ccusage data. |
| [cobra91/better-ccusage](https://github.com/cobra91/better-ccusage) | ~89 | MIT | Faster multi-provider JSONL analyzer. |
| [851-labs/tokenmaxxing](https://github.com/851-labs/tokenmaxxing) | ~76 | MIT | Local CLI on ccusage that syncs usage socially. |
| [Nihondo/AgentLimits](https://github.com/Nihondo/AgentLimits) | ~62 | MIT | macOS widgets for Codex/Claude limits + ccusage heatmap. |
| [majiayu000/quotabar](https://github.com/majiayu000/quotabar) | ~53 | MIT | Tauri v2 menubar — six-digit tray cost tiles. |
| [abhiunix/AgentHarbor](https://github.com/abhiunix/AgentHarbor) | ~12 | MIT | Native multi-agent rate-limit + session usage app. |
| [semantic-craft/iOS-vibebuddy](https://github.com/semantic-craft/iOS-vibebuddy) | ~90 | MIT | **v1.3.35** — iCloud push without APNs key + read-aloud personas; Recap removed. |
| [zhnd/lumo](https://github.com/zhnd/lumo) | ~147 | MIT | Local-first Claude Code usage/cost/session dashboard. |
| [fschmutz/claude-usage-panel](https://github.com/fschmutz/claude-usage-panel) | ~6 | MIT | Official usage API → GNOME/macOS/statusLine + MCP get_usage (Claude + Cursor). |
| [sylearn/AIUsage](https://github.com/sylearn/AIUsage) | ~684 | Apache-2.0 | Desktop multi-tool usage peer (still **v0.15.21**). |
| [BerriAI/litellm](https://github.com/BerriAI/litellm) | ~59.6k | NOASSERTION | Gateway spend/pricing — adjacent (**v1.102.1**; today team-member budget email alerts + cost-map). |
| [langfuse/langfuse](https://github.com/langfuse/langfuse) | ~35.1k | NOASSERTION | **v4.46.0** — OSS LLM observability + OTel — adjacent (skills + experiment compare). |
| [Helicone/helicone](https://github.com/Helicone/helicone) | ~6.2k | Apache-2.0 | Quiet since 16 Sep (**10th day**). OSS LLM observability proxy. Adjacent. |

**Also watch**
- [ClaudeCodeUsage/ClaudeCodeUsage](https://github.com/ClaudeCodeUsage/ClaudeCodeUsage) (~66★, MIT) — quiet after Thu contributor-credits docs; VS Code status-bar Claude Code usage/cost.
- [itvincent-git/codex-usage-desktop](https://github.com/itvincent-git/codex-usage-desktop) (~43★, MIT) — still **app-v3.6.1** macOS relaunch + Windows dup-fetch fix.
- [saeedkolivand/claude-usage-mac](https://github.com/saeedkolivand/claude-usage-mac) (~15★) — Claude Code macOS menu bar + desktop widget.
- [ofershap/cursor-usage-tracker](https://github.com/ofershap/cursor-usage-tracker) (~33★, MIT) — Cursor Enterprise FinOps, not personal autodetection.
- [DeepAgentLabs/agenticlens](https://github.com/DeepAgentLabs/agenticlens) (~59★, MIT) — agent token/cost/latency profiler (quiet since Thu).
- Micro/zero-★ Claude wrappers keep spawning daily — noise floor, not category peers.

**Build takeaways:** Stay ccusage-compatible (**v20.0.24**; watch tag **v20.0.25**). Differentiate on live quotas + auto-detect + UX vs openusage.sh (still product-quiet day 12). Today's sharpest signals: **TokenBar→Syrtis v2.0.0 rebrand**, **pane v0.4.55 Claude Cloud credits + StepFun**, **CodeZeno v2.15.19**, **Langfuse v4.46.0 skills**, **CPA-Manager-Plus v1.14.1**. Keep Langfuse/LiteLLM complementary; Helicone remains maintenance-mode adjacent (day 10).

Sources: GitHub API 2026-09-26 (Europe/London).

### Open source — 27 Sep 2026

Fresh Sunday GitHub API pass (Europe/London). Category still owned by **ccusage** + the two OpenUsage brands; **codeburn** remains the loud multi-tool local tracker. **Live today / overnight:** `Nanako0129/syrtis` (**v2.0.0→v2.0.1** Sat night — frosted-glass tooltips over history rows on macOS 27; **today** unreleased train — quota persist/load repair + sample-drop gate, Grok Build model-alias grouping, grokbot JWT history-owner decode, quota-gauge stale marker, demo quota fixtures; **~375★**, **+6★**), `semantic-craft/iOS-vibebuddy` (**v1.3.35→v1.3.37** today macOS — Claude allowance via ~15-min background `claude` CLI Haiku probe when status-line is absent (desktop/IDE sessions); spend_limit >100% shown as exhausted; **~90★**), `majiayu000/quotabar` (**v0.5.4** Sat — published **ccstats 0.9.0** SDK + shared agent-sessions parser from crates.io; notarized DMG staple; **~53★**), `Halloweedev/usagepal` (**v0.7.75→v0.7.77-beta.2** Sat — Budgets tab daily/per-day plan budgets + OpenCode Go V2 `session_message` cost rows; **~82★**), `851-labs/tokenmaxxing` (Sat **cli v0.7.0-alpha.1** — Windows bin launcher + silent sync task; **~76★**), `ccusage/ccusage` (still published **v20.0.24**; tag **v20.0.25** still no Release; continuous models.dev/LiteLLM pricing snapshots today; **~18,761★**), `janekbaraniewski/openusage` (**~212★**, **+1★** — still deps-only; product quiet continues into **day 13**), `ItsJazii/pane` (still **v0.4.55**; **~60★**, **+2★**), `CodeZeno/Claude-Code-Usage-Monitor` (still **v2.15.19**; **~542★**, **+3★**), `getagentseal/codeburn` (still **0.9.25**; **~11.3k★**), `robinebers/openusage` (latest Release still **v0.7.12**; repo notes **v0.7.13-beta.2**; **~4.3k★**), `junhoyeo/tokscale` (still **v4.17.0**; **~5.6k★**), `Piebald-AI/splitrail` (still **v3.10.1**; **~222★**), `langfuse/langfuse` (still **v4.46.0**; Sat dashboard multipart large query params; **~35.1k★**), `BerriAI/litellm` (still **v1.102.1**; today Gemini tools token_counter + open cost-discount/cache-aware routing PRs; **~59.7k★**). Helicone quiet since **16 Sep** (**11th day**).

| Repo | Stars | License | Why it matters |
|---|---:|---|---|
| [ccusage/ccusage](https://github.com/ccusage/ccusage) | ~18.8k | NOASSERTION | Still published **v20.0.24** (tag **v20.0.25** no Release) + live pricing snapshots — stay format-compatible (**~18,761★**). |
| [getagentseal/codeburn](https://github.com/getagentseal/codeburn) | ~11.3k | MIT | Still **0.9.25** — Fri Cursor CSV import (#1558) + store.db Agent sessions; mouse-track off-by-default. |
| [robinebers/openusage](https://github.com/robinebers/openusage) | ~4.3k | MIT | openusage.ai — Release **v0.7.12** / notes **v0.7.13-beta.2** Rate Limit Resets + Cursor Grok 4.7 + OpenCode 2 OAuth. |
| [janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage) | ~212 | MIT | openusage.sh local multi-tool TUI + SQLite — nearest product twin (**+1★**; product quiet day 13; tag **v0.25.0**). |
| [ItsJazii/pane](https://github.com/ItsJazii/pane) | ~60 | NOASSERTION | Still **v0.4.55** — Claude Cloud credits + StepFun wallet/plan — Windows tray OpenUsage port (**+2★**). |
| [Maciek-roboblog/Claude-Code-Usage-Monitor](https://github.com/Maciek-roboblog/Claude-Code-Usage-Monitor) | ~8.7k | MIT | Real-time Claude Code usage monitor with predictions/warnings. |
| [HermannBjorgvin/Clawdmeter](https://github.com/HermannBjorgvin/Clawdmeter) | ~2.2k | (check) | ESP32 desk dashboard — C6 AMOLED IMU auto-rotation still the recent hardware win. |
| [Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail) | ~222 | MIT | Still **v3.10.1** — GPT-6 Sol/Luna pricing on peak/off-peak + Opus 5.5 + Grok 4.7. |
| [juliantanx/aiusage](https://github.com/juliantanx/aiusage) | ~133 | MIT | Still **v1.5.19** — Antigravity model attribution + cloud sync. |
| [junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) | ~5.6k | MIT | Still **v4.17.0** after Fri TUI i18n + OpenClaw SQLite decode + Codex/Antigravity ledger. |
| [Iamshankhadeep/ccseva](https://github.com/Iamshankhadeep/ccseva) | ~807 | MIT | macOS menu bar for live Claude Code usage. |
| [ColeMurray/claude-code-otel](https://github.com/ColeMurray/claude-code-otel) | ~506 | MIT | Observability stack for Claude Code usage/perf/cost (OTel path). |
| [CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) | ~542 | MIT | Still **v2.15.19** — Direct3D OpenGL fallback + fractional DPI segments (**+3★**). |
| [leeguooooo/claude-code-usage-bar](https://github.com/leeguooooo/claude-code-usage-bar) | ~376 | MIT | Predict profiles keyed by session config dir — statusLine quota UX. |
| [foyzulkarim/claude-lens](https://github.com/foyzulkarim/claude-lens) | ~249 | MIT | Local dashboard: sessions, token costs, cache, tool calls, daily breakdowns. |
| [frankchiu-dev/claude-codex-usage-dashboard](https://github.com/frankchiu-dev/claude-codex-usage-dashboard) | ~173 | MIT | Local Windows dashboard for Claude Code + Codex limits. |
| [planetf1/otelite](https://github.com/planetf1/otelite) | ~94 | NOASSERTION | Still **v0.1.153** — storage ANALYZE + bounded ingest. |
| [alibaba/loongsuite-pilot](https://github.com/alibaba/loongsuite-pilot) | ~195 | Apache-2.0 | Quiet after Fri masking align — qwen-code / Codex/hooks path. |
| [seakee/CPA-Manager-Plus](https://github.com/seakee/CPA-Manager-Plus) | ~3.6k | MIT | Still **v1.14.1** — Usage Maintenance reverse-proxy HEAD probe fix. |
| [Rodiun/frugon](https://github.com/Rodiun/frugon) | ~213 | MIT | Local LLM bill-leak analyzer. |
| [mag123c/toktrack](https://github.com/mag123c/toktrack) | ~192 | MIT | Still **v2.17.4** — OpenCode v2 compaction/usage + stream SQLite. |
| [Halloweedev/usagepal](https://github.com/Halloweedev/usagepal) | ~82 | MIT | **v0.7.77-beta.2** — Budgets tab + OpenCode Go V2 costs — OpenUsage menu-bar fork. |
| [Nanako0129/syrtis](https://github.com/Nanako0129/syrtis) | ~375 | MIT | **Was TokenBar** — **v2.0.1** glass tooltips + **today** quota/Grok-alias unreleased train (**+6★**). |
| [tddworks/ClaudeBar](https://github.com/tddworks/ClaudeBar) | ~1.5k | (check) | Still **v0.4.93** — Touch Bar gauges colour by quota status. |
| [deviffyy/OpenQuota](https://github.com/deviffyy/OpenQuota) | ~221 | MIT | Cross-platform limits/spend tracker (quiet Sat). |
| [Dicklesworthstone/coding_agent_usage_tracker](https://github.com/Dicklesworthstone/coding_agent_usage_tracker) | ~87 | NOASSERTION | Single CLI for remaining quotas across Codex/Claude/Gemini/Cursor/Copilot. |
| [sculptdotfun/viberank](https://github.com/sculptdotfun/viberank) | ~117 | MIT | Public AI-coding usage leaderboard on ccusage data. |
| [cobra91/better-ccusage](https://github.com/cobra91/better-ccusage) | ~89 | MIT | Faster multi-provider JSONL analyzer. |
| [851-labs/tokenmaxxing](https://github.com/851-labs/tokenmaxxing) | ~76 | MIT | Sat **cli v0.7.0-alpha.1** — local CLI on ccusage that syncs usage socially. |
| [Nihondo/AgentLimits](https://github.com/Nihondo/AgentLimits) | ~62 | MIT | macOS widgets for Codex/Claude limits + ccusage heatmap. |
| [majiayu000/quotabar](https://github.com/majiayu000/quotabar) | ~53 | MIT | **v0.5.4** — ccstats 0.9.0 crates.io SDK + notarized DMG — Tauri v2 menubar. |
| [abhiunix/AgentHarbor](https://github.com/abhiunix/AgentHarbor) | ~12 | MIT | Native multi-agent rate-limit + session usage app (traffic snapshots only). |
| [semantic-craft/iOS-vibebuddy](https://github.com/semantic-craft/iOS-vibebuddy) | ~90 | MIT | **v1.3.37** — Claude allowance via background CLI probe when status-line absent. |
| [zhnd/lumo](https://github.com/zhnd/lumo) | ~147 | MIT | Local-first Claude Code usage/cost/session dashboard. |
| [fschmutz/claude-usage-panel](https://github.com/fschmutz/claude-usage-panel) | ~6 | MIT | Official usage API → GNOME/macOS/statusLine + MCP get_usage (Claude + Cursor). |
| [sylearn/AIUsage](https://github.com/sylearn/AIUsage) | ~687 | Apache-2.0 | Desktop multi-tool usage peer (still **v0.15.21**). |
| [BerriAI/litellm](https://github.com/BerriAI/litellm) | ~59.7k | NOASSERTION | Gateway spend/pricing — adjacent (**v1.102.1**; today Gemini tools counter + open cost-discount/cache-routing PRs). |
| [langfuse/langfuse](https://github.com/langfuse/langfuse) | ~35.1k | NOASSERTION | Still **v4.46.0** — OSS LLM observability + OTel — adjacent (Sat multipart dashboard queries). |
| [Helicone/helicone](https://github.com/Helicone/helicone) | ~6.2k | Apache-2.0 | Quiet since 16 Sep (**11th day**). OSS LLM observability proxy. Adjacent. |

**Also watch**
- [ClaudeCodeUsage/ClaudeCodeUsage](https://github.com/ClaudeCodeUsage/ClaudeCodeUsage) (~66★, MIT) — quiet after Thu contributor-credits docs; VS Code status-bar Claude Code usage/cost.
- [itvincent-git/codex-usage-desktop](https://github.com/itvincent-git/codex-usage-desktop) (~43★, MIT) — still **app-v3.6.1** macOS relaunch + Windows dup-fetch fix.
- [saeedkolivand/claude-usage-mac](https://github.com/saeedkolivand/claude-usage-mac) (~15★) — Claude Code macOS menu bar + desktop widget.
- [ofershap/cursor-usage-tracker](https://github.com/ofershap/cursor-usage-tracker) (~33★, MIT) — Cursor Enterprise FinOps, not personal autodetection.
- [DeepAgentLabs/agenticlens](https://github.com/DeepAgentLabs/agenticlens) (~59★, MIT) — agent token/cost/latency profiler (quiet).
- Micro/zero-★ Claude wrappers keep spawning daily — noise floor, not category peers.

**Build takeaways:** Stay ccusage-compatible (**v20.0.24**; watch tag **v20.0.25**). Differentiate on live quotas + auto-detect + UX vs openusage.sh (still product-quiet day 13). Today's sharpest signals: **Syrtis v2.0.1 + Sunday quota/Grok-alias train**, **VibeBuddy v1.3.37 Claude CLI allowance probe**, **QuotaBar v0.5.4 ccstats SDK**, **usagepal Budgets tab**. Keep Langfuse/LiteLLM complementary; Helicone remains maintenance-mode adjacent (day 11).

Sources: GitHub API 2026-09-27 (Europe/London).

### Open source — 28 Sep 2026

Fresh Monday GitHub API pass (Europe/London). Category still owned by **ccusage** + the two OpenUsage brands; **codeburn** remains the loud multi-tool local tracker. **Live today / overnight:** `ccusage/ccusage` (**v20.0.24→v20.0.26** Sun — npm published **20.0.26** (tag **v20.0.25** still never got a Release); Codex session-id filter; Claude requestless dedupe by timestamp; OpenCode fork-history exclude; stale-file skip + pricing spelling index; today continuous models.dev/LiteLLM pricing snapshots; **~18,772★**, **+11★**), `Nanako0129/syrtis` (**v2.0.1→v2.1.0** Sun evening — attribution onboarding cards, Grok Build model-alias unify, stale quota-gauge grey, quota persist/load repair; **today** unreleased train — Developer ID signing/notarize release job, first-run Overview setup cards, tray sand-shoal animation, per-lens popover scroll + Settings copy rewrite; **~377★**, **+2★**), `ItsJazii/pane` (still Release **v0.4.55**; **today** merged unreleased **widget mode** — always-on desktop dashboard + collapse/lock; **~62★**, **+2★**), `851-labs/tokenmaxxing` (still **cli v0.7.0-alpha.1**; **today** unauth CLI login IP rate-limit + Windows service/shim e2e; **~77★**, **+1★**), `janekbaraniewski/openusage` (**~213★**, **+1★** — still deps/CI-only; product quiet continues into **day 14**; tag **v0.25.0**), `getagentseal/codeburn` (still **0.9.25**; Sun Codex fork-replay preserve + GNOME paired-device combined usage + menubar dock-rings preference; **~11.3k★**), `tddworks/ClaudeBar` (still **v0.4.93**; **today** `claudebar://` URL-scheme open-popover harden; **~1.5k★**), `itvincent-git/codex-usage-desktop` (**app-v3.6.1→app-v3.7.0** Sun — per-model quota estimates + expandable compare charts + per-million consumption; **~43★**), `majiayu000/quotabar` (still **v0.5.4**; open tray empty-slot PR; **~53★**), `Halloweedev/usagepal` (still **v0.7.77-beta.2**; quiet since Sat; **~82★**), `semantic-craft/iOS-vibebuddy` (still **v1.3.37**; Sun night read-aloud opens with spoken project name; **~90★**), `CodeZeno/Claude-Code-Usage-Monitor` (still **v2.15.19**; **~544★**, **+2★**), `robinebers/openusage` (latest Release still **v0.7.12**; repo notes **v0.7.13-beta.2**; **~4.3k★**), `junhoyeo/tokscale` (still **v4.17.0**; **~5.6k★**), `Piebald-AI/splitrail` (still **v3.10.1**; **~222★**), `langfuse/langfuse` (still **v4.46.0**; today pricing drift confirm + mobile search/filters unify; **~35.1k★**), `BerriAI/litellm` (**v1.102.1→v1.103.0** today + **v1.104.0-rc.1**; Azure Mistral OCR pricing + Rust MCP gateway train; **~59.8k★**). Helicone quiet since **16 Sep** (**12th day**).

| Repo | Stars | License | Why it matters |
|---|---:|---|---|
| [ccusage/ccusage](https://github.com/ccusage/ccusage) | ~18.8k | NOASSERTION | Published **v20.0.26** (npm **20.0.26**; tag **v20.0.25** never released) — Codex session filter + Claude dedupe + OpenCode fork exclude (**~18,772★**). |
| [getagentseal/codeburn](https://github.com/getagentseal/codeburn) | ~11.3k | MIT | Still **0.9.25** — Sun Codex fork-replay preserve (#1569) + GNOME paired-device usage. |
| [robinebers/openusage](https://github.com/robinebers/openusage) | ~4.3k | MIT | openusage.ai — Release **v0.7.12** / notes **v0.7.13-beta.2** Rate Limit Resets + Cursor Grok 4.7 + OpenCode 2 OAuth. |
| [janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage) | ~213 | MIT | openusage.sh local multi-tool TUI + SQLite — nearest product twin (**+1★**; product quiet day 14; tag **v0.25.0**). |
| [ItsJazii/pane](https://github.com/ItsJazii/pane) | ~62 | NOASSERTION | Still Release **v0.4.55** — **today** unreleased widget mode merged (**+2★**). |
| [Maciek-roboblog/Claude-Code-Usage-Monitor](https://github.com/Maciek-roboblog/Claude-Code-Usage-Monitor) | ~8.7k | MIT | Real-time Claude Code usage monitor with predictions/warnings. |
| [HermannBjorgvin/Clawdmeter](https://github.com/HermannBjorgvin/Clawdmeter) | ~2.2k | (check) | ESP32 desk dashboard — C6 AMOLED IMU auto-rotation still the recent hardware win. |
| [Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail) | ~222 | MIT | Still **v3.10.1** — GPT-6 Sol/Luna pricing on peak/off-peak + Opus 5.5 + Grok 4.7. |
| [juliantanx/aiusage](https://github.com/juliantanx/aiusage) | ~134 | MIT | Still **v1.5.19** — Antigravity model attribution + cloud sync. |
| [junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) | ~5.6k | MIT | Still **v4.17.0** after Fri TUI i18n + OpenClaw SQLite decode + Codex/Antigravity ledger. |
| [Iamshankhadeep/ccseva](https://github.com/Iamshankhadeep/ccseva) | ~807 | MIT | macOS menu bar for live Claude Code usage. |
| [ColeMurray/claude-code-otel](https://github.com/ColeMurray/claude-code-otel) | ~506 | MIT | Observability stack for Claude Code usage/perf/cost (OTel path). |
| [CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) | ~544 | MIT | Still **v2.15.19** — Direct3D OpenGL fallback + fractional DPI segments (**+2★**). |
| [leeguooooo/claude-code-usage-bar](https://github.com/leeguooooo/claude-code-usage-bar) | ~375 | MIT | Predict profiles keyed by session config dir — statusLine quota UX. |
| [foyzulkarim/claude-lens](https://github.com/foyzulkarim/claude-lens) | ~249 | MIT | Local dashboard: sessions, token costs, cache, tool calls, daily breakdowns. |
| [frankchiu-dev/claude-codex-usage-dashboard](https://github.com/frankchiu-dev/claude-codex-usage-dashboard) | ~173 | MIT | Local Windows dashboard for Claude Code + Codex limits. |
| [planetf1/otelite](https://github.com/planetf1/otelite) | ~94 | NOASSERTION | Still **v0.1.153** — storage ANALYZE + bounded ingest. |
| [alibaba/loongsuite-pilot](https://github.com/alibaba/loongsuite-pilot) | ~198 | Apache-2.0 | Quiet on main after Fri agent interceptor — qwen-code / Codex/hooks path (**+3★**). |
| [seakee/CPA-Manager-Plus](https://github.com/seakee/CPA-Manager-Plus) | ~3.7k | MIT | Still **v1.14.1** — Usage Maintenance reverse-proxy HEAD probe fix. |
| [Rodiun/frugon](https://github.com/Rodiun/frugon) | ~214 | MIT | Local LLM bill-leak analyzer. |
| [mag123c/toktrack](https://github.com/mag123c/toktrack) | ~192 | MIT | Still **v2.17.4** — OpenCode v2 compaction/usage + stream SQLite. |
| [Halloweedev/usagepal](https://github.com/Halloweedev/usagepal) | ~82 | MIT | Still **v0.7.77-beta.2** — Budgets tab + OpenCode Go V2 costs — OpenUsage menu-bar fork (quiet since Sat). |
| [Nanako0129/syrtis](https://github.com/Nanako0129/syrtis) | ~377 | MIT | **Was TokenBar** — **v2.1.0** attribution cards + Grok Build unify + stale gauge; **today** Developer ID notarize + onboarding/sand-tray (**+2★**). |
| [tddworks/ClaudeBar](https://github.com/tddworks/ClaudeBar) | ~1.5k | (check) | Still **v0.4.93** — today `claudebar://` URL-scheme harden (Touch Bar gauges colour by quota). |
| [deviffyy/OpenQuota](https://github.com/deviffyy/OpenQuota) | ~221 | MIT | Cross-platform limits/spend tracker (quiet Sat). |
| [Dicklesworthstone/coding_agent_usage_tracker](https://github.com/Dicklesworthstone/coding_agent_usage_tracker) | ~87 | NOASSERTION | Single CLI for remaining quotas across Codex/Claude/Gemini/Cursor/Copilot. |
| [sculptdotfun/viberank](https://github.com/sculptdotfun/viberank) | ~117 | MIT | Public AI-coding usage leaderboard on ccusage data. |
| [cobra91/better-ccusage](https://github.com/cobra91/better-ccusage) | ~89 | MIT | Faster multi-provider JSONL analyzer. |
| [851-labs/tokenmaxxing](https://github.com/851-labs/tokenmaxxing) | ~77 | MIT | Still **cli v0.7.0-alpha.1** — today login rate-limit + Windows e2e; local CLI on ccusage that syncs usage socially. |
| [Nihondo/AgentLimits](https://github.com/Nihondo/AgentLimits) | ~63 | MIT | macOS widgets for Codex/Claude limits + ccusage heatmap. |
| [majiayu000/quotabar](https://github.com/majiayu000/quotabar) | ~53 | MIT | Still **v0.5.4** — ccstats 0.9.0 crates.io SDK + notarized DMG; open tray empty-slot PR. |
| [abhiunix/AgentHarbor](https://github.com/abhiunix/AgentHarbor) | ~12 | MIT | Native multi-agent rate-limit + session usage app (traffic snapshots only). |
| [semantic-craft/iOS-vibebuddy](https://github.com/semantic-craft/iOS-vibebuddy) | ~90 | MIT | Still **v1.3.37** — Claude allowance via background CLI probe when status-line absent. |
| [zhnd/lumo](https://github.com/zhnd/lumo) | ~147 | MIT | Local-first Claude Code usage/cost/session dashboard. |
| [fschmutz/claude-usage-panel](https://github.com/fschmutz/claude-usage-panel) | ~6 | MIT | Official usage API → GNOME/macOS/statusLine + MCP get_usage (Claude + Cursor). |
| [sylearn/AIUsage](https://github.com/sylearn/AIUsage) | ~688 | Apache-2.0 | Desktop multi-tool usage peer (still **v0.15.21**). |
| [BerriAI/litellm](https://github.com/BerriAI/litellm) | ~59.8k | NOASSERTION | Gateway spend/pricing — adjacent (**v1.103.0** today + **v1.104.0-rc.1**; Azure OCR pricing + Rust MCP gateway). |
| [langfuse/langfuse](https://github.com/langfuse/langfuse) | ~35.1k | NOASSERTION | Still **v4.46.0** — OSS LLM observability + OTel — adjacent (today pricing confirm + mobile search unify). |
| [Helicone/helicone](https://github.com/Helicone/helicone) | ~6.2k | Apache-2.0 | Quiet since 16 Sep (**12th day**). OSS LLM observability proxy. Adjacent. |

**Also watch**
- [ClaudeCodeUsage/ClaudeCodeUsage](https://github.com/ClaudeCodeUsage/ClaudeCodeUsage) (~66★, MIT) — quiet after Thu contributor-credits docs; VS Code status-bar Claude Code usage/cost.
- [itvincent-git/codex-usage-desktop](https://github.com/itvincent-git/codex-usage-desktop) (~43★, MIT) — **app-v3.7.0** Sun — per-model quota estimate charts + per-million consumption.
- [saeedkolivand/claude-usage-mac](https://github.com/saeedkolivand/claude-usage-mac) (~15★) — Claude Code macOS menu bar + desktop widget.
- [ofershap/cursor-usage-tracker](https://github.com/ofershap/cursor-usage-tracker) (~33★, MIT) — Cursor Enterprise FinOps, not personal autodetection.
- [DeepAgentLabs/agenticlens](https://github.com/DeepAgentLabs/agenticlens) (~59★, MIT) — agent token/cost/latency profiler (quiet).
- Micro/zero-★ Claude wrappers keep spawning daily — noise floor, not category peers.

**Build takeaways:** Stay ccusage-compatible (**v20.0.26** published; tag **v20.0.25** skipped as a Release). Differentiate on live quotas + auto-detect + UX vs openusage.sh (still product-quiet day 14). Today's sharpest signals: **ccusage v20.0.26**, **Syrtis v2.1.0 + Monday Developer-ID/onboarding/sand-tray train**, **pane unreleased widget mode**, **LiteLLM v1.103.0**, **codex-usage-desktop app-v3.7.0**. Keep Langfuse/LiteLLM complementary; Helicone remains maintenance-mode adjacent (day 12).

Sources: GitHub API 2026-09-28 (Europe/London).

### Open source — 29 Sep 2026

Fresh Tuesday GitHub API pass (Europe/London). Category still owned by **ccusage** + the two OpenUsage brands; **codeburn** remains the loud multi-tool local tracker. **Live today / overnight:** `Nanako0129/syrtis` (**v2.1.0→v2.2.0** Mon ~16:02 BST — Developer-ID signing + notarized installer DMG [#445](https://github.com/Nanako0129/syrtis/pull/445)/[#454](https://github.com/Nanako0129/syrtis/pull/454), **Sand shoal** animated tray + animation-pace curve [#441](https://github.com/Nanako0129/syrtis/pull/441), first-run Overview setup cards [#442](https://github.com/Nanako0129/syrtis/pull/442); main **~3 commits ahead** landing notarized-DMG copy [#456](https://github.com/Nanako0129/syrtis/pull/456); **~385★**, **+8★**), `langfuse/langfuse` (**v4.46.0→v4.47.0** today ~09:25 BST — dedicated trace transcripts [#17925](https://github.com/langfuse/langfuse/pull/17925), AI-gateway full-mode capture + env header [#17956](https://github.com/langfuse/langfuse/pull/17956)/[#17960](https://github.com/langfuse/langfuse/pull/17960), Anthropic/OpenAI 1-hour cache-write pricing; **~35.2k★**), `Piebald-AI/splitrail` (**v3.10.1→v3.10.2** Mon — Claude Sonnet 5.5 pricing [#266](https://github.com/Piebald-AI/splitrail/pull/266); **~222★**), `CodeZeno/Claude-Code-Usage-Monitor` (**v2.15.19→v2.15.22** ~00:11 BST Tue — Grok billing-period as 0% when unused + violet Compact Fluent [#149](https://github.com/CodeZeno/Claude-Code-Usage-Monitor/pull/149); **~545★**, **+1★**), `ItsJazii/pane` (still Release **v0.4.55**; Mon merged unreleased **widget mode** [#248](https://github.com/ItsJazii/pane/pull/248) — pinned/draggable/collapsible desktop bar; **~64★**, **+2★**), `ccusage/ccusage` (still **v20.0.26**; Tue = continuous models.dev/LiteLLM pricing snapshots only; **~18,793★**, **+21★**), `janekbaraniewski/openusage` (**~213★** flat — still deps/CI-only #383–#387; product quiet continues into **day 15**; tag **v0.25.0**; homepage copy drift still **34 / 35 / 36** providers vs README **36** docs pages), `majiayu000/quotabar` (**v0.5.4→v0.5.5** Mon — Grok CLI session renew + Windows tray popover DPI; **~53★**), `alibaba/loongsuite-pilot` (**v1.11.0** Mon — local Agent Interceptor for Qoder/Qwen/OpenClaw + Qwen Code subagent traces; today qoder image/tool-media fix; **~198★**), `leeguooooo/claude-code-usage-bar` (**v3.43.4** today — Windows installer auto-installs uv via winget; **~375★**), `851-labs/tokenmaxxing` (**cli v0.7.0-alpha.1→alpha.3** — Windows/Linux battle-test + exact-version upgrades; **~76★**), `itvincent-git/codex-usage-desktop` (**app-v3.7.0→app-v3.8.2** Mon — release train after Sun’s per-model charts; **~43★**), `getagentseal/codeburn` (still **0.9.25**; Mon unreleased — per-config Claude dock rings [#1570](https://github.com/getagentseal/codeburn/pull/1570), OpenClaw per-agent sqlite [#1526](https://github.com/getagentseal/codeburn/pull/1526), DeepSeek/Z.ai peak vs off-peak [#1561](https://github.com/getagentseal/codeburn/pull/1561); **~11.3k★**), `tddworks/ClaudeBar` (still **v0.4.93**; today/Mon unreleased Cursor Auto+API quotas from usage-summary + Z.ai API-key settings; **~1.5k★**), `robinebers/openusage` (latest Release still **v0.7.12**; notes **v0.7.13-beta.2**; today Claude Sonnet 5.5 pricing commit; **~4.3k★**), `junhoyeo/tokscale` (still **v4.17.0**; today Claude Code 1-hour cache-write pricing + TUI localization/import recovery [#1374](https://github.com/junhoyeo/tokscale/pull/1374)/[#1371](https://github.com/junhoyeo/tokscale/pull/1371); **~5.6k★**), `BerriAI/litellm` (still Latest **v1.103.0** + **v1.104.0-rc.1**; Tue OTel Langfuse cache/reasoning usage_details + Bedrock Mantle Claude 5.5 cost-map; **~59.8k★**). Helicone quiet since **16 Sep** (**13th day**).

| Repo | Stars | License | Why it matters |
|---|---:|---|---|
| [ccusage/ccusage](https://github.com/ccusage/ccusage) | ~18.8k | NOASSERTION | Still **v20.0.26** — Tue models.dev/LiteLLM pricing snapshots only (**~18,793★**, **+21★**). |
| [getagentseal/codeburn](https://github.com/getagentseal/codeburn) | ~11.3k | MIT | Still **0.9.25** — Mon unreleased per-config Claude dock rings + OpenClaw sqlite + DeepSeek/Z.ai peak/off-peak. |
| [robinebers/openusage](https://github.com/robinebers/openusage) | ~4.3k | MIT | openusage.ai — Release **v0.7.12** / notes **v0.7.13-beta.2**; today Claude Sonnet 5.5 pricing. |
| [janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage) | ~213 | MIT | openusage.sh local multi-tool TUI + SQLite — nearest product twin (product quiet **day 15**; tag **v0.25.0**; **34/35/36** provider copy drift). |
| [ItsJazii/pane](https://github.com/ItsJazii/pane) | ~64 | NOASSERTION | Still Release **v0.4.55** — Mon unreleased widget mode merged (#248) (**+2★**). |
| [Maciek-roboblog/Claude-Code-Usage-Monitor](https://github.com/Maciek-roboblog/Claude-Code-Usage-Monitor) | ~8.7k | MIT | Real-time Claude Code usage monitor with predictions/warnings. |
| [HermannBjorgvin/Clawdmeter](https://github.com/HermannBjorgvin/Clawdmeter) | ~2.2k | (check) | ESP32 desk dashboard — C6 AMOLED IMU auto-rotation still the recent hardware win. |
| [Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail) | ~222 | MIT | **v3.10.2** — Claude Sonnet 5.5 pricing (#266). |
| [juliantanx/aiusage](https://github.com/juliantanx/aiusage) | ~134 | MIT | Still **v1.5.19** — Antigravity model attribution + cloud sync. |
| [junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) | ~5.6k | MIT | Still **v4.17.0** — today 1-hour cache-write pricing + TUI localization/import recovery. |
| [Iamshankhadeep/ccseva](https://github.com/Iamshankhadeep/ccseva) | ~807 | MIT | macOS menu bar for live Claude Code usage. |
| [ColeMurray/claude-code-otel](https://github.com/ColeMurray/claude-code-otel) | ~507 | MIT | Observability stack for Claude Code usage/perf/cost (OTel path). |
| [CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) | ~545 | MIT | **v2.15.22** — Grok unused billing-period → 0% + violet Compact Fluent (**+1★**). |
| [leeguooooo/claude-code-usage-bar](https://github.com/leeguooooo/claude-code-usage-bar) | ~375 | MIT | **v3.43.4** — Windows installer auto-installs uv via winget. |
| [foyzulkarim/claude-lens](https://github.com/foyzulkarim/claude-lens) | ~250 | MIT | Local dashboard: sessions, token costs, cache, tool calls, daily breakdowns. |
| [frankchiu-dev/claude-codex-usage-dashboard](https://github.com/frankchiu-dev/claude-codex-usage-dashboard) | ~173 | MIT | Local Windows dashboard for Claude Code + Codex limits. |
| [planetf1/otelite](https://github.com/planetf1/otelite) | ~94 | NOASSERTION | Still **v0.1.153** — storage ANALYZE + bounded ingest. |
| [alibaba/loongsuite-pilot](https://github.com/alibaba/loongsuite-pilot) | ~198 | Apache-2.0 | **v1.11.0** — Agent Interceptor (Qoder/Qwen/OpenClaw) + Qwen Code subagent traces. |
| [seakee/CPA-Manager-Plus](https://github.com/seakee/CPA-Manager-Plus) | ~3.7k | MIT | Still **v1.14.1** — Usage Maintenance reverse-proxy HEAD probe fix. |
| [Rodiun/frugon](https://github.com/Rodiun/frugon) | ~214 | MIT | Local LLM bill-leak analyzer. |
| [mag123c/toktrack](https://github.com/mag123c/toktrack) | ~192 | MIT | Still **v2.17.4** — OpenCode v2 compaction/usage + stream SQLite. |
| [Halloweedev/usagepal](https://github.com/Halloweedev/usagepal) | ~82 | MIT | Still **v0.7.77-beta.2** — Budgets tab + OpenCode Go V2 costs — OpenUsage menu-bar fork (quiet). |
| [Nanako0129/syrtis](https://github.com/Nanako0129/syrtis) | ~385 | MIT | **Was TokenBar** — **v2.2.0** notarized DMG + Sand shoal tray + Overview setup cards (**+8★**; main ~3 ahead). |
| [tddworks/ClaudeBar](https://github.com/tddworks/ClaudeBar) | ~1.5k | (check) | Still **v0.4.93** — unreleased Cursor Auto/API quotas + Z.ai API-key settings. |
| [deviffyy/OpenQuota](https://github.com/deviffyy/OpenQuota) | ~221 | MIT | Cross-platform limits/spend tracker (quiet). |
| [Dicklesworthstone/coding_agent_usage_tracker](https://github.com/Dicklesworthstone/coding_agent_usage_tracker) | ~87 | NOASSERTION | Single CLI for remaining quotas across Codex/Claude/Gemini/Cursor/Copilot. |
| [sculptdotfun/viberank](https://github.com/sculptdotfun/viberank) | ~117 | MIT | Public AI-coding usage leaderboard on ccusage data. |
| [cobra91/better-ccusage](https://github.com/cobra91/better-ccusage) | ~89 | MIT | Faster multi-provider JSONL analyzer. |
| [851-labs/tokenmaxxing](https://github.com/851-labs/tokenmaxxing) | ~76 | MIT | **cli v0.7.0-alpha.3** — Windows/Linux battle-test + exact-version upgrades; local CLI on ccusage that syncs usage socially. |
| [Nihondo/AgentLimits](https://github.com/Nihondo/AgentLimits) | ~63 | MIT | macOS widgets for Codex/Claude limits + ccusage heatmap. |
| [majiayu000/quotabar](https://github.com/majiayu000/quotabar) | ~53 | MIT | **v0.5.5** — Grok CLI session renew + Windows tray popover DPI. |
| [abhiunix/AgentHarbor](https://github.com/abhiunix/AgentHarbor) | ~12 | MIT | Native multi-agent rate-limit + session usage app (traffic snapshots only). |
| [semantic-craft/iOS-vibebuddy](https://github.com/semantic-craft/iOS-vibebuddy) | ~91 | MIT | Still **v1.3.37** — chore cleanup PR; Claude allowance via background CLI probe (**+1★**). |
| [zhnd/lumo](https://github.com/zhnd/lumo) | ~147 | MIT | Local-first Claude Code usage/cost/session dashboard. |
| [fschmutz/claude-usage-panel](https://github.com/fschmutz/claude-usage-panel) | ~6 | MIT | Official usage API → GNOME/macOS/statusLine + MCP get_usage (Claude + Cursor). |
| [sylearn/AIUsage](https://github.com/sylearn/AIUsage) | ~690 | Apache-2.0 | Desktop multi-tool usage peer (still **v0.15.21**). |
| [BerriAI/litellm](https://github.com/BerriAI/litellm) | ~59.8k | NOASSERTION | Gateway spend/pricing — adjacent (still **v1.103.0** + **v1.104.0-rc.1**; Tue OTel Langfuse usage_details + Claude 5.5 cost-map). |
| [langfuse/langfuse](https://github.com/langfuse/langfuse) | ~35.2k | NOASSERTION | **v4.47.0** — OSS LLM observability + OTel — adjacent (trace transcripts + AI-gateway capture). |
| [Helicone/helicone](https://github.com/Helicone/helicone) | ~6.2k | Apache-2.0 | Quiet since 16 Sep (**13th day**). OSS LLM observability proxy. Adjacent. |

**Also watch**
- [ClaudeCodeUsage/ClaudeCodeUsage](https://github.com/ClaudeCodeUsage/ClaudeCodeUsage) (~66★, MIT) — quiet after Thu contributor-credits docs; VS Code status-bar Claude Code usage/cost.
- [itvincent-git/codex-usage-desktop](https://github.com/itvincent-git/codex-usage-desktop) (~43★, MIT) — **app-v3.8.2** Mon — follow-on to Sun’s per-model quota charts.
- [saeedkolivand/claude-usage-mac](https://github.com/saeedkolivand/claude-usage-mac) (~15★) — Claude Code macOS menu bar + desktop widget.
- [ofershap/cursor-usage-tracker](https://github.com/ofershap/cursor-usage-tracker) (~33★, MIT) — Cursor Enterprise FinOps, not personal autodetection.
- [DeepAgentLabs/agenticlens](https://github.com/DeepAgentLabs/agenticlens) (~59★, MIT) — agent token/cost/latency profiler (quiet).
- Micro/zero-★ Claude wrappers keep spawning daily — noise floor, not category peers.

**Build takeaways:** Stay ccusage-compatible (**v20.0.26** holds; Tue pricing-snapshot only). Differentiate on live quotas + auto-detect + UX vs openusage.sh (still product-quiet day 15; **34/35/36** provider copy drift). Today's sharpest signals: **Syrtis v2.2.0 notarized DMG + Sand shoal**, **Langfuse v4.47.0**, **Splitrail v3.10.2 Sonnet 5.5**, **CodeZeno v2.15.22 Grok billing**, **pane unreleased widget mode**, **QuotaBar v0.5.5**, **LoongSuite Pilot v1.11.0**. Keep Langfuse/LiteLLM complementary; Helicone remains maintenance-mode adjacent (day 13).

Sources: GitHub API 2026-09-29 (Europe/London).

### Open source — 30 Sep 2026

Fresh Wednesday GitHub API pass (Europe/London). Category still owned by **ccusage** + the two OpenUsage brands; **codeburn** remains the loud multi-tool local tracker. **Live today / overnight:** `janekbaraniewski/openusage` (**product-surface quiet broken** Tue evening — **not** quiet day 16 — Antigravity frozen quota snapshot [#378](https://github.com/janekbaraniewski/openusage/pull/378), Z.AI credit-based Coding Plan quotas [#388](https://github.com/janekbaraniewski/openusage/pull/388), Copilot GHE hostnames [#373](https://github.com/janekbaraniewski/openusage/pull/373), Codex latest-session-by-filename [#372](https://github.com/janekbaraniewski/openusage/pull/372), hermes MissingDB isolation [#391](https://github.com/janekbaraniewski/openusage/pull/391); tag still **v0.25.0**; **~215★**, **+2★**; homepage **34 / 35 / 36** provider copy drift holds), `robinebers/openusage` (**v0.7.13-beta.2→v0.7.13-beta.3** today ~10:28 BST — Codex one-card-per-account [#1321](https://github.com/robinebers/openusage/pull/1321), Ultrafast + GPT-6 Sol pricing [#1327](https://github.com/robinebers/openusage/pull/1327), Cursor grok-bot-cua as Grok 4.7 [#1329](https://github.com/robinebers/openusage/pull/1329), OpenCode 2 usage [#1323](https://github.com/robinebers/openusage/pull/1323); stable Release still **v0.7.12**; **~4.3k★**), `Piebald-AI/splitrail` (**v3.10.2→v3.10.3** Tue ~20:25 BST — GPT-6.1 Sol pricing [#268](https://github.com/Piebald-AI/splitrail/pull/268)/[#269](https://github.com/Piebald-AI/splitrail/pull/269); **~222★**), `BerriAI/litellm` (**v1.103.0→v1.103.1** ~01:59 BST Wed + **v1.104.0-rc.2** + **v1.105.0-dev.1**; **~59.9k★**), `851-labs/tokenmaxxing` (**cli v0.7.0-alpha.3→cli-v0.7.1** — Windows service install/repair + Linux deferred service repair; **~77★**), `majiayu000/quotabar` (**v0.5.5→v0.5.7** — v0.5.6 interface size 100/125/150% + v0.5.7 Codex credit-balance formatting; **~53★**), `leeguooooo/claude-code-usage-bar` (**v3.43.4→v3.44.0** Tue — ocs unread DMs + LAN bridge state; **~376★**), `itvincent-git/codex-usage-desktop` (**app-v3.8.2→app-v3.8.3** today — macOS child reaping after exit; **~44★**), `tddworks/ClaudeBar` (still **v0.4.93**; today unreleased Claude usage boot-screen [#320](https://github.com/tddworks/ClaudeBar/pull/320) + cost-fallback `$0.00` invent stop; **~1.5k★**), `getagentseal/codeburn` (still **0.9.25**; Tue unreleased Codex response usage [#1571](https://github.com/getagentseal/codeburn/pull/1571) + DeepSeek/gpt-5.6-codex pricing invariants [#1577](https://github.com/getagentseal/codeburn/pull/1577); **~11.3k★**), `ccusage/ccusage` (still **v20.0.26**; Wed = models.dev/LiteLLM pricing snapshots only; **~18,813★**, **+20★** vs Tue), `Nanako0129/syrtis` (still **v2.2.0** notarized DMG + Sand shoal; main still **~3 ahead** landing #456; **~387★**, **+2★**), `ItsJazii/pane` (still Release **v0.4.55**; widget mode [#248](https://github.com/ItsJazii/pane/pull/248) still **5 commits ahead** / unreleased; **~65★**), `langfuse/langfuse` (still **v4.47.0**; Wed/Tue pricing GPT-6.1 Sol + Claude Sonnet 5.5 defaults [#18055](https://github.com/langfuse/langfuse/pull/18055)/[#18023](https://github.com/langfuse/langfuse/pull/18023); **~35.2k★**), `junhoyeo/tokscale` (still **v4.17.0**; Tue unreleased 1-hour cache-write pricing + TUI localization/import recovery still untagged; **~5.6k★**), `fschmutz/claude-usage-panel` (still Release **v2.2.0**; Tue unreleased honest readings / parked-account refresh / Codex vault [#27](https://github.com/fschmutz/claude-usage-panel/pull/27); **~6★**), `CodeZeno/Claude-Code-Usage-Monitor` (still **v2.15.22**; quiet since Tue Grok billing; **~546★**). Helicone quiet since **16 Sep** (**14th day**).

| Repo | Stars | License | Why it matters |
|---|---:|---|---|
| [ccusage/ccusage](https://github.com/ccusage/ccusage) | ~18.8k | NOASSERTION | Still **v20.0.26** — Wed models.dev/LiteLLM pricing snapshots only (**~18,813★**, **+20★**). |
| [getagentseal/codeburn](https://github.com/getagentseal/codeburn) | ~11.3k | MIT | Still **0.9.25** — Tue unreleased Codex response usage (#1571) + DeepSeek/gpt-5.6-codex pricing invariants (#1577). |
| [robinebers/openusage](https://github.com/robinebers/openusage) | ~4.3k | MIT | openusage.ai — Release **v0.7.12** / notes **v0.7.13-beta.3** today (Codex multi-account + Ultrafast/GPT-6 Sol + OpenCode 2). |
| [janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage) | ~215 | MIT | openusage.sh local multi-tool TUI + SQLite — nearest product twin (**product quiet broken** Tue evening after day 15; tag still **v0.25.0**; **34/35/36** provider copy drift; **+2★**). |
| [ItsJazii/pane](https://github.com/ItsJazii/pane) | ~65 | NOASSERTION | Still Release **v0.4.55** — widget mode (#248) still **5 ahead** / unreleased. |
| [Maciek-roboblog/Claude-Code-Usage-Monitor](https://github.com/Maciek-roboblog/Claude-Code-Usage-Monitor) | ~8.7k | MIT | Real-time Claude Code usage monitor with predictions/warnings. |
| [HermannBjorgvin/Clawdmeter](https://github.com/HermannBjorgvin/Clawdmeter) | ~2.2k | (check) | ESP32 desk dashboard — C6 AMOLED IMU auto-rotation still the recent hardware win. |
| [Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail) | ~222 | MIT | **v3.10.3** — GPT-6.1 Sol pricing (#268/#269). |
| [juliantanx/aiusage](https://github.com/juliantanx/aiusage) | ~135 | MIT | Still **v1.5.19** — Antigravity model attribution + cloud sync (quiet since Thu). |
| [junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) | ~5.6k | MIT | Still **v4.17.0** — Tue unreleased 1-hour cache-write pricing + TUI localization/import recovery. |
| [Iamshankhadeep/ccseva](https://github.com/Iamshankhadeep/ccseva) | ~807 | MIT | macOS menu bar for live Claude Code usage. |
| [ColeMurray/claude-code-otel](https://github.com/ColeMurray/claude-code-otel) | ~507 | MIT | Observability stack for Claude Code usage/perf/cost (OTel path). |
| [CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) | ~546 | MIT | Still **v2.15.22** — quiet since Tue Grok unused billing-period → 0%. |
| [leeguooooo/claude-code-usage-bar](https://github.com/leeguooooo/claude-code-usage-bar) | ~376 | MIT | **v3.44.0** — ocs unread DMs + LAN bridge state on the ocs line. |
| [foyzulkarim/claude-lens](https://github.com/foyzulkarim/claude-lens) | ~250 | MIT | Local dashboard: sessions, token costs, cache, tool calls, daily breakdowns. |
| [frankchiu-dev/claude-codex-usage-dashboard](https://github.com/frankchiu-dev/claude-codex-usage-dashboard) | ~173 | MIT | Local Windows dashboard for Claude Code + Codex limits. |
| [planetf1/otelite](https://github.com/planetf1/otelite) | ~95 | NOASSERTION | Still **v0.1.153** — storage ANALYZE + bounded ingest. |
| [alibaba/loongsuite-pilot](https://github.com/alibaba/loongsuite-pilot) | ~198 | Apache-2.0 | Still **v1.11.0** — Tue qoder image/tool-media fix (#448). |
| [seakee/CPA-Manager-Plus](https://github.com/seakee/CPA-Manager-Plus) | ~3.7k | MIT | Still **v1.14.1** — Usage Maintenance reverse-proxy HEAD probe fix. |
| [Rodiun/frugon](https://github.com/Rodiun/frugon) | ~214 | MIT | Local LLM bill-leak analyzer. |
| [mag123c/toktrack](https://github.com/mag123c/toktrack) | ~192 | MIT | Still **v2.17.4** — Tue deps-only thiserror bump. |
| [Halloweedev/usagepal](https://github.com/Halloweedev/usagepal) | ~82 | MIT | Still **v0.7.77-beta.2** — Budgets tab + OpenCode Go V2 costs — OpenUsage menu-bar fork (quiet). |
| [Nanako0129/syrtis](https://github.com/Nanako0129/syrtis) | ~387 | MIT | **Was TokenBar** — still **v2.2.0** notarized DMG + Sand shoal (**+2★**; main ~3 ahead). |
| [tddworks/ClaudeBar](https://github.com/tddworks/ClaudeBar) | ~1.5k | (check) | Still **v0.4.93** — unreleased Claude boot-screen (#320) + cost-fallback invent stop. |
| [deviffyy/OpenQuota](https://github.com/deviffyy/OpenQuota) | ~221 | MIT | Cross-platform limits/spend tracker (quiet). |
| [Dicklesworthstone/coding_agent_usage_tracker](https://github.com/Dicklesworthstone/coding_agent_usage_tracker) | ~87 | NOASSERTION | Single CLI for remaining quotas across Codex/Claude/Gemini/Cursor/Copilot. |
| [sculptdotfun/viberank](https://github.com/sculptdotfun/viberank) | ~117 | MIT | Public AI-coding usage leaderboard on ccusage data. |
| [cobra91/better-ccusage](https://github.com/cobra91/better-ccusage) | ~90 | MIT | Faster multi-provider JSONL analyzer. |
| [851-labs/tokenmaxxing](https://github.com/851-labs/tokenmaxxing) | ~77 | MIT | **cli-v0.7.1** — Windows service + Linux deferred repair; local CLI on ccusage that syncs usage socially. |
| [Nihondo/AgentLimits](https://github.com/Nihondo/AgentLimits) | ~63 | MIT | macOS widgets for Codex/Claude limits + ccusage heatmap. |
| [majiayu000/quotabar](https://github.com/majiayu000/quotabar) | ~53 | MIT | **v0.5.7** — Codex credit-balance formatting (v0.5.6 interface size scaling). |
| [abhiunix/AgentHarbor](https://github.com/abhiunix/AgentHarbor) | ~12 | MIT | Native multi-agent rate-limit + session usage app (traffic snapshots only). |
| [semantic-craft/iOS-vibebuddy](https://github.com/semantic-craft/iOS-vibebuddy) | ~91 | MIT | Still **v1.3.37** — Claude allowance via background CLI probe. |
| [zhnd/lumo](https://github.com/zhnd/lumo) | ~147 | MIT | Local-first Claude Code usage/cost/session dashboard. |
| [fschmutz/claude-usage-panel](https://github.com/fschmutz/claude-usage-panel) | ~6 | MIT | Still **v2.2.0** — Tue unreleased honest readings / parked-account refresh / Codex vault (#27). |
| [sylearn/AIUsage](https://github.com/sylearn/AIUsage) | ~692 | Apache-2.0 | Desktop multi-tool usage peer (still **v0.15.21**). |
| [BerriAI/litellm](https://github.com/BerriAI/litellm) | ~59.9k | NOASSERTION | Gateway spend/pricing — adjacent (**v1.103.1** + **v1.104.0-rc.2** + **v1.105.0-dev.1**). |
| [langfuse/langfuse](https://github.com/langfuse/langfuse) | ~35.2k | NOASSERTION | Still **v4.47.0** — OSS LLM observability + OTel — adjacent (Wed GPT-6.1 Sol / Sonnet 5.5 pricing). |
| [Helicone/helicone](https://github.com/Helicone/helicone) | ~6.2k | Apache-2.0 | Quiet since 16 Sep (**14th day**). OSS LLM observability proxy. Adjacent. |

**Also watch**
- [ClaudeCodeUsage/ClaudeCodeUsage](https://github.com/ClaudeCodeUsage/ClaudeCodeUsage) (~66★, MIT) — quiet after Thu contributor-credits docs; VS Code status-bar Claude Code usage/cost.
- [itvincent-git/codex-usage-desktop](https://github.com/itvincent-git/codex-usage-desktop) (~44★, MIT) — **app-v3.8.3** today — macOS child reaping after exit.
- [saeedkolivand/claude-usage-mac](https://github.com/saeedkolivand/claude-usage-mac) (~15★) — Claude Code macOS menu bar + desktop widget.
- [ofershap/cursor-usage-tracker](https://github.com/ofershap/cursor-usage-tracker) (~33★, MIT) — Cursor Enterprise FinOps, not personal autodetection.
- [DeepAgentLabs/agenticlens](https://github.com/DeepAgentLabs/agenticlens) (~59★, MIT) — agent token/cost/latency profiler (quiet).
- Micro/zero-★ Claude wrappers keep spawning daily — noise floor, not category peers.

**Build takeaways:** Stay ccusage-compatible (**v20.0.26** holds; Wed pricing-snapshot only). Differentiate on live quotas + auto-detect + UX vs openusage.sh (**product surface resumed** Tue evening after 15 quiet days — still no new tag; **34/35/36** provider copy drift). Today's sharpest signals: **OpenUsage.sh product resume**, **openusage.ai v0.7.13-beta.3**, **Splitrail v3.10.3 GPT-6.1 Sol**, **LiteLLM v1.103.1**, **tokenmaxxing cli-v0.7.1**, **QuotaBar v0.5.7**, **usage-bar v3.44.0**. Keep Langfuse/LiteLLM complementary; Helicone remains maintenance-mode adjacent (day 14).

Sources: GitHub API 2026-09-30 (Europe/London).

### Open source — 1 Oct 2026

Fresh Thursday GitHub API pass (Europe/London). Category still owned by **ccusage** + the two OpenUsage brands; **codeburn** remains the loud multi-tool local tracker. **Live today / overnight:** `BerriAI/litellm` (**today** **v1.103.1→v1.103.2** ~07:37 BST; **~60.0k★**), `majiayu000/quotabar` (**Wed afternoon** **v0.5.7→v0.5.9** — v0.5.8 AppImage icon metadata + price catalog; v0.5.9 Linux AppImage launcher permissions / Firejail; **~53★**), `langfuse/langfuse` (**Wed** **v4.47.0→v4.48.0** — incident-session seeder + skills draft/hash fetch; **~35.3k★**), `itvincent-git/codex-usage-desktop` (**Wed** **app-v3.8.3→app-v3.9.0** — monthly five-hour + weekly quota consumption estimates + Windows Codex batch launch quoting; **~44★**), `tddworks/ClaudeBar` (still **v0.4.93**; **today** unreleased Kimi CLI probe / region option [#325](https://github.com/tddworks/ClaudeBar/pull/325) + Codex RPC passive until explicit refresh click [#322](https://github.com/tddworks/ClaudeBar/pull/322); **~1.5k★**), `ItsJazii/pane` (still Release **v0.4.55**; main now **~27 ahead** — widget mode [#248](https://github.com/ItsJazii/pane/pull/248) merged + card row-drag + glass collapsed bar + docs aimed at **0.4.56**; **~65★**), `ccusage/ccusage` (still **v20.0.26**; Thu = models.dev pricing snapshots only; **~18,827★**, **+~14★** vs Wed), `janekbaraniewski/openusage` (**quiet since Tue evening** product resume — no new commits Wed/Thu; tag still **v0.25.0**; homepage **34 / 35 / 36** provider copy drift holds; **~215★**), `robinebers/openusage` (hold notes **v0.7.13-beta.3** / stable Release **v0.7.12**; **~4.3k★**), `Piebald-AI/splitrail` (hold **v3.10.3**; **~222★**), `851-labs/tokenmaxxing` (still **cli-v0.7.1**; Wed evening unreleased Windows `npx.cmd` via cmd.exe [#130](https://github.com/851-labs/tokenmaxxing/pull/130); **~78★**), `leeguooooo/claude-code-usage-bar` (hold **v3.44.0**; **~376★**), `getagentseal/codeburn` (still **0.9.25**; Tue unreleased Codex response usage + pricing invariants remain untagged; **~11.3k★**), `Nanako0129/syrtis` (still **v2.2.0**; **~387★**), `junhoyeo/tokscale` (still **v4.17.0**; Tue unreleased cache-write pricing / TUI localization still untagged; **~5.6k★**), `sylearn/AIUsage` (**today** **v0.15.21→v0.15.22** — Claude Science local Skills/MCP without cloud login + effort retention; **~693★**), `CodeZeno/Claude-Code-Usage-Monitor` (still **v2.15.22**; **~552★**). Helicone quiet since **16 Sep** (**15th day**).

| Repo | Stars | License | Why it matters |
|---|---:|---|---|
| [ccusage/ccusage](https://github.com/ccusage/ccusage) | ~18.8k | NOASSERTION | Still **v20.0.26** — Thu models.dev pricing snapshots only (**~18,827★**). |
| [getagentseal/codeburn](https://github.com/getagentseal/codeburn) | ~11.3k | MIT | Still **0.9.25** — Tue unreleased Codex response usage (#1571) + pricing invariants (#1577). |
| [robinebers/openusage](https://github.com/robinebers/openusage) | ~4.3k | MIT | openusage.ai — Release **v0.7.12** / notes **v0.7.13-beta.3** (Codex multi-account + Ultrafast/GPT-6 Sol). |
| [janekbaraniewski/openusage](https://github.com/janekbaraniewski/openusage) | ~215 | MIT | openusage.sh local multi-tool TUI + SQLite — nearest product twin (quiet since Tue evening resume; tag still **v0.25.0**; **34/35/36** provider copy drift). |
| [ItsJazii/pane](https://github.com/ItsJazii/pane) | ~65 | NOASSERTION | Still Release **v0.4.55** — widget mode + card drag now on main (**~27 ahead** / aimed at 0.4.56). |
| [Maciek-roboblog/Claude-Code-Usage-Monitor](https://github.com/Maciek-roboblog/Claude-Code-Usage-Monitor) | ~8.7k | MIT | Real-time Claude Code usage monitor with predictions/warnings. |
| [HermannBjorgvin/Clawdmeter](https://github.com/HermannBjorgvin/Clawdmeter) | ~2.2k | (check) | ESP32 desk dashboard — hardware peer. |
| [Piebald-AI/splitrail](https://github.com/Piebald-AI/splitrail) | ~222 | MIT | Hold **v3.10.3** — GPT-6.1 Sol pricing. |
| [juliantanx/aiusage](https://github.com/juliantanx/aiusage) | ~135 | MIT | Still **v1.5.19** — Antigravity model attribution + cloud sync (quiet). |
| [junhoyeo/tokscale](https://github.com/junhoyeo/tokscale) | ~5.6k | MIT | Still **v4.17.0** — unreleased 1-hour cache-write pricing + TUI localization. |
| [Iamshankhadeep/ccseva](https://github.com/Iamshankhadeep/ccseva) | ~807 | MIT | macOS menu bar for live Claude Code usage. |
| [ColeMurray/claude-code-otel](https://github.com/ColeMurray/claude-code-otel) | ~507 | MIT | Observability stack for Claude Code usage/perf/cost (OTel path). |
| [CodeZeno/Claude-Code-Usage-Monitor](https://github.com/CodeZeno/Claude-Code-Usage-Monitor) | ~552 | MIT | Still **v2.15.22** — quiet. |
| [leeguooooo/claude-code-usage-bar](https://github.com/leeguooooo/claude-code-usage-bar) | ~376 | MIT | Hold **v3.44.0** — ocs unread DMs + LAN bridge state. |
| [foyzulkarim/claude-lens](https://github.com/foyzulkarim/claude-lens) | ~250 | MIT | Local dashboard: sessions, token costs, cache, tool calls, daily breakdowns. |
| [frankchiu-dev/claude-codex-usage-dashboard](https://github.com/frankchiu-dev/claude-codex-usage-dashboard) | ~173 | MIT | Local Windows dashboard for Claude Code + Codex limits. |
| [planetf1/otelite](https://github.com/planetf1/otelite) | ~95 | NOASSERTION | Still **v0.1.153** — storage ANALYZE + bounded ingest. |
| [alibaba/loongsuite-pilot](https://github.com/alibaba/loongsuite-pilot) | ~198 | Apache-2.0 | Still **v1.11.0** — qoder image/tool-media path. |
| [seakee/CPA-Manager-Plus](https://github.com/seakee/CPA-Manager-Plus) | ~3.7k | MIT | Still **v1.14.1** — Usage Maintenance reverse-proxy path. |
| [Rodiun/frugon](https://github.com/Rodiun/frugon) | ~214 | MIT | Local LLM bill-leak analyzer. |
| [mag123c/toktrack](https://github.com/mag123c/toktrack) | ~192 | MIT | Still **v2.17.4**. |
| [Halloweedev/usagepal](https://github.com/Halloweedev/usagepal) | ~82 | MIT | Still **v0.7.77-beta.2** — OpenUsage menu-bar fork (quiet). |
| [Nanako0129/syrtis](https://github.com/Nanako0129/syrtis) | ~387 | MIT | **Was TokenBar** — still **v2.2.0** notarized DMG + Sand shoal. |
| [tddworks/ClaudeBar](https://github.com/tddworks/ClaudeBar) | ~1.5k | (check) | Still **v0.4.93** — today unreleased Kimi CLI probe (#325) + Codex RPC click-gated refresh (#322). |
| [deviffyy/OpenQuota](https://github.com/deviffyy/OpenQuota) | ~221 | MIT | Cross-platform limits/spend tracker (quiet). |
| [Dicklesworthstone/coding_agent_usage_tracker](https://github.com/Dicklesworthstone/coding_agent_usage_tracker) | ~87 | NOASSERTION | Single CLI for remaining quotas across Codex/Claude/Gemini/Cursor/Copilot. |
| [sculptdotfun/viberank](https://github.com/sculptdotfun/viberank) | ~117 | MIT | Public AI-coding usage leaderboard on ccusage data. |
| [cobra91/better-ccusage](https://github.com/cobra91/better-ccusage) | ~90 | MIT | Faster multi-provider JSONL analyzer. |
| [851-labs/tokenmaxxing](https://github.com/851-labs/tokenmaxxing) | ~78 | MIT | **cli-v0.7.1** + unreleased Windows npx.cmd fix (#130); local CLI on ccusage that syncs usage socially. |
| [Nihondo/AgentLimits](https://github.com/Nihondo/AgentLimits) | ~63 | MIT | macOS widgets for Codex/Claude limits + ccusage heatmap. |
| [majiayu000/quotabar](https://github.com/majiayu000/quotabar) | ~53 | MIT | **v0.5.9** — AppImage launcher/icon fixes after v0.5.7 Codex credit formatting. |
| [abhiunix/AgentHarbor](https://github.com/abhiunix/AgentHarbor) | ~12 | MIT | Native multi-agent rate-limit + session usage app. |
| [semantic-craft/iOS-vibebuddy](https://github.com/semantic-craft/iOS-vibebuddy) | ~91 | MIT | Still **v1.3.37** — Claude allowance via background CLI probe. |
| [zhnd/lumo](https://github.com/zhnd/lumo) | ~147 | MIT | Local-first Claude Code usage/cost/session dashboard. |
| [fschmutz/claude-usage-panel](https://github.com/fschmutz/claude-usage-panel) | ~6 | MIT | Still **v2.2.0** — unreleased honest readings / Codex vault. |
| [sylearn/AIUsage](https://github.com/sylearn/AIUsage) | ~693 | Apache-2.0 | Desktop multi-tool usage peer — **today v0.15.22**. |
| [BerriAI/litellm](https://github.com/BerriAI/litellm) | ~60.0k | NOASSERTION | Gateway spend/pricing — adjacent (**v1.103.2**). |
| [langfuse/langfuse](https://github.com/langfuse/langfuse) | ~35.3k | NOASSERTION | **v4.48.0** — OSS LLM observability + OTel — adjacent. |
| [Helicone/helicone](https://github.com/Helicone/helicone) | ~6.2k | Apache-2.0 | Quiet since 16 Sep (**15th day**). OSS LLM observability proxy. Adjacent. |

**Also watch**
- [ClaudeCodeUsage/ClaudeCodeUsage](https://github.com/ClaudeCodeUsage/ClaudeCodeUsage) (~66★, MIT) — VS Code status-bar Claude Code usage/cost.
- [itvincent-git/codex-usage-desktop](https://github.com/itvincent-git/codex-usage-desktop) (~44★, MIT) — **app-v3.9.0** — monthly five-hour/weekly quota consumption estimates.
- [saeedkolivand/claude-usage-mac](https://github.com/saeedkolivand/claude-usage-mac) (~15★) — Claude Code macOS menu bar + desktop widget.
- [ofershap/cursor-usage-tracker](https://github.com/ofershap/cursor-usage-tracker) (~33★, MIT) — Cursor Enterprise FinOps, not personal autodetection.
- [DeepAgentLabs/agenticlens](https://github.com/DeepAgentLabs/agenticlens) (~59★, MIT) — agent token/cost/latency profiler (quiet).
- Micro/zero-★ Claude wrappers keep spawning daily — noise floor, not category peers.

**Build takeaways:** Stay ccusage-compatible (**v20.0.26** holds; Thu pricing-snapshot only). Differentiate on live quotas + auto-detect + UX vs openusage.sh (**still no new tag** after Tue evening resume; **34/35/36** provider copy drift). Today's sharpest signals: **LiteLLM v1.103.2**, **QuotaBar v0.5.9**, **Langfuse v4.48.0**, **codex-usage-desktop app-v3.9.0**, **ClaudeBar Kimi/Codex-RPC fixes (unreleased)**, **pane ~27 ahead toward 0.4.56 widget**, **AIUsage v0.15.22**. Keep Langfuse/LiteLLM complementary; Helicone remains maintenance-mode adjacent (day 15).

Sources: GitHub API 2026-10-01 (Europe/London).
