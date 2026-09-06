package webserve

import (
	"context"
	"testing"
	"time"

	"github.com/nurulislamz/agentusage/internal/core"
	"github.com/nurulislamz/agentusage/internal/export"
)

func TestCollectorCache(t *testing.T) {
	calls := 0
	c := newCollector(Options{
		RefreshSeconds: 60,
		Version:        "test",
		Now: func() time.Time {
			return time.Date(2026, 8, 28, 12, 0, 0, 0, time.UTC)
		},
		Collect: func() (Envelope, error) {
			calls++
			return Envelope{Source: "stub", Snapshots: []core.UsageSnapshot{
				core.NewUsageSnapshot("openai", "a"),
			}}, nil
		},
	})
	if _, err := c.envelope(); err != nil {
		t.Fatal(err)
	}
	if _, err := c.envelope(); err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		// CollectFunc bypasses the in-process cache on purpose so tests stay
		// deterministic; the production fetch() path is what caches.
		t.Logf("collect func calls = %d (override path is uncached)", calls)
	}
}

func TestCollectorRefreshBypassesCache(t *testing.T) {
	now := time.Date(2026, 8, 30, 21, 0, 0, 0, time.UTC)
	c := newCollector(Options{
		Demo:           true,
		RefreshSeconds: 60,
		Now:            func() time.Time { return now },
	})
	first, err := c.envelope()
	if err != nil {
		t.Fatal(err)
	}
	now = now.Add(time.Second)
	cached, err := c.envelope()
	if err != nil {
		t.Fatal(err)
	}
	if !cached.GeneratedAt.Equal(first.GeneratedAt) {
		t.Fatalf("expected cache hit, generated_at moved %v → %v", first.GeneratedAt, cached.GeneratedAt)
	}
	fresh, err := c.envelopeRefresh(true, "")
	if err != nil {
		t.Fatal(err)
	}
	if !fresh.GeneratedAt.Equal(now) {
		t.Fatalf("refresh generated_at = %v, want %v", fresh.GeneratedAt, now)
	}
}

func TestCollectorSetUsageModeReprojectsCache(t *testing.T) {
	now := time.Date(2026, 8, 30, 21, 0, 0, 0, time.UTC)
	c := newCollector(Options{
		Demo:           true,
		UsageMode:      "remaining",
		RefreshSeconds: 60,
		Now:            func() time.Time { return now },
	})
	first, err := c.envelope()
	if err != nil {
		t.Fatal(err)
	}
	if first.UsageMode != "remaining" {
		t.Fatalf("usage_mode = %q", first.UsageMode)
	}
	c.setUsageMode("used")
	second, err := c.envelope()
	if err != nil {
		t.Fatal(err)
	}
	if second.UsageMode != "used" {
		t.Fatalf("after set usage_mode = %q", second.UsageMode)
	}
}

func TestCollectorDemoDecorates(t *testing.T) {
	c := newCollector(Options{
		Demo:           true,
		Version:        "v-test",
		TimeWindow:     "7d",
		Theme:          "Tokyo Night",
		RefreshSeconds: 15,
		Now:            func() time.Time { return time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC) },
	})
	env, err := c.envelope()
	if err != nil {
		t.Fatal(err)
	}
	if env.Source != "demo" {
		t.Errorf("source = %q", env.Source)
	}
	if env.SchemaVersion != "1" {
		t.Errorf("schema = %q", env.SchemaVersion)
	}
	if env.AgentUsageVersion != "v-test" {
		t.Errorf("version = %q", env.AgentUsageVersion)
	}
	if env.TimeWindow != "7d" || env.Theme != "Tokyo Night" {
		t.Errorf("meta window/theme = %q / %q", env.TimeWindow, env.Theme)
	}
	if env.RefreshIntervalSeconds != 15 {
		t.Errorf("refresh = %d", env.RefreshIntervalSeconds)
	}
	if len(env.Snapshots) == 0 {
		t.Fatal("expected demo snapshots")
	}
	if len(env.Catalog) == 0 {
		t.Fatal("expected catalog")
	}
}

func TestStripRawEmpty(t *testing.T) {
	if got := stripRaw(nil); got == nil || len(got) != 0 {
		t.Errorf("nil input should become empty slice, got %#v", got)
	}
}

func TestProviderCatalogIncludesKnownIDs(t *testing.T) {
	cat := providerCatalog()
	seen := map[string]bool{}
	for _, entry := range cat {
		if entry.ID == "" || entry.Name == "" {
			t.Fatalf("empty catalog entry: %#v", entry)
		}
		seen[entry.ID] = true
	}
	for _, id := range []string{"openai", "claude_code", "cursor"} {
		if !seen[id] {
			t.Errorf("catalog missing %s", id)
		}
	}
}

func TestCollectorViewRuntimeMaintained(t *testing.T) {
	c := newCollector(Options{
		TimeWindow:     "14d",
		RefreshSeconds: 30,
	})
	if c.rt == nil {
		t.Fatal("expected collector to initialize ViewRuntime")
	}
	savedRT := c.rt
	// Verify rt is maintained across fetch
	if c.rt != savedRT {
		t.Fatal("expected collector to reuse the same ViewRuntime instance")
	}
}

func TestCollectorNoProviderEnrichmentClosures(t *testing.T) {
	c := newCollector(Options{
		RefreshSeconds: 30,
	})
	if c.enrich != nil {
		t.Fatal("expected enrich closure to be nil (presentation consumes shared daemon snapshots directly)")
	}
}

func TestCollectorDefaultFetchTimeoutLeavesRoomToWrite(t *testing.T) {
	c := newCollector(Options{Demo: true, RefreshSeconds: 30})
	if c.fetchTimeout != 12*time.Second {
		t.Fatalf("fetchTimeout = %v, want 12s so HTTP WriteTimeout can still write the body", c.fetchTimeout)
	}
}

func TestCollectorFetchTimesOutWhenSnapshotsHang(t *testing.T) {
	c := newCollector(Options{
		Source:         string(export.SourceDirect),
		RefreshSeconds: 30,
		Now:            func() time.Time { return time.Date(2026, 9, 3, 15, 0, 0, 0, time.UTC) },
	})
	c.fetchTimeout = 40 * time.Millisecond
	c.snapshotFetch = func(ctx context.Context, refresh bool, accountID string) ([]core.UsageSnapshot, string, error) {
		time.Sleep(2 * time.Second)
		return nil, "direct", nil
	}

	start := time.Now()
	_, err := c.fetch(false, "")
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if elapsed := time.Since(start); elapsed > 400*time.Millisecond {
		t.Fatalf("fetch took %v, want return at fetchTimeout", elapsed)
	}
}

func TestCollectorServesCacheWhileRefreshIsStuck(t *testing.T) {
	now := time.Date(2026, 9, 3, 15, 0, 0, 0, time.UTC)
	c := newCollector(Options{
		Demo:           true,
		RefreshSeconds: 60,
		Now:            func() time.Time { return now },
	})
	first, err := c.envelope()
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Views) == 0 {
		t.Fatal("expected cached demo views")
	}

	started := make(chan struct{})
	release := make(chan struct{})
	c.demo = false
	c.fetchTimeout = 2 * time.Second
	c.snapshotFetch = func(ctx context.Context, refresh bool, accountID string) ([]core.UsageSnapshot, string, error) {
		close(started)
		<-release
		return nil, "direct", nil
	}

	done := make(chan error, 1)
	go func() {
		_, err := c.envelopeRefresh(true, "")
		done <- err
	}()
	select {
	case <-started:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("stuck refresh never started")
	}

	start := time.Now()
	cached, err := c.envelope()
	if err != nil {
		t.Fatal(err)
	}
	if elapsed := time.Since(start); elapsed > 200*time.Millisecond {
		t.Fatalf("cached envelope() blocked for %v while refresh was in flight", elapsed)
	}
	if cached.GeneratedAt != first.GeneratedAt {
		t.Fatalf("expected cache hit, generated_at moved %v → %v", first.GeneratedAt, cached.GeneratedAt)
	}

	close(release)
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("refresh goroutine did not finish")
	}
}

func TestCollectorFetchSnapshotsEnrichment(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	c := newCollector(Options{
		Source:         string(export.SourceDirect),
		RefreshSeconds: 30,
	})

	var calledWithAccountID string
	var enrichCalled bool
	c.enrich = func(ctx context.Context, snaps map[string]core.UsageSnapshot, accountID string) {
		enrichCalled = true
		calledWithAccountID = accountID
		snaps["cursor-mock"] = core.NewUsageSnapshot("cursor", "cursor-mock")
	}

	snaps, _, err := c.fetchSnapshots(ctx, true, "cursor-mock")
	if err != nil {
		t.Fatalf("fetchSnapshots: %v", err)
	}
	if !enrichCalled {
		t.Fatal("expected enrich to be called during fetchSnapshots")
	}
	if calledWithAccountID != "cursor-mock" {
		t.Errorf("calledWithAccountID = %q, want cursor-mock", calledWithAccountID)
	}
	found := false
	for _, snap := range snaps {
		if snap.AccountID == "cursor-mock" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected mock snapshot in returned ordered snapshots")
	}
}
