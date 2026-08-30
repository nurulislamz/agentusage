# Local Web Dashboard (`agentusage serve`) Design

Date: 2026-08-30
Status: Implemented (native TUI-data chrome)
Author: cursor-agent

## 1. Problem Statement

agentUsage's live usage view is terminal-first. The browser dashboard must show
**the same data** as the Bubble Tea TUI so operators can open a tab without
learning a second information model. Rendering is native HTML/CSS (Deep Space
tokens, CSS gauges), not a painted terminal frame.

## 2. Goals

1. `agentusage serve` starts a local HTTP server with a self-contained dashboard.
2. Collection reuses the same path the TUI uses: prefer `daemon.ViewRuntime`
   (daemon socket / read-model), fall back to `export.Collect`.
3. The browser UI is a **split view** (header, navigator, detail pane, footer)
   using live theme tokens and OpenDesign `agentusage` DESIGN.md.
4. Detail cards are projected from the same Go section builders as the TUI
   (`buildDetailSections`) into typed JSON (`detail_cards`), not ANSI frames.
5. Default bind is loopback (`127.0.0.1:8080`) with the hub-style public-bind guard.

## 3. Non-Goals

1. React/Node frontend toolchain.
2. Changing TUI `View()` output (except shared helper extracts).
3. Multi-user accounts, TLS, or hosted SaaS.
4. Mutating settings / credentials from the browser.
5. Re-implementing analytics.

## 4. Architecture

```
agentusage serve
  → webserve.Server
      GET /api/v1/snapshots  → collector
           → daemon.ViewRuntime.ReadWithFallbackForWindow  (preferred)
           → export.Collect                                 (fallback)
           → tui.WebProjector (badges, gauges, detail_cards)
      GET /                 → embedded ui/ (HTML/CSS/JS)
```

`AccountView` carries structured nav fields plus `detail_cards` (heading / gauge /
timer / text rows). `frame_html` is unused by the SPA.

## 5. UI approach

Native chrome. Information and section order match the TUI. Visual treatment
follows `docs/DESIGN.md` / `~/open-design/design-systems/agentusage/`.

## 6. Security

- Default loopback bind.
- Non-loopback requires `AGENTUSAGE_SERVE_TOKEN` or `--allow-public`.
- Bearer auth on `/api/v1/*` when token set; `/healthz` stays open.
- Snapshot `Raw` maps stripped before JSON.

## 7. CLI

```
agentusage serve [--listen ADDR] [--source auto|direct|daemon] [--demo] [--open|--no-open] [--allow-public]
```
