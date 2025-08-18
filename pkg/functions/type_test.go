package functions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToStringFunction(t *testing.T) {
	fn := &ToStringFunction{}

	tests := []struct {
		name     string
		args     []interface{}
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "string value",
			args:     []interface{}{"hello"},
			expected: "hello",
			wantErr:  false,
		},
		{
			name:     "float value",
			args:     []interface{}{42.5},
			expected: "42.5",
			wantErr:  false,
		},
		{
			name:     "int value",
			args:     []interface{}{42},
			expected: "42",
			wantErr:  false,
		},
		{
			name:     "int64 value",
			args:     []interface{}{int64(42)},
			expected: "42",
			wantErr:  false,
		},
		{
			name:     "bool value",
			args:     []interface{}{true},
			expected: "true",
			wantErr:  false,
		},
		{
			name:     "null value",
			args:     []interface{}{nil},
			expected: "null",
			wantErr:  false,
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

func TestToNumberFunction(t *testing.T) {
	fn := &ToNumberFunction{}

	tests := []struct {
		name     string
		args     []interface{}
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "float value",
			args:     []interface{}{42.5},
			expected: 42.5,
			wantErr:  false,
		},
		{
			name:     "int value",
			args:     []interface{}{42},
			expected: 42.0,
			wantErr:  false,
		},
		{
			name:     "int64 value",
			args:     []interface{}{int64(42)},
			expected: 42.0,
			wantErr:  false,
		},
		{
			name:     "valid string value",
			args:     []interface{}{"42.5"},
			expected: 42.5,
			wantErr:  false,
		},
		{
			name:     "true boolean value",
			args:     []interface{}{true},
			expected: 1.0,
			wantErr:  false,
		},
		{
			name:     "false boolean value",
			args:     []interface{}{false},
			expected: 0.0,
			wantErr:  false,
		},
		{
			name:     "null value",
			args:     []interface{}{nil},
			expected: 0.0,
			wantErr:  false,
		},
		{
			name:    "invalid string value",
			args:    []interface{}{"not a number"},
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

func TestToBooleanFunction(t *testing.T) {
	fn := &ToBooleanFunction{}

	tests := []struct {
		name     string
		args     []interface{}
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "boolean value",
			args:     []interface{}{true},
			expected: true,
			wantErr:  false,
		},
		{
			name:     "true string value",
			args:     []interface{}{"true"},
			expected: true,
			wantErr:  false,
		},
		{
			name:     "false string value",
			args:     []interface{}{"false"},
			expected: false,
			wantErr:  false,
		},
		{
			name:     "non-zero float value",
			args:     []interface{}{42.5},
			expected: true,
			wantErr:  false,
		},
		{
			name:     "zero float value",
			args:     []interface{}{0.0},
			expected: false,
			wantErr:  false,
		},
		{
			name:     "non-zero int value",
			args:     []interface{}{42},
			expected: true,
			wantErr:  false,
		},
		{
			name:     "zero int value",
			args:     []interface{}{0},
			expected: false,
			wantErr:  false,
		},
		{
			name:     "non-zero int64 value",
			args:     []interface{}{int64(42)},
			expected: true,
			wantErr:  false,
		},
		{
			name:     "zero int64 value",
			args:     []interface{}{int64(0)},
			expected: false,
			wantErr:  false,
		},
		{
			name:     "null value",
			args:     []interface{}{nil},
			expected: false,
			wantErr:  false,
		},
		{
			name:    "invalid string value",
			args:    []interface{}{"not a boolean"},
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

func TestIsStringFunction(t *testing.T) {
	fn := &IsStringFunction{}

	tests := []struct {
		name     string
		args     []interface{}
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "string value",
			args:     []interface{}{"hello"},
			expected: true,
			wantErr:  false,
		},
		{
			name:     "non-string value",
			args:     []interface{}{42.5},
			expected: false,
			wantErr:  false,
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

func TestIsNumberFunction(t *testing.T) {
	fn := &IsNumberFunction{}

	tests := []struct {
		name     string
		args     []interface{}
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "float value",
			args:     []interface{}{42.5},
			expected: true,
			wantErr:  false,
		},
		{
			name:     "int value",
			args:     []interface{}{42},
			expected: true,
			wantErr:  false,
		},
		{
			name:     "int64 value",
			args:     []interface{}{int64(42)},
			expected: true,
			wantErr:  false,
		},
		{
			name:     "non-number value",
			args:     []interface{}{"not a number"},
			expected: false,
			wantErr:  false,
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

func TestIsBooleanFunction(t *testing.T) {
	fn := &IsBooleanFunction{}

	tests := []struct {
		name     string
		args     []interface{}
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "boolean value",
			args:     []interface{}{true},
			expected: true,
			wantErr:  false,
		},
		{
			name:     "non-boolean value",
			args:     []interface{}{"not a boolean"},
			expected: false,
			wantErr:  false,
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

func TestIsNullFunction(t *testing.T) {
	fn := &IsNullFunction{}

	tests := []struct {
		name     string
		args     []interface{}
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "null value",
			args:     []interface{}{nil},
			expected: true,
			wantErr:  false,
		},
		{
			name:     "non-null value",
			args:     []interface{}{"not null"},
			expected: false,
			wantErr:  false,
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
