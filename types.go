package jsonpath

import (
	"fmt"
	"regexp"
	"strings"
)

// filterCondition represents a filter condition in a filter expression
type filterCondition struct {
	field    string
	operator string
	value    interface{}
}

// exprNode represents a node in the filter expression tree
type exprNode interface {
	evaluate(item interface{}) (bool, error)
}

// conditionNode wraps a single filter condition
type conditionNode struct {
	cond filterCondition
}

func (n *conditionNode) evaluate(item interface{}) (bool, error) {
	return evaluateSingleCondition(n.cond, item)
}

// andNode represents an AND operation
type andNode struct {
	children []exprNode
}

func (n *andNode) evaluate(item interface{}) (bool, error) {
	for _, child := range n.children {
		result, err := child.evaluate(item)
		if err != nil {
			return false, err
		}
		if !result {
			return false, nil
		}
	}
	return true, nil
}

// orNode represents an OR operation
type orNode struct {
	children []exprNode
}

func (n *orNode) evaluate(item interface{}) (bool, error) {
	for _, child := range n.children {
		result, err := child.evaluate(item)
		if err != nil {
			return false, err
		}
		if result {
			return true, nil
		}
	}
	return false, nil
}

// evaluateSingleCondition evaluates a single filter condition against an item
func evaluateSingleCondition(cond filterCondition, item interface{}) (bool, error) {
	value, err := getFieldValue(item, cond.field)
	if err != nil {
		// For not_exists operator, field not found means it doesn't exist (true)
		if cond.operator == "not_exists" {
			return true, nil
		}
		return false, nil
	}

	switch cond.operator {
	case "exists":
		// Existence test: field exists and value is not nil
		return value != nil, nil
	case "not_exists":
		// Non-existence test: field doesn't exist or value is nil
		return value == nil, nil
	case "match":
		// RFC 9535 match() function: match(string, pattern)
		// Uses I-Regexp for full-string matching
		str, ok := value.(string)
		if !ok {
			return false, nil
		}
		pattern, ok := cond.value.(string)
		if !ok {
			return false, nil
		}
		// Convert I-Regexp to Go regexp
		goPattern, err := IRegexpToGoRegexp(pattern)
		if err != nil {
			return false, nil // Invalid pattern returns false
		}
		// Add anchors for full-string matching
		goPattern = "^(" + goPattern + ")$"
		re, err := regexp.Compile(goPattern)
		if err != nil {
			return false, nil
		}
		return re.MatchString(str), nil
	case "search":
		// RFC 9535 search() function: search(string, pattern)
		// Returns true if string contains a match for the pattern
		str, ok := value.(string)
		if !ok {
			return false, nil
		}
		pattern, ok := cond.value.(string)
		if !ok {
			return false, nil
		}
		// Convert I-Regexp to Go regexp
		goPattern, err := IRegexpToGoRegexp(pattern)
		if err != nil {
			return false, fmt.Errorf("invalid regex pattern: %s", pattern)
		}
		re, err := regexp.Compile(goPattern)
		if err != nil {
			return false, fmt.Errorf("invalid regex pattern: %s", pattern)
		}
		return re.MatchString(str), nil
	default:
		result, err := compareValues(value, cond.operator, cond.value)
		if err != nil {
			return false, fmt.Errorf("invalid operator: %s", cond.operator)
		}
		return result, nil
	}
}

// String returns the string representation of a filter condition
func (c filterCondition) String() string {
	field := strings.TrimPrefix(c.field, "@.")
	switch c.operator {
	case "exists":
		return fmt.Sprintf("@.%s", field)
	case "not_exists":
		return fmt.Sprintf("!@.%s", field)
	case "match":
		return fmt.Sprintf("match(@.%s, '%v')", field, c.value)
	case "search":
		return fmt.Sprintf("search(@.%s, '%v')", field, c.value)
	default:
		value := c.value
		if str, ok := value.(string); ok {
			value = "'" + str + "'"
		}
		return fmt.Sprintf("@.%s %s %v", field, c.operator, value)
	}
}
