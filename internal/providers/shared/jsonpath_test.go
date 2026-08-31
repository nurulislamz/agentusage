package shared

import (
	"encoding/json"
	"reflect"
	"sync"
	"testing"
)

func TestPathValue(t *testing.T) {
	data := map[string]any{
		"a": map[string]any{
			"b": map[string]any{
				"c": "target",
				"d": 123,
			},
			"non_map": "scalar",
		},
		"top": "value",
	}

	tests := []struct {
		name      string
		root      map[string]any
		path      []string
		wantValue any
		wantOk    bool
	}{
		{
			name:      "nil root",
			root:      nil,
			path:      []string{"a"},
			wantValue: nil,
			wantOk:    false,
		},
		{
			name:      "empty path",
			root:      data,
			path:      []string{},
			wantValue: data,
			wantOk:    true,
		},
		{
			name:      "single top-level key exists",
			root:      data,
			path:      []string{"top"},
			wantValue: "value",
			wantOk:    true,
		},
		{
			name:      "deeply nested key exists",
			root:      data,
			path:      []string{"a", "b", "c"},
			wantValue: "target",
			wantOk:    true,
		},
		{
			name:      "top level missing",
			root:      data,
			path:      []string{"missing"},
			wantValue: nil,
			wantOk:    false,
		},
		{
			name:      "nested key missing",
			root:      data,
			path:      []string{"a", "b", "missing"},
			wantValue: nil,
			wantOk:    false,
		},
		{
			name:      "intermediate node is not a map",
			root:      data,
			path:      []string{"a", "non_map", "child"},
			wantValue: nil,
			wantOk:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := PathValue(tt.root, tt.path...)
			if ok != tt.wantOk {
				t.Fatalf("PathValue() ok = %v, want %v", ok, tt.wantOk)
			}
			if ok && !reflect.DeepEqual(got, tt.wantValue) {
				t.Errorf("PathValue() = %v, want %v", got, tt.wantValue)
			}
		})
	}
}

func TestPathMap(t *testing.T) {
	data := map[string]any{
		"nested": map[string]any{
			"inner": map[string]any{"k": "v"},
			"list":  []any{1, 2, 3},
		},
	}

	t.Run("valid map returned", func(t *testing.T) {
		m, ok := PathMap(data, "nested", "inner")
		if !ok || m == nil {
			t.Fatalf("PathMap() ok = %v, want true", ok)
		}
		if m["k"] != "v" {
			t.Errorf("m[k] = %v, want v", m["k"])
		}
	})

	t.Run("node is not a map", func(t *testing.T) {
		m, ok := PathMap(data, "nested", "list")
		if ok || m != nil {
			t.Fatalf("PathMap() ok = %v, want false for non-map", ok)
		}
	})

	t.Run("path does not exist", func(t *testing.T) {
		m, ok := PathMap(data, "nested", "missing")
		if ok || m != nil {
			t.Fatalf("PathMap() ok = %v, want false for missing path", ok)
		}
	})
}

func TestPathSlice(t *testing.T) {
	data := map[string]any{
		"items":  []any{"x", "y", "z"},
		"scalar": 100,
	}

	t.Run("valid slice returned", func(t *testing.T) {
		s, ok := PathSlice(data, "items")
		if !ok || len(s) != 3 {
			t.Fatalf("PathSlice() ok = %v, len = %d, want 3", ok, len(s))
		}
		if s[0] != "x" || s[1] != "y" || s[2] != "z" {
			t.Errorf("s = %v, want [x y z]", s)
		}
	})

	t.Run("node is not a slice", func(t *testing.T) {
		s, ok := PathSlice(data, "scalar")
		if ok || s != nil {
			t.Fatalf("PathSlice() ok = %v, want false for non-slice", ok)
		}
	})

	t.Run("path does not exist", func(t *testing.T) {
		s, ok := PathSlice(data, "missing")
		if ok || s != nil {
			t.Fatalf("PathSlice() ok = %v, want false for missing", ok)
		}
	})
}

func TestFirstPathString(t *testing.T) {
	data := map[string]any{
		"empty_str":     "   ",
		"actual_str":    "  found_it  ",
		"json_num_str":  json.Number("12345"),
		"json_num_zero": json.Number("   "),
		"scalar_int":    999,
	}

	tests := []struct {
		name  string
		root  map[string]any
		paths [][]string
		want  string
	}{
		{
			name:  "first path matches string",
			root:  data,
			paths: [][]string{{"actual_str"}},
			want:  "found_it",
		},
		{
			name:  "first path empty string, second matches",
			root:  data,
			paths: [][]string{{"empty_str"}, {"actual_str"}},
			want:  "found_it",
		},
		{
			name:  "json.Number converted to trimmed string",
			root:  data,
			paths: [][]string{{"json_num_str"}},
			want:  "12345",
		},
		{
			name:  "json.Number empty skipped",
			root:  data,
			paths: [][]string{{"json_num_zero"}, {"actual_str"}},
			want:  "found_it",
		},
		{
			name:  "unsupported type skipped and falls through",
			root:  data,
			paths: [][]string{{"scalar_int"}, {"actual_str"}},
			want:  "found_it",
		},
		{
			name:  "no paths match",
			root:  data,
			paths: [][]string{{"missing1"}, {"missing2"}, {"empty_str"}},
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FirstPathString(tt.root, tt.paths...); got != tt.want {
				t.Errorf("FirstPathString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFirstPathNumber(t *testing.T) {
	data := map[string]any{
		"invalid_str": "not_a_num",
		"num_float":   12.34,
		"num_int":     56,
	}

	t.Run("finds first valid number", func(t *testing.T) {
		got := FirstPathNumber(data, []string{"invalid_str"}, []string{"num_float"}, []string{"num_int"})
		if got == nil || *got != 12.34 {
			t.Fatalf("FirstPathNumber() = %v, want 12.34", got)
		}
	})

	t.Run("returns nil if no paths match", func(t *testing.T) {
		got := FirstPathNumber(data, []string{"missing"}, []string{"invalid_str"})
		if got != nil {
			t.Fatalf("FirstPathNumber() = %v, want nil", got)
		}
	})
}

func TestNumberFromAny(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		wantVal float64
		wantOk  bool
	}{
		{name: "float64", input: float64(42.5), wantVal: 42.5, wantOk: true},
		{name: "float32", input: float32(10.5), wantVal: 10.5, wantOk: true},
		{name: "int", input: int(7), wantVal: 7.0, wantOk: true},
		{name: "int64", input: int64(1000), wantVal: 1000.0, wantOk: true},
		{name: "int32", input: int32(50), wantVal: 50.0, wantOk: true},
		{name: "json.Number valid", input: json.Number("123.456"), wantVal: 123.456, wantOk: true},
		{name: "json.Number invalid", input: json.Number("abc"), wantVal: 0, wantOk: false},
		{name: "string valid float", input: "  78.9  ", wantVal: 78.9, wantOk: true},
		{name: "string valid int", input: "42", wantVal: 42.0, wantOk: true},
		{name: "string invalid", input: "bad_number", wantVal: 0, wantOk: false},
		{name: "string empty", input: "  ", wantVal: 0, wantOk: false},
		{name: "bool unsupported", input: true, wantVal: 0, wantOk: false},
		{name: "nil unsupported", input: nil, wantVal: 0, wantOk: false},
		{name: "slice unsupported", input: []int{1, 2}, wantVal: 0, wantOk: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := NumberFromAny(tt.input)
			if ok != tt.wantOk {
				t.Fatalf("NumberFromAny(%v) ok = %v, want %v", tt.input, ok, tt.wantOk)
			}
			if ok && got != tt.wantVal {
				t.Errorf("NumberFromAny(%v) = %f, want %f", tt.input, got, tt.wantVal)
			}
		})
	}
}

func TestNumberPointers(t *testing.T) {
	t.Run("NumberToInt64Ptr nil", func(t *testing.T) {
		if got := NumberToInt64Ptr(nil); got != nil {
			t.Errorf("NumberToInt64Ptr(nil) = %v, want nil", got)
		}
	})

	t.Run("NumberToInt64Ptr non-nil", func(t *testing.T) {
		val := 42.9
		got := NumberToInt64Ptr(&val)
		if got == nil || *got != 42 {
			t.Fatalf("NumberToInt64Ptr(&42.9) = %v, want 42", got)
		}
	})

	t.Run("NumberToFloat64Ptr nil", func(t *testing.T) {
		if got := NumberToFloat64Ptr(nil); got != nil {
			t.Errorf("NumberToFloat64Ptr(nil) = %v, want nil", got)
		}
	})

	t.Run("NumberToFloat64Ptr non-nil", func(t *testing.T) {
		val := 123.456
		got := NumberToFloat64Ptr(&val)
		if got == nil || *got != 123.456 {
			t.Fatalf("NumberToFloat64Ptr(&123.456) = %v, want 123.456", got)
		}
		if got == &val {
			t.Errorf("NumberToFloat64Ptr should allocate a new copy, but returned identical pointer")
		}
	})
}

func TestJsonPath_Concurrency(t *testing.T) {
	root := map[string]any{
		"level1": map[string]any{
			"level2": map[string]any{
				"val":  "hello",
				"num":  json.Number("42"),
				"list": []any{"a", "b"},
			},
		},
	}

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = PathValue(root, "level1", "level2", "val")
			_, _ = PathMap(root, "level1", "level2")
			_, _ = PathSlice(root, "level1", "level2", "list")
			_ = FirstPathString(root, []string{"level1", "level2", "val"})
			_ = FirstPathNumber(root, []string{"level1", "level2", "num"})
		}()
	}
	wg.Wait()
}
