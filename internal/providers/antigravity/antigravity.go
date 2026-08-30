// Package antigravity fetches Antigravity account quota and usage from Google's
// Cloud Code internal API and local status-line captures.
package antigravity

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/nurulislamz/agentusage/internal/core"
	"github.com/nurulislamz/agentusage/internal/providers/providerbase"
	"github.com/nurulislamz/agentusage/internal/providers/shared"
	"github.com/nurulislamz/agentusage/internal/telemetry"
)

const (
	providerID           = "antigravity"
	defaultAccountID     = "antigravity"
	defaultUsageWindow   = "session"
	quotaNearLimitRatio  = 0.15
	statusFilePathEnvVar = "AGENTUSAGE_ANTIGRAVITY_STATUS_FILE"
)

// Provider exposes Antigravity quota and usage via retrieveUserQuotaSummary and status-line integration.
type Provider struct {
	providerbase.Base
}

// New returns the Antigravity provider.
func New() *Provider {
	return &Provider{
		Base: providerbase.New(core.ProviderSpec{
			ID: providerID,
			Info: core.ProviderInfo{
				Name:         "Antigravity CLI",
				Capabilities: []string{"local_config", "oauth", "quota", "statusline", "token_usage", "by_model", "by_workspace"},
				DocURL:       "https://antigravity.google/docs/cli/reference",
			},
			Auth: core.ProviderAuthSpec{
				Type:             core.ProviderAuthTypeLocal,
				DefaultAccountID: defaultAccountID,
			},
			Setup: core.ProviderSetupSpec{
				DocsURL: "https://antigravity.google/docs/cli/reference",
				Quickstart: []string{
					"Install the Antigravity CLI (`agy`) and sign in.",
					"Run `agentusage integrations install antigravity` to connect the status line.",
					"For multi-account boxes, use `agy-box <name>` (auto-detected).",
				},
			},
			Dashboard: dashboardWidget(),
		}),
	}
}

// DetailWidget keeps the provider's detail view focused on the generic coding tool usage sections.
func (p *Provider) DetailWidget() core.DetailWidget {
	return core.CodingToolDetailWidget(false)
}

// DefaultStatusFilePath returns the path written by the installed Antigravity status-line command.
func DefaultStatusFilePath() string {
	if path := strings.TrimSpace(os.Getenv(statusFilePathEnvVar)); path != "" {
		return path
	}
	stateDir, err := telemetry.DefaultStateDir()
	if err != nil || strings.TrimSpace(stateDir) == "" {
		return ""
	}
	return filepath.Join(stateDir, "antigravity-status.json")
}

func statusFilePath(acct core.AccountConfig) string {
	if path := strings.TrimSpace(acct.Path("status_file", "")); path != "" {
		return path
	}
	box := boxName(acct)
	stateDir, _ := telemetry.DefaultStateDir()
	if box != "" && stateDir != "" {
		boxPath := filepath.Join(stateDir, fmt.Sprintf("antigravity-%s-status.json", box))
		if fileExists(boxPath) {
			return boxPath
		}
	}
	if cDir := strings.TrimSpace(acct.Path("config_dir", "")); cDir != "" {
		if fileExists(filepath.Join(cDir, "antigravity-status.json")) {
			return filepath.Join(cDir, "antigravity-status.json")
		}
		if home, _ := os.UserHomeDir(); home != "" && cDir != filepath.Join(home, ".gemini", "antigravity-cli") {
			return ""
		}
	}
	return DefaultStatusFilePath()
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// HasChanged reports whether the status file or oauth token has been modified since the given time.
func (p *Provider) HasChanged(acct core.AccountConfig, since time.Time) (bool, error) {
	path := statusFilePath(acct)
	if path != "" {
		if info, err := os.Stat(path); err == nil && info.ModTime().After(since) {
			return true, nil
		}
	}
	tokPath := tokenFilePath(acct)
	if tokPath != "" {
		if info, err := os.Stat(tokPath); err == nil && info.ModTime().After(since) {
			return true, nil
		}
	}
	return false, nil
}

// Fetch loads status-line data and polls retrieveUserQuotaSummary, projecting into a snapshot.
func (p *Provider) Fetch(ctx context.Context, acct core.AccountConfig) (core.UsageSnapshot, error) {
	snap := core.NewUsageSnapshot(p.ID(), acct.ID)
	if err := ctx.Err(); err != nil {
		return snap, err
	}

	if dir := configDir(acct); dir != "" {
		snap.Raw["config_dir"] = dir
	}
	if box := boxName(acct); box != "" {
		snap.SetAttribute("box", box)
	}

	var statusLoaded bool
	var statusPayload statusLinePayload
	sPath := statusFilePath(acct)
	if sPath != "" {
		snap.Raw["status_file"] = sPath
		if fileExists(sPath) {
			if data, err := os.ReadFile(sPath); err == nil {
				if payload, pErr := parseStatusLinePayload(data); pErr == nil {
					statusPayload = payload
					statusLoaded = true
					projectSnapshot(&snap, payload)
					snap.Raw["status_source"] = "status_file"
				}
			}
		}
	}

	accessToken, tokenPath, tokenRefreshed, err := ensureAccessToken(ctx, acct, p.Client())
	if tokenPath != "" {
		snap.Raw["oauth_token_file"] = tokenPath
	}
	if err == nil && accessToken != "" {
		if tokenRefreshed {
			snap.Raw["oauth_status"] = "refreshed"
		} else {
			snap.Raw["oauth_status"] = "valid"
		}

		baseURL := strings.TrimSpace(acct.Hint("quota_endpoint", defaultQuotaEndpoint))
		summary, apiErr := retrieveUserQuotaSummary(ctx, accessToken, baseURL, p.Client())
		if apiErr != nil && isAuthHTTPError(apiErr) {
			// Retry once after pinging the box
			if pingErr := pingBoxForToken(ctx, acct); pingErr == nil {
				if tok, tPath, _, retryErr := ensureAccessToken(ctx, acct, p.Client()); retryErr == nil {
					accessToken = tok
					if tPath != "" {
						snap.Raw["oauth_token_file"] = tPath
					}
					summary, apiErr = retrieveUserQuotaSummary(ctx, accessToken, baseURL, p.Client())
				}
			}
		}
		if apiErr == nil {
			payload := statusLinePayload{
				Quota:      quotaMapFromSummary(summary),
				ReceivedAt: time.Now().UTC(),
				Product:    "antigravity",
			}
			if statusLoaded {
				payload.Model = statusPayload.Model
				payload.ContextWindow = statusPayload.ContextWindow
				payload.Workspace = statusPayload.Workspace
				payload.SessionID = statusPayload.SessionID
				payload.ConversationID = statusPayload.ConversationID
				payload.AgentState = statusPayload.AgentState
				payload.PlanTier = statusPayload.PlanTier
				if payload.Email == "" {
					payload.Email = statusPayload.Email
				}
			}
			projectSnapshot(&snap, payload)
			snap.Raw["quota_api"] = fmt.Sprintf("ok (%d buckets)", len(payload.Quota))
			snap.Raw["quota_source"] = "retrieveUserQuotaSummary"
			return snap, nil
		}
	}

	if statusLoaded {
		if snap.Status == "" || snap.Status == core.StatusUnknown {
			snap.Status = core.StatusOK
		}
		return snap, nil
	}

	if err != nil {
		snap.Status = core.StatusAuth
		snap.Message = "Antigravity OAuth token unavailable"
		snap.SetDiagnostic("auth_error", err.Error())
		snap.SetDiagnostic("setup", "Sign in with agy / agy-box or run agentusage integrations install antigravity")
		return snap, nil
	}

	snap.Status = core.StatusError
	snap.Message = "Antigravity quota API request failed"
	return snap, nil
}

func isAuthHTTPError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "HTTP 401") || strings.Contains(msg, "HTTP 403")
}

func projectSnapshot(snap *core.UsageSnapshot, payload statusLinePayload) {
	if snap == nil {
		return
	}

	snap.Timestamp = payloadReceivedAt(payload)
	snap.Status = statusFromQuota(payload)

	if payload.Product != "" {
		snap.SetAttribute("product", payload.Product)
	}
	if payload.PlanTier != "" {
		snap.SetAttribute("plan_tier", payload.PlanTier)
	}
	email := payload.Email
	if email == "" && payload.AuthInfo != nil {
		email = payload.AuthInfo.Email
	}
	if email != "" {
		snap.SetAttribute("account_email", email)
		snap.Raw["account_email"] = email
	}
	if payload.SessionID != "" {
		snap.SetAttribute("session_id", payload.SessionID)
	}
	if payload.ConversationID != "" {
		snap.SetAttribute("conversation_id", payload.ConversationID)
	}
	if payload.Version != "" {
		snap.SetAttribute("cli_version", payload.Version)
	}
	if payload.AgentState != "" {
		snap.Raw["agent_state"] = payload.AgentState
	}
	if workspace := statusWorkspace(payload); workspace != "" {
		snap.SetAttribute("workspace", workspace)
	}

	modelID := strings.TrimSpace(payload.Model.ID)
	modelName := strings.TrimSpace(payload.Model.DisplayName)
	if modelName == "" {
		modelName = modelID
	}
	if modelID != "" {
		snap.SetAttribute("model_id", modelID)
	}
	if modelName != "" {
		snap.SetAttribute("model", modelName)
	}
	if payload.Model.ParamSummary != "" {
		snap.SetAttribute("model_param", payload.Model.ParamSummary)
	}

	projectContextMetrics(snap, payload.ContextWindow)
	projectCurrentUsageMetrics(snap, payload.ContextWindow.CurrentUsage)
	projectQuotaMetrics(snap, payload)

	if total := cumulativeTotalTokens(payload.ContextWindow); total > 0 {
		snap.Metrics["total_tokens"] = core.Metric{
			Used:   core.Float64Ptr(float64(total)),
			Unit:   "tokens",
			Window: defaultUsageWindow,
		}
		if modelName != "" {
			record := core.ModelUsageRecord{
				RawModelID:  modelName,
				RawSource:   "statusline",
				Window:      defaultUsageWindow,
				InputTokens: core.Float64Ptr(float64(payload.ContextWindow.TotalInputTokens)),
				TotalTokens: core.Float64Ptr(float64(total)),
			}
			if payload.ContextWindow.TotalOutputTokens != nil {
				record.OutputTokens = core.Float64Ptr(float64(*payload.ContextWindow.TotalOutputTokens))
			}
			record.SetDimension("workspace", statusWorkspace(payload))
			snap.AppendModelUsage(record)
		}
	}

	if snap.Message == "" {
		if modelName != "" {
			snap.Message = fmt.Sprintf("Antigravity CLI (%s)", modelName)
		} else {
			snap.Message = "Antigravity quota"
		}
	}
}

func projectContextMetrics(snap *core.UsageSnapshot, contextWindow statusLineContextWindow) {
	used, remaining, hasPercent := contextPercentages(contextWindow)
	if hasPercent {
		snap.Metrics["context_window"] = core.Metric{
			Limit:     core.Float64Ptr(100),
			Used:      core.Float64Ptr(used),
			Remaining: core.Float64Ptr(remaining),
			Unit:      "%",
			Window:    defaultUsageWindow,
		}
	}

	if contextWindow.TotalInputTokens > 0 {
		snap.Metrics["total_input_tokens"] = core.Metric{
			Used:   core.Float64Ptr(float64(contextWindow.TotalInputTokens)),
			Unit:   "tokens",
			Window: defaultUsageWindow,
		}
	}
	if contextWindow.TotalOutputTokens != nil && *contextWindow.TotalOutputTokens > 0 {
		snap.Metrics["total_output_tokens"] = core.Metric{
			Used:   core.Float64Ptr(float64(*contextWindow.TotalOutputTokens)),
			Unit:   "tokens",
			Window: defaultUsageWindow,
		}
	}
	if !hasPercent && contextWindow.ContextWindowSize != nil && *contextWindow.ContextWindowSize > 0 {
		usedTokens := cumulativeTotalTokens(contextWindow)
		if usedTokens > 0 {
			snap.Metrics["context_window"] = core.Metric{
				Limit:  core.Float64Ptr(float64(*contextWindow.ContextWindowSize)),
				Used:   core.Float64Ptr(float64(usedTokens)),
				Unit:   "tokens",
				Window: defaultUsageWindow,
			}
		}
	}
}

func projectCurrentUsageMetrics(snap *core.UsageSnapshot, usage *statusLineCurrentUsage) {
	if usage == nil {
		return
	}
	setTokenMetric := func(key string, value int64) {
		if value <= 0 {
			return
		}
		snap.Metrics[key] = core.Metric{Used: core.Float64Ptr(float64(value)), Unit: "tokens", Window: "current"}
	}
	setTokenMetric("current_input_tokens", usage.InputTokens)
	setTokenMetric("current_output_tokens", usage.OutputTokens)
	setTokenMetric("current_cache_read_tokens", usage.CacheReadTokens)
	setTokenMetric("current_cache_write_tokens", usage.CacheWriteTokensValue())
	if total := usage.TotalTokens(); total > 0 {
		setTokenMetric("current_tokens", total)
	}
}

func contextPercentages(contextWindow statusLineContextWindow) (float64, float64, bool) {
	if contextWindow.UsedPercentage == nil && contextWindow.RemainingPercentage == nil {
		return 0, 0, false
	}
	used := 0.0
	remaining := 0.0
	if contextWindow.UsedPercentage != nil {
		used = clamp(*contextWindow.UsedPercentage, 0, 100)
	}
	if contextWindow.RemainingPercentage != nil {
		remaining = clamp(*contextWindow.RemainingPercentage, 0, 100)
	}
	if contextWindow.UsedPercentage == nil {
		used = 100 - remaining
	}
	if contextWindow.RemainingPercentage == nil {
		remaining = 100 - used
	}
	return used, remaining, true
}

func cumulativeTotalTokens(contextWindow statusLineContextWindow) int64 {
	out := int64(0)
	if contextWindow.TotalOutputTokens != nil {
		out = *contextWindow.TotalOutputTokens
	}
	return contextWindow.TotalInputTokens + out
}

func statusWorkspace(payload statusLinePayload) string {
	if current := strings.TrimSpace(payload.Workspace.CurrentDir); current != "" {
		return current
	}
	if project := strings.TrimSpace(payload.Workspace.ProjectDir); project != "" {
		return project
	}
	return strings.TrimSpace(payload.CWD)
}

// EnrichSnapshots runs a live Fetch on refresh so quota is immediate instead of
// waiting for the telemetry daemon poll cadence.
func (p *Provider) EnrichSnapshots(ctx context.Context, accounts []core.AccountConfig, snaps map[string]core.UsageSnapshot) {
	if p == nil {
		return
	}
	shared.EnrichSnapshotsWithFetch(ctx, providerID, p.Fetch, accounts, snaps, nil)
}

func projectQuotaMetrics(snap *core.UsageSnapshot, payload statusLinePayload) {
	if len(payload.Quota) == 0 {
		return
	}
	keys := make([]string, 0, len(payload.Quota))
	for key := range payload.Quota {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	now := time.Now().UTC()
	receivedAt := payloadReceivedAt(payload)

	modelID := strings.TrimSpace(payload.Model.ID)
	modelName := strings.TrimSpace(payload.Model.DisplayName)
	if modelName == "" {
		modelName = modelID
	}
	modelLower := strings.ToLower(modelName)
	activePool := ""
	if strings.Contains(modelLower, "gemini") {
		activePool = "gemini"
	} else if strings.Contains(modelLower, "claude") || strings.Contains(modelLower, "sonnet") || strings.Contains(modelLower, "opus") || strings.Contains(modelLower, "3p") || strings.Contains(modelLower, "gpt") {
		activePool = "claude"
	}

	isForActivePool := func(name string) bool {
		lower := strings.ToLower(name)
		if activePool == "gemini" {
			return strings.Contains(lower, "gemini")
		}
		if activePool == "claude" {
			return strings.Contains(lower, "claude") || strings.Contains(lower, "3p") || strings.Contains(lower, "opus") || strings.Contains(lower, "sonnet")
		}
		return true
	}

	worst := 1.0
	worstName := ""
	found := false
	for _, name := range keys {
		quota := payload.Quota[name]
		if quota.RemainingFraction == nil {
			continue
		}
		remaining := clamp(*quota.RemainingFraction, 0, 1)
		cleanName := sanitizeMetricName(name)

		window := "quota"
		var period time.Duration
		if strings.Contains(cleanName, "5h") {
			window = "5h"
			period = 5 * time.Hour
		} else if strings.Contains(cleanName, "weekly") || strings.Contains(cleanName, "7d") {
			window = "7d"
			period = 7 * 24 * time.Hour
		}

		reset := quotaResetTime(quota, receivedAt)

		// If the quota reset timestamp is in the past, advance it
		if !reset.IsZero() && period > 0 && reset.Before(now) {
			for reset.Before(now) {
				reset = reset.Add(period)
			}
			remaining = 1.0
		}

		remainingPercent := remaining * 100

		metric := core.Metric{
			Limit:     core.Float64Ptr(100),
			Used:      core.Float64Ptr(100 - remainingPercent),
			Remaining: core.Float64Ptr(remainingPercent),
			Unit:      "%",
			Window:    window,
		}

		key := "quota_" + cleanName
		snap.Metrics[key] = metric

		if !reset.IsZero() {
			snap.Resets[key] = reset
			snap.Resets[key+"_reset"] = reset
		}

		// Alias 3p keys to claude keys for backward and cross-widget compatibility
		if cleanName == "3p_5h" {
			snap.Metrics["quota_claude_5h"] = metric
			if !reset.IsZero() {
				snap.Resets["quota_claude_5h"] = reset
				snap.Resets["quota_claude_5h_reset"] = reset
			}
		} else if cleanName == "3p_weekly" {
			snap.Metrics["quota_claude_weekly"] = metric
			if !reset.IsZero() {
				snap.Resets["quota_claude_weekly"] = reset
				snap.Resets["quota_claude_weekly_reset"] = reset
			}
		}

		if !quota.Disabled {
			if !found || (isForActivePool(name) && !isForActivePool(worstName)) || (isForActivePool(name) == isForActivePool(worstName) && remaining < worst) {
				worst = remaining
				worstName = name
				found = true
			}
		}
	}

	// Synthesize 5h window if only weekly was emitted.
	if _, hasGemini5h := snap.Metrics["quota_gemini_5h"]; !hasGemini5h {
		if gWk, hasGeminiWk := snap.Metrics["quota_gemini_weekly"]; hasGeminiWk && gWk.Remaining != nil {
			rem := 100.0
			if *gWk.Remaining <= 0 {
				rem = 0.0
			}
			snap.Metrics["quota_gemini_5h"] = core.Metric{
				Limit:     core.Float64Ptr(100),
				Used:      core.Float64Ptr(100 - rem),
				Remaining: core.Float64Ptr(rem),
				Unit:      "%",
				Window:    "5h",
			}
		}
	}
	if _, hasClaude5h := snap.Metrics["quota_claude_5h"]; !hasClaude5h {
		if cWk, hasClaudeWk := snap.Metrics["quota_claude_weekly"]; hasClaudeWk && cWk.Remaining != nil {
			rem := 100.0
			if *cWk.Remaining <= 0 {
				rem = 0.0
			}
			snap.Metrics["quota_claude_5h"] = core.Metric{
				Limit:     core.Float64Ptr(100),
				Used:      core.Float64Ptr(100 - rem),
				Remaining: core.Float64Ptr(rem),
				Unit:      "%",
				Window:    "5h",
			}
			snap.Metrics["quota_3p_5h"] = snap.Metrics["quota_claude_5h"]
		}
	}

	if !found {
		return
	}

	var overallRemaining float64
	geminiRem, hasGemini := getPoolRemainingFraction(payload, "gemini")
	claudeRem, hasClaude := getPoolRemainingFraction(payload, "claude", "3p", "opus", "sonnet")

	if hasGemini && hasClaude {
		if activePool == "gemini" {
			overallRemaining = geminiRem
		} else if activePool == "claude" {
			overallRemaining = claudeRem
		} else {
			overallRemaining = math.Max(geminiRem, claudeRem)
		}
	} else if hasGemini {
		overallRemaining = geminiRem
	} else if hasClaude {
		overallRemaining = claudeRem
	} else if found {
		overallRemaining = worst
	} else {
		overallRemaining = 1.0
	}

	remainingPercent := overallRemaining * 100
	snap.Metrics["quota"] = core.Metric{
		Limit:     core.Float64Ptr(100),
		Used:      core.Float64Ptr(100 - remainingPercent),
		Remaining: core.Float64Ptr(remainingPercent),
		Unit:      "%",
		Window:    "quota",
	}
	if quota, ok := payload.Quota[worstName]; ok {
		if reset := quotaResetTime(quota, payloadReceivedAt(payload)); !reset.IsZero() {
			snap.Resets["quota"] = reset
			snap.Resets["quota_reset"] = reset
		}
	}
}

func getPoolRemainingFraction(payload statusLinePayload, poolKeywords ...string) (float64, bool) {
	worst := 1.0
	found := false
	for name, quota := range payload.Quota {
		if quota.RemainingFraction == nil || quota.Disabled {
			continue
		}
		cleanName := strings.ToLower(sanitizeMetricName(name))
		matches := false
		for _, kw := range poolKeywords {
			if strings.Contains(cleanName, kw) {
				matches = true
				break
			}
		}
		if !matches {
			continue
		}
		fraction := clamp(*quota.RemainingFraction, 0, 1)
		if !found || fraction < worst {
			worst = fraction
			found = true
		}
	}
	return worst, found
}

func statusFromQuota(payload statusLinePayload) core.Status {
	// First check context window usage
	if payload.ContextWindow.UsedPercentage != nil {
		used := *payload.ContextWindow.UsedPercentage
		if used >= 100 {
			return core.StatusLimited
		}
		if used >= 85 {
			return core.StatusNearLimit
		}
	}

	model := strings.ToLower(payload.Model.DisplayName)
	if model == "" {
		model = strings.ToLower(payload.Model.ID)
	}
	if strings.Contains(model, "gemini") {
		if rem, has := getPoolRemainingFraction(payload, "gemini"); has {
			if rem <= 0 {
				return core.StatusLimited
			}
			if rem < quotaNearLimitRatio {
				return core.StatusNearLimit
			}
			return core.StatusOK
		}
	} else if strings.Contains(model, "claude") || strings.Contains(model, "sonnet") || strings.Contains(model, "opus") || strings.Contains(model, "3p") || strings.Contains(model, "gpt") {
		if rem, has := getPoolRemainingFraction(payload, "claude", "3p", "opus", "sonnet"); has {
			if rem <= 0 {
				return core.StatusLimited
			}
			if rem < quotaNearLimitRatio {
				return core.StatusNearLimit
			}
			return core.StatusOK
		}
	}

	geminiRem, hasGemini := getPoolRemainingFraction(payload, "gemini")
	claudeRem, hasClaude := getPoolRemainingFraction(payload, "claude", "3p", "opus", "sonnet")

	if hasGemini && hasClaude {
		if geminiRem <= 0 && claudeRem <= 0 {
			return core.StatusLimited
		}
		if geminiRem < quotaNearLimitRatio && claudeRem < quotaNearLimitRatio {
			return core.StatusNearLimit
		}
		return core.StatusOK
	}

	worst, ok := worstQuotaFraction(payload)
	if !ok {
		return core.StatusOK
	}
	if worst <= 0 {
		return core.StatusLimited
	}
	if worst < quotaNearLimitRatio {
		return core.StatusNearLimit
	}
	return core.StatusOK
}

func worstQuotaFraction(payload statusLinePayload) (float64, bool) {
	worst := 1.0
	found := false
	for _, quota := range payload.Quota {
		if quota.RemainingFraction == nil || quota.Disabled {
			continue
		}
		fraction := clamp(*quota.RemainingFraction, 0, 1)
		if !found || fraction < worst {
			worst = fraction
			found = true
		}
	}
	return worst, found
}

func payloadReceivedAt(payload statusLinePayload) time.Time {
	if !payload.ReceivedAt.IsZero() {
		return payload.ReceivedAt.UTC()
	}
	return time.Now().UTC()
}

func quotaResetTime(quota statusLineQuota, receivedAt time.Time) time.Time {
	if reset := strings.TrimSpace(quota.ResetTime); reset != "" {
		if parsed, err := time.Parse(time.RFC3339Nano, reset); err == nil {
			return parsed.UTC()
		}
		if parsed, err := time.Parse(time.RFC3339, reset); err == nil {
			return parsed.UTC()
		}
	}
	if quota.ResetInSeconds != nil && *quota.ResetInSeconds > 0 {
		return receivedAt.Add(time.Duration(*quota.ResetInSeconds) * time.Second)
	}
	return time.Time{}
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func sanitizeMetricName(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var b strings.Builder
	lastUnderscore := false
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastUnderscore = false
			continue
		}
		if !lastUnderscore && b.Len() > 0 {
			b.WriteByte('_')
			lastUnderscore = true
		}
	}
	return strings.Trim(b.String(), "_")
}
