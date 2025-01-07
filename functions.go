package jsonpath

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// Function represents a JSONPath function
type Function interface {
	Call(args []interface{}) (interface{}, error)
	Name() string
}

// builtinFunction is a helper type for implementing Function interface
type builtinFunction struct {
	name     string
	callback func([]interface{}) (interface{}, error)
}

func (f *builtinFunction) Call(args []interface{}) (interface{}, error) {
	return f.callback(args)
}

func (f *builtinFunction) Name() string {
	return f.name
}

// regexCache 用于缓存编译后的正则表达式
var regexCache = make(map[string]*regexp.Regexp)
var regexCacheMutex sync.RWMutex

// getCompiledRegex 从缓存中获取或编译正则表达式
func getCompiledRegex(pattern string) (*regexp.Regexp, error) {
	// 先尝试从缓存中读取
	regexCacheMutex.RLock()
	if re, ok := regexCache[pattern]; ok {
		regexCacheMutex.RUnlock()
		return re, nil
	}
	regexCacheMutex.RUnlock()

	// 如果缓存中没有，则编译正则表达式
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	// 将编译后的正则表达式存入缓存
	regexCacheMutex.Lock()
	regexCache[pattern] = re
	regexCacheMutex.Unlock()

	return re, nil
}

// numberType 表示数值类型
type numberType int

const (
	numberTypeInteger numberType = iota
	numberTypeFloat
	numberTypeNaN
	numberTypeInfinity
	numberTypeNegativeInfinity
)

// numberValue 表示标准化的数值
type numberValue struct {
	typ   numberType
	value float64
}

// convertToNumber 将任意值转换为标准化的数值
func convertToNumber(v interface{}) (numberValue, error) {
	switch val := v.(type) {
	case int:
		return numberValue{typ: numberTypeInteger, value: float64(val)}, nil
	case int32:
		return numberValue{typ: numberTypeInteger, value: float64(val)}, nil
	case int64:
		return numberValue{typ: numberTypeInteger, value: float64(val)}, nil
	case float32:
		if isNaN32(float32(val)) {
			return numberValue{typ: numberTypeNaN}, nil
		}
		if isInf32(float32(val), 1) {
			return numberValue{typ: numberTypeInfinity, value: 1}, nil
		}
		if isInf32(float32(val), -1) {
			return numberValue{typ: numberTypeNegativeInfinity, value: -1}, nil
		}
		return numberValue{typ: numberTypeFloat, value: float64(val)}, nil
	case float64:
		if math.IsNaN(val) {
			return numberValue{typ: numberTypeNaN}, nil
		}
		if math.IsInf(val, 1) {
			return numberValue{typ: numberTypeInfinity, value: 1}, nil
		}
		if math.IsInf(val, -1) {
			return numberValue{typ: numberTypeNegativeInfinity, value: -1}, nil
		}
		if val == float64(int64(val)) {
			return numberValue{typ: numberTypeInteger, value: val}, nil
		}
		return numberValue{typ: numberTypeFloat, value: val}, nil
	case json.Number:
		if f, err := val.Float64(); err == nil {
			return convertToNumber(f)
		}
		if i, err := val.Int64(); err == nil {
			return numberValue{typ: numberTypeInteger, value: float64(i)}, nil
		}
		return numberValue{}, fmt.Errorf("invalid number: %v", val)
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return convertToNumber(f)
		}
		return numberValue{}, fmt.Errorf("invalid number string: %v", val)
	default:
		return numberValue{}, fmt.Errorf("cannot convert to number: %v", v)
	}
}

// compareNumberValues 比较两个数值
func compareNumberValues(a, b numberValue) int {
	// 处理特殊值
	if a.typ == numberTypeNaN || b.typ == numberTypeNaN {
		return 0 // NaN 等于 NaN，不等于其他任何值
	}
	if a.typ == numberTypeInfinity {
		if b.typ == numberTypeInfinity {
			return 0
		}
		return 1
	}
	if a.typ == numberTypeNegativeInfinity {
		if b.typ == numberTypeNegativeInfinity {
			return 0
		}
		return -1
	}
	if b.typ == numberTypeInfinity {
		return -1
	}
	if b.typ == numberTypeNegativeInfinity {
		return 1
	}

	// 处理普通数值
	diff := a.value - b.value
	if math.Abs(diff) < 1e-10 { // 使用精度阈值处理浮点数比较
		return 0
	}
	if diff > 0 {
		return 1
	}
	return -1
}

// formatNumber 格式化数值输出
func formatNumber(n numberValue) string {
	switch n.typ {
	case numberTypeNaN:
		return "NaN"
	case numberTypeInfinity:
		return "Infinity"
	case numberTypeNegativeInfinity:
		return "-Infinity"
	case numberTypeInteger:
		// 对于大整数，使用 strconv.FormatInt 可能会导致精度丢失
		// 所以我们需要特殊处理
		if n.value >= float64(math.MinInt64) && n.value <= float64(math.MaxInt64) {
			// 对于 MaxInt64，需要特殊处理，因为直接转换可能会导致溢出
			if n.value == float64(math.MaxInt64) {
				return "9223372036854775807"
			}
			return strconv.FormatInt(int64(n.value), 10)
		}
		// 对于超出 int64 范围的数，使用科学计数法
		return strconv.FormatFloat(n.value, 'g', -1, 64)
	default:
		// 对于浮点数，我们需要根据数值大小选择合适的格式化方式
		abs := math.Abs(n.value)
		if abs != 0 && abs < 0.0001 {
			// 使用科学计数法，但去掉前导零
			s := strconv.FormatFloat(n.value, 'e', -1, 64)
			if strings.Contains(s, "e-0") {
				s = strings.Replace(s, "e-0", "e-", 1)
			}
			return s
		}
		if abs >= 1e6 {
			// 对于大数，使用普通表示法
			return strconv.FormatFloat(n.value, 'f', 2, 64)
		}
		// 使用普通表示法，去掉尾随零
		s := strconv.FormatFloat(n.value, 'f', -1, 64)
		if strings.Contains(s, ".") {
			s = strings.TrimRight(s, "0")
			s = strings.TrimRight(s, ".")
		}
		return s
	}
}

// isNaN32 检查 float32 是否为 NaN
func isNaN32(f float32) bool {
	return f != f
}

// isInf32 检查 float32 是否为 Infinity
func isInf32(f float32, sign int) bool {
	return math.IsInf(float64(f), sign)
}

// globalFunctions is the registry of built-in functions
var globalFunctions = map[string]Function{
	"length": &builtinFunction{
		name: "length",
		callback: func(args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("length() requires exactly 1 argument")
			}

			// 如果参数是数组，返回数组长度
			if arr, ok := args[0].([]interface{}); ok {
				return float64(len(arr)), nil
			}

			// 如果参数是字符串，返回字符串长度
			if str, ok := args[0].(string); ok {
				return float64(len(str)), nil
			}

			// 如果参数是对象，返回对象的键数量
			if obj, ok := args[0].(map[string]interface{}); ok {
				return float64(len(obj)), nil
			}

			return nil, fmt.Errorf("length() argument must be string, array, or object")
		},
	},
	"keys": &builtinFunction{
		name: "keys",
		callback: func(args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("keys() requires exactly 1 argument")
			}

			// 确保参数是对象
			obj, ok := args[0].(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("keys() argument must be an object")
			}

			// 获取所有键并排序
			keys := make([]string, 0, len(obj))
			for k := range obj {
				keys = append(keys, k)
			}
			sort.Strings(keys) // 按字母顺序排序

			// 转换为 interface{} 切片
			result := make([]interface{}, len(keys))
			for i, k := range keys {
				result[i] = k
			}

			return result, nil
		},
	},
	"values": &builtinFunction{
		name: "values",
		callback: func(args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("values() requires exactly 1 argument")
			}

			// 确保参数是对象
			obj, ok := args[0].(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("values() argument must be an object")
			}

			// 获取所有键并排序，以确保值的顺序一致
			keys := make([]string, 0, len(obj))
			for k := range obj {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			// 按键的顺序获取值
			values := make([]interface{}, len(keys))
			for i, k := range keys {
				values[i] = obj[k]
			}

			return values, nil
		},
	},
	"count": &builtinFunction{
		name: "count",
		callback: func(args []interface{}) (interface{}, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("count() requires exactly 2 arguments: array and value")
			}

			// 确保第一个参数是数组
			arr, ok := args[0].([]interface{})
			if !ok {
				return nil, fmt.Errorf("count() first argument must be an array")
			}

			// 计算匹配值的数量
			count := 0
			for _, item := range arr {
				if reflect.DeepEqual(item, args[1]) {
					count++
				}
			}

			return float64(count), nil
		},
	},
	"min": &builtinFunction{
		name: "min",
		callback: func(args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("min() requires exactly 1 argument")
			}

			// 确保参数是数组
			arr, ok := args[0].([]interface{})
			if !ok {
				return nil, fmt.Errorf("min() argument must be an array")
			}

			if len(arr) == 0 {
				return nil, fmt.Errorf("min() cannot be applied to an empty array")
			}

			var minVal *numberValue
			for _, item := range arr {
				num, err := convertToNumber(item)
				if err != nil {
					continue // 跳过无效的数值
				}

				// 跳过 NaN
				if num.typ == numberTypeNaN {
					continue
				}

				if minVal == nil {
					minVal = &num
					continue
				}

				if compareNumberValues(num, *minVal) < 0 {
					minVal = &num
				}
			}

			if minVal == nil {
				return nil, fmt.Errorf("min() no valid numbers in array")
			}

			// 返回原始类型的值
			if minVal.typ == numberTypeInteger {
				return int64(minVal.value), nil
			}
			return minVal.value, nil
		},
	},
	"max": &builtinFunction{
		name: "max",
		callback: func(args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("max() requires exactly 1 argument")
			}

			// 确保参数是数组
			arr, ok := args[0].([]interface{})
			if !ok {
				return nil, fmt.Errorf("max() argument must be an array")
			}

			if len(arr) == 0 {
				return nil, fmt.Errorf("max() cannot be applied to an empty array")
			}

			var maxVal *numberValue
			for _, item := range arr {
				num, err := convertToNumber(item)
				if err != nil {
					continue // 跳过无效的数值
				}

				// 跳过 NaN
				if num.typ == numberTypeNaN {
					continue
				}

				if maxVal == nil {
					maxVal = &num
					continue
				}

				if compareNumberValues(num, *maxVal) > 0 {
					maxVal = &num
				}
			}

			if maxVal == nil {
				return nil, fmt.Errorf("max() no valid numbers in array")
			}

			// 返回原始类型的值
			if maxVal.typ == numberTypeInteger {
				return int64(maxVal.value), nil
			}
			return maxVal.value, nil
		},
	},
	"avg": &builtinFunction{
		name: "avg",
		callback: func(args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("avg() requires exactly 1 argument")
			}

			// 确保参数是数组
			arr, ok := args[0].([]interface{})
			if !ok {
				return nil, fmt.Errorf("avg() argument must be an array")
			}

			if len(arr) == 0 {
				return nil, fmt.Errorf("avg() cannot be applied to an empty array")
			}

			var sum float64
			count := 0

			for _, item := range arr {
				num, err := convertToNumber(item)
				if err != nil {
					continue // 跳过无效的数值
				}

				// 跳过特殊值
				if num.typ == numberTypeNaN ||
					num.typ == numberTypeInfinity ||
					num.typ == numberTypeNegativeInfinity {
					continue
				}

				sum += num.value
				count++
			}

			if count == 0 {
				return nil, fmt.Errorf("avg() no valid numbers in array")
			}

			result := sum / float64(count)
			if result == float64(int64(result)) {
				return int64(result), nil
			}
			return result, nil
		},
	},
	"sum": &builtinFunction{
		name: "sum",
		callback: func(args []interface{}) (interface{}, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("sum() requires exactly 1 argument")
			}

			// 确保参数是数组
			arr, ok := args[0].([]interface{})
			if !ok {
				return nil, fmt.Errorf("sum() argument must be an array")
			}

			if len(arr) == 0 {
				return nil, fmt.Errorf("sum() cannot be applied to an empty array")
			}

			var sum float64
			count := 0
			allIntegers := true

			for _, item := range arr {
				num, err := convertToNumber(item)
				if err != nil {
					continue // 跳过无效的数值
				}

				// 跳过特殊值
				if num.typ == numberTypeNaN ||
					num.typ == numberTypeInfinity ||
					num.typ == numberTypeNegativeInfinity {
					continue
				}

				if num.typ == numberTypeFloat {
					allIntegers = false
				}

				sum += num.value
				count++
			}

			if count == 0 {
				return nil, fmt.Errorf("sum() no valid numbers in array")
			}

			// 如果所有数都是整数且结果也是整数，返回整数类型
			if allIntegers && sum == float64(int64(sum)) {
				return int64(sum), nil
			}
			return sum, nil
		},
	},
	"match": &builtinFunction{
		name: "match",
		callback: func(args []interface{}) (interface{}, error) {
			// 1. 验证参数数量
			if len(args) != 2 {
				return nil, fmt.Errorf("match() requires exactly 2 arguments: string and pattern")
			}

			// 2. 获取并验证第二个参数（正则表达式模式）
			pattern, ok := args[1].(string)
			if !ok {
				return nil, fmt.Errorf("match() second argument must be a string pattern")
			}

			// 3. 处理空模式
			if pattern == "" {
				return false, nil
			}

			// 4. 获取第一个参数（要匹配的字符串）
			var str string
			switch v := args[0].(type) {
			case string:
				str = v
			default:
				// 对于非字符串值，返回 false
				return false, nil
			}

			// 5. 处理正则表达式模式
			var result strings.Builder
			var escaped bool
			var inCharClass bool

			for i := 0; i < len(pattern); i++ {
				ch := pattern[i]
				if escaped {
					switch ch {
					case 'd', 'D', 'w', 'W', 's', 'S', 'b', 'B':
						// 保持原样的特殊字符序列
						result.WriteByte('\\')
						result.WriteByte(ch)
					case 'n':
						result.WriteString(`\n`)
					case 'r':
						result.WriteString(`\r`)
					case 't':
						result.WriteString(`\t`)
					case '[', ']', '(', ')', '{', '}', '\\', '.', '*', '+', '?', '|', '^', '$':
						// 转义元字符
						result.WriteByte(ch)
					case 'p', 'P':
						// 处理 Unicode 属性
						result.WriteByte('\\')
						result.WriteByte(ch)
						if i+1 < len(pattern) && pattern[i+1] == '{' {
							i++ // 跳过 '{'
							result.WriteByte('{')
							for i+1 < len(pattern) && pattern[i+1] != '}' {
								i++
								result.WriteByte(pattern[i])
							}
							if i+1 < len(pattern) && pattern[i+1] == '}' {
								i++
								result.WriteByte('}')
							}
						}
					default:
						if inCharClass {
							result.WriteByte(ch)
						} else {
							result.WriteByte(ch)
						}
					}
					escaped = false
				} else if ch == '\\' {
					escaped = true
				} else if ch == '[' {
					inCharClass = true
					result.WriteByte(ch)
				} else if ch == ']' {
					inCharClass = false
					result.WriteByte(ch)
				} else {
					result.WriteByte(ch)
				}
			}

			// 处理末尾的反斜杠
			if escaped {
				result.WriteByte('\\')
			}

			pattern = result.String()

			// 6. 获取或编译正则表达式
			re, err := getCompiledRegex(pattern)
			if err != nil {
				return false, nil // 正则表达式语法错误时返回 false
			}

			// 7. 执行匹配
			return re.MatchString(str), nil
		},
	},
	"search": &builtinFunction{
		name: "search",
		callback: func(args []interface{}) (interface{}, error) {
			// 1. 验证参数数量
			if len(args) != 2 {
				return nil, fmt.Errorf("search() requires exactly 2 arguments")
			}

			// 2. 获取数组参数
			arr, ok := args[0].([]interface{})
			if !ok {
				return nil, fmt.Errorf("first argument must be an array")
			}

			// 3. 获取正则表达式参数
			pattern, ok := args[1].(string)
			if !ok {
				return nil, fmt.Errorf("second argument must be a string pattern")
			}

			// 4. 处理转义字符
			var result strings.Builder
			var escaped bool
			var inCharClass bool

			for i := 0; i < len(pattern); i++ {
				ch := pattern[i]
				if escaped {
					switch ch {
					case 'd', 'D', 'w', 'W', 's', 'S', 'b', 'B':
						// 保持原样的特殊字符序列
						result.WriteByte('\\')
						result.WriteByte(ch)
					case 'n':
						result.WriteString(`\n`)
					case 'r':
						result.WriteString(`\r`)
					case 't':
						result.WriteString(`\t`)
					case '[', ']', '(', ')', '{', '}', '\\', '.', '*', '+', '?', '|', '^', '$':
						// 转义元字符
						result.WriteByte(ch)
					case 'p', 'P':
						// 处理 Unicode 属性
						result.WriteByte('\\')
						result.WriteByte(ch)
						if i+1 < len(pattern) && pattern[i+1] == '{' {
							i++ // 跳过 '{'
							result.WriteByte('{')
							for i+1 < len(pattern) && pattern[i+1] != '}' {
								i++
								result.WriteByte(pattern[i])
							}
							if i+1 < len(pattern) && pattern[i+1] == '}' {
								i++
								result.WriteByte('}')
							}
						}
					default:
						if inCharClass {
							result.WriteByte(ch)
						} else {
							result.WriteByte(ch)
						}
					}
					escaped = false
				} else if ch == '\\' {
					escaped = true
				} else if ch == '[' {
					inCharClass = true
					result.WriteByte(ch)
				} else if ch == ']' {
					inCharClass = false
					result.WriteByte(ch)
				} else {
					result.WriteByte(ch)
				}
			}

			// 处理末尾的反斜杠
			if escaped {
				result.WriteByte('\\')
			}

			pattern = result.String()

			// 5. 获取或编译正则表达式
			re, err := getCompiledRegex(pattern)
			if err != nil {
				return nil, fmt.Errorf("invalid regular expression: %v", err)
			}

			// 6. 搜索匹配的元素
			matches := make([]interface{}, 0)
			for _, item := range arr {
				var str string
				switch v := item.(type) {
				case string:
					str = v
				case float64:
					str = strconv.FormatFloat(v, 'f', -1, 64)
				case int:
					str = strconv.Itoa(v)
				case bool:
					str = strconv.FormatBool(v)
				case nil:
					str = "null"
				default:
					// 尝试将其他类型转换为 JSON 字符串
					if jsonBytes, err := json.Marshal(v); err == nil {
						str = string(jsonBytes)
					} else {
						continue // 跳过无法转换的值
					}
				}

				if re.MatchString(str) {
					matches = append(matches, item) // 保持原始值类型
				}
			}

			return matches, nil
		},
	},
}

// GetFunction returns a registered function by name
func GetFunction(name string) (Function, error) {
	if f, exists := globalFunctions[name]; exists {
		return f, nil
	}
	return nil, fmt.Errorf("function %s not found", name)
}
