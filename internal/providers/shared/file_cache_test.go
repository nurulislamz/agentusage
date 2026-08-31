package shared

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestStatSignature(t *testing.T) {
	t.Run("existing file returns valid signature", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "test.jsonl")
		content := []byte("hello world")
		if err := os.WriteFile(tmpFile, content, 0o600); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		sig, err := StatSignature(tmpFile)
		if err != nil {
			t.Fatalf("StatSignature failed: %v", err)
		}
		if sig.Size != int64(len(content)) {
			t.Errorf("sig.Size = %d, want %d", sig.Size, len(content))
		}
		if sig.ModTime.IsZero() {
			t.Errorf("sig.ModTime is zero")
		}
	})

	t.Run("non-existent file returns error", func(t *testing.T) {
		tmpDir := t.TempDir()
		nonExistent := filepath.Join(tmpDir, "does_not_exist.jsonl")

		sig, err := StatSignature(nonExistent)
		if err == nil {
			t.Fatalf("expected error for non-existent file, got nil")
		}
		if sig.Size != 0 || !sig.ModTime.IsZero() {
			t.Errorf("expected zero FileSignature, got %+v", sig)
		}
	})
}

func TestFileSignature_Equal(t *testing.T) {
	now := time.Now().Truncate(time.Millisecond)
	earlier := now.Add(-1 * time.Hour)

	tests := []struct {
		name string
		a    FileSignature
		b    FileSignature
		want bool
	}{
		{
			name: "identical size and modtime",
			a:    FileSignature{ModTime: now, Size: 100},
			b:    FileSignature{ModTime: now, Size: 100},
			want: true,
		},
		{
			name: "different size, same modtime",
			a:    FileSignature{ModTime: now, Size: 100},
			b:    FileSignature{ModTime: now, Size: 200},
			want: false,
		},
		{
			name: "same size, different modtime",
			a:    FileSignature{ModTime: now, Size: 100},
			b:    FileSignature{ModTime: earlier, Size: 100},
			want: false,
		},
		{
			name: "different size and modtime",
			a:    FileSignature{ModTime: now, Size: 100},
			b:    FileSignature{ModTime: earlier, Size: 50},
			want: false,
		},
		{
			name: "both zero structs",
			a:    FileSignature{},
			b:    FileSignature{},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Equal(tt.b); got != tt.want {
				t.Errorf("FileSignature.Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileSignature_Grew(t *testing.T) {
	t0 := time.Now().Truncate(time.Millisecond)
	tEarlier := t0.Add(-1 * time.Minute)
	tLater := t0.Add(1 * time.Minute)

	tests := []struct {
		name string
		a    FileSignature // cached signature
		b    FileSignature // new signature on disk
		want bool
	}{
		{
			name: "same size and same modtime",
			a:    FileSignature{ModTime: t0, Size: 100},
			b:    FileSignature{ModTime: t0, Size: 100},
			want: true,
		},
		{
			name: "larger size and same modtime",
			a:    FileSignature{ModTime: t0, Size: 100},
			b:    FileSignature{ModTime: t0, Size: 200},
			want: true,
		},
		{
			name: "larger size and later modtime",
			a:    FileSignature{ModTime: t0, Size: 100},
			b:    FileSignature{ModTime: tLater, Size: 200},
			want: true,
		},
		{
			name: "same size and later modtime",
			a:    FileSignature{ModTime: t0, Size: 100},
			b:    FileSignature{ModTime: tLater, Size: 100},
			want: true,
		},
		{
			name: "shrunk size and later modtime",
			a:    FileSignature{ModTime: t0, Size: 200},
			b:    FileSignature{ModTime: tLater, Size: 100},
			want: false,
		},
		{
			name: "larger size but earlier modtime (e.g. clock skew/revert)",
			a:    FileSignature{ModTime: t0, Size: 100},
			b:    FileSignature{ModTime: tEarlier, Size: 200},
			want: false,
		},
		{
			name: "shrunk size and earlier modtime",
			a:    FileSignature{ModTime: t0, Size: 200},
			b:    FileSignature{ModTime: tEarlier, Size: 100},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Grew(tt.b); got != tt.want {
				t.Errorf("FileSignature.Grew() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileSignature_Concurrency(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "concurrent.jsonl")
	if err := os.WriteFile(tmpFile, []byte("start\n"), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			sig, err := StatSignature(tmpFile)
			if err != nil {
				t.Errorf("concurrent StatSignature failed: %v", err)
				return
			}
			if !sig.Equal(sig) {
				t.Errorf("sig should be equal to itself")
			}
			if !sig.Grew(sig) {
				t.Errorf("sig should satisfy Grew against itself")
			}
		}(i)
	}
	wg.Wait()
}
