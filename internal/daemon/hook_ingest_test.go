package daemon

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/nurulislamz/agentusage/internal/telemetry"
)

func TestIngestHookLocally_IngestsHookPayload(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "telemetry.db")
	spoolDir := filepath.Join(t.TempDir(), "spool")
	payload := []byte(`{"hook":"chat.message","timestamp":"2026-02-26T20:00:00Z","input":{"sessionID":"sess-1","agent":"main","messageID":"turn-1","variant":"default","model":{"providerID":"openrouter","modelID":"openai/gpt-oss-20b"}},"output":{"message":{"id":"msg-1","sessionID":"sess-1","role":"assistant"},"route":{"provider_name":"DeepInfra"},"usage":{"input_tokens":12,"output_tokens":4,"total_tokens":16,"cost_usd":0.00012}}}`)

	resp, err := IngestHookLocally(context.Background(), "opencode", "openrouter", payload, dbPath, spoolDir, false)
	if err != nil {
		t.Fatalf("ingest hook locally: %v", err)
	}
	if resp.Ingested == 0 {
		t.Fatalf("ingested = %d, want >0", resp.Ingested)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	var hookRawCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM usage_raw_events WHERE source_system='opencode' AND source_channel='hook'`).Scan(&hookRawCount); err != nil {
		t.Fatalf("query hook raw count: %v", err)
	}
	if hookRawCount == 0 {
		t.Fatalf("hook raw rows = %d, want >0", hookRawCount)
	}

	var upstream string
	if err := db.QueryRow(`SELECT COALESCE(json_extract(source_payload, '$._normalized.upstream_provider'), '') FROM usage_raw_events WHERE source_system='opencode' AND source_channel='hook' ORDER BY ingested_at DESC LIMIT 1`).Scan(&upstream); err != nil {
		t.Fatalf("query upstream provider: %v", err)
	}
	if upstream != "deepinfra" {
		t.Fatalf("upstream provider = %q, want deepinfra", upstream)
	}
}

func TestIngestHookLocally_SpoolOnly(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "telemetry.db")
	spoolDir := filepath.Join(t.TempDir(), "spool")
	payload := []byte(`{"hook":"tool.execute.after","timestamp":"2026-02-26T20:00:00Z","input":{"tool":"glob","sessionID":"sess-1","callID":"tool-1"},"output":{"title":"Glob"}}`)

	resp, err := IngestHookLocally(context.Background(), "opencode", "openrouter", payload, dbPath, spoolDir, true)
	if err != nil {
		t.Fatalf("spool-only ingest hook locally: %v", err)
	}
	if resp.Enqueued == 0 {
		t.Fatalf("enqueued = %d, want >0", resp.Enqueued)
	}
	if resp.Ingested != 0 {
		t.Fatalf("ingested = %d, want 0 in spool-only mode", resp.Ingested)
	}

	// Spool-only must not touch the SQLite store (avoids -shm races with a live daemon).
	if _, err := os.Stat(dbPath); !os.IsNotExist(err) {
		t.Fatalf("telemetry DB path %s should not be created in spool-only mode (stat err=%v)", dbPath, err)
	}

	entries, err := os.ReadDir(spoolDir)
	if err != nil {
		t.Fatalf("read spool dir: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected spool records after spool-only enqueue")
	}
}

func TestParseHookRequests_UnknownSource(t *testing.T) {
	_, err := ParseHookRequests("unknown_nonexistent_tool", "", []byte(`{}`))
	if err == nil {
		t.Error("expected error for unknown hook source")
	}
}

func TestIngestParsedHookLocally_EmptyRequests(t *testing.T) {
	resp, err := ingestParsedHookLocally(context.Background(), HookParseResult{
		SourceName: "opencode",
		Requests:   nil,
	}, "", "", false)
	if err != nil {
		t.Fatalf("ingestParsedHookLocally with empty requests error: %v", err)
	}
	if resp.Source != "opencode" || resp.Enqueued != 0 {
		t.Errorf("resp = %+v, want empty 0 enqueued", resp)
	}
}

func TestIngestParsedHookLocally_InvalidPaths(t *testing.T) {
	parsed := HookParseResult{
		SourceName: "opencode",
		Requests: []telemetry.IngestRequest{
			{
				SourceSystem:  "opencode",
				SourceChannel: "hook",
				AccountID:     "work",
				OccurredAt:    time.Now().UTC(),
				SessionID:     "s1",
			},
		},
	}

	// 1. Invalid DB Path
	_, err := ingestParsedHookLocally(context.Background(), parsed, "/dev/null/forbidden/db.sqlite", "", false)
	if err == nil {
		t.Error("expected error for invalid DB path")
	}

	// 2. Invalid Spool Path with spoolOnly
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "telemetry.db")
	_ = telemetry.NewSpool(filepath.Join(tempDir, "spool"))
	_, err = ingestParsedHookLocally(context.Background(), parsed, dbPath, "/dev/null/forbidden/spool", true)
	if err == nil {
		t.Error("expected error for invalid spool dir in spool-only mode")
	}

	// 3. Direct ingest with cancelled context triggers retries
	ctxCancel, cancel := context.WithCancel(context.Background())
	cancel()
	resp, _ := ingestParsedHookLocally(ctxCancel, parsed, dbPath, filepath.Join(tempDir, "spool"), false)
	_ = resp
}
