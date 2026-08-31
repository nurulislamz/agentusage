package daemon

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nurulislamz/agentusage/internal/config"
	"github.com/nurulislamz/agentusage/internal/core"
	"github.com/nurulislamz/agentusage/internal/telemetry"
)

func TestService_FullIntegration(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "telemetry.db")
	spoolDir := filepath.Join(tempDir, "spool")
	socketPath := filepath.Join(tempDir, "daemon.sock")

	cfg := Config{
		SocketPath:      socketPath,
		DBPath:          dbPath,
		SpoolDir:        spoolDir,
		CollectInterval: 10 * time.Minute,
		PollInterval:    10 * time.Minute,
		Verbose:         true,
		Export: config.ExportConfig{
			Target:          "http://127.0.0.1:9190",
			IntervalSeconds: 60,
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	svc, err := startService(ctx, cfg)
	if err != nil {
		t.Fatalf("startService error: %v", err)
	}
	defer svc.Close()

	// 1. Test IngestRequest and IngestBatch
	inTokens := int64(100)
	outTokens := int64(50)
	cost := 0.005
	req := telemetry.IngestRequest{
		SourceSystem:  telemetry.SourceSystem("claude_code"),
		SourceChannel: "hook",
		OccurredAt:    time.Now().UTC(),
		ProviderID:    "anthropic",
		AccountID:     "work-account",
		EventType:     "usage",
		TokenUsage:    core.TokenUsage{InputTokens: &inTokens, OutputTokens: &outTokens, CostUSD: &cost},
		ModelRaw:      "claude-3-5-sonnet",
	}

	tally, retries := svc.ingestBatch(ctx, []telemetry.IngestRequest{req})
	if tally.ingested != 1 || tally.failed != 0 || len(retries) != 0 {
		t.Errorf("ingestBatch tally = %+v, retries = %d", tally, len(retries))
	}

	// 2. Test IngestQuotaSnapshots and RecordBalanceObservations
	usedVal := 12.50
	limitVal := 100.00
	snap := core.UsageSnapshot{
		ProviderID: "openrouter",
		AccountID:  "openrouter-main",
		Timestamp:  time.Now().UTC(),
		Status:     core.StatusOK,
		Metrics: map[string]core.Metric{
			"credits": {
				Used:   &usedVal,
				Limit:  &limitVal,
				Unit:   "USD",
				Window: "1d",
			},
		},
	}

	err = svc.ingestQuotaSnapshots(ctx, map[string]core.UsageSnapshot{"openrouter-main": snap})
	if err != nil {
		t.Errorf("ingestQuotaSnapshots error: %v", err)
	}

	// 3. Test FlushBacklog
	flushRes, enqueued, warnings := svc.flushBacklog(ctx, nil, 100)
	if len(warnings) > 0 {
		t.Logf("flushBacklog warnings: %v", warnings)
	}
	_ = flushRes
	_ = enqueued

	// 4. Test HTTP Endpoints against live Service
	// Health
	reqHealth := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	wHealth := httptest.NewRecorder()
	svc.handleHealth(wHealth, reqHealth)
	if wHealth.Code != http.StatusOK {
		t.Errorf("handleHealth code = %d", wHealth.Code)
	}

	// ReadModel POST
	rmReq := ReadModelRequest{
		Accounts: []ReadModelAccount{
			{AccountID: "openrouter-main", ProviderID: "openrouter"},
		},
	}
	rmBody, _ := json.Marshal(rmReq)
	reqRM := httptest.NewRequest(http.MethodPost, "/v1/read-model", bytes.NewReader(rmBody))
	wRM := httptest.NewRecorder()
	svc.handleReadModel(wRM, reqRM)
	if wRM.Code != http.StatusOK {
		t.Errorf("handleReadModel code = %d, body: %s", wRM.Code, wRM.Body.String())
	}

	// Poll POST (kick)
	reqPoll := httptest.NewRequest(http.MethodPost, "/v1/poll", nil)
	wPoll := httptest.NewRecorder()
	svc.handlePoll(wPoll, reqPoll)
	if wPoll.Code != http.StatusOK {
		t.Errorf("handlePoll code = %d", wPoll.Code)
	}

	// 5. Test Retention and Orphan Pruning
	svc.pruneTelemetryOrphans(ctx)
	_ = svc.pruneOldData(ctx)

	// 6. Cancel context and verify shutdown
	cancel()
	time.Sleep(50 * time.Millisecond)

	// Stat socket after close
	_ = svc.Close()
}

func TestStartService_InvalidConfig(t *testing.T) {
	// Invalid DB Path (directory as DB)
	tempDir := t.TempDir()
	invalidDB := filepath.Join(tempDir, "nonexistent_parent_dir", "sub", "telemetry.db")
	_ = os.MkdirAll(invalidDB, 0755)

	cfg := Config{
		DBPath:   invalidDB,
		SpoolDir: tempDir,
	}

	_, err := startService(context.Background(), cfg)
	if err == nil {
		t.Error("expected error for invalid DBPath in startService")
	}

	// RunServer with invalid DB Path returns error
	runErr := RunServer(cfg)
	if runErr == nil {
		t.Error("expected error for invalid DBPath in RunServer")
	}
}
