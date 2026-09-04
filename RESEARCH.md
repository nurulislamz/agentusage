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
