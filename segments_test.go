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
			name: "single item",
			m: map[string]interface{}{
				"key": "value",
			},
			want: []interface{}{"value"},
		},
		{
			name: "multiple items",
			m: map[string]interface{}{
				"key1": "value1",
				"key2": 42,
				"key3": true,
			},
			want: []interface{}{"value1", 42, true},
		},
		{
			name: "nested structures",
			m: map[string]interface{}{
				"key1": map[string]interface{}{"nested": "value"},
				"key2": []interface{}{1, 2, 3},
			},
			want: []interface{}{
				map[string]interface{}{"nested": "value"},
				[]interface{}{1, 2, 3},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapToArray(tt.m)
			// 由于 map 的迭代顺序是不确定的，我们需要比较集合而不是切片
			if len(got) != len(tt.want) {
				t.Errorf("mapToArray() returned %d items, want %d", len(got), len(tt.want))
				return
			}
			// 创建一个映射来检查所有值是否存在
			wantMap := make(map[interface{}]bool)
			for _, v := range tt.want {
				wantMap[toString(v)] = true
			}
			for _, v := range got {
				if !wantMap[toString(v)] {
					t.Errorf("mapToArray() returned unexpected value: %v", v)
				}
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
			name:  "primitive value",
			value: 42,
			want:  []interface{}{},
		},
		{
			name:  "empty array",
			value: []interface{}{},
			want:  []interface{}{},
		},
		{
			name:  "array with primitives",
			value: []interface{}{1, "two", true},
			want:  []interface{}{1, "two", true},
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
			name:  "empty object",
			value: map[string]interface{}{},
			want:  []interface{}{},
		},
		{
			name: "object with primitives",
			value: map[string]interface{}{
				"a": 1,
				"b": "two",
				"c": true,
			},
			want: []interface{}{1, "two", true},
		},
		{
			name: "nested object",
			value: map[string]interface{}{
				"a": 1,
				"b": map[string]interface{}{
					"x": 2,
					"y": 3,
				},
				"c": 4,
			},
			want: []interface{}{1, map[string]interface{}{"x": 2, "y": 3}, 2, 3, 4},
		},
		{
			name: "complex nested structure",
			value: map[string]interface{}{
				"a": []interface{}{
					1,
					map[string]interface{}{
						"x": 2,
						"y": []interface{}{3, 4},
					},
					5,
				},
				"b": 6,
			},
			want: []interface{}{
				[]interface{}{1, map[string]interface{}{"x": 2, "y": []interface{}{3, 4}}, 5},
				1,
				map[string]interface{}{"x": 2, "y": []interface{}{3, 4}},
				2,
				[]interface{}{3, 4},
				3,
				4,
				5,
				6,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seg := &recursiveSegment{}
			var got []interface{}
			err := seg.recursiveCollect(tt.value, &got)

			if (err != nil) != tt.wantErr {
				t.Errorf("recursiveCollect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// 由于递归收集的顺序可能不同，我们需要比较集合而不是切片
			if len(got) != len(tt.want) {
				t.Errorf("recursiveCollect() returned %d items, want %d", len(got), len(tt.want))
				return
			}

			// 创建映射来检查所有值是否存在
			wantMap := make(map[interface{}]bool)
			for _, v := range tt.want {
				wantMap[toString(v)] = true
			}

			for _, v := range got {
				if !wantMap[toString(v)] {
					t.Errorf("recursiveCollect() returned unexpected value: %v", v)
				}
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
