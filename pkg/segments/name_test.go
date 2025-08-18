package segments

import (
	"reflect"
	"testing"
)

func TestNameSegment(t *testing.T) {
	tests := []struct {
		name     string
		segment  *NameSegment
		input    interface{}
		expected []interface{}
		wantErr  bool
	}{
		{
			name:    "simple property access",
			segment: NewNameSegment("name"),
			input: map[string]interface{}{
				"name": "test",
			},
			expected: []interface{}{"test"},
			wantErr:  false,
		},
		{
			name:    "non-existent property",
			segment: NewNameSegment("missing"),
			input: map[string]interface{}{
				"name": "test",
			},
			expected: nil,
			wantErr:  false,
		},
		{
			name:     "nil input",
			segment:  NewNameSegment("name"),
			input:    nil,
			expected: nil,
			wantErr:  false,
		},
		{
			name:     "non-object input",
			segment:  NewNameSegment("name"),
			input:    []interface{}{1, 2, 3},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.segment.Evaluate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NameSegment.Evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("NameSegment.Evaluate() = %v, want %v", got, tt.expected)
			}
		})
	}
}
