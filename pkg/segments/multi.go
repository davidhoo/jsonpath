package segments

import (
	"fmt"

	"github.com/davidhoo/jsonpath/pkg/errors"
)

// MultiIndexSegment 表示多索引段
type MultiIndexSegment struct {
	indices []int
}

// NewMultiIndexSegment 创建一个新的多索引段
func NewMultiIndexSegment(indices []int) *MultiIndexSegment {
	if indices == nil {
		indices = make([]int, 0)
	}
	return &MultiIndexSegment{indices: indices}
}

// Evaluate 实现 Segment 接口
func (s *MultiIndexSegment) Evaluate(value interface{}) ([]interface{}, error) {
	// 检查输入值是否为数组
	arr, ok := value.([]interface{})
	if !ok {
		return nil, errors.NewError(errors.ErrEvaluation, "cannot evaluate multi-index segment on non-array value", s.String())
	}

	// 如果索引列表为空，返回空结果
	if len(s.indices) == 0 {
		return []interface{}{}, nil
	}

	result := make([]interface{}, 0, len(s.indices))
	for _, idx := range s.indices {
		// 处理负数索引
		actualIdx := idx
		if idx < 0 {
			actualIdx = len(arr) + idx
		}

		// 检查索引是否越界
		if actualIdx < 0 || actualIdx >= len(arr) {
			continue
		}
		result = append(result, arr[actualIdx])
	}

	return result, nil
}

// String 实现 Segment 接口
func (s *MultiIndexSegment) String() string {
	result := "["
	for i, idx := range s.indices {
		if i > 0 {
			result += ","
		}
		result += fmt.Sprintf("%d", idx)
	}
	result += "]"
	return result
}

// MultiNameSegment 表示多名称段
type MultiNameSegment struct {
	names []string
}

// NewMultiNameSegment 创建一个新的多名称段
func NewMultiNameSegment(names []string) *MultiNameSegment {
	return &MultiNameSegment{names: names}
}

// Evaluate 实现 Segment 接口
func (s *MultiNameSegment) Evaluate(value interface{}) ([]interface{}, error) {
	// 检查输入值是否为对象
	obj, ok := value.(map[string]interface{})
	if !ok {
		return nil, errors.NewError(errors.ErrEvaluation, "cannot evaluate multi-name segment on non-object value", s.String())
	}

	var result []interface{}
	for _, name := range s.names {
		if val, exists := obj[name]; exists {
			result = append(result, val)
		}
	}

	return result, nil
}

// String 实现 Segment 接口
func (s *MultiNameSegment) String() string {
	result := "["
	for i, name := range s.names {
		if i > 0 {
			result += ","
		}
		result += fmt.Sprintf("'%s'", name)
	}
	result += "]"
	return result
}
