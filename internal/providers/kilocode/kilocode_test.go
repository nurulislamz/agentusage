package kilocode

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/nurulislamz/agentusage/internal/core"
	"github.com/nurulislamz/agentusage/internal/providers/roocode"
)

type fixedClock struct {
	t time.Time
}

func (f fixedClock) Now() time.Time {
	return f.t
}

func floatEquals(a, b, epsilon float64) bool {
	return math.Abs(a-b) <= epsilon
}

func createTestTask(t *testing.T, tasksRoot, taskName, uiJSON, convHistoryJSON string) string {
	t.Helper()
	taskDir := filepath.Join(tasksRoot, taskName)
	if err := os.MkdirAll(taskDir, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", taskDir, err)
	}
	if uiJSON != "" {
		uiPath := filepath.Join(taskDir, roocode.UIMessagesFile)
		if err := os.WriteFile(uiPath, []byte(uiJSON), 0o600); err != nil {
			t.Fatalf("write %s: %v", uiPath, err)
		}
	}
	if convHistoryJSON != "" {
		historyPath := filepath.Join(taskDir, roocode.APIConversationHistoryFile)
		if err := os.WriteFile(historyPath, []byte(convHistoryJSON), 0o600); err != nil {
			t.Fatalf("write %s: %v", historyPath, err)
		}
	}
	return taskDir
}

// -----------------------------------------------------------------------------
// Axis 1: Happy Paths
// -----------------------------------------------------------------------------

func TestProvider_BasicMetadata(t *testing.T) {
	p := New()
	if got, want := p.ID(), ID; got != want {
		t.Errorf("ID = %q, want %q", got, want)
	}
	if got, want := p.ID(), "kilo_code"; got != want {
		t.Errorf("ID = %q, want %q", got, want)
	}
	if got, want := DefaultAccountID, "kilo_code"; got != want {
		t.Errorf("DefaultAccountID = %q, want %q", got, want)
	}

	info := p.Describe()
	if info.Name != "Kilo Code" {
		t.Errorf("Describe().Name = %q, want Kilo Code", info.Name)
	}
	if info.DocURL != "https://github.com/Kilo-Org/kilocode" {
		t.Errorf("Describe().DocURL = %q, want Kilo DocURL", info.DocURL)
	}
	if len(info.Capabilities) != 4 {
		t.Errorf("Describe().Capabilities len = %d, want 4", len(info.Capabilities))
	}

	spec := p.Spec()
	if spec.Auth.Type != core.ProviderAuthTypeLocal {
		t.Errorf("auth type = %v, want local", spec.Auth.Type)
	}
	if spec.Auth.DefaultAccountID != "kilo_code" {
		t.Errorf("spec DefaultAccountID = %q, want kilo_code", spec.Auth.DefaultAccountID)
	}
	if len(spec.Setup.Quickstart) < 2 {
		t.Errorf("spec Quickstart steps = %d, want >= 2", len(spec.Setup.Quickstart))
	}

	dash := p.DashboardWidget()
	if dash.IsZero() {
		t.Fatal("DashboardWidget is zero")
	}
	if dash.ColorRole != core.DashboardColorRoleMauve {
		t.Errorf("DashboardWidget ColorRole = %v, want Mauve", dash.ColorRole)
	}
	if len(dash.CompactRows) == 0 {
		t.Errorf("DashboardWidget CompactRows is empty")
	}
}

func TestProvider_DetailWidget(t *testing.T) {
	p := New()
	widget := p.DetailWidget()
	if len(widget.Sections) != 7 {
		t.Errorf("DetailWidget sections len = %d, want 7", len(widget.Sections))
	}
	expectedStyles := []core.DetailSectionStyle{
		core.DetailSectionStyleUsage,
		core.DetailSectionStyleModels,
		core.DetailSectionStyleLanguages,
		core.DetailSectionStyleSpending,
		core.DetailSectionStyleTrends,
		core.DetailSectionStyleTokens,
		core.DetailSectionStyleActivity,
	}
	for i, style := range expectedStyles {
		if i < len(widget.Sections) && widget.Sections[i].Style != style {
			t.Errorf("section %d style = %v, want %v", i, widget.Sections[i].Style, style)
		}
	}
}

func TestProvider_Fetch_HappyPath_CompleteMetrics(t *testing.T) {
	tasksRoot := t.TempDir()
	now := time.Date(2026, 8, 31, 14, 0, 0, 0, time.UTC)
	tsMS := now.UnixMilli()

	uiContent := fmt.Sprintf(`[
		{"say":"api_req_started","ts":%d,"text":"{\"cost\":0.015,\"tokensIn\":1200,\"tokensOut\":350,\"cacheReads\":150,\"cacheWrites\":50,\"apiProtocol\":\"anthropic/v1\"}"},
		{"say":"api_req_started","ts":%d,"text":"{\"cost\":0.025,\"tokensIn\":800,\"tokensOut\":250,\"cacheReads\":50,\"cacheWrites\":20,\"apiProtocol\":\"anthropic/v1\"}"}
	]`, tsMS, tsMS+1000)
	convContent := `[{"role":"user","content":"<model>kilo-claude-3-7-sonnet</model>"}]`

	createTestTask(t, tasksRoot, "task-happy-1", uiContent, convContent)

	p := New()
	p.clock = fixedClock{t: now}
	acct := core.AccountConfig{ID: "kilo_code", Provider: "kilo_code", Auth: "local"}
	acct.SetPath("tasks_dir", tasksRoot)

	snap, err := p.Fetch(context.Background(), acct)
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}

	if snap.Status != core.StatusOK {
		t.Fatalf("Status = %v, want StatusOK (message: %s)", snap.Status, snap.Message)
	}
	if snap.ProviderID != "kilo_code" {
		t.Errorf("ProviderID = %q, want kilo_code", snap.ProviderID)
	}
	if snap.AccountID != "kilo_code" {
		t.Errorf("AccountID = %q, want kilo_code", snap.AccountID)
	}

	expectedMetrics := map[string]float64{
		"total_tasks":              1,
		"tasks_today":              1,
		"tasks_7d":                 1,
		"total_requests":           2,
		"total_input_tokens":       2000,
		"total_output_tokens":      600,
		"total_cache_read_tokens":  200,
		"total_cache_write_tokens": 70,
		"total_tokens":             2870,
		"total_cost_usd":           0.040,
		"today_cost_usd":           0.040,
	}

	for k, want := range expectedMetrics {
		m, ok := snap.Metrics[k]
		if !ok {
			t.Errorf("missing metric %q", k)
			continue
		}
		if m.Used == nil {
			t.Errorf("metric %q Used is nil", k)
			continue
		}
		if !floatEquals(*m.Used, want, 1e-6) {
			t.Errorf("metric %q = %v, want %v", k, *m.Used, want)
		}
	}

	// Verify DailySeries
	if len(snap.DailySeries["tasks"]) != 1 {
		t.Errorf("DailySeries[tasks] len = %d, want 1", len(snap.DailySeries["tasks"]))
	}
	if len(snap.DailySeries["tokens"]) != 1 {
		t.Errorf("DailySeries[tokens] len = %d, want 1", len(snap.DailySeries["tokens"]))
	}
	if len(snap.DailySeries["cost"]) != 1 {
		t.Errorf("DailySeries[cost] len = %d, want 1", len(snap.DailySeries["cost"]))
	}

	// Verify ModelUsage
	if len(snap.ModelUsage) != 1 {
		t.Fatalf("ModelUsage len = %d, want 1", len(snap.ModelUsage))
	}
	rec := snap.ModelUsage[0]
	if rec.RawModelID != "kilo-claude-3-7-sonnet" {
		t.Errorf("RawModelID = %q, want kilo-claude-3-7-sonnet", rec.RawModelID)
	}
	if rec.Requests == nil || *rec.Requests != 2 {
		t.Errorf("Requests = %v, want 2", rec.Requests)
	}
	if rec.Dimensions["upstream_provider"] != "anthropic" {
		t.Errorf("upstream_provider = %q, want anthropic", rec.Dimensions["upstream_provider"])
	}
	if rec.CostUSD == nil || !floatEquals(*rec.CostUSD, 0.040, 1e-6) {
		t.Errorf("CostUSD = %v, want 0.040", rec.CostUSD)
	}

	// Verify status message
	if snap.Message == "" || snap.Message == "Kilo Code OK" {
		t.Errorf("Message = %q, want detailed summary message", snap.Message)
	}
}

func TestProvider_Fetch_MultipleTasksAndModels(t *testing.T) {
	tasksRoot := t.TempDir()
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	todayMS := now.UnixMilli()
	twoDaysAgoMS := now.AddDate(0, 0, -2).UnixMilli()
	tenDaysAgoMS := now.AddDate(0, 0, -10).UnixMilli()

	// Task 1: Today (Anthropic / Claude)
	ui1 := fmt.Sprintf(`[{"say":"api_req_started","ts":%d,"text":"{\"cost\":0.05,\"tokensIn\":1000,\"tokensOut\":500,\"cacheReads\":100,\"cacheWrites\":50,\"apiProtocol\":\"anthropic\"}"}]`, todayMS)
	hist1 := `[{"content":"<model>claude-sonnet-4-5</model>"}]`
	createTestTask(t, tasksRoot, "task-1", ui1, hist1)

	// Task 2: 2 days ago (OpenAI / GPT-4o)
	ui2 := fmt.Sprintf(`[{"say":"api_req_started","ts":%d,"text":"{\"cost\":0.03,\"tokensIn\":800,\"tokensOut\":400,\"apiProtocol\":\"openai\"}"}]`, twoDaysAgoMS)
	hist2 := `[{"content":"<model>gpt-4o-mini</model>"}]`
	createTestTask(t, tasksRoot, "task-2", ui2, hist2)

	// Task 3: 10 days ago (DeepSeek / DeepSeek-V3)
	ui3 := fmt.Sprintf(`[{"say":"api_req_started","ts":%d,"text":"{\"cost\":0.02,\"tokensIn\":600,\"tokensOut\":300,\"apiProtocol\":\"deepseek\"}"}]`, tenDaysAgoMS)
	hist3 := `[{"content":"<model>deepseek-coder</model>"}]`
	createTestTask(t, tasksRoot, "task-3", ui3, hist3)

	p := New()
	p.clock = fixedClock{t: now}
	acct := core.AccountConfig{ID: "kilo_code", Provider: "kilo_code"}
	acct.SetPath("tasks_dir", tasksRoot)

	snap, err := p.Fetch(context.Background(), acct)
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}

	if got := *snap.Metrics["total_tasks"].Used; got != 3 {
		t.Errorf("total_tasks = %v, want 3", got)
	}
	if got := *snap.Metrics["tasks_today"].Used; got != 1 {
		t.Errorf("tasks_today = %v, want 1", got)
	}
	if got := *snap.Metrics["tasks_7d"].Used; got != 2 {
		t.Errorf("tasks_7d = %v, want 2 (tasks 1 and 2)", got)
	}
	if got := *snap.Metrics["total_requests"].Used; got != 3 {
		t.Errorf("total_requests = %v, want 3", got)
	}
	if got := *snap.Metrics["total_cost_usd"].Used; !floatEquals(got, 0.10, 1e-6) {
		t.Errorf("total_cost_usd = %v, want 0.10", got)
	}
	if got := *snap.Metrics["today_cost_usd"].Used; !floatEquals(got, 0.05, 1e-6) {
		t.Errorf("today_cost_usd = %v, want 0.05", got)
	}

	if len(snap.DailySeries["tasks"]) != 3 {
		t.Errorf("DailySeries[tasks] len = %d, want 3", len(snap.DailySeries["tasks"]))
	}
	if len(snap.ModelUsage) != 3 {
		t.Fatalf("ModelUsage len = %d, want 3", len(snap.ModelUsage))
	}

	modelMap := make(map[string]core.ModelUsageRecord)
	for _, rec := range snap.ModelUsage {
		modelMap[rec.RawModelID] = rec
	}

	if rec, ok := modelMap["claude-sonnet-4-5"]; !ok || rec.Dimensions["upstream_provider"] != "anthropic" {
		t.Errorf("claude-sonnet-4-5 record invalid: %+v", rec)
	}
	if rec, ok := modelMap["gpt-4o-mini"]; !ok || rec.Dimensions["upstream_provider"] != "openai" {
		t.Errorf("gpt-4o-mini record invalid: %+v", rec)
	}
	if rec, ok := modelMap["deepseek-coder"]; !ok || rec.Dimensions["upstream_provider"] != "deepseek" {
		t.Errorf("deepseek-coder record invalid: %+v", rec)
	}
}

func TestProvider_Fetch_AutoDiscoveryGlobalStorage(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpHome, ".config"))

	// Create VS Code Kilo Code extension directory structure
	extTasksDir := filepath.Join(tmpHome, ".config", "Code", "User", "globalStorage", roocode.KiloExtensionSubdir, "tasks")
	createTestTask(t, extTasksDir, "discovered-task-1",
		`[{"say":"api_req_started","ts":1716033600000,"text":"{\"cost\":0.02,\"tokensIn\":400,\"tokensOut\":200,\"apiProtocol\":\"anthropic\"}"}]`,
		`[{"content":"<model>kilo-discovered-model</model>"}]`,
	)

	p := New()
	p.clock = fixedClock{t: time.Date(2024, 5, 19, 12, 0, 0, 0, time.UTC)}
	acct := core.AccountConfig{ID: "kilo_code"} // No tasks_dir override

	snap, err := p.Fetch(context.Background(), acct)
	if err != nil {
		t.Fatalf("Fetch() auto-discovery error: %v", err)
	}

	if snap.Status != core.StatusOK {
		t.Fatalf("Status = %v, want StatusOK", snap.Status)
	}
	if got := *snap.Metrics["total_tasks"].Used; got != 1 {
		t.Errorf("total_tasks = %v, want 1", got)
	}
	if len(snap.ModelUsage) != 1 || snap.ModelUsage[0].RawModelID != "kilo-discovered-model" {
		t.Errorf("ModelUsage = %+v, want kilo-discovered-model", snap.ModelUsage)
	}
}

func TestProvider_Fetch_CrossVariantDedup(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpHome, ".config"))

	// Write identical task to Code and Cursor variants
	codeTasks := filepath.Join(tmpHome, ".config", "Code", "User", "globalStorage", roocode.KiloExtensionSubdir, "tasks")
	cursorTasks := filepath.Join(tmpHome, ".config", "Cursor", "User", "globalStorage", roocode.KiloExtensionSubdir, "tasks")

	sameUI := `[{"say":"api_req_started","ts":1716033600000,"text":"{\"cost\":0.01,\"tokensIn\":100,\"tokensOut\":50,\"apiProtocol\":\"anthropic\"}"}]`
	sameHist := `[{"content":"<model>dedup-model</model>"}]`

	createTestTask(t, codeTasks, "task-dup", sameUI, sameHist)
	createTestTask(t, cursorTasks, "task-dup", sameUI, sameHist)

	p := New()
	p.clock = fixedClock{t: time.Date(2024, 5, 19, 12, 0, 0, 0, time.UTC)}
	acct := core.AccountConfig{ID: "kilo_code"}

	snap, err := p.Fetch(context.Background(), acct)
	if err != nil {
		t.Fatalf("Fetch() dedup error: %v", err)
	}

	if snap.Status != core.StatusOK {
		t.Fatalf("Status = %v, want StatusOK", snap.Status)
	}
	// Deduplication ensures identical calls across variants don't double metrics
	if got := *snap.Metrics["total_requests"].Used; got != 1 {
		t.Errorf("total_requests = %v, want 1 after dedup", got)
	}
	if got := *snap.Metrics["total_tokens"].Used; got != 150 {
		t.Errorf("total_tokens = %v, want 150 after dedup", got)
	}
}

func TestProvider_Fetch_DefaultAccountIDAndEmptyProvider(t *testing.T) {
	tasksRoot := t.TempDir()
	createTestTask(t, tasksRoot, "task-1",
		`[{"say":"api_req_started","ts":1716033600000,"text":"{\"cost\":0.01,\"tokensIn\":50,\"tokensOut\":50,\"apiProtocol\":\"anthropic\"}"}]`,
		`[{"content":"<model>model-a</model>"}]`,
	)

	p := New()
	// Account with empty Provider
	acct := core.AccountConfig{ID: "custom-account", Provider: "   "}
	acct.SetPath("tasks_dir", tasksRoot)

	snap, err := p.Fetch(context.Background(), acct)
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}

	if snap.ProviderID != "kilo_code" {
		t.Errorf("snap.ProviderID = %q, want kilo_code", snap.ProviderID)
	}
	if snap.AccountID != "custom-account" {
		t.Errorf("snap.AccountID = %q, want custom-account", snap.AccountID)
	}
}

func TestProvider_ItemizedUsage_HappyPath(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpHome, ".config"))

	extTasksDir := filepath.Join(tmpHome, ".config", "Code", "User", "globalStorage", roocode.KiloExtensionSubdir, "tasks")
	createTestTask(t, extTasksDir, "itemized-task-1",
		`[
			{"say":"api_req_started","ts":1716033600000,"text":"{\"cost\":0.05,\"tokensIn\":500,\"tokensOut\":250,\"cacheReads\":50,\"cacheWrites\":25,\"apiProtocol\":\"anthropic\"}"},
			{"say":"api_req_started","ts":1716033605000,"text":"{\"cost\":0.02,\"tokensIn\":200,\"tokensOut\":100,\"cacheReads\":10,\"cacheWrites\":5,\"apiProtocol\":\"anthropic\"}"}
		]`,
		`[{"content":"<model>kilo-claude-3-5</model>"}]`,
	)

	p := New()
	events, err := p.ItemizedUsage()
	if err != nil {
		t.Fatalf("ItemizedUsage() error: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("ItemizedUsage() returned %d events, want 2", len(events))
	}

	ev1 := events[0]
	if ev1.ProviderID != "kilo_code" {
		t.Errorf("ev1.ProviderID = %q, want kilo_code", ev1.ProviderID)
	}
	if ev1.Session != "itemized-task-1" {
		t.Errorf("ev1.Session = %q, want itemized-task-1", ev1.Session)
	}
	if ev1.Model != "kilo-claude-3-5" {
		t.Errorf("ev1.Model = %q, want kilo-claude-3-5", ev1.Model)
	}
	if ev1.InputTokens != 500 || ev1.OutputTokens != 250 || ev1.CacheReadTokens != 50 || ev1.CacheCreationTokens != 25 {
		t.Errorf("ev1 tokens mismatch: In=%d, Out=%d, CacheRd=%d, CacheWr=%d",
			ev1.InputTokens, ev1.OutputTokens, ev1.CacheReadTokens, ev1.CacheCreationTokens)
	}
	if !floatEquals(ev1.CostUSD, 0.05, 1e-6) || !ev1.HasCost {
		t.Errorf("ev1 cost mismatch: CostUSD=%v, HasCost=%v", ev1.CostUSD, ev1.HasCost)
	}
	if ev1.Time.IsZero() {
		t.Errorf("ev1 Time is zero")
	}

	ev2 := events[1]
	if ev2.InputTokens != 200 || ev2.OutputTokens != 100 {
		t.Errorf("ev2 tokens mismatch: In=%d, Out=%d", ev2.InputTokens, ev2.OutputTokens)
	}
	if !floatEquals(ev2.CostUSD, 0.02, 1e-6) {
		t.Errorf("ev2 cost mismatch: CostUSD=%v", ev2.CostUSD)
	}
}

func TestProvider_HasChanged_Detected(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpHome, ".config"))

	extDir := filepath.Join(tmpHome, ".config", "Code", "User", "globalStorage", roocode.KiloExtensionSubdir)
	tasksDir := filepath.Join(extDir, "tasks")
	if err := os.MkdirAll(tasksDir, 0o755); err != nil {
		t.Fatal(err)
	}

	p := New()
	acct := core.AccountConfig{ID: "kilo_code"}

	// Past time: directory created recently, so since in the past returns true
	past := time.Now().Add(-1 * time.Hour)
	changed, err := p.HasChanged(acct, past)
	if err != nil {
		t.Fatalf("HasChanged() error: %v", err)
	}
	if !changed {
		t.Errorf("HasChanged(past) = false, want true")
	}

	// Future time: since is in the future, so returns false
	future := time.Now().Add(1 * time.Hour)
	changed, err = p.HasChanged(acct, future)
	if err != nil {
		t.Fatalf("HasChanged() error: %v", err)
	}
	if changed {
		t.Errorf("HasChanged(future) = true, want false")
	}
}

// -----------------------------------------------------------------------------
// Axis 2: Edge & Boundary Cases
// -----------------------------------------------------------------------------

func TestProvider_Fetch_NilClock(t *testing.T) {
	tasksRoot := t.TempDir()
	createTestTask(t, tasksRoot, "task-1",
		`[{"say":"api_req_started","ts":1716033600000,"text":"{\"cost\":0.01,\"tokensIn\":10,\"tokensOut\":10,\"apiProtocol\":\"anthropic\"}"}]`,
		`[{"content":"<model>kilo-model</model>"}]`,
	)

	p := New()
	p.clock = nil // Force nil clock to test fallback to core.SystemClock{}
	acct := core.AccountConfig{ID: "kilo_code"}
	acct.SetPath("tasks_dir", tasksRoot)

	snap, err := p.Fetch(context.Background(), acct)
	if err != nil {
		t.Fatalf("Fetch() with nil clock failed: %v", err)
	}
	if snap.Status != core.StatusOK {
		t.Errorf("Status = %v, want StatusOK", snap.Status)
	}
	if snap.Timestamp.IsZero() {
		t.Errorf("Timestamp is zero with nil clock")
	}
}

func TestProvider_Fetch_EmptyTasksDirectory(t *testing.T) {
	tasksRoot := t.TempDir() // Directory with 0 subdirectories

	p := New()
	p.clock = fixedClock{t: time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)}
	acct := core.AccountConfig{ID: "kilo_code"}
	acct.SetPath("tasks_dir", tasksRoot)

	snap, err := p.Fetch(context.Background(), acct)
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}
	// When tasks directory has 0 task subdirectories, resolveTaskDirs returns empty and Status is StatusUnknown
	if snap.Status != core.StatusUnknown {
		t.Errorf("Status = %v, want StatusUnknown", snap.Status)
	}
	if snap.Message != "Kilo Code extension data not found" {
		t.Errorf("Message = %q, want 'Kilo Code extension data not found'", snap.Message)
	}
}

func TestProvider_Fetch_TasksWithoutUIMessages(t *testing.T) {
	tasksRoot := t.TempDir()

	// Task with only conversation history (incomplete task)
	createTestTask(t, tasksRoot, "task-incomplete", "", `[{"content":"<model>test</model>"}]`)
	// Empty folder
	if err := os.MkdirAll(filepath.Join(tasksRoot, "task-empty"), 0o755); err != nil {
		t.Fatal(err)
	}

	p := New()
	acct := core.AccountConfig{ID: "kilo_code"}
	acct.SetPath("tasks_dir", tasksRoot)

	snap, err := p.Fetch(context.Background(), acct)
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}
	if snap.Status != core.StatusOK {
		t.Errorf("Status = %v, want StatusOK", snap.Status)
	}
	if snap.Message != "No Kilo Code usage recorded yet" {
		t.Errorf("Message = %q, want 'No Kilo Code usage recorded yet'", snap.Message)
	}
}

func TestProvider_Fetch_EmptyUIMessagesArray(t *testing.T) {
	tasksRoot := t.TempDir()
	createTestTask(t, tasksRoot, "task-empty-array", "[]", "")
	createTestTask(t, tasksRoot, "task-blank", "   \n\t  ", "")

	p := New()
	acct := core.AccountConfig{ID: "kilo_code"}
	acct.SetPath("tasks_dir", tasksRoot)

	snap, err := p.Fetch(context.Background(), acct)
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}
	if snap.Status != core.StatusOK {
		t.Errorf("Status = %v, want StatusOK", snap.Status)
	}
	if snap.Message != "No Kilo Code usage recorded yet" {
		t.Errorf("Message = %q, want 'No Kilo Code usage recorded yet'", snap.Message)
	}
}

func TestProvider_Fetch_NonAPIReqStartedMessages(t *testing.T) {
	tasksRoot := t.TempDir()

	// Mixed entries: text message, entry_type variant, type variant, unknown say
	uiJSON := `[
		{"say":"text","text":"Hello world"},
		{"say":"user_feedback","text":"Great work"},
		{"entry_type":"api_req_started","ts":1716033600000,"text":"{\"cost\":0.01,\"tokensIn\":100,\"tokensOut\":50,\"apiProtocol\":\"anthropic\"}"},
		{"type":"api_req_started","ts":1716033601000,"text":"{\"cost\":0.02,\"tokensIn\":200,\"tokensOut\":100,\"apiProtocol\":\"anthropic\"}"},
		{"say":"api_req_started","text":""}
	]`
	createTestTask(t, tasksRoot, "task-mixed", uiJSON, `[{"content":"<model>schema-variants</model>"}]`)

	p := New()
	p.clock = fixedClock{t: time.Date(2024, 5, 19, 12, 0, 0, 0, time.UTC)}
	acct := core.AccountConfig{ID: "kilo_code"}
	acct.SetPath("tasks_dir", tasksRoot)

	snap, err := p.Fetch(context.Background(), acct)
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}

	if got := *snap.Metrics["total_requests"].Used; got != 2 {
		t.Errorf("total_requests = %v, want 2 (from entry_type and type variants)", got)
	}
	if got := *snap.Metrics["total_tokens"].Used; got != 450 {
		t.Errorf("total_tokens = %v, want 450", got)
	}
}

func TestProvider_Fetch_BOMHandling(t *testing.T) {
	tasksRoot := t.TempDir()
	taskDir := filepath.Join(tasksRoot, "task-bom")
	if err := os.MkdirAll(taskDir, 0o755); err != nil {
		t.Fatal(err)
	}

	bom := []byte{0xEF, 0xBB, 0xBF}
	jsonBody := []byte(`[{"say":"api_req_started","ts":1716033600000,"text":"{\"cost\":0.01,\"tokensIn\":10,\"tokensOut\":10,\"apiProtocol\":\"anthropic\"}"}]`)
	fullContent := append(bom, jsonBody...)

	if err := os.WriteFile(filepath.Join(taskDir, roocode.UIMessagesFile), fullContent, 0o600); err != nil {
		t.Fatal(err)
	}

	p := New()
	acct := core.AccountConfig{ID: "kilo_code"}
	acct.SetPath("tasks_dir", tasksRoot)

	snap, err := p.Fetch(context.Background(), acct)
	if err != nil {
		t.Fatalf("Fetch() BOM error: %v", err)
	}
	if snap.Status != core.StatusOK {
		t.Errorf("Status = %v, want StatusOK", snap.Status)
	}
	if got := *snap.Metrics["total_requests"].Used; got != 1 {
		t.Errorf("total_requests = %v, want 1", got)
	}
}

func TestProvider_Fetch_NumericAndStringTimestampFormats(t *testing.T) {
	tasksRoot := t.TempDir()

	uiJSON := `[
		{"say":"api_req_started","ts":1716033600000,"text":"{\"cost\":0.01,\"tokensIn\":10,\"tokensOut\":10,\"apiProtocol\":\"openai\"}"},
		{"say":"api_req_started","ts":1716033601,"text":"{\"cost\":0.01,\"tokensIn\":10,\"tokensOut\":10,\"apiProtocol\":\"openai\"}"},
		{"say":"api_req_started","ts":"2024-05-18T12:00:02Z","text":"{\"cost\":0.01,\"tokensIn\":10,\"tokensOut\":10,\"apiProtocol\":\"openai\"}"},
		{"say":"api_req_started","ts":"2024-05-18T12:00:03.123456Z","text":"{\"cost\":0.01,\"tokensIn\":10,\"tokensOut\":10,\"apiProtocol\":\"openai\"}"},
		{"say":"api_req_started","ts":"2024-05-18 12:00:04","text":"{\"cost\":0.01,\"tokensIn\":10,\"tokensOut\":10,\"apiProtocol\":\"openai\"}"},
		{"say":"api_req_started","ts":0,"text":"{\"cost\":0.01,\"tokensIn\":10,\"tokensOut\":10,\"apiProtocol\":\"openai\"}"},
		{"say":"api_req_started","ts":"invalid-time-string","text":"{\"cost\":0.01,\"tokensIn\":10,\"tokensOut\":10,\"apiProtocol\":\"openai\"}"}
	]`
	createTestTask(t, tasksRoot, "task-timestamps", uiJSON, `[{"content":"<model>time-model</model>"}]`)

	p := New()
	p.clock = fixedClock{t: time.Date(2024, 5, 18, 12, 0, 0, 0, time.UTC)}
	acct := core.AccountConfig{ID: "kilo_code"}
	acct.SetPath("tasks_dir", tasksRoot)

	snap, err := p.Fetch(context.Background(), acct)
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}

	if got := *snap.Metrics["total_requests"].Used; got != 7 {
		t.Errorf("total_requests = %v, want 7", got)
	}
}

func TestProvider_Fetch_FloatAndCommaTokenCounts(t *testing.T) {
	tasksRoot := t.TempDir()

	uiJSON := `[
		{"say":"api_req_started","ts":1716033600000,"text":"{\"cost\":1.50,\"tokensIn\":1000.0,\"tokensOut\":500.0,\"cacheReads\":250.0,\"cacheWrites\":100,\"apiProtocol\":\"anthropic\"}"}
	]`
	createTestTask(t, tasksRoot, "task-coercion", uiJSON, `[{"content":"<model>coercion-model</model>"}]`)

	p := New()
	acct := core.AccountConfig{ID: "kilo_code"}
	acct.SetPath("tasks_dir", tasksRoot)

	snap, err := p.Fetch(context.Background(), acct)
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}

	if got := *snap.Metrics["total_input_tokens"].Used; got != 1000 {
		t.Errorf("total_input_tokens = %v, want 1000", got)
	}
	if got := *snap.Metrics["total_output_tokens"].Used; got != 500 {
		t.Errorf("total_output_tokens = %v, want 500", got)
	}
	if got := *snap.Metrics["total_cost_usd"].Used; !floatEquals(got, 1.50, 1e-6) {
		t.Errorf("total_cost_usd = %v, want 1.50", got)
	}
}

func TestProvider_Fetch_NegativeAndNaNValuesClamped(t *testing.T) {
	tasksRoot := t.TempDir()

	uiJSON := `[
		{"say":"api_req_started","ts":1716033600000,"text":"{\"cost\":-5.0,\"tokensIn\":-100,\"tokensOut\":-50,\"cacheReads\":-10,\"cacheWrites\":-5,\"apiProtocol\":\"anthropic\"}"}
	]`
	createTestTask(t, tasksRoot, "task-negatives", uiJSON, `[{"content":"<model>clamp-model</model>"}]`)

	p := New()
	acct := core.AccountConfig{ID: "kilo_code"}
	acct.SetPath("tasks_dir", tasksRoot)

	snap, err := p.Fetch(context.Background(), acct)
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}

	if snap.Status != core.StatusOK {
		t.Errorf("Status = %v, want StatusOK", snap.Status)
	}
	if _, ok := snap.Metrics["total_cost_usd"]; ok {
		t.Errorf("total_cost_usd should not be set when <= 0")
	}
	if _, ok := snap.Metrics["total_tokens"]; ok {
		t.Errorf("total_tokens should not be set when <= 0")
	}
	if len(snap.ModelUsage) != 1 {
		t.Fatalf("ModelUsage len = %d, want 1", len(snap.ModelUsage))
	}
	rec := snap.ModelUsage[0]
	if rec.TotalTokens == nil || *rec.TotalTokens != 0 {
		t.Errorf("ModelUsage TotalTokens = %v, want 0 after clamping", rec.TotalTokens)
	}
	if rec.CostUSD != nil {
		t.Errorf("ModelUsage CostUSD = %v, want nil when non-positive", rec.CostUSD)
	}
}

func TestProvider_Fetch_ModelFallbacks_SlugAndName(t *testing.T) {
	tasksRoot := t.TempDir()

	// Task 1: Fallback to <slug>
	createTestTask(t, tasksRoot, "task-slug",
		`[{"say":"api_req_started","ts":1716033600000,"text":"{\"cost\":0.01,\"tokensIn\":10,\"tokensOut\":10,\"apiProtocol\":\"openai\"}"}]`,
		`[{"content":"<slug>kilo-slug-model</slug>"}]`,
	)

	// Task 2: Fallback to <name>
	createTestTask(t, tasksRoot, "task-name",
		`[{"say":"api_req_started","ts":1716033600000,"text":"{\"cost\":0.01,\"tokensIn\":10,\"tokensOut\":10,\"apiProtocol\":\"anthropic\"}"}]`,
		`[{"content":"<name>kilo-name-model</name>"}]`,
	)

	// Task 3: No conversation history
	createTestTask(t, tasksRoot, "task-no-history",
		`[{"say":"api_req_started","ts":1716033600000,"text":"{\"cost\":0.01,\"tokensIn\":10,\"tokensOut\":10,\"apiProtocol\":\"google\"}"}]`,
		"",
	)

	p := New()
	acct := core.AccountConfig{ID: "kilo_code"}
	acct.SetPath("tasks_dir", tasksRoot)

	snap, err := p.Fetch(context.Background(), acct)
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}

	modelIDs := make(map[string]bool)
	for _, rec := range snap.ModelUsage {
		modelIDs[rec.RawModelID] = true
	}

	if !modelIDs["kilo-slug-model"] {
		t.Errorf("missing fallback model 'kilo-slug-model'")
	}
	if !modelIDs["kilo-name-model"] {
		t.Errorf("missing fallback model 'kilo-name-model'")
	}
	if !modelIDs["unknown"] {
		t.Errorf("missing 'unknown' fallback model when no history present")
	}
}

func TestProvider_Fetch_TokenAndCostFormattingThresholds(t *testing.T) {
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	tsMS := now.UnixMilli()

	tests := []struct {
		name        string
		taskCount   int
		tokensIn    int64
		tokensOut   int64
		cost        float64
		wantMsgPart []string
	}{
		{
			name:        "million tokens and dollar cost",
			taskCount:   1,
			tokensIn:    1_500_000,
			tokensOut:   500_000,
			cost:        12.50,
			wantMsgPart: []string{"1 task", "2.0M tokens", "$12.50"},
		},
		{
			name:        "thousand tokens and sub-dollar cost",
			taskCount:   2,
			tokensIn:    15_000,
			tokensOut:   5_000,
			cost:        0.0450,
			wantMsgPart: []string{"2 tasks", "20.0k tokens", "$0.0450"},
		},
		{
			name:        "small tokens and zero cost",
			taskCount:   1,
			tokensIn:    50,
			tokensOut:   20,
			cost:        0.0,
			wantMsgPart: []string{"1 task", "70 tokens"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tasksRoot := t.TempDir()
			for i := 0; i < tt.taskCount; i++ {
				uiJSON := fmt.Sprintf(`[{"say":"api_req_started","ts":%d,"text":"{\"cost\":%f,\"tokensIn\":%d,\"tokensOut\":%d,\"apiProtocol\":\"anthropic\"}"}]`,
					tsMS, tt.cost/float64(tt.taskCount), tt.tokensIn/int64(tt.taskCount), tt.tokensOut/int64(tt.taskCount))
				createTestTask(t, tasksRoot, fmt.Sprintf("task-%d", i), uiJSON, `[{"content":"<model>m</model>"}]`)
			}

			p := New()
			p.clock = fixedClock{t: now}
			acct := core.AccountConfig{ID: "kilo_code"}
			acct.SetPath("tasks_dir", tasksRoot)

			snap, err := p.Fetch(context.Background(), acct)
			if err != nil {
				t.Fatalf("Fetch() error: %v", err)
			}

			for _, wantPart := range tt.wantMsgPart {
				if !strings.Contains(snap.Message, wantPart) {
					t.Errorf("snap.Message = %q, want to contain %q", snap.Message, wantPart)
				}
			}
		})
	}
}

func TestProvider_HasChanged_TasksSubdirectoryMtime(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpHome, ".config"))

	extDir := filepath.Join(tmpHome, ".config", "Code", "User", "globalStorage", roocode.KiloExtensionSubdir)
	tasksDir := filepath.Join(extDir, "tasks")
	task1Dir := filepath.Join(tasksDir, "task-1")
	if err := os.MkdirAll(task1Dir, 0o755); err != nil {
		t.Fatal(err)
	}

	p := New()
	acct := core.AccountConfig{ID: "kilo_code"}

	// Set extDir mtime to old date
	oldTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	_ = os.Chtimes(extDir, oldTime, oldTime)
	_ = os.Chtimes(tasksDir, oldTime, oldTime)

	// But set task1Dir mtime to newer time
	recentTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	_ = os.Chtimes(task1Dir, recentTime, recentTime)

	since := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	changed, err := p.HasChanged(acct, since)
	if err != nil {
		t.Fatalf("HasChanged() error: %v", err)
	}
	if !changed {
		t.Errorf("HasChanged() should detect child task modification")
	}
}

// -----------------------------------------------------------------------------
// Axis 3: Negative & Error Branches
// -----------------------------------------------------------------------------

func TestProvider_Fetch_NoData_ReturnsStatusUnknown(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	p := New()
	acct := core.AccountConfig{ID: "kilo_code", Provider: "kilo_code", Auth: "local"}

	snap, err := p.Fetch(context.Background(), acct)
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}
	if snap.Status != core.StatusUnknown {
		t.Errorf("Status = %v, want StatusUnknown", snap.Status)
	}
	if snap.Message != "Kilo Code extension data not found" {
		t.Errorf("Message = %q, want 'Kilo Code extension data not found'", snap.Message)
	}
	if len(snap.Metrics) != 0 {
		t.Errorf("Metrics = %+v, want empty map on StatusUnknown", snap.Metrics)
	}
}

func TestProvider_Fetch_ContextCancelled(t *testing.T) {
	tasksRoot := t.TempDir()
	createTestTask(t, tasksRoot, "task-1",
		`[{"say":"api_req_started","ts":1716033600000,"text":"{\"cost\":0.01,\"tokensIn\":10,\"tokensOut\":10,\"apiProtocol\":\"anthropic\"}"}]`,
		`[{"content":"<model>m</model>"}]`,
	)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel context immediately

	p := New()
	acct := core.AccountConfig{ID: "kilo_code"}
	acct.SetPath("tasks_dir", tasksRoot)

	_, err := p.Fetch(ctx, acct)
	if err == nil {
		t.Fatal("expected error on cancelled context, got nil")
	}
	if err != context.Canceled {
		t.Errorf("got error %v, want %v", err, context.Canceled)
	}
}

func TestProvider_Fetch_MalformedJSON_DiagnosticRecorded(t *testing.T) {
	tasksRoot := t.TempDir()

	// Valid task
	createTestTask(t, tasksRoot, "task-good",
		`[{"say":"api_req_started","ts":1716033600000,"text":"{\"cost\":0.02,\"tokensIn\":50,\"tokensOut\":25,\"apiProtocol\":\"anthropic\"}"}]`,
		`[{"content":"<model>good-model</model>"}]`,
	)

	// Corrupted tasks: invalid array json, non-json text
	createTestTask(t, tasksRoot, "task-bad-1", `{"not": "an array"}`, "")
	createTestTask(t, tasksRoot, "task-bad-2", `[invalid json {]`, "")

	p := New()
	acct := core.AccountConfig{ID: "kilo_code"}
	acct.SetPath("tasks_dir", tasksRoot)

	snap, err := p.Fetch(context.Background(), acct)
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}

	if snap.Status != core.StatusOK {
		t.Errorf("Status = %v, want StatusOK", snap.Status)
	}
	if got := *snap.Metrics["total_tasks"].Used; got != 1 {
		t.Errorf("total_tasks = %v, want 1 (corrupt tasks skipped)", got)
	}
	if diag, ok := snap.Diagnostics["roocode_task_parse_errors"]; !ok || diag != "2" {
		t.Errorf("Diagnostics[roocode_task_parse_errors] = %q, want '2'", diag)
	}
}

func TestProvider_Fetch_AllTasksCorrupt_NoUsageRecorded(t *testing.T) {
	tasksRoot := t.TempDir()
	createTestTask(t, tasksRoot, "task-bad-1", `[invalid json`, "")
	createTestTask(t, tasksRoot, "task-bad-2", `{bad}`, "")

	p := New()
	acct := core.AccountConfig{ID: "kilo_code"}
	acct.SetPath("tasks_dir", tasksRoot)

	snap, err := p.Fetch(context.Background(), acct)
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}

	if snap.Status != core.StatusOK {
		t.Errorf("Status = %v, want StatusOK", snap.Status)
	}
	if snap.Message != "No Kilo Code usage recorded yet" {
		t.Errorf("Message = %q, want 'No Kilo Code usage recorded yet'", snap.Message)
	}
	if diag, ok := snap.Diagnostics["roocode_task_parse_errors"]; !ok || diag != "2" {
		t.Errorf("Diagnostics[roocode_task_parse_errors] = %q, want '2'", diag)
	}
}

func TestProvider_ItemizedUsage_NoDataOrCorrupt(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpHome, ".config"))

	p := New()

	// 1. No directories present
	events, err := p.ItemizedUsage()
	if err != nil {
		t.Fatalf("ItemizedUsage() on no data error: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("ItemizedUsage() returned %d events, want 0", len(events))
	}

	// 2. Corrupt task files present
	extTasksDir := filepath.Join(tmpHome, ".config", "Code", "User", "globalStorage", roocode.KiloExtensionSubdir, "tasks")
	createTestTask(t, extTasksDir, "task-corrupt", `[invalid json`, "")

	events, err = p.ItemizedUsage()
	if err != nil {
		t.Fatalf("ItemizedUsage() on corrupt data error: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("ItemizedUsage() returned %d events on corrupt, want 0", len(events))
	}
}

// -----------------------------------------------------------------------------
// Axis 4: Concurrency & Race Conditions
// -----------------------------------------------------------------------------

func TestProvider_Fetch_ConcurrentCalls(t *testing.T) {
	tasksRoot := t.TempDir()
	createTestTask(t, tasksRoot, "task-concurrent",
		`[{"say":"api_req_started","ts":1716033600000,"text":"{\"cost\":0.01,\"tokensIn\":100,\"tokensOut\":50,\"apiProtocol\":\"anthropic\"}"}]`,
		`[{"content":"<model>concurrent-model</model>"}]`,
	)

	p := New()
	acct := core.AccountConfig{ID: "kilo_code"}
	acct.SetPath("tasks_dir", tasksRoot)

	const goroutines = 25
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			snap, err := p.Fetch(context.Background(), acct)
			if err != nil {
				t.Errorf("goroutine %d Fetch() error: %v", idx, err)
				return
			}
			if snap.Status != core.StatusOK {
				t.Errorf("goroutine %d Status = %v, want StatusOK", idx, snap.Status)
			}
		}(i)
	}

	wg.Wait()
}

func TestProvider_ItemizedUsage_ConcurrentCalls(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpHome, ".config"))

	extTasksDir := filepath.Join(tmpHome, ".config", "Code", "User", "globalStorage", roocode.KiloExtensionSubdir, "tasks")
	createTestTask(t, extTasksDir, "task-c",
		`[{"say":"api_req_started","ts":1716033600000,"text":"{\"cost\":0.01,\"tokensIn\":100,\"tokensOut\":50,\"apiProtocol\":\"anthropic\"}"}]`,
		`[{"content":"<model>c-model</model>"}]`,
	)

	p := New()

	const goroutines = 25
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			events, err := p.ItemizedUsage()
			if err != nil {
				t.Errorf("goroutine %d ItemizedUsage() error: %v", idx, err)
				return
			}
			if len(events) != 1 {
				t.Errorf("goroutine %d len(events) = %d, want 1", idx, len(events))
			}
		}(i)
	}

	wg.Wait()
}

func TestProvider_HasChanged_ConcurrentCalls(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpHome, ".config"))

	extDir := filepath.Join(tmpHome, ".config", "Code", "User", "globalStorage", roocode.KiloExtensionSubdir)
	if err := os.MkdirAll(filepath.Join(extDir, "tasks"), 0o755); err != nil {
		t.Fatal(err)
	}

	p := New()
	acct := core.AccountConfig{ID: "kilo_code"}
	since := time.Now().Add(-1 * time.Hour)

	const goroutines = 25
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			changed, err := p.HasChanged(acct, since)
			if err != nil {
				t.Errorf("goroutine %d HasChanged() error: %v", idx, err)
				return
			}
			if !changed {
				t.Errorf("goroutine %d HasChanged() = false, want true", idx)
			}
		}(i)
	}

	wg.Wait()
}

func TestProvider_Concurrent_MixedOperations(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpHome, ".config"))

	extTasksDir := filepath.Join(tmpHome, ".config", "Code", "User", "globalStorage", roocode.KiloExtensionSubdir, "tasks")
	createTestTask(t, extTasksDir, "task-mixed",
		`[{"say":"api_req_started","ts":1716033600000,"text":"{\"cost\":0.02,\"tokensIn\":200,\"tokensOut\":100,\"apiProtocol\":\"anthropic\"}"}]`,
		`[{"content":"<model>m</model>"}]`,
	)

	p := New()
	acct := core.AccountConfig{ID: "kilo_code"}

	const workers = 30
	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(idx int) {
			defer wg.Done()
			switch idx % 5 {
			case 0:
				_, _ = p.Fetch(context.Background(), acct)
			case 1:
				_, _ = p.ItemizedUsage()
			case 2:
				_, _ = p.HasChanged(acct, time.Now().Add(-10*time.Minute))
			case 3:
				_ = p.DetailWidget()
			case 4:
				_ = New()
			}
		}(i)
	}

	wg.Wait()
}

// -----------------------------------------------------------------------------
// Axis 5: Domain Invariants
// -----------------------------------------------------------------------------

func TestProvider_DomainInvariants_SnapshotStructure(t *testing.T) {
	tasksRoot := t.TempDir()
	createTestTask(t, tasksRoot, "task-1",
		`[{"say":"api_req_started","ts":1716033600000,"text":"{\"cost\":0.01,\"tokensIn\":100,\"tokensOut\":50,\"apiProtocol\":\"anthropic\"}"}]`,
		`[{"content":"<model>inv-model</model>"}]`,
	)

	now := time.Date(2026, 8, 31, 15, 30, 0, 0, time.UTC)
	p := New()
	p.clock = fixedClock{t: now}
	acct := core.AccountConfig{ID: "account-123", Provider: "kilo_code"}
	acct.SetPath("tasks_dir", tasksRoot)

	snap, err := p.Fetch(context.Background(), acct)
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}

	// Invariant 1: Provider matches canonical ID
	if snap.ProviderID != ID {
		t.Errorf("snap.ProviderID = %q, want %q", snap.ProviderID, ID)
	}

	// Invariant 2: AccountID matches request AccountConfig
	if snap.AccountID != "account-123" {
		t.Errorf("snap.AccountID = %q, want account-123", snap.AccountID)
	}

	// Invariant 3: Timestamp matches clock
	if !snap.Timestamp.Equal(now) {
		t.Errorf("snap.Timestamp = %v, want %v", snap.Timestamp, now)
	}

	// Invariant 4: Status is OK when usage recorded
	if snap.Status != core.StatusOK {
		t.Errorf("snap.Status = %v, want StatusOK", snap.Status)
	}

	// Invariant 5: DailySeries maps are initialized
	if snap.DailySeries == nil {
		t.Errorf("snap.DailySeries is nil")
	}
	for _, seriesKey := range []string{"tasks", "tokens", "cost"} {
		if _, ok := snap.DailySeries[seriesKey]; !ok {
			t.Errorf("missing DailySeries key %q", seriesKey)
		}
	}
}

func TestProvider_DomainInvariants_TokenSumIdentity(t *testing.T) {
	tasksRoot := t.TempDir()
	createTestTask(t, tasksRoot, "task-1",
		`[
			{"say":"api_req_started","ts":1716033600000,"text":"{\"cost\":0.05,\"tokensIn\":1000,\"tokensOut\":500,\"cacheReads\":200,\"cacheWrites\":100,\"apiProtocol\":\"anthropic\"}"},
			{"say":"api_req_started","ts":1716033601000,"text":"{\"cost\":0.03,\"tokensIn\":600,\"tokensOut\":300,\"cacheReads\":50,\"cacheWrites\":25,\"apiProtocol\":\"openai\"}"}
		]`,
		`[{"content":"<model>sum-model</model>"}]`,
	)

	p := New()
	acct := core.AccountConfig{ID: "kilo_code"}
	acct.SetPath("tasks_dir", tasksRoot)

	snap, err := p.Fetch(context.Background(), acct)
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}

	input := *snap.Metrics["total_input_tokens"].Used
	output := *snap.Metrics["total_output_tokens"].Used
	cacheRd := *snap.Metrics["total_cache_read_tokens"].Used
	cacheWr := *snap.Metrics["total_cache_write_tokens"].Used
	total := *snap.Metrics["total_tokens"].Used

	// Invariant: total_tokens == input + output + cacheRead + cacheWrite
	if expectedTotal := input + output + cacheRd + cacheWr; total != expectedTotal {
		t.Errorf("Token sum identity violated: total=%v, want %v (in=%v, out=%v, cacheRd=%v, cacheWr=%v)",
			total, expectedTotal, input, output, cacheRd, cacheWr)
	}

	// Invariant: ModelUsage TotalTokens matches sum
	var modelUsageTotal float64
	for _, rec := range snap.ModelUsage {
		if rec.TotalTokens != nil {
			modelUsageTotal += *rec.TotalTokens
		}
	}
	if modelUsageTotal != total {
		t.Errorf("ModelUsage total tokens sum %v != snapshot total tokens %v", modelUsageTotal, total)
	}
}

func TestProvider_DomainInvariants_ItemizedEventsMatchSnapshot(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmpHome, ".config"))

	extTasksDir := filepath.Join(tmpHome, ".config", "Code", "User", "globalStorage", roocode.KiloExtensionSubdir, "tasks")
	createTestTask(t, extTasksDir, "task-item-1",
		`[
			{"say":"api_req_started","ts":1716033600000,"text":"{\"cost\":0.05,\"tokensIn\":500,\"tokensOut\":250,\"cacheReads\":50,\"cacheWrites\":25,\"apiProtocol\":\"anthropic\"}"},
			{"say":"api_req_started","ts":1716033605000,"text":"{\"cost\":0.02,\"tokensIn\":200,\"tokensOut\":100,\"cacheReads\":10,\"cacheWrites\":5,\"apiProtocol\":\"anthropic\"}"}
		]`,
		`[{"content":"<model>match-model</model>"}]`,
	)

	p := New()
	acct := core.AccountConfig{ID: "kilo_code"}

	snap, err := p.Fetch(context.Background(), acct)
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}

	events, err := p.ItemizedUsage()
	if err != nil {
		t.Fatalf("ItemizedUsage() error: %v", err)
	}

	var totalEvInput, totalEvOutput, totalEvCacheRd, totalEvCacheWr int
	var totalEvCost float64

	for _, ev := range events {
		totalEvInput += ev.InputTokens
		totalEvOutput += ev.OutputTokens
		totalEvCacheRd += ev.CacheReadTokens
		totalEvCacheWr += ev.CacheCreationTokens
		totalEvCost += ev.CostUSD

		if ev.ProviderID != ID {
			t.Errorf("event ProviderID = %q, want %q", ev.ProviderID, ID)
		}
		if !ev.HasCost {
			t.Errorf("event HasCost = false, want true")
		}
	}

	if float64(totalEvInput) != *snap.Metrics["total_input_tokens"].Used {
		t.Errorf("Itemized input %d != Snapshot input %v", totalEvInput, *snap.Metrics["total_input_tokens"].Used)
	}
	if float64(totalEvOutput) != *snap.Metrics["total_output_tokens"].Used {
		t.Errorf("Itemized output %d != Snapshot output %v", totalEvOutput, *snap.Metrics["total_output_tokens"].Used)
	}
	if float64(totalEvCacheRd) != *snap.Metrics["total_cache_read_tokens"].Used {
		t.Errorf("Itemized cacheRead %d != Snapshot cacheRead %v", totalEvCacheRd, *snap.Metrics["total_cache_read_tokens"].Used)
	}
	if float64(totalEvCacheWr) != *snap.Metrics["total_cache_write_tokens"].Used {
		t.Errorf("Itemized cacheWrite %d != Snapshot cacheWrite %v", totalEvCacheWr, *snap.Metrics["total_cache_write_tokens"].Used)
	}
	if !floatEquals(totalEvCost, *snap.Metrics["total_cost_usd"].Used, 1e-6) {
		t.Errorf("Itemized cost %v != Snapshot cost %v", totalEvCost, *snap.Metrics["total_cost_usd"].Used)
	}
}

func TestProvider_DomainInvariants_ProviderAndAuthSpecs(t *testing.T) {
	p := New()
	spec := p.Spec()

	if spec.ID != "kilo_code" {
		t.Errorf("spec.ID = %q, want kilo_code", spec.ID)
	}
	if spec.Auth.Type != core.ProviderAuthTypeLocal {
		t.Errorf("spec.Auth.Type = %v, want Local", spec.Auth.Type)
	}
	if spec.Auth.DefaultAccountID != "kilo_code" {
		t.Errorf("spec.Auth.DefaultAccountID = %q, want kilo_code", spec.Auth.DefaultAccountID)
	}

	caps := map[string]bool{}
	for _, c := range spec.Info.Capabilities {
		caps[c] = true
	}
	for _, requiredCap := range []string{"local_stats", "session_tracking", "model_tokens", "cost_estimation"} {
		if !caps[requiredCap] {
			t.Errorf("missing capability %q", requiredCap)
		}
	}
}

