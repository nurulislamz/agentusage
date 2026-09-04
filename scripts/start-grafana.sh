#!/usr/bin/env bash
# ==============================================================================
# start-grafana.sh: Manage Grafana + Prometheus for agentUsage agy-box monitor
# ==============================================================================
set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_FILE="$REPO_DIR/deploy/grafana/docker-compose.yml"
CMD="${1:-start}"

case "$CMD" in
  start|up)
    echo "🚀 Starting Prometheus and Grafana for agentUsage..."
    docker compose -f "$COMPOSE_FILE" up -d
    echo ""
    echo "✅ Grafana is running!"
    echo "📊 Dashboard URL: http://127.0.0.1:3000/d/agy-box-pings/antigravity-agy-box-ping-monitor"
    echo "📈 Prometheus:    http://127.0.0.1:9090"
    echo ""
    echo "Tip: Run '$0 expose' to generate a public cloudflared tunnel URL."
    ;;

  stop|down)
    echo "🛑 Stopping Grafana and Prometheus..."
    docker compose -f "$COMPOSE_FILE" down
    echo "Stopped."
    ;;

  restart)
    echo "🔄 Restarting..."
    docker compose -f "$COMPOSE_FILE" down
    docker compose -f "$COMPOSE_FILE" up -d
    echo "Restarted."
    ;;

  status)
    docker compose -f "$COMPOSE_FILE" ps
    ;;

  expose|tunnel)
    if ! command -v cloudflared >/dev/null 2>&1; then
      echo "Error: cloudflared binary not found." >&2
      exit 1
    fi
    echo "🌐 Starting cloudflared tunnel for Grafana (http://127.0.0.1:3000)..."
    exec cloudflared tunnel --url http://127.0.0.1:3000
    ;;

  *)
    echo "Usage: $0 {start|stop|restart|status|expose}"
    exit 1
    ;;
esac
