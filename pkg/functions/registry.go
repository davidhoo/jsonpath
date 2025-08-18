package functions

import (
	"fmt"

	"github.com/davidhoo/jsonpath/pkg/errors"
)

// registry 存储所有注册的函数
var registry = make(map[string]Function)

// Register 注册一个新的函数
func Register(name string, fn Function) {
	registry[name] = fn
}

// GetFunction 获取一个已注册的函数
func GetFunction(name string) (Function, error) {
	fn, ok := registry[name]
	if !ok {
		return nil, errors.NewError(errors.ErrInvalidFunction, fmt.Sprintf("function '%s' not found", name), "")
	}
	return fn, nil
}

// Unregister 注销一个函数
func Unregister(name string) {
	delete(registry, name)
}

// ListFunctions 列出所有注册的函数
func ListFunctions() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}
