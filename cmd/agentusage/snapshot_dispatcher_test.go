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
	enrichStarted := make(chan struct{})
	enrichContinue := make(chan struct{})
	enrichDone := make(chan struct{})

	d := &snapshotDispatcher{
		sendMsg: func(msg tea.Msg) {
			if sMsg, ok := msg.(tui.SnapshotsMsg); ok {
				msgCh <- sMsg
			}
		},
		enrich: func(snaps map[string]core.UsageSnapshot) {
			close(enrichStarted)
			<-enrichContinue
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
	var firstMsg tui.SnapshotsMsg
	select {
	case firstMsg = <-msgCh:
		if firstMsg.Snapshots["test-1"].Message != "base" {
			t.Fatalf("initial dispatch message = %q, want base", firstMsg.Snapshots["test-1"].Message)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for immediate dispatch message")
	}

	select {
	case <-enrichStarted:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for enrich to start")
	}

	// Mutate the original frame map and the already-sent message map the way the
	// TUI does (write Diagnostics). Neither must race with or corrupt enrich.
	frame.Snapshots["test-1"] = core.UsageSnapshot{ProviderID: "test", AccountID: "test-1", Message: "mutated-source"}
	{
		snap := firstMsg.Snapshots["test-1"]
		snap.EnsureMaps()
		snap.Diagnostics["display_branch"] = "test"
		firstMsg.Snapshots["test-1"] = snap
	}
	close(enrichContinue)

	select {
	case <-enrichDone:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for enrich to complete")
	}

	// Second message must arrive with enriched snapshots from an independent clone
	select {
	case msg := <-msgCh:
		if msg.Snapshots["test-1"].Message != "enriched" {
			t.Fatalf("enriched dispatch message = %q, want enriched", msg.Snapshots["test-1"].Message)
		}
		if firstMsg.Snapshots["test-1"].Message != "base" {
			t.Fatalf("first message mutated after send: %q", firstMsg.Snapshots["test-1"].Message)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for enriched dispatch message")
	}
}
