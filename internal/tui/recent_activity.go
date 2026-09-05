package tui

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/nurulislamz/agentusage/internal/core"
)

// RecentActivity captures whether an account/tool had activity in the last hour (< 60 minutes),
// when it occurred, and what percentage was used during that activity.
type RecentActivity struct {
	ActiveRecently bool
	LastActive     time.Time
	TimeAgo        string
	Percent        float64
	HasPercent     bool
}

// ResolveRecentActivity inspects the snapshot's attributes and metrics to determine
// recent activity status, relative time, and usage magnitude.
func ResolveRecentActivity(snap core.UsageSnapshot, now time.Time) RecentActivity {
	if now.IsZero() {
		now = time.Now()
	}

	var act RecentActivity

	// Check timestamp candidates in priority order
	timeCandidates := []string{
		snap.Attributes["last_active_at"],
		snap.Attributes["recent_activity_at"],
		snap.Attributes["telemetry_last_event_at"],
		snap.Raw["last_active_at"],
	}

	for _, tsStr := range timeCandidates {
		tsStr = strings.TrimSpace(tsStr)
		if tsStr == "" {
			continue
		}
		if t, ok := parseFlexibleTime(tsStr); ok {
			act.LastActive = t
			diff := now.Sub(t)
			if diff >= 0 && diff < 60*time.Minute {
				act.ActiveRecently = true
			}
			act.TimeAgo = FormatRecencyDuration(diff)
			break
		}
	}

	// Extract percentage used during recent activity
	pctCandidates := []string{
		snap.Attributes["recent_activity_pct"],
		snap.Attributes["recent_pct"],
		snap.Attributes["recent_activity_percent"],
	}

	for _, pctStr := range pctCandidates {
		pctStr = strings.TrimSpace(pctStr)
		pctStr = strings.TrimSuffix(pctStr, "%")
		if pctStr == "" {
			continue
		}
		if val, err := strconv.ParseFloat(pctStr, 64); err == nil && !math.IsNaN(val) && !math.IsInf(val, 0) && val >= 0 {
			act.Percent = val
			act.HasPercent = true
			break
		}
	}

	if !act.HasPercent {
		metricCandidates := []string{"recent_activity_pct", "recent_pct"}
		for _, mk := range metricCandidates {
			if m, ok := snap.Metrics[mk]; ok && m.Used != nil && *m.Used >= 0 {
				act.Percent = *m.Used
				act.HasPercent = true
				break
			}
		}
	}

	return act
}

// parseFlexibleTime attempts to parse timestamp strings in various formats.
func parseFlexibleTime(s string) (time.Time, bool) {
	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05.999999",
		"2006-01-02",
	}

	for _, layout := range formats {
		if t, err := time.Parse(layout, s); err == nil {
			return t, true
		}
	}

	// Try unix timestamp (seconds or milliseconds)
	if sec, err := strconv.ParseInt(s, 10, 64); err == nil && sec > 0 {
		if sec > 1e11 {
			return time.UnixMilli(sec), true
		}
		return time.Unix(sec, 0), true
	}

	return time.Time{}, false
}

// FormatRecencyDuration turns a time difference into a concise relative string.
func FormatRecencyDuration(d time.Duration) string {
	if d < 0 {
		return "just now"
	}
	if d < 1*time.Minute {
		return "just now"
	}
	if d < 60*time.Minute {
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	}
	days := int(d.Hours() / 24)
	return fmt.Sprintf("%dd ago", days)
}

// RenderActivityDot renders the bright green dot indicator (●) styled with colorGreen (#59D4A0).
func RenderActivityDot() string {
	return greenStyle.Render("●")
}

// RenderRecentActivityMicroBar renders a unicode block micro-bar representing recent activity usage.
func RenderRecentActivityMicroBar(percent float64, width int) string {
	if width <= 0 {
		width = 8
	}
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	filled := int(math.Round(percent / 100.0 * float64(width)))
	if percent > 0 && filled == 0 {
		filled = 1
	}
	if filled > width {
		filled = width
	}

	fill := greenStyle.Render(strings.Repeat("█", filled))
	emptyColor := colorSurface2
	if strings.TrimSpace(string(emptyColor)) == "" {
		emptyColor = colorDim
	}
	empty := lipgloss.NewStyle().Foreground(emptyColor).Render(strings.Repeat("░", width-filled))

	bracketStyle := dimStyle
	return bracketStyle.Render("[") + fill + empty + bracketStyle.Render("]")
}

// RenderRecentActivityLine renders the complete recent activity line with relative time and micro-bar.
func RenderRecentActivityLine(act RecentActivity, barW int) string {
	timePart := fmt.Sprintf("Active %s", act.TimeAgo)
	if !act.HasPercent {
		return timePart
	}

	bar := RenderRecentActivityMicroBar(act.Percent, barW)
	pctPart := lipgloss.NewStyle().Foreground(colorGreen).Bold(true).Render(fmt.Sprintf("%.1f%%", act.Percent))
	return fmt.Sprintf("%-16s %s %s", timePart, bar, pctPart)
}
