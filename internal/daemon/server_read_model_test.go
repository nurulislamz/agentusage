package daemon

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func TestComputeReadModel_TelemetryFailureDoesNotPoisonCache(t *testing.T) {
	badDB := filepath.Join(t.TempDir(), "not-a-database.db")
	if err := os.WriteFile(badDB, []byte("this is not a sqlite database"), 0o600); err != nil {
		t.Fatalf("write corrupt db stub: %v", err)
	}

	acct := core.AccountConfig{ID: "acct-1", Provider: "fake-prov"}
	req := ReadModelRequest{
		Accounts:   []ReadModelAccount{{AccountID: acct.ID, ProviderID: acct.Provider}},
		TimeWindow: core.TimeWindow30d,
	}
	cacheKey := ReadModelRequestKey(req)

	good := map[string]core.UsageSnapshot{
		acct.ID: {
			ProviderID: acct.Provider,
			AccountID:  acct.ID,
			Status:     core.StatusOK,
			Metrics: map[string]core.Metric{
				"tokens": {Used: floatPtr(42), Limit: floatPtr(100), Unit: "tokens"},
			},
			DailySeries: map[string][]core.TimePoint{
				"tokens": {{Date: "2026-09-28", Value: 42}},
			},
		},
	}

	origLoad := loadAccountsAndNormFunc
	origBuild := buildReadModelRequestFromConfigFunc
	origDisabled := disabledAccountsFromConfigFunc
	defer func() {
		loadAccountsAndNormFunc = origLoad
		buildReadModelRequestFromConfigFunc = origBuild
		disabledAccountsFromConfigFunc = origDisabled
	}()
	loadAccountsAndNormFunc = func() ([]core.AccountConfig, core.ModelNormalizationConfig, error) {
		return []core.AccountConfig{acct}, core.DefaultModelNormalizationConfig(), nil
	}
	buildReadModelRequestFromConfigFunc = func() (ReadModelRequest, error) {
		return req, nil
	}
	disabledAccountsFromConfigFunc = func() map[string]bool { return map[string]bool{} }

	svc := &Service{
		cfg:          Config{DBPath: badDB},
		rmCache:      newReadModelCache(),
		pollState:    make(map[string]*providerPollState),
		logThrottle:  core.NewLogThrottle(5, time.Minute),
		providerByID: map[string]core.UsageProvider{},
	}
	svc.pollState[acct.ID] = &providerPollState{
		hasSnap: true,
		lastSnap: core.UsageSnapshot{
			ProviderID: acct.Provider,
			AccountID:  acct.ID,
			Status:     core.StatusOK,
			Timestamp:  time.Now().UTC(),
			Metrics: map[string]core.Metric{
				"tokens": {Used: floatPtr(7), Limit: floatPtr(100), Unit: "tokens"},
			},
		},
	}
	svc.rmCache.set(cacheKey, good)

	// computeReadModel must surface the telemetry failure even when poll
	// enrichment still yields usable gauges.
	got, err := svc.computeReadModel(context.Background(), req)
	if err == nil {
		t.Fatal("computeReadModel: expected telemetry error for corrupt db, got nil")
	}
	if !SnapshotsHaveUsableData(got) {
		t.Fatalf("computeReadModel degraded result should still be usable via poll overlay: %+v", got)
	}
	if len(got[acct.ID].DailySeries) != 0 {
		t.Fatalf("degraded result unexpectedly kept DailySeries: %+v", got[acct.ID].DailySeries)
	}

	// publishReadModelSync must not replace the history-bearing cache entry.
	svc.publishReadModelSync(context.Background())
	cached, _, ok := svc.rmCache.get(cacheKey)
	if !ok {
		t.Fatal("cache entry missing after publishReadModelSync")
	}
	if len(cached[acct.ID].DailySeries["tokens"]) == 0 {
		t.Fatalf("publishReadModelSync poisoned cache; DailySeries lost: %+v", cached[acct.ID])
	}
	if cached[acct.ID].Metrics["tokens"].Used == nil || *cached[acct.ID].Metrics["tokens"].Used != 42 {
		t.Fatalf("publishReadModelSync overwrote good metrics: got %+v", cached[acct.ID].Metrics["tokens"])
	}

	// Forced refresh must serve last-good history instead of caching the
	// history-stripped poll overlay.
	body, _ := json.Marshal(ReadModelRequest{
		Accounts:   req.Accounts,
		TimeWindow: req.TimeWindow,
		Refresh:    true,
	})
	httpReq := httptest.NewRequest(http.MethodPost, "/v1/read-model", bytes.NewReader(body))
	w := httptest.NewRecorder()
	svc.handleReadModel(w, httpReq)
	if w.Code != http.StatusOK {
		t.Fatalf("handleReadModel status=%d body=%s", w.Code, w.Body.String())
	}
	var resp ReadModelResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Snapshots[acct.ID].DailySeries["tokens"]) == 0 {
		t.Fatalf("refresh response lost DailySeries; got %+v", resp.Snapshots[acct.ID])
	}
	cachedAfter, _, ok := svc.rmCache.get(cacheKey)
	if !ok || len(cachedAfter[acct.ID].DailySeries["tokens"]) == 0 {
		t.Fatalf("handleReadModel refresh poisoned cache: %+v", cachedAfter)
	}
}

func floatPtr(v float64) *float64 { return &v }
