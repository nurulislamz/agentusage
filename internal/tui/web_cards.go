package tui

import (
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/nurulislamz/agentusage/internal/config"
	"github.com/nurulislamz/agentusage/internal/core"
)

// WebDetailCard is one TUI detail section as typed rows for the browser.
type WebDetailCard struct {
	ID    string         `json:"id"`
	Title string         `json:"title"`
	Icon  string         `json:"icon,omitempty"`
	Color string         `json:"color,omitempty"`
	Rows  []WebDetailRow `json:"rows"`
}

// WebDetailRow is a heading, gauge, timer, or text/kv line.
type WebDetailRow struct {
	Kind    string   `json:"kind"`
	Label   string   `json:"label,omitempty"`
	Value   string   `json:"value,omitempty"`
	Hint    string   `json:"hint,omitempty"`
	Percent *float64 `json:"percent,omitempty"`
	Tone    string   `json:"tone,omitempty"`
}

var (
	quotaLineRe   = regexp.MustCompile(`(?i)([\d.]+)%\s+(remaining|used)`)
	trailingPctRe = regexp.MustCompile(`[\d.]+\s*%\s*$`)
	barePercentRe = regexp.MustCompile(`^[\d.]+%$`)
	barGlyphsRe   = regexp.MustCompile(`^[█░▒▓▀▄▌▐▏▎▍▋▊▉─━│┃┌┐└┘╭╮╰╯├┤┬┴┼\s]+$`)
)

func projectDetailCards(
	snap core.UsageSnapshot,
	widget core.DashboardWidget,
	w int,
	warnThresh, critThresh float64,
	timeWindow core.TimeWindow,
	hideCosts bool,
	now time.Time,
	usageMode string,
) []WebDetailCard {
	isUsed := usageMode == config.UsageModeUsed
	sections := buildDetailSections(snap, widget, w, warnThresh, critThresh, timeWindow, hideCosts, now, usageMode)
	out := make([]WebDetailCard, 0, len(sections))
	for _, sec := range sections {
		card := WebDetailCard{
			ID:    sec.id,
			Title: sec.title,
			Icon:  sec.icon,
			Color: colorHex(sec.color),
			Rows:  make([]WebDetailRow, 0),
		}
		if card.Icon == "" {
			card.Icon = sectionIcon(sec.title)
		}
		if strings.TrimSpace(card.Color) == "" {
			card.Color = colorHex(sectionColor(sec.title))
		}
		if strings.EqualFold(sec.title, "Timers") {
			card.Rows = projectTimerRows(snap, widget, now)
		} else {
			card.Rows = rowsFromSectionLines(sec.lines, isUsed)
		}
		if len(card.Rows) == 0 {
			continue
		}
		out = append(out, card)
	}
	return out
}

func projectTimerRows(snap core.UsageSnapshot, widget core.DashboardWidget, now time.Time) []WebDetailRow {
	keys := core.SortedStringKeys(snap.Resets)
	if len(keys) == 0 {
		return nil
	}
	rows := make([]WebDetailRow, 0, len(keys))
	for _, key := range keys {
		resetAt := snap.Resets[key]
		remaining := resetAt.Sub(now)
		dateStr := resetAt.Format("Jan 02 15:04")
		row := WebDetailRow{
			Kind:  "timer",
			Label: metricLabel(widget, key),
			Value: dateStr,
			Tone:  "ok",
		}
		if remaining <= 0 {
			row.Hint = "expired"
			row.Tone = "dim"
		} else {
			row.Hint = "in " + formatDuration(remaining)
			switch {
			case remaining < 15*time.Minute:
				row.Tone = "crit"
			case remaining < time.Hour:
				row.Tone = "warn"
			}
		}
		rows = append(rows, row)
	}
	return rows
}

func rowsFromSectionLines(lines []string, isUsedMode bool) []WebDetailRow {
	rows := make([]WebDetailRow, 0, len(lines))
	pendingLabel := ""
	flushText := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" {
			return
		}
		kind := "text"
		if strings.Contains(s, "◈") || looksLikeHeading(s) {
			kind = "heading"
		}
		rows = append(rows, WebDetailRow{Kind: kind, Value: s})
	}

	for _, raw := range lines {
		plain := strings.TrimSpace(StripANSI(raw))
		plain = collapseSpaces(plain)
		if plain == "" || isGaugeBarLine(plain) || isChartArtLine(plain) || (pendingLabel != "" && barePercentRe.MatchString(plain)) {
			continue
		}
		if loc := quotaLineRe.FindStringSubmatch(plain); loc != nil {
			pct, err := strconv.ParseFloat(loc[1], 64)
			if err != nil {
				flushText(plain)
				pendingLabel = ""
				continue
			}
			label := strings.TrimSpace(pendingLabel)
			pendingLabel = ""
			if label == "" {
				label = "Quota"
			}
			p := pct
			hint := strings.TrimSpace(quotaLineRe.ReplaceAllString(plain, ""))
			hint = strings.Trim(hint, "· ")
			rows = append(rows, WebDetailRow{
				Kind:    "gauge",
				Label:   label,
				Hint:    hint,
				Percent: &p,
				Tone:    quotaTone(pct, isUsedMode),
			})
			continue
		}
		if pendingLabel != "" {
			flushText(pendingLabel)
			pendingLabel = ""
		}
		if looksLikeGaugeLabel(plain) {
			pendingLabel = plain
			continue
		}
		flushText(plain)
	}
	if pendingLabel != "" {
		flushText(pendingLabel)
	}
	return rows
}

func isGaugeBarLine(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	stripped := strings.TrimSpace(trailingPctRe.ReplaceAllString(s, ""))
	if stripped == "" {
		return false
	}
	return barGlyphsRe.MatchString(stripped)
}

func isChartArtLine(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	if isGaugeBarLine(s) {
		return true
	}
	braille, box, other := 0, 0, 0
	for _, r := range s {
		if unicode.IsSpace(r) {
			continue
		}
		switch {
		case r >= 0x2800 && r <= 0x28FF:
			braille++
		case r >= 0x2500 && r <= 0x257F, r >= 0x2580 && r <= 0x259F:
			box++
		default:
			other++
		}
	}
	if box >= 2 {
		return true
	}
	if box >= 1 && braille >= 1 {
		return true
	}
	if braille >= 8 {
		return true
	}
	return braille+box > 0 && braille+box >= other
}

func looksLikeGaugeLabel(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	lower := strings.ToLower(s)
	return strings.Contains(lower, "limit remaining") ||
		strings.Contains(lower, "limit used") ||
		(strings.HasSuffix(lower, " remaining") && !strings.Contains(s, "%")) ||
		(strings.HasSuffix(lower, " used") && !strings.Contains(s, "%"))
}

func looksLikeHeading(s string) bool {
	trimmed := strings.TrimLeftFunc(s, func(r rune) bool {
		return r == '◈' || r == '✦' || r == '◇' || unicode.IsSpace(r)
	})
	if trimmed == "" {
		return false
	}
	letters := 0
	upper := 0
	for _, r := range trimmed {
		if unicode.IsLetter(r) {
			letters++
			if unicode.IsUpper(r) {
				upper++
			}
		}
	}
	return letters >= 4 && upper*2 >= letters
}

func collapseSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func quotaTone(percent float64, isUsedMode bool) string {
	if isUsedMode {
		switch {
		case percent >= 90:
			return "crit"
		case percent >= 75:
			return "warn"
		case percent >= 50:
			return "peach"
		default:
			return "ok"
		}
	}
	switch {
	case percent <= 10:
		return "crit"
	case percent <= 25:
		return "warn"
	case percent <= 50:
		return "peach"
	default:
		return "ok"
	}
}

func headerTone(snap core.UsageSnapshot) string {
	switch core.EffectiveStatus(snap) {
	case core.StatusOK:
		return "ok"
	case core.StatusNearLimit:
		return "warn"
	case core.StatusLimited, core.StatusError:
		return "crit"
	case core.StatusAuth:
		return "auth"
	default:
		return "dim"
	}
}
