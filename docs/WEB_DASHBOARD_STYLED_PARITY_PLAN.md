# Web dashboard: TUI information, web styling

Date: 2026-08-30
Branch: `cursor/web-tui-parity-836c`
Status: Implemented (native TUI-data chrome)

## OpenDesign

Use `~/open-design` as the design engine for this work:

| Binding | Path |
|---|---|
| Brand contract | `~/open-design/design-systems/agentusage/DESIGN.md` + `tokens.css` |
| Repo pointer | `docs/DESIGN.md` |
| Skill | `frontend-design` — dense operational, not SaaS |
| Follow-up | `impeccable-design-polish` after HTML exists |
| Craft | `craft/anti-ai-slop.md`, `typography`, `color` |

Agents paste `tokens.css` `:root` into the serve UI. Do not use OpenDesign’s generic `dashboard` or `cursor` systems — those are the wrong product.

Anti-slop that applies here: no indigo, no purple→blue hero gradient, no Inter, no glass cards, no colored left-border on rounded cards, no invented metrics. TUI glyphs (`⚡●◐✗⏱⌛◈`) stay because they are product copy.

## Goal

Replace the current “paint the TUI frame into a `<pre>`” web UI with a native HTML/CSS split dashboard that:

1. Shows **the same information, labels, numbers, grouping, and section order** as the Bubble Tea TUI.
2. Keeps the **web visual treatment** in the agentUsage DESIGN.md (Deep Space tokens, JetBrains Mono, CSS gauges, colored section *titles*, selected nav rail).
3. Does **not** invent extra widgets (no fake Team Budget / model-burn charts unless that account’s TUI detail actually has them).

Reference screen: live TUI split view (header + grouped navigator + detail cards + footer), e.g. `opencode-mohammed` selected with Usage / Timers / Info.

## Why the current web UI fails

`internal/webserve` currently renders a full `tui.Model.View()` frame, converts ANSI → HTML, and dumps it into `.tui-frame`. That copies terminal constraints (box-drawing, character grid, left-third click hit-test) into the browser. Information is technically there; the surface is unusable.

The API already has the structured fields needed for a native chrome (`views[]`, `theme_tokens`, `gauge_percent`, `status_badge`, `resets`, `detail_sections`). We stop painting `frame_html` as the UI.

## Information contract (must match TUI)

Every region below is copied from the TUI, not redesigned.

### Header (one line + hairline)

Left:

- `⚡` + gradient `agentUsage`
- Status counts from visible accounts: `N●` OK (green), `N◐` near-limit (yellow), `N✗` limited/error (red)
- Unmapped telemetry phrase when present (`⚠ 5 unmapped`)

Right (dim):

- `⊞ {n} providers`
- usage mode label (`Remaining` or `Used`)
- time window (`3 Days`, `30d`, …)
- unmapped call-to-action when actionable (`5 telemetry sources need mapping`)

No Analytics tab unless `experimental.analytics` is on (same as TUI).

### Navigator (left ~1/3, max width capped like TUI)

- Group by `provider_id`, header `PROVIDER_ID (count)`
- Active group uses provider accent (`✦`); idle groups dim (`◈`)
- Each account row, same two-line shape as TUI:
  - Line 1: status icon + `account_id` + `SnapshotStatusBadge` (`OK`, `LOW`, `AUTH`, `MONTHLY LIMIT`, `WEEKLY LIMIT`, `5H LIMIT`, …)
  - Line 2: remaining-mode mini gauge + summary percent (`0.00%`, `93.65%`, …) + `Resets in {duration}`
  - AUTH / no-gauge rows show the TUI summary text instead (`Authentication required`)
- Selected row: provider-colored left rail + name in provider color
- Overflow: `▼ N more` / `▲ N more` (real clickable rows, not a fake dropdown)

### Detail compact header (selected account)

Exactly `renderDetailCompactHeader`:

1. Status icon + **account_id** (left) · `provider_id` · optional email/plan · status badge (right)
2. Summary (`0.00%`) · cycle reset schedule (`Monthly resets in 9 days · Weekly resets in 5h 45m`)
3. Status-colored underline (crit when MONTHLY LIMIT, warn when LOW, ok otherwise)

### Detail cards (same sections, same order)

Cards come from the same `buildDetailSections` pipeline the TUI uses. For the OpenCode Go screenshot that means:

**Usage** (yellow/peach accent) — `⚡ Usage`

- Subheading: `◈ OPENCODE GO SUBSCRIPTION` (only if those metrics exist)
- One gauge block per TUI gauge, remaining-mode labels:
  - `Five Hour Limit Remaining` → percent + `N.NN% remaining` + `⏱ Resets in …`
  - `Weekly Limit Remaining`
  - `Monthly Limit Remaining` (empty bar + crit caption when 0%)
- Gauge fill color uses the same remaining-mode traffic light as `gaugeColor` / `RenderGauge`

Other providers keep **their** TUI usage blocks (Antigravity model pools, Cursor team budget, Copilot chat/completions, …). The web does not substitute a generic “dashboard” layout.

**Timers** (maroon accent) — `⏰ Timers`

- One row per `snap.Resets` key, TUI label (`Month Spend`, `Monthly`, `Usage 5h`, `Weekly`)
- Value: `Jan 02 15:04 (in 9d 3h)` with the same urgency dot colors

**Info / Attributes** (blue) and every other TUI card (Spending, Models, Clients, Tools, MCP, Diagnostics, …) appear **only when the TUI would show them**, in the same order.

### Footer

TUI footer status line: `auto-refresh ⟳ {interval} · p menu · u mode · r refresh · R refresh all · ? help` (web: omit TUI-only settings modal keys that are not wired, keep filter `/`, `j/k`, `r`, `t` theme).

### Keyboard / mouse

| Input | Behavior |
|---|---|
| `j` / `↓` | next account |
| `k` / `↑` | previous account |
| click row | select that account |
| `/` | filter bar (TUI search) |
| `r` / `R` | refresh |
| `t` | cycle theme tokens |
| detail scroll | native overflow, not ANSI `▼ more below` |

## Visual rules (web chrome only)

Keep the approved styling. Do not restyle information.

- Palette: live `theme_tokens` (Deep Space default: base `#0C0E16`, accent `#7EB8F7`, …)
- Font: JetBrains Mono
- Dense layout, 1px `surface1` borders, 6px card radius max
- Gauges: rounded CSS tracks; remaining-mode fill length = remaining percent (empty = 0% remaining)
- Section cards: colored top border / title color from `sectionColor` (Usage yellow, Timers maroon, Info blue)
- Status pills: tinted fill + semantic text (`MONTHLY LIMIT` crit, `LOW` warn, `OK` green, `AUTH` peach)
- No ASCII `╭─╮`, no unicode block `█` strips, no Inter/sans dashboard look
- No extra hero stats that the TUI detail header does not show

## Architecture

```
GET /api/v1/snapshots
  → existing collector (daemon / export / demo)
  → WebProjector (order, badges, gauge %, resets)
  → NEW: structured detail cards from the same section builders
  → JSON Envelope.views[]

Browser
  → native header / grouped nav / footer
  → native detail: compact header + CSS cards
  → frame_html unused (keep in JSON as debug fallback only)
```

**Do not reimplement provider usage logic in JavaScript.** OpenCode vs Antigravity vs Cursor gauge sets live in Go (`internal/tui/detail_sections.go` and friends). The browser only paints typed rows.

### New structured payload (additive)

Extend `AccountView` / projector with typed detail, for example:

```go
type DetailCard struct {
    ID    string       `json:"id"`    // "Usage", "Timers", "Info"
    Title string       `json:"title"`
    Icon  string       `json:"icon"`
    Color string       `json:"color"` // hex from theme section color
    Rows  []DetailRow  `json:"rows"`
}

type DetailRow struct {
    Kind    string   `json:"kind"` // heading | gauge | timer | text | kv
    Label   string   `json:"label,omitempty"`
    Value   string   `json:"value,omitempty"`
    Hint    string   `json:"hint,omitempty"`
    Percent *float64 `json:"percent,omitempty"` // remaining or used, same as TUI mode
    Tone    string   `json:"tone,omitempty"`    // ok | warn | crit | dim
}
```

Populate `Rows` from the same functions that currently emit lipgloss lines (`buildOpenCodeGoUsageLines`, `renderTimersSection`, attributes, …). TUI `View()` stays ANSI; web reads JSON.

Compact header fields already exist or are cheap to add: `summary`, `status_badge`, `reset_hint`, cycle schedule string, `provider_id`, `account_id`.

Envelope header stats (ok/warn/err counts, unmapped count, usage mode, window label) should be first-class JSON so the browser does not reverse-engineer them from frames.

## Files

| File | Change |
|---|---|
| `~/open-design/design-systems/agentusage/DESIGN.md` | Brand contract (already written) |
| `~/open-design/design-systems/agentusage/tokens.css` | Token `:root` to paste into serve CSS |
| `docs/DESIGN.md` | Pointer into OpenDesign |
| `internal/tui/web_view.go` | Emit `DetailCards` + header chrome fields; keep existing `WebAccountView` fields |
| `internal/tui/detail_sections.go` (and usage helpers) | Share remaining%/label/reset extraction so TUI lines and JSON cards stay in lockstep. Prefer extracting helpers over copying switch logic |
| `internal/webserve/types.go` | Add `DetailCards` on `AccountView`; add envelope `ok_count` / `warn_count` / `err_count` / `unmapped_*` if not already present |
| `internal/webserve/projection.go` | Map new fields; stop requiring `FrameHTML` for display |
| `internal/webserve/ui/index.html` | Restore header / nav / main / footer shell (not a `<pre>` frame) |
| `internal/webserve/ui/app.css` | Approved web styling; theme CSS variables from tokens |
| `internal/webserve/ui/app.js` | Render grouped nav + detail cards; `j/k`/click; no frame swap |
| `internal/webserve/ui/*_test.go` / `projection_test.go` / `web_view` tests | Golden JSON for OpenCode-style usage+timers; nav grouping |
| `docs/WEB_SERVE_DESIGN.md` | Update UI approach: native chrome, TUI data, not byte-parity frames |

**Do not change** TUI rendering output (`Model.View`) except tiny helper extracts. **Do not** add React/Node.

## Implementation sequence

1. **Envelope chrome** — counts, window, usage mode, unmapped phrase. Tests against fixture snapshots.
2. **Navigator** — group + badge + mini gauge + reset hint from existing `AccountView`. Tests: order matches `WebProjector.OrderSnapshots`; AUTH row has no fake 0% bar if TUI shows text instead.
3. **Detail header** — compact two-line header + status underline.
4. **Structured Usage/Timers/Info cards** — start with OpenCode Go (the screenshot). Then Antigravity and Cursor so we do not overfit one provider.
5. **Generic card fallback** — any section still only available as stripped lines renders as a native card of `text` rows (correct info, slightly less pretty) until a typed row exists.
6. **CSS** — paste OpenDesign `tokens.css` `:root`, then gauges, pills, rails, section title colors. Theme tokens from the daemon override `--bg`/`--accent`/… at runtime.
7. **`impeccable-design-polish`** — audit hierarchy, contrast, focus, anti-slop. Do not change copy.
7. **Retire frame painting** — `frame_html` optional/debug; click handling uses real DOM rows.
8. **Manual check** — `agentusage serve` against the same machine as the TUI screenshot: same accounts, same badges, same percents, same reset copy.

## Tests (minimum)

- Projector: OpenCode snapshot with `rolling_usage` / `weekly_usage` / `monthly_usage_pct` + resets → Usage card has three gauges with those percents and TUI labels.
- Timers card keys/labels match `core.MetricLabel` / sorted reset keys.
- Status badge `MONTHLY LIMIT` when monthly remaining is 0 (existing `SnapshotStatusBadge` tests stay).
- Nav grouping: two OpenCode accounts under `OPENCODE (2)`.
- CSS/JS: keep current serve tests (`frame_test` either deleted or switched to asserting JSON cards). Do not snapshot ANSI frames as the UI contract.

## Out of scope

- Analytics screen
- Settings / API key modal in the browser
- Changing TUI layout or copy
- Tile-grid overview (README screenshot) — live TUI is split view
- HTML mockup files (not part of the product)

## Success

Opening `agentusage serve` next to `agentusage` on the same data, the web view shows the same accounts, badges, percents, reset sentences, and detail sections as the terminal. The only difference is rendering: CSS gauges and cards instead of unicode boxes.
