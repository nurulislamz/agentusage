APP_NAME    := agentusage
MODULE      := github.com/nurulislamz/agentusage
VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT_HASH := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE  := $(shell date +%Y-%m-%dT%H:%M:%S%z)

BIN_DIR     := bin
CMD_DIR     := ./cmd/agentusage

# Append .exe to built binaries on Windows so they are runnable. GNU Make sets
# OS=Windows_NT on Windows (incl. Git Bash, where these recipes are run).
ifeq ($(OS),Windows_NT)
EXE         := .exe
else
EXE         :=
endif

GO          := go
GOFLAGS     :=
LDFLAGS     := -s -w \
               -X '$(MODULE)/internal/version.Version=$(VERSION)' \
               -X '$(MODULE)/internal/version.CommitHash=$(COMMIT_HASH)' \
               -X '$(MODULE)/internal/version.BuildDate=$(BUILD_DATE)'

GOLANGCI_LINT := golangci-lint

.PHONY: all
all: clean lint test build

.PHONY: help
help: ## Display this help screen
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

.PHONY: deps
deps: ## Download Go module dependencies
	$(GO) mod download
	$(GO) mod verify

.PHONY: tidy
tidy: ## Tidy Go module dependencies
	$(GO) mod tidy

.PHONY: fmt
fmt: ## Format Go source code
	$(GO) fmt ./...

.PHONY: vet
vet: ## Run go vet
	$(GO) vet ./...

.PHONY: lint
lint: ## Run linter (golangci-lint)
	@if command -v $(GOLANGCI_LINT) >/dev/null 2>&1; then \
		$(GOLANGCI_LINT) run ./...; \
	else \
		echo "Warning: $(GOLANGCI_LINT) not found, skipping."; \
	fi

.PHONY: test
test: ## Run unit tests with coverage
	$(GO) test $(GOFLAGS) -race -coverprofile=coverage.out -covermode=atomic ./...

.PHONY: verify-web-tui
verify-web-tui: ## Compare TUI detail information to the web snapshot payload
	$(GO) test $(GOFLAGS) -count=1 -run TestTUIWebInformationParity ./internal/webserve/

.PHONY: test-verbose
test-verbose: ## Run unit tests with verbose output
	$(GO) test $(GOFLAGS) -v -race ./...

.PHONY: run
run: ## Run the application locally
	$(GO) run $(CMD_DIR)

.PHONY: serve
serve: ## Run the local web dashboard (agentusage serve)
	$(GO) run $(CMD_DIR) serve $(ARGS)

# Extra words after box / box-list / box-rm become arguments, e.g.
#   make box agent-box NAME=physics
#   make box agent-box physics
ifneq ($(filter box box-list box-rm,$(MAKECMDGOALS)),)
  BOX_ARGS := $(filter-out box box-list box-rm,$(MAKECMDGOALS))
  .PHONY: $(BOX_ARGS)
  $(BOX_ARGS):
	@:
endif

.PHONY: box
box: ## Install box CLI to ~/.local/bin if needed, then create a profile: make box agent-box NAME=foo
	@./scripts/box.sh add $(BOX_ARGS) $(NAME)

.PHONY: box-list
box-list: ## List box profiles: make box-list [agent-box]
	@./scripts/box.sh list $(BOX_ARGS)

.PHONY: box-rm
box-rm: ## Remove a box profile: make box-rm agent-box NAME=foo
	@./scripts/box.sh rm $(BOX_ARGS) $(NAME)

.PHONY: test-scripts
test-scripts: ## Run shell script tests
	@./scripts/box_test.sh

.PHONY: build
build: deps ## Build the binary
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(APP_NAME)$(EXE) $(CMD_DIR)

.PHONY: install
install: build ## Install binary to ~/.local/bin and set up telemetry daemon service
	install -d $(HOME)/.local/bin
	install -m 755 $(BIN_DIR)/$(APP_NAME)$(EXE) $(HOME)/.local/bin/$(APP_NAME)$(EXE)
	@$(HOME)/.local/bin/$(APP_NAME)$(EXE) telemetry daemon install
	-@systemctl --user try-restart agentusage-serve.service >/dev/null 2>&1 || true

.PHONY: uninstall
uninstall: ## Uninstall binary from ~/.local/bin and remove telemetry daemon service
	@if [ -x "$(HOME)/.local/bin/$(APP_NAME)$(EXE)" ]; then \
		"$(HOME)/.local/bin/$(APP_NAME)$(EXE)" telemetry daemon uninstall 2>/dev/null || true; \
	fi
	rm -f $(HOME)/.local/bin/$(APP_NAME)$(EXE)



.PHONY: demo
demo: deps ## Build and run the demo with dummy data (for screenshots)
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(APP_NAME)-demo$(EXE) ./cmd/demo
	$(BIN_DIR)/$(APP_NAME)-demo$(EXE)

.PHONY: sync-tools
sync-tools: ## Regenerate all AI tool config files from canonical template
	@./scripts/sync-tool-configs.sh

.PHONY: icon-font
icon-font: ## Regenerate the provider icon font (internal/tmux/assets/openusage-icons.ttf)
	@python3 -m venv .venv-font 2>/dev/null || true
	@.venv-font/bin/pip install --quiet 'fonttools==4.63.0'
	@.venv-font/bin/python scripts/gen-icon-font.py

.PHONY: clean
clean: ## Clean build artifacts
	@rm -rf $(BIN_DIR) dist coverage.out
