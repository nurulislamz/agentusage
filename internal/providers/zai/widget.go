package zai

import (
	"github.com/nurulislamz/agentusage/internal/core"
	"github.com/nurulislamz/agentusage/internal/providers/providerbase"
)

func dashboardWidget() core.DashboardWidget {
	cfg := providerbase.DefaultDashboard(
		providerbase.WithColorRole(core.DashboardColorRoleSapphire),
		providerbase.WithGaugeMaxLines(2),
		providerbase.WithGaugePriority(
			"spend_limit",
			"plan_spend",
			"credit_balance",
			"available_balance",
			"usage_five_hour",
			"tokens_five_hour",
			"window_cost",
			"7d_api_cost",
			"today_api_cost",
			"window_requests",
			"window_tokens",
		),
		providerbase.WithSectionOrder(
			core.DashboardSectionHeader,
			core.DashboardSectionTopUsageProgress,
			core.DashboardSectionModelBurn,
			core.DashboardSectionClientBurn,
			core.DashboardSectionUpstreamProviders,
			core.DashboardSectionDailyUsage,
			core.DashboardSectionOtherData,
		),
		providerbase.WithCompactRows(
			core.DashboardCompactRow{Label: "Credits", Keys: []string{"available_balance", "credit_balance", "credit_active_grants", "today_api_cost", "7d_api_cost"}, MaxSegments: 5},
			core.DashboardCompactRow{Label: "Usage", Keys: []string{"spend_limit", "plan_spend", "usage_five_hour", "tokens_five_hour"}, MaxSegments: 4},
			core.DashboardCompactRow{Label: "Activity", Keys: []string{"today_requests", "window_requests", "active_models", "activity_providers"}, MaxSegments: 4},
			core.DashboardCompactRow{Label: "APIs", Keys: []string{"api_models_payload_bytes", "api_quota_limit_payload_bytes", "api_model_usage_payload_bytes", "api_credits_payload_bytes"}, MaxSegments: 4},
			core.DashboardCompactRow{Label: "Tokens", Keys: []string{"today_input_tokens", "today_output_tokens", "today_reasoning_tokens", "today_tokens"}, MaxSegments: 4},
		),
		providerbase.WithHideMetricPrefixes(
			"model_", "client_", "lang_", "tool_", "provider_", "source_", "endpoint_", "interface_",
			"today_", "7d_", "window_", "api_",
		),
		providerbase.WithRawGroups(
			core.DashboardRawGroup{
				Label: "Usage Split",
				Keys: []string{
					"model_usage", "model_usage_window", "model_usage_unit",
					"client_usage", "source_usage", "provider_usage",
				},
			},
			core.DashboardRawGroup{
				Label: "Account",
				Keys: []string{
					"provider_region", "plan_type", "subscription_status", "models_count", "active_model", "auth_type", "activity_models",
					"credit_grants_count", "credit_active_grants", "credit_expiring_30d",
				},
			},
			core.DashboardRawGroup{
				Label: "Activity",
				Keys: []string{
					"activity_days", "activity_models", "activity_clients", "activity_sources", "activity_providers", "activity_endpoints",
				},
			},
			core.DashboardRawGroup{
				Label: "API Payload",
				Keys: []string{
					"api_models_parse", "api_models_numeric_top", "api_models_numeric_omitted",
					"api_quota_limit_parse", "api_quota_limit_numeric_top", "api_quota_limit_numeric_omitted",
					"api_model_usage_parse", "api_model_usage_numeric_top", "api_model_usage_numeric_omitted",
					"api_credits_parse", "api_credits_numeric_top", "api_credits_numeric_omitted",
				},
			},
			core.DashboardRawGroup{
				Label: "Usage APIs",
				Keys: []string{
					"quota_api", "model_usage_api", "credits_api",
					"quota_limit_error", "model_usage_error", "credits_error",
				},
			},
		),
		providerbase.WithMetricLabels(map[string]string{
			"usage_five_hour":               "5h Token Usage",
			"tokens_five_hour":              "5h Tokens",
			"today_api_cost":                "Today Cost",
			"7d_api_cost":                   "7-Day Cost",
			"today_requests":                "Today Requests",
			"active_models":                 "Active Models",
			"window_requests":               "Window Requests",
			"window_tokens":                 "Window Tokens",
			"window_cost":                   "Window Cost",
			"spend_limit":                   "Budget Burn",
			"plan_spend":                    "Plan Burn",
			"credit_grants_count":           "Credit Grants",
			"credit_active_grants":          "Active Grants",
			"credit_expiring_30d":           "Expiring (30d)",
			"activity_providers":            "Providers",
			"available_balance":             "Available Balance",
			"api_models_payload_bytes":      "Models Payload",
			"api_quota_limit_payload_bytes": "Quota Payload",
			"api_model_usage_payload_bytes": "Model Usage Payload",
			"api_credits_payload_bytes":     "Credits Payload",
		}),
		providerbase.WithCompactLabels(map[string]string{
			"usage_five_hour":               "5h",
			"tokens_five_hour":              "5h tok",
			"today_api_cost":                "today",
			"7d_api_cost":                   "7d",
			"today_requests":                "req",
			"window_requests":               "win req",
			"window_tokens":                 "win tok",
			"active_models":                 "models",
			"window_cost":                   "win $",
			"spend_limit":                   "budget",
			"plan_spend":                    "plan",
			"credit_active_grants":          "grants",
			"activity_providers":            "prov",
			"available_balance":             "avail",
			"api_models_payload_bytes":      "models",
			"api_quota_limit_payload_bytes": "quota",
			"api_model_usage_payload_bytes": "model api",
			"api_credits_payload_bytes":     "credits api",
		}),
	)
	cfg.DisplayStyle = core.DashboardDisplayStyleDetailedCredits
	cfg.ShowClientComposition = true
	cfg.ClientCompositionHeading = "Clients"
	cfg.HideCreditsWhenBalancePresent = true
	cfg.SuppressZeroNonUsageMetrics = true
	cfg.ClientCompositionIncludeInterfaces = true
	return cfg
}
