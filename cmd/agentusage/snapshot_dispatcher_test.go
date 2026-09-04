package main

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nurulislamz/agentusage/internal/core"
	"github.com/nurulislamz/agentusage/internal/daemon"
	"github.com/nurulislamz/agentusage/internal/tui"
)

func TestApplyEnrich_ScopesToAccount(t *testing.T) {
	d := &snapshotDispatcher{
		enrich: func(snaps map[string]core.UsageSnapshot) {
			for id := range snaps {
				snap := snaps[id]
				snap.Message = "live"
				snaps[id] = snap
			}
		},
	}

	snaps := map[string]core.UsageSnapshot{
		"cursor-main": {ProviderID: "cursor", AccountID: "cursor-main", Message: "old"},
		"opencode":    {ProviderID: "opencode", AccountID: "opencode", Message: "old"},
	}
	d.applyEnrich(snaps, "opencode")

	if snaps["opencode"].Message != "live" {
		t.Fatalf("opencode message = %q, want live", snaps["opencode"].Message)
	}
	if snaps["cursor-main"].Message != "old" {
		t.Fatalf("cursor-main message = %q, want unchanged old", snaps["cursor-main"].Message)
	}
}

func TestSnapshotDispatcher_DispatchSendsImmediatelyAndEnrichesAsync(t *testing.T) {
	msgCh := make(chan tui.SnapshotsMsg, 10)
	enrichDone := make(chan struct{})

	d := &snapshotDispatcher{
		sendMsg: func(msg tea.Msg) {
			if sMsg, ok := msg.(tui.SnapshotsMsg); ok {
				msgCh <- sMsg
			}
		},
		enrich: func(snaps map[string]core.UsageSnapshot) {
			for id := range snaps {
				snap := snaps[id]
				snap.Message = "enriched"
				snaps[id] = snap
			}
			close(enrichDone)
		},
	}

	frame := daemon.SnapshotFrame{
		TimeWindow: core.TimeWindow30d,
		Snapshots: map[string]core.UsageSnapshot{
			"test-1": {ProviderID: "test", AccountID: "test-1", Message: "base"},
		},
	}

	d.dispatch(frame)

	// First message must arrive immediately with base snapshots
	select {
	case msg := <-msgCh:
		if msg.Snapshots["test-1"].Message != "base" {
			t.Fatalf("initial dispatch message = %q, want base", msg.Snapshots["test-1"].Message)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for immediate dispatch message")
	}

	// Wait for enrich to finish
	select {
	case <-enrichDone:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for enrich to complete")
	}

	// Second message must arrive with enriched snapshots
	select {
	case msg := <-msgCh:
		if msg.Snapshots["test-1"].Message != "enriched" {
			t.Fatalf("enriched dispatch message = %q, want enriched", msg.Snapshots["test-1"].Message)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for enriched dispatch message")
	}
}
