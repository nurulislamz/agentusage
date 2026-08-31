package tui

import "testing"

func TestPrettifyMCPName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "single character",
			input: "a",
			want:  "A",
		},
		{
			name:  "single multibyte rune",
			input: "ü",
			want:  "Ü",
		},
		{
			name:  "multibyte characters in words",
			input: "über_tool",
			want:  "Über Tool",
		},
		{
			name:  "french accented characters",
			input: "écran-mode",
			want:  "Écran Mode",
		},
		{
			name:  "dashes and underscores",
			input: "get_user-profile_data",
			want:  "Get User Profile Data",
		},
		{
			name:  "multiple spaces and symbols",
			input: "   fetch__all--records   ",
			want:  "Fetch All Records",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := prettifyMCPName(tc.input)
			if got != tc.want {
				t.Errorf("prettifyMCPName(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestPrettifyMCPFunctionName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty string returns empty",
			input: "",
			want:  "",
		},
		{
			name:  "single character function",
			input: "q",
			want:  "Q",
		},
		{
			name:  "multibyte function",
			input: "über_query",
			want:  "Über Query",
		},
		{
			name:  "snake case with uppercase input",
			input: "READ_FILE",
			want:  "Read File",
		},
		{
			name:  "kebab case",
			input: "execute-query",
			want:  "Execute Query",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := prettifyMCPFunctionName(tc.input)
			if got != tc.want {
				t.Errorf("prettifyMCPFunctionName(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestPrettifyMCPServerName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty returns unknown",
			input: "",
			want:  "unknown",
		},
		{
			name:  "whitespace returns unknown",
			input: "   ",
			want:  "unknown",
		},
		{
			name:  "claude ai prefix and mcp suffix",
			input: "claude_ai_weather_mcp",
			want:  "Weather",
		},
		{
			name:  "plugin prefix",
			input: "plugin_database_mcp",
			want:  "Database",
		},
		{
			name:  "repeated parts suffix removal",
			input: "github_tools_github",
			want:  "Github Tools",
		},
		{
			name:  "multibyte server name",
			input: "über_server",
			want:  "Über Server",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := prettifyMCPServerName(tc.input)
			if got != tc.want {
				t.Errorf("prettifyMCPServerName(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
