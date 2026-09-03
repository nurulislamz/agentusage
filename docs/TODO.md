# Project TODO

Backlog for upcoming infrastructure and UX work.

## 1. OpenTelemetry logging stack

**Goal:** Replace ad-hoc `log` / `StructuredLogger` output with structured, exportable observability via the OpenTelemetry stack.

**Current state:**
- Logging uses the stdlib `log` package and `internal/core/StructuredLogger` (`component=X level=Y event=Z`).
- Debug output is gated by `AGENTUSAGE_DEBUG=1`.
- No traces or metrics export today.

### Tasks

- [x] **Decide scope** — logs only vs logs + traces + metrics; identify which subsystems emit first (daemon, telemetry pipeline, provider polling).
- [x] **Add OTel SDK dependencies** — `go.opentelemetry.io/otel`, exporters (OTLP gRPC/HTTP), optional `slog` bridge.
- [x] **Create `internal/observability/` package** — bootstrap/shutdown helpers, resource attributes (service name, version from `internal/version`, commit, build date).
- [x] **Wire CLI flags / env config** — e.g. `AGENTUSAGE_OTEL_ENDPOINT`, `AGENTUSAGE_OTEL_INSECURE`, service name override; document in README and `configs/example_settings.json` if persisted.
- [x] **Migrate structured logging** — bridge `StructuredLogger` to OTel logs (or `slog` + OTel handler) without breaking existing `component=` / `event=` fields.
- [x] **Instrument key paths** — daemon server lifecycle, provider fetch latency/errors, telemetry ingest/dedup, hub HTTP handlers.
- [x] **Redact sensitive data** — ensure API keys, tokens, cookies, and raw auth headers never appear in log attributes.
- [x] **Graceful shutdown** — flush exporters on daemon/TUI exit (signal handlers in `cmd/agentusage`).
- [x] **Local dev stack** — optional `docker-compose` snippet with OTel Collector + Jaeger or Grafana Tempo for validation.
- [x] **Tests** — unit tests for bootstrap, attribute redaction, and no-op mode when OTel is disabled.

---

## 2. CI/CD dockerization

**Goal:** Build, test, and publish container images in CI for deployable components.

**Current state:**
- `Dockerfile.hub` exists for the hub/headless aggregation server only.
- `.github/workflows/ci.yaml` runs lint, vet, test, and native binary builds — no Docker jobs.
- CGO + `mattn/go-sqlite3` requires gcc/musl in the builder stage.

### Tasks

- [x] **Audit image targets** — confirm which binaries need images: `agentusage` daemon, `agentusage hub`, telemetry-only sidecar (if any).
- [x] **Add / refresh Dockerfiles** — multi-stage builds with pinned base images; align `Dockerfile.hub` module paths with current `cmd/agentusage` layout.
- [x] **Add `.dockerignore`** — exclude `.git`, test artifacts, `bin/`, coverage, and local config.
- [x] **CI: build job** — new workflow (or CI job) that runs `docker build` on PR and main; fail on build errors.
- [x] **CI: smoke test** — run container with `--headless` / health endpoint; basic startup + exit-code check.
- [x] **CI: publish on release** — push tagged images to GHCR (or chosen registry) from `release.yaml` / release-please flow.
- [x] **Version tagging** — tag images with semver, git SHA, and `latest` (main only); pass ldflags (`Version`, `CommitHash`, `BuildDate`) as build-args.
- [x] **Document usage** — README section: build locally, run daemon/hub in Docker, required volumes (`~/.config/agentusage`, socket paths), and CGO/runtime notes.
- [x] **Security hardening** — non-root runtime user, minimal Alpine/distroless final stage, scan images in CI (optional Trivy/Grype step).

---

## 3. In-app account add form

**Goal:** Let users add provider accounts from inside the app via a simple guided form — no hand-editing `settings.json`.

**Current state:**
- Accounts come from auto-detection or manual JSON config (`accounts` / `auto_detected_accounts`).
- Settings modal (tab **5 KEYS**) supports editing credentials for *existing* accounts (API key entry, browser-session connect).
- Copy still says “Add an account first under 1 PROV / 5 KEYS” — there is no first-class **Add account** flow.
- `Model.SetOnAddAccount` and `onAddAccount` exist for browser-session connect side effects; persistence wiring is partial.

### Tasks

- [ ] **UX design** — single “Add account” entry point (settings modal action or dedicated modal); fields: provider picker, account ID/label, auth method (API key / browser session / local path where applicable).
- [ ] **Drive form from `ProviderSpec`** — use each provider’s `Auth` metadata (default account ID, supported auth types, env var names, browser cookie refs) so the form stays in sync with the registry.
- [ ] **Implement add-account modal** — Bubble Tea form with validation, inline errors, and keyboard-first navigation consistent with existing settings UX.
- [ ] **Persist new accounts** — save to `config.Accounts` (manual accounts take precedence over auto-detected); call existing credential helpers for API keys / browser sessions.
- [ ] **Validate before save** — reuse `ValidateAPIKey`; for browser-session providers, optional “test connection” step; block duplicate `account.id + provider` pairs.
- [ ] **Refresh dashboard state** — append to `providerOrder`, enable by default, trigger `requestRefreshAll()` so the new tile appears immediately.
- [ ] **Edit / remove accounts** — optional follow-up: rename account ID, disable vs delete, clear credentials without removing the row.
- [ ] **Web dashboard parity** — if `serve` / web UI is in scope, expose the same add-account form over the existing web settings API.
- [ ] **Tests** — TUI input tests for form validation, duplicate rejection, and successful persist callback; config round-trip test for saved account shape.
- [ ] **Docs** — README snippet: “Adding an account from the app” with screenshots; deprecate JSON-only instructions where the form covers the flow.
