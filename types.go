package jsonpath

// filterCondition represents a filter condition in a filter expression
type filterCondition struct {
	field    string
	operator string
	value    interface{}
}
