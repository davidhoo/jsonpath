package segments

import (
	"fmt"
	"reflect"
)

// SliceSegment 表示切片段
type SliceSegment struct {
	start int // 起始索引
	end   int // 结束索引（-1 表示到末尾）
	step  int // 步长
}

// NewSliceSegment 创建一个新的切片段
func NewSliceSegment(start, end, step int) *SliceSegment {
	return &SliceSegment{
		start: start,
		end:   end,
		step:  step,
	}
}

// Evaluate 实现 Segment 接口
func (s *SliceSegment) Evaluate(value interface{}) ([]interface{}, error) {
	if value == nil {
		return nil, nil
	}

	// 使用反射获取值的类型
	val := reflect.ValueOf(value)
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		return nil, fmt.Errorf("cannot use slice on non-array value")
	}

	length := val.Len()
	if length == 0 {
		return []interface{}{}, nil
	}

	// 处理 start 索引
	start := s.start
	if start < 0 {
		start = length + start
	}
	if start < 0 {
		start = 0
	}
	if start > length {
		start = length
	}

	fmt.Printf("After start processing: start=%d\n", start)

	// 处理 end 索引
	end := s.end
	if end < 0 {
		// 对于负数索引，转换为正数索引
		end = length + end
	}
	if end < 0 {
		end = 0
	}
	if end > length {
		end = length
	}

	fmt.Printf("After end processing: end=%d\n", end)

	// 根据步长的正负决定遍历方向和边界
	var result []interface{}
	if s.step > 0 {
		// 正向遍历：[start, end)
		if end > start {
			for i := start; i < end; i += s.step {
				result = append(result, val.Index(i).Interface())
			}
		}
	} else {
		// 反向遍历：[start, end)
		if start >= length {
			start = length - 1
		}
		for i := start; i > end; i += s.step {
			result = append(result, val.Index(i).Interface())
		}
	}

	return result, nil
}

// normalizeIndex 规范化索引，处理负数索引
func (s *SliceSegment) normalizeIndex(index, length int) int {
	if index < 0 {
		// 负数索引从末尾开始计数
		index = length + index
	}
	// 确保索引在有效范围内
	if index < 0 {
		index = 0
	}
	if index > length {
		index = length
	}
	return index
}

// String 实现 Segment 接口
func (s *SliceSegment) String() string {
	if s.step == 1 {
		if s.start == 0 {
			return fmt.Sprintf("[:%d]", s.end)
		}
		if s.end == -1 {
			return fmt.Sprintf("[%d:]", s.start)
		}
		return fmt.Sprintf("[%d:%d]", s.start, s.end)
	}
	return fmt.Sprintf("[%d:%d:%d]", s.start, s.end, s.step)
}
