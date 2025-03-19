package functions

import (
	"errors"
	"testing"
)

func TestBuiltinFunction(t *testing.T) {
	tests := []struct {
		name     string
		function *BuiltinFunction
		args     []interface{}
		want     interface{}
		wantErr  error
	}{
		{
			name: "successful function call",
			function: &BuiltinFunction{
				name: "successful function call",
				callback: func(args []interface{}) (interface{}, error) {
					return "success", nil
				},
			},
			args:    []interface{}{"arg1", "arg2"},
			want:    "success",
			wantErr: nil,
		},
		{
			name: "function call with error",
			function: &BuiltinFunction{
				name: "function call with error",
				callback: func(args []interface{}) (interface{}, error) {
					return nil, errors.New("test error")
				},
			},
			args:    []interface{}{"arg1"},
			want:    nil,
			wantErr: errors.New("test error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 测试 Name 方法
			if got := tt.function.Name(); got != tt.name {
				t.Errorf("Name() = %v, want %v", got, tt.name)
			}

			// 测试 Call 方法
			got, err := tt.function.Call(tt.args)
			if err != nil && tt.wantErr != nil {
				if err.Error() != tt.wantErr.Error() {
					t.Errorf("Call() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("Call() unexpected error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("Call() = %v, want %v", got, tt.want)
			}
		})
	}
}
