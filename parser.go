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

	// Reject path that is only whitespace after $ (e.g. "$ ")
	if strings.TrimSpace(path) == "" {
		return nil, NewError(ErrSyntax, "invalid path: trailing whitespace after $", "$"+path)
	}

	// 移除前导点
	dotStripped := false
	if strings.HasPrefix(path, ".") {
		path = path[1:]
		dotStripped = true
	}

	// Reject whitespace between dot and name (e.g. "$. a" after stripping $)
	// Only applies when a dot was actually stripped from the path
	if dotStripped && len(path) > 0 && (path[0] == ' ' || path[0] == '\t' || path[0] == '\n' || path[0] == '\r') {
		return nil, NewError(ErrSyntax, "whitespace is not allowed between dot and member name", "$")
	}

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
	parenDepth := 0
	bracketDepth := 0

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
		case ch == '[' && !inQuote:
			bracketDepth++
			currentArg.WriteRune(ch)
		case ch == ']' && !inQuote:
			bracketDepth--
			currentArg.WriteRune(ch)
		case ch == '(' && !inQuote:
			parenDepth++
			currentArg.WriteRune(ch)
		case ch == ')' && !inQuote:
			parenDepth--
			currentArg.WriteRune(ch)
		case ch == ',' && !inQuote && parenDepth == 0 && bracketDepth == 0:
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
	// Reject bare recursive descent: $..
	if path == "" {
		return nil, NewError(ErrSyntax, "bare recursive descent is not allowed", "..")
	}

	// Reject whitespace after recursive descent: $.. a
	if path[0] == ' ' || path[0] == '\t' || path[0] == '\n' || path[0] == '\r' {
		return nil, NewError(ErrSyntax, "whitespace is not allowed between recursive descent and member name", "..")
	}

	var segments []segment
	segments = append(segments, &recursiveSegment{})

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
	afterDot := false
	parenDepth := 0

	// Use rune iteration to properly handle multi-byte UTF-8 characters
	runes := []rune(path)
	i := 0
	for i < len(runes) {
		r := runes[i]
		switch {
		case r == '[':
			if current != "" {
				seg, err := createDotSegment(current)
				if err != nil {
					return nil, err
				}
				segments = append(segments, seg)
				current = ""
			}
			afterDot = false

			// Find the matching closing bracket by counting bracket depth
			// This correctly handles nested brackets in filter expressions
			depth := 0
			j := i + 1
			inQuotes := false
			inSingleQuotes := false
			for j < len(runes) {
				ch := runes[j]
				// Handle escape sequences inside strings
				if (inQuotes || inSingleQuotes) && ch == '\\' && j+1 < len(runes) {
					j += 2 // skip the backslash and the next character
					continue
				}
				if ch == '"' && !inSingleQuotes {
					inQuotes = !inQuotes
				} else if ch == '\'' && !inQuotes {
					inSingleQuotes = !inSingleQuotes
				} else if !inQuotes && !inSingleQuotes {
					if ch == '[' {
						depth++
					} else if ch == ']' {
						if depth == 0 {
							break
						}
						depth--
					}
				}
				j++
			}
			if j >= len(runes) {
				return nil, NewError(ErrSyntax, "unclosed bracket", path)
			}
			bracketContent := string(runes[i+1 : j])
			seg, err := parseBracketSegment(bracketContent)
			if err != nil {
				return nil, err
			}
			segments = append(segments, seg)
			i = j // advance past the closing ']'

		case r == '(' && i == 0:
			// Handle leading parenthesis in dot notation
			parenDepth++
			current += string(r)
			afterDot = false

		case r == '(':
			parenDepth++
			current += string(r)
			afterDot = false

		case r == ')':
			parenDepth--
			current += string(r)
			afterDot = false

		case r == '.' && parenDepth == 0:
			if afterDot {
				// Second dot in ".." → recursive descent
				segments = append(segments, &recursiveSegment{})
				afterDot = false
			} else {
				if current != "" {
					seg, err := createDotSegment(current)
					if err != nil {
						return nil, err
					}
					segments = append(segments, seg)
					current = ""
				}
				afterDot = true
			}

		case (r == ' ' || r == '\t' || r == '\n' || r == '\r') && parenDepth == 0:
			// RFC 9535: whitespace is allowed between root and dot (e.g. "$ .a")
			// but NOT between dot and name (e.g. "$. a" is invalid).
			if afterDot {
				// Whitespace immediately after dot: invalid
				return nil, NewError(ErrSyntax, "whitespace is not allowed between dot and member name", path)
			}
			if current == "" {
				// Leading whitespace (e.g. "$ .a"), skip it
			} else {
				// Whitespace after a name: flush the name as a segment
				seg, err := createDotSegment(current)
				if err != nil {
					return nil, err
				}
				segments = append(segments, seg)
				current = ""
			}

		default:
			current += string(r)
			afterDot = false
		}
		i++
	}

	// 处理最后一个段
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
	// Only validate non-function names (functions are handled by nameSegmentV3.evaluateFunction)
	if !strings.Contains(name, "(") && !isValidMemberName(name) {
		return nil, NewError(ErrSyntax, fmt.Sprintf("invalid member name: %s", name), name)
	}
	return &nameSegment{name: name}, nil
}

// isValidMemberName checks if a name is valid for dot notation per RFC 9535.
// member-name-shorthand = name-first *name-char
// name-first = %x41-5A / "_" / %x61-7A / %x80-10FFFF  (letter / "_" / non-ASCII)
// name-char = name-first / %x30-39  (name-first / digit)
func isValidMemberName(name string) bool {
	if name == "" {
		return false
	}
	for i, r := range name {
		if i == 0 {
			if !((r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || r == '_' || r >= 0x80) {
				return false
			}
		} else {
			if !((r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r >= 0x80) {
				return false
			}
		}
	}
	return true
}

// 解析方括号段
func parseBracketSegment(content string) (segment, error) {
	// RFC 9535: whitespace is allowed around selectors in brackets
	content = strings.TrimSpace(content)

	// Reject empty brackets: $[]
	if content == "" {
		return nil, NewError(ErrSyntax, "empty bracket segment", "")
	}

	// Reject @ outside of filter expression (must be preceded by ?)
	if strings.HasPrefix(content, "@") {
		return nil, NewError(ErrSyntax, "@ is only allowed inside filter expressions", content)
	}

	// Reject $ outside of filter expression (must be preceded by ?)
	if strings.HasPrefix(content, "$") {
		return nil, NewError(ErrSyntax, "$ is only allowed inside filter expressions", content)
	}

	// Reject space-separated indices: $[0 2]
	// After trimming, check if content looks like "0 2" (numbers separated by space)
	// But allow whitespace in slice expressions (e.g., "1 :5:2", "1: 5:2")
	if strings.Contains(content, " ") && !strings.Contains(content, ",") && !strings.HasPrefix(content, "?") && !strings.HasPrefix(content, "'") && !strings.HasPrefix(content, "\"") && !strings.Contains(content, ":") {
		// Check if it looks like space-separated tokens (not just whitespace in a string)
		parts := strings.Fields(content)
		if len(parts) > 1 {
			return nil, NewError(ErrSyntax, "space is not a valid separator in bracket selector, use comma", content)
		}
	}

	// 处理通配符
	if content == "*" {
		return &wildcardSegment{}, nil
	}

	// 处理过滤器表达式
	if strings.HasPrefix(content, "?") {
		// Check if there are multiple selectors (commas at top level)
		if hasTopLevelComma(content) {
			return parseMultiIndexSegment(content)
		}
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

// hasTopLevelComma checks if content has commas at the top level (not inside parentheses, brackets, or quotes)
func hasTopLevelComma(content string) bool {
	inQuotes := false
	inSingleQuotes := false
	parenDepth := 0
	bracketDepth := 0
	for i := 0; i < len(content); i++ {
		ch := content[i]
		if (inQuotes || inSingleQuotes) && ch == '\\' && i+1 < len(content) {
			i++ // skip escaped character
			continue
		}
		if ch == '"' && !inSingleQuotes {
			inQuotes = !inQuotes
		} else if ch == '\'' && !inQuotes {
			inSingleQuotes = !inSingleQuotes
		} else if !inQuotes && !inSingleQuotes {
			if ch == '(' {
				parenDepth++
			} else if ch == ')' {
				parenDepth--
			} else if ch == '[' {
				bracketDepth++
			} else if ch == ']' {
				bracketDepth--
			} else if ch == ',' && parenDepth == 0 && bracketDepth == 0 {
				return true
			}
		}
	}
	return false
}

// splitTopLevel splits content by the given delimiter at the top level (not inside parentheses, brackets, or quotes)
func splitTopLevel(content string, delimiter byte) []string {
	var parts []string
	inQuotes := false
	inSingleQuotes := false
	parenDepth := 0
	bracketDepth := 0
	start := 0
	for i := 0; i < len(content); i++ {
		ch := content[i]
		if (inQuotes || inSingleQuotes) && ch == '\\' && i+1 < len(content) {
			i++ // skip escaped character
			continue
		}
		if ch == '"' && !inSingleQuotes {
			inQuotes = !inQuotes
		} else if ch == '\'' && !inQuotes {
			inSingleQuotes = !inSingleQuotes
		} else if !inQuotes && !inSingleQuotes {
			if ch == '(' {
				parenDepth++
			} else if ch == ')' {
				parenDepth--
			} else if ch == '[' {
				bracketDepth++
			} else if ch == ']' {
				bracketDepth--
			} else if ch == delimiter && parenDepth == 0 && bracketDepth == 0 {
				parts = append(parts, content[start:i])
				start = i + 1
			}
		}
	}
	parts = append(parts, content[start:])
	return parts
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

	// Parse atomic condition (everything until next &&, ||, or unmatched )) or ])
	start := p.pos
	depth := 0
	bracketDepth := 0
	inQuotes := false
	inSingleQuotes := false

	for p.pos < len(p.input) {
		ch := p.input[p.pos]

		// Handle escape sequences inside strings
		if (inQuotes || inSingleQuotes) && ch == '\\' && p.pos+1 < len(p.input) {
			p.pos += 2 // skip backslash and the escaped character
			continue
		}

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
		if ch == '[' {
			bracketDepth++
			p.pos++
			continue
		}
		if ch == ']' {
			if bracketDepth == 0 {
				break
			}
			bracketDepth--
			p.pos++
			continue
		}

		// Check for top-level && or ||
		if depth == 0 && bracketDepth == 0 && p.pos+1 < len(p.input) {
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
		case "match":
			newCond.operator = "not_match"
		case "not_match":
			newCond.operator = "match"
		case "search":
			newCond.operator = "not_search"
		case "not_search":
			newCond.operator = "search"
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

// normalizeFilterWhitespace handles whitespace in filter expressions
// Removes whitespace between ! and ( to support expressions like "!\n(@.a=='b')"
func normalizeFilterWhitespace(content string) string {
	if len(content) < 2 {
		return content
	}
	// Check if content starts with ! followed by whitespace and then (
	if content[0] == '!' {
		// Find the first non-whitespace character after !
		i := 1
		for i < len(content) && (content[i] == ' ' || content[i] == '\t' || content[i] == '\n' || content[i] == '\r') {
			i++
		}
		if i < len(content) && content[i] == '(' {
			// Remove whitespace between ! and (
			return "!" + content[i:]
		}
	}
	return content
}

// 解析过滤器表达式
func parseFilterSegment(content string) (segment, error) {
	// RFC 9535: allow whitespace in filter expressions
	content = strings.TrimSpace(content)

	// 检查是否是完整的函数调用格式: functionName(arg1, arg2)
	// 使用 tryParseFunctionCall 进行正确的括号匹配
	if funcName, argsStr, ok := tryParseFunctionCall(content); ok {
		cond, err := parseFilterFunctionCall(funcName, argsStr)
		if err != nil {
			return nil, NewError(ErrInvalidFilter, fmt.Sprintf("invalid filter syntax: %s", content), content)
		}
		return &filterSegment{expr: &conditionNode{cond: cond}}, nil
	}

	// 检查语法 - 支持 @, $, 函数调用, 和 ! 作为过滤器表达式的开头
	trimmed := strings.TrimSpace(content)
	isFunctionCallExpr := false
	if !strings.HasPrefix(trimmed, "@") && !strings.HasPrefix(trimmed, "$") &&
		!strings.HasPrefix(trimmed, "(@") && !strings.HasPrefix(trimmed, "($") &&
		!strings.HasPrefix(trimmed, "!") && !strings.HasPrefix(trimmed, "(!") &&
		!strings.HasPrefix(trimmed, "(") {
		// Check if it starts with a function name (e.g., count(@..*)>2, length(@.a)>=2)
		if idx := strings.Index(trimmed, "("); idx > 0 {
			funcName := trimmed[:idx]
			if !isValidFunctionName(funcName) {
				return nil, NewError(ErrInvalidFilter, fmt.Sprintf("invalid filter syntax: %s", content), content)
			}
			// Check if there's a top-level comparison operator (not inside function parens)
			// If so, it's a comparison expression, not a standalone function call
			if hasTopLevelOperator(trimmed) {
				// It's a comparison expression with function calls - pass to expression parser
				isFunctionCallExpr = true
			} else {
				// Standalone function call - pass to expression parser
				isFunctionCallExpr = true
			}
		} else {
			return nil, NewError(ErrInvalidFilter, fmt.Sprintf("invalid filter syntax: %s", content), content)
		}
	}

	// Normalize whitespace: remove whitespace between ! and (
	// e.g., "!\n(@.a=='b')" becomes "(!@.a=='b')"
	content = normalizeFilterWhitespace(content)

	// 取过滤器内容
	var filterContent string

	switch {
	case isFunctionCallExpr:
		// Function call expression (e.g., count(@..*)>2) - pass to expression parser
		filterContent = content
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
		// Check if outer parens match properly
		depth := 0
		matchIdx := -1
		for i := 0; i < len(content); i++ {
			if content[i] == '(' {
				depth++
			} else if content[i] == ')' {
				depth--
				if depth == 0 {
					matchIdx = i
					break
				}
			}
		}
		if matchIdx == len(content)-1 {
			// The first '(' matches the last ')' - strip outer (@...)
			filterContent = content[2 : len(content)-1]
		} else {
			// The first '(' doesn't match the last ')' - pass whole content to parser
			// e.g., "(@.a || @.b) && @.c"
			filterContent = content
		}
	case strings.HasPrefix(content, "($"):
		// Check if outer parens match properly
		depth := 0
		matchIdx := -1
		for i := 0; i < len(content); i++ {
			if content[i] == '(' {
				depth++
			} else if content[i] == ')' {
				depth--
				if depth == 0 {
					matchIdx = i
					break
				}
			}
		}
		if matchIdx == len(content)-1 {
			// The first '(' matches the last ')' - strip outer ($...)
			filterContent = content[2 : len(content)-1]
		} else {
			// The first '(' doesn't match the last ')' - pass whole content to parser
			filterContent = content
		}
	case strings.HasPrefix(content, "@"):
		filterContent = content
	case strings.HasPrefix(content, "$"):
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
		// Check if the outer parens actually match (not just first and last char)
		depth := 0
		matchingIdx := -1
		for i := 0; i < len(content); i++ {
			if content[i] == '(' {
				depth++
			} else if content[i] == ')' {
				depth--
				if depth == 0 {
					matchingIdx = i
					break
				}
			}
		}
		if matchingIdx == len(content)-1 {
			// The first '(' matches the last ')' - strip outer parens
			filterContent = content[1 : len(content)-1]
		} else {
			// The first '(' doesn't match the last ')' - pass whole content to parser
			// e.g., "(@.a || @.b) && @.c"
			filterContent = content
		}
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
// isNonSingularQuery checks if a path contains non-singular selectors
// (wildcard, slice, multi-index, descendant) which are not allowed in comparisons
func isNonSingularQuery(field string) bool {
	// Check for wildcard
	if strings.Contains(field, "*") {
		return true
	}
	// Check for descendant (..)
	if strings.Contains(field, "..") {
		return true
	}
	// Check for bracket expressions
	if strings.Contains(field, "[") {
		// Check for slice (:)
		if strings.Contains(field, ":") {
			return true
		}
		// Check for multi-index (,) - but need to be careful about commas in strings
		// Simple check: if there's a comma outside of quotes
		inQuotes := false
		for i := 0; i < len(field); i++ {
			if field[i] == '\'' || field[i] == '"' {
				inQuotes = !inQuotes
			} else if field[i] == ',' && !inQuotes {
				return true
			}
		}
	}
	return false
}

func parseFilterCondition(content string) (filterCondition, error) {
	// 检查是否是完整的函数调用（无比较操作符，如 match(@.a, 'pattern')）
	// 先用括号深度跟踪来确认整个内容是一个函数调用
	if funcName, argsStr, ok := tryParseFunctionCall(content); ok {
		return parseFilterFunctionCall(funcName, argsStr)
	}

	// 查找比较操作符 - 使用括号深度和方括号深度跟踪避免匹配函数参数和嵌套括号内的操作符
	var operator string
	var operatorIndex int
	var operatorFound bool

	// 按长度排序的操作符列表，确保先匹配较长的操作符
	operators := []string{"<=", ">=", "==", "!=", "<", ">"}
	for _, op := range operators {
		// 从左到右查找第一个在顶层（括号深度和方括号深度均为0）的操作符
		inQuotes := false
		inSingleQuotes := false
		parenDepth := 0
		bracketDepth := 0
		for i := 0; i <= len(content)-len(op); i++ {
			ch := content[i]
			if ch == '"' && !inSingleQuotes {
				inQuotes = !inQuotes
				continue
			}
			if ch == '\'' && !inQuotes {
				inSingleQuotes = !inSingleQuotes
				continue
			}
			if inQuotes || inSingleQuotes {
				continue
			}
			if ch == '(' {
				parenDepth++
				continue
			}
			if ch == ')' {
				parenDepth--
				continue
			}
			if ch == '[' {
				bracketDepth++
				continue
			}
			if ch == ']' {
				bracketDepth--
				continue
			}
			if parenDepth == 0 && bracketDepth == 0 && content[i:i+len(op)] == op {
				// Check it's not a false match (e.g., "!=", not part of "!==")
				isValid := true
				// Ensure we're not inside quotes (already handled above)
				if inQuotes || inSingleQuotes {
					isValid = false
				}
				if isValid {
					operator = op
					operatorIndex = i
					operatorFound = true
					break
				}
			}
		}
		if operatorFound {
			break
		}
	}

	if !operatorFound {
		// No operator found - this could be an existence test
		field := strings.TrimSpace(content)
		// Check for invalid operator-like characters at top level (not inside brackets)
		if hasTopLevelOperatorChar(field) {
			return filterCondition{}, NewError(ErrInvalidFilter, fmt.Sprintf("no valid operator found in condition: %s", content), content)
		}

		// Check if this is a function call without comparison (invalid per RFC 9535 for some functions)
		if _, _, isFunc := tryParseFunctionCall(field); isFunc {
			// Standalone function calls like count(@..*), length(@.a), value(@.a)
			// are invalid - they must be used in comparisons or as filter tests
			return filterCondition{}, NewError(ErrInvalidFilter, fmt.Sprintf("invalid filter syntax: %s", content), content)
		}

		// Determine if field is root ($) or current (@)
		isRoot := false
		if strings.HasPrefix(field, "$") {
			isRoot = true
		} else if strings.HasPrefix(field, "@") {
			// ok
		} else if strings.HasPrefix(field, ".") {
			field = "@" + field
		} else {
			return filterCondition{}, NewError(ErrInvalidFilter, fmt.Sprintf("invalid condition: %s", content), content)
		}

		// Strip field prefix (@ or $)
		// Handle @.. (descendant) before @. to avoid incorrect stripping
		if strings.HasPrefix(field, "@..") {
			field = field[1:] // strip just @, leaving ..
		} else if strings.HasPrefix(field, "$..") {
			field = field[1:] // strip just $, leaving ..
		} else {
			field = strings.TrimPrefix(field, "@.")
			field = strings.TrimPrefix(field, "$.")
			field = strings.TrimPrefix(field, "@")
			field = strings.TrimPrefix(field, "$")
		}
		return filterCondition{
			field:    field,
			operator: "exists",
			value:    nil,
			isRoot:   isRoot,
		}, nil
	}

	// 分割路径和值
	left := strings.TrimSpace(content[:operatorIndex])
	right := strings.TrimSpace(content[operatorIndex+len(operator):])

	// Check if the left side is a function call (e.g., length(@.a) == value($..c), count(@..*)>2)
	if leftFuncName, leftArgsStr, isLeftFunc := tryParseFunctionCall(left); isLeftFunc {
		// Reject match/search results being compared with booleans
		if leftFuncName == "match" || leftFuncName == "search" {
			if operator == "==" || operator == "!=" {
				if right == "true" || right == "false" {
					return filterCondition{}, NewError(ErrInvalidFilter, fmt.Sprintf("%s() result cannot be compared with boolean", leftFuncName), content)
				}
			}
		}

		// Validate function arguments
		if leftFuncName == "match" || leftFuncName == "search" {
			// For match/search, only validate param count (not types)
			if err := validateFunctionParamCount(leftFuncName, leftArgsStr); err != nil {
				return filterCondition{}, NewError(ErrInvalidFilter, err.Error(), content)
			}
		} else {
			if err := validateFunctionArgs(leftFuncName, leftArgsStr); err != nil {
				return filterCondition{}, NewError(ErrInvalidFilter, err.Error(), content)
			}
		}

		// Parse the right side value
		parsedValue, err := parseFilterValue(right)
		if err != nil {
			// Right side might also be a function call
			if rightFuncName, rightArgsStr, isRightFunc := tryParseFunctionCall(right); isRightFunc {
				// Validate right side function arguments
				if rightFuncName == "match" || rightFuncName == "search" {
					if err := validateFunctionParamCount(rightFuncName, rightArgsStr); err != nil {
						return filterCondition{}, NewError(ErrInvalidFilter, err.Error(), content)
					}
				} else {
					if err := validateFunctionArgs(rightFuncName, rightArgsStr); err != nil {
						return filterCondition{}, NewError(ErrInvalidFilter, err.Error(), content)
					}
				}
				// Both sides are function calls - store the whole expression for runtime evaluation
				return filterCondition{
					field:    left,
					operator: operator,
					value:    right,
					isRoot:   false,
				}, nil
			}
			return filterCondition{}, NewError(ErrInvalidFilter, fmt.Sprintf("invalid value: %s", right), content)
		}

		return filterCondition{
			field:    left,
			operator: operator,
			value:    parsedValue,
			isRoot:   false,
		}, nil
	}

	// Determine isRoot from the left side
	isRoot := false
	if strings.HasPrefix(left, "$") {
		isRoot = true
	} else if !strings.HasPrefix(left, "@") {
		if strings.HasPrefix(left, ".") {
			left = "@" + left
		} else {
			return filterCondition{}, NewError(ErrInvalidFilter, fmt.Sprintf("invalid condition: %s", content), content)
		}
	}

	// RFC 9535: non-singular paths are not allowed in comparisons
	if isNonSingularQuery(left) {
		return filterCondition{}, NewError(ErrInvalidFilter, "non-singular query is not allowed in comparison", content)
	}

	// 解析值
	parsedValue, err := parseFilterValue(right)
	if err != nil {
		// Right side might be a function call
		if _, _, isRightFunc := tryParseFunctionCall(right); isRightFunc {
			parsedValue = right // Store as string for runtime evaluation
		} else {
			return filterCondition{}, NewError(ErrInvalidFilter, fmt.Sprintf("invalid value: %s", right), content)
		}
	}

	// Strip field prefix (@ or $)
	field := strings.TrimPrefix(left, "@.")
	field = strings.TrimPrefix(field, "$.")
	field = strings.TrimPrefix(field, "@")
	field = strings.TrimPrefix(field, "$")

	return filterCondition{
		field:    field,
		operator: operator,
		value:    parsedValue,
		isRoot:   isRoot,
	}, nil
}

// tryParseFunctionCall attempts to parse content as a function call.
// Returns (funcName, argsStr, true) if successful, ("", "", false) otherwise.
func tryParseFunctionCall(content string) (string, string, bool) {
	content = strings.TrimSpace(content)
	if !strings.HasSuffix(content, ")") {
		return "", "", false
	}
	// Find the opening paren
	idx := strings.Index(content, "(")
	if idx <= 0 {
		return "", "", false
	}
	funcName := content[:idx]
	if !isValidFunctionName(funcName) {
		return "", "", false
	}
	// Find the matching ')' for the '(' at idx
	depth := 0
	matchingIdx := -1
	for i := idx; i < len(content); i++ {
		if content[i] == '(' {
			depth++
		} else if content[i] == ')' {
			depth--
			if depth == 0 {
				matchingIdx = i
				break
			}
		}
	}
	// The matching ')' must be at the end of the string
	if matchingIdx != len(content)-1 {
		return "", "", false
	}
	argsStr := content[idx+1 : len(content)-1]
	return funcName, argsStr, true
}

// hasTopLevelOperator checks if content has a comparison operator at the top level
// (not inside parentheses, brackets, or quotes)
func hasTopLevelOperator(content string) bool {
	operators := []string{"<=", ">=", "==", "!=", "<", ">"}
	for _, op := range operators {
		inQuotes := false
		inSingleQuotes := false
		parenDepth := 0
		bracketDepth := 0
		for i := 0; i <= len(content)-len(op); i++ {
			ch := content[i]
			if (inQuotes || inSingleQuotes) && ch == '\\' && i+1 < len(content) {
				i++
				continue
			}
			if ch == '"' && !inSingleQuotes {
				inQuotes = !inQuotes
				continue
			}
			if ch == '\'' && !inQuotes {
				inSingleQuotes = !inSingleQuotes
				continue
			}
			if inQuotes || inSingleQuotes {
				continue
			}
			if ch == '(' {
				parenDepth++
				continue
			}
			if ch == ')' {
				parenDepth--
				continue
			}
			if ch == '[' {
				bracketDepth++
				continue
			}
			if ch == ']' {
				bracketDepth--
				continue
			}
			if parenDepth == 0 && bracketDepth == 0 && content[i:i+len(op)] == op {
				return true
			}
		}
	}
	return false
}

// hasTopLevelOperatorChar checks if content has operator-like characters (=, <, >, !)
// at the top level (not inside brackets or quotes)
func hasTopLevelOperatorChar(content string) bool {
	inQuotes := false
	inSingleQuotes := false
	bracketDepth := 0
	for i := 0; i < len(content); i++ {
		ch := content[i]
		if (inQuotes || inSingleQuotes) && ch == '\\' && i+1 < len(content) {
			i++
			continue
		}
		if ch == '"' && !inSingleQuotes {
			inQuotes = !inQuotes
			continue
		}
		if ch == '\'' && !inQuotes {
			inSingleQuotes = !inSingleQuotes
			continue
		}
		if inQuotes || inSingleQuotes {
			continue
		}
		if ch == '[' {
			bracketDepth++
			continue
		}
		if ch == ']' {
			bracketDepth--
			continue
		}
		if bracketDepth == 0 && (ch == '=' || ch == '<' || ch == '>' || ch == '!') {
			return true
		}
	}
	return false
}

// validateFunctionArgs validates function arguments per RFC 9535 rules
func validateFunctionArgs(funcName, argsStr string) error {
	args, err := parseFunctionArgsList(argsStr)
	if err != nil {
		return fmt.Errorf("invalid function arguments: %v", err)
	}

	switch funcName {
	case "length":
		if len(args) != 1 {
			return fmt.Errorf("length() requires exactly 1 argument")
		}
		// length() argument must be a singular query (not non-singular)
		// But allow function calls like value() as arguments
		if argStr, ok := args[0].(string); ok {
			// Allow function calls (e.g., value($..c))
			if _, _, isFunc := tryParseFunctionCall(argStr); isFunc {
				// Function calls are allowed as arguments
			} else if isNonSingularQuery(argStr) {
				return fmt.Errorf("length() argument must be a singular query")
			}
		}
	case "count":
		if len(args) != 1 {
			return fmt.Errorf("count() requires exactly 1 argument")
		}
		// count() argument must be a nodelist (path starting with @ or $)
		argStr, ok := args[0].(string)
		if !ok {
			return fmt.Errorf("count() argument must be a nodelist")
		}
		if !strings.HasPrefix(argStr, "@") && !strings.HasPrefix(argStr, "$") {
			return fmt.Errorf("count() argument must be a nodelist")
		}
	case "value":
		if len(args) != 1 {
			return fmt.Errorf("value() requires exactly 1 argument")
		}
		// value() argument must be a nodelist (path starting with @ or $)
		argStr, ok := args[0].(string)
		if !ok {
			return fmt.Errorf("value() argument must be a nodelist")
		}
		if !strings.HasPrefix(argStr, "@") && !strings.HasPrefix(argStr, "$") {
			return fmt.Errorf("value() argument must be a nodelist")
		}
	case "match", "search":
		if len(args) != 2 {
			return fmt.Errorf("%s() requires exactly 2 arguments", funcName)
		}
	}
	return nil
}

// validateFunctionParamCount validates only the parameter count for functions
func validateFunctionParamCount(funcName, argsStr string) error {
	args, err := parseFunctionArgsList(argsStr)
	if err != nil {
		return fmt.Errorf("invalid function arguments: %v", err)
	}

	switch funcName {
	case "length", "count", "value":
		if len(args) != 1 {
			return fmt.Errorf("%s() requires exactly 1 argument", funcName)
		}
	case "match", "search":
		if len(args) != 2 {
			return fmt.Errorf("%s() requires exactly 2 arguments", funcName)
		}
	}
	return nil
}

// parseFilterFunctionCall 解析过滤器中的函数调用
func parseFilterFunctionCall(funcName, argsStr string) (filterCondition, error) {
	// 解析参数
	args, err := parseFunctionArgsList(argsStr)
	if err != nil {
		return filterCondition{}, NewError(ErrInvalidFilter, fmt.Sprintf("invalid function arguments: %v", err), funcName+"("+argsStr+")")
	}

	// Validate function arguments (but not for match/search - they accept any types)
	if funcName != "match" && funcName != "search" {
		if err := validateFunctionArgs(funcName, argsStr); err != nil {
			return filterCondition{}, NewError(ErrInvalidFilter, err.Error(), funcName+"("+argsStr+")")
		}
	}

	// 对于 match 和 search 函数，需要两个参数
	if funcName == "match" || funcName == "search" {
		if len(args) != 2 {
			return filterCondition{}, NewError(ErrInvalidFilter, fmt.Sprintf("%s() requires exactly 2 arguments", funcName), funcName+"("+argsStr+")")
		}

		// 第一个参数是字段路径或值
		field := fmt.Sprintf("%v", args[0])

		// 第二个参数是模式
		pattern := fmt.Sprintf("%v", args[1])

		return filterCondition{
			field:    strings.TrimPrefix(field, "@."),
			operator: funcName,
			value:    pattern,
		}, nil
	}

	// 对于其他函数，创建一个通用的函数调用条件
	// count, length, value must be used in comparisons (not standalone)
	if funcName == "count" || funcName == "length" || funcName == "value" {
		return filterCondition{}, NewError(ErrInvalidFilter, fmt.Sprintf("%s() result must be compared", funcName), funcName+"("+argsStr+")")
	}
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

	// Handle Nothing values
	_, isNothing1 := value1.(Nothing)
	_, isNothing2 := value2.(Nothing)
	if isNothing1 || isNothing2 {
		switch operator {
		case "==":
			return isNothing1 && isNothing2, nil // Nothing == Nothing → true
		case "!=":
			return !(isNothing1 && isNothing2), nil // Nothing != Nothing → false
		default:
			return false, nil
		}
	}

	// 处理 nil 值
	if value1 == nil || value2 == nil {
		switch operator {
		case "==":
			return value1 == value2, nil
		case "!=":
			return value1 != value2, nil
		case "<=", ">=":
			// null <= null and null >= null are true (same type comparison)
			return value1 == value2, nil
		default:
			return false, nil
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
			case "<=", ">=":
				// true <= true, false <= false are true (same type comparison)
				return bool1 == bool2, nil
			default:
				return false, nil
			}
		}
	}

	// 处理数组类型 - RFC 9535 支持深度比较
	if arr1, ok := value1.([]interface{}); ok {
		if arr2, ok := value2.([]interface{}); ok {
			switch operator {
			case "==":
				return deepCompareValues(arr1, arr2), nil
			case "!=":
				return !deepCompareValues(arr1, arr2), nil
			default:
				return false, nil
			}
		}
	}

	// 处理对象类型 - RFC 9535 支持深度比较
	if obj1, ok := value1.(map[string]interface{}); ok {
		if obj2, ok := value2.(map[string]interface{}); ok {
			switch operator {
			case "==":
				return deepCompareValues(obj1, obj2), nil
			case "!=":
				return !deepCompareValues(obj1, obj2), nil
			default:
				return false, nil
			}
		}
	}

	// RFC 9535: comparison between incompatible types
	// == returns false, != returns true
	if operator == "!=" {
		return true, nil
	}
	return false, nil
}

// deepCompareValues performs deep comparison of two values
func deepCompareValues(a, b interface{}) bool {
	switch v1 := a.(type) {
	case []interface{}:
		v2, ok := b.([]interface{})
		if !ok || len(v1) != len(v2) {
			return false
		}
		for i := range v1 {
			if !deepCompareValues(v1[i], v2[i]) {
				return false
			}
		}
		return true
	case map[string]interface{}:
		v2, ok := b.(map[string]interface{})
		if !ok || len(v1) != len(v2) {
			return false
		}
		for k, val1 := range v1 {
			val2, exists := v2[k]
			if !exists || !deepCompareValues(val1, val2) {
				return false
			}
		}
		return true
	default:
		return a == b
	}
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

// 解析多索引选择 - RFC 9535 支持混合选择器类型
func parseMultiIndexSegment(content string) (segment, error) {
	// 检查前导和尾随逗号
	if strings.HasPrefix(content, ",") {
		return nil, NewError(ErrInvalidPath, "leading comma in multi-index segment", content)
	}
	if strings.HasSuffix(content, ",") {
		return nil, NewError(ErrInvalidPath, "trailing comma in multi-index segment", content)
	}

	parts := splitTopLevel(content, ',')

	// 检查空索引
	for _, part := range parts {
		if strings.TrimSpace(part) == "" {
			return nil, NewError(ErrInvalidPath, "empty index in multi-index segment", content)
		}
	}

	// RFC 9535: 每个部分独立解析，支持混合选择器类型
	// 先尝试解析每个部分
	selectors := make([]segment, 0, len(parts))
	allIndices := true
	allNames := true

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)

		// 检查是否是过滤器表达式
		if strings.HasPrefix(trimmed, "?") {
			filter, err := parseFilterSegment(trimmed[1:])
			if err != nil {
				return nil, err
			}
			selectors = append(selectors, filter)
			allIndices = false
			allNames = false
			continue
		}

		// 检查是否是通配符
		if trimmed == "*" {
			selectors = append(selectors, &wildcardSegment{})
			allIndices = false
			allNames = false
			continue
		}

		// 检查是否是切片
		if strings.Contains(trimmed, ":") {
			slice, err := parseSliceSegment(trimmed)
			if err != nil {
				return nil, err
			}
			selectors = append(selectors, slice)
			allIndices = false
			allNames = false
			continue
		}

		// 检查是否是带引号的字符串
		if (strings.HasPrefix(trimmed, "'") && strings.HasSuffix(trimmed, "'")) ||
			(strings.HasPrefix(trimmed, "\"") && strings.HasSuffix(trimmed, "\"")) {
			quoteChar := trimmed[0]
			name := trimmed[1 : len(trimmed)-1]
			if !validateQuotedString(name, quoteChar) {
				return nil, NewError(ErrSyntax, fmt.Sprintf("invalid string literal: %s", trimmed), trimmed)
			}
			unescaped, err := unescapeString(name, quoteChar)
			if err != nil {
				return nil, NewError(ErrSyntax, fmt.Sprintf("invalid string literal: %s", trimmed), trimmed)
			}
			selectors = append(selectors, &nameSegment{name: unescaped})
			allIndices = false
			continue
		}

		// 尝试解析为数字索引
		if idx, err := strconv.Atoi(trimmed); err == nil {
			selectors = append(selectors, &indexSegment{index: idx})
			allNames = false
			continue
		}

		// 不是数字也不是带引号的字符串，可能是不带引号的字段名
		selectors = append(selectors, &nameSegment{name: trimmed})
		allIndices = false
	}

	// 如果所有部分都是索引，返回多索引段
	if allIndices {
		indices := make([]int, len(selectors))
		for i, sel := range selectors {
			indices[i] = sel.(*indexSegment).index
		}
		return &multiIndexSegment{indices: indices}, nil
	}

	// 如果所有部分都是名称，返回多名称段
	if allNames {
		names := make([]string, len(selectors))
		for i, sel := range selectors {
			names[i] = sel.(*nameSegment).name
		}
		return &multiNameSegment{names: names}, nil
	}

	// 混合类型，返回联合段
	return &unionSegment{selectors: selectors}, nil
}

// removeWhitespaceAroundColons removes whitespace around colons in slice expressions
func removeWhitespaceAroundColons(content string) string {
	var result []byte
	inQuotes := false
	inSingleQuotes := false
	for i := 0; i < len(content); i++ {
		ch := content[i]
		if (inQuotes || inSingleQuotes) && ch == '\\' && i+1 < len(content) {
			result = append(result, ch)
			i++
			result = append(result, content[i])
			continue
		}
		if ch == '"' && !inSingleQuotes {
			inQuotes = !inQuotes
			result = append(result, ch)
			continue
		}
		if ch == '\'' && !inQuotes {
			inSingleQuotes = !inSingleQuotes
			result = append(result, ch)
			continue
		}
		if inQuotes || inSingleQuotes {
			result = append(result, ch)
			continue
		}
		if ch == ':' {
			// Remove trailing whitespace before colon
			for len(result) > 0 {
				lastByte := result[len(result)-1]
				if lastByte == ' ' || lastByte == '\t' || lastByte == '\n' || lastByte == '\r' {
					result = result[:len(result)-1]
				} else {
					break
				}
			}
			result = append(result, ch)
			// Skip whitespace after colon
			for i+1 < len(content) {
				nextCh := content[i+1]
				if nextCh == ' ' || nextCh == '\t' || nextCh == '\n' || nextCh == '\r' {
					i++
				} else {
					break
				}
			}
		} else {
			result = append(result, ch)
		}
	}
	return string(result)
}

// 解析切片表达式
func parseSliceSegment(content string) (segment, error) {
	// Trim whitespace around colons (RFC 9535 allows whitespace in slice expressions)
	content = strings.TrimSpace(content)
	// Remove whitespace around colons
	content = removeWhitespaceAroundColons(content)

	parts := strings.Split(content, ":")
	if len(parts) > 3 {
		return nil, NewError(ErrSyntax, "slice has too many colons", content)
	}

	slice := &sliceSegment{start: 0, end: 0, step: 1}
	hasStart := false
	hasEnd := false

	// 解析起始索引
	if parts[0] != "" {
		if !validateIntegerLiteral(parts[0]) {
			return nil, NewError(ErrSyntax, fmt.Sprintf("invalid slice start index: %s", parts[0]), parts[0])
		}
		// Reject -0 (negative zero) per RFC 9535
		if parts[0] == "-0" {
			return nil, NewError(ErrSyntax, "negative zero is not a valid slice index", parts[0])
		}
		start, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, NewError(ErrSyntax, fmt.Sprintf("invalid slice start index: %s", parts[0]), parts[0])
		}
		if start < -9007199254740991 || start > 9007199254740991 {
			return nil, NewError(ErrSyntax, fmt.Sprintf("slice start index out of range: %d", start), parts[0])
		}
		slice.start = start
		hasStart = true
	}

	// 解析结束索引
	if len(parts) > 1 && parts[1] != "" {
		if !validateIntegerLiteral(parts[1]) {
			return nil, NewError(ErrSyntax, fmt.Sprintf("invalid slice end index: %s", parts[1]), parts[1])
		}
		// Reject -0 (negative zero) per RFC 9535
		if parts[1] == "-0" {
			return nil, NewError(ErrSyntax, "negative zero is not a valid slice index", parts[1])
		}
		end, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, NewError(ErrSyntax, fmt.Sprintf("invalid slice end index: %s", parts[1]), parts[1])
		}
		if end < -9007199254740991 || end > 9007199254740991 {
			return nil, NewError(ErrSyntax, fmt.Sprintf("slice end index out of range: %d", end), parts[1])
		}
		slice.end = end
		hasEnd = true
	}

	// 解析步长
	if len(parts) > 2 && parts[2] != "" {
		if !validateIntegerLiteral(parts[2]) {
			return nil, NewError(ErrSyntax, fmt.Sprintf("invalid slice step: %s", parts[2]), parts[2])
		}
		// Reject -0 (negative zero) per RFC 9535
		if parts[2] == "-0" {
			return nil, NewError(ErrSyntax, "negative zero is not a valid slice step", parts[2])
		}
		step, err := strconv.Atoi(parts[2])
		if err != nil {
			return nil, NewError(ErrSyntax, fmt.Sprintf("invalid slice step: %s", parts[2]), parts[2])
		}
		if step == 0 {
			// RFC 9535: zero step - the slice selects no elements (returns empty)
			// Store step as 0 to signal this behavior
			slice.step = 0
		} else if step < -9007199254740991 || step > 9007199254740991 {
			return nil, NewError(ErrSyntax, fmt.Sprintf("slice step out of range: %d", step), parts[2])
		} else {
			slice.step = step
		}
	}

	slice.hasStart = hasStart
	slice.hasEnd = hasEnd

	return slice, nil
}

// validateIntegerLiteral validates an integer literal per RFC 9535.
// Returns true if the string is a valid integer- or decimal-number.
func validateIntegerLiteral(s string) bool {
	if s == "" {
		return false
	}
	i := 0
	// Optional minus (no plus sign allowed)
	if i < len(s) && s[i] == '-' {
		i++
	}
	if i >= len(s) {
		return false
	}
	// Must start with a digit
	if s[i] < '0' || s[i] > '9' {
		return false
	}
	// Leading zero check: if first digit is '0', no more digits allowed
	if s[i] == '0' && i+1 < len(s) && s[i+1] >= '0' && s[i+1] <= '9' {
		return false
	}
	// Consume digits
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	// Must have consumed all characters
	return i == len(s)
}

// looksLikeNumber checks if content looks like it could be a number (starts with digit, -digit, or +digit)
func looksLikeNumber(s string) bool {
	if s == "" {
		return false
	}
	if s[0] >= '0' && s[0] <= '9' {
		return true
	}
	if (s[0] == '-' || s[0] == '+') && len(s) > 1 && s[1] >= '0' && s[1] <= '9' {
		return true
	}
	return false
}

// validateQuotedString validates a quoted string content per RFC 9535.
// Checks for invalid escape sequences and embedded control characters.
// The input should be the string content WITHOUT the surrounding quotes.
// quoteChar is the character used for quoting (' or ").
func validateQuotedString(s string, quoteChar byte) bool {
	i := 0
	for i < len(s) {
		ch := s[i]
		if ch == '\\' {
			// Escape sequence
			i++
			if i >= len(s) {
				return false // incomplete escape
			}
			esc := s[i]
			switch esc {
			case '\\', '/', 'b', 'f', 'n', 'r', 't':
				// Valid simple escapes (same for both quote types)
				i++
			case '"':
				// Double quote escape only valid in double-quoted strings
				if quoteChar != '"' {
					return false
				}
				i++
			case '\'':
				// Single quote escape only valid in single-quoted strings
				if quoteChar != '\'' {
					return false
				}
				i++
			case 'u':
				// Unicode escape: \uXXXX
				i++
				if i+4 > len(s) {
					return false // not enough hex digits
				}
				for j := 0; j < 4; j++ {
					if i+j >= len(s) || !isHexDigit(s[i+j]) {
						return false
					}
				}
				i += 4
			default:
				return false // invalid escape sequence
			}
		} else if ch == quoteChar {
			// Unescaped quote character - invalid
			return false
		} else if ch < 0x20 {
			// Control characters (U+0000-U+001F) are not allowed unescaped
			return false
		} else {
			i++
		}
	}
	return true
}

// unescapeString processes escape sequences in a quoted string per RFC 9535.
// The input should be the string content WITHOUT the surrounding quotes.
// quoteChar is the character used for quoting (' or ").
func unescapeString(s string, quoteChar byte) (string, error) {
	var result strings.Builder
	result.Grow(len(s))
	i := 0
	for i < len(s) {
		ch := s[i]
		if ch == '\\' {
			i++
			if i >= len(s) {
				return "", fmt.Errorf("incomplete escape")
			}
			esc := s[i]
			switch esc {
			case '"':
				if quoteChar != '"' {
					return "", fmt.Errorf("invalid escape: \\\" in single-quoted string")
				}
				result.WriteByte('"')
			case '\'':
				if quoteChar != '\'' {
					return "", fmt.Errorf("invalid escape: \\' in double-quoted string")
				}
				result.WriteByte('\'')
			case '\\':
				result.WriteByte('\\')
			case '/':
				result.WriteByte('/')
			case 'b':
				result.WriteByte('\b')
			case 'f':
				result.WriteByte('\f')
			case 'n':
				result.WriteByte('\n')
			case 'r':
				result.WriteByte('\r')
			case 't':
				result.WriteByte('\t')
			case 'u':
				// Unicode escape: \uXXXX
				i++
				if i+4 > len(s) {
					return "", fmt.Errorf("incomplete unicode escape")
				}
				hexStr := s[i : i+4]
				codePoint, err := strconv.ParseUint(hexStr, 16, 32)
				if err != nil {
					return "", fmt.Errorf("invalid unicode escape: %s", hexStr)
				}
				// Check for surrogate pairs
				if codePoint >= 0xD800 && codePoint <= 0xDFFF {
					// High surrogate: must be followed by low surrogate
					if codePoint <= 0xDBFF && i+10 <= len(s) && s[i+4] == '\\' && s[i+5] == 'u' {
						lowHex := s[i+6 : i+10]
						lowPoint, err := strconv.ParseUint(lowHex, 16, 32)
						if err == nil && lowPoint >= 0xDC00 && lowPoint <= 0xDFFF {
							// Valid surrogate pair
							combined := 0x10000 + (codePoint-0xD800)*0x400 + (lowPoint - 0xDC00)
							result.WriteRune(rune(combined))
							i += 10
							continue
						}
					}
					return "", fmt.Errorf("invalid surrogate pair")
				}
				result.WriteRune(rune(codePoint))
				i += 4
				continue
			default:
				return "", fmt.Errorf("invalid escape: \\%c", esc)
			}
			i++
		} else if ch == quoteChar {
			return "", fmt.Errorf("unescaped quote character")
		} else if ch < 0x20 {
			return "", fmt.Errorf("control character")
		} else {
			result.WriteByte(ch)
			i++
		}
	}
	return result.String(), nil
}

// isHexDigit checks if a byte is a hexadecimal digit
func isHexDigit(ch byte) bool {
	return (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

// parseIndexOrName parses a bracket content as an index or name selector
func parseIndexOrName(content string) (segment, error) {
	// 处理函数调用
	if strings.HasSuffix(content, ")") {
		return parseFunctionCall(content)
	}

	// Try to parse as an integer index with RFC 9535 validation
	if validateIntegerLiteral(content) {
		idx, err := strconv.Atoi(content)
		if err != nil {
			return nil, NewError(ErrSyntax, fmt.Sprintf("invalid index: %s", content), content)
		}
		// Reject -0 (negative zero)
		if content == "-0" {
			return nil, NewError(ErrSyntax, "negative zero is not a valid index", content)
		}
		// RFC 9535: index must be within IEEE 754 double precision range
		if idx < -9007199254740991 || idx > 9007199254740991 {
			return nil, NewError(ErrSyntax, fmt.Sprintf("index out of range: %d", idx), content)
		}
		return &indexSegment{index: idx}, nil
	}

	// If content looks like a number but fails validation, it's an invalid index
	if looksLikeNumber(content) {
		return nil, NewError(ErrSyntax, fmt.Sprintf("invalid index: %s", content), content)
	}

	// 处理字符串字面量
	if strings.HasPrefix(content, "'") && strings.HasSuffix(content, "'") && len(content) > 1 {
		inner := content[1 : len(content)-1]
		if !validateQuotedString(inner, '\'') {
			return nil, NewError(ErrSyntax, fmt.Sprintf("invalid string literal: %s", content), content)
		}
		unescaped, err := unescapeString(inner, '\'')
		if err != nil {
			return nil, NewError(ErrSyntax, fmt.Sprintf("invalid string literal: %s", content), content)
		}
		return &nameSegment{name: unescaped}, nil
	}

	// 处理双引号字符串字面量
	if strings.HasPrefix(content, "\"") && strings.HasSuffix(content, "\"") && len(content) > 1 {
		inner := content[1 : len(content)-1]
		if !validateQuotedString(inner, '"') {
			return nil, NewError(ErrSyntax, fmt.Sprintf("invalid string literal: %s", content), content)
		}
		unescaped, err := unescapeString(inner, '"')
		if err != nil {
			return nil, NewError(ErrSyntax, fmt.Sprintf("invalid string literal: %s", content), content)
		}
		return &nameSegment{name: unescaped}, nil
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

// validateNumberLiteral validates a number literal per RFC 9535 grammar.
// number = ["-"] (int / (int "." 1*DIGIT))
// int = "0" / (DIGIT1 *DIGIT)
func validateNumberLiteral(s string) bool {
	if s == "" {
		return false
	}
	i := 0
	// Optional minus
	if s[i] == '-' {
		i++
	}
	if i >= len(s) {
		return false
	}
	// Must have integer part
	if s[i] < '0' || s[i] > '9' {
		return false
	}
	// Leading zero check
	if s[i] == '0' {
		i++
		if i < len(s) && s[i] >= '0' && s[i] <= '9' {
			return false // leading zero
		}
	} else {
		// Non-zero digit, consume all digits
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			i++
		}
	}
	// Optional fractional part
	if i < len(s) && s[i] == '.' {
		i++
		if i >= len(s) || s[i] < '0' || s[i] > '9' {
			return false // must have digit after decimal
		}
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			i++
		}
	}
	// Optional exponent
	if i < len(s) && (s[i] == 'e' || s[i] == 'E') {
		i++
		if i < len(s) && (s[i] == '+' || s[i] == '-') {
			i++
		}
		if i >= len(s) || s[i] < '0' || s[i] > '9' {
			return false // must have digit after e/E
		}
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			i++
		}
	}
	// Must have consumed all characters
	return i == len(s)
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
	if strings.HasPrefix(valueStr, "'") && strings.HasSuffix(valueStr, "'") && len(valueStr) > 1 {
		inner := valueStr[1 : len(valueStr)-1]
		if !validateQuotedString(inner, '\'') {
			return nil, fmt.Errorf("invalid string literal: %s", valueStr)
		}
		unescaped, err := unescapeString(inner, '\'')
		if err != nil {
			return nil, fmt.Errorf("invalid string literal: %s", valueStr)
		}
		return unescaped, nil
	}
	if strings.HasPrefix(valueStr, "\"") && strings.HasSuffix(valueStr, "\"") && len(valueStr) > 1 {
		inner := valueStr[1 : len(valueStr)-1]
		if !validateQuotedString(inner, '"') {
			return nil, fmt.Errorf("invalid string literal: %s", valueStr)
		}
		unescaped, err := unescapeString(inner, '"')
		if err != nil {
			return nil, fmt.Errorf("invalid string literal: %s", valueStr)
		}
		return unescaped, nil
	}

	// Try to parse as number with RFC 9535 validation
	if validateNumberLiteral(valueStr) {
		num, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid number: %s", valueStr)
		}
		return num, nil
	}

	// 处理 $ 引用（根节点）
	if valueStr == "$" {
		return "$", nil
	}

	// 处理 @ 引用（当前元素）
	if valueStr == "@" {
		return "@", nil
	}

	// 处理 $.path 引用（根节点路径）
	if strings.HasPrefix(valueStr, "$.") || strings.HasPrefix(valueStr, "$[") {
		return valueStr, nil
	}

	// 处理 @.path 引用（当前元素路径）
	if strings.HasPrefix(valueStr, "@.") || strings.HasPrefix(valueStr, "@[") {
		return valueStr, nil
	}

	// 如果不是其他类型，返回错误
	return nil, fmt.Errorf("invalid value: %s", valueStr)
}
