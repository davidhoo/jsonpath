package jsonpath

import "strings"

// evaluator 定义了 JSONPath 表达式的评估器
type evaluator struct {
	segments []segment
}

// newEvaluator 创建一个新的评估器
func newEvaluator(segments []segment) *evaluator {
	return &evaluator{segments: segments}
}

// evaluate 评估 JSONPath 表达式
func (e *evaluator) evaluate(data interface{}) (interface{}, error) {
	current := []interface{}{data}

	// 执行每个段的评估
	for _, seg := range e.segments {
		var err error
		current, err = e.evaluateSegment(seg, current)
		if err != nil {
			return nil, err
		}
	}

	return e.processResult(current)
}

// evaluateSegment 评估单个段
func (e *evaluator) evaluateSegment(seg segment, current []interface{}) ([]interface{}, error) {
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

// processResult 处理评估结果
func (e *evaluator) processResult(current []interface{}) (interface{}, error) {
	// 如果没有段，直接返回���始数据
	if len(e.segments) == 0 && len(current) > 0 {
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
	if len(current) > 1 || e.lastSegmentNeedsArray() || e.hasArrayOperation() {
		return current, nil
	}

	// 单个结果的处理
	if len(current) == 1 {
		return e.processSingleResult(current[0]), nil
	}

	// 空结果返回空数组
	return current, nil
}

// processSingleResult 处理单个结果
func (e *evaluator) processSingleResult(result interface{}) interface{} {
	// 如果是字符串，直接返回
	if str, ok := result.(string); ok {
		return str
	}
	return result
}

// lastSegmentNeedsArray 检查最后一个段是否需要数组结果
func (e *evaluator) lastSegmentNeedsArray() bool {
	if len(e.segments) == 0 {
		return false
	}
	return isArrayOperation(e.segments[len(e.segments)-1].String())
}

// hasArrayOperation 检查是否有数组操作
func (e *evaluator) hasArrayOperation() bool {
	if len(e.segments) <= 1 {
		return false
	}

	// 检查除最后一个段外的所有段
	for _, seg := range e.segments[:len(e.segments)-1] {
		if isArrayOperation(seg.String()) {
			return true
		}
	}
	return false
}

// isArrayOperation 检查段是否包含数组操作
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
