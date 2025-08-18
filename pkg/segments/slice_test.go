package segments

import (
	"reflect"
	"testing"
)

func TestSliceSegment(t *testing.T) {
	tests := []struct {
		name     string
		segment  *SliceSegment
		input    interface{}
		expected []interface{}
		wantErr  bool
	}{
		{
			name:     "simple slice",
			segment:  NewSliceSegment(1, 3, 1),
			input:    []interface{}{0, 1, 2, 3, 4},
			expected: []interface{}{1, 2},
			wantErr:  false,
		},
		{
			name:     "slice with step",
			segment:  NewSliceSegment(0, 5, 2),
			input:    []interface{}{0, 1, 2, 3, 4},
			expected: []interface{}{0, 2, 4},
			wantErr:  false,
		},
		{
			name:     "negative step",
			segment:  NewSliceSegment(4, 0, -2),
			input:    []interface{}{0, 1, 2, 3, 4},
			expected: []interface{}{4, 2},
			wantErr:  false,
		},
		{
			name:     "slice to end",
			segment:  NewSliceSegment(2, -1, 1),
			input:    []interface{}{0, 1, 2, 3, 4},
			expected: []interface{}{2, 3, 4},
			wantErr:  false,
		},
		{
			name:     "slice from start",
			segment:  NewSliceSegment(0, 3, 1),
			input:    []interface{}{0, 1, 2, 3, 4},
			expected: []interface{}{0, 1, 2},
			wantErr:  false,
		},
		{
			name:     "negative indices",
			segment:  NewSliceSegment(-3, -1, 1),
			input:    []interface{}{0, 1, 2, 3, 4},
			expected: []interface{}{2, 3},
			wantErr:  false,
		},
		{
			name:     "out of bounds indices",
			segment:  NewSliceSegment(-10, 10, 1),
			input:    []interface{}{0, 1, 2},
			expected: []interface{}{0, 1, 2},
			wantErr:  false,
		},
		{
			name:     "empty array",
			segment:  NewSliceSegment(0, 1, 1),
			input:    []interface{}{},
			expected: []interface{}{},
			wantErr:  false,
		},
		{
			name:     "nil input",
			segment:  NewSliceSegment(0, 1, 1),
			input:    nil,
			expected: nil,
			wantErr:  false,
		},
		{
			name:     "non-array input",
			segment:  NewSliceSegment(0, 1, 1),
			input:    "string",
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.segment.Evaluate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("SliceSegment.Evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("SliceSegment.Evaluate() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSliceSegmentString(t *testing.T) {
	tests := []struct {
		name     string
		segment  *SliceSegment
		expected string
	}{
		{
			name:     "simple slice",
			segment:  NewSliceSegment(1, 3, 1),
			expected: "[1:3]",
		},
		{
			name:     "slice with step",
			segment:  NewSliceSegment(1, 3, 2),
			expected: "[1:3:2]",
		},
		{
			name:     "slice from start",
			segment:  NewSliceSegment(0, 3, 1),
			expected: "[:3]",
		},
		{
			name:     "slice to end",
			segment:  NewSliceSegment(1, -1, 1),
			expected: "[1:]",
		},
		{
			name:     "negative step",
			segment:  NewSliceSegment(3, 0, -1),
			expected: "[3:0:-1]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.segment.String(); got != tt.expected {
				t.Errorf("SliceSegment.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}
