package webserve

import (
	"time"

	"github.com/nurulislamz/agentusage/internal/core"
)

func f64(v float64) *float64 { return &v }

func demoSnapshots(now time.Time) []core.UsageSnapshot {
	return []core.UsageSnapshot{
		demoCursor(now),
		demoClaude(now),
		demoCopilot(now),
		demoGemini(now),
		demoOpenRouter(now),
		demoOllama(now),
		demoOpenCode(now),
		demoCommandCode(now),
		demoCodex(now),
	}
}

func demoSeries(now time.Time, values ...float64) []core.TimePoint {
	out := make([]core.TimePoint, 0, len(values))
	start := now.UTC().AddDate(0, 0, -(len(values) - 1))
	for i, v := range values {
		out = append(out, core.TimePoint{
			Date:  start.AddDate(0, 0, i).Format("2006-01-02"),
			Value: v,
		})
	}
	return out
}

func demoClaude(now time.Time) core.UsageSnapshot {
	snap := core.NewUsageSnapshot("claude_code", "claude-code")
	snap.Timestamp = now
	snap.Status = core.StatusOK
	snap.Message = "$9.20 / $20.00 · in 13d"
	snap.Metrics = map[string]core.Metric{
		"today_api_cost":      {Used: f64(9.20), Unit: "USD", Window: "today"},
		"7d_api_cost":         {Used: f64(187.40), Unit: "USD", Window: "7d"},
		"all_time_api_cost":   {Used: f64(912.50), Unit: "USD"},
		"burn_rate":           {Used: f64(8.40), Unit: "USD/h"},
		"today_input_tokens":  {Used: f64(1_240_000), Unit: "tokens", Window: "today"},
		"today_output_tokens": {Used: f64(318_000), Unit: "tokens", Window: "today"},
		"usage_five_hour":     {Used: f64(38), Limit: f64(100), Remaining: f64(62), Unit: "%", Window: "rolling-5h"},
		"usage_seven_day":     {Used: f64(54), Limit: f64(100), Remaining: f64(46), Unit: "%", Window: "rolling-7d"},
		"sessions_today":      {Used: f64(12), Unit: "sessions", Window: "today"},
		"messages_today":      {Used: f64(184), Unit: "messages", Window: "today"},
	}
	snap.Attributes = map[string]string{
		"plan_type":           "Max 5",
		"user_email":          "dev@acme-corp.dev",
		"last_active_at":      now.Add(-12 * time.Minute).Format(time.RFC3339),
		"recent_activity_pct": "2.5",
	}
	snap.ModelUsage = []core.ModelUsageRecord{
		{RawModelID: "claude-opus-4-6", Canonical: "claude-opus-4-6", CanonicalFamily: "claude", CostUSD: f64(31.20), InputTokens: f64(820000), OutputTokens: f64(210000), Window: "today", Confidence: 1},
		{RawModelID: "claude-sonnet-4-6", Canonical: "claude-sonnet-4-6", CanonicalFamily: "claude", CostUSD: f64(8.40), InputTokens: f64(310000), OutputTokens: f64(82000), Window: "today", Confidence: 1},
		{RawModelID: "claude-haiku-4-5", Canonical: "claude-haiku-4-5", CanonicalFamily: "claude", CostUSD: f64(2.58), InputTokens: f64(110000), OutputTokens: f64(26000), Window: "today", Confidence: 1},
	}
	snap.DailySeries = map[string][]core.TimePoint{
		"cost":     demoSeries(now, 18.2, 21.4, 16.8, 29.1, 24.6, 33.0, 42.18),
		"tokens":   demoSeries(now, 820000, 910000, 740000, 1.1e6, 980000, 1.3e6, 1.558e6),
		"requests": demoSeries(now, 92, 101, 84, 128, 114, 146, 184),
	}
	snap.Resets = map[string]time.Time{
		"plan_spend":      now.Add(13 * 24 * time.Hour),
		"usage_five_hour": now.Add(2*time.Hour + 15*time.Minute),
	}
	return snap
}

func demoCursor(now time.Time) core.UsageSnapshot {
	snap := core.NewUsageSnapshot("cursor", "cursor-ide")
	snap.Timestamp = now
	snap.Status = core.StatusOK
	snap.Message = "$3.45 / $20.00 · in 11d"
	snap.Metrics = map[string]core.Metric{
		"quota":         {Used: f64(17.25), Limit: f64(100.0), Remaining: f64(82.75), Unit: "%"},
		"plan_spend":    {Used: f64(3.45), Limit: f64(20.00), Remaining: f64(16.55), Unit: "USD"},
		"team_budget":   {Used: f64(1572.00), Limit: f64(3600.00), Remaining: f64(2028.00), Unit: "USD"},
		"billing_cycle": {Used: f64(48.7), Limit: f64(100.0), Remaining: f64(51.3), Unit: "%"},
		"today_cost":    {Used: f64(3.45), Unit: "USD", Window: "today"},
		"requests_today": {Used: f64(412), Unit: "requests", Window: "today"},
		"billing_input_tokens":  {Used: f64(597100), Unit: "tokens", Window: "month"},
		"billing_output_tokens": {Used: f64(320100), Unit: "tokens", Window: "month"},
		"code_added":    {Used: f64(139), Unit: "lines"},
		"code_removed":  {Used: f64(335), Unit: "lines"},
	}
	snap.Attributes = map[string]string{
		"plan_type":           "Pro",
		"user_email":          "demo.user@acme-corp.dev",
		"billing_cycle_dates": "Feb 11 → Mar 12",
		"billing_remaining":   "11d 17h remaining",
		"team_remaining":      "$2,028 remaining",
		"team_spent":          "$1,572 / $3,600",
		"code_meta":           "2.2k files · 3.6k commits · 65% AI-generated",
		"last_active_at":      now.Add(-34 * time.Minute).Format(time.RFC3339),
		"recent_activity_pct": "8.0",
	}
	snap.ModelUsage = []core.ModelUsageRecord{
		{RawModelID: "claude-4.5-opus", Canonical: "claude-4.5-opus", CostUSD: f64(1.18), Window: "billing", Confidence: 1},
		{RawModelID: "composer-1.5", Canonical: "composer-1.5", CostUSD: f64(0.76), Window: "billing", Confidence: 1},
		{RawModelID: "gemini-3-flash", Canonical: "gemini-3-flash", CostUSD: f64(0.41), Window: "billing", Confidence: 1},
		{RawModelID: "gpt-5.2", Canonical: "gpt-5.2", CostUSD: f64(0.32), Window: "billing", Confidence: 1},
		{RawModelID: "grok-4", Canonical: "grok-4", CostUSD: f64(0.28), Window: "billing", Confidence: 1},
	}
	snap.DailySeries = map[string][]core.TimePoint{
		"cost":     demoSeries(now, 1.2, 1.8, 1.4, 2.5, 2.1, 3.0, 3.45),
		"requests": demoSeries(now, 280, 310, 240, 390, 340, 420, 412),
	}
	snap.Resets = map[string]time.Time{
		"plan_spend": now.Add(11*24*time.Hour + 17*time.Hour),
	}
	return snap
}

func demoGemini(now time.Time) core.UsageSnapshot {
	snap := core.NewUsageSnapshot("gemini_cli", "gemini-cli")
	snap.Timestamp = now
	snap.Status = core.StatusOK
	snap.Message = "$8.50 / $20.00 · in 08d"
	snap.Metrics = map[string]core.Metric{
		"plan_spend": {Used: f64(8.50), Limit: f64(20.00), Remaining: f64(11.50), Unit: "USD"},
		"today_cost": {Used: f64(1.20), Unit: "USD", Window: "today"},
	}
	snap.Attributes = map[string]string{
		"plan_type":      "standard",
		"last_active_at": now.Add(-18 * time.Minute).Format(time.RFC3339),
	}
	snap.Resets = map[string]time.Time{
		"plan_spend": now.Add(8 * 24 * time.Hour),
	}
	return snap
}

func demoOpenRouter(now time.Time) core.UsageSnapshot {
	snap := core.NewUsageSnapshot("openrouter", "openrouter")
	snap.Timestamp = now
	snap.Status = core.StatusOK
	snap.Message = "$15.80 / $50.00 · in 04d"
	snap.Metrics = map[string]core.Metric{
		"plan_spend":          {Used: f64(15.80), Limit: f64(50), Remaining: f64(34.20), Unit: "USD", Window: "current"},
		"credit_balance":      {Used: f64(34.20), Limit: f64(50), Remaining: f64(15.80), Unit: "USD", Window: "current"},
		"today_cost":          {Used: f64(0.18), Unit: "USD", Window: "today"},
		"7d_api_cost":         {Used: f64(6.50), Unit: "USD", Window: "7d"},
		"today_requests":      {Used: f64(21), Unit: "requests", Window: "today"},
		"today_input_tokens":  {Used: f64(746000), Unit: "tokens", Window: "today"},
		"today_output_tokens": {Used: f64(89300), Unit: "tokens", Window: "today"},
	}
	snap.ModelUsage = []core.ModelUsageRecord{
		{RawModelID: "moonshotai/kimi-k2.5", Canonical: "kimi-k2.5", CostUSD: f64(3.76), Window: "activity", Confidence: 1},
		{RawModelID: "qwen/qwen3-coder-flash", Canonical: "qwen3-coder-flash", CostUSD: f64(2.44), Window: "activity", Confidence: 1},
	}
	snap.DailySeries = map[string][]core.TimePoint{
		"cost": demoSeries(now, 0.9, 1.1, 0.7, 1.4, 1.0, 1.2, 0.18),
	}
	snap.Resets = map[string]time.Time{
		"plan_spend": now.Add(4 * 24 * time.Hour),
	}
	return snap
}

func demoCopilot(now time.Time) core.UsageSnapshot {
	snap := core.NewUsageSnapshot("copilot", "copilot")
	snap.Timestamp = now
	snap.Status = core.StatusOK
	snap.Message = "$2.20 / $10.00 · in 06d"
	snap.Metrics = map[string]core.Metric{
		"plan_spend":       {Used: f64(2.20), Limit: f64(10), Remaining: f64(7.80), Unit: "USD"},
		"premium_requests": {Used: f64(186), Limit: f64(300), Remaining: f64(114), Unit: "requests", Window: "month"},
		"chat_requests":    {Used: f64(94), Unit: "requests", Window: "today"},
	}
	snap.DailySeries = map[string][]core.TimePoint{
		"requests": demoSeries(now, 70, 82, 64, 91, 88, 102, 94),
	}
	snap.Resets = map[string]time.Time{
		"plan_spend": now.Add(6 * 24 * time.Hour),
	}
	return snap
}

func demoCodex(now time.Time) core.UsageSnapshot {
	snap := core.NewUsageSnapshot("codex", "codex-cli")
	snap.Timestamp = now
	snap.Status = core.StatusOK
	snap.Message = "$11.40 today"
	snap.Metrics = map[string]core.Metric{
		"today_cost":          {Used: f64(11.40), Unit: "USD", Window: "today"},
		"7d_api_cost":         {Used: f64(48.20), Unit: "USD", Window: "7d"},
		"today_input_tokens":  {Used: f64(420000), Unit: "tokens", Window: "today"},
		"today_output_tokens": {Used: f64(96000), Unit: "tokens", Window: "today"},
	}
	snap.ModelUsage = []core.ModelUsageRecord{
		{RawModelID: "gpt-5.1-codex", Canonical: "gpt-5.1-codex", CostUSD: f64(11.40), Window: "today", Confidence: 1},
	}
	snap.DailySeries = map[string][]core.TimePoint{
		"cost": demoSeries(now, 6.2, 8.1, 5.4, 9.8, 7.6, 10.2, 11.40),
	}
	snap.Attributes = map[string]string{
		"last_active_at":      now.Add(-6 * time.Minute).Format(time.RFC3339),
		"recent_activity_pct": "18.0",
	}
	return snap
}

func demoOllama(now time.Time) core.UsageSnapshot {
	snap := core.NewUsageSnapshot("ollama", "ollama-local")
	snap.Timestamp = now
	snap.Status = core.StatusOK
	snap.Message = "$10.50 / $30.00 · in 06d"
	snap.Metrics = map[string]core.Metric{
		"plan_spend":     {Used: f64(10.50), Limit: f64(30.00), Remaining: f64(19.50), Unit: "USD"},
		"requests_today": {Used: f64(38), Unit: "requests", Window: "today"},
		"models_loaded":  {Used: f64(4), Unit: "models"},
	}
	snap.Attributes = map[string]string{"base_url": "http://127.0.0.1:11434"}
	snap.DailySeries = map[string][]core.TimePoint{
		"requests": demoSeries(now, 12, 18, 9, 22, 16, 28, 38),
	}
	snap.Resets = map[string]time.Time{
		"plan_spend": now.Add(6 * 24 * time.Hour),
	}
	return snap
}

func demoOpenCode(now time.Time) core.UsageSnapshot {
	snap := core.NewUsageSnapshot("opencode", "opencode-pro")
	snap.Timestamp = now
	snap.Status = core.StatusOK
	snap.Message = "$42.50 balance · 85.00% 5h · 72.00% weekly"
	snap.Metrics = map[string]core.Metric{
		"rolling_usage":     {Used: f64(15), Remaining: f64(85), Unit: "percent", Window: "rolling-5h"},
		"weekly_usage":      {Used: f64(28), Remaining: f64(72), Unit: "percent", Window: "7d"},
		"monthly_usage_pct": {Used: f64(40), Remaining: f64(60), Unit: "percent", Window: "month"},
		"console_balance":   {Remaining: f64(42.50), Unit: "USD", Window: "current"},
		"monthly_usage":     {Used: f64(17.50), Unit: "USD", Window: "month"},
		"monthly_limit":     {Limit: f64(60.00), Used: f64(17.50), Remaining: f64(42.50), Unit: "USD", Window: "month"},
	}
	snap.Resets = map[string]time.Time{
		"rolling_usage":     now.Add(2*time.Hour + 30*time.Minute),
		"weekly_usage":      now.Add(3*24*time.Hour + 12*time.Hour),
		"monthly_usage_pct": now.Add(18*24*time.Hour + 6*time.Hour),
	}
	snap.Attributes = map[string]string{
		"subscription_plan":   "Pro",
		"auth_scope":          "zen+console",
		"last_active_at":      now.Add(-48 * time.Minute).Format(time.RFC3339),
		"recent_activity_pct": "1.5",
	}
	return snap
}

func demoCommandCode(now time.Time) core.UsageSnapshot {
	snap := core.NewUsageSnapshot("command_code", "command-code")
	snap.Timestamp = now
	snap.Status = core.StatusOK
	snap.Message = "Command Code (GOAT) · 80.0% wk rem"
	snap.Attributes = map[string]string{
		"plan_name":         "GOAT",
		"plan_id":           "individual-goat",
		"monthly_cap":       "$70.00",
		"monthly_used":      "$35.84",
		"monthly_remaining": "$34.16",
		"weekly_cap":        "$35.00",
		"weekly_used":       "$7.00",
		"five_hour_cap":     "$14.00",
		"five_hour_used":    "$0.00",
	}
	snap.Metrics = map[string]core.Metric{
		"monthly_subscription": {Limit: f64(100), Used: f64(51.2), Remaining: f64(48.8), Unit: "percent", Window: "month"},
		"monthly_credits":      {Limit: f64(70.0), Used: f64(35.84), Remaining: f64(34.16), Unit: "USD", Window: "month"},
		"weekly_usage":         {Used: f64(20.0), Remaining: f64(80.0), Unit: "percent", Window: "7d"},
		"five_hour_usage":      {Used: f64(0.0), Remaining: f64(100.0), Unit: "percent", Window: "5h"},
		"balance":              {Remaining: f64(34.16), Unit: "USD"},
		"total_cost":           {Used: f64(35.84), Unit: "USD", Window: "billing-period"},
		"total_tokens":         {Used: f64(1420000), Unit: "tokens", Window: "billing-period"},
	}
	snap.Resets = map[string]time.Time{
		"monthly_subscription": now.Add(15 * 24 * time.Hour),
		"weekly_usage":         now.Add(3 * 24 * time.Hour),
		"five_hour_usage":      now.Add(4 * time.Hour),
	}
	return snap
}
