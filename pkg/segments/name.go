package segments

// NameSegment 表示名称段
type NameSegment struct {
	name string
}

// NewNameSegment 创建一个新的名称段
func NewNameSegment(name string) *NameSegment {
	return &NameSegment{name: name}
}

// Evaluate 实现 Segment 接口
func (s *NameSegment) Evaluate(value interface{}) ([]interface{}, error) {
	// TODO: 实现名称段的评估逻辑
	return nil, nil
}

// String 实现 Segment 接口
func (s *NameSegment) String() string {
	return s.name
}
