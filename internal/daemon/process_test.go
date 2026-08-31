package daemon

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/nurulislamz/agentusage/internal/version"
)

func TestIsReleaseSemver(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		wantOK bool
	}{
		{name: "release", input: "v0.4.0", wantOK: true},
		{name: "release with spaces", input: "  v1.2.3  ", wantOK: true},
		{name: "dev", input: "dev", wantOK: false},
		{name: "dirty snapshot", input: "v0.4.0-11-g0aa98a4-dirty", wantOK: false},
		{name: "missing patch", input: "v0.4", wantOK: false},
		{name: "missing v", input: "0.4.0", wantOK: false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := IsReleaseSemver(tt.input); got != tt.wantOK {
				t.Fatalf("IsReleaseSemver(%q) = %v, want %v", tt.input, got, tt.wantOK)
			}
		})
	}
}

func TestHealthCurrent(t *testing.T) {
	origVersion := version.Version
	t.Cleanup(func() {
		version.Version = origVersion
	})
	registryHash := ProviderRegistryHash()

	t.Run("release requires exact daemon version", func(t *testing.T) {
		version.Version = "v0.4.0"
		health := HealthResponse{
			DaemonVersion:    "dev",
			APIVersion:       APIVersion,
			ProviderRegistry: registryHash,
		}
		if HealthCurrent(health) {
			t.Fatal("HealthCurrent() = true, want false")
		}
	})

	t.Run("release accepts exact daemon version", func(t *testing.T) {
		version.Version = "v0.4.0"
		health := HealthResponse{
			DaemonVersion:    "v0.4.0",
			APIVersion:       APIVersion,
			ProviderRegistry: registryHash,
		}
		if !HealthCurrent(health) {
			t.Fatal("HealthCurrent() = false, want true")
		}
	})

	t.Run("local snapshot accepts running dev daemon", func(t *testing.T) {
		version.Version = "v0.4.0-11-g0aa98a4-dirty"
		health := HealthResponse{
			DaemonVersion:    "dev",
			APIVersion:       APIVersion,
			ProviderRegistry: registryHash,
		}
		if !HealthCurrent(health) {
			t.Fatal("HealthCurrent() = false, want true")
		}
	})

	t.Run("api mismatch stays incompatible", func(t *testing.T) {
		version.Version = "v0.4.0-11-g0aa98a4-dirty"
		health := HealthResponse{
			DaemonVersion:    "dev",
			APIVersion:       "v2",
			ProviderRegistry: registryHash,
		}
		if HealthCurrent(health) {
			t.Fatal("HealthCurrent() = true, want false")
		}
	})

	t.Run("missing provider registry hash is incompatible for release builds", func(t *testing.T) {
		version.Version = "v0.4.0"
		health := HealthResponse{
			DaemonVersion: "v0.4.0",
			APIVersion:    APIVersion,
		}
		if HealthCurrent(health) {
			t.Fatal("HealthCurrent() = true, want false")
		}
	})

	t.Run("missing provider registry hash is tolerated for local snapshots", func(t *testing.T) {
		version.Version = "v0.4.0-11-g0aa98a4-dirty"
		health := HealthResponse{
			DaemonVersion: "dev",
			APIVersion:    APIVersion,
		}
		if !HealthCurrent(health) {
			t.Fatal("HealthCurrent() = false, want true")
		}
	})
}

func TestClassifyEnsureError_EdgeCases(t *testing.T) {
	// 1. Nil error -> running
	stNil := ClassifyEnsureError(nil)
	if stNil.Status != DaemonStatusRunning {
		t.Errorf("ClassifyEnsureError(nil) = %v, want Running", stNil.Status)
	}

	// 2. Not installed
	stNotInstalled := ClassifyEnsureError(fmt.Errorf("service not installed"))
	if stNotInstalled.Status != DaemonStatusNotInstalled {
		t.Errorf("ClassifyEnsureError('not installed') = %v, want NotInstalled", stNotInstalled.Status)
	}
	if stNotInstalled.InstallHint == "" {
		t.Error("InstallHint should not be empty for not installed state")
	}

	// 3. Out of date
	stOutdated := ClassifyEnsureError(fmt.Errorf("telemetry daemon is out of date"))
	if stOutdated.Status != DaemonStatusOutdated {
		t.Errorf("ClassifyEnsureError('out of date') = %v, want Outdated", stOutdated.Status)
	}

	// 4. Unsupported
	stUnsupported := ClassifyEnsureError(fmt.Errorf("unsupported on plan9"))
	if stUnsupported.Status != DaemonStatusError {
		t.Errorf("ClassifyEnsureError('unsupported') = %v, want Error", stUnsupported.Status)
	}

	// 5. Generic error
	stGen := ClassifyEnsureError(fmt.Errorf("connection refused"))
	if stGen.Status != DaemonStatusError || stGen.Message != "connection refused" {
		t.Errorf("ClassifyEnsureError(generic) = %+v", stGen)
	}
}

func TestHealthVersion(t *testing.T) {
	if got := HealthVersion(HealthResponse{DaemonVersion: "v1.2.3"}); got != "v1.2.3" {
		t.Errorf("HealthVersion = %q, want 'v1.2.3'", got)
	}
	if got := HealthVersion(HealthResponse{DaemonVersion: ""}); got != "unknown" {
		t.Errorf("HealthVersion = %q, want 'unknown'", got)
	}
}

func TestTailTextLines_And_TailFile(t *testing.T) {
	text := "line1\nline2\nline3\nline4\nline5"

	// 1. Full text within limit
	if got := TailTextLines(text, 10); got != text {
		t.Errorf("TailTextLines(10) = %q, want full text", got)
	}

	// 2. Truncated tail
	if got := TailTextLines(text, 2); got != "line4\nline5" {
		t.Errorf("TailTextLines(2) = %q, want 'line4\\nline5'", got)
	}

	// 3. Empty text
	if got := TailTextLines("", 5); got != "" {
		t.Errorf("TailTextLines(\"\") = %q, want empty", got)
	}

	// 4. Default limit when <= 0
	if got := TailTextLines(text, 0); got != text {
		t.Errorf("TailTextLines(0) = %q, want text", got)
	}

	// 5. TailFile with temp file
	tmp := t.TempDir()
	filePath := filepath.Join(tmp, "tail_test.txt")
	if err := os.WriteFile(filePath, []byte(text), 0644); err != nil {
		t.Fatal(err)
	}
	if got := TailFile(filePath, 2); got != "line4\nline5" {
		t.Errorf("TailFile(2) = %q, want 'line4\\nline5'", got)
	}

	// 6. TailFile missing file returns empty
	if got := TailFile(filepath.Join(tmp, "missing.txt"), 5); got != "" {
		t.Errorf("TailFile(missing) = %q, want empty", got)
	}
}

func TestStartupDiagnostics(t *testing.T) {
	tmp := t.TempDir()
	logPath := filepath.Join(tmp, "daemon.stderr.log")
	_ = os.WriteFile(logPath, []byte("error: some daemon log"), 0644)

	// 1. Linux manager
	mgrLinux := ServiceManager{
		Kind:       "linux",
		exePath:    "/usr/local/bin/agentusage",
		socketPath: "/tmp/test.sock",
		unitPath:   "/home/user/.config/systemd/user/agentusage.service",
		stateDir:   tmp,
	}
	diagLinux := StartupDiagnostics(mgrLinux, "/tmp/test.sock")
	if !strings.Contains(diagLinux, "manager_kind=linux") {
		t.Errorf("StartupDiagnostics missing manager_kind: %s", diagLinux)
	}

	// 2. Darwin manager
	mgrDarwin := ServiceManager{
		Kind:       "darwin",
		exePath:    "/usr/local/bin/agentusage",
		socketPath: "/tmp/test.sock",
		unitPath:   "/home/user/Library/LaunchAgents/com.agentusage.daemon.plist",
		stateDir:   tmp,
	}
	diagDarwin := StartupDiagnostics(mgrDarwin, "/tmp/test.sock")
	if !strings.Contains(diagDarwin, "manager_kind=darwin") {
		t.Errorf("StartupDiagnostics missing manager_kind: %s", diagDarwin)
	}

	// 3. Windows manager
	mgrWin := ServiceManager{
		Kind:       "windows",
		exePath:    "C:\\agentusage.exe",
		socketPath: "C:\\daemon.sock",
		stateDir:   tmp,
	}
	diagWin := StartupDiagnostics(mgrWin, "C:\\daemon.sock")
	if !strings.Contains(diagWin, "manager_kind=windows") {
		t.Errorf("StartupDiagnostics missing manager_kind: %s", diagWin)
	}
}

func TestEnsureRunning_EmptySocket(t *testing.T) {
	_, err := EnsureRunning(context.Background(), "", false)
	if err == nil || !strings.Contains(err.Error(), "socket path is empty") {
		t.Errorf("EnsureRunning(\"\") err = %v, want empty socket path error", err)
	}
}

func TestWaitForHealth_NilAndTimeout(t *testing.T) {
	// 1. Nil client
	if err := WaitForHealth(context.Background(), nil, 100*time.Millisecond); err == nil {
		t.Error("WaitForHealth(nil) should return error")
	}

	// 2. Offline client with short timeout
	client := &Client{
		SocketPath: "/tmp/nonexistent_socket_test.sock",
		http:       http.DefaultClient,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	if err := WaitForHealth(ctx, client, 250*time.Millisecond); err == nil {
		t.Error("WaitForHealth on offline client should return error")
	}
}

func TestSpawnDaemonProcess(t *testing.T) {
	err := spawnDaemonProcess("/tmp/test.sock", true)
	if err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Errorf("spawnDaemonProcess err = %v, want unsupported message", err)
	}
}

func TestStartViaManagedService_And_EnsureViaServiceManager(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	client := NewClient("/tmp/test.sock")

	// 1. startViaManagedService needsUpgrade with transient exe
	mgrTransient := ServiceManager{
		Kind:    "linux",
		exePath: "/tmp/go-build123/exe/main",
	}
	_, err := startViaManagedService(ctx, client, mgrTransient, true, "/tmp/test.sock")
	if err == nil || !strings.Contains(err.Error(), "upgrade telemetry daemon service") {
		t.Errorf("expected upgrade error, got: %v", err)
	}

	// 2. startViaManagedService not installed
	mgrNotInstalled := ServiceManager{
		Kind:     "linux",
		unitPath: "/tmp/nonexistent.service",
	}
	_, err = startViaManagedService(ctx, client, mgrNotInstalled, false, "/tmp/test.sock")
	if err == nil || !strings.Contains(err.Error(), "telemetry daemon service is not installed") {
		t.Errorf("expected not installed error, got: %v", err)
	}

	// 2b. startViaManagedService installed but fails to start
	tmpUnit := filepath.Join(t.TempDir(), "agentusage.service")
	_ = os.WriteFile(tmpUnit, []byte(""), 0644)
	mgrInstalled := ServiceManager{
		Kind:     "linux",
		unitPath: tmpUnit,
	}
	_, err = startViaManagedService(ctx, client, mgrInstalled, false, "/tmp/test.sock")
	if err == nil {
		t.Error("expected error when service fails to start")
	}

	// 3. ensureViaServiceManager with needsUpgrade and transient exe
	_, err = ensureViaServiceManager(ctx, client, "/tmp/test.sock", false, true, HealthResponse{
		DaemonVersion: "v0.0.1",
	})
	if err == nil {
		t.Error("expected error from ensureViaServiceManager with needsUpgrade on transient binary")
	}
}

func TestWaitAndVerifyDaemon_OfflineClient(t *testing.T) {
	offlineClient := &Client{
		SocketPath: "/tmp/nonexistent_test_wait.sock",
		http:       http.DefaultClient,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := waitAndVerifyDaemon(ctx, offlineClient, "/tmp/nonexistent_test_wait.sock")
	if err == nil {
		t.Error("expected error when waiting for offline daemon")
	}
}

func TestEnsureRunning_OfflineSocket(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	_, err := EnsureRunning(ctx, "/tmp/nonexistent_socket_test.sock", false)
	if err == nil {
		t.Error("expected error ensuring running on nonexistent socket")
	}
}
