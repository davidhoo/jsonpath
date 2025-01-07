package jsonpath

import (
	"testing"

	"golang.org/x/text/language"
)

func TestStandardCompareStrings(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		operator string
		b        string
		want     bool
	}{
		// 基本比较测试
		{"equal strings", "hello", "==", "hello", true},
		{"not equal strings", "hello", "!=", "world", true},
		{"less than", "apple", "<", "banana", true},
		{"less than or equal", "apple", "<=", "apple", true},
		{"greater than", "banana", ">", "apple", true},
		{"greater than or equal", "banana", ">=", "banana", true},

		// Unicode 字符测试
		{"unicode equal", "你好", "==", "你好", true},
		{"unicode not equal", "你好", "!=", "世界", true},
		{"unicode less than", "一", "<", "二", true},
		{"unicode greater than", "二", ">", "一", true},

		// 特殊字符测试
		{"special chars equal", "hello!", "==", "hello!", true},
		{"special chars not equal", "hello!", "!=", "hello?", true},
		{"special chars less than", "hello!", "<", "hello?", true},
		{"special chars greater than", "hello?", ">", "hello!", true},

		// 空字符串测试
		{"empty string equal", "", "==", "", true},
		{"empty string not equal", "", "!=", "hello", true},
		{"empty string less than", "", "<", "hello", true},
		{"empty string greater than", "hello", ">", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := standardCompareStrings(tt.a, tt.operator, tt.b)
			if got != tt.want {
				t.Errorf("standardCompareStrings(%q, %q, %q) = %v, want %v",
					tt.a, tt.operator, tt.b, got, tt.want)
			}
		})
	}
}

func TestStringComparerWithOptions(t *testing.T) {
	tests := []struct {
		name     string
		options  stringCompareOptions
		a        string
		operator string
		b        string
		want     bool
	}{
		// 大小写不敏感测试
		{
			name: "case insensitive equal",
			options: stringCompareOptions{
				caseSensitive: false,
				locale:        language.English,
			},
			a:        "Hello",
			operator: "==",
			b:        "hello",
			want:     true,
		},
		{
			name: "case insensitive not equal",
			options: stringCompareOptions{
				caseSensitive: false,
				locale:        language.English,
			},
			a:        "Hello",
			operator: "!=",
			b:        "world",
			want:     true,
		},

		// 不同语言环境测试
		{
			name: "german locale",
			options: stringCompareOptions{
				caseSensitive: true,
				locale:        language.German,
			},
			a:        "ä",
			operator: "<",
			b:        "b",
			want:     true,
		},
		{
			name: "chinese locale",
			options: stringCompareOptions{
				caseSensitive: true,
				locale:        language.Chinese,
			},
			a:        "你",
			operator: "<",
			b:        "我",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := newStringComparer(tt.options)
			var got bool
			switch tt.operator {
			case "==":
				got = sc.equals(tt.a, tt.b)
			case "!=":
				got = !sc.equals(tt.a, tt.b)
			case "<":
				got = sc.lessThan(tt.a, tt.b)
			case "<=":
				got = sc.lessThanOrEqual(tt.a, tt.b)
			case ">":
				got = sc.greaterThan(tt.a, tt.b)
			case ">=":
				got = sc.greaterThanOrEqual(tt.a, tt.b)
			}
			if got != tt.want {
				t.Errorf("stringComparer.compare(%q, %q, %q) with options %+v = %v, want %v",
					tt.a, tt.operator, tt.b, tt.options, got, tt.want)
			}
		})
	}
}

func TestCompareStringsFunction(t *testing.T) {
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
