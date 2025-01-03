package jsonpath

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseRecursive(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		wantLen     int
		wantErr     bool
		errType     ErrorType
		errContains string
	}{
		{
			name:    "empty path",
			path:    "",
			wantLen: 1, // 只包含递归段
		},
		{
			name:    "simple recursive",
			path:    "name",
			wantLen: 2, // 递归段 + 名称段
		},
		{
			name:    "recursive with dot notation",
			path:    ".name",
			wantLen: 2,
		},
		{
			name:    "recursive with nested path",
			path:    "books.title",
			wantLen: 3, // 递归段 + books段 + title段
		},
		{
			name:    "recursive with bracket notation",
			path:    "[0].name",
			wantLen: 3, // 递归段 + 索引段 + 名称段
		},
		{
			name:    "recursive with filter",
			path:    "[?(@.price > 10)]",
			wantLen: 2, // 递归段 + 过滤器段
		},
		{
			name:        "invalid filter syntax",
			path:        "[?(@.price >=)]",
			wantErr:     true,
			errType:     ErrInvalidFilter,
			errContains: "invalid value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRecursive(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRecursive() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				if jsonErr, ok := err.(*Error); ok {
					if tt.errType != jsonErr.Type {
						t.Errorf("parseRecursive() error type = %v, want %v", jsonErr.Type, tt.errType)
					}
					if tt.errContains != "" && !strings.Contains(jsonErr.Message, tt.errContains) {
						t.Errorf("parseRecursive() error = %v, want error containing %v", jsonErr.Message, tt.errContains)
					}
				} else {
					t.Errorf("parseRecursive() error is not a JSONPath error: %v", err)
				}
				return
			}
			if len(got) != tt.wantLen {
				t.Errorf("parseRecursive() returned %d segments, want %d", len(got), tt.wantLen)
			}
			// 验证第一个段是否为递归段
			if len(got) > 0 {
				if _, ok := got[0].(*recursiveSegment); !ok {
					t.Error("First segment is not a recursiveSegment")
				}
			}
		})
	}
}

func TestRecursiveSegmentString(t *testing.T) {
	seg := &recursiveSegment{}
	expected := ".."
	if got := seg.String(); got != expected {
		t.Errorf("recursiveSegment.String() = %v, want %v", got, expected)
	}
}

func TestParseMultiIndexSegment(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    *multiIndexSegment
		wantErr bool
	}{
		{
			name:    "single index",
			content: "1",
			want:    &multiIndexSegment{indices: []int{1}},
			wantErr: false,
		},
		{
			name:    "multiple indices",
			content: "1,2,3",
			want:    &multiIndexSegment{indices: []int{1, 2, 3}},
			wantErr: false,
		},
		{
			name:    "negative indices",
			content: "-1,-2,-3",
			want:    &multiIndexSegment{indices: []int{-1, -2, -3}},
			wantErr: false,
		},
		{
			name:    "mixed indices",
			content: "0,1,-1,2,-2",
			want:    &multiIndexSegment{indices: []int{0, 1, -1, 2, -2}},
			wantErr: false,
		},
		{
			name:    "with spaces",
			content: "1, 2, 3",
			want:    &multiIndexSegment{indices: []int{1, 2, 3}},
			wantErr: false,
		},
		{
			name:    "invalid index",
			content: "1,a,3",
			wantErr: true,
		},
		{
			name:    "empty index",
			content: "1,,3",
			wantErr: true,
		},
		{
			name:    "trailing comma",
			content: "1,2,",
			wantErr: true,
		},
		{
			name:    "leading comma",
			content: ",1,2",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseMultiIndexSegment(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseMultiIndexSegment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("parseMultiIndexSegment() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestMultiIndexSegmentString(t *testing.T) {
	seg := &multiIndexSegment{indices: []int{1, 2, 3}}
	expected := "[1,2,3]"
	if got := seg.String(); got != expected {
		t.Errorf("multiIndexSegment.String() = %v, want %v", got, expected)
	}
}

func TestParseSliceSegment(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    *sliceSegment
		wantErr bool
	}{
		{
			name:    "empty slice",
			content: ":",
			want:    &sliceSegment{start: 0, end: 0, step: 1},
			wantErr: false,
		},
		{
			name:    "start only",
			content: "1:",
			want:    &sliceSegment{start: 1, end: 0, step: 1},
			wantErr: false,
		},
		{
			name:    "end only",
			content: ":2",
			want:    &sliceSegment{start: 0, end: 2, step: 1},
			wantErr: false,
		},
		{
			name:    "start and end",
			content: "1:2",
			want:    &sliceSegment{start: 1, end: 2, step: 1},
			wantErr: false,
		},
		{
			name:    "negative indices",
			content: "-2:-1",
			want:    &sliceSegment{start: -2, end: -1, step: 1},
			wantErr: false,
		},
		{
			name:    "with step",
			content: "1:5:2",
			want:    &sliceSegment{start: 1, end: 5, step: 2},
			wantErr: false,
		},
		{
			name:    "negative step",
			content: "5:1:-1",
			want:    &sliceSegment{start: 5, end: 1, step: -1},
			wantErr: false,
		},
		{
			name:    "empty start with step",
			content: ":5:2",
			want:    &sliceSegment{start: 0, end: 5, step: 2},
			wantErr: false,
		},
		{
			name:    "empty end with step",
			content: "1::2",
			want:    &sliceSegment{start: 1, end: 0, step: 2},
			wantErr: false,
		},
		{
			name:    "invalid start",
			content: "a:2",
			wantErr: true,
		},
		{
			name:    "invalid end",
			content: "1:b",
			wantErr: true,
		},
		{
			name:    "invalid step",
			content: "1:2:c",
			wantErr: true,
		},
		{
			name:    "zero step",
			content: "1:2:0",
			wantErr: true,
		},
		{
			name:    "too many parts",
			content: "1:2:3:4",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSliceSegment(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSliceSegment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("parseSliceSegment() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestSliceSegmentString(t *testing.T) {
	tests := []struct {
		name     string
		segment  *sliceSegment
		expected string
	}{
		{
			name:     "full slice",
			segment:  &sliceSegment{start: 1, end: 5, step: 2},
			expected: "[1:5:2]",
		},
		{
			name:     "without step",
			segment:  &sliceSegment{start: 1, end: 5, step: 1},
			expected: "[1:5]",
		},
		{
			name:     "negative indices",
			segment:  &sliceSegment{start: -2, end: -1, step: 1},
			expected: "[-2:-1]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.segment.String(); got != tt.expected {
				t.Errorf("sliceSegment.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseFunctionCall(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantName    string
		wantArgs    []interface{}
		wantErr     bool
		errContains string
	}{
		// 无参数函数
		{
			name:     "no args",
			content:  "length()",
			wantName: "length",
			wantArgs: []interface{}{},
		},
		{
			name:     "empty args",
			content:  "count(  )",
			wantName: "count",
			wantArgs: []interface{}{},
		},

		// 数字参数
		{
			name:     "single number arg",
			content:  "min(42)",
			wantName: "min",
			wantArgs: []interface{}{float64(42)},
		},
		{
			name:     "multiple number args",
			content:  "sum(1, 2, 3)",
			wantName: "sum",
			wantArgs: []interface{}{float64(1), float64(2), float64(3)},
		},
		{
			name:     "decimal number args",
			content:  "avg(1.5, 2.7, 3.9)",
			wantName: "avg",
			wantArgs: []interface{}{float64(1.5), float64(2.7), float64(3.9)},
		},
		{
			name:     "negative number args",
			content:  "range(-1, -2, -3)",
			wantName: "range",
			wantArgs: []interface{}{float64(-1), float64(-2), float64(-3)},
		},

		// 字符串参数
		{
			name:     "single string arg",
			content:  "match('pattern')",
			wantName: "match",
			wantArgs: []interface{}{"pattern"},
		},
		{
			name:     "multiple string args",
			content:  "concat('hello', 'world')",
			wantName: "concat",
			wantArgs: []interface{}{"hello", "world"},
		},
		{
			name:     "string args with spaces",
			content:  "format('hello world', 'goodbye world')",
			wantName: "format",
			wantArgs: []interface{}{"hello world", "goodbye world"},
		},
		{
			name:     "empty string args",
			content:  "join('', '')",
			wantName: "join",
			wantArgs: []interface{}{"", ""},
		},

		// 混合参数
		{
			name:     "mixed string and number args",
			content:  "format('value', 42)",
			wantName: "format",
			wantArgs: []interface{}{"value", float64(42)},
		},
		{
			name:     "complex mixed args",
			content:  "transform('data', 1.5, 'options', -3)",
			wantName: "transform",
			wantArgs: []interface{}{"data", float64(1.5), "options", float64(-3)},
		},

		// 错误情况
		{
			name:        "missing opening parenthesis",
			content:     "length",
			wantErr:     true,
			errContains: "missing opening parenthesis",
		},
		{
			name:        "missing closing parenthesis",
			content:     "length(",
			wantErr:     true,
			errContains: "missing closing parenthesis",
		},
		{
			name:        "invalid argument type",
			content:     "func(invalid)",
			wantErr:     true,
			errContains: "unsupported argument type",
		},
		{
			name:        "unclosed string argument",
			content:     "func('unclosed)",
			wantErr:     true,
			errContains: "unsupported argument type",
		},
		{
			name:        "unmatched quotes",
			content:     "func('test\", 42)",
			wantErr:     true,
			errContains: "unsupported argument type",
		},

		// 边界情况
		{
			name:     "function name with underscore",
			content:  "my_func()",
			wantName: "my_func",
			wantArgs: []interface{}{},
		},
		{
			name:     "function name with numbers",
			content:  "func123()",
			wantName: "func123",
			wantArgs: []interface{}{},
		},
		{
			name:     "very large number",
			content:  "big(9999999999)",
			wantName: "big",
			wantArgs: []interface{}{float64(9999999999)},
		},
		{
			name:     "very small number",
			content:  "small(-9999999999)",
			wantName: "small",
			wantArgs: []interface{}{float64(-9999999999)},
		},
		{
			name:     "string with escaped quotes",
			content:  "escape('I''m here')",
			wantName: "escape",
			wantArgs: []interface{}{"I'm here"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFunctionCall(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseFunctionCall() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("parseFunctionCall() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}

			fs, ok := got.(*functionSegment)
			if !ok {
				t.Errorf("parseFunctionCall() returned %T, want *functionSegment", got)
				return
			}

			if fs.name != tt.wantName {
				t.Errorf("parseFunctionCall() name = %v, want %v", fs.name, tt.wantName)
			}

			if !reflect.DeepEqual(fs.args, tt.wantArgs) {
				t.Errorf("parseFunctionCall() args = %v, want %v", fs.args, tt.wantArgs)
			}
		})
	}
}

func TestFunctionSegmentString(t *testing.T) {
	tests := []struct {
		name     string
		segment  *functionSegment
		expected string
	}{
		{
			name:     "no args",
			segment:  &functionSegment{name: "length", args: []interface{}{}},
			expected: "length()",
		},
		{
			name:     "single number arg",
			segment:  &functionSegment{name: "min", args: []interface{}{float64(1)}},
			expected: "min(1)",
		},
		{
			name:     "multiple args",
			segment:  &functionSegment{name: "sum", args: []interface{}{float64(1), float64(2), float64(3)}},
			expected: "sum(1,2,3)",
		},
		{
			name:     "string arg",
			segment:  &functionSegment{name: "match", args: []interface{}{"pattern"}},
			expected: "match('pattern')",
		},
		{
			name:     "mixed args",
			segment:  &functionSegment{name: "format", args: []interface{}{"value", float64(42)}},
			expected: "format('value',42)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.segment.String(); got != tt.expected {
				t.Errorf("functionSegment.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseIndexOrName(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantType    string
		wantValue   interface{}
		wantErr     bool
		errContains string
	}{
		// 数字索引
		{
			name:      "positive index",
			content:   "42",
			wantType:  "indexSegment",
			wantValue: 42,
		},
		{
			name:      "negative index",
			content:   "-1",
			wantType:  "indexSegment",
			wantValue: -1,
		},
		{
			name:      "zero index",
			content:   "0",
			wantType:  "indexSegment",
			wantValue: 0,
		},
		{
			name:      "max int",
			content:   "9223372036854775807",
			wantType:  "indexSegment",
			wantValue: 9223372036854775807,
		},
		{
			name:      "min int",
			content:   "-9223372036854775808",
			wantType:  "indexSegment",
			wantValue: -9223372036854775808,
		},

		// 字符串字面量
		{
			name:      "quoted string",
			content:   "'hello'",
			wantType:  "nameSegment",
			wantValue: "hello",
		},
		{
			name:      "empty quoted string",
			content:   "''",
			wantType:  "nameSegment",
			wantValue: "",
		},
		{
			name:      "quoted string with spaces",
			content:   "'hello world'",
			wantType:  "nameSegment",
			wantValue: "hello world",
		},
		{
			name:      "quoted string with special characters",
			content:   "'hello.world'",
			wantType:  "nameSegment",
			wantValue: "hello.world",
		},
		{
			name:      "quoted string with numbers",
			content:   "'123'",
			wantType:  "nameSegment",
			wantValue: "123",
		},
		{
			name:      "quoted string with unicode",
			content:   "'你好'",
			wantType:  "nameSegment",
			wantValue: "你好",
		},

		// 普通名称
		{
			name:      "unquoted string",
			content:   "hello",
			wantType:  "nameSegment",
			wantValue: "hello",
		},
		{
			name:      "unquoted string with underscore",
			content:   "hello_world",
			wantType:  "nameSegment",
			wantValue: "hello_world",
		},
		{
			name:      "unquoted string with numbers",
			content:   "hello123",
			wantType:  "nameSegment",
			wantValue: "hello123",
		},
		{
			name:      "unquoted string starting with underscore",
			content:   "_hello",
			wantType:  "nameSegment",
			wantValue: "_hello",
		},

		// 函数调用
		{
			name:      "function call without args",
			content:   "length()",
			wantType:  "functionSegment",
			wantValue: "length",
		},
		{
			name:      "function call with number arg",
			content:   "min(1)",
			wantType:  "functionSegment",
			wantValue: "min",
		},
		{
			name:      "function call with string arg",
			content:   "match('pattern')",
			wantType:  "functionSegment",
			wantValue: "match",
		},
		{
			name:      "function call with multiple args",
			content:   "format('value', 42)",
			wantType:  "functionSegment",
			wantValue: "format",
		},

		// 边界情况和错误
		{
			name:      "single quote",
			content:   "'",
			wantType:  "nameSegment",
			wantValue: "'",
		},
		{
			name:      "unclosed quote",
			content:   "'hello",
			wantType:  "nameSegment",
			wantValue: "'hello",
		},
		{
			name:      "empty string",
			content:   "",
			wantType:  "nameSegment",
			wantValue: "",
		},
		{
			name:      "whitespace only",
			content:   "   ",
			wantType:  "nameSegment",
			wantValue: "   ",
		},
		{
			name:      "special characters",
			content:   "@#$%",
			wantType:  "nameSegment",
			wantValue: "@#$%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseIndexOrName(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseIndexOrName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("parseIndexOrName() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}

			// 检查段类型
			gotType := reflect.TypeOf(got).String()
			if !strings.HasSuffix(gotType, tt.wantType) {
				t.Errorf("parseIndexOrName() returned %v, want type %v", gotType, tt.wantType)
				return
			}

			// 检查段值
			switch v := got.(type) {
			case *indexSegment:
				if v.index != tt.wantValue.(int) {
					t.Errorf("parseIndexOrName() index = %v, want %v", v.index, tt.wantValue)
				}
			case *nameSegment:
				if v.name != tt.wantValue.(string) {
					t.Errorf("parseIndexOrName() name = %v, want %v", v.name, tt.wantValue)
				}
			case *functionSegment:
				if v.name != tt.wantValue.(string) {
					t.Errorf("parseIndexOrName() function name = %v, want %v", v.name, tt.wantValue)
				}
			default:
				t.Errorf("parseIndexOrName() returned unexpected type %T", got)
			}
		})
	}
}

func TestParseFilterValue(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    interface{}
		wantErr bool
	}{
		// 测试用例
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFilterValue(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseFilterValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseFilterValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
