package antigravity

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/nurulislamz/agentusage/internal/core"
	"github.com/nurulislamz/agentusage/internal/providers/shared"
)

func TestProvider_MetadataAndWidgets(t *testing.T) {
	p := New()
	if p.ID() != "antigravity" {
		t.Errorf("ID() = %q, want antigravity", p.ID())
	}
	spec := p.Spec()
	if spec.Info.Name != "Antigravity CLI" {
		t.Errorf("Name = %q, want 'Antigravity CLI'", spec.Info.Name)
	}
	if spec.Auth.DefaultAccountID != "antigravity" {
		t.Errorf("DefaultAccountID = %q, want antigravity", spec.Auth.DefaultAccountID)
	}
	if len(spec.Setup.Quickstart) == 0 {
		t.Error("expected non-empty quickstart")
	}

	detail := p.DetailWidget()
	if len(detail.Sections) == 0 {
		t.Error("expected non-empty detail sections")
	}

	widget := dashboardWidget()
	if widget.ColorRole != core.DashboardColorRoleMauve {
		t.Errorf("ColorRole = %q, want %q", widget.ColorRole, core.DashboardColorRoleMauve)
	}
	if len(widget.CompactRows) != 2 {
		t.Errorf("CompactRows len = %d, want 2", len(widget.CompactRows))
	}
}

func TestParseHookPayloadRejected(t *testing.T) {
	if _, err := New().ParseHookPayload([]byte(`{}`), shared.TelemetryCollectOptions{}); err == nil {
		t.Fatal("expected status-line hook rejection")
	}
}

func TestFetchQuotaFromAPI(t *testing.T) {
	quotaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1internal:retrieveUserQuotaSummary" {
			http.NotFound(w, r)
			return
		}
		if got := r.Header.Get("User-Agent"); got != "antigravity" {
			t.Errorf("User-Agent = %q, want antigravity", got)
		}
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") {
			t.Errorf("missing bearer auth: %q", r.Header.Get("Authorization"))
		}
		_, _ = w.Write([]byte(`{
			"groups": [
				{
					"displayName": "Gemini Models",
					"buckets": [
						{"bucketId":"gemini-weekly","remainingFraction":0.3787134,"resetTime":"2030-09-02T03:31:19Z","window":"weekly"},
						{"bucketId":"gemini-5h","remainingFraction":0,"resetTime":"2030-08-29T16:49:43Z","window":"5h"}
					]
				},
				{
					"displayName": "Claude and GPT models",
					"buckets": [
						{"bucketId":"3p-weekly","remainingFraction":0,"resetTime":"2030-09-02T23:42:57Z","window":"weekly"},
						{"bucketId":"3p-5h","remainingFraction":0.6616888,"resetTime":"2030-08-29T19:39:50Z","window":"5h"}
					]
				}
			]
		}`))
	}))
	defer quotaServer.Close()

	configDir := t.TempDir()
	writeTestToken(t, filepath.Join(configDir, oauthTokenFile), "access-token", "2030-01-01T00:00:00Z", "refresh-token")

	p := New()
	p.HTTPClient = quotaServer.Client()
	snap, err := p.Fetch(context.Background(), core.AccountConfig{
		ID:       "antigravity-chaos",
		Provider: "antigravity",
		ProviderPaths: map[string]string{
			"config_dir": configDir,
		},
		RuntimeHints: map[string]string{
			"box_name":       "chaos",
			"quota_endpoint": quotaServer.URL + "/v1internal",
		},
	})
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if snap.Status != core.StatusLimited {
		t.Fatalf("status = %q, want %q", snap.Status, core.StatusLimited)
	}
	if got := *snap.Metrics["quota_gemini_5h"].Remaining; got != 0 {
		t.Errorf("quota_gemini_5h = %v, want 0", got)
	}
	if got := *snap.Metrics["quota_gemini_weekly"].Remaining; math.Abs(got-37.87134) > 0.001 {
		t.Errorf("quota_gemini_weekly = %v, want ~37.87134", got)
	}
	if got := *snap.Metrics["quota_claude_5h"].Remaining; math.Abs(got-66.16888) > 0.001 {
		t.Errorf("quota_claude_5h = %v, want ~66.16888", got)
	}
	if snap.Raw["quota_source"] != "retrieveUserQuotaSummary" {
		t.Errorf("quota_source = %q", snap.Raw["quota_source"])
	}
	if snap.Attributes["box"] != "chaos" {
		t.Errorf("box = %q, want chaos", snap.Attributes["box"])
	}
	if snap.Raw["oauth_status"] != "valid" {
		t.Errorf("oauth_status = %q, want valid", snap.Raw["oauth_status"])
	}
}

func TestFetch_ContextCancelled(t *testing.T) {
	p := New()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := p.Fetch(ctx, core.AccountConfig{
		ID:       "antigravity",
		Provider: "antigravity",
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Fetch() error = %v, want context.Canceled", err)
	}
}

func TestFetchMissingTokenIsAuth(t *testing.T) {
	p := New()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	snap, err := p.Fetch(ctx, core.AccountConfig{
		ID:       "antigravity",
		Provider: "antigravity",
		Binary:   filepath.Join(t.TempDir(), "no-such-agy"),
		ProviderPaths: map[string]string{
			"config_dir": t.TempDir(),
		},
	})
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if snap.Status != core.StatusAuth {
		t.Fatalf("status = %q, want %q", snap.Status, core.StatusAuth)
	}
	if snap.Diagnostics["auth_error"] == "" {
		t.Error("expected auth_error diagnostic")
	}
}

func TestFetchUsesValidAccessToken(t *testing.T) {
	quotaHits := 0
	quotaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		quotaHits++
		if r.Header.Get("Authorization") != "Bearer new-access" {
			t.Errorf("Authorization = %q", r.Header.Get("Authorization"))
		}
		_, _ = w.Write([]byte(`{
			"groups":[{"buckets":[
				{"bucketId":"gemini-5h","remainingFraction":0.5,"resetTime":"2030-01-01T00:00:00Z"},
				{"bucketId":"gemini-weekly","remainingFraction":0.9,"resetTime":"2030-01-08T00:00:00Z"}
			]}]
		}`))
	}))
	defer quotaServer.Close()

	configDir := t.TempDir()
	writeTestToken(t, filepath.Join(configDir, oauthTokenFile), "new-access", time.Now().UTC().Add(time.Hour).Format(time.RFC3339Nano), "refresh-token")

	p := New()
	p.HTTPClient = quotaServer.Client()
	snap, err := p.Fetch(context.Background(), core.AccountConfig{
		ID:       "antigravity-chaos",
		Provider: "antigravity",
		ProviderPaths: map[string]string{
			"config_dir": configDir,
		},
		RuntimeHints: map[string]string{
			"model":          "gemini-2.5-pro",
			"quota_endpoint": quotaServer.URL + "/v1internal",
		},
	})
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if snap.Status != core.StatusOK {
		t.Fatalf("status = %q, want ok; msg=%s diag=%v", snap.Status, snap.Message, snap.Diagnostics)
	}
	if quotaHits != 1 {
		t.Fatalf("quota hits = %d, want 1", quotaHits)
	}
	if got := *snap.Metrics["quota_gemini_5h"].Remaining; got != 50 {
		t.Errorf("quota_gemini_5h = %v, want 50", got)
	}
	if snap.Attributes["model"] != "gemini-2.5-pro" {
		t.Errorf("model = %q, want gemini-2.5-pro", snap.Attributes["model"])
	}
}

func TestFetch_EmptyBucketsIsError(t *testing.T) {
	quotaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"groups": []}`))
	}))
	defer quotaServer.Close()

	configDir := t.TempDir()
	writeTestToken(t, filepath.Join(configDir, oauthTokenFile), "valid-token", "2030-01-01T00:00:00Z", "")

	p := New()
	p.HTTPClient = quotaServer.Client()
	snap, err := p.Fetch(context.Background(), core.AccountConfig{
		ID:       "antigravity",
		Provider: "antigravity",
		ProviderPaths: map[string]string{
			"config_dir": configDir,
		},
		RuntimeHints: map[string]string{
			"quota_endpoint": quotaServer.URL + "/v1internal",
		},
	})
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if snap.Status != core.StatusError {
		t.Fatalf("status = %q, want %q", snap.Status, core.StatusError)
	}
	if !strings.Contains(snap.Message, "no buckets") {
		t.Errorf("Message = %q, want no buckets error", snap.Message)
	}
}

func TestFetch_APINonAuthError(t *testing.T) {
	quotaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}))
	defer quotaServer.Close()

	configDir := t.TempDir()
	writeTestToken(t, filepath.Join(configDir, oauthTokenFile), "valid-token", "2030-01-01T00:00:00Z", "")

	p := New()
	p.HTTPClient = quotaServer.Client()
	snap, err := p.Fetch(context.Background(), core.AccountConfig{
		ID:       "antigravity",
		Provider: "antigravity",
		ProviderPaths: map[string]string{
			"config_dir": configDir,
		},
		RuntimeHints: map[string]string{
			"quota_endpoint": quotaServer.URL + "/v1internal",
		},
	})
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if snap.Status != core.StatusError {
		t.Fatalf("status = %q, want %q", snap.Status, core.StatusError)
	}
	if !strings.Contains(snap.Message, "request failed") {
		t.Errorf("Message = %q, want request failed", snap.Message)
	}
}

func TestFetch_APIAuth401Error_RetrySucceeds(t *testing.T) {
	attempt := 0
	quotaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt++
		if attempt == 1 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`{
			"groups":[{"buckets":[
				{"bucketId":"gemini-5h","remainingFraction":0.8,"resetTime":"2030-01-01T00:00:00Z"}
			]}]
		}`))
	}))
	defer quotaServer.Close()

	home := t.TempDir()
	t.Setenv("HOME", home)
	configDir := t.TempDir()
	tokenPath := filepath.Join(configDir, oauthTokenFile)
	writeTestToken(t, tokenPath, "initial-token", "2030-01-01T00:00:00Z", "")

	// Create a mock agy binary that updates the token file on ping
	binDir := filepath.Join(home, ".local", "bin")
	_ = os.MkdirAll(binDir, 0o755)
	mockAgy := filepath.Join(binDir, "agy")
	mockScript := fmt.Sprintf("#!/bin/sh\ncat << 'EOF' > %q\n{\n  \"token\": {\n    \"access_token\": \"refreshed-token\",\n    \"expiry\": \"2030-01-01T00:00:00Z\"\n  }\n}\nEOF\nexit 0\n", tokenPath)
	if err := os.WriteFile(mockAgy, []byte(mockScript), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+":"+os.Getenv("PATH"))

	p := New()
	p.HTTPClient = quotaServer.Client()
	snap, err := p.Fetch(context.Background(), core.AccountConfig{
		ID:       "antigravity",
		Provider: "antigravity",
		Binary:   mockAgy,
		ProviderPaths: map[string]string{
			"config_dir": configDir,
		},
		RuntimeHints: map[string]string{
			"quota_endpoint": quotaServer.URL + "/v1internal",
		},
	})
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if snap.Status != core.StatusOK {
		t.Fatalf("status = %q, want ok; msg=%s", snap.Status, snap.Message)
	}
	if snap.Raw["oauth_status"] != "refreshed_after_401" {
		t.Errorf("oauth_status = %q, want refreshed_after_401", snap.Raw["oauth_status"])
	}
}

func TestFetch_APIAuth401Error_RetryFails(t *testing.T) {
	quotaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer quotaServer.Close()

	configDir := t.TempDir()
	writeTestToken(t, filepath.Join(configDir, oauthTokenFile), "rejected-token", "2030-01-01T00:00:00Z", "")

	p := New()
	p.HTTPClient = quotaServer.Client()
	snap, err := p.Fetch(context.Background(), core.AccountConfig{
		ID:       "antigravity",
		Provider: "antigravity",
		Binary:   "/no/such/agy-binary",
		ProviderPaths: map[string]string{
			"config_dir": configDir,
		},
		RuntimeHints: map[string]string{
			"quota_endpoint": quotaServer.URL + "/v1internal",
		},
	})
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if snap.Status != core.StatusAuth {
		t.Fatalf("status = %q, want %q", snap.Status, core.StatusAuth)
	}
	if !strings.Contains(snap.Message, "rejected credentials") {
		t.Errorf("Message = %q, want credentials rejected", snap.Message)
	}
}

func TestIsAuthHTTPError(t *testing.T) {
	if isAuthHTTPError(nil) {
		t.Error("nil error should not be auth HTTP error")
	}
	if !isAuthHTTPError(errors.New("retrieveUserQuotaSummary HTTP 401: Unauthorized")) {
		t.Error("HTTP 401 should be auth error")
	}
	if !isAuthHTTPError(errors.New("retrieveUserQuotaSummary HTTP 403: Forbidden")) {
		t.Error("HTTP 403 should be auth error")
	}
	if isAuthHTTPError(errors.New("retrieveUserQuotaSummary HTTP 500: Server Error")) {
		t.Error("HTTP 500 should not be auth error")
	}
	if isAuthHTTPError(errors.New("network connection reset")) {
		t.Error("network error should not be auth error")
	}
}

func TestProjectSnapshot_AttributesAndEdgeCases(t *testing.T) {
	// Nil snapshot should be safe no-op
	projectSnapshot(nil, statusLinePayload{})

	var snap core.UsageSnapshot
	snap.Attributes = make(map[string]string)
	snap.Metrics = make(map[string]core.Metric)
	snap.Resets = make(map[string]time.Time)

	payload := statusLinePayload{
		Product:  "antigravity-pro",
		PlanTier: "enterprise",
		Email:    "developer@example.com",
		Model: statusLineModel{
			ID:          "gemini-2.5-flash",
			DisplayName: "Gemini 2.5 Flash",
		},
		Quota: map[string]statusLineQuota{
			"gemini_5h": {
				RemainingFraction: core.Float64Ptr(0.75),
			},
		},
	}

	projectSnapshot(&snap, payload)

	if snap.Attributes["product"] != "antigravity-pro" {
		t.Errorf("product = %q", snap.Attributes["product"])
	}
	if snap.Attributes["plan_tier"] != "enterprise" {
		t.Errorf("plan_tier = %q", snap.Attributes["plan_tier"])
	}
	if snap.Attributes["account_email"] != "developer@example.com" {
		t.Errorf("account_email = %q", snap.Attributes["account_email"])
	}
	if snap.Attributes["model"] != "Gemini 2.5 Flash" {
		t.Errorf("model = %q", snap.Attributes["model"])
	}
	if snap.Attributes["model_id"] != "gemini-2.5-flash" {
		t.Errorf("model_id = %q", snap.Attributes["model_id"])
	}
	if snap.Message != "Antigravity quota" {
		t.Errorf("Message = %q, want 'Antigravity quota'", snap.Message)
	}

	// Test model fallback to ID when DisplayName is empty
	var snap2 core.UsageSnapshot
	snap2.Attributes = make(map[string]string)
	snap2.Metrics = make(map[string]core.Metric)
	snap2.Resets = make(map[string]time.Time)

	payload2 := statusLinePayload{
		Model: statusLineModel{
			ID: "claude-3-7-sonnet",
		},
		Quota: map[string]statusLineQuota{
			"claude_5h": {
				RemainingFraction: core.Float64Ptr(0.9),
			},
		},
	}
	projectSnapshot(&snap2, payload2)
	if snap2.Attributes["model"] != "claude-3-7-sonnet" {
		t.Errorf("model = %q, want claude-3-7-sonnet", snap2.Attributes["model"])
	}
}

func TestProjectQuotaMetrics_AdvancesPastReset(t *testing.T) {
	past := time.Now().UTC().Add(-10 * time.Minute)
	payload := statusLinePayload{
		ReceivedAt: past.Add(-5 * time.Minute),
		Quota: map[string]statusLineQuota{
			"gemini-weekly": {
				RemainingFraction: core.Float64Ptr(0),
				ResetTime:         past.Format(time.RFC3339Nano),
			},
			"gemini-5h": {
				RemainingFraction: core.Float64Ptr(0),
				ResetTime:         past.Format(time.RFC3339Nano),
			},
		},
	}
	var snap core.UsageSnapshot
	snap.Metrics = make(map[string]core.Metric)
	snap.Resets = make(map[string]time.Time)

	projectQuotaMetrics(&snap, payload)

	if rem := metricRemaining(t, snap, "quota_gemini_weekly"); rem != 100 {
		t.Fatalf("quota_gemini_weekly remaining = %v, want 100 after reset passed", rem)
	}
	if rem := metricRemaining(t, snap, "quota_gemini_5h"); rem != 100 {
		t.Fatalf("quota_gemini_5h remaining = %v, want 100 after reset passed", rem)
	}
	reset, ok := snap.Resets["quota_gemini_weekly"]
	if !ok {
		t.Fatal("missing quota_gemini_weekly reset")
	}
	if !reset.After(time.Now().UTC()) {
		t.Fatalf("reset time = %v, expected future timestamp after advancing weekly period", reset)
	}
}

func TestProjectQuotaMetrics_SynthesisAndAliasing(t *testing.T) {
	// Synthesize 5h from weekly when 5h is missing
	payload := statusLinePayload{
		Quota: map[string]statusLineQuota{
			"gemini_weekly": {
				RemainingFraction: core.Float64Ptr(0.85),
				ResetTime:         "2030-01-01T00:00:00Z",
			},
			"claude_weekly": {
				RemainingFraction: core.Float64Ptr(0.0),
				ResetTime:         "2030-01-01T00:00:00Z",
			},
			"disabled_pool": {
				RemainingFraction: core.Float64Ptr(0.1),
				Disabled:          true,
			},
			"nil_pool": {
				RemainingFraction: nil,
			},
		},
	}
	var snap core.UsageSnapshot
	snap.Metrics = make(map[string]core.Metric)
	snap.Resets = make(map[string]time.Time)

	projectQuotaMetrics(&snap, payload)

	if got := *snap.Metrics["quota_gemini_5h"].Remaining; got != 100 {
		t.Errorf("quota_gemini_5h synthesized = %v, want 100", got)
	}
	if got := *snap.Metrics["quota_claude_5h"].Remaining; got != 0 {
		t.Errorf("quota_claude_5h synthesized = %v, want 0 when weekly <= 0", got)
	}
	if got := *snap.Metrics["quota_3p_5h"].Remaining; got != 0 {
		t.Errorf("quota_3p_5h synthesized = %v, want 0", got)
	}

	// 3p_5h and 3p_weekly aliasing to claude_5h and claude_weekly
	var snap2 core.UsageSnapshot
	snap2.Metrics = make(map[string]core.Metric)
	snap2.Resets = make(map[string]time.Time)

	payload2 := statusLinePayload{
		Quota: map[string]statusLineQuota{
			"3p_5h": {
				RemainingFraction: core.Float64Ptr(0.44),
				ResetTime:         "2030-01-01T00:00:00Z",
			},
			"3p_weekly": {
				RemainingFraction: core.Float64Ptr(0.99),
				ResetTime:         "2030-01-08T00:00:00Z",
			},
		},
	}
	projectQuotaMetrics(&snap2, payload2)
	if got := *snap2.Metrics["quota_claude_5h"].Remaining; got != 44 {
		t.Errorf("quota_claude_5h alias = %v, want 44", got)
	}
	if got := *snap2.Metrics["quota_claude_weekly"].Remaining; got != 99 {
		t.Errorf("quota_claude_weekly alias = %v, want 99", got)
	}
}

func TestProjectQuotaMetrics_ActivePoolSelection(t *testing.T) {
	// Active pool Claude selects Claude worst over Gemini
	payload := statusLinePayload{
		Model: statusLineModel{DisplayName: "Claude 3.7 Sonnet"},
		Quota: map[string]statusLineQuota{
			"gemini_5h": {RemainingFraction: core.Float64Ptr(0.1)},
			"claude_5h": {RemainingFraction: core.Float64Ptr(0.8)},
		},
	}
	var snap core.UsageSnapshot
	snap.Metrics = make(map[string]core.Metric)
	snap.Resets = make(map[string]time.Time)

	projectQuotaMetrics(&snap, payload)
	if got := *snap.Metrics["quota"].Remaining; got != 80 {
		t.Errorf("overall quota for claude model = %v, want 80", got)
	}

	// Active pool Gemini selects Gemini worst over Claude
	payload.Model.DisplayName = "Gemini 2.5 Pro"
	snap = core.UsageSnapshot{
		Metrics: make(map[string]core.Metric),
		Resets:  make(map[string]time.Time),
	}
	projectQuotaMetrics(&snap, payload)
	if got := *snap.Metrics["quota"].Remaining; got != 10 {
		t.Errorf("overall quota for gemini model = %v, want 10", got)
	}

	// No active pool takes max of both pools
	payload.Model.DisplayName = "custom-orchestrator"
	snap = core.UsageSnapshot{
		Metrics: make(map[string]core.Metric),
		Resets:  make(map[string]time.Time),
	}
	projectQuotaMetrics(&snap, payload)
	if got := *snap.Metrics["quota"].Remaining; got != 80 {
		t.Errorf("overall quota for untargeted model = %v, want 80 (max of 10 and 80)", got)
	}

	// Empty quota yields no overall quota metric
	snapEmpty := core.UsageSnapshot{
		Metrics: make(map[string]core.Metric),
		Resets:  make(map[string]time.Time),
	}
	projectQuotaMetrics(&snapEmpty, statusLinePayload{})
	if _, ok := snapEmpty.Metrics["quota"]; ok {
		t.Error("expected no overall quota metric for empty payload")
	}
}

func TestStatusFromQuota_ComprehensiveMatrix(t *testing.T) {
	tests := []struct {
		name       string
		model      string
		snapAttr   string
		quotas     map[string]statusLineQuota
		wantStatus core.Status
	}{
		{
			name:  "gemini model - exhausted (0%)",
			model: "gemini-2.5-pro",
			quotas: map[string]statusLineQuota{
				"gemini_5h": {RemainingFraction: core.Float64Ptr(0.0)},
			},
			wantStatus: core.StatusLimited,
		},
		{
			name:  "gemini model - near limit (10%)",
			model: "gemini-2.5-pro",
			quotas: map[string]statusLineQuota{
				"gemini_5h": {RemainingFraction: core.Float64Ptr(0.10)},
			},
			wantStatus: core.StatusNearLimit,
		},
		{
			name:  "gemini model - ok (50%)",
			model: "gemini-2.5-pro",
			quotas: map[string]statusLineQuota{
				"gemini_5h": {RemainingFraction: core.Float64Ptr(0.50)},
			},
			wantStatus: core.StatusOK,
		},
		{
			name:  "claude model - exhausted",
			model: "claude-3-7-sonnet",
			quotas: map[string]statusLineQuota{
				"claude_5h": {RemainingFraction: core.Float64Ptr(0.0)},
			},
			wantStatus: core.StatusLimited,
		},
		{
			name:  "claude model - near limit",
			model: "claude-3-7-sonnet",
			quotas: map[string]statusLineQuota{
				"claude_5h": {RemainingFraction: core.Float64Ptr(0.14)},
			},
			wantStatus: core.StatusNearLimit,
		},
		{
			name:  "claude model - ok",
			model: "claude-3-7-sonnet",
			quotas: map[string]statusLineQuota{
				"claude_5h": {RemainingFraction: core.Float64Ptr(0.90)},
			},
			wantStatus: core.StatusOK,
		},
		{
			name:  "no model specified - both exhausted -> Limited",
			model: "",
			quotas: map[string]statusLineQuota{
				"gemini_5h": {RemainingFraction: core.Float64Ptr(0.0)},
				"claude_5h": {RemainingFraction: core.Float64Ptr(0.0)},
			},
			wantStatus: core.StatusLimited,
		},
		{
			name:  "no model specified - both near limit -> NearLimit",
			model: "",
			quotas: map[string]statusLineQuota{
				"gemini_5h": {RemainingFraction: core.Float64Ptr(0.10)},
				"claude_5h": {RemainingFraction: core.Float64Ptr(0.12)},
			},
			wantStatus: core.StatusNearLimit,
		},
		{
			name:  "no model specified - one 0 and other near limit -> NearLimit",
			model: "",
			quotas: map[string]statusLineQuota{
				"gemini_5h": {RemainingFraction: core.Float64Ptr(0.0)},
				"claude_5h": {RemainingFraction: core.Float64Ptr(0.10)},
			},
			wantStatus: core.StatusNearLimit,
		},
		{
			name:  "no model specified - reverse one 0 and other near limit -> NearLimit",
			model: "",
			quotas: map[string]statusLineQuota{
				"gemini_5h": {RemainingFraction: core.Float64Ptr(0.10)},
				"claude_5h": {RemainingFraction: core.Float64Ptr(0.0)},
			},
			wantStatus: core.StatusNearLimit,
		},
		{
			name:  "no model specified - one 0 and other healthy -> OK",
			model: "",
			quotas: map[string]statusLineQuota{
				"gemini_5h": {RemainingFraction: core.Float64Ptr(0.0)},
				"claude_5h": {RemainingFraction: core.Float64Ptr(0.80)},
			},
			wantStatus: core.StatusOK,
		},
		{
			name:  "generic pool only - exhausted -> Limited",
			model: "",
			quotas: map[string]statusLineQuota{
				"custom_pool": {RemainingFraction: core.Float64Ptr(0.0)},
			},
			wantStatus: core.StatusLimited,
		},
		{
			name:  "generic pool only - near limit -> NearLimit",
			model: "",
			quotas: map[string]statusLineQuota{
				"custom_pool": {RemainingFraction: core.Float64Ptr(0.08)},
			},
			wantStatus: core.StatusNearLimit,
		},
		{
			name:  "generic pool only - healthy -> OK",
			model: "",
			quotas: map[string]statusLineQuota{
				"custom_pool": {RemainingFraction: core.Float64Ptr(0.75)},
			},
			wantStatus: core.StatusOK,
		},
		{
			name:       "empty quota -> OK",
			model:      "",
			quotas:     map[string]statusLineQuota{},
			wantStatus: core.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			snap := &core.UsageSnapshot{
				Attributes: make(map[string]string),
			}
			if tc.snapAttr != "" {
				snap.Attributes["model"] = tc.snapAttr
			}
			payload := statusLinePayload{
				Model: statusLineModel{DisplayName: tc.model},
				Quota: tc.quotas,
			}
			got := statusFromQuota(snap, payload)
			if got != tc.wantStatus {
				t.Errorf("statusFromQuota() = %q, want %q", got, tc.wantStatus)
			}
		})
	}
}

func TestWorstQuotaFraction_AndClamp(t *testing.T) {
	// Clamp tests
	if got := clamp(-5, 0, 10); got != 0 {
		t.Errorf("clamp(-5, 0, 10) = %v, want 0", got)
	}
	if got := clamp(15, 0, 10); got != 10 {
		t.Errorf("clamp(15, 0, 10) = %v, want 10", got)
	}
	if got := clamp(5, 0, 10); got != 5 {
		t.Errorf("clamp(5, 0, 10) = %v, want 5", got)
	}

	// Empty payload
	if _, ok := worstQuotaFraction(statusLinePayload{}); ok {
		t.Error("worstQuotaFraction should be false for empty payload")
	}

	// Payload with all nil or disabled
	payload := statusLinePayload{
		Quota: map[string]statusLineQuota{
			"q1": {RemainingFraction: nil},
			"q2": {RemainingFraction: core.Float64Ptr(0.2), Disabled: true},
		},
	}
	if _, ok := worstQuotaFraction(payload); ok {
		t.Error("worstQuotaFraction should be false when all disabled/nil")
	}

	// Payload with valid items
	payload.Quota["q3"] = statusLineQuota{RemainingFraction: core.Float64Ptr(0.35)}
	payload.Quota["q4"] = statusLineQuota{RemainingFraction: core.Float64Ptr(0.85)}
	if worst, ok := worstQuotaFraction(payload); !ok || math.Abs(worst-0.35) > 0.001 {
		t.Errorf("worstQuotaFraction = (%v, %v), want (0.35, true)", worst, ok)
	}
}

func TestSanitizeMetricName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"gemini-5h", "gemini_5h"},
		{"  Claude 3.7 Sonnet Weekly!  ", "claude_3_7_sonnet_weekly"},
		{"___leading_and_trailing___", "leading_and_trailing"},
		{"3P-WEEKLY", "3p_weekly"},
		{"alpha--beta__gamma", "alpha_beta_gamma"},
	}
	for _, tc := range tests {
		got := sanitizeMetricName(tc.input)
		if got != tc.want {
			t.Errorf("sanitizeMetricName(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestQuotaResetTime_Parsing(t *testing.T) {
	receivedAt := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)

	// RFC3339Nano
	qNano := statusLineQuota{ResetTime: "2026-08-31T17:00:00.123456789Z"}
	if got := quotaResetTime(qNano, receivedAt); got.IsZero() || got.Nanosecond() != 123456789 {
		t.Errorf("quotaResetTime RFC3339Nano = %v", got)
	}

	// RFC3339
	qRFC := statusLineQuota{ResetTime: "2026-08-31T17:00:00Z"}
	if got := quotaResetTime(qRFC, receivedAt); got.IsZero() || got.Hour() != 17 {
		t.Errorf("quotaResetTime RFC3339 = %v", got)
	}

	// ResetInSeconds fallback
	sec := int64(3600)
	qSec := statusLineQuota{ResetInSeconds: &sec}
	wantSec := receivedAt.Add(time.Hour)
	if got := quotaResetTime(qSec, receivedAt); !got.Equal(wantSec) {
		t.Errorf("quotaResetTime ResetInSeconds = %v, want %v", got, wantSec)
	}

	// Empty ResetTime and nil ResetInSeconds
	qEmpty := statusLineQuota{}
	if got := quotaResetTime(qEmpty, receivedAt); !got.IsZero() {
		t.Errorf("quotaResetTime empty = %v, want zero time", got)
	}

	// Invalid ResetTime format and nil ResetInSeconds
	qInvalid := statusLineQuota{ResetTime: "invalid-time-format"}
	if got := quotaResetTime(qInvalid, receivedAt); !got.IsZero() {
		t.Errorf("quotaResetTime invalid = %v, want zero time", got)
	}
}

func TestPayloadReceivedAt(t *testing.T) {
	customTime := time.Date(2026, 8, 31, 10, 0, 0, 0, time.UTC)
	pCustom := statusLinePayload{ReceivedAt: customTime}
	if got := payloadReceivedAt(pCustom); !got.Equal(customTime) {
		t.Errorf("payloadReceivedAt(custom) = %v, want %v", got, customTime)
	}

	pZero := statusLinePayload{}
	if got := payloadReceivedAt(pZero); got.IsZero() || time.Since(got) > time.Minute {
		t.Errorf("payloadReceivedAt(zero) = %v, expected recent time", got)
	}
}

func TestQuotaMapFromSummary(t *testing.T) {
	summary := quotaSummaryResponse{
		Groups: []quotaGroup{
			{
				Buckets: []quotaBucket{
					{BucketID: "gemini-5h", RemainingFraction: core.Float64Ptr(0.5), ResetTime: "2030-01-01T00:00:00Z"},
					{BucketID: "   ", RemainingFraction: core.Float64Ptr(0.1)}, // empty BucketID should be skipped
				},
			},
		},
	}
	got := quotaMapFromSummary(summary)
	if len(got) != 1 || got["gemini-5h"].RemainingFraction == nil || *got["gemini-5h"].RemainingFraction != 0.5 {
		t.Fatalf("unexpected map: %+v", got)
	}
}

func TestRetrieveUserQuotaSummary_Branches(t *testing.T) {
	// 1. Nil client creates default client, empty baseURL uses defaultQuotaEndpoint
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, ":retrieveUserQuotaSummary") {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(`{"groups":[]}`))
	}))
	defer ts.Close()

	resp, err := retrieveUserQuotaSummary(context.Background(), "token", ts.URL+"/v1internal", nil)
	if err != nil {
		t.Fatalf("retrieveUserQuotaSummary() error = %v", err)
	}
	if len(resp.Groups) != 0 {
		t.Errorf("groups len = %d, want 0", len(resp.Groups))
	}

	// 2. HTTP Non-200 Error
	tsErr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "permission denied", http.StatusForbidden)
	}))
	defer tsErr.Close()

	_, err = retrieveUserQuotaSummary(context.Background(), "token", tsErr.URL+"/v1internal", tsErr.Client())
	if err == nil || !strings.Contains(err.Error(), "HTTP 403") {
		t.Errorf("expected HTTP 403 error, got %v", err)
	}

	// 3. Malformed JSON
	tsBadJSON := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{not-json`))
	}))
	defer tsBadJSON.Close()

	_, err = retrieveUserQuotaSummary(context.Background(), "token", tsBadJSON.URL+"/v1internal", tsBadJSON.Client())
	if err == nil || !strings.Contains(err.Error(), "parse retrieveUserQuotaSummary") {
		t.Errorf("expected unmarshal error, got %v", err)
	}

	// 4. Invalid URL / request creation error
	_, err = retrieveUserQuotaSummary(context.Background(), "token", "http://invalid domain with spaces", nil)
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}

func TestEnrichSnapshots_NilAndMultiple(t *testing.T) {
	// Nil provider should safe no-op
	var nilProvider *Provider
	nilProvider.EnrichSnapshots(context.Background(), nil, nil)

	quotaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
			"groups":[{"buckets":[
				{"bucketId":"gemini-5h","remainingFraction":0.42,"resetTime":"2030-01-01T00:00:00Z"},
				{"bucketId":"gemini-weekly","remainingFraction":0.8,"resetTime":"2030-01-08T00:00:00Z"}
			]}]
		}`))
	}))
	defer quotaServer.Close()

	configDir := t.TempDir()
	writeTestToken(t, filepath.Join(configDir, oauthTokenFile), "access-token", "2030-01-01T00:00:00Z", "refresh-token")

	p := New()
	p.HTTPClient = quotaServer.Client()
	snaps := map[string]core.UsageSnapshot{
		"antigravity-1": {
			ProviderID: "antigravity",
			AccountID:  "antigravity-1",
			Status:     core.StatusUnknown,
			Metrics:    map[string]core.Metric{},
		},
		"antigravity-2": {
			ProviderID: "antigravity",
			AccountID:  "antigravity-2",
			Status:     core.StatusUnknown,
			Metrics:    map[string]core.Metric{},
		},
	}
	p.EnrichSnapshots(context.Background(), []core.AccountConfig{
		{
			ID:       "antigravity-1",
			Provider: "antigravity",
			ProviderPaths: map[string]string{
				"config_dir": configDir,
			},
			RuntimeHints: map[string]string{
				"quota_endpoint": quotaServer.URL + "/v1internal",
			},
		},
		{
			ID:       "antigravity-2",
			Provider: "antigravity",
			ProviderPaths: map[string]string{
				"config_dir": configDir,
			},
			RuntimeHints: map[string]string{
				"quota_endpoint": quotaServer.URL + "/v1internal",
			},
		},
	}, snaps)

	if snaps["antigravity-1"].Status != core.StatusOK {
		t.Fatalf("snap1 status = %q, want ok", snaps["antigravity-1"].Status)
	}
	if snaps["antigravity-2"].Status != core.StatusOK {
		t.Fatalf("snap2 status = %q, want ok", snaps["antigravity-2"].Status)
	}
}

func TestFetch_ConcurrencyUnderRace(t *testing.T) {
	quotaServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{
			"groups":[{"buckets":[
				{"bucketId":"gemini-5h","remainingFraction":0.6,"resetTime":"2030-01-01T00:00:00Z"},
				{"bucketId":"claude-5h","remainingFraction":0.7,"resetTime":"2030-01-01T00:00:00Z"}
			]}]
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
			snap, err := p.Fetch(context.Background(), core.AccountConfig{
				ID:       fmt.Sprintf("antigravity-%d", idx),
				Provider: "antigravity",
				ProviderPaths: map[string]string{
					"config_dir": configDir,
				},
				RuntimeHints: map[string]string{
					"quota_endpoint": quotaServer.URL + "/v1internal",
				},
			})
			if err != nil {
				t.Errorf("Fetch concurrent error = %v", err)
				return
			}
			if snap.Status != core.StatusOK {
				t.Errorf("snap status = %q, want ok", snap.Status)
			}
		}(i)
	}
	wg.Wait()
}

func writeTestToken(t *testing.T, path, access, expiry, refresh string) {
	t.Helper()
	payload := oauthTokenFilePayload{
		Token: oauthToken{
			AccessToken:  access,
			RefreshToken: refresh,
			TokenType:    "Bearer",
			Expiry:       expiry,
		},
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func metricRemaining(t *testing.T, snap core.UsageSnapshot, key string) float64 {
	t.Helper()
	metric, ok := snap.Metrics[key]
	if !ok || metric.Remaining == nil {
		t.Fatalf("metric %q missing remaining value: %+v", key, metric)
	}
	return *metric.Remaining
}
