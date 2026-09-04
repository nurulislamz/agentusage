package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/nurulislamz/agentusage/internal/config"
	"github.com/nurulislamz/agentusage/internal/core"
)

func newTDDTestModel() Model {
	ids := []string{"openrouter", "gemini_cli", "ollama", "copilot", "cursor", "claude_code", "codex"}
	snaps := make(map[string]core.UsageSnapshot, len(ids))
	for _, id := range ids {
		snaps[id] = core.UsageSnapshot{
			AccountID:  id,
			ProviderID: id,
			Timestamp:  time.Now(),
			Status:     core.StatusOK,
		}
	}
	m := NewModel(
		0.20,
		0.05,
		false,
		config.DashboardConfig{},
		nil,
		core.TimeWindow30d,
	)
	m.width = 120
	m.height = 35
	m.hasData = true
	m.snapshots = snaps
	m.sortedIDs = ids
	m.dashboardView = dashboardViewSplit
	m.mode = modeList
	return m
}

// ---------------------------------------------------------------------------
// 1. Footer & Header Clickable Button Tests
// ---------------------------------------------------------------------------

func TestFooterClick_MenuButtonOpensSettings(t *testing.T) {
	m := newTDDTestModel()
	footerY := m.height - 1

	// In footer: "p menu" is roughly columns 20-30
	footerText := ansi.Strip(m.renderFooterStatusLine(m.width))
	menuIdx := strings.Index(footerText, "menu")
	if menuIdx < 0 {
		t.Fatalf("menu not found in footer text: %q", footerText)
	}

	updated, _ := m.Update(tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
		X:      menuIdx,
		Y:      footerY,
	})
	got := updated.(Model)
	if !got.settings.show {
		t.Fatalf("expected clicking 'menu' in footer to open settings, but settings.show is false")
	}
}

func TestFooterClick_LayoutButtonCyclesView(t *testing.T) {
	m := newTDDTestModel()
	initialView := m.activeDashboardView()
	footerY := m.height - 1

	footerText := ansi.Strip(m.renderFooterStatusLine(m.width))
	layoutIdx := strings.Index(footerText, "layout")
	if layoutIdx < 0 {
		t.Fatalf("layout not found in footer text: %q", footerText)
	}

	updated, _ := m.Update(tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
		X:      layoutIdx,
		Y:      footerY,
	})
	got := updated.(Model)
	if got.activeDashboardView() == initialView {
		t.Fatalf("expected clicking 'layout' in footer to cycle view from %v, got %v", initialView, got.activeDashboardView())
	}
}

func TestFooterClick_ModeButtonTogglesUsageMode(t *testing.T) {
	m := newTDDTestModel()
	initialMode := m.usageMode
	footerY := m.height - 1

	footerText := ansi.Strip(m.renderFooterStatusLine(m.width))
	modeIdx := strings.Index(footerText, "mode")
	if modeIdx < 0 {
		t.Fatalf("mode not found in footer text: %q", footerText)
	}

	updated, _ := m.Update(tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
		X:      modeIdx,
		Y:      footerY,
	})
	got := updated.(Model)
	if got.usageMode == initialMode {
		t.Fatalf("expected clicking 'mode' in footer to toggle usage mode from %v, got %v", initialMode, got.usageMode)
	}
}

func TestFooterClick_HelpButtonTogglesHelp(t *testing.T) {
	m := newTDDTestModel()
	footerY := m.height - 1

	footerText := ansi.Strip(m.renderFooterStatusLine(m.width))
	helpIdx := strings.Index(footerText, "help")
	if helpIdx < 0 {
		t.Fatalf("help not found in footer text: %q", footerText)
	}

	updated, _ := m.Update(tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
		X:      helpIdx,
		Y:      footerY,
	})
	got := updated.(Model)
	if !got.showHelp {
		t.Fatalf("expected clicking 'help' in footer to show help overlay, got showHelp=false")
	}
}

func TestFooterClick_DetailModeEscBackExitsDetail(t *testing.T) {
	m := newTDDTestModel()
	m.mode = modeDetail
	footerY := m.height - 1

	footerText := ansi.Strip(m.renderFooterStatusLine(m.width))
	backIdx := strings.Index(footerText, "Esc back")
	if backIdx < 0 {
		backIdx = strings.Index(footerText, "back")
	}
	if backIdx < 0 {
		t.Fatalf("back not found in footer text: %q", footerText)
	}

	updated, _ := m.Update(tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
		X:      backIdx + 2,
		Y:      footerY,
	})
	got := updated.(Model)
	if got.mode != modeList {
		t.Fatalf("expected clicking 'Esc back' in footer to exit detail mode, got %v", got.mode)
	}
}

// ---------------------------------------------------------------------------
// 2. Scrolling in All 6 Dashboard Views
// ---------------------------------------------------------------------------

func TestMouseScroll_SplitViewNavigatorVsDetail(t *testing.T) {
	m := newTDDTestModel()
	m.cursor = 0
	m.detailOffset = 0

	// 1. Mouse wheel down on left list (X=10) should advance cursor
	updated, _ := m.Update(tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonWheelDown,
		X:      10,
		Y:      10,
	})
	got := updated.(Model)
	if got.cursor <= 0 {
		t.Fatalf("expected mouse scroll down on left column to advance cursor, got %d", got.cursor)
	}

	// 2. Mouse wheel down on right detail pane (X=80) should advance detailOffset
	updated2, _ := got.Update(tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonWheelDown,
		X:      80,
		Y:      10,
	})
	got2 := updated2.(Model)
	if got2.detailOffset <= 0 {
		t.Fatalf("expected mouse scroll down on right detail pane to advance detailOffset, got %d", got2.detailOffset)
	}
}

func TestMouseScroll_NonSplitViewsScrollsItems(t *testing.T) {
	views := []dashboardViewMode{
		dashboardViewMatrix,
		dashboardViewBento,
		dashboardViewBars,
		dashboardViewDials,
		dashboardViewStrips,
	}

	for _, v := range views {
		t.Run(string(v), func(t *testing.T) {
			m := newTDDTestModel()
			m.setDashboardView(v)
			m.cursor = 0

			// Mouse wheel down anywhere in the content area should advance cursor or scroll view
			updated, _ := m.Update(tea.MouseMsg{
				Action: tea.MouseActionPress,
				Button: tea.MouseButtonWheelDown,
				X:      m.width / 2,
				Y:      m.height / 2,
			})
			got := updated.(Model)
			if got.cursor == 0 && got.tileOffset == 0 {
				t.Fatalf("view %v: expected mouse wheel down to advance cursor or scroll view, got cursor=%d, tileOffset=%d", v, got.cursor, got.tileOffset)
			}
		})
	}
}

func TestKeyboardScroll_NonSplitViewsPageAndJump(t *testing.T) {
	views := []dashboardViewMode{
		dashboardViewMatrix,
		dashboardViewBento,
		dashboardViewBars,
		dashboardViewDials,
		dashboardViewStrips,
	}

	for _, v := range views {
		t.Run(string(v), func(t *testing.T) {
			m := newTDDTestModel()
			m.setDashboardView(v)
			m.cursor = 0

			// PgDown should advance cursor
			updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyPgDown})
			got := updated.(Model)
			if got.cursor <= 0 {
				t.Fatalf("view %v: expected PgDown to advance cursor from 0, got %d", v, got.cursor)
			}

			// End / G should jump to last item
			updatedEnd, _ := got.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
			gotEnd := updatedEnd.(Model)
			if gotEnd.cursor != len(m.sortedIDs)-1 {
				t.Fatalf("view %v: expected 'G' to jump to last item %d, got %d", v, len(m.sortedIDs)-1, gotEnd.cursor)
			}

			// Home / g should jump to first item
			updatedHome, _ := gotEnd.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
			gotHome := updatedHome.(Model)
			if gotHome.cursor != 0 {
				t.Fatalf("view %v: expected 'g' to jump to first item 0, got %d", v, gotHome.cursor)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 3. Item Selection Mouse Clicks Across All Views
// ---------------------------------------------------------------------------

func TestMouseLeftClick_MatrixViewSelectsRow(t *testing.T) {
	m := newTDDTestModel()
	m.setDashboardView(dashboardViewMatrix)
	m.cursor = 0

	// Matrix view has header lines (agentUsage + separator = 2)
	// Provider header for OPENROUTER is line 2
	// Table header is line 3
	// Account 0 (openrouter) is line 4
	// Next provider GEMINI_CLI is line 6
	// Table header is line 7
	// Account 1 (gemini_cli) is line 8
	// Click on Account 1 row
	clickY := 8
	updated, _ := m.Update(tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
		X:      30,
		Y:      clickY,
	})
	got := updated.(Model)
	if got.cursor != 1 {
		t.Fatalf("expected clicking matrix row at Y=%d to select item 1, got cursor=%d", clickY, got.cursor)
	}
}

func TestBento2DNavigation_LeftRightArrows(t *testing.T) {
	m := newTDDTestModel()
	m.setDashboardView(dashboardViewBento)
	m.cursor = 0

	// In Bento view, pressing "right" or "l" should navigate to next item without forcing detail mode
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRight})
	got := updated.(Model)
	if got.mode == modeDetail {
		t.Fatalf("expected right arrow in bento view to navigate to next tile, but it entered modeDetail")
	}
	if got.cursor != 1 {
		t.Fatalf("expected right arrow to move cursor to 1, got %d", got.cursor)
	}

	// Pressing "left" or "h" should navigate back to previous item
	updatedLeft, _ := got.Update(tea.KeyMsg{Type: tea.KeyLeft})
	gotLeft := updatedLeft.(Model)
	if gotLeft.cursor != 0 {
		t.Fatalf("expected left arrow to move cursor back to 0, got %d", gotLeft.cursor)
	}
}

// ---------------------------------------------------------------------------
// 4. Cockpit Detail Section Navigation
// ---------------------------------------------------------------------------

func TestCockpitDetailSectionStarts_MatchesRenderedSections(t *testing.T) {
	m := newTDDTestModel()
	m.mode = modeDetail
	m.cursor = 0

	// In detail mode, navigateDetailSection(1) should jump to the next section
	// The first section in RenderCockpit is "⚡ USAGE & QUOTAS", followed by "⏱ TIMERS & SCHEDULE"
	starts := m.detailSectionStarts()
	if len(starts) == 0 {
		t.Fatalf("expected detailSectionStarts() to return section start lines for cockpit, got empty")
	}

	// First start should be at or after the hero area (>= 3)
	if starts[0] < 3 {
		t.Fatalf("expected first section start to be >= 3 (after hero), got %d", starts[0])
	}

	// Pressing Tab should advance detailOffset to starts[0]
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	got := updated.(Model)
	if got.detailOffset != starts[0] {
		t.Fatalf("expected Tab to advance detailOffset to first section %d, got %d", starts[0], got.detailOffset)
	}

	// Pressing Tab again should advance to starts[1] if available
	if len(starts) > 1 {
		updated2, _ := got.Update(tea.KeyMsg{Type: tea.KeyTab})
		got2 := updated2.(Model)
		if got2.detailOffset != starts[1] {
			t.Fatalf("expected Tab to advance detailOffset to second section %d, got %d", starts[1], got2.detailOffset)
		}

		// Shift+Tab should navigate back to starts[0]
		updatedBack, _ := got2.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
		gotBack := updatedBack.(Model)
		if gotBack.detailOffset != starts[0] {
			t.Fatalf("expected Shift+Tab to navigate back to %d, got %d", starts[0], gotBack.detailOffset)
		}
	}
}

// ---------------------------------------------------------------------------
// 5. Card Click Selection in Bento and Bars Views
// ---------------------------------------------------------------------------

func TestMouseLeftClick_BentoCardSelectsItem(t *testing.T) {
	m := newTDDTestModel()
	m.setDashboardView(dashboardViewBento)
	m.cursor = 0

	// In Bento view, provider 0 header is line 2, tile 0 is lines 3..9
	// Provider 1 header is line 10, tile 1 is lines 11..17
	clickY := 12
	updated, _ := m.Update(tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
		X:      20,
		Y:      clickY,
	})
	got := updated.(Model)
	if got.cursor != 1 {
		t.Fatalf("expected clicking Bento card at Y=%d to select item 1, got cursor=%d", clickY, got.cursor)
	}
}

func TestMouseLeftClick_BarsCardSelectsItem(t *testing.T) {
	m := newTDDTestModel()
	m.setDashboardView(dashboardViewBars)
	m.cursor = 0

	// In Bars view, provider 0 header is line 2, card 0 is lines 3..8
	// Provider 1 header is line 9, card 1 is lines 10..15
	clickY := 11
	updated, _ := m.Update(tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
		X:      20,
		Y:      clickY,
	})
	got := updated.(Model)
	if got.cursor != 1 {
		t.Fatalf("expected clicking Bars card at Y=%d to select item 1, got cursor=%d", clickY, got.cursor)
	}
}

// ---------------------------------------------------------------------------
// 6. Settings Modal Interactions
// ---------------------------------------------------------------------------

func TestSettingsModal_MouseWheelScrollsBody(t *testing.T) {
	m := newTDDTestModel()
	m.settings.show = true
	m.settings.tab = settingsTabProviders
	m.settings.bodyOffset = 0

	// Mouse wheel down inside modal
	updated, _ := m.Update(tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonWheelDown,
		X:      m.width / 2,
		Y:      m.height / 2,
	})
	got := updated.(Model)
	if got.settings.bodyOffset <= 0 {
		t.Fatalf("expected mouse wheel down in settings to advance bodyOffset, got %d", got.settings.bodyOffset)
	}
}

