package tui

import (
	"math"
	"strings"
	"time"

	"github.com/nurulislamz/agentusage/internal/config"
	"github.com/nurulislamz/agentusage/internal/core"
)

// WebUsageLine is one quota window for the navigator / usage-only boards.
type WebUsageLine struct {
	Label   string   `json:"label"`
	Short   string   `json:"short,omitempty"`
	Percent *float64 `json:"percent,omitempty"`
	Value   string   `json:"value,omitempty"`
	Hint    string   `json:"hint,omitempty"`
	ResetIn string   `json:"reset_in,omitempty"`
	Tone    string   `json:"tone,omitempty"`
	Urgent  bool     `json:"urgent,omitempty"`
	Group   string   `json:"group,omitempty"`
}

func projectUsageLines(snap core.UsageSnapshot, widget core.DashboardWidget, cards []WebDetailCard, now time.Time) []WebUsageLine {
	timers := projectTimerRows(snap, widget, now)
	used := make([]bool, len(timers))
	card := findUsageCard(cards)
	hasGauge := false
	for _, row := range card.Rows {
		if row.Kind == "gauge" {
			hasGauge = true
			break
		}
	}

	lines := make([]WebUsageLine, 0, len(card.Rows)+len(timers))
	claimed := make(map[string]bool)
	for _, row := range card.Rows {
		switch row.Kind {
		case "gauge":
			line := WebUsageLine{
				Label:   row.Label,
				Short:   shortUsageLabel(row.Label),
				Percent: row.Percent,
				Hint:    row.Hint,
				Tone:    row.Tone,
			}
			if i, timer := matchTimerRow(row.Label, timers, used); i >= 0 {
				used[i] = true
				line = applyTimerToUsageLine(line, timer)
			} else if line.ResetIn == "" {
				line.ResetIn = resetInFromHint(row.Hint)
			}
			line.Group = resolveUsageGroup(line, snap, widget, claimed)
			lines = append(lines, line)
		case "kv":
			if skipUsageKV(row.Label, hasGauge) {
				continue
			}
			line := WebUsageLine{
				Label: row.Label,
				Short: shortUsageLabel(row.Label),
				Value: row.Value,
				Hint:  row.Hint,
				Tone:  row.Tone,
			}
			if i, timer := matchTimerRow(row.Label+" "+row.Value, timers, used); i >= 0 {
				used[i] = true
				line = applyTimerToUsageLine(line, timer)
			}
			line.Group = resolveUsageGroup(line, snap, widget, claimed)
			lines = append(lines, line)
		}
	}

	if !hasGauge {
		for i, timer := range timers {
			if used[i] {
				continue
			}
			line := WebUsageLine{
				Label: timer.Label,
				Short: shortUsageLabel(timer.Label),
				Tone:  timer.Tone,
				Group: usageLineGroup(timer.Label, widget),
			}
			lines = append(lines, applyTimerToUsageLine(line, timer))
		}
	}
	return lines
}

func ensureUsageLines(view *WebAccountView, usageMode string) {
	if view == nil || len(view.UsageLines) > 0 {
		return
	}
	isUsed := usageMode == config.UsageModeUsed
	if view.HasGauge {
		pct := view.GaugePercent
		view.UsageLines = []WebUsageLine{{
			Label:   "Usage",
			Short:   "Usage",
			Percent: &pct,
			Value:   view.Summary,
			Tone:    quotaTone(pct, isUsed),
		}}
		return
	}
	if s := strings.TrimSpace(view.Summary); s != "" {
		view.UsageLines = []WebUsageLine{{
			Label: "Usage",
			Short: "Usage",
			Value: s,
			Tone:  "dim",
		}}
	}
}

func findUsageCard(cards []WebDetailCard) WebDetailCard {
	for _, c := range cards {
		if isUsageCardTitle(c.ID, c.Title) {
			return c
		}
	}
	for _, c := range cards {
		for _, r := range c.Rows {
			if r.Kind == "gauge" {
				return c
			}
		}
	}
	return WebDetailCard{}
}

func isUsageCardTitle(id, title string) bool {
	id = strings.ToLower(strings.TrimSpace(id))
	title = strings.ToLower(strings.TrimSpace(title))
	switch id {
	case "usage", "hero", "overview", "quota":
		return true
	}
	return title == "usage"
}

func skipUsageKV(label string, hasGauge bool) bool {
	lower := strings.ToLower(strings.TrimSpace(label))
	switch lower {
	case "tokens", "activity", "models", "tools", "spend", "spending":
		return true
	case "quota", "credits", "subscription":
		return hasGauge
	default:
		return false
	}
}

func matchTimerRow(label string, timers []WebDetailRow, used []bool) (int, WebDetailRow) {
	want := usageWindowKey(label)
	if want == "" {
		return -1, WebDetailRow{}
	}
	for i, t := range timers {
		if i < len(used) && used[i] {
			continue
		}
		if usageWindowKey(t.Label) == want {
			return i, t
		}
	}
	return -1, WebDetailRow{}
}

func usageLineGroup(label string, widget core.DashboardWidget) string {
	lower := strings.ToLower(label)
	for _, row := range widget.CompactRows {
		rowLabel := strings.TrimSpace(row.Label)
		if rowLabel == "" {
			continue
		}
		if strings.Contains(lower, strings.ToLower(rowLabel)) {
			return compactGroupTitle(rowLabel)
		}
	}
	if strings.Contains(lower, "gemini") {
		return "Gemini"
	}
	if strings.Contains(lower, "claude") || strings.Contains(lower, "opus") || strings.Contains(lower, "sonnet") || strings.Contains(lower, "gpt") || strings.Contains(lower, "3p") {
		return "Claude / GPT"
	}
	return ""
}

func resolveUsageGroup(line WebUsageLine, snap core.UsageSnapshot, widget core.DashboardWidget, claimed map[string]bool) string {
	if g := usageLineGroup(line.Label+" "+line.Short+" "+line.Value, widget); g != "" {
		return g
	}
	if line.Percent == nil || len(widget.CompactRows) == 0 {
		return ""
	}
	for _, row := range widget.CompactRows {
		for _, key := range row.Keys {
			if claimed[key] {
				continue
			}
			m, ok := snap.Metrics[key]
			if !ok || m.Remaining == nil {
				continue
			}
			if math.Abs(*m.Remaining-*line.Percent) > 0.05 {
				continue
			}
			claimed[key] = true
			return compactGroupTitle(row.Label)
		}
	}
	return ""
}

func compactGroupTitle(label string) string {
	lower := strings.ToLower(label)
	if strings.Contains(lower, "gemini") {
		return "Gemini"
	}
	if strings.Contains(lower, "claude") || strings.Contains(lower, "opus") || strings.Contains(lower, "gpt") {
		return "Claude / GPT"
	}
	return strings.TrimSpace(label)
}

func applyTimerToUsageLine(line WebUsageLine, timer WebDetailRow) WebUsageLine {
	resetIn := strings.TrimSpace(strings.TrimPrefix(strings.ToLower(timer.Hint), "in "))
	if resetIn == "" && strings.EqualFold(timer.Hint, "expired") {
		resetIn = "expired"
	}
	if resetIn == "" {
		resetIn = resetInFromHint(timer.Hint)
	}
	if line.ResetIn == "" {
		line.ResetIn = resetIn
	}
	if line.Hint == "" {
		line.Hint = timer.Hint
	}
	if timer.Tone != "" && (line.Tone == "" || line.Tone == "ok") {
		switch timer.Tone {
		case "crit", "warn":
			line.Tone = timer.Tone
		}
	}
	line.Urgent = timer.Tone == "crit" || timer.Tone == "warn" || line.Urgent
	return line
}

func resetInFromHint(hint string) string {
	h := strings.TrimSpace(hint)
	lower := strings.ToLower(h)
	for _, prefix := range []string{"resets in ", "reset in ", "in "} {
		if strings.HasPrefix(lower, prefix) {
			return strings.TrimSpace(h[len(prefix):])
		}
	}
	return ""
}

func usageWindowKey(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ReplaceAll(s, "_", " ")
	switch {
	case strings.Contains(s, "five hour"), strings.Contains(s, "5 hour"), strings.Contains(s, "5h"), strings.Contains(s, "rolling"):
		return "5h"
	case strings.Contains(s, "week"):
		return "week"
	case strings.Contains(s, "month"):
		return "month"
	case strings.Contains(s, "session"):
		return "session"
	case strings.Contains(s, "daily"), strings.Contains(s, "today"), strings.HasPrefix(s, "day "):
		return "day"
	default:
		return strings.Join(strings.Fields(s), " ")
	}
}

func shortUsageLabel(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	switch usageWindowKey(s) {
	case "5h":
		return "5h"
	case "week":
		return "Week"
	case "month":
		return "Month"
	case "session":
		return "Session"
	case "day":
		return "Day"
	}
	lower := strings.ToLower(s)
	for _, suffix := range []string{
		" limit remaining", " limit used", " remaining", " used", " limit",
	} {
		if strings.HasSuffix(lower, suffix) {
			s = strings.TrimSpace(s[:len(s)-len(suffix)])
			break
		}
	}
	if len(s) > 14 {
		return s[:13] + "…"
	}
	return s
}

func nextResetFromLines(lines []WebUsageLine, pills []WebResetPill) string {
	best := ""
	for _, line := range lines {
		if line.ResetIn == "" || strings.EqualFold(line.ResetIn, "expired") {
			continue
		}
		if best == "" || resetDurationRank(line.ResetIn) < resetDurationRank(best) {
			best = line.ResetIn
		}
	}
	if best != "" {
		return best
	}
	for _, p := range pills {
		if p.Duration == "" {
			continue
		}
		if best == "" || resetDurationRank(p.Duration) < resetDurationRank(best) {
			best = p.Duration
		}
	}
	return best
}

func resetDurationRank(s string) time.Duration {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "")
	var d time.Duration
	n := 0
	unit := byte(0)
	flush := func() {
		if n == 0 || unit == 0 {
			n = 0
			unit = 0
			return
		}
		switch unit {
		case 'd':
			d += time.Duration(n) * 24 * time.Hour
		case 'h':
			d += time.Duration(n) * time.Hour
		case 'm':
			d += time.Duration(n) * time.Minute
		case 's':
			d += time.Duration(n) * time.Second
		}
		n = 0
		unit = 0
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
			continue
		}
		if c == 'd' || c == 'h' || c == 'm' || c == 's' {
			unit = c
			flush()
			continue
		}
		n = 0
		unit = 0
	}
	if d == 0 {
		return time.Hour * 24 * 365
	}
	return d
}
