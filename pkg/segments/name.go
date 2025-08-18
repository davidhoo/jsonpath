package segments

import (
	"fmt"
	"reflect"
)

// NameSegment 表示名称段
type NameSegment struct {
	name string
}

// NewNameSegment 创建一个新的名称段
func NewNameSegment(name string) *NameSegment {
	return &NameSegment{name: name}
}

// Evaluate 实现 Segment 接口
func (s *NameSegment) Evaluate(value interface{}) ([]interface{}, error) {
	if value == nil {
		return nil, nil
	}

	// 使用反射获取对象的属性
	val := reflect.ValueOf(value)
	if val.Kind() != reflect.Map {
		return nil, fmt.Errorf("cannot access property '%s' on non-object value", s.name)
	}

	// 获取属性值
	propVal := val.MapIndex(reflect.ValueOf(s.name))
	if !propVal.IsValid() {
		return nil, nil // 属性不存在时返回空结果
	}

	return []interface{}{propVal.Interface()}, nil
}

// String 实现 Segment 接口
func (s *NameSegment) String() string {
	return s.name
}
