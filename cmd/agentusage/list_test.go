package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/nurulislamz/agentusage/internal/core"
)

func TestBuildListItems_SortingAndFields(t *testing.T) {
	accounts := []core.AccountConfig{
		{ID: "cursor-physics", Provider: "cursor", Auth: "local"},
		{ID: "antigravity-nurulz", Provider: "antigravity", Auth: "local"},
		{ID: "antigravity", Provider: "antigravity", Auth: "local"},
		{ID: "claude-code", Provider: "claude_code", Auth: "local"},
	}

	items := buildListItems(accounts)
	if len(items) != 4 {
		t.Fatalf("expected 4 items, got %d", len(items))
	}

	// Providers should be sorted alphabetically, then IDs
	if items[0].Provider != "antigravity" || items[0].ID != "antigravity" {
		t.Errorf("expected first item to be antigravity/antigravity, got %s/%s", items[0].Provider, items[0].ID)
	}
	if items[1].Provider != "antigravity" || items[1].ID != "antigravity-nurulz" {
		t.Errorf("expected second item to be antigravity/antigravity-nurulz, got %s/%s", items[1].Provider, items[1].ID)
	}
	if items[2].Provider != "claude_code" {
		t.Errorf("expected third item provider claude_code, got %s", items[2].Provider)
	}
	if items[3].Provider != "cursor" {
		t.Errorf("expected fourth item provider cursor, got %s", items[3].Provider)
	}
}

func TestRunList_JSONOutput(t *testing.T) {
	var buf bytes.Buffer
	opts := listOptions{format: "json"}

	if err := runList(&buf, opts); err != nil {
		t.Fatalf("runList error: %v", err)
	}

	var items []AccountListItem
	if err := json.Unmarshal(buf.Bytes(), &items); err != nil {
		t.Fatalf("failed to decode JSON output: %v\noutput:\n%s", err, buf.String())
	}

	if len(items) == 0 {
		t.Fatal("expected at least one account item in JSON output")
	}

	foundAntigravity := false
	for _, it := range items {
		if strings.HasPrefix(it.ID, "antigravity") {
			foundAntigravity = true
			break
		}
	}
	if !foundAntigravity {
		t.Errorf("expected antigravity account in items, got: %+v", items)
	}
}

func TestRunList_QuietIDsOnly(t *testing.T) {
	var buf bytes.Buffer
	opts := listOptions{idsOnly: true}

	if err := runList(&buf, opts); err != nil {
		t.Fatalf("runList error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) == 0 {
		t.Fatal("expected at least one ID line")
	}

	for _, line := range lines {
		if strings.Contains(line, "\t") || strings.Contains(line, " ") {
			t.Errorf("expected ID line without spaces/tabs, got %q", line)
		}
	}
}

func TestRunList_TableOutput(t *testing.T) {
	var buf bytes.Buffer
	opts := listOptions{format: "table"}

	if err := runList(&buf, opts); err != nil {
		t.Fatalf("runList error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "ID") || !strings.Contains(out, "PROVIDER") || !strings.Contains(out, "STATUS") {
		t.Errorf("missing expected table headers in output:\n%s", out)
	}
}

func TestRunList_InvalidFormat(t *testing.T) {
	var buf bytes.Buffer
	opts := listOptions{format: "xml"}

	err := runList(&buf, opts)
	if err == nil {
		t.Fatal("expected error for invalid format, got nil")
	}
	if !strings.Contains(err.Error(), "unknown format") {
		t.Errorf("unexpected error message: %v", err)
	}
}
