# Usage Simplification Implementation Plan

> **For agentic workers:** Use `executing-plans` to implement the bounded tasks below. A BXO coordinator may dispatch the task packets in `.ai/` according to `.coord/plan.yml`; review each dependency before starting its consumers. This document authorizes planning only, not autonomous execution or deployment.

**Goal:** Open agentUsage and see accurate usage using existing account logins, without status-line setup, agent prompts, or routine daemon administration.

**Architecture:** Retain the existing daemon and SQLite read model as the single collection owner. Route terminal, web, and machine-readable queries through that service. Antigravity refreshes OAuth credentials directly; other providers retain their supported auth methods, and local history supplements APIs without becoming a competing quota source.

**Tech Stack:** Go, Cobra, Bubble Tea, net/http, SQLite with CGO, existing OS service support, PlantUML.

**Status:** Proposed implementation handoff, based on the working tree inspected on 2026-09-05. No application behavior was changed for this plan. Default assumption: retain useful token/spend history. “Refresh-token-only” applies to Antigravity renewal, not every provider's authentication scheme.

## Decision and alternatives

Recommended: consolidate around the existing service, remove obsolete status-line acquisition, and make the default launch self-sufficient. This reuses the working read model, multi-account support, and history rather than replacing the application.

| Approach | Benefit | Cost / decision |
|---|---|---|
| Consolidate existing service (chosen) | One polling/auth owner; TUI, web and CLI agree; preserves history | Requires lifecycle and client cleanup |
| Delete daemon and SQLite; poll only while a view is open | Smallest quota-only product | Loses persistent history and shared polling across processes; a different product scope |
| Only remove Antigravity status-line code | Small, fast patch | Leaves direct CLI fetches, UI enrichment, and manual service setup; insufficient alone |

Retain supported providers, account IDs, windows, layouts, JSON compatibility, and useful token/spend data. Do not build a generic OAuth framework, new event bus, new database, or replacement frontend. Do not remove history collectors just because their comments mention status lines.

## Evidence from this checkout

| Area | Verified behavior | Consequence |
|---|---|---|
| `internal/providers/antigravity/antigravity.go` | `Fetch` already calls `retrieveUserQuotaSummary`; projects through `statusLinePayload`; auth retry calls `pingBoxForToken` | Finish the API conversion; remove misleading payload and process fallback |
| `internal/providers/antigravity/auth.go` | Renewal uses saved refresh token plus `OPENUSAGE_ANTIGRAVITY_CLIENT_ID` and `OPENUSAGE_ANTIGRAVITY_CLIENT_SECRET`; errors/missing tokens fall back to agent execution | Refresh token alone is not sufficient under the current contract |
| `scripts/boxes/agy-box`, `scripts/boxes/agent-box` | Inject `openusage antigravity statusline` and `openusage cursor statusline` | Remove obsolete registration and migrate only owned entries |
| `internal/integrations/definitions.go` | Registers Antigravity and Cursor status-line integrations | Remove install capability; retain narrowly scoped migration matching |
| `internal/detect/detect.go`, `internal/daemon/server_watch.go` | Discovery uses status files; watcher kicks quota polling for Antigravity status writes | Discover credentials/configuration and poll on time/explicit refresh instead |
| `cmd/agentusage/get.go` | `fetchAccountSnapshot` calls providers directly | Bypasses service ownership, caching, and read-model projection |
| `cmd/agentusage/dashboard.go`, `internal/webserve/collect.go` | Frontends construct provider enrichment paths | Can repeat quota fetches outside the daemon |
| `internal/daemon/runtime.go` | Methods named `ReadWithFallback*` reconnect to daemon; they do not directly poll providers | Existing user guide's direct fallback claim is inaccurate |
| `internal/daemon/process.go` | Supported service-manager platforms require an installed service; other platforms can spawn a process | Default launch still exposes installation details |
| `internal/providers/claude_code/claude_code.go` | Cookie auth is tried before local OAuth | Prefer saved OAuth where available; no blanket refresh support claim |
| `internal/providers/codex/live_usage.go` | Live usage reads an access token from `auth.json` | Do not assume refresh/rotation support is already implemented |
| `internal/providers/claude_code/usage_cache.go`, `internal/tmux/context.go` | Five-hour cache has a tmux reader and writer | Removing a status-line-named file blindly would break another consumer |
| `cmd/agentusage/main.go`, `main_test.go` | Root already exposes `get`, `list`, `detect`, `doctor`, `serve`, `daemon`; statusline commands are removed | Simplify current commands; do not resurrect obsolete routes |

BXO itself was not located in this checkout. The concrete box integration scope here is `scripts/boxes/` and `scripts/box.sh`. Any external BXO change needs its repository and a separate file plan; do not invent BXO commands or edit another project from this task.

## Target user experience and command contract

1. Sign in using the provider's normal tool once. agentUsage discovers local profiles and box credential paths without starting boxes or running inference.
2. Run `agentusage`. Show cached usage immediately when available, then update automatically. First use starts an ordinary local helper if no managed service is installed; it does not require a daemon install step.
3. Show remaining quota, reset time, and last successful update per account. Keep history/detail views for providers with suitable data. No data means unavailable, never an invented zero or unlimited quota.
4. Expired Antigravity access tokens renew directly over HTTP when refresh credentials are available. Missing/revoked credentials produce a sign-in action for the affected account; other accounts continue.
5. `r`, web refresh, and `get` use the same collection owner. Rendering, layout/window switches, `list`, and diagnostics do not independently poll provider APIs.
6. Closing the UI closes that client. The shared helper can remain available for history and other clients; do not kill it when another consumer may be using it. OS startup persistence remains an explicit advanced choice.

| Surface | Target behavior | Compatibility |
|---|---|---|
| `agentusage` | Main usage view; account setup and refresh in the UI | Keep existing navigation/layouts |
| `agentusage get <id>` | Script-friendly one-account result, via service; refresh when cached quota is stale | Preserve flags, field names, units, exit behavior and stdout purity |
| `agentusage list` | Shared account resolver, with availability; no live quota calls | Preserve machine output and account matching |
| `agentusage serve` | Optional browser presentation of the same snapshots | Keep existing flags for this change |
| `agentusage doctor` | One diagnostic entry point: credentials, accounts, helper, stale data | Add `--detect` for detailed discovery; read-only by default |
| `agentusage detect` | Hidden deprecated alias for the same discovery implementation | Keep old output/flags; diagnostics to stderr only |
| `agentusage daemon ...` | Advanced grouped help for service administration and existing hook ingestion | Keep names/flags for scripts, service files and installed hooks |
| Status-line commands | No supported usage collection dependency | Already removed; clean their installers instead of adding stubs |
| Demo / parity verification | Development workflows, not onboarding steps | Do not expand or remove their runtime implementation here |

Keep five primary user actions (open, get, list, serve, diagnose). Group daemon administration separately. Hiding a command is only UX cleanup; completion requires removing duplicate collection and obsolete integrations as well.

## Collection and credential contracts

### Single collection owner

Use `core.UsageSnapshot` and existing `daemon.SnapshotFrame`; no parallel DTO. Extend the existing `ViewRuntime` seam and daemon RPCs instead of adding another application framework.

Proposed public methods on `ViewRuntime` (introduce in T2, consume in T4):

```go
func (r *ViewRuntime) ReadForWindow(ctx context.Context, window core.TimeWindow) SnapshotFrame
func (r *ViewRuntime) RefreshForWindow(ctx context.Context, window core.TimeWindow) SnapshotFrame
```

The first replaces misleading `ReadWithFallbackForWindow`; the second already exists and should be fixed/extended rather than duplicated. Keep existing callers compiling with thin wrappers until T4 removes the last old call. Existing request IDs remain responsible for discarding responses from superseded time-window selections.

The daemon owns provider instances, quota fetches, history collection, normalization, and publication. Move any necessary enrichment into this path once; remove frontend fetch-and-overlay closures after proving equivalent fields. History collectors own tokens/spend; live quota owns limits/resets. Never add those sources together as duplicate usage. Preserve `limit_snapshot` projection and existing dedup identities.

Coalesce concurrent refresh requests into the same in-flight polling generation, returning only after its snapshot is published. Merely serializing duplicate polls behind a mutex is insufficient. Scheduled and explicit refresh share rate-limit backoff; API failures do not reset the last-success timestamp. Returning cached data must identify its age and the latest error. Account filtering can initially occur after the shared read: do not add a second RPC solely for `get`.

### Provider auth policy

| Provider class | Policy for this change |
|---|---|
| Antigravity | Valid saved access token -> quota API; expired/rejected token -> direct OAuth refresh -> one API retry; no `agy`, `agy-box`, Docker, or prompt execution |
| Claude Code | Prefer local OAuth credentials; retain existing cookie fallback only when OAuth is unavailable or explicitly rejected; do not fail over on 429/transient errors |
| Codex | Preserve supported saved access-token API flow and local history; expired login yields clear auth-required status unless refresh ownership is separately verified |
| Cursor, OpenCode, other providers | Preserve existing supported API keys, local credentials or session auth; remove obsolete status-line installation without claiming all support refresh grants |

For Antigravity, retain existing environment variable names as aliases. Do not embed a discovered client secret or copy raw credentials into the plan. Verify the issuer/client pairing and token-file ownership using sanitized fixtures and the actual installed tool's documented contract before enabling renewal. If client configuration is missing, still use a valid access token; when renewal is needed show an actionable auth message. This is a configuration condition, not a reason to launch a coding agent.

Auth state machine acceptance:

- Valid access token: zero token endpoint requests; one quota request.
- Missing file / invalid JSON: account-local `StatusAuth`, no subprocess or repeated token requests.
- Expired token + configured refresh: one refresh; preserve old refresh token if response omits one; persist rotated token atomically with `0600` mode, preserving unrelated JSON fields.
- Quota 401: force refresh even if stored expiry is in the future; retry quota once. A second 401 ends in `StatusAuth`.
- Quota 403: classify as authorization failure without repeatedly refreshing; do not treat every 403 as expiry.
- Invalid grant: auth-required until credentials change; no retry loop. Token endpoint timeout/5xx: retryable provider error with bounded backoff, not a revoked-login diagnosis.
- Quota 429: `StatusLimited`, respect retry timing; no token refresh. Network/5xx: `StatusError`, retain last successful values with their original timestamp.
- Concurrent clients/accounts: one renewal per canonical credential path, independent profiles stay independent. Lock/re-read before renewal. Before replacing the file, detect an external change; reload instead of overwriting a newer login. If the owning tool cannot share safe rotation semantics, require re-login rather than inventing a second refresh-token writer.
- Diagnostics include safe status/error codes, never raw response bodies, tokens or cookies. All operations honor cancellation and timeouts.

## Migration and deletion rules

- Stop writing status-line settings in both box scripts on add and launch. Remove Antigravity/Cursor definitions from install choices and provider readiness requirements.
- Add an explicit `doctor --fix-legacy-statuslines` migration using structural JSON parsing and exact known command/argument matching. Remove only agentUsage/openusage-owned Antigravity/Cursor status-line entries, preserve custom commands and all unrelated JSON, and save a backup first. Malformed files are left untouched with a per-file error. Repeating migration is a no-op.
- Default startup and normal `doctor` never rewrite third-party settings. Existing hooks for history-only providers continue to work; hook absence does not mark API quota unavailable.
- Retain old `status_file` config hints as ignored compatibility input for one release. Stop creating accounts solely from stale status files. Preserve explicit configured profiles even when credentials are currently missing.
- Delete Antigravity's status-line payload/decode/projection machinery only after API response fixtures cover its quota semantics. Keep the telemetry collector interface where needed, with a concise unsupported-hook response; no parallel legacy parser.
- The shared Claude five-hour cache remains until tmux has been migrated to the shared read model. This plan does not silently remove it.
- Keep SQLite schema/history intact. Do not delete user's credentials, token files, local transcripts, daemon databases, spool files or settings during cleanup.

## Execution packets and dependencies

Use worktrees under `.worktrees/usage-simplify-<task-id>`. Existing uncommitted work affects dashboard, daemon service, telemetry projection, providers and UI. The coordinator must capture an agreed baseline containing those edits before branching; do not reset, stash, overwrite, or silently exclude them. Each executor reports the baseline commit, changed files, tests, and remaining limitations. The packets are for BXO/Antigravity execution later; no agents were launched by this planning change.

Order: **T1 and T3 may run in parallel; T2 follows T1; T4 follows T2 and T3; T5 follows T4; T6 reviews the integrated result.** These dependencies serialize shared command/test files and keep incompatible frontend/runtime revisions apart.

### T1 — Direct Antigravity auth and API projection

Role: delegated-executor. Depends on: none.

Files: `internal/providers/antigravity/auth.go`, `auth_unix.go`, `auth_windows.go`, `antigravity.go`, `types.go`, `metrics.go`, `telemetry.go`, plus tests in that directory. Create `auth_lock.go` / platform variants only if safe credential ownership requires them. No daemon, CLI, scripts or integration edits.

- [ ] Add `httptest` / injectable transport tests for every auth-state case above, using temporary token files and sentinel secrets. Record the pre-change failures.
- [ ] Replace `ensureAccessToken` ping paths with explicit auth/refresh outcomes; replace string-based HTTP auth classification with typed status errors; make 401 force renewal.
- [ ] Add per-path coordination, preservation of unknown credential fields, rotation/write-failure tests, and deterministic cancellation/backoff checks.
- [ ] Replace `statusLinePayload` with API-specific quota projection input; preserve model buckets, reset times, empty-response behavior and metric units. Remove dead status-line parsing after checking references.
- [ ] Remove `pingBoxForToken`, CLI resolution, associated platform process setup and imports once unused. Keep `EnrichSnapshots` temporarily for T4 callers.
- [ ] Run `CGO_ENABLED=1 go test -race ./internal/providers/antigravity`; expect PASS. A sentinel executable named `agy`/`agy-box` must never be invoked by any auth scenario.

### T2 — One daemon polling and lifecycle path

Role: delegated-executor. Depends on: T1.

Files: `internal/daemon/runtime.go`, `process.go`, `server_poll.go`, `server_http.go`, `server_read_model.go`, `poll_scheduler.go`, `types.go`, `accounts.go`, and related daemon tests. `service.go` only for necessary lifecycle integration; preserve its existing edits. No frontend or watcher edits (watcher belongs to T3).

- [ ] Add tests for default launch with no service installed, an existing managed service, an already healthy helper, an outdated helper, and two simultaneous first clients.
- [ ] Prefer healthy helper -> installed managed service -> ordinary helper spawn. Use existing socket/process ownership mechanisms so only one helper binds; no automatic OS service installation or compiler execution on open. Fail clearly on permission/bind/version incompatibility with bounded retry.
- [ ] Introduce `ReadForWindow` as the clear read method; preserve thin old-name wrappers until T4. Return last-good snapshots plus existing daemon state on disconnect; cold start has a clear loading/error state.
- [ ] Test simultaneous timer/manual refreshes with a counting fake provider and a barrier: one active generation, shared completion, published result before response. Make freshness/backoff and last-success timestamps deterministic with a fake clock.
- [ ] Centralize any live quota enrichment needed for parity in the daemon fetch/projection path, preserving history and authoritative quota precedence. Cancel workers on helper shutdown.
- [ ] Run `CGO_ENABLED=1 go test -race ./internal/daemon ./internal/telemetry`; expect PASS. Test missing-service lifecycle with fakes, not by installing/uninstalling the developer's service.

### T3 — Remove status-line coupling and prepare migration

Role: delegated-executor. Depends on: none; disjoint from T1/T2.

Files: `internal/integrations/definitions.go`, `manager.go`, `match.go` and related tests; create `internal/integrations/legacy_statusline.go` and its test; `internal/detect/detect.go` and tests; `internal/daemon/server_watch.go`, `poll_kick_test.go`, `server_more_test.go`, `change_detection_test.go`; `scripts/boxes/agy-box`, `scripts/boxes/agent-box`, `scripts/box_test.sh`. No other daemon files or command files.

- [ ] Add fixtures for no config, owned command, quoted executable path, custom command, malformed JSON, multiple box profiles, stale status file only, and repeated cleanup.
- [ ] Stop box add/launch injection; remove active Antigravity/Cursor status-line definitions and readiness dependencies. Implement narrowly scoped cleanup helper for T5, without running it on startup.
- [ ] Discover profiles from configured directories/credential paths; stop emitting `status_file` hints and treating status files as quota discovery evidence. Preserve account IDs and local/box overrides.
- [ ] Delete `isAntigravityStatusFile` polling kicks and associated status-file-specific tests. Preserve unrelated local history watchers and poll scheduling tests.
- [ ] Run `CGO_ENABLED=1 go test -race ./internal/integrations ./internal/detect ./internal/daemon` and `bash -n scripts/boxes/agy-box scripts/boxes/agent-box`. Add shell fixture coverage to `scripts/box_test.sh`, then run `bash scripts/box_test.sh` using fake tools and temporary HOME.

### T4 — Make every presentation consume shared usage

Role: delegated-executor. Depends on: T2, T3.

Files: `cmd/agentusage/get.go`, `get_test.go`, `dashboard.go`, `dashboard_update_test.go`, `snapshot_dispatcher.go`, `snapshot_dispatcher_test.go`; `internal/webserve/collect.go` and collector tests; `internal/daemon/runtime.go`, `runtime_broadcast_test.go` for retiring old wrappers; `internal/providers/antigravity/antigravity.go` for retiring enrichment wrapper; `internal/providers/claude_code/claude_code.go`, `usage_api_oauth_test.go` for auth ordering. Other provider enrichment wrappers only if reference checks prove they became unused.

- [ ] Add parity fixtures: the same account/window/generation produces identical quota values, reset times and status in `get` JSON, web data and TUI model. Keep existing JSON shape and account resolution tests.
- [ ] Replace `fetchAccountSnapshot` direct provider dispatch with the shared service read/refresh. Preserve `get` timeout and error semantics; service failure must not look like a successful zero quota response.
- [ ] Remove provider construction and fetch/overlay closures from dashboard/web. Route manual refresh through the shared generation and normal renders through reads. Preserve request IDs and reject stale-window responses.
- [ ] Prefer Claude OAuth before cookies. Cover fallback on missing/rejected OAuth, and no cookie fallback on network errors/429. Preserve local history and shared tmux cache.
- [ ] Rename remaining misleading runtime read calls and remove compatibility wrappers after repository-wide reference checks.
- [ ] Run `CGO_ENABLED=1 go test -race ./cmd/agentusage ./internal/webserve ./internal/tui ./internal/daemon ./internal/providers/claude_code ./internal/providers/antigravity ./internal/tmux`; expect PASS. Verify multiple views no longer multiply fake upstream request counts.

### T5 — Smaller command help, migration UX and documentation

Role: delegated-executor. Depends on: T4.

Files: `cmd/agentusage/main.go`, `main_test.go`, `doctor.go`, `doctor_test.go`, `detect.go`, `detect_test.go`, `daemon.go`; `internal/tui/model_input.go`, `model_panels.go`, `help.go`, `model_install_test.go`; `README.md`, `RUNBOOK.md`, `docs/TELEMETRY_INTEGRATIONS.md`, `docs/user-flow-diagram.md`, `docs/COMMAND_FLOW_DIAGRAMS.md`, relevant `docs/diagrams/*.svg` and diagram tests. No provider/runtime edits.

- [ ] Group help into everyday usage and advanced service administration. Hide/deprecate `detect` while preserving its execution/flags/output; share its implementation with `doctor --detect`.
- [ ] Wire `doctor --fix-legacy-statuslines` to T3's cleanup helper. Display affected paths and outcomes without tokens; backup before mutation. Normal diagnostics remain read-only.
- [ ] Remove obsolete status-line/tmux installation checks from quota readiness and onboarding; retain optional history integration controls where still useful. Default setup should say sign in/open/refresh.
- [ ] Update root command tests to assert visible/help groups separately from compatibility routes, rather than mistaking hidden aliases for deleted commands.
- [ ] Replace the archived flows in the user guide with verified final flows; regenerate corresponding SVGs, and update diagram count assertions only if the guide's structure changes. Correct architecture guide's obsolete direct-fallback and demo-command claims.
- [ ] Run `CGO_ENABLED=1 go test -race ./cmd/agentusage ./internal/integrations ./internal/tui` and compile/render changed PlantUML diagrams. Verify `--help`, `get --help`, `doctor --help`, and legacy `detect` output using temporary config roots.

### T6 — Independent acceptance review

Role: reviewer. Depends on: T5. Read-only; fixes return to the task owner.

- [ ] Review the integrated diff against this plan and record baseline/head IDs. Confirm unrelated working-tree edits survived and all task file scopes were respected.
- [ ] Run `CGO_ENABLED=1 go build ./cmd/agentusage`, `make test`, `make vet`, and `make lint`. Report lint skipped if the binary is unavailable; do not call that a lint pass.
- [ ] Exercise a fake-provider end-to-end sequence: cold launch without managed service, automatic helper start, valid/expired/revoked login, two clients refreshing, quota 429, offline cached display, account switch, helper reconnect, clean UI exit.
- [ ] Verify zero coding-agent process launches, no quota dependency on status-line files/hooks, no secret-bearing diagnostics, and no duplicate upstream fetches for one overlapping refresh generation.
- [ ] Exercise box add/launch and migration with fixture directories; verify custom status lines, existing history and account IDs survive. No real credential mutation is needed for acceptance.
- [ ] Report PASS / CONDITIONAL PASS / FAIL with exact evidence and unresolved limitations. A conditional result is not permission to ship. Deployment, real-account migration and external BXO edits are separate execution decisions.

## Definition of done and rollback

Done means the default view works without status-line or managed-service setup; Antigravity renews through HTTP when configured; TUI/web/get read the same snapshots; failures retain honest freshness/status; API quota no longer depends on history hooks; and all compatibility tests pass. Measure reduction by removed direct-fetch/enrichment call sites, obsolete installer definitions, process fallback branches and duplicate polls—not line count alone.

Land T1, T3, T2, T4, T5 as reviewable commits in dependency order. Each commit must build. Keep migration explicit and idempotent so reverting code does not require database rollback. Restore a backed-up third-party settings file only when deliberately undoing that migration; never overwrite subsequent user edits. Retain old credential keys/config fields during the compatibility release.
