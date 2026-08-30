package antigravity

import (
	"bytes"
	"encoding/json"
	"time"
)

// Internal projection types shared by the API parser, status-line parser, and quota metrics.
type statusLinePayload struct {
	CWD            string                     `json:"cwd,omitempty"`
	SessionID      string                     `json:"session_id,omitempty"`
	ConversationID string                     `json:"conversation_id,omitempty"`
	TranscriptPath string                     `json:"transcript_path,omitempty"`
	Model          statusLineModel            `json:"model,omitempty"`
	Workspace      statusLineWorkspace        `json:"workspace,omitempty"`
	Version        string                     `json:"version,omitempty"`
	Product        string                     `json:"product,omitempty"`
	ContextWindow  statusLineContextWindow    `json:"context_window,omitempty"`
	Quota          map[string]statusLineQuota `json:"quota,omitempty"`
	AgentState     string                     `json:"agent_state,omitempty"`
	PlanTier       string                     `json:"plan_tier,omitempty"`
	Email          string                     `json:"email,omitempty"`
	AuthInfo       *statusLineAuthInfo        `json:"auth_info,omitempty"`
	ReceivedAt     time.Time                  `json:"received_at,omitempty"`
}

type statusLineModel struct {
	ID           string `json:"id,omitempty"`
	DisplayName  string `json:"display_name,omitempty"`
	ParamSummary string `json:"param_summary,omitempty"`
}

type statusLineWorkspace struct {
	CurrentDir string `json:"current_dir,omitempty"`
	ProjectDir string `json:"project_dir,omitempty"`
}

type statusLineContextWindow struct {
	TotalInputTokens    int64                   `json:"total_input_tokens,omitempty"`
	TotalOutputTokens   *int64                  `json:"total_output_tokens,omitempty"`
	ContextWindowSize   *int64                  `json:"context_window_size,omitempty"`
	UsedPercentage      *float64                `json:"used_percentage,omitempty"`
	RemainingPercentage *float64                `json:"remaining_percentage,omitempty"`
	CurrentUsage        *statusLineCurrentUsage `json:"current_usage,omitempty"`
}

type statusLineCurrentUsage struct {
	InputTokens         int64 `json:"input_tokens,omitempty"`
	OutputTokens        int64 `json:"output_tokens,omitempty"`
	CacheReadTokens     int64 `json:"cache_read_input_tokens,omitempty"`
	CacheWriteTokens    int64 `json:"cache_creation_input_tokens,omitempty"`
	AlternateCacheWrite int64 `json:"cache_write_tokens,omitempty"`
}

func (u *statusLineCurrentUsage) CacheWriteTokensValue() int64 {
	if u == nil {
		return 0
	}
	if u.CacheWriteTokens > 0 {
		return u.CacheWriteTokens
	}
	return u.AlternateCacheWrite
}

func (u *statusLineCurrentUsage) TotalTokens() int64 {
	if u == nil {
		return 0
	}
	return u.InputTokens + u.OutputTokens + u.CacheReadTokens + u.CacheWriteTokensValue()
}

type statusLineAuthInfo struct {
	Email       string `json:"email,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
	UserID      any    `json:"userId,omitempty"`
	AuthID      string `json:"authId,omitempty"`
}

type statusLineQuota struct {
	RemainingFraction *float64 `json:"remaining_fraction,omitempty"`
	ResetTime         string   `json:"reset_time,omitempty"`
	ResetInSeconds    *int64   `json:"reset_in_seconds,omitempty"`
	Disabled          bool     `json:"disabled,omitempty"`
}

// UnmarshalJSON accepts the documented object form and numeric fraction forms.
func (q *statusLineQuota) UnmarshalJSON(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if bytes.Equal(trimmed, []byte("null")) || len(trimmed) == 0 {
		return nil
	}
	if trimmed[0] == '{' {
		type quotaAlias statusLineQuota
		var decoded quotaAlias
		if err := json.Unmarshal(trimmed, &decoded); err != nil {
			return err
		}
		*q = statusLineQuota(decoded)
		return nil
	}
	var fraction float64
	if err := json.Unmarshal(trimmed, &fraction); err != nil {
		return err
	}
	q.RemainingFraction = &fraction
	return nil
}
