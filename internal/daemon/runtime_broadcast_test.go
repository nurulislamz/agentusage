package daemon

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/nurulislamz/agentusage/internal/core"
)

func TestSnapshotFingerprint_Empty(t *testing.T) {
	fp := snapshotFingerprint(nil)
	if fp != "" {
		t.Fatalf("nil snapshots: got %q, want empty", fp)
	}

	fp = snapshotFingerprint(map[string]core.UsageSnapshot{})
	if fp != "" {
		t.Fatalf("empty snapshots: got %q, want empty", fp)
	}
}

func TestSnapshotFingerprint_Deterministic(t *testing.T) {
	ts := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)
	snaps := map[string]core.UsageSnapshot{
		"openai":    {Timestamp: ts},
		"anthropic": {Timestamp: ts.Add(time.Hour)},
	}

	fp1 := snapshotFingerprint(snaps)
	fp2 := snapshotFingerprint(snaps)
	if fp1 != fp2 {
		t.Fatalf("fingerprint not deterministic: %q != %q", fp1, fp2)
	}
}

func TestSnapshotFingerprint_DiffersOnTimestampChange(t *testing.T) {
	ts := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)
	snaps1 := map[string]core.UsageSnapshot{
		"openai": {Timestamp: ts},
	}
	snaps2 := map[string]core.UsageSnapshot{
		"openai": {Timestamp: ts.Add(time.Second)},
	}

	fp1 := snapshotFingerprint(snaps1)
	fp2 := snapshotFingerprint(snaps2)
	if fp1 == fp2 {
		t.Fatal("fingerprints should differ when timestamp changes")
	}
}

func TestSnapshotFingerprint_DiffersOnKeyChange(t *testing.T) {
	ts := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)
	snaps1 := map[string]core.UsageSnapshot{
		"openai": {Timestamp: ts},
	}
	snaps2 := map[string]core.UsageSnapshot{
		"anthropic": {Timestamp: ts},
	}

	fp1 := snapshotFingerprint(snaps1)
	fp2 := snapshotFingerprint(snaps2)
	if fp1 == fp2 {
		t.Fatal("fingerprints should differ when keys change")
	}
}

func TestSnapshotFingerprint_DiffersOnMetricCountChange(t *testing.T) {
	ts := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)
	snaps1 := map[string]core.UsageSnapshot{
		"claude_code": {Timestamp: ts, Metrics: map[string]core.Metric{"rpm": {}}},
	}
	snaps2 := map[string]core.UsageSnapshot{
		"claude_code": {Timestamp: ts, Metrics: map[string]core.Metric{"rpm": {}, "tpm": {}}},
	}

	fp1 := snapshotFingerprint(snaps1)
	fp2 := snapshotFingerprint(snaps2)
	if fp1 == fp2 {
		t.Fatal("fingerprints should differ when metric count changes (telemetry enrichment)")
	}
}

func TestViewRuntime_LifecycleAndNilSafety(t *testing.T) {
	// 1. Nil receiver safety
	var nilRuntime *ViewRuntime
	if client := nilRuntime.CurrentClient(); client != nil {
		t.Errorf("nil.CurrentClient() = %v, want nil", client)
	}
	nilRuntime.SetClient(nil)
	if st := nilRuntime.State(); st.Status != DaemonStatusUnknown {
		t.Errorf("nil.State() = %v, want Unknown", st.Status)
	}
	nilRuntime.SetTimeWindow(core.TimeWindow7d)
	if tw := nilRuntime.TimeWindow(); tw != core.TimeWindow30d {
		t.Errorf("nil.TimeWindow() = %v, want 30d", tw)
	}
	nilRuntime.ResetEnsureThrottle()
	if c := nilRuntime.EnsureClient(nil); c != nil {
		t.Errorf("nil.EnsureClient() = %v, want nil", c)
	}

	// 2. Normal ViewRuntime lifecycle
	rt := NewViewRuntime(nil, "/tmp/test.sock", false)
	if st := rt.State(); st.Status != DaemonStatusConnecting {
		t.Errorf("initial state = %v, want Connecting", st.Status)
	}

	mockClient := &Client{SocketPath: "/tmp/test.sock"}
	rt.SetClient(mockClient)
	if rt.CurrentClient() != mockClient {
		t.Errorf("CurrentClient() = %v, want %v", rt.CurrentClient(), mockClient)
	}

	rt.SetTimeWindow(core.TimeWindow7d)
	if rt.TimeWindow() != core.TimeWindow7d {
		t.Errorf("TimeWindow() = %v, want 7d", rt.TimeWindow())
	}

	rt.ResetEnsureThrottle()
	if rt.CurrentClient() != nil {
		t.Error("ResetEnsureThrottle should clear client")
	}
}

func TestViewRuntime_ReadWithFallback_And_Refresh(t *testing.T) {
	ctx := context.Background()

	// 1. Fallback when client fails or is nil
	rt := NewViewRuntime(nil, "", false)
	frame := rt.ReadWithFallback(ctx)
	if rt.State().Status != DaemonStatusError && rt.State().Status != DaemonStatusConnecting {
		t.Errorf("rt.State().Status = %v", rt.State().Status)
	}
	_ = frame

	// 2. ReadWithFallbackForWindow with offline client
	offlineClient := &Client{
		SocketPath: "/tmp/nonexistent_socket_test.sock",
		http:       &http.Client{Timeout: 100 * time.Millisecond},
	}
	rt.SetClient(offlineClient)
	frame2 := rt.ReadWithFallbackForWindow(ctx, core.TimeWindow30d)
	if rt.State().Status != DaemonStatusError && rt.State().Status != DaemonStatusConnecting {
		t.Errorf("rt.State().Status = %v", rt.State().Status)
	}
	_ = frame2

	// 3. RefreshForWindow with offline client returns frame
	frame3 := rt.RefreshForWindow(ctx, core.TimeWindow7d)
	_ = frame3
}
