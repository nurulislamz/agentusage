package core

import (
	"slices"
	"strconv"
	"strings"
)

type modelMetricKind string

const (
	modelMetricInput      modelMetricKind = "input"
	modelMetricOutput     modelMetricKind = "output"
	modelMetricCached     modelMetricKind = "cached"
	modelMetricCacheRead  modelMetricKind = "cache_read"
	modelMetricCacheWrite modelMetricKind = "cache_write"
	modelMetricReasoning  modelMetricKind = "reasoning"
	modelMetricCostUSD    modelMetricKind = "cost_usd"
	modelMetricRequests   modelMetricKind = "requests"
)

type modelWindowKey struct {
	model  string
	window string
}

type modelUsageAccumulator struct {
	rawModelID      string
	window          string
	inputTokens     float64
	hasInput        bool
	outputTokens    float64
	hasOutput       bool
	cachedTokens    float64
	hasCached       bool
	reasoningTokens float64
	hasReasoning    bool
	costUSD         float64
	hasCost         bool
	requests        float64
	hasRequests     bool
}

func BuildModelUsageFromSnapshotMetrics(s UsageSnapshot) []ModelUsageRecord {
	if len(s.Metrics) == 0 && len(s.Raw) == 0 {
		return nil
	}

	records := make(map[modelWindowKey]int)
	items := make([]modelUsageAccumulator, 0, len(s.Metrics))

	ensure := func(rawModelID, window string) *modelUsageAccumulator {
		rawModelID = strings.TrimSpace(rawModelID)
		window = strings.TrimSpace(window)
		if rawModelID == "" {
			rawModelID = "unknown"
		}
		if window == "" {
			window = "unknown"
		}
		key := modelWindowKey{model: rawModelID, window: window}
		if idx, ok := records[key]; ok {
			return &items[idx]
		}
		idx := len(items)
		records[key] = idx
		items = append(items, modelUsageAccumulator{
			rawModelID: rawModelID,
			window:     window,
		})
		return &items[idx]
	}

	for key, metric := range s.Metrics {
		if metric.Used == nil {
			continue
		}
		rawModel, kind, ok := parseModelMetricKey(key)
		if !ok {
			continue
		}
		acc := ensure(rawModel, metric.Window)
		applyModelMetricAcc(acc, kind, *metric.Used)
	}

	for key, rawValue := range s.Raw {
		rawModel, kind, ok := parseModelMetricKey(key)
		if !ok {
			continue
		}
		val, ok := parseModelRawValue(rawValue)
		if !ok {
			continue
		}
		acc := ensure(rawModel, "unknown")
		applyModelMetricAcc(acc, kind, val)
	}

	if len(items) == 0 {
		return nil
	}

	slices.SortFunc(items, func(a, b modelUsageAccumulator) int {
		if a.rawModelID != b.rawModelID {
			return strings.Compare(a.rawModelID, b.rawModelID)
		}
		return strings.Compare(a.window, b.window)
	})

	out := make([]ModelUsageRecord, 0, len(items))
	for i := range items {
		acc := &items[i]
		var (
			inputPtr     *float64
			outputPtr    *float64
			cachedPtr    *float64
			reasoningPtr *float64
			costPtr      *float64
			requestsPtr  *float64
			totalPtr     *float64
		)
		hasAny := false
		if acc.hasInput {
			inputPtr = Float64Ptr(acc.inputTokens)
			hasAny = true
		}
		if acc.hasOutput {
			outputPtr = Float64Ptr(acc.outputTokens)
			hasAny = true
		}
		if acc.hasCached {
			cachedPtr = Float64Ptr(acc.cachedTokens)
			hasAny = true
		}
		if acc.hasReasoning {
			reasoningPtr = Float64Ptr(acc.reasoningTokens)
			hasAny = true
		}
		if acc.hasCost {
			costPtr = Float64Ptr(acc.costUSD)
			hasAny = true
		}
		if acc.hasRequests {
			requestsPtr = Float64Ptr(acc.requests)
			hasAny = true
		}
		if acc.hasInput || acc.hasOutput {
			total := acc.inputTokens + acc.outputTokens
			totalPtr = Float64Ptr(total)
			hasAny = true
		}

		if !hasAny {
			continue
		}

		dims := make(map[string]string, 2)
		if s.ProviderID != "" {
			dims["provider_id"] = s.ProviderID
		}
		if s.AccountID != "" {
			dims["account_id"] = s.AccountID
		}

		out = append(out, ModelUsageRecord{
			RawModelID:      acc.rawModelID,
			RawSource:       "metrics_fallback",
			Window:          acc.window,
			Dimensions:      dims,
			InputTokens:     inputPtr,
			OutputTokens:    outputPtr,
			CachedTokens:    cachedPtr,
			ReasoningTokens: reasoningPtr,
			TotalTokens:     totalPtr,
			CostUSD:         costPtr,
			Requests:        requestsPtr,
		})
	}

	return out
}

func parseModelMetricKey(key string) (rawModelID string, kind modelMetricKind, ok bool) {
	if strings.HasPrefix(key, "model_") {
		inner := key[6:]
		switch {
		case strings.HasSuffix(inner, "_input_tokens"):
			return inner[:len(inner)-len("_input_tokens")], modelMetricInput, true
		case strings.HasSuffix(inner, "_output_tokens"):
			return inner[:len(inner)-len("_output_tokens")], modelMetricOutput, true
		case strings.HasSuffix(inner, "_cached_tokens"):
			return inner[:len(inner)-len("_cached_tokens")], modelMetricCached, true
		case strings.HasSuffix(inner, "_cache_read_tokens"):
			return inner[:len(inner)-len("_cache_read_tokens")], modelMetricCacheRead, true
		case strings.HasSuffix(inner, "_cache_write_tokens"):
			return inner[:len(inner)-len("_cache_write_tokens")], modelMetricCacheWrite, true
		case strings.HasSuffix(inner, "_reasoning_tokens"):
			return inner[:len(inner)-len("_reasoning_tokens")], modelMetricReasoning, true
		case strings.HasSuffix(inner, "_cost_usd"):
			return inner[:len(inner)-len("_cost_usd")], modelMetricCostUSD, true
		case strings.HasSuffix(inner, "_cost"):
			return inner[:len(inner)-len("_cost")], modelMetricCostUSD, true
		case strings.HasSuffix(inner, "_requests"):
			return inner[:len(inner)-len("_requests")], modelMetricRequests, true
		default:
			return "", "", false
		}
	}
	if strings.HasPrefix(key, "input_tokens_") {
		return key[len("input_tokens_"):], modelMetricInput, true
	}
	if strings.HasPrefix(key, "output_tokens_") {
		return key[len("output_tokens_"):], modelMetricOutput, true
	}
	return "", "", false
}

func parseModelRawValue(raw string) (float64, bool) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return 0, false
	}
	start := 0
	for start < len(s) && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	end := start
	hasComma := false
	for end < len(s) && s[end] != ' ' && s[end] != '\t' && s[end] != '\n' && s[end] != '\r' {
		if s[end] == ',' {
			hasComma = true
		}
		end++
	}
	if start >= end {
		return 0, false
	}
	token := s[start:end]
	if hasComma {
		token = strings.ReplaceAll(token, ",", "")
	}
	v, err := strconv.ParseFloat(token, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

func applyModelMetricAcc(acc *modelUsageAccumulator, kind modelMetricKind, value float64) {
	if acc == nil || value <= 0 {
		return
	}
	switch kind {
	case modelMetricInput:
		acc.inputTokens += value
		acc.hasInput = true
	case modelMetricOutput:
		acc.outputTokens += value
		acc.hasOutput = true
	case modelMetricCached, modelMetricCacheRead, modelMetricCacheWrite:
		acc.cachedTokens += value
		acc.hasCached = true
	case modelMetricReasoning:
		acc.reasoningTokens += value
		acc.hasReasoning = true
	case modelMetricCostUSD:
		acc.costUSD += value
		acc.hasCost = true
	case modelMetricRequests:
		acc.requests += value
		acc.hasRequests = true
	}
}
