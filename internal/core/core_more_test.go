package core

import (
	"context"
	"testing"

	"github.com/nurulislamz/agentusage/internal/observability"
)

func TestStructuredLogger_Methods(t *testing.T) {
	// 1. Nil logger safety
	var nilLogger *StructuredLogger
	nilLogger.Infof("event", "msg %s", "arg")
	nilLogger.Warnf("event", "msg %s", "arg")

	// 2. Default logger (always true)
	logger := NewLogger("test-component")
	logger.Infof("test_event", "hello %s", "world")
	logger.Infof("empty_format", "")
	logger.Warnf("warn_event", "warning %d", 42)

	// 3. WithVerbose
	verboseFalse := logger.WithVerbose(func() bool { return false })
	verboseFalse.Infof("silent_event", "should not log")

	verboseTrue := logger.WithVerbose(func() bool { return true })
	verboseTrue.Infof("loud_event", "should log")

	verboseNil := logger.WithVerbose(nil)
	verboseNil.Infof("default_event", "should log")

	// 4. Test with observability enabled
	observability.ResetForTesting()
	defer observability.ResetForTesting()

	_ = observability.Init(context.Background(), observability.Config{
		Enabled:     true,
		Endpoint:    "http://127.0.0.1:4318",
		Insecure:    true,
		ServiceName: "agentusage-test",
	})

	logger.Infof("otel_event", "hello otel %s", "world")
	logger.Warnf("otel_warn", "warning otel %d", 99)
	logger.Infof("empty_msg", "")
}

func TestTracef_And_DebugEnabled(t *testing.T) {
	// Calling DebugEnabled and Tracef should be safe
	_ = DebugEnabled()
	Tracef("trace message with arg %d", 123)

	// Test with observability enabled
	observability.ResetForTesting()
	defer observability.ResetForTesting()

	_ = observability.Init(context.Background(), observability.Config{
		Enabled:     true,
		Endpoint:    "http://127.0.0.1:4318",
		Insecure:    true,
		ServiceName: "agentusage-test",
	})
	Tracef("trace message to otel %s", "debug")
}

func TestSystemClock_Now(t *testing.T) {
	clock := SystemClock{}
	now := clock.Now()
	if now.IsZero() {
		t.Error("SystemClock.Now() returned zero time")
	}
}

func TestSortedCompactStrings_Cases(t *testing.T) {
	// 1. Nil / empty
	if got := SortedCompactStrings(nil); got != nil {
		t.Errorf("SortedCompactStrings(nil) = %v, want nil", got)
	}
	if got := SortedCompactStrings([]string{"", "  ", "\t"}); got != nil {
		t.Errorf("SortedCompactStrings(whitespace) = %v, want nil", got)
	}

	// 2. Deduplication and sorting
	input := []string{" beta ", "alpha", "beta", "  gamma  ", "alpha"}
	got := SortedCompactStrings(input)
	want := []string{"alpha", "beta", "gamma"}
	if len(got) != len(want) {
		t.Fatalf("SortedCompactStrings len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestTokenUsage_SumTotalTokens_And_HasTokenData(t *testing.T) {
	// 1. SumTotalTokens computes total
	in := int64(100)
	out := int64(50)
	reasoning := int64(25)
	cacheRead := int64(10)
	cacheWrite := int64(5)

	u := TokenUsage{
		InputTokens:      &in,
		OutputTokens:     &out,
		ReasoningTokens:  &reasoning,
		CacheReadTokens:  &cacheRead,
		CacheWriteTokens: &cacheWrite,
	}
	u.SumTotalTokens()
	if u.TotalTokens == nil || *u.TotalTokens != 190 {
		t.Errorf("TotalTokens = %v, want 190", u.TotalTokens)
	}

	// SumTotalTokens does not overwrite existing total
	existing := int64(999)
	u.TotalTokens = &existing
	u.SumTotalTokens()
	if *u.TotalTokens != 999 {
		t.Errorf("TotalTokens overwritten = %d, want 999", *u.TotalTokens)
	}

	// 2. HasTokenData
	emptyUsage := TokenUsage{}
	if emptyUsage.HasTokenData() {
		t.Error("emptyUsage.HasTokenData() = true, want false")
	}

	cost := 0.05
	costUsage := TokenUsage{CostUSD: &cost}
	if !costUsage.HasTokenData() {
		t.Error("costUsage.HasTokenData() = false, want true")
	}

	// 3. Int64Ptr
	ptr := Int64Ptr(42)
	if ptr == nil || *ptr != 42 {
		t.Errorf("Int64Ptr(42) = %v", ptr)
	}
}

func TestAccountConfig_PathsAndHints(t *testing.T) {
	acct := AccountConfig{
		ID:       "test-account",
		Provider: "test-provider",
		ProviderPaths: map[string]string{
			"primary_path": "/path/provider",
		},
		Paths: map[string]string{
			"legacy_path":  "/path/legacy",
			"primary_path": "/path/legacy_primary",
		},
		RuntimeHints: map[string]string{
			"hint_path": "/path/hint",
			"custom":    "val",
		},
	}

	// 1. Path priority: ProviderPaths > Paths > RuntimeHints > fallback
	if got := acct.Path("primary_path", "/fallback"); got != "/path/provider" {
		t.Errorf("Path(primary_path) = %q, want /path/provider", got)
	}
	if got := acct.Path("legacy_path", "/fallback"); got != "/path/legacy" {
		t.Errorf("Path(legacy_path) = %q, want /path/legacy", got)
	}
	if got := acct.Path("hint_path", "/fallback"); got != "/path/hint" {
		t.Errorf("Path(hint_path) = %q, want /path/hint", got)
	}
	if got := acct.Path("missing_path", "/fallback"); got != "/fallback" {
		t.Errorf("Path(missing_path) = %q, want /fallback", got)
	}
	if got := acct.Path("missing_path", ""); got != "" {
		t.Errorf("Path(missing_path with empty fallback) = %q, want empty", got)
	}

	// 2. SetPath
	var nilAcct *AccountConfig
	nilAcct.SetPath("k", "v") // safe no-op

	acct2 := AccountConfig{}
	acct2.SetPath("", "val")
	acct2.SetPath("key", "")
	if len(acct2.ProviderPaths) != 0 {
		t.Errorf("expected empty ProviderPaths, got %+v", acct2.ProviderPaths)
	}
	acct2.SetPath("data_dir", "/var/data")
	if acct2.ProviderPaths["data_dir"] != "/var/data" {
		t.Errorf("SetPath failed, got %q", acct2.ProviderPaths["data_dir"])
	}

	// 3. Hint and SetHint
	if got := acct.Hint("custom", "fallback"); got != "val" {
		t.Errorf("Hint(custom) = %q, want val", got)
	}
	if got := acct.Hint("missing", "fallback"); got != "fallback" {
		t.Errorf("Hint(missing) = %q, want fallback", got)
	}
	nilAcct.SetHint("k", "v")
	acct2.SetHint("session_id", "s123")
	if acct2.RuntimeHints["session_id"] != "s123" {
		t.Errorf("SetHint failed, got %q", acct2.RuntimeHints["session_id"])
	}

	// 4. PathMap
	emptyAcct := AccountConfig{}
	if got := emptyAcct.PathMap(); got != nil {
		t.Errorf("emptyAcct.PathMap() = %v, want nil", got)
	}
	pm := acct.PathMap()
	if pm["primary_path"] != "/path/provider" || pm["legacy_path"] != "/path/legacy" {
		t.Errorf("PathMap() = %+v", pm)
	}
}

func TestAccountConfig_ResolveAPIKey(t *testing.T) {
	// 1. Direct APIKey
	acct1 := AccountConfig{APIKey: "sk-direct"}
	if key := acct1.ResolveAPIKey(); key != "sk-direct" {
		t.Errorf("ResolveAPIKey = %q, want sk-direct", key)
	}

	// 2. Direct Token
	acct2 := AccountConfig{Token: "token-direct"}
	if key := acct2.ResolveAPIKey(); key != "token-direct" {
		t.Errorf("ResolveAPIKey = %q, want token-direct", key)
	}

	// 3. APIKeyEnv
	t.Setenv("TEST_CORE_API_KEY_ENV", "sk-from-env")
	acct3 := AccountConfig{APIKeyEnv: "TEST_CORE_API_KEY_ENV"}
	if key := acct3.ResolveAPIKey(); key != "sk-from-env" {
		t.Errorf("ResolveAPIKey = %q, want sk-from-env", key)
	}

	// 4. Fallback when env not set
	acct4 := AccountConfig{APIKeyEnv: "NONEXISTENT_CORE_ENV_VAR"}
	_ = acct4.ResolveAPIKey()
}

func TestUsageSnapshot_NewAuthSnapshot_And_MergeAccounts(t *testing.T) {
	// 1. NewAuthSnapshot
	authSnap := NewAuthSnapshot("claude_code", "acc-1", "Claude Code auth required")
	if authSnap.ProviderID != "claude_code" || authSnap.Status != StatusAuth {
		t.Errorf("NewAuthSnapshot = %+v", authSnap)
	}

	// 2. MergeAccounts
	manual := []AccountConfig{
		{ID: "openrouter-work", Provider: "openrouter"},
		{ID: "shared-id", Provider: "manual-prov"},
	}
	autodetect := []AccountConfig{
		{ID: "shared-id", Provider: "auto-prov"},
		{ID: "ollama-local", Provider: "ollama"},
	}
	merged := MergeAccounts(manual, autodetect)
	if len(merged) != 3 {
		t.Fatalf("MergeAccounts len = %d, want 3", len(merged))
	}
	if merged[0].ID != "openrouter-work" || merged[1].ID != "shared-id" || merged[2].ID != "ollama-local" {
		t.Errorf("merged = %+v", merged)
	}

	// 3. SetAttribute and SetDiagnostic
	snap := UsageSnapshot{}
	snap.SetAttribute("plan", "enterprise")
	snap.SetDiagnostic("rate_limit", "3000")
	if snap.Attributes["plan"] != "enterprise" || snap.Diagnostics["rate_limit"] != "3000" {
		t.Errorf("snap = %+v", snap)
	}
}

func TestUsageBreakdowns_Helpers(t *testing.T) {
	// 1. HasMCPUsage
	snapEmpty := UsageSnapshot{}
	if HasMCPUsage(snapEmpty) {
		t.Error("HasMCPUsage(empty) = true, want false")
	}

	snapMCP := UsageSnapshot{
		Metrics: map[string]Metric{
			"mcp_slack_total": {Used: Float64Ptr(2)},
		},
	}
	if !HasMCPUsage(snapMCP) {
		t.Error("HasMCPUsage(mcp) = false, want true")
	}

	// 2. IncludeDetailMetricKey
	if !IncludeDetailMetricKey("cost_5h") {
		t.Error("IncludeDetailMetricKey(cost_5h) = false, want true")
	}

	// 3. ExtractUpstreamProviderBreakdown
	breakdown, _ := ExtractUpstreamProviderBreakdown(snapMCP)
	_ = breakdown
}

func TestModelUsage_AppendModelUsage(t *testing.T) {
	snap := UsageSnapshot{}
	in, out := 100.0, 50.0
	cost := 0.01
	rec := ModelUsageRecord{
		RawModelID:   "gpt-4o",
		InputTokens:  &in,
		OutputTokens: &out,
		CostUSD:      &cost,
	}

	snap.AppendModelUsage(rec)
	snap.AppendModelUsage(ModelUsageRecord{}) // empty RawModelID ignored

	if len(snap.ModelUsage) != 1 {
		t.Fatalf("AppendModelUsage count = %d, want 1", len(snap.ModelUsage))
	}
	if snap.ModelUsage[0].RawModelID != "gpt-4o" {
		t.Errorf("AppendModelUsage = %+v", snap.ModelUsage[0])
	}
}

func TestAnalyticsSnapshot_Helpers(t *testing.T) {
	tokens := 1000.0
	cost := 0.05
	snap := UsageSnapshot{
		ProviderID: "claude_code",
		AccountID:  "claude-work",
		Status:     StatusOK,
		ModelUsage: []ModelUsageRecord{
			{
				RawModelID:  "claude-3-7-sonnet",
				InputTokens: &tokens,
				CostUSD:     &cost,
			},
		},
		DailySeries: map[string][]TimePoint{
			"tokens_total": {{Date: "2025-01-01", Value: 1000}},
		},
	}

	// 1. ExtractAnalyticsModelUsage
	usage := ExtractAnalyticsModelUsage(snap)
	if len(usage) != 1 {
		t.Fatalf("ExtractAnalyticsModelUsage count = %d, want 1", len(usage))
	}

	// 2. SelectAnalyticsWeightSeries
	series := SelectAnalyticsWeightSeries(snap.DailySeries)
	if len(series) == 0 {
		t.Error("SelectAnalyticsWeightSeries returned empty")
	}

	// 3. analyticsModelDisplayName
	name := analyticsModelDisplayName(ModelUsageRecord{
		RawModelID: "claude-3-7-sonnet",
	})
	if name != "claude-3-7-sonnet" {
		t.Errorf("analyticsModelDisplayName = %q, want 'claude-3-7-sonnet'", name)
	}

	nameCanonical := analyticsModelDisplayName(ModelUsageRecord{
		Dimensions: map[string]string{
			"canonical_group_id": "Claude 3.7",
		},
	})
	if nameCanonical != "Claude 3.7" {
		t.Errorf("analyticsModelDisplayName with group = %q, want 'Claude 3.7'", nameCanonical)
	}
}

func TestDashboardWidget_DetailSections_And_MissingMetrics(t *testing.T) {
	// 1. DefaultDetailSectionOrder
	order := DefaultDetailSectionOrder()
	if len(order) == 0 {
		t.Error("DefaultDetailSectionOrder returned empty")
	}

	// 2. IsKnownDetailStandardSection and DetailSectionLabel
	for _, sec := range order {
		if !IsKnownDetailStandardSection(sec) {
			t.Errorf("IsKnownDetailStandardSection(%q) = false, want true", sec)
		}
		label := DetailSectionLabel(sec)
		if label == "" {
			t.Errorf("DetailSectionLabel(%q) returned empty", sec)
		}
	}
	if IsKnownDetailStandardSection("unknown_section") {
		t.Error("IsKnownDetailStandardSection(unknown) = true, want false")
	}
	if got := DetailSectionLabel("custom_sec"); got != "custom_sec" {
		t.Errorf("DetailSectionLabel(custom_sec) = %q, want 'custom_sec'", got)
	}

	// 3. MissingMetrics
	w := DashboardWidget{
		DataSpec: WidgetDataSpec{
			RequiredMetricKeys: []string{"rpm", "tpm", ""},
		},
	}
	snap := UsageSnapshot{
		Metrics: map[string]Metric{
			"rpm": {Limit: Float64Ptr(100)},
		},
	}
	missing := w.MissingMetrics(snap)
	if len(missing) != 1 || missing[0] != "tpm" {
		t.Errorf("MissingMetrics = %+v, want ['tpm']", missing)
	}

	// 4. Widget IsZero
	wEmpty := DashboardWidget{}
	if !wEmpty.IsZero() {
		t.Error("wEmpty.IsZero() = false, want true")
	}
}

func TestHasLanguageUsage_TrueAndFalse(t *testing.T) {
	snapEmpty := UsageSnapshot{}
	if HasLanguageUsage(snapEmpty) {
		t.Error("HasLanguageUsage(empty) = true, want false")
	}

	snapLang := UsageSnapshot{
		Metrics: map[string]Metric{
			"lang_go": {Used: Float64Ptr(10)},
		},
	}
	if !HasLanguageUsage(snapLang) {
		t.Error("HasLanguageUsage(lang_go) = false, want true")
	}
}

func TestModelIdentity_NormalizeCanonicalModel(t *testing.T) {
	models := []struct {
		provider string
		model    string
	}{
		{"google", "gemini-2.0-flash-exp"},
		{"google", "gemini-1.5-pro-latest"},
		{"anthropic", "claude-3-5-sonnet-20241022"},
		{"anthropic", "claude-3-opus-20240229"},
		{"deepseek", "deepseek-chat"},
		{"deepseek", "deepseek-reasoner"},
		{"openai", "gpt-4o-2024-08-06"},
		{"openai", "o1-mini-2024-09-12"},
		{"qwen", "qwen-2.5-coder-32b"},
		{"custom", "custom-unknown-model"},
	}

	cfg := DefaultModelNormalizationConfig()
	for _, tc := range models {
		res := normalizeCanonicalModel(tc.provider, tc.model, cfg)
		if res.LineageID == "" {
			t.Errorf("normalizeCanonicalModel(%s, %s).LineageID is empty", tc.provider, tc.model)
		}
	}
}

func TestSnapshotNormalize_WithConfig(t *testing.T) {
	snap := UsageSnapshot{
		ProviderID: "claude_code",
		AccountID:  "acc-1",
		DailySeries: map[string][]TimePoint{
			"cost": {
				{Date: "2025-01-01", Value: 1.5},
				{Date: "2025-01-02", Value: 2.5},
			},
		},
	}
	norm := NormalizeUsageSnapshotWithConfig(snap, DefaultModelNormalizationConfig())
	if norm.ProviderID != "claude_code" {
		t.Errorf("norm.ProviderID = %q, want claude_code", norm.ProviderID)
	}
}

func TestAnalyticsNormalize_AliasMetricKeyAndInto(t *testing.T) {
	// 1. aliasMetricKey nil safety and basic alias
	aliasMetricKey(nil, "src", "tgt", Metric{})
	aliasMetricKey(&UsageSnapshot{}, "", "tgt", Metric{})
	aliasMetricKey(&UsageSnapshot{}, "src", "", Metric{})

	snap := &UsageSnapshot{
		Metrics: map[string]Metric{
			"src": {Used: Float64Ptr(10)},
		},
	}
	aliasMetricKey(snap, "src", "tgt", Metric{Used: Float64Ptr(10)})
	if _, ok := snap.Metrics["tgt"]; !ok {
		t.Error("aliasMetricKey did not insert target metric")
	}

	// 2. aliasMetricInto nil safety and alias
	aliasMetricInto(nil, "can")
	aliasMetricInto(snap, "")

	snap2 := &UsageSnapshot{
		Metrics: map[string]Metric{
			"alias1": {Used: Float64Ptr(20)},
		},
	}
	aliasMetricInto(snap2, "canonical", "alias1", "alias2")
	if _, ok := snap2.Metrics["canonical"]; !ok {
		t.Error("aliasMetricInto did not resolve alias1 into canonical")
	}

	// 3. bestWindowCostMetric
	if _, ok := bestWindowCostMetric(nil); ok {
		t.Error("bestWindowCostMetric(nil) returned true")
	}
	snapCost := &UsageSnapshot{
		Metrics: map[string]Metric{
			"today_api_cost": {Used: Float64Ptr(5.5)},
		},
	}
	m, ok := bestWindowCostMetric(snapCost)
	if !ok || m.Used == nil || *m.Used != 5.5 {
		t.Errorf("bestWindowCostMetric = %+v, want 5.5", m)
	}
}

func TestModelNormalizationConfig_Normalize(t *testing.T) {
	cfgEmpty := ModelNormalizationConfig{}
	norm := NormalizeModelNormalizationConfig(cfgEmpty)
	if norm.GroupBy != ModelNormalizationGroupLineage || norm.MinConfidence != 0.80 {
		t.Errorf("NormalizeModelNormalizationConfig(empty) = %+v", norm)
	}

	cfgInvalid := ModelNormalizationConfig{
		GroupBy:       "invalid_group",
		MinConfidence: 2.5,
	}
	norm2 := NormalizeModelNormalizationConfig(cfgInvalid)
	if norm2.GroupBy != ModelNormalizationGroupLineage || norm2.MinConfidence != 1.0 {
		t.Errorf("NormalizeModelNormalizationConfig(invalid) = %+v", norm2)
	}
}

func TestUsageSnapshot_MetaValue(t *testing.T) {
	snap := UsageSnapshot{
		Attributes:  map[string]string{"attr_key": "attr_val"},
		Diagnostics: map[string]string{"diag_key": "diag_val"},
		Raw:         map[string]string{"raw_key": "raw_val"},
	}

	if got, ok := snap.MetaValue("attr_key"); !ok || got != "attr_val" {
		t.Errorf("MetaValue(attr_key) = %q, %v, want attr_val, true", got, ok)
	}
	if got, ok := snap.MetaValue("diag_key"); !ok || got != "diag_val" {
		t.Errorf("MetaValue(diag_key) = %q, %v, want diag_val, true", got, ok)
	}
	if got, ok := snap.MetaValue("raw_key"); !ok || got != "raw_val" {
		t.Errorf("MetaValue(raw_key) = %q, %v, want raw_val, true", got, ok)
	}
	if got, ok := snap.MetaValue("missing_key"); ok || got != "" {
		t.Errorf("MetaValue(missing_key) = %q, %v, want empty, false", got, ok)
	}
}

func TestModelIdentity_KnownVendors(t *testing.T) {
	vendors := []string{"anthropic", "openai", "google", "deepseek", "meta", "mistral"}
	for _, v := range vendors {
		if !isKnownVendor(v) {
			t.Errorf("isKnownVendor(%q) = false, want true", v)
		}
	}
	if isKnownVendor("unknown_vendor_xyz") {
		t.Error("isKnownVendor(unknown) = true, want false")
	}
}

func TestModelUsageFromMetrics_ParseAndApply(t *testing.T) {
	// 1. parseModelRawValue
	if _, ok := parseModelRawValue(""); ok {
		t.Error("parseModelRawValue('') returned true")
	}
	if v, ok := parseModelRawValue("   1,234.50   "); !ok || v != 1234.50 {
		t.Errorf("parseModelRawValue('1,234.50') = %v, %v, want 1234.50, true", v, ok)
	}
	if _, ok := parseModelRawValue("not-a-number"); ok {
		t.Error("parseModelRawValue(not-a-number) returned true")
	}

	// 2. applyModelMetricAcc
	applyModelMetricAcc(nil, modelMetricInput, 10)

	acc := &modelUsageAccumulator{}
	applyModelMetricAcc(acc, modelMetricInput, 100)
	applyModelMetricAcc(acc, modelMetricOutput, 50)
	applyModelMetricAcc(acc, modelMetricCached, 25)
	applyModelMetricAcc(acc, modelMetricReasoning, 10)
	applyModelMetricAcc(acc, modelMetricCostUSD, 0.05)
	applyModelMetricAcc(acc, modelMetricRequests, 2)
	applyModelMetricAcc(acc, modelMetricInput, -5) // <= 0 ignored

	if acc.inputTokens != 100 || acc.outputTokens != 50 || acc.cachedTokens != 25 || acc.reasoningTokens != 10 || acc.costUSD != 0.05 || acc.requests != 2 {
		t.Errorf("accumulator = %+v", acc)
	}
}

func TestAnalyticsNormalize_NormalizeSeriesPoints(t *testing.T) {
	// 1. Nil / empty
	if got := normalizeSeriesPoints(nil); got != nil {
		t.Errorf("normalizeSeriesPoints(nil) = %v, want nil", got)
	}

	// 2. Single item invalid
	if got := normalizeSeriesPoints([]TimePoint{{Date: " ", Value: 10}}); got != nil {
		t.Errorf("normalizeSeriesPoints(empty date) = %v, want nil", got)
	}
	if got := normalizeSeriesPoints([]TimePoint{{Date: "2025-01-01", Value: -1}}); got != nil {
		t.Errorf("normalizeSeriesPoints(negative val) = %v, want nil", got)
	}

	// 3. Single item trimmed
	gotSingle := normalizeSeriesPoints([]TimePoint{{Date: " 2025-01-01 ", Value: 5}})
	if len(gotSingle) != 1 || gotSingle[0].Date != "2025-01-01" {
		t.Errorf("normalizeSeriesPoints(single) = %+v", gotSingle)
	}

	// 4. Multiple unsorted with duplicates
	points := []TimePoint{
		{Date: "2025-01-02", Value: 10},
		{Date: "2025-01-01", Value: 5},
		{Date: "2025-01-01", Value: 3},
		{Date: "   ", Value: 100},
	}
	res := normalizeSeriesPoints(points)
	if len(res) != 2 {
		t.Fatalf("normalizeSeriesPoints len = %d, want 2", len(res))
	}
	if res[0].Date != "2025-01-01" || res[0].Value != 8 {
		t.Errorf("res[0] = %+v, want 2025-01-01: 8", res[0])
	}
	if res[1].Date != "2025-01-02" || res[1].Value != 10 {
		t.Errorf("res[1] = %+v, want 2025-01-02: 10", res[1])
	}
}

func TestUsageBreakdowns_UpstreamAndClientBreakdown(t *testing.T) {
	// 1. ExtractUpstreamProviderBreakdown
	snapUpstream := UsageSnapshot{
		Metrics: map[string]Metric{
			"upstream_openai_cost_usd":        {Used: Float64Ptr(1.50)},
			"upstream_openai_input_tokens":    {Used: Float64Ptr(1000)},
			"upstream_openai_output_tokens":   {Used: Float64Ptr(500)},
			"upstream_openai_requests":        {Used: Float64Ptr(10)},
			"upstream_anthropic_cost_usd":     {Used: Float64Ptr(2.00)},
			"upstream_anthropic_input_tokens": {Used: Float64Ptr(2000)},
		},
	}
	upstreams, usedUp := ExtractUpstreamProviderBreakdown(snapUpstream)
	if len(upstreams) != 2 {
		t.Fatalf("ExtractUpstreamProviderBreakdown count = %d, want 2", len(upstreams))
	}
	if !usedUp["upstream_openai_cost_usd"] {
		t.Error("usedUp missing upstream_openai_cost_usd")
	}

	// 2. ExtractClientBreakdown
	snapClient := UsageSnapshot{
		Metrics: map[string]Metric{
			"client_cursor_total_tokens":     {Used: Float64Ptr(1500)},
			"client_cursor_input_tokens":     {Used: Float64Ptr(1000)},
			"client_cursor_output_tokens":    {Used: Float64Ptr(500)},
			"client_cursor_cached_tokens":    {Used: Float64Ptr(200)},
			"client_cursor_reasoning_tokens": {Used: Float64Ptr(100)},
			"client_cursor_requests":         {Used: Float64Ptr(15)},
			"client_cursor_sessions":         {Used: Float64Ptr(3)},
			"source_claude_requests":         {Used: Float64Ptr(20)},
			"source_claude_requests_today":   {Used: Float64Ptr(5)},
		},
	}
	clients, usedClient := ExtractClientBreakdown(snapClient)
	if len(clients) == 0 {
		t.Error("ExtractClientBreakdown returned empty")
	}
	if !usedClient["client_cursor_total_tokens"] {
		t.Error("usedClient missing client_cursor_total_tokens")
	}
}
