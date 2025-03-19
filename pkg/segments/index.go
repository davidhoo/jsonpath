package segments

import "strconv"

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
	// TODO: 实现索引段的评估逻辑
	return nil, nil
}

// String 实现 Segment 接口
func (s *IndexSegment) String() string {
	return strconv.Itoa(s.index)
}
