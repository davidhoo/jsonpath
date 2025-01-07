package jsonpath

import (
	"fmt"
	"math"
	"testing"
)

// NumberTestSuite 包含所有数值相关的测试用例
type NumberTestSuite struct {
	// 可以添加共享的测试数据或辅助函数
}

// floatEquals 用于比较浮点数是否相等，考虑精度误差
func (s *NumberTestSuite) floatEquals(a, b float64) bool {
	const epsilon = 1e-6
	return math.Abs(a-b) < epsilon
}

// TestConvertToNumber 测试数值转换功能
func (s *NumberTestSuite) TestConvertToNumber(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    numberValue
		wantErr bool
	}{
		// 整数类型测试
		{
			name:    "convert int",
			input:   42,
			want:    numberValue{typ: numberTypeInteger, value: 42},
			wantErr: false,
		},
		{
			name:    "convert int32",
			input:   int32(42),
			want:    numberValue{typ: numberTypeInteger, value: 42},
			wantErr: false,
		},
		{
			name:    "convert int64",
			input:   int64(42),
			want:    numberValue{typ: numberTypeInteger, value: 42},
			wantErr: false,
		},
		// 浮点数类型测试
		{
			name:    "convert float32",
			input:   float32(3.14),
			want:    numberValue{typ: numberTypeFloat, value: 3.14},
			wantErr: false,
		},
		{
			name:    "convert float64",
			input:   3.14,
			want:    numberValue{typ: numberTypeFloat, value: 3.14},
			wantErr: false,
		},
		// 字符串类型测试
		{
			name:    "convert string integer",
			input:   "42",
			want:    numberValue{typ: numberTypeInteger, value: 42},
			wantErr: false,
		},
		{
			name:    "convert string float",
			input:   "3.14",
			want:    numberValue{typ: numberTypeFloat, value: 3.14},
			wantErr: false,
		},
		// 特殊值测试
		{
			name:    "convert NaN",
			input:   math.NaN(),
			want:    numberValue{typ: numberTypeNaN},
			wantErr: false,
		},
		{
			name:    "convert +Inf",
			input:   math.Inf(1),
			want:    numberValue{typ: numberTypeInfinity, value: 1},
			wantErr: false,
		},
		{
			name:    "convert -Inf",
			input:   math.Inf(-1),
			want:    numberValue{typ: numberTypeNegativeInfinity, value: -1},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertToNumber(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertToNumber() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.typ != tt.want.typ {
				t.Errorf("convertToNumber() type = %v, want %v", got.typ, tt.want.typ)
			}
			if got.typ != numberTypeNaN && !s.floatEquals(got.value, tt.want.value) {
				t.Errorf("convertToNumber() value = %v, want %v", got.value, tt.want.value)
			}
		})
	}
}

// CompareTestSuite 包含所有比较相关的测试用例
type CompareTestSuite struct {
	comparer *stringComparer
}

// Setup 初始化比较测试套件
func (s *CompareTestSuite) Setup() {
	s.comparer = newStringComparer()
}

// TestCompareValues 测试值比较功能
func (s *CompareTestSuite) TestCompareValues(t *testing.T) {
	for _, tt := range []struct {
		name     string
		a        interface{}
		operator string
		b        interface{}
		want     bool
	}{
		{
			name:     "compare equal integers",
			a:        42,
			operator: "==",
			b:        42,
			want:     true,
		},
		{
			name:     "compare different integers",
			a:        42,
			operator: ">",
			b:        24,
			want:     true,
		},
		{
			name:     "compare string with number",
			a:        "42",
			operator: "!=",
			b:        "24",
			want:     true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := standardCompareStrings(fmt.Sprint(tt.a), tt.operator, fmt.Sprint(tt.b))
			if got != tt.want {
				t.Errorf("compareValues() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestCompareBooleans 测试布尔值比较功能
func (s *CompareTestSuite) TestCompareBooleans(t *testing.T) {
	if s.comparer == nil {
		s.Setup()
	}
	for _, tt := range []struct {
		name string
		a    bool
		b    bool
		want int
	}{
		{
			name: "compare true with false",
			a:    true,
			b:    false,
			want: 1,
		},
		{
			name: "compare false with true",
			a:    false,
			b:    true,
			want: -1,
		},
		{
			name: "compare equal booleans",
			a:    true,
			b:    true,
			want: 0,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := s.comparer.compare(fmt.Sprint(tt.a), fmt.Sprint(tt.b))
			if got != tt.want {
				t.Errorf("compareBooleans() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestFunctions 运行所有函数测试
func TestFunctions(t *testing.T) {
	numberSuite := &NumberTestSuite{}
	compareSuite := &CompareTestSuite{}
	compareSuite.Setup()

	t.Run("NumberTestSuite", func(t *testing.T) {
		t.Run("TestConvertToNumber", func(t *testing.T) {
			numberSuite.TestConvertToNumber(t)
		})
	})

	t.Run("CompareTestSuite", func(t *testing.T) {
		t.Run("TestCompareValues", func(t *testing.T) {
			compareSuite.TestCompareValues(t)
		})
		t.Run("TestCompareBooleans", func(t *testing.T) {
			compareSuite.TestCompareBooleans(t)
		})
	})
}
