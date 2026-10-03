package telemetry

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/nurulislamz/agentusage/internal/core"
)

func TestDecodeStoredLimitSnapshot_RejectsEmptyPayload(t *testing.T) {
	_, ok := decodeStoredLimitSnapshot("antigravity", "antigravity-mohammed", "{}", time.Now().UTC().Format(time.RFC3339Nano))
	if ok {
		t.Fatal("expected empty {} payload to be rejected")
	}
}

func TestDecodeStoredLimitSnapshot_RejectsMetriclessStatusOK(t *testing.T) {
	payload := `{"snapshot":{"provider_id":"cursor","account_id":"cursor","status":"OK","metrics":{},"attributes":{"email":"user@example.com"}}}`
	_, ok := decodeStoredLimitSnapshot("cursor", "cursor", payload, time.Now().UTC().Format(time.RFC3339Nano))
	if ok {
		t.Fatal("expected metric-less StatusOK payload to be rejected")
	}
}

func TestBuildLimitSnapshotRequests_SkipsMetriclessStatusOK(t *testing.T) {
	reqs := BuildLimitSnapshotRequests(map[string]core.UsageSnapshot{
		"cursor": {
			ProviderID: "cursor",
			AccountID:  "cursor",
			Timestamp:  time.Now().UTC(),
			Status:     core.StatusOK,
			Metrics:    map[string]core.Metric{},
			Attributes: map[string]string{"email": "user@example.com"},
		},
	})
	if len(reqs) != 0 {
		t.Fatalf("requests = %d, want 0 for metric-less StatusOK", len(reqs))
	}
}

func TestHydrateSkipsNewerMetriclessStatusOKRoot(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "telemetry.db")
	store, err := OpenStore(dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	ingestor := NewQuotaSnapshotIngestor(store)
	used := 42.0
	goodTS := time.Date(2026, 10, 3, 10, 0, 0, 0, time.UTC)
	if err := ingestor.Ingest(context.Background(), map[string]core.UsageSnapshot{
		"cursor": {
			ProviderID: "cursor",
			AccountID:  "cursor",
			Timestamp:  goodTS,
			Status:     core.StatusOK,
			Metrics: map[string]core.Metric{
				"plan_percent_used": {Used: &used, Limit: core.Float64Ptr(100), Remaining: core.Float64Ptr(58), Unit: "%", Window: "monthly"},
			},
		},
	}); err != nil {
		t.Fatalf("good ingest: %v", err)
	}

	// Simulate a historical poison row already in the DB (newer occurred_at,
	// StatusOK, no metrics) — BuildLimitSnapshotRequests now refuses these,
	// so insert the envelope directly.
	poisonTS := goodTS.Add(time.Hour)
	poisonPayload := `{"snapshot":{"provider_id":"cursor","account_id":"cursor","status":"OK","message":"","metrics":{},"resets":{},"attributes":{"email":"user@example.com"},"diagnostics":{}}}`
	rawID := "raw-poison-cursor"
	ingestedAt := poisonTS.Format(time.RFC3339Nano)
	if _, err := store.db.Exec(`
		INSERT INTO usage_raw_events (
			raw_event_id, ingested_at, source_system, source_channel, source_schema_version,
			source_payload, source_payload_hash, workspace_id, agent_session_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, rawID, ingestedAt, string(SourceSystemPoller), string(SourceChannelAPI), providerSnapshotSchemaVersion,
		poisonPayload, "poison", "", ""); err != nil {
		t.Fatalf("insert poison raw: %v", err)
	}
	if _, err := store.db.Exec(`
		INSERT INTO usage_events (
			event_id, occurred_at, agent_name, event_type, status, dedup_key,
			raw_event_id, normalization_version, provider_id, account_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, "evt-poison-cursor", poisonTS.Format(time.RFC3339Nano), "provider_poller",
		string(EventTypeLimitSnapshot), "ok", "dedup-poison-cursor",
		rawID, "v1", "cursor", "cursor"); err != nil {
		t.Fatalf("insert poison event: %v", err)
	}

	templates := map[string]core.UsageSnapshot{
		"cursor": {ProviderID: "cursor", AccountID: "cursor", Status: core.StatusUnknown},
	}
	got, err := ApplyCanonicalTelemetryViewWithOptions(context.Background(), dbPath, templates, ReadModelOptions{})
	if err != nil {
		t.Fatalf("read model: %v", err)
	}
	snap := got["cursor"]
	metric, ok := snap.Metrics["plan_percent_used"]
	if !ok || metric.Used == nil || *metric.Used != 42 {
		t.Fatalf("expected prior plan_percent_used=42 after skipping poison root, got %#v", snap.Metrics)
	}
}

func TestQuotaSnapshotIngest_RepairsEmptyDedupedPayload(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "telemetry.db")
	store, err := OpenStore(dbPath)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	ingestor := NewQuotaSnapshotIngestor(store)
	ts := time.Date(2026, 8, 29, 17, 13, 13, 0, time.UTC)

	// Seed a same-second turn/dedup key with a usable (but soon-corrupted) row,
	// then force the stored payload to {} the way broken production rows look.
	seedRem := 1.0
	seed := map[string]core.UsageSnapshot{
		"antigravity-mohammed": {
			ProviderID: "antigravity",
			AccountID:  "antigravity-mohammed",
			Timestamp:  ts,
			Status:     core.StatusOK,
			Metrics: map[string]core.Metric{
				"quota_gemini_weekly": {Remaining: &seedRem, Unit: "%", Window: "7d"},
			},
		},
	}
	if err := ingestor.Ingest(context.Background(), seed); err != nil {
		t.Fatalf("seed ingest: %v", err)
	}

	if _, err := store.db.Exec(`UPDATE usage_raw_events SET source_payload='{}', source_payload_hash='00'`); err != nil {
		t.Fatalf("force empty payload: %v", err)
	}

	rem := 96.0
	full := map[string]core.UsageSnapshot{
		"antigravity-mohammed": {
			ProviderID: "antigravity",
			AccountID:  "antigravity-mohammed",
			Timestamp:  ts,
			Status:     core.StatusOK,
			Metrics: map[string]core.Metric{
				"quota_gemini_weekly": {Remaining: &rem, Unit: "%", Window: "7d"},
			},
		},
	}
	if err := ingestor.Ingest(context.Background(), full); err != nil {
		t.Fatalf("repair ingest: %v", err)
	}

	templates := map[string]core.UsageSnapshot{
		"antigravity-mohammed": {
			ProviderID: "antigravity",
			AccountID:  "antigravity-mohammed",
			Status:     core.StatusUnknown,
		},
	}
	got, err := ApplyCanonicalTelemetryViewWithOptions(context.Background(), dbPath, templates, ReadModelOptions{})
	if err != nil {
		t.Fatalf("read model: %v", err)
	}
	snap := got["antigravity-mohammed"]
	if snap.Status != core.StatusOK {
		t.Fatalf("status = %q, want OK after payload repair", snap.Status)
	}
	if _, ok := snap.Metrics["quota_gemini_weekly"]; !ok {
		t.Fatalf("expected quota metrics after payload repair, got %#v", snap.Metrics)
	}
}
