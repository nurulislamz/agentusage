package tui

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/nurulislamz/agentusage/internal/core"
)

// renderMatrixView renders the Dense Roster Matrix HUD layout.
func (m Model) renderMatrixView(w, h int) string {
	ids := m.filteredIDs()
	if len(ids) == 0 {
		empty := []string{
			"",
			dimStyle.Render("  No provider accounts found."),
		}
		return padToSize(strings.Join(empty, "\n"), w, h)
	}

	now := m.viewNow()
	var lines []string

	// Group views by provider
	type providerGroup struct {
		providerID   string
		providerName string
		accountIDs   []string
		indices      []int
	}

	var groups []providerGroup
	groupMap := make(map[string]int)

	for i, id := range ids {
		snap := m.snapshots[id]
		pid := snap.ProviderID
		if pid == "" {
			pid = "other"
		}
		if gIdx, exists := groupMap[pid]; exists {
			groups[gIdx].accountIDs = append(groups[gIdx].accountIDs, id)
			groups[gIdx].indices = append(groups[gIdx].indices, i)
		} else {
			groupMap[pid] = len(groups)
			groups = append(groups, providerGroup{
				providerID:   pid,
				providerName: strings.ToUpper(pid),
				accountIDs:   []string{id},
				indices:      []int{i},
			})
		}
	}

	cursorLineIdx := 0

	for _, grp := range groups {
		pColor := providerThemeColor(grp.providerID)
		hasAttention := false
		for _, id := range grp.accountIDs {
			snap := m.snapshots[id]
			status := core.EffectiveStatus(snap)
			if status == core.StatusLimited || status == core.StatusError {
				hasAttention = true
				break
			}
		}

		statusBadge := lipgloss.NewStyle().Foreground(colorGreen).Render("ALL OK")
		if hasAttention {
			statusBadge = lipgloss.NewStyle().Foreground(colorCrit).Bold(true).Render("ATTENTION")
		}

		headerTitle := lipgloss.NewStyle().Bold(true).Foreground(pColor).Render("◈ " + strings.ToUpper(grp.providerName))
		agentCount := dimStyle.Render(fmt.Sprintf("(%d %s)", len(grp.accountIDs), pluralize(len(grp.accountIDs), "agent", "agents")))
		header := fmt.Sprintf(" %s %s  %s", headerTitle, agentCount, statusBadge)
		lines = append(lines, header)

		// Table Header
		tblHdr := "   " + dimStyle.Render(padRight("ACCOUNT", 20)+" "+padRight("STATUS", 8)+" "+padRight("QUOTA 1", 18)+" "+padRight("QUOTA 2", 18)+" "+padRight("NEXT RESET", 14))
		if w >= 95 {
			tblHdr += " " + dimStyle.Render(padRight("TREND", 10))
		}
		lines = append(lines, tblHdr)

		for rowIdx, id := range grp.accountIDs {
			globalIdx := grp.indices[rowIdx]
			snap := m.snapshots[id]
			selected := (globalIdx == m.cursor)
			if selected {
				cursorLineIdx = len(lines)
			}

			widget := dashboardWidget(snap.ProviderID)
			hideCosts := m.resolveHideCosts(snap)
			di := computeDisplayInfo(snap, widget, hideCosts, m.usageMode)
			cards := projectDetailCards(snap, widget, w, m.warnThreshold, m.critThreshold, m.timeWindow, hideCosts, now, m.usageMode)
			uLines := projectUsageLines(snap, widget, cards, now)
			if len(uLines) == 0 && di.gaugePercent >= 0 {
				pct := di.gaugePercent
				uLines = []WebUsageLine{{
					Label:   "Usage",
					Short:   "Usage",
					Percent: &pct,
					Value:   di.summary,
					Tone:    quotaTone(pct, m.isUsageModeUsed()),
				}}
			}

			// 1. Account column
			effStatus := core.EffectiveStatus(snap)
			statusCol := StatusColor(effStatus)
			statusIco := StatusIcon(effStatus)
			ico := lipgloss.NewStyle().Foreground(statusCol).Render(statusIco)
			accName := snap.AccountID
			if len(accName) > 16 {
				accName = accName[:15] + "…"
			}
			accStyled := lipgloss.NewStyle().Foreground(colorText).Render(accName)
			if selected {
				accStyled = lipgloss.NewStyle().Bold(true).Foreground(pColor).Render(accName)
			}
			accCol := padRight(fmt.Sprintf("%s %s", ico, accStyled), 20)

			// 2. Status badge
			badge := SnapshotStatusBadge(snap)
			badgeCol := padRight(badge, 8)

			// 3. Quotas
			q1 := dimStyle.Render("—")
			q2 := dimStyle.Render("—")
			if len(uLines) > 0 {
				q1 = formatMatrixQuotaCell(uLines[0])
			}
			if len(uLines) > 1 {
				q2 = formatMatrixQuotaCell(uLines[1])
			}

			// 4. Reset
			next := nextResetFromLines(uLines, nil)
			resetStr := dimStyle.Render("—")
			if next != "" {
				resetStr = lipgloss.NewStyle().Foreground(colorPeach).Render("⏱ " + next)
			}
			resetCol := padRight(resetStr, 14)

			// 5. Trend (sparkline)
			trendCol := ""
			if w >= 95 {
				dailyPoints := firstDailySeries(snap, "cost", "analytics_cost", "tokens", "requests")
				if len(dailyPoints) > 1 {
					vals := make([]float64, len(dailyPoints))
					for vi, p := range dailyPoints {
						vals[vi] = math.Max(0, p.Value)
					}
					trendCol = RenderSparkline(vals, 8, pColor)
				} else {
					trendCol = dimStyle.Render("—")
				}
			}

			rowContent := fmt.Sprintf("%s %s %s %s %s", accCol, badgeCol, padRight(q1, 18), padRight(q2, 18), resetCol)
			if trendCol != "" {
				rowContent += " " + trendCol
			}

			indicator := "  "
			if selected {
				indicator = lipgloss.NewStyle().Bold(true).Foreground(pColor).Render("❯ ")
			}

			lines = append(lines, indicator+rowContent)
		}
		lines = append(lines, "")
	}

	// Viewport scrolling
	totalLines := len(lines)
	if h <= 0 {
		return strings.Join(lines, "\n")
	}
	if totalLines <= h {
		return padToSize(strings.Join(lines, "\n"), w, h)
	}

	start := cursorLineIdx - (h / 2)
	if start < 0 {
		start = 0
	}
	if start+h > totalLines {
		start = totalLines - h
		if start < 0 {
			start = 0
		}
	}
	end := min(start+h, totalLines)

	visible := append([]string(nil), lines[start:end]...)
	if start > 0 && len(visible) > 0 {
		visible[0] = lipgloss.NewStyle().Foreground(colorAccent).Render("  ▲ scroll up")
	}
	if end < totalLines && len(visible) > 1 {
		visible[len(visible)-1] = lipgloss.NewStyle().Foreground(colorAccent).Render("  ▼ more below")
	}

	return padToSize(strings.Join(visible, "\n"), w, h)
}

func formatMatrixQuotaCell(l WebUsageLine) string {
	lbl := l.Short
	if lbl == "" {
		lbl = l.Label
	}
	if len(lbl) > 5 {
		lbl = lbl[:4] + "…"
	}
	if l.Percent != nil {
		pct := *l.Percent
		toneCol := toneLipglossColor(l.Tone)
		bar := renderSubmenuMiniBar(pct, 4, toneCol)
		pctStr := lipgloss.NewStyle().Foreground(toneCol).Bold(true).Render(fmt.Sprintf("%3.0f%%", pct))
		return fmt.Sprintf("%s %s %s", padRight(lbl, 4), bar, pctStr)
	}
	val := l.Value
	if val == "" {
		val = "—"
	}
	if len(val) > 10 {
		val = val[:9] + "…"
	}
	return fmt.Sprintf("%s %s", padRight(lbl, 4), val)
}

func renderSubmenuMiniBar(percent float64, width int, toneColor lipgloss.Color) string {
	if width <= 0 {
		width = 4
	}
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	filled := int(math.Round(percent / 100 * float64(width)))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}
	fill := lipgloss.NewStyle().Foreground(toneColor).Render(strings.Repeat("=", filled))
	empty := dimStyle.Render(strings.Repeat("-", width-filled))
	return "[" + fill + empty + "]"
}

func pluralize(n int, singular, plural string) string {
	if n == 1 {
		return singular
	}
	return plural
}
