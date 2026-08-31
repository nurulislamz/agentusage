package core

import (
	"testing"
	"time"
)

var (
	benchFixedTime = time.Date(2026, 2, 22, 12, 0, 0, 0, time.UTC)
	sinkSnapshot   UsageSnapshot
	sinkRecords    []ModelUsageRecord
	sinkEntries    []AnalyticsModelUsageEntry
	sinkCost       float64
	sinkIdentity   canonicalModelIdentity
)

func makeBenchFallbackSnapshot() UsageSnapshot {
	return UsageSnapshot{
		ProviderID: "openai",
		AccountID:  "acct-prod",
		Timestamp:  benchFixedTime,
		Metrics: map[string]Metric{
			"model_gpt_4o_input_tokens":             {Used: Float64Ptr(15000), Unit: "tokens", Window: "1d"},
			"model_gpt_4o_output_tokens":            {Used: Float64Ptr(2500), Unit: "tokens", Window: "1d"},
			"model_gpt_4o_cost_usd":                 {Used: Float64Ptr(0.08), Unit: "USD", Window: "1d"},
			"model_claude_3_5_sonnet_input_tokens":  {Used: Float64Ptr(30000), Unit: "tokens", Window: "1d"},
			"model_claude_3_5_sonnet_output_tokens": {Used: Float64Ptr(4000), Unit: "tokens", Window: "1d"},
			"model_claude_3_5_sonnet_cost_usd":      {Used: Float64Ptr(0.15), Unit: "USD", Window: "1d"},
			"model_gemini_1_5_pro_input_tokens":     {Used: Float64Ptr(10000), Unit: "tokens", Window: "1d"},
			"model_gemini_1_5_pro_output_tokens":    {Used: Float64Ptr(1200), Unit: "tokens", Window: "1d"},
			"model_deepseek_chat_input_tokens":      {Used: Float64Ptr(5000), Unit: "tokens", Window: "1d"},
			"model_deepseek_chat_output_tokens":     {Used: Float64Ptr(800), Unit: "tokens", Window: "1d"},
			"today_api_cost":                        {Used: Float64Ptr(0.23), Unit: "USD", Window: "1d"},
			"7d_api_cost":                           {Used: Float64Ptr(1.45), Unit: "USD", Window: "7d"},
			"all_time_api_cost":                     {Used: Float64Ptr(12.30), Unit: "USD", Window: "all-time"},
		},
		Raw: map[string]string{
			"quota_api":     "live",
			"api_error":     "HTTP 500 transient warning",
			"account_email": "dev@example.com",
			"rate_limit":    "tier-4",
		},
	}
}

func makeBenchPrePopulatedSnapshot() UsageSnapshot {
	return UsageSnapshot{
		ProviderID: "openrouter",
		AccountID:  "acct-or-01",
		Timestamp:  benchFixedTime,
		Metrics: map[string]Metric{
			"today_api_cost":    {Used: Float64Ptr(1.85), Unit: "USD", Window: "1d"},
			"7d_api_cost":       {Used: Float64Ptr(12.40), Unit: "USD", Window: "7d"},
			"30d_api_cost":      {Used: Float64Ptr(45.00), Unit: "USD", Window: "30d"},
			"all_time_api_cost": {Used: Float64Ptr(120.50), Unit: "USD", Window: "all-time"},
		},
		ModelUsage: []ModelUsageRecord{
			{
				RawModelID:   "anthropic/claude-3-5-sonnet-20241022",
				InputTokens:  Float64Ptr(120000),
				OutputTokens: Float64Ptr(18000),
				CostUSD:      Float64Ptr(0.63),
				Requests:     Float64Ptr(45),
				Window:       "1d",
			},
			{
				RawModelID:   "openai/gpt-4o-2024-08-06",
				InputTokens:  Float64Ptr(85000),
				OutputTokens: Float64Ptr(12000),
				CostUSD:      Float64Ptr(0.35),
				Requests:     Float64Ptr(30),
				Window:       "1d",
			},
			{
				RawModelID:   "google/gemini-1.5-pro-002",
				InputTokens:  Float64Ptr(50000),
				OutputTokens: Float64Ptr(8000),
				CostUSD:      Float64Ptr(0.12),
				Requests:     Float64Ptr(20),
				Window:       "1d",
			},
			{
				RawModelID:   "openai/o1-mini-2024-09-12",
				InputTokens:  Float64Ptr(40000),
				OutputTokens: Float64Ptr(15000),
				CostUSD:      Float64Ptr(0.55),
				Requests:     Float64Ptr(15),
				Window:       "1d",
			},
			{
				RawModelID:   "deepseek/deepseek-chat",
				InputTokens:  Float64Ptr(60000),
				OutputTokens: Float64Ptr(9000),
				CostUSD:      Float64Ptr(0.20),
				Requests:     Float64Ptr(25),
				Window:       "1d",
			},
		},
		DailySeries: map[string][]TimePoint{
			"analytics_cost":   {{Date: "2026-02-20", Value: 1.20}, {Date: "2026-02-21", Value: 1.50}},
			"analytics_tokens": {{Date: "2026-02-20", Value: 150000}, {Date: "2026-02-21", Value: 200000}},
			"usage_model_gpt5": {{Date: "2026-02-20", Value: 50000}, {Date: "2026-02-21", Value: 75000}},
		},
		Raw: map[string]string{
			"org_name":      "Engineering",
			"tier":          "scale",
			"sync_status":   "ok",
			"warning_limit": "approaching 80%",
		},
	}
}

func makeBenchLargeSnapshot() UsageSnapshot {
	modelNames := []string{
		"claude-3-5-sonnet-20241022", "claude-3-5-haiku-20241022", "claude-3-opus-20240229",
		"gpt-4o-2024-08-06", "gpt-4o-mini-2024-07-18", "o1-preview-2024-09-12", "o1-mini-2024-09-12",
		"gemini-1.5-pro-002", "gemini-1.5-flash-002", "gemini-2.0-flash-exp",
		"deepseek-chat", "deepseek-reasoner", "mistral-large-2407", "codestral-2405",
		"meta-llama/llama-3.1-405b-instruct", "meta-llama/llama-3.1-70b-instruct",
		"qwen/qwen-2.5-coder-32b-instruct", "x-ai/grok-beta",
	}

	records := make([]ModelUsageRecord, 0, len(modelNames)*2)
	metrics := make(map[string]Metric, len(modelNames)*4+10)
	dailySeries := make(map[string][]TimePoint, len(modelNames)+5)
	raw := make(map[string]string, 20)

	for i, name := range modelNames {
		inp := float64((i + 1) * 10000)
		out := float64((i + 1) * 2000)
		cost := float64(i+1) * 0.05
		req := float64((i + 1) * 5)
		records = append(records, ModelUsageRecord{
			RawModelID:   name,
			InputTokens:  Float64Ptr(inp),
			OutputTokens: Float64Ptr(out),
			CostUSD:      Float64Ptr(cost),
			Requests:     Float64Ptr(req),
			Window:       "1d",
		})

		sanitized := sanitizeAnalyticsMetricID(name)
		metrics["model_"+sanitized+"_input_tokens"] = Metric{Used: Float64Ptr(inp), Unit: "tokens", Window: "1d"}
		metrics["model_"+sanitized+"_output_tokens"] = Metric{Used: Float64Ptr(out), Unit: "tokens", Window: "1d"}
		metrics["model_"+sanitized+"_cost_usd"] = Metric{Used: Float64Ptr(cost), Unit: "USD", Window: "1d"}
		metrics["model_"+sanitized+"_requests"] = Metric{Used: Float64Ptr(req), Unit: "requests", Window: "1d"}

		dailySeries["tokens_model_"+sanitized] = []TimePoint{
			{Date: "2026-02-19", Value: inp * 0.8},
			{Date: "2026-02-20", Value: inp * 0.9},
			{Date: "2026-02-21", Value: inp},
		}
	}

	metrics["today_api_cost"] = Metric{Used: Float64Ptr(5.50), Unit: "USD", Window: "1d"}
	metrics["7d_api_cost"] = Metric{Used: Float64Ptr(35.00), Unit: "USD", Window: "7d"}
	metrics["all_time_api_cost"] = Metric{Used: Float64Ptr(250.00), Unit: "USD", Window: "all-time"}

	raw["account_type"] = "enterprise"
	raw["rate_tier"] = "tier-5"
	raw["api_status"] = "healthy"

	return UsageSnapshot{
		ProviderID:  "openrouter",
		AccountID:   "acct-large-01",
		Timestamp:   benchFixedTime,
		Metrics:     metrics,
		ModelUsage:  records,
		DailySeries: dailySeries,
		Raw:         raw,
	}
}

func BenchmarkNormalizeUsageSnapshot_FallbackMetrics(b *testing.B) {
	cfg := DefaultModelNormalizationConfig()
	snap := makeBenchFallbackSnapshot()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkSnapshot = NormalizeUsageSnapshotWithConfig(snap, cfg)
	}
}

func BenchmarkNormalizeUsageSnapshot_PrePopulated(b *testing.B) {
	cfg := DefaultModelNormalizationConfig()
	snap := makeBenchPrePopulatedSnapshot()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkSnapshot = NormalizeUsageSnapshotWithConfig(snap, cfg)
	}
}

func BenchmarkNormalizeUsageSnapshot_Large(b *testing.B) {
	cfg := DefaultModelNormalizationConfig()
	snap := makeBenchLargeSnapshot()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkSnapshot = NormalizeUsageSnapshotWithConfig(snap, cfg)
	}
}

func BenchmarkBuildModelUsageFromSnapshotMetrics(b *testing.B) {
	snap := makeBenchFallbackSnapshot()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkRecords = BuildModelUsageFromSnapshotMetrics(snap)
	}
}

func BenchmarkExtractAnalyticsModelUsage_PrePopulated(b *testing.B) {
	snap := makeBenchPrePopulatedSnapshot()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkEntries = ExtractAnalyticsModelUsage(snap)
	}
}

func BenchmarkExtractAnalyticsModelUsage_Fallback(b *testing.B) {
	snap := makeBenchFallbackSnapshot()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkEntries = ExtractAnalyticsModelUsage(snap)
	}
}

func BenchmarkSumAnalyticsModelCost_PrePopulated(b *testing.B) {
	snap := makeBenchPrePopulatedSnapshot()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkCost = sumAnalyticsModelCost(snap)
	}
}

func BenchmarkSumAnalyticsModelCost_Fallback(b *testing.B) {
	snap := makeBenchFallbackSnapshot()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkCost = sumAnalyticsModelCost(snap)
	}
}

func BenchmarkNormalizeModelUsageRecords(b *testing.B) {
	snap := makeBenchPrePopulatedSnapshot()
	cfg := DefaultModelNormalizationConfig()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sinkRecords = normalizeModelUsageRecords(snap, cfg)
	}
}

func BenchmarkNormalizeCanonicalModel(b *testing.B) {
	cfg := DefaultModelNormalizationConfig()
	models := []struct {
		provider string
		model    string
	}{
		{"claude_code", "claude-3-5-sonnet-20241022"},
		{"openai", "gpt-4o-2024-08-06"},
		{"gemini_cli", "gemini-1.5-pro-002"},
		{"openrouter", "deepseek/deepseek-chat"},
		{"cursor", "claude-4.6-opus-high-thinking"},
		{"groq", "meta-llama/llama-3.1-70b-versatile"},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m := models[i%len(models)]
		sinkIdentity = normalizeCanonicalModel(m.provider, m.model, cfg)
	}
}

func BenchmarkNormalizeAnalyticsDailySeries(b *testing.B) {
	snap := makeBenchPrePopulatedSnapshot()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := snap.DeepClone()
		normalizeAnalyticsDailySeries(&s)
		sinkSnapshot = s
	}
}
