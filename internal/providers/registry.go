package providers

import (
	"strings"

	"github.com/nurulislamz/agentusage/internal/core"
	"github.com/nurulislamz/agentusage/internal/providers/alibaba_cloud"
	"github.com/nurulislamz/agentusage/internal/providers/amp"
	"github.com/nurulislamz/agentusage/internal/providers/anthropic"
	"github.com/nurulislamz/agentusage/internal/providers/antigravity"
	"github.com/nurulislamz/agentusage/internal/providers/azure_openai"
	"github.com/nurulislamz/agentusage/internal/providers/claude_code"
	"github.com/nurulislamz/agentusage/internal/providers/codebuff"
	"github.com/nurulislamz/agentusage/internal/providers/codex"
	"github.com/nurulislamz/agentusage/internal/providers/command_code"
	"github.com/nurulislamz/agentusage/internal/providers/copilot"
	"github.com/nurulislamz/agentusage/internal/providers/crush"
	"github.com/nurulislamz/agentusage/internal/providers/cursor"
	"github.com/nurulislamz/agentusage/internal/providers/deepseek"
	"github.com/nurulislamz/agentusage/internal/providers/droid"
	"github.com/nurulislamz/agentusage/internal/providers/gemini_api"
	"github.com/nurulislamz/agentusage/internal/providers/gemini_cli"
	"github.com/nurulislamz/agentusage/internal/providers/goose"
	"github.com/nurulislamz/agentusage/internal/providers/groq"
	"github.com/nurulislamz/agentusage/internal/providers/hermes"
	"github.com/nurulislamz/agentusage/internal/providers/kilocode"
	"github.com/nurulislamz/agentusage/internal/providers/kimi_cli"
	"github.com/nurulislamz/agentusage/internal/providers/kiro"
	"github.com/nurulislamz/agentusage/internal/providers/mistral"
	"github.com/nurulislamz/agentusage/internal/providers/moonshot"
	"github.com/nurulislamz/agentusage/internal/providers/mux"
	"github.com/nurulislamz/agentusage/internal/providers/ollama"
	"github.com/nurulislamz/agentusage/internal/providers/openai"
	"github.com/nurulislamz/agentusage/internal/providers/openclaw"
	"github.com/nurulislamz/agentusage/internal/providers/opencode"
	"github.com/nurulislamz/agentusage/internal/providers/openrouter"
	"github.com/nurulislamz/agentusage/internal/providers/perplexity"
	"github.com/nurulislamz/agentusage/internal/providers/pi"
	"github.com/nurulislamz/agentusage/internal/providers/qwen_cli"
	"github.com/nurulislamz/agentusage/internal/providers/roocode"
	"github.com/nurulislamz/agentusage/internal/providers/shared"
	"github.com/nurulislamz/agentusage/internal/providers/xai"
	"github.com/nurulislamz/agentusage/internal/providers/zai"
	"github.com/nurulislamz/agentusage/internal/providers/zed"
)

func AllProviders() []core.UsageProvider {
	return []core.UsageProvider{
		openai.New(),
		anthropic.New(),
		azure_openai.New(),
		alibaba_cloud.New(),
		openrouter.New(),
		perplexity.New(),
		groq.New(),
		mistral.New(),
		moonshot.New(),
		deepseek.New(),
		xai.New(),
		zai.New(),
		opencode.New(),
		gemini_api.New(),
		gemini_cli.New(),
		antigravity.New(),
		ollama.New(),
		copilot.New(),
		cursor.New(),
		claude_code.New(),
		codex.New(),
		amp.New(),
		goose.New(),
		hermes.New(),
		mux.New(),
		droid.New(),
		crush.New(),
		roocode.New(),
		kilocode.New(),
		kiro.New(),
		zed.New(),
		codebuff.New(),
		kimi_cli.New(),
		openclaw.New(),
		pi.New(),
		qwen_cli.New(),
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
