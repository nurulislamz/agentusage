package daemon

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/mod/semver"

	"github.com/nurulislamz/agentusage/internal/telemetry"
	"github.com/nurulislamz/agentusage/internal/version"
)

func ClassifyEnsureError(err error) DaemonState {
	if err == nil {
		return DaemonState{Status: DaemonStatusRunning}
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, "not installed"):
		return DaemonState{
			Status:      DaemonStatusNotInstalled,
			Message:     "Background helper is not set up.",
			InstallHint: "agentusage daemon install",
		}
	case strings.Contains(msg, "out of date"):
		return DaemonState{
			Status:  DaemonStatusOutdated,
			Message: msg,
		}
	case strings.Contains(msg, "unsupported on"):
		return DaemonState{
			Status:  DaemonStatusError,
			Message: msg,
		}
	default:
		return DaemonState{
			Status:  DaemonStatusError,
			Message: msg,
		}
	}
}

type serviceLifecycleManager interface {
	IsSupported() bool
	IsInstalled() bool
	Start() error
	InstallHint() string
}

var (
	serviceManagerFactory = func(socketPath string) (serviceLifecycleManager, error) {
		return NewServiceManager(socketPath)
	}
	spawnDaemonFunc = spawnDaemonProcess
)

func EnsureRunning(ctx context.Context, socketPath string, verbose bool) (*Client, error) {
	socketPath = strings.TrimSpace(socketPath)
	if socketPath == "" {
		return nil, fmt.Errorf("daemon socket path is empty")
	}
	client := NewClient(socketPath)

	// 1. Prefer healthy helper: if already running and current, use it directly.
	health, healthErr := WaitForHealthInfo(ctx, client, 1200*time.Millisecond)
	if healthErr == nil {
		if HealthCurrent(health) {
			return client, nil
		}
		// Running but outdated -> fail clearly on version incompatibility without auto-installing service.
		manager, _ := serviceManagerFactory(socketPath)
		hint := "agentusage daemon install"
		if manager != nil && manager.InstallHint() != "" {
			hint = manager.InstallHint()
		}
		return nil, fmt.Errorf(
			"telemetry daemon is out of date (running=%s expected=%s); run `%s` to upgrade",
			HealthVersion(health), strings.TrimSpace(version.Version), hint,
		)
	}

	// 2. Installed managed service: if supported and installed, start it.
	manager, err := serviceManagerFactory(socketPath)
	if err == nil && manager.IsSupported() && manager.IsInstalled() {
		return startViaManagedService(ctx, client, manager, false, socketPath)
	}

	// 3. Ordinary helper spawn: spawn background helper process.
	if err := spawnDaemonFunc(socketPath, verbose); err != nil {
		return nil, fmt.Errorf("start telemetry daemon: %w", err)
	}
	if err := waitAndVerifyDaemon(ctx, client, socketPath); err != nil {
		return nil, err
	}
	return client, nil
}

func ensureViaServiceManager(
	ctx context.Context,
	client *Client,
	socketPath string,
	verbose bool,
	needsUpgrade bool,
	health HealthResponse,
) (*Client, error) {
	manager, err := serviceManagerFactory(socketPath)
	if err != nil {
		return nil, err
	}

	if needsUpgrade {
		return nil, fmt.Errorf(
			"telemetry daemon is out of date (running=%s expected=%s); run `%s` to upgrade",
			HealthVersion(health), strings.TrimSpace(version.Version), manager.InstallHint(),
		)
	}

	if manager.IsSupported() && manager.IsInstalled() {
		return startViaManagedService(ctx, client, manager, needsUpgrade, socketPath)
	}

	if err := spawnDaemonFunc(socketPath, verbose); err != nil {
		if waitErr := waitAndVerifyDaemon(ctx, client, socketPath); waitErr == nil {
			return client, nil
		}
		return nil, fmt.Errorf("start telemetry daemon: %w", err)
	}
	if err := waitAndVerifyDaemon(ctx, client, socketPath); err != nil {
		return nil, err
	}
	return client, nil
}

func startViaManagedService(
	ctx context.Context,
	client *Client,
	manager serviceLifecycleManager,
	needsUpgrade bool,
	socketPath string,
) (*Client, error) {
	if needsUpgrade {
		return nil, fmt.Errorf("telemetry daemon is out of date; run `%s` to upgrade", manager.InstallHint())
	}
	if !manager.IsInstalled() {
		return nil, fmt.Errorf("telemetry daemon service is not installed; run `%s`", manager.InstallHint())
	}
	if err := manager.Start(); err != nil {
		// If start returned an ambiguous manager-level error, still check whether
		// a daemon reached health on the socket before failing hard.
		if waitErr := waitAndVerifyDaemon(ctx, client, socketPath); waitErr == nil {
			return client, nil
		}
		if sm, ok := manager.(ServiceManager); ok {
			return nil, fmt.Errorf("start telemetry daemon service: %w\n%s", err, StartupDiagnostics(sm, socketPath))
		}
		return nil, fmt.Errorf("start telemetry daemon service: %w", err)
	}
	if err := waitAndVerifyDaemon(ctx, client, socketPath); err != nil {
		if sm, ok := manager.(ServiceManager); ok {
			return nil, fmt.Errorf("%w\n%s", err, StartupDiagnostics(sm, socketPath))
		}
		return nil, err
	}
	return client, nil
}

func waitAndVerifyDaemon(ctx context.Context, client *Client, socketPath string) error {
	if err := WaitForHealth(ctx, client, 25*time.Second); err != nil {
		return err
	}
	health, err := WaitForHealthInfo(ctx, client, 1500*time.Millisecond)
	if err != nil {
		return nil
	}
	if !HealthCurrent(health) {
		return fmt.Errorf(
			"telemetry daemon is out of date (running=%s expected=%s)",
			HealthVersion(health), strings.TrimSpace(version.Version),
		)
	}
	return nil
}

func HealthVersion(health HealthResponse) string {
	if v := strings.TrimSpace(health.DaemonVersion); v != "" {
		return v
	}
	return "unknown"
}

func HealthCurrent(health HealthResponse) bool {
	expected := strings.TrimSpace(version.Version)
	if expected == "" || strings.EqualFold(expected, "dev") || !IsReleaseSemver(expected) {
		return HealthAPICompatible(health) && HealthProviderRegistryCompatible(health)
	}
	return strings.TrimSpace(health.DaemonVersion) == expected &&
		HealthAPICompatible(health) &&
		HealthProviderRegistryCompatible(health)
}

func HealthAPICompatible(health HealthResponse) bool {
	apiVersion := strings.TrimSpace(health.APIVersion)
	return apiVersion == "" || apiVersion == APIVersion
}

func HealthProviderRegistryCompatible(health HealthResponse) bool {
	expected := ProviderRegistryHash()
	if expected == "" {
		return true
	}
	got := strings.TrimSpace(health.ProviderRegistry)
	if got == "" {
		current := strings.TrimSpace(version.Version)
		// Backward-compatible for local/dev snapshots so `go run` workflows don't
		// force service reinstalls against transient executable paths.
		if current == "" || strings.EqualFold(current, "dev") || !IsReleaseSemver(current) {
			return true
		}
		return false
	}
	return strings.EqualFold(got, expected)
}

func IsReleaseSemver(value string) bool {
	v := strings.TrimSpace(value)
	if !semver.IsValid(v) {
		return false
	}
	if semver.Prerelease(v) != "" || semver.Build(v) != "" {
		return false
	}
	return v == semver.Canonical(v)
}

func WaitForHealth(ctx context.Context, client *Client, timeout time.Duration) error {
	_, err := WaitForHealthInfo(ctx, client, timeout)
	return err
}

func WaitForHealthInfo(
	ctx context.Context,
	client *Client,
	timeout time.Duration,
) (HealthResponse, error) {
	if client == nil {
		return HealthResponse{}, fmt.Errorf("daemon client is nil")
	}
	pingCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	deadline := time.Now().Add(timeout)
	var lastErr error
	for time.Now().Before(deadline) {
		if pingCtx.Err() != nil {
			break
		}
		hc, hcCancel := context.WithTimeout(pingCtx, 700*time.Millisecond)
		health, err := client.HealthInfo(hc)
		hcCancel()
		if err == nil {
			return health, nil
		}
		lastErr = err
		time.Sleep(220 * time.Millisecond)
	}
	if pingCtx.Err() != nil && pingCtx.Err() != context.Canceled {
		return HealthResponse{}, pingCtx.Err()
	}
	if lastErr != nil {
		return HealthResponse{}, fmt.Errorf("telemetry daemon did not become ready at %s: %w", client.SocketPath, lastErr)
	}
	return HealthResponse{}, fmt.Errorf("telemetry daemon did not become ready at %s", client.SocketPath)
}

func StartupDiagnostics(manager ServiceManager, socketPath string) string {
	lines := []string{
		fmt.Sprintf("manager_kind=%s", strings.TrimSpace(manager.Kind)),
		fmt.Sprintf("service_supported=%t", manager.IsSupported()),
		fmt.Sprintf("service_installed=%t", manager.IsInstalled()),
		fmt.Sprintf("socket_path=%s", strings.TrimSpace(socketPath)),
		fmt.Sprintf("unit_path=%s", strings.TrimSpace(manager.unitPath)),
		fmt.Sprintf("service_executable=%s", strings.TrimSpace(manager.exePath)),
		fmt.Sprintf("service_executable_transient=%t", isTransientExecutablePath(manager.exePath)),
	}
	if _, err := os.Stat(strings.TrimSpace(manager.exePath)); err == nil {
		lines = append(lines, "service_executable_exists=true")
	} else {
		lines = append(lines, fmt.Sprintf("service_executable_exists=false err=%v", err))
	}
	if hint := strings.TrimSpace(manager.StatusHint()); hint != "" {
		lines = append(lines, "status_cmd="+hint)
	}
	if owner := SocketOwnerSummary(socketPath); strings.TrimSpace(owner) != "" {
		lines = append(lines, "socket_owner:\n"+owner)
	}
	if stderrPath := strings.TrimSpace(manager.StderrLogPath()); stderrPath != "" {
		lines = append(lines, "stderr_log="+stderrPath)
		if tail := TailFile(stderrPath, 30); strings.TrimSpace(tail) != "" {
			lines = append(lines, "stderr_tail:\n"+tail)
		}
	}
	if stdoutPath := strings.TrimSpace(manager.StdoutLogPath()); stdoutPath != "" {
		lines = append(lines, "stdout_log="+stdoutPath)
		if tail := TailFile(stdoutPath, 30); strings.TrimSpace(tail) != "" {
			lines = append(lines, "stdout_tail:\n"+tail)
		}
	}
	if manager.Kind == "darwin" {
		var launchctlErr error
		for _, domain := range manager.domainCandidates() {
			target := domain + "/" + LaunchdDaemonLabel
			out, err := RunCommand("launchctl", "print", target)
			if err != nil {
				launchctlErr = err
				continue
			}
			if tail := TailTextLines(out, 30); strings.TrimSpace(tail) != "" {
				lines = append(lines, "launchctl_print_tail("+target+"):\n"+tail)
			}
			launchctlErr = nil
			break
		}
		if launchctlErr != nil {
			lines = append(lines, fmt.Sprintf("launchctl_print_error=%v", launchctlErr))
		}
	}
	if manager.Kind == "linux" {
		if out, err := RunCommand("systemctl", "--user", "status", SystemdDaemonUnit, "--no-pager", "-n", "30"); err == nil {
			if tail := TailTextLines(out, 30); strings.TrimSpace(tail) != "" {
				lines = append(lines, "systemctl_status_tail:\n"+tail)
			}
		}
		if out, err := RunCommand("journalctl", "--user-unit", SystemdDaemonUnit, "-n", "30", "--no-pager"); err == nil {
			if tail := TailTextLines(out, 30); strings.TrimSpace(tail) != "" {
				lines = append(lines, "journalctl_tail:\n"+tail)
			}
		}
	}
	if manager.Kind == "windows" {
		if out, err := RunCommand("schtasks", "/Query", "/TN", WindowsScheduledTask, "/V", "/FO", "LIST"); err == nil {
			if tail := TailTextLines(out, 40); strings.TrimSpace(tail) != "" {
				lines = append(lines, "schtasks_query_tail:\n"+tail)
			}
		}
	}
	return strings.Join(lines, "\n")
}

func TailFile(path string, maxLines int) string {
	raw, err := os.ReadFile(strings.TrimSpace(path))
	if err != nil {
		return ""
	}
	return TailTextLines(string(raw), maxLines)
}

func TailTextLines(text string, maxLines int) string {
	text = strings.TrimSpace(strings.ReplaceAll(text, "\r\n", "\n"))
	if text == "" {
		return ""
	}
	if maxLines <= 0 {
		maxLines = 20
	}
	parts := strings.Split(text, "\n")
	if len(parts) <= maxLines {
		return strings.Join(parts, "\n")
	}
	return strings.Join(parts[len(parts)-maxLines:], "\n")
}

func spawnDaemonProcess(socketPath string, verbose bool) error {
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve executable: %w", err)
	}
	exe = ResolveStableExecutable(exe)

	stateDir, err := telemetry.DefaultStateDir()
	if err == nil && strings.TrimSpace(stateDir) != "" {
		_ = os.MkdirAll(stateDir, 0o755)
	}

	args := []string{"daemon", "run", "--socket-path", socketPath}
	if verbose {
		args = append(args, "--verbose")
	}

	cmd := exec.Command(exe, args...)
	if stateDir != "" {
		stdoutLog, err := os.OpenFile(filepath.Join(stateDir, "daemon.stdout.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err == nil {
			cmd.Stdout = stdoutLog
		}
		stderrLog, err := os.OpenFile(filepath.Join(stateDir, "daemon.stderr.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err == nil {
			cmd.Stderr = stderrLog
		}
	}

	if err := cmd.Start(); err != nil {
		return err
	}
	if cmd.Process != nil {
		_ = cmd.Process.Release()
	}
	return nil
}
