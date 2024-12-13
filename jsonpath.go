package jsonpath

import (
	"encoding/json"
	"fmt"
)

// Query executes a JSONPath query on a JSON string and returns the result
func Query(jsonStr string, path string) (interface{}, error) {
	// Parse JSON
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, fmt.Errorf("invalid JSON: %v", err)
	}

	// Parse path
	segments, err := parse(path)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %v", err)
	}

	// Evaluate segments
	result := []interface{}{data}
	for _, seg := range segments {
		var newResult []interface{}
		for _, val := range result {
			evaluated, err := seg.evaluate(val)
			if err != nil {
				return nil, err
			}
			newResult = append(newResult, evaluated...)
		}
		result = newResult
	}

	// 根据最后一个段的类型决定返回格式
	if len(segments) > 0 {
		switch segments[len(segments)-1].(type) {
		case *filterSegment:
			return result, nil
		case *functionSegment:
			if len(result) == 1 {
				return result[0], nil
			}
			return result, nil
		}
	}

	// 对于其他情况，如果只有一个结果，返回单个值
	if len(result) == 1 {
		return result[0], nil
	}
	return result, nil
}
