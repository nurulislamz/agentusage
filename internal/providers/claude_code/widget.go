package claude_code

import (
	"github.com/nurulislamz/agentusage/internal/core"
	"github.com/nurulislamz/agentusage/internal/providers/providerbase"
)

func dashboardWidget() core.DashboardWidget {
	return providerbase.CodingToolDashboard(
		providerbase.WithColorRole(core.DashboardColorRoleLavender),
		providerbase.WithGaugePriority(
			"usage_five_hour", "usage_seven_day", "usage_seven_day_sonnet", "usage_seven_day_opus", "usage_seven_day_cowork",
			"cache_hit_ratio",
		),
		providerbase.WithCompactRows(
			core.DashboardCompactRow{Label: "Credits", Keys: []string{"today_api_cost", "5h_block_cost", "7d_api_cost", "all_time_api_cost"}, MaxSegments: 5},
			core.DashboardCompactRow{Label: "Usage", Keys: []string{"usage_five_hour", "usage_seven_day", "usage_seven_day_sonnet", "usage_seven_day_opus"}, MaxSegments: 4},
			core.DashboardCompactRow{Label: "Activity", Keys: []string{"messages_today", "sessions_today"}, MaxSegments: 4},
			core.DashboardCompactRow{Label: "Tokens", Keys: []string{"today_input_tokens", "today_output_tokens", "7d_input_tokens", "7d_output_tokens"}, MaxSegments: 4},
		),
		providerbase.WithHideMetricPrefixes(
			"tokens_today_", "input_tokens_", "output_tokens_", "provider_",
			"today_", "7d_", "all_time_", "5h_block_", "project_", "agent_",
		),
		providerbase.WithHideMetricKeys(
			"all_time_input_tokens", "all_time_output_tokens", "all_time_cache_read_tokens", "all_time_cache_create_tokens",
			"all_time_cache_create_5m_tokens", "all_time_cache_create_1h_tokens", "all_time_reasoning_tokens",
		),
		providerbase.WithMetricLabels(map[string]string{
			"today_api_cost":         "Today Cost",
			"5h_block_cost":          "5h Cost",
			"7d_api_cost":            "7-Day Cost",
			"all_time_api_cost":      "All-Time Cost",
			"usage_five_hour":        "5-Hour Usage",
			"usage_seven_day":        "7-Day Usage",
			"usage_seven_day_sonnet": "7d Sonnet Usage",
			"usage_seven_day_opus":   "7d Opus Usage",
			"usage_seven_day_cowork": "7d Team Usage",
		}),
		providerbase.WithCompactLabels(map[string]string{
			"today_api_cost":         "today",
			"5h_block_cost":          "5h",
			"7d_api_cost":            "7d",
			"all_time_api_cost":      "all",
			"usage_five_hour":        "5h",
			"usage_seven_day":        "7d",
			"usage_seven_day_sonnet": "sonnet",
			"usage_seven_day_opus":   "opus",
			"messages_today":         "msgs",
			"sessions_today":         "sess",
			"today_input_tokens":     "in",
			"today_output_tokens":    "out",
			"7d_input_tokens":        "7d in",
			"7d_output_tokens":       "7d out",
		}),
	)
}
