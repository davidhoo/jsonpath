package jsonpath

import (
	"strings"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

// stringCompareOptions 定义字符串比较选项
type stringCompareOptions struct {
	caseSensitive bool         // 是否大小写敏感
	locale        language.Tag // 用于排序的语言环境
}

// defaultStringCompareOptions 返回默认的字符串比较选项
func defaultStringCompareOptions() stringCompareOptions {
	return stringCompareOptions{
		caseSensitive: true,
		locale:        language.English, // 默认使用英语排序规则
	}
}

// stringComparer 提供标准化的字符串比较功能
type stringComparer struct {
	options  stringCompareOptions
	collator *collate.Collator
}

// newStringComparer 创建一个新的字符串比较器
func newStringComparer(opts ...stringCompareOptions) *stringComparer {
	options := defaultStringCompareOptions()
	if len(opts) > 0 {
		options = opts[0]
	}

	return &stringComparer{
		options:  options,
		collator: collate.New(options.locale),
	}
}

// compare 比较两个字符串
// 返回值:
// -1: a < b
//
//	0: a == b
//	1: a > b
func (sc *stringComparer) compare(a, b string) int {
	if sc.options.caseSensitive {
		// 当大小写敏感且使用默认英语环境时，使用 strings.Compare
		if sc.options.locale == language.English {
			return strings.Compare(a, b)
		}
		// 当大小写敏感且使用其他语言环境时，使用 collator
		result := sc.collator.CompareString(a, b)
		if result < 0 {
			return -1
		} else if result > 0 {
			return 1
		}
		return 0
	}

	// 当大小写不敏感时，使用 collator 并转换为小写
	result := sc.collator.CompareString(strings.ToLower(a), strings.ToLower(b))
	if result < 0 {
		return -1
	} else if result > 0 {
		return 1
	}
	return 0
}

// equals 检查两个字符串是否相等
func (sc *stringComparer) equals(a, b string) bool {
	return sc.compare(a, b) == 0
}

// lessThan 检查 a 是否小于 b
func (sc *stringComparer) lessThan(a, b string) bool {
	return sc.compare(a, b) < 0
}

// lessThanOrEqual 检查 a 是否小于等于 b
func (sc *stringComparer) lessThanOrEqual(a, b string) bool {
	return sc.compare(a, b) <= 0
}

// greaterThan 检查 a 是否大于 b
func (sc *stringComparer) greaterThan(a, b string) bool {
	return sc.compare(a, b) > 0
}

// greaterThanOrEqual 检查 a 是否大于等于 b
func (sc *stringComparer) greaterThanOrEqual(a, b string) bool {
	return sc.compare(a, b) >= 0
}

// standardCompareStrings 使用指定的操作符比较两个字符串
func standardCompareStrings(a string, operator string, b string) bool {
	sc := newStringComparer()

	switch operator {
	case "==":
		return sc.equals(a, b)
	case "!=":
		return !sc.equals(a, b)
	case "<":
		return sc.lessThan(a, b)
	case "<=":
		return sc.lessThanOrEqual(a, b)
	case ">":
		return sc.greaterThan(a, b)
	case ">=":
		return sc.greaterThanOrEqual(a, b)
	}
	return false
}
