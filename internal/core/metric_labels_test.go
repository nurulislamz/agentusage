package core

import "testing"

func TestPrettifyMetricKey(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "override plan_percent_used",
			input: "plan_percent_used",
			want:  "Plan Used",
		},
		{
			name:  "override plan_total_spend_usd",
			input: "plan_total_spend_usd",
			want:  "Total Plan Spend",
		},
		{
			name:  "override spend_limit",
			input: "spend_limit",
			want:  "Spend Limit",
		},
		{
			name:  "override individual_spend",
			input: "individual_spend",
			want:  "Individual Spend",
		},
		{
			name:  "override context_window",
			input: "context_window",
			want:  "Context Window",
		},
		{
			name:  "single-letter segments",
			input: "a_b",
			want:  "A B",
		},
		{
			name:  "single letter single segment",
			input: "x",
			want:  "X",
		},
		{
			name:  "multibyte unicode uber_metric",
			input: "über_metric",
			want:  "Über Metric",
		},
		{
			name:  "multibyte unicode ecran_mode",
			input: "écran_mode",
			want:  "Écran Mode",
		},
		{
			name:  "acronym replacements RPM",
			input: "request_rpm",
			want:  "Request RPM",
		},
		{
			name:  "acronym replacements USD TPM RPD TPD API",
			input: "total_usd_tokens_tpm_rpd_tpd_api",
			want:  "Total USD Tokens TPM RPD TPD API",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "consecutive underscores",
			input: "foo__bar",
			want:  "Foo  Bar",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := PrettifyMetricKey(tc.input)
			if got != tc.want {
				t.Errorf("PrettifyMetricKey(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestNormalizeMetricLabel(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "5h block replacement",
			input: "5h Block",
			want:  "Usage 5h",
		},
		{
			name:  "5-Hour Usage replacement",
			input: "5-Hour Usage",
			want:  "Usage 5h",
		},
		{
			name:  "5h usage replacement",
			input: "5h Usage",
			want:  "Usage 5h",
		},
		{
			name:  "7-day usage replacement",
			input: "7-Day Usage",
			want:  "Usage 7d",
		},
		{
			name:  "7d usage replacement",
			input: "7d Usage",
			want:  "Usage 7d",
		},
		{
			name:  "trimmed whitespace",
			input: "  5h Block  ",
			want:  "Usage 5h",
		},
		{
			name:  "empty string",
			input: "   ",
			want:  "",
		},
		{
			name:  "unchanged label",
			input: "Custom Metric",
			want:  "Custom Metric",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := NormalizeMetricLabel(tc.input)
			if got != tc.want {
				t.Errorf("NormalizeMetricLabel(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestMetricLabel(t *testing.T) {
	widget := DashboardWidget{
		MetricLabelOverrides: map[string]string{
			"custom_key": "My Custom Label",
			"block_key":  "5h Block",
		},
	}

	if got := MetricLabel(widget, "custom_key"); got != "My Custom Label" {
		t.Errorf("MetricLabel override = %q, want %q", got, "My Custom Label")
	}

	if got := MetricLabel(widget, "block_key"); got != "Usage 5h" {
		t.Errorf("MetricLabel normalized override = %q, want %q", got, "Usage 5h")
	}

	if got := MetricLabel(widget, "requests_rpm"); got != "Requests RPM" {
		t.Errorf("MetricLabel fallback = %q, want %q", got, "Requests RPM")
	}
}
