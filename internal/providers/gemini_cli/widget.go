package gemini_cli

import (
	"github.com/nurulislamz/agentusage/internal/core"
	"github.com/nurulislamz/agentusage/internal/providers/providerbase"
)

func dashboardWidget() core.DashboardWidget {
	return providerbase.CodingToolDashboard(
		providerbase.WithColorRole(core.DashboardColorRoleBlue),
		providerbase.WithGaugeMaxLines(1),
		providerbase.WithGaugePriority(
			"quota", "quota_pro", "quota_flash", "context_window", "tokens_today", "7d_tokens", "messages_today", "sessions_today",
			"client_cli_total_tokens", "client_cli_input_tokens", "cache_hit_ratio",
		),
		providerbase.WithCompactRows(
			core.DashboardCompactRow{Label: "Usage", Keys: []string{"quota", "quota_models_exhausted", "quota_models_low", "quota_models_tracked"}, MaxSegments: 4},
			core.DashboardCompactRow{Label: "Activity", Keys: []string{"messages_today", "sessions_today", "total_conversations"}, MaxSegments: 4},
			core.DashboardCompactRow{Label: "Tokens", Keys: []string{"tokens_today", "7d_tokens", "today_input_tokens", "today_output_tokens"}, MaxSegments: 4},
			core.DashboardCompactRow{Label: "Today Tok", Keys: []string{"today_cached_tokens", "today_reasoning_tokens", "today_tool_tokens"}, MaxSegments: 3},
			core.DashboardCompactRow{Label: "7d Tok", Keys: []string{"7d_input_tokens", "7d_output_tokens", "7d_cached_tokens", "7d_reasoning_tokens", "7d_tool_tokens"}, MaxSegments: 5},
			core.DashboardCompactRow{Label: "Totals", Keys: []string{"total_input_tokens", "total_output_tokens", "total_cached_tokens", "total_reasoning_tokens", "total_tool_tokens", "total_tokens"}, MaxSegments: 6},
			core.DashboardCompactRow{Label: "Efficiency", Keys: []string{"cache_efficiency", "reasoning_share", "tool_token_share", "avg_tokens_per_turn"}, MaxSegments: 4},
		),
		providerbase.WithSectionOrder(
			core.DashboardSectionHeader,
			core.DashboardSectionTopUsageProgress,
			core.DashboardSectionModelBurn,
			core.DashboardSectionClientBurn,
			core.DashboardSectionDailyUsage,
			core.DashboardSectionOtherData,
		),
		providerbase.WithHideMetricPrefixes(
			"tokens_model_", "tokens_client_",
		),
		providerbase.WithHideMetricKeys(
			"total_messages", "total_sessions", "total_turns",
			"client_cli_messages", "client_cli_turns",
		),
		providerbase.WithRawGroups(
			core.DashboardRawGroup{
				Label: "Usage Split",
				Keys:  []string{"model_usage", "client_usage"},
			},
		),
		providerbase.WithMetricLabels(map[string]string{
			"client_cli_input_tokens": "CLI Input Tokens",
			"client_cli_total_tokens": "CLI Total Tokens",
			"total_turns":             "All-Time Turns",
			"tokens_today":            "Today Tokens",
			"7d_tokens":               "7-Day Tokens",
			"quota":                   "Usage (Worst Model)",
			"quota_pro":               "Pro Usage",
			"quota_flash":             "Flash Usage",
			"quota_models_tracked":    "Tracked Usage Models",
			"quota_models_low":        "Low Usage Models",
			"quota_models_exhausted":  "Exhausted Usage Models",
			"today_input_tokens":      "Today Input Tokens",
			"today_output_tokens":     "Today Output Tokens",
			"today_cached_tokens":     "Today Cached Tokens",
			"today_reasoning_tokens":  "Today Reasoning Tokens",
			"today_tool_tokens":       "Today Tool Tokens",
			"7d_input_tokens":         "7-Day Input Tokens",
			"7d_output_tokens":        "7-Day Output Tokens",
			"7d_cached_tokens":        "7-Day Cached Tokens",
			"7d_reasoning_tokens":     "7-Day Reasoning Tokens",
			"7d_tool_tokens":          "7-Day Tool Tokens",
			"cache_efficiency":        "Cache Efficiency",
			"reasoning_share":         "Reasoning Share",
			"tool_token_share":        "Tool Token Share",
		}),
		providerbase.WithCompactLabels(map[string]string{
			"client_cli_input_tokens": "cli in",
			"client_cli_total_tokens": "cli total",
			"tokens_today":            "today tok",
			"7d_tokens":               "7d tok",
			"quota":                   "all",
			"quota_pro":               "pro",
			"quota_flash":             "flash",
			"today_input_tokens":      "in",
			"today_output_tokens":     "out",
			"today_cached_tokens":     "cached",
			"today_reasoning_tokens":  "reason",
			"today_tool_tokens":       "tools",
			"7d_input_tokens":         "in",
			"7d_output_tokens":        "out",
			"7d_cached_tokens":        "cached",
			"7d_reasoning_tokens":     "reason",
			"7d_tool_tokens":          "tools",
			"total_input_tokens":      "in",
			"total_output_tokens":     "out",
			"total_cached_tokens":     "cached",
			"total_reasoning_tokens":  "reason",
			"total_tool_tokens":       "tools",
			"total_tokens":            "all",
			"avg_tokens_per_turn":     "tok/turn",
			"cache_efficiency":        "cache %",
			"reasoning_share":         "reason %",
			"tool_token_share":        "tool %",
			"quota_models_exhausted":  "exhausted",
		}),
	)
}
