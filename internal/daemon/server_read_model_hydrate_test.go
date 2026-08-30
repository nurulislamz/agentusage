package daemon

import (
	"context"
	"testing"
	"time"

	"github.com/nurulislamz/agentusage/internal/core"
)

type mockHydrateProvider struct{}

func (m mockHydrateProvider) ID() string                            { return "mock" }
func (m mockHydrateProvider) Describe() core.ProviderInfo           { return core.ProviderInfo{Name: "mock"} }
func (m mockHydrateProvider) Spec() core.ProviderSpec               { return core.ProviderSpec{} }
func (m mockHydrateProvider) DashboardWidget() core.DashboardWidget { return core.DashboardWidget{} }
func (m mockHydrateProvider) DetailWidget() core.DetailWidget       { return core.DetailWidget{} }
func (m mockHydrateProvider) HasChanged(acct core.AccountConfig, since time.Time) (bool, error) {
	return true, nil
}
func (m mockHydrateProvider) Fetch(ctx context.Context, acct core.AccountConfig) (core.UsageSnapshot, error) {
	rem := 96.0
	return core.UsageSnapshot{
		ProviderID: acct.Provider,
		AccountID:  acct.ID,
		Status:     core.StatusOK,
		Metrics: map[string]core.Metric{
			"quota_gemini": {Remaining: &rem},
		},
	}, nil
}

func TestEnrichReadModelSnapshots_HydratesUnknownFromStatusFile(t *testing.T) {
	svc := &Service{
		providerByID: map[string]core.UsageProvider{
			"mock": mockHydrateProvider{},
		},
	}
	accounts := []core.AccountConfig{{
		ID:       "mock-acct",
		Provider: "mock",
	}}
	snaps := map[string]core.UsageSnapshot{
		"mock-acct": {
			ProviderID: "mock",
			AccountID:  "mock-acct",
			Status:     core.StatusUnknown,
		},
	}

	got := svc.enrichReadModelSnapshots(context.Background(), accounts, core.DefaultModelNormalizationConfig(), snaps)
	snap := got["mock-acct"]
	if snap.Status != core.StatusOK {
		t.Fatalf("status = %q, want OK", snap.Status)
	}
	if len(snap.Metrics) == 0 {
		t.Fatalf("expected quota metrics after hydration, got metrics=%v", snap.Metrics)
	}
}

func TestOverlayPollStateSnapshots_PrefersPolledData(t *testing.T) {
	now := time.Now().UTC()
	rem := 70.33
	svc := &Service{
		pollState: map[string]*providerPollState{
			"antigravity-physics": {
				hasSnap: true,
				lastSnap: core.UsageSnapshot{
					ProviderID: "antigravity",
					AccountID:  "antigravity-physics",
					Status:     core.StatusOK,
					Timestamp:  now,
					Metrics: map[string]core.Metric{
						"quota_gemini_weekly": {Remaining: &rem},
					},
				},
			},
		},
	}
	snaps := map[string]core.UsageSnapshot{
		"antigravity-physics": {
			ProviderID: "antigravity",
			AccountID:  "antigravity-physics",
			Status:     core.StatusUnknown,
			Timestamp:  now.Add(-time.Hour),
		},
	}

	got := svc.overlayPollStateSnapshots(snaps)
	if got["antigravity-physics"].Status != core.StatusOK {
		t.Fatalf("status = %q, want OK", got["antigravity-physics"].Status)
	}
}
