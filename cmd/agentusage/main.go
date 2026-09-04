package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/nurulislamz/agentusage/internal/config"
	"github.com/nurulislamz/agentusage/internal/core"
	"github.com/nurulislamz/agentusage/internal/observability"
	"github.com/nurulislamz/agentusage/internal/version"
	"github.com/spf13/cobra"
)

func main() {
	core.EnsureUserLocalBinOnPATH()

	if core.DebugEnabled() {
		log.SetOutput(os.Stderr)
	} else {
		log.SetOutput(io.Discard)
	}

	// Initialize observability from environment variables by default
	_ = observability.Init(context.Background(), observability.ResolveConfig(observability.Config{}))
	defer func() {
		_ = observability.Flush(context.Background())
		_ = observability.Shutdown(context.Background())
	}()

	root := newRootCommand()

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func newRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:     "agentusage",
		Aliases: []string{"agu", "openusage"},
		Short:   "agentUsage is a terminal dashboard for monitoring AI coding tool usage and spend.",
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			if cfg, err := config.Load(); err == nil && cfg.Observability.Enabled {
				_ = observability.Init(context.Background(), observability.Config{
					Enabled:     cfg.Observability.Enabled,
					Endpoint:    cfg.Observability.Endpoint,
					Insecure:    cfg.Observability.Insecure,
					ServiceName: cfg.Observability.ServiceName,
					Headers:     cfg.Observability.Headers,
				})
			}
		},
		Version: version.Version,
		Run: func(_ *cobra.Command, _ []string) {
			// Loaded here rather than in main so an unreadable config only fails
			// the dashboard. Subcommands load their own config, and the ones that
			// do not need it (help) keep
			// working — including the ones you reach for to dig yourself out.
			cfg, err := config.Load()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
				fmt.Fprintf(os.Stderr, "Config path: %s\n", config.ConfigPath())
				fmt.Fprintf(os.Stderr, "Move that file aside to start from defaults.\n")
				os.Exit(1)
			}
			runDashboard(cfg)
		},
	}

	root.CompletionOptions.DisableDefaultCmd = true

	root.AddCommand(newListCommand())
	root.AddCommand(newGetCommand())
	root.AddCommand(newDetectCommand())
	root.AddCommand(newDoctorCommand())
	root.AddCommand(newServeCommand())
	root.AddCommand(newDaemonCommand())
	root.InitDefaultHelpCmd()

	return root
}
