package errors

import (
	"testing"
)

func TestError(t *testing.T) {
	tests := []struct {
		name    string
		errType ErrorType
		msg     string
		path    string
		want    string
	}{
		{
			name:    "syntax error",
			errType: ErrSyntax,
			msg:     "invalid syntax",
			path:    "$.test",
			want:    "invalid syntax",
		},
		{
			name:    "invalid path error",
			errType: ErrInvalidPath,
			msg:     "invalid path",
			path:    "$.test",
			want:    "invalid path",
		},
		{
			name:    "invalid filter error",
			errType: ErrInvalidFilter,
			msg:     "invalid filter",
			path:    "$.test",
			want:    "invalid filter",
		},
		{
			name:    "evaluation error",
			errType: ErrEvaluation,
			msg:     "evaluation failed",
			path:    "$.test",
			want:    "evaluation failed",
		},
		{
			name:    "invalid function error",
			errType: ErrInvalidFunction,
			msg:     "invalid function",
			path:    "$.test",
			want:    "invalid function",
		},
		{
			name:    "invalid argument error",
			errType: ErrInvalidArgument,
			msg:     "invalid argument",
			path:    "$.test",
			want:    "invalid argument",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewError(tt.errType, tt.msg, tt.path)
			if err == nil {
				t.Fatal("NewError() returned nil")
			}

			// 测试 Error() 方法
			if got := err.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}

			// 测试类型断言
			if e, ok := err.(*Error); !ok {
				t.Error("NewError() did not return *Error type")
			} else {
				if e.Type != tt.errType {
					t.Errorf("Error.Type = %v, want %v", e.Type, tt.errType)
				}
				if e.Message != tt.msg {
					t.Errorf("Error.Message = %v, want %v", e.Message, tt.msg)
				}
				if e.Path != tt.path {
					t.Errorf("Error.Path = %v, want %v", e.Path, tt.path)
				}
			}
		})
	}
}
