package tui

import (
	"strings"
	"time"

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
}

func projectUsageLines(snap core.UsageSnapshot, widget core.DashboardWidget, cards []WebDetailCard, now time.Time) []WebUsageLine {
	timers := projectTimerRows(snap, widget, now)
	used := make([]bool, len(timers))
	card := findUsageCard(cards)

	lines := make([]WebUsageLine, 0, len(card.Rows)+len(timers))
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
			if i, timer := matchTimerRow(row.Label, timers); i >= 0 {
				used[i] = true
				line = applyTimerToUsageLine(line, timer)
			} else if line.ResetIn == "" {
				line.ResetIn = resetInFromHint(row.Hint)
			}
			lines = append(lines, line)
		case "kv":
			if looksLikeUsageKV(row.Label) {
				lines = append(lines, WebUsageLine{
					Label: row.Label,
					Short: shortUsageLabel(row.Label),
					Value: row.Value,
					Hint:  row.Hint,
					Tone:  row.Tone,
				})
			}
		}
	}

	for i, timer := range timers {
		if used[i] {
			continue
		}
		line := WebUsageLine{
			Label: timer.Label,
			Short: shortUsageLabel(timer.Label),
			Tone:  timer.Tone,
		}
		lines = append(lines, applyTimerToUsageLine(line, timer))
	}

	if len(lines) == 0 && strings.TrimSpace(snap.Message) != "" {
		lines = append(lines, WebUsageLine{
			Label: "Status",
			Short: "Status",
			Value: strings.TrimSpace(snap.Message),
			Tone:  "dim",
		})
	}
	return lines
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

func looksLikeUsageKV(label string) bool {
	lower := strings.ToLower(label)
	if strings.Contains(lower, "cost") || strings.Contains(lower, "spend") || strings.Contains(lower, "$") {
		return false
	}
	return strings.Contains(lower, "limit") ||
		strings.Contains(lower, "quota") ||
		strings.Contains(lower, "cap") ||
		strings.Contains(lower, "remaining") ||
		strings.Contains(lower, "used") ||
		strings.Contains(lower, "allowance") ||
		strings.Contains(lower, "window") ||
		strings.Contains(lower, "reset")
}

func matchTimerRow(label string, timers []WebDetailRow) (int, WebDetailRow) {
	want := usageWindowKey(label)
	if want == "" {
		return -1, WebDetailRow{}
	}
	for i, t := range timers {
		if usageWindowKey(t.Label) == want {
			return i, t
		}
	}
	return -1, WebDetailRow{}
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
