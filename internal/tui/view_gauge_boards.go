package tui

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/nurulislamz/agentusage/internal/core"
)

type gaugeBoardGroup struct {
	providerID   string
	providerName string
	accountIDs   []string
	indices      []int
}

func (m Model) groupIDsByProvider(ids []string) []gaugeBoardGroup {
	var groups []gaugeBoardGroup
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
			groups = append(groups, gaugeBoardGroup{
				providerID:   pid,
				providerName: strings.ToUpper(pid),
				accountIDs:   []string{id},
				indices:      []int{i},
			})
		}
	}
	return groups
}

// renderBarsView renders the Linear gauges · OpenUsage-style cards layout.
func (m Model) renderBarsView(w, h int) string {
	ids := m.filteredIDs()
	if len(ids) == 0 {
		return padToSize("\n"+dimStyle.Render("  No provider accounts found."), w, h)
	}

	now := m.viewNow()
	var lines []string
	groups := m.groupIDsByProvider(ids)
	cardW := clamp(w-4, 30, 90)
	cursorLineIdx := 0

	for _, grp := range groups {
		pColor := providerThemeColor(grp.providerID)
		header := m.renderBoardGroupHeader(grp, pColor)
		lines = append(lines, header)

		for rowIdx, id := range grp.accountIDs {
			globalIdx := grp.indices[rowIdx]
			snap := m.snapshots[id]
			selected := (globalIdx == m.cursor)
			if selected {
				cursorLineIdx = len(lines)
			}

			lines = append(lines, m.renderBarCard(snap, selected, cardW, now))
		}
		lines = append(lines, "")
	}

	return m.scrollBoardLines(lines, cursorLineIdx, w, h)
}

func (m Model) renderBarCard(snap core.UsageSnapshot, selected bool, cardW int, now time.Time) string {
	pColor := providerThemeColor(snap.ProviderID)
	effStatus := core.EffectiveStatus(snap)
	statusCol := StatusColor(effStatus)
	statusIco := StatusIcon(effStatus)

	widget := dashboardWidget(snap.ProviderID)
	hideCosts := m.resolveHideCosts(snap)
	di := computeDisplayInfo(snap, widget, hideCosts, m.usageMode)
	cards := projectDetailCards(snap, widget, cardW, m.warnThreshold, m.critThreshold, m.timeWindow, hideCosts, now, m.usageMode)
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

	// Head
	ico := lipgloss.NewStyle().Foreground(statusCol).Render(statusIco)
	name := snap.AccountID
	nameStyled := lipgloss.NewStyle().Bold(true).Foreground(colorText).Render(name)
	if selected {
		nameStyled = lipgloss.NewStyle().Bold(true).Foreground(pColor).Render(name)
	}
	leftHead := fmt.Sprintf("%s %s", ico, nameStyled)
	badge := SnapshotStatusBadge(snap)
	headGap := cardW - lipgloss.Width(leftHead) - lipgloss.Width(badge) - 4
	if headGap < 1 {
		headGap = 1
	}
	headLine := leftHead + strings.Repeat(" ", headGap) + badge

	// Body: Linear gauges
	var bodyRows []string
	if len(uLines) > 0 {
		for _, l := range uLines {
			lbl := l.Short
			if lbl == "" {
				lbl = l.Label
			}
			lblPadded := padRight(lbl, 12)

			if l.Percent != nil {
				pct := *l.Percent
				toneCol := toneLipglossColor(l.Tone)
				barW := clamp(cardW/3, 10, 24)
				bar := RenderGauge(pct, barW, m.warnThreshold, m.critThreshold)
				pctStr := lipgloss.NewStyle().Foreground(toneCol).Bold(true).Render(fmt.Sprintf("%5.1f%%", pct))
				hint := ""
				if l.ResetIn != "" {
					hint = dimStyle.Render(" · ⏱ in " + l.ResetIn)
				}
				bodyRows = append(bodyRows, fmt.Sprintf("  %s %s %s%s", lblPadded, bar, pctStr, hint))
			} else {
				val := l.Value
				if val == "" {
					val = "—"
				}
				bodyRows = append(bodyRows, fmt.Sprintf("  %s %s", lblPadded, val))
			}
		}
	} else {
		di := computeDisplayInfo(snap, widget, hideCosts, m.usageMode)
		txt := di.summary
		if txt == "" {
			txt = "No active quotas"
		}
		bodyRows = append(bodyRows, "  "+dimStyle.Render(txt))
	}

	// Foot: Activity sparkline if present
	spark := ""
	dailyPoints := firstDailySeries(snap, "cost", "analytics_cost", "tokens", "requests")
	if len(dailyPoints) > 1 {
		vals := make([]float64, len(dailyPoints))
		for vi, p := range dailyPoints {
			vals[vi] = math.Max(0, p.Value)
		}
		spark = "  " + RenderSparkline(vals, 12, pColor)
	}
	if spark != "" {
		bodyRows = append(bodyRows, spark)
	}

	content := headLine + "\n" + strings.Join(bodyRows, "\n")
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorSurface1).
		Padding(0, 1).
		Width(cardW).
		MarginBottom(1)

	if selected {
		boxStyle = boxStyle.
			BorderForeground(pColor).
			Bold(true)
	}

	return boxStyle.Render(content)
}

// renderDialsView renders the Radial gauges · at-a-glance remaining layout.
func (m Model) renderDialsView(w, h int) string {
	ids := m.filteredIDs()
	if len(ids) == 0 {
		return padToSize("\n"+dimStyle.Render("  No provider accounts found."), w, h)
	}

	now := m.viewNow()
	var lines []string
	groups := m.groupIDsByProvider(ids)
	cardW := clamp(w-4, 30, 90)
	cursorLineIdx := 0

	for _, grp := range groups {
		pColor := providerThemeColor(grp.providerID)
		header := m.renderBoardGroupHeader(grp, pColor)
		lines = append(lines, header)

		for rowIdx, id := range grp.accountIDs {
			globalIdx := grp.indices[rowIdx]
			snap := m.snapshots[id]
			selected := (globalIdx == m.cursor)
			if selected {
				cursorLineIdx = len(lines)
			}

			lines = append(lines, m.renderDialCard(snap, selected, cardW, now))
		}
		lines = append(lines, "")
	}

	return m.scrollBoardLines(lines, cursorLineIdx, w, h)
}

func (m Model) renderDialCard(snap core.UsageSnapshot, selected bool, cardW int, now time.Time) string {
	pColor := providerThemeColor(snap.ProviderID)
	effStatus := core.EffectiveStatus(snap)
	statusCol := StatusColor(effStatus)
	statusIco := StatusIcon(effStatus)

	widget := dashboardWidget(snap.ProviderID)
	hideCosts := m.resolveHideCosts(snap)
	di := computeDisplayInfo(snap, widget, hideCosts, m.usageMode)
	cards := projectDetailCards(snap, widget, cardW, m.warnThreshold, m.critThreshold, m.timeWindow, hideCosts, now, m.usageMode)
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

	// Head
	ico := lipgloss.NewStyle().Foreground(statusCol).Render(statusIco)
	name := snap.AccountID
	nameStyled := lipgloss.NewStyle().Bold(true).Foreground(colorText).Render(name)
	if selected {
		nameStyled = lipgloss.NewStyle().Bold(true).Foreground(pColor).Render(name)
	}
	leftHead := fmt.Sprintf("%s %s", ico, nameStyled)
	badge := SnapshotStatusBadge(snap)
	headGap := cardW - lipgloss.Width(leftHead) - lipgloss.Width(badge) - 4
	if headGap < 1 {
		headGap = 1
	}
	headLine := leftHead + strings.Repeat(" ", headGap) + badge

	// Body: Radial ASCII dials
	var dialRows []string
	if len(uLines) > 0 {
		for _, l := range uLines {
			lbl := l.Short
			if lbl == "" {
				lbl = l.Label
			}
			if l.Percent != nil {
				pct := *l.Percent
				toneCol := toneLipglossColor(l.Tone)
				// Radial arc gauge representation
				arcTop := lipgloss.NewStyle().Foreground(toneCol).Render(fmt.Sprintf("╭─ %3.0f%% ─╮", pct))
				arcBot := lipgloss.NewStyle().Foreground(toneCol).Render("╰────────╯")
				hint := ""
				if l.ResetIn != "" {
					hint = " · ⏱ in " + l.ResetIn
				}
				dialRows = append(dialRows, fmt.Sprintf("  %s  %s%s", arcTop, padRight(lbl, 10), dimStyle.Render(hint)))
				dialRows = append(dialRows, fmt.Sprintf("  %s", arcBot))
			} else {
				val := l.Value
				if val == "" {
					val = "—"
				}
				dialRows = append(dialRows, fmt.Sprintf("  ( %s ) %s", padRight(lbl, 6), val))
			}
		}
	} else {
		di := computeDisplayInfo(snap, widget, hideCosts, m.usageMode)
		txt := di.summary
		if txt == "" {
			txt = "No active quotas"
		}
		dialRows = append(dialRows, "  "+dimStyle.Render(txt))
	}

	content := headLine + "\n" + strings.Join(dialRows, "\n")
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorSurface1).
		Padding(0, 1).
		Width(cardW).
		MarginBottom(1)

	if selected {
		boxStyle = boxStyle.
			BorderForeground(pColor).
			Bold(true)
	}

	return boxStyle.Render(content)
}

// renderStripsView renders the Grafana bar-gauge wall layout.
func (m Model) renderStripsView(w, h int) string {
	ids := m.filteredIDs()
	if len(ids) == 0 {
		return padToSize("\n"+dimStyle.Render("  No provider accounts found."), w, h)
	}

	now := m.viewNow()
	var lines []string
	groups := m.groupIDsByProvider(ids)
	stripW := clamp(w-4, 30, 110)
	cursorLineIdx := 0

	for _, grp := range groups {
		pColor := providerThemeColor(grp.providerID)
		header := m.renderBoardGroupHeader(grp, pColor)
		lines = append(lines, header)

		for rowIdx, id := range grp.accountIDs {
			globalIdx := grp.indices[rowIdx]
			snap := m.snapshots[id]
			selected := (globalIdx == m.cursor)
			if selected {
				cursorLineIdx = len(lines)
			}

			lines = append(lines, m.renderStripRow(snap, selected, stripW, now))
		}
		lines = append(lines, "")
	}

	return m.scrollBoardLines(lines, cursorLineIdx, w, h)
}

func (m Model) renderStripRow(snap core.UsageSnapshot, selected bool, stripW int, now time.Time) string {
	pColor := providerThemeColor(snap.ProviderID)
	effStatus := core.EffectiveStatus(snap)
	statusCol := StatusColor(effStatus)
	statusIco := StatusIcon(effStatus)

	widget := dashboardWidget(snap.ProviderID)
	hideCosts := m.resolveHideCosts(snap)
	di := computeDisplayInfo(snap, widget, hideCosts, m.usageMode)
	cards := projectDetailCards(snap, widget, stripW, m.warnThreshold, m.critThreshold, m.timeWindow, hideCosts, now, m.usageMode)
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

	// Line 1: ID header
	ico := lipgloss.NewStyle().Foreground(statusCol).Render(statusIco)
	name := snap.AccountID
	nameStyled := lipgloss.NewStyle().Bold(true).Foreground(colorText).Render(name)
	if selected {
		nameStyled = lipgloss.NewStyle().Bold(true).Foreground(pColor).Render(name)
	}
	leftPart := fmt.Sprintf("%s %s", ico, nameStyled)
	badge := SnapshotStatusBadge(snap)
	gap1 := stripW - lipgloss.Width(leftPart) - lipgloss.Width(badge)
	if gap1 < 1 {
		gap1 = 1
	}
	idLine := leftPart + strings.Repeat(" ", gap1) + badge

	var rowLines []string
	rowLines = append(rowLines, idLine)

	// Wide bar-gauge wall strips
	if len(uLines) > 0 {
		for _, l := range uLines {
			lbl := l.Short
			if lbl == "" {
				lbl = l.Label
			}
			lblStr := padRight(lbl, 8)

			if l.Percent != nil {
				pct := *l.Percent
				toneCol := toneLipglossColor(l.Tone)
				barWidth := clamp(stripW-32, 12, 50)
				bar := renderStripBar(pct, barWidth, toneCol)
				pctStr := lipgloss.NewStyle().Foreground(toneCol).Bold(true).Render(fmt.Sprintf("%3.0f%%", pct))
				hint := ""
				if l.ResetIn != "" {
					hint = " · ⏱ in " + l.ResetIn
				}
				rowLines = append(rowLines, fmt.Sprintf("  %s %s %s%s", lblStr, bar, pctStr, dimStyle.Render(hint)))
			} else {
				val := l.Value
				if val == "" {
					val = "—"
				}
				rowLines = append(rowLines, fmt.Sprintf("  %s %s", lblStr, val))
			}
		}
	} else {
		di := computeDisplayInfo(snap, widget, hideCosts, m.usageMode)
		txt := di.summary
		if txt == "" {
			txt = "No active quotas"
		}
		rowLines = append(rowLines, "  "+dimStyle.Render(txt))
	}

	indicator := "  "
	if selected {
		indicator = lipgloss.NewStyle().Bold(true).Foreground(pColor).Render("┃ ")
	}

	for i := range rowLines {
		rowLines[i] = indicator + rowLines[i]
	}
	rowLines = append(rowLines, "  "+surface1Style.Render(strings.Repeat("─", stripW-4)))

	return strings.Join(rowLines, "\n")
}

func renderStripBar(percent float64, width int, toneColor lipgloss.Color) string {
	if width <= 0 {
		width = 20
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
	fill := lipgloss.NewStyle().Foreground(toneColor).Render(strings.Repeat("█", filled))
	empty := dimStyle.Render(strings.Repeat("░", width-filled))
	return fill + empty
}

func (m Model) renderBoardGroupHeader(grp gaugeBoardGroup, pColor lipgloss.Color) string {
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
	return fmt.Sprintf(" %s %s  %s", headerTitle, agentCount, statusBadge)
}

func (m Model) scrollBoardLines(lines []string, cursorLineIdx, w, h int) string {
	totalLines := len(lines)
	if h <= 0 {
		return strings.Join(lines, "\n")
	}
	if totalLines <= h {
		return padToSize(strings.Join(lines, "\n"), w, h)
	}

	start := cursorLineIdx - (h / 3)
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
