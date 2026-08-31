package daemon

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
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

func TestStartBroadcaster_And_WarmUp(t *testing.T) {
	rt := NewViewRuntime(nil, "", false)
	var handledFrames []SnapshotFrame
	handler := func(f SnapshotFrame) {
		handledFrames = append(handledFrames, f)
	}
	stateEmitted := 0
	stateHandler := func(st DaemonState) {
		stateEmitted++
	}

	// 1. WarmUp with canceled context
	ctxCancel, cancel := context.WithCancel(context.Background())
	cancel()

	cancelled := warmUp(ctxCancel, rt, handler, func() { stateEmitted++ })
	if !cancelled {
		t.Error("warmUp with canceled ctx should return true")
	}

	// 2. WarmUp with live mock client
	mockClient := &Client{
		SocketPath: "/tmp/mock.sock",
		http: &http.Client{
			Transport: mockDaemonRoundTripper(func(req *http.Request) (*http.Response, error) {
				used := 10.0
				resp := ReadModelResponse{
					Snapshots: map[string]core.UsageSnapshot{
						"acc1": {
							ProviderID: "prov1",
							AccountID:  "acc1",
							Status:     core.StatusOK,
							Metrics: map[string]core.Metric{
								"cost": {Used: &used, Unit: "USD"},
							},
						},
					},
				}
				body, _ := json.Marshal(resp)
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(bytes.NewReader(body)),
					Header:     make(http.Header),
				}, nil
			}),
		},
	}
	rtMock := NewViewRuntime(nil, "/tmp/mock.sock", false)
	rtMock.SetClient(mockClient)

	cancelledLive := warmUp(context.Background(), rtMock, handler, func() { stateEmitted++ })
	if cancelledLive {
		t.Error("warmUp with live mock client should return false")
	}

	// 3. StartBroadcaster with live mock client
	ctxBroadcast, cancelBroadcast := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancelBroadcast()
	StartBroadcaster(ctxBroadcast, rtMock, 10*time.Millisecond, handler, stateHandler)
	time.Sleep(50 * time.Millisecond)
}

type mockDaemonRoundTripper func(*http.Request) (*http.Response, error)

func (m mockDaemonRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return m(r)
}

func TestFetchReadModel_Success_And_RefreshForWindow(t *testing.T) {
	ctx := context.Background()
	rt := NewViewRuntime(nil, "/tmp/mock.sock", false)

	mockClient := &Client{
		SocketPath: "/tmp/mock.sock",
		http: &http.Client{
			Transport: mockDaemonRoundTripper(func(req *http.Request) (*http.Response, error) {
				if strings.Contains(req.URL.Path, "read-model") {
					resp := ReadModelResponse{
						Snapshots: map[string]core.UsageSnapshot{
							"acc1": {ProviderID: "prov1", AccountID: "acc1", Status: core.StatusOK},
						},
					}
					body, _ := json.Marshal(resp)
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewReader(body)),
						Header:     make(http.Header),
					}, nil
				}
				if strings.Contains(req.URL.Path, "poll") {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(`{"status":"polled"}`)),
						Header:     make(http.Header),
					}, nil
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader(`{}`)),
					Header:     make(http.Header),
				}, nil
			}),
		},
	}

	rt.SetClient(mockClient)

	// 1. fetchReadModel success
	snaps, err := rt.fetchReadModel(ctx, mockClient, ReadModelRequest{})
	if err != nil {
		t.Fatalf("fetchReadModel error: %v", err)
	}
	if len(snaps) != 1 || snaps["acc1"].AccountID != "acc1" {
		t.Errorf("snapshots = %+v, want acc1", snaps)
	}
	if rt.State().Status != DaemonStatusRunning {
		t.Errorf("state status = %v, want Running", rt.State().Status)
	}

	// 2. RefreshForWindow with live client
	frame := rt.RefreshForWindow(ctx, core.TimeWindow7d)
	if len(frame.Snapshots) != 1 {
		t.Errorf("RefreshForWindow frame = %+v", frame)
	}
}

func TestEnsureClient_Branches(t *testing.T) {
	// 1. Nil runtime
	var nilRT *ViewRuntime
	if client := nilRT.EnsureClient(context.Background()); client != nil {
		t.Error("nilRT.EnsureClient should return nil")
	}

	// 2. Empty socket path
	rtEmpty := NewViewRuntime(nil, "", false)
	if client := rtEmpty.EnsureClient(context.Background()); client != nil {
		t.Error("rtEmpty.EnsureClient should return nil")
	}
	if rtEmpty.State().Status != DaemonStatusError {
		t.Errorf("rtEmpty status = %v, want DaemonStatusError", rtEmpty.State().Status)
	}

	// 3. TimeWindow getter/setter
	rt := NewViewRuntime(nil, "/tmp/test.sock", false)
	rt.SetTimeWindow(core.TimeWindow7d)
	if tw := rt.TimeWindow(); tw != core.TimeWindow7d {
		t.Errorf("TimeWindow() = %v, want 7d", tw)
	}

	// 4. ResetEnsureThrottle
	rt.ResetEnsureThrottle()
}
