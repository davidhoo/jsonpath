package functions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLengthFunction(t *testing.T) {
	fn := &LengthFunction{}

	tests := []struct {
		name     string
		args     []interface{}
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "array length",
			args:     []interface{}{[]interface{}{1, 2, 3}},
			expected: float64(3),
			wantErr:  false,
		},
		{
			name:     "string length",
			args:     []interface{}{"hello"},
			expected: float64(5),
			wantErr:  false,
		},
		{
			name:    "invalid argument",
			args:    []interface{}{123},
			wantErr: true,
		},
		{
			name:    "no arguments",
			args:    []interface{}{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fn.Call(tt.args)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestFirstFunction(t *testing.T) {
	fn := &FirstFunction{}

	tests := []struct {
		name     string
		args     []interface{}
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "first element",
			args:     []interface{}{[]interface{}{1, 2, 3}},
			expected: 1,
			wantErr:  false,
		},
		{
			name:     "empty array",
			args:     []interface{}{[]interface{}{}},
			expected: nil,
			wantErr:  false,
		},
		{
			name:    "invalid argument",
			args:    []interface{}{"not an array"},
			wantErr: true,
		},
		{
			name:    "no arguments",
			args:    []interface{}{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fn.Call(tt.args)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestLastFunction(t *testing.T) {
	fn := &LastFunction{}

	tests := []struct {
		name     string
		args     []interface{}
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "last element",
			args:     []interface{}{[]interface{}{1, 2, 3}},
			expected: 3,
			wantErr:  false,
		},
		{
			name:     "empty array",
			args:     []interface{}{[]interface{}{}},
			expected: nil,
			wantErr:  false,
		},
		{
			name:    "invalid argument",
			args:    []interface{}{"not an array"},
			wantErr: true,
		},
		{
			name:    "no arguments",
			args:    []interface{}{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fn.Call(tt.args)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, got)
		})
	}
}
