package segments

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
)

func TestWildcardSegment(t *testing.T) {
	tests := []struct {
		name     string
		segment  *WildcardSegment
		input    interface{}
		expected []interface{}
		wantErr  bool
	}{
		{
			name:    "object wildcard",
			segment: NewWildcardSegment(),
			input: map[string]interface{}{
				"name": "test",
				"age":  30,
				"city": "beijing",
			},
			expected: []interface{}{"test", 30, "beijing"},
			wantErr:  false,
		},
		{
			name:     "array wildcard",
			segment:  NewWildcardSegment(),
			input:    []interface{}{1, 2, 3},
			expected: []interface{}{1, 2, 3},
			wantErr:  false,
		},
		{
			name:     "empty object",
			segment:  NewWildcardSegment(),
			input:    map[string]interface{}{},
			expected: []interface{}{},
			wantErr:  false,
		},
		{
			name:     "empty array",
			segment:  NewWildcardSegment(),
			input:    []interface{}{},
			expected: []interface{}{},
			wantErr:  false,
		},
		{
			name:     "nil input",
			segment:  NewWildcardSegment(),
			input:    nil,
			expected: nil,
			wantErr:  false,
		},
		{
			name:     "non-object and non-array input",
			segment:  NewWildcardSegment(),
			input:    "string",
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.segment.Evaluate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("WildcardSegment.Evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			// 对于对象通配符测试，我们需要对结果进行排序以使其顺序无关
			if _, ok := tt.input.(map[string]interface{}); ok {
				// 将结果转换为字符串切片以便排序
				gotStr := make([]string, len(got))
				expectedStr := make([]string, len(tt.expected))
				for i, v := range got {
					gotStr[i] = fmt.Sprintf("%v", v)
				}
				for i, v := range tt.expected {
					expectedStr[i] = fmt.Sprintf("%v", v)
				}
				sort.Strings(gotStr)
				sort.Strings(expectedStr)
				if !reflect.DeepEqual(gotStr, expectedStr) {
					t.Errorf("WildcardSegment.Evaluate() = %v, want %v", gotStr, expectedStr)
				}
			} else if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("WildcardSegment.Evaluate() = %v, want %v", got, tt.expected)
			}
		})
	}
}
