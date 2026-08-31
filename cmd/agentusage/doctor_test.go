package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestDoctorCommandExecution(t *testing.T) {
	var buf bytes.Buffer
	runDoctorDiagnostics(&buf, false)
	output := buf.String()

	if !strings.Contains(output, "agentUsage Doctor") {
		t.Errorf("expected header in doctor output, got:\n%s", output)
	}
	if !strings.Contains(output, "[ OK ] System:") {
		t.Errorf("expected system check in doctor output, got:\n%s", output)
	}
	if !strings.Contains(output, "Result:") {
		t.Errorf("expected result summary in doctor output, got:\n%s", output)
	}
}

func TestDoctorCheckerFormatting(t *testing.T) {
	var buf bytes.Buffer
	d := &doctorChecker{out: &buf}

	d.ok("check 1: %s", "passed")
	d.info("check 2: %s", "informational")
	d.warn("check 3: %s", "warning")
	d.fail("check 4: %s", "failed")

	if d.okCount != 1 {
		t.Errorf("expected okCount=1, got %d", d.okCount)
	}
	if d.warnCount != 1 {
		t.Errorf("expected warnCount=1, got %d", d.warnCount)
	}
	if d.failCount != 1 {
		t.Errorf("expected failCount=1, got %d", d.failCount)
	}

	out := buf.String()
	if !strings.Contains(out, "[ OK ] check 1: passed") {
		t.Errorf("missing [ OK ] format: %s", out)
	}
	if !strings.Contains(out, "[INFO] check 2: informational") {
		t.Errorf("missing [INFO] format: %s", out)
	}
	if !strings.Contains(out, "[WARN] check 3: warning") {
		t.Errorf("missing [WARN] format: %s", out)
	}
	if !strings.Contains(out, "[FAIL] check 4: failed") {
		t.Errorf("missing [FAIL] format: %s", out)
	}
}

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1024 * 1024, "1.0 MB"},
		{5 * 1024 * 1024, "5.0 MB"},
	}

	for _, tt := range tests {
		got := formatFileSize(tt.bytes)
		if got != tt.expected {
			t.Errorf("formatFileSize(%d) = %q, expected %q", tt.bytes, got, tt.expected)
		}
	}
}
