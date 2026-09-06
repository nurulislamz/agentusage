package main

import (
	"context"
	"sync/atomic"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nurulislamz/agentusage/internal/daemon"
	"github.com/nurulislamz/agentusage/internal/tui"
)

type snapshotDispatcher struct {
	program *tea.Program
	nextID  atomic.Uint64
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
	d.send(frame, requestID)
}

func (d *snapshotDispatcher) refresh(ctx context.Context, rt *daemon.ViewRuntime, req tui.RefreshRequest) uint64 {
	requestID := d.nextID.Add(1)
	go func() {
		frame := rt.RefreshForWindow(ctx, req.TimeWindow)
		d.send(frame, requestID)
	}()
	return requestID
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
