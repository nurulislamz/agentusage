package daemon

import (
	"context"
	"time"

	"github.com/nurulislamz/agentusage/internal/core"
)

func (s *Service) enrichReadModelSnapshots(
	ctx context.Context,
	accounts []core.AccountConfig,
	modelNorm core.ModelNormalizationConfig,
	snaps map[string]core.UsageSnapshot,
) map[string]core.UsageSnapshot {
	if s == nil || len(snaps) == 0 {
		return snaps
	}
	out := make(map[string]core.UsageSnapshot, len(snaps))
	for id, snap := range snaps {
		out[id] = snap
	}
	out = s.overlayPollStateSnapshots(out)
	out = s.hydrateUnknownLocalSnapshots(ctx, accounts, modelNorm, out)
	return out
}

func (s *Service) overlayPollStateSnapshots(snaps map[string]core.UsageSnapshot) map[string]core.UsageSnapshot {
	s.pollStateMu.Lock()
	defer s.pollStateMu.Unlock()

	for accountID, snap := range snaps {
		state := s.pollState[accountID]
		if state == nil || !state.hasSnap {
			continue
		}
		polled := state.lastSnap
		if !snapshotMoreUsableThan(polled, snap) {
			continue
		}
		snaps[accountID] = polled
	}
	return snaps
}

func snapshotMoreUsableThan(candidate, current core.UsageSnapshot) bool {
	if !snapshotHasUsableData(candidate) {
		return false
	}
	if !snapshotHasUsableData(current) {
		return true
	}
	// A newer metric-less StatusOK (e.g. Cursor cli-config email fallback)
	// must not replace a hydrated snapshot that still carries quota Metrics/
	// Resets. Auth/Error/Limited may still surface without quota signals.
	if snapshotHasQuotaSignals(current) && !snapshotHasQuotaSignals(candidate) && !snapshotStatusMayReplaceQuota(candidate.Status) {
		return false
	}
	if candidate.Timestamp.After(current.Timestamp) {
		return true
	}
	if current.Status == core.StatusUnknown && candidate.Status != core.StatusUnknown {
		return true
	}
	return false
}

func snapshotHasUsableData(snap core.UsageSnapshot) bool {
	if snap.Status != "" && snap.Status != core.StatusUnknown {
		return true
	}
	if snapshotHasQuotaSignals(snap) {
		return true
	}
	return false
}

func snapshotHasQuotaSignals(snap core.UsageSnapshot) bool {
	return len(snap.Metrics) > 0 || len(snap.Resets) > 0
}

func snapshotStatusMayReplaceQuota(status core.Status) bool {
	switch status {
	case core.StatusAuth, core.StatusError, core.StatusLimited:
		return true
	default:
		return false
	}
}

func (s *Service) hydrateUnknownLocalSnapshots(
	ctx context.Context,
	accounts []core.AccountConfig,
	modelNorm core.ModelNormalizationConfig,
	snaps map[string]core.UsageSnapshot,
) map[string]core.UsageSnapshot {
	for _, acct := range accounts {
		current, ok := snaps[acct.ID]
		if !ok || snapshotHasUsableData(current) {
			continue
		}
		provider, ok := s.providerByID[acct.Provider]
		if !ok {
			continue
		}
		if _, ok := provider.(core.ChangeDetector); !ok {
			continue
		}

		fetchCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		fresh, err := provider.Fetch(fetchCtx, acct)
		cancel()
		if err != nil {
			continue
		}
		fresh = core.NormalizeUsageSnapshotWithConfig(fresh, modelNorm)
		if snapshotHasUsableData(fresh) {
			snaps[acct.ID] = fresh
		}
	}
	return snaps
}
