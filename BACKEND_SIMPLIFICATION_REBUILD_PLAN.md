# Backend simplification rebuild

## Context

Rebuild the backend from the ground up as three pieces only: a provider list where each provider is one direct HTTP (or local-file) fetch, a small SQLite store holding accounts plus tokens/refresh tokens/expiry, and a config sync that upserts the configured accounts/credentials into that store. The current tree couples `core.UsageProvider.Fetch` to widget/status/analytics/telemetry layers (`DashboardWidget`, `DetailWidget`, `core.Status*`, `shared.ApplyStatus*`, `shared.FinalizeStatus`, `shared.ProcessStandardResponse`, `shared.ProbeRateLimits`, `ChangeDetector`, per-provider caches/TTLs, `internal/telemetry` event/rollup/read-model pipeline); all of that is deleted, not ported. No new test files are added per request; proof is `go build` plus direct live-fetch smoke checks. Fingerprint safety is a hard constraint: no browser-UA spoofing, no cookie-scrape polling, no HTML-scrape polling, no retry storms.

## Approach

### Step 1 — Create new `internal/backend` package shell alongside the old tree (no deletions yet)

Create `internal/backend/types.go` with the only three types in the new backend, replacing `core.AccountConfig` / `core.UsageSnapshot` / `core.ProviderSpec` for fetch purposes:

```go
package backend

type Account struct {
    ID, Provider, AuthType, APIKeyEnv, BaseURL string
}
type Creds struct {
    APIKey, AccessToken, RefreshToken, Cookie string
    ExpiresAt int64 // unix seconds, 0 = unknown
}
type Snapshot struct {
    Provider, AccountID string
    FetchedAt int64 // unix seconds
    Data map[string]string // raw endpoint fields only, e.g. "balance_usd", "reset_at"
}
type Fetcher func(ctx context.Context, http *http.Client, acct Account, creds Creds) (Snapshot, error)
```

Do not reuse `core.UsageProvider`, `core.ProviderSpec`, `core.DashboardWidget`, `core.DetailWidget`, `core.Metric`, or `core.Status*`. Do not add widget/status/analytics helpers. Old code keeps compiling because nothing is deleted in this step.

### Step 2 — Create minimal SQLite token store in `internal/backend/store.go`

New SQLite file (default path `~/.local/share/agentusage/backend.db`, unverified — confirm first against `internal/telemetry/paths.go` state-dir resolution and reuse that resolver if it already picks the state dir). Single `sql.Open("sqlite3", ...)` handle, `SetMaxOpenConns(1)`, `PRAGMA journal_mode=WAL; PRAGMA busy_timeout=5000;`. Exact DDL, executed on open:

```sql
CREATE TABLE IF NOT EXISTS accounts(
  id TEXT PRIMARY KEY, provider TEXT NOT NULL, auth_type TEXT NOT NULL,
  api_key_env TEXT NOT NULL DEFAULT '', base_url TEXT NOT NULL DEFAULT '',
  updated_at INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS tokens(
  account_id TEXT PRIMARY KEY REFERENCES accounts(id) ON DELETE CASCADE,
  api_key TEXT NOT NULL DEFAULT '', access_token TEXT NOT NULL DEFAULT '',
  refresh_token TEXT NOT NULL DEFAULT '', cookie TEXT NOT NULL DEFAULT '',
  expires_at INTEGER NOT NULL DEFAULT 0, updated_at INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS snapshots(
  account_id TEXT PRIMARY KEY REFERENCES accounts(id) ON DELETE CASCADE,
  fetched_at INTEGER NOT NULL, data_json TEXT NOT NULL DEFAULT '{}'
);
```

Functions, exact signatures:

```go
func Open(path string) (*Store, error)
func (s *Store) SyncAccount(acct Account, creds Creds) error // upsert accounts row + tokens row, sets updated_at=time.Now().Unix()
func (s *Store) Creds(accountID string) (Account, Creds, error) // single-row read; sql.ErrNoRows if missing
func (s *Store) SaveSnapshot(snap Snapshot) error // upsert snapshots row, data_json=json.Marshal(snap.Data)
```

Refresh-token write-back: `antigravity` is today the only provider whose refresh flow rewrites the provider's own token file (`internal/providers/antigravity/auth.go` flock-guarded `writeOAuthToken`); gemini/claude/codex/cursor tokens are read-only from provider-native files. The new backend keeps exactly one write-back: after a successful antigravity refresh, `SyncAccount` the new tokens AND rewrite the provider token file at the resolved `tokenFilePath` under the same file lock discipline. All other providers never write outside the store.

No events/rollup/dedup/retention/balance_observations/spool/read-model tables. Copy nothing from `internal/telemetry/store.go` except the single-connection + WAL pragma shape; avoid its migration list.

### Step 3 — Config-to-store sync in `internal/backend/sync.go`

Single entry point, exact signature:

```go
func SyncConfigToStore(s *Store, accounts []Account, credsByAccount map[string]Creds) error
```

Behavior: for each account, call `SyncAccount`; do not delete rows for accounts absent from config in v1 (one inline line: stale-row pruning is out of scope to avoid deleting user tokens on a transient config read). Caller wires it at daemon/dashboard startup from existing `internal/config` load (`LoadCredentials` account IDs + `settings.json` account list — map `APIKeyEnv`/`BaseURL`/`Auth` fields into `Account`, `Keys[accountID]` into `Creds.APIKey`). Secrets never go into logs; on sync error return the error, no partial in-memory cache.

### Step 4 — One shared fingerprint-safe HTTP helper in `internal/backend/http.go`

All HTTP providers call this and nothing else. Exact signature:

```go
func Do(ctx context.Context, client *http.Client, method, url string, headers map[string]string, body []byte) (status int, respBody []byte, respHeaders http.Header, err error)
```

Rules baked in, no per-provider overrides: `http.NewRequestWithContext`, 30s client timeout set by caller (`&http.Client{Timeout: 30*time.Second}`), `Accept: application/json` only, no default `User-Agent` (Go's `Go-http-client/2.0` stays — do not set a browser or CLI UA unless the provider row below explicitly requires it), no `Referer`, no cookie jar. Caller passes at most `Authorization: Bearer <token>` (or the provider's documented scheme) plus any documented version header from the table. Error policy: network error returns `err`; HTTP 401/403 returns `ErrAuth` (sentinel `errors.New("backend: auth failed")` with status attached via `fmt.Errorf("backend: auth failed: status %d: %w", status, ErrAuth)`); HTTP 429 returns `ErrLimited` and copies `Retry-After` into the returned headers only (caller sleeps, never retries in a loop); other non-2xx returns `fmt.Errorf("backend: %s status %d: %.200s", provider, status, body)`. No automatic retries. Polling is sequential per account owned by the daemon ticker, never concurrent per account; minimum 60s between polls of the same account (enforced by daemon ticker, not by providers).

Delete-or-avoid list (do not port): `shared.CreateStandardRequest`, `shared.FetchJSON`, `shared.ProbeRateLimits`, `shared.ProcessStandardResponse`, `shared.ApplyStatusFromResponse`, `shared.ApplyStatusFromCode`, `shared.ApplyStandardRateLimits`, `shared.FinalizeStatus`, `providerbase.Base.Client` caching, per-provider `cacheMu`/`ttl*` caches (`copilot ttlSnapshot`, `cursor livePlanCacheTTL`), `ChangeDetector.HasChanged`.

### Step 5 — Per-provider direct fetchers in `internal/backend/providers/*.go`

One file per HTTP provider, one exported func each, exact shape `func FetchX(ctx, client, acct, creds) (backend.Snapshot, error)` calling only `backend.Do` (or plain `os.ReadFile` for local providers). Base-URL rule for every row: `url = acct.BaseURL if non-empty else default below` (this preserves the existing `shared.ResolveBaseURL` behavior without importing it). Auth rule: token comes from `creds` argument (populated from the store), never read from env/disk inside the fetcher. Each fetcher does exactly the listed call(s), parses only the fields it needs into `Snapshot.Data`, and returns. No status-line writes, no widget construction, no metric normalization.

HTTP providers (method, URL, auth — all read this session except where marked):

- `openai`: `GET {base=https://api.openai.com/v1}/models/gpt-4.1-mini`, `Authorization: Bearer <apiKey>`. Surface: rate-limit headers only.
- `anthropic`: `GET {base=https://api.anthropic.com/v1}/messages`, headers `x-api-key: <apiKey>`, `anthropic-version: 2023-06-01`, and nothing else — do not port the current behavior where `CreateStandardRequest` additionally injects `Authorization: Bearer <key>` on top. Note: current code probes via GET; keep one GET probe, do not invent a POST body.
- `azure_openai`: `GET {endpoint}/openai/deployments?api-version=2024-10-21`, header `api-key: <key>` (no Bearer). Endpoint = `acct.BaseURL` else `AZURE_OPENAI_ENDPOINT` else `https://<AZURE_RESOURCE_NAME>.openai.azure.com`.
- `alibaba_cloud`: `GET {base=https://dashscope.aliyuncs.com/api/v1}/quotas`, Bearer.
- `openrouter`: `GET {base=https://openrouter.ai/api/v1}/key` (fallback `/auth/key`), then `GET /credits`, then `GET /generation?limit=100&offset=N` pages (+ optional `/keys`, `/activity`, `/api/internal/v1/transaction-analytics?window=1mo` only if needed; that last one trims `/api/v1` off the base). Bearer.
- `perplexity`: `GET https://console.perplexity.ai/rest/pplx-api/v2/groups`, `GET /rest/pplx-api/v2/groups/{org_id}`, `GET /rest/pplx-api/v2/groups/{org_id}/usage-analytics`, every call appends `?version=2.18&source=default`. Cookie `__Secure-next-auth.session-token=<v>` from `creds.Cookie` plus headers `x-app-apiversion: 2.18`, `x-app-domain: api-console`. Keep the existing `agentusage/perplexity-console` UA exactly; add nothing browser-like. Implement org-list + balance/usage only; drop anything that needs HTML scraping.
- `groq`: `GET https://api.groq.com/openai/v1/models`. Bearer. Single GET probe, same shape as openai.
- `mistral`: `GET {base=https://api.mistral.ai/v1}/billing/subscription`, `GET /billing/usage?...`, `GET /models`. Bearer.
- `moonshot`: `GET {base=https://api.moonshot.ai}/v1/users/me`, `GET /v1/users/me/balance` (`.cn` base `https://api.moonshot.cn` when account uses China). Bearer.
- `deepseek`: `GET {base=https://api.deepseek.com}/user/balance`, `GET /v1/models`. Bearer.
- `xai`: `GET {base=https://api.x.ai/v1}/api-key`. Bearer.
- `zai`: `GET {codingBase}/models` where codingBase defaults `https://api.z.ai/api/coding/paas/v4` (China `https://open.bigmodel.cn/api/coding/paas/v4`), plus `GET {monitorBase}/api/monitor/usage/quota/limit` (plus `/api/monitor/usage/model-usage`, `/tool-usage`, `/api/paas/v4/user/credit_grants` only if needed) where monitorBase defaults `https://api.z.ai` (China `https://open.bigmodel.cn`). Auth: first try `Authorization: <raw key>`, on 401/403 retry once with `Authorization: Bearer <key>`.
- `opencode`: `GET {base=https://opencode.ai}/zen/v1/models`, `GET /zen/go/v1/usage`. Bearer. Do not port `console_rpc.go` `_server` hash-ID RPC or `/auth` cookie flow — that pinned-hash browser-console scraping is the highest ban/fragility risk after the claude cookie path.
- `gemini_api`: `GET {base=https://generativelanguage.googleapis.com/v1beta}/models?key=<apiKey>`. Key-in-query, no Authorization header.
- `gemini_cli`: `POST https://oauth2.googleapis.com/token` (refresh_token → access_token), then `POST https://cloudcode-pa.googleapis.com/v1internal:loadCodeAssist`, then `POST https://cloudcode-pa.googleapis.com/v1internal:retrieveUserQuota`. Bearer. Exact JSON shapes copied from `internal/providers/gemini_cli/api_usage.go` request structs; no client-metadata spoofing beyond what those structs already send.
- `antigravity`: `POST https://oauth2.googleapis.com/token` (refresh), then `POST https://daily-cloudcode-pa.googleapis.com/v1internal:retrieveUserQuotaSummary` with body `{}`, headers `Authorization: Bearer`, `Content-Type: application/json`, `User-Agent: antigravity`. Keep that exact UA (it is the product's own identifier, not a browser spoof); do not add Chrome headers.
- `ollama`: local `GET http://127.0.0.1:11434/api/version`, `GET /api/tags`, `GET /api/ps` (no auth); cloud `POST https://ollama.com/api/me`, `GET https://ollama.com/api/tags` with Bearer. Do not port `GET https://ollama.com/settings` HTML scrape.
- `cursor`: `POST https://api2.cursor.sh/aiserver.v1.DashboardService/GetCurrentPeriodUsage`, `POST .../GetPlanInfo`, `POST .../GetHardLimit`, headers `Authorization: Bearer <jwt from creds>`, `Content-Type: application/json`, `Connect-Protocol-Version: 1`, body `{}`. No extra UA.
- `copilot`: shell out to `gh api` exactly as today — `GET /user`, `GET /copilot_internal/user`, `GET /rate_limit`, `GET /orgs/{org}/copilot/billing`, `GET /orgs/{org}/copilot/metrics` via `gh api <path>` subprocess (no direct `api.github.com` token use, no new HTTP client). Token comes from the `gh` session, not the store.
- `codex`: `GET https://chatgpt.com/backend-api/wham/usage` (fallback `{base}/api/codex/usage` when base lacks `/backend-api`), headers `Authorization: Bearer`, `ChatGPT-Account-Id: <id when known>`, `Accept: application/json`, `User-Agent: codex-cli[/<version> when known]`. Keep the `codex-cli` UA (official CLI identifier); never send a browser UA here.
- `claude_code`: `GET https://api.anthropic.com/api/oauth/usage`, headers `Authorization: Bearer <oauth token>`, `anthropic-beta: oauth-2025-04-20`. This is the only HTTP call. Do not port `GET https://claude.ai/api/organizations/{org}/usage` cookie path with `Mozilla/5.0...Chrome/131`, `Referer: https://claude.ai/settings/usage`, `anthropic-client-platform: web_claude_ai` — that browser-mimicry cookie scrape is the top ban risk in the tree and is deleted.
- `command_code`: `GET {base=https://api.commandcode.ai}/alpha/whoami`, `GET /alpha/billing/credits`, `GET /alpha/billing/subscriptions`, `GET /alpha/usage/summary`. Bearer, `User-Agent: cli` (keep this exact minimal UA, it is already non-browser).

Local-file-only providers (no HTTP at all — read session JSONL/SQLite under the tool's state dir via `os.ReadFile`/`sql.Open` read-only, return counts/tokens into `Snapshot.Data`): `amp`, `goose`, `hermes`, `mux`, `droid`, `crush`, `roocode`, `kilo_code` (package `kilocode`), `kiro_cli` (package `kiro`), `zed`, `codebuff`, `kimi_cli`, `pi`, `qwen_cli`, `openclaw`. Exact file layouts verified per provider dir in the scout (each is a `local`-auth provider with no `http.NewRequest` call site); implement each as a ≤40-line `FetchX` that returns empty `Data` with no error when its directory is absent.

Registry: `internal/backend/providers/registry.go` with `func Fetchers() map[string]backend.Fetcher` mapping all 37 IDs from `internal/providers/registry.go AllProviders` to the new `FetchX` funcs, preserving the two non-obvious IDs (`kilocode` package → `kilo_code`, `kiro` package → `kiro_cli`). No `Describe`/`Spec`/widget entries.

### Step 6 — Rewire callers, then delete the old backend

Only after Steps 1–5 compile: point the daemon poller (`internal/daemon/server_poll.go`, the only scheduled `Fetch` caller) and dashboard loader at `backend.Fetchers` + `Store.Creds` (one fetch per account per tick, sequential, 60s minimum interval, stop on `ErrAuth` until re-sync; the daemon's poll `ctx` deadline — 8s today — governs each fetch, the 30s client timeout is only a ceiling). Drop the `ChangeDetector.HasChanged` gate and the `snapshotFingerprint` broadcast gate — fetch directly every tick. Then delete: `internal/providers/*` fetch/widget/spec files, `internal/providers/shared` status/rate-limit helpers, `internal/providers/providerbase` widget constructors, `internal/core` widget/detail/status/analytics/cost-visibility files backing them (`detail_widget.go`, `dashboard_widget.go`, `analytics_*.go`, `cost_visibility.go` and their tests), the `internal/telemetry` event/dedup/rollup/read-model/spool/collector/pipeline files superseded by `snapshots` — explicitly `usage_reconciliation_windows` + `EventTypeReconcileAdjust` (no writer/reader), `_migrations` + the 4 one-shot repair SQLs in `store.go:341-390`, `usage_rollup_daily` + `daemon_meta` watermark + the whole retention tier (`RollupDaily`, `PruneOldEvents`/`PruneOrphanRawEvents`/`PruneRawEventPayloads`, retention loop — the rollup is read only by tests, never by the read model), one of the two spool systems (`telemetry/spool.go` pipeline spool vs `server_spool.go` hook-spool dir; keep neither — the new backend has no queue), the dedup/read-model cache tier (`dedup.go` key branch, `usage_view_cache.go` global cache, `poll_scheduler.go` FNV hash, read-model cache in daemon `types.go`), and dead packages `internal/tmux`, `internal/ccevents`, `internal/report`, `internal/hub` (no importers; CLI commands already removed) — plus the residual statusline coupling: `internal/providers/cursor/statusline.go` capture/render (cursor fetch must not depend on `status_file`; keep only the payload parser+types) and `internal/providers/claude_code/usage_cache.go` 5h-cache writes plus the dead `settingsConfig.StatusLine` field (daemon reads `Snapshot.Data`, never the cache file). The CLI `statusline` command is already gone and `internal/integrations/legacy_statusline.go` cleanup stays untouched. `internal/config/config.go` `modifyConfig`/`saveLocked` and `credentials.go` `writeCredentials` (0600, `credMu`) stay as the config-write mechanism Step 3 builds on; do not redesign them. Default clean cutover: no aliases, no `UsageProvider` compatibility shim. Exact deletion list is whatever `grep -r "DashboardWidget\|DetailWidget\|StatusAuth\|StatusLimited\|FinalizeStatus\|ChangeDetector\|ProviderSpec" --include=*.go` returns at that point — delete every hit outside `internal/backend`.

## Critical files & anchors

- `internal/core/provider.go:241-255` — `UsageProvider` interface (`DashboardWidget`/`DetailWidget`/`Fetch`) being replaced by `backend.Fetcher`.
- `internal/providers/registry.go:47-87` — canonical 37-provider ID list the new registry must cover 1:1.
- `internal/providers/shared/helpers.go:16-99` — `CreateStandardRequest`/`ApplyStatus*`/`FinalizeStatus`/`FetchJSON` status-line machinery being deleted, not ported.
- `internal/telemetry/store.go:211-323` — current event/rollup schema proving how much is deleted (replaced by 3-table DDL in Approach Step 2).
- `internal/providers/claude_code/claude_code.go:486-501` — browser-UA cookie auth being explicitly banned from the rebuild (keep only the OAuth fetcher).

## Verification

No new `_test.go` files. Commands run from repo root with `CGO_ENABLED=1` (sqlite):

1. `go build ./...` — must exit 0 after every Approach step (Steps 1–5 add alongside; Step 6 deletes).
2. `go vet ./internal/backend/...` — must exit 0 after Step 5.
3. Live smoke per touched HTTP provider (example: `OPENAI_API_KEY=<key> go run ./cmd/agentusage telemetry daemon` or a throwaway `go run` calling `backend.Fetchers()["openai"]` with store creds): input = real key in store via sync; expected = `Snapshot.Data` populated with that endpoint's fields and no `User-Agent: Mozilla` / `Referer` / cookie headers on the wire (verify by pointing `acct.BaseURL` at a local request-dumping server and reading the received headers). Repeat for `codex`, `cursor`, `claude_code` (OAuth only), `gemini_cli`, `antigravity` before cutover.
4. Ban-risk check: `grep -rn "Mozilla/5.0\|Referer\|web_claude_ai\|/_server?id=\|/settings" internal/backend/` must return zero rows after Step 5 (only allowed UAs: exact `antigravity`, `codex-cli*`, `cli`).

## Assumptions & contingencies

- Assumption: SQLite becomes the runtime token home and config is the editor (user's stated model), even though today secrets live in `credentials.json` (0600) and SQLite holds usage events. If during execution the team prefers to keep secrets in `credentials.json` and use SQLite only for snapshot cache, do that instead: keep `SyncAccount` writing only non-secret fields + snapshot rows, and read secrets from `internal/config.LoadCredentials` at fetch time — the fetcher signatures do not change.
- Assumption: dropping all historical telemetry/rollup data is acceptable in a ground-up rebuild (no migration of `usage_events`/`usage_rollup_daily`). If history must survive, do not port the pipeline: freeze the old DB file as a read-only archive and start the new 3-table DB at a new path instead.
- Assumption: local-file providers need only best-effort counts (empty `Data`, no error when absent). If any of the 15 listed local providers actually requires an HTTP call found during its confirm-first read, implement it as a `backend.Do` fetcher following Step 4 rules rather than forcing it into the file reader.
