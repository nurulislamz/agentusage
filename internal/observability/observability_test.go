package observability

import (
	"context"
	"log/slog"
	"testing"
)

func TestRedactSensitive(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"api_key=sk-1234567890abcdef", "api_key=[REDACTED]"},
		{"Authorization: Bearer my-secret-token", "Authorization=[REDACTED]"},
		{"password=mysecretpassword", "password=[REDACTED]"},
		{"normal event message with count=42", "normal event message with count=42"},
	}

	for _, tt := range tests {
		got := RedactSensitive(tt.input)
		if got != tt.want {
			t.Errorf("RedactSensitive(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestResolveConfig(t *testing.T) {
	t.Setenv("AGENTUSAGE_OTEL_ENABLED", "true")
	t.Setenv("AGENTUSAGE_OTEL_ENDPOINT", "localhost:4318")
	t.Setenv("AGENTUSAGE_OTEL_INSECURE", "true")
	t.Setenv("AGENTUSAGE_OTEL_SERVICE_NAME", "test-agentusage")

	cfg := ResolveConfig(Config{})
	if !cfg.Enabled {
		t.Errorf("expected Enabled to be true")
	}
	if cfg.Endpoint != "localhost:4318" {
		t.Errorf("Endpoint = %q, want %q", cfg.Endpoint, "localhost:4318")
	}
	if !cfg.Insecure {
		t.Errorf("expected Insecure to be true")
	}
	if cfg.ServiceName != "test-agentusage" {
		t.Errorf("ServiceName = %q, want %q", cfg.ServiceName, "test-agentusage")
	}
}

func TestNoopWhenDisabled(t *testing.T) {
	_ = Init(context.Background(), Config{Enabled: false})
	if IsEnabled() {
		t.Errorf("expected IsEnabled() = false")
	}

	// Should not panic
	EmitLog(context.Background(), slog.LevelInfo, "test", "event", "message")
	_ = Flush(context.Background())
	_ = Shutdown(context.Background())
}
