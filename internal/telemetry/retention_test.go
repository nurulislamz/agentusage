package telemetry

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPruneCorruptBackupsKeepsNewest(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "telemetry.db")

	// Three corrupt snapshots with increasing modtimes, plus a -wal sidecar
	// for the newest that should survive alongside it.
	mk := func(name string, age time.Duration) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
		mt := time.Now().Add(-age)
		if err := os.Chtimes(p, mt, mt); err != nil {
			t.Fatal(err)
		}
		return p
	}
	oldest := mk("telemetry.db.corrupt.20260101T000000", 72*time.Hour)
	middle := mk("telemetry.db.corrupt.bak", 48*time.Hour)
	newest := mk("telemetry.db.corrupt.20260601T000000", 1*time.Hour)
	// An unrelated file must be left untouched.
	keepMe := mk("telemetry.db", 0)

	pruneCorruptBackups(dbPath, 1)

	if _, err := os.Stat(newest); err != nil {
		t.Errorf("newest corrupt backup should be kept: %v", err)
	}
	if _, err := os.Stat(keepMe); err != nil {
		t.Errorf("live db file must not be touched: %v", err)
	}
	for _, gone := range []string{oldest, middle} {
		if _, err := os.Stat(gone); !os.IsNotExist(err) {
			t.Errorf("expected %s to be removed, err=%v", filepath.Base(gone), err)
		}
	}
}

func TestPruneOldEventsBatchedRemovesBacklog(t *testing.T) {
	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "telemetry.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	// Use the same WAL + single-connection setup the production store uses.
	// Without it the bulk insert below runs as thousands of FULL-synced
	// auto-commit transactions, which on Windows is slow enough to blow the
	// package test timeout.
	if err := configureSQLiteConnection(db); err != nil {
		t.Fatalf("configure: %v", err)
	}

	store := NewStore(db)
	if err := store.Init(context.Background()); err != nil {
		t.Fatalf("init: %v", err)
	}

	// Insert a backlog larger than one batch so the loop must iterate, split
	// across "old" (beyond retention) and "recent" (within retention).
	const oldCount = pruneEventsBatch + 1500
	const recentCount = 50
	// Batch every insert into a single transaction. The prune logic under test
	// is unaffected by how rows arrive, and one commit instead of ~2n keeps the
	// test fast and deadlock-free across platforms.
	insert := func(n int, occurredAt time.Time) {
		tx, err := db.Begin()
		if err != nil {
			t.Fatalf("begin tx: %v", err)
		}
		for i := 0; i < n; i++ {
			id := fmt.Sprintf("%s-%d", occurredAt.Format("20060102"), i)
			ts := occurredAt.Format(time.RFC3339Nano)
			if _, err := tx.Exec(`INSERT INTO usage_raw_events
				(raw_event_id, ingested_at, source_system, source_channel, source_schema_version, source_payload, source_payload_hash)
				VALUES (?, ?, 'test', 'api', 'v1', '{}', ?)`, id, ts, id); err != nil {
				t.Fatalf("insert raw: %v", err)
			}
			if _, err := tx.Exec(`INSERT INTO usage_events
				(event_id, occurred_at, provider_id, account_id, agent_name, event_type, status, dedup_key, raw_event_id, normalization_version)
				VALUES (?, ?, 'p', 'a', 'test', 'tool_usage', 'ok', ?, ?, 'v1')`, id, ts, id, id); err != nil {
				t.Fatalf("insert event: %v", err)
			}
		}
		if err := tx.Commit(); err != nil {
			t.Fatalf("commit tx: %v", err)
		}
	}
	insert(oldCount, time.Now().Add(-90*24*time.Hour))
	insert(recentCount, time.Now().Add(-1*time.Hour))

	deleted, complete, err := store.PruneOldEvents(context.Background(), 30, "9999-12-31")
	if err != nil {
		t.Fatalf("prune: %v", err)
	}
	if !complete {
		t.Errorf("expected backlog to be fully drained")
	}
	if deleted != oldCount {
		t.Errorf("deleted = %d, want %d", deleted, oldCount)
	}

	var remaining int
	if err := db.QueryRow(`SELECT COUNT(*) FROM usage_events`).Scan(&remaining); err != nil {
		t.Fatal(err)
	}
	if remaining != recentCount {
		t.Errorf("remaining = %d, want %d (recent events must survive)", remaining, recentCount)
	}
}

func TestPruneRawEventPayloads_WipesAnalyticsDimensions(t *testing.T) {
	// Regression lock: analytics language / code-stats still read file paths and
	// line counts from usage_raw_events.source_payload. Clearing that JSON after
	// 1h (former daemon collect-loop behavior) permanently zeroed those views
	// for every event older than an hour.
	_, db, store := openUsageViewRawTestStore(t)

	mustIngestUsageEvent(t, store, IngestRequest{
		SourceSystem:  SourceSystem("claude_code"),
		SourceChannel: SourceChannelHook,
		OccurredAt:    time.Now().UTC().Add(-90 * time.Minute),
		ProviderID:    "anthropic",
		AccountID:     "anthropic-main",
		AgentName:     "claude_code",
		EventType:     EventTypeToolUsage,
		ToolName:      "Edit",
		ToolCallID:    "tool-edit-1",
		SessionID:     "sess-analytics",
		Status:        EventStatusOK,
		Payload: map[string]any{
			"file":          "/workspace/internal/daemon/server.go",
			"lines_added":   12,
			"lines_removed": 3,
		},
	}, "ingest tool event with analytics payload")

	// Make the raw row eligible for a 1-hour payload prune.
	if _, err := db.Exec(`UPDATE usage_raw_events SET ingested_at = datetime('now', '-2 hours')`); err != nil {
		t.Fatalf("backdate ingested_at: %v", err)
	}

	filter := usageFilter{ProviderIDs: []string{"anthropic"}}
	langsBefore, err := queryLanguageAgg(context.Background(), db, filter)
	if err != nil {
		t.Fatalf("language agg before prune: %v", err)
	}
	if len(langsBefore) == 0 || langsBefore[0].Language != "go" {
		t.Fatalf("expected go language agg before prune, got %+v", langsBefore)
	}
	codeBefore, err := queryCodeStatsAgg(context.Background(), db, filter)
	if err != nil {
		t.Fatalf("code stats before prune: %v", err)
	}
	if codeBefore.FilesChanged < 1 || codeBefore.LinesAdded != 12 || codeBefore.LinesRemoved != 3 {
		t.Fatalf("expected code stats before prune, got %+v", codeBefore)
	}

	pruned, err := store.PruneRawEventPayloads(context.Background(), 1, 1000)
	if err != nil {
		t.Fatalf("PruneRawEventPayloads: %v", err)
	}
	if pruned != 1 {
		t.Fatalf("pruned = %d, want 1", pruned)
	}

	var payload string
	if err := db.QueryRow(`SELECT source_payload FROM usage_raw_events LIMIT 1`).Scan(&payload); err != nil {
		t.Fatalf("read payload: %v", err)
	}
	if payload != "{}" {
		t.Fatalf("source_payload = %q, want {}", payload)
	}

	langsAfter, err := queryLanguageAgg(context.Background(), db, filter)
	if err != nil {
		t.Fatalf("language agg after prune: %v", err)
	}
	if len(langsAfter) != 0 {
		t.Fatalf("language agg should be empty after payload prune, got %+v", langsAfter)
	}
	codeAfter, err := queryCodeStatsAgg(context.Background(), db, filter)
	if err != nil {
		t.Fatalf("code stats after prune: %v", err)
	}
	if codeAfter.FilesChanged != 0 || codeAfter.LinesAdded != 0 || codeAfter.LinesRemoved != 0 {
		t.Fatalf("code stats should be zero after payload prune, got %+v", codeAfter)
	}
}

func TestPruneOldEventsCancelledContextReturnsProgress(t *testing.T) {
	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "telemetry.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()
	store := NewStore(db)
	if err := store.Init(context.Background()); err != nil {
		t.Fatalf("init: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled

	// Should return cleanly (0, not-complete, nil) rather than erroring on a dead context.
	deleted, complete, err := store.PruneOldEvents(ctx, 30, "9999-12-31")
	if err != nil {
		t.Errorf("cancelled prune should not error, got %v", err)
	}
	if deleted != 0 {
		t.Errorf("deleted = %d, want 0 on pre-cancelled context", deleted)
	}
	if complete {
		t.Errorf("complete should be false when context is cancelled before draining")
	}
}
