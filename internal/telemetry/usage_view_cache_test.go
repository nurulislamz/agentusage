package telemetry

import (
	"context"
	"testing"
	"time"

	"github.com/nurulislamz/agentusage/internal/core"
)

// TestLoadUsageViewCached_ColdStart verifies that a cold cache reports a
// miss but still returns the current fingerprint so the caller can populate
// the cache after building the aggregate.
func TestLoadUsageViewCached_ColdStart(t *testing.T) {
	globalUsageViewCache.reset()
	_, db, store := openUsageViewRawTestStore(t)
	mustIngestUsageEvent(t, store, IngestRequest{
		SourceSystem:  SourceSystem("opencode"),
		SourceChannel: SourceChannelHook,
		OccurredAt:    time.Date(2026, 5, 17, 12, 0, 0, 0, time.UTC),
		ProviderID:    "openrouter",
		AccountID:     "openrouter",
		AgentName:     "opencode",
		EventType:     EventTypeMessageUsage,
		SessionID:     "sess-cold",
		MessageID:     "msg-cold",
		ModelRaw:      "qwen/qwen3-coder-flash",
		TokenUsage: core.TokenUsage{
			InputTokens:  int64Ptr(10),
			OutputTokens: int64Ptr(5),
			TotalTokens:  int64Ptr(15),
			Requests:     int64Ptr(1),
		},
	}, "ingest cold event")

	filter := usageFilter{ProviderIDs: []string{"openrouter"}}
	agg, fp, hit, err := loadUsageViewCached(context.Background(), db, "test-ns", filter)
	if err != nil {
		t.Fatalf("cache lookup: %v", err)
	}
	if hit {
		t.Fatal("expected cache miss on cold start")
	}
	if agg != nil {
		t.Fatal("expected nil agg on miss")
	}
	if fp.empty() {
		t.Fatalf("expected non-zero fingerprint on cold start, got %+v", fp)
	}
}

// TestLoadUsageViewCached_HitWhenFingerprintMatches stores an agg then verifies
// a subsequent lookup with unchanged data returns a cache hit.
func TestLoadUsageViewCached_HitWhenFingerprintMatches(t *testing.T) {
	globalUsageViewCache.reset()
	_, db, store := openUsageViewRawTestStore(t)
	mustIngestUsageEvent(t, store, IngestRequest{
		SourceSystem:  SourceSystem("opencode"),
		SourceChannel: SourceChannelHook,
		OccurredAt:    time.Date(2026, 5, 17, 12, 0, 0, 0, time.UTC),
		ProviderID:    "openrouter",
		AccountID:     "openrouter",
		AgentName:     "opencode",
		EventType:     EventTypeMessageUsage,
		SessionID:     "sess-hit",
		MessageID:     "msg-hit",
		ModelRaw:      "qwen/qwen3-coder-flash",
		TokenUsage: core.TokenUsage{
			InputTokens:  int64Ptr(10),
			OutputTokens: int64Ptr(5),
			TotalTokens:  int64Ptr(15),
			Requests:     int64Ptr(1),
		},
	}, "ingest hit event")

	filter := usageFilter{ProviderIDs: []string{"openrouter"}}
	_, fp, _, err := loadUsageViewCached(context.Background(), db, "test-ns", filter)
	if err != nil {
		t.Fatalf("cache lookup (probe): %v", err)
	}
	expectedAgg := &telemetryUsageAgg{EventCount: 1, LastOccurred: "2026-05-17T12:00:00Z"}
	storeUsageViewCache("test-ns", filter, expectedAgg, fp)

	agg, _, hit, err := loadUsageViewCached(context.Background(), db, "test-ns", filter)
	if err != nil {
		t.Fatalf("cache lookup (hit): %v", err)
	}
	if !hit {
		t.Fatal("expected cache hit when fingerprint unchanged")
	}
	if agg == nil || agg.EventCount != 1 {
		t.Fatalf("expected stored agg to be returned, got %+v", agg)
	}
}

// TestLoadUsageViewCached_InvalidatedOnNewEvent verifies that ingesting a new
// event after caching reports a miss because MAX(rowid) shifts.
func TestLoadUsageViewCached_InvalidatedOnNewEvent(t *testing.T) {
	globalUsageViewCache.reset()
	_, db, store := openUsageViewRawTestStore(t)
	mustIngestUsageEvent(t, store, IngestRequest{
		SourceSystem:  SourceSystem("opencode"),
		SourceChannel: SourceChannelHook,
		OccurredAt:    time.Date(2026, 5, 17, 12, 0, 0, 0, time.UTC),
		ProviderID:    "openrouter",
		AccountID:     "openrouter",
		AgentName:     "opencode",
		EventType:     EventTypeMessageUsage,
		SessionID:     "sess-inv-1",
		MessageID:     "msg-inv-1",
		ModelRaw:      "qwen/qwen3-coder-flash",
		TokenUsage: core.TokenUsage{
			InputTokens: int64Ptr(10),
			TotalTokens: int64Ptr(10),
			Requests:    int64Ptr(1),
		},
	}, "ingest first event")

	filter := usageFilter{ProviderIDs: []string{"openrouter"}}
	_, fp, _, err := loadUsageViewCached(context.Background(), db, "test-ns", filter)
	if err != nil {
		t.Fatalf("cache lookup (probe): %v", err)
	}
	storeUsageViewCache("test-ns", filter, &telemetryUsageAgg{EventCount: 1}, fp)

	// Ingest a second event — fingerprint changes.
	mustIngestUsageEvent(t, store, IngestRequest{
		SourceSystem:  SourceSystem("opencode"),
		SourceChannel: SourceChannelHook,
		OccurredAt:    time.Date(2026, 5, 17, 12, 1, 0, 0, time.UTC),
		ProviderID:    "openrouter",
		AccountID:     "openrouter",
		AgentName:     "opencode",
		EventType:     EventTypeMessageUsage,
		SessionID:     "sess-inv-2",
		MessageID:     "msg-inv-2",
		ModelRaw:      "qwen/qwen3-coder-flash",
		TokenUsage: core.TokenUsage{
			InputTokens: int64Ptr(20),
			TotalTokens: int64Ptr(20),
			Requests:    int64Ptr(1),
		},
	}, "ingest second event")

	_, _, hit, err := loadUsageViewCached(context.Background(), db, "test-ns", filter)
	if err != nil {
		t.Fatalf("cache lookup after ingest: %v", err)
	}
	if hit {
		t.Fatal("expected cache miss after new event ingest (fingerprint changed)")
	}
}

// TestLoadUsageViewCached_InvalidatedOnPrune verifies that deleting rows (as
// retention pruning does) drops COUNT(*) and invalidates the cached entry.
func TestLoadUsageViewCached_InvalidatedOnPrune(t *testing.T) {
	globalUsageViewCache.reset()
	_, db, store := openUsageViewRawTestStore(t)
	mustIngestUsageEvent(t, store, IngestRequest{
		SourceSystem:  SourceSystem("opencode"),
		SourceChannel: SourceChannelHook,
		OccurredAt:    time.Date(2026, 5, 17, 12, 0, 0, 0, time.UTC),
		ProviderID:    "openrouter",
		AccountID:     "openrouter",
		AgentName:     "opencode",
		EventType:     EventTypeMessageUsage,
		SessionID:     "sess-prune",
		MessageID:     "msg-prune",
		ModelRaw:      "qwen/qwen3-coder-flash",
		TokenUsage: core.TokenUsage{
			InputTokens: int64Ptr(10),
			TotalTokens: int64Ptr(10),
			Requests:    int64Ptr(1),
		},
	}, "ingest prune event")

	filter := usageFilter{ProviderIDs: []string{"openrouter"}}
	_, fp, _, err := loadUsageViewCached(context.Background(), db, "test-ns", filter)
	if err != nil {
		t.Fatalf("cache lookup (probe): %v", err)
	}
	storeUsageViewCache("test-ns", filter, &telemetryUsageAgg{EventCount: 1}, fp)

	// Simulate prune: delete the row.
	if _, err := db.Exec("DELETE FROM usage_events WHERE provider_id = 'openrouter'"); err != nil {
		t.Fatalf("prune events: %v", err)
	}

	_, _, hit, err := loadUsageViewCached(context.Background(), db, "test-ns", filter)
	if err != nil {
		t.Fatalf("cache lookup after prune: %v", err)
	}
	if hit {
		t.Fatal("expected cache miss after prune (count changed)")
	}
}

// TestLoadUsageViewCached_EmptyNamespaceDisablesCache verifies that callers
// that pass an empty namespace (e.g. tests) bypass the cache entirely.
func TestLoadUsageViewCached_EmptyNamespaceDisablesCache(t *testing.T) {
	globalUsageViewCache.reset()
	_, db, store := openUsageViewRawTestStore(t)
	mustIngestUsageEvent(t, store, IngestRequest{
		SourceSystem:  SourceSystem("opencode"),
		SourceChannel: SourceChannelHook,
		OccurredAt:    time.Date(2026, 5, 17, 12, 0, 0, 0, time.UTC),
		ProviderID:    "openrouter",
		AccountID:     "openrouter",
		AgentName:     "opencode",
		EventType:     EventTypeMessageUsage,
		SessionID:     "sess-bypass",
		MessageID:     "msg-bypass",
		ModelRaw:      "qwen/qwen3-coder-flash",
		TokenUsage: core.TokenUsage{
			InputTokens: int64Ptr(10),
			TotalTokens: int64Ptr(10),
			Requests:    int64Ptr(1),
		},
	}, "ingest bypass event")

	filter := usageFilter{ProviderIDs: []string{"openrouter"}}
	storeUsageViewCache("", filter, &telemetryUsageAgg{EventCount: 999}, usageViewFingerprint{maxRowID: 1, count: 1})
	agg, _, hit, err := loadUsageViewCached(context.Background(), db, "", filter)
	if err != nil {
		t.Fatalf("cache lookup with empty namespace: %v", err)
	}
	if hit {
		t.Fatal("empty namespace should always miss")
	}
	if agg != nil {
		t.Fatalf("empty namespace should return nil agg, got %+v", agg)
	}
}

// TestLoadUsageViewCached_InvalidatedOnDedupEnrich verifies that an in-place
// UPDATE from enrichEventByDedupKey (same rowid, same COUNT(*)) still misses
// the cache because token/cost sums changed.
func TestLoadUsageViewCached_InvalidatedOnDedupEnrich(t *testing.T) {
	globalUsageViewCache.reset()
	_, db, store := openUsageViewRawTestStore(t)
	jsonl := IngestRequest{
		SourceSystem:  SourceSystem("claude_code"),
		SourceChannel: SourceChannelJSONL,
		OccurredAt:    time.Date(2026, 5, 17, 12, 0, 0, 0, time.UTC),
		ProviderID:    "claude_code",
		AccountID:     "claude_code",
		AgentName:     "claude",
		EventType:     EventTypeMessageUsage,
		SessionID:     "sess-enrich",
		MessageID:     "msg-enrich",
		ModelRaw:      "claude-sonnet",
		TokenUsage: core.TokenUsage{
			InputTokens:  int64Ptr(10),
			OutputTokens: int64Ptr(0),
			TotalTokens:  int64Ptr(10),
			CostUSD:      float64Ptr(0.01),
			Requests:     int64Ptr(1),
		},
	}
	mustIngestUsageEvent(t, store, jsonl, "ingest jsonl event")

	filter := usageFilter{ProviderIDs: []string{"claude_code"}, AccountID: "claude_code"}
	_, fp, _, err := loadUsageViewCached(context.Background(), db, "test-ns", filter)
	if err != nil {
		t.Fatalf("cache lookup (probe): %v", err)
	}
	storeUsageViewCache("test-ns", filter, &telemetryUsageAgg{EventCount: 1, Activity: telemetryActivityAgg{TotalTokens: 10, TotalCost: 0.01}}, fp)

	hook := jsonl
	hook.SourceChannel = SourceChannelHook
	hook.TokenUsage = core.TokenUsage{
		InputTokens:  int64Ptr(80),
		OutputTokens: int64Ptr(20),
		TotalTokens:  int64Ptr(100),
		CostUSD:      float64Ptr(0.50),
		Requests:     int64Ptr(1),
	}
	result, err := store.Ingest(context.Background(), hook)
	if err != nil {
		t.Fatalf("hook enrich: %v", err)
	}
	if !result.Deduped {
		t.Fatal("expected hook ingest to be deduped")
	}

	var rowCount, totalTokens int64
	if err := db.QueryRow(`SELECT COUNT(*), COALESCE(SUM(total_tokens), 0) FROM usage_events`).Scan(&rowCount, &totalTokens); err != nil {
		t.Fatalf("query after enrich: %v", err)
	}
	if rowCount != 1 {
		t.Fatalf("canonical row count = %d, want 1 (in-place update)", rowCount)
	}
	if totalTokens != 100 {
		t.Fatalf("canonical total_tokens = %d, want 100", totalTokens)
	}

	_, _, hit, err := loadUsageViewCached(context.Background(), db, "test-ns", filter)
	if err != nil {
		t.Fatalf("cache lookup after enrich: %v", err)
	}
	if hit {
		t.Fatal("expected cache miss after in-place dedup enrich (fingerprint must include token/cost sums)")
	}
}

// TestApplyCanonicalUsageView_RefreshesAfterDedupEnrich is the user-facing
// path: collector JSONL warms the view cache, then a hook for the same
// message_id overlays higher tokens/cost. The next apply must not keep the
// pre-enrich window_tokens / window_cost.
func TestApplyCanonicalUsageView_RefreshesAfterDedupEnrich(t *testing.T) {
	globalUsageViewCache.reset()
	dbPath, store := openUsageViewTestStore(t)

	jsonl := IngestRequest{
		SourceSystem:  SourceSystem("claude_code"),
		SourceChannel: SourceChannelJSONL,
		OccurredAt:    time.Date(2026, 5, 17, 12, 0, 0, 0, time.UTC),
		ProviderID:    "claude_code",
		AccountID:     "claude_code",
		AgentName:     "claude",
		EventType:     EventTypeMessageUsage,
		SessionID:     "sess-enrich-view",
		MessageID:     "msg-enrich-view",
		ModelRaw:      "claude-sonnet",
		TokenUsage: core.TokenUsage{
			InputTokens:  int64Ptr(10),
			OutputTokens: int64Ptr(0),
			TotalTokens:  int64Ptr(10),
			CostUSD:      float64Ptr(0.01),
			Requests:     int64Ptr(1),
		},
	}
	mustIngestUsageEvent(t, store, jsonl, "ingest jsonl event")

	snaps := map[string]core.UsageSnapshot{
		"claude_code": {
			ProviderID: "claude_code",
			AccountID:  "claude_code",
			Metrics:    map[string]core.Metric{},
		},
	}
	first, err := applyCanonicalUsageViewForTest(context.Background(), dbPath, snaps)
	if err != nil {
		t.Fatalf("first view: %v", err)
	}
	if got := metricUsed(first["claude_code"].Metrics["window_tokens"]); got != 10 {
		t.Fatalf("after jsonl window_tokens = %v, want 10", got)
	}
	if got := metricUsed(first["claude_code"].Metrics["window_cost"]); got != 0.01 {
		t.Fatalf("after jsonl window_cost = %v, want 0.01", got)
	}

	hook := jsonl
	hook.SourceChannel = SourceChannelHook
	hook.TokenUsage = core.TokenUsage{
		InputTokens:  int64Ptr(80),
		OutputTokens: int64Ptr(20),
		TotalTokens:  int64Ptr(100),
		CostUSD:      float64Ptr(0.50),
		Requests:     int64Ptr(1),
	}
	result, err := store.Ingest(context.Background(), hook)
	if err != nil {
		t.Fatalf("hook enrich: %v", err)
	}
	if !result.Deduped {
		t.Fatal("expected hook ingest to be deduped")
	}

	second, err := applyCanonicalUsageViewForTest(context.Background(), dbPath, snaps)
	if err != nil {
		t.Fatalf("second view: %v", err)
	}
	if got := metricUsed(second["claude_code"].Metrics["window_tokens"]); got != 100 {
		t.Fatalf("after hook enrich window_tokens = %v, want 100 (stale cache would keep 10)", got)
	}
	if got := metricUsed(second["claude_code"].Metrics["window_cost"]); got != 0.50 {
		t.Fatalf("after hook enrich window_cost = %v, want 0.50 (stale cache would keep 0.01)", got)
	}
}
