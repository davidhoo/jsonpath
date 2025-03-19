package segments

// RecursiveSegment 表示递归下降段
type RecursiveSegment struct{}

// Evaluate 实现 Segment 接口
func (s *RecursiveSegment) Evaluate(value interface{}) ([]interface{}, error) {
	// TODO: 实现递归下降逻辑
	return nil, nil
}

// String 实现 Segment 接口
func (s *RecursiveSegment) String() string {
	return ".."
}
