package tui

import (
	"github.com/nurulislamz/agentusage/internal/config"
)

type dashboardViewMode string

const (
	dashboardViewSplit  dashboardViewMode = dashboardViewMode(config.DashboardViewSplit)
	dashboardViewMatrix dashboardViewMode = dashboardViewMode(config.DashboardViewMatrix)
	dashboardViewBento  dashboardViewMode = dashboardViewMode(config.DashboardViewBento)
	dashboardViewBars   dashboardViewMode = dashboardViewMode(config.DashboardViewBars)
	dashboardViewDials  dashboardViewMode = dashboardViewMode(config.DashboardViewDials)
	dashboardViewStrips dashboardViewMode = dashboardViewMode(config.DashboardViewStrips)
)

type dashboardViewOption struct {
	ID          dashboardViewMode
	Label       string
	Description string
}

var dashboardViewOptions = []dashboardViewOption{
	{
		ID:          dashboardViewSplit,
		Label:       "Split",
		Description: "Glanceable Submenu + Deep Inspector Cockpit",
	},
	{
		ID:          dashboardViewMatrix,
		Label:       "Matrix",
		Description: "Dense Roster Matrix HUD",
	},
	{
		ID:          dashboardViewBento,
		Label:       "Bento",
		Description: "Viewport Bento Glance Tiles",
	},
	{
		ID:          dashboardViewBars,
		Label:       "Bars",
		Description: "Linear gauges · OpenUsage-style cards",
	},
	{
		ID:          dashboardViewDials,
		Label:       "Dials",
		Description: "Radial gauges · at-a-glance remaining",
	},
	{
		ID:          dashboardViewStrips,
		Label:       "Strips",
		Description: "Grafana bar-gauge wall",
	},
}

func normalizeDashboardViewMode(raw string) dashboardViewMode {
	switch dashboardViewMode(raw) {
	case dashboardViewSplit, dashboardViewMatrix, dashboardViewBento, dashboardViewBars, dashboardViewDials, dashboardViewStrips:
		return dashboardViewMode(raw)
	default:
		return dashboardViewSplit
	}
}

func dashboardViewLabel(mode dashboardViewMode) string {
	for _, opt := range dashboardViewOptions {
		if opt.ID == mode {
			return opt.Label
		}
	}
	return "Split"
}

func dashboardViewIndex(mode dashboardViewMode) int {
	for i, opt := range dashboardViewOptions {
		if opt.ID == mode {
			return i
		}
	}
	return 0
}

func dashboardViewByIndex(index int) dashboardViewMode {
	if len(dashboardViewOptions) == 0 {
		return dashboardViewSplit
	}
	i := index % len(dashboardViewOptions)
	if i < 0 {
		i += len(dashboardViewOptions)
	}
	return dashboardViewOptions[i].ID
}

func (m Model) configuredDashboardView() dashboardViewMode {
	if m.dashboardView != "" {
		return m.dashboardView
	}
	return dashboardViewSplit
}

func (m Model) activeDashboardView() dashboardViewMode {
	if m.dashboardView != "" {
		return m.dashboardView
	}
	return dashboardViewSplit
}

func (m Model) dashboardViewStatusLabel() string {
	return dashboardViewLabel(m.activeDashboardView())
}

func (m *Model) setDashboardView(mode dashboardViewMode) {
	m.dashboardView = normalizeDashboardViewMode(string(mode))
	m.mode = modeList
	m.detailOffset = 0
	m.detailTab = 0
	m.tileOffset = 0
	m.invalidateTileBodyCache()
	m.invalidateDetailCache()
	m.invalidateRenderCaches()
}

func (m Model) nextDashboardView(step int) dashboardViewMode {
	cur := m.activeDashboardView()
	idx := dashboardViewIndex(cur)
	return dashboardViewByIndex(idx + step)
}
