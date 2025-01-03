package jsonpath

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func TestMapToArray(t *testing.T) {
	tests := []struct {
		name string
		m    map[string]interface{}
		want []interface{}
	}{
		{
			name: "empty map",
			m:    map[string]interface{}{},
			want: []interface{}{},
		},
		{
			name: "single key-value pair",
			m: map[string]interface{}{
				"key": "value",
			},
			want: []interface{}{"value"},
		},
		{
			name: "multiple key-value pairs",
			m: map[string]interface{}{
				"name": "John",
				"age":  30,
				"city": "New York",
			},
			want: []interface{}{"John", 30, "New York"},
		},
		{
			name: "nested objects",
			m: map[string]interface{}{
				"name": "John",
				"address": map[string]interface{}{
					"city":  "New York",
					"state": "NY",
				},
				"age": 30,
			},
			want: []interface{}{
				"John",
				map[string]interface{}{
					"city":  "New York",
					"state": "NY",
				},
				30,
			},
		},
		{
			name: "array values",
			m: map[string]interface{}{
				"numbers": []interface{}{1, 2, 3},
				"letters": []interface{}{"a", "b", "c"},
			},
			want: []interface{}{
				[]interface{}{1, 2, 3},
				[]interface{}{"a", "b", "c"},
			},
		},
		{
			name: "mixed types",
			m: map[string]interface{}{
				"string":  "text",
				"number":  42,
				"boolean": true,
				"null":    nil,
				"array":   []interface{}{1, "two", true},
				"object": map[string]interface{}{
					"key": "value",
				},
			},
			want: []interface{}{
				"text",
				42,
				true,
				nil,
				[]interface{}{1, "two", true},
				map[string]interface{}{
					"key": "value",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapToArray(tt.m)
			// 由于 map 的迭代顺序是不确定的，我们需要排序后再比较
			sortSlice := func(s []interface{}) {
				sort.Slice(s, func(i, j int) bool {
					return fmt.Sprintf("%v", s[i]) < fmt.Sprintf("%v", s[j])
				})
			}
			sortSlice(got)
			sortSlice(tt.want)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mapToArray() = %v, want %v", got, tt.want)
			}
		})
	}
}

// toString 将任意值转换为字符串以便比较
func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case int:
		return strconv.Itoa(val)
	case bool:
		return strconv.FormatBool(val)
	case map[string]interface{}:
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var result strings.Builder
		for _, k := range keys {
			result.WriteString(k)
			result.WriteString(":")
			result.WriteString(toString(val[k]))
			result.WriteString(",")
		}
		return result.String()
	case []interface{}:
		var result strings.Builder
		for _, item := range val {
			result.WriteString(toString(item))
			result.WriteString(",")
		}
		return result.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

func TestRecursiveCollect(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		want    []interface{}
		wantErr bool
	}{
		{
			name:  "simple array",
			value: []interface{}{1, 2, 3},
			want:  []interface{}{1, 2, 3},
		},
		{
			name: "nested array",
			value: []interface{}{
				1,
				[]interface{}{2, 3},
				4,
			},
			want: []interface{}{1, []interface{}{2, 3}, 2, 3, 4},
		},
		{
			name: "simple object",
			value: map[string]interface{}{
				"a": 1,
				"b": 2,
			},
			want: []interface{}{1, 2},
		},
		{
			name: "nested object",
			value: map[string]interface{}{
				"a": 1,
				"b": map[string]interface{}{
					"c": 2,
					"d": 3,
				},
				"e": 4,
			},
			want: []interface{}{1, map[string]interface{}{"c": 2, "d": 3}, 2, 3, 4},
		},
		{
			name: "mixed nested structure",
			value: map[string]interface{}{
				"a": []interface{}{1, 2},
				"b": map[string]interface{}{
					"c": []interface{}{3, 4},
					"d": 5,
				},
				"e": 6,
			},
			want: []interface{}{
				[]interface{}{1, 2},
				1, 2,
				map[string]interface{}{"c": []interface{}{3, 4}, "d": 5},
				[]interface{}{3, 4},
				3, 4,
				5,
				6,
			},
		},
		{
			name:  "primitive value",
			value: 42,
			want:  nil,
		},
		{
			name:  "nil value",
			value: nil,
			want:  nil,
		},
		{
			name: "empty structures",
			value: map[string]interface{}{
				"a": []interface{}{},
				"b": map[string]interface{}{},
			},
			want: []interface{}{[]interface{}{}, map[string]interface{}{}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &recursiveSegment{}
			got, err := s.evaluate(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("recursiveSegment.evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.want == nil {
				if len(got) != 0 {
					t.Errorf("recursiveSegment.evaluate() = %v, want empty result", got)
				}
				return
			}

			// 由于 map 的迭代顺序是不确定的，我们需要排序后再比较
			sortSlice := func(s []interface{}) {
				sort.Slice(s, func(i, j int) bool {
					return fmt.Sprintf("%v", s[i]) < fmt.Sprintf("%v", s[j])
				})
			}
			sortSlice(got)
			sortSlice(tt.want)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("recursiveSegment.evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculateIndex(t *testing.T) {
	tests := []struct {
		name   string
		idx    int
		length int
		want   int
	}{
		{
			name:   "positive index within bounds",
			idx:    2,
			length: 5,
			want:   2,
		},
		{
			name:   "zero index",
			idx:    0,
			length: 5,
			want:   0,
		},
		{
			name:   "index equals length",
			idx:    5,
			length: 5,
			want:   5,
		},
		{
			name:   "index exceeds length",
			idx:    7,
			length: 5,
			want:   5,
		},
		{
			name:   "negative index within bounds",
			idx:    -2,
			length: 5,
			want:   3,
		},
		{
			name:   "negative index equals negative length",
			idx:    -5,
			length: 5,
			want:   0,
		},
		{
			name:   "negative index exceeds length",
			idx:    -6,
			length: 5,
			want:   0,
		},
		{
			name:   "empty array",
			idx:    1,
			length: 0,
			want:   0,
		},
		{
			name:   "negative index with empty array",
			idx:    -1,
			length: 0,
			want:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateIndex(tt.idx, tt.length); got != tt.want {
				t.Errorf("calculateIndex(%v, %v) = %v, want %v",
					tt.idx, tt.length, got, tt.want)
			}
		})
	}
}

func TestGenerateIndices(t *testing.T) {
	tests := []struct {
		name  string
		start int
		end   int
		step  int
		want  []int
	}{
		{
			name:  "positive step forward",
			start: 0,
			end:   5,
			step:  1,
			want:  []int{0, 1, 2, 3, 4},
		},
		{
			name:  "positive step with gap",
			start: 0,
			end:   6,
			step:  2,
			want:  []int{0, 2, 4},
		},
		{
			name:  "negative step backward",
			start: 5,
			end:   0,
			step:  -1,
			want:  []int{5, 4, 3, 2, 1},
		},
		{
			name:  "negative step with gap",
			start: 6,
			end:   0,
			step:  -2,
			want:  []int{6, 4, 2},
		},
		{
			name:  "start equals end",
			start: 3,
			end:   3,
			step:  1,
			want:  nil,
		},
		{
			name:  "start after end with positive step",
			start: 5,
			end:   3,
			step:  1,
			want:  nil,
		},
		{
			name:  "start before end with negative step",
			start: 3,
			end:   5,
			step:  -1,
			want:  nil,
		},
		{
			name:  "large step",
			start: 0,
			end:   10,
			step:  3,
			want:  []int{0, 3, 6, 9},
		},
		{
			name:  "zero step",
			start: 0,
			end:   5,
			step:  0,
			want:  nil,
		},
		{
			name:  "single element range",
			start: 1,
			end:   2,
			step:  1,
			want:  []int{1},
		},
		{
			name:  "negative indices",
			start: -3,
			end:   -1,
			step:  1,
			want:  []int{-3, -2},
		},
		{
			name:  "step larger than range",
			start: 0,
			end:   5,
			step:  10,
			want:  []int{0},
		},
		{
			name:  "negative step larger than range",
			start: 5,
			end:   0,
			step:  -10,
			want:  []int{5},
		},
		{
			name:  "empty range at start",
			start: 0,
			end:   0,
			step:  1,
			want:  nil,
		},
		{
			name:  "empty range at end",
			start: 5,
			end:   5,
			step:  1,
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateIndices(tt.start, tt.end, tt.step)
			if tt.want == nil {
				if got != nil && len(got) != 0 {
					t.Errorf("generateIndices(%v, %v, %v) = %v, want empty slice",
						tt.start, tt.end, tt.step, got)
				}
			} else if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("generateIndices(%v, %v, %v) = %v, want %v",
					tt.start, tt.end, tt.step, got, tt.want)
			}
		})
	}
}

func TestNameSegmentEvaluate(t *testing.T) {
	tests := []struct {
		name        string
		segment     *nameSegment
		value       interface{}
		want        []interface{}
		wantErr     bool
		errContains string
	}{
		{
			name:    "simple field access",
			segment: &nameSegment{name: "name"},
			value: map[string]interface{}{
				"name": "John",
				"age":  30,
			},
			want: []interface{}{"John"},
		},
		{
			name:    "nested field access",
			segment: &nameSegment{name: "address"},
			value: map[string]interface{}{
				"name": "John",
				"address": map[string]interface{}{
					"city": "New York",
				},
			},
			want: []interface{}{map[string]interface{}{
				"city": "New York",
			}},
		},
		{
			name:    "field not found",
			segment: &nameSegment{name: "phone"},
			value: map[string]interface{}{
				"name": "John",
				"age":  30,
			},
			wantErr:     true,
			errContains: "field phone not found",
		},
		{
			name:        "value is not an object",
			segment:     &nameSegment{name: "name"},
			value:       "not an object",
			wantErr:     true,
			errContains: "value is not an object",
		},
		{
			name:    "function call without arguments",
			segment: &nameSegment{name: "length()"},
			value:   []interface{}{1, 2, 3},
			want:    []interface{}{float64(3)},
		},
		{
			name:        "function call with arguments",
			segment:     &nameSegment{name: "min(1,2,3)"},
			value:       []interface{}{4, 5, 6},
			wantErr:     true,
			errContains: "invalid argument: min() requires exactly 1 argument",
		},
		{
			name:    "function call with string argument",
			segment: &nameSegment{name: "match('pattern')"},
			value:   "test pattern",
			want:    []interface{}{true},
		},
		{
			name:        "invalid function call syntax",
			segment:     &nameSegment{name: "invalid("},
			value:       []interface{}{1, 2, 3},
			wantErr:     true,
			errContains: "invalid function call syntax",
		},
		{
			name:        "unknown function",
			segment:     &nameSegment{name: "unknown()"},
			value:       []interface{}{1, 2, 3},
			wantErr:     true,
			errContains: "unknown function",
		},
		{
			name:        "invalid function arguments",
			segment:     &nameSegment{name: "min('invalid')"},
			value:       []interface{}{1, 2, 3},
			wantErr:     true,
			errContains: "invalid argument",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.segment.evaluate(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("nameSegment.evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("nameSegment.evaluate() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("nameSegment.evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNameSegmentString(t *testing.T) {
	segment := &nameSegment{name: "test"}
	if got := segment.String(); got != "test" {
		t.Errorf("nameSegment.String() = %v, want %v", got, "test")
	}
}

func TestIndexSegmentEvaluate(t *testing.T) {
	tests := []struct {
		name        string
		segment     *indexSegment
		value       interface{}
		want        []interface{}
		wantErr     bool
		errContains string
	}{
		{
			name:    "positive index within bounds",
			segment: &indexSegment{index: 1},
			value:   []interface{}{1, 2, 3},
			want:    []interface{}{2},
		},
		{
			name:    "zero index",
			segment: &indexSegment{index: 0},
			value:   []interface{}{1, 2, 3},
			want:    []interface{}{1},
		},
		{
			name:    "negative index",
			segment: &indexSegment{index: -1},
			value:   []interface{}{1, 2, 3},
			want:    []interface{}{3},
		},
		{
			name:    "index out of bounds (positive)",
			segment: &indexSegment{index: 3},
			value:   []interface{}{1, 2, 3},
			want:    []interface{}{},
		},
		{
			name:    "index out of bounds (negative)",
			segment: &indexSegment{index: -4},
			value:   []interface{}{1, 2, 3},
			want:    []interface{}{},
		},
		{
			name:    "empty array",
			segment: &indexSegment{index: 0},
			value:   []interface{}{},
			want:    []interface{}{},
		},
		{
			name:        "value is not an array",
			segment:     &indexSegment{index: 0},
			value:       "not an array",
			wantErr:     true,
			errContains: "value is not an array",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.segment.evaluate(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("indexSegment.evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("indexSegment.evaluate() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("indexSegment.evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIndexSegmentString(t *testing.T) {
	tests := []struct {
		name    string
		segment *indexSegment
		want    string
	}{
		{
			name:    "positive index",
			segment: &indexSegment{index: 1},
			want:    "[1]",
		},
		{
			name:    "zero index",
			segment: &indexSegment{index: 0},
			want:    "[0]",
		},
		{
			name:    "negative index",
			segment: &indexSegment{index: -1},
			want:    "[-1]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.segment.String(); got != tt.want {
				t.Errorf("indexSegment.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIndexSegmentNormalizeIndex(t *testing.T) {
	tests := []struct {
		name    string
		segment *indexSegment
		length  int
		want    int
	}{
		{
			name:    "positive index",
			segment: &indexSegment{index: 1},
			length:  3,
			want:    1,
		},
		{
			name:    "zero index",
			segment: &indexSegment{index: 0},
			length:  3,
			want:    0,
		},
		{
			name:    "negative index",
			segment: &indexSegment{index: -1},
			length:  3,
			want:    2,
		},
		{
			name:    "negative index with length 1",
			segment: &indexSegment{index: -1},
			length:  1,
			want:    0,
		},
		{
			name:    "negative index equals negative length",
			segment: &indexSegment{index: -3},
			length:  3,
			want:    0,
		},
		{
			name:    "negative index exceeds length",
			segment: &indexSegment{index: -4},
			length:  3,
			want:    -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.segment.normalizeIndex(tt.length); got != tt.want {
				t.Errorf("indexSegment.normalizeIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}
