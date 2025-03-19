package errors

// ErrorType 表示错误类型
type ErrorType int

const (
	// ErrSyntax 表示 JSONPath 表达式语法错误
	ErrSyntax ErrorType = iota
	// ErrInvalidPath 表示无效的路径格式
	ErrInvalidPath
	// ErrInvalidFilter 表示无效的过滤表达式
	ErrInvalidFilter
	// ErrEvaluation 表示评估过程中的错误
	ErrEvaluation
	// ErrInvalidFunction 表示无效的函数调用
	ErrInvalidFunction
	// ErrInvalidArgument 表示无效的参数
	ErrInvalidArgument
)

// Error 表示一个 JSONPath 错误
type Error struct {
	Type    ErrorType // 错误类型
	Message string    // 错误消息
	Path    string    // 发生错误的 JSONPath 表达式
}

// Error 实现 error 接口
func (e *Error) Error() string {
	return e.Message
}

// NewError 创建一个新的 JSONPath 错误
func NewError(typ ErrorType, msg string, path string) error {
	return &Error{
		Type:    typ,
		Message: msg,
		Path:    path,
	}
}
