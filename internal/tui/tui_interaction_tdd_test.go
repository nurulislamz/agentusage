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
	m.providerOrder = ids
	m.providerEnabled = make(map[string]bool)
	m.accountProviders = make(map[string]string)
	for _, id := range ids {
		m.providerEnabled[id] = true
		m.accountProviders[id] = id
	}
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

func TestBoardViews_ScrollingDownShowsSelectedBottomItemInView(t *testing.T) {
	views := []dashboardViewMode{
		dashboardViewBars,
		dashboardViewDials,
		dashboardViewStrips,
	}

	for _, v := range views {
		t.Run(string(v), func(t *testing.T) {
			m := newTDDTestModel()
			m.setDashboardView(v)
			m.height = 24
			lastIdx := len(m.sortedIDs) - 1
			lastName := m.sortedIDs[lastIdx] // "codex"

			// Move cursor down to last item via "j"
			for i := 0; i < lastIdx; i++ {
				updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
				m = updated.(Model)
			}

			if m.cursor != lastIdx {
				t.Fatalf("expected cursor to be %d, got %d", lastIdx, m.cursor)
			}

			view := m.View()
			if !strings.Contains(view, lastName) {
				t.Fatalf("view %v at height 24 with cursor on last item %q did not render %q in View():\n%s", v, lastName, lastName, view)
			}
		})
	}
}

func TestBentoView_ScrollingDownShowsSelectedBottomItemInView(t *testing.T) {
	m := newTDDTestModel()
	m.setDashboardView(dashboardViewBento)
	m.height = 20
	lastIdx := len(m.sortedIDs) - 1
	lastName := m.sortedIDs[lastIdx] // "codex"

	// Move cursor down to last item via "j"
	for i := 0; i < lastIdx; i++ {
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		m = updated.(Model)
	}

	if m.cursor != lastIdx {
		t.Fatalf("expected cursor to be %d, got %d", lastIdx, m.cursor)
	}

	view := m.View()
	if !strings.Contains(view, lastName) {
		t.Fatalf("view bento at height 20 with cursor on last item %q did not render %q in View():\n%s", lastName, lastName, view)
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

func TestEnterKey_EntersDetailModeInAllViews(t *testing.T) {
	views := []dashboardViewMode{
		dashboardViewSplit,
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
			m.cursor = 1

			// Press Enter
			updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
			got := updated.(Model)
			if got.mode != modeDetail {
				t.Fatalf("view %v: expected Enter to switch to modeDetail, got %v", v, got.mode)
			}
			if got.cursor != 1 {
				t.Fatalf("view %v: expected cursor 1 to be preserved in modeDetail, got %d", v, got.cursor)
			}

			// Press Esc to exit detail mode
			updatedEsc, _ := got.Update(tea.KeyMsg{Type: tea.KeyEsc})
			gotEsc := updatedEsc.(Model)
			if gotEsc.mode != modeList {
				t.Fatalf("view %v: expected Esc to return to modeList, got %v", v, gotEsc.mode)
			}
			if gotEsc.activeDashboardView() != v {
				t.Fatalf("view %v: expected activeDashboardView to be %v after exiting detail, got %v", v, v, gotEsc.activeDashboardView())
			}
		})
	}
}

func TestMouseClick_EntersDetailModeInAllViews(t *testing.T) {
	t.Run("split_click_right_pane_stays_in_modelist", func(t *testing.T) {
		m := newTDDTestModel()
		m.setDashboardView(dashboardViewSplit)
		m.cursor = 0

		// Click on right detail pane (X=80, Y=10)
		updated, _ := m.Update(tea.MouseMsg{
			Action: tea.MouseActionPress,
			Button: tea.MouseButtonLeft,
			X:      80,
			Y:      10,
		})
		got := updated.(Model)
		if got.mode != modeList {
			t.Fatalf("expected clicking right detail pane in split view to stay in modeList, got %v", got.mode)
		}
	})

	t.Run("split_click_left_pane_selects_item_in_modelist", func(t *testing.T) {
		m := newTDDTestModel()
		m.setDashboardView(dashboardViewSplit)
		m.cursor = 0

		// Click on item 1 (X=10, Y=7)
		updated, _ := m.Update(tea.MouseMsg{
			Action: tea.MouseActionPress,
			Button: tea.MouseButtonLeft,
			X:      10,
			Y:      7,
		})
		got := updated.(Model)
		if got.mode != modeList {
			t.Fatalf("expected clicking left pane in split view to remain in modeList, got %v", got.mode)
		}
		if got.cursor != 1 {
			t.Fatalf("expected clicking item 1 in left pane to select cursor 1, got %d", got.cursor)
		}
	})

	views := []dashboardViewMode{
		dashboardViewMatrix,
		dashboardViewBento,
		dashboardViewBars,
		dashboardViewDials,
		dashboardViewStrips,
	}

	for _, v := range views {
		t.Run(string(v)+"_click_selected_enters_detail", func(t *testing.T) {
			m := newTDDTestModel()
			m.setDashboardView(v)
			m.cursor = 0

			// In all these views, first item is at Y around 3 or 4
			// First click selects it (already cursor=0). Second click on selected item enters detail mode!
			clickY := 3
			if v == dashboardViewMatrix {
				clickY = 4
			}
			updated, _ := m.Update(tea.MouseMsg{
				Action: tea.MouseActionPress,
				Button: tea.MouseButtonLeft,
				X:      20,
				Y:      clickY,
			})
			got := updated.(Model)
			if got.mode != modeDetail {
				t.Fatalf("view %v: expected clicking selected item to enter modeDetail, got %v", v, got.mode)
			}
		})
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
	if got.settings.bodyOffset <= 0 && got.settings.cursor <= 0 {
		t.Fatalf("expected mouse wheel down in settings to advance bodyOffset or cursor, got bodyOffset=%d, cursor=%d", got.settings.bodyOffset, got.settings.cursor)
	}
}

func TestSettingsModal_MouseWheelScrollsActiveTabCursor(t *testing.T) {
	t.Run("providers_tab", func(t *testing.T) {
		m := newTDDTestModel()
		m.settings.show = true
		m.settings.tab = settingsTabProviders
		m.settings.cursor = 0

		updated, _ := m.Update(tea.MouseMsg{
			Action: tea.MouseActionPress,
			Button: tea.MouseButtonWheelDown,
			X:      m.width / 2,
			Y:      m.height / 2,
		})
		got := updated.(Model)
		if got.settings.cursor <= 0 {
			t.Fatalf("expected mouse wheel down in providers settings to advance cursor, got %d", got.settings.cursor)
		}
	})

	t.Run("themes_tab", func(t *testing.T) {
		m := newTDDTestModel()
		m.settings.show = true
		m.settings.tab = settingsTabTheme
		m.settings.themeCursor = 0

		updated, _ := m.Update(tea.MouseMsg{
			Action: tea.MouseActionPress,
			Button: tea.MouseButtonWheelDown,
			X:      m.width / 2,
			Y:      m.height / 2,
		})
		got := updated.(Model)
		if got.settings.themeCursor <= 0 {
			t.Fatalf("expected mouse wheel down in themes settings to advance themeCursor, got %d", got.settings.themeCursor)
		}
	})

	t.Run("api_keys_tab", func(t *testing.T) {
		m := newTDDTestModel()
		m.settings.show = true
		m.settings.tab = settingsTabAPIKeys
		m.settings.cursor = 0

		updated, _ := m.Update(tea.MouseMsg{
			Action: tea.MouseActionPress,
			Button: tea.MouseButtonWheelDown,
			X:      m.width / 2,
			Y:      m.height / 2,
		})
		got := updated.(Model)
		if got.settings.cursor <= 0 {
			t.Fatalf("expected mouse wheel down in API keys settings to advance cursor, got %d", got.settings.cursor)
		}
	})

	t.Run("sections_tab_list_and_preview", func(t *testing.T) {
		m := newTDDTestModel()
		m.settings.show = true
		m.settings.tab = settingsTabWidgetSections
		m.settings.sectionRowCursor = 0
		m.settings.previewOffset = 0

		// Wheel down on left half -> advances sectionRowCursor
		updatedLeft, _ := m.Update(tea.MouseMsg{
			Action: tea.MouseActionPress,
			Button: tea.MouseButtonWheelDown,
			X:      m.width / 4,
			Y:      m.height / 2,
		})
		gotLeft := updatedLeft.(Model)
		if gotLeft.settings.sectionRowCursor <= 0 {
			t.Fatalf("expected mouse wheel down on left side of sections tab to advance sectionRowCursor, got %d", gotLeft.settings.sectionRowCursor)
		}

		// Wheel down on right half -> advances previewOffset
		updatedRight, _ := m.Update(tea.MouseMsg{
			Action: tea.MouseActionPress,
			Button: tea.MouseButtonWheelDown,
			X:      3 * m.width / 4,
			Y:      m.height / 2,
		})
		gotRight := updatedRight.(Model)
		if gotRight.settings.previewOffset <= 0 {
			t.Fatalf("expected mouse wheel down on right side of sections tab to advance previewOffset, got %d", gotRight.settings.previewOffset)
		}
	})
}

func TestSettingsModal_ClickRowActivatesOrEntersItem(t *testing.T) {
	t.Run("providers_click_toggles_provider", func(t *testing.T) {
		m := newTDDTestModel()
		m.settings.show = true
		m.settings.tab = settingsTabProviders
		ids := m.settingsIDs()
		if len(ids) == 0 {
			t.Fatal("no settings IDs")
		}
		targetID := ids[0]
		initialState := m.isProviderEnabled(targetID)

		// Calculate modal bounds
		contentW := clamp(m.width-24, 68, 92)
		panelH := clamp(20, 8, m.height-14) + 9
		modalStartX := (m.width - (contentW + 2)) / 2
		modalStartY := (m.height - panelH) / 2
		bodyStartY := modalStartY + 5
		// Row 0 in providers tab: 4 header lines + 1 group header = offset 5
		clickY := bodyStartY + 5

		updated, _ := m.Update(tea.MouseMsg{
			Action: tea.MouseActionPress,
			Button: tea.MouseButtonLeft,
			X:      modalStartX + 10,
			Y:      clickY,
		})
		got := updated.(Model)
		if got.isProviderEnabled(targetID) == initialState {
			t.Fatalf("expected clicking provider row to toggle enabled state from %v, got %v", initialState, got.isProviderEnabled(targetID))
		}
	})

	t.Run("themes_click_selects_theme", func(t *testing.T) {
		m := newTDDTestModel()
		m.settings.show = true
		m.settings.tab = settingsTabTheme
		themes := AvailableThemes()
		if len(themes) < 2 {
			t.Fatal("need at least 2 themes to test")
		}

		contentW := clamp(m.width-24, 68, 92)
		panelH := clamp(20, 8, m.height-14) + 9
		modalStartX := (m.width - (contentW + 2)) / 2
		modalStartY := (m.height - panelH) / 2
		bodyStartY := modalStartY + 5
		// In themes tab: 5 header lines + 1 (theme index 1) = offset 6
		clickY := bodyStartY + 6

		updated, _ := m.Update(tea.MouseMsg{
			Action: tea.MouseActionPress,
			Button: tea.MouseButtonLeft,
			X:      modalStartX + 10,
			Y:      clickY,
		})
		got := updated.(Model)
		if got.settings.themeCursor != 1 {
			t.Fatalf("expected clicking theme row 1 to select themeCursor=1, got %d", got.settings.themeCursor)
		}
	})

	t.Run("api_keys_click_enters_edit_mode", func(t *testing.T) {
		m := newTDDTestModel()
		m.settings.show = true
		m.settings.tab = settingsTabAPIKeys
		ids := m.apiKeysTabIDs()
		if len(ids) == 0 {
			t.Fatal("no API key IDs")
		}

		contentW := clamp(m.width-24, 68, 92)
		panelH := clamp(20, 8, m.height-14) + 9
		modalStartX := (m.width - (contentW + 2)) / 2
		modalStartY := (m.height - panelH) / 2
		bodyStartY := modalStartY + 5
		// Header lines in API keys tab: 5 header lines -> row 0 is offset 5
		clickY := bodyStartY + 5

		updated, _ := m.Update(tea.MouseMsg{
			Action: tea.MouseActionPress,
			Button: tea.MouseButtonLeft,
			X:      modalStartX + 10,
			Y:      clickY,
		})
		got := updated.(Model)
		if !got.settings.apiKeyEditing && !got.settings.browserPicker.active {
			t.Fatalf("expected clicking API key row to enter apiKeyEditing or browserPicker, got editing=%v, picker=%v", got.settings.apiKeyEditing, got.settings.browserPicker.active)
		}
	})

	t.Run("sections_click_toggles_section_and_switches_subtab", func(t *testing.T) {
		m := newTDDTestModel()
		m.settings.show = true
		m.settings.tab = settingsTabWidgetSections
		m.settings.sectionSubTab = 0

		contentW := clamp(m.width-24, 68, 92)
		panelH := clamp(20, 8, m.height-14) + 9
		modalStartX := (m.width - (contentW + 2)) / 2
		modalStartY := (m.height - panelH) / 2
		bodyStartY := modalStartY + 5

		// Click at relY = 0 (bodyStartY) -> switches subtab to 1
		updatedSubTab, _ := m.Update(tea.MouseMsg{
			Action: tea.MouseActionPress,
			Button: tea.MouseButtonLeft,
			X:      modalStartX + 10,
			Y:      bodyStartY,
		})
		gotSubTab := updatedSubTab.(Model)
		if gotSubTab.settings.sectionSubTab != 1 {
			t.Fatalf("expected clicking subtab header at relY=0 to switch sectionSubTab to 1, got %d", gotSubTab.settings.sectionSubTab)
		}

		// Click at relY = 4 (bodyStartY + 4) -> toggles section row 0
		entriesBefore := gotSubTab.detailWidgetSectionEntries()
		if len(entriesBefore) > 0 {
			initEnabled := entriesBefore[0].Enabled
			updatedRow, _ := gotSubTab.Update(tea.MouseMsg{
				Action: tea.MouseActionPress,
				Button: tea.MouseButtonLeft,
				X:      modalStartX + 10,
				Y:      bodyStartY + 4,
			})
			gotRow := updatedRow.(Model)
			entriesAfter := gotRow.detailWidgetSectionEntries()
			if entriesAfter[0].Enabled == initEnabled {
				t.Fatalf("expected clicking section row 0 to toggle enabled state from %v, got %v", initEnabled, entriesAfter[0].Enabled)
			}
		}
	})
}

func TestAnalyticsView_ScrollingAndTabSwitching(t *testing.T) {
	m := newTDDTestModel()
	m.experimentalAnalytics = true

	// 1. Click "2:Analytics" tab in header (Y=0)
	// Header format: ⚡ agentUsage 1:Dashboard 2:Analytics
	// Click near X=30 on line 0 to hit "2:Analytics"
	updated, _ := m.Update(tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
		X:      30,
		Y:      0,
	})
	got := updated.(Model)
	if got.screen != screenAnalytics {
		// Try tab key
		tabUpdated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
		got = tabUpdated.(Model)
	}
	if got.screen != screenAnalytics {
		t.Fatalf("expected switching to screenAnalytics, got %v", got.screen)
	}

	// 2. Keyboard scrolling in analytics
	got.analyticsScrollY = 0
	jUpdated, _ := got.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	gotJ := jUpdated.(Model)
	if gotJ.analyticsScrollY <= 0 {
		t.Fatalf("expected 'j' in analytics to increment analyticsScrollY, got %d", gotJ.analyticsScrollY)
	}

	kUpdated, _ := gotJ.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	gotK := kUpdated.(Model)
	if gotK.analyticsScrollY != 0 {
		t.Fatalf("expected 'k' in analytics to decrement analyticsScrollY back to 0, got %d", gotK.analyticsScrollY)
	}

	// 3. Mouse wheel scrolling in analytics
	wheelDownUpdated, _ := gotK.Update(tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonWheelDown,
		X:      m.width / 2,
		Y:      m.height / 2,
	})
	gotWheelDown := wheelDownUpdated.(Model)
	if gotWheelDown.analyticsScrollY <= 0 {
		t.Fatalf("expected mouse wheel down in analytics to increment analyticsScrollY, got %d", gotWheelDown.analyticsScrollY)
	}

	wheelUpUpdated, _ := gotWheelDown.Update(tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonWheelUp,
		X:      m.width / 2,
		Y:      m.height / 2,
	})
	gotWheelUp := wheelUpUpdated.(Model)
	if gotWheelUp.analyticsScrollY >= gotWheelDown.analyticsScrollY {
		t.Fatalf("expected mouse wheel up in analytics to decrement analyticsScrollY, got %d", gotWheelUp.analyticsScrollY)
	}

	// 4. Clicking tab "1:Dashboard" switches back
	dashUpdated, _ := gotWheelUp.Update(tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
		X:      16,
		Y:      0,
	})
	gotDash := dashUpdated.(Model)
	if gotDash.screen != screenDashboard {
		t.Fatalf("expected clicking Dashboard tab to switch back to screenDashboard, got %v", gotDash.screen)
	}
}

func TestDetailView_ScrollingNavigationAndExiting(t *testing.T) {
	m := newTDDTestModel()
	m.mode = modeDetail
	m.detailOffset = 0

	// 1. "j" scrolls down
	jUpdated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	gotJ := jUpdated.(Model)
	if gotJ.detailOffset != 1 {
		t.Fatalf("expected 'j' in detail mode to increment detailOffset to 1, got %d", gotJ.detailOffset)
	}

	// 2. "k" scrolls up
	kUpdated, _ := gotJ.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	gotK := kUpdated.(Model)
	if gotK.detailOffset != 0 {
		t.Fatalf("expected 'k' in detail mode to decrement detailOffset to 0, got %d", gotK.detailOffset)
	}

	// 3. Mouse wheel down scrolls detail down
	wheelDown, _ := gotK.Update(tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonWheelDown,
		X:      m.width / 2,
		Y:      m.height / 2,
	})
	gotWD := wheelDown.(Model)
	if gotWD.detailOffset <= 0 {
		t.Fatalf("expected mouse wheel down in detail mode to increment detailOffset, got %d", gotWD.detailOffset)
	}

	// 4. Mouse wheel up scrolls detail up
	wheelUp, _ := gotWD.Update(tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonWheelUp,
		X:      m.width / 2,
		Y:      m.height / 2,
	})
	gotWU := wheelUp.(Model)
	if gotWU.detailOffset >= gotWD.detailOffset {
		t.Fatalf("expected mouse wheel up in detail mode to decrement detailOffset, got %d", gotWU.detailOffset)
	}

	// 5. Jump to bottom "G" and top "g"
	gUpdated, _ := gotWU.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	gotG := gUpdated.(Model)
	if gotG.detailOffset < 100 {
		t.Fatalf("expected 'G' to jump detailOffset to bottom, got %d", gotG.detailOffset)
	}

	homeUpdated, _ := gotG.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	gotHome := homeUpdated.(Model)
	if gotHome.detailOffset != 0 {
		t.Fatalf("expected 'g' to reset detailOffset to 0, got %d", gotHome.detailOffset)
	}

	// 6. Clicking top header exits detail mode
	clickHeader, _ := gotHome.Update(tea.MouseMsg{
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
		X:      m.width / 2,
		Y:      2, // headerLines + 1 (line 3 is <= headerLines+1)
	})
	gotExit := clickHeader.(Model)
	if gotExit.mode != modeList {
		t.Fatalf("expected clicking header area to exit detail mode, got %v", gotExit.mode)
	}
}
