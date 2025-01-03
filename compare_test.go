package jsonpath

import (
	"math"
	"testing"
)

func TestCompareValues(t *testing.T) {
	tests := []struct {
		name    string
		value1  interface{}
		op      string
		value2  interface{}
		want    bool
		wantErr bool
	}{
		{
			name:    "nil equals nil",
			value1:  nil,
			op:      "==",
			value2:  nil,
			want:    true,
			wantErr: false,
		},
		{
			name:    "nil not equals value",
			value1:  nil,
			op:      "!=",
			value2:  42,
			want:    true,
			wantErr: false,
		},
		{
			name:    "nil other operator",
			value1:  nil,
			op:      ">",
			value2:  42,
			want:    false,
			wantErr: true,
		},
		{
			name:    "numbers equal",
			value1:  float64(42),
			op:      "==",
			value2:  float64(42),
			want:    true,
			wantErr: false,
		},
		{
			name:    "numbers not equal",
			value1:  float64(42),
			op:      "!=",
			value2:  float64(43),
			want:    true,
			wantErr: false,
		},
		{
			name:    "numbers greater than",
			value1:  float64(43),
			op:      ">",
			value2:  float64(42),
			want:    true,
			wantErr: false,
		},
		{
			name:    "numbers less than",
			value1:  float64(42),
			op:      "<",
			value2:  float64(43),
			want:    true,
			wantErr: false,
		},
		{
			name:    "numbers greater than or equal",
			value1:  float64(42),
			op:      ">=",
			value2:  float64(42),
			want:    true,
			wantErr: false,
		},
		{
			name:    "numbers less than or equal",
			value1:  float64(42),
			op:      "<=",
			value2:  float64(42),
			want:    true,
			wantErr: false,
		},
		{
			name:    "strings equal",
			value1:  "hello",
			op:      "==",
			value2:  "hello",
			want:    true,
			wantErr: false,
		},
		{
			name:    "strings not equal",
			value1:  "hello",
			op:      "!=",
			value2:  "world",
			want:    true,
			wantErr: false,
		},
		{
			name:    "strings greater than",
			value1:  "world",
			op:      ">",
			value2:  "hello",
			want:    true,
			wantErr: false,
		},
		{
			name:    "strings less than",
			value1:  "hello",
			op:      "<",
			value2:  "world",
			want:    true,
			wantErr: false,
		},
		{
			name:    "strings match",
			value1:  "hello",
			op:      "match",
			value2:  "^h.*o$",
			want:    true,
			wantErr: false,
		},
		{
			name:    "strings not match",
			value1:  "hello",
			op:      "match",
			value2:  "^w.*d$",
			want:    false,
			wantErr: false,
		},
		{
			name:    "match operator with invalid pattern",
			value1:  "hello",
			op:      "match",
			value2:  "[",
			want:    false,
			wantErr: true,
		},
		{
			name:    "match operator with non-string value",
			value1:  42,
			op:      "match",
			value2:  "^h.*o$",
			want:    false,
			wantErr: true,
		},
		{
			name:    "match operator with non-string pattern",
			value1:  "hello",
			op:      "match",
			value2:  42,
			want:    false,
			wantErr: true,
		},
		{
			name:    "booleans equal",
			value1:  true,
			op:      "==",
			value2:  true,
			want:    true,
			wantErr: false,
		},
		{
			name:    "booleans not equal",
			value1:  true,
			op:      "!=",
			value2:  false,
			want:    true,
			wantErr: false,
		},
		{
			name:    "bool invalid operator",
			value1:  true,
			op:      "<",
			value2:  false,
			want:    false,
			wantErr: true,
		},
		{
			name:    "incompatible types - string and number",
			value1:  "42",
			op:      "==",
			value2:  float64(42),
			want:    false,
			wantErr: true,
		},
		{
			name:    "incompatible types - bool and number",
			value1:  true,
			op:      "==",
			value2:  float64(1),
			want:    false,
			wantErr: true,
		},
		{
			name:    "incompatible types - string and bool",
			value1:  "true",
			op:      "==",
			value2:  true,
			want:    false,
			wantErr: true,
		},
		{
			name:    "incompatible types - array and number",
			value1:  []interface{}{1, 2, 3},
			op:      "==",
			value2:  float64(42),
			want:    false,
			wantErr: true,
		},
		{
			name:    "incompatible types - object and number",
			value1:  map[string]interface{}{"key": "value"},
			op:      "==",
			value2:  float64(42),
			want:    false,
			wantErr: true,
		},
		{
			name:    "invalid operator",
			value1:  float64(42),
			op:      "invalid",
			value2:  float64(42),
			want:    false,
			wantErr: true,
		},
		{
			name:    "empty operator",
			value1:  float64(42),
			op:      "",
			value2:  float64(42),
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := compareValues(tt.value1, tt.op, tt.value2)
			if (err != nil) != tt.wantErr {
				t.Errorf("compareValues() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("compareValues() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompareStrings(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		operator string
		b        string
		want     bool
	}{
		// 相等性比较
		{
			name:     "equal strings",
			a:        "hello",
			operator: "==",
			b:        "hello",
			want:     true,
		},
		{
			name:     "not equal strings",
			a:        "hello",
			operator: "!=",
			b:        "world",
			want:     true,
		},
		{
			name:     "equal empty strings",
			a:        "",
			operator: "==",
			b:        "",
			want:     true,
		},
		{
			name:     "empty and non-empty strings",
			a:        "",
			operator: "!=",
			b:        "hello",
			want:     true,
		},

		// 大小比较
		{
			name:     "less than",
			a:        "apple",
			operator: "<",
			b:        "banana",
			want:     true,
		},
		{
			name:     "less than or equal (equal)",
			a:        "apple",
			operator: "<=",
			b:        "apple",
			want:     true,
		},
		{
			name:     "less than or equal (less)",
			a:        "apple",
			operator: "<=",
			b:        "banana",
			want:     true,
		},
		{
			name:     "greater than",
			a:        "banana",
			operator: ">",
			b:        "apple",
			want:     true,
		},
		{
			name:     "greater than or equal (equal)",
			a:        "banana",
			operator: ">=",
			b:        "banana",
			want:     true,
		},
		{
			name:     "greater than or equal (greater)",
			a:        "banana",
			operator: ">=",
			b:        "apple",
			want:     true,
		},

		// 特殊字符
		{
			name:     "strings with spaces",
			a:        "hello world",
			operator: "==",
			b:        "hello world",
			want:     true,
		},
		{
			name:     "strings with newlines",
			a:        "hello\nworld",
			operator: "==",
			b:        "hello\nworld",
			want:     true,
		},
		{
			name:     "strings with tabs",
			a:        "hello\tworld",
			operator: "==",
			b:        "hello\tworld",
			want:     true,
		},

		// Unicode 字符
		{
			name:     "unicode strings equal",
			a:        "你好",
			operator: "==",
			b:        "你好",
			want:     true,
		},
		{
			name:     "unicode strings not equal",
			a:        "你好",
			operator: "!=",
			b:        "世界",
			want:     true,
		},
		{
			name:     "unicode strings comparison",
			a:        "a",
			operator: "<",
			b:        "你",
			want:     true, // ASCII 字符小于 Unicode 字符
		},

		// 大小写敏感性
		{
			name:     "case sensitive equal",
			a:        "Hello",
			operator: "==",
			b:        "hello",
			want:     false,
		},
		{
			name:     "case sensitive less than",
			a:        "Hello",
			operator: "<",
			b:        "hello",
			want:     true, // 大写字母的 ASCII 值小于小写字母
		},

		// 无效操作符
		{
			name:     "invalid operator",
			a:        "hello",
			operator: "invalid",
			b:        "world",
			want:     false,
		},
		{
			name:     "empty operator",
			a:        "hello",
			operator: "",
			b:        "world",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := compareStrings(tt.a, tt.operator, tt.b); got != tt.want {
				t.Errorf("compareStrings() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompareBooleans(t *testing.T) {
	tests := []struct {
		name     string
		a        bool
		operator string
		b        bool
		want     bool
	}{
		// 相等性比较
		{
			name:     "true equals true",
			a:        true,
			operator: "==",
			b:        true,
			want:     true,
		},
		{
			name:     "false equals false",
			a:        false,
			operator: "==",
			b:        false,
			want:     true,
		},
		{
			name:     "true not equals false",
			a:        true,
			operator: "!=",
			b:        false,
			want:     true,
		},
		{
			name:     "false not equals true",
			a:        false,
			operator: "!=",
			b:        true,
			want:     true,
		},
		{
			name:     "true not equals true",
			a:        true,
			operator: "!=",
			b:        true,
			want:     false,
		},
		{
			name:     "false not equals false",
			a:        false,
			operator: "!=",
			b:        false,
			want:     false,
		},

		// 无效操作符
		{
			name:     "invalid operator",
			a:        true,
			operator: "invalid",
			b:        false,
			want:     false,
		},
		{
			name:     "empty operator",
			a:        true,
			operator: "",
			b:        false,
			want:     false,
		},
		{
			name:     "less than operator",
			a:        true,
			operator: "<",
			b:        false,
			want:     false,
		},
		{
			name:     "greater than operator",
			a:        true,
			operator: ">",
			b:        false,
			want:     false,
		},
		{
			name:     "less than or equal operator",
			a:        true,
			operator: "<=",
			b:        false,
			want:     false,
		},
		{
			name:     "greater than or equal operator",
			a:        true,
			operator: ">=",
			b:        false,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := compareBooleans(tt.a, tt.operator, tt.b); got != tt.want {
				t.Errorf("compareBooleans() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompareNumbers(t *testing.T) {
	tests := []struct {
		name     string
		a        float64
		operator string
		b        float64
		want     bool
	}{
		// 相等性比较
		{
			name:     "equal numbers",
			a:        42.0,
			operator: "==",
			b:        42.0,
			want:     true,
		},
		{
			name:     "not equal numbers",
			a:        42.0,
			operator: "!=",
			b:        43.0,
			want:     true,
		},
		{
			name:     "equal with decimals",
			a:        42.5,
			operator: "==",
			b:        42.5,
			want:     true,
		},
		{
			name:     "not equal with decimals",
			a:        42.5,
			operator: "!=",
			b:        42.6,
			want:     true,
		},

		// 大小比较
		{
			name:     "less than",
			a:        42.0,
			operator: "<",
			b:        43.0,
			want:     true,
		},
		{
			name:     "less than or equal (equal)",
			a:        42.0,
			operator: "<=",
			b:        42.0,
			want:     true,
		},
		{
			name:     "less than or equal (less)",
			a:        42.0,
			operator: "<=",
			b:        43.0,
			want:     true,
		},
		{
			name:     "greater than",
			a:        43.0,
			operator: ">",
			b:        42.0,
			want:     true,
		},
		{
			name:     "greater than or equal (equal)",
			a:        42.0,
			operator: ">=",
			b:        42.0,
			want:     true,
		},
		{
			name:     "greater than or equal (greater)",
			a:        43.0,
			operator: ">=",
			b:        42.0,
			want:     true,
		},

		// 特殊值
		{
			name:     "zero equals zero",
			a:        0.0,
			operator: "==",
			b:        0.0,
			want:     true,
		},
		{
			name:     "negative equals negative",
			a:        -42.0,
			operator: "==",
			b:        -42.0,
			want:     true,
		},
		{
			name:     "positive infinity equals positive infinity",
			a:        math.Inf(1),
			operator: "==",
			b:        math.Inf(1),
			want:     true,
		},
		{
			name:     "negative infinity equals negative infinity",
			a:        math.Inf(-1),
			operator: "==",
			b:        math.Inf(-1),
			want:     true,
		},
		{
			name:     "NaN not equals NaN",
			a:        math.NaN(),
			operator: "==",
			b:        math.NaN(),
			want:     false,
		},

		// 边界值
		{
			name:     "max float64",
			a:        math.MaxFloat64,
			operator: "==",
			b:        math.MaxFloat64,
			want:     true,
		},
		{
			name:     "smallest positive float64",
			a:        math.SmallestNonzeroFloat64,
			operator: "==",
			b:        math.SmallestNonzeroFloat64,
			want:     true,
		},
		{
			name:     "epsilon difference",
			a:        1.0,
			operator: "!=",
			b:        1.0000000000000002,
			want:     true,
		},

		// 无效操作符
		{
			name:     "invalid operator",
			a:        42.0,
			operator: "invalid",
			b:        42.0,
			want:     false,
		},
		{
			name:     "empty operator",
			a:        42.0,
			operator: "",
			b:        42.0,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := compareNumbers(tt.a, tt.operator, tt.b); got != tt.want {
				t.Errorf("compareNumbers() = %v, want %v", got, tt.want)
			}
		})
	}
}
