package parser

import (
	"strconv"
	"strings"

	"github.com/davidhoo/jsonpath/pkg/errors"
	"github.com/davidhoo/jsonpath/pkg/segments"
)

// parseBracketSegment 解析括号表达式
func parseBracketSegment(content string) (segments.Segment, error) {
	// 移除括号
	content = strings.TrimPrefix(content, "[")
	content = strings.TrimSuffix(content, "]")

	// 处理空内容
	if content == "" {
		return nil, errors.NewError(errors.ErrSyntax, "empty bracket expression", content)
	}

	// 根据内容类型分发到不同的解析器
	switch {
	case content == "*":
		return &segments.WildcardSegment{}, nil
	case strings.HasPrefix(content, "?"):
		return parseFilterSegment(content)
	case strings.Contains(content, ","):
		return parseMultiIndexSegment(content)
	case strings.Contains(content, ":"):
		return parseSliceSegment(content)
	case strings.HasPrefix(content, "'") || strings.HasPrefix(content, "\""):
		// 移除引号
		content = strings.Trim(content, "'\"")
		return segments.NewNameSegment(content), nil
	case strings.HasPrefix(content, "@"):
		return parseFunctionCall(content)
	default:
		// 尝试解析为数字索引
		if idx, err := strconv.Atoi(content); err == nil {
			return segments.NewIndexSegment(idx), nil
		}
		return segments.NewNameSegment(content), nil
	}
}

// parseFilterSegment 解析过滤器表达式
func parseFilterSegment(content string) (segments.Segment, error) {
	// 移除括号
	content = strings.TrimPrefix(content, "[")
	content = strings.TrimSuffix(content, "]")

	// 移除 ? 前缀
	content = strings.TrimPrefix(content, "?")

	// 检查表达式是否为空
	if content == "" {
		return nil, errors.NewError(errors.ErrSyntax, "empty filter expression", content)
	}

	// 检查表达式是否以 @ 开头并且有括号
	if !strings.HasPrefix(content, "(@") {
		return nil, errors.NewError(errors.ErrSyntax, "filter expression must start with @", content)
	}

	// 移除开头的 ( 和结尾的 )
	content = strings.TrimPrefix(content, "(")
	content = strings.TrimSuffix(content, ")")

	// 检查表达式是否以 @ 开头
	if !strings.HasPrefix(content, "@") {
		return nil, errors.NewError(errors.ErrSyntax, "filter expression must start with @", content)
	}

	// 移除 @ 前缀，但保留后面的表达式
	content = strings.TrimPrefix(content, "@")

	// 检查表达式是否为空
	if content == "" {
		return nil, errors.NewError(errors.ErrSyntax, "empty filter expression after @", content)
	}

	// 如果表达式以 . 开头，移除它
	content = strings.TrimPrefix(content, ".")

	return segments.NewFilterSegment(content), nil
}

// parseMultiIndexSegment 解析多索引段
func parseMultiIndexSegment(content string) (segments.Segment, error) {
	// 移除括号
	content = strings.TrimPrefix(content, "[")
	content = strings.TrimSuffix(content, "]")

	// 分割多个索引
	parts := strings.Split(content, ",")
	indices := make([]int, 0, len(parts))
	names := make([]string, 0, len(parts))
	isIndex := true

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return nil, errors.NewError(errors.ErrSyntax, "empty multi-index segment", content)
		}

		// 处理带引号的名称
		if strings.HasPrefix(part, "'") || strings.HasPrefix(part, "\"") {
			if isIndex && len(indices) > 0 {
				return nil, errors.NewError(errors.ErrSyntax, "mixed index and name segments", content)
			}
			isIndex = false
			name := strings.Trim(part, "'\"")
			names = append(names, name)
			continue
		}

		// 尝试解析为数字
		idx, err := strconv.Atoi(part)
		if err != nil {
			if isIndex && len(indices) > 0 {
				return nil, errors.NewError(errors.ErrSyntax, "mixed index and name segments", content)
			}
			isIndex = false
			names = append(names, part)
			continue
		}

		if !isIndex && len(names) > 0 {
			return nil, errors.NewError(errors.ErrSyntax, "mixed index and name segments", content)
		}
		isIndex = true
		indices = append(indices, idx)
	}

	if isIndex {
		return segments.NewMultiIndexSegment(indices), nil
	}
	return segments.NewMultiNameSegment(names), nil
}

// parseSliceSegment 解析切片表达式
func parseSliceSegment(content string) (segments.Segment, error) {
	// 移除括号
	content = strings.TrimPrefix(content, "[")
	content = strings.TrimSuffix(content, "]")

	// 分割切片参数
	parts := strings.Split(content, ":")
	if len(parts) != 2 && len(parts) != 3 {
		return nil, errors.NewError(errors.ErrSyntax, "invalid slice expression", content)
	}

	var start, end, step int
	var err error

	// 解析起始位置
	if parts[0] == "" {
		start = 0
	} else {
		start, err = strconv.Atoi(parts[0])
		if err != nil {
			return nil, errors.NewError(errors.ErrSyntax, "invalid start index", parts[0])
		}
	}

	// 解析结束位置
	if parts[1] == "" {
		end = -1 // 表示到末尾
	} else {
		end, err = strconv.Atoi(parts[1])
		if err != nil {
			return nil, errors.NewError(errors.ErrSyntax, "invalid end index", parts[1])
		}
	}

	// 解析步长
	if len(parts) == 3 {
		if parts[2] == "" {
			step = 1
		} else {
			step, err = strconv.Atoi(parts[2])
			if err != nil {
				return nil, errors.NewError(errors.ErrSyntax, "invalid step", parts[2])
			}
			if step == 0 {
				return nil, errors.NewError(errors.ErrSyntax, "step cannot be zero", parts[2])
			}
		}
	} else {
		step = 1
	}

	return segments.NewSliceSegment(start, end, step), nil
}

// parseIndexOrName 解析索引或名称表达式
func parseIndexOrName(content string) (segments.Segment, error) {
	// TODO: 实现索引或名称表达式解析
	return nil, nil
}

// parseFunctionCall 解析函数调用表达式
func parseFunctionCall(content string) (segments.Segment, error) {
	// 移除括号
	content = strings.TrimPrefix(content, "[")
	content = strings.TrimSuffix(content, "]")

	// 移除 @ 前缀
	content = strings.TrimPrefix(content, "@")

	// 检查表达式是否为空
	if content == "" {
		return nil, errors.NewError(errors.ErrSyntax, "empty function call", content)
	}

	// 移除开头的 . 如果存在
	content = strings.TrimPrefix(content, ".")

	// 查找函数名和参数
	openParen := strings.Index(content, "(")
	if openParen == -1 {
		return nil, errors.NewError(errors.ErrSyntax, "missing opening parenthesis", content)
	}

	closeParen := strings.LastIndex(content, ")")
	if closeParen == -1 {
		return nil, errors.NewError(errors.ErrSyntax, "missing closing parenthesis", content)
	}

	if closeParen != len(content)-1 {
		return nil, errors.NewError(errors.ErrSyntax, "extra characters after closing parenthesis", content)
	}

	// 提取函数名
	name := strings.TrimSpace(content[:openParen])
	if name == "" {
		return nil, errors.NewError(errors.ErrSyntax, "empty function name", content)
	}

	// 提取参数
	argsStr := strings.TrimSpace(content[openParen+1 : closeParen])
	var arguments []string
	if argsStr != "" {
		arguments = strings.Split(argsStr, ",")
		for i, arg := range arguments {
			arguments[i] = strings.TrimSpace(arg)
		}
	}

	return segments.NewFunctionSegment(name, arguments), nil
}
