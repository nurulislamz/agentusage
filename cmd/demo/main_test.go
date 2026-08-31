package main

import (
	"context"
	"flag"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nurulislamz/agentusage/internal/config"
	"github.com/nurulislamz/agentusage/internal/core"
	"github.com/nurulislamz/agentusage/internal/providers"
	"github.com/nurulislamz/agentusage/internal/tui"
)

func TestBuildDemoSnapshots_IncludesAllDemoProviders(t *testing.T) {
	snaps := buildDemoSnapshots()
	if len(snaps) == 0 {
		t.Fatal("buildDemoSnapshots returned no snapshots")
	}

	byProvider := make(map[string]string)
	for accountID, snap := range snaps {
		if snap.AccountID == "" {
			t.Fatalf("snapshot for key %q has empty account id", accountID)
		}
		if accountID != snap.AccountID {
			t.Fatalf("snapshot key/account mismatch: key=%q account=%q", accountID, snap.AccountID)
		}
		if snap.ProviderID == "" {
			t.Fatalf("snapshot %q has empty provider id", accountID)
		}
		if snap.Status == "" {
			t.Fatalf("snapshot %q has empty status", accountID)
		}
		if snap.Metrics == nil {
			t.Fatalf("snapshot %q has nil metrics map", accountID)
		}
		if existing, ok := byProvider[snap.ProviderID]; ok {
			t.Fatalf("provider %q appears multiple times (%q, %q)", snap.ProviderID, existing, accountID)
		}
		byProvider[snap.ProviderID] = accountID
	}

	for providerID := range demoProviderIDs {
		if _, ok := byProvider[providerID]; !ok {
			t.Fatalf("missing demo snapshot for provider %q", providerID)
		}
	}

	if len(snaps) != len(demoProviderIDs) {
		t.Fatalf("expected %d snapshots, got %d", len(demoProviderIDs), len(snaps))
	}
}

func TestBuildDemoSnapshots_WidgetCoverage(t *testing.T) {
	snaps := buildDemoSnapshots()

	type expectation struct {
		hasModelBurnData bool
		hasClientMixData bool
	}

	want := map[string]expectation{
		"claude_code": {hasModelBurnData: true, hasClientMixData: true},
		"codex":       {hasModelBurnData: true, hasClientMixData: true},
		"copilot":     {hasModelBurnData: true, hasClientMixData: true},
		"gemini_cli":  {hasModelBurnData: true, hasClientMixData: true},
		"openrouter":  {hasModelBurnData: true, hasClientMixData: true},
	}

	for providerID, exp := range want {
		snap, ok := snapshotByProvider(snaps, providerID)
		if !ok {
			t.Fatalf("missing snapshot for provider %q", providerID)
		}
		if exp.hasModelBurnData && !hasModelBurnMetrics(snap) {
			t.Fatalf("provider %q missing model burn metrics", providerID)
		}
		if exp.hasClientMixData && !hasClientMixMetrics(snap) {
			t.Fatalf("provider %q missing client mix metrics", providerID)
		}
	}
}

func TestBuildDemoAccounts_IncludesAllDemoProviders(t *testing.T) {
	accounts := buildDemoAccounts()
	if len(accounts) == 0 {
		t.Fatal("buildDemoAccounts returned no accounts")
	}

	byProvider := make(map[string]core.AccountConfig, len(accounts))
	for _, account := range accounts {
		if account.ID == "" {
			t.Fatalf("account for provider %q has empty ID", account.Provider)
		}
		if account.Provider == "" {
			t.Fatalf("account %q has empty provider ID", account.ID)
		}
		if _, ok := byProvider[account.Provider]; ok {
			t.Fatalf("duplicate account for provider %q", account.Provider)
		}
		byProvider[account.Provider] = account
	}

	for providerID := range demoProviderIDs {
		if _, ok := byProvider[providerID]; !ok {
			t.Fatalf("missing account for provider %q", providerID)
		}
	}

	if len(accounts) != len(demoProviderIDs) {
		t.Fatalf("expected %d accounts, got %d", len(demoProviderIDs), len(accounts))
	}
}

func TestBuildDemoProviders_FetchesMockedSnapshots(t *testing.T) {
	scenario := newDemoScenario(time.Date(2026, 4, 16, 10, 0, 0, 0, time.UTC), defaultDemoConfig())
	wrapped := buildDemoProviders(providers.AllProviders(), scenario)
	if len(wrapped) == 0 {
		t.Fatal("buildDemoProviders returned no providers")
	}

	byProvider := make(map[string]core.UsageProvider, len(wrapped))
	for _, provider := range wrapped {
		byProvider[provider.ID()] = provider
	}

	for _, account := range buildDemoAccounts() {
		provider, ok := byProvider[account.Provider]
		if !ok {
			t.Fatalf("missing wrapped provider %q", account.Provider)
		}

		snap, err := provider.Fetch(context.Background(), account)
		if err != nil {
			t.Fatalf("fetch for provider %q failed: %v", account.Provider, err)
		}
		if snap.ProviderID != account.Provider {
			t.Fatalf("provider mismatch for account %q: got %q want %q", account.ID, snap.ProviderID, account.Provider)
		}
		if snap.AccountID != account.ID {
			t.Fatalf("account mismatch for provider %q: got %q want %q", account.Provider, snap.AccountID, account.ID)
		}
		if snap.Status == "" {
			t.Fatalf("empty status for provider %q", account.Provider)
		}
		if snap.Metrics == nil {
			t.Fatalf("nil metrics for provider %q", account.Provider)
		}
	}
}

func TestBuildDemoSnapshotsForPhase_ProgressesDeterministically(t *testing.T) {
	anchor := time.Date(2026, 4, 16, 10, 0, 0, 0, time.UTC)
	early := buildDemoSnapshotsForPhase(anchor, 0)
	mid := buildDemoSnapshotsForPhase(anchor, 3)
	late := buildDemoSnapshotsForPhase(anchor, len(demoPhaseShares)-1)

	checks := []struct {
		providerID string
		metricKey  string
	}{
		{providerID: "claude_code", metricKey: "5h_block_cost"},
		{providerID: "gemini_cli", metricKey: "quota"},
		{providerID: "openrouter", metricKey: "usage_monthly"},
	}

	for _, tc := range checks {
		earlySnap, ok := snapshotByProvider(early, tc.providerID)
		if !ok {
			t.Fatalf("missing early snapshot for provider %q", tc.providerID)
		}
		midSnap, ok := snapshotByProvider(mid, tc.providerID)
		if !ok {
			t.Fatalf("missing mid snapshot for provider %q", tc.providerID)
		}
		lateSnap, ok := snapshotByProvider(late, tc.providerID)
		if !ok {
			t.Fatalf("missing late snapshot for provider %q", tc.providerID)
		}

		earlyValue, ok := metricUsed(earlySnap.Metrics, tc.metricKey)
		if !ok {
			t.Fatalf("provider %q missing early metric %q", tc.providerID, tc.metricKey)
		}
		midValue, ok := metricUsed(midSnap.Metrics, tc.metricKey)
		if !ok {
			t.Fatalf("provider %q missing mid metric %q", tc.providerID, tc.metricKey)
		}
		lateValue, ok := metricUsed(lateSnap.Metrics, tc.metricKey)
		if !ok {
			t.Fatalf("provider %q missing late metric %q", tc.providerID, tc.metricKey)
		}

		if !(earlyValue < midValue && midValue < lateValue) {
			t.Fatalf("provider %q metric %q is not monotonic across phases: early=%.2f mid=%.2f late=%.2f", tc.providerID, tc.metricKey, earlyValue, midValue, lateValue)
		}
	}

	earlyOpenRouter, _ := snapshotByProvider(early, "openrouter")
	lateOpenRouter, _ := snapshotByProvider(late, "openrouter")
	earlyLast := earlyOpenRouter.DailySeries["analytics_tokens"][len(earlyOpenRouter.DailySeries["analytics_tokens"])-1].Value
	lateLast := lateOpenRouter.DailySeries["analytics_tokens"][len(lateOpenRouter.DailySeries["analytics_tokens"])-1].Value
	if earlyLast >= lateLast {
		t.Fatalf("expected latest demo series point to grow across phases: early=%.2f late=%.2f", earlyLast, lateLast)
	}
}

func TestDemoScenario_StopsAtFinalFrame(t *testing.T) {
	scenario := newDemoScenario(time.Date(2026, 4, 16, 10, 0, 0, 0, time.UTC), defaultDemoConfig())
	last := len(demoPhaseShares) - 1

	for range len(demoPhaseShares) + 3 {
		scenario.Advance()
	}

	if scenario.CurrentPhase() != last {
		t.Fatalf("expected scenario to stop at phase %d, got %d", last, scenario.CurrentPhase())
	}

	account := core.AccountConfig{ID: "codex-cli", Provider: "codex"}
	snap, ok := scenario.Snapshot(account.ID, account.Provider)
	if !ok {
		t.Fatal("missing codex snapshot at final phase")
	}

	extraAdvanced := scenario.Advance()
	if extraAdvanced {
		t.Fatal("expected scenario advance to stop once the final frame is reached")
	}

	nextSnap, ok := scenario.Snapshot(account.ID, account.Provider)
	if !ok {
		t.Fatal("missing codex snapshot after extra advance")
	}

	if snap.Timestamp != nextSnap.Timestamp {
		t.Fatalf("final frame changed after extra advance: %s != %s", snap.Timestamp, nextSnap.Timestamp)
	}
}

func TestDemoScenario_LoopsWhenEnabled(t *testing.T) {
	cfg := defaultDemoConfig()
	cfg.interval = 2 * time.Second
	cfg.loop = true
	scenario := newDemoScenario(time.Date(2026, 4, 16, 10, 0, 0, 0, time.UTC), cfg)
	account := core.AccountConfig{ID: "codex-cli", Provider: "codex"}

	for range len(demoPhaseShares) - 1 {
		if !scenario.Advance() {
			t.Fatal("expected advance through pre-loop frames to succeed")
		}
	}

	lastSnap, ok := scenario.Snapshot(account.ID, account.Provider)
	if !ok {
		t.Fatal("missing final-frame codex snapshot")
	}

	if !scenario.Advance() {
		t.Fatal("expected loop-enabled scenario to wrap")
	}

	if scenario.CurrentPhase() != 0 {
		t.Fatalf("expected loop-enabled scenario to wrap to phase 0, got %d", scenario.CurrentPhase())
	}

	loopedSnap, ok := scenario.Snapshot(account.ID, account.Provider)
	if !ok {
		t.Fatal("missing wrapped codex snapshot")
	}

	if !loopedSnap.Timestamp.After(lastSnap.Timestamp) {
		t.Fatalf("expected wrapped snapshot timestamp to move forward: %s <= %s", loopedSnap.Timestamp, lastSnap.Timestamp)
	}
}

func TestParseDemoConfig(t *testing.T) {
	cfg, err := parseDemoConfig([]string{"-interval", "750ms", "-loop"})
	if err != nil {
		t.Fatalf("parseDemoConfig returned error: %v", err)
	}
	if cfg.interval != 750*time.Millisecond {
		t.Fatalf("unexpected interval: got %s want %s", cfg.interval, 750*time.Millisecond)
	}
	if !cfg.loop {
		t.Fatal("expected loop flag to be true")
	}
}

func TestParseDemoConfig_RejectsZeroInterval(t *testing.T) {
	if _, err := parseDemoConfig([]string{"-interval", "0s"}); err == nil {
		t.Fatal("expected zero interval to be rejected")
	}
}

func TestBuildDemoSnapshots_RichProviderDetails(t *testing.T) {
	snaps := buildDemoSnapshots()

	type providerExpect struct {
		metrics []string
		raw     []string
		resets  []string
		series  []string
	}

	expectations := map[string]providerExpect{
		"gemini_cli": {
			metrics: []string{
				"quota",
				"quota_model_gemini_2_5_pro_requests",
				"tool_calls_success",
				"tool_calls_total",
				"tool_success_rate",
				"composer_lines_added",
				"composer_files_changed",
				"lang_go",
			},
			raw: []string{
				"language_usage",
			},
			resets: []string{
				"quota_model_gemini_2_5_pro_requests_reset",
			},
			series: []string{
				"analytics_tokens",
			},
		},
		"cursor": {
			metrics: []string{
				"interface_composer",
				"composer_accepted_lines",
				"tool_calls_total",
			},
			raw: []string{
				"billing_cycle_start",
				"billing_cycle_end",
			},
			resets: []string{
				"billing_cycle_end",
			},
			series: []string{
				"usage_model_claude-4.6-opus-high-thinking",
			},
		},
		"claude_code": {
			metrics: []string{
				"tool_bash",
				"client_api_server_total_tokens",
				"project_platform_core_requests",
				"lang_go",
				"composer_lines_added",
				"total_prompts",
			},
			raw: []string{
				"block_start",
				"block_end",
				"language_usage",
				"project_usage",
			},
			series: []string{
				"analytics_tokens",
				"tokens_client_api_server",
				"usage_model_synthetic",
				"usage_project_platform_core",
			},
		},
		"codex": {
			metrics: []string{
				"model_gpt_5_4_input_tokens",
				"client_cli_total_tokens",
				"project_dashboard_shell_requests",
			},
			raw: []string{
				"project_usage",
			},
			series: []string{
				"analytics_tokens",
				"tokens_client_cli",
				"usage_project_dashboard_shell",
			},
		},
		"openrouter": {
			// OpenRouter (an API router) has model, client (app), and provider
			// breakdowns, but no per-tool or per-language telemetry — so no
			// lang_*/tool_* metrics here.
			metrics: []string{
				"analytics_7d_tokens",
				"model_qwen_qwen3-coder-flash_cost_usd",
				"client_recipe_blog_total_tokens",
				"provider_alibaba_cost_usd",
			},
			raw: []string{
				"client_usage",
			},
			series: []string{
				"analytics_tokens",
				"tokens_client_recipe_blog",
			},
		},
		"copilot": {
			metrics: []string{
				"gh_core_rpm",
				"gh_graphql_rpm",
				"model_claude_haiku_4_5_input_tokens",
				"client_vscode_total_tokens",
				"tool_calls_total",
				"tool_success_rate",
				"composer_lines_added",
				"composer_files_changed",
				"lang_go",
			},
			raw: []string{
				"language_usage",
			},
			resets: []string{
				"gh_core_rpm_reset",
			},
			series: []string{
				"tokens_client_vscode",
			},
		},
	}

	for providerID, exp := range expectations {
		snap, ok := snapshotByProvider(snaps, providerID)
		if !ok {
			t.Fatalf("missing snapshot for provider %q", providerID)
		}

		for _, key := range exp.metrics {
			if _, ok := snap.Metrics[key]; !ok {
				t.Fatalf("provider %q missing metric %q", providerID, key)
			}
		}
		for _, key := range exp.raw {
			if _, ok := snap.Raw[key]; !ok {
				t.Fatalf("provider %q missing raw %q", providerID, key)
			}
		}
		for _, key := range exp.resets {
			if _, ok := snap.Resets[key]; !ok {
				t.Fatalf("provider %q missing reset %q", providerID, key)
			}
		}
		for _, key := range exp.series {
			if _, ok := snap.DailySeries[key]; !ok {
				t.Fatalf("provider %q missing daily series %q", providerID, key)
			}
		}
	}
}

func TestBuildDemoSnapshots_UsesNonLinearDailyPatterns(t *testing.T) {
	snaps := buildDemoSnapshots()

	cases := []struct {
		providerID  string
		key         string
		minPoints   int
		minSpanDays int
	}{
		{providerID: "claude_code", key: "analytics_requests", minPoints: 10, minSpanDays: 14},
		{providerID: "codex", key: "analytics_requests", minPoints: 3, minSpanDays: 6},
		{providerID: "cursor", key: "analytics_tokens", minPoints: 5, minSpanDays: 7},
		{providerID: "openrouter", key: "analytics_tokens", minPoints: 8, minSpanDays: 16},
	}

	for _, tc := range cases {
		snap, ok := snapshotByProvider(snaps, tc.providerID)
		if !ok {
			t.Fatalf("missing snapshot for provider %q", tc.providerID)
		}
		pts := snap.DailySeries[tc.key]
		if len(pts) < tc.minPoints {
			t.Fatalf("provider %q series %q too short: got %d want >= %d", tc.providerID, tc.key, len(pts), tc.minPoints)
		}
		if span := seriesSpanDays(t, pts); span < tc.minSpanDays {
			t.Fatalf("provider %q series %q spans only %d days; want >= %d", tc.providerID, tc.key, span, tc.minSpanDays)
		}
		if isStrictlyIncreasing(pts) {
			t.Fatalf("provider %q series %q is still a straight ramp", tc.providerID, tc.key)
		}
	}
}

func snapshotByProvider(snaps map[string]core.UsageSnapshot, providerID string) (core.UsageSnapshot, bool) {
	for _, snap := range snaps {
		if snap.ProviderID == providerID {
			return snap, true
		}
	}
	return core.UsageSnapshot{}, false
}

func hasModelBurnMetrics(snap core.UsageSnapshot) bool {
	for key, m := range snap.Metrics {
		if m.Used == nil {
			continue
		}
		if strings.HasPrefix(key, "model_") && (strings.HasSuffix(key, "_cost_usd") || strings.HasSuffix(key, "_cost")) {
			return true
		}
		if strings.HasPrefix(key, "model_") && (strings.HasSuffix(key, "_input_tokens") || strings.HasSuffix(key, "_output_tokens")) {
			return true
		}
	}
	return false
}

func hasClientMixMetrics(snap core.UsageSnapshot) bool {
	for key, m := range snap.Metrics {
		if m.Used == nil {
			continue
		}
		if strings.HasPrefix(key, "client_") && strings.HasSuffix(key, "_total_tokens") {
			return true
		}
	}
	return false
}

func seriesSpanDays(t *testing.T, pts []core.TimePoint) int {
	t.Helper()
	if len(pts) < 2 {
		return 0
	}
	first, err := time.Parse("2006-01-02", pts[0].Date)
	if err != nil {
		t.Fatalf("parse first date %q: %v", pts[0].Date, err)
	}
	last, err := time.Parse("2006-01-02", pts[len(pts)-1].Date)
	if err != nil {
		t.Fatalf("parse last date %q: %v", pts[len(pts)-1].Date, err)
	}
	return int(last.Sub(first).Hours() / 24)
}

func isStrictlyIncreasing(pts []core.TimePoint) bool {
	if len(pts) < 2 {
		return false
	}
	for i := 1; i < len(pts); i++ {
		if pts[i].Value <= pts[i-1].Value {
			return false
		}
	}
	return true
}

func TestIoDiscard_Write(t *testing.T) {
	d := ioDiscard{}
	n, err := d.Write([]byte("hello world"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 11 {
		t.Fatalf("expected 11 bytes written, got %d", n)
	}
}

func TestParseDemoConfig_Variants(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantInterval time.Duration
		wantLoop     bool
		wantErr      bool
		errCheck     func(error) bool
	}{
		{
			name:         "default args",
			args:         []string{},
			wantInterval: defaultDemoRefreshInterval,
			wantLoop:     false,
		},
		{
			name:         "custom interval and loop",
			args:         []string{"-interval", "2s", "-loop"},
			wantInterval: 2 * time.Second,
			wantLoop:     true,
		},
		{
			name:     "help flag",
			args:     []string{"-help"},
			wantErr:  true,
			errCheck: func(err error) bool { return err == flag.ErrHelp },
		},
		{
			name:    "invalid flag",
			args:    []string{"-nonexistent"},
			wantErr: true,
		},
		{
			name:    "negative interval",
			args:    []string{"-interval", "-1s"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := parseDemoConfig(tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseDemoConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if tt.errCheck != nil && !tt.errCheck(err) {
					t.Fatalf("parseDemoConfig() error %v failed errCheck", err)
				}
				return
			}
			if cfg.interval != tt.wantInterval {
				t.Errorf("cfg.interval = %v, want %v", cfg.interval, tt.wantInterval)
			}
			if cfg.loop != tt.wantLoop {
				t.Errorf("cfg.loop = %v, want %v", cfg.loop, tt.wantLoop)
			}
		})
	}
}

func TestDemoHelpers_PtrAndDemoSeries(t *testing.T) {
	v := 42.5
	p := ptr(v)
	if p == nil || *p != v {
		t.Fatalf("ptr(%v) = %v, want %v", v, p, &v)
	}

	now := time.Date(2026, 4, 16, 12, 0, 0, 0, time.UTC)
	if s := demoSeries(now); s != nil {
		t.Fatalf("expected nil for empty demoSeries, got %v", s)
	}

	pts := demoSeries(now, 10.0, 20.0, 30.0)
	if len(pts) != 3 {
		t.Fatalf("expected 3 points, got %d", len(pts))
	}
	if pts[0].Date != "2026-04-14" || pts[0].Value != 10.0 {
		t.Errorf("unexpected first point: %+v", pts[0])
	}
	if pts[2].Date != "2026-04-16" || pts[2].Value != 30.0 {
		t.Errorf("unexpected last point: %+v", pts[2])
	}
}

func TestDemoHelpers_DemoPatternSeries(t *testing.T) {
	now := time.Date(2026, 4, 16, 12, 0, 0, 0, time.UTC)
	if s := demoPatternSeries(now, 100); s != nil {
		t.Fatalf("expected nil for empty pattern, got %v", s)
	}

	// Pattern with zero and negative weights (should be ignored)
	ptsIgnored := demoPatternSeries(now, 100, demoPoint(5, 0.0), demoPoint(3, -0.5))
	if len(ptsIgnored) != 0 {
		t.Fatalf("expected 0 points when weights are non-positive, got %d", len(ptsIgnored))
	}

	pts := demoPatternSeries(now, 100, demoPoint(3, 0.5), demoPoint(1, 0.8), demoPoint(5, 0.2))
	if len(pts) != 3 {
		t.Fatalf("expected 3 points, got %d", len(pts))
	}
	// Verify sorting by date ascending (DaysAgo 5 -> DaysAgo 3 -> DaysAgo 1)
	if pts[0].Date != "2026-04-11" || pts[0].Value != 20.0 {
		t.Errorf("point 0 mismatch: %+v", pts[0])
	}
	if pts[1].Date != "2026-04-13" || pts[1].Value != 50.0 {
		t.Errorf("point 1 mismatch: %+v", pts[1])
	}
	if pts[2].Date != "2026-04-15" || pts[2].Value != 80.0 {
		t.Errorf("point 2 mismatch: %+v", pts[2])
	}
}

func TestDemoHelpers_RoundDemoSeriesValue(t *testing.T) {
	tests := []struct {
		input float64
		want  float64
	}{
		{input: 1234.56, want: 1235.0},
		{input: 1000.0, want: 1000.0},
		{input: 123.456, want: 123.5},
		{input: 100.0, want: 100.0},
		{input: 12.3456, want: 12.35},
		{input: 0.123, want: 0.12},
		{input: 0.0, want: 0.0},
	}
	for _, tt := range tests {
		got := roundDemoSeriesValue(tt.input)
		if got != tt.want {
			t.Errorf("roundDemoSeriesValue(%v) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestDemoHelpers_DemoMessageForSnapshot(t *testing.T) {
	// 1. OpenRouter with credit_balance
	snapOR := core.UsageSnapshot{
		ProviderID: "openrouter",
		Metrics: map[string]core.Metric{
			"credit_balance": {Remaining: ptr(42.50)},
		},
	}
	if msg := demoMessageForSnapshot(snapOR); msg != "$42.50 credits remaining" {
		t.Errorf("unexpected openrouter msg: %q", msg)
	}

	// 2. OpenRouter without credit_balance (fallback)
	snapORFallback := core.UsageSnapshot{
		ProviderID: "openrouter",
		Metrics:    map[string]core.Metric{},
		Message:    "Fallback message",
	}
	if msg := demoMessageForSnapshot(snapORFallback); msg != "Fallback message" {
		t.Errorf("unexpected openrouter fallback msg: %q", msg)
	}

	// 3. Cursor with spend_limit and plan_spend
	snapCursor := core.UsageSnapshot{
		ProviderID: "cursor",
		Metrics: map[string]core.Metric{
			"plan_spend":  {Used: ptr(75.50)},
			"spend_limit": {Limit: ptr(200.0), Remaining: ptr(124.50)},
		},
	}
	if msg := demoMessageForSnapshot(snapCursor); msg != "Team — $75.50 / $200 team spend ($124.50 remaining)" {
		t.Errorf("unexpected cursor msg: %q", msg)
	}

	// 4. Cursor missing metrics (fallback)
	snapCursorFallback := core.UsageSnapshot{
		ProviderID: "cursor",
		Metrics: map[string]core.Metric{
			"plan_spend": {Used: ptr(75.50)},
		},
		Message: "Cursor fallback",
	}
	if msg := demoMessageForSnapshot(snapCursorFallback); msg != "Cursor fallback" {
		t.Errorf("unexpected cursor fallback msg: %q", msg)
	}

	// 5. Other provider
	snapOther := core.UsageSnapshot{
		ProviderID: "codex",
		Message:    "Codex ok",
	}
	if msg := demoMessageForSnapshot(snapOther); msg != "Codex ok" {
		t.Errorf("unexpected other msg: %q", msg)
	}
}

func TestDemoHelpers_MetricExtractors(t *testing.T) {
	metrics := map[string]core.Metric{
		"full":     {Used: ptr(10), Limit: ptr(100), Remaining: ptr(90)},
		"nil_ptrs": {},
	}

	// metricUsed
	if val, ok := metricUsed(metrics, "full"); !ok || val != 10 {
		t.Errorf("metricUsed(full) = (%v, %v), want (10, true)", val, ok)
	}
	if _, ok := metricUsed(metrics, "nil_ptrs"); ok {
		t.Errorf("metricUsed(nil_ptrs) want false, got true")
	}
	if _, ok := metricUsed(metrics, "missing"); ok {
		t.Errorf("metricUsed(missing) want false, got true")
	}

	// metricLimit
	if val, ok := metricLimit(metrics, "full"); !ok || val != 100 {
		t.Errorf("metricLimit(full) = (%v, %v), want (100, true)", val, ok)
	}
	if _, ok := metricLimit(metrics, "nil_ptrs"); ok {
		t.Errorf("metricLimit(nil_ptrs) want false, got true")
	}
	if _, ok := metricLimit(metrics, "missing"); ok {
		t.Errorf("metricLimit(missing) want false, got true")
	}

	// metricRemaining
	if val, ok := metricRemaining(metrics, "full"); !ok || val != 90 {
		t.Errorf("metricRemaining(full) = (%v, %v), want (90, true)", val, ok)
	}
	if _, ok := metricRemaining(metrics, "nil_ptrs"); ok {
		t.Errorf("metricRemaining(nil_ptrs) want false, got true")
	}
	if _, ok := metricRemaining(metrics, "missing"); ok {
		t.Errorf("metricRemaining(missing) want false, got true")
	}
}

func TestDemoHelpers_RoundLike(t *testing.T) {
	// Rounding like integer
	if got := roundLike(10.0, 5.43); got != 5.0 {
		t.Errorf("roundLike(10.0, 5.43) = %v, want 5.0", got)
	}
	if got := roundLike(10.0, 5.67); got != 6.0 {
		t.Errorf("roundLike(10.0, 5.67) = %v, want 6.0", got)
	}

	// Rounding like float
	if got := roundLike(10.25, 5.436); got != 5.44 {
		t.Errorf("roundLike(10.25, 5.436) = %v, want 5.44", got)
	}
}

func TestDemoAccountID_AllCases(t *testing.T) {
	tests := []struct {
		providerID string
		want       string
	}{
		{"claude_code", "claude-code"},
		{"codex", "codex-cli"},
		{"cursor", "cursor-ide"},
		{"gemini_cli", "gemini-cli"},
		{"openrouter", "openrouter"},
		{"copilot", "copilot"},
		{"ollama", "ollama"},
		{"custom_provider", "custom_provider"},
	}
	for _, tt := range tests {
		if got := demoAccountID(tt.providerID); got != tt.want {
			t.Errorf("demoAccountID(%q) = %q, want %q", tt.providerID, got, tt.want)
		}
	}
}

func TestDemoProvider_DelegationAndFetchBranches(t *testing.T) {
	allReal := providers.AllProviders()
	scenario := newDemoScenario(time.Now(), defaultDemoConfig())
	demoProvs := buildDemoProviders(allReal, scenario)

	for _, dp := range demoProvs {
		// Test delegation methods
		_ = dp.ID()
		_ = dp.Describe()
		_ = dp.Spec()
		_ = dp.DashboardWidget()
		_ = dp.DetailWidget()
	}

	// 1. Fetch with scenario
	var p0 core.UsageProvider
	for _, p := range demoProvs {
		if p.ID() == "gemini_cli" {
			p0 = p
			break
		}
	}
	if p0 == nil {
		t.Fatal("gemini_cli provider not found in demoProvs")
	}
	acct0 := core.AccountConfig{ID: "gemini-cli", Provider: "gemini_cli"}
	snap, err := p0.Fetch(context.Background(), acct0)
	if err != nil {
		t.Fatalf("fetch with scenario error: %v", err)
	}
	if snap.ProviderID != "gemini_cli" || snap.AccountID != "gemini-cli" {
		t.Errorf("unexpected snapshot account/provider: %+v", snap)
	}

	// 2. Fetch without scenario (scenario is nil), matching snaps[acct.ID]
	pNoScenario := buildDemoProviders(allReal, nil)
	pGemini := pNoScenario[0]
	for _, p := range pNoScenario {
		if p.ID() == "gemini_cli" {
			pGemini = p
			break
		}
	}
	snap2, err := pGemini.Fetch(context.Background(), acct0)
	if err != nil {
		t.Fatalf("fetch without scenario error: %v", err)
	}
	if snap2.ProviderID != "gemini_cli" {
		t.Errorf("unexpected provider in snap2: %s", snap2.ProviderID)
	}

	// 3. Fetch without scenario, account ID not matching map key but matching provider
	snap3, err := pGemini.Fetch(context.Background(), core.AccountConfig{ID: "custom-gemini-acct", Provider: "gemini_cli"})
	if err != nil {
		t.Fatalf("fetch fallback error: %v", err)
	}
	if snap3.AccountID != "custom-gemini-acct" || snap3.ProviderID != "gemini_cli" {
		t.Errorf("unexpected snap3: %+v", snap3)
	}

	// 4. Fetch without scenario, unknown provider
	pUnknown := &demoProvider{
		base:     &mockUnknownProvider{id: "unknown_prov"},
		scenario: nil,
	}
	snap4, err := pUnknown.Fetch(context.Background(), core.AccountConfig{ID: "unknown-acct", Provider: "unknown_prov"})
	if err != nil {
		t.Fatalf("fetch unknown provider error: %v", err)
	}
	if snap4.ProviderID != "unknown_prov" || snap4.AccountID != "unknown-acct" || snap4.Status != core.StatusOK {
		t.Errorf("unexpected snap4: %+v", snap4)
	}
}

type mockUnknownProvider struct {
	id string
}

func (m *mockUnknownProvider) ID() string                            { return m.id }
func (m *mockUnknownProvider) Describe() core.ProviderInfo           { return core.ProviderInfo{} }
func (m *mockUnknownProvider) Spec() core.ProviderSpec               { return core.ProviderSpec{} }
func (m *mockUnknownProvider) DashboardWidget() core.DashboardWidget { return core.DashboardWidget{} }
func (m *mockUnknownProvider) DetailWidget() core.DetailWidget       { return core.DetailWidget{} }
func (m *mockUnknownProvider) Fetch(context.Context, core.AccountConfig) (core.UsageSnapshot, error) {
	return core.UsageSnapshot{}, nil
}

func TestDemoScenario_EdgeCases(t *testing.T) {
	// Zero start time and non-positive interval
	s := newDemoScenario(time.Time{}, demoConfig{interval: -5 * time.Second})
	if s.interval != defaultDemoRefreshInterval {
		t.Errorf("expected default interval, got %v", s.interval)
	}

	// Empty scenario frames snapshot lookup
	emptyScenario := &demoScenario{}
	if _, ok := emptyScenario.Snapshot("any", "any"); ok {
		t.Errorf("expected ok=false on empty scenario")
	}

	// clampDemoPhase edge cases
	if c := clampDemoPhase(-10); c != 0 {
		t.Errorf("clampDemoPhase(-10) = %d, want 0", c)
	}
	if c := clampDemoPhase(100); c != len(demoPhaseShares)-1 {
		t.Errorf("clampDemoPhase(100) = %d, want %d", c, len(demoPhaseShares)-1)
	}
	if c := clampDemoPhase(2); c != 2 {
		t.Errorf("clampDemoPhase(2) = %d, want 2", c)
	}

	// Snapshot lookup fallback by provider ID
	snap, ok := s.Snapshot("non-standard-account-id", "claude_code")
	if !ok {
		t.Fatal("expected to find snapshot by providerID fallback")
	}
	if snap.ProviderID != "claude_code" {
		t.Errorf("expected provider claude_code, got %s", snap.ProviderID)
	}

	// Snapshot not found
	if _, ok := s.Snapshot("unknown", "unknown"); ok {
		t.Errorf("expected not found for unknown provider")
	}
}

func TestDemoScenario_ScaleMetricAndConstants(t *testing.T) {
	// 1. shouldKeepDemoMetricConstant keys
	constantKeys := []string{
		"context_window", "composer_context_pct", "quota_models_tracked",
		"quota_models_low", "quota_models_exhausted", "mcp_servers_active",
		"custom_servers_active", "any_active_key",
	}
	for _, k := range constantKeys {
		m := core.Metric{Used: ptr(100.0), Unit: "servers"}
		scaled := scaleDemoMetric(k, m, 0.5)
		if *scaled.Used != 100.0 {
			t.Errorf("key %q should remain constant at 100.0, got %v", k, *scaled.Used)
		}
	}

	// 2. Metric with Limit and Remaining, Used is nil
	mLimitRemaining := core.Metric{
		Limit:     ptr(100.0),
		Remaining: ptr(80.0),
	}
	scaledLR := scaleDemoMetric("quota", mLimitRemaining, 0.5)
	// used = 20, share = 0.5 -> scaledUsed = 10, remaining = 100 - 10 = 90
	if scaledLR.Remaining == nil || *scaledLR.Remaining != 90.0 {
		t.Errorf("scaledLR remaining = %v, want 90.0", scaledLR.Remaining)
	}

	// 3. Metric with Limit, Used and Remaining where scaledUsed exceeds Limit
	mClampRemaining := core.Metric{
		Limit:     ptr(10.0),
		Used:      ptr(20.0),
		Remaining: ptr(0.0),
	}
	scaledClamp := scaleDemoMetric("test_metric", mClampRemaining, 1.0)
	if scaledClamp.Remaining == nil || *scaledClamp.Remaining != 0.0 {
		t.Errorf("scaledClamp remaining = %v, want 0.0", scaledClamp.Remaining)
	}

	// 4. Metric with Remaining only (Limit nil)
	mRemOnly := core.Metric{
		Remaining: ptr(50.0),
	}
	scaledRemOnly := scaleDemoMetric("test_rem", mRemOnly, 0.5)
	if scaledRemOnly.Remaining == nil || *scaledRemOnly.Remaining != 50.0 {
		t.Errorf("scaledRemOnly remaining = %v, want 50.0", scaledRemOnly.Remaining)
	}

	// 5. Metric with Remaining and Limit (Limit > 0, Used = Limit - Remaining)
	mRemLimit := core.Metric{
		Limit:     ptr(100.0),
		Remaining: ptr(40.0),
	}
	scaledRemLimit := scaleDemoMetric("test_rem_limit", mRemLimit, 0.5)
	// used = 60 * 0.5 = 30, remaining = 100 - 30 = 70
	if scaledRemLimit.Remaining == nil || *scaledRemLimit.Remaining != 70.0 {
		t.Errorf("scaledRemLimit remaining = %v, want 70.0", scaledRemLimit.Remaining)
	}
}

func TestDemoScenario_ScaleHelpers(t *testing.T) {
	// scaleDemoValue with final == 0
	if v := scaleDemoValue(10.0, 0.0, 0.5); v != 0 {
		t.Errorf("scaleDemoValue(10, 0, 0.5) = %v, want 0", v)
	}

	// scaleDemoRemaining with Remaining == nil
	if v := scaleDemoRemaining(core.Metric{}, 0.5); v != 0 {
		t.Errorf("scaleDemoRemaining(nil) = %v, want 0", v)
	}

	// scaleDemoRemaining with Limit remaining clamp below 0
	mNeg := core.Metric{
		Limit:     ptr(10.0),
		Remaining: ptr(-5.0),
	}
	if v := scaleDemoRemaining(mNeg, 2.0); v != 0 {
		t.Errorf("scaleDemoRemaining(neg) = %v, want 0", v)
	}

	// scaleDemoFloatPtr
	if v := scaleDemoFloatPtr(nil, 0.5); v != nil {
		t.Errorf("scaleDemoFloatPtr(nil) = %v, want nil", v)
	}
	if v := scaleDemoFloatPtr(ptr(20.0), 0.5); v == nil || *v != 10.0 {
		t.Errorf("scaleDemoFloatPtr(20, 0.5) = %v, want 10.0", v)
	}

	// scaleDemoSeries
	if s := scaleDemoSeries(nil, 0.5); s != nil {
		t.Errorf("scaleDemoSeries(nil) = %v, want nil", s)
	}
	if s := scaleDemoSeries([]core.TimePoint{}, 0.5); s != nil {
		t.Errorf("scaleDemoSeries([]) = %v, want nil", s)
	}
	pts := []core.TimePoint{{Date: "2026-04-15", Value: 100}, {Date: "2026-04-16", Value: 200}}
	scaledPts := scaleDemoSeries(pts, 0.5)
	if len(scaledPts) != 2 || scaledPts[1].Value != 100 {
		t.Errorf("unexpected scaledPts: %v", scaledPts)
	}

	// scaleDemoModelUsage
	records := []core.ModelUsageRecord{
		{
			RawModelID:      "gpt-4",
			InputTokens:     ptr(100.0),
			OutputTokens:    ptr(200.0),
			CachedTokens:    ptr(50.0),
			ReasoningTokens: ptr(30.0),
			TotalTokens:     ptr(350.0),
			CostUSD:         ptr(1.50),
			Requests:        ptr(10.0),
		},
	}
	scaledRecs := scaleDemoModelUsage(records, 0.5)
	if len(scaledRecs) != 1 || *scaledRecs[0].InputTokens != 50.0 || *scaledRecs[0].CostUSD != 0.75 {
		t.Errorf("unexpected scaled model usage: %+v", scaledRecs[0])
	}
}

func TestDemoScenario_StatusTransitions(t *testing.T) {
	// Codex status changes based on plan_percent_used
	makeCodexSnap := func(used *float64) core.UsageSnapshot {
		s := core.UsageSnapshot{
			ProviderID: "codex",
			Status:     core.StatusOK,
			Metrics:    map[string]core.Metric{},
		}
		if used != nil {
			s.Metrics["plan_percent_used"] = core.Metric{Used: used}
		}
		return s
	}

	// >= 100 -> StatusLimited
	snapLimited := makeCodexSnap(ptr(100.0))
	if st := demoStatusForSnapshot(snapLimited); st != core.StatusLimited {
		t.Errorf("status for 100%% = %v, want StatusLimited", st)
	}

	// >= 85 -> StatusNearLimit
	snapNearLimit := makeCodexSnap(ptr(88.0))
	if st := demoStatusForSnapshot(snapNearLimit); st != core.StatusNearLimit {
		t.Errorf("status for 88%% = %v, want StatusNearLimit", st)
	}

	// < 85 -> StatusOK
	snapOK := makeCodexSnap(ptr(50.0))
	if st := demoStatusForSnapshot(snapOK); st != core.StatusOK {
		t.Errorf("status for 50%% = %v, want StatusOK", st)
	}

	// Without plan_percent_used
	snapNoMetric := makeCodexSnap(nil)
	if st := demoStatusForSnapshot(snapNoMetric); st != core.StatusOK {
		t.Errorf("status without metric = %v, want StatusOK", st)
	}

	// Non-codex provider
	snapGemini := core.UsageSnapshot{ProviderID: "gemini_cli", Status: core.StatusNearLimit}
	if st := demoStatusForSnapshot(snapGemini); st != core.StatusNearLimit {
		t.Errorf("status for gemini = %v, want StatusNearLimit", st)
	}
}

func TestScopeSnapshotToWindow_Variations(t *testing.T) {
	// 1. Empty DailySeries returns snap unchanged
	snapEmpty := core.UsageSnapshot{ProviderID: "test", Metrics: map[string]core.Metric{"cost": {Used: ptr(10)}}}
	scopedEmpty := scopeSnapshotToWindow(snapEmpty, core.TimeWindow7d)
	if len(scopedEmpty.DailySeries) != 0 {
		t.Errorf("expected empty series to remain empty")
	}

	// 2. DailySeries with cost and analytics_cost
	now := time.Now()
	snapWithSeries := core.UsageSnapshot{
		ProviderID: "test",
		Metrics: map[string]core.Metric{
			"model_gpt4_cost": {Used: ptr(100.0)},
			"today_spend":     {Used: ptr(10.0)},
			"other_rate":      {Used: ptr(5.0)},
		},
		DailySeries: map[string][]core.TimePoint{
			"analytics_cost":     demoSeries(now, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10), // 10 days * 10 = 100 total
			"analytics_tokens":   demoSeries(now, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100),
			"analytics_requests": demoSeries(now, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1),
		},
	}

	// Scope to 3d window (3 days * 10 = 30 cost, fraction = 30/100 = 0.30)
	scoped3d := scopeSnapshotToWindow(snapWithSeries, core.TimeWindow3d)
	if wc, ok := scoped3d.Metrics["window_cost"]; !ok || wc.Used == nil || *wc.Used != 30.0 {
		t.Errorf("window_cost for 3d = %v, want 30.0", wc.Used)
	}
	if wt, ok := scoped3d.Metrics["window_tokens"]; !ok || wt.Used == nil || *wt.Used != 300.0 {
		t.Errorf("window_tokens for 3d = %v, want 300.0", wt.Used)
	}
	if wr, ok := scoped3d.Metrics["window_requests"]; !ok || wr.Used == nil || *wr.Used != 3.0 {
		t.Errorf("window_requests for 3d = %v, want 3.0", wr.Used)
	}
	// model_gpt4_cost should be scaled by 0.30 -> 30.0
	if mc, ok := scoped3d.Metrics["model_gpt4_cost"]; !ok || mc.Used == nil || *mc.Used != 30.0 {
		t.Errorf("scaled model_gpt4_cost = %v, want 30.0", mc.Used)
	}
	// today_spend and other_rate should NOT be scaled
	if ts, ok := scoped3d.Metrics["today_spend"]; !ok || ts.Used == nil || *ts.Used != 10.0 {
		t.Errorf("today_spend = %v, want 10.0", ts.Used)
	}
}

func TestKeepEntities(t *testing.T) {
	if k := keepEntities(core.TimeWindow1d); k != 3 {
		t.Errorf("keepEntities(1d) = %d, want 3", k)
	}
	if k := keepEntities(core.TimeWindow3d); k != 4 {
		t.Errorf("keepEntities(3d) = %d, want 4", k)
	}
	if k := keepEntities(core.TimeWindow7d); k != 5 {
		t.Errorf("keepEntities(7d) = %d, want 5", k)
	}
	if k := keepEntities(core.TimeWindow30d); k != 0 {
		t.Errorf("keepEntities(30d) = %d, want 0", k)
	}
	if k := keepEntities(core.TimeWindowAll); k != 0 {
		t.Errorf("keepEntities(all) = %d, want 0", k)
	}
}

func TestMatchSuffix(t *testing.T) {
	for _, suf := range breakdownSuffixes {
		key := "test_entity" + suf
		if got := matchSuffix(key); got != suf {
			t.Errorf("matchSuffix(%q) = %q, want %q", key, got, suf)
		}
	}
	if got := matchSuffix("no_matching_suffix"); got != "" {
		t.Errorf("matchSuffix(no_match) = %q, want empty", got)
	}
}

func TestTrimWindowVariant(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"tool_bash_today", "tool_bash"},
		{"tool_bash_1d", "tool_bash"},
		{"tool_bash_7d", "tool_bash"},
		{"tool_bash_30d", "tool_bash"},
		{"tool_bash", "tool_bash"},
	}
	for _, tt := range tests {
		if got := trimWindowVariant(tt.input); got != tt.want {
			t.Errorf("trimWindowVariant(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestPruneBreakdownsForWindow_Entities(t *testing.T) {
	metrics := map[string]core.Metric{
		// Models: ranked by cost (cost > alt)
		"model_m1_cost_usd":  {Used: ptr(100.0)},
		"model_m2_cost_usd":  {Used: ptr(50.0)},
		"model_m3_cost_usd":  {Used: ptr(20.0)},
		"model_m4_cost_usd":  {Used: ptr(10.0)},
		"model_m5_tokens":    {Used: ptr(500.0)}, // no cost
		// Clients
		"client_c1_cost_usd": {Used: ptr(10.0)},
		"client_c2_cost_usd": {Used: ptr(20.0)},
		"client_c3_cost_usd": {Used: ptr(30.0)},
		"client_c4_cost_usd": {Used: ptr(40.0)},
		// Languages
		"lang_go":            {Used: ptr(100.0)},
		"lang_py":            {Used: ptr(80.0)},
		"lang_ts":            {Used: ptr(60.0)},
		"lang_rs":            {Used: ptr(40.0)},
		"lang_c":             {Used: ptr(20.0)},
		// Tools
		"tool_calls_total":   {Used: ptr(1000.0)}, // aggregate, kept
		"tool_bash":          {Used: ptr(500.0)},
		"tool_bash_today":    {Used: ptr(50.0)},
		"tool_read":          {Used: ptr(400.0)},
		"tool_write":         {Used: ptr(300.0)},
		"tool_grep":          {Used: ptr(200.0)},
		"tool_glob":          {Used: ptr(100.0)},
		// MCP
		"mcp_servers_active": {Used: ptr(3.0)}, // aggregate, kept
		"mcp_github_total":   {Used: ptr(300.0)},
		"mcp_github_issue":   {Used: ptr(100.0)},
		"mcp_fs_total":       {Used: ptr(200.0)},
		"mcp_db_total":       {Used: ptr(100.0)},
		"mcp_slack_total":    {Used: ptr(50.0)},
	}
	dailySeries := map[string][]core.TimePoint{
		"tokens_model_m4": {{Date: "2026-04-16", Value: 10}},
		"usage_model_m4":  {{Date: "2026-04-16", Value: 10}},
		"usage_mcp_slack": {{Date: "2026-04-16", Value: 5}},
	}

	snap := core.UsageSnapshot{
		ProviderID:  "test",
		Metrics:     metrics,
		DailySeries: dailySeries,
	}

	// Prune for 1d window (keepEntities = 3)
	pruneBreakdownsForWindow(&snap, core.TimeWindow1d)

	// In models: m1, m2, m3 should be kept; m4 and m5 should be dropped
	if _, ok := snap.Metrics["model_m1_cost_usd"]; !ok {
		t.Errorf("model_m1 should be kept")
	}
	if _, ok := snap.Metrics["model_m4_cost_usd"]; ok {
		t.Errorf("model_m4 should be dropped")
	}
	if _, ok := snap.DailySeries["tokens_model_m4"]; ok {
		t.Errorf("tokens_model_m4 daily series should be deleted")
	}

	// In languages: top 3 (go, py, ts) kept; rs and c dropped
	if _, ok := snap.Metrics["lang_go"]; !ok {
		t.Errorf("lang_go should be kept")
	}
	if _, ok := snap.Metrics["lang_rs"]; ok {
		t.Errorf("lang_rs should be dropped")
	}

	// In tools: aggregate tool_calls_total kept, top 3 tools (bash, read, write) kept; grep and glob dropped
	if _, ok := snap.Metrics["tool_calls_total"]; !ok {
		t.Errorf("tool_calls_total aggregate should be kept")
	}
	if _, ok := snap.Metrics["tool_bash"]; !ok {
		t.Errorf("tool_bash should be kept")
	}
	if _, ok := snap.Metrics["tool_bash_today"]; !ok {
		t.Errorf("tool_bash_today companion should be kept with tool_bash")
	}
	if _, ok := snap.Metrics["tool_glob"]; ok {
		t.Errorf("tool_glob should be dropped")
	}

	// In MCP: servers_active kept, top 3 (github, fs, db) kept; slack dropped
	if _, ok := snap.Metrics["mcp_servers_active"]; !ok {
		t.Errorf("mcp_servers_active aggregate should be kept")
	}
	if _, ok := snap.Metrics["mcp_github_total"]; !ok {
		t.Errorf("mcp_github_total should be kept")
	}
	if _, ok := snap.Metrics["mcp_slack_total"]; ok {
		t.Errorf("mcp_slack_total should be dropped")
	}
	if _, ok := snap.DailySeries["usage_mcp_slack"]; ok {
		t.Errorf("usage_mcp_slack daily series should be deleted")
	}
}

func TestScaleByWindow_Patterns(t *testing.T) {
	shouldScale := []string{
		"model_gpt4_cost",
		"provider_anthropic_tokens",
		"client_cli_requests",
		"lang_go_count",
		"tool_bash_calls",
	}
	for _, k := range shouldScale {
		if !scaleByWindow(k) {
			t.Errorf("scaleByWindow(%q) = false, want true", k)
		}
	}

	shouldNotScale := []string{
		"model_error_rate",
		"token_balance",
		"quota_remaining",
		"credits_used",
		"today_cost",
		"7d_cost",
		"30d_cost",
		"all_time_tokens",
		"usage_summary",
		"keys_count",
		"analytics_cost",
		"plan_percent",
		"unknown_metric",
	}
	for _, k := range shouldNotScale {
		if scaleByWindow(k) {
			t.Errorf("scaleByWindow(%q) = true, want false", k)
		}
	}
}

func TestConcurrency_ScenarioAndProviders(t *testing.T) {
	scenario := newDemoScenario(time.Now(), defaultDemoConfig())
	providersList := buildDemoProviders(providers.AllProviders(), scenario)
	accounts := buildDemoAccounts()

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				if j%3 == 0 {
					scenario.Advance()
				}
				_ = scenario.CurrentPhase()

				acct := accounts[workerID%len(accounts)]
				prov := providersList[workerID%len(providersList)]
				snap, err := prov.Fetch(context.Background(), acct)
				if err == nil {
					_ = scopeSnapshotToWindow(snap, core.TimeWindow7d)
				}
			}
		}(i)
	}
	wg.Wait()
}

func TestWidgetRenderingAndTUIModel(t *testing.T) {
	scenario := newDemoScenario(time.Now(), defaultDemoConfig())
	demoProviders := buildDemoProviders(providers.AllProviders(), scenario)
	accounts := buildDemoAccounts()

	// 1. Render Dashboard and Detail Widgets for all providers
	for _, p := range demoProviders {
		acct := core.AccountConfig{ID: demoAccountID(p.ID()), Provider: p.ID()}
		snap, err := p.Fetch(context.Background(), acct)
		if err != nil {
			t.Fatalf("fetch for %q failed: %v", p.ID(), err)
		}

		dashWidget := p.DashboardWidget()
		detailWidget := p.DetailWidget()

		// Verify widgets describe and spec match
		_ = p.Describe()
		_ = p.Spec()

		if p.ID() == "claude_code" || p.ID() == "codex" || p.ID() == "cursor" {
			if snap.Metrics == nil || len(snap.Metrics) == 0 {
				t.Fatalf("expected metrics for %q", p.ID())
			}
		}
		_ = dashWidget
		_ = detailWidget
	}

	// 2. Initialize full TUI Model and verify View()
	model := tui.NewModel(
		0.20,
		0.05,
		false,
		config.DashboardConfig{HideSectionsWithNoData: true},
		accounts,
		core.TimeWindow30d,
	)

	// Set window size
	updatedModel, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	model = updatedModel.(tui.Model)

	// Send snapshots message
	snaps := make(map[string]core.UsageSnapshot, len(accounts))
	for _, p := range demoProviders {
		acct := core.AccountConfig{ID: demoAccountID(p.ID()), Provider: p.ID()}
		snap, _ := p.Fetch(context.Background(), acct)
		snaps[acct.ID] = scopeSnapshotToWindow(snap, core.TimeWindow30d)
	}

	updatedModel, _ = model.Update(tui.SnapshotsMsg{
		Snapshots:  snaps,
		TimeWindow: core.TimeWindow30d,
		RequestID:  1,
	})
	model = updatedModel.(tui.Model)

	// Verify model view renders
	view := model.View()
	if view == "" {
		t.Fatal("expected non-empty model view")
	}

	// Test window callbacks
	var currentWindow atomic.Value
	currentWindow.Store(core.TimeWindow30d)
	model.SetOnTimeWindowChange(func(w core.TimeWindow) { currentWindow.Store(w) })
	model.SetOnRefresh(func(req tui.RefreshRequest) uint64 {
		currentWindow.Store(req.TimeWindow)
		return 2
	})
}

func TestPruning_BranchEdgeCases(t *testing.T) {
	// 1. pruneEntityMetrics with empty entity, no suffix, nil Used, and ranking
	metrics := map[string]core.Metric{
		"model_mix_source":  {Used: ptr(1.0)},  // no suffix
		"model__cost_usd":   {Used: ptr(1.0)},  // empty entity name
		"model_m1_cost_usd": {},               // nil Used
		"model_m2_cost_usd": {Used: ptr(50.0)},
		"model_m3_tokens":   {Used: ptr(100.0)},
	}
	dropped := pruneEntityMetrics(metrics, "model_", 1)
	if len(dropped) == 0 {
		t.Errorf("expected dropped entities")
	}

	// 2. pruneToolMetrics with empty entity, MCP tool name, nil Used
	toolMetrics := map[string]core.Metric{
		"tool_":            {Used: ptr(1.0)}, // empty name
		"tool_today":       {Used: ptr(1.0)}, // trimWindowVariant becomes ""
		"tool_mcp_fs_read": {Used: ptr(5.0)}, // IsMCPToolMetricName == true
		"tool_git":         {},               // nil Used
		"tool_curl":        {Used: ptr(10.0)},
	}
	pruneToolMetrics(toolMetrics, 1)

	// 3. pruneMCPServers with servers_active, empty rest, server without underscore
	mcpMetrics := map[string]core.Metric{
		"mcp_servers_active": {Used: ptr(1.0)},
		"mcp_":               {Used: ptr(1.0)},
		"mcp_serverwithoutunderscore": {Used: ptr(10.0)},
	}
	droppedMCP := pruneMCPServers(mcpMetrics, 5) // len <= keep
	if droppedMCP != nil {
		t.Errorf("expected nil droppedMCP when count <= keep, got %v", droppedMCP)
	}

	// 4. pruneFlatMetrics when len <= keep
	flatMetrics := map[string]core.Metric{
		"lang_go": {Used: ptr(10.0)},
	}
	pruneFlatMetrics(flatMetrics, "lang_", 5)
	if _, ok := flatMetrics["lang_go"]; !ok {
		t.Errorf("expected lang_go to be retained")
	}
}

func TestDemoMetricUsed_EdgeCases(t *testing.T) {
	// Empty metric
	if val, ok := demoMetricUsed(core.Metric{}); ok || val != 0 {
		t.Errorf("expected (0, false) for empty metric, got (%v, %v)", val, ok)
	}

	// Metric with Limit only
	if val, ok := demoMetricUsed(core.Metric{Limit: ptr(100.0)}); ok || val != 0 {
		t.Errorf("expected (0, false) for limit-only metric, got (%v, %v)", val, ok)
	}

	// Metric with Remaining only
	if val, ok := demoMetricUsed(core.Metric{Remaining: ptr(50.0)}); ok || val != 0 {
		t.Errorf("expected (0, false) for remaining-only metric, got (%v, %v)", val, ok)
	}
}


