package webserve

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSnapshotsIncludeProjectedViews(t *testing.T) {
	srv := testServer(t, Options{Demo: true})
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
		t.Fatal("expected projected views")
	}
	if env.ThemeTokens.Base == "" {
		t.Fatal("expected theme tokens")
	}
	view := env.Views[0]
	if view.Key == "" || len(view.TileLines) == 0 {
		t.Fatalf("incomplete view: %+v", view)
	}
	if view.DetailHTML == "" {
		t.Fatal("expected detail_html from TUI RenderDetailContent")
	}
	if view.BadgeHTML == "" {
		t.Fatal("expected badge_html")
	}
	if view.IconHTML == "" {
		t.Fatal("expected icon_html")
	}
	if view.LastRefreshed == "" {
		t.Fatal("expected last_refreshed on demo views")
	}
	if len(view.UsageLines) == 0 {
		t.Fatal("expected usage_lines for navigator summaries")
	}
}
