package jsonpath

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// ErrFunctionNotFound 表示函数未找到
var ErrFunctionNotFound = fmt.Errorf("function not found")

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

	// 分割参数
	var args []interface{}
	var currentArg strings.Builder
	var inQuote bool
	var quoteChar rune
	var inObject int

	for _, ch := range argsStr {
		switch {
		case ch == '\'' || ch == '"':
			if !inQuote {
				inQuote = true
				quoteChar = ch
			} else if quoteChar == ch {
				inQuote = false
			}
			currentArg.WriteRune(ch)
		case ch == '{':
			inObject++
			currentArg.WriteRune(ch)
		case ch == '}':
			inObject--
			currentArg.WriteRune(ch)
		case ch == ',' && !inQuote && inObject == 0:
			// 处理当前参数
			arg := strings.TrimSpace(currentArg.String())
			if arg != "" {
				parsedArg, err := parseSingleArg(arg)
				if err != nil {
					return nil, err
				}
				args = append(args, parsedArg)
			}
			currentArg.Reset()
		default:
			currentArg.WriteRune(ch)
		}
	}

	// 处理最后一个参数
	arg := strings.TrimSpace(currentArg.String())
	if arg != "" {
		parsedArg, err := parseSingleArg(arg)
		if err != nil {
			return nil, err
		}
		args = append(args, parsedArg)
	}

	return args, nil
}

// 解析单个参数
func parseSingleArg(arg string) (interface{}, error) {
	// 处理数字
	if num, err := strconv.ParseFloat(arg, 64); err == nil {
		return num, nil
	}

	// 处理字符串（带引号）
	if (strings.HasPrefix(arg, "'") && strings.HasSuffix(arg, "'")) ||
		(strings.HasPrefix(arg, "\"") && strings.HasSuffix(arg, "\"")) {
		return arg[1 : len(arg)-1], nil
	}

	// 处理对象（JSON格式）
	if strings.HasPrefix(arg, "{") && strings.HasSuffix(arg, "}") {
		var obj interface{}
		if err := json.Unmarshal([]byte(arg), &obj); err == nil {
			return obj, nil
		}
	}

	return nil, fmt.Errorf("invalid argument format: %s", arg)
}

func (s *nameSegment) evaluate(value interface{}) ([]interface{}, error) {
	// 处理函数调用
	if strings.Contains(s.name, "(") {
		// 检查函数调用语法
		openParen := strings.Index(s.name, "(")
		closeParen := strings.LastIndex(s.name, ")")
		if openParen == -1 || closeParen == -1 || openParen > closeParen {
			return nil, fmt.Errorf("invalid function call syntax: malformed function call")
		}

		// 解析函数名和参数
		funcName := s.name[:openParen]
		argsStr := s.name[openParen+1 : closeParen]

		// 获取函数
		fn, err := GetFunction(funcName)
		if err != nil {
			return nil, fmt.Errorf("unknown function: %s", funcName)
		}

		// 解析参数
		var args []interface{}
		if argsStr != "" {
			parsedArgs, err := parseFunctionArgs(argsStr)
			if err != nil {
				return nil, fmt.Errorf("invalid argument: %v", err)
			}
			args = append([]interface{}{value}, parsedArgs...)
		} else {
			args = []interface{}{value}
		}

		// 调用函数
		result, err := fn.Call(args)
		if err != nil {
			return nil, fmt.Errorf("invalid argument: %v", err)
		}

		// 确保返回值是正确的类型
		switch v := result.(type) {
		case int:
			result = float64(v)
		case int64:
			result = float64(v)
		case int32:
			result = float64(v)
		case float32:
			result = float64(v)
		}

		return []interface{}{result}, nil
	}

	// 处理对象字段访问
	obj, ok := value.(map[string]interface{})
	if !ok {
		return []interface{}{}, nil
	}

	val, exists := obj[s.name]
	if !exists {
		return []interface{}{}, nil
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
		return []interface{}{}, nil
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
		return []interface{}{}, nil
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
	result = append(result, value)
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
	step = s.step
	if step == 0 {
		step = 1
	}

	// 处理起始索引
	start = s.start
	if start == 0 {
		if step > 0 {
			start = 0
		} else {
			start = length - 1
		}
	} else if start < 0 {
		start = length + start
		if start < 0 {
			if step > 0 {
				start = 0
			} else {
				start = -1
			}
		}
	} else if start >= length {
		if step > 0 {
			start = length
		} else {
			start = length - 1
		}
	}

	// 处理结束索引
	end = s.end
	if end == 0 {
		if step > 0 {
			end = length
		} else {
			end = -1
		}
	} else if end < 0 {
		end = length + end
		if end < 0 {
			if step > 0 {
				end = 0
			} else {
				end = -1
			}
		}
	} else if end >= length {
		if step > 0 {
			end = length
		} else {
			end = length - 1
		}
	}

	return start, end, step
}

// 根据步长生成索引序列
func generateIndices(start, end, step int) []int {
	// 处理零步长
	if step == 0 {
		step = 1
	}

	// 检查无效的范围
	if step > 0 && start >= end {
		return nil
	}
	if step < 0 && start <= end {
		return nil
	}

	// 计算需要生成的索引数量
	count := 0
	if step > 0 {
		count = (end - start + step - 1) / step
	} else {
		count = (start - end - step - 1) / (-step)
	}

	// 预分配切片
	indices := make([]int, 0, count)

	// 生成索引
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
		return []interface{}{}, nil
	}

	// 获取规范化的范围
	start, end, step := s.normalizeRange(len(arr))

	// 生成索引序列
	indices := generateIndices(start, end, step)
	if len(indices) == 0 {
		return nil, nil
	}

	// 获取结果元素
	result := make([]interface{}, len(indices))
	for i, idx := range indices {
		if idx >= 0 && idx < len(arr) {
			result[i] = arr[idx]
		}
	}

	return result, nil
}

// 字符串表示
func (s *sliceSegment) String() string {
	var result strings.Builder
	result.WriteString("[")
	result.WriteString(strconv.Itoa(s.start))
	result.WriteString(":")
	result.WriteString(strconv.Itoa(s.end))
	if s.step != 1 {
		result.WriteString(":")
		result.WriteString(strconv.Itoa(s.step))
	}
	result.WriteString("]")
	return result.String()
}

// 过滤器段
type filterSegment struct {
	expr exprNode
}

// evaluate 评估过滤器段
func (s *filterSegment) evaluate(data interface{}) ([]interface{}, error) {
	// 如果数据不是 map 或 slice，返回空结果
	if data == nil {
		return nil, nil
	}

	// If no expression, return nil
	if s.expr == nil {
		return nil, nil
	}

	// 处理 map 类型
	if m, ok := data.(map[string]interface{}); ok {
		result, err := s.expr.evaluate(m)
		if err != nil {
			return nil, err
		}
		if result {
			return []interface{}{m}, nil
		}
		return nil, nil
	}

	// 处理 slice 类型
	if arr, ok := data.([]interface{}); ok {
		var results []interface{}
		for _, item := range arr {
			result, err := s.expr.evaluate(item)
			if err != nil {
				return nil, err
			}
			if result {
				results = append(results, item)
			}
		}
		return results, nil
	}

	return nil, nil
}

// String 返回过滤器段的字符串表示
func (s *filterSegment) String() string {
	return "[?" + exprToString(s.expr) + "]"
}

func exprToString(node exprNode) string {
	switch n := node.(type) {
	case *conditionNode:
		return n.cond.String()
	case *andNode:
		parts := make([]string, len(n.children))
		for i, child := range n.children {
			parts[i] = exprToString(child)
		}
		return strings.Join(parts, " && ")
	case *orNode:
		parts := make([]string, len(n.children))
		for i, child := range n.children {
			parts[i] = exprToString(child)
		}
		return strings.Join(parts, " || ")
	default:
		return ""
	}
}

// 多索引段
type multiIndexSegment struct {
	indices []int
}

func (s *multiIndexSegment) evaluate(value interface{}) ([]interface{}, error) {
	arr, ok := value.([]interface{})
	if !ok {
		return []interface{}{}, nil
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

	// 准备函数参数
	var args []interface{}
	if len(s.args) == 0 {
		// 如果没有参数，使用当前值作为唯一参数
		args = []interface{}{value}
	} else {
		// 如果有参数，直接使用提供的参数
		args = s.args
	}

	// 调用函数
	result, err := fn.Call(args)
	if err != nil {
		return nil, err
	}

	// 处理数值类型转换
	switch v := result.(type) {
	case int:
		result = float64(v)
	case int32:
		result = float64(v)
	case int64:
		result = float64(v)
	case float32:
		result = float64(v)
	}

	// 如果结果已经是数组，直接返回
	if arr, ok := result.([]interface{}); ok {
		return arr, nil
	}

	// 将单个值包装在数组中返回
	return []interface{}{result}, nil
}

func (s *functionSegment) String() string {
	args := make([]string, len(s.args))
	for i, arg := range s.args {
		switch v := arg.(type) {
		case string:
			args[i] = fmt.Sprintf("'%s'", v)
		case float64:
			args[i] = strconv.FormatFloat(v, 'f', -1, 64)
		case int:
			args[i] = strconv.Itoa(v)
		case bool:
			args[i] = strconv.FormatBool(v)
		case nil:
			args[i] = "null"
		default:
			args[i] = fmt.Sprintf("%v", v)
		}
	}
	return fmt.Sprintf("%s(%s)", s.name, strings.Join(args, ","))
}

// 多字段段
type multiNameSegment struct {
	names []string
}

func (s *multiNameSegment) evaluate(value interface{}) ([]interface{}, error) {
	obj, ok := value.(map[string]interface{})
	if !ok {
		return []interface{}{}, nil
	}

	var result []interface{}

	for _, name := range s.names {
		// 获取字段值
		if val, exists := obj[name]; exists {
			result = append(result, val)
		}
	}

	return result, nil
}

func (s *multiNameSegment) String() string {
	names := make([]string, len(s.names))
	for i, name := range s.names {
		names[i] = fmt.Sprintf("'%s'", name)
	}
	return fmt.Sprintf("[%s]", strings.Join(names, ","))
}
