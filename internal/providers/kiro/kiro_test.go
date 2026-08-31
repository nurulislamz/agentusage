package kiro

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/nurulislamz/agentusage/internal/core"
)

// ============================================================================
// Axis 1: Happy Paths
// ============================================================================

func TestKiro_HappyPath_FetchWithBothSources(t *testing.T) {
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, "sessions")
	if err := os.MkdirAll(sessionsDir, 0o755); err != nil {
		t.Fatalf("mkdir sessions: %v", err)
	}

	// Create JSON session with companion JSONL
	sessJSON := filepath.Join(sessionsDir, "sess-1.json")
	sessData := `{
		"session_id": "sess-1",
		"cwd": "/workspace/project-a",
		"updated_at": "2026-05-18T10:00:00Z",
		"session_state": {
			"rts_model_state": {
				"model_info": {
					"model_id": "claude-3-7-sonnet",
					"context_window_tokens": 200000
				}
			},
			"conversation_metadata": {
				"user_turn_metadatas": [
					{
						"input_tokens": 1500,
						"output_tokens": 300,
						"request_start_time": "2026-05-18T09:59:00Z",
						"request_end_time": "2026-05-18T10:00:00Z"
					}
				]
			}
		}
	}`
	if err := os.WriteFile(sessJSON, []byte(sessData), 0o600); err != nil {
		t.Fatalf("write sessJSON: %v", err)
	}

	sessJSONL := filepath.Join(sessionsDir, "sess-1.jsonl")
	jsonlContent := `{"kind":"UserMessage","data":{"id":"u1"}}
{"kind":"AssistantMessage","data":{"message_id":"msg-1","timestamp":"2026-05-18T10:00:00Z","content":[{"text":"hello world"}],"metadata":{"response_size":40}}}`
	if err := os.WriteFile(sessJSONL, []byte(jsonlContent), 0o600); err != nil {
		t.Fatalf("write sessJSONL: %v", err)
	}

	// Create SQLite DB with conversations_v2
	dbPath := filepath.Join(tmpDir, "data.sqlite3")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	_, err = db.Exec(`CREATE TABLE conversations_v2 (key TEXT, conversation_id TEXT, value TEXT)`)
	if err != nil {
		t.Fatalf("create table: %v", err)
	}
	dbValue := `{
		"session_id": "sess-2",
		"cwd": "/workspace/project-b",
		"updated_at": "2026-05-18T11:00:00Z",
		"session_state": {
			"rts_model_state": {
				"model_info": {
					"model_id": "amazon.nova-pro",
					"context_window_tokens": 100000
				}
			},
			"conversation_metadata": {
				"user_turn_metadatas": [
					{"input_tokens": 2000, "output_tokens": 500}
				]
			}
		},
		"history": [{}, {}, {}]
	}`
	_, err = db.Exec(`INSERT INTO conversations_v2 (key, conversation_id, value) VALUES (?, ?, ?)`, "key-2", "sess-2", dbValue)
	if err != nil {
		t.Fatalf("insert dbValue: %v", err)
	}
	db.Close()

	p := New()
	p.clock = fixedClock{t: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)}
	acct := core.AccountConfig{ID: DefaultAccountID, Provider: ID, Auth: "local"}
	acct.SetPath(PathHintSessionsDirKey, sessionsDir)
	acct.SetPath(PathHintDBKey, dbPath)

	snap, err := p.Fetch(context.Background(), acct)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if snap.Status != core.StatusOK {
		t.Fatalf("status = %v, want OK (message=%s)", snap.Status, snap.Message)
	}

	if val := snap.Metrics["total_conversations"].Used; val == nil || *val != 2 {
		t.Errorf("total_conversations = %v, want 2", val)
	}
	if val := snap.Metrics["conversations_with_tokens"].Used; val == nil || *val != 2 {
		t.Errorf("conversations_with_tokens = %v, want 2", val)
	}
	if val := snap.Metrics["total_input_tokens"].Used; val == nil || *val != 3500 {
		t.Errorf("total_input_tokens = %v, want 3500", val)
	}
	if val := snap.Metrics["total_output_tokens"].Used; val == nil || *val != 800 {
		t.Errorf("total_output_tokens = %v, want 800", val)
	}
	if val := snap.Metrics["total_tokens"].Used; val == nil || *val != 4300 {
		t.Errorf("total_tokens = %v, want 4300", val)
	}
	if val := snap.Metrics["total_messages"].Used; val == nil || *val != 4 {
		t.Errorf("total_messages = %v, want 4", val)
	}

	if len(snap.DailySeries["conversations"]) == 0 {
		t.Error("DailySeries[conversations] is empty")
	}
	if len(snap.DailySeries["tokens"]) == 0 {
		t.Error("DailySeries[tokens] is empty")
	}
	if len(snap.ModelUsage) != 2 {
		t.Fatalf("len(ModelUsage) = %d, want 2", len(snap.ModelUsage))
	}
}

func TestKiro_HappyPath_ItemizedUsage(t *testing.T) {
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, "sessions")
	if err := os.MkdirAll(sessionsDir, 0o755); err != nil {
		t.Fatalf("mkdir sessions: %v", err)
	}

	sessJSON := filepath.Join(sessionsDir, "sess-itemized.json")
	sessData := `{
		"session_id": "sess-itemized",
		"cwd": "/workspace/itemized",
		"updated_at": "2026-05-18T10:00:00Z",
		"session_state": {
			"rts_model_state": {
				"model_info": {
					"model_id": "claude-3-7-sonnet"
				}
			},
			"conversation_metadata": {
				"user_turn_metadatas": [
					{"input_tokens": 100, "output_tokens": 20}
				]
			}
		}
	}`
	if err := os.WriteFile(sessJSON, []byte(sessData), 0o600); err != nil {
		t.Fatalf("write sessJSON: %v", err)
	}

	t.Setenv("KIRO_SESSIONS_DIR", sessionsDir)
	t.Setenv("KIRO_DATA_DIR", tmpDir)

	// Create data.sqlite3
	dbPath := filepath.Join(tmpDir, "data.sqlite3")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	_, _ = db.Exec(`CREATE TABLE conversations_v2 (key TEXT, conversation_id TEXT, value TEXT)`)
	dbValue := `{
		"session_id": "sess-db-itemized",
		"cwd": "/workspace/db",
		"updated_at": "2026-05-18T11:00:00Z",
		"session_state": {
			"rts_model_state": {
				"model_info": {"model_id": "nova-pro"}
			},
			"conversation_metadata": {
				"user_turn_metadatas": [{"input_tokens": 50, "output_tokens": 10}]
			}
		}
	}`
	_, _ = db.Exec(`INSERT INTO conversations_v2 (key, conversation_id, value) VALUES (?, ?, ?)`, "key-db", "sess-db-itemized", dbValue)
	db.Close()

	p := New()
	events, err := p.ItemizedUsage()
	if err != nil {
		t.Fatalf("ItemizedUsage: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("len(events) = %d, want 2", len(events))
	}
	if events[0].ProviderID != ID {
		t.Errorf("ProviderID = %q, want %q", events[0].ProviderID, ID)
	}
}

func TestKiro_HappyPath_LegacyConversationsTable(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "data.sqlite3")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	_, err = db.Exec(`CREATE TABLE conversations (key TEXT, value TEXT)`)
	if err != nil {
		t.Fatalf("create table: %v", err)
	}
	dbValue := `{
		"session_id": "legacy-session-123",
		"cwd": "/workspace/legacy",
		"updated_at": "2026-05-18T14:00:00Z",
		"session_state": {
			"rts_model_state": {
				"model_info": {"model_id": "legacy-model", "context_window_tokens": 8000}
			},
			"conversation_metadata": {
				"user_turn_metadatas": [{"input_tokens": 400, "output_tokens": 80}]
			}
		}
	}`
	_, err = db.Exec(`INSERT INTO conversations (key, value) VALUES (?, ?)`, "legacy-key", dbValue)
	if err != nil {
		t.Fatalf("insert row: %v", err)
	}
	db.Close()

	convs, err := queryKiroConversations(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("queryKiroConversations: %v", err)
	}
	if len(convs) != 1 {
		t.Fatalf("len(convs) = %d, want 1", len(convs))
	}
	c := convs[0]
	if c.ConversationID != "legacy-session-123" {
		t.Errorf("ConversationID = %q, want legacy-session-123", c.ConversationID)
	}
	if c.Model != "legacy-model" {
		t.Errorf("Model = %q, want legacy-model", c.Model)
	}
	if c.InputTokens != 400 || c.OutputTokens != 80 {
		t.Errorf("tokens = %d/%d, want 400/80", c.InputTokens, c.OutputTokens)
	}
}

func TestKiro_HappyPath_SessionHeaderOnlyTurnMetadata(t *testing.T) {
	tmpDir := t.TempDir()
	headerPath := filepath.Join(tmpDir, "turn-only.json")
	headerData := `{
		"session_id": "turn-only-sess",
		"cwd": "/workspace/turns",
		"updated_at": "2026-05-18T08:00:00Z",
		"session_state": {
			"rts_model_state": {
				"model_info": {
					"model_id": "claude-sonnet",
					"context_window_tokens": 10000
				}
			},
			"conversation_metadata": {
				"user_turn_metadatas": [
					{
						"input_tokens": 500,
						"output_tokens": 150,
						"request_start_time": "2026-05-18T08:01:00Z",
						"request_end_time": "2026-05-18T08:02:00Z"
					},
					{
						"context_usage_percentage": 0.1,
						"output_tokens": 100,
						"request_end_time": "2026-05-18T08:05:00Z"
					}
				]
			}
		}
	}`
	if err := os.WriteFile(headerPath, []byte(headerData), 0o600); err != nil {
		t.Fatalf("write header: %v", err)
	}

	conv, err := parseKiroSession(headerPath)
	if err != nil {
		t.Fatalf("parseKiroSession: %v", err)
	}
	if conv == nil {
		t.Fatal("conv is nil")
	}
	if conv.ConversationID != "turn-only-sess" {
		t.Errorf("ConversationID = %q, want turn-only-sess", conv.ConversationID)
	}
	if conv.MessageCount != 2 {
		t.Errorf("MessageCount = %d, want 2", conv.MessageCount)
	}
	if conv.InputTokens != 1500 {
		t.Errorf("InputTokens = %d, want 1500", conv.InputTokens)
	}
	if conv.OutputTokens != 250 {
		t.Errorf("OutputTokens = %d, want 250", conv.OutputTokens)
	}
	if conv.TotalTokens != 1750 {
		t.Errorf("TotalTokens = %d, want 1750", conv.TotalTokens)
	}
}

func TestKiro_HappyPath_HistoryTokensExtraction(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "data.sqlite3")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	_, err = db.Exec(`CREATE TABLE conversations_v2 (key TEXT, conversation_id TEXT, value TEXT)`)
	if err != nil {
		t.Fatalf("create table: %v", err)
	}

	dbValue := `{
		"session_id": "hist-sess",
		"cwd": "/workspace/hist",
		"updated_at": "2026-05-18T10:00:00Z",
		"session_state": {
			"rts_model_state": {
				"model_info": {
					"model_id": "claude-opus",
					"context_window_tokens": 100000
				}
			}
		},
		"history": [
			{
				"role": "user",
				"content": "do something"
			},
			{
				"role": "assistant",
				"input_tokens": 1200,
				"output_tokens": 450
			},
			{
				"role": "assistant",
				"context_usage_percentage": 0.05,
				"response_size": 40
			}
		]
	}`
	_, err = db.Exec(`INSERT INTO conversations_v2 (key, conversation_id, value) VALUES (?, ?, ?)`, "key-hist", "hist-sess", dbValue)
	if err != nil {
		t.Fatalf("insert row: %v", err)
	}
	db.Close()

	convs, err := queryKiroConversations(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("queryKiroConversations: %v", err)
	}
	if len(convs) != 1 {
		t.Fatalf("len(convs) = %d, want 1", len(convs))
	}
	c := convs[0]
	if c.InputTokens != 6200 {
		t.Errorf("InputTokens = %d, want 6200", c.InputTokens)
	}
	if c.OutputTokens != 460 {
		t.Errorf("OutputTokens = %d, want 460", c.OutputTokens)
	}
	if !c.HasTokens {
		t.Error("HasTokens = false, want true")
	}
}

func TestKiro_HappyPath_HasChanged(t *testing.T) {
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, "sessions")
	if err := os.MkdirAll(sessionsDir, 0o755); err != nil {
		t.Fatalf("mkdir sessions: %v", err)
	}
	dbPath := filepath.Join(tmpDir, "data.sqlite3")
	if err := os.WriteFile(dbPath, []byte("data"), 0o600); err != nil {
		t.Fatalf("write db: %v", err)
	}

	p := New()
	acct := core.AccountConfig{ID: DefaultAccountID, Provider: ID}
	acct.SetPath(PathHintSessionsDirKey, sessionsDir)
	acct.SetPath(PathHintDBKey, dbPath)

	now := time.Now()
	past := now.Add(-1 * time.Hour)
	future := now.Add(1 * time.Hour)

	changed, err := p.HasChanged(acct, past)
	if err != nil {
		t.Fatalf("HasChanged(past): %v", err)
	}
	if !changed {
		t.Errorf("HasChanged(past) = false, want true")
	}

	changedFuture, err := p.HasChanged(acct, future)
	if err != nil {
		t.Fatalf("HasChanged(future): %v", err)
	}
	if changedFuture {
		t.Errorf("HasChanged(future) = true, want false")
	}

	emptyAcct := core.AccountConfig{}
	t.Setenv("KIRO_DATA_DIR", "")
	t.Setenv("KIRO_SESSIONS_DIR", "")
	changedEmpty, err := p.HasChanged(emptyAcct, past)
	if err != nil {
		t.Fatalf("HasChanged(empty): %v", err)
	}
	if changedEmpty {
		t.Errorf("HasChanged(empty) = true, want false")
	}
}

// ============================================================================
// Axis 2: Edge & Boundary Cases
// ============================================================================

func TestKiro_Edge_EmptyAndMalformedJSON(t *testing.T) {
	t.Run("parseKiroValue empty string", func(t *testing.T) {
		conv, ok := parseKiroValue("k1", "id1", "   ")
		if ok || conv.Key != "" {
			t.Errorf("parseKiroValue with whitespace = (%v, %v), want zero/false", conv, ok)
		}
	})

	t.Run("parseKiroValue malformed JSON returns fallback session", func(t *testing.T) {
		conv, ok := parseKiroValue("k1", "id1", "{not-json")
		if !ok {
			t.Fatalf("parseKiroValue with invalid JSON returned ok=false, want true fallback")
		}
		if conv.Key != "k1" || conv.ConversationID != "id1" || conv.Source != "sqlite" {
			t.Errorf("conv = %+v, unexpected fallback fields", conv)
		}
	})

	t.Run("parseKiroValue empty object", func(t *testing.T) {
		conv, ok := parseKiroValue("k1", "", "{}")
		if !ok {
			t.Fatalf("parseKiroValue with empty object returned ok=false")
		}
		if conv.Model != "" || conv.HasTokens {
			t.Errorf("unexpected tokens or model in empty object: %+v", conv)
		}
	})

	t.Run("parseKiroSessionHeader missing file", func(t *testing.T) {
		h, err := parseKiroSessionHeader("/path/does/not/exist/sess.json")
		if err != nil || h != nil {
			t.Errorf("parseKiroSessionHeader missing file = (%v, %v), want (nil, nil)", h, err)
		}
	})

	t.Run("parseKiroSessionHeader malformed JSON", func(t *testing.T) {
		tmp := filepath.Join(t.TempDir(), "bad.json")
		_ = os.WriteFile(tmp, []byte("bad-json"), 0o600)
		h, err := parseKiroSessionHeader(tmp)
		if err != nil || h != nil {
			t.Errorf("parseKiroSessionHeader bad JSON = (%v, %v), want (nil, nil)", h, err)
		}
	})

	t.Run("parseKiroSessionJSONL missing file", func(t *testing.T) {
		events, err := parseKiroSessionJSONL("/path/does/not/exist/sess.jsonl", nil)
		if err != nil || events != nil {
			t.Errorf("parseKiroSessionJSONL missing file = (%v, %v), want (nil, nil)", events, err)
		}
	})

	t.Run("parseKiroSessionJSONL with various message kinds and malformed lines", func(t *testing.T) {
		tmpDir := t.TempDir()
		jsonlPath := filepath.Join(tmpDir, "mixed.jsonl")
		content := `
{"kind":"UserMessage","data":{"id":"u1"}}
malformed line
{"kind":"AssistantMessage","data":{"id":"a1","timestamp":"2026-05-18T10:00:00Z","metadata":{"response_size":20}}}
{"kind":"ToolUse","data":{}}
{"kind":"AssistantMessage","data":"invalid-json-data"}
`
		_ = os.WriteFile(jsonlPath, []byte(content), 0o600)
		events, err := parseKiroSessionJSONL(jsonlPath, &kiroHeader{})
		if err != nil {
			t.Fatalf("parseKiroSessionJSONL: %v", err)
		}
		if len(events) != 1 {
			t.Fatalf("len(events) = %d, want 1 valid AssistantMessage", len(events))
		}
	})
}

func TestKiro_Edge_NumericAndTypeParsers(t *testing.T) {
	t.Run("readInt64 types", func(t *testing.T) {
		if v := readInt64(map[string]any{"a": float64(42)}, "a"); v != 42 {
			t.Errorf("readInt64(float64) = %d, want 42", v)
		}
		if v := readInt64(map[string]any{"a": float64(-5)}, "a"); v != 0 {
			t.Errorf("readInt64(negative float64) = %d, want 0", v)
		}
		if v := readInt64(map[string]any{"a": int64(100)}, "a"); v != 100 {
			t.Errorf("readInt64(int64) = %d, want 100", v)
		}
		if v := readInt64(map[string]any{"a": int64(-10)}, "a"); v != 0 {
			t.Errorf("readInt64(negative int64) = %d, want 0", v)
		}
		if v := readInt64(map[string]any{"a": int(20)}, "a"); v != 20 {
			t.Errorf("readInt64(int) = %d, want 20", v)
		}
		if v := readInt64(map[string]any{"a": int(-20)}, "a"); v != 0 {
			t.Errorf("readInt64(negative int) = %d, want 0", v)
		}
		if v := readInt64(map[string]any{"a": "string"}, "a"); v != 0 {
			t.Errorf("readInt64(string) = %d, want 0", v)
		}
		if v := readInt64(map[string]any{}, "missing"); v != 0 {
			t.Errorf("readInt64(missing) = %d, want 0", v)
		}
	})

	t.Run("readFloat64 types", func(t *testing.T) {
		if v := readFloat64(map[string]any{"f": float64(0.75)}, "f"); v != 0.75 {
			t.Errorf("readFloat64(float64) = %v, want 0.75", v)
		}
		if v := readFloat64(map[string]any{"f": float64(-1.5)}, "f"); v != 0 {
			t.Errorf("readFloat64(negative float64) = %v, want 0", v)
		}
		if v := readFloat64(map[string]any{"f": int64(50)}, "f"); v != 50.0 {
			t.Errorf("readFloat64(int64) = %v, want 50.0", v)
		}
		if v := readFloat64(map[string]any{"f": int64(-50)}, "f"); v != 0 {
			t.Errorf("readFloat64(negative int64) = %v, want 0", v)
		}
		if v := readFloat64(map[string]any{"f": int(10)}, "f"); v != 10.0 {
			t.Errorf("readFloat64(int) = %v, want 10.0", v)
		}
		if v := readFloat64(map[string]any{"f": int(-10)}, "f"); v != 0 {
			t.Errorf("readFloat64(negative int) = %v, want 0", v)
		}
		if v := readFloat64(map[string]any{"f": "not-a-number"}, "f"); v != 0 {
			t.Errorf("readFloat64(string) = %v, want 0", v)
		}
		if v := readFloat64(map[string]any{}, "missing"); v != 0 {
			t.Errorf("readFloat64(missing) = %v, want 0", v)
		}
	})

	t.Run("readOptionalInt64 types", func(t *testing.T) {
		v, ok := readOptionalInt64(map[string]any{"keyA": float64(10)}, "missing", "keyA")
		if !ok || v != 10 {
			t.Errorf("readOptionalInt64(float64) = (%d, %v), want (10, true)", v, ok)
		}
		v, ok = readOptionalInt64(map[string]any{"keyA": float64(-10)}, "keyA")
		if !ok || v != 0 {
			t.Errorf("readOptionalInt64(negative float64) = (%d, %v), want (0, true)", v, ok)
		}
		v, ok = readOptionalInt64(map[string]any{"keyB": int64(25)}, "keyB")
		if !ok || v != 25 {
			t.Errorf("readOptionalInt64(int64) = (%d, %v), want (25, true)", v, ok)
		}
		v, ok = readOptionalInt64(map[string]any{"keyB": int64(-25)}, "keyB")
		if !ok || v != 0 {
			t.Errorf("readOptionalInt64(negative int64) = (%d, %v), want (0, true)", v, ok)
		}
		v, ok = readOptionalInt64(map[string]any{"keyC": int(30)}, "keyC")
		if !ok || v != 30 {
			t.Errorf("readOptionalInt64(int) = (%d, %v), want (30, true)", v, ok)
		}
		v, ok = readOptionalInt64(map[string]any{"keyC": int(-30)}, "keyC")
		if !ok || v != 0 {
			t.Errorf("readOptionalInt64(negative int) = (%d, %v), want (0, true)", v, ok)
		}
		v, ok = readOptionalInt64(map[string]any{"keyD": "invalid"}, "keyD")
		if ok || v != 0 {
			t.Errorf("readOptionalInt64(invalid string) = (%d, %v), want (0, false)", v, ok)
		}
		v, ok = readOptionalInt64(map[string]any{}, "none")
		if ok || v != 0 {
			t.Errorf("readOptionalInt64(missing) = (%d, %v), want (0, false)", v, ok)
		}
	})

	t.Run("textLength and estimateTokensFromChars", func(t *testing.T) {
		if l := textLength("hello"); l != 5 {
			t.Errorf("textLength(string) = %d, want 5", l)
		}
		if l := textLength([]any{"foo", "bar", map[string]any{"text": "baz"}}); l != 9 {
			t.Errorf("textLength(slice) = %d, want 9", l)
		}
		nestedMap := map[string]any{
			"content": "abc",
			"extra":   []any{"d", "e"},
		}
		if l := textLength(nestedMap); l != 5 {
			t.Errorf("textLength(nestedMap) = %d, want 5", l)
		}
		if l := textLength(12345); l != 0 {
			t.Errorf("textLength(int) = %d, want 0", l)
		}

		if tok := estimateTokensFromChars(0); tok != 0 {
			t.Errorf("estimateTokensFromChars(0) = %d, want 0", tok)
		}
		if tok := estimateTokensFromChars(-10); tok != 0 {
			t.Errorf("estimateTokensFromChars(-10) = %d, want 0", tok)
		}
		if tok := estimateTokensFromChars(4); tok != 1 {
			t.Errorf("estimateTokensFromChars(4) = %d, want 1", tok)
		}
		if tok := estimateTokensFromChars(5); tok != 2 {
			t.Errorf("estimateTokensFromChars(5) = %d, want 2", tok)
		}
	})

	t.Run("parseTimestamp formats", func(t *testing.T) {
		formats := []string{
			"2026-05-18T12:34:56.789Z",
			"2026-05-18T12:34:56Z",
			"2026-05-18 12:34:56.789",
			"2026-05-18 12:34:56",
			"2026-05-18",
		}
		for _, f := range formats {
			ts, ok := parseTimestamp(f)
			if !ok || ts.IsZero() {
				t.Errorf("parseTimestamp(%q) failed: ok=%v, ts=%v", f, ok, ts)
			}
		}
		if _, ok := parseTimestamp(""); ok {
			t.Error("parseTimestamp(\"\") = true, want false")
		}
		if _, ok := parseTimestamp("not-a-date"); ok {
			t.Error("parseTimestamp(\"not-a-date\") = true, want false")
		}
	})

	t.Run("pickTime and pickNonEmpty", func(t *testing.T) {
		t1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		if got := pickTime(time.Time{}, t1); !got.Equal(t1) {
			t.Errorf("pickTime = %v, want %v", got, t1)
		}
		if got := pickTime(time.Time{}, time.Time{}); !got.IsZero() {
			t.Errorf("pickTime(zeros) = %v, want zero", got)
		}

		if got := pickNonEmpty("   ", "default"); got != "default" {
			t.Errorf("pickNonEmpty = %q, want default", got)
		}
		if got := pickNonEmpty("first", "second"); got != "first" {
			t.Errorf("pickNonEmpty = %q, want first", got)
		}
	})
}

func TestKiro_Edge_PathResolutionsAndEnv(t *testing.T) {
	tmpDir := t.TempDir()
	customDataDir := filepath.Join(tmpDir, "custom_data")
	_ = os.MkdirAll(customDataDir, 0o755)
	dbFile := filepath.Join(customDataDir, "data.sqlite3")
	_ = os.WriteFile(dbFile, []byte(""), 0o600)

	customSessionsDir := filepath.Join(tmpDir, "custom_sessions")
	_ = os.MkdirAll(customSessionsDir, 0o755)

	t.Setenv("KIRO_DATA_DIR", customDataDir)
	t.Setenv("KIRO_SESSIONS_DIR", customSessionsDir)
	t.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir, "xdg_share"))

	resolvedDB := resolveDBPath(core.AccountConfig{})
	if resolvedDB != dbFile {
		t.Errorf("resolveDBPath() = %q, want %q", resolvedDB, dbFile)
	}

	resolvedSess := resolveSessionsDir(core.AccountConfig{})
	if resolvedSess != customSessionsDir {
		t.Errorf("resolveSessionsDir() = %q, want %q", resolvedSess, customSessionsDir)
	}

	overrideDir := filepath.Join(tmpDir, "override_sess")
	_ = os.MkdirAll(overrideDir, 0o755)
	overrideDB := filepath.Join(tmpDir, "override.sqlite3")
	_ = os.WriteFile(overrideDB, []byte(""), 0o600)

	acct := core.AccountConfig{}
	acct.SetPath(PathHintDBKey, overrideDB)
	acct.SetPath(PathHintSessionsDirKey, overrideDir)

	if got := resolveDBPath(acct); got != overrideDB {
		t.Errorf("resolveDBPath with override = %q, want %q", got, overrideDB)
	}
	if got := resolveSessionsDir(acct); got != overrideDir {
		t.Errorf("resolveSessionsDir with override = %q, want %q", got, overrideDir)
	}

	if fileExists("") || dirExists("") {
		t.Error("fileExists/dirExists on empty string returned true")
	}
	if fileExists(overrideDir) {
		t.Error("fileExists on directory returned true")
	}
	if dirExists(overrideDB) {
		t.Error("dirExists on file returned true")
	}
}

func TestKiro_Edge_EstimateTokensFromSessionState(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "data.sqlite3")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	_, err = db.Exec(`CREATE TABLE conversations_v2 (key TEXT, conversation_id TEXT, value TEXT)`)
	if err != nil {
		t.Fatalf("create table: %v", err)
	}

	// Turn metadata with only context_usage_percentage (no explicit input/output tokens)
	dbValue := `{
		"session_id": "estimate-sess",
		"cwd": "/workspace/estimate",
		"updated_at": "2026-05-18T10:00:00Z",
		"session_state": {
			"rts_model_state": {
				"model_info": {
					"model_id": "claude-sonnet",
					"context_window_tokens": 50000
				}
			},
			"conversation_metadata": {
				"user_turn_metadatas": [
					{"context_usage_percentage": 0.1},
					{"context_usage_percentage": -0.5}
				]
			}
		}
	}`
	_, err = db.Exec(`INSERT INTO conversations_v2 (key, conversation_id, value) VALUES (?, ?, ?)`, "key-est", "estimate-sess", dbValue)
	if err != nil {
		t.Fatalf("insert row: %v", err)
	}
	db.Close()

	convs, err := queryKiroConversations(context.Background(), dbPath)
	if err != nil {
		t.Fatalf("queryKiroConversations: %v", err)
	}
	if len(convs) != 1 {
		t.Fatalf("len(convs) = %d, want 1", len(convs))
	}
	c := convs[0]
	// 0.1 * 50000 = 5000 tokens
	if c.InputTokens != 5000 {
		t.Errorf("InputTokens = %d, want 5000", c.InputTokens)
	}
	if !c.HasTokens {
		t.Error("HasTokens = false, want true")
	}

	// Directly test estimateTokensFromSessionState with edge cases
	var rawState map[string]json.RawMessage
	in, out, tot, ok := estimateTokensFromSessionState(rawState, 0)
	if ok || in != 0 || out != 0 || tot != 0 {
		t.Errorf("estimateTokensFromSessionState(zero contextWindow) = (%d,%d,%d,%v), want zeros/false", in, out, tot, ok)
	}

	// State without metadata
	rawState = map[string]json.RawMessage{"other": []byte("{}")}
	in, out, tot, ok = estimateTokensFromSessionState(rawState, 1000)
	if ok {
		t.Errorf("estimateTokensFromSessionState(no metadata) = true, want false")
	}

	// State with invalid metadata JSON
	rawState = map[string]json.RawMessage{"conversation_metadata": []byte("invalid-json")}
	in, out, tot, ok = estimateTokensFromSessionState(rawState, 1000)
	if ok {
		t.Errorf("estimateTokensFromSessionState(invalid json) = true, want false")
	}
}

func TestKiro_Edge_ExtractModelInfoEdges(t *testing.T) {
	// Missing rts_model_state
	if m := extractModelFromSessionState(map[string]json.RawMessage{}); m != "" {
		t.Errorf("extractModelFromSessionState(empty) = %q, want empty", m)
	}
	// Invalid rts_model_state
	if m := extractModelFromSessionState(map[string]json.RawMessage{"rts_model_state": []byte("bad")}); m != "" {
		t.Errorf("extractModelFromSessionState(bad rts) = %q, want empty", m)
	}
	// Missing model_info
	if m := extractModelFromSessionState(map[string]json.RawMessage{"rts_model_state": []byte("{}")}); m != "" {
		t.Errorf("extractModelFromSessionState(no model_info) = %q, want empty", m)
	}
	// Bad model_info
	if m := extractModelFromSessionState(map[string]json.RawMessage{"rts_model_state": []byte(`{"model_info":"not-a-map"}`)}); m != "" {
		t.Errorf("extractModelFromSessionState(bad model_info) = %q, want empty", m)
	}
	// Non-string model_id
	if m := extractModelFromSessionState(map[string]json.RawMessage{"rts_model_state": []byte(`{"model_info":{"model_id":123}}`)}); m != "" {
		t.Errorf("extractModelFromSessionState(non-string model_id) = %q, want empty", m)
	}

	// Context window with missing model_info
	if cw := extractContextWindowFromSessionState(map[string]json.RawMessage{}); cw != 0 {
		t.Errorf("extractContextWindowFromSessionState(empty) = %d, want 0", cw)
	}
}

func TestKiro_Edge_TokensFromTurnMetadataEdges(t *testing.T) {
	if in, out, tot, ok := tokensFromTurnMetadata(nil); ok || in != 0 || out != 0 || tot != 0 {
		t.Errorf("tokensFromTurnMetadata(nil) = (%d,%d,%d,%v), want zeros/false", in, out, tot, ok)
	}

	header := &kiroHeader{
		TurnMetadatas: []kiroTurnMetadata{
			{HasInputTokens: false, HasOutputTokens: false},
		},
	}
	if in, out, tot, ok := tokensFromTurnMetadata(header); ok {
		t.Errorf("tokensFromTurnMetadata(no tokens) = true, want false (got %d,%d,%d)", in, out, tot)
	}
}

func TestKiro_Edge_Fetch_OneSourceFailsOtherSucceeds(t *testing.T) {
	tmpDir := t.TempDir()
	// Create valid session file
	sessDir := filepath.Join(tmpDir, "sess")
	_ = os.MkdirAll(sessDir, 0o755)
	sessPath := filepath.Join(sessDir, "valid.json")
	_ = os.WriteFile(sessPath, []byte(`{"session_id":"sess-ok","session_state":{"rts_model_state":{"model_info":{"model_id":"claude"}}}}`), 0o600)

	// Create corrupt sqlite file
	badDB := filepath.Join(tmpDir, "bad.sqlite3")
	_ = os.WriteFile(badDB, []byte("corrupt-sqlite"), 0o600)

	p := New()
	acct := core.AccountConfig{ID: DefaultAccountID, Provider: ID}
	acct.SetPath(PathHintSessionsDirKey, sessDir)
	acct.SetPath(PathHintDBKey, badDB)

	snap, err := p.Fetch(context.Background(), acct)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if snap.Status != core.StatusOK {
		t.Errorf("snap.Status = %v, want StatusOK", snap.Status)
	}
	if snap.Diagnostics["query_error"] == "" {
		t.Error("expected query_error diagnostic for corrupt sqlite source")
	}

	// Test zero conversations recorded
	emptySessDir := filepath.Join(tmpDir, "empty_sess")
	_ = os.MkdirAll(emptySessDir, 0o755)
	acctEmpty := core.AccountConfig{ID: DefaultAccountID, Provider: ID}
	acctEmpty.SetPath(PathHintSessionsDirKey, emptySessDir)
	snapEmpty, err := p.Fetch(context.Background(), acctEmpty)
	if err != nil {
		t.Fatalf("Fetch empty: %v", err)
	}
	if snapEmpty.Status != core.StatusOK || snapEmpty.Message != "No Kiro CLI conversations recorded" {
		t.Errorf("snapEmpty = status:%v, msg:%q", snapEmpty.Status, snapEmpty.Message)
	}
}

// ============================================================================
// Axis 3: Negative & Error Branches
// ============================================================================

func TestKiro_Negative_DatabaseErrors(t *testing.T) {
	t.Run("openReadOnly empty path", func(t *testing.T) {
		db, err := openReadOnly("")
		if err == nil || db != nil {
			t.Errorf("openReadOnly(\"\") = (%v, %v), want error", db, err)
		}
	})

	t.Run("pingContext closed DB", func(t *testing.T) {
		tmp := filepath.Join(t.TempDir(), "dummy.sqlite3")
		_ = os.WriteFile(tmp, []byte(""), 0o600)
		db, err := openReadOnly(tmp)
		if err != nil {
			t.Fatalf("openReadOnly: %v", err)
		}
		db.Close()
		err = pingContext(context.Background(), db)
		if err == nil {
			t.Error("pingContext on closed DB succeeded, want error")
		}
	})

	t.Run("detectConversationsTable non-sqlite file", func(t *testing.T) {
		corruptFile := filepath.Join(t.TempDir(), "corrupt.sqlite3")
		_ = os.WriteFile(corruptFile, []byte("this is not a sqlite database"), 0o600)
		db, err := openReadOnly(corruptFile)
		if err != nil {
			t.Fatalf("openReadOnly: %v", err)
		}
		defer db.Close()
		_, ok, err := detectConversationsTable(context.Background(), db)
		if err == nil {
			t.Errorf("detectConversationsTable on corrupt DB = (%v, %v), want error", ok, err)
		}
	})

	t.Run("queryKiroConversations table without value column", func(t *testing.T) {
		dbPath := filepath.Join(t.TempDir(), "no_val.sqlite3")
		db, err := sql.Open("sqlite3", dbPath)
		if err != nil {
			t.Fatalf("open sqlite: %v", err)
		}
		_, _ = db.Exec(`CREATE TABLE conversations_v2 (key TEXT, other TEXT)`)
		db.Close()

		convs, err := queryKiroConversations(context.Background(), dbPath)
		if err != nil {
			t.Fatalf("queryKiroConversations with missing value col returned error: %v", err)
		}
		if len(convs) != 0 {
			t.Errorf("queryKiroConversations returned %d convs, want 0", len(convs))
		}
	})
}

func TestKiro_Negative_FetchSourcesError(t *testing.T) {
	tmpDir := t.TempDir()
	badDB := filepath.Join(tmpDir, "bad.sqlite3")
	_ = os.WriteFile(badDB, []byte("garbage content"), 0o600)

	sessDir := filepath.Join(tmpDir, "sess")
	_ = os.MkdirAll(sessDir, 0o755)

	p := New()
	acct := core.AccountConfig{ID: DefaultAccountID, Provider: ID}
	acct.SetPath(PathHintDBKey, badDB)
	acct.SetPath(PathHintSessionsDirKey, sessDir)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Pre-cancel context

	snap, err := p.Fetch(ctx, acct)
	if err == nil {
		t.Errorf("Fetch with pre-cancelled ctx & bad DB expected error, got nil (status=%v)", snap.Status)
	}
	if snap.Status != core.StatusError {
		t.Errorf("snap.Status = %v, want StatusError", snap.Status)
	}
	if snap.Diagnostics["query_error"] == "" {
		t.Error("expected query_error diagnostic")
	}
}

func TestKiro_Negative_ReadKiroFileSessionsContextCancel(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, "s1.json"), []byte("{}"), 0o600)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := readKiroFileSessions(ctx, tmpDir)
	if err == nil {
		t.Error("readKiroFileSessions with cancelled context expected error, got nil")
	}
}

// ============================================================================
// Axis 4: Concurrency & Race Conditions
// ============================================================================

func TestKiro_Concurrency_FetchAndItemizedUnderRace(t *testing.T) {
	tmpDir := t.TempDir()
	sessionsDir := filepath.Join(tmpDir, "sessions")
	_ = os.MkdirAll(sessionsDir, 0o755)

	for i := 0; i < 5; i++ {
		name := fmt.Sprintf("session-%d.json", i)
		data := fmt.Sprintf(`{
			"session_id": "session-%d",
			"cwd": "/workspace/concurrent",
			"updated_at": "2026-05-18T12:00:00Z",
			"session_state": {
				"rts_model_state": {"model_info": {"model_id": "claude-concurrent", "context_window_tokens": 100000}},
				"conversation_metadata": {"user_turn_metadatas": [{"input_tokens": %d, "output_tokens": %d}]}
			}
		}`, i, (i+1)*100, (i+1)*20)
		_ = os.WriteFile(filepath.Join(sessionsDir, name), []byte(data), 0o600)
	}

	dbPath := filepath.Join(tmpDir, "data.sqlite3")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	_, _ = db.Exec(`CREATE TABLE conversations_v2 (key TEXT, conversation_id TEXT, value TEXT)`)
	for i := 5; i < 10; i++ {
		val := fmt.Sprintf(`{
			"session_id": "session-%d",
			"cwd": "/workspace/concurrent-db",
			"updated_at": "2026-05-18T12:00:00Z",
			"session_state": {
				"rts_model_state": {"model_info": {"model_id": "claude-concurrent"}},
				"conversation_metadata": {"user_turn_metadatas": [{"input_tokens": %d, "output_tokens": %d}]}
			}
		}`, i, (i+1)*100, (i+1)*20)
		_, _ = db.Exec(`INSERT INTO conversations_v2 (key, conversation_id, value) VALUES (?, ?, ?)`, fmt.Sprintf("k-%d", i), fmt.Sprintf("session-%d", i), val)
	}
	db.Close()

	p := New()
	acct := core.AccountConfig{ID: DefaultAccountID, Provider: ID}
	acct.SetPath(PathHintSessionsDirKey, sessionsDir)
	acct.SetPath(PathHintDBKey, dbPath)

	t.Setenv("KIRO_SESSIONS_DIR", sessionsDir)
	t.Setenv("KIRO_DATA_DIR", tmpDir)

	var wg sync.WaitGroup
	workers := 10
	iterations := 5

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for it := 0; it < iterations; it++ {
				snap, fetchErr := p.Fetch(context.Background(), acct)
				if fetchErr != nil {
					t.Errorf("worker %d Fetch error: %v", workerID, fetchErr)
					return
				}
				if snap.Status != core.StatusOK {
					t.Errorf("worker %d Fetch status = %v, want OK", workerID, snap.Status)
					return
				}

				events, itemErr := p.ItemizedUsage()
				if itemErr != nil {
					t.Errorf("worker %d ItemizedUsage error: %v", workerID, itemErr)
					return
				}
				if len(events) != 10 {
					t.Errorf("worker %d len(events) = %d, want 10", workerID, len(events))
					return
				}

				_, _ = p.HasChanged(acct, time.Now().Add(-1*time.Minute))
			}
		}(w)
	}

	wg.Wait()
}

// ============================================================================
// Axis 5: Domain Invariants & Metadata
// ============================================================================

func TestKiro_DomainInvariants_WidgetsAndMetadata(t *testing.T) {
	p := New()

	if p.ID() != ID {
		t.Errorf("ID = %q, want %q", p.ID(), ID)
	}
	if p.Spec().Info.Name != "Kiro CLI" {
		t.Errorf("Name = %q, want 'Kiro CLI'", p.Spec().Info.Name)
	}

	dash := p.DashboardWidget()
	if dash.IsZero() {
		t.Error("DashboardWidget is zero")
	}

	detail := p.DetailWidget()
	if len(detail.Sections) == 0 {
		t.Error("DetailWidget has no sections")
	}

	pNilClock := &Provider{clock: nil}
	tNil := pNilClock.now()
	if tNil.IsZero() {
		t.Error("pNilClock.now() returned zero time")
	}
}

func TestKiro_DomainInvariants_MetricNormalizationAndStatusMessage(t *testing.T) {
	snap := core.NewUsageSnapshot(ID, DefaultAccountID)
	snap.DailySeries = make(map[string][]core.TimePoint)

	// Test empty conversations
	populateSnapshot(&snap, []kiroConversation{}, time.Now())
	if len(snap.Metrics) != 0 {
		t.Errorf("len(snap.Metrics) = %d, want 0 for empty convs", len(snap.Metrics))
	}
	if msg := buildStatusMessage(snap); msg != "OK" {
		t.Errorf("buildStatusMessage for empty snap = %q, want 'OK'", msg)
	}

	// Test 1 conversation with tokens
	c1 := kiroConversation{
		ConversationID:  "c1",
		Model:           "claude",
		InputTokens:     100,
		OutputTokens:    50,
		TotalTokens:     150,
		HasTokens:       true,
		MessageCount:    1,
		HasMessageCount: true,
		UpdatedAt:       time.Date(2026, 5, 18, 10, 0, 0, 0, time.UTC),
	}
	populateSnapshot(&snap, []kiroConversation{c1}, time.Now())
	if msg := buildStatusMessage(snap); msg != "1 conversation, 150 tokens (est.)" {
		t.Errorf("buildStatusMessage(1 conv) = %q, want '1 conversation, 150 tokens (est.)'", msg)
	}

	// Test multiple conversations
	c2 := kiroConversation{
		ConversationID:  "c2",
		Model:           "claude",
		InputTokens:     200,
		OutputTokens:    100,
		TotalTokens:     300,
		HasTokens:       true,
		MessageCount:    2,
		HasMessageCount: true,
		UpdatedAt:       time.Date(2026, 5, 18, 11, 0, 0, 0, time.UTC),
	}
	snap2 := core.NewUsageSnapshot(ID, DefaultAccountID)
	snap2.DailySeries = make(map[string][]core.TimePoint)
	populateSnapshot(&snap2, []kiroConversation{c1, c2}, time.Now())
	if msg := buildStatusMessage(snap2); msg != "2 conversations, 450 tokens (est.)" {
		t.Errorf("buildStatusMessage(2 convs) = %q, want '2 conversations, 450 tokens (est.)'", msg)
	}
}

func TestKiro_DomainInvariants_MergeKiroConversation(t *testing.T) {
	primary := kiroConversation{
		Key:            "k1",
		ConversationID: "id1",
		Source:         "jsonl",
		Workspace:      "/ws/1",
		Model:          "model1",
		UpdatedAt:      time.Date(2026, 5, 18, 10, 0, 0, 0, time.UTC),
		InputTokens:    100,
		OutputTokens:   50,
		TotalTokens:    150,
		HasTokens:      true,
	}

	secondary := kiroConversation{
		Key:             "k1",
		ConversationID:  "id1",
		Source:          "sqlite",
		Workspace:       "/ws/2",
		Model:           "model2",
		UpdatedAt:       time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
		MessageCount:    5,
		HasMessageCount: true,
	}

	merged := mergeKiroConversation(primary, secondary)
	if merged.Model != "model1" {
		t.Errorf("merged.Model = %q, want model1 (primary preserved)", merged.Model)
	}
	if merged.Workspace != "/ws/1" {
		t.Errorf("merged.Workspace = %q, want /ws/1 (primary preserved)", merged.Workspace)
	}
	if !merged.UpdatedAt.Equal(secondary.UpdatedAt) {
		t.Errorf("merged.UpdatedAt = %v, want %v (newer timestamp wins)", merged.UpdatedAt, secondary.UpdatedAt)
	}
	if merged.Source != "jsonl+sqlite" {
		t.Errorf("merged.Source = %q, want jsonl+sqlite", merged.Source)
	}
	if !merged.HasMessageCount || merged.MessageCount != 5 {
		t.Errorf("merged.MessageCount = %d/%v, want 5/true", merged.MessageCount, merged.HasMessageCount)
	}
	if !merged.HasTokens || merged.TotalTokens != 150 {
		t.Errorf("merged.TotalTokens = %d/%v, want 150/true", merged.TotalTokens, merged.HasTokens)
	}

	// Test primary with empty values taking secondary values
	primaryEmpty := kiroConversation{}
	secondaryFull := kiroConversation{
		Model:           "secondary-model",
		Workspace:       "/ws/secondary",
		UpdatedAt:       time.Date(2026, 5, 18, 15, 0, 0, 0, time.UTC),
		Source:          "sqlite",
		InputTokens:     500,
		OutputTokens:    100,
		TotalTokens:     600,
		HasTokens:       true,
		MessageCount:    3,
		HasMessageCount: true,
	}
	mergedFull := mergeKiroConversation(primaryEmpty, secondaryFull)
	if mergedFull.Model != "secondary-model" || mergedFull.Workspace != "/ws/secondary" || mergedFull.Source != "sqlite" {
		t.Errorf("mergedFull unexpected values: %+v", mergedFull)
	}
	if !mergedFull.HasTokens || mergedFull.TotalTokens != 600 {
		t.Errorf("mergedFull tokens: %d/%v", mergedFull.TotalTokens, mergedFull.HasTokens)
	}
	if !mergedFull.HasMessageCount || mergedFull.MessageCount != 3 {
		t.Errorf("mergedFull message count: %d/%v", mergedFull.MessageCount, mergedFull.HasMessageCount)
	}

	cEmpty := kiroConversation{Source: "raw"}
	key := kiroConversationKey(cEmpty)
	if key != "" {
		t.Errorf("kiroConversationKey for empty conv = %q, want empty", key)
	}

	cKeyOnly := kiroConversation{Key: "custom-key"}
	if key := kiroConversationKey(cKeyOnly); key != "key:custom-key" {
		t.Errorf("kiroConversationKey(cKeyOnly) = %q, want key:custom-key", key)
	}

	mergedList := mergeKiroConversations([]kiroConversation{cEmpty, cEmpty})
	if len(mergedList) != 2 {
		t.Errorf("mergeKiroConversations for items without key len = %d, want 2", len(mergedList))
	}
}
