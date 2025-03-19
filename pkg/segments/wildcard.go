package segments

// WildcardSegment 表示通配符段
type WildcardSegment struct{}

// Evaluate 实现 Segment 接口
func (s *WildcardSegment) Evaluate(value interface{}) ([]interface{}, error) {
	// TODO: 实现通配符段的评估逻辑
	return nil, nil
}

// String 实现 Segment 接口
func (s *WildcardSegment) String() string {
	return "*"
}
