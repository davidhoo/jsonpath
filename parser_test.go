package jsonpath

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseRecursive(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		wantLen     int
		wantErr     bool
		errType     ErrorType
		errContains string
	}{
		{
			name:    "empty path",
			path:    "",
			wantLen: 1, // 只包含递归段
		},
		{
			name:    "simple recursive",
			path:    "name",
			wantLen: 2, // 递归段 + 名称段
		},
		{
			name:    "recursive with dot notation",
			path:    ".name",
			wantLen: 2,
		},
		{
			name:    "recursive with nested path",
			path:    "books.title",
			wantLen: 3, // 递归段 + books段 + title段
		},
		{
			name:    "recursive with bracket notation",
			path:    "[0].name",
			wantLen: 3, // 递归段 + 索引段 + 名称段
		},
		{
			name:    "recursive with filter",
			path:    "[?(@.price > 10)]",
			wantLen: 2, // 递归段 + 过滤器段
		},
		{
			name:        "invalid filter syntax",
			path:        "[?(@.price >=)]",
			wantErr:     true,
			errType:     ErrInvalidFilter,
			errContains: "invalid value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRecursive(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRecursive() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				if jsonErr, ok := err.(*Error); ok {
					if tt.errType != jsonErr.Type {
						t.Errorf("parseRecursive() error type = %v, want %v", jsonErr.Type, tt.errType)
					}
					if tt.errContains != "" && !strings.Contains(jsonErr.Message, tt.errContains) {
						t.Errorf("parseRecursive() error = %v, want error containing %v", jsonErr.Message, tt.errContains)
					}
				} else {
					t.Errorf("parseRecursive() error is not a JSONPath error: %v", err)
				}
				return
			}
			if len(got) != tt.wantLen {
				t.Errorf("parseRecursive() returned %d segments, want %d", len(got), tt.wantLen)
			}
			// 验证第一个段是否为递归段
			if len(got) > 0 {
				if _, ok := got[0].(*recursiveSegment); !ok {
					t.Error("First segment is not a recursiveSegment")
				}
			}
		})
	}
}

func TestRecursiveSegmentString(t *testing.T) {
	seg := &recursiveSegment{}
	expected := ".."
	if got := seg.String(); got != expected {
		t.Errorf("recursiveSegment.String() = %v, want %v", got, expected)
	}
}

func TestParseMultiIndexSegment(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantIndices []int
		wantErr     bool
		errContains string
	}{
		{
			name:        "single index",
			content:     "1",
			wantIndices: []int{1},
		},
		{
			name:        "multiple indices",
			content:     "1,2,3",
			wantIndices: []int{1, 2, 3},
		},
		{
			name:        "indices with spaces",
			content:     "1, 2, 3",
			wantIndices: []int{1, 2, 3},
		},
		{
			name:        "negative indices",
			content:     "-1,-2,-3",
			wantIndices: []int{-1, -2, -3},
		},
		{
			name:        "invalid index",
			content:     "1,a,3",
			wantErr:     true,
			errContains: "invalid index",
		},
		{
			name:        "empty index",
			content:     "1,,3",
			wantErr:     true,
			errContains: "invalid index",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseMultiIndexSegment(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseMultiIndexSegment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("parseMultiIndexSegment() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}

			segment, ok := got.(*multiIndexSegment)
			if !ok {
				t.Fatal("parseMultiIndexSegment() returned wrong type")
			}

			if !reflect.DeepEqual(segment.indices, tt.wantIndices) {
				t.Errorf("parseMultiIndexSegment() indices = %v, want %v", segment.indices, tt.wantIndices)
			}
		})
	}
}

func TestMultiIndexSegmentString(t *testing.T) {
	seg := &multiIndexSegment{indices: []int{1, 2, 3}}
	expected := "[1,2,3]"
	if got := seg.String(); got != expected {
		t.Errorf("multiIndexSegment.String() = %v, want %v", got, expected)
	}
}

func TestParseSliceSegment(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantSlice   *sliceSegment
		wantErr     bool
		errContains string
	}{
		{
			name:      "empty slice",
			content:   ":",
			wantSlice: &sliceSegment{start: 0, end: 0, step: 1},
		},
		{
			name:      "start only",
			content:   "1:",
			wantSlice: &sliceSegment{start: 1, end: 0, step: 1},
		},
		{
			name:      "end only",
			content:   ":2",
			wantSlice: &sliceSegment{start: 0, end: 2, step: 1},
		},
		{
			name:      "start and end",
			content:   "1:2",
			wantSlice: &sliceSegment{start: 1, end: 2, step: 1},
		},
		{
			name:      "with step",
			content:   "1:5:2",
			wantSlice: &sliceSegment{start: 1, end: 5, step: 2},
		},
		{
			name:      "negative indices",
			content:   "-2:-1",
			wantSlice: &sliceSegment{start: -2, end: -1, step: 1},
		},
		{
			name:      "negative step",
			content:   "5:1:-1",
			wantSlice: &sliceSegment{start: 5, end: 1, step: -1},
		},
		{
			name:        "invalid start",
			content:     "a:",
			wantErr:     true,
			errContains: "invalid start index",
		},
		{
			name:        "invalid end",
			content:     ":b",
			wantErr:     true,
			errContains: "invalid end index",
		},
		{
			name:        "invalid step",
			content:     "1:2:c",
			wantErr:     true,
			errContains: "invalid step",
		},
		{
			name:        "zero step",
			content:     "1:2:0",
			wantErr:     true,
			errContains: "step cannot be zero",
		},
		{
			name:        "too many parts",
			content:     "1:2:3:4",
			wantErr:     true,
			errContains: "invalid slice syntax",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSliceSegment(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSliceSegment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("parseSliceSegment() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}

			segment, ok := got.(*sliceSegment)
			if !ok {
				t.Fatal("parseSliceSegment() returned wrong type")
			}

			if !reflect.DeepEqual(segment, tt.wantSlice) {
				t.Errorf("parseSliceSegment() = %+v, want %+v", segment, tt.wantSlice)
			}
		})
	}
}

func TestSliceSegmentString(t *testing.T) {
	tests := []struct {
		name     string
		segment  *sliceSegment
		expected string
	}{
		{
			name:     "full slice",
			segment:  &sliceSegment{start: 1, end: 5, step: 2},
			expected: "[1:5:2]",
		},
		{
			name:     "without step",
			segment:  &sliceSegment{start: 1, end: 5, step: 1},
			expected: "[1:5]",
		},
		{
			name:     "negative indices",
			segment:  &sliceSegment{start: -2, end: -1, step: 1},
			expected: "[-2:-1]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.segment.String(); got != tt.expected {
				t.Errorf("sliceSegment.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseFunctionCall(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantFunc    *functionSegment
		wantErr     bool
		errContains string
	}{
		{
			name:     "no args",
			content:  "length()",
			wantFunc: &functionSegment{name: "length", args: []interface{}{}},
		},
		{
			name:     "single number arg",
			content:  "min(1)",
			wantFunc: &functionSegment{name: "min", args: []interface{}{float64(1)}},
		},
		{
			name:     "multiple number args",
			content:  "sum(1, 2, 3)",
			wantFunc: &functionSegment{name: "sum", args: []interface{}{float64(1), float64(2), float64(3)}},
		},
		{
			name:     "string arg",
			content:  "match('pattern')",
			wantFunc: &functionSegment{name: "match", args: []interface{}{"pattern"}},
		},
		{
			name:     "mixed args",
			content:  "format('value', 42)",
			wantFunc: &functionSegment{name: "format", args: []interface{}{"value", float64(42)}},
		},
		{
			name:        "missing parenthesis",
			content:     "length",
			wantErr:     true,
			errContains: "missing opening parenthesis",
		},
		{
			name:        "invalid arg type",
			content:     "sum(1, invalid, 3)",
			wantErr:     true,
			errContains: "unsupported argument type",
		},
		{
			name:        "unclosed string",
			content:     "match('pattern)",
			wantErr:     true,
			errContains: "unsupported argument type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFunctionCall(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseFunctionCall() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("parseFunctionCall() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}

			segment, ok := got.(*functionSegment)
			if !ok {
				t.Fatal("parseFunctionCall() returned wrong type")
			}

			if segment.name != tt.wantFunc.name {
				t.Errorf("parseFunctionCall() name = %v, want %v", segment.name, tt.wantFunc.name)
			}

			if !reflect.DeepEqual(segment.args, tt.wantFunc.args) {
				t.Errorf("parseFunctionCall() args = %v, want %v", segment.args, tt.wantFunc.args)
			}
		})
	}
}

func TestFunctionSegmentString(t *testing.T) {
	tests := []struct {
		name     string
		segment  *functionSegment
		expected string
	}{
		{
			name:     "no args",
			segment:  &functionSegment{name: "length", args: []interface{}{}},
			expected: "length()",
		},
		{
			name:     "single number arg",
			segment:  &functionSegment{name: "min", args: []interface{}{float64(1)}},
			expected: "min(1)",
		},
		{
			name:     "multiple args",
			segment:  &functionSegment{name: "sum", args: []interface{}{float64(1), float64(2), float64(3)}},
			expected: "sum(1,2,3)",
		},
		{
			name:     "string arg",
			segment:  &functionSegment{name: "match", args: []interface{}{"pattern"}},
			expected: "match('pattern')",
		},
		{
			name:     "mixed args",
			segment:  &functionSegment{name: "format", args: []interface{}{"value", float64(42)}},
			expected: "format('value',42)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.segment.String(); got != tt.expected {
				t.Errorf("functionSegment.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}
