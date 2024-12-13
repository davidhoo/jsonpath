package jsonpath

import (
	"fmt"
	"sort"
)

// Function represents a JSONPath function
type Function interface {
	Call(args []interface{}) (interface{}, error)
	Name() string
}

// builtinFunction is a helper type for implementing Function interface
type builtinFunction struct {
	name     string
	callback func([]interface{}) (interface{}, error)
}

func (f *builtinFunction) Call(args []interface{}) (interface{}, error) {
	return f.callback(args)
}

func (f *builtinFunction) Name() string {
	return f.name
}

// globalFunctions is the registry of built-in functions
var globalFunctions = map[string]Function{
	"length": &builtinFunction{
		name: "length",
		callback: func(args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("length() requires exactly 1 argument")
			}

			// 如果参数是数组，返回数组长度
			if arr, ok := args[0].([]interface{}); ok {
				return float64(len(arr)), nil
			}

			// 如果参数是字符串，返回字符串长度
			if str, ok := args[0].(string); ok {
				return float64(len(str)), nil
			}

			// 如果参数是对象，返回对象的键数量
			if obj, ok := args[0].(map[string]interface{}); ok {
				return float64(len(obj)), nil
			}

			return nil, fmt.Errorf("length() argument must be string, array, or object")
		},
	},
	"keys": &builtinFunction{
		name: "keys",
		callback: func(args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("keys() requires exactly 1 argument")
			}

			// 确保参数是对象
			obj, ok := args[0].(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("keys() argument must be an object")
			}

			// 获取所有键并排序
			keys := make([]string, 0, len(obj))
			for k := range obj {
				keys = append(keys, k)
			}
			sort.Strings(keys) // 按字母顺序排序

			// 转换为 interface{} 切片
			result := make([]interface{}, len(keys))
			for i, k := range keys {
				result[i] = k
			}

			return result, nil
		},
	},
}

// GetFunction returns a registered function by name
func GetFunction(name string) (Function, error) {
	if f, exists := globalFunctions[name]; exists {
		return f, nil
	}
	return nil, fmt.Errorf("function %s not found", name)
}
