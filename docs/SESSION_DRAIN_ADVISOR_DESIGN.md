# Session Drain Advisor Design

Date: 2026-09-03
Status: Proposed
Author: Cloud Agent

## 0. Pre-Design Quiz Answers

Background-agent intake: the request was to plan a toggleable suggestion for which AI to use, with the goal of draining one 5-hour session fully before switching to the next. Answers below are locked from that request plus codebase exploration; they are not implementation.

1. **Problem solved:** Users with several subscription coding tools (Claude Code, Cursor, Codex, Copilot, Antigravity, OpenCode, Command Code, …) hop between them and open overlapping 5-hour rate-limit windows. Unused quota in an open window expires. The dashboard should tell them which AI to stay on until that 5h session is drained, then which one to use next.
2. **Beneficiaries:** Power users running multiple paid coding agents. Secondary: tmux / web dashboard users who want the same glanceable hint.
3. **Affected subsystems:** `core` (pure advisor), `config` (persisted toggle), `TUI` (banner, USE badge, keybinding, settings), `webserve` (parity API + chrome), `tmux` (suggestion segment). No provider Fetch changes.
4. **Out of scope:** Auto-launching or switching the actual tool; model-tier advice (Opus vs Sonnet); changing poll/fetch; analytics screen; persisting a manual “pin this session”; reordering tiles automatically.
5. **Overlapping docs:** `tmux` active-tool detection (`internal/tmux/active.go`) answers “which tool is running now” via recency/process/priority. This feature answers “which tool should you use next for quota.” Gauge projection (`internal/tui/tiles_gauge.go`) and short-window classification (`internal/tui/cycle_reset.go`) already know 5h windows; reuse their rules, do not fork them.
6. **MVP:** Default-off toggle; pure ranking over current snapshots; TUI header/footer hint + USE badge on the recommended tile; persist like usage mode; web parity; tests for the ranking function.
7. **Public interfaces:** additive `dashboard.session_advisor` bool; `SaveDashboardSessionAdvisor`; TUI key `a`; web `POST /api/v1/session-advisor`; tmux `{suggest}` / `suggest` segment.
8. **Backward compatibility:** additive config, default off. Missing field = current behavior (no suggestion). No snapshot schema change.

## 1. Problem Statement

Subscription coding tools almost all expose a **rolling ~5-hour usage window** (Claude Code billing blocks, Cursor/Codex `rate_limit_*` with `Window: "5h"`, OpenCode `rolling_usage`, Command Code `five_hour_usage`, Antigravity `quota_*_5h`, Claude Code / Zai / Ollama `usage_five_hour`). Starting a second tool while the first window is still open burns two clocks at once. The dashboard already shows per-tile 5h gauges and reset countdowns, but it never **recommends a single tool to drain**.

Users want:

> Stay on one AI until its 5h session is used up, then go to the next one — and be able to turn that advice off.

## 2. Goals

1. Rank enabled accounts that have a 5h session window and pick exactly one **Use now** target.
2. Prefer finishing an already-open, not-yet-drained 5h session over starting a fresh window.
3. After that session is drained (or the tool is otherwise unusable this window), recommend the next eligible account in the user’s dashboard provider order.
4. Make the advisor **toggleable** from the TUI, settings, config, and web dashboard. Default **off**.
5. Share one pure function across TUI, web, and tmux so the suggestion cannot drift.

## 3. Non-Goals

1. Controlling or launching Claude/Cursor/Codex/etc.
2. Advising which **model** to pick inside a tool.
3. Auto-reordering dashboard tiles (highlight only).
4. Changing provider adapters, telemetry ingest, or daemon polling.
5. Treating pay-as-you-go credit APIs (OpenAI/Anthropic/OpenRouter keys) as 5h session tools.
6. A new rotation list in config. Dashboard provider order (Settings → Providers, Shift+J/K) is the rotation.
7. Persisting “I started this session on purpose” beyond what snapshots already show.

## 4. Impact Analysis

### Affected Subsystems

| Subsystem | Impact | Summary |
|-----------|--------|---------|
| core types | minor | New pure `AdviseSession` helper + 5h window detection shared by all UIs |
| config | minor | Additive `dashboard.session_advisor` bool + save helper |
| TUI | major | Toggle, header/footer copy, USE badge, settings row, help |
| webserve | minor | Envelope field, POST toggle, banner in web chrome |
| tmux | minor | `{suggest}` variable + `suggest` segment; Detect strategy unchanged in MVP |
| providers | none | Consume existing snapshot metrics/resets |
| detect / daemon / telemetry | none | Ranking is a view over already-fetched snapshots |
| CLI | none | No new cobra command |

### Existing Design / Code Overlap

- **`internal/tmux/active.go`**: recency-based “what is running.” Complementary. Do not reuse Detect as the advisor; a recently-focused Cursor session should not override “Claude 5h is 70% used, stay on Claude.”
- **`internal/tui/cycle_reset.go` `isShortWindowResetKey`**: too broad (1d, rpm, tpm). Advisor must detect **5h session windows only**.
- **`internal/tui/gauge.go` `gaugeWindowDuration`**: already maps `"5h"` / `"rolling-5h"` to `5 * time.Hour`. Export or move a shared predicate into `core`.
- **Usage mode (`u`)**: display remaining vs used. Advisor is independent; both can be on.
- **Hide-costs (`c`)**: per-account. Advisor is global.

## 5. Detailed Design

### 5.1 User-visible behavior

When the advisor is **off** (default): no extra chrome.

When **on**:

```
⚡ agentUsage  Dashboard  4●          🎯 Use Claude Code · 62% of 5h · 1h 54m left
─────────────────────────────────────────────────────────────────────────────
  ┃ claude_code          USE  5H LOW
  │ cursor                    OK
  │ codex                     OK
```

Footer (dashboard, idle):

```
 auto-refresh ⟳ 30s · a advisor on · p menu · u mode · r refresh · ? help
```

Tile / list badge for the recommended account: accent `USE` (or `USE NEXT` when the current 5h is drained and we are pointing at a fresh window). Existing status badges stay (`OK`, `5H LOW`, `5H LIMIT`, …). `USE` is an extra left-of-badge marker, not a replacement.

Header copy (truncate to width):

| Reason | Copy |
|--------|------|
| `stick` | `Use {name} · {used%} of 5h · {resetIn} left` |
| `switch` | `Switch to {name} · 5h drained on {prev}` |
| `start` | `Start {name} · no 5h window open` |
| `wait` | `All 5h sessions drained · next reset {soonest}` |
| `none` | hide banner (no eligible 5h tools) |

`{name}` is the account id (same as tile titles), not a marketing name.

### 5.2 Toggle

| Surface | Behavior |
|---------|----------|
| Key `a` | Toggle on dashboard (same layer as `u` / `w` / `t`). Persist immediately. |
| Settings → Providers | First row above the account list: `Session drain advisor  [on]/[off]`. Space/Enter on that row toggles. Cursor index 0 is the toggle; accounts start at 1. |
| Config | `"dashboard": { "session_advisor": true }` |
| Web | `POST /api/v1/session-advisor` with `{ "session_advisor": true }` or empty body to toggle, mirroring `/api/v1/usage-mode` |

Default: **false**. `omitempty` when false so existing `settings.json` files stay byte-stable.

Help overlay: add `{ "a", "Toggle 5h session advisor" }` next to the usage-mode binding.

### 5.3 Config schema

```go
type DashboardConfig struct {
    // ...existing fields...
    SessionAdvisor bool `json:"session_advisor,omitempty"`
}
```

```go
func SaveDashboardSessionAdvisor(enabled bool) error
func SaveDashboardSessionAdvisorTo(path string, enabled bool) error
```

`tui.Services` and `dashboardapp.Service` gain `SaveDashboardSessionAdvisor(enabled bool) error`. Update the two test fakes (`mockUsageModeService`, `fakeServices`).

Normalization: any JSON value other than boolean `true` loads as off (missing, `false`, garbage). Do not invent string enums.

### 5.4 Core algorithm (pure)

New file: `internal/core/session_advisor.go`. No TUI imports.

```go
type SessionPhase string

const (
    SessionPhaseIdle     SessionPhase = "idle"      // no open 5h window
    SessionPhaseDraining SessionPhase = "draining"  // window open, remaining > drain floor
    SessionPhaseDrained  SessionPhase = "drained"   // window open or limited, remaining ~ 0
    SessionPhaseBlocked  SessionPhase = "blocked"   // auth/error/weekly-monthly exhausted
    SessionPhaseNone     SessionPhase = "none"      // not a 5h session tool
)

type SessionAdviceReason string

const (
    AdviceStick  SessionAdviceReason = "stick"
    AdviceSwitch SessionAdviceReason = "switch"
    AdviceStart  SessionAdviceReason = "start"
    AdviceWait   SessionAdviceReason = "wait"
)

type SessionWindow struct {
    AccountID    string
    ProviderID   string
    Phase        SessionPhase
    UsedPct      float64        // 0–100, -1 if unknown
    RemainingPct float64
    ResetAt      time.Time
    MetricKey    string
}

type SessionAdvice struct {
    AccountID     string
    ProviderID    string
    Reason        SessionAdviceReason
    UsedPct       float64
    RemainingPct  float64
    ResetAt       time.Time
    PrevAccountID string // set on switch
    NextAccountID string // upcoming after current drains; empty on wait/none
}

func AdviseSession(snaps []UsageSnapshot, order []string, now time.Time) (SessionAdvice, bool)
```

`order` is the dashboard account order (enabled accounts only). Snapshots not in `order` are ignored. Accounts in `order` with no snapshot yet are `idle` if the provider is known to have 5h windows, else `none`.

#### Detecting a 5h session window

A metric is a 5h **quota** window when **all** of:

1. It is not a cost/count metric (`*_cost`, `*_msgs`, `*_tokens` with unit tokens, `5h_block_cost`).
2. Either:
   - `strings.EqualFold(metric.Window, "5h")` or `"rolling-5h"` or contains `"5h block"`, **or**
   - the key is in the known set: `usage_five_hour`, `five_hour_usage`, `rolling_usage`, `billing_block` (reset-only fallback), `quota_gemini_5h`, `quota_claude_5h`, `quota_3p_5h`, `quota_opus_sonnet_5h`, `rate_limit_primary`, `rate_limit_secondary` **when** their `Window` is 5h.
3. `MetricUsedPercent(key, metric) >= 0` **or** a matching `Resets[key]` / `Resets["billing_block"]` exists.

Export `IsFiveHourSessionMetric(key string, m Metric) bool` so TUI gauges and the advisor cannot disagree later.

**Multi-pool providers (Antigravity):** fold every 5h quota on that snapshot into one `SessionWindow`:

- Phase `blocked` if weekly/monthly is exhausted (`monthlyQuotaExhausted` equivalent, plus Antigravity weekly remaining ≤ 0).
- Else phase `drained` if **every** started 5h pool is drained.
- Else phase `draining` if **any** 5h pool is started and not drained. `UsedPct` = max used% among draining pools (the one to finish).
- Else `idle`.

This keeps advice at **account/tile** granularity: “keep using Antigravity until its open 5h pools are empty,” not “switch to Cursor because Gemini 5h is empty while Claude 5h is still unused.”

#### Phase rules

Let `used = MetricUsedPercent`, `drainFloor = 99.5` (constant, not user-tunable in MVP).

| Phase | Condition |
|-------|-----------|
| `blocked` | `Status` is Auth/Error, **or** weekly/monthly remaining ≤ 0, **or** provider status Limited **and** the exhausted window is weekly/monthly (existing `SnapshotStatusBadge` priority: weekly/monthly beats 5h) |
| `drained` | not blocked, and (5h remaining ≤ 0 **or** used ≥ drainFloor **or** 5H LIMIT status) |
| `draining` | not blocked/drained, and (used > 0.5 **or** reset is in the future with elapsed > 1 minute). Elapsed = `5h - time.Until(resetAt)` when reset exists |
| `idle` | has 5h metrics/resets but window not started (used ≈ 0 and no in-progress reset) |
| `none` | no 5h quota metric |

Credits-only API providers (openai, anthropic, groq, …) stay `none` unless they actually emit a 5h quota window.

#### Ranking

Walk enabled accounts in `order`. Partition by phase.

```
if any draining:
    pick draining with highest UsedPct
    tie-break: soonest ResetAt
    tie-break: earlier in order
    reason = stick
    next = first eligible after this one (draining else idle else none)
else if any idle:
    pick first idle in order
    reason = start   // or switch if some drained exist (PrevAccountID = that drained id)
else if any drained:
    reason = wait    // all 5h tools are empty this window
    AccountID empty / bool=false?  still return advice with Reason wait and soonest ResetAt
else:
    return ok=false  // nothing to say
```

**Switch vs start:** if the chosen account is `idle` and at least one other account is `drained`, reason is `switch` and `PrevAccountID` is the drained account with the soonest remaining reset (the one we just finished). If nothing was drained, reason is `start`.

**Why highest used% among draining, not “currently focused tile”:** the whole point is to refuse hopping. If Claude is 70% into its 5h and Cursor is 10% into a second window the user already opened, keep draining Claude. When Claude hits `drained`, Cursor (already open) ranks above an idle Codex because draining beats idle.

**Blocked accounts** never win. They are skipped for `next` as well.

### 5.5 TUI wiring

1. `Model.sessionAdvisor bool`, loaded from `DashboardConfig.SessionAdvisor` in `NewModel` (same place as `usageMode`).
2. `toggleSessionAdvisor()` + `persistDashboardSessionAdvisorCmd()` mirroring `toggleUsageMode`.
3. `handleKey`: `case "a":` on dashboard when filter/settings/help are closed.
4. `renderHeader`: when on and advice exists, put the 🎯 sentence in the right-side `info` slot (today: `⊞ N providers`). If width is tight, drop the provider count first, then truncate the advice with `…`.
5. `renderTile` / `renderListItemWithGroup`: if advice.AccountID matches, prefix the right-side badges with accent `USE` (`switch`/`start` use `NEXT`).
6. Footer hint includes `a advisor on/off`.
7. Settings Providers: inject a toggle row. Shift+J/K reorder must skip the toggle row (still only reorder accounts).
8. Invalidate render caches on toggle (same as usage mode).

Do **not** auto-move the list cursor onto the recommended tile. Suggestion is informational; keyboard focus stays where the user left it.

### 5.6 Web dashboard parity

Follow the usage-mode pattern in `internal/webserve`:

- `Envelope.SessionAdvisor bool`
- `Envelope.Advice *SessionAdvice` (JSON, omitempty when off or none)
- `POST /api/v1/session-advisor` with CSRF origin checks copied from usage-mode
- `collector.setSessionAdvisor` reprojects the cache
- Web chrome: banner above the card grid; recommended card gets a `USE` chip
- `app.js` key `a` calls the POST (alongside `u` → `cycleUsageMode`)

Demo snapshots in `internal/webserve/demo.go` should include one draining 5h tool and one idle 5h tool so the banner is visible in `make demo` / serve demo.

### 5.7 tmux

MVP: **do not** change `Detect()`. Recency remains “what is running.”

Add:

- Context field `Advice core.SessionAdvice` (zero when advisor off or no 5h tools)
- Variable `{suggest}` → `claude_code` (account id) or empty
- Variable `{suggest_reason}` → `stick` / `switch` / …
- Segment `suggest` → `use claude_code 62%` (compact, for status-right)

`BuildContext` reads `settings.json` dashboard flag (already loaded via config) and runs `AdviseSession` on `AllSnapshots`. If the flag is off, leave Advice zero so templates that reference `{suggest}` emit empty (existing tmux anti-flicker empty-segment behavior).

Follow-up (not MVP): optional Detect strategy `advisor` that pins the suggested provider when the flag is on. Call that out so we do not silently change status-bar identity in v1.

### 5.8 Constants

```go
const (
    SessionWindowDuration = 5 * time.Hour
    SessionStartUsedPct   = 0.5
    SessionDrainUsedPct   = 99.5
    SessionElapsedFloor   = time.Minute
)
```

Not user-facing in MVP. Weekly/monthly still win over 5h via existing status inference.

## 6. Alternatives Considered

### Persist a sticky “current session” account id
Would survive a refresh where used% flickers to 0. Rejected for MVP: extra config state, and snapshots already encode whether a window is open. Revisit if we see false `start` flips on poll gaps.

### Reorder tiles so the recommended tool is first
Noisy and fights the user’s Shift+J/K order (which is also the rotation). Badge + banner are enough.

### Use tmux Detect recency as the stick target
That optimizes for “whatever you touched last,” which is the opposite of drain-then-switch when the user already hopped.

### Include credit APIs as overflow after all 5h tools are drained
Useful later (`AdviceWait` could become `use openrouter`). Out of scope: the request is specifically 5h session draining.

### Per-provider advisor enable list
Overkill. Disabled dashboard accounts are already out of `order`.

### Key `d` for “drain”
`d` is unused but less mnemonic than `a` (advisor), and `a` is free in `handleKey` / `handleListKey`.

## 7. Risks

| Risk | Mitigation |
|------|------------|
| Inconsistent 5h key names across 37 providers | Window-string detection + small known-key allowlist; table test per fixture snapshot |
| Antigravity multi-pool false switch | Fold pools per account (5.4) |
| Header overflow on narrow terminals | Truncate advice; drop `⊞ N providers` first |
| Settings cursor off-by-one after adding toggle row | Tests for reorder + Space on row 0 |
| Services interface churn | Only two test fakes + `dashboardapp.Service` |
| Tmux users think `{tool}` will follow the advisor | Document Detect vs Advise; leave Detect alone |
| Polling used%=0 while a block is still open | Treat in-progress `ResetAt` (elapsed > 1m) as draining even if used is 0 |

## 8. Test plan

All new logic is table-driven `testing` tests. No mocks beyond existing TUI service fakes.

### `internal/core/session_advisor_test.go`

| Case | Expect |
|------|--------|
| One draining, one idle | stick on draining |
| Two draining, 70% vs 20% | stick on 70% |
| Two draining, same used%, sooner reset | stick on sooner reset |
| Draining + drained | stick on draining |
| All idle | start first in order |
| One drained, rest idle | switch to first idle, Prev=drained |
| All drained | wait, ok=true, ResetAt=soonest |
| Only credits APIs | ok=false |
| Weekly exhausted + healthy 5h | blocked, skipped |
| Auth snapshot | blocked, skipped |
| Antigravity Gemini 40% / Claude 0% | draining (stay) |
| Antigravity both 5h empty, weekly OK | drained |
| Disabled account in snaps but not in order | ignored |
| Empty order | ok=false |
| `rate_limit_primary` Window 5h (Codex/Cursor shape) | counted |
| `5h_block_cost` only, no quota % | `none` (cost is not a session) |
| used 0, reset in 3h50m (elapsed 10m) | draining |
| used 0, no reset | idle |

### Config

- Default load: `SessionAdvisor == false`
- `SaveDashboardSessionAdvisorTo` round-trip true/false
- Unknown JSON: off

### TUI

- `a` toggles and persists (clone `TestUsageModeToggle_KeybindingAndPersistence`)
- Header contains `Use ` when on + draining fixture
- Header omits advice when off
- Recommended tile/list row contains `USE`
- Help overlay lists `a`
- Settings Providers row 0 toggles flag; Shift+J on account 1 does not move the toggle row

### Web

- GET envelope includes `session_advisor` + `advice`
- POST toggles and reprojects
- CSRF origin rejection (copy usage-mode test)

### tmux

- `{suggest}` empty when flag off
- `{suggest}` account id when on
- `suggest` segment compact string

## 9. Implementation tasks

### Task 1: Core advisor
Files: `internal/core/session_advisor.go`, `internal/core/session_advisor_test.go`
Depends on: none
Description: `IsFiveHourSessionMetric`, phase classification, `AdviseSession`. No UI.

### Task 2: Config persistence
Files: `internal/config/config.go`, `internal/config/config_test.go`, `configs/example_settings.json`
Depends on: none
Description: field, normalize, `SaveDashboardSessionAdvisor[To]`, default false.

### Task 3: Dashboard service + TUI toggle
Files: `internal/dashboardapp/service.go`, `internal/tui/model.go`, `internal/tui/model_input.go`, `internal/tui/model_commands.go`, `internal/tui/usage_mode_test.go` (fake), `internal/tui/telemetry_mapping_input_test.go` (fake), new `session_advisor_toggle_test.go`
Depends on: Task 2
Description: Services method, model flag, `a` key, persist cmd.

### Task 4: TUI chrome
Files: `internal/tui/model_view.go`, `internal/tui/tiles.go`, `internal/tui/model_panels.go`, `internal/tui/help.go`, `internal/tui/settings_modal.go`, `internal/tui/settings_modal_sections.go`, `internal/tui/settings_modal_input.go`
Depends on: Tasks 1, 3
Description: header/footer, USE/NEXT badge, settings row, help. Tests for header, badge, settings cursor.

### Task 5: Web parity
Files: `internal/webserve/types.go`, `internal/webserve/server.go`, `internal/webserve/collect.go`, `internal/webserve/projection.go`, `internal/webserve/ui/app.js`, `internal/webserve/demo.go`, tests
Depends on: Tasks 1, 2
Description: envelope, POST, CSRF, demo fixture, key handler.

### Task 6: tmux segment
Files: `internal/tmux/context.go`, `internal/tmux/segments.go`, `internal/tmux/formatter.go` (if alias map), tests, `docs/TMUX_RUNBOOK.md` (variables table)
Depends on: Tasks 1, 2
Description: `{suggest}`, `{suggest_reason}`, `suggest` segment. Detect unchanged.

### Task 7: Docs
Files: `README.md` (keybindings), `docs/SESSION_DRAIN_ADVISOR_DESIGN.md` (this file, status → Accepted after review)
Depends on: Tasks 3–6
Description: document `a`, config key, that provider order is the rotation.

### Dependency graph

- Tasks 1 and 2: parallel
- Task 3 depends on 2
- Task 4 depends on 1 and 3
- Task 5 depends on 1 and 2 (parallel with 3/4)
- Task 6 depends on 1 and 2 (parallel with 5)
- Task 7 last

## 10. Acceptance

1. Fresh config: no banner, no USE badge, `a` turns both on and writes `session_advisor: true`.
2. Fixture: Claude 5h 62% remaining 38%, Cursor 5h unused → banner `Use claude_code`, Claude tile `USE`.
3. Same fixture after Claude 5h used 100% → banner `Switch to cursor`, Cursor `NEXT`.
4. `a` again restores previous chrome.
5. Provider with only `today_api_cost` never selected.
6. `go test ./internal/core/ ./internal/config/ ./internal/tui/ ./internal/webserve/ ./internal/tmux/` and `make vet` pass.

## 11. Follow-ups (explicitly not this change)

1. Detect strategy `advisor` so tmux `{tool}` follows the suggestion.
2. After all 5h tools wait, suggest a credits API as overflow.
3. User-tunable drain floor / session length.
4. Persist sticky account if poll flicker proves real.
5. Per-model-family advice inside Antigravity.
