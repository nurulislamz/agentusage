package webserve

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// DetachConfig starts a background `agentusage serve` child.
type DetachConfig struct {
	Executable string
	Args       []string
	PIDPath    string
	LogPath    string
	HealthURL  string
	Timeout    time.Duration
	ExtraEnv   []string
}

func PIDFile(stateDir string) string {
	return filepath.Join(stateDir, "serve.pid")
}

func LogFile(stateDir string) string {
	return filepath.Join(stateDir, "serve.log")
}

func ValidateServeMode(detach, stop, verify bool) error {
	n := 0
	if detach {
		n++
	}
	if stop {
		n++
	}
	if verify {
		n++
	}
	if n > 1 {
		return fmt.Errorf("serve: --detach, --stop, and --verify are mutually exclusive")
	}
	return nil
}

// ChildServeArgs builds argv for a detached child from os.Args.
// It drops the executable, --detach/--stop/--open, and always passes --no-open.
func ChildServeArgs(osArgs []string) []string {
	out := make([]string, 0, len(osArgs))
	start := 0
	if len(osArgs) > 0 {
		start = 1
	}
	hasNoOpen := false
	for _, a := range osArgs[start:] {
		if isBoolFlag(a, "detach") || isBoolFlag(a, "stop") || isBoolFlag(a, "open") {
			continue
		}
		if isBoolFlag(a, "no-open") {
			hasNoOpen = true
		}
		out = append(out, a)
	}
	if len(out) == 0 {
		out = append(out, "serve")
	}
	if !hasNoOpen {
		out = append(out, "--no-open")
	}
	return out
}

func isBoolFlag(arg, name string) bool {
	return arg == "--"+name || strings.HasPrefix(arg, "--"+name+"=")
}

func HealthzURL(listenAddr, basePath string) string {
	addr := normalizeListenAddr(listenAddr)
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		prefix, _ := normalizeBasePath(basePath)
		return "http://" + addr + prefix + "/healthz"
	}
	if host == "" || host == "0.0.0.0" || host == "::" {
		host = "127.0.0.1"
	}
	if strings.Contains(host, ":") {
		host = "[" + host + "]"
	}
	prefix, _ := normalizeBasePath(basePath)
	return "http://" + host + ":" + port + prefix + "/healthz"
}

func WritePID(path string, pid int) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("serve: empty pid file path")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("serve: pid dir: %w", err)
	}
	return os.WriteFile(path, []byte(strconv.Itoa(pid)+"\n"), 0o600)
}

func ReadPID(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, fmt.Errorf("serve: invalid pid file %s: %w", path, err)
	}
	return pid, nil
}

func RunningPID(pidPath string) (int, bool) {
	pid, err := ReadPID(pidPath)
	if err != nil || pid <= 0 || !ProcessAlive(pid) {
		return 0, false
	}
	return pid, true
}

func WaitHealthy(url string, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 500 * time.Millisecond}
	var last error
	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
			last = fmt.Errorf("status %d", resp.StatusCode)
		} else {
			last = err
		}
		time.Sleep(50 * time.Millisecond)
	}
	if last == nil {
		last = fmt.Errorf("timeout")
	}
	return fmt.Errorf("serve: %s not ready: %w", url, last)
}

func StartDetached(cfg DetachConfig) (int, error) {
	if pid, ok := RunningPID(cfg.PIDPath); ok {
		return pid, fmt.Errorf("serve: already running (pid %d); stop it with agentusage serve --stop", pid)
	}
	exe := strings.TrimSpace(cfg.Executable)
	if exe == "" {
		var err error
		exe, err = os.Executable()
		if err != nil {
			return 0, fmt.Errorf("serve: executable: %w", err)
		}
	}
	if err := os.MkdirAll(filepath.Dir(cfg.LogPath), 0o755); err != nil {
		return 0, fmt.Errorf("serve: log dir: %w", err)
	}
	logFile, err := os.OpenFile(cfg.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return 0, fmt.Errorf("serve: open log: %w", err)
	}
	defer logFile.Close()

	stdin, err := os.Open(os.DevNull)
	if err != nil {
		return 0, fmt.Errorf("serve: open %s: %w", os.DevNull, err)
	}
	defer stdin.Close()

	cmd := exec.Command(exe, cfg.Args...)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Stdin = stdin
	cmd.Env = append(append([]string{}, os.Environ()...), cfg.ExtraEnv...)
	cmd.SysProcAttr = detachSysProcAttr()
	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("serve: detach start: %w", err)
	}
	pid := cmd.Process.Pid
	if err := WritePID(cfg.PIDPath, pid); err != nil {
		_ = terminatePID(pid)
		return 0, err
	}
	_ = cmd.Process.Release()

	if strings.TrimSpace(cfg.HealthURL) != "" {
		timeout := cfg.Timeout
		if timeout <= 0 {
			timeout = 10 * time.Second
		}
		if err := WaitHealthy(cfg.HealthURL, timeout); err != nil {
			_ = StopDetached(cfg.PIDPath)
			return 0, fmt.Errorf("%w\n  logs: %s", err, cfg.LogPath)
		}
	}
	return pid, nil
}

func StopDetached(pidPath string) error {
	pid, err := ReadPID(pidPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if pid <= 0 {
		_ = os.Remove(pidPath)
		return nil
	}
	if ProcessAlive(pid) {
		_ = terminatePID(pid)
		reapPID(pid, 5*time.Second)
		if ProcessAlive(pid) {
			_ = killPID(pid)
			reapPID(pid, 2*time.Second)
		}
		if ProcessAlive(pid) {
			return fmt.Errorf("serve: pid %d did not exit", pid)
		}
	}
	_ = os.Remove(pidPath)
	return nil
}

func reapPID(pid int, timeout time.Duration) {
	done := make(chan struct{})
	go func() {
		if proc, err := os.FindProcess(pid); err == nil {
			_, _ = proc.Wait()
		}
		close(done)
	}()
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for {
		if !ProcessAlive(pid) {
			return
		}
		select {
		case <-done:
			for time.Now().Before(deadline) && ProcessAlive(pid) {
				time.Sleep(50 * time.Millisecond)
			}
			return
		case <-ticker.C:
			if !time.Now().Before(deadline) {
				return
			}
		}
	}
}
