package tui

import (
	"strings"
	"testing"

	"github.com/nurulislamz/agentusage/internal/core"
)

func TestBuildDetailCodexCreditForecastSection(t *testing.T) {
	used := 2572.322
	limit := 7500.0
	rate := 200.0
	runout := 24.638
	snap := core.UsageSnapshot{
		Metrics: map[string]core.Metric{
			"codex_credit_limit": {
				Used:  &used,
				Limit: &limit,
				Unit:  "credits",
			},
			"codex_credit_burn_rate":    {Used: &rate, Unit: "credits/hour"},
			"codex_credit_runout_hours": {Used: &runout, Unit: "h"},
		},
	}

	lines := buildDetailCodexCreditForecastSection(snap, 100)
	output := strings.Join(lines, "\n")
	for _, want := range []string{"Credit Usage", "34.30%", "Credit Rate: cannot render as bar or graph", "Credit Forecast: cannot render as bar or graph"} {
		if !strings.Contains(output, want) {
			t.Errorf("expected forecast output to contain %q, got %q", want, output)
		}
	}

	// When daily series exists for burn rate, sparkline graph is rendered
	snapWithSeries := snap
	snapWithSeries.DailySeries = map[string][]core.TimePoint{
		"codex_credit_burn_rate": {
			{Value: 100},
			{Value: 150},
			{Value: 200},
		},
	}
	linesWithSeries := buildDetailCodexCreditForecastSection(snapWithSeries, 100)
	outputWithSeries := strings.Join(linesWithSeries, "\n")
	if strings.Contains(outputWithSeries, "Credit Rate: cannot render as bar or graph") {
		t.Errorf("expected sparkline graph for burn rate with series, but got error banner in %q", outputWithSeries)
	}
	if !strings.Contains(outputWithSeries, "Credit Rate") {
		t.Errorf("expected Credit Rate line in %q", outputWithSeries)
	}
}
