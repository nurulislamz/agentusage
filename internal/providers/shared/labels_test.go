package shared

import (
	"testing"

	"github.com/nurulislamz/agentusage/internal/core"
)

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

