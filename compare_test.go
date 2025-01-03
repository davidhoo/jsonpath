package jsonpath

import (
	"testing"
)

func TestCompareValues(t *testing.T) {
	tests := []struct {
		name     string
		a        interface{}
		operator string
		b        interface{}
		want     bool
		wantErr  bool
	}{
		// nil 值比较
		{
			name:     "nil equals nil",
			a:        nil,
			operator: "==",
			b:        nil,
			want:     true,
		},
		{
			name:     "nil not equals value",
			a:        nil,
			operator: "==",
			b:        42,
			want:     false,
		},
		{
			name:     "value not equals nil",
			a:        42,
			operator: "==",
			b:        nil,
			want:     false,
		},
		{
			name:     "nil not equals nil",
			a:        nil,
			operator: "!=",
			b:        nil,
			want:     false,
		},
		{
			name:     "nil other operator",
			a:        nil,
			operator: ">",
			b:        nil,
			want:     false,
		},

		// 数字比较 (float64)
		{
			name:     "float64 equals",
			a:        float64(42),
			operator: "==",
			b:        float64(42),
			want:     true,
		},
		{
			name:     "float64 not equals",
			a:        float64(42),
			operator: "!=",
			b:        float64(43),
			want:     true,
		},
		{
			name:     "float64 less than",
			a:        float64(42),
			operator: "<",
			b:        float64(43),
			want:     true,
		},
		{
			name:     "float64 less than or equal",
			a:        float64(42),
			operator: "<=",
			b:        float64(42),
			want:     true,
		},
		{
			name:     "float64 greater than",
			a:        float64(43),
			operator: ">",
			b:        float64(42),
			want:     true,
		},
		{
			name:     "float64 greater than or equal",
			a:        float64(42),
			operator: ">=",
			b:        float64(42),
			want:     true,
		},

		// 数字比较 (int64)
		{
			name:     "int64 equals",
			a:        int64(42),
			operator: "==",
			b:        int64(42),
			want:     true,
		},
		{
			name:     "int64 with float64",
			a:        int64(42),
			operator: "==",
			b:        float64(42),
			want:     true,
		},
		{
			name:     "float64 with int64",
			a:        float64(42),
			operator: "==",
			b:        int64(42),
			want:     true,
		},

		// 字符串比较
		{
			name:     "string equals",
			a:        "hello",
			operator: "==",
			b:        "hello",
			want:     true,
		},
		{
			name:     "string not equals",
			a:        "hello",
			operator: "!=",
			b:        "world",
			want:     true,
		},
		{
			name:     "string less than",
			a:        "apple",
			operator: "<",
			b:        "banana",
			want:     true,
		},
		{
			name:     "string greater than",
			a:        "banana",
			operator: ">",
			b:        "apple",
			want:     true,
		},

		// 布尔值比较
		{
			name:     "bool equals true",
			a:        true,
			operator: "==",
			b:        true,
			want:     true,
		},
		{
			name:     "bool equals false",
			a:        false,
			operator: "==",
			b:        false,
			want:     true,
		},
		{
			name:     "bool not equals",
			a:        true,
			operator: "!=",
			b:        false,
			want:     true,
		},

		// 类型不匹配
		{
			name:     "incompatible types",
			a:        "42",
			operator: "==",
			b:        42,
			wantErr:  true,
		},

		// 正则匹配
		{
			name:     "match operator with matching string",
			a:        "hello world",
			operator: "match",
			b:        "world$",
			want:     true,
		},
		{
			name:     "match operator with non-matching string",
			a:        "hello world",
			operator: "match",
			b:        "^world",
			want:     false,
		},
		{
			name:     "match operator with invalid pattern",
			a:        "hello world",
			operator: "match",
			b:        "[",
			want:     false,
		},
		{
			name:     "match operator with non-string value",
			a:        42,
			operator: "match",
			b:        "42",
			wantErr:  true,
		},
		{
			name:     "match operator with non-string pattern",
			a:        "42",
			operator: "match",
			b:        42,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := compareValues(tt.a, tt.operator, tt.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("compareValues() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
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
		{
			name:     "case sensitive comparison",
			a:        "Apple",
			operator: "<",
			b:        "apple",
			want:     true,
		},
		{
			name:     "empty strings equal",
			a:        "",
			operator: "==",
			b:        "",
			want:     true,
		},
		{
			name:     "empty string less than non-empty",
			a:        "",
			operator: "<",
			b:        "a",
			want:     true,
		},
		{
			name:     "invalid operator",
			a:        "hello",
			operator: "invalid",
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
			name:     "true equals false",
			a:        true,
			operator: "==",
			b:        false,
			want:     false,
		},
		{
			name:     "false equals true",
			a:        false,
			operator: "==",
			b:        true,
			want:     false,
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
		{
			name:     "invalid operator",
			a:        true,
			operator: ">",
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
