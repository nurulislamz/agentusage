package providers

import (
	"strings"

	"github.com/nurulislamz/openusage/internal/core"
	"github.com/nurulislamz/openusage/internal/providers/alibaba_cloud"
	"github.com/nurulislamz/openusage/internal/providers/amp"
	"github.com/nurulislamz/openusage/internal/providers/anthropic"
	"github.com/nurulislamz/openusage/internal/providers/antigravity"
	"github.com/nurulislamz/openusage/internal/providers/azure_openai"
	"github.com/nurulislamz/openusage/internal/providers/claude_code"
	"github.com/nurulislamz/openusage/internal/providers/codebuff"
	"github.com/nurulislamz/openusage/internal/providers/codex"
	"github.com/nurulislamz/openusage/internal/providers/command_code"
	"github.com/nurulislamz/openusage/internal/providers/copilot"
	"github.com/nurulislamz/openusage/internal/providers/crush"
	"github.com/nurulislamz/openusage/internal/providers/cursor"
	"github.com/nurulislamz/openusage/internal/providers/deepseek"
	"github.com/nurulislamz/openusage/internal/providers/droid"
	"github.com/nurulislamz/openusage/internal/providers/gemini_api"
	"github.com/nurulislamz/openusage/internal/providers/gemini_cli"
	"github.com/nurulislamz/openusage/internal/providers/goose"
	"github.com/nurulislamz/openusage/internal/providers/groq"
	"github.com/nurulislamz/openusage/internal/providers/hermes"
	"github.com/nurulislamz/openusage/internal/providers/kilocode"
	"github.com/nurulislamz/openusage/internal/providers/kimi_cli"
	"github.com/nurulislamz/openusage/internal/providers/kiro"
	"github.com/nurulislamz/openusage/internal/providers/mistral"
	"github.com/nurulislamz/openusage/internal/providers/moonshot"
	"github.com/nurulislamz/openusage/internal/providers/mux"
	"github.com/nurulislamz/openusage/internal/providers/ollama"
	"github.com/nurulislamz/openusage/internal/providers/openai"
	"github.com/nurulislamz/openusage/internal/providers/openclaw"
	"github.com/nurulislamz/openusage/internal/providers/opencode"
	"github.com/nurulislamz/openusage/internal/providers/openrouter"
	"github.com/nurulislamz/openusage/internal/providers/perplexity"
	"github.com/nurulislamz/openusage/internal/providers/pi"
	"github.com/nurulislamz/openusage/internal/providers/qwen_cli"
	"github.com/nurulislamz/openusage/internal/providers/roocode"
	"github.com/nurulislamz/openusage/internal/providers/shared"
	"github.com/nurulislamz/openusage/internal/providers/xai"
	"github.com/nurulislamz/openusage/internal/providers/zai"
	"github.com/nurulislamz/openusage/internal/providers/zed"
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
