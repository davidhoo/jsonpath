package segments

import (
	"strconv"
	"strings"

	"github.com/davidhoo/jsonpath/pkg/functions"
)

// FunctionSegment 表示函数调用段
type FunctionSegment struct {
	name      string
	arguments []string
}

// NewFunctionSegment 创建一个新的函数调用段
func NewFunctionSegment(name string, arguments []string) *FunctionSegment {
	return &FunctionSegment{
		name:      name,
		arguments: arguments,
	}
}

// Evaluate 实现 Segment 接口
func (s *FunctionSegment) Evaluate(value interface{}) ([]interface{}, error) {
	// 获取函数
	fn, err := functions.GetFunction(s.name)
	if err != nil {
		return nil, err
	}

	// 解析参数
	args, err := s.parseArguments(value)
	if err != nil {
		return nil, err
	}

	// 调用函数
	result, err := fn.Call(args)
	if err != nil {
		return nil, err
	}

	return []interface{}{result}, nil
}

// parseArguments 解析函数参数
func (s *FunctionSegment) parseArguments(value interface{}) ([]interface{}, error) {
	var args []interface{}
	for _, arg := range s.arguments {
		// 尝试解析为数字
		if num, err := strconv.ParseFloat(arg, 64); err == nil {
			args = append(args, num)
			continue
		}

		// 尝试解析为布尔值
		if arg == "true" {
			args = append(args, true)
			continue
		}
		if arg == "false" {
			args = append(args, false)
			continue
		}

		// 尝试解析为 null
		if arg == "null" {
			args = append(args, nil)
			continue
		}

		// 尝试从当前值中获取属性
		if strings.HasPrefix(arg, "@") {
			val, err := getValueFromPath(arg, value)
			if err != nil {
				return nil, err
			}
			args = append(args, val)
			continue
		}

		// 作为字符串处理
		args = append(args, arg)
	}
	return args, nil
}

// String 实现 Segment 接口
func (s *FunctionSegment) String() string {
	result := s.name + "("
	for i, arg := range s.arguments {
		if i > 0 {
			result += ","
		}
		result += arg
	}
	result += ")"
	return result
}
