package segments

import (
	"reflect"
	"sort"
	"testing"
)

func TestRecursiveSegment(t *testing.T) {
	tests := []struct {
		name     string
		segment  *RecursiveSegment
		input    interface{}
		expected []interface{}
		wantErr  bool
	}{
		{
			name:    "simple object",
			segment: NewRecursiveSegment(),
			input: map[string]interface{}{
				"name": "test",
				"age":  30,
			},
			expected: []interface{}{
				map[string]interface{}{
					"name": "test",
					"age":  30,
				},
				"test",
				30,
			},
			wantErr: false,
		},
		{
			name:    "nested object",
			segment: NewRecursiveSegment(),
			input: map[string]interface{}{
				"person": map[string]interface{}{
					"name": "test",
					"age":  30,
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"person": map[string]interface{}{
						"name": "test",
						"age":  30,
					},
				},
				map[string]interface{}{
					"name": "test",
					"age":  30,
				},
				"test",
				30,
			},
			wantErr: false,
		},
		{
			name:    "array with objects",
			segment: NewRecursiveSegment(),
			input: []interface{}{
				map[string]interface{}{"name": "test1"},
				map[string]interface{}{"name": "test2"},
			},
			expected: []interface{}{
				[]interface{}{
					map[string]interface{}{"name": "test1"},
					map[string]interface{}{"name": "test2"},
				},
				map[string]interface{}{"name": "test1"},
				"test1",
				map[string]interface{}{"name": "test2"},
				"test2",
			},
			wantErr: false,
		},
		{
			name:     "nil input",
			segment:  NewRecursiveSegment(),
			input:    nil,
			expected: nil,
			wantErr:  false,
		},
		{
			name:     "primitive value",
			segment:  NewRecursiveSegment(),
			input:    "test",
			expected: []interface{}{"test"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.segment.Evaluate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("RecursiveSegment.Evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			// 由于递归遍历的顺序可能不同，我们需要对结果进行排序后比较
			if got != nil && tt.expected != nil {
				// 将结果转换为字符串切片以便排序
				gotStr := make([]string, len(got))
				expectedStr := make([]string, len(tt.expected))
				for i, v := range got {
					gotStr[i] = toString(v)
				}
				for i, v := range tt.expected {
					expectedStr[i] = toString(v)
				}
				sort.Strings(gotStr)
				sort.Strings(expectedStr)
				if !reflect.DeepEqual(gotStr, expectedStr) {
					t.Errorf("RecursiveSegment.Evaluate() = %v, want %v", gotStr, expectedStr)
				}
			} else if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("RecursiveSegment.Evaluate() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// toString 将任意值转换为字符串，用于排序比较
func toString(v interface{}) string {
	switch val := v.(type) {
	case map[string]interface{}:
		// 对于对象，我们需要确保键的顺序一致
		keys := make([]string, 0, len(val))
		for k := range val {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		result := "{"
		for i, k := range keys {
			if i > 0 {
				result += ","
			}
			result += k + ":" + toString(val[k])
		}
		return result + "}"
	case []interface{}:
		// 对于数组，我们需要确保元素的顺序一致
		result := "["
		for i, v := range val {
			if i > 0 {
				result += ","
			}
			result += toString(v)
		}
		return result + "]"
	default:
		return reflect.ValueOf(v).String()
	}
}
