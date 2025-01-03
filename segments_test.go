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
			name:  "positive step",
			start: 0,
			end:   5,
			step:  2,
			want:  []int{0, 2, 4},
		},
		{
			name:  "negative step",
			start: 4,
			end:   -1,
			step:  -2,
			want:  []int{4, 2, 0},
		},
		{
			name:  "zero step",
			start: 0,
			end:   5,
			step:  0,
			want:  []int{0, 1, 2, 3, 4},
		},
		{
			name:  "invalid range with positive step",
			start: 5,
			end:   0,
			step:  1,
			want:  nil,
		},
		{
			name:  "invalid range with negative step",
			start: 0,
			end:   5,
			step:  -1,
			want:  nil,
		},
		{
			name:  "single element range",
			start: 0,
			end:   1,
			step:  1,
			want:  []int{0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateIndices(tt.start, tt.end, tt.step)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("generateIndices(%d, %d, %d) = %v, want %v",
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

func TestCalculateStep(t *testing.T) {
	tests := []struct {
		name string
		step int
		want int
	}{
		{
			name: "zero step",
			step: 0,
			want: 1,
		},
		{
			name: "positive step",
			step: 2,
			want: 2,
		},
		{
			name: "negative step",
			step: -1,
			want: -1,
		},
		{
			name: "large positive step",
			step: 1000,
			want: 1000,
		},
		{
			name: "large negative step",
			step: -1000,
			want: -1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateStep(tt.step); got != tt.want {
				t.Errorf("calculateStep() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalizeRange(t *testing.T) {
	tests := []struct {
		name      string
		segment   *sliceSegment
		length    int
		wantStart int
		wantEnd   int
		wantStep  int
	}{
		{
			name:      "default values with positive step",
			segment:   &sliceSegment{start: 0, end: 0, step: 1},
			length:    5,
			wantStart: 0,
			wantEnd:   5,
			wantStep:  1,
		},
		{
			name:      "default values with negative step",
			segment:   &sliceSegment{start: 0, end: 0, step: -1},
			length:    5,
			wantStart: 4,
			wantEnd:   -1,
			wantStep:  -1,
		},
		{
			name:      "positive indices",
			segment:   &sliceSegment{start: 1, end: 3, step: 1},
			length:    5,
			wantStart: 1,
			wantEnd:   3,
			wantStep:  1,
		},
		{
			name:      "negative indices",
			segment:   &sliceSegment{start: -2, end: -1, step: 1},
			length:    5,
			wantStart: 3,
			wantEnd:   4,
			wantStep:  1,
		},
		{
			name:      "out of range indices",
			segment:   &sliceSegment{start: 10, end: 20, step: 1},
			length:    5,
			wantStart: 5,
			wantEnd:   5,
			wantStep:  1,
		},
		{
			name:      "negative out of range indices",
			segment:   &sliceSegment{start: -10, end: -8, step: 1},
			length:    5,
			wantStart: 0,
			wantEnd:   0,
			wantStep:  1,
		},
		{
			name:      "custom step",
			segment:   &sliceSegment{start: 0, end: 5, step: 2},
			length:    5,
			wantStart: 0,
			wantEnd:   5,
			wantStep:  2,
		},
		{
			name:      "negative step",
			segment:   &sliceSegment{start: 4, end: 0, step: -1},
			length:    5,
			wantStart: 4,
			wantEnd:   -1,
			wantStep:  -1,
		},
		{
			name:      "zero step",
			segment:   &sliceSegment{start: 0, end: 5, step: 0},
			length:    5,
			wantStart: 0,
			wantEnd:   5,
			wantStep:  1,
		},
		{
			name:      "empty array",
			segment:   &sliceSegment{start: 0, end: 0, step: 1},
			length:    0,
			wantStart: 0,
			wantEnd:   0,
			wantStep:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStart, gotEnd, gotStep := tt.segment.normalizeRange(tt.length)
			if gotStart != tt.wantStart {
				t.Errorf("normalizeRange() start = %v, want %v", gotStart, tt.wantStart)
			}
			if gotEnd != tt.wantEnd {
				t.Errorf("normalizeRange() end = %v, want %v", gotEnd, tt.wantEnd)
			}
			if gotStep != tt.wantStep {
				t.Errorf("normalizeRange() step = %v, want %v", gotStep, tt.wantStep)
			}
		})
	}
}

func TestGetArrayElements(t *testing.T) {
	tests := []struct {
		name    string
		arr     []interface{}
		indices []int
		want    []interface{}
	}{
		{
			name:    "empty array",
			arr:     []interface{}{},
			indices: []int{0, 1, 2},
			want:    nil,
		},
		{
			name:    "empty indices",
			arr:     []interface{}{1, 2, 3},
			indices: []int{},
			want:    nil,
		},
		{
			name:    "valid indices",
			arr:     []interface{}{1, 2, 3, 4, 5},
			indices: []int{0, 2, 4},
			want:    []interface{}{1, 3, 5},
		},
		{
			name:    "out of range indices",
			arr:     []interface{}{1, 2, 3},
			indices: []int{-1, 3, 4},
			want:    nil,
		},
		{
			name:    "mixed valid and invalid indices",
			arr:     []interface{}{1, 2, 3},
			indices: []int{0, -1, 1, 3, 2},
			want:    []interface{}{1, 2, 3},
		},
		{
			name:    "duplicate indices",
			arr:     []interface{}{1, 2, 3},
			indices: []int{0, 0, 1, 1, 2, 2},
			want:    []interface{}{1, 1, 2, 2, 3, 3},
		},
		{
			name:    "reverse order indices",
			arr:     []interface{}{1, 2, 3},
			indices: []int{2, 1, 0},
			want:    []interface{}{3, 2, 1},
		},
		{
			name:    "mixed type array",
			arr:     []interface{}{1, "two", true, 4.0, nil},
			indices: []int{0, 2, 4},
			want:    []interface{}{1, true, nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getArrayElements(tt.arr, tt.indices)
			if tt.want == nil {
				if got != nil && len(got) != 0 {
					t.Errorf("getArrayElements() = %v, want nil or empty slice", got)
				}
			} else if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getArrayElements() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSliceSegmentEvaluate(t *testing.T) {
	tests := []struct {
		name    string
		segment *sliceSegment
		value   interface{}
		want    []interface{}
		wantErr bool
	}{
		{
			name: "full slice",
			segment: &sliceSegment{
				start: 0,
				end:   0,
				step:  1,
			},
			value:   []interface{}{1, 2, 3, 4, 5},
			want:    []interface{}{1, 2, 3, 4, 5},
			wantErr: false,
		},
		{
			name: "positive step slice",
			segment: &sliceSegment{
				start: 1,
				end:   4,
				step:  1,
			},
			value:   []interface{}{1, 2, 3, 4, 5},
			want:    []interface{}{2, 3, 4},
			wantErr: false,
		},
		{
			name: "negative step slice",
			segment: &sliceSegment{
				start: 0,
				end:   0,
				step:  -1,
			},
			value:   []interface{}{1, 2, 3, 4, 5},
			want:    []interface{}{5, 4, 3, 2, 1},
			wantErr: false,
		},
		{
			name: "step with gap",
			segment: &sliceSegment{
				start: 0,
				end:   5,
				step:  2,
			},
			value:   []interface{}{1, 2, 3, 4, 5},
			want:    []interface{}{1, 3, 5},
			wantErr: false,
		},
		{
			name: "negative indices",
			segment: &sliceSegment{
				start: -2,
				end:   0,
				step:  1,
			},
			value:   []interface{}{1, 2, 3, 4, 5},
			want:    []interface{}{4, 5},
			wantErr: false,
		},
		{
			name: "out of range indices",
			segment: &sliceSegment{
				start: 5,
				end:   10,
				step:  1,
			},
			value:   []interface{}{1, 2, 3, 4, 5},
			want:    nil,
			wantErr: false,
		},
		{
			name: "non-array value",
			segment: &sliceSegment{
				start: 0,
				end:   5,
				step:  1,
			},
			value:   "not an array",
			want:    nil,
			wantErr: true,
		},
		{
			name: "nil value",
			segment: &sliceSegment{
				start: 0,
				end:   5,
				step:  1,
			},
			value:   nil,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.segment.evaluate(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("sliceSegment.evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if tt.want == nil {
					if got != nil && len(got) != 0 {
						t.Errorf("sliceSegment.evaluate() = %v, want nil or empty slice", got)
					}
				} else if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("sliceSegment.evaluate() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestFilterSegmentString(t *testing.T) {
	tests := []struct {
		name    string
		segment *filterSegment
		want    string
	}{
		{
			name: "single condition",
			segment: &filterSegment{
				conditions: []filterCondition{
					{field: "@.price", operator: ">", value: float64(10)},
				},
			},
			want: "[?@.price > 10]",
		},
		{
			name: "multiple conditions with AND",
			segment: &filterSegment{
				conditions: []filterCondition{
					{field: "@.price", operator: ">", value: float64(10)},
					{field: "@.category", operator: "==", value: "book"},
				},
				operators: []string{"&&"},
			},
			want: "[?@.price > 10 && @.category == 'book']",
		},
		{
			name: "multiple conditions with OR",
			segment: &filterSegment{
				conditions: []filterCondition{
					{field: "@.price", operator: "<", value: float64(5)},
					{field: "@.price", operator: ">", value: float64(100)},
				},
				operators: []string{"||"},
			},
			want: "[?@.price < 5 || @.price > 100]",
		},
		{
			name: "complex conditions",
			segment: &filterSegment{
				conditions: []filterCondition{
					{field: "@.price", operator: ">", value: float64(10)},
					{field: "@.category", operator: "==", value: "book"},
					{field: "@.inStock", operator: "==", value: true},
				},
				operators: []string{"&&", "&&"},
			},
			want: "[?@.price > 10 && @.category == 'book' && @.inStock == true]",
		},
		{
			name: "empty conditions",
			segment: &filterSegment{
				conditions: []filterCondition{},
			},
			want: "[?]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.segment.String(); got != tt.want {
				t.Errorf("filterSegment.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMultiIndexSegmentEvaluate(t *testing.T) {
	tests := []struct {
		name    string
		segment *multiIndexSegment
		value   interface{}
		want    []interface{}
		wantErr bool
	}{
		{
			name: "simple indices",
			segment: &multiIndexSegment{
				indices: []int{0, 2, 4},
			},
			value:   []interface{}{1, 2, 3, 4, 5},
			want:    []interface{}{1, 3, 5},
			wantErr: false,
		},
		{
			name: "negative indices",
			segment: &multiIndexSegment{
				indices: []int{-1, -2},
			},
			value:   []interface{}{1, 2, 3, 4, 5},
			want:    []interface{}{5, 4},
			wantErr: false,
		},
		{
			name: "mixed indices",
			segment: &multiIndexSegment{
				indices: []int{0, -1, 2},
			},
			value:   []interface{}{1, 2, 3, 4, 5},
			want:    []interface{}{1, 5, 3},
			wantErr: false,
		},
		{
			name: "out of range indices",
			segment: &multiIndexSegment{
				indices: []int{5, 6, -6},
			},
			value:   []interface{}{1, 2, 3, 4, 5},
			want:    nil,
			wantErr: false,
		},
		{
			name: "empty array",
			segment: &multiIndexSegment{
				indices: []int{0, 1, 2},
			},
			value:   []interface{}{},
			want:    nil,
			wantErr: false,
		},
		{
			name: "empty indices",
			segment: &multiIndexSegment{
				indices: []int{},
			},
			value:   []interface{}{1, 2, 3},
			want:    nil,
			wantErr: false,
		},
		{
			name: "duplicate indices",
			segment: &multiIndexSegment{
				indices: []int{0, 0, 1, 1},
			},
			value:   []interface{}{1, 2, 3},
			want:    []interface{}{1, 1, 2, 2},
			wantErr: false,
		},
		{
			name: "non-array value",
			segment: &multiIndexSegment{
				indices: []int{0, 1},
			},
			value:   "not an array",
			want:    nil,
			wantErr: true,
		},
		{
			name: "nil value",
			segment: &multiIndexSegment{
				indices: []int{0, 1},
			},
			value:   nil,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.segment.evaluate(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("multiIndexSegment.evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if tt.want == nil {
					if got != nil && len(got) != 0 {
						t.Errorf("multiIndexSegment.evaluate() = %v, want nil or empty slice", got)
					}
				} else if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("multiIndexSegment.evaluate() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestFunctionSegmentEvaluate(t *testing.T) {
	tests := []struct {
		name    string
		segment *functionSegment
		value   interface{}
		want    []interface{}
		wantErr bool
	}{
		{
			name: "length function without args",
			segment: &functionSegment{
				name: "length",
				args: []interface{}{},
			},
			value:   []interface{}{1, 2, 3},
			want:    []interface{}{float64(3)},
			wantErr: false,
		},
		{
			name: "min function with array",
			segment: &functionSegment{
				name: "min",
				args: []interface{}{[]interface{}{3, 1, 4, 1, 5}},
			},
			value:   nil,
			want:    []interface{}{float64(1)},
			wantErr: false,
		},
		{
			name: "max function with array",
			segment: &functionSegment{
				name: "max",
				args: []interface{}{[]interface{}{3, 1, 4, 1, 5}},
			},
			value:   nil,
			want:    []interface{}{float64(5)},
			wantErr: false,
		},
		{
			name: "avg function with array",
			segment: &functionSegment{
				name: "avg",
				args: []interface{}{[]interface{}{1, 2, 3, 4, 5}},
			},
			value:   nil,
			want:    []interface{}{float64(3)},
			wantErr: false,
		},
		{
			name: "unknown function",
			segment: &functionSegment{
				name: "unknown",
				args: []interface{}{},
			},
			value:   []interface{}{1, 2, 3},
			want:    nil,
			wantErr: true,
		},
		{
			name: "function with invalid args",
			segment: &functionSegment{
				name: "min",
				args: []interface{}{"not a number"},
			},
			value:   nil,
			want:    nil,
			wantErr: true,
		},
		{
			name: "function with empty array",
			segment: &functionSegment{
				name: "min",
				args: []interface{}{[]interface{}{}},
			},
			value:   nil,
			want:    nil,
			wantErr: true,
		},
		{
			name: "function with nil value",
			segment: &functionSegment{
				name: "length",
				args: []interface{}{nil},
			},
			value:   nil,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.segment.evaluate(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("functionSegment.evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("functionSegment.evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWildcardSegmentEvaluate(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		want    []interface{}
		wantErr bool
	}{
		{
			name:    "array value",
			value:   []interface{}{1, 2, 3},
			want:    []interface{}{1, 2, 3},
			wantErr: false,
		},
		{
			name: "object value",
			value: map[string]interface{}{
				"a": 1,
				"b": 2,
				"c": 3,
			},
			want:    []interface{}{1, 2, 3},
			wantErr: false,
		},
		{
			name:    "empty array",
			value:   []interface{}{},
			want:    []interface{}{},
			wantErr: false,
		},
		{
			name:    "empty object",
			value:   map[string]interface{}{},
			want:    []interface{}{},
			wantErr: false,
		},
		{
			name: "nested array",
			value: []interface{}{
				[]interface{}{1, 2},
				map[string]interface{}{"a": 3},
				4,
			},
			want: []interface{}{
				[]interface{}{1, 2},
				map[string]interface{}{"a": 3},
				4,
			},
			wantErr: false,
		},
		{
			name: "nested object",
			value: map[string]interface{}{
				"arr": []interface{}{1, 2},
				"obj": map[string]interface{}{"a": 3},
				"num": 4,
			},
			want: []interface{}{
				[]interface{}{1, 2},
				map[string]interface{}{"a": 3},
				4,
			},
			wantErr: false,
		},
		{
			name:    "nil value",
			value:   nil,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "string value",
			value:   "not an array or object",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "number value",
			value:   42,
			want:    nil,
			wantErr: true,
		},
		{
			name:    "boolean value",
			value:   true,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &wildcardSegment{}
			got, err := s.evaluate(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("wildcardSegment.evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// 由于 map 的迭代顺序是不确定的，我们需要排序后再比较
				sortSlice := func(s []interface{}) {
					sort.Slice(s, func(i, j int) bool {
						return fmt.Sprintf("%v", s[i]) < fmt.Sprintf("%v", s[j])
					})
				}
				sortSlice(got)
				sortSlice(tt.want)
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("wildcardSegment.evaluate() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestWildcardSegmentString(t *testing.T) {
	s := &wildcardSegment{}
	want := "*"
	if got := s.String(); got != want {
		t.Errorf("wildcardSegment.String() = %v, want %v", got, want)
	}
}

func TestFilterSegmentEvaluate(t *testing.T) {
	tests := []struct {
		name    string
		segment *filterSegment
		value   interface{}
		want    []interface{}
		wantErr bool
	}{
		{
			name: "single condition - array",
			segment: &filterSegment{
				conditions: []filterCondition{
					{field: "@.price", operator: ">", value: float64(10)},
				},
			},
			value: []interface{}{
				map[string]interface{}{"price": float64(5)},
				map[string]interface{}{"price": float64(15)},
				map[string]interface{}{"price": float64(8)},
				map[string]interface{}{"price": float64(20)},
			},
			want: []interface{}{
				map[string]interface{}{"price": float64(15)},
				map[string]interface{}{"price": float64(20)},
			},
			wantErr: false,
		},
		{
			name: "single condition - object",
			segment: &filterSegment{
				conditions: []filterCondition{
					{field: "@.price", operator: ">", value: float64(10)},
				},
			},
			value: map[string]interface{}{"price": float64(15)},
			want: []interface{}{
				map[string]interface{}{"price": float64(15)},
			},
			wantErr: false,
		},
		{
			name: "multiple conditions with AND",
			segment: &filterSegment{
				conditions: []filterCondition{
					{field: "@.price", operator: ">", value: float64(10)},
					{field: "@.category", operator: "==", value: "book"},
				},
				operators: []string{"&&"},
			},
			value: []interface{}{
				map[string]interface{}{"price": float64(15), "category": "book"},
				map[string]interface{}{"price": float64(5), "category": "book"},
				map[string]interface{}{"price": float64(20), "category": "movie"},
				map[string]interface{}{"price": float64(25), "category": "book"},
			},
			want: []interface{}{
				map[string]interface{}{"price": float64(15), "category": "book"},
				map[string]interface{}{"price": float64(25), "category": "book"},
			},
			wantErr: false,
		},
		{
			name: "multiple conditions with OR",
			segment: &filterSegment{
				conditions: []filterCondition{
					{field: "@.price", operator: "<", value: float64(10)},
					{field: "@.price", operator: ">", value: float64(20)},
				},
				operators: []string{"||"},
			},
			value: []interface{}{
				map[string]interface{}{"price": float64(5)},
				map[string]interface{}{"price": float64(15)},
				map[string]interface{}{"price": float64(25)},
			},
			want: []interface{}{
				map[string]interface{}{"price": float64(5)},
				map[string]interface{}{"price": float64(25)},
			},
			wantErr: false,
		},
		{
			name: "regex match condition",
			segment: &filterSegment{
				conditions: []filterCondition{
					{field: "@.name", operator: "match", value: "^J.*n$"},
				},
			},
			value: []interface{}{
				map[string]interface{}{"name": "John"},
				map[string]interface{}{"name": "Jane"},
				map[string]interface{}{"name": "Bob"},
			},
			want: []interface{}{
				map[string]interface{}{"name": "John"},
			},
			wantErr: false,
		},
		{
			name: "invalid regex pattern",
			segment: &filterSegment{
				conditions: []filterCondition{
					{field: "@.name", operator: "match", value: "[invalid"},
				},
			},
			value: []interface{}{
				map[string]interface{}{"name": "John"},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "field not found",
			segment: &filterSegment{
				conditions: []filterCondition{
					{field: "@.missing", operator: ">", value: float64(10)},
				},
			},
			value: []interface{}{
				map[string]interface{}{"price": float64(15)},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "invalid operator",
			segment: &filterSegment{
				conditions: []filterCondition{
					{field: "@.price", operator: "invalid", value: float64(10)},
				},
			},
			value: []interface{}{
				map[string]interface{}{"price": float64(15)},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "non-object array elements",
			segment: &filterSegment{
				conditions: []filterCondition{
					{field: "@.price", operator: ">", value: float64(10)},
				},
			},
			value: []interface{}{
				"string",
				42,
				true,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "nil value",
			segment: &filterSegment{
				conditions: []filterCondition{
					{field: "@.price", operator: ">", value: float64(10)},
				},
			},
			value:   nil,
			want:    nil,
			wantErr: false,
		},
		{
			name: "non-array/object value",
			segment: &filterSegment{
				conditions: []filterCondition{
					{field: "@.price", operator: ">", value: float64(10)},
				},
			},
			value:   "not an array or object",
			want:    nil,
			wantErr: false,
		},
		{
			name: "empty conditions",
			segment: &filterSegment{
				conditions: []filterCondition{},
			},
			value: []interface{}{
				map[string]interface{}{"price": float64(15)},
			},
			want:    nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.segment.evaluate(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("filterSegment.evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if tt.want == nil {
					if got != nil && len(got) != 0 {
						t.Errorf("filterSegment.evaluate() = %v, want nil or empty slice", got)
					}
				} else if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("filterSegment.evaluate() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
