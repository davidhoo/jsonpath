package jsonpath

import (
	"fmt"
	"regexp"
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
		str, ok := value.(string)
		if !ok {
			return false, nil
		}
		pattern, ok := cond.value.(string)
		if !ok {
			return false, nil
		}
		re, err := regexp.Compile(pattern)
		if err != nil {
			return false, nil
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
