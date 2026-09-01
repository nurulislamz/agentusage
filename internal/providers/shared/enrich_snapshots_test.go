package shared

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nurulislamz/agentusage/internal/core"
)

func TestSnapshotUsable(t *testing.T) {
	tests := []struct {
		name string
		snap core.UsageSnapshot
		want bool
	}{
		{
			name: "StatusAuth with no metrics",
			snap: core.UsageSnapshot{Status: core.StatusAuth},
			want: false,
		},
		{
			name: "StatusAuth with metrics",
			snap: core.UsageSnapshot{
				Status:  core.StatusAuth,
				Metrics: map[string]core.Metric{"cost": {Used: core.Float64Ptr(5)}},
			},
			want: true,
		},
		{
			name: "StatusError with no metrics",
			snap: core.UsageSnapshot{Status: core.StatusError},
			want: false,
		},
		{
			name: "StatusError with metrics",
			snap: core.UsageSnapshot{
				Status:  core.StatusError,
				Metrics: map[string]core.Metric{"cost": {Used: core.Float64Ptr(5)}},
			},
			want: true,
		},
		{
			name: "StatusOK with no metrics",
			snap: core.UsageSnapshot{Status: core.StatusOK},
			want: true,
		},
		{
			name: "StatusLimited with no metrics",
			snap: core.UsageSnapshot{Status: core.StatusLimited},
			want: true,
		},
		{
			name: "StatusUnknown with no metrics",
			snap: core.UsageSnapshot{Status: core.StatusUnknown},
			want: false,
		},
		{
			name: "StatusUnknown with metrics",
			snap: core.UsageSnapshot{
				Status:  core.StatusUnknown,
				Metrics: map[string]core.Metric{"tokens": {Used: core.Float64Ptr(100)}},
			},
			want: true,
		},
		{
			name: "Empty status with no metrics",
			snap: core.UsageSnapshot{},
			want: false,
		},
		{
			name: "Empty status with metrics",
			snap: core.UsageSnapshot{
				Metrics: map[string]core.Metric{"tokens": {Used: core.Float64Ptr(100)}},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SnapshotUsable(tt.snap); got != tt.want {
				t.Errorf("SnapshotUsable(%s) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestAccountMatchesProvider(t *testing.T) {
	tests := []struct {
		name       string
		providerID string
		accountID  string
		snap       core.UsageSnapshot
		want       bool
	}{
		{
			name:       "snap provider ID matches",
			providerID: "claude",
			accountID:  "custom-name",
			snap:       core.UsageSnapshot{ProviderID: "claude"},
			want:       true,
		},
		{
			name:       "account ID equals provider ID",
			providerID: "claude",
			accountID:  "claude",
			snap:       core.UsageSnapshot{},
			want:       true,
		},
		{
			name:       "account ID has prefix providerID-",
			providerID: "claude",
			accountID:  "claude-work",
			snap:       core.UsageSnapshot{},
			want:       true,
		},
		{
			name:       "account ID does not match",
			providerID: "claude",
			accountID:  "openai-main",
			snap:       core.UsageSnapshot{ProviderID: "openai"},
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := accountMatchesProvider(tt.providerID, tt.accountID, tt.snap); got != tt.want {
				t.Errorf("accountMatchesProvider(%s) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestEnrichSnapshotsWithFetch(t *testing.T) {
	t.Run("nil fetch function returns immediately", func(t *testing.T) {
		snaps := map[string]core.UsageSnapshot{"test": {}}
		EnrichSnapshotsWithFetch(context.Background(), "prov", nil, nil, snaps, nil)
	})

	t.Run("empty snaps map returns immediately", func(t *testing.T) {
		fetch := func(_ context.Context, _ core.AccountConfig) (core.UsageSnapshot, error) {
			return core.UsageSnapshot{}, nil
		}
		EnrichSnapshotsWithFetch(context.Background(), "prov", fetch, nil, map[string]core.UsageSnapshot{}, nil)
	})

	t.Run("no matching snapshots returns immediately", func(t *testing.T) {
		fetchCalled := false
		fetch := func(_ context.Context, _ core.AccountConfig) (core.UsageSnapshot, error) {
			fetchCalled = true
			return core.UsageSnapshot{}, nil
		}
		snaps := map[string]core.UsageSnapshot{"other-1": {ProviderID: "other", AccountID: "other-1"}}
		EnrichSnapshotsWithFetch(context.Background(), "prov", fetch, nil, snaps, nil)
		if fetchCalled {
			t.Error("fetch should not be called when no snapshots match provider")
		}
	})

	t.Run("context cancelled skips results", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		fetch := func(_ context.Context, _ core.AccountConfig) (core.UsageSnapshot, error) {
			return core.UsageSnapshot{Status: core.StatusOK}, nil
		}
		snaps := map[string]core.UsageSnapshot{
			"prov-1": {ProviderID: "prov", AccountID: "prov-1", Status: core.StatusUnknown},
		}
		EnrichSnapshotsWithFetch(ctx, "prov", fetch, nil, snaps, nil)
		if snaps["prov-1"].Status != core.StatusUnknown {
			t.Errorf("expected snapshot to remain unchanged on canceled context, got %v", snaps["prov-1"].Status)
		}
	})

	t.Run("fetch error or unusable snapshot skips update", func(t *testing.T) {
		fetch := func(_ context.Context, acct core.AccountConfig) (core.UsageSnapshot, error) {
			if acct.ID == "prov-err" {
				return core.UsageSnapshot{}, errors.New("network fail")
			}
			return core.UsageSnapshot{Status: core.StatusUnknown}, nil // unusable
		}
		snaps := map[string]core.UsageSnapshot{
			"prov-err":      {ProviderID: "prov", AccountID: "prov-err", Message: "initial"},
			"prov-unusable": {ProviderID: "prov", AccountID: "prov-unusable", Message: "initial"},
		}
		EnrichSnapshotsWithFetch(context.Background(), "prov", fetch, nil, snaps, nil)
		if snaps["prov-err"].Message != "initial" || snaps["prov-unusable"].Message != "initial" {
			t.Error("snapshots should remain unchanged when fetch errors or produces unusable snapshot")
		}
	})

	t.Run("successful fetch with custom merge", func(t *testing.T) {
		accts := []core.AccountConfig{
			{ID: "prov-main", Provider: "prov"},
		}
		snaps := map[string]core.UsageSnapshot{
			"prov-main": {
				ProviderID: "prov",
				AccountID:  "prov-main",
				Status:     core.StatusUnknown,
				Message:    "base_msg",
			},
		}
		fetch := func(_ context.Context, acct core.AccountConfig) (core.UsageSnapshot, error) {
			snap := core.NewUsageSnapshot(acct.Provider, acct.ID)
			snap.Status = core.StatusOK
			snap.Message = "live_msg"
			return snap, nil
		}
		merge := func(base, fresh core.UsageSnapshot) core.UsageSnapshot {
			base.Status = fresh.Status
			base.Message = base.Message + "+" + fresh.Message
			return base
		}

		EnrichSnapshotsWithFetch(context.Background(), "prov", fetch, accts, snaps, merge)
		got := snaps["prov-main"]
		if got.Status != core.StatusOK {
			t.Errorf("Status = %v, want StatusOK", got.Status)
		}
		if got.Message != "base_msg+live_msg" {
			t.Errorf("Message = %q, want base_msg+live_msg", got.Message)
		}
	})
}

func TestOverlayLiveFetch(t *testing.T) {
	t.Run("unusable fresh snapshot returns base", func(t *testing.T) {
		base := core.UsageSnapshot{ProviderID: "prov", Message: "base"}
		fresh := core.UsageSnapshot{Status: core.StatusUnknown} // unusable
		got := OverlayLiveFetch(base, fresh)
		if got.Message != "base" {
			t.Errorf("expected base to be returned unchanged, got %+v", got)
		}
	})

	t.Run("all fresh fields overlaid", func(t *testing.T) {
		now := time.Now()
		base := core.UsageSnapshot{
			ProviderID:  "prov",
			AccountID:   "acct-1",
			Metrics:     map[string]core.Metric{"old_m": {Used: core.Float64Ptr(1)}},
			Resets:      map[string]time.Time{"old_r": now.Add(-time.Hour)},
			Attributes:  map[string]string{"old_a": "1"},
			Diagnostics: map[string]string{"old_d": "1"},
			Raw:         map[string]string{"old_raw": "1"},
		}

		freshNow := now.Add(time.Hour)
		fresh := core.UsageSnapshot{
			ProviderID:  "prov",
			AccountID:   "acct-1",
			Timestamp:   freshNow,
			Status:      core.StatusOK,
			Message:     "all systems go",
			Metrics:     map[string]core.Metric{"new_m": {Used: core.Float64Ptr(2)}},
			Resets:      map[string]time.Time{"new_r": freshNow},
			Attributes:  map[string]string{"new_a": "2"},
			Diagnostics: map[string]string{"new_d": "2"},
			Raw:         map[string]string{"new_raw": "2"},
		}

		got := OverlayLiveFetch(base, fresh)
		if !got.Timestamp.Equal(freshNow) {
			t.Errorf("Timestamp = %v, want %v", got.Timestamp, freshNow)
		}
		if got.Status != core.StatusOK {
			t.Errorf("Status = %v, want StatusOK", got.Status)
		}
		if got.Message != "all systems go" {
			t.Errorf("Message = %q, want 'all systems go'", got.Message)
		}
		if len(got.Metrics) != 2 || len(got.Resets) != 2 || len(got.Attributes) != 2 || len(got.Diagnostics) != 2 || len(got.Raw) != 2 {
			t.Errorf("expected maps merged, got %+v", got)
		}
	})
}
