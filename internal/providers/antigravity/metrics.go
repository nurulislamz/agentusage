package antigravity

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type pingMetricsState struct {
	mu           sync.RWMutex
	pingsByLabel map[string]int64
	durations    map[string]float64
	lastPingTime map[string]int64
	lastStatus   map[string]int
}

var pingMetrics = &pingMetricsState{
	pingsByLabel: make(map[string]int64),
	durations:    make(map[string]float64),
	lastPingTime: make(map[string]int64),
	lastStatus:   make(map[string]int),
}

// RecordBoxPing records metrics and emits structured logging for an agy-box ping.
func RecordBoxPing(box, accountID, reason string, duration time.Duration, err error) {
	if box == "" {
		box = "default"
	}
	if accountID == "" {
		accountID = "antigravity"
	}
	if reason == "" {
		reason = "unknown"
	}
	status := "success"
	statusVal := 1
	if err != nil {
		status = "error"
		statusVal = 0
	}

	pingMetrics.mu.Lock()
	labelKey := fmt.Sprintf("%s|%s|%s|%s", box, accountID, reason, status)
	pingMetrics.pingsByLabel[labelKey]++
	pingMetrics.durations[box] = duration.Seconds()
	nowUnix := time.Now().UTC().Unix()
	pingMetrics.lastPingTime[box] = nowUnix
	pingMetrics.lastStatus[box] = statusVal
	pingMetrics.mu.Unlock()

	// Console / daemon log
	if err != nil {
		log.Printf("[antigravity] agy-box ping box=%s account=%s reason=%s duration=%s status=error err=%v",
			box, accountID, reason, duration.Round(time.Millisecond), err)
	} else {
		log.Printf("[antigravity] agy-box ping box=%s account=%s reason=%s duration=%s status=success",
			box, accountID, reason, duration.Round(time.Millisecond))
	}

	// Persistent log line in ~/.local/state/agentusage/agy-pings.log (synchronous to avoid test cleanup races)
	if home, _ := os.UserHomeDir(); home != "" {
		logDir := filepath.Join(home, ".local", "state", "agentusage")
		logFile := filepath.Join(logDir, "agy-pings.log")
		if err := os.MkdirAll(logDir, 0o755); err == nil {
			if f, oerr := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); oerr == nil {
				defer f.Close()
				errMsg := ""
				if err != nil {
					errMsg = err.Error()
				}
				line := fmt.Sprintf("%s box=%s account=%s reason=%s duration_ms=%d status=%s error=%q\n",
					time.Now().UTC().Format(time.RFC3339), box, accountID, reason, duration.Milliseconds(), status, errMsg)
				_, _ = f.WriteString(line)
			}
		}
	}
}

// WritePrometheusMetrics serializes recorded agy-box metrics to w in Prometheus format.
func WritePrometheusMetrics(w io.Writer) {
	pingMetrics.mu.RLock()
	defer pingMetrics.mu.RUnlock()

	fmt.Fprintln(w, "# HELP agy_box_pings_total Total count of agy-box ping invocations")
	fmt.Fprintln(w, "# TYPE agy_box_pings_total counter")
	keys := make([]string, 0, len(pingMetrics.pingsByLabel))
	for k := range pingMetrics.pingsByLabel {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		parts := strings.Split(k, "|")
		if len(parts) == 4 {
			fmt.Fprintf(w, "agy_box_pings_total{box=%q,account_id=%q,reason=%q,status=%q} %d\n",
				parts[0], parts[1], parts[2], parts[3], pingMetrics.pingsByLabel[k])
		}
	}

	fmt.Fprintln(w, "# HELP agy_box_ping_duration_seconds Duration of last agy-box ping in seconds")
	fmt.Fprintln(w, "# TYPE agy_box_ping_duration_seconds gauge")
	boxes := make([]string, 0, len(pingMetrics.durations))
	for b := range pingMetrics.durations {
		boxes = append(boxes, b)
	}
	sort.Strings(boxes)
	for _, b := range boxes {
		fmt.Fprintf(w, "agy_box_ping_duration_seconds{box=%q} %.4f\n", b, pingMetrics.durations[b])
	}

	fmt.Fprintln(w, "# HELP agy_box_last_ping_timestamp_seconds Unix timestamp of last agy-box ping")
	fmt.Fprintln(w, "# TYPE agy_box_last_ping_timestamp_seconds gauge")
	for _, b := range boxes {
		if ts, ok := pingMetrics.lastPingTime[b]; ok {
			fmt.Fprintf(w, "agy_box_last_ping_timestamp_seconds{box=%q} %d\n", b, ts)
		}
	}

	fmt.Fprintln(w, "# HELP agy_box_last_ping_status Status of last agy-box ping (1=success, 0=error)")
	fmt.Fprintln(w, "# TYPE agy_box_last_ping_status gauge")
	for _, b := range boxes {
		if st, ok := pingMetrics.lastStatus[b]; ok {
			fmt.Fprintf(w, "agy_box_last_ping_status{box=%q} %d\n", b, st)
		}
	}
}

// ResetMetricsForTesting clears metrics state for isolated tests.
func ResetMetricsForTesting() {
	pingMetrics.mu.Lock()
	defer pingMetrics.mu.Unlock()
	pingMetrics.pingsByLabel = make(map[string]int64)
	pingMetrics.durations = make(map[string]float64)
	pingMetrics.lastPingTime = make(map[string]int64)
	pingMetrics.lastStatus = make(map[string]int)
}
