package jsonpath

import (
	"fmt"
)

// 点号访问
type dotSegment struct {
	key string
}

func (d *dotSegment) evaluate(value interface{}) ([]interface{}, error) {
	if m, ok := value.(map[string]interface{}); ok {
		if v, exists := m[d.key]; exists {
			return []interface{}{v}, nil
		}
	}
	return nil, fmt.Errorf("key %s not found", d.key)
}

func (d *dotSegment) String() string {
	return "." + d.key
}

// 数组索引访问
type indexSegment struct {
	index int
}

func (i *indexSegment) evaluate(value interface{}) ([]interface{}, error) {
	if arr, ok := value.([]interface{}); ok {
		if i.index < 0 {
			i.index = len(arr) + i.index
		}
		if i.index >= 0 && i.index < len(arr) {
			return []interface{}{arr[i.index]}, nil
		}
	}
	return nil, fmt.Errorf("invalid array index %d", i.index)
}

func (i *indexSegment) String() string {
	return fmt.Sprintf("[%d]", i.index)
}

// 通配符
type wildcardSegment struct{}

func (w *wildcardSegment) evaluate(value interface{}) ([]interface{}, error) {
	switch v := value.(type) {
	case []interface{}:
		return v, nil
	case map[string]interface{}:
		result := make([]interface{}, 0, len(v))
		for _, val := range v {
			result = append(result, val)
		}
		return result, nil
	}
	return nil, fmt.Errorf("wildcard can only be applied to array or object")
}

func (w *wildcardSegment) String() string {
	return "[*]"
}

// 过滤器
type filterSegment struct {
	condition string
	operator  string
	field     string
	value     float64
}

func (f *filterSegment) evaluate(value interface{}) ([]interface{}, error) {
	arr, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("filter can only be applied to array")
	}

	var result []interface{}
	for _, item := range arr {
		match, err := f.matchCondition(item)
		if err != nil {
			continue
		}
		if match {
			result = append(result, item)
		}
	}

	// 如果结果为空，返回空数组而不是 nil
	if result == nil {
		result = make([]interface{}, 0)
	}
	return result, nil
}

func (f *filterSegment) matchCondition(item interface{}) (bool, error) {
	obj, ok := item.(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("filter item must be object")
	}

	fieldValue, ok := obj[f.field]
	if !ok {
		return false, fmt.Errorf("field %s not found", f.field)
	}

	// 转换字段值为 float64 以进行比较
	var numValue float64
	switch v := fieldValue.(type) {
	case float64:
		numValue = v
	case int:
		numValue = float64(v)
	default:
		return false, fmt.Errorf("field %s is not a number", f.field)
	}

	switch f.operator {
	case "<":
		return numValue < f.value, nil
	case "<=":
		return numValue <= f.value, nil
	case ">":
		return numValue > f.value, nil
	case ">=":
		return numValue >= f.value, nil
	case "==":
		return numValue == f.value, nil
	case "!=":
		return numValue != f.value, nil
	default:
		return false, fmt.Errorf("unsupported operator: %s", f.operator)
	}
}

func (f *filterSegment) String() string {
	return fmt.Sprintf("[?(@.%s %s %v)]", f.field, f.operator, f.value)
}

// 递归下降
type recursiveSegment struct {
	key string
}

func (r *recursiveSegment) evaluate(value interface{}) ([]interface{}, error) {
	var results []interface{}
	seen := make(map[interface{}]bool) // 用于去重

	var walk func(v interface{}) error
	walk = func(v interface{}) error {
		switch val := v.(type) {
		case map[string]interface{}:
			// 先检查当前层级的匹配
			if r.key == "*" {
				for _, v := range val {
					if !seen[v] {
						results = append(results, v)
						seen[v] = true
					}
				}
			} else if v, ok := val[r.key]; ok {
				if !seen[v] {
					results = append(results, v)
					seen[v] = true
				}
			}
			// 按照键的顺序递归搜索
			for _, v := range val {
				if err := walk(v); err != nil {
					return err
				}
			}
		case []interface{}:
			for _, item := range val {
				if err := walk(item); err != nil {
					return err
				}
			}
		}
		return nil
	}

	if err := walk(value); err != nil {
		return nil, err
	}
	return results, nil
}

func (r *recursiveSegment) String() string {
	if r.key == "*" {
		return ".."
	}
	return ".." + r.key
}

// 多索引选择器
type multiIndexSegment struct {
	indices []int
}

func (m *multiIndexSegment) evaluate(value interface{}) ([]interface{}, error) {
	arr, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("multi-index can only be applied to array")
	}

	result := make([]interface{}, 0, len(m.indices))
	for _, idx := range m.indices {
		if idx < 0 {
			idx = len(arr) + idx
		}
		if idx >= 0 && idx < len(arr) {
			result = append(result, arr[idx])
		}
	}
	return result, nil
}

func (m *multiIndexSegment) String() string {
	return fmt.Sprintf("[%v]", m.indices)
}

// 数组切片
type sliceSegment struct {
	start int
	end   int
	step  int
}

func (s *sliceSegment) evaluate(value interface{}) ([]interface{}, error) {
	arr, ok := value.([]interface{})
	if !ok {
		return nil, fmt.Errorf("slice can only be applied to array")
	}

	length := len(arr)
	start, end := s.normalizeRange(length)

	// 创建一个新的切片来存储结果
	result := make([]interface{}, 0)

	// 根据步长的方向遍历数组
	if s.step > 0 {
		// 正向遍历
		for i := start; i < end && i < length; i += s.step {
			if i >= 0 {
				result = append(result, arr[i])
			}
		}
	} else {
		// 反向遍历
		for i := start; i > end && i >= 0; i += s.step {
			if i < length {
				result = append(result, arr[i])
			}
		}
	}

	return result, nil
}

func (s *sliceSegment) normalizeRange(length int) (start, end int) {
	// 如果步长为 0，设置为默认值 1
	if s.step == 0 {
		s.step = 1
	}

	// 处理起始索引
	start = s.start
	if s.step > 0 {
		// 正向切片
		if start < 0 {
			start = length + start
		}
		if start < 0 {
			start = 0
		}
		if start >= length {
			start = length
		}
	} else {
		// 反向切片
		if start == 0 {
			start = length - 1
		} else if start < 0 {
			start = length + start
		}
		if start < 0 {
			start = 0
		}
		if start >= length {
			start = length - 1
		}
	}

	// 处理结束索引
	end = s.end
	if s.step > 0 {
		// 正向切片
		if end == 0 {
			end = length
		} else if end < 0 {
			end = length + end
		}
		if end < 0 {
			end = 0
		}
		if end > length {
			end = length
		}
	} else {
		// 反向切片
		if end == 0 {
			end = -1
		} else if end < 0 {
			end = length + end
		}
		if end < -1 {
			end = -1
		}
		if end >= length {
			end = length - 1
		}
	}

	return start, end
}

func (s *sliceSegment) String() string {
	return fmt.Sprintf("[%d:%d:%d]", s.start, s.end, s.step)
}
