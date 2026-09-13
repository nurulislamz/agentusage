package core

import "strings"

const QuotaNearLimitPercent = 15.0

// EffectiveStatus returns the status that should be shown for a snapshot.
// Explicit provider status (such as OK, LIMITED, AUTH) is preserved.
// When telemetry has quota metrics but no persisted status, infer health
// from the tightest quota bucket.
func EffectiveStatus(snap UsageSnapshot) Status {
	if snap.Status != "" && snap.Status != StatusUnknown {
		return snap.Status
	}
	if status, ok := InferStatusFromQuotaMetrics(snap); ok {
		return status
	}
	return snap.Status
}

func isQuotaOrLimitMetric(key string, met Metric) bool {
	k := strings.ToLower(key)
	if strings.Contains(k, "success_rate") || strings.Contains(k, "hit_ratio") ||
		strings.Contains(k, "percentage_tracked") || strings.Contains(k, "ai_code_percentage") {
		return false
	}
	if strings.HasPrefix(k, "quota") || strings.HasPrefix(k, "rate_limit") ||
		strings.HasPrefix(k, "plan_") || strings.HasPrefix(k, "cursor_plan") {
		return true
	}
	if k == "rpm" || k == "tpm" || k == "rpd" || k == "tpd" {
		return true
	}
	if strings.Contains(k, "subscription") || strings.Contains(k, "credits") ||
		strings.Contains(k, "budget") || strings.Contains(k, "limit") {
		return true
	}
	if strings.Contains(k, "usage") && (strings.Contains(k, "weekly") || strings.Contains(k, "monthly") ||
		strings.Contains(k, "five_hour") || strings.Contains(k, "5h") || strings.Contains(k, "rolling")) {
		return true
	}
	if met.Window != "" && (met.Remaining != nil || (met.Limit != nil && *met.Limit > 0)) {
		return true
	}
	return false
}

func InferStatusFromQuotaMetrics(snap UsageSnapshot) (Status, bool) {
	worst := -1.0
	found := false
	for key, met := range snap.Metrics {
		if !isQuotaOrLimitMetric(key, met) {
			continue
		}
		rem := -1.0
		if met.Remaining != nil {
			if met.Limit != nil && *met.Limit > 0 {
				rem = (*met.Remaining / *met.Limit) * 100
			} else {
				rem = *met.Remaining
			}
		} else if met.Used != nil {
			if met.Limit != nil && *met.Limit > 0 {
				rem = (1.0 - (*met.Used / *met.Limit)) * 100
			} else if met.Unit == "%" || met.Unit == "percent" {
				rem = 100 - *met.Used
			}
		} else if pct := met.Percent(); pct >= 0 {
			rem = pct
		}
		if rem < 0 {
			rem = 0
		}
		if rem > 100 {
			rem = 100
		}
		if !found || rem < worst {
			worst = rem
			found = true
		}
	}
	if !found {
		return "", false
	}
	if worst <= 0.5 {
		return StatusLimited, true
	}
	if worst < QuotaNearLimitPercent {
		return StatusNearLimit, true
	}
	return StatusOK, true
}
