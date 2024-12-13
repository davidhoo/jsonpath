package jsonpath

import (
	"encoding/json"
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

// 解析函数参数
func parseFunctionArgs(argsStr string) ([]interface{}, error) {
	if argsStr == "" {
		return nil, nil
	}

	// 处理数字
	if num, err := strconv.ParseFloat(argsStr, 64); err == nil {
		return []interface{}{num}, nil
	}

	// 处理字符串（带引号）
	if (strings.HasPrefix(argsStr, "'") && strings.HasSuffix(argsStr, "'")) ||
		(strings.HasPrefix(argsStr, "\"") && strings.HasSuffix(argsStr, "\"")) {
		return []interface{}{argsStr[1 : len(argsStr)-1]}, nil
	}

	// 处理对象（JSON格式）
	if strings.HasPrefix(argsStr, "{") && strings.HasSuffix(argsStr, "}") {
		var obj interface{}
		if err := json.Unmarshal([]byte(argsStr), &obj); err == nil {
			return []interface{}{obj}, nil
		}
	}

	return nil, fmt.Errorf("invalid argument format: %s", argsStr)
}

func (s *nameSegment) evaluate(value interface{}) ([]interface{}, error) {
	// 处理函数调用
	if strings.HasSuffix(s.name, ")") {
		// 解析函数名和参数
		openParen := strings.Index(s.name, "(")
		if openParen == -1 {
			return nil, fmt.Errorf("invalid function call syntax")
		}
		funcName := s.name[:openParen]
		argsStr := s.name[openParen+1 : len(s.name)-1]

		// 获取函数
		fn, err := GetFunction(funcName)
		if err != nil {
			return nil, err
		}

		// 解析参数
		var args []interface{}
		if argsStr != "" {
			parsedArgs, err := parseFunctionArgs(argsStr)
			if err != nil {
				return nil, err
			}
			args = append([]interface{}{value}, parsedArgs...)
		} else {
			args = []interface{}{value}
		}

		// 调用函数
		result, err := fn.Call(args)
		if err != nil {
			return nil, err
		}

		return []interface{}{result}, nil
	}

	// 处理对象字段访问
	obj, ok := value.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("value is not an object")
	}

	val, exists := obj[s.name]
	if !exists {
		return nil, fmt.Errorf("field %s not found", s.name)
	}

	return []interface{}{val}, nil
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
	// 确保输入是数组
	arr, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("filter can only be applied to arrays")
	}

	var result []interface{}
	for _, item := range arr {
		matches, err := s.evaluateFilter(item)
		if err != nil {
			continue
		}
		if matches {
			result = append(result, item)
		}
	}

	return result, nil
}

func (s *filterSegment) evaluateFilter(item interface{}) (bool, error) {
	// 获取字段值
	var fieldValue interface{}
	var err error

	if s.field == "" {
		fieldValue = item
	} else {
		fieldValue, err = getFieldValue(item, s.field)
		if err != nil {
			return false, err
		}
	}

	// 转换为数字进行比较
	numValue, ok := fieldValue.(float64)
	if !ok {
		return false, fmt.Errorf("field value is not a number")
	}

	// 执行比较
	switch s.operator {
	case "==":
		return numValue == s.value, nil
	case "!=":
		return numValue != s.value, nil
	case "<":
		return numValue < s.value, nil
	case "<=":
		return numValue <= s.value, nil
	case ">":
		return numValue > s.value, nil
	case ">=":
		return numValue >= s.value, nil
	default:
		return false, fmt.Errorf("unsupported operator: %s", s.operator)
	}
}

// getFieldValue 获取对象中指定字段的值
func getFieldValue(obj interface{}, field string) (interface{}, error) {
	// 移除开头的点
	if strings.HasPrefix(field, ".") {
		field = field[1:]
	}

	// 分割字段路径
	parts := strings.Split(field, ".")
	current := obj

	for _, part := range parts {
		// 确保当前值是对象
		m, ok := current.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("value is not an object")
		}

		// 获取下一级字段值
		var exists bool
		current, exists = m[part]
		if !exists {
			return nil, fmt.Errorf("field %s not found", part)
		}
	}

	return current, nil
}

func (s *filterSegment) String() string {
	return fmt.Sprintf("[?(@%s%s%v)]", s.field, s.operator, s.value)
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

// functionSegment represents a function call in the JSONPath expression
type functionSegment struct {
	name string
	args []interface{}
}

func (s *functionSegment) evaluate(value interface{}) ([]interface{}, error) {
	// 获取函数
	fn, err := GetFunction(s.name)
	if err != nil {
		return nil, err
	}

	// 如果没有参数，使用当前值为参数
	if len(s.args) == 0 {
		s.args = []interface{}{value}
	}

	// 调用函数
	result, err := fn.Call(s.args)
	if err != nil {
		return nil, err
	}

	return []interface{}{result}, nil
}

func (s *functionSegment) String() string {
	args := make([]string, len(s.args))
	for i, arg := range s.args {
		switch v := arg.(type) {
		case string:
			args[i] = fmt.Sprintf("'%s'", v)
		default:
			args[i] = fmt.Sprintf("%v", v)
		}
	}
	return fmt.Sprintf("%s(%s)", s.name, strings.Join(args, ","))
}
