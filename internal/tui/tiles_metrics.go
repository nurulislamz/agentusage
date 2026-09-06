package tui

import (
	"fmt"
	"math"
	"slices"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/nurulislamz/agentusage/internal/core"
)

type compactMetricRowSpec struct {
	label       string
	keys        []string
	match       func(string, core.Metric) bool
	maxSegments int
}

func buildTileCompactMetricSummaryLines(snap core.UsageSnapshot, widget core.DashboardWidget, innerW int) ([]string, map[string]bool) {
	return buildTileCompactMetricSummaryLinesWithHide(snap, widget, innerW, false)
}

// buildTileCompactMetricSummaryLinesWithHide returns nil, nil because usage metrics
// must always be displayed as bars or graphs rather than compact numeric summaries.
func buildTileCompactMetricSummaryLinesWithHide(snap core.UsageSnapshot, widget core.DashboardWidget, innerW int, hideCosts bool) ([]string, map[string]bool) {
	return nil, nil
}

// isMonetaryMetricKey reports whether a metric key/value would render a dollar
// amount in the TUI. Used by hide-costs filters at every call site that emits
// USD-formatted text.
func isMonetaryMetricKey(key string, met core.Metric) bool {
	if key == "burn_rate" {
		return true
	}
	if met.Unit == "USD" {
		return true
	}
	if strings.HasSuffix(key, "_usd") || strings.HasSuffix(key, "_cost") {
		return true
	}
	if strings.Contains(key, "cost") || strings.Contains(key, "spend") || strings.Contains(key, "price") {
		return true
	}
	return false
}

func collectCompactMetricSegments(spec compactMetricRowSpec, widget core.DashboardWidget, metrics map[string]core.Metric, consumed map[string]bool, hideCosts bool) ([]string, []string) {
	maxSegments := spec.maxSegments
	if maxSegments <= 0 {
		maxSegments = 4
	}

	var segments []string
	var used []string
	seenLabels := map[string]bool{}
	add := func(key string, met core.Metric) {
		if len(segments) >= maxSegments {
			return
		}
		segment := compactMetricSegment(widget, key, met)
		if segment == "" {
			return
		}
		// Deduplicate: if a previous segment already resolved to the same
		// label (e.g. two metrics both showing "7d"), skip the later one.
		resolvedLabel := resolvedCompactLabel(widget, key, met)
		if seenLabels[resolvedLabel] {
			return
		}
		seenLabels[resolvedLabel] = true
		segments = append(segments, segment)
		used = append(used, key)
	}

	for _, key := range spec.keys {
		if len(segments) >= maxSegments {
			break
		}
		if consumed[key] {
			continue
		}
		met, ok := metrics[key]
		if !ok {
			continue
		}
		if hideCosts && isMonetaryMetricKey(key, met) {
			continue
		}
		add(key, met)
	}

	if spec.match != nil && len(segments) < maxSegments {
		keys := core.SortedStringKeys(metrics)
		for _, key := range keys {
			if len(segments) >= maxSegments {
				break
			}
			if consumed[key] || slices.Contains(spec.keys, key) {
				continue
			}
			met := metrics[key]
			if !spec.match(key, met) {
				continue
			}
			if hideCosts && isMonetaryMetricKey(key, met) {
				continue
			}
			add(key, met)
		}
	}

	return segments, used
}

func compactMetricSegment(widget core.DashboardWidget, key string, met core.Metric) string {
	value := compactMetricValue(key, met)
	if value == "" {
		return ""
	}
	label := compactMetricLabel(widget, key)
	// When the metric carries a Window tag, replace any hardcoded time-window
	// prefix in the label with the actual window value so labels stay in sync
	// with the selected time range.
	if w := strings.TrimSpace(met.Window); w != "" && label != "" {
		label = replaceTimePrefix(label, w)
	}
	if label == "" {
		return value
	}
	return label + " " + value
}

// resolvedCompactLabel returns the final label string that compactMetricSegment
// would use for deduplication purposes (without the value part).
func resolvedCompactLabel(widget core.DashboardWidget, key string, met core.Metric) string {
	label := compactMetricLabel(widget, key)
	if w := strings.TrimSpace(met.Window); w != "" && label != "" {
		label = replaceTimePrefix(label, w)
	}
	return strings.ToLower(strings.TrimSpace(label))
}

// replaceTimePrefix swaps a hardcoded time prefix (today, 7d, 30d, all, 1d)
// at the start of a label with the metric's actual window tag.
func replaceTimePrefix(label, window string) string {
	prefixes := []string{"today ", "7d ", "30d ", "1d ", "all "}
	low := strings.ToLower(label)
	for _, p := range prefixes {
		if strings.HasPrefix(low, p) {
			rest := label[len(p):]
			if rest == "" {
				return window
			}
			return window + " " + rest
		}
	}
	// Exact match (label IS just the time tag, no suffix).
	switch strings.ToLower(label) {
	case "today", "7d", "30d", "1d", "all":
		return window
	}
	return label
}

func compactMetricLabel(widget core.DashboardWidget, key string) string {
	if widget.CompactMetricLabelOverrides != nil {
		if label, ok := widget.CompactMetricLabelOverrides[key]; ok && label != "" {
			return label
		}
	}

	if strings.HasPrefix(key, "org_") && strings.HasSuffix(key, "_seats") {
		org := strings.TrimSuffix(strings.TrimPrefix(key, "org_"), "_seats")
		if org != "" {
			return truncateToWidth(org, 8)
		}
		return "seats"
	}

	if strings.HasPrefix(key, "rate_limit_") {
		return strings.TrimPrefix(key, "rate_limit_")
	}

	labels := map[string]string{
		"plan_spend":           "plan",
		"plan_included":        "incl",
		"plan_bonus":           "bonus",
		"spend_limit":          "cap",
		"individual_spend":     "mine",
		"plan_percent_used":    "used",
		"plan_total_spend_usd": "plan",
		"plan_limit_usd":       "limit",
		"credit_balance":       "balance",
		"credits":              "credits",
		"monthly_spend":        "month",
		"context_window":       "ctx",
		"messages_today":       "msgs",
		"sessions_today":       "sess",
		"tool_calls_today":     "tools",
		"chat_quota":           "chat",
		"completions_quota":    "comp",
		"rpm":                  "rpm",
		"tpm":                  "tpm",
		"rpd":                  "rpd",
		"tpd":                  "tpd",
	}
	return labels[key]
}

func compactMetricValue(key string, met core.Metric) string {
	if key == "burn_rate" && met.Used != nil {
		return fmt.Sprintf("%s/h", formatUSD(*met.Used))
	}

	used, hasUsed := metricUsedValue(met)
	isUSD := isTileUSDMetric(key, met)
	isPct := met.Unit == "%"

	if strings.HasPrefix(key, "quota") && met.Remaining != nil {
		if isPct {
			return fmt.Sprintf("%.0f%%", *met.Remaining)
		}
		if isUSD {
			return fmt.Sprintf("%s left", formatUSD(*met.Remaining))
		}
		return fmt.Sprintf("%s left", compactMetricAmount(*met.Remaining, met.Unit))
	}

	if met.Limit != nil {
		if hasUsed {
			if isPct {
				return fmt.Sprintf("%.0f%%", used)
			}
			if isUSD {
				return fmt.Sprintf("%s/%s", formatUSD(used), formatUSD(*met.Limit))
			}
			return fmt.Sprintf("%s/%s", compactMetricAmount(used, met.Unit), compactMetricAmount(*met.Limit, met.Unit))
		}
		if met.Remaining != nil && isPct {
			return fmt.Sprintf("%.0f%%", 100-*met.Remaining)
		}
	}

	if hasUsed {
		if isPct {
			return fmt.Sprintf("%.0f%%", used)
		}
		if isUSD {
			return formatUSD(used)
		}
		return compactMetricAmount(used, met.Unit)
	}

	if met.Remaining != nil {
		if isPct {
			return fmt.Sprintf("%.0f%% left", *met.Remaining)
		}
		if isUSD {
			return fmt.Sprintf("%s left", formatUSD(*met.Remaining))
		}
		return fmt.Sprintf("%s left", compactMetricAmount(*met.Remaining, met.Unit))
	}

	return ""
}

func metricUsedValue(met core.Metric) (float64, bool) {
	if met.Used != nil {
		return *met.Used, true
	}
	if met.Limit != nil && met.Remaining != nil {
		return *met.Limit - *met.Remaining, true
	}
	return 0, false
}

func isTileUSDMetric(key string, met core.Metric) bool {
	return met.Unit == "USD" || strings.HasSuffix(key, "_usd") ||
		strings.Contains(key, "cost") || strings.Contains(key, "spend") ||
		strings.Contains(key, "price")
}

func compactMetricAmount(v float64, unit string) string {
	switch unit {
	case "tokens", "requests", "messages", "completions", "conversations", "seats", "quota", "lines":
		return shortCompact(v)
	case "":
		return shortCompact(v)
	default:
		return fmt.Sprintf("%s %s", shortCompact(v), unit)
	}
}

func (m Model) buildTileMetricLines(snap core.UsageSnapshot, widget core.DashboardWidget, innerW int, skipKeys map[string]bool) []string {
	return m.buildTileMetricLinesWithHide(snap, widget, innerW, skipKeys, false)
}

func RenderMetricErrorLine(label string, innerW int) string {
	errBadge := lipgloss.NewStyle().
		Foreground(colorMantle).
		Background(colorCrit).
		Bold(true).
		Padding(0, 1).
		Render("ERROR")
	errMsg := lipgloss.NewStyle().
		Foreground(colorCrit).
		Bold(true).
		Render(fmt.Sprintf("%s: cannot render as bar or graph", label))
	full := errBadge + " " + errMsg
	if innerW > 0 && lipgloss.Width(full) > innerW {
		return truncateToWidth(full, innerW)
	}
	return full
}

func findDailySeriesForMetric(snap core.UsageSnapshot, key string) []core.TimePoint {
	if snap.DailySeries == nil {
		return nil
	}
	candidates := []string{
		key,
		"usage_" + key,
		"tokens_" + key,
		"usage_model_" + key,
		"usage_client_" + key,
		"usage_source_" + key,
		"usage_project_" + key,
	}
	for _, c := range candidates {
		if pts, ok := snap.DailySeries[c]; ok && len(pts) > 0 {
			return pts
		}
	}
	return nil
}

func (m Model) buildTileMetricLinesWithHide(snap core.UsageSnapshot, widget core.DashboardWidget, innerW int, skipKeys map[string]bool, hideCosts bool) []string {
	if len(snap.Metrics) == 0 {
		return nil
	}

	keys := core.SortedStringKeys(snap.Metrics)

	maxLabel := innerW/2 - 1
	if maxLabel < 8 {
		maxLabel = 8
	}
	if maxLabel > 18 {
		maxLabel = 18
	}

	var lines []string
	for _, key := range keys {
		if skipKeys != nil && skipKeys[key] {
			continue
		}
		if hasAnyPrefix(key, widget.HideMetricPrefixes) || slices.Contains(widget.HideMetricKeys, key) {
			continue
		}
		met := snap.Metrics[key]
		if shouldSuppressMetricLine(widget, key, met, snap.Metrics) {
			continue
		}
		if hideCosts && isMonetaryMetricKey(key, met) {
			continue
		}

		label := metricLabel(widget, key)
		if len(label) > maxLabel {
			label = label[:maxLabel-1] + "…"
		}

		// 1. Can it be rendered as a bar/gauge?
		if metricHasGauge(key, met) {
			usedPct := metricUsedPercent(key, met)
			if usedPct < 0 {
				usedPct = 0
			}
			if usedPct > 100 {
				usedPct = 100
			}
			gaugeW := innerW - maxLabel - 10
			if gaugeW < 6 {
				gaugeW = 6
			}
			var gauge string
			if m.isUsageModeUsed() {
				gauge = RenderUsageGauge(usedPct, gaugeW, m.warnThreshold, m.critThreshold)
			} else {
				remainingPct := 100 - usedPct
				if remainingPct < 0 {
					remainingPct = 0
				}
				if remainingPct > 100 {
					remainingPct = 100
				}
				gauge = RenderGauge(remainingPct, gaugeW, m.warnThreshold, m.critThreshold)
			}
			labelR := lipgloss.NewStyle().Foreground(colorSubtext).Width(maxLabel).Render(label)
			lines = append(lines, labelR+" "+gauge)
			continue
		}

		// 2. Can it be rendered as a graph/sparkline from daily series?
		points := findDailySeriesForMetric(snap, key)
		if len(points) > 1 {
			vals := make([]float64, len(points))
			for i, p := range points {
				vals[i] = math.Max(0, p.Value)
			}
			sparkW := innerW - maxLabel - 4
			if sparkW < 6 {
				sparkW = 6
			}
			spark := RenderSparkline(vals, sparkW, ProviderColor(snap.ProviderID))
			labelR := lipgloss.NewStyle().Foreground(colorSubtext).Width(maxLabel).Render(label)
			lines = append(lines, labelR+" "+spark)
			continue
		}

		// 3. Unable to render as a bar or graph -> big error message!
		lines = append(lines, RenderMetricErrorLine(metricLabel(widget, key), innerW))
	}
	return lines
}

func shouldSuppressMetricLine(widget core.DashboardWidget, key string, met core.Metric, all map[string]core.Metric) bool {
	// Key-level usage on /key is often zero/no-limit even when account has non-zero /credits totals.
	// Hide noisy zero rows and prefer the higher-signal credit_balance summary.
	if widget.HideCreditsWhenBalancePresent && key == "credits" {
		if _, hasBalance := all["credit_balance"]; hasBalance {
			return true
		}
	}

	if slices.Contains(widget.SuppressZeroMetricKeys, key) {
		if met.Used == nil || *met.Used == 0 {
			return true
		}
	}

	if widget.SuppressZeroNonUsageMetrics && met.Used != nil && *met.Used == 0 && met.Limit == nil && met.Remaining == nil {
		return true
	}

	return false
}

func hasAnyPrefix(value string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(value, prefix) {
			return true
		}
	}
	return false
}

func formatTileMetricValue(key string, met core.Metric) string {
	isUSD := met.Unit == "USD" || strings.HasSuffix(key, "_usd") ||
		strings.Contains(key, "cost") || strings.Contains(key, "spend") ||
		strings.Contains(key, "price")
	isPct := met.Unit == "%"

	if met.Limit != nil && met.Used != nil {
		if isUSD {
			return fmt.Sprintf("$%s / $%s", formatNumber(*met.Used), formatNumber(*met.Limit))
		}
		if isPct {
			return fmt.Sprintf("%.0f%%", *met.Used)
		}
		unit := met.Unit
		switch unit {
		case "tokens":
			unit = "tok"
		case "requests":
			unit = "req"
		case "messages":
			unit = "messages"
		}
		if unit != "" {
			return fmt.Sprintf("%s / %s %s", formatNumber(*met.Used), formatNumber(*met.Limit), unit)
		}
		return fmt.Sprintf("%s / %s", formatNumber(*met.Used), formatNumber(*met.Limit))
	}
	if met.Limit != nil && met.Remaining != nil {
		used := *met.Limit - *met.Remaining
		usedPct := used / *met.Limit * 100
		return fmt.Sprintf("%s / %s (%.0f%%)", formatNumber(used), formatNumber(*met.Limit), usedPct)
	}
	if met.Used != nil {
		if isUSD {
			return fmt.Sprintf("$%s", formatNumber(*met.Used))
		}
		if isPct {
			return fmt.Sprintf("%.0f%%", *met.Used)
		}
		unit := met.Unit
		switch unit {
		case "tokens":
			unit = "tok"
		case "requests":
			unit = "req"
		}
		if unit == "" {
			return formatNumber(*met.Used)
		}
		return fmt.Sprintf("%s %s", formatNumber(*met.Used), unit)
	}
	if met.Remaining != nil {
		return fmt.Sprintf("%s avail", formatNumber(*met.Remaining))
	}
	return ""
}

func renderDotLeaderRow(label, value string, totalW int) string {
	labelR := lipgloss.NewStyle().Foreground(colorSubtext).Render(label)
	valueR := lipgloss.NewStyle().Foreground(colorText).Bold(true).Render(value)
	lw := lipgloss.Width(labelR)
	vw := lipgloss.Width(valueR)
	dotsW := totalW - lw - vw - 2
	if dotsW < 1 {
		dotsW = 1
	}
	dots := tileDotLeaderStyle.Render(strings.Repeat("·", dotsW))
	return labelR + " " + dots + " " + valueR
}

func prioritizeMetricKeys(keys, priority []string) []string {
	if len(priority) == 0 || len(keys) == 0 {
		return keys
	}
	keySet := make(map[string]bool, len(keys))
	for _, k := range keys {
		keySet[k] = true
	}
	seen := make(map[string]bool, len(keys))
	ordered := make([]string, 0, len(keys))
	for _, key := range priority {
		if keySet[key] && !seen[key] {
			ordered = append(ordered, key)
			seen[key] = true
		}
	}
	for _, key := range keys {
		if !seen[key] {
			ordered = append(ordered, key)
		}
	}
	return ordered
}

func shortCompact(v float64) string {
	if v >= 1_000_000 {
		return fmt.Sprintf("%.1fM", v/1_000_000)
	}
	if v >= 1_000 {
		return fmt.Sprintf("%.1fk", v/1_000)
	}
	return fmt.Sprintf("%.0f", v)
}

func truncateToWidth(s string, maxW int) string {
	if maxW <= 0 || lipgloss.Width(s) <= maxW {
		return s
	}
	r := []rune(s)
	for len(r) > 0 && lipgloss.Width(string(r)+"…") > maxW {
		r = r[:len(r)-1]
	}
	if len(r) == 0 {
		return "…"
	}
	return string(r) + "…"
}

func intersperse(items []string, sep string) []string {
	if len(items) <= 1 {
		return items
	}
	result := make([]string, 0, len(items)*2-1)
	for i, item := range items {
		if i > 0 {
			result = append(result, sep)
		}
		result = append(result, item)
	}
	return result
}
