package shared

import (
	"testing"

	"github.com/nurulislamz/agentusage/internal/core"
)

func TestDefaultCodeStatsConfig(t *testing.T) {
	cfg := DefaultCodeStatsConfig()
	if cfg.LinesAdded != "composer_lines_added" {
		t.Errorf("LinesAdded = %q, want composer_lines_added", cfg.LinesAdded)
	}
	if cfg.LinesRemoved != "composer_lines_removed" {
		t.Errorf("LinesRemoved = %q, want composer_lines_removed", cfg.LinesRemoved)
	}
	if cfg.FilesChanged != "composer_files_changed" {
		t.Errorf("FilesChanged = %q, want composer_files_changed", cfg.FilesChanged)
	}
	if cfg.Commits != "scored_commits" {
		t.Errorf("Commits = %q, want scored_commits", cfg.Commits)
	}
	if cfg.AIPercent != "ai_code_percentage" {
		t.Errorf("AIPercent = %q, want ai_code_percentage", cfg.AIPercent)
	}
	if cfg.Prompts != "total_prompts" {
		t.Errorf("Prompts = %q, want total_prompts", cfg.Prompts)
	}
}

func TestCodingToolHidePrefixes(t *testing.T) {
	prefixes := CodingToolHidePrefixes()
	expected := []string{
		"model_", "source_", "client_", "mode_", "interface_",
		"subagent_", "lang_", "tool_",
	}
	if len(prefixes) != len(expected) {
		t.Fatalf("len(CodingToolHidePrefixes) = %d, want %d", len(prefixes), len(expected))
	}
	for i, exp := range expected {
		if prefixes[i] != exp {
			t.Errorf("prefix[%d] = %q, want %q", i, prefixes[i], exp)
		}
	}
}

func TestCodingToolSectionOrder(t *testing.T) {
	sections := CodingToolSectionOrder()
	expected := []core.DashboardStandardSection{
		core.DashboardSectionHeader,
		core.DashboardSectionTopUsageProgress,
		core.DashboardSectionModelBurn,
		core.DashboardSectionClientBurn,
		core.DashboardSectionToolUsage,
		core.DashboardSectionMCPUsage,
		core.DashboardSectionLanguageBurn,
		core.DashboardSectionCodeStats,
		core.DashboardSectionOtherData,
	}
	if len(sections) != len(expected) {
		t.Fatalf("len(CodingToolSectionOrder) = %d, want %d", len(sections), len(expected))
	}
	for i, exp := range expected {
		if sections[i] != exp {
			t.Errorf("section[%d] = %v, want %v", i, sections[i], exp)
		}
	}
}

func TestCodeStatsLabelsConsistency(t *testing.T) {
	keys := []string{
		"composer_lines_added",
		"composer_lines_removed",
		"composer_files_changed",
		"scored_commits",
		"total_prompts",
		"ai_code_percentage",
		"cache_hit_ratio",
	}

	for _, k := range keys {
		if _, ok := CodeStatsMetricLabels[k]; !ok {
			t.Errorf("missing metric label for key %q", k)
		}
		if _, ok := CodeStatsCompactLabels[k]; !ok {
			t.Errorf("missing compact label for key %q", k)
		}
	}
}
