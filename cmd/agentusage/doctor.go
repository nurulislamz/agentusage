package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"

	"github.com/nurulislamz/agentusage/internal/config"
	"github.com/nurulislamz/agentusage/internal/daemon"
	"github.com/nurulislamz/agentusage/internal/detect"
	"github.com/nurulislamz/agentusage/internal/integrations"
	"github.com/nurulislamz/agentusage/internal/telemetry"
	"github.com/nurulislamz/agentusage/internal/version"
)

type doctorChecker struct {
	out       io.Writer
	okCount   int
	warnCount int
	failCount int
}

func (d *doctorChecker) ok(format string, a ...any) {
	d.okCount++
	fmt.Fprintf(d.out, "[ OK ] "+format+"\n", a...)
}

func (d *doctorChecker) info(format string, a ...any) {
	fmt.Fprintf(d.out, "[INFO] "+format+"\n", a...)
}

func (d *doctorChecker) warn(format string, a ...any) {
	d.warnCount++
	fmt.Fprintf(d.out, "[WARN] "+format+"\n", a...)
}

func (d *doctorChecker) fail(format string, a ...any) {
	d.failCount++
	fmt.Fprintf(d.out, "[FAIL] "+format+"\n", a...)
}

func newDoctorCommand() *cobra.Command {
	var verbose bool
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Run comprehensive system and environment diagnostics",
		Long: `Run full health and diagnostic checks across agentUsage configuration,
telemetry daemon, database integrity, auto-detected tools, and integration hooks.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			runDoctorDiagnostics(os.Stdout, verbose)
			return nil
		},
	}
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "include verbose diagnostic details")
	return cmd
}

func runDoctorDiagnostics(out io.Writer, verbose bool) {
	d := &doctorChecker{out: out}

	fmt.Fprintln(out, "agentUsage Doctor")
	fmt.Fprintln(out, strings.Repeat("-", 40))

	checkDoctorSystem(d)
	checkDoctorConfig(d)
	checkDoctorDaemon(d, verbose)
	checkDoctorToolsAndIntegrations(d, verbose)
	checkDoctorStatuslineAndTmux(d)

	fmt.Fprintln(out)
	if d.failCount == 0 && d.warnCount == 0 {
		fmt.Fprintf(out, "Result: All systems healthy (%d checks passed).\n", d.okCount)
	} else if d.failCount == 0 {
		fmt.Fprintf(out, "Result: Healthy with %d warning(s) (%d passed).\n", d.warnCount, d.okCount)
	} else {
		fmt.Fprintf(out, "Result: Issues detected (%d failed, %d warning(s), %d passed).\n", d.failCount, d.warnCount, d.okCount)
	}
}

func checkDoctorSystem(d *doctorChecker) {
	ver := version.String()
	if strings.TrimSpace(ver) == "" {
		ver = "unknown"
	}
	d.ok("System: %s/%s (%s)", runtime.GOOS, runtime.GOARCH, ver)

	if exe, err := os.Executable(); err == nil {
		d.ok("Binary: %s", exe)
	}

	cterm := strings.ToLower(strings.TrimSpace(os.Getenv("COLORTERM")))
	term := strings.ToLower(strings.TrimSpace(os.Getenv("TERM")))
	if strings.Contains(cterm, "truecolor") || strings.Contains(cterm, "24bit") {
		d.ok("Terminal: truecolor supported (COLORTERM=%s)", cterm)
	} else if strings.Contains(term, "256") {
		d.info("Terminal: 256-color mode (TERM=%s)", term)
	} else {
		d.warn("Terminal: standard ANSI (COLORTERM unset, TERM=%s)", term)
	}
}

func checkDoctorConfig(d *doctorChecker) {
	cfgPath := config.ConfigPath()
	if fi, err := os.Stat(cfgPath); err == nil {
		cfg, loadErr := config.Load()
		if loadErr != nil {
			d.fail("Config: %s (unparseable: %v)", cfgPath, loadErr)
		} else {
			accountCount := len(cfg.Accounts)
			themeName := cfg.Theme
			if themeName == "" {
				themeName = "default"
			}
			d.ok("Config: %s (%d configured accounts, theme: %s, size: %d bytes)", cfgPath, accountCount, themeName, fi.Size())
		}
	} else if os.IsNotExist(err) {
		d.info("Config: %s (using auto-detected defaults)", cfgPath)
	} else {
		d.warn("Config: %s (%v)", cfgPath, err)
	}

	credPath := config.CredentialsPath()
	if fi, err := os.Stat(credPath); err == nil {
		mode := fi.Mode().Perm()
		if runtime.GOOS != "windows" && mode != 0o600 && mode != 0o400 {
			d.warn("Credentials permissions: %s has mode %#o (recommended: 0600)", credPath, mode)
		}
		creds, credErr := config.LoadCredentials()
		if credErr != nil {
			d.fail("Credentials: %s (unparseable: %v)", credPath, credErr)
		} else {
			keyCount := len(creds.Keys)
			sessCount := len(creds.Sessions)
			d.ok("Credentials: %s (%d API keys, %d browser sessions)", credPath, keyCount, sessCount)
		}
	} else if os.IsNotExist(err) {
		d.info("Credentials: %s (no manual keys stored)", credPath)
	} else {
		d.warn("Credentials: %s (%v)", credPath, err)
	}

	stateDir, err := telemetry.DefaultStateDir()
	if err != nil {
		d.fail("State directory: %v", err)
	} else if fi, err := os.Stat(stateDir); err == nil && fi.IsDir() {
		d.ok("State directory: %s", stateDir)
	} else {
		d.info("State directory: %s (will be created automatically)", stateDir)
	}
}

func checkDoctorDaemon(d *doctorChecker, verbose bool) {
	socketPath := daemon.ResolveSocketPath()
	if strings.TrimSpace(socketPath) == "" {
		defaultSock, _ := telemetry.DefaultSocketPath()
		socketPath = defaultSock
	}

	manager, err := daemon.NewServiceManager(socketPath)
	if err == nil && manager.IsSupported() {
		if manager.IsInstalled() {
			d.ok("Daemon Service: installed (%s)", manager.Kind)
		} else {
			d.info("Daemon Service: not installed")
		}
	}

	client := daemon.NewClient(socketPath)
	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()

	health, healthErr := client.HealthInfo(ctx)
	if healthErr == nil {
		ver := daemon.HealthVersion(health)
		d.ok("Daemon Runtime: active & healthy (version: %s, socket: %s)", ver, socketPath)
	} else {
		if manager.IsSupported() && manager.IsInstalled() {
			d.warn("Daemon Runtime: service installed but not responding (%v)", healthErr)
		} else {
			d.info("Daemon Runtime: not running (direct poll mode active)")
		}
	}

	dbPath, err := telemetry.DefaultDBPath()
	if err == nil {
		if fi, statErr := os.Stat(dbPath); statErr == nil {
			if db, openErr := sql.Open("sqlite3", fmt.Sprintf("file:%s?mode=ro", dbPath)); openErr == nil {
				var integrity string
				_ = db.QueryRow("PRAGMA integrity_check(1);").Scan(&integrity)
				_ = db.Close()
				if strings.EqualFold(integrity, "ok") {
					d.ok("Telemetry Database: %s (healthy, %s)", dbPath, formatFileSize(fi.Size()))
				} else {
					d.warn("Telemetry Database: %s (integrity check: %s)", dbPath, integrity)
				}
			} else {
				d.ok("Telemetry Database: %s (%s)", dbPath, formatFileSize(fi.Size()))
			}
		} else if os.IsNotExist(statErr) {
			d.info("Telemetry Database: %s (will be initialized on first event)", dbPath)
		}
	}
}

func checkDoctorToolsAndIntegrations(d *doctorChecker, verbose bool) {
	result := detect.AutoDetect()
	detect.ApplyCredentials(&result)

	if len(result.Tools) > 0 {
		var toolNames []string
		for _, t := range result.Tools {
			toolNames = append(toolNames, t.Name)
		}
		sort.Strings(toolNames)
		d.ok("Detected Tools (%d): %s", len(result.Tools), strings.Join(toolNames, ", "))
	} else {
		d.info("Detected Tools: none found")
	}

	if len(result.Accounts) > 0 {
		provSet := make(map[string]struct{})
		for _, a := range result.Accounts {
			provSet[a.Provider] = struct{}{}
		}
		var provList []string
		for p := range provSet {
			provList = append(provList, p)
		}
		sort.Strings(provList)
		d.ok("Active Accounts (%d across %d providers): %s", len(result.Accounts), len(provList), strings.Join(provList, ", "))
	} else {
		d.warn("Active Accounts: no provider API keys or accounts detected")
	}

	dirs := integrations.NewDefaultDirs()
	matches := integrations.MatchDetected(integrations.AllDefinitions(), result, dirs)
	var installedHooks []string
	var actionableHooks []string

	for _, m := range matches {
		if m.Status.Installed {
			installedHooks = append(installedHooks, string(m.Definition.ID))
		} else if m.Actionable {
			actionableHooks = append(actionableHooks, string(m.Definition.ID))
		}
	}

	if len(installedHooks) > 0 {
		sort.Strings(installedHooks)
		d.ok("Integration Hooks: %s", strings.Join(installedHooks, ", "))
	}
	if len(actionableHooks) > 0 {
		sort.Strings(actionableHooks)
		d.info("Recommended Hooks: %s", strings.Join(actionableHooks, ", "))
	}
}

func checkDoctorStatuslineAndTmux(d *doctorChecker) {
	home, err := os.UserHomeDir()
	if err == nil {
		claudeSettings := filepath.Join(home, ".claude", "settings.json")
		if data, err := os.ReadFile(claudeSettings); err == nil {
			if strings.Contains(string(data), "agentusage statusline") || strings.Contains(string(data), "openusage statusline") {
				d.ok("Claude Code Statusline: configured in %s", claudeSettings)
			}
		}
	}

	if _, err := exec.LookPath("tmux"); err == nil {
		if conf, err := detectTmuxConfForDoctor(); err == nil && conf != "" {
			if present, _ := checkSentinelPresent(conf); present {
				d.ok("tmux Integration: segment configured in %s", conf)
			}
		}
	}
}

func detectTmuxConfForDoctor() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	candidates := []string{
		filepath.Join(home, ".config", "tmux", "tmux.conf"),
		filepath.Join(home, ".tmux.conf"),
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c, nil
		}
	}
	return "", nil
}

func checkSentinelPresent(path string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	content := string(data)
	return strings.Contains(content, "agentusage tmux") || strings.Contains(content, "openusage tmux"), nil
}

func formatFileSize(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	} else if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	}
	return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
}
