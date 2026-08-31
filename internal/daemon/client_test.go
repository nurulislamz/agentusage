package daemon

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nurulislamz/agentusage/internal/core"
)

func TestClient_HealthInfo(t *testing.T) {
	// 1. Unconfigured client
	var nilClient *Client
	if _, err := nilClient.HealthInfo(context.Background()); err == nil {
		t.Error("nilClient.HealthInfo should return error")
	}
	emptyClient := &Client{SocketPath: ""}
	if _, err := emptyClient.HealthInfo(context.Background()); err == nil {
		t.Error("emptyClient.HealthInfo should return error")
	}

	// 2. Success response
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/healthz" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode(HealthResponse{
			Status:        "ok",
			DaemonVersion: "1.2.3",
			APIVersion:    "v1",
		})
	}))
	defer srv.Close()

	c := &Client{
		SocketPath: "/tmp/mock.sock",
	}

	// Direct handler test with RoundTripper mock
	mockTransport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		rec := httptest.NewRecorder()
		if req.URL.Path == "/healthz" {
			_ = json.NewEncoder(rec).Encode(HealthResponse{
				Status:        "ok",
				DaemonVersion: "1.2.3",
				APIVersion:    "v1",
			})
		} else {
			rec.WriteHeader(http.StatusNotFound)
		}
		return rec.Result(), nil
	})

	c.http = &http.Client{Transport: mockTransport}

	h, err := c.HealthInfo(context.Background())
	if err != nil {
		t.Fatalf("HealthInfo failed: %v", err)
	}
	if h.DaemonVersion != "1.2.3" || h.Status != "ok" {
		t.Errorf("HealthInfo = %+v, want status ok and version 1.2.3", h)
	}

	// Empty body returns status: ok
	emptyTransport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		rec := httptest.NewRecorder()
		rec.WriteHeader(http.StatusOK)
		return rec.Result(), nil
	})
	c.http = &http.Client{Transport: emptyTransport}
	hEmpty, err := c.HealthInfo(context.Background())
	if err != nil {
		t.Fatalf("HealthInfo on empty body: %v", err)
	}
	if hEmpty.Status != "ok" {
		t.Errorf("hEmpty.Status = %q, want 'ok'", hEmpty.Status)
	}

	// Error response
	errTransport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		rec := httptest.NewRecorder()
		rec.WriteHeader(http.StatusInternalServerError)
		return rec.Result(), nil
	})
	c.http = &http.Client{Transport: errTransport}
	if _, err := c.HealthInfo(context.Background()); err == nil {
		t.Error("expected error for 500 status in HealthInfo")
	}
}

func TestClient_ReadModel(t *testing.T) {
	mockTransport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		rec := httptest.NewRecorder()
		if req.URL.Path == "/v1/read-model" && req.Method == http.MethodPost {
			_ = json.NewEncoder(rec).Encode(ReadModelResponse{
				Snapshots: map[string]core.UsageSnapshot{
					"acc1": {ProviderID: "test-prov", AccountID: "acc1"},
				},
			})
		} else {
			rec.WriteHeader(http.StatusBadRequest)
		}
		return rec.Result(), nil
	})

	c := &Client{
		SocketPath: "/tmp/mock.sock",
		http:       &http.Client{Transport: mockTransport},
	}

	snaps, err := c.ReadModel(context.Background(), ReadModelRequest{
		Accounts: []ReadModelAccount{{ProviderID: "test-prov", AccountID: "acc1"}},
	})
	if err != nil {
		t.Fatalf("ReadModel failed: %v", err)
	}
	if len(snaps) != 1 || snaps["acc1"].ProviderID != "test-prov" {
		t.Errorf("ReadModel snapshots = %+v", snaps)
	}

	// Error path
	errTransport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		rec := httptest.NewRecorder()
		rec.WriteHeader(http.StatusInternalServerError)
		_, _ = rec.WriteString("internal read model crash")
		return rec.Result(), nil
	})
	c.http = &http.Client{Transport: errTransport}
	if _, err := c.ReadModel(context.Background(), ReadModelRequest{}); err == nil {
		t.Error("expected error for 500 status in ReadModel")
	}
}

func TestClient_IngestHook(t *testing.T) {
	var capturedPath string
	var capturedQuery string
	mockTransport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		capturedPath = req.URL.Path
		capturedQuery = req.URL.RawQuery
		rec := httptest.NewRecorder()
		_ = json.NewEncoder(rec).Encode(HookResponse{
			Source:   "opencode",
			Ingested: 3,
		})
		return rec.Result(), nil
	})

	c := &Client{
		SocketPath: "/tmp/mock.sock",
		http:       &http.Client{Transport: mockTransport},
	}

	resp, err := c.IngestHook(context.Background(), "opencode", "work-account", []byte(`{"event":"test"}`))
	if err != nil {
		t.Fatalf("IngestHook failed: %v", err)
	}
	if resp.Source != "opencode" || resp.Ingested != 3 {
		t.Errorf("IngestHook resp = %+v", resp)
	}
	if capturedPath != "/v1/hook/opencode" {
		t.Errorf("capturedPath = %q, want '/v1/hook/opencode'", capturedPath)
	}
	if capturedQuery != "account_id=work-account" {
		t.Errorf("capturedQuery = %q, want 'account_id=work-account'", capturedQuery)
	}

	// Error path
	errTransport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		rec := httptest.NewRecorder()
		rec.WriteHeader(http.StatusBadRequest)
		_, _ = rec.WriteString("invalid hook format")
		return rec.Result(), nil
	})
	c.http = &http.Client{Transport: errTransport}
	if _, err := c.IngestHook(context.Background(), "opencode", "", []byte(`bad`)); err == nil {
		t.Error("expected error for 400 status in IngestHook")
	}
}

func TestClient_RequestPoll_And_Wait(t *testing.T) {
	// 1. Unconfigured client
	var nilClient *Client
	if err := nilClient.RequestPoll(context.Background()); err == nil {
		t.Error("nilClient.RequestPoll should error")
	}

	// 2. Normal RequestPoll
	var capturedQuery string
	mockTransport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		capturedQuery = req.URL.RawQuery
		rec := httptest.NewRecorder()
		rec.WriteHeader(http.StatusOK)
		_, _ = rec.WriteString(`{"status":"kicked"}`)
		return rec.Result(), nil
	})

	c := &Client{
		SocketPath: "/tmp/mock.sock",
		http:       &http.Client{Transport: mockTransport},
	}

	if err := c.RequestPoll(context.Background()); err != nil {
		t.Fatalf("RequestPoll failed: %v", err)
	}
	if capturedQuery != "" {
		t.Errorf("RequestPoll query = %q, want empty", capturedQuery)
	}

	// 3. RequestPollWait
	if err := c.RequestPollWait(context.Background()); err != nil {
		t.Fatalf("RequestPollWait failed: %v", err)
	}
	if capturedQuery != "wait=1" {
		t.Errorf("RequestPollWait query = %q, want 'wait=1'", capturedQuery)
	}

	// 4. Server error
	errTransport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		rec := httptest.NewRecorder()
		rec.WriteHeader(http.StatusInternalServerError)
		_, _ = rec.WriteString("poll failed")
		return rec.Result(), nil
	})
	c.http = &http.Client{Transport: errTransport}
	if err := c.RequestPoll(context.Background()); err == nil {
		t.Error("expected error for 500 status in RequestPoll")
	}
}

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
