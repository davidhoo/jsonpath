package segments

import (
	"fmt"
	"reflect"
	"strconv"
)

// IndexSegment 表示索引段
type IndexSegment struct {
	index int
}

// NewIndexSegment 创建一个新的索引段
func NewIndexSegment(index int) *IndexSegment {
	return &IndexSegment{index: index}
}

// Evaluate 实现 Segment 接口
func (s *IndexSegment) Evaluate(value interface{}) ([]interface{}, error) {
	if value == nil {
		return nil, nil
	}

	// 使用反射获取数组
	val := reflect.ValueOf(value)
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		return nil, fmt.Errorf("cannot access index %d on non-array value", s.index)
	}

	// 检查索引是否越界
	if s.index < 0 {
		s.index = val.Len() + s.index
	}
	if s.index < 0 || s.index >= val.Len() {
		return nil, nil // 索引越界时返回空结果
	}

	return []interface{}{val.Index(s.index).Interface()}, nil
}

// String 实现 Segment 接口
func (s *IndexSegment) String() string {
	return strconv.Itoa(s.index)
}
