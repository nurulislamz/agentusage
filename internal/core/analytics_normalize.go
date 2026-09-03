package core

import (
	"slices"
	"strings"
	"time"
)

func normalizeAnalyticsMetrics(s *UsageSnapshot) {
	if s == nil {
		return
	}
	s.EnsureMaps()
	normalizeAnalyticsCostMetrics(s)
	normalizeAnalyticsBreakdownMetrics(s)
}

func normalizeAnalyticsCostMetrics(s *UsageSnapshot) {
	aliasMetricInto(s, "today_api_cost", "today_cost", "daily_cost_usd", "usage_daily")
	aliasMetricInto(s, "7d_api_cost", "7d_cost", "usage_weekly")
	aliasMetricInto(s, "30d_api_cost", "monthly_cost")
	aliasMetricInto(s, "all_time_api_cost", "total_cost_usd", "billing_total_cost", "composer_cost", "cli_cost", "total_cost")

	if _, ok := s.Metrics["window_cost"]; !ok {
		if metric, ok := bestWindowCostMetric(s); ok {
			s.Metrics["window_cost"] = metric
		}
	}
	if _, ok := s.Metrics["window_tokens"]; !ok {
		if total := sumAnalyticsModelTokens(*s); total > 0 {
			s.Metrics["window_tokens"] = Metric{Used: Float64Ptr(total), Unit: "tokens", Window: inferredAnalyticsWindow(*s)}
		}
	}
	if _, ok := s.Metrics["window_requests"]; !ok {
		if total := sumAnalyticsModelRequests(*s); total > 0 {
			s.Metrics["window_requests"] = Metric{Used: Float64Ptr(total), Unit: "requests", Window: inferredAnalyticsWindow(*s)}
		}
	}
}

func normalizeAnalyticsBreakdownMetrics(s *UsageSnapshot) {
	for key, metric := range s.Metrics {
		switch {
		case strings.HasPrefix(key, "model_") && strings.HasSuffix(key, "_cost"):
			aliasMetricKey(s, key, strings.TrimSuffix(key, "_cost")+"_cost_usd", metric)
		case strings.HasPrefix(key, "provider_") && strings.HasSuffix(key, "_cost") && !strings.HasSuffix(key, "_byok_cost"):
			aliasMetricKey(s, key, strings.TrimSuffix(key, "_cost")+"_cost_usd", metric)
		case strings.HasPrefix(key, "provider_") && strings.HasSuffix(key, "_prompt_tokens"):
			aliasMetricKey(s, key, strings.TrimSuffix(key, "_prompt_tokens")+"_input_tokens", metric)
		case strings.HasPrefix(key, "provider_") && strings.HasSuffix(key, "_completion_tokens"):
			aliasMetricKey(s, key, strings.TrimSuffix(key, "_completion_tokens")+"_output_tokens", metric)
		}
	}

	synthesizeSelfProviderBreakdown(s)
}

func aliasMetricInto(s *UsageSnapshot, canonical string, aliases ...string) {
	if s == nil || canonical == "" {
		return
	}
	if _, exists := s.Metrics[canonical]; exists {
		return
	}
	for _, alias := range aliases {
		if metric, ok := s.Metrics[alias]; ok {
			s.Metrics[canonical] = metric
			return
		}
	}
}

func aliasMetricKey(s *UsageSnapshot, source, target string, metric Metric) {
	if s == nil || source == "" || target == "" {
		return
	}
	if _, exists := s.Metrics[target]; exists {
		return
	}
	s.Metrics[target] = metric
}

func bestWindowCostMetric(s *UsageSnapshot) (Metric, bool) {
	if s == nil {
		return Metric{}, false
	}
	for _, key := range []string{
		"window_cost",
		"today_api_cost",
		"7d_api_cost",
		"30d_api_cost",
		"all_time_api_cost",
		"billing_total_cost",
		"composer_cost",
		"total_cost_usd",
		"total_cost",
		"cli_cost",
		"plan_total_spend_usd",
		"individual_spend",
	} {
		if metric, ok := s.Metrics[key]; ok && metric.Used != nil && *metric.Used > 0 {
			return metric, true
		}
	}
	modelCost := sumAnalyticsModelCost(*s)
	if modelCost > 0 {
		return Metric{Used: Float64Ptr(modelCost), Unit: "USD", Window: inferredAnalyticsWindow(*s)}, true
	}
	return Metric{}, false
}

func synthesizeSelfProviderBreakdown(s *UsageSnapshot) {
	if s == nil {
		return
	}
	if hasAnalyticsProviderMetrics(*s) {
		return
	}

	cost := sumAnalyticsModelCost(*s)
	input := 0.0
	output := 0.0
	requests := 0.0
	if len(s.ModelUsage) > 0 {
		for i := range s.ModelUsage {
			rec := &s.ModelUsage[i]
			if rec.InputTokens != nil {
				input += *rec.InputTokens
			}
			if rec.OutputTokens != nil {
				output += *rec.OutputTokens
			}
			if rec.TotalTokens != nil && rec.InputTokens == nil && rec.OutputTokens == nil {
				input += *rec.TotalTokens
			}
			if rec.Requests != nil {
				requests += *rec.Requests
			}
		}
	} else {
		for key, metric := range s.Metrics {
			if metric.Used == nil || *metric.Used <= 0 {
				continue
			}
			switch {
			case (strings.HasPrefix(key, "model_") && strings.HasSuffix(key, "_input_tokens")) || strings.HasPrefix(key, "input_tokens_"):
				input += *metric.Used
			case (strings.HasPrefix(key, "model_") && strings.HasSuffix(key, "_output_tokens")) || strings.HasPrefix(key, "output_tokens_"):
				output += *metric.Used
			case strings.HasPrefix(key, "model_") && strings.HasSuffix(key, "_requests"):
				requests += *metric.Used
			}
		}
		for key, rawVal := range s.Raw {
			switch {
			case (strings.HasPrefix(key, "model_") && strings.HasSuffix(key, "_input_tokens")) || strings.HasPrefix(key, "input_tokens_"):
				if val, ok := parseModelRawValue(rawVal); ok {
					input += val
				}
			case (strings.HasPrefix(key, "model_") && strings.HasSuffix(key, "_output_tokens")) || strings.HasPrefix(key, "output_tokens_"):
				if val, ok := parseModelRawValue(rawVal); ok {
					output += val
				}
			case strings.HasPrefix(key, "model_") && strings.HasSuffix(key, "_requests"):
				if val, ok := parseModelRawValue(rawVal); ok {
					requests += val
				}
			}
		}
	}

	if requests <= 0 {
		if metric, ok := s.Metrics["window_requests"]; ok && metric.Used != nil {
			requests = *metric.Used
		}
	}
	if cost <= 0 && input <= 0 && output <= 0 && requests <= 0 {
		return
	}

	providerKey := sanitizeAnalyticsMetricID(s.ProviderID)
	if providerKey == "" {
		providerKey = "unknown"
	}
	window := inferredAnalyticsWindow(*s)
	if cost > 0 {
		s.Metrics["provider_"+providerKey+"_cost_usd"] = Metric{Used: Float64Ptr(cost), Unit: "USD", Window: window}
	}
	if input > 0 {
		s.Metrics["provider_"+providerKey+"_input_tokens"] = Metric{Used: Float64Ptr(input), Unit: "tokens", Window: window}
	}
	if output > 0 {
		s.Metrics["provider_"+providerKey+"_output_tokens"] = Metric{Used: Float64Ptr(output), Unit: "tokens", Window: window}
	}
	if requests > 0 {
		s.Metrics["provider_"+providerKey+"_requests"] = Metric{Used: Float64Ptr(requests), Unit: "requests", Window: window}
	}
}

func hasAnalyticsProviderMetrics(s UsageSnapshot) bool {
	for key := range s.Metrics {
		if strings.HasPrefix(key, "provider_") {
			return true
		}
	}
	return false
}

func inferredAnalyticsWindow(s UsageSnapshot) string {
	for _, key := range []string{"window_cost", "window_tokens", "window_requests", "today_api_cost", "7d_api_cost", "30d_api_cost", "all_time_api_cost"} {
		if metric, ok := s.Metrics[key]; ok && strings.TrimSpace(metric.Window) != "" {
			return metric.Window
		}
	}
	return "all-time"
}

func sumAnalyticsModelTokens(s UsageSnapshot) float64 {
	if len(s.ModelUsage) > 0 {
		total := 0.0
		for i := range s.ModelUsage {
			rec := &s.ModelUsage[i]
			if rec.InputTokens != nil {
				total += *rec.InputTokens
			}
			if rec.OutputTokens != nil {
				total += *rec.OutputTokens
			}
			if rec.TotalTokens != nil && rec.InputTokens == nil && rec.OutputTokens == nil {
				total += *rec.TotalTokens
			}
		}
		return total
	}
	total := 0.0
	for key, metric := range s.Metrics {
		if metric.Used == nil || *metric.Used <= 0 {
			continue
		}
		if (strings.HasPrefix(key, "model_") && (strings.HasSuffix(key, "_input_tokens") || strings.HasSuffix(key, "_output_tokens"))) ||
			strings.HasPrefix(key, "input_tokens_") || strings.HasPrefix(key, "output_tokens_") {
			total += *metric.Used
		}
	}
	for key, rawVal := range s.Raw {
		if (strings.HasPrefix(key, "model_") && (strings.HasSuffix(key, "_input_tokens") || strings.HasSuffix(key, "_output_tokens"))) ||
			strings.HasPrefix(key, "input_tokens_") || strings.HasPrefix(key, "output_tokens_") {
			if val, ok := parseModelRawValue(rawVal); ok && val > 0 {
				total += val
			}
		}
	}
	return total
}

func sumAnalyticsModelRequests(s UsageSnapshot) float64 {
	total := 0.0
	for _, rec := range s.ModelUsage {
		if rec.Requests != nil {
			total += *rec.Requests
		}
	}
	return total
}

func sanitizeAnalyticsMetricID(raw string) string {
	// Preserve non-separator runes (including non-ASCII model IDs). The ASCII-only
	// scanner previously dropped those characters and collapsed distinct models
	// into the same series key (e.g. "模型甲" and "模型乙" both became "unknown").
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(value))
	lastUnderscore := false
	for _, r := range value {
		switch r {
		case '/', '-', ' ', '.', ':', '_':
			if !lastUnderscore && b.Len() > 0 {
				b.WriteByte('_')
				lastUnderscore = true
			}
		default:
			b.WriteRune(r)
			lastUnderscore = false
		}
	}
	return strings.Trim(b.String(), "_")
}

func normalizeAnalyticsDailySeries(s *UsageSnapshot) {
	if s == nil {
		return
	}
	s.EnsureMaps()
	if s.DailySeries == nil {
		s.DailySeries = make(map[string][]TimePoint)
	}

	normalizeExistingSeriesAliases(s)
	synthesizeCoreSeriesFromMetrics(s)
	synthesizeModelSeriesFromRecords(s)

	for key, points := range s.DailySeries {
		s.DailySeries[key] = normalizeSeriesPoints(points)
	}
}

func normalizeExistingSeriesAliases(s *UsageSnapshot) {
	aliasInto(s, "cost", "analytics_cost", "daily_cost")
	aliasInto(s, "tokens_total", "analytics_tokens", "tokens")
	aliasInto(s, "requests", "analytics_requests")

	type seriesMerge struct {
		target string
		points []TimePoint
	}
	var toMerge []seriesMerge
	for key, points := range s.DailySeries {
		switch {
		case strings.HasPrefix(key, "tokens_model_"):
			model := key[len("tokens_model_"):]
			toMerge = append(toMerge, seriesMerge{target: "tokens_" + model, points: points})
		case strings.HasPrefix(key, "usage_model_"):
			model := key[len("usage_model_"):]
			toMerge = append(toMerge,
				seriesMerge{target: "tokens_model_" + model, points: points},
				seriesMerge{target: "tokens_" + model, points: points},
			)
		}
	}
	for _, tm := range toMerge {
		mergeSeries(s, tm.target, tm.points)
	}
}

func aliasInto(s *UsageSnapshot, canonical string, aliases ...string) {
	if len(s.DailySeries[canonical]) > 0 {
		return
	}
	for _, alias := range aliases {
		if len(s.DailySeries[alias]) > 0 {
			s.DailySeries[canonical] = append([]TimePoint(nil), s.DailySeries[alias]...)
			return
		}
	}
}

func synthesizeCoreSeriesFromMetrics(s *UsageSnapshot) {
	todayDate := analyticsReferenceTime(s).Format("2006-01-02")

	metricUsed := func(keys ...string) float64 {
		for _, k := range keys {
			if m, ok := s.Metrics[k]; ok && m.Used != nil && *m.Used > 0 {
				return *m.Used
			}
		}
		return 0
	}

	cost1 := metricUsed("today_api_cost", "daily_cost_usd", "today_cost", "usage_daily")
	tok1 := metricUsed("analytics_tokens")
	req1 := metricUsed("analytics_requests")

	if len(s.DailySeries["cost"]) == 0 {
		if cost1 > 0 {
			s.DailySeries["cost"] = []TimePoint{{Date: todayDate, Value: cost1}}
		}
	}

	if len(s.DailySeries["tokens_total"]) == 0 {
		if tok1 > 0 {
			s.DailySeries["tokens_total"] = []TimePoint{{Date: todayDate, Value: tok1}}
		}
	}

	if len(s.DailySeries["requests"]) == 0 {
		if req1 > 0 {
			s.DailySeries["requests"] = []TimePoint{{Date: todayDate, Value: req1}}
		}
	}
}

func synthesizeModelSeriesFromRecords(s *UsageSnapshot) {
	if len(s.ModelUsage) == 0 {
		return
	}
	date := analyticsReferenceTime(s).Format("2006-01-02")

	perModel := make(map[string]float64, len(s.ModelUsage))
	for i := range s.ModelUsage {
		rec := &s.ModelUsage[i]
		model := strings.TrimSpace(rec.RawModelID)
		if model == "" {
			model = strings.TrimSpace(rec.CanonicalLineageID)
		}
		if model == "" {
			continue
		}
		total := float64(0)
		if rec.TotalTokens != nil {
			total += *rec.TotalTokens
		} else {
			if rec.InputTokens != nil {
				total += *rec.InputTokens
			}
			if rec.OutputTokens != nil {
				total += *rec.OutputTokens
			}
		}
		if total <= 0 {
			continue
		}
		perModel[normalizeSeriesModelKey(model)] += total
	}

	for model, total := range perModel {
		legacyKey := "tokens_" + model
		canonicalKey := "tokens_model_" + model
		if len(s.DailySeries[canonicalKey]) == 0 {
			s.DailySeries[canonicalKey] = []TimePoint{{Date: date, Value: total}}
		}
		if len(s.DailySeries[legacyKey]) == 0 {
			s.DailySeries[legacyKey] = []TimePoint{{Date: date, Value: total}}
		}
	}
}

func mergeSeries(s *UsageSnapshot, key string, points []TimePoint) {
	if key == "" || len(points) == 0 {
		return
	}
	s.DailySeries[key] = normalizeSeriesPoints(append(s.DailySeries[key], points...))
}

func normalizeSeriesPoints(points []TimePoint) []TimePoint {
	if len(points) == 0 {
		return nil
	}
	if len(points) == 1 {
		d := strings.TrimSpace(points[0].Date)
		if d == "" || points[0].Value <= 0 {
			return nil
		}
		if d == points[0].Date {
			return points
		}
		return []TimePoint{{Date: d, Value: points[0].Value}}
	}

	isCleanSorted := true
	for i := 0; i < len(points); i++ {
		d := strings.TrimSpace(points[i].Date)
		if d == "" || points[i].Value <= 0 || d != points[i].Date {
			isCleanSorted = false
			break
		}
		if i > 0 && points[i-1].Date >= points[i].Date {
			isCleanSorted = false
			break
		}
	}
	if isCleanSorted {
		return points
	}

	agg := make(map[string]float64, len(points))
	for _, p := range points {
		date := strings.TrimSpace(p.Date)
		if date == "" || p.Value <= 0 {
			continue
		}
		agg[date] += p.Value
	}
	if len(agg) == 0 {
		return nil
	}
	keys := make([]string, 0, len(agg))
	for k := range agg {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	out := make([]TimePoint, 0, len(keys))
	for _, k := range keys {
		out = append(out, TimePoint{Date: k, Value: agg[k]})
	}
	return out
}

func normalizeSeriesModelKey(model string) string {
	res := sanitizeAnalyticsMetricID(model)
	if res == "" {
		return "unknown"
	}
	return res
}

func analyticsReferenceTime(s *UsageSnapshot) time.Time {
	if s != nil && !s.Timestamp.IsZero() {
		return s.Timestamp.UTC()
	}
	return time.Now().UTC()
}
