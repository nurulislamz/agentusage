package crush

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/nurulislamz/agentusage/internal/core"
)

type fixedClock struct {
	t time.Time
}

func (f fixedClock) Now() time.Time {
	return f.t
}

type testDBOptions struct {
	hasProviderColumn bool
	missingSessions   bool
	missingMessages   bool
	sessions          []testSessionRow
	messages          []testMessageRow
}

type testSessionRow struct {
	id               string
	parentSessionID  *string
	messageCount     *int64
	promptTokens     *int64
	completionTokens *int64
	cost             *float64
	createdAt        *int64
	updatedAt        *int64
}

type testMessageRow struct {
	id        string
	sessionID string
	role      string
	model     string
	provider  string
	createdAt int64
}

func ptrInt64(v int64) *int64       { return &v }
func ptrFloat64(v float64) *float64 { return &v }
func ptrString(v string) *string    { return &v }

func createTestCrushDB(t *testing.T, dbPath string, opts testDBOptions) string {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(dbPath), err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open sqlite %s: %v", dbPath, err)
	}
	defer db.Close()

	if !opts.missingSessions {
		_, err := db.Exec(`
			CREATE TABLE sessions (
				id TEXT PRIMARY KEY,
				parent_session_id TEXT,
				message_count INTEGER,
				prompt_tokens INTEGER,
				completion_tokens INTEGER,
				cost REAL,
				created_at INTEGER,
				updated_at INTEGER
			);
		`)
		if err != nil {
			t.Fatalf("create sessions table: %v", err)
		}

		for _, s := range opts.sessions {
			_, err := db.Exec(`
				INSERT INTO sessions (id, parent_session_id, message_count, prompt_tokens, completion_tokens, cost, created_at, updated_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			`, s.id, s.parentSessionID, s.messageCount, s.promptTokens, s.completionTokens, s.cost, s.createdAt, s.updatedAt)
			if err != nil {
				t.Fatalf("insert session %s: %v", s.id, err)
			}
		}
	}

	if !opts.missingMessages {
		if opts.hasProviderColumn {
			_, err := db.Exec(`
				CREATE TABLE messages (
					id TEXT PRIMARY KEY,
					session_id TEXT,
					role TEXT,
					model TEXT,
					provider TEXT,
					created_at INTEGER
				);
			`)
			if err != nil {
				t.Fatalf("create messages table: %v", err)
			}
			for _, m := range opts.messages {
				_, err := db.Exec(`
					INSERT INTO messages (id, session_id, role, model, provider, created_at)
					VALUES (?, ?, ?, ?, ?, ?)
				`, m.id, m.sessionID, m.role, m.model, m.provider, m.createdAt)
				if err != nil {
					t.Fatalf("insert message %s: %v", m.id, err)
				}
			}
		} else {
			_, err := db.Exec(`
				CREATE TABLE messages (
					id TEXT PRIMARY KEY,
					session_id TEXT,
					role TEXT,
					model TEXT,
					created_at INTEGER
				);
			`)
			if err != nil {
				t.Fatalf("create messages table: %v", err)
			}
			for _, m := range opts.messages {
				_, err := db.Exec(`
					INSERT INTO messages (id, session_id, role, model, created_at)
					VALUES (?, ?, ?, ?, ?)
				`, m.id, m.sessionID, m.role, m.model, m.createdAt)
				if err != nil {
					t.Fatalf("insert message %s: %v", m.id, err)
				}
			}
		}
	}

	return dbPath
}

// -----------------------------------------------------------------------------
// Axis 1: Happy Paths
// -----------------------------------------------------------------------------

func TestProvider_Fetch_HappyPath_Complete(t *testing.T) {
	tmp := t.TempDir()
	now := time.Date(2026, 8, 31, 15, 0, 0, 0, time.UTC)
	todayMS := now.UnixMilli()
	twoDaysAgoMS := now.AddDate(0, 0, -2).UnixMilli()
	tenDaysAgoMS := now.AddDate(0, 0, -10).UnixMilli()

	dbPath := filepath.Join(tmp, "project1", ".crush", "crush.db")
	createTestCrushDB(t, dbPath, testDBOptions{
		hasProviderColumn: true,
		sessions: []testSessionRow{
			{
				id:               "sess-today",
				messageCount:     ptrInt64(10),
				promptTokens:     ptrInt64(1000),
				completionTokens: ptrInt64(200),
				cost:             ptrFloat64(0.045),
				createdAt:        ptrInt64(todayMS),
				updatedAt:        ptrInt64(todayMS),
			},
			{
				id:               "sess-2d",
				messageCount:     ptrInt64(5),
				promptTokens:     ptrInt64(500),
				completionTokens: ptrInt64(100),
				cost:             ptrFloat64(0.020),
				createdAt:        ptrInt64(twoDaysAgoMS),
				updatedAt:        ptrInt64(twoDaysAgoMS),
			},
			{
				id:               "sess-10d",
				messageCount:     ptrInt64(8),
				promptTokens:     ptrInt64(800),
				completionTokens: ptrInt64(150),
				cost:             ptrFloat64(0.035),
				createdAt:        ptrInt64(tenDaysAgoMS),
				updatedAt:        ptrInt64(tenDaysAgoMS),
			},
		},
		messages: []testMessageRow{
			{
				id:        "msg-1",
				sessionID: "sess-today",
				role:      "assistant",
				model:     "claude-3-5-sonnet-20241022",
				provider:  "anthropic",
				createdAt: todayMS,
			},
			{
				id:        "msg-2",
				sessionID: "sess-2d",
				role:      "assistant",
				model:     "claude-3-5-sonnet-20241022",
				provider:  "anthropic",
				createdAt: twoDaysAgoMS,
			},
			{
				id:        "msg-3",
				sessionID: "sess-10d",
				role:      "assistant",
				model:     "gpt-4o",
				provider:  "openai",
				createdAt: tenDaysAgoMS,
			},
		},
	})

	provider := New()
	provider.clock = fixedClock{t: now}

	acct := core.AccountConfig{ID: "test-crush"}
	acct.SetPath(PathHintSingleDBKey, dbPath)

	snap, err := provider.Fetch(context.Background(), acct)
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}

	if snap.Status != core.StatusOK {
		t.Errorf("Status = %v, want StatusOK; message: %s", snap.Status, snap.Message)
	}

	// Verify Metrics
	metrics := snap.Metrics
	if got := *metrics["total_sessions"].Used; got != 3 {
		t.Errorf("total_sessions = %v, want 3", got)
	}
	if got := *metrics["sessions_today"].Used; got != 1 {
		t.Errorf("sessions_today = %v, want 1", got)
	}
	if got := *metrics["sessions_7d"].Used; got != 2 {
		t.Errorf("sessions_7d = %v, want 2", got)
	}
	if got := *metrics["total_tokens"].Used; got != 2750 {
		t.Errorf("total_tokens = %v, want 2750", got)
	}
	if got := *metrics["total_input_tokens"].Used; got != 2300 {
		t.Errorf("total_input_tokens = %v, want 2300", got)
	}
	if got := *metrics["total_output_tokens"].Used; got != 450 {
		t.Errorf("total_output_tokens = %v, want 450", got)
	}
	if got := *metrics["total_cost_usd"].Used; got < 0.099 || got > 0.101 {
		t.Errorf("total_cost_usd = %v, want ~0.10", got)
	}
	if got := *metrics["total_projects"].Used; got != 1 {
		t.Errorf("total_projects = %v, want 1", got)
	}

	// Verify DailySeries
	if len(snap.DailySeries["sessions"]) != 3 {
		t.Errorf("DailySeries[sessions] count = %d, want 3", len(snap.DailySeries["sessions"]))
	}
	if len(snap.DailySeries["tokens"]) != 3 {
		t.Errorf("DailySeries[tokens] count = %d, want 3", len(snap.DailySeries["tokens"]))
	}
	if len(snap.DailySeries["cost_usd"]) != 3 {
		t.Errorf("DailySeries[cost_usd] count = %d, want 3", len(snap.DailySeries["cost_usd"]))
	}

	// Verify ModelUsage breakdown
	if len(snap.ModelUsage) != 2 {
		t.Fatalf("ModelUsage count = %d, want 2", len(snap.ModelUsage))
	}
	var claudeRec, gptRec *core.ModelUsageRecord
	for i := range snap.ModelUsage {
		rec := &snap.ModelUsage[i]
		if rec.RawModelID == "claude-3-5-sonnet-20241022" {
			claudeRec = rec
		} else if rec.RawModelID == "gpt-4o" {
			gptRec = rec
		}
	}
	if claudeRec == nil || gptRec == nil {
		t.Fatalf("Expected both claude and gpt-4o records in ModelUsage")
	}
	if *claudeRec.TotalTokens != 1800 {
		t.Errorf("claude TotalTokens = %v, want 1800", *claudeRec.TotalTokens)
	}
	if claudeRec.Dimensions["upstream_provider"] != "anthropic" {
		t.Errorf("claude upstream_provider = %q, want anthropic", claudeRec.Dimensions["upstream_provider"])
	}
	if gptRec.Dimensions["upstream_provider"] != "openai" {
		t.Errorf("gpt-4o upstream_provider = %q, want openai", gptRec.Dimensions["upstream_provider"])
	}

	// Verify Status Message
	if snap.Message == "" || snap.Message == "OK" {
		t.Errorf("Message = %q, want formatted summary", snap.Message)
	}
}

func TestProvider_Fetch_MultipleProjects_Aggregation(t *testing.T) {
	tmp := t.TempDir()
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	nowMS := now.UnixMilli()

	dbA := filepath.Join(tmp, "projectA", ".crush", "crush.db")
	dbB := filepath.Join(tmp, "projectB", ".crush", "crush.db")

	createTestCrushDB(t, dbA, testDBOptions{
		hasProviderColumn: true,
		sessions: []testSessionRow{
			{
				id:               "sess-a1",
				messageCount:     ptrInt64(2),
				promptTokens:     ptrInt64(300),
				completionTokens: ptrInt64(100),
				cost:             ptrFloat64(0.50),
				createdAt:        ptrInt64(nowMS),
			},
		},
		messages: []testMessageRow{
			{
				id:        "msg-a1",
				sessionID: "sess-a1",
				role:      "assistant",
				model:     "claude-3-5-sonnet-20241022",
				provider:  "anthropic",
				createdAt: nowMS,
			},
		},
	})

	createTestCrushDB(t, dbB, testDBOptions{
		hasProviderColumn: true,
		sessions: []testSessionRow{
			{
				id:               "sess-b1",
				messageCount:     ptrInt64(4),
				promptTokens:     ptrInt64(700),
				completionTokens: ptrInt64(200),
				cost:             ptrFloat64(1.25),
				createdAt:        ptrInt64(nowMS),
			},
		},
		messages: []testMessageRow{
			{
				id:        "msg-b1",
				sessionID: "sess-b1",
				role:      "assistant",
				model:     "gemini-2.0-flash",
				provider:  "google",
				createdAt: nowMS,
			},
		},
	})

	provider := New()
	provider.clock = fixedClock{t: now}

	acct := core.AccountConfig{ID: "crush-multi"}
	acct.SetPath(PathHintDBsKey, dbA+string(os.PathListSeparator)+dbB)

	snap, err := provider.Fetch(context.Background(), acct)
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}

	if snap.Status != core.StatusOK {
		t.Errorf("Status = %v, want StatusOK", snap.Status)
	}
	if got := *snap.Metrics["total_sessions"].Used; got != 2 {
		t.Errorf("total_sessions = %v, want 2", got)
	}
	if got := *snap.Metrics["total_tokens"].Used; got != 1300 {
		t.Errorf("total_tokens = %v, want 1300", got)
	}
	if got := *snap.Metrics["total_cost_usd"].Used; got < 1.74 || got > 1.76 {
		t.Errorf("total_cost_usd = %v, want ~1.75", got)
	}
	if got := *snap.Metrics["total_projects"].Used; got != 2 {
		t.Errorf("total_projects = %v, want 2", got)
	}
	if snap.Raw["db_paths.0"] != dbA || snap.Raw["db_paths.1"] != dbB {
		t.Errorf("Raw db_paths mismatch: %v, %v", snap.Raw["db_paths.0"], snap.Raw["db_paths.1"])
	}
	if snap.Raw["db_count"] != "2" {
		t.Errorf("Raw db_count = %q, want 2", snap.Raw["db_count"])
	}
}

func TestProvider_Fetch_OlderSchemaWithoutProviderColumn(t *testing.T) {
	tmp := t.TempDir()
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	nowMS := now.UnixMilli()

	dbPath := filepath.Join(tmp, "legacy", ".crush", "crush.db")
	createTestCrushDB(t, dbPath, testDBOptions{
		hasProviderColumn: false, // Older schema without provider column
		sessions: []testSessionRow{
			{
				id:               "sess-legacy",
				messageCount:     ptrInt64(3),
				promptTokens:     ptrInt64(400),
				completionTokens: ptrInt64(100),
				createdAt:        ptrInt64(nowMS),
			},
		},
		messages: []testMessageRow{
			{
				id:        "msg-leg",
				sessionID: "sess-legacy",
				role:      "assistant",
				model:     "claude-3-haiku",
				createdAt: nowMS,
			},
		},
	})

	provider := New()
	provider.clock = fixedClock{t: now}
	acct := core.AccountConfig{ID: "crush"}
	acct.SetPath(PathHintSingleDBKey, dbPath)

	snap, err := provider.Fetch(context.Background(), acct)
	if err != nil {
		t.Fatalf("Fetch() error on legacy schema: %v", err)
	}

	if snap.Status != core.StatusOK {
		t.Errorf("Status = %v, want StatusOK", snap.Status)
	}
	if len(snap.ModelUsage) != 1 {
		t.Fatalf("ModelUsage count = %d, want 1", len(snap.ModelUsage))
	}
	if snap.ModelUsage[0].RawModelID != "claude-3-haiku" {
		t.Errorf("RawModelID = %q, want claude-3-haiku", snap.ModelUsage[0].RawModelID)
	}
	if snap.ModelUsage[0].Dimensions != nil && snap.ModelUsage[0].Dimensions["upstream_provider"] != "" {
		t.Errorf("upstream_provider should be empty for older schema, got %q", snap.ModelUsage[0].Dimensions["upstream_provider"])
	}
}

func TestProvider_Fetch_FiltersChildSessions(t *testing.T) {
	tmp := t.TempDir()
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	nowMS := now.UnixMilli()

	parentID := "sess-root"
	childID := "sess-child"

	dbPath := filepath.Join(tmp, "subagents", ".crush", "crush.db")
	createTestCrushDB(t, dbPath, testDBOptions{
		hasProviderColumn: true,
		sessions: []testSessionRow{
			{
				id:               parentID,
				parentSessionID:  nil, // Root session
				messageCount:     ptrInt64(10),
				promptTokens:     ptrInt64(2000),
				completionTokens: ptrInt64(500),
				cost:             ptrFloat64(0.15),
				createdAt:        ptrInt64(nowMS),
			},
			{
				id:               childID,
				parentSessionID:  &parentID, // Forked child session — should be skipped
				messageCount:     ptrInt64(5),
				promptTokens:     ptrInt64(1000),
				completionTokens: ptrInt64(250),
				cost:             ptrFloat64(0.075),
				createdAt:        ptrInt64(nowMS),
			},
		},
		messages: []testMessageRow{
			{
				id:        "msg-p",
				sessionID: parentID,
				role:      "assistant",
				model:     "claude-3-5-sonnet-20241022",
				provider:  "anthropic",
				createdAt: nowMS,
			},
			{
				id:        "msg-c",
				sessionID: childID,
				role:      "assistant",
				model:     "claude-3-5-sonnet-20241022",
				provider:  "anthropic",
				createdAt: nowMS,
			},
		},
	})

	provider := New()
	provider.clock = fixedClock{t: now}
	acct := core.AccountConfig{ID: "crush"}
	acct.SetPath(PathHintSingleDBKey, dbPath)

	snap, err := provider.Fetch(context.Background(), acct)
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}

	if got := *snap.Metrics["total_sessions"].Used; got != 1 {
		t.Errorf("total_sessions = %v, want 1 (child must be filtered)", got)
	}
	if got := *snap.Metrics["total_tokens"].Used; got != 2500 {
		t.Errorf("total_tokens = %v, want 2500", got)
	}
}

func TestProvider_ItemizedUsage(t *testing.T) {
	tmp := t.TempDir()
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	nowMS := now.UnixMilli()

	dbPath := filepath.Join(tmp, "project", ".crush", "crush.db")
	createTestCrushDB(t, dbPath, testDBOptions{
		hasProviderColumn: true,
		sessions: []testSessionRow{
			{
				id:               "sess-item-1",
				messageCount:     ptrInt64(4),
				promptTokens:     ptrInt64(600),
				completionTokens: ptrInt64(150),
				cost:             ptrFloat64(0.08),
				createdAt:        ptrInt64(nowMS),
			},
		},
		messages: []testMessageRow{
			{
				id:        "msg-item-1",
				sessionID: "sess-item-1",
				role:      "assistant",
				model:     "claude-3-5-sonnet-20241022",
				provider:  "anthropic",
				createdAt: nowMS,
			},
		},
	})

	registry := writeRegistry(t, filepath.Join(tmp, "registry"), []crushProject{
		{Path: filepath.Join(tmp, "project"), DataDir: ".crush"},
	})
	t.Setenv(EnvRegistry, registry)

	provider := New()
	events, err := provider.ItemizedUsage()
	if err != nil {
		t.Fatalf("ItemizedUsage() error: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("ItemizedUsage() returned %d events, want 1", len(events))
	}

	ev := events[0]
	if ev.ProviderID != ID {
		t.Errorf("ProviderID = %q, want %q", ev.ProviderID, ID)
	}
	if ev.Session != "sess-item-1" {
		t.Errorf("Session = %q, want sess-item-1", ev.Session)
	}
	if ev.Model != "claude-3-5-sonnet-20241022" {
		t.Errorf("Model = %q, want claude-3-5-sonnet-20241022", ev.Model)
	}
	if ev.InputTokens != 600 {
		t.Errorf("InputTokens = %d, want 600", ev.InputTokens)
	}
	if ev.OutputTokens != 150 {
		t.Errorf("OutputTokens = %d, want 150", ev.OutputTokens)
	}
	if ev.CostUSD != 0.08 || !ev.HasCost {
		t.Errorf("CostUSD = %v, HasCost = %v, want 0.08, true", ev.CostUSD, ev.HasCost)
	}
}

func TestProvider_ItemizedUsage_SkippedOnError(t *testing.T) {
	tmp := t.TempDir()
	corruptDB := filepath.Join(tmp, "bad", ".crush", "crush.db")
	seedDB(t, corruptDB) // raw bytes, not a real sqlite db

	registry := writeRegistry(t, filepath.Join(tmp, "registry"), []crushProject{
		{Path: filepath.Join(tmp, "bad"), DataDir: ".crush"},
	})
	t.Setenv(EnvRegistry, registry)

	provider := New()
	events, err := provider.ItemizedUsage()
	if err != nil {
		t.Fatalf("ItemizedUsage() unexpected error: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("ItemizedUsage() should return 0 events on corrupt DB, got %d", len(events))
	}
}

func TestProvider_HasChanged(t *testing.T) {
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "project", ".crush", "crush.db")
	seedDB(t, dbPath)

	provider := New()

	// 1. Account with no DBs -> returns false, nil
	acctEmpty := core.AccountConfig{ID: "empty"}
	changed, err := provider.HasChanged(acctEmpty, time.Now())
	if err != nil || changed {
		t.Errorf("HasChanged(empty) = (%v, %v), want (false, nil)", changed, err)
	}

	// 2. DB modified in the past, since is before mtime -> true
	past := time.Now().Add(-1 * time.Hour)
	acctValid := core.AccountConfig{ID: "valid"}
	acctValid.SetPath(PathHintSingleDBKey, dbPath)

	changed, err = provider.HasChanged(acctValid, past)
	if err != nil || !changed {
		t.Errorf("HasChanged(past) = (%v, %v), want (true, nil)", changed, err)
	}

	// 3. since is in the future -> false
	future := time.Now().Add(1 * time.Hour)
	changed, err = provider.HasChanged(acctValid, future)
	if err != nil || changed {
		t.Errorf("HasChanged(future) = (%v, %v), want (false, nil)", changed, err)
	}
}

func TestDiscoverDBPaths_Integration(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmp)

	projDir := filepath.Join(tmp, "registered-proj")
	seedDB(t, filepath.Join(projDir, ".crush", "crush.db"))

	writeRegistry(t, filepath.Join(tmp, "crush"), []crushProject{
		{Path: projDir, DataDir: ".crush"},
	})

	dbs := DiscoverDBPaths()
	if len(dbs) != 1 {
		t.Fatalf("DiscoverDBPaths() returned %d DBs, want 1", len(dbs))
	}
}

func TestWidgets_DashboardAndDetail(t *testing.T) {
	p := New()
	detail := p.DetailWidget()
	if detail.Sections == nil && len(detail.Sections) == 0 {
		// DetailWidget returns a core.DetailWidget struct
	}

	dash := dashboardWidget()
	if dash.ColorRole != core.DashboardColorRoleSky {
		t.Errorf("ColorRole = %v, want Sky", dash.ColorRole)
	}
	if len(dash.CompactRows) != 4 {
		t.Errorf("CompactRows len = %d, want 4", len(dash.CompactRows))
	}
}

// -----------------------------------------------------------------------------
// Axis 2: Edge / Boundary Cases
// -----------------------------------------------------------------------------

func TestQuerySessions_FilterZeroMessagesZeroCost(t *testing.T) {
	tmp := t.TempDir()
	now := time.Now().UnixMilli()
	dbPath := filepath.Join(tmp, "test_filter.db")

	createTestCrushDB(t, dbPath, testDBOptions{
		hasProviderColumn: true,
		sessions: []testSessionRow{
			{
				id:           "sess-zero-both",
				messageCount: ptrInt64(0),
				cost:         ptrFloat64(0.0),
				createdAt:    ptrInt64(now),
			},
			{
				id:           "sess-has-cost-only",
				messageCount: ptrInt64(0),
				cost:         ptrFloat64(0.01),
				createdAt:    ptrInt64(now),
			},
			{
				id:           "sess-has-msgs-only",
				messageCount: ptrInt64(2),
				cost:         ptrFloat64(0.0),
				createdAt:    ptrInt64(now),
			},
		},
	})

	sessions, err := querySessions(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("querySessions error: %v", err)
	}

	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(sessions))
	}
	ids := map[string]bool{sessions[0].ID: true, sessions[1].ID: true}
	if !ids["sess-has-cost-only"] || !ids["sess-has-msgs-only"] {
		t.Errorf("unexpected session IDs returned: %+v", ids)
	}
}

func TestQuerySessions_NullAndNegativeValues(t *testing.T) {
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "null_and_neg.db")

	createTestCrushDB(t, dbPath, testDBOptions{
		hasProviderColumn: true,
		sessions: []testSessionRow{
			{
				id:               "sess-nulls",
				messageCount:     ptrInt64(1),
				promptTokens:     nil,
				completionTokens: nil,
				cost:             nil,
				createdAt:        nil,
				updatedAt:        nil,
			},
			{
				id:               "sess-negatives",
				messageCount:     ptrInt64(1),
				promptTokens:     ptrInt64(-100),
				completionTokens: ptrInt64(-50),
				cost:             ptrFloat64(-0.5),
				createdAt:        ptrInt64(-1234),
				updatedAt:        ptrInt64(0),
			},
		},
	})

	sessions, err := querySessions(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("querySessions error: %v", err)
	}

	if len(sessions) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(sessions))
	}

	for _, s := range sessions {
		if s.PromptTokens < 0 || s.CompletionTokens < 0 {
			t.Errorf("session %s has negative tokens: prompt=%d, compl=%d", s.ID, s.PromptTokens, s.CompletionTokens)
		}
		if s.HasCost {
			t.Errorf("session %s has HasCost=true unexpectedly", s.ID)
		}
		if !s.CreatedAt.IsZero() && s.ID == "sess-negatives" {
			t.Errorf("session %s CreatedAt should be zero, got %v", s.ID, s.CreatedAt)
		}
	}
}

func TestPopulateSnapshot_CreatedAtFallbackToUpdatedAt(t *testing.T) {
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	updatedTime := now.AddDate(0, 0, -1) // yesterday

	sessions := []crushSession{
		{
			ID:               "fallback-sess",
			PromptTokens:     100,
			CompletionTokens: 50,
			Cost:             0.05,
			HasCost:          true,
			CreatedAt:        time.Time{}, // Zero
			UpdatedAt:        updatedTime,
			Model:            "test-model",
		},
		{
			ID:               "no-date-sess",
			PromptTokens:     200,
			CompletionTokens: 100,
			CreatedAt:        time.Time{},
			UpdatedAt:        time.Time{},
			Model:            "test-model",
		},
	}

	snap := core.NewUsageSnapshot("crush", "test")
	populateSnapshot(&snap, sessions, 1, now)

	// Verify yesterday bucket got the fallback session
	yesterdayKey := updatedTime.UTC().Format("2006-01-02")
	var foundYesterday bool
	for _, tp := range snap.DailySeries["sessions"] {
		if tp.Date == yesterdayKey {
			foundYesterday = true
			if tp.Value != 1 {
				t.Errorf("yesterday sessions = %v, want 1", tp.Value)
			}
		}
	}
	if !foundYesterday {
		t.Errorf("DailySeries missing yesterday bucket %s", yesterdayKey)
	}
}

func TestPopulateSnapshot_EmptyModelDefaultsToUnknown(t *testing.T) {
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	sessions := []crushSession{
		{
			ID:               "sess-nomodel",
			PromptTokens:     100,
			CompletionTokens: 50,
			Model:            "", // empty
		},
	}

	snap := core.NewUsageSnapshot("crush", "test")
	populateSnapshot(&snap, sessions, 1, now)

	if len(snap.ModelUsage) != 1 {
		t.Fatalf("ModelUsage len = %d, want 1", len(snap.ModelUsage))
	}
	if snap.ModelUsage[0].RawModelID != "unknown" {
		t.Errorf("RawModelID = %q, want unknown", snap.ModelUsage[0].RawModelID)
	}
}

func TestFormatHelpers_And_StatusMessage(t *testing.T) {
	// formatCount
	if got := formatCount(1, "project"); got != "1 project" {
		t.Errorf("formatCount(1, project) = %q", got)
	}
	if got := formatCount(3, "project"); got != "3 projects" {
		t.Errorf("formatCount(3, project) = %q", got)
	}
	if got := formatCount(1, "session"); got != "1 session" {
		t.Errorf("formatCount(1, session) = %q", got)
	}
	if got := formatCount(10, "session"); got != "10 sessions" {
		t.Errorf("formatCount(10, session) = %q", got)
	}

	// formatCostUSD
	if got := formatCostUSD(1.50); got != "$1.50" {
		t.Errorf("formatCostUSD(1.50) = %q, want $1.50", got)
	}
	if got := formatCostUSD(0.0042); got != "$0.0042" {
		t.Errorf("formatCostUSD(0.0042) = %q, want $0.0042", got)
	}
	if got := formatCostUSD(0.0); got != "$0.0000" {
		t.Errorf("formatCostUSD(0.0) = %q, want $0.0000", got)
	}

	// setUsedMetric edge cases
	snap := core.NewUsageSnapshot("crush", "test")
	setUsedMetric(&snap, "zero_val", 0, "units", "all-time")
	setUsedMetric(&snap, "neg_val", -5, "units", "all-time")
	if len(snap.Metrics) != 0 {
		t.Errorf("expected 0 metrics for <=0 values, got %d", len(snap.Metrics))
	}

	// setUsedMetric with nil Metrics map initializes it
	nilMetricsSnap := core.UsageSnapshot{}
	setUsedMetric(&nilMetricsSnap, "val", 100, "tokens", "all-time")
	if nilMetricsSnap.Metrics["val"].Used == nil || *nilMetricsSnap.Metrics["val"].Used != 100 {
		t.Errorf("setUsedMetric with nil map failed to set value")
	}

	// buildStatusMessage with empty snapshot
	if got := buildStatusMessage(snap); got != "OK" {
		t.Errorf("buildStatusMessage(empty) = %q, want OK", got)
	}

	// buildStatusMessage with projects and cost
	setUsedMetric(&snap, "total_projects", 2, "projects", "all-time")
	setUsedMetric(&snap, "total_cost_usd", 4.50, "USD", "all-time")
	msg := buildStatusMessage(snap)
	if msg != "2 projects, $4.50" {
		t.Errorf("buildStatusMessage = %q, want '2 projects, $4.50'", msg)
	}
}

func TestResolveRegistryPath_Priority(t *testing.T) {
	// Account hint wins
	acct := core.AccountConfig{ID: "test"}
	acct.SetPath(PathHintRegistryKey, "/override/registry.json")
	if got := resolveRegistryPath(acct); got != "/override/registry.json" {
		t.Errorf("resolveRegistryPath(hint) = %q", got)
	}

	// Env var wins over default
	acctNoHint := core.AccountConfig{ID: "test"}
	t.Setenv(EnvRegistry, "/env/registry.json")
	if got := resolveRegistryPath(acctNoHint); got != "/env/registry.json" {
		t.Errorf("resolveRegistryPath(env) = %q", got)
	}
}

func TestDefaultRegistryPath_MissingHome(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("HOME", "")
	if runtime.GOOS != "windows" {
		got := defaultRegistryPath()
		if got != "" {
			t.Errorf("defaultRegistryPath with empty HOME should return '', got %q", got)
		}
	}
}

func TestResolveProjectDB_EdgeCases(t *testing.T) {
	// Empty project path with relative data dir -> ""
	if got := resolveProjectDB(crushProject{Path: "", DataDir: ".crush"}); got != "" {
		t.Errorf("resolveProjectDB(empty path) = %q, want empty", got)
	}

	// Whitespace trimming
	got := resolveProjectDB(crushProject{Path: "  /proj  ", DataDir: "  .data  "})
	want := filepath.Join("/proj", ".data", "crush.db")
	if got != want {
		t.Errorf("resolveProjectDB(whitespace) = %q, want %q", got, want)
	}
}

func TestSplitPathList_EdgeCases(t *testing.T) {
	sep := string(os.PathListSeparator)

	// Empty string
	if got := splitPathList(""); len(got) != 0 {
		t.Errorf("splitPathList('') = %v, want empty", got)
	}

	// Only separators and whitespace
	input := fmt.Sprintf(" %s %s   %s ", sep, sep, sep)
	if got := splitPathList(input); len(got) != 0 {
		t.Errorf("splitPathList(separators) = %v, want empty", got)
	}

	// Duplicates and trimming
	input = fmt.Sprintf(" /path/a %s /path/b %s /path/a %s /path/c ", sep, sep, sep)
	got := splitPathList(input)
	if len(got) != 3 || got[0] != "/path/a" || got[1] != "/path/b" || got[2] != "/path/c" {
		t.Errorf("splitPathList(duplicates) = %v", got)
	}
}

func TestFileExists_EdgeCases(t *testing.T) {
	tmp := t.TempDir()
	filePath := filepath.Join(tmp, "file.txt")
	if err := os.WriteFile(filePath, []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}

	if fileExists("") {
		t.Errorf("fileExists('') should be false")
	}
	if fileExists(filepath.Join(tmp, "nonexistent")) {
		t.Errorf("fileExists(nonexistent) should be false")
	}
	if fileExists(tmp) {
		t.Errorf("fileExists(directory) should be false")
	}
	if !fileExists(filePath) {
		t.Errorf("fileExists(regular file) should be true")
	}
}

func TestProvider_Now_NilReceiverAndNilClock(t *testing.T) {
	var nilP *Provider
	if nilP.now().IsZero() {
		t.Errorf("nilP.now() returned zero time")
	}

	pNoClock := &Provider{clock: nil}
	if pNoClock.now().IsZero() {
		t.Errorf("pNoClock.now() returned zero time")
	}
}

func TestMillisToTime_ZeroAndNegative(t *testing.T) {
	if got := millisToTime(sql.NullInt64{Valid: false}); !got.IsZero() {
		t.Errorf("millisToTime(invalid) = %v, want zero", got)
	}
	if got := millisToTime(sql.NullInt64{Valid: true, Int64: 0}); !got.IsZero() {
		t.Errorf("millisToTime(0) = %v, want zero", got)
	}
	if got := millisToTime(sql.NullInt64{Valid: true, Int64: -100}); !got.IsZero() {
		t.Errorf("millisToTime(-100) = %v, want zero", got)
	}
}

func TestNonNegativeInt64_NegativeAndNull(t *testing.T) {
	if got := nonNegativeInt64(sql.NullInt64{Valid: false}); got != 0 {
		t.Errorf("nonNegativeInt64(invalid) = %d, want 0", got)
	}
	if got := nonNegativeInt64(sql.NullInt64{Valid: true, Int64: -50}); got != 0 {
		t.Errorf("nonNegativeInt64(-50) = %d, want 0", got)
	}
	if got := nonNegativeInt64(sql.NullInt64{Valid: true, Int64: 42}); got != 42 {
		t.Errorf("nonNegativeInt64(42) = %d, want 42", got)
	}
}

// -----------------------------------------------------------------------------
// Axis 3: Negative / Error Branches
// -----------------------------------------------------------------------------

func TestProvider_Fetch_NoDBsFound_StatusUnknown(t *testing.T) {
	provider := New()
	acct := core.AccountConfig{ID: "empty-acct"}
	acct.SetPath(PathHintSingleDBKey, "/nonexistent/db.db")

	snap, err := provider.Fetch(context.Background(), acct)
	if err != nil {
		t.Fatalf("Fetch() should not return err on missing DBs: %v", err)
	}

	if snap.Status != core.StatusUnknown {
		t.Errorf("Status = %v, want StatusUnknown", snap.Status)
	}
	if snap.Message != "No Crush project databases found" {
		t.Errorf("Message = %q, want 'No Crush project databases found'", snap.Message)
	}
}

func TestProvider_Fetch_AllDBsCorrupt_StatusError(t *testing.T) {
	tmp := t.TempDir()
	badDB1 := filepath.Join(tmp, "bad1.db")
	badDB2 := filepath.Join(tmp, "bad2.db")
	seedDB(t, badDB1)
	seedDB(t, badDB2)

	provider := New()
	acct := core.AccountConfig{ID: "all-corrupt"}
	acct.SetPath(PathHintDBsKey, badDB1+string(os.PathListSeparator)+badDB2)

	snap, err := provider.Fetch(context.Background(), acct)
	if err == nil {
		t.Fatal("Fetch() should return error when all DBs fail")
	}

	if snap.Status != core.StatusError {
		t.Errorf("Status = %v, want StatusError", snap.Status)
	}
	if snap.Message != "Failed to read any Crush database" {
		t.Errorf("Message = %q, want 'Failed to read any Crush database'", snap.Message)
	}
	if snap.Diagnostics["query_errors"] == "" {
		t.Errorf("Expected query_errors diagnostic to be populated")
	}
}

func TestProvider_Fetch_PartialDBFailure_StatusOKWithDiagnostics(t *testing.T) {
	tmp := t.TempDir()
	now := time.Now().UnixMilli()

	badDB := filepath.Join(tmp, "bad.db")
	seedDB(t, badDB)

	goodDB := filepath.Join(tmp, "good.db")
	createTestCrushDB(t, goodDB, testDBOptions{
		hasProviderColumn: true,
		sessions: []testSessionRow{
			{
				id:           "sess-good",
				messageCount: ptrInt64(3),
				promptTokens: ptrInt64(100),
				cost:         ptrFloat64(0.01),
				createdAt:    ptrInt64(now),
			},
		},
	})

	provider := New()
	acct := core.AccountConfig{ID: "partial-fail"}
	acct.SetPath(PathHintDBsKey, badDB+string(os.PathListSeparator)+goodDB)

	snap, err := provider.Fetch(context.Background(), acct)
	if err != nil {
		t.Fatalf("Fetch() unexpected error: %v", err)
	}

	if snap.Status != core.StatusOK {
		t.Errorf("Status = %v, want StatusOK", snap.Status)
	}
	if snap.Diagnostics["query_errors"] == "" {
		t.Errorf("Expected query_errors diagnostic for bad DB")
	}
	if got := *snap.Metrics["total_sessions"].Used; got != 1 {
		t.Errorf("total_sessions = %v, want 1", got)
	}
}

func TestOpenReadOnly_EmptyPath(t *testing.T) {
	_, err := openReadOnly("")
	if err == nil {
		t.Fatal("openReadOnly('') expected error")
	}
}

func TestPingContext_Canceled(t *testing.T) {
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "test.db")
	createTestCrushDB(t, dbPath, testDBOptions{})

	db, err := openReadOnly(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := pingContext(ctx, db); err == nil {
		t.Fatal("pingContext with canceled ctx expected error")
	}
}

func TestQuerySessions_ContextCanceled(t *testing.T) {
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "test.db")
	createTestCrushDB(t, dbPath, testDBOptions{})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := querySessions(ctx, dbPath)
	if err == nil {
		t.Fatal("querySessions with canceled ctx expected error")
	}
}

func TestQuerySessions_MissingSessionsTable(t *testing.T) {
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "missing_sessions.db")
	createTestCrushDB(t, dbPath, testDBOptions{
		missingSessions: true,
		missingMessages: false,
	})

	_, err := querySessions(context.Background(), dbPath)
	if err == nil {
		t.Fatal("querySessions on DB without sessions table expected error")
	}
}

func TestQuerySessions_MissingMessagesTable(t *testing.T) {
	tmp := t.TempDir()
	now := time.Now().UnixMilli()
	dbPath := filepath.Join(tmp, "missing_messages.db")
	createTestCrushDB(t, dbPath, testDBOptions{
		missingSessions: false,
		missingMessages: true,
		sessions: []testSessionRow{
			{
				id:           "sess-1",
				messageCount: ptrInt64(2),
				createdAt:    ptrInt64(now),
			},
		},
	})

	// When messages table is missing, latestAssistantModel fails
	_, err := querySessions(context.Background(), dbPath)
	if err == nil {
		t.Fatal("querySessions on DB without messages table expected error")
	}
}

func TestLatestAssistantModel_NoAssistantMessage(t *testing.T) {
	tmp := t.TempDir()
	now := time.Now().UnixMilli()
	dbPath := filepath.Join(tmp, "no_assistant.db")
	createTestCrushDB(t, dbPath, testDBOptions{
		hasProviderColumn: true,
		sessions: []testSessionRow{
			{
				id:           "sess-user-only",
				messageCount: ptrInt64(1),
				createdAt:    ptrInt64(now),
			},
		},
		messages: []testMessageRow{
			{
				id:        "msg-u",
				sessionID: "sess-user-only",
				role:      "user",
				createdAt: now,
			},
		},
	})

	sessions, err := querySessions(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("querySessions error: %v", err)
	}

	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
	if sessions[0].Model != "" || sessions[0].Provider != "" {
		t.Errorf("expected empty model and provider, got model=%q provider=%q", sessions[0].Model, sessions[0].Provider)
	}
}

// -----------------------------------------------------------------------------
// Axis 4: Concurrency & Race Conditions
// -----------------------------------------------------------------------------

func TestProvider_Fetch_Concurrent_Race(t *testing.T) {
	tmp := t.TempDir()
	now := time.Now().UnixMilli()
	dbPath := filepath.Join(tmp, "concurrent.db")
	createTestCrushDB(t, dbPath, testDBOptions{
		hasProviderColumn: true,
		sessions: []testSessionRow{
			{
				id:               "sess-c1",
				messageCount:     ptrInt64(5),
				promptTokens:     ptrInt64(100),
				completionTokens: ptrInt64(50),
				cost:             ptrFloat64(0.01),
				createdAt:        ptrInt64(now),
			},
		},
		messages: []testMessageRow{
			{
				id:        "msg-c1",
				sessionID: "sess-c1",
				role:      "assistant",
				model:     "claude-3-5-sonnet-20241022",
				provider:  "anthropic",
				createdAt: now,
			},
		},
	})

	provider := New()
	acct := core.AccountConfig{ID: "crush-race"}
	acct.SetPath(PathHintSingleDBKey, dbPath)

	const workers = 20
	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(workerID int) {
			defer wg.Done()
			snap, err := provider.Fetch(context.Background(), acct)
			if err != nil {
				t.Errorf("worker %d: Fetch() error = %v", workerID, err)
				return
			}
			if snap.Status != core.StatusOK {
				t.Errorf("worker %d: Status = %v", workerID, snap.Status)
			}
		}(i)
	}

	wg.Wait()
}

func TestProvider_ItemizedUsage_Concurrent_Race(t *testing.T) {
	tmp := t.TempDir()
	now := time.Now().UnixMilli()
	dbPath := filepath.Join(tmp, "proj", ".crush", "crush.db")
	createTestCrushDB(t, dbPath, testDBOptions{
		hasProviderColumn: true,
		sessions: []testSessionRow{
			{
				id:           "sess-race-item",
				messageCount: ptrInt64(2),
				promptTokens: ptrInt64(150),
				createdAt:    ptrInt64(now),
			},
		},
	})

	registry := writeRegistry(t, filepath.Join(tmp, "registry"), []crushProject{
		{Path: filepath.Join(tmp, "proj"), DataDir: ".crush"},
	})
	t.Setenv(EnvRegistry, registry)

	provider := New()

	const workers = 15
	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			events, err := provider.ItemizedUsage()
			if err != nil {
				t.Errorf("ItemizedUsage() error: %v", err)
				return
			}
			if len(events) != 1 {
				t.Errorf("expected 1 event, got %d", len(events))
			}
		}()
	}

	wg.Wait()
}

func TestProvider_HasChanged_Concurrent_Race(t *testing.T) {
	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "project", ".crush", "crush.db")
	seedDB(t, dbPath)

	provider := New()
	acct := core.AccountConfig{ID: "race-changed"}
	acct.SetPath(PathHintSingleDBKey, dbPath)
	past := time.Now().Add(-1 * time.Minute)

	const workers = 15
	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			changed, err := provider.HasChanged(acct, past)
			if err != nil || !changed {
				t.Errorf("HasChanged() = (%v, %v), want (true, nil)", changed, err)
			}
		}()
	}

	wg.Wait()
}

func TestQuerySessions_ConcurrentReads_Race(t *testing.T) {
	tmp := t.TempDir()
	now := time.Now().UnixMilli()
	dbPath := filepath.Join(tmp, "read_race.db")
	createTestCrushDB(t, dbPath, testDBOptions{
		hasProviderColumn: true,
		sessions: []testSessionRow{
			{
				id:           "sess-r1",
				messageCount: ptrInt64(3),
				promptTokens: ptrInt64(200),
				createdAt:    ptrInt64(now),
			},
		},
		messages: []testMessageRow{
			{
				id:        "msg-r1",
				sessionID: "sess-r1",
				role:      "assistant",
				model:     "claude-3-5-sonnet",
				provider:  "anthropic",
				createdAt: now,
			},
		},
	})

	const readers = 20
	var wg sync.WaitGroup
	wg.Add(readers)

	for i := 0; i < readers; i++ {
		go func() {
			defer wg.Done()
			sessions, err := querySessions(context.Background(), dbPath)
			if err != nil {
				t.Errorf("querySessions concurrent error: %v", err)
				return
			}
			if len(sessions) != 1 {
				t.Errorf("expected 1 session, got %d", len(sessions))
			}
		}()
	}

	wg.Wait()
}

// -----------------------------------------------------------------------------
// Axis 5: Domain Invariants
// -----------------------------------------------------------------------------

func TestDomainInvariants_MetricsSumAndWindows(t *testing.T) {
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	nowMS := now.UnixMilli()
	twoDaysAgoMS := now.AddDate(0, 0, -2).UnixMilli()

	sessions := []crushSession{
		{
			ID:               "s1",
			PromptTokens:     1000,
			CompletionTokens: 300,
			Cost:             0.05,
			HasCost:          true,
			CreatedAt:        time.UnixMilli(nowMS).UTC(),
			Model:            "model-a",
			Provider:         "provider-a",
		},
		{
			ID:               "s2",
			PromptTokens:     500,
			CompletionTokens: 200,
			Cost:             0.02,
			HasCost:          true,
			CreatedAt:        time.UnixMilli(twoDaysAgoMS).UTC(),
			Model:            "model-b",
			Provider:         "provider-b",
		},
	}

	snap := core.NewUsageSnapshot("crush", "test")
	populateSnapshot(&snap, sessions, 2, now)

	// Invariant 1: total_tokens = total_input_tokens + total_output_tokens
	totTokens := *snap.Metrics["total_tokens"].Used
	totIn := *snap.Metrics["total_input_tokens"].Used
	totOut := *snap.Metrics["total_output_tokens"].Used
	if totTokens != totIn+totOut {
		t.Errorf("Invariant violated: total_tokens (%v) != in (%v) + out (%v)", totTokens, totIn, totOut)
	}

	// Invariant 2: Windows correctly assigned
	if snap.Metrics["total_tokens"].Window != allTimeWindow {
		t.Errorf("total_tokens window = %q, want %q", snap.Metrics["total_tokens"].Window, allTimeWindow)
	}
	if snap.Metrics["sessions_today"].Window != "today" {
		t.Errorf("sessions_today window = %q, want 'today'", snap.Metrics["sessions_today"].Window)
	}
	if snap.Metrics["sessions_7d"].Window != "7d" {
		t.Errorf("sessions_7d window = %q, want '7d'", snap.Metrics["sessions_7d"].Window)
	}

	// Invariant 3: ModelUsage sums match top-level totals
	var sumModelIn, sumModelOut, sumModelReqs float64
	for _, rec := range snap.ModelUsage {
		sumModelIn += *rec.InputTokens
		sumModelOut += *rec.OutputTokens
		sumModelReqs += *rec.Requests
	}
	if sumModelIn != totIn {
		t.Errorf("ModelUsage in tokens sum (%v) != top level (%v)", sumModelIn, totIn)
	}
	if sumModelOut != totOut {
		t.Errorf("ModelUsage out tokens sum (%v) != top level (%v)", sumModelOut, totOut)
	}
	if sumModelReqs != *snap.Metrics["total_sessions"].Used {
		t.Errorf("ModelUsage requests sum (%v) != total_sessions (%v)", sumModelReqs, *snap.Metrics["total_sessions"].Used)
	}
}
