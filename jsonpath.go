package jsonpath

import (
	"fmt"
	"strings"
)

// JSONPath 表示一个编译好的 JSONPath 表达式
type JSONPath struct {
	segments []segment
}

// segment 表示路径中的一个片段
type segment interface {
	evaluate(value interface{}) ([]interface{}, error)
	String() string
}

// 编译 JSONPath 表达式
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

// 执行 JSONPath 查询
func (jp *JSONPath) Execute(data interface{}) (interface{}, error) {
	var current []interface{} = []interface{}{data}
	var lastErr error

	for _, seg := range jp.segments {
		var nextCurrent []interface{}
		for _, item := range current {
			results, err := seg.evaluate(item)
			if err != nil {
				lastErr = err
				continue // 跳过错误，继续处理其他项
			}
			nextCurrent = append(nextCurrent, results...)
		}
		if len(nextCurrent) == 0 && lastErr != nil {
			return nil, lastErr
		}
		current = nextCurrent
	}

	// 如果没有段，直接返回原始数据
	if len(jp.segments) == 0 {
		return data, nil
	}

	// 获取最后一个段的类型
	lastSeg := jp.segments[len(jp.segments)-1]
	lastSegStr := lastSeg.String()

	// 检查是否有任何数组操作（除了单个索引访问）
	hasArrayOp := false
	for _, seg := range jp.segments[:len(jp.segments)-1] {
		segStr := seg.String()
		if strings.Contains(segStr, "[") {
			// 检查是否是简单的单个索引访问
			if !strings.Contains(segStr, ":") &&
				!strings.Contains(segStr, "*") &&
				!strings.Contains(segStr, "?") &&
				!strings.Contains(segStr, ",") {
				continue
			}
			hasArrayOp = true
			break
		} else if strings.Contains(segStr, "..") {
			hasArrayOp = true
			break
		}
	}

	// 以下情况需要返回数组：
	// 1. 结果包含多个元素
	// 2. 使用了数组操作（切片、通配符、过滤器等）
	// 3. 使用了递归下降
	// 4. 路径中包含任何数组操作（除了单个索引访问）
	if len(current) > 1 ||
		strings.Contains(lastSegStr, ":") || // 切片操作
		strings.Contains(lastSegStr, "*") || // 通配符
		strings.Contains(lastSegStr, "?") || // 过滤器
		strings.Contains(lastSegStr, ",") || // 多索引选择
		strings.HasPrefix(lastSegStr, "..") || // 递归下降
		hasArrayOp { // 路径中包含数组操作
		if current == nil {
			current = make([]interface{}, 0)
		}
		return current, nil
	}

	// 对于单个结果，如果是简单的点号访问或单个索引访问，返回单个值
	if len(current) == 1 {
		if str, ok := current[0].(string); ok {
			return str, nil
		}
		return current[0], nil
	}

	// 其他情况返回数组
	if current == nil {
		current = make([]interface{}, 0)
	}
	return current, nil
}
