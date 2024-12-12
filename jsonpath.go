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

// 检查段是否包含数组操作
func isArrayOperation(segStr string) bool {
	// 如果不包含方括号，只需检查是否是递归下降
	if !strings.Contains(segStr, "[") {
		return strings.Contains(segStr, "..")
	}

	// 检查是否是简单的单个索引访问
	return strings.Contains(segStr, ":") ||
		strings.Contains(segStr, "*") ||
		strings.Contains(segStr, "?") ||
		strings.Contains(segStr, ",")
}

// 检查是否有数组操作
func (jp *JSONPath) hasArrayOperation() bool {
	if len(jp.segments) <= 1 {
		return false
	}

	// 检查除最后一个段外的所有段
	for _, seg := range jp.segments[:len(jp.segments)-1] {
		if isArrayOperation(seg.String()) {
			return true
		}
	}
	return false
}

// 检查最后一个段是否需要数组结果
func (jp *JSONPath) lastSegmentNeedsArray() bool {
	if len(jp.segments) == 0 {
		return false
	}
	return isArrayOperation(jp.segments[len(jp.segments)-1].String())
}

// 处理单个结果
func processSingleResult(result interface{}) interface{} {
	// 如果是字符串，直接返回
	if str, ok := result.(string); ok {
		return str
	}
	return result
}

// 处理执行结果
func (jp *JSONPath) processResult(current []interface{}) (interface{}, error) {
	// 如果没有段，直接返回原始数据
	if len(jp.segments) == 0 && len(current) > 0 {
		return current[0], nil
	}

	// 确保 current 不为 nil
	if current == nil {
		current = make([]interface{}, 0)
	}

	// 需要返回数组的情况：
	// 1. 结果包含多个元素
	// 2. 最后一个段需要数组结果
	// 3. 路径中包含数组操作
	if len(current) > 1 || jp.lastSegmentNeedsArray() || jp.hasArrayOperation() {
		return current, nil
	}

	// 单个结果的处理
	if len(current) == 1 {
		return processSingleResult(current[0]), nil
	}

	// 空结果返回空数组
	return current, nil
}

// 执行段的评估
func (jp *JSONPath) evaluateSegment(seg segment, current []interface{}) ([]interface{}, error) {
	var nextCurrent []interface{}
	var lastErr error

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
	return nextCurrent, nil
}

// Execute 执行 JSONPath 查询
func (jp *JSONPath) Execute(data interface{}) (interface{}, error) {
	current := []interface{}{data}

	// 执行每个段的查询
	for _, seg := range jp.segments {
		var err error
		current, err = jp.evaluateSegment(seg, current)
		if err != nil {
			return nil, err
		}
	}

	return jp.processResult(current)
}
