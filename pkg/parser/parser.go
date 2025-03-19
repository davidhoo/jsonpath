package parser

import (
	"strings"

	"github.com/davidhoo/jsonpath/pkg/errors"
	"github.com/davidhoo/jsonpath/pkg/segments"
)

// Parse 解析 JSONPath 表达式并返回段列表
func Parse(path string) ([]segments.Segment, error) {
	// 处理空路径
	if path == "" {
		return nil, nil
	}

	// 检查并移除 $ 前缀
	if !strings.HasPrefix(path, "$") {
		return nil, errors.NewError(errors.ErrSyntax, "path must start with $", path)
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

// parseRecursive 解析递归下降路径
func parseRecursive(path string) ([]segments.Segment, error) {
	var segs []segments.Segment
	segs = append(segs, &segments.RecursiveSegment{})

	// 如果路径为空，直接返回
	if path == "" {
		return segs, nil
	}

	// 解析剩余部分
	remainingSegments, err := parseRegular(path)
	if err != nil {
		return nil, err
	}

	return append(segs, remainingSegments...), nil
}

// parseRegular 解析常规路径
func parseRegular(path string) ([]segments.Segment, error) {
	var result []segments.Segment
	var current strings.Builder
	var inBracket bool
	var bracketCount int

	for i := 0; i < len(path); i++ {
		ch := path[i]

		switch {
		case ch == '[':
			if !inBracket {
				// 处理前面的点表示法部分
				if current.Len() > 0 {
					seg, err := createDotSegment(current.String())
					if err != nil {
						return nil, err
					}
					result = append(result, seg)
					current.Reset()
				}
				inBracket = true
			}
			bracketCount++
			current.WriteByte(ch)

		case ch == ']':
			bracketCount--
			current.WriteByte(ch)
			if bracketCount == 0 {
				// 处理括号表达式
				seg, err := parseBracketSegment(current.String())
				if err != nil {
					return nil, err
				}
				result = append(result, seg)
				current.Reset()
				inBracket = false
			}

		case ch == '.' && !inBracket:
			// 处理点分隔的部分
			if current.Len() > 0 {
				seg, err := createDotSegment(current.String())
				if err != nil {
					return nil, err
				}
				result = append(result, seg)
				current.Reset()
			}

		default:
			current.WriteByte(ch)
		}
	}

	// 处理最后一部分
	if current.Len() > 0 {
		if inBracket {
			return nil, errors.NewError(errors.ErrSyntax, "unclosed bracket expression", path)
		}
		seg, err := createDotSegment(current.String())
		if err != nil {
			return nil, err
		}
		result = append(result, seg)
	}

	return result, nil
}

// createDotSegment 创建点表示法段
func createDotSegment(name string) (segments.Segment, error) {
	if name == "" {
		return nil, errors.NewError(errors.ErrSyntax, "empty name segment", name)
	}
	return segments.NewNameSegment(name), nil
}
