# Project TODO

Backlog for upcoming infrastructure and UX work.

## 1. OpenTelemetry logging stack

**Goal:** Replace ad-hoc `log` / `StructuredLogger` output with structured, exportable observability via the OpenTelemetry stack.

**Current state:**
- Logging uses the stdlib `log` package and `internal/core/StructuredLogger` (`component=X level=Y event=Z`).
- Debug output is gated by `AGENTUSAGE_DEBUG=1`.
- No traces or metrics export today.

### Tasks

- [ ] **Decide scope** — logs only vs logs + traces + metrics; identify which subsystems emit first (daemon, telemetry pipeline, provider polling).
- [ ] **Add OTel SDK dependencies** — `go.opentelemetry.io/otel`, exporters (OTLP gRPC/HTTP), optional `slog` bridge.
- [ ] **Create `internal/observability/` package** — bootstrap/shutdown helpers, resource attributes (service name, version from `internal/version`, commit, build date).
- [ ] **Wire CLI flags / env config** — e.g. `AGENTUSAGE_OTEL_ENDPOINT`, `AGENTUSAGE_OTEL_INSECURE`, service name override; document in README and `configs/example_settings.json` if persisted.
- [ ] **Migrate structured logging** — bridge `StructuredLogger` to OTel logs (or `slog` + OTel handler) without breaking existing `component=` / `event=` fields.
- [ ] **Instrument key paths** — daemon server lifecycle, provider fetch latency/errors, telemetry ingest/dedup, hub HTTP handlers.
- [ ] **Redact sensitive data** — ensure API keys, tokens, cookies, and raw auth headers never appear in log attributes.
- [ ] **Graceful shutdown** — flush exporters on daemon/TUI exit (signal handlers in `cmd/agentusage`).
- [ ] **Local dev stack** — optional `docker-compose` snippet with OTel Collector + Jaeger or Grafana Tempo for validation.
- [ ] **Tests** — unit tests for bootstrap, attribute redaction, and no-op mode when OTel is disabled.

---

## 2. CI/CD dockerization

**Goal:** Build, test, and publish container images in CI for deployable components.

**Current state:**
- `Dockerfile.hub` exists for the hub/headless aggregation server only.
- `.github/workflows/ci.yaml` runs lint, vet, test, and native binary builds — no Docker jobs.
- CGO + `mattn/go-sqlite3` requires gcc/musl in the builder stage.

### Tasks

- [ ] **Audit image targets** — confirm which binaries need images: `agentusage` daemon, `agentusage hub`, telemetry-only sidecar (if any).
- [ ] **Add / refresh Dockerfiles** — multi-stage builds with pinned base images; align `Dockerfile.hub` module paths with current `cmd/agentusage` layout.
- [ ] **Add `.dockerignore`** — exclude `.git`, test artifacts, `bin/`, coverage, and local config.
- [ ] **CI: build job** — new workflow (or CI job) that runs `docker build` on PR and main; fail on build errors.
- [ ] **CI: smoke test** — run container with `--headless` / health endpoint; basic startup + exit-code check.
- [ ] **CI: publish on release** — push tagged images to GHCR (or chosen registry) from `release.yaml` / release-please flow.
- [ ] **Version tagging** — tag images with semver, git SHA, and `latest` (main only); pass ldflags (`Version`, `CommitHash`, `BuildDate`) as build-args.
- [ ] **Document usage** — README section: build locally, run daemon/hub in Docker, required volumes (`~/.config/agentusage`, socket paths), and CGO/runtime notes.
- [ ] **Security hardening** — non-root runtime user, minimal Alpine/distroless final stage, scan images in CI (optional Trivy/Grype step).

---

## 3. In-app account add form

**Goal:** Let users add provider accounts from inside the app via a simple guided form — no hand-editing `settings.json`.

**Current state:**
- Implemented in-app guided form accessible via `a` shortcut on the dashboard and in settings tabs (**1 PROV** / **5 KEYS**).
- Driven dynamically by `ProviderSpec` metadata (auth types, default account ID, doc URL, env vars, cookie config).
- Live validation with `ValidateAPIKey`, direct key entry or env var toggle, browser session connect, duplicate account collision prevention.
- Atomic persistence to `config.Accounts` and `credentials.json` with immediate dashboard refresh.

### Tasks

- [x] **UX design** — single “Add account” entry point (accessible via `a` key on dashboard and in settings modal); fields: provider picker, account ID/label, auth method (API key / browser session / local path).
- [x] **Drive form from `ProviderSpec`** — use each provider’s `Auth` metadata (default account ID, supported auth types, env var names, browser cookie refs) so the form stays in sync with the registry.
- [x] **Implement add-account modal** — Bubble Tea form with validation, inline errors, and keyboard-first navigation consistent with existing settings UX.
- [x] **Persist new accounts** — save to `config.Accounts` (manual accounts take precedence over auto-detected); call existing credential helpers for API keys / browser sessions.
- [x] **Validate before save** — reuse `ValidateAPIKey`; for browser-session providers, browser picker step; check existing accounts and auto-generate non-colliding IDs.
- [x] **Refresh dashboard state** — append to `providerOrder`, enable by default, trigger `requestRefreshAll()` so the new tile appears immediately.
- [ ] **Edit / remove accounts** — optional follow-up: rename account ID, disable vs delete, clear credentials without removing the row.
- [ ] **Web dashboard parity** — if `serve` / web UI is in scope, expose the same add-account form over the existing web settings API.
- [x] **Tests** — TUI input tests for form validation, duplicate rejection, default IDs, modal navigation, and save callbacks; config persistence unit tests.
- [x] **Docs** — Documented in help overlay (`?`) and settings modal hints.

