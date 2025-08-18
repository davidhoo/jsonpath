package segments

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/davidhoo/jsonpath/pkg/errors"
)

// FilterSegment 表示过滤器段
type FilterSegment struct {
	expression string
}

// NewFilterSegment 创建一个新的过滤器段
func NewFilterSegment(expression string) *FilterSegment {
	return &FilterSegment{
		expression: expression,
	}
}

// Evaluate 实现 Segment 接口
func (s *FilterSegment) Evaluate(value interface{}) ([]interface{}, error) {
	// 检查输入值是否为数组
	arr, ok := value.([]interface{})
	if !ok {
		return nil, errors.NewError(errors.ErrEvaluation, "cannot evaluate filter segment on non-array value", s.String())
	}

	var result []interface{}
	for _, item := range arr {
		// 解析表达式
		expr, err := parseFilterExpression(s.expression)
		if err != nil {
			return nil, err
		}

		// 评估表达式
		matches, err := evaluateFilterExpression(expr, item)
		if err != nil {
			return nil, err
		}

		if matches {
			result = append(result, item)
		}
	}

	return result, nil
}

// filterExpression 表示过滤表达式
type filterExpression struct {
	left     string
	operator string
	right    string
}

// parseFilterExpression 解析过滤表达式
func parseFilterExpression(expr string) (*filterExpression, error) {
	// 移除空白字符
	expr = strings.TrimSpace(expr)

	// 支持的运算符
	operators := []string{"==", "!=", ">=", "<=", ">", "<", "=~", "!~"}
	for _, op := range operators {
		if idx := strings.Index(expr, op); idx != -1 {
			left := strings.TrimSpace(expr[:idx])
			right := strings.TrimSpace(expr[idx+len(op):])
			return &filterExpression{
				left:     left,
				operator: op,
				right:    right,
			}, nil
		}
	}

	return nil, errors.NewError(errors.ErrInvalidFilter, "invalid filter expression", expr)
}

// evaluateFilterExpression 评估过滤表达式
func evaluateFilterExpression(expr *filterExpression, value interface{}) (bool, error) {
	// 获取左侧值
	leftVal, err := getValueFromPath(expr.left, value)
	if err != nil {
		return false, err
	}

	// 获取右侧值
	rightVal, err := parseValue(expr.right)
	if err != nil {
		return false, err
	}

	// 根据运算符进行比较
	switch expr.operator {
	case "==":
		return compareValues(leftVal, rightVal) == 0, nil
	case "!=":
		return compareValues(leftVal, rightVal) != 0, nil
	case ">=":
		return compareValues(leftVal, rightVal) >= 0, nil
	case "<=":
		return compareValues(leftVal, rightVal) <= 0, nil
	case ">":
		return compareValues(leftVal, rightVal) > 0, nil
	case "<":
		return compareValues(leftVal, rightVal) < 0, nil
	case "=~":
		return matchRegex(leftVal, rightVal)
	case "!~":
		matches, err := matchRegex(leftVal, rightVal)
		if err != nil {
			return false, err
		}
		return !matches, nil
	default:
		return false, errors.NewError(errors.ErrInvalidFilter, "unsupported operator", expr.operator)
	}
}

// getValueFromPath 从路径获取值
func getValueFromPath(path string, value interface{}) (interface{}, error) {
	// 移除 @ 前缀
	path = strings.TrimPrefix(path, "@")

	// 如果路径为空，返回当前值
	if path == "" {
		return value, nil
	}

	// 分割路径
	parts := strings.Split(path, ".")
	current := value

	for _, part := range parts {
		// 检查当前值是否为对象
		obj, ok := current.(map[string]interface{})
		if !ok {
			return nil, errors.NewError(errors.ErrEvaluation, "cannot access property of non-object value", part)
		}

		// 获取属性值
		val, ok := obj[part]
		if !ok {
			return nil, errors.NewError(errors.ErrEvaluation, "property not found", part)
		}

		current = val
	}

	return current, nil
}

// parseValue 解析值
func parseValue(value string) (interface{}, error) {
	// 尝试解析为数字
	if num, err := strconv.ParseFloat(value, 64); err == nil {
		return num, nil
	}

	// 尝试解析为布尔值
	if value == "true" {
		return true, nil
	}
	if value == "false" {
		return false, nil
	}

	// 尝试解析为 null
	if value == "null" {
		return nil, nil
	}

	// 如果是字符串字面量（带引号）
	if strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") {
		return value[1 : len(value)-1], nil
	}
	if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
		return value[1 : len(value)-1], nil
	}

	// 默认作为字符串处理
	return value, nil
}

// compareValues 比较两个值
func compareValues(a, b interface{}) int {
	// 如果两个值都是数字
	if aNum, ok := a.(float64); ok {
		if bNum, ok := b.(float64); ok {
			if aNum < bNum {
				return -1
			}
			if aNum > bNum {
				return 1
			}
			return 0
		}
	}

	// 如果两个值都是字符串
	if aStr, ok := a.(string); ok {
		if bStr, ok := b.(string); ok {
			return strings.Compare(aStr, bStr)
		}
	}

	// 如果两个值都是布尔值
	if aBool, ok := a.(bool); ok {
		if bBool, ok := b.(bool); ok {
			if aBool == bBool {
				return 0
			}
			if aBool {
				return 1
			}
			return -1
		}
	}

	// 如果两个值都是 null
	if a == nil && b == nil {
		return 0
	}

	// 如果只有一个值是 null
	if a == nil {
		return -1
	}
	if b == nil {
		return 1
	}

	// 其他情况，转换为字符串比较
	return strings.Compare(fmt.Sprintf("%v", a), fmt.Sprintf("%v", b))
}

// matchRegex 使用正则表达式匹配
func matchRegex(value, pattern interface{}) (bool, error) {
	// 将值转换为字符串
	valueStr, ok := value.(string)
	if !ok {
		valueStr = fmt.Sprintf("%v", value)
	}

	// 将模式转换为字符串
	patternStr, ok := pattern.(string)
	if !ok {
		return false, errors.NewError(errors.ErrInvalidFilter, "regex pattern must be a string", fmt.Sprintf("%v", pattern))
	}

	// TODO: 实现正则表达式匹配
	// 这里需要添加正则表达式匹配的实现
	return strings.Contains(valueStr, patternStr), nil
}

// String 实现 Segment 接口
func (s *FilterSegment) String() string {
	return fmt.Sprintf("[?(%s)]", s.expression)
}
