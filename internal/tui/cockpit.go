package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/nurulislamz/agentusage/internal/config"
	"github.com/nurulislamz/agentusage/internal/core"
)

func (m Model) renderCockpit(snap core.UsageSnapshot, w int) string {
	hideCosts := m.resolveHideCosts(snap)
	return RenderCockpit(snap, m.viewNow(), w, m.warnThreshold, m.critThreshold, m.timeWindow, hideCosts, m.usageMode)
}

// RenderCockpit renders the deep inspector cockpit layout matching the web UI:
// - Hero: status icon + account_id (left) ... provider_name · detail + status badge pill (right)
// - Subhero: summary remaining · cycle schedule (left) ... Last refreshed X ago (right)
// - Status accent hairline
// - '⚡ USAGE & QUOTAS' card with grouped linear gauges, bars, and reset hints
// - '⏱ TIMERS & SCHEDULE' card with urgency dots, labels, formatted timestamps and duration hints
// - '📈 ACTIVITY & TREND' card with stats table / sparkline when telemetry points exist
// - Any extra detail cards (e.g. Info)
func RenderCockpit(
	snap core.UsageSnapshot,
	now time.Time,
	w int,
	warnThresh, critThresh float64,
	timeWindow core.TimeWindow,
	hideCosts bool,
	usageMode string,
) string {
	if w < 20 {
		w = 20
	}

	pColor := providerThemeColor(snap.ProviderID)
	effStatus := core.EffectiveStatus(snap)
	statusCol := StatusColor(effStatus)
	statusIco := StatusIcon(effStatus)

	widget := dashboardWidget(snap.ProviderID)
	di := computeDisplayInfo(snap, widget, hideCosts, usageMode)

	// 1. Hero: status icon + account_id (left) ... provider_name · detail + status badge pill (right)
	iconStr := lipgloss.NewStyle().Foreground(statusCol).Render(statusIco)
	nameStr := lipgloss.NewStyle().Bold(true).Foreground(colorText).Render(snap.AccountID)
	heroLeft := fmt.Sprintf("%s %s", iconStr, nameStr)

	metaParts := []string{snap.ProviderID}
	if di.detail != "" {
		metaParts = append(metaParts, di.detail)
	}
	metaStr := dimStyle.Render(strings.Join(metaParts, " · "))
	badge := SnapshotStatusBadge(snap)
	heroRight := metaStr + "  " + badge

	gap := w - lipgloss.Width(heroLeft) - lipgloss.Width(heroRight)
	var heroLine string
	if gap >= 1 {
		heroLine = heroLeft + strings.Repeat(" ", gap) + heroRight
	} else {
		heroLine = heroLeft + "  " + heroRight
	}

	// 2. Subhero: summary remaining · cycle schedule (left) ... Last refreshed X ago (right)
	summaryText := di.summary
	if summaryText != "" && strings.HasSuffix(strings.TrimSpace(summaryText), "%") {
		summaryText += " remaining"
	}
	subheroLeftParts := []string{}
	if summaryText != "" {
		subheroLeftParts = append(subheroLeftParts, lipgloss.NewStyle().Bold(true).Foreground(colorText).Render(summaryText))
	}
	if sched := formatCycleResetSchedule(snap, now); sched != "" {
		subheroLeftParts = append(subheroLeftParts, dimStyle.Render(sched))
	}
	subheroLeft := strings.Join(subheroLeftParts, " · ")

	subheroRight := dimStyle.Render(formatLastRefreshed(snap.Timestamp, now))
	subGap := w - lipgloss.Width(subheroLeft) - lipgloss.Width(subheroRight)
	var subheroLine string
	if subGap >= 1 {
		subheroLine = subheroLeft + strings.Repeat(" ", subGap) + subheroRight
	} else {
		subheroLine = subheroLeft + "  " + subheroRight
	}

	// 3. Status accent hairline
	hairline := lipgloss.NewStyle().Foreground(statusCol).Render(strings.Repeat("━", w))

	var sections []string
	sections = append(sections, heroLine, subheroLine, hairline)

	// 4. '⚡ USAGE & QUOTAS' card
	cards := projectDetailCards(snap, widget, w, warnThresh, critThresh, timeWindow, hideCosts, now, usageMode)
	lines := projectUsageLines(snap, widget, cards, now)
	if len(lines) == 0 && di.gaugePercent >= 0 {
		pct := di.gaugePercent
		lines = []WebUsageLine{{
			Label:   "Usage",
			Short:   "Usage",
			Percent: &pct,
			Value:   di.summary,
			Tone:    quotaTone(pct, usageMode == config.UsageModeUsed),
		}}
	}

	var quotaLines []string
	quotaLines = append(quotaLines, lipgloss.NewStyle().Bold(true).Foreground(colorBlue).Render("⚡ USAGE & QUOTAS"))
	quotaLines = append(quotaLines, surface1Style.Render(strings.Repeat("─", w)))

	if len(lines) > 0 {
		currentGroup := ""
		for _, l := range lines {
			if l.Group != "" && l.Group != currentGroup {
				currentGroup = l.Group
				quotaLines = append(quotaLines, lipgloss.NewStyle().Bold(true).Foreground(pColor).Render("  ◈ "+strings.ToUpper(currentGroup)))
			}

			label := l.Label
			if label == "" {
				label = l.Short
			}
			labelPadded := padRight(label, 24)

			if l.Percent != nil {
				pct := *l.Percent
				toneCol := toneLipglossColor(l.Tone)
				barW := clamp(w/4, 8, 20)
				bar := RenderGauge(pct, barW, warnThresh, critThresh)
				pctCaption := fmt.Sprintf("%5.2f%%", pct)
				if usageMode == config.UsageModeUsed {
					pctCaption = fmt.Sprintf("%.2f%% used", pct)
				}
				pctStr := lipgloss.NewStyle().Foreground(toneCol).Bold(true).Render(pctCaption)
				hint := ""
				if l.ResetIn != "" {
					hint = dimStyle.Render(" · ⏱ in " + l.ResetIn)
				} else if l.Hint != "" {
					hint = dimStyle.Render(" · " + l.Hint)
				}
				quotaLines = append(quotaLines, fmt.Sprintf("  %s %s %s%s", labelPadded, bar, pctStr, hint))
			} else {
				val := l.Value
				if val == "" {
					val = "—"
				}
				quotaLines = append(quotaLines, fmt.Sprintf("  %s %s", labelPadded, lipgloss.NewStyle().Foreground(colorText).Render(val)))
			}
		}
	} else if summaryText != "" {
		quotaLines = append(quotaLines, "  "+summaryText)
	} else {
		quotaLines = append(quotaLines, dimStyle.Render("  No quota data available"))
	}
	sections = append(sections, strings.Join(quotaLines, "\n"))

	// 5. '⏱ TIMERS & SCHEDULE' card
	timers := projectTimerRows(snap, widget, now)
	var timerLines []string
	timerLines = append(timerLines, lipgloss.NewStyle().Bold(true).Foreground(colorTeal).Render("⏱ TIMERS & SCHEDULE"))
	timerLines = append(timerLines, surface1Style.Render(strings.Repeat("─", w)))

	if len(timers) > 0 {
		for _, t := range timers {
			dotCol := toneLipglossColor(t.Tone)
			dot := lipgloss.NewStyle().Foreground(dotCol).Render("●")
			lbl := padRight(t.Label, 20)
			val := lipgloss.NewStyle().Foreground(colorText).Render(t.Value)
			hint := ""
			if t.Hint != "" {
				hint = dimStyle.Render(" (" + t.Hint + ")")
			}
			timerLines = append(timerLines, fmt.Sprintf("  %s %s %s%s", dot, lbl, val, hint))
		}
	} else {
		next := nextResetFromLines(lines, nil)
		if next != "" {
			dot := lipgloss.NewStyle().Foreground(colorGreen).Render("●")
			lbl := padRight("Next Reset", 20)
			timerLines = append(timerLines, fmt.Sprintf("  %s %s in %s", dot, lbl, next))
		} else {
			timerLines = append(timerLines, dimStyle.Render("  No upcoming reset timers"))
		}
	}
	sections = append(sections, strings.Join(timerLines, "\n"))

	// 6. '📈 ACTIVITY & TREND' card
	dailyPoints := firstDailySeries(snap, "cost", "analytics_cost", "tokens", "requests")
	if len(dailyPoints) > 0 {
		var actLines []string
		actLines = append(actLines, lipgloss.NewStyle().Bold(true).Foreground(colorPeach).Render("📈 ACTIVITY & TREND"))
		actLines = append(actLines, surface1Style.Render(strings.Repeat("─", w)))

		vals := make([]float64, len(dailyPoints))
		sum := 0.0
		for i, p := range dailyPoints {
			v := p.Value
			if v < 0 {
				v = 0
			}
			vals[i] = v
			sum += v
		}

		sparkW := clamp(w/3, 8, 28)
		spark := RenderSparkline(vals, sparkW, pColor)

		last := dailyPoints[len(dailyPoints)-1]
		stats := []string{fmt.Sprintf("Today: %s", shortCompact(last.Value))}
		if len(dailyPoints) > 1 {
			prev := dailyPoints[len(dailyPoints)-2]
			stats = append(stats, fmt.Sprintf("Yesterday: %s", shortCompact(prev.Value)))
		}
		stats = append(stats, fmt.Sprintf("%dd Total: %s", len(dailyPoints), shortCompact(sum)))
		statsLine := dimStyle.Render(strings.Join(stats, "   "))

		if spark != "" {
			actLines = append(actLines, "  "+spark+"   "+statsLine)
		} else {
			actLines = append(actLines, "  "+statsLine)
		}
		sections = append(sections, strings.Join(actLines, "\n"))
	}

	// 7. Any extra detail cards (e.g. Info, Attributes, etc.)
	for _, c := range cards {
		idLower := strings.ToLower(c.ID)
		titleLower := strings.ToLower(c.Title)
		if isCockpitBuiltinSection(idLower, titleLower) {
			continue
		}
		var extraLines []string
		extraLines = append(extraLines, lipgloss.NewStyle().Bold(true).Foreground(colorLavender).Render(c.Title))
		extraLines = append(extraLines, surface1Style.Render(strings.Repeat("─", w)))
		for _, r := range c.Rows {
			if r.Kind == "heading" {
				extraLines = append(extraLines, lipgloss.NewStyle().Bold(true).Render("  "+r.Value))
			} else if r.Label != "" {
				extraLines = append(extraLines, fmt.Sprintf("  %s %s", padRight(r.Label, 20), r.Value))
			} else if r.Value != "" {
				extraLines = append(extraLines, "  "+r.Value)
			}
		}
		if len(extraLines) > 2 {
			sections = append(sections, strings.Join(extraLines, "\n"))
		}
	}

	return strings.Join(sections, "\n\n")
}

func isCockpitBuiltinSection(id, title string) bool {
	switch id {
	case "usage", "timers", "activity", "hero", "overview", "quota", "trends":
		return true
	}
	switch title {
	case "usage", "timers", "activity", "trends", "quotas":
		return true
	}
	return false
}

func toneLipglossColor(tone string) lipgloss.Color {
	switch strings.ToLower(strings.TrimSpace(tone)) {
	case "crit":
		return colorCrit
	case "warn":
		return colorYellow
	case "ok":
		return colorGreen
	default:
		return colorText
	}
}
