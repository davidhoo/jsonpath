package jsonpath

import (
	"encoding/json"
	"math"
	"testing"
)

// floatEquals 用于比较浮点数是否相等，考虑精度误差
func floatEquals(a, b float64) bool {
	const epsilon = 1e-6
	return math.Abs(a-b) < epsilon
}

func TestConvertToNumber(t *testing.T) {
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
		{
			name:    "convert float64 integer value",
			input:   42.0,
			want:    numberValue{typ: numberTypeInteger, value: 42},
			wantErr: false,
		},

		// 特殊浮点数测试
		{
			name:    "convert NaN float32",
			input:   float32(math.NaN()),
			want:    numberValue{typ: numberTypeNaN},
			wantErr: false,
		},
		{
			name:    "convert NaN float64",
			input:   math.NaN(),
			want:    numberValue{typ: numberTypeNaN},
			wantErr: false,
		},
		{
			name:    "convert +Inf float32",
			input:   float32(math.Inf(1)),
			want:    numberValue{typ: numberTypeInfinity},
			wantErr: false,
		},
		{
			name:    "convert +Inf float64",
			input:   math.Inf(1),
			want:    numberValue{typ: numberTypeInfinity},
			wantErr: false,
		},
		{
			name:    "convert -Inf float32",
			input:   float32(math.Inf(-1)),
			want:    numberValue{typ: numberTypeNegativeInfinity},
			wantErr: false,
		},
		{
			name:    "convert -Inf float64",
			input:   math.Inf(-1),
			want:    numberValue{typ: numberTypeNegativeInfinity},
			wantErr: false,
		},

		// json.Number 类型测试
		{
			name:    "convert json.Number integer",
			input:   json.Number("42"),
			want:    numberValue{typ: numberTypeInteger, value: 42},
			wantErr: false,
		},
		{
			name:    "convert json.Number float",
			input:   json.Number("3.14"),
			want:    numberValue{typ: numberTypeFloat, value: 3.14},
			wantErr: false,
		},
		{
			name:    "convert invalid json.Number",
			input:   json.Number("invalid"),
			wantErr: true,
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
		{
			name:    "convert invalid string",
			input:   "invalid",
			wantErr: true,
		},

		// 边界值测试
		{
			name:    "convert max int64",
			input:   int64(math.MaxInt64),
			want:    numberValue{typ: numberTypeInteger, value: float64(math.MaxInt64)},
			wantErr: false,
		},
		{
			name:    "convert min int64",
			input:   int64(math.MinInt64),
			want:    numberValue{typ: numberTypeInteger, value: float64(math.MinInt64)},
			wantErr: false,
		},

		// 无效类型测试
		{
			name:    "convert bool",
			input:   true,
			wantErr: true,
		},
		{
			name:    "convert nil",
			input:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertToNumber(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertToNumber() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.typ != tt.want.typ {
					t.Errorf("convertToNumber() type = %v, want %v", got.typ, tt.want.typ)
				}
				if got.typ != numberTypeNaN && !floatEquals(got.value, tt.want.value) {
					t.Errorf("convertToNumber() value = %v, want %v", got.value, tt.want.value)
				}
			}
		})
	}
}

func TestCompareNumberValues(t *testing.T) {
	tests := []struct {
		name     string
		a        numberValue
		b        numberValue
		expected int
	}{
		{
			name:     "equal integers",
			a:        numberValue{typ: numberTypeInteger, value: 42},
			b:        numberValue{typ: numberTypeInteger, value: 42},
			expected: 0,
		},
		{
			name:     "less than integer",
			a:        numberValue{typ: numberTypeInteger, value: 41},
			b:        numberValue{typ: numberTypeInteger, value: 42},
			expected: -1,
		},
		{
			name:     "greater than integer",
			a:        numberValue{typ: numberTypeInteger, value: 43},
			b:        numberValue{typ: numberTypeInteger, value: 42},
			expected: 1,
		},
		{
			name:     "equal floats",
			a:        numberValue{typ: numberTypeFloat, value: 3.14},
			b:        numberValue{typ: numberTypeFloat, value: 3.14},
			expected: 0,
		},
		{
			name:     "less than float",
			a:        numberValue{typ: numberTypeFloat, value: 3.13},
			b:        numberValue{typ: numberTypeFloat, value: 3.14},
			expected: -1,
		},
		{
			name:     "greater than float",
			a:        numberValue{typ: numberTypeFloat, value: 3.15},
			b:        numberValue{typ: numberTypeFloat, value: 3.14},
			expected: 1,
		},
		{
			name:     "float precision test",
			a:        numberValue{typ: numberTypeFloat, value: 0.1 + 0.2},
			b:        numberValue{typ: numberTypeFloat, value: 0.3},
			expected: 0,
		},
		{
			name:     "NaN equals NaN",
			a:        numberValue{typ: numberTypeNaN},
			b:        numberValue{typ: numberTypeNaN},
			expected: 0,
		},
		{
			name:     "NaN compared with number",
			a:        numberValue{typ: numberTypeNaN},
			b:        numberValue{typ: numberTypeInteger, value: 42},
			expected: 0,
		},
		{
			name:     "number compared with NaN",
			a:        numberValue{typ: numberTypeInteger, value: 42},
			b:        numberValue{typ: numberTypeNaN},
			expected: 0,
		},
		{
			name:     "Infinity equals Infinity",
			a:        numberValue{typ: numberTypeInfinity},
			b:        numberValue{typ: numberTypeInfinity},
			expected: 0,
		},
		{
			name:     "Infinity greater than number",
			a:        numberValue{typ: numberTypeInfinity},
			b:        numberValue{typ: numberTypeInteger, value: math.MaxFloat64},
			expected: 1,
		},
		{
			name:     "number less than Infinity",
			a:        numberValue{typ: numberTypeInteger, value: math.MaxFloat64},
			b:        numberValue{typ: numberTypeInfinity},
			expected: -1,
		},
		{
			name:     "-Infinity equals -Infinity",
			a:        numberValue{typ: numberTypeNegativeInfinity},
			b:        numberValue{typ: numberTypeNegativeInfinity},
			expected: 0,
		},
		{
			name:     "-Infinity less than number",
			a:        numberValue{typ: numberTypeNegativeInfinity},
			b:        numberValue{typ: numberTypeInteger, value: -math.MaxFloat64},
			expected: -1,
		},
		{
			name:     "number greater than -Infinity",
			a:        numberValue{typ: numberTypeInteger, value: -math.MaxFloat64},
			b:        numberValue{typ: numberTypeNegativeInfinity},
			expected: 1,
		},
		{
			name:     "Infinity greater than -Infinity",
			a:        numberValue{typ: numberTypeInfinity},
			b:        numberValue{typ: numberTypeNegativeInfinity},
			expected: 1,
		},
		{
			name:     "-Infinity less than Infinity",
			a:        numberValue{typ: numberTypeNegativeInfinity},
			b:        numberValue{typ: numberTypeInfinity},
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareNumberValues(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("compareNumberValues(%v, %v) = %v, want %v",
					tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    numberValue
		expected string
	}{
		{
			name:     "format integer",
			input:    numberValue{typ: numberTypeInteger, value: 42},
			expected: "42",
		},
		{
			name:     "format negative integer",
			input:    numberValue{typ: numberTypeInteger, value: -42},
			expected: "-42",
		},
		{
			name:     "format zero",
			input:    numberValue{typ: numberTypeInteger, value: 0},
			expected: "0",
		},
		{
			name:     "format float",
			input:    numberValue{typ: numberTypeFloat, value: 3.14},
			expected: "3.14",
		},
		{
			name:     "format negative float",
			input:    numberValue{typ: numberTypeFloat, value: -3.14},
			expected: "-3.14",
		},
		{
			name:     "format float with trailing zeros",
			input:    numberValue{typ: numberTypeFloat, value: 2.0},
			expected: "2",
		},
		{
			name:     "format large float",
			input:    numberValue{typ: numberTypeFloat, value: 1234567.89},
			expected: "1234567.89",
		},
		{
			name:     "format small float",
			input:    numberValue{typ: numberTypeFloat, value: 0.0000001},
			expected: "1e-7",
		},
		{
			name:     "format NaN",
			input:    numberValue{typ: numberTypeNaN},
			expected: "NaN",
		},
		{
			name:     "format Infinity",
			input:    numberValue{typ: numberTypeInfinity},
			expected: "Infinity",
		},
		{
			name:     "format -Infinity",
			input:    numberValue{typ: numberTypeNegativeInfinity},
			expected: "-Infinity",
		},
		{
			name:     "format max int64",
			input:    numberValue{typ: numberTypeInteger, value: float64(math.MaxInt64)},
			expected: "9223372036854775807",
		},
		{
			name:     "format min int64",
			input:    numberValue{typ: numberTypeInteger, value: float64(math.MinInt64)},
			expected: "-9223372036854775808",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatNumber(tt.input)
			if result != tt.expected {
				t.Errorf("formatNumber(%v) = %v, want %v",
					tt.input, result, tt.expected)
			}
		})
	}
}
