package daemon

import (
	"context"
	"database/sql"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nurulislamz/agentusage/internal/core"
	"github.com/nurulislamz/agentusage/internal/telemetry"
)

func TestPruneTelemetryOrphans_PreservesAnalyticsPayloads(t *testing.T) {
	// Collect-loop orphan pruning used to call PruneRawEventPayloads(1h), which
	// cleared source_payload while analytics still json_extracts language /
	// code-stats fields from it. Ensure pruneTelemetryOrphans no longer does that.
	dbPath := filepath.Join(t.TempDir(), "telemetry.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	store := telemetry.NewStore(db)
	if err := store.Init(context.Background()); err != nil {
		t.Fatalf("init store: %v", err)
	}

	if _, err := store.Ingest(context.Background(), telemetry.IngestRequest{
		SourceSystem:  telemetry.SourceSystem("claude_code"),
		SourceChannel: telemetry.SourceChannelHook,
		OccurredAt:    time.Now().UTC().Add(-2 * time.Hour),
		ProviderID:    "anthropic",
		AccountID:     "anthropic-main",
		AgentName:     "claude_code",
		EventType:     telemetry.EventTypeToolUsage,
		ToolName:      "Edit",
		ToolCallID:    "tool-preserve-1",
		SessionID:     "sess-preserve",
		Status:        telemetry.EventStatusOK,
		Payload: map[string]any{
			"file":          "/workspace/internal/daemon/server.go",
			"lines_added":   4,
			"lines_removed": 1,
		},
	}); err != nil {
		t.Fatalf("ingest: %v", err)
	}
	if _, err := db.Exec(`UPDATE usage_raw_events SET ingested_at = datetime('now', '-2 hours')`); err != nil {
		t.Fatalf("backdate ingested_at: %v", err)
	}

	svc := &Service{
		store:       store,
		logThrottle: core.NewLogThrottle(5, time.Minute),
	}
	svc.pruneTelemetryOrphans(context.Background())

	var payload string
	if err := db.QueryRow(`SELECT source_payload FROM usage_raw_events LIMIT 1`).Scan(&payload); err != nil {
		t.Fatalf("read payload: %v", err)
	}
	if payload == "" || payload == "{}" {
		t.Fatalf("source_payload was cleared by pruneTelemetryOrphans: %q", payload)
	}
	if !strings.Contains(payload, "server.go") || !strings.Contains(payload, "lines_added") {
		t.Fatalf("source_payload missing analytics fields: %s", payload)
	}
}
