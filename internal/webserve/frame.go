package webserve

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/nurulislamz/agentusage/internal/config"
	"github.com/nurulislamz/agentusage/internal/core"
	"github.com/nurulislamz/agentusage/internal/tui"
)

const (
	defaultFrameWidth  = 120
	defaultFrameHeight = 40
)

// renderTUIFrame builds one full TUI dashboard frame with cursorIdx selected.
// Uses the real Bubble Tea model Update/View path so the HTML matches the
// terminal byte-for-byte (aside from ANSI→HTML conversion).
func renderTUIFrame(cfg config.Config, snaps []core.UsageSnapshot, cursorIdx, width, height int) string {
	ensureTrueColor()
	if width < 60 {
		width = defaultFrameWidth
	}
	if height < 16 {
		height = defaultFrameHeight
	}

	accounts := core.MergeAccounts(cfg.Accounts, cfg.AutoDetectedAccounts)
	if len(accounts) == 0 {
		accounts = accountsFromSnapshots(snaps)
	}

	tw := core.ParseTimeWindow(cfg.Data.TimeWindow)
	m := tui.NewModel(
		cfg.UI.WarnThreshold,
		cfg.UI.CritThreshold,
		cfg.Experimental.Analytics,
		cfg.Dashboard,
		accounts,
		tw,
	)

	snapMap := tui.SnapshotsToMap(snaps)
	if len(snapMap) == 0 {
		return ""
	}

	updated, _ := m.Update(tea.WindowSizeMsg{Width: width, Height: height})
	m = updated.(tui.Model)
	updated, _ = m.Update(tui.SnapshotsMsg{Snapshots: snapMap, TimeWindow: tw})
	m = updated.(tui.Model)

	if cursorIdx < 0 {
		cursorIdx = 0
	}
	for i := 0; i < cursorIdx; i++ {
		updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		m = updated.(tui.Model)
	}

	return m.View()
}

func accountsFromSnapshots(snaps []core.UsageSnapshot) []core.AccountConfig {
	out := make([]core.AccountConfig, 0, len(snaps))
	seen := make(map[string]bool, len(snaps))
	for _, snap := range snaps {
		id := strings.TrimSpace(snap.AccountID)
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, core.AccountConfig{
			ID:       id,
			Provider: snap.ProviderID,
		})
	}
	return out
}
