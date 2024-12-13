package jsonpath

import (
	"fmt"
	"strconv"
	"strings"
)

// 解析 JSONPath 表达式
func parse(path string) ([]segment, error) {
	// 处理空路径
	if path == "" {
		return nil, nil
	}

	// 检查并移除 $ 前缀
	if !strings.HasPrefix(path, "$") {
		return nil, fmt.Errorf("path must start with $")
	}
	path = path[1:]

	// 如果路径只有 $，返回空段列表
	if path == "" {
		return nil, nil
	}

	// 如果下一个字符是点，移除它
	if strings.HasPrefix(path, ".") {
		path = path[1:]
	}

	// 处理递归下降
	if strings.HasPrefix(path, ".") {
		return parseRecursive(path[1:])
	}

	// 处理常规路径
	return parseRegular(path)
}

// 解析递归下降路径
func parseRecursive(path string) ([]segment, error) {
	var segments []segment
	segments = append(segments, &recursiveSegment{})

	// 如果路径为空，直接返回
	if path == "" {
		return segments, nil
	}

	// 处理后续路径
	if strings.HasPrefix(path, ".") {
		path = path[1:]
	}

	// 如果还有路径，继续解析
	if path != "" {
		remainingSegments, err := parseRegular(path)
		if err != nil {
			return nil, err
		}
		segments = append(segments, remainingSegments...)
	}

	return segments, nil
}

// 解析常规路径
func parseRegular(path string) ([]segment, error) {
	var segments []segment
	var current string
	var inBracket bool
	var bracketContent string

	for i := 0; i < len(path); i++ {
		char := path[i]

		switch {
		case char == '[':
			if inBracket {
				return nil, fmt.Errorf("nested brackets not allowed")
			}
			if current != "" {
				seg, err := createDotSegment(current)
				if err != nil {
					return nil, err
				}
				segments = append(segments, seg)
				current = ""
			}
			inBracket = true

		case char == ']':
			if !inBracket {
				return nil, fmt.Errorf("unexpected closing bracket")
			}
			seg, err := parseBracketSegment(bracketContent)
			if err != nil {
				return nil, err
			}
			segments = append(segments, seg)
			bracketContent = ""
			inBracket = false

		case char == '.' && !inBracket:
			if current != "" {
				seg, err := createDotSegment(current)
				if err != nil {
					return nil, err
				}
				segments = append(segments, seg)
				current = ""
			}

		default:
			if inBracket {
				bracketContent += string(char)
			} else {
				current += string(char)
			}
		}
	}

	// 处理最后一个段
	if inBracket {
		return nil, fmt.Errorf("unclosed bracket")
	}
	if current != "" {
		seg, err := createDotSegment(current)
		if err != nil {
			return nil, err
		}
		segments = append(segments, seg)
	}

	return segments, nil
}

// 创建点表示法段
func createDotSegment(name string) (segment, error) {
	if name == "*" {
		return &wildcardSegment{}, nil
	}
	return &nameSegment{name: name}, nil
}

// 解析方括号段
func parseBracketSegment(content string) (segment, error) {
	// 处理通配符
	if content == "*" {
		return &wildcardSegment{}, nil
	}

	// 处理过滤器表达式
	if strings.HasPrefix(content, "?") {
		return parseFilterSegment(content[1:])
	}

	// 处理多索引选择
	if strings.Contains(content, ",") {
		return parseMultiIndexSegment(content)
	}

	// 处理切片表达式
	if strings.Contains(content, ":") {
		return parseSliceSegment(content)
	}

	// 处理索引或名称
	return parseIndexOrName(content)
}

// 解析过滤器表达式
func parseFilterSegment(content string) (segment, error) {
	// 检查语法
	if !strings.HasPrefix(content, "(@") || !strings.HasSuffix(content, ")") {
		return nil, fmt.Errorf("invalid filter syntax: %s", content)
	}

	// 提取过滤器内容
	content = content[2 : len(content)-1]

	// 解析操作符和值
	operators := []string{"<=", ">=", "==", "!=", "<", ">"}
	var operator string
	var field string
	var value string

	for _, op := range operators {
		if parts := strings.Split(content, op); len(parts) == 2 {
			field = strings.TrimSpace(parts[0])
			value = strings.TrimSpace(parts[1])
			operator = op
			break
		}
	}

	if operator == "" {
		return nil, fmt.Errorf("invalid filter operator: %s", content)
	}

	// 从字段中移除 @ 前缀
	if strings.HasPrefix(field, "@.") {
		field = strings.TrimPrefix(field, "@.")
	} else if strings.HasPrefix(field, "@") {
		field = ""
	}

	// 解析值
	numValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid filter value: %s", value)
	}

	return &filterSegment{
		field:    field,
		operator: operator,
		value:    numValue,
	}, nil
}

// 解析多索引选择
func parseMultiIndexSegment(content string) (segment, error) {
	parts := strings.Split(content, ",")
	indices := make([]int, 0, len(parts))

	for _, part := range parts {
		idx, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return nil, fmt.Errorf("invalid index: %s", part)
		}
		indices = append(indices, idx)
	}

	return &multiIndexSegment{indices: indices}, nil
}

// 解析切片表达式
func parseSliceSegment(content string) (segment, error) {
	parts := strings.Split(content, ":")
	if len(parts) > 3 {
		return nil, fmt.Errorf("invalid slice syntax")
	}

	slice := &sliceSegment{start: 0, end: 0, step: 1}

	// 解析起始索引
	if parts[0] != "" {
		start, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid start index: %s", parts[0])
		}
		slice.start = start
	}

	// 解析结束索引
	if len(parts) > 1 && parts[1] != "" {
		end, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid end index: %s", parts[1])
		}
		slice.end = end
	}

	// 解析步长
	if len(parts) > 2 && parts[2] != "" {
		step, err := strconv.Atoi(parts[2])
		if err != nil {
			return nil, fmt.Errorf("invalid step: %s", parts[2])
		}
		if step == 0 {
			return nil, fmt.Errorf("step cannot be zero")
		}
		slice.step = step
	}

	return slice, nil
}

// 解析索引或名称
func parseIndexOrName(content string) (segment, error) {
	// 处理函数调用
	if strings.HasSuffix(content, ")") {
		return parseFunctionCall(content)
	}

	// 尝试解析为数字索引
	if idx, err := strconv.Atoi(content); err == nil {
		return &indexSegment{index: idx}, nil
	}

	// 处理字符串字面量
	if strings.HasPrefix(content, "'") && strings.HasSuffix(content, "'") {
		return &nameSegment{name: content[1 : len(content)-1]}, nil
	}

	return &nameSegment{name: content}, nil
}

// 解析函数调用
func parseFunctionCall(content string) (segment, error) {
	// 找到函数名和参数列表
	openParen := strings.Index(content, "(")
	if openParen == -1 {
		return nil, fmt.Errorf("invalid function call syntax: missing opening parenthesis")
	}

	name := content[:openParen]
	argsStr := content[openParen+1 : len(content)-1]

	// ���析参数
	var args []interface{}
	if argsStr != "" {
		// 简单参数解析，暂时只支持数字和字符串
		argParts := strings.Split(argsStr, ",")
		args = make([]interface{}, len(argParts))
		for i, arg := range argParts {
			arg = strings.TrimSpace(arg)
			// 尝试解析为数字
			if num, err := strconv.ParseFloat(arg, 64); err == nil {
				args[i] = num
				continue
			}
			// 处理字符串参数
			if strings.HasPrefix(arg, "'") && strings.HasSuffix(arg, "'") {
				args[i] = arg[1 : len(arg)-1]
				continue
			}
			return nil, fmt.Errorf("unsupported argument type: %s", arg)
		}
	}

	return &functionSegment{name: name, args: args}, nil
}
