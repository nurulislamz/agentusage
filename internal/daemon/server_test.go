package daemon

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/nurulislamz/agentusage/internal/core"
	"github.com/nurulislamz/agentusage/internal/providers"
)

func shortSocketPath(t *testing.T, suffix string) string {
	t.Helper()
	// Use the OS temp dir (short enough to stay under the AF_UNIX sun_path
	// limit) rather than a hardcoded /tmp, which does not exist on Windows.
	return filepath.Join(os.TempDir(), fmt.Sprintf("agentusage-%d-%s.sock", time.Now().UnixNano(), strings.TrimSpace(suffix)))
}

func TestEnsureSocketPathAvailable_ActiveSocketReturnsError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix sockets are not supported in this test")
	}

	socketPath := shortSocketPath(t, "active")
	_ = os.Remove(socketPath)
	t.Cleanup(func() { _ = os.Remove(socketPath) })
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("listen unix socket: %v", err)
	}
	defer listener.Close()

	err = EnsureSocketPathAvailable(socketPath)
	if err == nil {
		t.Fatal("expected error for active daemon socket")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "already running") {
		t.Fatalf("error = %q, want already running message", err)
	}
}

func TestEnsureSocketPathAvailable_RemovesStaleSocket(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix sockets are not supported in this test")
	}

	socketPath := shortSocketPath(t, "stale")
	_ = os.Remove(socketPath)
	t.Cleanup(func() { _ = os.Remove(socketPath) })
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("listen unix socket: %v", err)
	}
	if err := listener.Close(); err != nil {
		t.Fatalf("close listener: %v", err)
	}

	if _, statErr := os.Stat(socketPath); statErr != nil && !os.IsNotExist(statErr) {
		t.Fatalf("stat socket before ensure: %v", statErr)
	}

	if err := EnsureSocketPathAvailable(socketPath); err != nil {
		t.Fatalf("ensure socket path available: %v", err)
	}

	if _, statErr := os.Stat(socketPath); !os.IsNotExist(statErr) {
		t.Fatalf("expected stale socket to be removed, stat err = %v", statErr)
	}
}

func TestEnsureSocketPathAvailable_RejectsRegularFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		// On Windows an AF_UNIX socket path is materialized as a regular file
		// and never reports os.ModeSocket, so we deliberately cannot distinguish
		// a leftover socket from any other file: EnsureSocketPathAvailable treats
		// it as a possibly-stale socket and removes it after a failed dial probe
		// (see socket_windows.go). The "reject regular file" semantic is
		// macOS/Linux-only.
		t.Skip("regular-file rejection is not applicable to Windows AF_UNIX sockets")
	}
	socketPath := shortSocketPath(t, "file")
	_ = os.Remove(socketPath)
	t.Cleanup(func() { _ = os.Remove(socketPath) })
	if err := os.WriteFile(socketPath, []byte("not-a-socket"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	err := EnsureSocketPathAvailable(socketPath)
	if err == nil {
		t.Fatal("expected error for regular file at socket path")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "not a socket") {
		t.Fatalf("error = %q, want not a socket message", err)
	}
}

func TestDefaultCollectOptions_GeminiHasSessionsDir(t *testing.T) {
	source, ok := providers.TelemetrySourceBySystem("gemini_cli")
	if !ok {
		t.Skip("gemini_cli telemetry source not found in registry")
	}
	opts := source.DefaultCollectOptions()

	if got := opts.Paths["sessions_dir"]; got == "" {
		t.Fatal("expected non-empty sessions_dir from gemini DefaultCollectOptions")
	}
	if _, ok := opts.Paths["projects_dir"]; ok {
		t.Fatalf("unexpected claude projects_dir in gemini opts: %+v", opts.Paths)
	}
}

func TestStartSocketServer_Permissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix sockets are not supported in this test")
	}

	socketPath := shortSocketPath(t, "perms")
	_ = os.Remove(socketPath)
	t.Cleanup(func() { _ = os.Remove(socketPath) })

	svc := &Service{
		cfg: Config{
			SocketPath: socketPath,
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := svc.startSocketServer(ctx); err != nil {
		t.Fatalf("startSocketServer: %v", err)
	}

	info, err := os.Stat(socketPath)
	if err != nil {
		t.Fatalf("stat socket: %v", err)
	}

	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("socket permissions = %o, want 0600", perm)
	}
}

func TestHandleHook_PayloadLimit(t *testing.T) {
	svc := &Service{}
	oversized := bytes.Repeat([]byte("a"), 5<<20) // 5 MiB (limit is 4 MiB)
	req := httptest.NewRequest(http.MethodPost, "/v1/hook/opencode", bytes.NewReader(oversized))
	w := httptest.NewRecorder()

	svc.handleHook(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d (RequestEntityTooLarge)", w.Code, http.StatusRequestEntityTooLarge)
	}
}

func TestHandleHealth(t *testing.T) {
	svc := &Service{}
	req := httptest.NewRequest(http.MethodGet, "/v1/health", nil)
	w := httptest.NewRecorder()

	svc.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if !strings.Contains(w.Body.String(), `"status":"ok"`) {
		t.Errorf("health body = %s, want status:ok", w.Body.String())
	}
}

func TestHandlePoll_Methods(t *testing.T) {
	svc := &Service{
		pollKick: make(chan struct{}, 1),
	}

	// 1. GET returns 405
	reqGet := httptest.NewRequest(http.MethodGet, "/v1/poll", nil)
	wGet := httptest.NewRecorder()
	svc.handlePoll(wGet, reqGet)
	if wGet.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET /v1/poll status = %d, want 405", wGet.Code)
	}

	// 2. POST without wait returns 200 kicked
	reqPost := httptest.NewRequest(http.MethodPost, "/v1/poll", nil)
	wPost := httptest.NewRecorder()
	svc.handlePoll(wPost, reqPost)
	if wPost.Code != http.StatusOK {
		t.Errorf("POST /v1/poll status = %d, want 200", wPost.Code)
	}
	if !strings.Contains(wPost.Body.String(), `"status":"kicked"`) {
		t.Errorf("POST /v1/poll body = %s, want status:kicked", wPost.Body.String())
	}

	// 3. POST with wait=1 returns 200 polled
	reqWait := httptest.NewRequest(http.MethodPost, "/v1/poll?wait=1", nil)
	wWait := httptest.NewRecorder()
	svc.handlePoll(wWait, reqWait)
	if wWait.Code != http.StatusOK {
		t.Errorf("POST /v1/poll?wait=1 status = %d, want 200", wWait.Code)
	}
	if !strings.Contains(wWait.Body.String(), `"status":"polled"`) {
		t.Errorf("POST /v1/poll?wait=1 body = %s, want status:polled", wWait.Body.String())
	}
}

func TestHandleHook_Validation(t *testing.T) {
	svc := &Service{
		pollKick: make(chan struct{}, 1),
	}

	// 1. GET returns 405
	reqGet := httptest.NewRequest(http.MethodGet, "/v1/hook/opencode", nil)
	wGet := httptest.NewRecorder()
	svc.handleHook(wGet, reqGet)
	if wGet.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET /v1/hook status = %d, want 405", wGet.Code)
	}

	// 2. Missing source -> 400
	reqNoSource := httptest.NewRequest(http.MethodPost, "/v1/hook/", bytes.NewReader([]byte("{}")))
	wNoSource := httptest.NewRecorder()
	svc.handleHook(wNoSource, reqNoSource)
	if wNoSource.Code != http.StatusBadRequest {
		t.Errorf("POST /v1/hook/ (no source) status = %d, want 400", wNoSource.Code)
	}

	// 3. Empty payload -> 400
	reqEmpty := httptest.NewRequest(http.MethodPost, "/v1/hook/opencode", bytes.NewReader([]byte("   ")))
	wEmpty := httptest.NewRecorder()
	svc.handleHook(wEmpty, reqEmpty)
	if wEmpty.Code != http.StatusBadRequest {
		t.Errorf("POST /v1/hook/opencode (empty) status = %d, want 400", wEmpty.Code)
	}

	// 4. Invalid JSON payload -> 400
	reqInvalid := httptest.NewRequest(http.MethodPost, "/v1/hook/opencode", bytes.NewReader([]byte("not json!")))
	wInvalid := httptest.NewRecorder()
	svc.handleHook(wInvalid, reqInvalid)
	if wInvalid.Code != http.StatusBadRequest {
		t.Errorf("POST /v1/hook/opencode (invalid) status = %d, want 400", wInvalid.Code)
	}

	// 5. Valid hook payload returns 200
	validPayload := []byte(`{"hook":"chat.message","timestamp":"2026-02-26T20:00:00Z","input":{"sessionID":"sess-1","agent":"main","messageID":"turn-1","variant":"default","model":{"providerID":"openrouter","modelID":"openai/gpt-oss-20b"}},"output":{"message":{"id":"msg-1","sessionID":"sess-1","role":"assistant"},"route":{"provider_name":"DeepInfra"},"usage":{"input_tokens":12,"output_tokens":4,"total_tokens":16,"cost_usd":0.00012}}}`)
	reqValid := httptest.NewRequest(http.MethodPost, "/v1/hook/opencode", bytes.NewReader(validPayload))
	wValid := httptest.NewRecorder()
	svc.handleHook(wValid, reqValid)
	if wValid.Code != http.StatusOK {
		t.Errorf("POST /v1/hook/opencode status = %d, want 200", wValid.Code)
	}
}

func TestHandleReadModel_Validation(t *testing.T) {
	cache := newReadModelCache()
	reqAccounts := []ReadModelAccount{{AccountID: "acc1", ProviderID: "prov1"}}
	rmReq := ReadModelRequest{Accounts: reqAccounts}
	key := ReadModelRequestKey(rmReq)
	cache.set(key, map[string]core.UsageSnapshot{
		"acc1": {AccountID: "acc1", ProviderID: "prov1", Status: core.StatusOK},
	})

	svc := &Service{
		rmCache: cache,
	}

	// 1. GET returns 405
	reqGet := httptest.NewRequest(http.MethodGet, "/v1/read-model", nil)
	wGet := httptest.NewRecorder()
	svc.handleReadModel(wGet, reqGet)
	if wGet.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET /v1/read-model status = %d, want 405", wGet.Code)
	}

	// 2. Invalid JSON body -> 400
	reqInvalid := httptest.NewRequest(http.MethodPost, "/v1/read-model", bytes.NewReader([]byte("invalid json")))
	wInvalid := httptest.NewRecorder()
	svc.handleReadModel(wInvalid, reqInvalid)
	if wInvalid.Code != http.StatusBadRequest {
		t.Errorf("POST /v1/read-model (invalid) status = %d, want 400", wInvalid.Code)
	}

	// 3. Cache Hit -> 200 with cached snapshots
	reqBody, _ := json.Marshal(rmReq)
	reqCache := httptest.NewRequest(http.MethodPost, "/v1/read-model", bytes.NewReader(reqBody))
	wCache := httptest.NewRecorder()
	svc.handleReadModel(wCache, reqCache)
	if wCache.Code != http.StatusOK {
		t.Errorf("POST /v1/read-model cache hit status = %d, want 200", wCache.Code)
	}

	// 4. Empty accounts request -> reads from config
	reqEmpty := httptest.NewRequest(http.MethodPost, "/v1/read-model", bytes.NewReader([]byte(`{"accounts":[]}`)))
	wEmpty := httptest.NewRecorder()
	svc.handleReadModel(wEmpty, reqEmpty)
	if wEmpty.Code != http.StatusOK {
		t.Errorf("POST /v1/read-model empty accounts status = %d, want 200", wEmpty.Code)
	}

	// 5. Refresh=true bypasses cache
	rmReqRefresh := ReadModelRequest{Accounts: reqAccounts, Refresh: true}
	reqRefreshBody, _ := json.Marshal(rmReqRefresh)
	reqRefresh := httptest.NewRequest(http.MethodPost, "/v1/read-model", bytes.NewReader(reqRefreshBody))
	wRefresh := httptest.NewRecorder()
	svc.handleReadModel(wRefresh, reqRefresh)
	if wRefresh.Code != http.StatusOK {
		t.Errorf("POST /v1/read-model refresh=true status = %d, want 200", wRefresh.Code)
	}
}

func TestReadModelCache_OperationsAndEviction(t *testing.T) {
	cache := newReadModelCache()

	// 1. Get on non-existent or empty key
	if _, _, ok := cache.get(""); ok {
		t.Error("get(\"\") returned true, want false")
	}
	if _, _, ok := cache.get("missing"); ok {
		t.Error("get(\"missing\") returned true, want false")
	}

	// 2. Set and Get
	snaps := map[string]core.UsageSnapshot{
		"acc1": {
			ProviderID: "test-prov",
			AccountID:  "acc1",
			Status:     core.StatusOK,
		},
	}
	cache.set("key1", snaps)
	cached, updatedAt, ok := cache.get("key1")
	if !ok {
		t.Fatal("get(\"key1\") returned false after set")
	}
	if len(cached) != 1 || cached["acc1"].ProviderID != "test-prov" {
		t.Errorf("cached snapshot = %+v", cached)
	}
	if updatedAt.IsZero() {
		t.Error("updatedAt should not be zero")
	}

	// 3. In-flight refresh locking
	if !cache.beginRefresh("key1") {
		t.Error("beginRefresh(\"key1\") should return true on first call")
	}
	if cache.beginRefresh("key1") {
		t.Error("beginRefresh(\"key1\") should return false when already in flight")
	}
	if cache.beginRefresh("") {
		t.Error("beginRefresh(\"\") should return false")
	}
	cache.endRefresh("key1")
	if !cache.beginRefresh("key1") {
		t.Error("beginRefresh(\"key1\") should return true after endRefresh")
	}
	cache.endRefresh("key1")

	// 4. Cache eviction when exceeding maxEntries (50)
	for i := 0; i < 60; i++ {
		k := fmt.Sprintf("k-%d", i)
		cache.set(k, snaps)
	}
	cache.mu.RLock()
	totalEntries := len(cache.entries)
	cache.mu.RUnlock()
	if totalEntries > 50 {
		t.Errorf("cache entries = %d, want <= 50 after eviction", totalEntries)
	}
}

type fakeCountingProvider struct {
	mu           sync.Mutex
	fetchCount   int
	barrier      chan struct{}
	fetchErr     error
	snapshotFunc func() core.UsageSnapshot
}

func (f *fakeCountingProvider) ID() string { return "fake-prov" }
func (f *fakeCountingProvider) Describe() core.ProviderInfo {
	return core.ProviderInfo{Name: "Fake Provider"}
}
func (f *fakeCountingProvider) Spec() core.ProviderSpec               { return core.ProviderSpec{} }
func (f *fakeCountingProvider) DashboardWidget() core.DashboardWidget { return core.DashboardWidget{} }
func (f *fakeCountingProvider) DetailWidget() core.DetailWidget       { return core.DetailWidget{} }

func (f *fakeCountingProvider) Fetch(ctx context.Context, acct core.AccountConfig) (core.UsageSnapshot, error) {
	f.mu.Lock()
	f.fetchCount++
	barrier := f.barrier
	f.mu.Unlock()

	if barrier != nil {
		select {
		case <-barrier:
		case <-ctx.Done():
			return core.UsageSnapshot{}, ctx.Err()
		}
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	if f.fetchErr != nil {
		return core.UsageSnapshot{}, f.fetchErr
	}
	if f.snapshotFunc != nil {
		return f.snapshotFunc(), nil
	}
	used := 50.0
	return core.UsageSnapshot{
		ProviderID: acct.Provider,
		AccountID:  acct.ID,
		Status:     core.StatusOK,
		Metrics: map[string]core.Metric{
			"requests": {Used: &used},
		},
	}, nil
}

func TestServer_SimultaneousRefreshes_CoalescedGenerationAndSharedCompletion(t *testing.T) {
	tempDir := t.TempDir()
	barrier := make(chan struct{})
	prov := &fakeCountingProvider{
		barrier: barrier,
	}
	acct := core.AccountConfig{ID: "acct-barrier", Provider: "fake-prov"}

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
		return ReadModelRequest{
			Accounts:   []ReadModelAccount{{AccountID: acct.ID, ProviderID: acct.Provider}},
			TimeWindow: core.TimeWindow30d,
		}, nil
	}
	disabledAccountsFromConfigFunc = func() map[string]bool {
		return map[string]bool{}
	}

	svc := &Service{
		cfg: Config{
			DBPath: filepath.Join(tempDir, "empty.db"),
		},
		providerByID:  map[string]core.UsageProvider{"fake-prov": prov},
		pollScheduler: newPollScheduler(30 * time.Second),
		rmCache:       newReadModelCache(),
		pollState:     make(map[string]*providerPollState),
		pollKick:      make(chan struct{}, 1),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	errCh := make(chan error, 3)

	// Refresh 1: Scheduled / background timer poll
	wg.Add(1)
	go func() {
		defer wg.Done()
		svc.pollProviders(ctx)
	}()

	// Wait briefly for pollProviders to begin generation and hit barrier
	time.Sleep(50 * time.Millisecond)

	// Refresh 2: Manual HTTP poll with wait=1
	wg.Add(1)
	go func() {
		defer wg.Done()
		req := httptest.NewRequest(http.MethodPost, "/v1/poll?wait=1", nil)
		w := httptest.NewRecorder()
		svc.handlePoll(w, req)
		if w.Code != http.StatusOK {
			errCh <- fmt.Errorf("manual poll 1 status=%d", w.Code)
		}
	}()

	// Refresh 3: Another simultaneous manual HTTP poll with wait=1
	wg.Add(1)
	go func() {
		defer wg.Done()
		req := httptest.NewRequest(http.MethodPost, "/v1/poll?wait=1", nil)
		w := httptest.NewRecorder()
		svc.handlePoll(w, req)
		if w.Code != http.StatusOK {
			errCh <- fmt.Errorf("manual poll 2 status=%d", w.Code)
		}
	}()

	// Give goroutines time to reach the barrier
	time.Sleep(50 * time.Millisecond)

	// Verify that while blocked at the barrier:
	// 1. One active generation exists
	activeGen := svc.pollScheduler.ActiveGenerationID()
	if activeGen == 0 {
		t.Error("expected active generation while blocked at barrier")
	}

	// 2. Upstream provider Fetch was invoked exactly once (coalesced!)
	prov.mu.Lock()
	fCount := prov.fetchCount
	prov.mu.Unlock()
	if fCount != 1 {
		t.Fatalf("expected 1 fetch while coalesced at barrier, got %d", fCount)
	}

	// Release the barrier
	close(barrier)
	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Fatalf("goroutine error: %v", err)
	}

	// Upstream fetch count must still be 1 (no duplicate calls after barrier released)
	prov.mu.Lock()
	fCount = prov.fetchCount
	prov.mu.Unlock()
	if fCount != 1 {
		t.Fatalf("expected exactly 1 fetch total across all simultaneous refreshes, got %d", fCount)
	}

	// Active generation must be ended (0)
	if svc.pollScheduler.ActiveGenerationID() != 0 {
		t.Errorf("active generation after completion = %d, want 0", svc.pollScheduler.ActiveGenerationID())
	}

	// Result must be published in read model cache
	req := ReadModelRequest{
		Accounts:   []ReadModelAccount{{AccountID: acct.ID, ProviderID: acct.Provider}},
		TimeWindow: core.TimeWindow30d,
	}
	cacheKey := ReadModelRequestKey(req)
	cached, _, ok := svc.rmCache.get(cacheKey)
	if !ok || len(cached) == 0 {
		t.Fatal("read model cache missing published result after generation end")
	}
	snap := cached[acct.ID]
	if snap.Status != core.StatusOK {
		t.Errorf("snapshot status = %v, want StatusOK", snap.Status)
	}
	if snap.Metrics["requests"].Used == nil || *snap.Metrics["requests"].Used != 50.0 {
		t.Errorf("snapshot metric requests = %v, want 50.0", snap.Metrics["requests"])
	}
}

func TestServer_DeterministicFreshness_FakeClock_And_LastErrorRetained(t *testing.T) {
	tempDir := t.TempDir()
	t0 := time.Date(2026, 9, 6, 10, 0, 0, 0, time.UTC)
	clk := newTestFakeClock(t0)

	prov := &fakeCountingProvider{}
	acct := core.AccountConfig{ID: "acct-clock", Provider: "fake-prov"}

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
		return ReadModelRequest{
			Accounts:   []ReadModelAccount{{AccountID: acct.ID, ProviderID: acct.Provider}},
			TimeWindow: core.TimeWindow30d,
		}, nil
	}
	disabledAccountsFromConfigFunc = func() map[string]bool {
		return map[string]bool{}
	}

	svc := (&Service{
		cfg: Config{
			DBPath: filepath.Join(tempDir, "empty.db"),
		},
		providerByID:  map[string]core.UsageProvider{"fake-prov": prov},
		pollScheduler: newPollScheduler(30 * time.Second),
		rmCache:       newReadModelCache(),
		pollState:     make(map[string]*providerPollState),
		pollKick:      make(chan struct{}, 1),
	}).WithClock(clk)

	ctx := context.Background()

	// 1. Initial successful poll at t0
	svc.pollProviders(ctx)

	req := ReadModelRequest{
		Accounts:   []ReadModelAccount{{AccountID: acct.ID, ProviderID: acct.Provider}},
		TimeWindow: core.TimeWindow30d,
	}
	cacheKey := ReadModelRequestKey(req)
	cached, _, ok := svc.rmCache.get(cacheKey)
	if !ok {
		t.Fatal("cache missing after initial poll")
	}
	snap := cached[acct.ID]
	if snap.Status != core.StatusOK {
		t.Fatalf("initial status = %v, want StatusOK", snap.Status)
	}
	// Timestamp stamped by fake clock at t0
	if !snap.Timestamp.Equal(t0) {
		t.Fatalf("snapshot timestamp = %v, want deterministic clock time %v", snap.Timestamp, t0)
	}

	// 2. Advance fake clock by 1 hour
	t1 := t0.Add(1 * time.Hour)
	clk.Advance(1 * time.Hour)

	// Simulate provider failure on upstream API
	prov.mu.Lock()
	prov.fetchErr = errors.New("upstream API 500 internal server error")
	prov.mu.Unlock()

	// Run poll during failure
	svc.pollProviders(ctx)

	cached2, _, ok2 := svc.rmCache.get(cacheKey)
	if !ok2 {
		t.Fatal("cache missing after second poll")
	}
	snap2 := cached2[acct.ID]

	// Must retain last successful values and last-success timestamp
	if snap2.Status != core.StatusError {
		t.Errorf("status on failure = %v, want StatusError", snap2.Status)
	}
	if !snap2.Timestamp.Equal(t0) {
		t.Errorf("snapshot timestamp = %v, want preserved last-success timestamp %v", snap2.Timestamp, t0)
	}
	if snap2.Metrics["requests"].Used == nil || *snap2.Metrics["requests"].Used != 50.0 {
		t.Errorf("retained metric requests = %v, want 50.0", snap2.Metrics["requests"])
	}

	// Must identify error diagnostics and error timestamp
	if snap2.Diagnostics["last_error"] != "upstream API 500 internal server error" {
		t.Errorf("diagnostics last_error = %q", snap2.Diagnostics["last_error"])
	}
	if snap2.Diagnostics["last_error_at"] != t1.Format(time.RFC3339) {
		t.Errorf("diagnostics last_error_at = %q, want %q", snap2.Diagnostics["last_error_at"], t1.Format(time.RFC3339))
	}
}

func TestServer_ScheduledAndExplicitRefresh_ShareRateLimitBackoff(t *testing.T) {
	tempDir := t.TempDir()
	t0 := time.Date(2026, 9, 6, 10, 0, 0, 0, time.UTC)
	clk := newTestFakeClock(t0)

	resetAt := t0.Add(2 * time.Hour)
	prov := &fakeCountingProvider{
		snapshotFunc: func() core.UsageSnapshot {
			return core.UsageSnapshot{
				ProviderID: "fake-prov",
				AccountID:  "acct-ratelimit",
				Status:     core.StatusLimited,
				Resets: map[string]time.Time{
					"requests": resetAt,
				},
			}
		},
	}
	acct := core.AccountConfig{ID: "acct-ratelimit", Provider: "fake-prov"}

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
		return ReadModelRequest{
			Accounts:   []ReadModelAccount{{AccountID: acct.ID, ProviderID: acct.Provider}},
			TimeWindow: core.TimeWindow30d,
		}, nil
	}
	disabledAccountsFromConfigFunc = func() map[string]bool {
		return map[string]bool{}
	}

	svc := (&Service{
		cfg: Config{
			DBPath: filepath.Join(tempDir, "empty.db"),
		},
		providerByID:  map[string]core.UsageProvider{"fake-prov": prov},
		pollScheduler: newPollScheduler(30 * time.Second),
		rmCache:       newReadModelCache(),
		pollState:     make(map[string]*providerPollState),
		pollKick:      make(chan struct{}, 1),
	}).WithClock(clk)

	ctx := context.Background()

	// 1. Initial poll returns StatusLimited with resetAt = t0 + 2h
	svc.pollProviders(ctx)
	if prov.fetchCount != 1 {
		t.Fatalf("fetchCount = %d, want 1", prov.fetchCount)
	}

	// 2. Advance clock to t0 + 30m (still within rate-limit window)
	clk.Advance(30 * time.Minute)

	// Scheduled check
	if svc.pollScheduler.ShouldPoll(acct.ID, false) {
		t.Fatal("pollScheduler.ShouldPoll returned true while rate limited")
	}

	// Explicit manual poll: handlePoll with wait=1
	req := httptest.NewRequest(http.MethodPost, "/v1/poll?wait=1", nil)
	w := httptest.NewRecorder()
	svc.handlePoll(w, req)

	// Provider Fetch should NOT be invoked again because rate-limit is shared
	if prov.fetchCount != 1 {
		t.Fatalf("fetchCount after explicit poll during rate limit = %d, want 1 (should not hammer upstream)", prov.fetchCount)
	}

	// 3. Advance clock past resetAt (t0 + 2h + 1m)
	clk.Advance(91 * time.Minute)

	if !svc.pollScheduler.ShouldPoll(acct.ID, false) {
		t.Fatal("pollScheduler.ShouldPoll returned false after rate limit expired")
	}

	// Now explicit poll can fetch
	req2 := httptest.NewRequest(http.MethodPost, "/v1/poll?wait=1", nil)
	w2 := httptest.NewRecorder()
	svc.handlePoll(w2, req2)

	if prov.fetchCount != 2 {
		t.Fatalf("fetchCount after rate limit expired = %d, want 2", prov.fetchCount)
	}
}
