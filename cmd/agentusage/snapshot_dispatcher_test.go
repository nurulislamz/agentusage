package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nurulislamz/agentusage/internal/core"
	"github.com/nurulislamz/agentusage/internal/daemon"
	"github.com/nurulislamz/agentusage/internal/tui"
)

type mockDaemonRT func(*http.Request) (*http.Response, error)

func (m mockDaemonRT) RoundTrip(r *http.Request) (*http.Response, error) {
	return m(r)
}

func TestSnapshotDispatcher_DispatchSendsSnapshotsMsg(t *testing.T) {
	msgCh := make(chan tui.SnapshotsMsg, 10)

	d := &snapshotDispatcher{
		sendMsg: func(msg tea.Msg) {
			if sMsg, ok := msg.(tui.SnapshotsMsg); ok {
				msgCh <- sMsg
			}
		},
	}

	frame := daemon.SnapshotFrame{
		TimeWindow: core.TimeWindow30d,
		Snapshots: map[string]core.UsageSnapshot{
			"test-1": {ProviderID: "test", AccountID: "test-1", Message: "base"},
		},
	}

	d.dispatch(frame)

	select {
	case msg := <-msgCh:
		if msg.RequestID == 0 {
			t.Error("expected non-zero RequestID")
		}
		if msg.TimeWindow != core.TimeWindow30d {
			t.Errorf("got TimeWindow %v, want 30d", msg.TimeWindow)
		}
		if msg.Snapshots["test-1"].Message != "base" {
			t.Fatalf("dispatch message = %q, want base", msg.Snapshots["test-1"].Message)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for dispatch message")
	}
}

func TestSnapshotDispatcher_RefreshRoutesThroughViewRuntime(t *testing.T) {
	msgCh := make(chan tui.SnapshotsMsg, 10)

	d := &snapshotDispatcher{
		sendMsg: func(msg tea.Msg) {
			if sMsg, ok := msg.(tui.SnapshotsMsg); ok {
				msgCh <- sMsg
			}
		},
	}

	pollCount := 0
	readModelCount := 0
	mockClient := daemon.NewMockClient("/tmp/mock.sock", mockDaemonRT(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path == "/v1/poll" {
			pollCount++
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"status":"polled"}`))),
				Header:     make(http.Header),
			}, nil
		}
		if req.URL.Path == "/v1/read-model" {
					readModelCount++
					resp := daemon.ReadModelResponse{
						Snapshots: map[string]core.UsageSnapshot{
							"acct-refreshed": {
								ProviderID: "prov",
								AccountID:  "acct-refreshed",
								Status:     core.StatusOK,
							},
						},
					}
					body, _ := json.Marshal(resp)
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(bytes.NewReader(body)),
						Header:     make(http.Header),
					}, nil
				}
				return &http.Response{StatusCode: http.StatusNotFound}, nil
			}))

	rt := daemon.NewViewRuntime(nil, "/tmp/mock.sock", false)
	rt.SetClient(mockClient)

	reqID := d.refresh(context.Background(), rt, tui.RefreshRequest{
		TimeWindow: core.TimeWindow7d,
	})
	if reqID == 0 {
		t.Fatal("expected non-zero requestID returned by refresh")
	}

	select {
	case msg := <-msgCh:
		if msg.RequestID != reqID {
			t.Errorf("msg.RequestID = %d, want %d", msg.RequestID, reqID)
		}
		if msg.TimeWindow != core.TimeWindow7d {
			t.Errorf("msg.TimeWindow = %v, want 7d", msg.TimeWindow)
		}
		if _, ok := msg.Snapshots["acct-refreshed"]; !ok {
			t.Fatalf("expected acct-refreshed in snapshots: %+v", msg.Snapshots)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for refresh SnapshotsMsg")
	}

	if pollCount != 1 {
		t.Errorf("pollCount = %d, want 1", pollCount)
	}
	if readModelCount != 1 {
		t.Errorf("readModelCount = %d, want 1", readModelCount)
	}
}
