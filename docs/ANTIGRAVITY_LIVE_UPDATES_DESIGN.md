# Antigravity Live Updates Design

Date: 2026-08-28
Status: Implemented
Author: cursor-agent

## 1. Problem Statement

Antigravity already pushes fresh usage JSON on every CLI state change via the status-line hook, but OpenUsage only re-reads that file on the daemon's fixed poll interval (default 30s), so the dashboard feels stale even though the data on disk is live.

## 2. Goals

1. When `agy` updates the status line, the OpenUsage dashboard reflects the new quota/context/model within about one second (daemon running).
2. Keep the existing status-line contract unchanged — still a JSON stdin → one-line stdout bridge.
3. Fail soft: if the daemon is not running, status-line capture still writes the file (today's behavior).

## 3. Non-Goals

1. Removing the periodic poll loop (it remains the fallback and covers API providers).
2. WebSocket push from daemon to TUI/web (TUI already re-fetches the read model on a short tick once data is ingested).
3. Live updates without `agy` running (no Antigravity usage API exists for this integration).
4. Reworking OpenCode or other providers in this change.

## 4. Impact Analysis

### Affected Subsystems

| Subsystem | Impact | Summary |
|-----------|--------|---------|
| core types | none | — |
| providers | none | Capture path unchanged; notify lives in CLI |
| TUI | none | Existing read-model refresh picks up new ingest |
| config | none | No new settings required |
| detect | none | — |
| daemon | minor | Poll-kick channel, `/v1/poll`, watch OpenUsage state dir |
| telemetry | none | Existing hook ingest reused |
| CLI | minor | `openusage antigravity statusline` notifies daemon after write |

### Existing Design Doc Overlap

- `docs/DAEMON_POWER_OPTIMIZATION_V2_DESIGN.md` deferred full fsnotify-driven collection; this change only kicks an existing poll, not a new collection architecture.
- `docs/site/docs/providers/antigravity.md` documents the status-line bridge; update after ship.

## 5. Detailed Design

### 5.1 Poll kick channel

`Service` gains a buffered kick channel:

```go
pollKick chan struct{} // capacity 1; coalesces bursts
```

`RequestPoll()` non-blocking-sends on it. `runPollLoop` selects on the ticker **or** `pollKick` and calls `pollProviders`.

### 5.2 `POST /v1/poll`

New UDS endpoint. Empty body. Response `{"status":"kicked"}`. Calls `RequestPoll()`. Used by the status-line CLI and available for other local tools.

### 5.3 Watch OpenUsage state directory

Extend `collectWatchDirs()` with `telemetry.DefaultStateDir()` (where `antigravity*-status.json` lives).

On `Write|Create|Rename` for paths matching `antigravity` + `status` + `.json`, call `RequestPoll()` with a short debounce (~300ms) so rename-atomic writes coalesce.

Existing watch behavior for other dirs (markDataIngested only) stays; antigravity needs a full Fetch poll because tile gauges come from `limit_snapshot` ingest of `Fetch()`, not from read-model cache alone.

### 5.4 Status-line CLI notify

After a successful `CaptureStatusLine` in `cmd/openusage/antigravity.go`:

1. Best-effort `IngestHook(ctx, "antigravity", accountID, payload)` (telemetry events).
2. Best-effort `Client.RequestPoll(ctx)` → `POST /v1/poll`.

Timeouts are short (≤1s). Errors are ignored so a missing daemon never breaks Antigravity's footer.

Account ID: prefer `AGY_ACCOUNT`, else email slug from payload if cheap to parse, else omit.

### 5.5 Backward Compatibility

- No config schema changes.
- Status-line stdout contract unchanged.
- Daemon without clients of `/v1/poll` behaves as today plus watching the state dir.

## 6. Alternatives Considered

### Pure shorter poll interval

Rejected as primary fix: wastes CPU when idle and still lags under load. Kick-on-write is event-driven.

### fsnotify only (no status-line notify)

Useful as a backup (e.g. another process writes the file) but status-line → `/v1/poll` is more reliable across container bind-mount edge cases and avoids depending solely on inotify.

### Have CaptureStatusLine import daemon

Rejected: `daemon` already imports `providers`; would create an import cycle. Notify from the CLI command instead.

## 7. Implementation Tasks

### Task 1: Poll kick + `/v1/poll`
Files: `internal/daemon/server.go`, `server_poll.go`, `server_http.go`, `client.go`, tests
Depends on: none
Description: Add kick channel, RequestPoll, wire runPollLoop, HTTP handler, client method.
Tests: RequestPoll coalesces; handler returns 200; poll loop reacts to kick without waiting for ticker.

### Task 2: Watch state dir for antigravity status files
Files: `internal/daemon/server_watch.go`, tests
Depends on: Task 1
Description: Watch DefaultStateDir; match antigravity status filenames; RequestPoll with short debounce.
Tests: filename matcher; collectWatchDirs includes state dir when present.

### Task 3: Status-line CLI notify
Files: `cmd/openusage/antigravity.go`, optional small helper + test
Depends on: Task 1
Description: After capture, best-effort hook ingest + RequestPoll.
Tests: helper no-ops / ignores dial errors; does not fail the command.

### Task 4: Docs
Files: `docs/site/docs/providers/antigravity.md`, `docs/site/docs/daemon/integrations.md`
Depends on: Tasks 1–3
Description: Document that live dashboard updates require the telemetry daemon and an active `agy` session.

### Dependency Graph

```
Task 1 → Task 2
Task 1 → Task 3
Tasks 2, 3 → Task 4
```
