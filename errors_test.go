package jsonpath

import "testing"

func TestError(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		wantText string
	}{
		{
			name: "syntax error",
			err: &Error{
				Type:    ErrSyntax,
				Message: "invalid syntax",
				Path:    "$.invalid[",
			},
			wantText: "invalid syntax",
		},
		{
			name: "invalid path",
			err: &Error{
				Type:    ErrInvalidPath,
				Message: "path not found",
				Path:    "$.notexist",
			},
			wantText: "path not found",
		},
		{
			name: "invalid filter",
			err: &Error{
				Type:    ErrInvalidFilter,
				Message: "invalid filter expression",
				Path:    "$[?(@.invalid)]",
			},
			wantText: "invalid filter expression",
		},
		{
			name: "evaluation error",
			err: &Error{
				Type:    ErrEvaluation,
				Message: "cannot evaluate expression",
				Path:    "$.items[(@.price * @.quantity)]",
			},
			wantText: "cannot evaluate expression",
		},
		{
			name: "invalid function",
			err: &Error{
				Type:    ErrInvalidFunction,
				Message: "unknown function",
				Path:    "$.items.unknown()",
			},
			wantText: "unknown function",
		},
		{
			name: "invalid argument",
			err: &Error{
				Type:    ErrInvalidArgument,
				Message: "invalid argument type",
				Path:    "$.items.length(42)",
			},
			wantText: "invalid argument type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.wantText {
				t.Errorf("Error() = %v, want %v", got, tt.wantText)
			}
		})
	}
}

func TestNewError(t *testing.T) {
	tests := []struct {
		name     string
		typ      ErrorType
		msg      string
		path     string
		wantType ErrorType
		wantMsg  string
		wantPath string
	}{
		{
			name:     "create syntax error",
			typ:      ErrSyntax,
			msg:      "invalid syntax",
			path:     "$.invalid[",
			wantType: ErrSyntax,
			wantMsg:  "invalid syntax",
			wantPath: "$.invalid[",
		},
		{
			name:     "create invalid path error",
			typ:      ErrInvalidPath,
			msg:      "path not found",
			path:     "$.notexist",
			wantType: ErrInvalidPath,
			wantMsg:  "path not found",
			wantPath: "$.notexist",
		},
		{
			name:     "create invalid filter error",
			typ:      ErrInvalidFilter,
			msg:      "invalid filter expression",
			path:     "$[?(@.invalid)]",
			wantType: ErrInvalidFilter,
			wantMsg:  "invalid filter expression",
			wantPath: "$[?(@.invalid)]",
		},
		{
			name:     "create evaluation error",
			typ:      ErrEvaluation,
			msg:      "cannot evaluate expression",
			path:     "$.items[(@.price * @.quantity)]",
			wantType: ErrEvaluation,
			wantMsg:  "cannot evaluate expression",
			wantPath: "$.items[(@.price * @.quantity)]",
		},
		{
			name:     "create invalid function error",
			typ:      ErrInvalidFunction,
			msg:      "unknown function",
			path:     "$.items.unknown()",
			wantType: ErrInvalidFunction,
			wantMsg:  "unknown function",
			wantPath: "$.items.unknown()",
		},
		{
			name:     "create invalid argument error",
			typ:      ErrInvalidArgument,
			msg:      "invalid argument type",
			path:     "$.items.length(42)",
			wantType: ErrInvalidArgument,
			wantMsg:  "invalid argument type",
			wantPath: "$.items.length(42)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewError(tt.typ, tt.msg, tt.path)
			if err, ok := got.(*Error); ok {
				if err.Type != tt.wantType {
					t.Errorf("NewError().Type = %v, want %v", err.Type, tt.wantType)
				}
				if err.Message != tt.wantMsg {
					t.Errorf("NewError().Message = %v, want %v", err.Message, tt.wantMsg)
				}
				if err.Path != tt.wantPath {
					t.Errorf("NewError().Path = %v, want %v", err.Path, tt.wantPath)
				}
			} else {
				t.Errorf("NewError() returned %T, want *Error", got)
			}
		})
	}
}
