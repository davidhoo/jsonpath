package jsonpath

import (
	"fmt"
	"strconv"
	"strings"
)

func parse(path string) ([]segment, error) {
	var segments []segment
	var current string

	for len(path) > 0 {
		switch path[0] {
		case '.':
			path = path[1:]
			if len(path) == 0 {
				return nil, fmt.Errorf("unexpected end after dot")
			}

			// 处理递归下降
			if path[0] == '.' {
				path = path[1:]
				end := strings.IndexAny(path, ".[")
				if end == -1 {
					end = len(path)
				}
				key := path[:end]
				segments = append(segments, &recursiveSegment{key: key})
				path = path[end:]
				continue
			}

			if path[0] == '*' {
				segments = append(segments, &wildcardSegment{})
				path = path[1:]
				continue
			}

			end := strings.IndexAny(path, ".[")
			if end == -1 {
				end = len(path)
			}
			segments = append(segments, &dotSegment{key: path[:end]})
			path = path[end:]

		case '[':
			path = path[1:]
			end := strings.Index(path, "]")
			if end == -1 {
				return nil, fmt.Errorf("unclosed bracket")
			}

			current = path[:end]
			path = path[end+1:]

			// 处理切片
			if strings.Contains(current, ":") {
				start, end, step, err := parseSlice(current)
				if err != nil {
					return nil, err
				}
				segments = append(segments, &sliceSegment{
					start: start,
					end:   end,
					step:  step,
				})
				continue
			}

			// 处理多索引
			if strings.Contains(current, ",") {
				indices, err := parseMultiIndex(current)
				if err != nil {
					return nil, err
				}
				segments = append(segments, &multiIndexSegment{indices: indices})
				continue
			}

			if current == "*" {
				segments = append(segments, &wildcardSegment{})
			} else if idx, err := strconv.Atoi(current); err == nil {
				segments = append(segments, &indexSegment{index: idx})
			} else if strings.HasPrefix(current, "?") {
				expr := current[2 : len(current)-1] // 移除 "?(" 和 ")"
				field, op, value, err := parseFilterExpression(expr)
				if err != nil {
					return nil, fmt.Errorf("invalid filter expression: %v", err)
				}
				segments = append(segments, &filterSegment{
					field:    field,
					operator: op,
					value:    value,
				})
			} else {
				return nil, fmt.Errorf("invalid bracket expression: %s", current)
			}

		default:
			return nil, fmt.Errorf("unexpected character: %c", path[0])
		}
	}

	return segments, nil
}

func parseFilterExpression(expr string) (field string, operator string, value float64, err error) {
	// 支持的操作符
	operators := []string{"<=", ">=", "<", ">", "==", "!="}

	if !strings.HasPrefix(expr, "@.") {
		return "", "", 0, fmt.Errorf("filter must start with @.")
	}
	expr = strings.TrimSpace(expr[2:]) // 移除 "@." 并清理空格

	// 查找操作符
	var op string
	var opIndex int = -1
	for _, operator := range operators {
		if idx := strings.Index(expr, operator); idx != -1 {
			if opIndex == -1 || idx < opIndex {
				op = operator
				opIndex = idx
			}
		}
	}

	if opIndex == -1 {
		return "", "", 0, fmt.Errorf("no valid operator found")
	}

	field = strings.TrimSpace(expr[:opIndex])
	if field == "" {
		return "", "", 0, fmt.Errorf("empty field name")
	}

	valueStr := strings.TrimSpace(expr[opIndex+len(op):])
	value, err = strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return "", "", 0, fmt.Errorf("invalid numeric value: %s", valueStr)
	}

	return field, op, value, nil
}

// 解析切片表��式
func parseSlice(expr string) (start, end, step int, err error) {
	parts := strings.Split(expr, ":")
	if len(parts) > 3 {
		return 0, 0, 0, fmt.Errorf("invalid slice format")
	}

	// 设置默认值
	start, end, step = 0, 0, 1

	// 解析 start
	if parts[0] != "" {
		start, err = strconv.Atoi(parts[0])
		if err != nil {
			return 0, 0, 0, fmt.Errorf("invalid start index: %s", parts[0])
		}
	}

	// 解析 end
	if len(parts) > 1 {
		if parts[1] != "" {
			end, err = strconv.Atoi(parts[1])
			if err != nil {
				return 0, 0, 0, fmt.Errorf("invalid end index: %s", parts[1])
			}
		}
	}

	// 解析 step
	if len(parts) > 2 {
		if parts[2] != "" {
			step, err = strconv.Atoi(parts[2])
			if err != nil {
				return 0, 0, 0, fmt.Errorf("invalid step: %s", parts[2])
			}
			if step == 0 {
				return 0, 0, 0, fmt.Errorf("step cannot be zero")
			}
		} else if len(parts) == 3 {
			step = -1 // 处理 [::-1] 的情况
		}
	}

	return start, end, step, nil
}

// 解析多索引表达式
func parseMultiIndex(expr string) ([]int, error) {
	parts := strings.Split(expr, ",")
	indices := make([]int, 0, len(parts))

	for _, part := range parts {
		idx, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return nil, fmt.Errorf("invalid index: %s", part)
		}
		indices = append(indices, idx)
	}

	return indices, nil
}
