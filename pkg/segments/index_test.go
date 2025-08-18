package segments

import (
	"reflect"
	"testing"
)

func TestIndexSegment(t *testing.T) {
	tests := []struct {
		name     string
		segment  *IndexSegment
		input    interface{}
		expected []interface{}
		wantErr  bool
	}{
		{
			name:     "simple array access",
			segment:  NewIndexSegment(1),
			input:    []interface{}{1, 2, 3},
			expected: []interface{}{2},
			wantErr:  false,
		},
		{
			name:     "negative index",
			segment:  NewIndexSegment(-1),
			input:    []interface{}{1, 2, 3},
			expected: []interface{}{3},
			wantErr:  false,
		},
		{
			name:     "out of bounds index",
			segment:  NewIndexSegment(3),
			input:    []interface{}{1, 2, 3},
			expected: nil,
			wantErr:  false,
		},
		{
			name:     "nil input",
			segment:  NewIndexSegment(0),
			input:    nil,
			expected: nil,
			wantErr:  false,
		},
		{
			name:     "non-array input",
			segment:  NewIndexSegment(0),
			input:    map[string]interface{}{"key": "value"},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.segment.Evaluate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("IndexSegment.Evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("IndexSegment.Evaluate() = %v, want %v", got, tt.expected)
			}
		})
	}
}
