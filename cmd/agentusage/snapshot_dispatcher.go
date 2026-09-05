package main

import (
	"context"
	"strings"
	"sync/atomic"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nurulislamz/agentusage/internal/core"
	"github.com/nurulislamz/agentusage/internal/daemon"
	"github.com/nurulislamz/agentusage/internal/tui"
)

type snapshotDispatcher struct {
	program *tea.Program
	nextID  atomic.Uint64
	enrich  func(map[string]core.UsageSnapshot)
	sendMsg func(tea.Msg)
}

func (d *snapshotDispatcher) bind(program *tea.Program) {
	d.program = program
	if program != nil {
		d.sendMsg = program.Send
	}
}

func (d *snapshotDispatcher) dispatch(frame daemon.SnapshotFrame) {
	requestID := d.nextID.Add(1)
	// Clone before send so the TUI never shares a map with the enrich goroutine
	// (TUI may write Diagnostics while DeepClone/enrich iterates or writes).
	base := core.DeepCloneSnapshots(frame.Snapshots)
	d.send(daemon.SnapshotFrame{
		Snapshots:  base,
		TimeWindow: frame.TimeWindow,
	}, requestID)
	if d == nil || d.enrich == nil || len(base) == 0 {
		return
	}
	go func() {
		// Second clone: enrich mutates the map; keep the already-sent base intact.
		enriched := core.DeepCloneSnapshots(base)
		d.enrich(enriched)
		d.send(daemon.SnapshotFrame{
			Snapshots:  enriched,
			TimeWindow: frame.TimeWindow,
		}, requestID)
	}()
}

func (d *snapshotDispatcher) refresh(ctx context.Context, rt *daemon.ViewRuntime, req tui.RefreshRequest) uint64 {
	requestID := d.nextID.Add(1)
	go func() {
		frame := rt.ReadWithFallbackForWindow(ctx, req.TimeWindow)
		d.applyEnrich(frame.Snapshots, req.AccountID)
		d.send(frame, requestID)
	}()
	return requestID
}

func (d *snapshotDispatcher) applyEnrich(snaps map[string]core.UsageSnapshot, accountID string) {
	if d == nil || d.enrich == nil || len(snaps) == 0 {
		return
	}
	accountID = strings.TrimSpace(accountID)
	if accountID == "" {
		d.enrich(snaps)
		return
	}
	subset := map[string]core.UsageSnapshot{}
	if snap, ok := snaps[accountID]; ok {
		subset[accountID] = snap
	}
	if len(subset) == 0 {
		return
	}
	d.enrich(subset)
	if snap, ok := subset[accountID]; ok {
		snaps[accountID] = snap
	}
}

func (d *snapshotDispatcher) send(frame daemon.SnapshotFrame, requestID uint64) {
	if d == nil {
		return
	}
	sendFn := d.sendMsg
	if sendFn == nil && d.program != nil {
		sendFn = d.program.Send
	}
	if sendFn == nil {
		return
	}
	sendFn(tui.SnapshotsMsg{
		Snapshots:  frame.Snapshots,
		TimeWindow: frame.TimeWindow,
		RequestID:  requestID,
	})
}
