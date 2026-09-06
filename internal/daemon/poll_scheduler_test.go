package daemon

import (
	"sync"
	"testing"
	"time"

	"github.com/nurulislamz/agentusage/internal/core"
)

func TestPollScheduler_ShouldPoll_FirstPollAlwaysRuns(t *testing.T) {
	ps := newPollScheduler(30 * time.Second)
	if !ps.ShouldPoll("acct1", false) {
		t.Error("first poll should always run")
	}
}

func TestPollScheduler_ShouldPoll_RespectsBaseInterval(t *testing.T) {
	ps := newPollScheduler(30 * time.Second)

	// First poll runs and records.
	ps.ShouldPoll("acct1", false)
	ps.RecordPoll("acct1", true) // changed

	// Immediately after: should not poll again.
	if ps.ShouldPoll("acct1", false) {
		t.Error("should not poll immediately after previous poll")
	}
}

func TestPollScheduler_HTTPExponentialBackoff(t *testing.T) {
	ps := newPollScheduler(30 * time.Second)

	ps.ShouldPoll("acct1", false) // init

	tests := []struct {
		noChangeCount int
		wantInterval  time.Duration
	}{
		{0, 30 * time.Second},
		{1, 60 * time.Second},
		{2, 2 * time.Minute},
		{3, 4 * time.Minute},
		{4, 8 * time.Minute},
		{10, 8 * time.Minute},
	}

	for _, tt := range tests {
		ps.mu.Lock()
		ps.states["acct1"].consecutiveNoChange = tt.noChangeCount
		got := ps.effectiveIntervalLocked(ps.states["acct1"])
		ps.mu.Unlock()
		if got != tt.wantInterval {
			t.Errorf("noChange=%d: got %s, want %s", tt.noChangeCount, got, tt.wantInterval)
		}
	}
}

func TestPollScheduler_BackoffTiers_LocalProvider(t *testing.T) {
	ps := newPollScheduler(30 * time.Second)

	ps.ShouldPoll("acct1", true) // hasLocalDetector=true

	tests := []struct {
		noChangeCount int
		wantInterval  time.Duration
	}{
		{0, 30 * time.Second},
		{3, 60 * time.Second},
		{6, 180 * time.Second},
		{11, 240 * time.Second},
		{21, 480 * time.Second}, // 16x cap for local providers
	}

	for _, tt := range tests {
		ps.mu.Lock()
		ps.states["acct1"].consecutiveNoChange = tt.noChangeCount
		got := ps.effectiveIntervalLocked(ps.states["acct1"])
		ps.mu.Unlock()
		if got != tt.wantInterval {
			t.Errorf("noChange=%d: got %s, want %s", tt.noChangeCount, got, tt.wantInterval)
		}
	}
}

func TestPollScheduler_ResetOnChange(t *testing.T) {
	ps := newPollScheduler(30 * time.Second)

	ps.ShouldPoll("acct1", false)

	// Simulate 10 no-change polls.
	for i := 0; i < 10; i++ {
		ps.RecordPoll("acct1", false)
	}

	ps.mu.Lock()
	noChange := ps.states["acct1"].consecutiveNoChange
	ps.mu.Unlock()
	if noChange != 10 {
		t.Fatalf("expected 10 consecutive no-change, got %d", noChange)
	}

	// A changed poll resets to 0.
	ps.RecordPoll("acct1", true)

	ps.mu.Lock()
	noChange = ps.states["acct1"].consecutiveNoChange
	ps.mu.Unlock()
	if noChange != 0 {
		t.Errorf("expected 0 after change, got %d", noChange)
	}
}

func TestPollScheduler_SnapshotChanged(t *testing.T) {
	ps := newPollScheduler(30 * time.Second)

	snap1 := core.UsageSnapshot{
		Status: core.StatusOK,
		Metrics: map[string]core.Metric{
			"requests": {Used: ptr(100.0)},
		},
	}

	// First time is always "changed".
	if !ps.SnapshotChanged("acct1", snap1) {
		t.Error("first snapshot should be reported as changed")
	}

	// Same snapshot: not changed.
	if ps.SnapshotChanged("acct1", snap1) {
		t.Error("identical snapshot should not be reported as changed")
	}

	// Different snapshot: changed.
	snap2 := core.UsageSnapshot{
		Status: core.StatusOK,
		Metrics: map[string]core.Metric{
			"requests": {Used: ptr(200.0)},
		},
	}
	if !ps.SnapshotChanged("acct1", snap2) {
		t.Error("different snapshot should be reported as changed")
	}
}

func TestPollScheduler_HTTPMinimumInterval(t *testing.T) {
	ps := newPollScheduler(10 * time.Second)
	ps.ShouldPoll("acct1", false)

	ps.mu.Lock()
	ps.states["acct1"].consecutiveNoChange = 0
	got := ps.effectiveIntervalLocked(ps.states["acct1"])
	ps.mu.Unlock()
	if got != 30*time.Second {
		t.Errorf("HTTP provider at 1x with 10s base: got %s, want 30s floor", got)
	}
}

func TestPollScheduler_UnknownAccount(t *testing.T) {
	ps := newPollScheduler(30 * time.Second)
	ps.mu.Lock()
	got := ps.baseInterval
	if state := ps.states["nonexistent"]; state != nil {
		got = ps.effectiveIntervalLocked(state)
	}
	ps.mu.Unlock()
	if got != 30*time.Second {
		t.Errorf("unknown account should return base interval, got %s", got)
	}
}

func TestHTTPBasePollInterval(t *testing.T) {
	if got := HTTPBasePollInterval(1 * time.Second); got != 30*time.Second {
		t.Errorf("HTTPBasePollInterval(1s) = %s, want 30s floor", got)
	}
	if got := HTTPBasePollInterval(60 * time.Second); got != 60*time.Second {
		t.Errorf("HTTPBasePollInterval(60s) = %s, want 60s", got)
	}
}

func ptr(f float64) *float64 { return &f }

type testFakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func newTestFakeClock(t time.Time) *testFakeClock {
	return &testFakeClock{now: t}
}

func (f *testFakeClock) Now() time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.now
}

func (f *testFakeClock) Advance(d time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.now = f.now.Add(d)
}

func TestPollScheduler_FakeClock_DeterministicAdvance(t *testing.T) {
	t0 := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)
	clk := newTestFakeClock(t0)

	ps := newPollScheduler(30 * time.Second)
	ps.SetClock(clk)

	// First poll should run
	if !ps.ShouldPoll("acct1", false) {
		t.Fatal("first poll should run")
	}
	ps.RecordPoll("acct1", true)

	// Advance clock by 10s (< 30s base interval)
	clk.Advance(10 * time.Second)
	if ps.ShouldPoll("acct1", false) {
		t.Fatal("should not poll 10s after previous poll with 30s interval")
	}

	// Advance clock by another 25s (total 35s >= 30s)
	clk.Advance(25 * time.Second)
	if !ps.ShouldPoll("acct1", false) {
		t.Fatal("should poll 35s after previous poll with 30s interval")
	}
}

func TestPollScheduler_RateLimitBackoff(t *testing.T) {
	t0 := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)
	clk := newTestFakeClock(t0)

	ps := newPollScheduler(30 * time.Second)
	ps.SetClock(clk)

	// Initial poll
	if !ps.ShouldPoll("acct1", false) {
		t.Fatal("initial poll should run")
	}

	// Record rate limit until t0 + 60s
	resetAt := t0.Add(60 * time.Second)
	ps.RecordRateLimit("acct1", resetAt)

	// Immediate poll should be blocked by rate limit
	if ps.ShouldPoll("acct1", false) {
		t.Fatal("should be blocked while rate limited")
	}

	// Advance 30s (still before resetAt)
	clk.Advance(30 * time.Second)
	if ps.ShouldPoll("acct1", false) {
		t.Fatal("should still be blocked before rate limit reset")
	}

	// Advance another 35s (past resetAt)
	clk.Advance(35 * time.Second)
	if !ps.ShouldPoll("acct1", false) {
		t.Fatal("should allow poll after rate limit reset expired")
	}

	// Test fallback default backoff with future time
	ps.RecordRateLimit("acct2", t0.Add(5*time.Minute))
	if ps.ShouldPoll("acct2", false) {
		t.Fatal("should be blocked with future rate limit date")
	}
	// Clear rate limit
	ps.ClearRateLimit("acct2")
	if !ps.ShouldPoll("acct2", false) {
		t.Fatal("should allow poll after clearing rate limit")
	}
}

func TestPollScheduler_GenerationCoalescing(t *testing.T) {
	ps := newPollScheduler(30 * time.Second)

	// Begin first generation
	genID1, isLeader1, done1 := ps.BeginGeneration()
	if genID1 != 1 || !isLeader1 {
		t.Fatalf("first generation ID = %d, isLeader = %v, want 1 and true", genID1, isLeader1)
	}
	if ps.ActiveGenerationID() != 1 {
		t.Fatalf("active generation ID = %d, want 1", ps.ActiveGenerationID())
	}

	// Second concurrent begin returns the exact same generation
	genID2, isLeader2, done2 := ps.BeginGeneration()
	if genID2 != genID1 || isLeader2 {
		t.Fatalf("concurrent generation ID = %d (want %d), isLeader = %v (want false)", genID2, genID1, isLeader2)
	}
	if done2 != done1 {
		t.Fatal("concurrent generation Done channel must be identical")
	}

	// End generation
	ps.EndGeneration(genID1)

	select {
	case <-done1:
		// Expected: Done channel closed
	default:
		t.Fatal("expected done1 channel to be closed after EndGeneration")
	}

	if ps.ActiveGenerationID() != 0 {
		t.Fatalf("active generation ID = %d, want 0 after EndGeneration", ps.ActiveGenerationID())
	}

	// Subsequent generation gets incremented ID
	genID3, isLeader3, _ := ps.BeginGeneration()
	if genID3 != 2 || !isLeader3 {
		t.Fatalf("next generation ID = %d, isLeader = %v, want 2 and true", genID3, isLeader3)
	}
	ps.EndGeneration(genID3)
}
