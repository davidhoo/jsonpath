package segments

import (
	"fmt"
	"reflect"
)

// WildcardSegment 表示通配符段
type WildcardSegment struct{}

// NewWildcardSegment 创建一个新的通配符段
func NewWildcardSegment() *WildcardSegment {
	return &WildcardSegment{}
}

// Evaluate 实现 Segment 接口
func (s *WildcardSegment) Evaluate(value interface{}) ([]interface{}, error) {
	if value == nil {
		return nil, nil
	}

	// 使用反射获取值的类型
	val := reflect.ValueOf(value)
	switch val.Kind() {
	case reflect.Map:
		// 对于对象，返回所有属性值
		result := make([]interface{}, 0, val.Len())
		for _, key := range val.MapKeys() {
			result = append(result, val.MapIndex(key).Interface())
		}
		return result, nil
	case reflect.Slice, reflect.Array:
		// 对于数组，返回所有元素
		result := make([]interface{}, val.Len())
		for i := 0; i < val.Len(); i++ {
			result[i] = val.Index(i).Interface()
		}
		return result, nil
	default:
		return nil, fmt.Errorf("cannot use wildcard on non-object and non-array value")
	}
}

// String 实现 Segment 接口
func (s *WildcardSegment) String() string {
	return "*"
}
