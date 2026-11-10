package telemetry

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/nurulislamz/agentusage/internal/core"
)

// usageViewCache caches the fully aggregated telemetryUsageAgg for a given
// filter, keyed by (dbPath, providerIDs, accountID, since, todaySince).
//
// Each entry stores a content fingerprint captured at the time the aggregate
// was built. On lookup a cheap probe query runs the same fingerprint against
// the underlying usage_events table; if it matches we skip the expensive
// materialize + 14 aggregate queries.
//
// The fingerprint covers:
//   - MAX(rowid): inserts invalidate cleanly
//   - COUNT(*): retention pruning / deletes invalidate cleanly
//   - SUM of token/cost/request columns: in-place UPDATEs from
//     enrichEventByDedupKey (hook overlaying a collector JSONL row) also
//     invalidate. Rowid and COUNT do not change on UPDATE, so omitting
//     those sums left window_tokens / window_cost stuck at the pre-enrich
//     totals until another insert, delete, or process restart.
//
// The cache is process-global. It is bounded by a fixed maximum number of
// entries; on overflow the oldest entry by insertion order is dropped. The
// expected cardinality is small (one entry per provider × account × window
// class, typically <50), so a simple cap is sufficient.
const usageViewCacheMaxEntries = 64

type usageViewFingerprint struct {
	maxRowID int64
	count    int64
	tokens   int64
	costE6   int64
	requests int64
}

func (a usageViewFingerprint) equal(b usageViewFingerprint) bool {
	return a.maxRowID == b.maxRowID &&
		a.count == b.count &&
		a.tokens == b.tokens &&
		a.costE6 == b.costE6 &&
		a.requests == b.requests
}

func (a usageViewFingerprint) empty() bool {
	return a.maxRowID == 0 && a.count == 0
}

type usageViewCacheEntry struct {
	agg        *telemetryUsageAgg
	fp         usageViewFingerprint
	insertedAt time.Time
}

type usageViewCacheStore struct {
	mu      sync.Mutex
	entries map[string]*usageViewCacheEntry
	// order tracks insertion order for FIFO eviction; we keep it cheap and
	// simple rather than building a full LRU.
	order []string
}

var globalUsageViewCache = &usageViewCacheStore{
	entries: make(map[string]*usageViewCacheEntry),
}

// usageViewCacheNamespace identifies the database backing a cache entry. It
// flows from ApplyCanonicalTelemetryViewWithOptions down into the loader so
// the cache survives the per-call sql.Open/Close lifecycle.
type usageViewCacheNamespace string

func usageViewCacheKey(namespace usageViewCacheNamespace, filter usageFilter) string {
	providers := strings.Join(normalizeProviderIDs(filter.ProviderIDs), ",")
	since := int64(0)
	if !filter.Since.IsZero() {
		since = filter.Since.UTC().UnixNano()
	}
	today := int64(0)
	if !filter.TodaySince.IsZero() {
		today = filter.TodaySince.UTC().UnixNano()
	}
	return fmt.Sprintf("%s|%s|%s|%d|%d", namespace, providers, strings.TrimSpace(filter.AccountID), since, today)
}

func (s *usageViewCacheStore) lookup(key string) (*usageViewCacheEntry, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.entries[key]
	return entry, ok
}

func (s *usageViewCacheStore) store(key string, entry *usageViewCacheEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.entries[key]; !exists {
		s.order = append(s.order, key)
		for len(s.order) > usageViewCacheMaxEntries {
			drop := s.order[0]
			s.order = s.order[1:]
			delete(s.entries, drop)
		}
	}
	s.entries[key] = entry
}

// reset is exposed for tests so they can run against a clean cache.
func (s *usageViewCacheStore) reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = make(map[string]*usageViewCacheEntry)
	s.order = nil
}

// probeUsageEventsFingerprint runs a cheap identity query against usage_events
// using the same provider/account filter as the materialization step. It
// deliberately ignores the Since cutoff so the probe stays index-driven: the
// cache key already includes Since, so an entry only matches when Since is
// identical between builder and reader.
func probeUsageEventsFingerprint(ctx context.Context, db *sql.DB, filter usageFilter) (usageViewFingerprint, error) {
	providers := normalizeProviderIDs(filter.ProviderIDs)
	if len(providers) == 0 {
		return usageViewFingerprint{}, nil
	}

	args := make([]any, 0, len(providers)+1)
	var b strings.Builder
	b.WriteString(`SELECT
		COALESCE(MAX(rowid), 0),
		COUNT(*),
		COALESCE(SUM(
			COALESCE(total_tokens, 0) +
			COALESCE(input_tokens, 0) +
			COALESCE(output_tokens, 0) +
			COALESCE(reasoning_tokens, 0) +
			COALESCE(cache_read_tokens, 0) +
			COALESCE(cache_write_tokens, 0)
		), 0),
		COALESCE(CAST(ROUND(SUM(COALESCE(cost_usd, 0)) * 1000000) AS INTEGER), 0),
		COALESCE(SUM(COALESCE(requests, 0)), 0)
		FROM usage_events
		WHERE event_type IN ('message_usage', 'tool_usage')
		  AND provider_id IN (`)
	for i, p := range providers {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString("?")
		args = append(args, p)
	}
	b.WriteString(")")

	accountID := strings.TrimSpace(filter.AccountID)
	if accountID != "" {
		b.WriteString(" AND account_id = ?")
		args = append(args, accountID)
	}

	var fp usageViewFingerprint
	if err := db.QueryRowContext(ctx, b.String(), args...).Scan(
		&fp.maxRowID, &fp.count, &fp.tokens, &fp.costE6, &fp.requests,
	); err != nil {
		return usageViewFingerprint{}, fmt.Errorf("probe usage_events fingerprint: %w", err)
	}
	return fp, nil
}

// loadUsageViewCached returns a cached aggregate if the fingerprint is
// unchanged. The hit bool is true on cache hit; on miss the caller must
// fall back to a full materialize+aggregate pass and then call
// storeUsageViewCache to populate the cache.
func loadUsageViewCached(
	ctx context.Context,
	db *sql.DB,
	namespace usageViewCacheNamespace,
	filter usageFilter,
) (*telemetryUsageAgg, usageViewFingerprint, bool, error) {
	if namespace == "" {
		return nil, usageViewFingerprint{}, false, nil
	}
	key := usageViewCacheKey(namespace, filter)
	entry, hit := globalUsageViewCache.lookup(key)
	fp, err := probeUsageEventsFingerprint(ctx, db, filter)
	if err != nil {
		return nil, usageViewFingerprint{}, false, err
	}
	if !hit || entry == nil {
		return nil, fp, false, nil
	}
	if !entry.fp.equal(fp) {
		core.Tracef("[usage_view_cache] miss (fingerprint changed) providers=%v account=%s maxRowID=%d→%d count=%d→%d tokens=%d→%d costE6=%d→%d requests=%d→%d",
			filter.ProviderIDs, filter.AccountID,
			entry.fp.maxRowID, fp.maxRowID, entry.fp.count, fp.count,
			entry.fp.tokens, fp.tokens, entry.fp.costE6, fp.costE6,
			entry.fp.requests, fp.requests)
		return nil, fp, false, nil
	}
	core.Tracef("[usage_view_cache] hit providers=%v account=%s maxRowID=%d count=%d age=%s",
		filter.ProviderIDs, filter.AccountID, fp.maxRowID, fp.count, time.Since(entry.insertedAt))
	return entry.agg, fp, true, nil
}

func storeUsageViewCache(
	namespace usageViewCacheNamespace,
	filter usageFilter,
	agg *telemetryUsageAgg,
	fp usageViewFingerprint,
) {
	if namespace == "" || agg == nil {
		return
	}
	key := usageViewCacheKey(namespace, filter)
	globalUsageViewCache.store(key, &usageViewCacheEntry{
		agg:        agg,
		fp:         fp,
		insertedAt: time.Now(),
	})
}
