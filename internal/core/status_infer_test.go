package core

import "testing"

func TestInferStatusFromQuotaMetrics(t *testing.T) {
	okSnap := UsageSnapshot{
		Metrics: map[string]Metric{
			"quota_gemini_weekly": {Remaining: floatPtr(80)},
		},
	}
	if got, ok := InferStatusFromQuotaMetrics(okSnap); !ok || got != StatusOK {
		t.Fatalf("got %q ok=%v, want OK true", got, ok)
	}

	warnSnap := UsageSnapshot{
		Metrics: map[string]Metric{
			"quota_gemini_5h": {Remaining: floatPtr(10)},
		},
	}
	if got, ok := InferStatusFromQuotaMetrics(warnSnap); !ok || got != StatusNearLimit {
		t.Fatalf("got %q ok=%v, want NEAR_LIMIT true", got, ok)
	}

	limitedSnap := UsageSnapshot{
		Metrics: map[string]Metric{
			"quota_gemini_weekly": {Remaining: floatPtr(0)},
		},
	}
	if got, ok := InferStatusFromQuotaMetrics(limitedSnap); !ok || got != StatusLimited {
		t.Fatalf("got %q ok=%v, want LIMITED true", got, ok)
	}
}

func TestEffectiveStatus_PrefersExplicitStatus(t *testing.T) {
	snap := UsageSnapshot{
		Status: StatusAuth,
		Metrics: map[string]Metric{
			"quota_gemini_weekly": {Remaining: floatPtr(80)},
		},
	}
	if got := EffectiveStatus(snap); got != StatusAuth {
		t.Fatalf("got %q, want AUTH", got)
	}
}

func TestEffectiveStatus_InfersFromQuotaWhenUnknown(t *testing.T) {
	snap := UsageSnapshot{
		Status: StatusUnknown,
		Metrics: map[string]Metric{
			"quota_gemini_weekly": {Remaining: floatPtr(80)},
		},
	}
	if got := EffectiveStatus(snap); got != StatusOK {
		t.Fatalf("got %q, want OK", got)
	}
}

func TestEffectiveStatus_InfersFromQuotaWhenEmptyOrUnknown(t *testing.T) {
	snap := UsageSnapshot{
		Status: "",
		Metrics: map[string]Metric{
			"monthly_subscription": {
				Limit:     floatPtr(100),
				Used:      floatPtr(99.83),
				Remaining: floatPtr(0.17),
				Unit:      "percent",
				Window:    "month",
			},
		},
	}
	if got := EffectiveStatus(snap); got != StatusLimited {
		t.Fatalf("got %q, want LIMITED", got)
	}
}

func TestEffectiveStatus_IgnoresToolSuccessRate(t *testing.T) {
	snap := UsageSnapshot{
		Status: StatusOK,
		Metrics: map[string]Metric{
			"tool_success_rate": {Used: floatPtr(100), Unit: "%"},
		},
	}
	if got := EffectiveStatus(snap); got != StatusOK {
		t.Fatalf("got %q, want OK", got)
	}
}

func floatPtr(v float64) *float64 { return &v }


