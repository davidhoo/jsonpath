package functions

// Function 表示一个 JSONPath 函数
type Function interface {
	// Call 调用函数并返回结果
	Call(args []interface{}) (interface{}, error)
	// Name 返回函数名称
	Name() string
}

// BuiltinFunction 是 Function 接口的辅助实现类型
type BuiltinFunction struct {
	name     string
	callback func([]interface{}) (interface{}, error)
}

// Call 实现 Function 接口
func (f *BuiltinFunction) Call(args []interface{}) (interface{}, error) {
	return f.callback(args)
}

// Name 实现 Function 接口
func (f *BuiltinFunction) Name() string {
	return f.name
}
