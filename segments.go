package jsonpath

import (
	"fmt"
	"strconv"
	"strings"
)

// 段类型定义
type segment interface {
	evaluate(value interface{}) ([]interface{}, error)
	String() string
}

// 名称段
type nameSegment struct {
	name string
}

func (s *nameSegment) evaluate(value interface{}) ([]interface{}, error) {
	// 处理对象
	if obj, ok := value.(map[string]interface{}); ok {
		if val, exists := obj[s.name]; exists {
			return []interface{}{val}, nil
		}
		return []interface{}{}, nil
	}
	return nil, fmt.Errorf("value is not an object")
}

func (s *nameSegment) String() string {
	return s.name
}

// 索引段
type indexSegment struct {
	index int
}

func (s *indexSegment) evaluate(value interface{}) ([]interface{}, error) {
	arr, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("value is not an array")
	}

	idx := s.normalizeIndex(len(arr))
	if idx < 0 || idx >= len(arr) {
		return []interface{}{}, nil
	}

	return []interface{}{arr[idx]}, nil
}

func (s *indexSegment) normalizeIndex(length int) int {
	if s.index < 0 {
		return length + s.index
	}
	return s.index
}

func (s *indexSegment) String() string {
	return fmt.Sprintf("[%d]", s.index)
}

// 通配符段
type wildcardSegment struct{}

func (s *wildcardSegment) evaluate(value interface{}) ([]interface{}, error) {
	switch v := value.(type) {
	case []interface{}:
		return v, nil
	case map[string]interface{}:
		return mapToArray(v), nil
	default:
		return nil, fmt.Errorf("value is neither array nor object")
	}
}

func mapToArray(m map[string]interface{}) []interface{} {
	result := make([]interface{}, 0, len(m))
	for _, v := range m {
		result = append(result, v)
	}
	return result
}

func (s *wildcardSegment) String() string {
	return "*"
}

// 递归下降段
type recursiveSegment struct{}

func (s *recursiveSegment) evaluate(value interface{}) ([]interface{}, error) {
	var result []interface{}
	err := s.recursiveCollect(value, &result)
	return result, err
}

func (s *recursiveSegment) recursiveCollect(value interface{}, result *[]interface{}) error {
	switch v := value.(type) {
	case []interface{}:
		return s.collectFromArray(v, result)
	case map[string]interface{}:
		return s.collectFromObject(v, result)
	default:
		return nil
	}
}

func (s *recursiveSegment) collectFromArray(arr []interface{}, result *[]interface{}) error {
	for _, item := range arr {
		*result = append(*result, item)
		if err := s.recursiveCollect(item, result); err != nil {
			return err
		}
	}
	return nil
}

func (s *recursiveSegment) collectFromObject(obj map[string]interface{}, result *[]interface{}) error {
	for _, value := range obj {
		*result = append(*result, value)
		if err := s.recursiveCollect(value, result); err != nil {
			return err
		}
	}
	return nil
}

func (s *recursiveSegment) String() string {
	return ".."
}

// 切片段
type sliceSegment struct {
	start, end, step int
}

// 计算切片范围的实际索引
func calculateIndex(idx, length int) int {
	if idx < 0 {
		idx = length + idx
	}
	if idx < 0 {
		idx = 0
	}
	if idx > length {
		idx = length
	}
	return idx
}

// 计算切片步长
func calculateStep(step int) int {
	if step == 0 {
		return 1
	}
	return step
}

// 规范化切片范围
func (s *sliceSegment) normalizeRange(length int) (start, end, step int) {
	// 处理步长
	step = calculateStep(s.step)

	// 处理起始索引
	start = s.start
	if start == 0 {
		if step > 0 {
			start = 0
		} else {
			start = length - 1
		}
	} else {
		start = calculateIndex(start, length)
	}

	// 处理结束索引
	end = s.end
	if end == 0 {
		if step > 0 {
			end = length
		} else {
			end = -1
		}
	} else {
		end = calculateIndex(end, length)
	}

	return start, end, step
}

// 根据步长生成索引序列
func generateIndices(start, end, step int) []int {
	var indices []int
	if step > 0 {
		for i := start; i < end; i += step {
			indices = append(indices, i)
		}
	} else {
		for i := start; i > end; i += step {
			indices = append(indices, i)
		}
	}
	return indices
}

// 从数组中获取指定索引的元素
func getArrayElements(arr []interface{}, indices []int) []interface{} {
	var result []interface{}
	for _, idx := range indices {
		if idx >= 0 && idx < len(arr) {
			result = append(result, arr[idx])
		}
	}
	return result
}

// 评估切片表达式
func (s *sliceSegment) evaluate(value interface{}) ([]interface{}, error) {
	// 检查值是否为数组
	arr, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("value is not an array")
	}

	// 获取规范化的范围
	start, end, step := s.normalizeRange(len(arr))

	// 生成索引序列
	indices := generateIndices(start, end, step)

	// 获取结果元素
	return getArrayElements(arr, indices), nil
}

// 字符串表示
func (s *sliceSegment) String() string {
	return fmt.Sprintf("[%d:%d:%d]", s.start, s.end, s.step)
}

// 过滤器段
type filterSegment struct {
	field    string
	operator string
	value    float64
}

func (s *filterSegment) evaluate(value interface{}) ([]interface{}, error) {
	arr, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("filter can only be applied to array")
	}

	var result []interface{}
	for _, item := range arr {
		match, err := s.matchCondition(item)
		if err != nil {
			continue // 跳过错误的项
		}
		if match {
			result = append(result, item)
		}
	}

	return result, nil
}

func (s *filterSegment) matchCondition(item interface{}) (bool, error) {
	// 检查项是否为对象
	obj, ok := item.(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("filter item must be object")
	}

	// 获取字段值
	fieldValue, ok := obj[s.field]
	if !ok {
		return false, fmt.Errorf("field %s not found", s.field)
	}

	// 转换字段值为数字
	var numValue float64
	switch v := fieldValue.(type) {
	case float64:
		numValue = v
	case int:
		numValue = float64(v)
	default:
		return false, fmt.Errorf("field %s is not a number", s.field)
	}

	// 比较值
	switch s.operator {
	case "<":
		return numValue < s.value, nil
	case "<=":
		return numValue <= s.value, nil
	case ">":
		return numValue > s.value, nil
	case ">=":
		return numValue >= s.value, nil
	case "==":
		return numValue == s.value, nil
	case "!=":
		return numValue != s.value, nil
	default:
		return false, fmt.Errorf("unsupported operator: %s", s.operator)
	}
}

func (s *filterSegment) String() string {
	return fmt.Sprintf("[?(@.%s %s %v)]", s.field, s.operator, s.value)
}

// 多索引段
type multiIndexSegment struct {
	indices []int
}

func (s *multiIndexSegment) evaluate(value interface{}) ([]interface{}, error) {
	arr, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("multi-index can only be applied to array")
	}

	var result []interface{}
	length := len(arr)

	for _, idx := range s.indices {
		// 处理负索引
		if idx < 0 {
			idx = length + idx
		}

		// 检查索引范围
		if idx >= 0 && idx < length {
			result = append(result, arr[idx])
		}
	}

	return result, nil
}

func (s *multiIndexSegment) String() string {
	indices := make([]string, len(s.indices))
	for i, idx := range s.indices {
		indices[i] = strconv.Itoa(idx)
	}
	return fmt.Sprintf("[%s]", strings.Join(indices, ","))
}
