# Feature Progress Ledger: agy-box-ping-monitoring

## Seam Contract
- Interface/Signature:
  - `antigravity.RecordBoxPing(box string, accountID string, reason string, duration time.Duration, err error)`
  - `antigravity.WritePrometheusMetrics(w io.Writer)`
  - HTTP `/metrics` endpoint in webserve
  - Logging: structured log output in daemon (`[antigravity] agy-box ping ...`) + persistent lines in `~/.local/state/agentusage/agy-pings.log`
  - Grafana Dashboard: pre-provisioned Prometheus datasource and `agy-boxes.json` dashboard running on port 3000
- Done-When:
  - `go test ./internal/providers/antigravity/... -v` passes
  - `go test ./internal/webserve/... -v` passes
  - `make vet` passes
  - Grafana dashboard running, exposable, and displaying ping metrics and logs

## Tasks & File Plan
- [x] Task 1: Antigravity Box Ping Logging, Tracker & Prometheus Metrics (`internal/providers/antigravity/metrics.go`, `internal/providers/antigravity/metrics_test.go`, `internal/providers/antigravity/auth.go`, `internal/providers/antigravity/antigravity.go`)
- [x] Task 2: Web Server Endpoints for Metrics & Pings (`internal/webserve/server.go`, `internal/webserve/server_test.go`)
- [x] Task 3: Grafana & Prometheus Provisioning & Dashboard (`deploy/grafana/docker-compose.yml`, `deploy/grafana/prometheus.yml`, `deploy/grafana/dashboards/dashboards.yaml`, `deploy/grafana/dashboards/agy-boxes.json`, `deploy/grafana/datasources/prometheus.yaml`, `scripts/start-grafana.sh`)
- [x] Task 4: Runtime Verification & Exposing Dashboard (Start containers, verify metrics scraping, verify Grafana UI, test ping logging live, ensure persistent running)

## Locked Files (Do Not Touch in Later Passes)
- `internal/providers/antigravity/metrics.go`
- `internal/providers/antigravity/metrics_test.go`
- `internal/providers/antigravity/auth.go`
- `internal/providers/antigravity/antigravity.go`
- `internal/webserve/server.go`
- `internal/webserve/server_test.go`
- `deploy/grafana/docker-compose.yml`
- `deploy/grafana/prometheus.yml`
- `deploy/grafana/provisioning/dashboards/agy-boxes.json`
- `deploy/grafana/provisioning/dashboards/dashboards.yaml`
- `deploy/grafana/provisioning/datasources/prometheus.yaml`
- `scripts/start-grafana.sh`

## Out of Scope
- Modifying other providers (OpenAI, Claude, etc.)
- Modifying core database schemas
