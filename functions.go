package jsonpath

import "fmt"

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
}

// GetFunction returns a registered function by name
func GetFunction(name string) (Function, error) {
	if f, exists := globalFunctions[name]; exists {
		return f, nil
	}
	return nil, fmt.Errorf("function %s not found", name)
}
