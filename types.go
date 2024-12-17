package jsonpath

import (
	"fmt"
	"strings"
)

// filterCondition represents a filter condition in a filter expression
type filterCondition struct {
	field    string
	operator string
	value    interface{}
}

// String returns the string representation of a filter condition
func (c *filterCondition) String() string {
	var result strings.Builder
	result.WriteString("@.")
	result.WriteString(c.field)
	result.WriteString(c.operator)
	switch v := c.value.(type) {
	case string:
		result.WriteString("\"" + v + "\"")
	default:
		result.WriteString(fmt.Sprintf("%v", v))
	}
	return result.String()
}
