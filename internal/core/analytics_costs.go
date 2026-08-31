package core

import "strings"

type AnalyticsCostSummary struct {
	TotalCostUSD float64
	TodayCostUSD float64
	WeekCostUSD  float64
	BurnRateUSD  float64
}

func ExtractAnalyticsCostSummary(s UsageSnapshot) AnalyticsCostSummary {
	metricTotal := firstPositiveMetricUsed(s,
		0,
		"window_cost",
		"total_cost_usd",
		"all_time_api_cost",
		"billing_total_cost",
		"composer_cost",
		"total_cost",
		"cli_cost",
		"plan_total_spend_usd",
		"individual_spend",
		"monthly_cost",
	)
	modelTotal := sumAnalyticsModelCost(s)
	total := metricTotal
	if modelTotal > total {
		total = modelTotal
	}

	return AnalyticsCostSummary{
		TotalCostUSD: total,
		TodayCostUSD: firstPositiveMetricUsed(s,
			0,
			"today_api_cost",
			"daily_cost_usd",
			"today_cost",
			"usage_daily",
		),
		WeekCostUSD: firstPositiveMetricUsed(s,
			0,
			"7d_api_cost",
			"7d_cost",
			"usage_weekly",
		),
		BurnRateUSD: firstPositiveMetricUsed(s,
			0,
			"burn_rate",
		),
	}
}

func sumAnalyticsModelCost(s UsageSnapshot) float64 {
	if len(s.ModelUsage) > 0 {
		total := 0.0
		for i := range s.ModelUsage {
			if s.ModelUsage[i].CostUSD != nil && *s.ModelUsage[i].CostUSD > 0 {
				total += *s.ModelUsage[i].CostUSD
			}
		}
		return total
	}
	total := 0.0
	for key, metric := range s.Metrics {
		if metric.Used != nil && *metric.Used > 0 {
			if strings.HasPrefix(key, "model_") && (strings.HasSuffix(key, "_cost_usd") || strings.HasSuffix(key, "_cost")) {
				total += *metric.Used
			}
		}
	}
	for key, rawVal := range s.Raw {
		if strings.HasPrefix(key, "model_") && (strings.HasSuffix(key, "_cost_usd") || strings.HasSuffix(key, "_cost")) {
			if val, ok := parseModelRawValue(rawVal); ok && val > 0 {
				total += val
			}
		}
	}
	return total
}

func firstPositiveMetricUsed(s UsageSnapshot, fallback float64, keys ...string) float64 {
	if fallback > 0 {
		return fallback
	}
	for _, key := range keys {
		if metric, ok := s.Metrics[key]; ok && metric.Used != nil && *metric.Used > 0 {
			return *metric.Used
		}
	}
	return 0
}
