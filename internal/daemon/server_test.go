package daemon

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/nurulislamz/agentusage/internal/core"
	"github.com/nurulislamz/agentusage/internal/providers"
)

func shortSocketPath(t *testing.T, suffix string) string {
	t.Helper()
	// Use the OS temp dir (short enough to stay under the AF_UNIX sun_path
	// limit) rather than a hardcoded /tmp, which does not exist on Windows.
	return filepath.Join(os.TempDir(), fmt.Sprintf("agentusage-%d-%s.sock", time.Now().UnixNano(), strings.TrimSpace(suffix)))
}

func TestEnsureSocketPathAvailable_ActiveSocketReturnsError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix sockets are not supported in this test")
	}

	socketPath := shortSocketPath(t, "active")
	_ = os.Remove(socketPath)
	t.Cleanup(func() { _ = os.Remove(socketPath) })
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("listen unix socket: %v", err)
	}
	defer listener.Close()

	err = EnsureSocketPathAvailable(socketPath)
	if err == nil {
		t.Fatal("expected error for active daemon socket")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "already running") {
		t.Fatalf("error = %q, want already running message", err)
	}
}

func TestEnsureSocketPathAvailable_RemovesStaleSocket(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix sockets are not supported in this test")
	}

	socketPath := shortSocketPath(t, "stale")
	_ = os.Remove(socketPath)
	t.Cleanup(func() { _ = os.Remove(socketPath) })
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("listen unix socket: %v", err)
	}
	if err := listener.Close(); err != nil {
		t.Fatalf("close listener: %v", err)
	}

	if _, statErr := os.Stat(socketPath); statErr != nil && !os.IsNotExist(statErr) {
		t.Fatalf("stat socket before ensure: %v", statErr)
	}

	if err := EnsureSocketPathAvailable(socketPath); err != nil {
		t.Fatalf("ensure socket path available: %v", err)
	}

	if _, statErr := os.Stat(socketPath); !os.IsNotExist(statErr) {
		t.Fatalf("expected stale socket to be removed, stat err = %v", statErr)
	}
}

func TestEnsureSocketPathAvailable_RejectsRegularFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		// On Windows an AF_UNIX socket path is materialized as a regular file
		// and never reports os.ModeSocket, so we deliberately cannot distinguish
		// a leftover socket from any other file: EnsureSocketPathAvailable treats
		// it as a possibly-stale socket and removes it after a failed dial probe
		// (see socket_windows.go). The "reject regular file" semantic is
		// macOS/Linux-only.
		t.Skip("regular-file rejection is not applicable to Windows AF_UNIX sockets")
	}
	socketPath := shortSocketPath(t, "file")
	_ = os.Remove(socketPath)
	t.Cleanup(func() { _ = os.Remove(socketPath) })
	if err := os.WriteFile(socketPath, []byte("not-a-socket"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	err := EnsureSocketPathAvailable(socketPath)
	if err == nil {
		t.Fatal("expected error for regular file at socket path")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "not a socket") {
		t.Fatalf("error = %q, want not a socket message", err)
	}
}

func TestDefaultCollectOptions_GeminiHasSessionsDir(t *testing.T) {
	source, ok := providers.TelemetrySourceBySystem("gemini_cli")
	if !ok {
		t.Skip("gemini_cli telemetry source not found in registry")
	}
	opts := source.DefaultCollectOptions()

	if got := opts.Paths["sessions_dir"]; got == "" {
		t.Fatal("expected non-empty sessions_dir from gemini DefaultCollectOptions")
	}
	if _, ok := opts.Paths["projects_dir"]; ok {
		t.Fatalf("unexpected claude projects_dir in gemini opts: %+v", opts.Paths)
	}
}

func TestStartSocketServer_Permissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix sockets are not supported in this test")
	}

	socketPath := shortSocketPath(t, "perms")
	_ = os.Remove(socketPath)
	t.Cleanup(func() { _ = os.Remove(socketPath) })

	svc := &Service{
		cfg: Config{
			SocketPath: socketPath,
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := svc.startSocketServer(ctx); err != nil {
		t.Fatalf("startSocketServer: %v", err)
	}

	info, err := os.Stat(socketPath)
	if err != nil {
		t.Fatalf("stat socket: %v", err)
	}

	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("socket permissions = %o, want 0600", perm)
	}
}

func TestHandleHook_PayloadLimit(t *testing.T) {
	svc := &Service{}
	oversized := bytes.Repeat([]byte("a"), 5<<20) // 5 MiB (limit is 4 MiB)
	req := httptest.NewRequest(http.MethodPost, "/v1/hook/opencode", bytes.NewReader(oversized))
	w := httptest.NewRecorder()

	svc.handleHook(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d (RequestEntityTooLarge)", w.Code, http.StatusRequestEntityTooLarge)
	}
}

func TestHandleHealth(t *testing.T) {
	svc := &Service{}
	req := httptest.NewRequest(http.MethodGet, "/v1/health", nil)
	w := httptest.NewRecorder()

	svc.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if !strings.Contains(w.Body.String(), `"status":"ok"`) {
		t.Errorf("health body = %s, want status:ok", w.Body.String())
	}
}

func TestHandlePoll_Methods(t *testing.T) {
	svc := &Service{
		pollKick: make(chan struct{}, 1),
	}

	// 1. GET returns 405
	reqGet := httptest.NewRequest(http.MethodGet, "/v1/poll", nil)
	wGet := httptest.NewRecorder()
	svc.handlePoll(wGet, reqGet)
	if wGet.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET /v1/poll status = %d, want 405", wGet.Code)
	}

	// 2. POST without wait returns 200 kicked
	reqPost := httptest.NewRequest(http.MethodPost, "/v1/poll", nil)
	wPost := httptest.NewRecorder()
	svc.handlePoll(wPost, reqPost)
	if wPost.Code != http.StatusOK {
		t.Errorf("POST /v1/poll status = %d, want 200", wPost.Code)
	}
	if !strings.Contains(wPost.Body.String(), `"status":"kicked"`) {
		t.Errorf("POST /v1/poll body = %s, want status:kicked", wPost.Body.String())
	}

	// 3. POST with wait=1 returns 200 polled
	reqWait := httptest.NewRequest(http.MethodPost, "/v1/poll?wait=1", nil)
	wWait := httptest.NewRecorder()
	svc.handlePoll(wWait, reqWait)
	if wWait.Code != http.StatusOK {
		t.Errorf("POST /v1/poll?wait=1 status = %d, want 200", wWait.Code)
	}
	if !strings.Contains(wWait.Body.String(), `"status":"polled"`) {
		t.Errorf("POST /v1/poll?wait=1 body = %s, want status:polled", wWait.Body.String())
	}
}

func TestHandleHook_Validation(t *testing.T) {
	svc := &Service{
		pollKick: make(chan struct{}, 1),
	}

	// 1. GET returns 405
	reqGet := httptest.NewRequest(http.MethodGet, "/v1/hook/opencode", nil)
	wGet := httptest.NewRecorder()
	svc.handleHook(wGet, reqGet)
	if wGet.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET /v1/hook status = %d, want 405", wGet.Code)
	}

	// 2. Missing source -> 400
	reqNoSource := httptest.NewRequest(http.MethodPost, "/v1/hook/", bytes.NewReader([]byte("{}")))
	wNoSource := httptest.NewRecorder()
	svc.handleHook(wNoSource, reqNoSource)
	if wNoSource.Code != http.StatusBadRequest {
		t.Errorf("POST /v1/hook/ (no source) status = %d, want 400", wNoSource.Code)
	}

	// 3. Empty payload -> 400
	reqEmpty := httptest.NewRequest(http.MethodPost, "/v1/hook/opencode", bytes.NewReader([]byte("   ")))
	wEmpty := httptest.NewRecorder()
	svc.handleHook(wEmpty, reqEmpty)
	if wEmpty.Code != http.StatusBadRequest {
		t.Errorf("POST /v1/hook/opencode (empty) status = %d, want 400", wEmpty.Code)
	}

	// 4. Invalid JSON payload -> 400
	reqInvalid := httptest.NewRequest(http.MethodPost, "/v1/hook/opencode", bytes.NewReader([]byte("not json!")))
	wInvalid := httptest.NewRecorder()
	svc.handleHook(wInvalid, reqInvalid)
	if wInvalid.Code != http.StatusBadRequest {
		t.Errorf("POST /v1/hook/opencode (invalid) status = %d, want 400", wInvalid.Code)
	}
}

func TestHandleReadModel_Validation(t *testing.T) {
	svc := &Service{}

	// 1. GET returns 405
	reqGet := httptest.NewRequest(http.MethodGet, "/v1/read-model", nil)
	wGet := httptest.NewRecorder()
	svc.handleReadModel(wGet, reqGet)
	if wGet.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET /v1/read-model status = %d, want 405", wGet.Code)
	}

	// 2. Invalid JSON body -> 400
	reqInvalid := httptest.NewRequest(http.MethodPost, "/v1/read-model", bytes.NewReader([]byte("invalid json")))
	wInvalid := httptest.NewRecorder()
	svc.handleReadModel(wInvalid, reqInvalid)
	if wInvalid.Code != http.StatusBadRequest {
		t.Errorf("POST /v1/read-model (invalid) status = %d, want 400", wInvalid.Code)
	}
}

func TestReadModelCache_OperationsAndEviction(t *testing.T) {
	cache := newReadModelCache()

	// 1. Get on non-existent or empty key
	if _, _, ok := cache.get(""); ok {
		t.Error("get(\"\") returned true, want false")
	}
	if _, _, ok := cache.get("missing"); ok {
		t.Error("get(\"missing\") returned true, want false")
	}

	// 2. Set and Get
	snaps := map[string]core.UsageSnapshot{
		"acc1": {
			ProviderID: "test-prov",
			AccountID:  "acc1",
			Status:     core.StatusOK,
		},
	}
	cache.set("key1", snaps)
	cached, updatedAt, ok := cache.get("key1")
	if !ok {
		t.Fatal("get(\"key1\") returned false after set")
	}
	if len(cached) != 1 || cached["acc1"].ProviderID != "test-prov" {
		t.Errorf("cached snapshot = %+v", cached)
	}
	if updatedAt.IsZero() {
		t.Error("updatedAt should not be zero")
	}

	// 3. In-flight refresh locking
	if !cache.beginRefresh("key1") {
		t.Error("beginRefresh(\"key1\") should return true on first call")
	}
	if cache.beginRefresh("key1") {
		t.Error("beginRefresh(\"key1\") should return false when already in flight")
	}
	if cache.beginRefresh("") {
		t.Error("beginRefresh(\"\") should return false")
	}
	cache.endRefresh("key1")
	if !cache.beginRefresh("key1") {
		t.Error("beginRefresh(\"key1\") should return true after endRefresh")
	}
	cache.endRefresh("key1")

	// 4. Cache eviction when exceeding maxEntries (50)
	for i := 0; i < 60; i++ {
		k := fmt.Sprintf("k-%d", i)
		cache.set(k, snaps)
	}
	cache.mu.RLock()
	totalEntries := len(cache.entries)
	cache.mu.RUnlock()
	if totalEntries > 50 {
		t.Errorf("cache entries = %d, want <= 50 after eviction", totalEntries)
	}
}
