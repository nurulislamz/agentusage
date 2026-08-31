package daemon

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nurulislamz/agentusage/internal/core"
	"github.com/nurulislamz/agentusage/internal/telemetry"
)

func TestSnapshotResetPassed(t *testing.T) {
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	since := now.Add(-time.Hour)

	// 1. Zero since or empty resets
	if snapshotResetPassed(core.UsageSnapshot{}, time.Time{}, now) {
		t.Error("expected false for zero since")
	}
	if snapshotResetPassed(core.UsageSnapshot{}, since, now) {
		t.Error("expected false for empty resets")
	}

	// 2. Reset passed between since and now
	snapPassed := core.UsageSnapshot{
		Resets: map[string]time.Time{
			"cycle": now.Add(-30 * time.Minute),
		},
	}
	if !snapshotResetPassed(snapPassed, since, now) {
		t.Error("expected true when reset passed between since and now")
	}

	// 3. Reset in future
	snapFuture := core.UsageSnapshot{
		Resets: map[string]time.Time{
			"cycle": now.Add(30 * time.Minute),
		},
	}
	if snapshotResetPassed(snapFuture, since, now) {
		t.Error("expected false when reset is in future")
	}

	// 4. Reset before since
	snapPast := core.UsageSnapshot{
		Resets: map[string]time.Time{
			"cycle": since.Add(-10 * time.Minute),
		},
	}
	if snapshotResetPassed(snapPast, since, now) {
		t.Error("expected false when reset was before since")
	}
}

func TestIsAntigravityStatusFile_EdgeCases(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"antigravity-status.json", true},
		{"antigravity_work_status.json", true},
		{"/path/to/antigravity.status.json", true},
		{".antigravity-status.tmp", false},
		{"antigravity.txt", false},
		{"claude-status.json", false},
		{"", false},
	}

	for _, tc := range cases {
		if got := isAntigravityStatusFile(tc.path); got != tc.want {
			t.Errorf("isAntigravityStatusFile(%q) = %v, want %v", tc.path, got, tc.want)
		}
	}
}

func TestCollectWatchDirs(t *testing.T) {
	dirs := collectWatchDirs()
	// Should return slice of existing directories without panicking
	for _, d := range dirs {
		info, err := os.Stat(d)
		if err != nil || !info.IsDir() {
			t.Errorf("collectWatchDirs returned invalid dir: %s", d)
		}
	}
}

func TestServerLogging_Infof_Warnf_ShouldLog(t *testing.T) {
	// 1. Nil service
	var nilSvc *Service
	nilSvc.infof("event", "format %s", "arg")
	nilSvc.warnf("event", "format %s", "arg")
	if nilSvc.shouldLog("key", time.Second) {
		t.Error("nilSvc.shouldLog should return false")
	}

	// 2. Verbose = false (logs suppressed)
	svcQuiet := &Service{
		cfg: Config{Verbose: false},
	}
	svcQuiet.infof("event", "")
	svcQuiet.warnf("event", "")

	// 3. Verbose = true
	svcVerbose := &Service{
		cfg:         Config{Verbose: true},
		logThrottle: core.NewLogThrottle(5, time.Minute),
	}
	svcVerbose.infof("event", "")
	svcVerbose.infof("event", "msg=%s", "test")
	svcVerbose.warnf("event", "")
	svcVerbose.warnf("event", "err=%s", "test-err")

	if !svcVerbose.shouldLog("key1", time.Millisecond) {
		t.Error("svcVerbose.shouldLog should return true on first call")
	}
}

func TestServerSpool_CleanupAndFlush_NilAndEmpty(t *testing.T) {
	var nilSvc *Service
	nilSvc.flushSpoolBacklog(context.Background(), 100)
	nilSvc.cleanupSpool()

	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "telemetry.db")
	store, err := telemetry.OpenStore(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	spool := telemetry.NewSpool(filepath.Join(tmp, "spool"))
	pipeline := telemetry.NewPipeline(store, spool)

	// Enqueue test request into spool
	reqs := []telemetry.IngestRequest{
		{
			SourceSystem:  "opencode",
			SourceChannel: "hook",
			AccountID:     "work",
			OccurredAt:    time.Now().UTC(),
			SessionID:     "sess1",
			TurnID:        "turn1",
			MessageID:     "msg1",
			EventType:     telemetry.EventTypeTurnCompleted,
		},
	}
	_, _ = pipeline.EnqueueRequests(reqs)

	svc := &Service{
		cfg: Config{
			SpoolDir: filepath.Join(tmp, "spool"),
			Verbose:  true,
		},
		pipeline:    pipeline,
		store:       store,
		logThrottle: core.NewLogThrottle(5, time.Minute),
	}
	svc.flushSpoolBacklog(context.Background(), 100)
	svc.cleanupSpool()
}

func TestNewServiceManager_And_NewClient(t *testing.T) {
	mgr, err := NewServiceManager("/tmp/test.sock")
	if err != nil {
		t.Fatalf("NewServiceManager error: %v", err)
	}
	if mgr.socketPath != "/tmp/test.sock" {
		t.Errorf("mgr.socketPath = %q, want '/tmp/test.sock'", mgr.socketPath)
	}

	client := NewClient("/tmp/test.sock")
	if client == nil || client.SocketPath != "/tmp/test.sock" {
		t.Errorf("NewClient = %+v", client)
	}
}

type mockChangeDetectorProvider struct {
	core.UsageProvider
	changed bool
	err     error
}

func (m *mockChangeDetectorProvider) Spec() core.ProviderSpec {
	return core.ProviderSpec{
		CreditMetrics: map[string]core.BalanceSemantics{
			"sessions": core.BalanceCumulative,
		},
	}
}

func (m *mockChangeDetectorProvider) HasChanged(_ core.AccountConfig, _ time.Time) (bool, error) {
	return m.changed, m.err
}

func TestSkipUnchangedProvider(t *testing.T) {
	svc := &Service{
		pollState: make(map[string]*providerPollState),
	}
	acct := core.AccountConfig{ID: "acc1", Provider: "mock"}

	// 1. Provider does not implement ChangeDetector -> returns nil
	if snap := svc.skipUnchangedProvider(&struct{ core.UsageProvider }{}, acct); snap != nil {
		t.Error("expected nil for non-ChangeDetector provider")
	}

	// 2. Provider implements ChangeDetector, but state is nil -> returns nil
	detector := &mockChangeDetectorProvider{changed: false}
	if snap := svc.skipUnchangedProvider(detector, acct); snap != nil {
		t.Error("expected nil when pollState has no entry")
	}

	// 3. State hasSnap=true and changed=false -> returns cached snapshot
	cachedSnap := core.UsageSnapshot{ProviderID: "mock", AccountID: "acc1", Status: core.StatusOK}
	svc.pollState["acc1"] = &providerPollState{
		hasSnap:     true,
		lastSnap:    cachedSnap,
		lastFetchAt: time.Now().Add(-5 * time.Minute),
	}
	if snap := svc.skipUnchangedProvider(detector, acct); snap == nil || snap.AccountID != "acc1" {
		t.Errorf("expected cached snapshot, got %+v", snap)
	}

	// 4. Changed=true -> returns nil
	detector.changed = true
	if snap := svc.skipUnchangedProvider(detector, acct); snap != nil {
		t.Error("expected nil when detector.HasChanged is true")
	}
}

func TestProcessHookSpool_And_Cleanup(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "telemetry.db")
	store, err := telemetry.OpenStore(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	svc := &Service{
		store:       store,
		logThrottle: core.NewLogThrottle(5, time.Minute),
	}

	hookDir := filepath.Join(tempDir, "hooks")
	_ = os.MkdirAll(hookDir, 0755)

	// 1. Write a valid hook file
	validPayload := []byte(`{"hook":"chat.message","timestamp":"2026-02-26T20:00:00Z","input":{"sessionID":"s1","messageID":"m1"},"output":{"usage":{"input_tokens":10,"output_tokens":5}}}`)
	rawFile := rawHookFile{
		Source:    "opencode",
		AccountID: "work",
		Payload:   validPayload,
	}
	rawBytes, _ := json.Marshal(rawFile)
	_ = os.WriteFile(filepath.Join(hookDir, "hook1.json"), rawBytes, 0644)

	// 2. Write a corrupt/invalid hook file
	_ = os.WriteFile(filepath.Join(hookDir, "corrupt.json"), []byte("invalid json!"), 0644)

	// 3. Write a .tmp file
	_ = os.WriteFile(filepath.Join(hookDir, "test.json.tmp"), []byte("tmp"), 0644)

	// Run processHookSpool
	svc.processHookSpool(context.Background(), hookDir)

	// Verify hook1 was ingested and removed
	if _, err := os.Stat(filepath.Join(hookDir, "hook1.json")); !os.IsNotExist(err) {
		t.Error("expected hook1.json to be processed and removed")
	}

	// 4. Create an old json file (> 24h old)
	oldPath := filepath.Join(hookDir, "old.json")
	_ = os.WriteFile(oldPath, []byte("{}"), 0644)
	oldTime := time.Now().Add(-48 * time.Hour)
	_ = os.Chtimes(oldPath, oldTime, oldTime)

	// Run cleanupHookSpool
	svc.cleanupHookSpool(hookDir)
	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Error("expected old.json to be pruned by cleanupHookSpool")
	}
	if _, err := os.Stat(filepath.Join(hookDir, "test.json.tmp")); !os.IsNotExist(err) {
		t.Error("expected test.json.tmp to be cleaned up")
	}
}

func TestPushToExporter(t *testing.T) {
	// 1. Nil exporter is safe no-op
	svc := &Service{}
	svc.pushToExporter(context.Background(), map[string]core.UsageSnapshot{"acc": {ProviderID: "test"}})

	// 2. Nil snapshots is safe no-op
	svc.pushToExporter(context.Background(), nil)
}

func TestRunCollectLoop_And_WatchLoop_Cancel(t *testing.T) {
	svc := &Service{
		cfg: Config{
			CollectInterval: 10 * time.Millisecond,
		},
		logThrottle: core.NewLogThrottle(5, time.Minute),
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // canceled immediately

	// Both loops should exit cleanly on ctx.Done()
	svc.runCollectLoop(ctx)
	svc.runWatchLoop(ctx)
}

func TestRunPollLoop_Cancel(t *testing.T) {
	svc := &Service{
		cfg: Config{
			PollInterval: 10 * time.Millisecond,
		},
		logThrottle: core.NewLogThrottle(5, time.Minute),
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	svc.runPollLoop(ctx)
}

func TestRecordBalanceObservations(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "telemetry.db")
	store, err := telemetry.OpenStore(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	if err := store.RunMigrations(context.Background()); err != nil {
		t.Fatal(err)
	}

	used := 50.0
	rem := 150.0
	snapshots := map[string]core.UsageSnapshot{
		"codex-main": {
			ProviderID: "codex",
			AccountID:  "codex-main",
			Timestamp:  time.Now().UTC(),
			Metrics: map[string]core.Metric{
				"sessions": {
					Used:      &used,
					Remaining: &rem,
					Window:    "daily",
				},
			},
		},
	}

	// Nil store is safe
	nilSvc := &Service{}
	nilSvc.recordBalanceObservations(context.Background(), snapshots)

	// Live store records observation
	svc := &Service{
		store: store,
		providerByID: map[string]core.UsageProvider{
			"codex": &mockChangeDetectorProvider{},
		},
	}
	svc.recordBalanceObservations(context.Background(), snapshots)
}

func TestPollProviders_WithMockQuotaIngest(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "telemetry.db")
	store, err := telemetry.OpenStore(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	if err := store.RunMigrations(context.Background()); err != nil {
		t.Fatal(err)
	}

	quotaIngest := telemetry.NewQuotaSnapshotIngestor(store)
	ps := newPollScheduler(30 * time.Second)

	svc := &Service{
		store:         store,
		quotaIngest:   quotaIngest,
		pollScheduler: ps,
		pollState:     make(map[string]*providerPollState),
		providerByID:  make(map[string]core.UsageProvider),
		logThrottle:   core.NewLogThrottle(5, time.Minute),
	}

	// Run poll with canceled ctx
	ctxCancel, cancel := context.WithCancel(context.Background())
	cancel()
	svc.pollProviders(ctxCancel)
}

func TestPruneOldData_LiveStore(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "telemetry.db")
	store, err := telemetry.OpenStore(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	if err := store.RunMigrations(context.Background()); err != nil {
		t.Fatal(err)
	}

	svc := &Service{
		store:       store,
		logThrottle: core.NewLogThrottle(5, time.Minute),
	}

	// Should safely run and prune without error
	complete := svc.pruneOldData(context.Background())
	if !complete {
		t.Error("pruneOldData should return complete=true on empty store")
	}
}

func TestFlushBacklog_WithRetries(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "telemetry.db")
	store, err := telemetry.OpenStore(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	if err := store.RunMigrations(context.Background()); err != nil {
		t.Fatal(err)
	}

	spool := telemetry.NewSpool(filepath.Join(tempDir, "spool"))
	pipeline := telemetry.NewPipeline(store, spool)

	svc := &Service{
		pipeline: pipeline,
	}

	retries := []telemetry.IngestRequest{
		{
			SourceSystem:  "opencode",
			SourceChannel: "hook",
			AccountID:     "work",
			OccurredAt:    time.Now().UTC(),
			SessionID:     "sess1",
			TurnID:        "turn1",
			MessageID:     "msg1",
			EventType:     telemetry.EventTypeTurnCompleted,
		},
	}

	flush, enqueued, warnings := svc.flushBacklog(context.Background(), retries, 100)
	if enqueued != 1 {
		t.Errorf("enqueued = %d, want 1", enqueued)
	}
	if flush.Ingested != 1 {
		t.Errorf("flush.Ingested = %d, want 1", flush.Ingested)
	}
	_ = warnings
}

func TestRunWatchLoop_LiveEvents(t *testing.T) {
	stateDir, err := telemetry.DefaultStateDir()
	if err != nil || stateDir == "" {
		t.Skip("no default state dir")
	}
	_ = os.MkdirAll(stateDir, 0755)

	svc := &Service{
		pollKick:    make(chan struct{}, 1),
		logThrottle: core.NewLogThrottle(5, time.Minute),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	go svc.runWatchLoop(ctx)

	time.Sleep(30 * time.Millisecond)
	// Write antigravity status file
	testStatus := filepath.Join(stateDir, "antigravity_status.json")
	_ = os.WriteFile(testStatus, []byte("{}"), 0644)
	defer os.Remove(testStatus)

	// Write generic file
	testOther := filepath.Join(stateDir, "other.json")
	_ = os.WriteFile(testOther, []byte("{}"), 0644)
	defer os.Remove(testOther)

	time.Sleep(50 * time.Millisecond)
}
