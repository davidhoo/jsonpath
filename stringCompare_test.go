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
