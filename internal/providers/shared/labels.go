package shared

import "github.com/nurulislamz/agentusage/internal/core"

// CodingToolHidePrefixes returns the set of metric prefixes hidden by most coding-tool providers.
func CodingToolHidePrefixes() []string {
	return []string{
		"model_", "source_", "client_", "mode_", "interface_",
		"subagent_", "lang_", "tool_",
	}
}

// CodingToolSectionOrder returns the standard section order used by coding-tool providers.
func CodingToolSectionOrder() []core.DashboardStandardSection {
	return []core.DashboardStandardSection{
		core.DashboardSectionHeader,
		core.DashboardSectionTopUsageProgress,
		core.DashboardSectionModelBurn,
		core.DashboardSectionClientBurn,
		core.DashboardSectionOtherData,
	}
}

