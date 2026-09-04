package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestActiveDashboardView_ReturnsConfigured(t *testing.T) {
	views := []dashboardViewMode{
		dashboardViewSplit,
		dashboardViewMatrix,
		dashboardViewBento,
		dashboardViewBars,
		dashboardViewDials,
		dashboardViewStrips,
	}

	for _, v := range views {
		m := Model{
			dashboardView: v,
			width:         120,
			sortedIDs:     []string{"a", "b", "c"},
			snapshots:     testSnapshots("a", "b", "c"),
		}

		if got := m.activeDashboardView(); got != v {
			t.Fatalf("activeDashboardView = %q, want %q", got, v)
		}
	}
}

func TestNormalizeDashboardViewMode(t *testing.T) {
	cases := []struct {
		input string
		want  dashboardViewMode
	}{
		{"split", dashboardViewSplit},
		{"matrix", dashboardViewMatrix},
		{"bento", dashboardViewBento},
		{"bars", dashboardViewBars},
		{"dials", dashboardViewDials},
		{"strips", dashboardViewStrips},
		{"grid", dashboardViewSplit},
		{"stacked", dashboardViewSplit},
		{"tabs", dashboardViewSplit},
		{"compare", dashboardViewSplit},
		{"list", dashboardViewSplit},
		{"unknown", dashboardViewSplit},
	}
	for _, c := range cases {
		if got := normalizeDashboardViewMode(c.input); got != c.want {
			t.Fatalf("normalizeDashboardViewMode(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestDashboardViewOptions_ContainsAll6Views(t *testing.T) {
	if len(dashboardViewOptions) != 6 {
		t.Fatalf("len(dashboardViewOptions) = %d, want 6", len(dashboardViewOptions))
	}
	expected := []dashboardViewMode{
		dashboardViewSplit,
		dashboardViewMatrix,
		dashboardViewBento,
		dashboardViewBars,
		dashboardViewDials,
		dashboardViewStrips,
	}
	for i, exp := range expected {
		if dashboardViewOptions[i].ID != exp {
			t.Fatalf("dashboardViewOptions[%d].ID = %q, want %q", i, dashboardViewOptions[i].ID, exp)
		}
	}
}

func TestNextDashboardView_Cycling(t *testing.T) {
	m := Model{dashboardView: dashboardViewSplit}
	order := []dashboardViewMode{
		dashboardViewSplit,
		dashboardViewMatrix,
		dashboardViewBento,
		dashboardViewBars,
		dashboardViewDials,
		dashboardViewStrips,
	}

	for i := 0; i < len(order); i++ {
		cur := order[i]
		m.dashboardView = cur
		next := m.nextDashboardView(1)
		expectedNext := order[(i+1)%len(order)]
		if next != expectedNext {
			t.Fatalf("from %q, nextDashboardView(1) = %q, want %q", cur, next, expectedNext)
		}
		prev := m.nextDashboardView(-1)
		expectedPrev := order[(i-1+len(order))%len(order)]
		if prev != expectedPrev {
			t.Fatalf("from %q, nextDashboardView(-1) = %q, want %q", cur, prev, expectedPrev)
		}
	}
}

func TestKeyboardLayoutCycling_VAndTab(t *testing.T) {
	m := Model{
		dashboardView: dashboardViewSplit,
		screen:        screenDashboard,
		mode:          modeList,
		hasData:       true,
	}

	// Press 'v' to cycle Split -> Matrix
	mUpdated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("v")})
	m = mUpdated.(Model)
	if m.activeDashboardView() != dashboardViewMatrix {
		t.Fatalf("after 'v', activeDashboardView = %q, want %q", m.activeDashboardView(), dashboardViewMatrix)
	}
	if cmd == nil {
		t.Fatal("expected persist command after 'v'")
	}

	// Press 'tab' to cycle Matrix -> Bento
	mUpdated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = mUpdated.(Model)
	if m.activeDashboardView() != dashboardViewBento {
		t.Fatalf("after 'tab', activeDashboardView = %q, want %q", m.activeDashboardView(), dashboardViewBento)
	}
	if cmd == nil {
		t.Fatal("expected persist command after 'tab'")
	}

	// Press 'V' to cycle Bento -> Matrix
	mUpdated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("V")})
	m = mUpdated.(Model)
	if m.activeDashboardView() != dashboardViewMatrix {
		t.Fatalf("after 'V', activeDashboardView = %q, want %q", m.activeDashboardView(), dashboardViewMatrix)
	}
	if cmd == nil {
		t.Fatal("expected persist command after 'V'")
	}

	// Press 'shift+tab' to cycle Matrix -> Split
	mUpdated, cmd = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	m = mUpdated.(Model)
	if m.activeDashboardView() != dashboardViewSplit {
		t.Fatalf("after 'shift+tab', activeDashboardView = %q, want %q", m.activeDashboardView(), dashboardViewSplit)
	}
	if cmd == nil {
		t.Fatal("expected persist command after 'shift+tab'")
	}
}

func TestBoardViews_NavigationAndDetailToggle(t *testing.T) {
	views := []dashboardViewMode{
		dashboardViewMatrix,
		dashboardViewBento,
		dashboardViewBars,
		dashboardViewDials,
		dashboardViewStrips,
	}

	for _, v := range views {
		t.Run(string(v), func(t *testing.T) {
			m := Model{
				dashboardView: v,
				screen:        screenDashboard,
				mode:          modeList,
				hasData:       true,
				sortedIDs:     []string{"a", "b", "c"},
				snapshots:     testSnapshots("a", "b", "c"),
				cursor:        0,
			}

			// Press 'j' to move cursor down
			mUpdated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
			m = mUpdated.(Model)
			if m.cursor != 1 {
				t.Fatalf("[%s] cursor = %d, want 1", v, m.cursor)
			}

			// Press 'k' to move cursor up
			mUpdated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
			m = mUpdated.(Model)
			if m.cursor != 0 {
				t.Fatalf("[%s] cursor = %d, want 0", v, m.cursor)
			}

			// Press 'Enter' to enter detail mode (Cockpit)
			mUpdated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
			m = mUpdated.(Model)
			if m.mode != modeDetail {
				t.Fatalf("[%s] expected modeDetail after Enter, got %v", v, m.mode)
			}

			// Press 'Esc' to exit detail mode back to board layout
			mUpdated, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
			m = mUpdated.(Model)
			if m.mode != modeList {
				t.Fatalf("[%s] expected modeList after Esc, got %v", v, m.mode)
			}
			if m.activeDashboardView() != v {
				t.Fatalf("[%s] expected active layout preserved, got %q", v, m.activeDashboardView())
			}
		})
	}
}

func TestFooterAndHelp_ShowLayoutSwitcher(t *testing.T) {
	m := Model{
		dashboardView: dashboardViewSplit,
		screen:        screenDashboard,
		mode:          modeList,
		width:         120,
		height:        30,
	}

	footer := m.renderFooterStatusLine(120)
	if !strings.Contains(footer, "v layout") {
		t.Fatalf("expected 'v layout' in footer, got: %q", footer)
	}

	help := m.renderHelpOverlay(120, 30)
	if !strings.Contains(help, "v / V / Tab") {
		t.Fatalf("expected 'v / V / Tab' in help overlay, got: %q", help)
	}
}
