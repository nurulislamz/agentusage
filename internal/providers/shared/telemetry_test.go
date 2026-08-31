package shared

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestTelemetryCollectOptions(t *testing.T) {
	t.Run("Path with nil Paths", func(t *testing.T) {
		opts := TelemetryCollectOptions{}
		if got := opts.Path("key", " default "); got != "default" {
			t.Errorf("Path() = %q, want %q", got, "default")
		}
	})

	t.Run("Path with present key", func(t *testing.T) {
		opts := TelemetryCollectOptions{
			Paths: map[string]string{
				"custom": " /path/to/custom ",
				"empty":  "   ",
			},
		}
		if got := opts.Path("custom", "/fallback"); got != "/path/to/custom" {
			t.Errorf("Path(custom) = %q, want %q", got, "/path/to/custom")
		}
		if got := opts.Path("empty", " /fallback "); got != "/fallback" {
			t.Errorf("Path(empty) = %q, want %q", got, "/fallback")
		}
		if got := opts.Path("missing", "/fallback"); got != "/fallback" {
			t.Errorf("Path(missing) = %q, want %q", got, "/fallback")
		}
	})

	t.Run("PathsFor with nil PathLists", func(t *testing.T) {
		opts := TelemetryCollectOptions{}
		fb := []string{"/a", "/b"}
		if got := opts.PathsFor("key", fb); len(got) != 2 || got[0] != "/a" {
			t.Errorf("PathsFor() = %v, want %v", got, fb)
		}
	})

	t.Run("PathsFor with present and empty keys", func(t *testing.T) {
		opts := TelemetryCollectOptions{
			PathLists: map[string][]string{
				"items": {"/x", "/y"},
				"empty": {},
			},
		}
		fb := []string{"/fallback"}
		if got := opts.PathsFor("items", fb); len(got) != 2 || got[0] != "/x" {
			t.Errorf("PathsFor(items) = %v, want [/x /y]", got)
		}
		if got := opts.PathsFor("empty", fb); len(got) != 1 || got[0] != "/fallback" {
			t.Errorf("PathsFor(empty) = %v, want %v", got, fb)
		}
		if got := opts.PathsFor("missing", fb); len(got) != 1 || got[0] != "/fallback" {
			t.Errorf("PathsFor(missing) = %v, want %v", got, fb)
		}
	})
}

func TestParseTimestampString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		wantY   int
		wantM   time.Month
		wantD   int
	}{
		{
			name:    "RFC3339Nano",
			input:   "2026-08-31T20:09:18.123456789Z",
			wantErr: false,
			wantY:   2026, wantM: time.August, wantD: 31,
		},
		{
			name:    "RFC3339",
			input:   "2026-08-31T20:09:18Z",
			wantErr: false,
			wantY:   2026, wantM: time.August, wantD: 31,
		},
		{
			name:    "ISO millis Z",
			input:   "2026-08-31T20:09:18.123Z",
			wantErr: false,
			wantY:   2026, wantM: time.August, wantD: 31,
		},
		{
			name:    "ISO seconds Z",
			input:   "2026-08-31T20:09:18Z",
			wantErr: false,
			wantY:   2026, wantM: time.August, wantD: 31,
		},
		{
			name:    "Datetime space separated",
			input:   "2026-08-31 20:09:18",
			wantErr: false,
			wantY:   2026, wantM: time.August, wantD: 31,
		},
		{
			name:    "Date only",
			input:   "2026-08-31",
			wantErr: false,
			wantY:   2026, wantM: time.August, wantD: 31,
		},
		{
			name:    "Unix timestamp seconds string",
			input:   "1756670958",
			wantErr: false,
			wantY:   2025,
		},
		{
			name:    "Unix timestamp millis string",
			input:   "1756670958000",
			wantErr: false,
			wantY:   2025,
		},
		{
			name:    "Unix timestamp micros string",
			input:   "1756670958000000",
			wantErr: false,
			wantY:   2025,
		},
		{
			name:    "Invalid format",
			input:   "not-a-timestamp",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTimestampString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseTimestampString(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr {
				if got.Year() != tt.wantY {
					t.Errorf("Year = %d, want %d", got.Year(), tt.wantY)
				}
				if tt.wantM != 0 && got.Month() != tt.wantM {
					t.Errorf("Month = %v, want %v", got.Month(), tt.wantM)
				}
				if tt.wantD != 0 && got.Day() != tt.wantD {
					t.Errorf("Day = %d, want %d", got.Day(), tt.wantD)
				}
			}
		})
	}
}

func TestFlexParseTime(t *testing.T) {
	t.Run("valid timestamp", func(t *testing.T) {
		got := FlexParseTime("2026-01-15T12:00:00Z")
		if got.IsZero() || got.Year() != 2026 {
			t.Errorf("FlexParseTime() = %v, want year 2026", got)
		}
	})

	t.Run("invalid timestamp returns zero time", func(t *testing.T) {
		got := FlexParseTime("invalid")
		if !got.IsZero() {
			t.Errorf("FlexParseTime(invalid) = %v, want zero time", got)
		}
	})
}

func TestUnixAuto(t *testing.T) {
	t.Run("seconds", func(t *testing.T) {
		ts := int64(1700000000)
		got := UnixAuto(ts)
		if got.Unix() != ts {
			t.Errorf("UnixAuto(seconds) = %d, want %d", got.Unix(), ts)
		}
	})

	t.Run("milliseconds", func(t *testing.T) {
		ts := int64(1700000000000)
		got := UnixAuto(ts)
		if got.Unix() != 1700000000 {
			t.Errorf("UnixAuto(milli) = %d, want 1700000000", got.Unix())
		}
	})

	t.Run("microseconds", func(t *testing.T) {
		ts := int64(1700000000000000)
		got := UnixAuto(ts)
		if got.Unix() != 1700000000 {
			t.Errorf("UnixAuto(micro) = %d, want 1700000000", got.Unix())
		}
	})
}

func TestParseFlexibleTimestamp(t *testing.T) {
	t.Run("empty string", func(t *testing.T) {
		ts, ok := ParseFlexibleTimestamp("   ")
		if ok || ts != 0 {
			t.Fatalf("ParseFlexibleTimestamp(empty) = (%d, %v), want (0, false)", ts, ok)
		}
	})

	t.Run("valid timestamp", func(t *testing.T) {
		ts, ok := ParseFlexibleTimestamp("2026-01-01T00:00:00Z")
		if !ok || ts == 0 {
			t.Fatalf("ParseFlexibleTimestamp() = (%d, %v), want valid unix ts", ts, ok)
		}
	})

	t.Run("invalid string", func(t *testing.T) {
		ts, ok := ParseFlexibleTimestamp("unparseable")
		if ok || ts != 0 {
			t.Fatalf("ParseFlexibleTimestamp(invalid) = (%d, %v), want (0, false)", ts, ok)
		}
	})
}

func TestSanitizeWorkspace(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty", input: "", want: ""},
		{name: "dot", input: ".", want: "."},
		{name: "root slash", input: "/", want: "/"},
		{name: "unix path", input: "/home/user/projects/my-app", want: "my-app"},
		{name: "trailing slash", input: "/home/user/projects/my-app/", want: "my-app"},
		{name: "whitespace padded", input: "  /var/log/syslog  ", want: "syslog"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SanitizeWorkspace(tt.input); got != tt.want {
				t.Errorf("SanitizeWorkspace(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestExpandHome(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty", input: "", want: ""},
		{name: "tilde only", input: "~", want: home},
		{name: "tilde with slash", input: "~/test/dir", want: filepath.Join(home, "test/dir")},
		{name: "absolute path", input: "/tmp/data", want: "/tmp/data"},
		{name: "relative path", input: "foo/bar", want: "foo/bar"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExpandHome(tt.input); got != tt.want {
				t.Errorf("ExpandHome(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCollectFilesByExt(t *testing.T) {
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "a.json")
	file2 := filepath.Join(tmpDir, "b.JSONL")
	file3 := filepath.Join(tmpDir, "c.txt")
	subDir := filepath.Join(tmpDir, "sub")
	file4 := filepath.Join(subDir, "d.json")

	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{file1, file2, file3, file4} {
		if err := os.WriteFile(f, []byte("data"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	exts := map[string]bool{
		".json":  true,
		".jsonl": true,
	}

	t.Run("collect from directory root", func(t *testing.T) {
		files, err := CollectFilesByExt([]string{tmpDir}, exts)
		if err != nil {
			t.Fatalf("CollectFilesByExt() error = %v", err)
		}
		// file1 (.json), file2 (.JSONL case-insensitive), file4 (.json)
		if len(files) != 3 {
			t.Fatalf("len(files) = %d, want 3: %v", len(files), files)
		}
	})

	t.Run("collect from individual file root matching", func(t *testing.T) {
		files, err := CollectFilesByExt([]string{file1}, exts)
		if err != nil {
			t.Fatalf("CollectFilesByExt() error = %v", err)
		}
		if len(files) != 1 || files[0] != file1 {
			t.Fatalf("files = %v, want [%s]", files, file1)
		}
	})

	t.Run("collect from individual file root not matching", func(t *testing.T) {
		files, err := CollectFilesByExt([]string{file3}, exts)
		if err != nil {
			t.Fatalf("CollectFilesByExt() error = %v", err)
		}
		if len(files) != 0 {
			t.Fatalf("files = %v, want []", files)
		}
	})

	t.Run("non-existent path skipped", func(t *testing.T) {
		files, err := CollectFilesByExt([]string{filepath.Join(tmpDir, "missing")}, exts)
		if err != nil {
			t.Fatalf("CollectFilesByExt() error = %v", err)
		}
		if len(files) != 0 {
			t.Fatalf("files = %v, want []", files)
		}
	})

	t.Run("empty roots", func(t *testing.T) {
		files, err := CollectFilesByExt([]string{"", "   "}, exts)
		if err != nil {
			t.Fatalf("CollectFilesByExt() error = %v", err)
		}
		if len(files) != 0 {
			t.Fatalf("files = %v, want []", files)
		}
	})
}

func TestCollectFilesWithStat(t *testing.T) {
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "a.json")
	file2 := filepath.Join(tmpDir, "b.txt")
	subDir := filepath.Join(tmpDir, "sub")
	file3 := filepath.Join(subDir, "c.json")

	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{file1, file2, file3} {
		if err := os.WriteFile(f, []byte("data"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	exts := map[string]bool{".json": true}

	t.Run("collects files and stats", func(t *testing.T) {
		result, err := CollectFilesWithStat([]string{tmpDir}, exts)
		if err != nil {
			t.Fatalf("CollectFilesWithStat() error = %v", err)
		}
		if len(result) != 2 {
			t.Fatalf("len(result) = %d, want 2", len(result))
		}
		if _, ok := result[file1]; !ok {
			t.Errorf("missing file1 in result: %v", result)
		}
		if _, ok := result[file3]; !ok {
			t.Errorf("missing file3 in result: %v", result)
		}
	})

	t.Run("collects matching individual file", func(t *testing.T) {
		result, err := CollectFilesWithStat([]string{file1}, exts)
		if err != nil {
			t.Fatalf("CollectFilesWithStat() error = %v", err)
		}
		if len(result) != 1 || result[file1] == nil {
			t.Fatalf("result = %v, want map with file1", result)
		}
	})

	t.Run("skips non-matching individual file", func(t *testing.T) {
		result, err := CollectFilesWithStat([]string{file2}, exts)
		if err != nil {
			t.Fatalf("CollectFilesWithStat() error = %v", err)
		}
		if len(result) != 0 {
			t.Fatalf("result = %v, want empty", result)
		}
	})

	t.Run("skips non-existent path", func(t *testing.T) {
		result, err := CollectFilesWithStat([]string{filepath.Join(tmpDir, "missing")}, exts)
		if err != nil {
			t.Fatalf("CollectFilesWithStat() error = %v", err)
		}
		if len(result) != 0 {
			t.Fatalf("result = %v, want empty", result)
		}
	})

	t.Run("empty roots", func(t *testing.T) {
		result, err := CollectFilesWithStat([]string{""}, exts)
		if err != nil {
			t.Fatalf("CollectFilesWithStat() error = %v", err)
		}
		if len(result) != 0 {
			t.Fatalf("result = %v, want empty", result)
		}
	})
}

func TestExtractFilePathsFromPayload(t *testing.T) {
	payload := map[string]any{
		"file_path": "/home/user/src/main.go",
		"target":    "pkg/utils/helper.go",
		"cwd":       "/home/user/src",
		"params": map[string]any{
			"paths": []any{
				"./config/settings.json",
				"`internal/service.go`",
				"\"app.ts\"",
				"https://example.com/ignored.go",
				"-v",
				"simpleword", // no slash, backslash or dot -> ignored
			},
		},
		"unhinted_key": "some/ignored/path.go",
	}

	got := ExtractFilePathsFromPayload(payload)
	expected := []string{
		"/home/user/src",
		"/home/user/src/main.go",
		"app.ts",
		"config/settings.json",
		"internal/service.go",
		"pkg/utils/helper.go",
	}

	if len(got) != len(expected) {
		t.Fatalf("ExtractFilePathsFromPayload() returned %v (len %d), want %v (len %d)", got, len(got), expected, len(expected))
	}
	for i, exp := range expected {
		if got[i] != exp {
			t.Errorf("got[%d] = %q, want %q", i, got[i], exp)
		}
	}
}

func TestExtractPathTokens_EdgeCases(t *testing.T) {
	t.Run("empty string", func(t *testing.T) {
		if got := extractPathTokens("   "); got != nil {
			t.Errorf("extractPathTokens(empty) = %v, want nil", got)
		}
	})

	t.Run("urls and flags filtered", func(t *testing.T) {
		raw := "http://test.com https://test.com file://foo/bar --verbose -flag /valid/path.go"
		got := extractPathTokens(raw)
		if len(got) != 1 || got[0] != "/valid/path.go" {
			t.Errorf("extractPathTokens() = %v, want [/valid/path.go]", got)
		}
	})

	t.Run("trimmed quotes and punctuation", func(t *testing.T) {
		raw := "('path/to/a.txt', [\"path/to/b.py\"]) <path/to/c.rs>:"
		got := extractPathTokens(raw)
		if len(got) != 3 {
			t.Fatalf("len(got) = %d, want 3: %v", len(got), got)
		}
	})
}

func TestTelemetry_Concurrency(t *testing.T) {
	payload := map[string]any{
		"file_path": "/var/log/app.log",
		"paths":     []any{"/a/b.go", "/c/d.rs"},
	}

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = ExtractFilePathsFromPayload(payload)
			_ = SanitizeWorkspace("/a/b/c")
			_ = FlexParseTime("2026-08-31T20:00:00Z")
		}()
	}
	wg.Wait()
}
