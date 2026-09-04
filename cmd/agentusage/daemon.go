package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/nurulislamz/agentusage/internal/config"
	"github.com/nurulislamz/agentusage/internal/daemon"
	"github.com/nurulislamz/agentusage/internal/detect"
	"github.com/nurulislamz/agentusage/internal/integrations"
	"github.com/nurulislamz/agentusage/internal/telemetry"
	"github.com/spf13/cobra"
)

func newDaemonCommand() *cobra.Command {
	var (
		socketPath      string
		dbPath          string
		spoolDir        string
		interval        time.Duration
		collectInterval time.Duration
		pollInterval    time.Duration
		verbose         bool
	)

	runDaemon := func(_ *cobra.Command, _ []string) error {
		cfgFile, loadErr := config.Load()
		if loadErr != nil {
			log.Printf("warning: failed to load config, using defaults: %v", loadErr)
			cfgFile = config.DefaultConfig()
		}

		resolvedInterval := interval
		if resolvedInterval <= 0 {
			resolvedInterval = time.Duration(cfgFile.UI.RefreshIntervalSeconds) * time.Second
		}
		if resolvedInterval <= 0 {
			resolvedInterval = 30 * time.Second
		}

		resolvedCollect := collectInterval
		if resolvedCollect <= 0 {
			resolvedCollect = resolvedInterval
		}
		resolvedPoll := pollInterval
		if resolvedPoll <= 0 {
			resolvedPoll = resolvedInterval
		}

		// Check for actionable integrations and print advisory hints.
		detected := detect.AutoDetect()
		dirs := integrations.NewDefaultDirs()
		matches := integrations.MatchDetected(integrations.AllDefinitions(), detected, dirs)
		var actionableIDs []string
		for _, m := range matches {
			if m.Actionable {
				actionableIDs = append(actionableIDs, string(m.Definition.ID))
			}
		}
		if len(actionableIDs) > 0 {
			fmt.Fprintf(os.Stderr, "hint: detected tools with missing integrations: %s\n", strings.Join(actionableIDs, ", "))
		}

		return daemon.RunServer(daemon.Config{
			DBPath:          strings.TrimSpace(dbPath),
			SpoolDir:        strings.TrimSpace(spoolDir),
			SocketPath:      strings.TrimSpace(socketPath),
			CollectInterval: resolvedCollect,
			PollInterval:    resolvedPoll,
			Verbose:         verbose,
			Export:          cfgFile.Export,
		})
	}

	cmd := &cobra.Command{
		Use:   "daemon",
		Short: "Manage the background telemetry daemon",
		Long:  "Start the telemetry daemon. Use subcommands to install, uninstall, check status, or send hooks.",
		Example: strings.Join([]string{
			"  agentusage daemon",
			"  agentusage daemon run",
			"  agentusage daemon --verbose",
			"  agentusage daemon install",
			"  agentusage daemon status",
			"  agentusage daemon uninstall",
		}, "\n"),
		RunE: runDaemon,
	}

	defaultSocketPath, _ := telemetry.DefaultSocketPath()
	defaultDBPath, _ := telemetry.DefaultDBPath()
	defaultSpoolDir, _ := telemetry.DefaultSpoolDir()

	cmd.PersistentFlags().StringVar(&socketPath, "socket-path", defaultSocketPath, "path to daemon unix socket")
	addDaemonRunFlags(cmd, &dbPath, &spoolDir, &interval, &collectInterval, &pollInterval, &verbose, defaultDBPath, defaultSpoolDir)

	runCmd := newDaemonRunCommand(runDaemon)
	addDaemonRunFlags(runCmd, &dbPath, &spoolDir, &interval, &collectInterval, &pollInterval, &verbose, defaultDBPath, defaultSpoolDir)
	cmd.AddCommand(runCmd)
	cmd.AddCommand(newDaemonInstallCommand())
	cmd.AddCommand(newDaemonUninstallCommand())
	cmd.AddCommand(newDaemonStatusCommand())
	cmd.AddCommand(newDaemonHookCommand())

	return cmd
}

func addDaemonRunFlags(
	cmd *cobra.Command,
	dbPath *string,
	spoolDir *string,
	interval *time.Duration,
	collectInterval *time.Duration,
	pollInterval *time.Duration,
	verbose *bool,
	defaultDBPath string,
	defaultSpoolDir string,
) {
	cmd.Flags().StringVar(dbPath, "db-path", defaultDBPath, "path to telemetry sqlite database")
	cmd.Flags().StringVar(spoolDir, "spool-dir", defaultSpoolDir, "path to telemetry spool directory")
	cmd.Flags().DurationVar(interval, "interval", 0, "default collector/poller interval (0 uses config or 30s)")
	cmd.Flags().DurationVar(collectInterval, "collect-interval", 0, "collector interval override (0 uses --interval)")
	cmd.Flags().DurationVar(pollInterval, "poll-interval", 0, "provider poll interval override (0 uses --interval)")
	cmd.Flags().BoolVar(verbose, "verbose", false, "enable daemon logs")
}

func newDaemonRunCommand(runE func(cmd *cobra.Command, args []string) error) *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Run the telemetry daemon server",
		RunE:  runE,
	}
}

func newDaemonInstallCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install the telemetry daemon as a system service",
		RunE: func(cmd *cobra.Command, _ []string) error {
			socketPath, _ := cmd.Flags().GetString("socket-path")
			if err := daemon.InstallService(strings.TrimSpace(socketPath)); err != nil {
				return err
			}
			fmt.Println("telemetry daemon service installed")
			return nil
		},
	}
}

func newDaemonUninstallCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall the telemetry daemon system service",
		RunE: func(cmd *cobra.Command, _ []string) error {
			socketPath, _ := cmd.Flags().GetString("socket-path")
			if err := daemon.UninstallService(strings.TrimSpace(socketPath)); err != nil {
				return err
			}
			fmt.Println("telemetry daemon service uninstalled")
			return nil
		},
	}
}

func newDaemonStatusCommand() *cobra.Command {
	var details bool
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show telemetry daemon status",
		RunE: func(cmd *cobra.Command, _ []string) error {
			socketPath, _ := cmd.Flags().GetString("socket-path")
			return daemon.ServiceStatus(cmd.Context(), strings.TrimSpace(socketPath), details)
		},
	}
	cmd.Flags().BoolVar(&details, "details", false, "include verbose startup diagnostics")
	return cmd
}

func newDaemonHookCommand() *cobra.Command {
	var (
		socketPath string
		accountID  string
		dbPath     string
		spoolDir   string
		spoolOnly  bool
		verbose    bool
	)

	cmd := &cobra.Command{
		Use:   "hook <source> [payload]",
		Short: "Send a hook payload to the telemetry daemon via stdin or an argument",
		Long: strings.Join([]string{
			"Send a hook payload to the telemetry daemon.",
			"",
			"The payload is read from stdin by default. If a non-empty positional",
			"payload argument is provided after <source>, it is used instead of stdin.",
			"This supports tools (e.g. Codex's notify) that pass the event JSON as argv.",
		}, "\n"),
		Example: strings.Join([]string{
			"  agentusage daemon hook opencode < /tmp/opencode-hook-event.json",
			"  agentusage daemon hook codex < /tmp/codex-notify-payload.json",
			"  agentusage daemon hook claude_code < /tmp/claude-hook-payload.json",
			"  agentusage daemon hook codex '{\"type\":\"agent-turn-complete\"}'",
		}, "\n"),
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			sourceName := strings.ToLower(strings.TrimSpace(args[0]))
			if sourceName == "" {
				return fmt.Errorf("hook source name cannot be empty")
			}

			var payload []byte
			if len(args) == 2 && strings.TrimSpace(args[1]) != "" {
				payload = []byte(args[1])
			} else {
				data, err := io.ReadAll(cmd.InOrStdin())
				if err != nil {
					return fmt.Errorf("read hook payload from stdin: %w", err)
				}
				payload = data
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Second)
			defer cancel()

			var daemonErr error
			if !spoolOnly {
				client := daemon.NewClient(strings.TrimSpace(socketPath))
				result, err := client.IngestHook(ctx, sourceName, strings.TrimSpace(accountID), payload)
				if err == nil {
					if verbose {
						fmt.Printf("telemetry hook %s via daemon enqueued=%d processed=%d ingested=%d deduped=%d failed=%d\n",
							sourceName,
							result.Enqueued,
							result.Processed,
							result.Ingested,
							result.Deduped,
							result.Failed,
						)
						for _, w := range result.Warnings {
							fmt.Printf("warning: %s\n", w)
						}
					}
					return nil
				}
				daemonErr = err
			}

			result, err := daemon.IngestHookLocally(
				ctx,
				sourceName,
				strings.TrimSpace(accountID),
				payload,
				strings.TrimSpace(dbPath),
				strings.TrimSpace(spoolDir),
				spoolOnly,
			)
			if err != nil {
				if daemonErr != nil {
					return fmt.Errorf("send hook payload to telemetry daemon: %w (local fallback failed: %v)", daemonErr, err)
				}
				return fmt.Errorf("ingest hook payload locally: %w", err)
			}

			if verbose {
				if daemonErr != nil && !spoolOnly {
					fmt.Printf("telemetry hook %s via local-fallback daemon_error=%v enqueued=%d processed=%d ingested=%d deduped=%d failed=%d\n",
						sourceName,
						daemonErr,
						result.Enqueued,
						result.Processed,
						result.Ingested,
						result.Deduped,
						result.Failed,
					)
				} else {
					fmt.Printf("telemetry hook %s via local-ingest enqueued=%d processed=%d ingested=%d deduped=%d failed=%d\n",
						sourceName,
						result.Enqueued,
						result.Processed,
						result.Ingested,
						result.Deduped,
						result.Failed,
					)
				}
				for _, w := range result.Warnings {
					fmt.Printf("warning: %s\n", w)
				}
			}
			return nil
		},
	}

	defaultSocketPath, _ := telemetry.DefaultSocketPath()
	defaultDBPath, _ := telemetry.DefaultDBPath()
	defaultSpoolDir, _ := telemetry.DefaultSpoolDir()
	cmd.Flags().StringVar(&socketPath, "socket-path", defaultSocketPath, "path to daemon unix socket")
	cmd.Flags().StringVar(&accountID, "account-id", "", "optional logical account id override for ingested hook events")
	cmd.Flags().StringVar(&dbPath, "db-path", defaultDBPath, "path to telemetry sqlite database (used by local fallback)")
	cmd.Flags().StringVar(&spoolDir, "spool-dir", defaultSpoolDir, "path to telemetry spool directory (used by local fallback)")
	cmd.Flags().BoolVar(&spoolOnly, "spool-only", false, "enqueue hook payload to local spool without immediate DB ingest")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "print detailed ingest summary")

	return cmd
}
