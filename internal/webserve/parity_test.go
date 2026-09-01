package webserve

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/nurulislamz/agentusage/internal/config"
	"github.com/nurulislamz/agentusage/internal/core"
)

func TestTUIWebInformationParity_DemoHTTP(t *testing.T) {
	now := time.Date(2026, 8, 30, 18, 15, 0, 0, time.UTC)
	opts := Options{
		Demo:       true,
		Theme:      "Deep Space",
		TimeWindow: "3d",
		UsageMode:  config.UsageModeRemaining,
		Now:        func() time.Time { return now },
		ListenAddr: "127.0.0.1:0",
	}
	srv, err := NewServer(opts)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/snapshots", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d body %s", w.Code, w.Body.String())
	}
	var env Envelope
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}
	if len(env.Views) == 0 {
		t.Fatal("expected views from web port")
	}
	issues := VerifyTUIWebParity(opts, env)
	if len(issues) > 0 {
		t.Fatalf("tui/web information mismatch:\n%s", joinIssues(issues))
	}
}

func TestTUIWebInformationParity_OpenCode(t *testing.T) {
	now := time.Date(2026, 8, 30, 18, 15, 0, 0, time.UTC)
	zero, thirtyTwo, hundred := 0.0, 32.0, 100.0
	snap := core.UsageSnapshot{
		AccountID:  "opencode-mohammed",
		ProviderID: "opencode",
		Status:     core.StatusLimited,
		Timestamp:  now,
		Metrics: map[string]core.Metric{
			"rolling_usage":     {Remaining: &hundred},
			"weekly_usage":      {Remaining: &thirtyTwo},
			"monthly_usage_pct": {Remaining: &zero},
		},
		Resets: map[string]time.Time{
			"rolling_usage":     now.Add(4*time.Hour + 59*time.Minute),
			"weekly_usage":      now.Add(5*time.Hour + 45*time.Minute),
			"monthly_usage_pct": now.Add(9*24*time.Hour + 3*time.Hour),
		},
	}
	opts := Options{
		Theme:      "Deep Space",
		TimeWindow: "3d",
		UsageMode:  config.UsageModeRemaining,
		Now:        func() time.Time { return now },
		ListenAddr: "127.0.0.1:0",
		Collect: func() (Envelope, error) {
			return Envelope{Source: "direct", Snapshots: []core.UsageSnapshot{snap}}, nil
		},
	}
	srv, err := NewServer(opts)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/snapshots", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var env Envelope
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}
	issues := VerifyTUIWebParity(opts, env)
	if len(issues) > 0 {
		t.Fatalf("tui/web information mismatch:\n%s", joinIssues(issues))
	}
	if len(env.Views) != 1 {
		t.Fatalf("views = %d", len(env.Views))
	}
	gauges := 0
	for _, card := range env.Views[0].DetailCards {
		for _, row := range card.Rows {
			if row.Kind == "gauge" {
				gauges++
			}
		}
	}
	if gauges < 3 {
		t.Fatalf("expected 3 OpenCode gauges on the web port, got %d", gauges)
	}
}

func TestTUIWebInformationParity_CommandCode(t *testing.T) {
	now := time.Date(2026, 8, 30, 18, 15, 0, 0, time.UTC)
	snap := core.UsageSnapshot{
		AccountID:  "command-code",
		ProviderID: "command_code",
		Status:     core.StatusOK,
		Timestamp:  now,
		Attributes: map[string]string{
			"plan_name":         "GOAT",
			"plan_id":           "individual-goat",
			"monthly_cap":       "$70.00",
			"monthly_used":      "$35.84",
			"monthly_remaining": "$34.16",
			"weekly_cap":        "$35.00",
			"weekly_used":       "$7.00",
			"five_hour_cap":     "$14.00",
			"five_hour_used":    "$0.00",
		},
		Metrics: map[string]core.Metric{
			"monthly_subscription": {Limit: f64(100), Used: f64(51.2), Remaining: f64(48.8), Unit: "percent", Window: "month"},
			"monthly_credits":      {Limit: f64(70.0), Used: f64(35.84), Remaining: f64(34.16), Unit: "USD", Window: "month"},
			"weekly_usage":         {Used: f64(20.0), Remaining: f64(80.0), Unit: "percent", Window: "7d"},
			"five_hour_usage":      {Used: f64(0.0), Remaining: f64(100.0), Unit: "percent", Window: "5h"},
			"balance":              {Remaining: f64(34.16), Unit: "USD"},
			"total_cost":           {Used: f64(35.84), Unit: "USD", Window: "billing-period"},
			"total_tokens":         {Used: f64(1420000), Unit: "tokens", Window: "billing-period"},
		},
		Resets: map[string]time.Time{
			"monthly_subscription": now.Add(15 * 24 * time.Hour),
			"weekly_usage":         now.Add(3 * 24 * time.Hour),
			"five_hour_usage":      now.Add(4 * time.Hour),
		},
	}
	opts := Options{
		Theme:      "Deep Space",
		TimeWindow: "3d",
		UsageMode:  config.UsageModeRemaining,
		Now:        func() time.Time { return now },
		ListenAddr: "127.0.0.1:0",
		Collect: func() (Envelope, error) {
			return Envelope{Source: "direct", Snapshots: []core.UsageSnapshot{snap}}, nil
		},
	}
	srv, err := NewServer(opts)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/snapshots", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var env Envelope
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}
	issues := VerifyTUIWebParity(opts, env)
	if len(issues) > 0 {
		t.Fatalf("tui/web information mismatch:\n%s", joinIssues(issues))
	}
	if len(env.Views) != 1 {
		t.Fatalf("views = %d", len(env.Views))
	}
	gauges := 0
	for _, card := range env.Views[0].DetailCards {
		for _, row := range card.Rows {
			if row.Kind == "gauge" {
				gauges++
			}
		}
	}
	if gauges < 3 {
		t.Fatalf("expected >=3 Command Code gauges on the web port, got %d", gauges)
	}
}

func joinIssues(issues []ParityIssue) string {
	parts := make([]string, len(issues))
	for i, issue := range issues {
		parts[i] = "  " + issue.String()
	}
	return strings.Join(parts, "\n")
}
