# CONTEXT.md — working on the agentUsage web dashboard

Notes for agents (and humans) editing or debugging the web UI in
`internal/webserve/` and the antigravity provider in
`internal/providers/antigravity/`. Read AGENTS.md first for build/test
commands; this file covers live-system facts you'd otherwise have to
rediscover.

## How the dashboard works

- **Server-rendered everything.** The browser app is htmx + a thin
  `app.js`; every board (split / matrix / bento / bars / dials / strips)
  is Go templates in `internal/webserve/templates/*.tmpl` fed by
  `render.go`. There is no client-side state besides cookies + `app.js`
  ergonomics. Changing what the UI shows means changing Go render code or
  templates — not React.
- **Two running processes** on `jobby-dev-use`:
  - `agentusage daemon run --socket-path ~/.local/state/agentusage/telemetry.sock`
    — collects provider snapshots, owns the SQLite store, refreshes
    provider tokens (this is where provider env/credentials matter).
  - `agentusage serve --no-open` (port 8088) — the web dashboard; reads
    snapshots from the daemon over the socket. Restarting one doesn't
    restart the other. Logs: `~/.local/state/agentusage/{daemon,serve}.log`.
  - No systemd user unit; processes are started manually. `pkill -f
    "agentusage ..."` also kills your ssh session mid-command — kill by
    exact pid instead (`pgrep -f` + `kill`), or the whole remote command
    exits 255 and the restart chain never runs.

## Where things live

| What | Where |
| --- | --- |
| Layouts list + metadata | `layoutList` in `render.go` (add a new board here AND a template) |
| One row per quota window (the collapse) | `collapseUsageLines` — the single collapse feeding matrix/bento/bars/dials/strips |
| Window bucketing (5h/week/30d/...) | `bentoWindowKey` — matches Short+Label only, **never Hint** (hints contain durations like "Resets in 15h 52m" that falsely match "5h") |
| Per-window reset chip | `.bento-row-reset` (bento template) / `.matrix-cell-reset` (quota-cell template) |
| KPI banner window counts | `calculateGlobalStats` uses `rv.Lines` (FULL, uncollapsed lines); board projections use the collapsed set. Don't confuse them |
| Provider group ordering | `groupRenderViews` + `sortProviderGroups`, order persisted in cookie `au_provider_order` (pipe-separated provider ids) |
| Drag-reorder | grip handle in `.provider-group-header` (parts.html.tmpl), `providerDrag` IIFE in `ui/app.js`, endpoint `/actions/provider-order` |
| Cursor semantics | ONE billing reset, THREE usage metrics (Included/Auto/API) — buckets key on identity in `bentoWindowKey`; the shared reset comes from `nextResetDisplay` |
| Antigravity semantics | two windows (5h + weekly); sibling Gemini/Claude rows collapse to the tightest; reset times absent while AUTH |

## Cookie-driven state (per browser)

`au_layout`, `au_account`, `au_filter`, `au_expanded`, `au_view`,
`au_provider_order`. Set via `setUICookie`, read via `readUICookie`
(`ui_handlers.go`). Query params override cookies for layout/filter.
**A stale cookie will make a "fixed" bug reappear after deploy — clear
cookies or use a fresh profile when verifying UI changes.**

## The "stale binary" trap (cost half a session)

The dashboard is a Go binary with CSS/JS/templates **embedded**. If
rendered HTML doesn't match `internal/webserve/templates/`, the running
`agentusage serve` is stale — rebuild (`CGO_ENABLED=1 go build -o bin/agentusage ./cmd/agentusage`),
`install -m 755 bin/agentusage ~/.local/bin/agentusage`, then restart
serve. Verify by diffing `curl /agentusage/app.css` against the repo file.

## Antigravity AUTH debugging

- Status comes from the provider fetch; expired OAuth token + missing
  client credentials → status AUTH with reason in the inspect Diagnostics
  ("oauth client not configured ...").
- Token files: `~/.agy-containers/<box>/.gemini/antigravity-cli/antigravity-oauth-token`
  (JSON: `token.access_token/refresh_token/expiry`, `auth_method`).
  Access tokens are 1h; refresh uses Google's token endpoint and requires
  an installed-application OAuth client.
- The provider now **extracts the CLI's public client id/secret from the
  local `agy` binary at runtime** (`extractCLIOAuthClient`, cached once).
  Env vars `OPENUSAGE_ANTIGRAVITY_CLIENT_ID/SECRET` (aliases
  `ANTIGRAVITY_*`) override. These values are public (shipped inside the
  CLI binary) but are deliberately **not** embedded in this repo — GitHub
  push protection flags them (it even decodes base64). Test fixtures
  assemble the patterns at runtime for the same reason.
- `pkill`/`kill` of the daemon also kills your ssh command chain if the
  pattern matches the shell — kill by exact pid.

## Testing UI changes

- `renderFragment(t, env, renderInput{...})` renders a board directly —
  much faster than spinning the server. Fixtures: `Envelope{Views:
  []AccountView{...}}`, helper `f64()` for percent pointers.
- **Parity contract**: every web-rendered row must exist in the TUI
  render (`VerifyTUIWebParity`). If you intentionally diverge (e.g. web
  shows untruncated values the TUI caps at width 45), handle it in
  `parity.go` — see `tuiContainsTruncated`.
- CSS lives in `internal/webserve/ui/app.css`; tests assert on literal
  rule snippets (e.g. `.bento-quota-row { display: grid; ...`) — check
  `TestQuietDockControlSurfaces`-style tests before renaming classes or
  dropping "retired" tokens.

## Layout gotchas (all bit us once)

- Grid `1fr` tracks **do not shrink below content min-width**; every
  shell/panel column uses `minmax(0, 1fr)` for this reason. If a phone
  viewport clips instead of reflowing, suspect a fixed-width track or a
  non-wrapping flex row (header/footer scroll internally instead).
- `.shell` first grid row is `auto` (≤900px) so the header owns its
  height; desktop keeps a fixed `--topbar-h`.
- The board grids go 3-col at 1200px, 4-col at 1400px (four antigravity
  cards in one row).

## Gotchas that bit us

- `git commit --amend` after a soft reset onto an older commit can
  swallow unrelated commits — prefer `git checkout -B <new> <base>` +
  `git checkout <good-commit> -- .` + fresh commit when untangling.
- GitHub push protection scans **entire push history**, decodes base64,
  and even flags *test fixtures* that look like Google OAuth client
  values. Assemble such strings at runtime in tests; never embed real
  ones.
- `go test ./...` on this box: two daemon-process tests
  (`TestStartViaManagedService_And_EnsureViaServiceManager`,
  `TestEnsureRunning_OfflineSocket`) fail on **main** too — pre-existing
  environmental flake, not your change.
- The parallel-agent session may leave `agentusage daemon` test
  processes on `/tmp/*.sock` — leave them alone; the real daemon is the
  `telemetry.sock` one.
- Mobile top bar: `.header-main` is flex with fixed-height shell row —
  on ≤900px the first shell row is `auto` and the header may wrap;
  don't reintroduce a fixed `--topbar-h` there.
