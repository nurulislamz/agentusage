package tui

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/nurulislamz/agentusage/internal/core"
)

// renderBentoView renders the Viewport Bento Glance Tiles layout.
func (m Model) renderBentoView(w, h int) string {
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

	tileW := 36
	if w < 40 {
		tileW = max(28, w-4)
	}
	cols := max(1, (w-2)/(tileW+2))

	cursorStart, cursorEnd := 0, 0

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

		// Render tiles in rows of `cols`
		var tileBlocks []string
		for rowIdx, id := range grp.accountIDs {
			globalIdx := grp.indices[rowIdx]
			snap := m.snapshots[id]
			selected := (globalIdx == m.cursor)

			tileBlocks = append(tileBlocks, m.renderBentoTile(snap, selected, tileW, now))
		}

		// Group tiles into rows
		for i := 0; i < len(tileBlocks); i += cols {
			end := min(i+cols, len(tileBlocks))
			rowTiles := tileBlocks[i:end]
			rowJoined := lipgloss.JoinHorizontal(lipgloss.Top, rowTiles...)
			rowLines := strings.Split(rowJoined, "\n")

			rowContainsSelected := false
			for colIdx := i; colIdx < end; colIdx++ {
				if grp.indices[colIdx] == m.cursor {
					rowContainsSelected = true
					break
				}
			}

			if rowContainsSelected {
				cursorStart = len(lines)
				cursorEnd = cursorStart + len(rowLines)
			}

			lines = append(lines, rowLines...)
		}
		lines = append(lines, "")
	}

	totalLines := len(lines)
	if h <= 0 {
		return strings.Join(lines, "\n")
	}
	if totalLines <= h {
		return padToSize(strings.Join(lines, "\n"), w, h)
	}

	tileH := cursorEnd - cursorStart
	start := cursorStart - 2
	if tileH < h {
		if cursorEnd >= start+h {
			start = cursorEnd - h + 2
		}
	} else {
		start = cursorStart
	}
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

func (m Model) renderBentoTile(snap core.UsageSnapshot, selected bool, tileW int, now time.Time) string {
	pColor := providerThemeColor(snap.ProviderID)
	effStatus := core.EffectiveStatus(snap)
	statusCol := StatusColor(effStatus)
	statusIco := StatusIcon(effStatus)

	widget := dashboardWidget(snap.ProviderID)
	hideCosts := m.resolveHideCosts(snap)
	di := computeDisplayInfo(snap, widget, hideCosts, m.usageMode)
	cards := projectDetailCards(snap, widget, tileW, m.warnThreshold, m.critThreshold, m.timeWindow, hideCosts, now, m.usageMode)
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

	// Header line
	ico := lipgloss.NewStyle().Foreground(statusCol).Render(statusIco)
	name := snap.AccountID
	maxName := tileW - 14
	if len(name) > maxName && maxName > 3 {
		name = name[:maxName-1] + "…"
	}
	nameStyled := lipgloss.NewStyle().Bold(true).Foreground(colorText).Render(name)
	if selected {
		nameStyled = lipgloss.NewStyle().Bold(true).Foreground(pColor).Render(name)
	}
	leftHead := fmt.Sprintf("%s %s", ico, nameStyled)
	badge := SnapshotStatusBadge(snap)
	headGap := (tileW - 4) - lipgloss.Width(leftHead) - lipgloss.Width(badge)
	if headGap < 1 {
		headGap = 1
	}
	headLine := leftHead + strings.Repeat(" ", headGap) + badge

	// Quota rows (up to 3)
	var bodyRows []string
	if len(uLines) > 0 {
		for _, l := range uLines[:min(3, len(uLines))] {
			lbl := l.Short
			if lbl == "" {
				lbl = l.Label
			}
			if len(lbl) > 7 {
				lbl = lbl[:6] + "…"
			}
			lblStr := padRight(lbl, 7)

			if l.Percent != nil {
				pct := *l.Percent
				toneCol := toneLipglossColor(l.Tone)
				bar := renderSubmenuMiniBar(pct, 4, toneCol)
				pctStr := lipgloss.NewStyle().Foreground(toneCol).Bold(true).Render(fmt.Sprintf("%3.0f%%", pct))
				bodyRows = append(bodyRows, fmt.Sprintf("%s %s %s", lblStr, bar, pctStr))
			} else {
				val := l.Value
				if val == "" {
					val = "—"
				}
				if len(val) > tileW-14 {
					val = val[:tileW-15] + "…"
				}
				bodyRows = append(bodyRows, fmt.Sprintf("%s %s", lblStr, val))
			}
		}
	} else {
		di := computeDisplayInfo(snap, widget, hideCosts, m.usageMode)
		txt := di.summary
		if txt == "" {
			txt = "No active quotas"
		}
		if len(txt) > tileW-6 {
			txt = txt[:tileW-7] + "…"
		}
		bodyRows = append(bodyRows, dimStyle.Render(txt))
	}
	for len(bodyRows) < 3 {
		bodyRows = append(bodyRows, "")
	}

	// Footer: next reset and sparkline
	next := nextResetFromLines(uLines, nil)
	resetStr := dimStyle.Render("—")
	if next != "" {
		resetStr = lipgloss.NewStyle().Foreground(colorPeach).Render("⏱ " + next)
	}

	spark := ""
	dailyPoints := firstDailySeries(snap, "cost", "analytics_cost", "tokens", "requests")
	if len(dailyPoints) > 1 {
		vals := make([]float64, len(dailyPoints))
		for vi, p := range dailyPoints {
			vals[vi] = math.Max(0, p.Value)
		}
		spark = RenderSparkline(vals, 6, pColor)
	}

	footGap := (tileW - 4) - lipgloss.Width(resetStr) - lipgloss.Width(spark)
	if footGap < 1 {
		footGap = 1
	}
	footLine := resetStr + strings.Repeat(" ", footGap) + spark

	content := headLine + "\n" + strings.Join(bodyRows, "\n") + "\n" + footLine

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorSurface1).
		Padding(0, 1).
		Width(tileW).
		MarginRight(1).
		MarginBottom(1)

	if selected {
		boxStyle = boxStyle.
			BorderForeground(pColor).
			Bold(true)
	}

	return boxStyle.Render(content)
}
