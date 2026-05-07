package jsonpath

import (
	"encoding/json"
	"fmt"
)

// Query executes a JSONPath query on JSON data and returns a NodeList.
// Each Node contains a Location (Normalized Path) and the corresponding Value.
func Query(data interface{}, path string) (NodeList, error) {
	// If data is a string, parse it as JSON
	if jsonStr, ok := data.(string); ok {
		var parsedData interface{}
		if err := json.Unmarshal([]byte(jsonStr), &parsedData); err != nil {
			return nil, fmt.Errorf("invalid JSON: %v", err)
		}
		data = parsedData
	}

	// Parse path into segments
	segments, err := parse(path)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %v", err)
	}

	// Convert to v3 segments
	v3Segments := wrapSegments(segments)

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

	return nodeList, nil
}
