package providers

import (
	"strings"

	"github.com/nurulislamz/agentusage/internal/core"
	"github.com/nurulislamz/agentusage/internal/providers/antigravity"
	"github.com/nurulislamz/agentusage/internal/providers/claude_code"
	"github.com/nurulislamz/agentusage/internal/providers/codex"
	"github.com/nurulislamz/agentusage/internal/providers/command_code"
	"github.com/nurulislamz/agentusage/internal/providers/copilot"
	"github.com/nurulislamz/agentusage/internal/providers/cursor"
	"github.com/nurulislamz/agentusage/internal/providers/deepseek"
	"github.com/nurulislamz/agentusage/internal/providers/gemini_cli"
	"github.com/nurulislamz/agentusage/internal/providers/opencode"
	"github.com/nurulislamz/agentusage/internal/providers/openrouter"
	"github.com/nurulislamz/agentusage/internal/providers/shared"
)

func AllProviders() []core.UsageProvider {
	return []core.UsageProvider{
		opencode.New(),
		gemini_cli.New(),
		antigravity.New(),
		copilot.New(),
		cursor.New(),
		deepseek.New(),
		openrouter.New(),
		claude_code.New(),
		codex.New(),
		command_code.New(),
	}
}

func TelemetrySourceBySystem(system string) (shared.TelemetrySource, bool) {
	target := strings.TrimSpace(system)
	if target == "" {
		return nil, false
	}
	for _, provider := range AllProviders() {
		source, ok := provider.(shared.TelemetrySource)
		if !ok {
			continue
		}
		if strings.EqualFold(source.System(), target) {
			return source, true
		}
	}
	return nil, false
}
