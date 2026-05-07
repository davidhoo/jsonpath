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
	isRoot   bool // true if field references root ($) instead of current element (@)
}

// exprNode represents a node in the filter expression tree
type exprNode interface {
	evaluate(item interface{}, root interface{}) (bool, error)
}

// conditionNode wraps a single filter condition
type conditionNode struct {
	cond filterCondition
}

func (n *conditionNode) evaluate(item interface{}, root interface{}) (bool, error) {
	return evaluateSingleCondition(n.cond, item, root)
}

// andNode represents an AND operation
type andNode struct {
	children []exprNode
}

func (n *andNode) evaluate(item interface{}, root interface{}) (bool, error) {
	for _, child := range n.children {
		result, err := child.evaluate(item, root)
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

func (n *orNode) evaluate(item interface{}, root interface{}) (bool, error) {
	for _, child := range n.children {
		result, err := child.evaluate(item, root)
		if err != nil {
			return false, err
		}
		if result {
			return true, nil
		}
	}
	return false, nil
}

// isNonSingularPath checks if a field path contains non-singular selectors (wildcard, slice, multi-index)
func isNonSingularPath(field string) bool {
	return strings.Contains(field, "*") || strings.Contains(field, "[") || strings.Contains(field, "..")
}

// evaluateSingleCondition evaluates a single filter condition against an item
func evaluateSingleCondition(cond filterCondition, item interface{}, root interface{}) (bool, error) {
	// Determine the context for field lookup
	var context interface{}
	if cond.isRoot {
		context = root
	} else {
		context = item
	}

	// Handle bare existence test for root ($[?$])
	if cond.field == "" && cond.operator == "exists" && cond.isRoot {
		return root != nil, nil
	}

	// Handle non-singular paths (wildcard, slice, multi-index, descendant)
	if isNonSingularPath(cond.field) {
		// Build the path expression
		var pathExpr string
		if cond.isRoot {
			pathExpr = "$" + cond.field
		} else {
			pathExpr = cond.field
		}

		// Evaluate the path
		var results NodeList
		var err error
		if cond.isRoot {
			results, err = Query(root, pathExpr)
		} else {
			// For relative paths, wrap the item and adjust the path
			// e.g., @.field.* becomes $[0].field.* when evaluated against item
			adjustedPath := "$" + pathExpr
			tempRoot := map[string]interface{}{"_": item}
			adjustedPath = "$._" + strings.TrimPrefix(pathExpr, "@")
			results, err = Query(tempRoot, adjustedPath)
		}
		if err != nil {
			return false, nil
		}
		hasResults := len(results) > 0
		switch cond.operator {
		case "exists":
			return hasResults, nil
		case "not_exists":
			return !hasResults, nil
		default:
			// RFC 9535: non-singular path in comparison is not allowed
			// Return false for comparison operators
			return false, nil
		}
	}

	var value interface{}
	var valueErr error
	var isAbsent bool

	// For root references with complex paths (containing [), evaluate as JSONPath
	if cond.isRoot && strings.Contains(cond.field, "[") {
		pathExpr := "$" + cond.field
		results, err := Query(root, pathExpr)
		if err != nil {
			return false, nil
		}
		hasResults := len(results) > 0
		switch cond.operator {
		case "exists":
			return hasResults, nil
		case "not_exists":
			return !hasResults, nil
		default:
			if !hasResults {
				isAbsent = true
			} else {
				value = results[0].Value
			}
		}
	} else {
		value, valueErr = getFieldValue(context, cond.field)
		if valueErr != nil {
			isAbsent = true
		}
	}

	// Handle absent values per RFC 9535
	if isAbsent {
		switch cond.operator {
		case "exists":
			return false, nil
		case "not_exists":
			return true, nil
		default:
			// RFC 9535: comparison with absent value
			// Resolve the comparison value
			resolvedValue := resolveFilterValue(cond.value, item, root)

			// If comparison value is also absent (nil from unresolved path), compare as Nothing
			if resolvedValue == nil {
				// Check if the comparison value itself is an absent reference
				if cond.value != nil {
					if str, ok := cond.value.(string); ok {
						// Check if it's a reference to an absent field
						if strings.HasPrefix(str, "@.") || strings.HasPrefix(str, "$.") {
							// Both sides absent: == returns true, != returns false (per RFC 9535)
							switch cond.operator {
							case "==":
								return true, nil
							case "!=":
								return false, nil
							default:
								return false, nil
							}
						}
					}
				}
			}

			// For !=null, absent != null is true (RFC 9535: absent is not null)
			if cond.operator == "!=" && resolvedValue == nil {
				// Check if comparing to literal null
				if cond.value == nil {
					return true, nil
				}
			}

			return false, nil
		}
	}

	// Resolve $ and @ references in the comparison value
	resolvedValue := resolveFilterValue(cond.value, item, root)

	switch cond.operator {
	case "exists":
		// If we got here, the field was found
		return true, nil
	case "not_exists":
		// If we got here, the field was found (so it exists)
		return false, nil
	case "match":
		// RFC 9535 match() function: match(string, pattern)
		// Uses I-Regexp for full-string matching
		str, ok := value.(string)
		if !ok {
			return false, nil
		}
		pattern, ok := resolvedValue.(string)
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
		pattern, ok := resolvedValue.(string)
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
		result, err := compareValues(value, cond.operator, resolvedValue)
		if err != nil {
			return false, fmt.Errorf("invalid operator: %s", cond.operator)
		}
		return result, nil
	}
}

// resolveFilterValue resolves $ and @ references in filter values
func resolveFilterValue(value interface{}, item interface{}, root interface{}) interface{} {
	str, ok := value.(string)
	if !ok {
		return value
	}

	switch str {
	case "$":
		// Root reference - return root value
		return root
	case "@":
		// Current element reference - return current item
		return item
	default:
		// Check for $.path or @.path references
		if strings.HasPrefix(str, "$.") || strings.HasPrefix(str, "$[") {
			// Resolve path from root using JSONPath engine for complex paths
			if strings.Contains(str, "[") {
				// Complex path with brackets - use JSONPath engine
				results, err := Query(root, str)
				if err != nil || len(results) == 0 {
					return nil
				}
				if len(results) == 1 {
					return results[0].Value
				}
				// Return array of values for multiple results
				values := make([]interface{}, len(results))
				for i, r := range results {
					values[i] = r.Value
				}
				return values
			}
			// Simple path - use getFieldValue
			path := strings.TrimPrefix(str, "$")
			if path == "" {
				return root
			}
			resolved, err := getFieldValue(root, path)
			if err != nil {
				return nil
			}
			return resolved
		}
		if strings.HasPrefix(str, "@.") || strings.HasPrefix(str, "@[") {
			// Resolve path from current element using JSONPath engine for complex paths
			if strings.Contains(str, "[") {
				// Complex path with brackets - use JSONPath engine
				// Wrap item in an array to make it accessible as $[0]
				tempRoot := []interface{}{item}
				adjustedPath := "$[0]" + strings.TrimPrefix(str, "@")
				results, err := Query(tempRoot, adjustedPath)
				if err != nil || len(results) == 0 {
					return nil
				}
				if len(results) == 1 {
					return results[0].Value
				}
				// Return array of values for multiple results
				values := make([]interface{}, len(results))
				for i, r := range results {
					values[i] = r.Value
				}
				return values
			}
			// Simple path - use getFieldValue
			path := strings.TrimPrefix(str, "@")
			if path == "" {
				return item
			}
			resolved, err := getFieldValue(item, path)
			if err != nil {
				return nil
			}
			return resolved
		}
		return value
	}
}

// String returns the string representation of a filter condition
func (c filterCondition) String() string {
	prefix := "@"
	if c.isRoot {
		prefix = "$"
	}
	field := strings.TrimPrefix(c.field, "@.")
	field = strings.TrimPrefix(field, "$.")
	switch c.operator {
	case "exists":
		if field == "" {
			return prefix
		}
		return fmt.Sprintf("%s.%s", prefix, field)
	case "not_exists":
		return fmt.Sprintf("!%s.%s", prefix, field)
	case "match":
		return fmt.Sprintf("match(%s.%s, '%v')", prefix, field, c.value)
	case "search":
		return fmt.Sprintf("search(%s.%s, '%v')", prefix, field, c.value)
	default:
		value := c.value
		if str, ok := value.(string); ok {
			value = "'" + str + "'"
		}
		return fmt.Sprintf("%s.%s %s %v", prefix, field, c.operator, value)
	}
}
