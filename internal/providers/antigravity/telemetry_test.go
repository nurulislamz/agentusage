package antigravity

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/nurulislamz/agentusage/internal/core"
	"github.com/nurulislamz/agentusage/internal/providers/shared"
)

func TestTelemetrySource_SystemAndDefaults(t *testing.T) {
	p := New()
	if got := p.System(); got != "antigravity" {
		t.Errorf("System() = %q, want 'antigravity'", got)
	}

	opts := p.DefaultCollectOptions()
	if _, ok := opts.Paths["config_dir"]; !ok {
		t.Errorf("DefaultCollectOptions() missing config_dir path: %+v", opts)
	}
}

func TestTelemetrySource_Collect_ContextCancelled(t *testing.T) {
	p := New()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	events, err := p.Collect(ctx, shared.TelemetryCollectOptions{})
	if err == nil {
		t.Fatal("expected context cancelled error, got nil")
	}
	if len(events) != 0 {
		t.Errorf("expected 0 events, got %d", len(events))
	}
}

func TestTelemetrySource_Collect_SuccessfulEvents(t *testing.T) {
	quotaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
			"groups": [
				{
					"displayName": "Gemini Models",
					"buckets": [
						{"bucketId":"gemini-weekly","remainingFraction":0.80,"resetTime":"2030-09-02T03:31:19Z","window":"weekly"},
						{"bucketId":"gemini-5h","remainingFraction":0.50,"resetTime":"2030-08-29T16:49:43Z","window":"5h"}
					]
				}
			]
		}`))
	}))
	defer quotaServer.Close()

	configDir := t.TempDir()
	tokenFile := filepath.Join(configDir, "custom-token-file")
	writeTestToken(t, tokenFile, "access-token", "2030-01-01T00:00:00Z", "refresh-token")

	p := New()
	p.HTTPClient = quotaServer.Client()

	opts := shared.TelemetryCollectOptions{
		Paths: map[string]string{
			"account_id":       "box-dev-1",
			"config_dir":       configDir,
			"box_name":         "dev-1",
			"oauth_token_file": tokenFile,
			"quota_endpoint":   quotaServer.URL + "/v1internal",
		},
	}

	events, err := p.Collect(context.Background(), opts)
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("Collect() returned %d events, want 1", len(events))
	}

	evt := events[0]
	if evt.SchemaVersion != telemetrySchemaVersion {
		t.Errorf("SchemaVersion = %q, want %q", evt.SchemaVersion, telemetrySchemaVersion)
	}
	if evt.Channel != shared.TelemetryChannelAPI {
		t.Errorf("Channel = %q, want %q", evt.Channel, shared.TelemetryChannelAPI)
	}
	if evt.AccountID != "box-dev-1" {
		t.Errorf("AccountID = %q, want box-dev-1", evt.AccountID)
	}
	if evt.SessionID != "box-dev-1" {
		t.Errorf("SessionID = %q, want box-dev-1", evt.SessionID)
	}
	if evt.EventType != shared.TelemetryEventTypeRawEnvelope {
		t.Errorf("EventType = %q, want %q", evt.EventType, shared.TelemetryEventTypeRawEnvelope)
	}
	if evt.Status != shared.TelemetryStatusOK {
		t.Errorf("Status = %q, want %q", evt.Status, shared.TelemetryStatusOK)
	}
	if evt.Payload["source"] != "antigravity_quota_api" {
		t.Errorf("Payload.source = %v", evt.Payload["source"])
	}
	if evt.Payload["box"] != "dev-1" {
		t.Errorf("Payload.box = %v, want dev-1", evt.Payload["box"])
	}
	if evt.MessageID == "" {
		t.Error("expected non-empty MessageID")
	}
}

func TestTelemetrySource_Collect_AuthOrErrorYieldsNilEvents(t *testing.T) {
	// 1. Missing token yields StatusAuth -> nil events, nil error
	p := New()
	opts := shared.TelemetryCollectOptions{
		Paths: map[string]string{
			"config_dir": t.TempDir(),
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	events, err := p.Collect(ctx, opts)
	if err != nil {
		t.Fatalf("Collect() with auth error returned err: %v", err)
	}
	if events != nil {
		t.Fatalf("Collect() with auth error returned events: %v, want nil", events)
	}

	// 2. Empty buckets yields StatusError -> nil events, nil error
	quotaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"groups": []}`))
	}))
	defer quotaServer.Close()

	configDir := t.TempDir()
	writeTestToken(t, filepath.Join(configDir, oauthTokenFile), "access-token", "2030-01-01T00:00:00Z", "")

	p.HTTPClient = quotaServer.Client()
	opts = shared.TelemetryCollectOptions{
		Paths: map[string]string{
			"config_dir":     configDir,
			"quota_endpoint": quotaServer.URL + "/v1internal",
		},
	}

	events, err = p.Collect(context.Background(), opts)
	if err != nil {
		t.Fatalf("Collect() with status error returned err: %v", err)
	}
	if events != nil {
		t.Fatalf("Collect() with status error returned events: %v, want nil", events)
	}
}

func TestQuotaTelemetryEvent_Fallbacks(t *testing.T) {
	// Zero timestamp should default to current time
	snapZeroTime := core.UsageSnapshot{
		ProviderID: providerID,
		AccountID:  "antigravity",
		Status:     core.StatusOK,
		Attributes: map[string]string{"box": "main"},
		Metrics: map[string]core.Metric{
			"quota_gemini_5h": {Remaining: core.Float64Ptr(75.0)},
		},
		Raw: map[string]string{"quota_api": "ok (1 buckets)"},
	}

	evt := quotaTelemetryEvent(snapZeroTime, "")
	if evt.OccurredAt.IsZero() || time.Since(evt.OccurredAt) > time.Minute {
		t.Errorf("OccurredAt = %v, expected recent time", evt.OccurredAt)
	}
	if evt.AccountID != defaultAccountID {
		t.Errorf("AccountID = %q, want default %q", evt.AccountID, defaultAccountID)
	}

	// Non-zero timestamp and non-empty accountID
	customTime := time.Date(2026, 8, 31, 14, 30, 0, 0, time.UTC)
	snapCustom := snapZeroTime
	snapCustom.Timestamp = customTime

	evtCustom := quotaTelemetryEvent(snapCustom, "custom-acct")
	if !evtCustom.OccurredAt.Equal(customTime) {
		t.Errorf("OccurredAt = %v, want %v", evtCustom.OccurredAt, customTime)
	}
	if evtCustom.AccountID != "custom-acct" {
		t.Errorf("AccountID = %q, want custom-acct", evtCustom.AccountID)
	}
}

func TestQuotaRevision_DeterministicAndSensitive(t *testing.T) {
	snap1 := core.UsageSnapshot{
		Status: core.StatusOK,
		Metrics: map[string]core.Metric{
			"quota_b": {Remaining: core.Float64Ptr(50.0)},
			"quota_a": {Remaining: core.Float64Ptr(80.0)},
		},
	}
	snap2 := core.UsageSnapshot{
		Status: core.StatusOK,
		Metrics: map[string]core.Metric{
			"quota_a": {Remaining: core.Float64Ptr(80.0)},
			"quota_b": {Remaining: core.Float64Ptr(50.0)},
		},
	}
	snap3 := core.UsageSnapshot{
		Status: core.StatusOK,
		Metrics: map[string]core.Metric{
			"quota_a": {Remaining: core.Float64Ptr(80.0)},
			"quota_b": {Remaining: nil}, // nil remaining
		},
	}
	snap4 := core.UsageSnapshot{
		Status: core.StatusLimited,
		Metrics: map[string]core.Metric{
			"quota_a": {Remaining: core.Float64Ptr(80.0)},
			"quota_b": {Remaining: core.Float64Ptr(50.0)},
		},
	}

	rev1 := quotaRevision(snap1)
	rev2 := quotaRevision(snap2)
	rev3 := quotaRevision(snap3)
	rev4 := quotaRevision(snap4)

	if rev1 != rev2 {
		t.Errorf("rev1 (%q) != rev2 (%q) despite identical metrics in different order", rev1, rev2)
	}
	if rev1 == rev3 {
		t.Errorf("rev1 (%q) should differ from rev3 (%q) with nil metric remaining", rev1, rev3)
	}
	if rev1 == rev4 {
		t.Errorf("rev1 (%q) should differ from rev4 (%q) with different status", rev1, rev4)
	}
	if len(rev1) != 16 { // 8 bytes in hex = 16 hex chars
		t.Errorf("revision length = %d, want 16 hex characters", len(rev1))
	}
}

func TestSortedMetricKeys(t *testing.T) {
	snap := core.UsageSnapshot{
		Metrics: map[string]core.Metric{
			"zebra": {},
			"apple": {},
			"mango": {},
		},
	}
	keys := sortedMetricKeys(snap)
	if len(keys) != 3 || keys[0] != "apple" || keys[1] != "mango" || keys[2] != "zebra" {
		t.Errorf("sortedMetricKeys() = %v, want [apple mango zebra]", keys)
	}
}

func TestTelemetry_ConcurrencyUnderRace(t *testing.T) {
	quotaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
			"groups":[{"buckets":[{"bucketId":"gemini-5h","remainingFraction":0.9,"resetTime":"2030-01-01T00:00:00Z"}]}]
		}`))
	}))
	defer quotaServer.Close()

	configDir := t.TempDir()
	writeTestToken(t, filepath.Join(configDir, oauthTokenFile), "access-token", "2030-01-01T00:00:00Z", "")

	p := New()
	p.HTTPClient = quotaServer.Client()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			opts := shared.TelemetryCollectOptions{
				Paths: map[string]string{
					"account_id":     "concurrent-account",
					"config_dir":     configDir,
					"quota_endpoint": quotaServer.URL + "/v1internal",
				},
			}
			events, err := p.Collect(context.Background(), opts)
			if err != nil {
				t.Errorf("Collect concurrent error = %v", err)
				return
			}
			if len(events) != 1 {
				t.Errorf("Collect returned %d events, want 1", len(events))
			}
		}(i)
	}
	wg.Wait()
}
