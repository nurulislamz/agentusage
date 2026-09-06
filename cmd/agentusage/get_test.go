package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/nurulislamz/agentusage/internal/config"
	"github.com/nurulislamz/agentusage/internal/core"
	"github.com/nurulislamz/agentusage/internal/daemon"
	"github.com/nurulislamz/agentusage/internal/tui"
	"github.com/nurulislamz/agentusage/internal/webserve"
)

func TestFindAccount(t *testing.T) {
	accounts := []core.AccountConfig{
		{ID: "antigravity", Provider: "antigravity"},
		{ID: "antigravity-nurulz", Provider: "antigravity"},
		{ID: "cursor-physics", Provider: "cursor"},
		{ID: "opencode-mohammed", Provider: "opencode"},
	}
	accounts[1].SetHint("box_name", "nurulz")
	accounts[2].SetHint("box_name", "physics")

	// 1. Exact ID
	acct, ok := findAccount(accounts, "antigravity-nurulz")
	if !ok || acct.ID != "antigravity-nurulz" {
		t.Errorf("expected to find antigravity-nurulz, got %+v, ok=%v", acct, ok)
	}

	// 2. Case-insensitive exact ID
	acct, ok = findAccount(accounts, "ANTIGRAVITY-NURULZ")
	if !ok || acct.ID != "antigravity-nurulz" {
		t.Errorf("expected case-insensitive match, got %+v, ok=%v", acct, ok)
	}

	// 3. Antigravity box name match (nurulz -> antigravity-nurulz)
	acct, ok = findAccount(accounts, "nurulz")
	if !ok || acct.ID != "antigravity-nurulz" {
		t.Errorf("expected box name nurulz to resolve to antigravity-nurulz, got %+v, ok=%v", acct, ok)
	}

	// 4. Cursor box name match (physics -> cursor-physics)
	acct, ok = findAccount(accounts, "physics")
	if !ok || acct.ID != "cursor-physics" {
		t.Errorf("expected physics to resolve to cursor-physics, got %+v, ok=%v", acct, ok)
	}

	// 5. Antigravity agy- prefix alias (agy-nurulz -> antigravity-nurulz)
	acct, ok = findAccount(accounts, "agy-nurulz")
	if !ok || acct.ID != "antigravity-nurulz" {
		t.Errorf("expected agy-nurulz to resolve to antigravity-nurulz, got %+v, ok=%v", acct, ok)
	}

	// 6. Delimiter aliases with slash (cursor/physics -> cursor-physics, agy/nurulz -> antigravity-nurulz)
	acct, ok = findAccount(accounts, "cursor/physics")
	if !ok || acct.ID != "cursor-physics" {
		t.Errorf("expected cursor/physics to resolve to cursor-physics, got %+v, ok=%v", acct, ok)
	}
	acct, ok = findAccount(accounts, "agy/nurulz")
	if !ok || acct.ID != "antigravity-nurulz" {
		t.Errorf("expected agy/nurulz to resolve to antigravity-nurulz, got %+v, ok=%v", acct, ok)
	}

	// 7. Unknown account
	_, ok = findAccount(accounts, "nonexistent-box")
	if ok {
		t.Error("expected nonexistent-box to return false")
	}
}

func TestBuildGetResponse_FiveHourDefault(t *testing.T) {
	acct := core.AccountConfig{
		ID:       "antigravity-nurulz",
		Provider: "antigravity",
	}

	limit := 100.0
	geminiRemaining := 85.0
	geminiUsed := 15.0
	claudeRemaining := 95.0
	claudeUsed := 5.0

	resetTime := time.Now().UTC().Add(2 * time.Hour)

	snap := core.UsageSnapshot{
		ProviderID: "antigravity",
		AccountID:  "antigravity-nurulz",
		Status:     core.StatusOK,
		Metrics: map[string]core.Metric{
			"quota_gemini_5h": {
				Limit:     &limit,
				Used:      &geminiUsed,
				Remaining: &geminiRemaining,
				Unit:      "%",
				Window:    "5h",
			},
			"quota_claude_5h": {
				Limit:     &limit,
				Used:      &claudeUsed,
				Remaining: &claudeRemaining,
				Unit:      "%",
				Window:    "5h",
			},
		},
		Resets: map[string]time.Time{
			"quota_gemini_5h": resetTime,
		},
	}

	resp := buildGetResponse(acct, snap, "5h")

	if resp.ID != "antigravity-nurulz" {
		t.Errorf("expected ID antigravity-nurulz, got %s", resp.ID)
	}
	if resp.Window != "5h" {
		t.Errorf("expected Window 5h, got %s", resp.Window)
	}
	if resp.Remaining == nil || *resp.Remaining != 85.0 {
		t.Errorf("expected bottleneck remaining 85.0, got %v", resp.Remaining)
	}
	if len(resp.Pools) != 2 {
		t.Fatalf("expected 2 pools, got %d", len(resp.Pools))
	}
	if p, ok := resp.Pools["gemini_5h"]; !ok || *p.Remaining != 85.0 {
		t.Errorf("expected gemini_5h pool with 85.0 remaining, got %+v", p)
	}
	if p, ok := resp.Pools["claude_5h"]; !ok || *p.Remaining != 95.0 {
		t.Errorf("expected claude_5h pool with 95.0 remaining, got %+v", p)
	}
	if resp.ResetsIn == "" || !strings.Contains(resp.ResetsIn, "h") {
		t.Errorf("expected formatted ResetsIn countdown, got %q", resp.ResetsIn)
	}
}

func TestBuildGetResponse_WeeklyWindow(t *testing.T) {
	acct := core.AccountConfig{
		ID:       "antigravity-nurulz",
		Provider: "antigravity",
	}

	limit := 100.0
	geminiWeeklyRem := 60.0
	geminiWeeklyUsed := 40.0

	snap := core.UsageSnapshot{
		ProviderID: "antigravity",
		AccountID:  "antigravity-nurulz",
		Status:     core.StatusOK,
		Metrics: map[string]core.Metric{
			"quota_gemini_5h": {
				Limit:  &limit,
				Window: "5h",
			},
			"quota_gemini_weekly": {
				Limit:     &limit,
				Used:      &geminiWeeklyUsed,
				Remaining: &geminiWeeklyRem,
				Unit:      "%",
				Window:    "7d",
			},
		},
	}

	resp := buildGetResponse(acct, snap, "weekly")

	if resp.Window != "weekly" {
		t.Errorf("expected Window weekly, got %s", resp.Window)
	}
	if resp.Remaining == nil || *resp.Remaining != 60.0 {
		t.Errorf("expected remaining 60.0, got %v", resp.Remaining)
	}
	if _, ok := resp.Pools["gemini_weekly"]; !ok {
		t.Errorf("expected gemini_weekly pool in response: %+v", resp.Pools)
	}
}

func TestFormatDuration(t *testing.T) {
	cases := []struct {
		d    time.Duration
		want string
	}{
		{0, "now"},
		{-5 * time.Minute, "now"},
		{30 * time.Second, "30s"},
		{2*time.Minute + 15*time.Second, "2m 15s"},
		{3*time.Hour + 20*time.Minute, "3h 20m"},
		{26 * time.Hour, "1d 2h"},
	}

	for _, c := range cases {
		got := formatDuration(c.d)
		if got != c.want {
			t.Errorf("formatDuration(%v) = %q, want %q", c.d, got, c.want)
		}
	}
}

func TestFetchAccountSnapshot_SuccessAndPolling(t *testing.T) {
	pollCount := 0
	readModelCount := 0
	expectedSnap := core.UsageSnapshot{
		ProviderID: "antigravity",
		AccountID:  "antigravity-nurulz",
		Status:     core.StatusOK,
	}

	mockClient := daemon.NewMockClient("/tmp/mock.sock", mockDaemonRT(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path == "/v1/poll" {
			pollCount++
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"status":"polled"}`))),
				Header:     make(http.Header),
			}, nil
		}
		if req.URL.Path == "/v1/read-model" {
			readModelCount++
			resp := daemon.ReadModelResponse{
				Snapshots: map[string]core.UsageSnapshot{
					"antigravity-nurulz": expectedSnap,
				},
			}
			body, _ := json.Marshal(resp)
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader(body)),
				Header:     make(http.Header),
			}, nil
		}
		return &http.Response{StatusCode: http.StatusNotFound}, nil
	}))

	origEnsure := ensureDaemonClientFunc
	ensureDaemonClientFunc = func(ctx context.Context) (*daemon.Client, error) {
		return mockClient, nil
	}
	defer func() { ensureDaemonClientFunc = origEnsure }()

	acct := core.AccountConfig{ID: "antigravity-nurulz", Provider: "antigravity"}
	cfg := config.DefaultConfig()

	snap, err := fetchAccountSnapshot(context.Background(), acct, cfg)
	if err != nil {
		t.Fatalf("fetchAccountSnapshot unexpected error: %v", err)
	}
	if snap.AccountID != "antigravity-nurulz" {
		t.Errorf("snap.AccountID = %q, want antigravity-nurulz", snap.AccountID)
	}
	if pollCount != 1 {
		t.Errorf("pollCount = %d, want 1", pollCount)
	}
	if readModelCount != 1 {
		t.Errorf("readModelCount = %d, want 1", readModelCount)
	}
}

func TestFetchAccountSnapshot_ErrorSemantics(t *testing.T) {
	origEnsure := ensureDaemonClientFunc
	defer func() { ensureDaemonClientFunc = origEnsure }()

	acct := core.AccountConfig{ID: "antigravity-nurulz", Provider: "antigravity"}
	cfg := config.DefaultConfig()

	// 1. Daemon unavailable must error and NOT return zero quota
	ensureDaemonClientFunc = func(ctx context.Context) (*daemon.Client, error) {
		return nil, errors.New("daemon process failed to start")
	}
	snap, err := fetchAccountSnapshot(context.Background(), acct, cfg)
	if err == nil {
		t.Fatal("expected error when daemon is unavailable, got nil")
	}
	if snap.Status != "" {
		t.Errorf("expected empty snapshot status on error, got %v", snap.Status)
	}

	// 2. Context deadline exceeded during poll wait
	mockClientTimeout := daemon.NewMockClient("/tmp/mock.sock", mockDaemonRT(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path == "/v1/poll" {
			time.Sleep(50 * time.Millisecond)
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewReader([]byte(`{}`)))}, nil
	}))
	ensureDaemonClientFunc = func(ctx context.Context) (*daemon.Client, error) {
		return mockClientTimeout, nil
	}
	ctxTimeout, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	_, err = fetchAccountSnapshot(ctxTimeout, acct, cfg)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}

	// 3. Account missing from read-model snapshots
	mockClientEmpty := daemon.NewMockClient("/tmp/mock.sock", mockDaemonRT(func(req *http.Request) (*http.Response, error) {
		resp := daemon.ReadModelResponse{
			Snapshots: map[string]core.UsageSnapshot{},
		}
		body, _ := json.Marshal(resp)
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	}))
	ensureDaemonClientFunc = func(ctx context.Context) (*daemon.Client, error) {
		return mockClientEmpty, nil
	}
	_, err = fetchAccountSnapshot(context.Background(), acct, cfg)
	if err == nil {
		t.Fatal("expected error when account missing from snapshots")
	}
	if !strings.Contains(err.Error(), "no usage snapshot found") {
		t.Errorf("error = %q, want 'no usage snapshot found'", err.Error())
	}
}

func TestParityFixtures_Get_Web_TUI(t *testing.T) {
	fixedNow := time.Now().UTC().Truncate(time.Second)
	resetTime := fixedNow.Add(2 * time.Hour)
	limit := 100.0
	used := 15.0
	remaining := 85.0

	snap := core.UsageSnapshot{
		ProviderID: "antigravity",
		AccountID:  "antigravity-nurulz",
		Status:     core.StatusOK,
		Timestamp:  fixedNow,
		Metrics: map[string]core.Metric{
			"quota_gemini_5h": {
				Limit:     &limit,
				Used:      &used,
				Remaining: &remaining,
				Unit:      "%",
				Window:    "5h",
			},
		},
		Resets: map[string]time.Time{
			"quota_gemini_5h": resetTime,
		},
	}

	acct := core.AccountConfig{
		ID:       "antigravity-nurulz",
		Provider: "antigravity",
	}

	// 1. Presentation: get command JSON response
	getResp := buildGetResponse(acct, snap, "5h")
	if getResp.Remaining == nil || *getResp.Remaining != 85.0 {
		t.Fatalf("get remaining = %v, want 85.0", getResp.Remaining)
	}
	if !strings.EqualFold(getResp.Status, "ok") {
		t.Fatalf("get status = %q, want ok", getResp.Status)
	}
	if getResp.ResetsIn == "" || (!strings.Contains(getResp.ResetsIn, "2h") && !strings.Contains(getResp.ResetsIn, "1h")) {
		t.Fatalf("get resets_in = %q, want approx 2h countdown", getResp.ResetsIn)
	}

	// 2. Presentation: webserve envelope
	opts := webserve.Options{
		TimeWindow: "5h",
		Theme:      "Deep Space",
		Now:        func() time.Time { return fixedNow },
		Collect: func() (webserve.Envelope, error) {
			return webserve.Envelope{
				TimeWindow: "5h",
				Snapshots:  []core.UsageSnapshot{snap},
			}, nil
		},
	}
	srv, err := webserve.NewServer(opts)
	if err != nil {
		t.Fatalf("webserve.NewServer: %v", err)
	}
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/snapshots", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("web serve status = %d", w.Code)
	}
	var env webserve.Envelope
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatalf("decode web envelope: %v", err)
	}
	if len(env.Views) != 1 {
		t.Fatalf("web views count = %d, want 1", len(env.Views))
	}
	webView := env.Views[0]
	if webView.AccountID != "antigravity-nurulz" {
		t.Errorf("web account ID = %q, want antigravity-nurulz", webView.AccountID)
	}

	// Verify TUI-Web parity tool reports 0 issues
	issues := webserve.VerifyTUIWebParity(opts, env)
	if len(issues) > 0 {
		t.Fatalf("VerifyTUIWebParity issues: %+v", issues)
	}

	// 3. Presentation: TUI model detail projection
	cfg := config.DefaultConfig()
	cfg.Data.TimeWindow = "5h"
	proj := tui.NewWebProjectorFromConfig(cfg)
	proj.Now = fixedNow
	tuiSnapMap := map[string]core.UsageSnapshot{
		"antigravity-nurulz": snap,
	}
	ordered := proj.OrderSnapshots(tuiSnapMap)
	if len(ordered) != 1 {
		t.Fatalf("tui ordered snapshots count = %d, want 1", len(ordered))
	}
	tuiSnap := ordered[0]
	if tuiSnap.Metrics["quota_gemini_5h"].Remaining == nil || *tuiSnap.Metrics["quota_gemini_5h"].Remaining != 85.0 {
		t.Errorf("tui remaining = %v, want 85.0", tuiSnap.Metrics["quota_gemini_5h"].Remaining)
	}
	if tuiSnap.Status != core.StatusOK {
		t.Errorf("tui status = %v, want StatusOK", tuiSnap.Status)
	}
}
