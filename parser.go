package jsonpath

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// 解析 JSONPath 表达式
func parse(path string) ([]segment, error) {
	// 处理空路径
	if path == "" {
		return nil, nil
	}

	// 检查是否是函数调用格式: functionName(arg1, arg2, ...)
	// 这是 RFC 9535 的 match() 和 search() 函数语法
	if idx := strings.Index(path, "("); idx > 0 && strings.HasSuffix(path, ")") {
		funcName := path[:idx]
		// 验证函数名是有效的标识符
		if isValidFunctionName(funcName) {
			argsStr := path[idx+1 : len(path)-1]
			return parseTopLevelFunctionCall(funcName, argsStr)
		}
	}

	// 检查并移除 $ 前缀
	if !strings.HasPrefix(path, "$") {
		return nil, NewError(ErrSyntax, "path must start with $", path)
	}
	path = strings.TrimPrefix(path, "$")

	// 如果路径只有 $，返回空段列表
	if path == "" {
		return nil, nil
	}

	// 移除前导点
	path = strings.TrimPrefix(path, ".")

	// 处理递归下降
	if strings.HasPrefix(path, ".") {
		return parseRecursive(path[1:])
	}

	// 处理常规路径
	return parseRegular(path)
}

// isValidFunctionName 检查是否是有效的函数名
func isValidFunctionName(name string) bool {
	if name == "" {
		return false
	}
	for i, r := range name {
		if i == 0 {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_') {
				return false
			}
		} else {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
				return false
			}
		}
	}
	return true
}

// parseTopLevelFunctionCall 解析顶层函数调用
func parseTopLevelFunctionCall(funcName, argsStr string) ([]segment, error) {
	// 解析参数
	args, err := parseFunctionArgsList(argsStr)
	if err != nil {
		return nil, err
	}

	// 创建函数段
	return []segment{&functionSegment{name: funcName, args: args}}, nil
}

// parseFunctionArgsList 解析函数参数列表
func parseFunctionArgsList(argsStr string) ([]interface{}, error) {
	if strings.TrimSpace(argsStr) == "" {
		return nil, nil
	}

	var args []interface{}
	var currentArg strings.Builder
	var inQuote bool
	var quoteChar rune
	depth := 0

	for i := 0; i < len(argsStr); i++ {
		ch := rune(argsStr[i])

		switch {
		case (ch == '\'' || ch == '"') && !inQuote:
			// 开始引号
			inQuote = true
			quoteChar = ch
			// 不将引号写入 currentArg
		case ch == quoteChar && inQuote:
			// 结束引号
			inQuote = false
			quoteChar = 0
			// 不将引号写入 currentArg
		case ch == '\\' && inQuote && i+1 < len(argsStr):
			// 处理转义字符
			nextCh := rune(argsStr[i+1])
			if nextCh == quoteChar || nextCh == '\\' {
				// 转义的引号或反斜杠
				currentArg.WriteRune(nextCh)
				i++ // 跳过下一个字符
			} else {
				// 其他转义序列，保持原样
				currentArg.WriteRune(ch)
			}
		case ch == '(' && !inQuote:
			depth++
			currentArg.WriteRune(ch)
		case ch == ')' && !inQuote:
			depth--
			currentArg.WriteRune(ch)
		case ch == ',' && !inQuote && depth == 0:
			arg := strings.TrimSpace(currentArg.String())
			if arg != "" {
				parsedArg, err := parseSingleFunctionArg(arg)
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
		parsedArg, err := parseSingleFunctionArg(arg)
		if err != nil {
			return nil, err
		}
		args = append(args, parsedArg)
	}

	return args, nil
}

// parseSingleFunctionArg 解析单个函数参数
func parseSingleFunctionArg(arg string) (interface{}, error) {
	arg = strings.TrimSpace(arg)

	// 尝试解析为数字
	if num, err := strconv.ParseFloat(arg, 64); err == nil {
		return num, nil
	}

	// 处理布尔值
	if arg == "true" {
		return true, nil
	}
	if arg == "false" {
		return false, nil
	}

	// 处理 null
	if arg == "null" {
		return nil, nil
	}

	// 如果以 $ 开头，它是一个路径引用
	if strings.HasPrefix(arg, "$") {
		return arg, nil
	}

	// 处理 @ 引用（在过滤器上下文中）
	if strings.HasPrefix(arg, "@") {
		return arg, nil
	}

	// 其他情况都作为字符串处理（包括正则表达式模式）
	return arg, nil
}

// 解析递归下降路径
func parseRecursive(path string) ([]segment, error) {
	var segments []segment
	segments = append(segments, &recursiveSegment{})

	// 如果路径为空，直接返回
	if path == "" {
		return segments, nil
	}

	// 移除前导点
	path = strings.TrimPrefix(path, ".")

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
				return nil, NewError(ErrSyntax, "nested brackets not allowed", path)
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
				return nil, NewError(ErrSyntax, "unexpected closing bracket", path)
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
		return nil, NewError(ErrSyntax, "unclosed bracket", path)
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

	// 处理多索引选择或多字段选择
	if strings.Contains(content, ",") ||
		((strings.HasPrefix(content, "'") && strings.HasSuffix(content, "'")) && strings.Contains(content[1:len(content)-1], "','")) ||
		((strings.HasPrefix(content, "\"") && strings.HasSuffix(content, "\"")) && strings.Contains(content[1:len(content)-1], "\",\"")) {
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
	return expr
}

// expressionParser is a recursive descent parser for filter expressions
type expressionParser struct {
	input string
	pos   int
}

// parseFilterExpression parses a filter expression string into an expression tree
func parseFilterExpression(input string) (exprNode, error) {
	p := &expressionParser{input: input, pos: 0}
	node, err := p.parseOr()
	if err != nil {
		return nil, err
	}
	p.skipSpaces()
	if p.pos < len(p.input) {
		return nil, NewError(ErrInvalidFilter, fmt.Sprintf("unexpected character at position %d: %c", p.pos, p.input[p.pos]), input)
	}
	return node, nil
}

func (p *expressionParser) skipSpaces() {
	for p.pos < len(p.input) && p.input[p.pos] == ' ' {
		p.pos++
	}
}

func (p *expressionParser) parseOr() (exprNode, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}

	children := []exprNode{left}
	for {
		p.skipSpaces()
		if p.pos+1 < len(p.input) && p.input[p.pos:p.pos+2] == "||" {
			p.pos += 2
			right, err := p.parseAnd()
			if err != nil {
				return nil, err
			}
			children = append(children, right)
		} else {
			break
		}
	}

	if len(children) == 1 {
		return children[0], nil
	}
	return &orNode{children: children}, nil
}

func (p *expressionParser) parseAnd() (exprNode, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}

	children := []exprNode{left}
	for {
		p.skipSpaces()
		if p.pos+1 < len(p.input) && p.input[p.pos:p.pos+2] == "&&" {
			p.pos += 2
			right, err := p.parseUnary()
			if err != nil {
				return nil, err
			}
			children = append(children, right)
		} else {
			break
		}
	}

	if len(children) == 1 {
		return children[0], nil
	}
	return &andNode{children: children}, nil
}

func (p *expressionParser) parseUnary() (exprNode, error) {
	p.skipSpaces()
	if p.pos < len(p.input) && p.input[p.pos] == '!' {
		p.pos++
		inner, err := p.parsePrimary()
		if err != nil {
			return nil, err
		}
		return negateNode(inner)
	}
	return p.parsePrimary()
}

func (p *expressionParser) parsePrimary() (exprNode, error) {
	p.skipSpaces()

	if p.pos >= len(p.input) {
		return nil, NewError(ErrInvalidFilter, "unexpected end of expression", p.input)
	}

	// Handle parenthesized expression
	if p.input[p.pos] == '(' {
		p.pos++ // skip '('
		node, err := p.parseOr()
		if err != nil {
			return nil, err
		}
		p.skipSpaces()
		if p.pos >= len(p.input) || p.input[p.pos] != ')' {
			return nil, NewError(ErrInvalidFilter, "missing closing parenthesis", p.input)
		}
		p.pos++ // skip ')'
		return node, nil
	}

	// Parse atomic condition (everything until next &&, ||, or unmatched ))
	start := p.pos
	depth := 0
	inQuotes := false
	inSingleQuotes := false

	for p.pos < len(p.input) {
		ch := p.input[p.pos]

		if ch == '"' && !inSingleQuotes {
			inQuotes = !inQuotes
			p.pos++
			continue
		}
		if ch == '\'' && !inQuotes {
			inSingleQuotes = !inSingleQuotes
			p.pos++
			continue
		}

		if inQuotes || inSingleQuotes {
			p.pos++
			continue
		}

		if ch == '(' {
			depth++
			p.pos++
			continue
		}
		if ch == ')' {
			if depth == 0 {
				break
			}
			depth--
			p.pos++
			continue
		}

		// Check for top-level && or ||
		if depth == 0 && p.pos+1 < len(p.input) {
			op := p.input[p.pos : p.pos+2]
			if op == "&&" || op == "||" {
				break
			}
		}

		p.pos++
	}

	condStr := strings.TrimSpace(p.input[start:p.pos])
	if condStr == "" {
		return nil, NewError(ErrInvalidFilter, "empty condition", p.input)
	}

	cond, err := parseFilterCondition(condStr)
	if err != nil {
		return nil, err
	}
	return &conditionNode{cond: cond}, nil
}

// negateNode applies negation to an expression node
func negateNode(node exprNode) (exprNode, error) {
	switch n := node.(type) {
	case *conditionNode:
		newCond := n.cond
		switch newCond.operator {
		case "==":
			newCond.operator = "!="
		case "!=":
			newCond.operator = "=="
		case "<":
			newCond.operator = ">="
		case "<=":
			newCond.operator = ">"
		case ">":
			newCond.operator = "<="
		case ">=":
			newCond.operator = "<"
		case "exists":
			newCond.operator = "not_exists"
		case "not_exists":
			newCond.operator = "exists"
		default:
			return nil, NewError(ErrInvalidFilter, fmt.Sprintf("cannot negate operator: %s", newCond.operator), "")
		}
		return &conditionNode{cond: newCond}, nil
	case *andNode:
		children := make([]exprNode, len(n.children))
		for i, child := range n.children {
			negated, err := negateNode(child)
			if err != nil {
				return nil, err
			}
			children[i] = negated
		}
		return &orNode{children: children}, nil
	case *orNode:
		children := make([]exprNode, len(n.children))
		for i, child := range n.children {
			negated, err := negateNode(child)
			if err != nil {
				return nil, err
			}
			children[i] = negated
		}
		return &andNode{children: children}, nil
	default:
		return nil, NewError(ErrInvalidFilter, "cannot negate expression", "")
	}
}

// 解析过滤器表达式
func parseFilterSegment(content string) (segment, error) {
	// 检查是否是函数调用格式: functionName(arg1, arg2)
	// 这是 RFC 9535 的 match() 和 search() 函数语法
	if idx := strings.Index(content, "("); idx > 0 && strings.HasSuffix(content, ")") {
		funcName := content[:idx]
		if isValidFunctionName(funcName) {
			// 这是一个函数调用，解析为过滤器表达式
			argsStr := content[idx+1 : len(content)-1]
			cond, err := parseFilterFunctionCall(funcName, argsStr)
			if err != nil {
				return nil, NewError(ErrInvalidFilter, fmt.Sprintf("invalid filter syntax: %s", content), content)
			}
			return &filterSegment{expr: &conditionNode{cond: cond}}, nil
		}
	}

	// 检查语法
	if !strings.HasPrefix(content, "@") && !strings.HasPrefix(content, "(@") && !strings.HasPrefix(content, "!") && !strings.HasPrefix(content, "(!") && !strings.HasPrefix(content, "(") {
		return nil, NewError(ErrInvalidFilter, fmt.Sprintf("invalid filter syntax: %s", content), content)
	}

	// 取过滤器内容
	var filterContent string

	switch {
	case strings.HasPrefix(content, "(!"):
		if !strings.HasSuffix(content, ")") {
			return nil, NewError(ErrInvalidFilter, "invalid filter syntax: missing closing parenthesis", content)
		}
		filterContent = content[2 : len(content)-1]
		// Apply De Morgan's laws: !(A && B) => !A || !B, !(A || B) => !A && !B
		expr, err := parseFilterExpression(filterContent)
		if err != nil {
			return nil, NewError(ErrInvalidFilter, fmt.Sprintf("error parsing filter expression: %v", err), content)
		}
		negated, err := negateNode(expr)
		if err != nil {
			return nil, NewError(ErrInvalidFilter, fmt.Sprintf("error negating expression: %v", err), content)
		}
		return &filterSegment{expr: negated}, nil
	case strings.HasPrefix(content, "!@"):
		// Keep the ! in the content for the parser to handle as unary operator
		filterContent = content
	case strings.HasPrefix(content, "(@"):
		if !strings.HasSuffix(content, ")") {
			return nil, NewError(ErrInvalidFilter, "invalid filter syntax: missing closing parenthesis", content)
		}
		filterContent = content[2 : len(content)-1]
	case strings.HasPrefix(content, "@"):
		filterContent = content
	case strings.HasPrefix(content, "!"):
		// Keep the ! in the content for the parser to handle as unary operator
		filterContent = content
		// Check for !(expr) pattern
		if len(content) > 1 && content[1] == '(' {
			if !strings.HasSuffix(content, ")") {
				return nil, NewError(ErrInvalidFilter, "invalid filter syntax: missing closing parenthesis", content)
			}
			// Apply De Morgan's laws for !(expr)
			innerContent := content[2 : len(content)-1]
			expr, err := parseFilterExpression(innerContent)
			if err != nil {
				return nil, NewError(ErrInvalidFilter, fmt.Sprintf("error parsing filter expression: %v", err), content)
			}
			negated, err := negateNode(expr)
			if err != nil {
				return nil, NewError(ErrInvalidFilter, fmt.Sprintf("error negating expression: %v", err), content)
			}
			return &filterSegment{expr: negated}, nil
		}
	case strings.HasPrefix(content, "("):
		if !strings.HasSuffix(content, ")") {
			return nil, NewError(ErrInvalidFilter, "invalid filter syntax: missing closing parenthesis", content)
		}
		filterContent = content[1 : len(content)-1]
	default:
		filterContent = content
	}

	// 标准化表达式
	filterContent = normalizeFilterExpression(filterContent)

	// 解析表达式为树结构
	expr, err := parseFilterExpression(filterContent)
	if err != nil {
		return nil, NewError(ErrInvalidFilter, fmt.Sprintf("error parsing filter expression: %v", err), content)
	}

	return &filterSegment{expr: expr}, nil
}

// 解析过滤器条件
func parseFilterCondition(content string) (filterCondition, error) {
	// 检查是否是函数调用格式: functionName(arg1, arg2)
	// 这是 RFC 9535 的 match() 和 search() 函数语法
	if idx := strings.Index(content, "("); idx > 0 && strings.HasSuffix(content, ")") {
		funcName := content[:idx]
		if isValidFunctionName(funcName) {
			argsStr := content[idx+1 : len(content)-1]
			return parseFilterFunctionCall(funcName, argsStr)
		}
	}

	// 确保条件以 @ 开头
	if !strings.HasPrefix(content, "@") {
		if strings.HasPrefix(content, ".") {
			content = "@" + content
		} else {
			return filterCondition{}, NewError(ErrInvalidFilter, fmt.Sprintf("invalid condition: %s", content), content)
		}
	}

	// 检查是否是旧式方法调用: @.field.match(pattern)
	if strings.Contains(content, ".match(") {
		parts := strings.Split(content, ".match(")
		if len(parts) != 2 || !strings.HasSuffix(parts[1], ")") {
			return filterCondition{}, NewError(ErrInvalidFilter, "invalid match function syntax", content)
		}

		field := strings.TrimSpace(parts[0])
		pattern := strings.TrimSpace(parts[1][:len(parts[1])-1])

		// 验证字段格式
		if !strings.HasPrefix(field, "@") {
			return filterCondition{}, NewError(ErrInvalidFilter, "filter condition must start with @", content)
		}

		// 移除引号
		pattern = strings.Trim(pattern, "\"'")

		return filterCondition{
			field:    strings.TrimPrefix(field, "@."),
			operator: "match",
			value:    pattern,
		}, nil
	}

	// 查找比较操作符
	var operator string
	var operatorIndex int
	var operatorFound bool

	// 按长度排序的操作符列表，确保先匹配较长的操作符
	operators := []string{"<=", ">=", "==", "!=", "<", ">"}
	for _, op := range operators {
		idx := strings.Index(content, op)
		if idx != -1 {
			// 确保这是一个独立的操作符，不是符串值的一部分
			inQuotes := false
			inParens := 0
			isValid := true
			for i := 0; i < idx; i++ {
				switch content[i] {
				case '"':
					inQuotes = !inQuotes
				case '(':
					inParens++
				case ')':
					inParens--
				}
			}
			if inQuotes || inParens > 0 {
				isValid = false
			}
			if isValid {
				operator = op
				operatorIndex = idx
				operatorFound = true
				break
			}
		}
	}

	if !operatorFound {
		// No operator found - validate that this is a valid field path
		field := strings.TrimSpace(content)
		// Check for invalid operator-like characters
		if strings.ContainsAny(field, "=<>!") {
			return filterCondition{}, NewError(ErrInvalidFilter, fmt.Sprintf("no valid operator found in condition: %s", content), content)
		}
		return filterCondition{
			field:    strings.TrimPrefix(field, "@."),
			operator: "exists",
			value:    nil,
		}, nil
	}

	// 分割路径和值
	field := strings.TrimSpace(content[:operatorIndex])
	value := strings.TrimSpace(content[operatorIndex+len(operator):])

	// 解析值
	parsedValue, err := parseFilterValue(value)
	if err != nil {
		return filterCondition{}, NewError(ErrInvalidFilter, fmt.Sprintf("invalid value: %s", value), content)
	}

	return filterCondition{
		field:    strings.TrimPrefix(field, "@."),
		operator: operator,
		value:    parsedValue,
	}, nil
}

// parseFilterFunctionCall 解析过滤器中的函数调用
func parseFilterFunctionCall(funcName, argsStr string) (filterCondition, error) {
	// 解析参数
	args, err := parseFunctionArgsList(argsStr)
	if err != nil {
		return filterCondition{}, NewError(ErrInvalidFilter, fmt.Sprintf("invalid function arguments: %v", err), funcName+"("+argsStr+")")
	}

	// 对于 match 和 search 函数，需要两个参数
	if funcName == "match" || funcName == "search" {
		if len(args) != 2 {
			return filterCondition{}, NewError(ErrInvalidFilter, fmt.Sprintf("%s() requires exactly 2 arguments", funcName), funcName+"("+argsStr+")")
		}

		// 第一个参数是字段路径
		field, ok := args[0].(string)
		if !ok {
			return filterCondition{}, NewError(ErrInvalidFilter, fmt.Sprintf("%s() first argument must be a string", funcName), funcName+"("+argsStr+")")
		}

		// 第二个参数是模式
		pattern, ok := args[1].(string)
		if !ok {
			return filterCondition{}, NewError(ErrInvalidFilter, fmt.Sprintf("%s() second argument must be a string", funcName), funcName+"("+argsStr+")")
		}

		return filterCondition{
			field:    strings.TrimPrefix(field, "@."),
			operator: funcName,
			value:    pattern,
		}, nil
	}

	// 对于其他函数，创建一个通用的函数调用条件
	return filterCondition{
		field:    "",
		operator: "function:" + funcName,
		value:    args,
	}, nil
}

// compareValues compares two values based on the operator
func compareValues(value1 interface{}, operator string, value2 interface{}) (bool, error) {
	// 检查操作符是否有效
	validOperators := map[string]bool{
		"==":    true,
		"!=":    true,
		">":     true,
		"<":     true,
		">=":    true,
		"<=":    true,
		"match": true,
	}
	if !validOperators[operator] {
		return false, fmt.Errorf("invalid operator: %s", operator)
	}

	// 处理 nil 值
	if value1 == nil || value2 == nil {
		switch operator {
		case "==":
			return value1 == value2, nil
		case "!=":
			return value1 != value2, nil
		default:
			return false, fmt.Errorf("invalid operator for nil values: %s", operator)
		}
	}

	// 处理数字类型
	num1, num2, isNum := normalizeNumbers(value1, value2)
	if isNum {
		switch operator {
		case "==":
			return num1 == num2, nil
		case "!=":
			return num1 != num2, nil
		case ">":
			return num1 > num2, nil
		case "<":
			return num1 < num2, nil
		case ">=":
			return num1 >= num2, nil
		case "<=":
			return num1 <= num2, nil
		default:
			return false, fmt.Errorf("invalid operator for numbers: %s", operator)
		}
	}

	// 处理字符串类型
	if str1, ok := value1.(string); ok {
		if str2, ok := value2.(string); ok {
			switch operator {
			case "==":
				return str1 == str2, nil
			case "!=":
				return str1 != str2, nil
			case ">":
				return str1 > str2, nil
			case "<":
				return str1 < str2, nil
			case ">=":
				return str1 >= str2, nil
			case "<=":
				return str1 <= str2, nil
			case "match":
				re, err := regexp.Compile(str2)
				if err != nil {
					return false, fmt.Errorf("invalid regex pattern: %s", str2)
				}
				return re.MatchString(str1), nil
			default:
				return false, fmt.Errorf("invalid operator for strings: %s", operator)
			}
		}
		if operator == "match" {
			return false, fmt.Errorf("pattern must be a string")
		}
	}
	if operator == "match" {
		return false, fmt.Errorf("value must be a string")
	}

	// 处理布尔类型
	if bool1, ok := value1.(bool); ok {
		if bool2, ok := value2.(bool); ok {
			switch operator {
			case "==":
				return bool1 == bool2, nil
			case "!=":
				return bool1 != bool2, nil
			default:
				return false, fmt.Errorf("invalid operator for booleans: %s", operator)
			}
		}
	}

	return false, fmt.Errorf("incompatible types for comparison")
}

// normalizeNumbers 将两个值转换为 float64 类型
func normalizeNumbers(value1, value2 interface{}) (float64, float64, bool) {
	var num1, num2 float64
	var ok1, ok2 bool

	// 尝试将 value1 转换为 float64
	switch v := value1.(type) {
	case float64:
		num1, ok1 = v, true
	case int64:
		num1, ok1 = float64(v), true
	case int:
		num1, ok1 = float64(v), true
	}

	// 尝试将 value2 转换为 float64
	switch v := value2.(type) {
	case float64:
		num2, ok2 = v, true
	case int64:
		num2, ok2 = float64(v), true
	case int:
		num2, ok2 = float64(v), true
	}

	return num1, num2, ok1 && ok2
}

// compareStrings compares two strings using the specified operator
func compareStrings(a string, operator string, b string) bool {
	return standardCompareStrings(a, operator, b)
}

// getFieldValue 获取对象中指定字段的值
func getFieldValue(obj interface{}, field string) (interface{}, error) {
	// 移除 @ 和前导点
	field = strings.TrimPrefix(field, "@")
	field = strings.TrimPrefix(field, ".")

	// 如果字段为空，返回对象本身
	if field == "" {
		return obj, nil
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

// isValidIdentifier 检查字符串是否是有效的标识符
func isValidIdentifier(s string) bool {
	if s == "" {
		return false
	}

	// 检查第一个字符是否是字母或下划线
	first := s[0]
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || first == '_') {
		return false
	}

	// 检查其余字符是否是字母、数字或下划线
	for i := 1; i < len(s); i++ {
		c := s[i]
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}

	return true
}

// 解析多索引选择
func parseMultiIndexSegment(content string) (segment, error) {
	// 检查前导和尾随逗号
	if strings.HasPrefix(content, ",") {
		return nil, NewError(ErrInvalidPath, "leading comma in multi-index segment", content)
	}
	if strings.HasSuffix(content, ",") {
		return nil, NewError(ErrInvalidPath, "trailing comma in multi-index segment", content)
	}

	parts := strings.Split(content, ",")

	// 检查空索引
	for _, part := range parts {
		if strings.TrimSpace(part) == "" {
			return nil, NewError(ErrInvalidPath, "empty index in multi-index segment", content)
		}
	}

	// 检查是否包含字符串字段名
	// 检查是否包含字符串字段名
	hasString := false
	hasQuotedString := false
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		// 检查是否是字符串字段名（带引号）
		if (strings.HasPrefix(trimmed, "'") && strings.HasSuffix(trimmed, "'")) ||
			(strings.HasPrefix(trimmed, "\"") && strings.HasSuffix(trimmed, "\"")) {
			hasString = true
			hasQuotedString = true
			break
		}
		// 检查是否是非数字的字段名
		if _, err := strconv.Atoi(trimmed); err != nil {
			// 如果不是带引号的字符串，但也不是数字，可能是无效索引或不带引号的字段名
			// 我们需要进一步判断
			hasString = true
			break
		}
	}

	// 如果有引号字符串，或者所有非数字部分都是有效的字段名，则作为多字段段处理
	if hasString && hasQuotedString {
		// 有引号字符串，按多字段段处理
		names := make([]string, 0, len(parts))
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			// 处理带引号的字符串
			if (strings.HasPrefix(trimmed, "'") && strings.HasSuffix(trimmed, "'")) ||
				(strings.HasPrefix(trimmed, "\"") && strings.HasSuffix(trimmed, "\"")) {
				names = append(names, trimmed[1:len(trimmed)-1])
			} else {
				// 处理不带引号的字段名
				names = append(names, trimmed)
			}
		}
		return &multiNameSegment{names: names}, nil
	}

	// 检查是否所有部分都是有效的字段名（不带引号但也不是纯数字）
	if hasString && !hasQuotedString {
		// 检查是否所有部分都是数字或者所有部分都是有效的标识符
		allNumbers := true
		allValidNames := true

		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if _, err := strconv.Atoi(trimmed); err != nil {
				// 不是数字
				allNumbers = false
				if !isValidIdentifier(trimmed) {
					allValidNames = false
				}
			} else {
				// 是数字，但在混合情况下不应该作为字段名
				allValidNames = false
			}
		}

		// 如果混合了数字和非数字，这是无效的
		if !allNumbers && !allValidNames {
			return nil, NewError(ErrInvalidPath, "cannot mix numeric indices and field names", content)
		}

		if allValidNames {
			// 所有部分都是有效标识符，按多字段段处理
			names := make([]string, 0, len(parts))
			for _, part := range parts {
				trimmed := strings.TrimSpace(part)
				names = append(names, trimmed)
			}
			return &multiNameSegment{names: names}, nil
		}
	}

	// 否则解析为多索引段
	indices := make([]int, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		idx, err := strconv.Atoi(trimmed)
		if err != nil {
			return nil, NewError(ErrInvalidPath, fmt.Sprintf("invalid index: %s", trimmed), content)
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
	if strings.HasPrefix(content, "'") && strings.HasSuffix(content, "'") && len(content) > 1 {
		return &nameSegment{name: content[1 : len(content)-1]}, nil
	}

	// 处理双引号字符串字面量
	if strings.HasPrefix(content, "\"") && strings.HasSuffix(content, "\"") && len(content) > 1 {
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

	// 检查闭合括号
	if !strings.HasSuffix(content, ")") {
		return nil, fmt.Errorf("invalid function call syntax: missing closing parenthesis")
	}

	name := content[:openParen]
	argsStr := content[openParen+1 : len(content)-1]

	// 解析参数
	args := make([]interface{}, 0)
	if argsStr != "" {
		// 简单参数解析，暂时只支持数字和字符串
		argParts := strings.Split(argsStr, ",")
		for _, arg := range argParts {
			arg = strings.TrimSpace(arg)
			if arg == "" {
				continue // 跳过空参数
			}
			// 尝试解析为数字
			if num, err := strconv.ParseFloat(arg, 64); err == nil {
				args = append(args, num)
				continue
			}
			// 处理字符串参数
			if strings.HasPrefix(arg, "'") && strings.HasSuffix(arg, "'") {
				// 处理转义引号
				str := arg[1 : len(arg)-1]
				str = strings.ReplaceAll(str, "''", "'")
				args = append(args, str)
				continue
			}
			return nil, fmt.Errorf("unsupported argument type: %s", arg)
		}
	}

	return &functionSegment{name: name, args: args}, nil
}

// parseFilterValue parses a filter value string into an appropriate type
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
