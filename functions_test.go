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
