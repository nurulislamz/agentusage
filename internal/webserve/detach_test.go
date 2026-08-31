package webserve

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	if os.Getenv("AGENTUSAGE_TEST_DETACH_CHILD") == "1" {
		if path := os.Getenv("AGENTUSAGE_TEST_DETACH_READY"); path != "" {
			_ = os.WriteFile(path, []byte("ok"), 0o600)
		}
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
		<-c
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestChildServeArgsStripsDetachAndAddsNoOpen(t *testing.T) {
	got := ChildServeArgs([]string{"agentusage", "serve", "--detach"})
	want := []string{"serve", "--no-open"}
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestChildServeArgsKeepsListenAndBasePath(t *testing.T) {
	got := ChildServeArgs([]string{
		"/tmp/exe", "serve", "--listen", "127.0.0.1:8088",
		"--detach", "--base-path", "/agentusage",
	})
	joined := strings.Join(got, " ")
	if !strings.Contains(joined, "--listen 127.0.0.1:8088") {
		t.Fatalf("missing listen: %q", got)
	}
	if !strings.Contains(joined, "--base-path /agentusage") {
		t.Fatalf("missing base-path: %q", got)
	}
	if strings.Contains(joined, "--detach") {
		t.Fatalf("child still has --detach: %q", got)
	}
	if got[len(got)-1] != "--no-open" {
		t.Fatalf("expected trailing --no-open, got %q", got)
	}
}

func TestChildServeArgsStripsOpen(t *testing.T) {
	got := ChildServeArgs([]string{"agentusage", "serve", "--open", "--detach=true"})
	joined := strings.Join(got, " ")
	if strings.Contains(joined, "--open") && !strings.Contains(joined, "--no-open") {
		t.Fatalf("should not keep --open: %q", got)
	}
	if strings.Contains(joined, "--detach") {
		t.Fatalf("should strip --detach=true: %q", got)
	}
}

func TestHealthzURL(t *testing.T) {
	cases := map[string]struct {
		listen, base, want string
	}{
		"defaults":    {"", "", "http://127.0.0.1:8080/healthz"},
		"prefix":      {"127.0.0.1:8088", "/agentusage", "http://127.0.0.1:8088/agentusage/healthz"},
		"all-ifaces":  {":8088", "", "http://127.0.0.1:8088/healthz"},
		"empty-slash": {"localhost:9090", "/", "http://localhost:9090/healthz"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := HealthzURL(tc.listen, tc.base); got != tc.want {
				t.Errorf("HealthzURL(%q, %q) = %q, want %q", tc.listen, tc.base, got, tc.want)
			}
		})
	}
}

func TestPIDFileRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := PIDFile(dir)
	if filepath.Base(path) != "serve.pid" {
		t.Fatalf("PIDFile base = %q", filepath.Base(path))
	}
	if err := WritePID(path, 4242); err != nil {
		t.Fatal(err)
	}
	got, err := ReadPID(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != 4242 {
		t.Fatalf("pid = %d", got)
	}
}

func TestWaitHealthy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	t.Cleanup(srv.Close)
	if err := WaitHealthy(srv.URL+"/healthz", time.Second); err != nil {
		t.Fatal(err)
	}
}

func TestWaitHealthyTimesOut(t *testing.T) {
	err := WaitHealthy("http://127.0.0.1:1/healthz", 50*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout")
	}
}

func TestStartAndStopDetached(t *testing.T) {
	dir := t.TempDir()
	pidPath := PIDFile(dir)
	logPath := LogFile(dir)
	ready := filepath.Join(dir, "ready")

	pid, err := StartDetached(DetachConfig{
		Executable: os.Args[0],
		PIDPath:    pidPath,
		LogPath:    logPath,
		ExtraEnv: []string{
			"AGENTUSAGE_TEST_DETACH_CHILD=1",
			"AGENTUSAGE_TEST_DETACH_READY=" + ready,
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if pid <= 0 {
		t.Fatalf("pid = %d", pid)
	}
	t.Cleanup(func() { _ = StopDetached(pidPath) })

	if !ProcessAlive(pid) {
		t.Fatal("child not alive")
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(ready); err == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if _, err := os.Stat(ready); err != nil {
		t.Fatalf("child never wrote ready file: %v", err)
	}

	running, ok := RunningPID(pidPath)
	if !ok || running != pid {
		t.Fatalf("RunningPID = %d, %v, want %d", running, ok, pid)
	}

	_, err = StartDetached(DetachConfig{
		Executable: os.Args[0],
		PIDPath:    pidPath,
		LogPath:    logPath,
		ExtraEnv:   []string{"AGENTUSAGE_TEST_DETACH_CHILD=1"},
	})
	if err == nil || !strings.Contains(err.Error(), "already running") {
		t.Fatalf("second start err = %v", err)
	}

	if err := StopDetached(pidPath); err != nil {
		t.Fatal(err)
	}
	deadline = time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && ProcessAlive(pid) {
		time.Sleep(20 * time.Millisecond)
	}
	if ProcessAlive(pid) {
		t.Fatal("child still alive after stop")
	}
	if _, err := os.Stat(pidPath); !os.IsNotExist(err) {
		t.Fatalf("pidfile still present: %v", err)
	}
}

func TestStartDetachedReplacesStalePIDFile(t *testing.T) {
	dir := t.TempDir()
	pidPath := PIDFile(dir)
	if err := WritePID(pidPath, 999999); err != nil {
		t.Fatal(err)
	}
	pid, err := StartDetached(DetachConfig{
		Executable: os.Args[0],
		PIDPath:    pidPath,
		LogPath:    LogFile(dir),
		ExtraEnv:   []string{"AGENTUSAGE_TEST_DETACH_CHILD=1"},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = StopDetached(pidPath) })
	if pid == 999999 {
		t.Fatal("did not replace stale pid")
	}
	if !ProcessAlive(pid) {
		t.Fatal("replacement child not alive")
	}
}

func TestStopDetachedIdempotentWhenNotRunning(t *testing.T) {
	if err := StopDetached(PIDFile(t.TempDir())); err != nil {
		t.Fatal(err)
	}
}

func TestValidateServeMode(t *testing.T) {
	if err := ValidateServeMode(true, true, false); err == nil {
		t.Fatal("detach+stop should error")
	}
	if err := ValidateServeMode(true, false, true); err == nil {
		t.Fatal("detach+verify should error")
	}
	if err := ValidateServeMode(false, true, true); err == nil {
		t.Fatal("stop+verify should error")
	}
	if err := ValidateServeMode(true, false, false); err != nil {
		t.Fatal(err)
	}
}
