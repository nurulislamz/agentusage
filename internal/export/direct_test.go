package export

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/nurulislamz/agentusage/internal/core"
)

type ctxIgnoringProvider struct {
	id   string
	wait time.Duration
}

func (p ctxIgnoringProvider) ID() string                  { return p.id }
func (p ctxIgnoringProvider) Describe() core.ProviderInfo { return core.ProviderInfo{Name: p.id} }
func (p ctxIgnoringProvider) Spec() core.ProviderSpec     { return core.ProviderSpec{} }
func (p ctxIgnoringProvider) DashboardWidget() core.DashboardWidget {
	return core.DashboardWidget{}
}
func (p ctxIgnoringProvider) DetailWidget() core.DetailWidget { return core.DetailWidget{} }

func (p ctxIgnoringProvider) Fetch(_ context.Context, acct core.AccountConfig) (core.UsageSnapshot, error) {
	time.Sleep(p.wait)
	snap := core.NewUsageSnapshot(acct.Provider, acct.ID)
	snap.Status = core.StatusOK
	return snap, nil
}

type fastProvider struct{ id string }

func (p fastProvider) ID() string                  { return p.id }
func (p fastProvider) Describe() core.ProviderInfo { return core.ProviderInfo{Name: p.id} }
func (p fastProvider) Spec() core.ProviderSpec     { return core.ProviderSpec{} }
func (p fastProvider) DashboardWidget() core.DashboardWidget {
	return core.DashboardWidget{}
}
func (p fastProvider) DetailWidget() core.DetailWidget { return core.DetailWidget{} }

func (p fastProvider) Fetch(_ context.Context, acct core.AccountConfig) (core.UsageSnapshot, error) {
	snap := core.NewUsageSnapshot(acct.Provider, acct.ID)
	snap.Status = core.StatusOK
	snap.Message = "ok"
	return snap, nil
}

func TestCollectSnapshotsReturnsWhenContextCancels(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()

	accounts := []core.AccountConfig{
		{ID: "fast-1", Provider: "openai"},
		{ID: "stuck-1", Provider: "cursor"},
	}
	providers := map[string]core.UsageProvider{
		"openai": fastProvider{id: "openai"},
		"cursor": ctxIgnoringProvider{id: "cursor", wait: 8 * time.Second},
	}

	start := time.Now()
	done := make(chan []core.UsageSnapshot, 1)
	go func() {
		done <- collectSnapshots(ctx, accounts, providers, core.ModelNormalizationConfig{}, time.Now)
	}()

	var snaps []core.UsageSnapshot
	select {
	case snaps = <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("collectSnapshots ignored context cancel and kept waiting on Fetch")
	}
	if elapsed := time.Since(start); elapsed > 400*time.Millisecond {
		t.Fatalf("collectSnapshots took %v, want return shortly after cancel", elapsed)
	}

	byID := make(map[string]core.UsageSnapshot, len(snaps))
	for _, snap := range snaps {
		byID[snap.AccountID] = snap
	}
	if snap, ok := byID["fast-1"]; !ok || snap.Status != core.StatusOK {
		t.Fatalf("fast-1 = %+v, want completed OK snapshot", snap)
	}
	stuck, ok := byID["stuck-1"]
	if !ok {
		t.Fatal("missing timeout snapshot for stuck-1")
	}
	if stuck.Status != core.StatusError {
		t.Fatalf("stuck-1 status = %q, want ERROR", stuck.Status)
	}
	if !strings.Contains(stuck.Message, "timed out") {
		t.Fatalf("stuck-1 message = %q, want timed out", stuck.Message)
	}
}
