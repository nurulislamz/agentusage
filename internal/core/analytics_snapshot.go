package core

import (
	"slices"
	"strings"

	"github.com/samber/lo"
)

type AnalyticsModelUsageEntry struct {
	Name         string
	CostUSD      float64
	InputTokens  float64
	OutputTokens float64
	Confidence   float64
	Window       string
}

type NamedSeries struct {
	Name   string
	Points []TimePoint
}

func ExtractAnalyticsModelUsage(s UsageSnapshot) []AnalyticsModelUsageEntry {
	records := s.ModelUsage
	if len(records) == 0 {
		records = BuildModelUsageFromSnapshotMetrics(s)
	}
	if len(records) == 0 {
		return nil
	}

	type agg struct {
		name       string
		cost       float64
		input      float64
		output     float64
		confidence float64
		window     string
	}

	byModel := make(map[string]int, len(records))
	items := make([]agg, 0, len(records))

	for _, rec := range records {
		name := analyticsModelDisplayName(rec)
		if name == "" {
			continue
		}
		idx, ok := byModel[name]
		if !ok {
			idx = len(items)
			byModel[name] = idx
			items = append(items, agg{name: name, window: rec.Window})
		}
		entry := &items[idx]
		if rec.CostUSD != nil && *rec.CostUSD > 0 {
			entry.cost += *rec.CostUSD
		}
		if rec.InputTokens != nil {
			entry.input += *rec.InputTokens
		}
		if rec.OutputTokens != nil {
			entry.output += *rec.OutputTokens
		}
		if rec.TotalTokens != nil && rec.InputTokens == nil && rec.OutputTokens == nil {
			entry.input += *rec.TotalTokens
		}
		if rec.Confidence > entry.confidence {
			entry.confidence = rec.Confidence
		}
		if entry.window == "" {
			entry.window = rec.Window
		}
	}

	out := make([]AnalyticsModelUsageEntry, 0, len(items))
	for i := range items {
		e := &items[i]
		if e.cost <= 0 && e.input <= 0 && e.output <= 0 {
			continue
		}
		out = append(out, AnalyticsModelUsageEntry{
			Name:         e.name,
			CostUSD:      e.cost,
			InputTokens:  e.input,
			OutputTokens: e.output,
			Confidence:   e.confidence,
			Window:       e.window,
		})
	}
	slices.SortFunc(out, func(a, b AnalyticsModelUsageEntry) int {
		ta := a.InputTokens + a.OutputTokens
		tb := b.InputTokens + b.OutputTokens
		if ta != tb {
			if ta > tb {
				return -1
			}
			return 1
		}
		if a.CostUSD != b.CostUSD {
			if a.CostUSD > b.CostUSD {
				return -1
			}
			return 1
		}
		return strings.Compare(a.Name, b.Name)
	})
	return out
}

func ExtractAnalyticsModelSeries(series map[string][]TimePoint) []NamedSeries {
	keys := analyticsModelSeriesKeys(series)

	out := make([]NamedSeries, 0, len(keys))
	for _, key := range keys {
		name := strings.TrimPrefix(key, "tokens_model_")
		if name == key {
			name = strings.TrimPrefix(key, "usage_model_")
		}
		if name == key {
			name = strings.TrimPrefix(key, "tokens_")
		}
		if name == "" || len(series[key]) == 0 {
			continue
		}
		out = append(out, NamedSeries{Name: name, Points: series[key]})
	}
	return out
}

func SelectAnalyticsWeightSeries(series map[string][]TimePoint) []TimePoint {
	for _, key := range []string{
		"tokens_total",
		"messages",
		"sessions",
		"tool_calls",
		"requests",
		"tab_accepted",
		"composer_accepted",
	} {
		if pts := series[key]; len(pts) > 0 {
			return pts
		}
	}
	for _, named := range ExtractAnalyticsModelSeries(series) {
		if len(named.Points) > 0 {
			return named.Points
		}
	}
	keys := lo.Filter(SortedStringKeys(series), func(key string, _ int) bool {
		return strings.HasPrefix(key, "usage_client_")
	})
	for _, key := range keys {
		if len(series[key]) > 0 {
			return series[key]
		}
	}
	return nil
}

func hasAnalyticsTokenSeries(series map[string][]TimePoint) bool {
	for key, points := range series {
		if strings.HasPrefix(key, "tokens_model_") && len(points) > 0 {
			return true
		}
	}
	return false
}

func analyticsModelSeriesKeys(series map[string][]TimePoint) []string {
	hasCanonicalTokenSeries := hasAnalyticsTokenSeries(series)
	hasUsageSeries := false
	for key, points := range series {
		if strings.HasPrefix(key, "usage_model_") && len(points) > 0 {
			hasUsageSeries = true
			break
		}
	}

	keys := lo.Filter(SortedStringKeys(series), func(key string, _ int) bool {
		switch {
		case strings.HasPrefix(key, "tokens_model_"):
			return true
		case strings.HasPrefix(key, "usage_model_"):
			return !hasCanonicalTokenSeries
		case strings.HasPrefix(key, "tokens_"):
			if hasCanonicalTokenSeries || hasUsageSeries {
				return false
			}
			return key != "tokens_total" && !strings.HasPrefix(key, "tokens_client_")
		default:
			return false
		}
	})
	return keys
}

func analyticsModelDisplayName(rec ModelUsageRecord) string {
	if rec.Dimensions != nil {
		if groupID := strings.TrimSpace(rec.Dimensions["canonical_group_id"]); groupID != "" {
			return groupID
		}
	}
	if raw := strings.TrimSpace(rec.RawModelID); raw != "" {
		return raw
	}
	if canonical := strings.TrimSpace(rec.CanonicalLineageID); canonical != "" {
		return canonical
	}
	return "unknown"
}
