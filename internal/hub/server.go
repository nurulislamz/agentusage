package hub

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/nurulislamz/agentusage/internal/core"
	"github.com/nurulislamz/agentusage/internal/observability"
)

const (
	// maxPushBodyBytes caps /v1/push request bodies to guard against a
	// misbehaving or malicious worker exhausting hub memory. 4 MiB is
	// comfortably larger than any realistic snapshot batch.
	maxPushBodyBytes = 4 << 20
)

// Server receives RemoteEnvelope pushes from worker machines over TCP HTTP.
//
// When authToken is non-empty, /v1/push and /v1/snapshots require the header
// "Authorization: Bearer <token>". /healthz is always unauthenticated so that
// liveness probes work without secrets. When authToken is empty, all endpoints
// are open — suitable only for trusted LAN deployments.
type Server struct {
	addr      string
	store     *Store
	authToken string
}

func NewServer(addr string, store *Store) *Server {
	return NewServerWithAuth(addr, store, "")
}

// NewServerWithAuth creates a Server that requires Bearer token auth on
// mutating / data endpoints when authToken is non-empty. The caller is
// expected to have resolved any environment-variable fallback (e.g.
// AGENTUSAGE_HUB_TOKEN) before invoking this — see cmd/agentusage/hub.go's
// resolveHubRuntime for the canonical resolution path.
func NewServerWithAuth(addr string, store *Store, authToken string) *Server {
	return &Server{addr: addr, store: store, authToken: strings.TrimSpace(authToken)}
}

// AuthEnabled reports whether the server requires a Bearer token.
func (s *Server) AuthEnabled() bool {
	return s.authToken != ""
}

func (s *Server) ListenAndServe(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/push", s.handlePush)
	mux.HandleFunc("/v1/snapshots", s.handleSnapshots)
	mux.HandleFunc("/healthz", s.handleHealth)

	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("hub: listen %s: %w", s.addr, err)
	}

	observability.EmitInfo("hub", "server_start", "addr=%s auth=%t", s.addr, s.AuthEnabled())

	srv := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Serve(ln)
	}()

	select {
	case <-ctx.Done():
		observability.EmitInfo("hub", "server_stop", "reason=context_done")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}

// checkAuth returns true if the request is authorized (or auth is disabled).
// When returning false, it has already written a 401 response.
func (s *Server) checkAuth(w http.ResponseWriter, r *http.Request) bool {
	if s.authToken == "" {
		return true
	}
	header := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		observability.EmitWarn("hub", "auth_failed", "path=%s reason=missing_bearer_token", r.URL.Path)
		w.Header().Set("WWW-Authenticate", `Bearer realm="agentusage-hub"`)
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing bearer token"})
		return false
	}
	got := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	// Constant-time compare so an attacker can't enumerate the token
	// byte-by-byte via response-timing differences.
	if subtle.ConstantTimeCompare([]byte(got), []byte(s.authToken)) != 1 {
		observability.EmitWarn("hub", "auth_failed", "path=%s reason=invalid_token", r.URL.Path)
		w.Header().Set("WWW-Authenticate", `Bearer realm="agentusage-hub", error="invalid_token"`)
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid bearer token"})
		return false
	}
	return true
}

func (s *Server) handlePush(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	if !s.checkAuth(w, r) {
		return
	}

	start := time.Now()
	r.Body = http.MaxBytesReader(w, r.Body, maxPushBodyBytes)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		// MaxBytesReader returns a "http: request body too large" error when
		// the cap is exceeded; report 413 in that case, 400 otherwise.
		if strings.Contains(err.Error(), "request body too large") {
			observability.EmitWarn("hub", "push_rejected", "reason=body_too_large")
			writeJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"error": "request body too large"})
			return
		}
		observability.EmitWarn("hub", "push_rejected", "reason=read_body_failed error=%v", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "read body failed"})
		return
	}
	if len(strings.TrimSpace(string(body))) == 0 {
		observability.EmitWarn("hub", "push_rejected", "reason=empty_body")
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "empty body"})
		return
	}
	var env core.RemoteEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		observability.EmitWarn("hub", "push_rejected", "reason=invalid_json error=%v", err)
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}
	if strings.TrimSpace(env.Machine) == "" {
		observability.EmitWarn("hub", "push_rejected", "reason=missing_machine_name")
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "machine name required"})
		return
	}
	s.store.Ingest(env)
	observability.EmitInfo(
		"hub",
		"push_ingested",
		"machine=%s snapshots=%d duration_ms=%d",
		env.Machine,
		len(env.Snapshots),
		time.Since(start).Milliseconds(),
	)
	writeJSON(w, http.StatusOK, pushResponse{OK: true})
}

func (s *Server) handleSnapshots(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	if !s.checkAuth(w, r) {
		return
	}
	snaps := s.store.Snapshots()
	observability.EmitInfo("hub", "snapshots_read", "count=%d", len(snaps))
	writeJSON(w, http.StatusOK, snaps)
}

// handleHealth is always unauthenticated. It leaks only the list of machine
// names, which is considered non-sensitive enough to keep liveness probes
// simple in containerised deployments.
func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{
		Status:   "ok",
		Machines: s.store.MachineNames(),
	})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
