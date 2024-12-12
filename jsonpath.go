package jsonpath

import (
	"fmt"
	"strings"
)

// JSONPath 表示一个编译好的 JSONPath 表达式
type JSONPath struct {
	segments []segment
}

// Compile 编译 JSONPath 表达式
func Compile(path string) (*JSONPath, error) {
	if !strings.HasPrefix(path, "$") {
		return nil, fmt.Errorf("path must start with $")
	}

	segments, err := parse(path[1:])
	if err != nil {
		return nil, err
	}

	return &JSONPath{segments: segments}, nil
}

// Execute 执行 JSONPath 查询
func (jp *JSONPath) Execute(data interface{}) (interface{}, error) {
	evaluator := newEvaluator(jp.segments)
	return evaluator.evaluate(data)
}

// String 返回 JSONPath 表达式的字符串表示
func (jp *JSONPath) String() string {
	var builder strings.Builder
	builder.WriteString("$")
	for _, seg := range jp.segments {
		str := seg.String()
		if strings.HasPrefix(str, "[") || strings.HasPrefix(str, "..") {
			builder.WriteString(str)
		} else {
			builder.WriteString(".")
			builder.WriteString(str)
		}
	}
	return builder.String()
}
