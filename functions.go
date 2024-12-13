package jsonpath

import (
	"encoding/json"
	"fmt"
	"reflect"
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
	"values": &builtinFunction{
		name: "values",
		callback: func(args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("values() requires exactly 1 argument")
			}

			// 确保参数是对象
			obj, ok := args[0].(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("values() argument must be an object")
			}

			// 获取所有键并排序，以确保值的顺序一致
			keys := make([]string, 0, len(obj))
			for k := range obj {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			// 按键的顺序获取值
			values := make([]interface{}, len(keys))
			for i, k := range keys {
				values[i] = obj[k]
			}

			return values, nil
		},
	},
	"count": &builtinFunction{
		name: "count",
		callback: func(args []interface{}) (interface{}, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("count() requires exactly 2 arguments: array and value")
			}

			// 确保第一个参数是数组
			arr, ok := args[0].([]interface{})
			if !ok {
				return nil, fmt.Errorf("count() first argument must be an array")
			}

			// 计算匹配值的数量
			count := 0
			for _, item := range arr {
				if reflect.DeepEqual(item, args[1]) {
					count++
				}
			}

			return float64(count), nil
		},
	},
	"min": &builtinFunction{
		name: "min",
		callback: func(args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("min() requires exactly 1 argument")
			}

			// 确保参数是数组
			arr, ok := args[0].([]interface{})
			if !ok {
				return nil, fmt.Errorf("min() argument must be an array")
			}

			if len(arr) == 0 {
				return nil, fmt.Errorf("min() cannot be applied to an empty array")
			}

			// 找到第一个数值作为初始值
			var minVal float64
			initialized := false

			for _, item := range arr {
				var num float64
				switch v := item.(type) {
				case float64:
					num = v
				case int:
					num = float64(v)
				case json.Number:
					var err error
					num, err = v.Float64()
					if err != nil {
						continue // 跳过无效的数值
					}
				default:
					continue // 跳过非数值类型
				}

				if !initialized {
					minVal = num
					initialized = true
					continue
				}

				if num < minVal {
					minVal = num
				}
			}

			if !initialized {
				return nil, fmt.Errorf("min() no valid numbers in array")
			}

			return minVal, nil
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
