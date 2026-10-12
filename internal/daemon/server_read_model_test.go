package daemon

import (
	"context"
	"testing"
	"time"

	"github.com/nurulislamz/agentusage/internal/core"
)

func TestReadModelCacheIntervalRespectsPollInterval(t *testing.T) {
	tests := []struct {
		name string
		in   time.Duration
		want time.Duration
	}{
		{name: "default", in: 0, want: 30 * time.Second},
		{name: "minimum", in: time.Second, want: 5 * time.Second},
		{name: "normal", in: 30 * time.Second, want: 30 * time.Second},
		{name: "long", in: time.Hour, want: time.Hour},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := readModelCacheInterval(tt.in); got != tt.want {
				t.Fatalf("readModelCacheInterval(%s) = %s, want %s", tt.in, got, tt.want)
			}
		})
	}
}

func TestServiceContext(t *testing.T) {
	// 1. Nil service with nil fallback
	var nilSvc *Service
	if ctx := nilSvc.serviceContext(nil); ctx == nil {
		t.Error("serviceContext(nil) should return non-nil context")
	}

	// 2. Nil service with custom fallback
	customCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if ctx := nilSvc.serviceContext(customCtx); ctx != customCtx {
		t.Error("serviceContext(custom) should return fallback context")
	}

	// 3. Service with set ctx
	svcCtx, cancel2 := context.WithCancel(context.Background())
	defer cancel2()
	svc := &Service{ctx: svcCtx}
	if ctx := svc.serviceContext(customCtx); ctx != svcCtx {
		t.Error("serviceContext should prefer svc.ctx over fallback")
	}
}

func TestComputeReadModel_Empty(t *testing.T) {
	svc := &Service{}
	res, err := svc.computeReadModel(context.Background(), ReadModelRequest{
		Accounts: nil,
	})
	if err != nil {
		t.Fatalf("computeReadModel on empty accounts err: %v", err)
	}
	if len(res) != 0 {
		t.Errorf("computeReadModel = %+v, want empty map", res)
	}
}

func TestRefreshReadModelCacheFromConfig(t *testing.T) {
	svc := &Service{
		rmCache:     newReadModelCache(),
		logThrottle: core.NewLogThrottle(5, time.Minute),
	}

	// Should safely run and trigger async refresh without error
	svc.refreshReadModelCacheFromConfig(context.Background())
}

func TestReadModelCache_SetIfGenerationRejectsStaleAsync(t *testing.T) {
	cache := newReadModelCache()
	stale := map[string]core.UsageSnapshot{
		"acc1": {
			ProviderID: "cursor",
			AccountID:  "acc1",
			Status:     core.StatusOK,
			DailySeries: map[string][]core.TimePoint{
				"cost": {{Date: "2026-10-01", Value: 1}},
			},
		},
	}
	fresh := map[string]core.UsageSnapshot{
		"acc1": {
			ProviderID: "cursor",
			AccountID:  "acc1",
			Status:     core.StatusOK,
			DailySeries: map[string][]core.TimePoint{
				"cost": {{Date: "2026-10-11", Value: 9}},
			},
			ModelUsage: []core.ModelUsageRecord{{RawModelID: "gpt"}},
		},
	}

	cache.set("key", stale)
	gen, ok := cache.beginRefresh("key")
	if !ok {
		t.Fatal("beginRefresh failed")
	}
	if gen != 1 {
		t.Fatalf("generation = %d, want 1", gen)
	}

	// Sync publish lands while the async compute is still in flight.
	cache.set("key", fresh)

	if cache.setIfGeneration("key", stale, gen) {
		t.Fatal("setIfGeneration accepted stale async result after sync publish")
	}
	cached, _, ok := cache.get("key")
	if !ok {
		t.Fatal("fresh cache entry missing")
	}
	points := cached["acc1"].DailySeries["cost"]
	if len(points) != 1 || points[0].Date != "2026-10-11" {
		t.Fatalf("cache rolled back to stale view: %+v", points)
	}
	if len(cached["acc1"].ModelUsage) != 1 {
		t.Fatalf("ModelUsage wiped by stale async: %+v", cached["acc1"].ModelUsage)
	}
	cache.endRefresh("key")

	// Unchanged generation still accepts the async write.
	gen2, ok := cache.beginRefresh("key")
	if !ok {
		t.Fatal("beginRefresh after end failed")
	}
	newer := map[string]core.UsageSnapshot{
		"acc1": {
			ProviderID: "cursor",
			AccountID:  "acc1",
			Status:     core.StatusOK,
			DailySeries: map[string][]core.TimePoint{
				"cost": {{Date: "2026-10-12", Value: 3}},
			},
		},
	}
	if !cache.setIfGeneration("key", newer, gen2) {
		t.Fatal("setIfGeneration rejected current generation")
	}
	cached, _, ok = cache.get("key")
	if !ok || cached["acc1"].DailySeries["cost"][0].Date != "2026-10-12" {
		t.Fatalf("expected async write to land, got %+v", cached["acc1"].DailySeries)
	}
	cache.endRefresh("key")
}
