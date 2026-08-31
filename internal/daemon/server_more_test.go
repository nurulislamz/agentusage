package daemon

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/nurulislamz/agentusage/internal/core"
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
	svc := &Service{
		cfg: Config{
			SpoolDir: tmp,
			Verbose:  true,
		},
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
