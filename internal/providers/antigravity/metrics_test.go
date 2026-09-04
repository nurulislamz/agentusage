package antigravity

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestRecordBoxPing_AndWritePrometheus(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	ResetMetricsForTesting()

	RecordBoxPing("chaos", "antigravity-chaos", "missing_token", 250*time.Millisecond, nil)
	RecordBoxPing("chaos", "antigravity-chaos", "token_expired", 500*time.Millisecond, nil)
	RecordBoxPing("missing", "antigravity-missing", "missing_token", 100*time.Millisecond, errors.New("exit status 1"))

	var buf bytes.Buffer
	WritePrometheusMetrics(&buf)
	out := buf.String()

	if !strings.Contains(out, "agy_box_pings_total{box=\"chaos\",account_id=\"antigravity-chaos\",reason=\"missing_token\",status=\"success\"} 1") {
		t.Errorf("missing expected chaos success metric: %s", out)
	}
	if !strings.Contains(out, "agy_box_pings_total{box=\"chaos\",account_id=\"antigravity-chaos\",reason=\"token_expired\",status=\"success\"} 1") {
		t.Errorf("missing expected chaos expired metric: %s", out)
	}
	if !strings.Contains(out, "agy_box_pings_total{box=\"missing\",account_id=\"antigravity-missing\",reason=\"missing_token\",status=\"error\"} 1") {
		t.Errorf("missing expected missing error metric: %s", out)
	}
	if !strings.Contains(out, "agy_box_last_ping_status{box=\"chaos\"} 1") {
		t.Errorf("expected chaos status 1: %s", out)
	}
	if !strings.Contains(out, "agy_box_last_ping_status{box=\"missing\"} 0") {
		t.Errorf("expected missing status 0: %s", out)
	}
}
