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

	// Handle function calls as field (e.g., count(@..*)>2, length(@.a)==2)
	// This must be checked BEFORE non-singular path check since function calls
	// may contain non-singular selectors (like @..*) as arguments
	if funcName, argsStr, isFunc := isFunctionCall(cond.field); isFunc {
		funcResult, err := evaluateFilterFunction(funcName, argsStr, item, root)
		if err != nil {
			return false, nil
		}
		// Check if the comparison value is also a function call
		resolvedValue := resolveFilterValue(cond.value, item, root)
		if valueStr, ok := cond.value.(string); ok {
			if valFuncName, valArgsStr, isValFunc := isFunctionCall(valueStr); isValFunc {
				valResult, valErr := evaluateFilterFunction(valFuncName, valArgsStr, item, root)
				if valErr != nil {
					resolvedValue = Nothing{}
				} else {
					resolvedValue = valResult
				}
			}
		}
		result, err := compareValues(funcResult, cond.operator, resolvedValue)
		if err != nil {
			return false, nil
		}
		return result, nil
	}

	// Handle function calls as comparison value (e.g., @.a == length(@.b))
	// Skip match/search operators - they have dedicated handling below
	if cond.operator != "match" && cond.operator != "search" && cond.operator != "not_match" && cond.operator != "not_search" {
		if valueStr, ok := cond.value.(string); ok {
			if funcName, argsStr, isFunc := isFunctionCall(valueStr); isFunc {
				funcResult, err := evaluateFilterFunction(funcName, argsStr, item, root)
				if err != nil {
					return false, nil
				}
				var fieldValue interface{}
				var fieldErr error
				if cond.isRoot {
					fieldValue, fieldErr = getFieldValue(root, cond.field)
				} else {
					fieldValue, fieldErr = getFieldValue(item, cond.field)
				}
				if fieldErr != nil {
					// Field is absent - treat as Nothing and compare with function result
					switch cond.operator {
					case "exists":
						return false, nil
					case "not_exists":
						return true, nil
					default:
						return compareValues(Nothing{}, cond.operator, funcResult)
					}
				}
				result, err := compareValues(fieldValue, cond.operator, funcResult)
				if err != nil {
					return false, nil
				}
				return result, nil
			}
		}
	}

	// Handle non-singular paths (wildcard, slice, multi-index, descendant)
	if isNonSingularPath(cond.field) {
		// Build the path expression
		var pathExpr string
		if cond.isRoot {
			if strings.HasPrefix(cond.field, "..") {
				pathExpr = "$" + cond.field
			} else {
				pathExpr = "$." + cond.field
			}
		} else {
			if strings.HasPrefix(cond.field, "..") {
				pathExpr = "@" + cond.field
			} else {
				pathExpr = "@." + cond.field
			}
		}

		// Evaluate the path
		var results NodeList
		var err error
		if cond.isRoot {
			results, err = Query(root, pathExpr)
		} else {
			// For relative paths, wrap the item in an array and adjust the path
			tempRoot := []interface{}{item}
			var adjustedPath string
			if strings.HasPrefix(cond.field, "..") {
				adjustedPath = "$[0]" + cond.field
			} else {
				adjustedPath = "$[0]." + cond.field
			}
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

			// Check if the comparison value is a function call and evaluate it
			if valueStr, ok := cond.value.(string); ok {
				if valFuncName, valArgsStr, isValFunc := isFunctionCall(valueStr); isValFunc {
					valResult, valErr := evaluateFilterFunction(valFuncName, valArgsStr, item, root)
					if valErr != nil {
						resolvedValue = Nothing{}
					} else {
						resolvedValue = valResult
					}
				} else if resolvedValue == nil && (strings.HasPrefix(valueStr, "@") || strings.HasPrefix(valueStr, "$")) {
					// Path reference that resolved to nil → treat as Nothing
					resolvedValue = Nothing{}
				}
			}

			// Treat absent field as Nothing
			nothingValue := Nothing{}
			result, err := compareValues(nothingValue, cond.operator, resolvedValue)
			if err != nil {
				return false, nil
			}
			return result, nil
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
		// Evaluate function call pattern if needed
		matchPattern := resolvedValue
		if valueStr, ok := cond.value.(string); ok {
			if valFuncName, valArgsStr, isValFunc := isFunctionCall(valueStr); isValFunc {
				valResult, valErr := evaluateFilterFunction(valFuncName, valArgsStr, item, root)
				if valErr != nil {
					return false, nil
				}
				matchPattern = valResult
			}
		}
		pattern, ok := matchPattern.(string)
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
		// Evaluate function call pattern if needed
		searchPattern := resolvedValue
		if valueStr, ok := cond.value.(string); ok {
			if valFuncName, valArgsStr, isValFunc := isFunctionCall(valueStr); isValFunc {
				valResult, valErr := evaluateFilterFunction(valFuncName, valArgsStr, item, root)
				if valErr != nil {
					return false, nil
				}
				searchPattern = valResult
			}
		}
		pattern, ok := searchPattern.(string)
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
	case "not_match":
		// Negated match: returns true if string does NOT match the pattern
		str, ok := value.(string)
		if !ok {
			return true, nil
		}
		pattern, ok := resolvedValue.(string)
		if !ok {
			return true, nil
		}
		goPattern, err := IRegexpToGoRegexp(pattern)
		if err != nil {
			return true, nil
		}
		goPattern = "^(" + goPattern + ")$"
		re, err := regexp.Compile(goPattern)
		if err != nil {
			return true, nil
		}
		return !re.MatchString(str), nil
	case "not_search":
		// Negated search: returns true if string does NOT contain a match
		str, ok := value.(string)
		if !ok {
			return true, nil
		}
		pattern, ok := resolvedValue.(string)
		if !ok {
			return true, nil
		}
		goPattern, err := IRegexpToGoRegexp(pattern)
		if err != nil {
			return true, nil
		}
		re, err := regexp.Compile(goPattern)
		if err != nil {
			return true, nil
		}
		return !re.MatchString(str), nil
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

// isFunctionCall checks if a string looks like a function call (e.g., "count(@..*)")
func isFunctionCall(s string) (string, string, bool) {
	s = strings.TrimSpace(s)
	if !strings.HasSuffix(s, ")") {
		return "", "", false
	}
	idx := strings.Index(s, "(")
	if idx <= 0 {
		return "", "", false
	}
	funcName := s[:idx]
	if !isValidFunctionName(funcName) {
		return "", "", false
	}
	// Find the matching ')' for the '(' at idx
	depth := 0
	matchingIdx := -1
	for i := idx; i < len(s); i++ {
		if s[i] == '(' {
			depth++
		} else if s[i] == ')' {
			depth--
			if depth == 0 {
				matchingIdx = i
				break
			}
		}
	}
	// The matching ')' must be at the end of the string
	if matchingIdx != len(s)-1 {
		return "", "", false
	}
	argsStr := s[idx+1 : len(s)-1]
	return funcName, argsStr, true
}

// evaluateFilterFunction evaluates a function call in a filter context
func evaluateFilterFunction(funcName, argsStr string, item interface{}, root interface{}) (interface{}, error) {
	fn, err := GetFunction(funcName)
	if err != nil {
		return nil, err
	}

	// Parse the arguments
	args, err := parseFunctionArgsList(argsStr)
	if err != nil {
		return nil, err
	}

	// Resolve @ and $ path references in arguments
	resolvedArgs := make([]interface{}, len(args))
	for i, arg := range args {
		if str, ok := arg.(string); ok {
			// Normalize whitespace in paths (e.g., "$ [0] .a" -> "$[0].a")
			if strings.HasPrefix(str, "$") || strings.HasPrefix(str, "@") {
				str = normalizePathWhitespace(str)
			}
			// Check if the argument is itself a function call (e.g., value($..c))
			if innerFuncName, innerArgsStr, isInnerFunc := isFunctionCall(str); isInnerFunc {
				innerResult, innerErr := evaluateFilterFunction(innerFuncName, innerArgsStr, item, root)
				if innerErr != nil {
					// Function returned error → treat as Nothing
					resolvedArgs[i] = Nothing{}
				} else {
					resolvedArgs[i] = innerResult
				}
			} else if str == "@" {
				resolvedArgs[i] = item
			} else if str == "$" {
				resolvedArgs[i] = root
			} else if strings.HasPrefix(str, "@") {
				// Evaluate @.path or @[...] against the current item
				tempRoot := []interface{}{item}
				adjustedPath := "$[0]" + strings.TrimPrefix(str, "@")
				results, err := Query(tempRoot, adjustedPath)
				if err != nil {
					resolvedArgs[i] = nil
				} else {
					// For count() and value(), pass the nodelist as an array
					if funcName == "count" || funcName == "value" {
						values := make([]interface{}, len(results))
						for j, r := range results {
							values[j] = r.Value
						}
						resolvedArgs[i] = values
					} else if len(results) == 1 {
						resolvedArgs[i] = results[0].Value
					} else if len(results) > 1 {
						values := make([]interface{}, len(results))
						for j, r := range results {
							values[j] = r.Value
						}
						resolvedArgs[i] = values
					} else {
						resolvedArgs[i] = nil
					}
				}
			} else if strings.HasPrefix(str, "$") {
				// Evaluate $.path or $[...] against the root
				results, err := Query(root, str)
				if err != nil {
					resolvedArgs[i] = nil
				} else {
					// For count() and value(), pass the nodelist as an array
					if funcName == "count" || funcName == "value" {
						values := make([]interface{}, len(results))
						for j, r := range results {
							values[j] = r.Value
						}
						resolvedArgs[i] = values
					} else if len(results) == 1 {
						resolvedArgs[i] = results[0].Value
					} else if len(results) > 1 {
						values := make([]interface{}, len(results))
						for j, r := range results {
							values[j] = r.Value
						}
						resolvedArgs[i] = values
					} else {
						resolvedArgs[i] = nil
					}
				}
			} else {
				resolvedArgs[i] = arg
			}
		} else {
			resolvedArgs[i] = arg
		}
	}

	// Check for Nothing arguments - if any argument is Nothing, the function returns Nothing
	for _, arg := range resolvedArgs {
		if _, ok := arg.(Nothing); ok {
			return Nothing{}, nil
		}
	}

	result, err := fn.Call(resolvedArgs)
	if err != nil {
		// Function error → return Nothing instead of error
		return Nothing{}, nil
	}

	// Normalize numeric types
	switch v := result.(type) {
	case int:
		result = float64(v)
	case int64:
		result = float64(v)
	case int32:
		result = float64(v)
	case float32:
		result = float64(v)
	}

	return result, nil
}

// normalizePathWhitespace removes whitespace from paths while preserving brackets and quoted strings
func normalizePathWhitespace(path string) string {
	var result []byte
	inQuotes := false
	inSingleQuotes := false
	inBrackets := false
	for i := 0; i < len(path); i++ {
		ch := path[i]
		if (inQuotes || inSingleQuotes) && ch == '\\' && i+1 < len(path) {
			result = append(result, ch)
			i++
			result = append(result, path[i])
			continue
		}
		if ch == '"' && !inSingleQuotes {
			inQuotes = !inQuotes
			result = append(result, ch)
			continue
		}
		if ch == '\'' && !inQuotes {
			inSingleQuotes = !inSingleQuotes
			result = append(result, ch)
			continue
		}
		if inQuotes || inSingleQuotes {
			result = append(result, ch)
			continue
		}
		if ch == '[' {
			inBrackets = true
			result = append(result, ch)
			continue
		}
		if ch == ']' {
			inBrackets = false
			result = append(result, ch)
			continue
		}
		// Skip whitespace outside of brackets and quotes
		if (ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r') && !inBrackets {
			continue
		}
		result = append(result, ch)
	}
	return string(result)
}
