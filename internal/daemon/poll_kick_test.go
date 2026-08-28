package daemon

import (
	"testing"
	"time"
)

func TestRequestPoll_Coalesces(t *testing.T) {
	s := &Service{pollKick: make(chan struct{}, 1)}
	s.RequestPoll()
	s.RequestPoll()
	s.RequestPoll()

	select {
	case <-s.pollKick:
	default:
		t.Fatal("expected one pending kick")
	}
	select {
	case <-s.pollKick:
		t.Fatal("expected kicks to coalesce to a single pending signal")
	default:
	}
}

func TestRequestPoll_NilSafe(t *testing.T) {
	var s *Service
	s.RequestPoll() // must not panic
	(&Service{}).RequestPoll()
}

func TestIsAntigravityStatusFile(t *testing.T) {
	cases := []struct {
		name string
		want bool
	}{
		{"antigravity-status.json", true},
		{"antigravity-mohammed-status.json", true},
		{"/tmp/antigravity-nurulz-status.json", true},
		{".antigravity-status-abc.tmp", false},
		{"telemetry.db", false},
		{"antigravity-status.json.bak", false},
		{"other-status.json", false},
	}
	for _, tc := range cases {
		if got := isAntigravityStatusFile(tc.name); got != tc.want {
			t.Errorf("isAntigravityStatusFile(%q)=%v want %v", tc.name, got, tc.want)
		}
	}
}

func TestRequestPoll_DoesNotBlock(t *testing.T) {
	s := &Service{pollKick: make(chan struct{}, 1)}
	done := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			s.RequestPoll()
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("RequestPoll blocked under burst")
	}
}
