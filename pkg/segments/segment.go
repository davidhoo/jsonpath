package segments

// Segment 表示 JSONPath 中的一个段
type Segment interface {
	// Evaluate 评估段并返回结果
	Evaluate(value interface{}) ([]interface{}, error)
	// String 返回段的字符串表示
	String() string
}
