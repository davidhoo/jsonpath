package functions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAbsFunction(t *testing.T) {
	fn := &AbsFunction{}

	tests := []struct {
		name     string
		args     []interface{}
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "positive float",
			args:     []interface{}{42.5},
			expected: 42.5,
			wantErr:  false,
		},
		{
			name:     "negative float",
			args:     []interface{}{-42.5},
			expected: 42.5,
			wantErr:  false,
		},
		{
			name:     "positive int",
			args:     []interface{}{42},
			expected: 42.0,
			wantErr:  false,
		},
		{
			name:     "negative int",
			args:     []interface{}{-42},
			expected: 42.0,
			wantErr:  false,
		},
		{
			name:     "positive int64",
			args:     []interface{}{int64(42)},
			expected: 42.0,
			wantErr:  false,
		},
		{
			name:     "negative int64",
			args:     []interface{}{int64(-42)},
			expected: 42.0,
			wantErr:  false,
		},
		{
			name:    "invalid argument",
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

func TestCeilFunction(t *testing.T) {
	fn := &CeilFunction{}

	tests := []struct {
		name     string
		args     []interface{}
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "positive float",
			args:     []interface{}{42.1},
			expected: 43.0,
			wantErr:  false,
		},
		{
			name:     "negative float",
			args:     []interface{}{-42.1},
			expected: -42.0,
			wantErr:  false,
		},
		{
			name:     "integer float",
			args:     []interface{}{42.0},
			expected: 42.0,
			wantErr:  false,
		},
		{
			name:     "positive int",
			args:     []interface{}{42},
			expected: 42.0,
			wantErr:  false,
		},
		{
			name:     "positive int64",
			args:     []interface{}{int64(42)},
			expected: 42.0,
			wantErr:  false,
		},
		{
			name:    "invalid argument",
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

func TestFloorFunction(t *testing.T) {
	fn := &FloorFunction{}

	tests := []struct {
		name     string
		args     []interface{}
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "positive float",
			args:     []interface{}{42.9},
			expected: 42.0,
			wantErr:  false,
		},
		{
			name:     "negative float",
			args:     []interface{}{-42.9},
			expected: -43.0,
			wantErr:  false,
		},
		{
			name:     "integer float",
			args:     []interface{}{42.0},
			expected: 42.0,
			wantErr:  false,
		},
		{
			name:     "positive int",
			args:     []interface{}{42},
			expected: 42.0,
			wantErr:  false,
		},
		{
			name:     "positive int64",
			args:     []interface{}{int64(42)},
			expected: 42.0,
			wantErr:  false,
		},
		{
			name:    "invalid argument",
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

func TestRoundFunction(t *testing.T) {
	fn := &RoundFunction{}

	tests := []struct {
		name     string
		args     []interface{}
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "round up",
			args:     []interface{}{42.5},
			expected: 43.0,
			wantErr:  false,
		},
		{
			name:     "round down",
			args:     []interface{}{42.4},
			expected: 42.0,
			wantErr:  false,
		},
		{
			name:     "integer float",
			args:     []interface{}{42.0},
			expected: 42.0,
			wantErr:  false,
		},
		{
			name:     "positive int",
			args:     []interface{}{42},
			expected: 42.0,
			wantErr:  false,
		},
		{
			name:     "positive int64",
			args:     []interface{}{int64(42)},
			expected: 42.0,
			wantErr:  false,
		},
		{
			name:    "invalid argument",
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

func TestMaxFunction(t *testing.T) {
	fn := &MaxFunction{}

	tests := []struct {
		name     string
		args     []interface{}
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "single number",
			args:     []interface{}{42.5},
			expected: 42.5,
			wantErr:  false,
		},
		{
			name:     "two numbers",
			args:     []interface{}{42.5, 43.5},
			expected: 43.5,
			wantErr:  false,
		},
		{
			name:     "multiple numbers",
			args:     []interface{}{42.5, 43.5, 41.5},
			expected: 43.5,
			wantErr:  false,
		},
		{
			name:     "mixed types",
			args:     []interface{}{42, 43.5, int64(41)},
			expected: 43.5,
			wantErr:  false,
		},
		{
			name:    "invalid argument",
			args:    []interface{}{42.5, "not a number"},
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

func TestMinFunction(t *testing.T) {
	fn := &MinFunction{}

	tests := []struct {
		name     string
		args     []interface{}
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "single number",
			args:     []interface{}{42.5},
			expected: 42.5,
			wantErr:  false,
		},
		{
			name:     "two numbers",
			args:     []interface{}{42.5, 41.5},
			expected: 41.5,
			wantErr:  false,
		},
		{
			name:     "multiple numbers",
			args:     []interface{}{42.5, 41.5, 43.5},
			expected: 41.5,
			wantErr:  false,
		},
		{
			name:     "mixed types",
			args:     []interface{}{42, 41.5, int64(43)},
			expected: 41.5,
			wantErr:  false,
		},
		{
			name:    "invalid argument",
			args:    []interface{}{42.5, "not a number"},
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
