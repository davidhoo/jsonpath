package jsonpath

import (
	"encoding/json"
	"fmt"
)

// Query executes a JSONPath query on JSON data and returns the result
func Query(data interface{}, path string) (interface{}, error) {
	// If data is a string, parse it as JSON
	if jsonStr, ok := data.(string); ok {
		var parsedData interface{}
		if err := json.Unmarshal([]byte(jsonStr), &parsedData); err != nil {
			return nil, fmt.Errorf("invalid JSON: %v", err)
		}
		data = parsedData
	}

	// Parse path into old segments
	oldSegments, err := parse(path)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %v", err)
	}

	// Convert to v3 segments
	v3Segments := wrapSegments(oldSegments)

	// Build root node
	root := Node{Location: "$", Value: data}

	// Evaluate segments using v3 pipeline
	nodeList := NodeList{root}
	for _, seg := range v3Segments {
		var newNodeList NodeList
		for _, n := range nodeList {
			evaluated, err := seg.evaluate(n)
			if err != nil {
				return nil, err
			}
			newNodeList = append(newNodeList, evaluated...)
		}
		nodeList = newNodeList
	}

	// Extract values from nodelist for backward compatibility
	result := make([]interface{}, len(nodeList))
	for i, n := range nodeList {
		result[i] = n.Value
	}

	// Determine return format based on last segment type
	if len(oldSegments) > 0 {
		switch oldSegments[len(oldSegments)-1].(type) {
		case *filterSegment:
			return result, nil
		case *functionSegment:
			if len(result) == 1 {
				return result[0], nil
			}
			return result, nil
		}
	}

	// For other cases, return single value if only one result
	if len(result) == 1 {
		return result[0], nil
	}
	if result == nil {
		return []interface{}{}, nil
	}
	return result, nil
}
