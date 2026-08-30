# Local Web Dashboard (`openusage serve`) Design

Date: 2026-08-30
Status: Implemented (TUI parity rewrite)
Author: cursor-agent

## 1. Problem Statement

OpenUsage's live usage view is terminal-first. The browser dashboard must show
**the same data and the same visual language** as the Bubble Tea TUI so operators
can open a tab without learning a second UI.

## 2. Goals

1. `openusage serve` starts a local HTTP server with a self-contained dashboard.
2. Collection reuses the same path the TUI uses: prefer `daemon.ViewRuntime`
   (daemon socket / read-model), fall back to `export.Collect` (same as
   `openusage export --source auto`).
3. The browser UI is a **faithful port of the TUI split view** (header, navigator,
   detail pane, footer) using the active theme tokens.
4. Detail content is produced by the same Go renderer as the TUI
   (`tui.RenderDetailContent`) and converted ANSI → HTML.
5. Default bind is loopback (`127.0.0.1:8080`) with the hub-style public-bind guard.

## 3. Non-Goals

1. React/Node frontend toolchain.
2. Changing TUI code in the serve/web PR.
3. Multi-user accounts, TLS, or hosted SaaS.
4. Mutating settings / credentials from the browser.
5. Re-implementing analytics unless `experimental.analytics` is already on and
   exposed through the same snapshot envelope (future).

## 4. Architecture

```
openusage serve
  → webserve.Server
      GET /api/v1/snapshots  → collector
           → daemon.ViewRuntime.ReadWithFallbackForWindow  (preferred)
           → export.Collect                                 (fallback)
           → tui.WebProjector + tui.RenderDetailContent
           → ANSIToHTML → Envelope.views[].detail_html
      GET /                 → embedded ui/ (HTML/CSS/JS)
```

### Envelope (v1)

```go
type Envelope struct {
    SchemaVersion, GeneratedAt, OpenUsageVersion, Source, TimeWindow, Theme string
    RefreshIntervalSeconds int
    UsageMode string
    Catalog []CatalogEntry
    ThemeTokens ThemeTokens
    Views []AccountView      // TUI-projected + HTML fragments
    Snapshots []UsageSnapshot // sanitized (Raw stripped)
}
```

`AccountView` carries structured nav fields plus HTML fragments:

- `detail_html` — exact TUI detail pane
- `badge_html`, `icon_html`, `strip_html`, `summary_html` — navigator chrome
- `reset_hint` — cycle reset countdown text

### UI

Vanilla SPA embedded with `go:embed`. Layout mirrors TUI:

| Region | TUI | Web |
|--------|-----|-----|
| Header | brand + tabs + status counts + meta | same |
| Left (~⅓) | provider navigator with group headers, compact block strips | same |
| Right | `RenderDetailContent` cards | `detail_html` in monospace `<pre>` |
| Footer | key hints + theme/source | same shortcuts (↑↓ / r t) |

Theme CSS variables are injected from `theme_tokens` so the browser matches the
configured TUI theme (Deep Space default).

## 5. Security

- Default loopback bind.
- Non-loopback requires `OPENUSAGE_SERVE_TOKEN` or `--allow-public`.
- Bearer auth on `/api/v1/*` when token set; `/healthz` stays open.
- Snapshot `Raw` maps stripped before JSON.

## 6. CLI

```
openusage serve [--listen ADDR] [--source auto|direct|daemon] [--demo] [--open|--no-open] [--allow-public]
```

## 7. Compatibility

No TUI package changes required for this rewrite. Webserve consumes exported TUI
helpers only (`WebProjector`, `RenderDetailContent`, `RenderCompactBlockStrip`,
`SnapshotStatusBadge`, theme tokens).
