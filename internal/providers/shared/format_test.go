package shared

import (
	"sync"
	"testing"
)

func TestFormatTokenCount(t *testing.T) {
	tests := []struct {
		name  string
		value int
		want  string
	}{
		{name: "zero", value: 0, want: "0"},
		{name: "small int", value: 42, want: "42"},
		{name: "just below 1K", value: 999, want: "999"},
		{name: "exact 1K", value: 1000, want: "1.0K"},
		{name: "fractional K", value: 1500, want: "1.5K"},
		{name: "large K", value: 999499, want: "999.5K"},
		{name: "just below 1M", value: 999999, want: "1000.0K"},
		{name: "exact 1M", value: 1_000_000, want: "1.0M"},
		{name: "fractional M", value: 2_345_678, want: "2.3M"},
		{name: "exact 1B", value: 1_000_000_000, want: "1.0B"},
		{name: "large B", value: 12_500_000_000, want: "12.5B"},
		{name: "negative value", value: -50, want: "-50"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatTokenCount(tt.value); got != tt.want {
				t.Errorf("FormatTokenCount(%d) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestFormatTokenCountF(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		want  string
	}{
		{name: "zero", value: 0.0, want: "0"},
		{name: "small float", value: 42.4, want: "42"},
		{name: "small float rounded", value: 42.6, want: "43"},
		{name: "just below 1K", value: 999.0, want: "999"},
		{name: "exact 1K", value: 1000.0, want: "1.0K"},
		{name: "fractional K", value: 1540.0, want: "1.5K"},
		{name: "exact 1M", value: 1_000_000.0, want: "1.0M"},
		{name: "fractional M", value: 2_350_000.0, want: "2.4M"},
		{name: "exact 1B", value: 1_000_000_000.0, want: "1.0B"},
		{name: "fractional B", value: 5_780_000_000.0, want: "5.8B"},
		{name: "negative float", value: -12.3, want: "-12"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatTokenCountF(tt.value); got != tt.want {
				t.Errorf("FormatTokenCountF(%f) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{name: "empty string with maxLen 0", input: "", maxLen: 0, want: ""},
		{name: "empty string with maxLen 5", input: "", maxLen: 5, want: ""},
		{name: "string shorter than maxLen", input: "hello", maxLen: 10, want: "hello"},
		{name: "string equal to maxLen", input: "hello", maxLen: 5, want: "hello"},
		{name: "string longer than maxLen", input: "hello world", maxLen: 8, want: "hello w…"},
		{name: "maxLen is 1", input: "hello", maxLen: 1, want: "…"},
		{name: "maxLen is 0", input: "hello", maxLen: 0, want: "…"},
		{name: "maxLen is negative", input: "hello", maxLen: -2, want: "…"},
		{name: "unicode multi-byte runes within limit", input: "👋🌍✨🚀", maxLen: 4, want: "👋🌍✨🚀"},
		{name: "unicode multi-byte runes truncated", input: "👋🌍✨🚀🎉", maxLen: 4, want: "👋🌍✨…"},
		{name: "mixed ASCII and CJK", input: "Hello世界你好", maxLen: 8, want: "Hello世界…"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Truncate(tt.input, tt.maxLen); got != tt.want {
				t.Errorf("Truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestFormat_Concurrency(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_ = FormatTokenCount(n * 1000)
			_ = FormatTokenCountF(float64(n) * 1.5e6)
			_ = Truncate("concurrent formatting string test", 10)
		}(i)
	}
	wg.Wait()
}
