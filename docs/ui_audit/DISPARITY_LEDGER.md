# Comprehensive UI Disparity Ledger & Audit Report
**Source of Truth**: Gold Standard Design Concepts (`docs/redesign_concepts/14_ceramic_swiss_studio.jpg`, `01_split_cockpit_view.jpg`, etc.)  
**Audited Target**: Live Web Dashboard (`internal/webserve`, 1920x1080 Viewport)  
**Status**: Visual Audit Complete (Targets 1–10 Complete)  
**Generated Artifacts**: `docs/ui_audit/annotated_*.png` (9 annotated visual defect maps)

---

## Executive Summary

An exhaustive visual inspection of the live web UI was conducted across all dashboard layouts, cockpits, global navigation, and search/filter states against the gold standard design mockups (`14_ceramic_swiss_studio.jpg`). Across 9 audit targets, 108 visual defects were identified and mapped with pixel-coordinate bounding boxes and issue tags.

These 108 findings consolidate into **8 core architectural disparities** between the current web rendering engine and the gold standard design:

1. **CSS Grid Track Collapse (1-Column Stack / Viewport Dead Void)**: Bento, Bars, and Dials views collapse into a single narrow ~359px column, leaving ~1520px of dead void on 1080p displays and pushing 3–4 providers completely off-screen.
2. **Internal TUI Error String Leakage**: Handling metrics without defined limits leaks raw ANSI/TUI text (`cannot render as bar or graph`) into web cards, tables, and strips.
3. **Severe Active Tab Contrast Failure**: White text (`#FFFFFF`) rendered on light-gray pills (`#D2D2D7`) produces a 1.09:1–1.5:1 contrast ratio, severely violating WCAG AA (4.5:1 min).
4. **Catastrophic Empty State Shell Destruction**: Filtering for a non-matching query hides the entire `#app` container, destroying header, search input, and dock, trapping the user without UI to clear the query.
5. **Theme Token Desynchronization & Brand Inconsistency**: Hardcoded "Ceramic Studio" header text clashes with the bottom dock's "Cupertino Light" switcher, and the header brand icon lacks the vector squircle SVG emblem.
6. **Missing Gauges & Blank Strips for Non-Percentage Quotas**: Currency and token budgets fail to calculate percentages, resulting in empty 0% progress capsules, missing radial dials, and completely empty 1880px card bodies.
7. **Stuttering & Redundant Formatting Artifacts**: Repeated labels (`5h 5h 100%`), duplicated reset strings, dangling punctuation (`Next reset:`, `· $2,028 remaining`, solitary `.`), and concatenated badge text (`• codex-cliUsage`).
8. **Cockpit Geometry & Polish Defects**: Full-width 1566px accent lines bleeding across panels, unstyled raw 11px Unicode glyphs for refresh buttons, and static non-interactive collapsible chevrons.

---

## Consolidated Disparity Ledger & Root Cause Analysis

### 1. CSS Grid Track Collapse & Viewport Overflow (Bento, Bars, Dials)
- **Visual Evidence**: `annotated_bento_view.png`, `annotated_bars_view.png`, `annotated_dials_view.png`
- **Disparity**: Instead of a modular 3-column responsive grid (as shown in `14_ceramic_swiss_studio.jpg`), cards are stacked in a single 359px wide column against the left margin. The remaining ~1520px of screen real estate is an empty white void. Consequently, 3 to 4 provider cards (Ollama, OpenCode, Pi) are pushed below `y=1080` without scrollbars or pagination, and bottom cards (`cursor-ide`) are bisected by the dock bar.
- **Root Cause**: `app.css` defines `.bento-tiles-grid` and `.board-grid` with rigid single-column or flex container behavior that fails to specify `display: grid; grid-template-columns: repeat(auto-fill, minmax(360px, 1fr)); gap: 16px;`.
- **Remediation**: Update `.bento-tiles-grid`, `.board-bars`, and `.board-dials` in `internal/webserve/ui/app.css` to use responsive multi-column grid tracks with appropriate auto-fill constraints.

### 2. Raw Error String Leakage (`cannot render as bar or graph`)
- **Visual Evidence**: `annotated_matrix_view.png`, `annotated_bento_view.png`, `annotated_bars_view.png`, `annotated_dials_view.png`, `annotated_strips_view.png`
- **Disparity**: Red error badges displaying `ERROR <metric>: cannot render as bar or graph` appear inside cards (e.g. OpenAI Codex CLI), table cells, and strip rows.
- **Root Cause**: `buildDetailSections()` in the TUI projection subsystem outputs an error string when a metric configured with a bar/gauge widget has no upper limit or series. This string is scraped into `WebDetailRow` and rendered in HTML.
- **Remediation**: In `internal/webserve/render.go`, filter or intercept strings matching `cannot render as bar or graph`. Render scalar metrics as clean text badges or key-value items rather than leaking internal diagnostic errors.

### 3. Active Tab Contrast Violation (WCAG AA)
- **Visual Evidence**: `annotated_global_shell.png`, `annotated_split_view.png`, `annotated_matrix_view.png`, etc.
- **Disparity**: Active layout navigation pills (`[Split]`, `[Matrix]`, `[Bento]`, `[Bars]`, `[Dials]`, `[Strips]`) render pure white text (`#FFFFFF`) against a `#D2D2D7` light grey active pill background. The contrast ratio is ~1.5:1 (well below the 4.5:1 WCAG AA minimum).
- **Root Cause**: `.layout-btn.active` in `app.css` sets background to `var(--tab-active-bg)` without setting `color: var(--tab-active-fg, #1a1a1c)` or uses an inverted text token.
- **Remediation**: In `internal/webserve/ui/app.css`, explicitly style `.layout-btn.active` with high-contrast dark text (`color: #1a1a1c; font-weight: 600;`).

### 4. Catastrophic Empty State Shell Destruction
- **Visual Evidence**: `annotated_filter_view.png`
- **Disparity**: Entering a search filter query with 0 matches (e.g. `/nomatch`) hides the entire `#app` container. The header, brand logo, search bar, and dock vanish completely, displaying an unstyled centered message stating "Start the telemetry daemon (agentusage daemon)". The user is completely trapped with no UI or input to clear the search.
- **Root Cause**: In `templates/app.html.tmpl`, `<div id="app" ... {{if not .HasData}} hidden{{end}}>` unconditionally hides `#app` whenever `.HasData` is false (which happens during empty filter results). In `templates/shell.html.tmpl`, `#empty-state` assumes empty data is always caused by daemon stoppage.
- **Remediation**:
  - Keep `#app` visible when `.Filtered` is true.
  - In `templates/app.html.tmpl`, render an in-panel empty state card with a prominent "Clear filter" button and preserved search bar.
  - Correct the empty state messaging when filters are active.

### 5. Theme Token Desynchronization & Brand Inconsistency
- **Visual Evidence**: `annotated_global_shell.png`
- **Disparity**: The header displays hardcoded text `Ceramic Studio` while the bottom dock displays `Cupertino Light · loopback` and a dropdown for `Cupertino Light`. Furthermore, the header brand icon is plain text `aU` inside a box rather than the SVG squircle emblem.
- **Root Cause**: `templates/app.html.tmpl` hardcodes `<span class="brand-studio">Ceramic Studio</span>` instead of referencing `.Env.ThemeTokens.Name` or a synchronized brand model, and `app.html.tmpl` uses `<span class="brand-mark">aU</span>` instead of an SVG emblem.
- **Remediation**: Synchronize theme token presentation across header and dock; ensure the header brand icon matches the gold standard squircle SVG emblem.

### 6. Missing Progress Gauges & Blank Strips for Non-Percentage Quotas
- **Visual Evidence**: `annotated_bars_view.png`, `annotated_strips_view.png`, `annotated_detail_cockpits.png`
- **Disparity**: Accounts with currency quotas (e.g. Gemini CLI `$8.50 / $20.00`) render no progress bar in Bars view, an empty 48px track in Left Navigation, and a completely empty 1880px wide white strip in Strips view. In Dials view, 3 of 4 cards render zero radial gauges.
- **Root Cause**: `render.go` only generates gauges if `l.Pct != nil` or `v.HasGauge == true`. When providers provide numeric limits in values (like `$8.50 / $20.00`) without explicit percentage fields, percentage calculation is skipped.
- **Remediation**: In `internal/webserve/render.go`, parse currency/token limits into calculated percentages so all linear bars, radial dials, and strip gauges render complete metrics.

### 7. Stuttering & Punctuation Formatting Artifacts
- **Visual Evidence**: `annotated_bars_view.png`, `annotated_strips_view.png`, `annotated_detail_cockpits.png`
- **Disparity**:
  - `5h 5h 100%`: Timeband pill duplicates the metric label.
  - `Resets in 3h59m`: Duplicated on both left and right columns of the same gauge row in Command Code.
  - `Refreshed: Last refreshed just now`: Double prefix.
  - `Next reset:`: Dangling colon with missing timestamp in cockpit hero subtitle.
  - `· $2,028 remaining`: Flex space-between pushes caption text to far right with dangling leading dot `·`.
  - `• codex-cliUsage`: Account title and badge concatenated without whitespace in Strips view.
- **Root Cause**: String formatting templates lack trimming of delimiters and omit duplicate checks between pills and titles.
- **Remediation**: Clean template helpers (`cleanResetCaption`, `stripDuplicatePillLabel`, `formatCaption`) in `render.go` and `templates/parts.html.tmpl`.

### 8. Cockpit Polish & Micro-Interactions
- **Visual Evidence**: `annotated_split_view.png`, `annotated_global_shell.png`
- **Disparity**: The 1566px accent hairline under the hero bleeds across the entire screen; the cockpit refresh button is an unstyled raw Unicode `⟳` glyph; Model Burn has a static non-interactive chevron `▸` and an unscaled sparkline.
- **Root Cause**: CSS `.accent-line` has full width without max-width or flex boundary; `.btn-cockpit-refresh` lacks styling rules; Model Burn title lacks `<details>` or toggle state.
- **Remediation**: Apply scoped CSS rules in `app.css` for `.accent-line`, `.btn-cockpit-refresh`, and interactive `.burn-card` toggles.

---

## Annotated Target Mapping Matrix

| Target | Target View | Annotated Artifact | Defect Count | Primary Disparities Identified |
|---|---|---|:---:|---|
| **Target 1** | Split Cockpit (Default) | `docs/ui_audit/annotated_split_view.png` | 12 | Active tab contrast (1.5:1), unstyled 11px refresh glyph, accent line bleed, KPI caption dangling dot, model burn sparkline lack of baseline/units, static chevron |
| **Target 2** | Matrix HUD Grid | `docs/ui_audit/annotated_matrix_view.png` | 12 | `cannot render as bar or graph` error leakage, 6x redundant `<thead>` repetition, viewport truncation at bottom, empty quota placeholders, table drawer toggle column omission |
| **Target 3** | Bento Glance Grid | `docs/ui_audit/annotated_bento_view.png` | 12 | 1-column grid collapse (~1620px dead void), 3 providers pushed off-screen, bottom card bisected by dock, error string leakage, premature 64px text truncation, stuttering labels |
| **Target 4** | Bars Telemetry View | `docs/ui_audit/annotated_bars_view.png` | 12 | Single 359px narrow column (~1520px dead void), 4 providers pushed off-screen, active tab contrast failure, missing linear gauge on currency quota, duplicate reset captions |
| **Target 5** | Dials & Gauges View | `docs/ui_audit/annotated_dials_view.png` | 12 | Single column collapse, 3 of 4 cards omit radial dials entirely, error string leakage, bottom card dock bisection, omitted card footers and drilldown links |
| **Target 6** | Strips Wall View | `docs/ui_audit/annotated_strips_view.png` | 12 | Account & badge concatenation (`• codex-cliUsage`), 1880px empty card body in Gemini CLI, error string leakage, stretched horizontal distortion, orphan 'In' metric label |
| **Target 7** | Search / Empty State | `docs/ui_audit/annotated_filter_view.png` | 12 | Catastrophic `#app` destruction on no-match (user trapped), misleading daemon stopped hint, passive filter text lacking clear button, empty 0% mini-gauge in left nav |
| **Target 8** | Detail Alternate Cockpits | `docs/ui_audit/annotated_detail_cockpits.png` | 12 | Dangling empty colon `Next reset:`, misleading 0.0% green zero-quota KPI rendering, solitary period `.` artifact, multi-reset caption clutter, full-width accent line bleed |
| **Target 9** | Global Shell & Dock | `docs/ui_audit/annotated_global_shell.png` | 12 | Plain text `aU` instead of SVG squircle emblem, active tab contrast failure, theme switcher desync (Ceramic vs Cupertino), dock crowding, dock theme dropdown redundancy |
| **Total** | **Full Application Suite** | **9 Annotated Artifacts** | **108** | **8 Core Deduplicated Root Causes** |
