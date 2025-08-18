package functions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConcatFunction(t *testing.T) {
	fn := &ConcatFunction{}

	tests := []struct {
		name     string
		args     []interface{}
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "concat strings",
			args:     []interface{}{"hello", " ", "world"},
			expected: "hello world",
			wantErr:  false,
		},
		{
			name:     "concat with numbers",
			args:     []interface{}{"value: ", 42},
			expected: "value: 42",
			wantErr:  false,
		},
		{
			name:     "concat with booleans",
			args:     []interface{}{"status: ", true},
			expected: "status: true",
			wantErr:  false,
		},
		{
			name:     "concat with null",
			args:     []interface{}{"value: ", nil},
			expected: "value: null",
			wantErr:  false,
		},
		{
			name:    "no arguments",
			args:    []interface{}{},
			wantErr: true,
		},
		{
			name:    "single argument",
			args:    []interface{}{"hello"},
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

func TestContainsFunction(t *testing.T) {
	fn := &ContainsFunction{}

	tests := []struct {
		name     string
		args     []interface{}
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "contains substring",
			args:     []interface{}{"hello world", "world"},
			expected: true,
			wantErr:  false,
		},
		{
			name:     "does not contain substring",
			args:     []interface{}{"hello world", "xyz"},
			expected: false,
			wantErr:  false,
		},
		{
			name:     "empty substring",
			args:     []interface{}{"hello world", ""},
			expected: true,
			wantErr:  false,
		},
		{
			name:    "invalid first argument",
			args:    []interface{}{123, "world"},
			wantErr: true,
		},
		{
			name:    "invalid second argument",
			args:    []interface{}{"hello world", 123},
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

func TestStartsWithFunction(t *testing.T) {
	fn := &StartsWithFunction{}

	tests := []struct {
		name     string
		args     []interface{}
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "starts with prefix",
			args:     []interface{}{"hello world", "hello"},
			expected: true,
			wantErr:  false,
		},
		{
			name:     "does not start with prefix",
			args:     []interface{}{"hello world", "world"},
			expected: false,
			wantErr:  false,
		},
		{
			name:     "empty prefix",
			args:     []interface{}{"hello world", ""},
			expected: true,
			wantErr:  false,
		},
		{
			name:    "invalid first argument",
			args:    []interface{}{123, "hello"},
			wantErr: true,
		},
		{
			name:    "invalid second argument",
			args:    []interface{}{"hello world", 123},
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

func TestEndsWithFunction(t *testing.T) {
	fn := &EndsWithFunction{}

	tests := []struct {
		name     string
		args     []interface{}
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "ends with suffix",
			args:     []interface{}{"hello world", "world"},
			expected: true,
			wantErr:  false,
		},
		{
			name:     "does not end with suffix",
			args:     []interface{}{"hello world", "hello"},
			expected: false,
			wantErr:  false,
		},
		{
			name:     "empty suffix",
			args:     []interface{}{"hello world", ""},
			expected: true,
			wantErr:  false,
		},
		{
			name:    "invalid first argument",
			args:    []interface{}{123, "world"},
			wantErr: true,
		},
		{
			name:    "invalid second argument",
			args:    []interface{}{"hello world", 123},
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
