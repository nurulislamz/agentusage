package webserve

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/janekbaraniewski/openusage/internal/config"
	"github.com/janekbaraniewski/openusage/internal/core"
	"github.com/janekbaraniewski/openusage/internal/tui"
)

func TestRenderTUIFrame_MatchesBrandAndAccount(t *testing.T) {
	ensureTrueColor()
	_ = tui.ThemeTokensForName("Deep Space")

	now := time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC)
	snaps := []core.UsageSnapshot{demoClaude(now), demoCursor(now)}
	cfg := config.DefaultConfig()
	cfg.Theme = "Deep Space"
	cfg.Dashboard.UsageMode = config.UsageModeRemaining

	frame := renderTUIFrame(cfg, snaps, 0, 120, 36)
	plain := tui.StripANSI(frame)
	if !strings.Contains(plain, "OpenUsage") {
		t.Fatalf("missing brand:\n%s", plain)
	}
	if !strings.Contains(plain, "claude-code") {
		t.Fatalf("missing selected account:\n%s", plain)
	}
	if !strings.Contains(plain, "Usage") {
		t.Fatalf("missing detail usage section:\n%s", plain)
	}

	frame1 := renderTUIFrame(cfg, snaps, 1, 120, 36)
	plain1 := tui.StripANSI(frame1)
	if !strings.Contains(plain1, "cursor-ide") {
		t.Fatalf("cursor selection missing cursor-ide:\n%s", plain1)
	}
}

func TestSnapshotsIncludeFrameHTML(t *testing.T) {
	srv := testServer(t, Options{Demo: true, Theme: "Deep Space"})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/snapshots", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var env Envelope
	if err := json.NewDecoder(w.Body).Decode(&env); err != nil {
		t.Fatal(err)
	}
	if len(env.Views) == 0 {
		t.Fatal("expected views")
	}
	if env.Views[0].FrameHTML == "" {
		t.Fatal("expected frame_html from full TUI render")
	}
	if !strings.Contains(env.Views[0].FrameHTML, "OpenUsage") && !strings.Contains(env.Views[0].FrameHTML, "span") {
		t.Fatalf("frame_html looks empty/plain: %q", env.Views[0].FrameHTML[:min(80, len(env.Views[0].FrameHTML))])
	}
}
