<p align="center">
  <img src="./assets/logo.gif" alt="agentUsage logo">
</p>

<p align="center"><strong>terminal-first local quota and usage tracking for AI coding agents, IDEs, and LLM APIs.</strong></p>

---

agentUsage is a terminal-first local dashboard for monitoring AI coding tool usage, spend, quotas, and rate limits across coding agents, IDEs, and LLM APIs. It auto-detects installed tools and API keys on your workstation with zero configuration required.

![agentUsage dashboard](./assets/dashboard.png)

## Installation

### macOS (Homebrew)

```bash
brew install nurulislamz/tap/agentusage
```

### Quick install script (Linux & macOS)

```bash
curl -fsSL https://github.com/nurulislamz/agentusage/releases/latest/download/install.sh | bash
```

### From source (Go 1.25+)

```bash
go install github.com/nurulislamz/agentusage/cmd/agentusage@latest
```

> **Note:** Requires CGO (`CGO_ENABLED=1`) for SQLite storage.

### Quick run

```bash
agentusage
```

## Docker

agentUsage provides multi-stage container images published to GitHub Container Registry (`ghcr.io/nurulislamz/agentusage` and `ghcr.io/nurulislamz/agentusage-hub`).

Images are built with Alpine 3.21, include SQLite CGO support, and run under the unprivileged `agentusage` user (`UID 1000:1000`, `HOME=/home/agentusage`).

### Running the Headless Hub Server

The hub aggregates usage snapshots from multiple worker machines:

```bash
docker run -d \
  --name agentusage-hub \
  -p 9190:9190 \
  -e AGENTUSAGE_HUB_TOKEN="your-secret-token" \
  ghcr.io/nurulislamz/agentusage:latest
```

> **Note:** When binding a public port, set `AGENTUSAGE_HUB_TOKEN` to require Bearer token authentication from worker machines.

### Running the Web Dashboard

Serve the browser-based dashboard in a container by mounting your config and telemetry state volumes:

```bash
docker run -d \
  --name agentusage-web \
  -p 8080:8080 \
  -v ~/.config/agentusage:/home/agentusage/.config/agentusage \
  -v ~/.local/state/agentusage:/home/agentusage/.local/state/agentusage \
  -e AGENTUSAGE_SERVE_TOKEN="your-secret-token" \
  ghcr.io/nurulislamz/agentusage:latest serve --listen :8080 --no-open
```

### Running the Telemetry Daemon in Docker

Run the background SQLite telemetry collector and expose its Unix domain socket or state directory:

```bash
docker run -d \
  --name agentusage-daemon \
  -v ~/.config/agentusage:/home/agentusage/.config/agentusage \
  -v ~/.local/state/agentusage:/home/agentusage/.local/state/agentusage \
  ghcr.io/nurulislamz/agentusage:latest telemetry daemon run
```

### Required Volumes and Paths

| Host Path | Container Path (`$HOME=/home/agentusage`) | Purpose |
|---|---|---|
| `~/.config/agentusage` | `/home/agentusage/.config/agentusage` | Settings (`settings.json`) and secure API credentials (`credentials.json`) |
| `~/.local/state/agentusage` | `/home/agentusage/.local/state/agentusage` | SQLite database (`telemetry.db`) and daemon socket (`daemon.sock`) |

### Building Locally

Build and run Docker images locally with `make`:

```bash
make docker-build        # Build unified agentusage image (agentusage:latest)
make docker-build-hub    # Build dedicated hub image (agentusage-hub:latest)
make docker-run-hub      # Run headless hub on port 9190
make docker-run-serve    # Run web dashboard on port 8080
make docker-run-daemon   # Run telemetry daemon in container
```

### Environment Variables

| Variable | Description |
|---|---|
| `AGENTUSAGE_HUB_TOKEN` | Bearer token for authenticating hub connections and worker pushes |
| `AGENTUSAGE_SERVE_TOKEN` | Bearer token for securing web dashboard access (`agentusage serve`) |
| `AGENTUSAGE_SERVE_BASE_PATH` | URL prefix for serving behind reverse proxies (e.g. `/agentusage`) |
| `AGENTUSAGE_DEBUG` | Set to `1` to enable debug logging |

### CGO and Runtime Architecture

- **CGO Required**: `CGO_ENABLED=1` is enabled in the builder stage with `gcc` and `musl-dev` because `mattn/go-sqlite3` requires CGO for local SQLite telemetry caching and the Cursor provider.
- **Security Hardening**: Runtime containers drop privileges to user `agentusage` (`UID 1000:1000`), include minimal runtime dependencies (`ca-certificates`, `tzdata`, `wget`, `curl`), and include automated healthcheck monitoring.

## Supported Providers

agentUsage supports 37 providers covering coding agents, IDEs, and LLM API platforms. All providers are automatically detected when available.

### Coding Agents & IDEs (24 providers)

| Provider | Detection | What it tracks |
|---|---|---|
| **Claude Code** | `claude` binary + `~/.claude` directory | Daily activity, per-model tokens, 5-hour billing block computation, burn rate, cost estimation |
| **Cursor** | `cursor` binary + local SQLite DBs (`state.vscdb`) | Plan spend and limits, per-model aggregation, Composer sessions, AI code scoring |
| **GitHub Copilot** | `gh` CLI with Copilot extension or `copilot` binary + `~/.copilot` | Chat and completions quota, org billing, session tracking |
| **Codex CLI** | `codex` binary + `~/.codex` directory | Session tokens, per-model and per-client breakdown, credits, rate limits |
| **Gemini CLI** | `gemini` binary + `~/.gemini` directory | OAuth status, conversation count, per-model tokens, quota API |
| **Antigravity CLI** | `agy` binary + `~/.gemini/antigravity-cli` (or `~/.agy-containers/<box>/...`) | Live quota buckets, session tokens, model quotas, OAuth token refresh |
| **OpenCode** | `OPENCODE_API_KEY` / `ZEN_API_KEY` or `~/.local/share/opencode/auth.json` | Credits, activity, generation stats, Zen models + balance, monthly limit, subscription |
| **Ollama** | `OLLAMA_HOST` env var or `ollama` binary | Local server models, per-model usage, optional cloud billing |
| **Amp** | `amp` binary + `~/.amp` directory | Session tokens, per-turn usage and costs |
| **Goose** | `goose` binary + `~/.config/goose` (or `~/.local/share/goose`) | Session tokens, model breakdowns, local logs |
| **Hermes Agent** | `hermes` binary + `~/.hermes` directory | Conversation logs, session tokens, cost estimation |
| **Mux** | `mux` binary + `~/.mux` directory | Multiplexed sessions, turn usage, token counts |
| **Droid** | `droid` binary + `~/.factory` directory | Droid sessions, model tokens, activity |
| **Crush** | `crush` binary + `~/.crush` directory | Charm Crush agent sessions, per-model token usage |
| **Roo Code** | Roo Code VS Code extension storage | Session tokens, model costs, workspace activity |
| **Kilo Code** | `~/.kilocode` directory or storage | Kilo Code sessions, token usage, cost tracking |
| **Kiro** | `kiro` binary + `~/.kiro` directory | Kiro sessions, model usage, turn metrics |
| **Zed** | `zed` binary + `~/.config/zed` directory | Zed assistant/agent session tokens, model usage |
| **Codebuff** | `codebuff` binary + `~/.codebuff` directory | Codebuff generation tokens, sessions, cost estimation |
| **Kimi CLI** | `kimi` binary + `~/.kimi` directory | Kimi CLI session tokens, model breakdown |
| **OpenClaw** | `openclaw` binary + `~/.openclaw` directory | OpenClaw session tokens, workspace metrics |
| **Pi** | `pi` binary + `~/.pi` directory | Pi coding agent session tokens, turn metrics |
| **Qwen CLI** | `qwen` binary + `~/.qwen` directory | Qwen CLI session tokens, model usage |
| **Command Code** | `command-code` binary + `~/.command-code` directory | Command Code session logs, token usage, cost tracking |

### API Platforms (13 providers)

| Provider | Detection | What it tracks |
|---|---|---|
| **OpenAI** | `OPENAI_API_KEY` environment variable | Rate limits via lightweight header probing |
| **Anthropic** | `ANTHROPIC_API_KEY` environment variable | Rate limits via lightweight header probing |
| **Azure OpenAI** | `AZURE_OPENAI_API_KEY` or `AZURE_API_KEY` + `AZURE_RESOURCE_NAME` (or `AZURE_OPENAI_ENDPOINT`) | Rate limits via lightweight header probing on the resource endpoint |
| **Alibaba Cloud** | `ALIBABA_CLOUD_API_KEY` or `DASHSCOPE_API_KEY` | Quotas, credits, daily usage, per-model tracking (DashScope) |
| **OpenRouter** | `OPENROUTER_API_KEY` environment variable | Credits, activity, generation stats, per-model breakdown across endpoints |
| **Perplexity** | Browser session at `console.perplexity.ai` | Tier (0–5), available balance, lifetime spend, auto-reload settings, payment method, 30-day analytics |
| **Groq** | `GROQ_API_KEY` environment variable | Rate limits, daily usage windows |
| **Mistral AI** | `MISTRAL_API_KEY` environment variable | Subscription info, usage endpoints |
| **Moonshot** | `MOONSHOT_API_KEY` environment variable | Balance breakdown (cash + voucher), org limits, tier; supports `api.moonshot.ai` and `api.moonshot.cn` |
| **DeepSeek** | `DEEPSEEK_API_KEY` environment variable | Rate limits, account balance |
| **xAI** | `XAI_API_KEY` environment variable | Rate limits, API key info |
| **Z.AI** | `ZAI_API_KEY` / `ZHIPUAI_API_KEY` or `~/.chelper/config.yaml` | Coding plan quotas, model/tool usage, daily trends, credit balance |
| **Gemini API** | `GEMINI_API_KEY` or `GOOGLE_API_KEY` | Rate limits, per-model limits |

## Development & Make Commands

```bash
make build          # Build binary to ./bin/agentusage
make run            # Run the application locally
make test           # Run unit tests with coverage
make test-verbose   # Run unit tests with verbose output
make lint           # Run linter (golangci-lint)
make vet            # Run go vet
make fmt            # Format Go source code
make deps           # Download and verify Go dependencies
make tidy           # Tidy Go module dependencies
make demo           # Build and run demo dashboard with simulated data
make serve          # Run local web dashboard
make docker-build   # Build Docker image for agentusage
make docker-run-hub # Run headless hub container
make install        # Install binary to ~/.local/bin and configure daemon service
make uninstall      # Uninstall binary from ~/.local/bin and remove daemon service
```

## License

[MIT](LICENSE)
