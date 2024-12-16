package jsonpath

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
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

// 标准化过滤器表达式
func normalizeFilterExpression(expr string) string {
	expr = strings.TrimSpace(expr)
	// 移除最外层括号
	if strings.HasPrefix(expr, "(") && strings.HasSuffix(expr, ")") {
		expr = strings.TrimSpace(expr[1 : len(expr)-1])
	}
	return expr
}

// 应用 De Morgan 定律转换表达式
func applyDeMorgan(expr string) (string, error) {
	// 处理空表达式
	if expr == "" {
		return "", fmt.Errorf("empty expression")
	}

	// 如果表达式被括号包围，先移除括号
	if strings.HasPrefix(expr, "(") && strings.HasSuffix(expr, ")") {
		expr = strings.TrimSpace(expr[1 : len(expr)-1])
	}

	// 解析表达式
	conditions, operators, err := splitLogicalOperators(expr)
	if err != nil {
		return "", err
	}

	// 转换每个条件
	for i := range conditions {
		cond := strings.TrimSpace(conditions[i])
		// 如果条件本身是一个复合表达式（带括号），递归处理
		if strings.HasPrefix(cond, "(") && strings.HasSuffix(cond, ")") {
			inner, err := applyDeMorgan(cond)
			if err != nil {
				return "", err
			}
			conditions[i] = inner
			continue
		}

		// 确保条件以 @ 开头
		if !strings.HasPrefix(cond, "@") {
			if strings.HasPrefix(cond, ".") {
				cond = "@" + cond
			} else {
				return "", fmt.Errorf("invalid condition: %s", cond)
			}
		}

		// 反转比较操作符
		for _, op := range []string{"<=", ">=", "==", "!=", "<", ">"} {
			if idx := strings.Index(cond, op); idx != -1 {
				prefix := cond[:idx]
				suffix := cond[idx+len(op):]
				var newOp string
				switch op {
				case "==":
					newOp = "!="
				case "!=":
					newOp = "=="
				case "<":
					newOp = ">="
				case "<=":
					newOp = ">"
				case ">":
					newOp = "<="
				case ">=":
					newOp = "<"
				}
				conditions[i] = prefix + newOp + suffix
				break
			}
		}
	}

	// 转换逻辑运算符
	for i := range operators {
		switch operators[i] {
		case "&&":
			operators[i] = "||"
		case "||":
			operators[i] = "&&"
		}
	}

	// 重建表达式
	var result strings.Builder
	for i, cond := range conditions {
		if i > 0 {
			result.WriteString(" " + operators[i-1] + " ")
		}
		// 如果条件包含空格或逻辑运算符，需要加括号
		if strings.Contains(cond, " ") || strings.Contains(cond, "&&") || strings.Contains(cond, "||") {
			result.WriteString("(" + cond + ")")
		} else {
			result.WriteString(cond)
		}
	}

	return result.String(), nil
}

// 解析过滤器表达��
func parseFilterSegment(content string) (segment, error) {
	// 检查语法
	if !strings.HasPrefix(content, "@") && !strings.HasPrefix(content, "(@") && !strings.HasPrefix(content, "!") && !strings.HasPrefix(content, "(!") {
		return nil, fmt.Errorf("invalid filter syntax: %s", content)
	}

	// 提取过滤器内容
	var isNegated bool
	var filterContent string
	var isCompoundNegation bool

	switch {
	case strings.HasPrefix(content, "(!"):
		if !strings.HasSuffix(content, ")") {
			return nil, fmt.Errorf("invalid filter syntax: missing closing parenthesis")
		}
		isNegated = true
		isCompoundNegation = true
		filterContent = content[2 : len(content)-1]
	case strings.HasPrefix(content, "!@"):
		isNegated = true
		filterContent = content[1:]
	case strings.HasPrefix(content, "(@"):
		if !strings.HasSuffix(content, ")") {
			return nil, fmt.Errorf("invalid filter syntax: missing closing parenthesis")
		}
		filterContent = content[2 : len(content)-1]
	case strings.HasPrefix(content, "@"):
		filterContent = content
	case strings.HasPrefix(content, "!"):
		isNegated = true
		filterContent = content[1:]
		if strings.HasPrefix(filterContent, "(") {
			if !strings.HasSuffix(filterContent, ")") {
				return nil, fmt.Errorf("invalid filter syntax: missing closing parenthesis")
			}
			isCompoundNegation = true
			filterContent = filterContent[1 : len(filterContent)-1]
		}
	default:
		filterContent = content
	}

	// 如果是复合否定表达式，应用 De Morgan 定律
	if isNegated && isCompoundNegation {
		var err error
		filterContent, err = applyDeMorgan(filterContent)
		if err != nil {
			return nil, fmt.Errorf("error applying De Morgan's laws: %v", err)
		}
		isNegated = false // 已经处理过否定
	}

	// 标准化表达式
	filterContent = normalizeFilterExpression(filterContent)

	// 分割逻辑运算符
	conditions, operators, err := splitLogicalOperators(filterContent)
	if err != nil {
		return nil, err
	}

	// 解析每个条件
	filterConditions := make([]filterCondition, 0, len(conditions))
	for i, condStr := range conditions {
		// 处理括号内的条件
		if strings.HasPrefix(condStr, "(") && strings.HasSuffix(condStr, ")") {
			// 递归处理括号内的条件
			innerContent := strings.TrimSpace(condStr[1 : len(condStr)-1])
			innerConditions, innerOperators, err := splitLogicalOperators(innerContent)
			if err != nil {
				return nil, err
			}

			// 解析内部条件
			for j, innerCondStr := range innerConditions {
				cond, err := parseFilterCondition(strings.TrimSpace(innerCondStr))
				if err != nil {
					return nil, err
				}
				filterConditions = append(filterConditions, cond)
				if j < len(innerOperators) {
					operators = append(operators, innerOperators[j])
				}
			}
		} else {
			// 解析普通条件
			cond, err := parseFilterCondition(strings.TrimSpace(condStr))
			if err != nil {
				return nil, err
			}
			// 如果是简单否定表达式，应用否定到第一个条件
			if isNegated && i == 0 {
				switch cond.operator {
				case "==":
					cond.operator = "!="
				case "!=":
					cond.operator = "=="
				case "<":
					cond.operator = ">="
				case "<=":
					cond.operator = ">"
				case ">":
					cond.operator = "<="
				case ">=":
					cond.operator = "<"
				}
				isNegated = false // 已经处理过否定
			}
			filterConditions = append(filterConditions, cond)
		}
	}

	return &filterSegment{
		conditions: filterConditions,
		operators:  operators,
	}, nil
}

// 解析单个过滤条件
func parseFilterCondition(condStr string) (filterCondition, error) {
	condStr = strings.TrimSpace(condStr)

	// 处理括号
	if strings.HasPrefix(condStr, "(") && strings.HasSuffix(condStr, ")") {
		condStr = strings.TrimSpace(condStr[1 : len(condStr)-1])
	}

	// 确保条件以 @ 开头
	if !strings.HasPrefix(condStr, "@") {
		// 如果不是以 @ 开头，尝试添加 @
		if strings.HasPrefix(condStr, ".") {
			condStr = "@" + condStr
		} else {
			return filterCondition{}, fmt.Errorf("filter condition must start with @ or .: %s", condStr)
		}
	}

	// 解析操作符和值
	operators := []string{"<=", ">=", "==", "!=", "<", ">"}
	var operator string
	var field string
	var valueStr string

	// 找到第一个有效的操作符
	opIndex := -1
	opLen := 0
	for _, op := range operators {
		idx := strings.Index(condStr, op)
		if idx != -1 {
			// 确保这是一个独立的操作符，不是字符串值的一部分
			inQuotes := false
			inParens := 0
			for _, ch := range condStr[:idx] {
				if ch == '"' || ch == '\'' {
					inQuotes = !inQuotes
				} else if ch == '(' {
					inParens++
				} else if ch == ')' {
					inParens--
				}
			}
			if !inQuotes && inParens == 0 {
				opIndex = idx
				opLen = len(op)
				operator = op
				break
			}
		}
	}

	if opIndex == -1 {
		return filterCondition{}, fmt.Errorf("invalid filter operator: %s", condStr)
	}

	field = strings.TrimSpace(condStr[:opIndex])
	valueStr = strings.TrimSpace(condStr[opIndex+opLen:])

	// 处理字段名（移除 @ 和可能的前导点）
	field = field[1:] // 移除 @
	if strings.HasPrefix(field, ".") {
		field = field[1:]
	}

	// 解析值
	value, err := parseFilterValue(valueStr)
	if err != nil {
		return filterCondition{}, fmt.Errorf("invalid filter value: %v", err)
	}

	return filterCondition{
		field:    field,
		operator: operator,
		value:    value,
	}, nil
}

// 解析过滤器值
func parseFilterValue(valueStr string) (interface{}, error) {
	valueStr = strings.TrimSpace(valueStr)

	// 处理 null
	if valueStr == "null" {
		return nil, nil
	}

	// 处理布尔值
	if valueStr == "true" {
		return true, nil
	}
	if valueStr == "false" {
		return false, nil
	}

	// 处理字符串（带引号）
	if (strings.HasPrefix(valueStr, "'") && strings.HasSuffix(valueStr, "'")) ||
		(strings.HasPrefix(valueStr, "\"") && strings.HasSuffix(valueStr, "\"")) {
		return valueStr[1 : len(valueStr)-1], nil
	}

	// 尝试解析为数字
	if num, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return num, nil
	}

	// 如果不是其他类型，返回错误
	return nil, fmt.Errorf("invalid value: %s", valueStr)
}

// 检查是否是有效的标识符
func isValidIdentifier(s string) bool {
	if len(s) == 0 {
		return false
	}

	// 第一个字符必须是字母或下划线
	if !unicode.IsLetter(rune(s[0])) && s[0] != '_' {
		return false
	}

	// 其余字符必须是字母、数字或下划线
	for i := 1; i < len(s); i++ {
		if !unicode.IsLetter(rune(s[i])) && !unicode.IsDigit(rune(s[i])) && s[i] != '_' {
			return false
		}
	}

	return true
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
		// 确保当前是对象
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

	// 解析参数
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

// 分割逻辑运算符
func splitLogicalOperators(content string) ([]string, []string, error) {
	var conditions []string
	var operators []string
	var current strings.Builder
	var inQuotes bool
	var quoteChar rune
	var inParens int
	var lastChar rune

	for i := 0; i < len(content); {
		char := rune(content[i])

		// 处理括号
		if char == '(' {
			inParens++
			current.WriteRune(char)
			i++
			lastChar = char
			continue
		}
		if char == ')' {
			inParens--
			current.WriteRune(char)
			i++
			lastChar = char
			continue
		}

		// 处理引号
		if (char == '"' || char == '\'') && (lastChar != '\\') {
			if !inQuotes {
				inQuotes = true
				quoteChar = char
			} else if char == quoteChar {
				inQuotes = false
			}
			current.WriteRune(char)
			i++
			lastChar = char
			continue
		}

		// 在引号或括号内的内容直接添加
		if inQuotes || inParens > 0 {
			current.WriteRune(char)
			i++
			lastChar = char
			continue
		}

		// 检查��辑运算符
		if i+1 < len(content) {
			op := content[i : i+2]
			if (op == "&&" || op == "||") && !inQuotes && inParens == 0 {
				// 添加当前条件
				if current.Len() > 0 {
					conditions = append(conditions, strings.TrimSpace(current.String()))
					current.Reset()
				}
				// 添加运算符
				operators = append(operators, op)
				i += 2
				lastChar = rune(op[1])
				continue
			}
		}

		current.WriteRune(char)
		i++
		lastChar = char
	}

	// 添加最后一个条件
	if current.Len() > 0 {
		conditions = append(conditions, strings.TrimSpace(current.String()))
	}

	// 验证条件和运算符数量
	if len(conditions) != len(operators)+1 {
		return nil, nil, fmt.Errorf("invalid number of conditions and operators")
	}

	return conditions, operators, nil
}
