package segments

import (
	"reflect"
)

// RecursiveSegment 表示递归段
type RecursiveSegment struct{}

// NewRecursiveSegment 创建一个新的递归段
func NewRecursiveSegment() *RecursiveSegment {
	return &RecursiveSegment{}
}

// Evaluate 实现 Segment 接口
func (s *RecursiveSegment) Evaluate(value interface{}) ([]interface{}, error) {
	if value == nil {
		return nil, nil
	}

	// 使用递归收集所有值
	result := make([]interface{}, 0)
	s.collectValues(value, &result)
	return result, nil
}

// collectValues 递归收集所有值
func (s *RecursiveSegment) collectValues(value interface{}, result *[]interface{}) {
	if value == nil {
		return
	}

	// 将当前值添加到结果中
	*result = append(*result, value)

	// 使用反射获取值的类型
	val := reflect.ValueOf(value)
	switch val.Kind() {
	case reflect.Map:
		// 递归处理对象的所有属性
		for _, key := range val.MapKeys() {
			s.collectValues(val.MapIndex(key).Interface(), result)
		}
	case reflect.Slice, reflect.Array:
		// 递归处理数组的所有元素
		for i := 0; i < val.Len(); i++ {
			s.collectValues(val.Index(i).Interface(), result)
		}
	}
}

// String 实现 Segment 接口
func (s *RecursiveSegment) String() string {
	return ".."
}
