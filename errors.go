package jsonpath

// ErrorType represents the type of error that occurred
type ErrorType int

const (
	ErrSyntax          ErrorType = iota // Syntax error in JSONPath expression
	ErrInvalidPath                      // Invalid path format
	ErrInvalidFilter                    // Invalid filter expression
	ErrEvaluation                       // Error during evaluation
	ErrInvalidFunction                  // Invalid function call
	ErrInvalidArgument                  // Invalid argument
)

// Error represents a JSONPath error
type Error struct {
	Type    ErrorType // Type of error
	Message string    // Error message
	Path    string    // JSONPath expression where error occurred
}

// Error implements the error interface
func (e *Error) Error() string {
	return e.Message
}

// NewError creates a new JSONPath error
func NewError(typ ErrorType, msg string, path string) error {
	return &Error{
		Type:    typ,
		Message: msg,
		Path:    path,
	}
}
