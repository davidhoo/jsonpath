# Phase 3: v3.0.0 类型系统基础实施计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 引入RFC 9535核心类型，不影响现有代码（纯新增）

**Architecture:** 定义Node、NodeList、Nothing、LogicalType核心类型，实现Normalized Path生成器和I-Regexp解析器

**Tech Stack:** Go, JSON serialization, Unicode, Regular expressions

---

## Task 1: 定义核心类型

**Files:**
- Create: `types_v3.go`

**Step 1: 创建核心类型定义**

创建 `types_v3.go`：

```go
package jsonpath

import "encoding/json"

// Node 表示RFC 9535节点（location + value）
type Node struct {
	Location string      `json:"location"`
	Value    interface{} `json:"value"`
}

// NodeList 表示节点列表
type NodeList []Node

// MarshalJSON 实现json.Marshaler接口
func (nl NodeList) MarshalJSON() ([]byte, error) {
	return json.Marshal([]Node(nl))
}

// Nothing 表示Nothing值（与null不同）
type Nothing struct{}

// String 返回Nothing的字符串表示
func (n Nothing) String() string {
	return "Nothing"
}

// LogicalType 表示三值逻辑
type LogicalType int8

const (
	LogicalNothing LogicalType = iota
	LogicalFalse
	LogicalTrue
)

// String 返回LogicalType的字符串表示
func (lt LogicalType) String() string {
	switch lt {
	case LogicalNothing:
		return "nothing"
	case LogicalFalse:
		return "false"
	case LogicalTrue:
		return "true"
	default:
		return "unknown"
	}
}
```

**Step 2: 添加类型测试**

创建 `types_v3_test.go`：

```go
package jsonpath

import (
	"encoding/json"
	"testing"
)

func TestNodeCreation(t *testing.T) {
	node := Node{
		Location: "$['name']",
		Value:    "test",
	}
	
	if node.Location != "$['name']" {
		t.Errorf("Expected location $['name'], got %s", node.Location)
	}
	
	if node.Value != "test" {
		t.Errorf("Expected value test, got %v", node.Value)
	}
}

func TestNodeListJSON(t *testing.T) {
	nodeList := NodeList{
		{Location: "$[0]", Value: 1},
		{Location: "$[1]", Value: 2},
	}
	
	jsonData, err := json.Marshal(nodeList)
	if err != nil {
		t.Fatalf("Failed to marshal NodeList: %v", err)
	}
	
	expected := `[{"location":"$[0]","value":1},{"location":"$[1]","value":2}]`
	if string(jsonData) != expected {
		t.Errorf("Expected %s, got %s", expected, string(jsonData))
	}
}

func TestLogicalTypeString(t *testing.T) {
	tests := []struct {
		lt       LogicalType
		expected string
	}{
		{LogicalNothing, "nothing"},
		{LogicalFalse, "false"},
		{LogicalTrue, "true"},
	}
	
	for _, tt := range tests {
		if tt.lt.String() != tt.expected {
			t.Errorf("Expected %s, got %s", tt.expected, tt.lt.String())
		}
	}
}

func TestNothingString(t *testing.T) {
	n := Nothing{}
	if n.String() != "Nothing" {
		t.Errorf("Expected Nothing, got %s", n.String())
	}
}
```

**Step 3: 运行测试确认通过**

```bash
go test -run TestNode -v
go test -run TestLogicalType -v
go test -run TestNothing -v
```

Expected: PASS

**Step 4: 运行全量测试确认无回归**

```bash
go test ./...
```

Expected: 所有测试通过

**Step 5: Commit**

```bash
git add types_v3.go types_v3_test.go
git commit -m "feat: add Node, NodeList, Nothing, LogicalType types"
```

---

## Task 2: 实现Normalized Path生成器

**Files:**
- Create: `normalized_path.go`
- Create: `normalized_path_test.go`

**Step 1: 创建Normalized Path生成器**

创建 `normalized_path.go`：

```go
package jsonpath

import (
	"fmt"
	"strings"
	"unicode"
)

// NormalizedPathGenerator 生成RFC 9535 Normalized Path
type NormalizedPathGenerator struct {
	segments []string
}

// NewNormalizedPathGenerator 创建新的Normalized Path生成器
func NewNormalizedPathGenerator() *NormalizedPathGenerator {
	return &NormalizedPathGenerator{
		segments: []string{"$"},
	}
}

// AddMember 添加对象成员段
func (npg *NormalizedPathGenerator) AddMember(name string) *NormalizedPathGenerator {
	escaped := escapeMemberName(name)
	npg.segments = append(npg.segments, fmt.Sprintf("['%s']", escaped))
	return npg
}

// AddIndex 添加数组索引段
func (npg *NormalizedPathGenerator) AddIndex(index int) *NormalizedPathGenerator {
	npg.segments = append(npg.segments, fmt.Sprintf("[%d]", index))
	return npg
}

// String 返回Normalized Path字符串
func (npg *NormalizedPathGenerator) String() string {
	return strings.Join(npg.segments, "")
}

// escapeMemberName 转义成员名中的特殊字符
func escapeMemberName(name string) string {
	var result strings.Builder
	
	for _, r := range name {
		switch {
		case r == '\'':
			result.WriteString("\\'")
		case r == '\\':
			result.WriteString("\\\\")
		case unicode.IsControl(r):
			result.WriteString(fmt.Sprintf("\\u%04x", r))
		default:
			result.WriteRune(r)
		}
	}
	
	return result.String()
}

// GenerateNormalizedPath 从路径段列表生成Normalized Path
func GenerateNormalizedPath(segments []interface{}) string {
	npg := NewNormalizedPathGenerator()
	
	for _, seg := range segments {
		switch s := seg.(type) {
		case string:
			npg.AddMember(s)
		case int:
			npg.AddIndex(s)
		}
	}
	
	return npg.String()
}
```

**Step 2: 添加Normalized Path测试**

创建 `normalized_path_test.go`：

```go
package jsonpath

import "testing"

func TestNormalizedPathGeneration(t *testing.T) {
	tests := []struct {
		name     string
		segments []interface{}
		expected string
	}{
		{
			name:     "Root",
			segments: []interface{}{},
			expected: "$",
		},
		{
			name:     "Simple member",
			segments: []interface{}{"store"},
			expected: "$['store']",
		},
		{
			name:     "Nested members",
			segments: []interface{}{"store", "book"},
			expected: "$['store']['book']",
		},
		{
			name:     "Array index",
			segments: []interface{}{"store", "book", 0},
			expected: "$['store']['book'][0]",
		},
		{
			name:     "Member with single quote",
			segments: []interface{}{"it's"},
			expected: "$['it\\'s']",
		},
		{
			name:     "Member with backslash",
			segments: []interface{}{"back\\slash"},
			expected: "$['back\\\\slash']",
		},
		{
			name:     "Empty member name",
			segments: []interface{}{""},
			expected: "$['']",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateNormalizedPath(tt.segments)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestNormalizedPathGenerator(t *testing.T) {
	npg := NewNormalizedPathGenerator()
	npg.AddMember("store").AddMember("book").AddIndex(0)
	
	expected := "$['store']['book'][0]"
	if npg.String() != expected {
		t.Errorf("Expected %s, got %s", expected, npg.String())
	}
}
```

**Step 3: 运行测试确认通过**

```bash
go test -run TestNormalizedPath -v
```

Expected: PASS

**Step 4: 运行全量测试确认无回归**

```bash
go test ./...
```

Expected: 所有测试通过

**Step 5: Commit**

```bash
git add normalized_path.go normalized_path_test.go
git commit -m "feat: implement Normalized Path generator"
```

---

## Task 3: 实现I-Regexp解析器

**Files:**
- Create: `iregexp.go`
- Create: `iregexp_test.go`

**Step 1: 创建I-Regexp解析器**

创建 `iregexp.go`：

```go
package jsonpath

import (
	"fmt"
	"regexp"
	"strings"
)

// IRegexpParser 解析I-Regexp语法
type IRegexpParser struct {
	pattern string
	pos     int
}

// NewIRegexpParser 创建新的I-Regexp解析器
func NewIRegexpParser(pattern string) *IRegexpParser {
	return &IRegexpParser{
		pattern: pattern,
		pos:     0,
	}
}

// IsValidIRegexp 验证是否为合法的I-Regexp
func IsValidIRegexp(pattern string) bool {
	parser := NewIRegexpParser(pattern)
	_, err := parser.Parse()
	return err == nil
}

// IRegexpToGoRegexp 将I-Regexp转换为Go regexp
func IRegexpToGoRegexp(pattern string) (string, error) {
	parser := NewIRegexpParser(pattern)
	return parser.Parse()
}

// Parse 解析I-Regexp模式
func (p *IRegexpParser) Parse() (string, error) {
	result, err := p.parseAlternation()
	if err != nil {
		return "", err
	}
	
	if p.pos < len(p.pattern) {
		return "", fmt.Errorf("unexpected character at position %d: %c", p.pos, p.pattern[p.pos])
	}
	
	return result, nil
}

// parseAlternation 解析交替运算 |
func (p *IRegexpParser) parseAlternation() (string, error) {
	left, err := p.parseSequence()
	if err != nil {
		return "", err
	}
	
	for p.pos < len(p.pattern) && p.pattern[p.pos] == '|' {
		p.pos++ // skip '|'
		right, err := p.parseSequence()
		if err != nil {
			return "", err
		}
		left = left + "|" + right
	}
	
	return left, nil
}

// parseSequence 解析序列
func (p *IRegexpParser) parseSequence() (string, error) {
	var result strings.Builder
	
	for p.pos < len(p.pattern) && p.pattern[p.pos] != '|' && p.pattern[p.pos] != ')' {
		atom, err := p.parseAtom()
		if err != nil {
			return "", err
		}
		result.WriteString(atom)
	}
	
	return result.String(), nil
}

// parseAtom 解析原子
func (p *IRegexpParser) parseAtom() (string, error) {
	if p.pos >= len(p.pattern) {
		return "", fmt.Errorf("unexpected end of pattern")
	}
	
	ch := p.pattern[p.pos]
	
	switch {
	case ch == '(':
		return p.parseGroup()
	case ch == '[':
		return p.parseCharacterClass()
	case ch == '\\':
		return p.parseEscape()
	case ch == '.':
		p.pos++
		return ".", nil
	case ch == '^' || ch == '$':
		p.pos++
		return string(ch), nil
	case ch == '*' || ch == '+' || ch == '?':
		return "", fmt.Errorf("quantifier without operand at position %d", p.pos)
	default:
		p.pos++
		return string(ch), nil
	}
}

// parseGroup 解析分组
func (p *IRegexpParser) parseGroup() (string, error) {
	p.pos++ // skip '('
	
	// 检查是否为特殊分组
	if p.pos < len(p.pattern) && p.pattern[p.pos] == '?' {
		return "", fmt.Errorf("unsupported group type at position %d", p.pos)
	}
	
	inner, err := p.parseAlternation()
	if err != nil {
		return "", err
	}
	
	if p.pos >= len(p.pattern) || p.pattern[p.pos] != ')' {
		return "", fmt.Errorf("missing closing parenthesis")
	}
	p.pos++ // skip ')'
	
	// 检查量词
	quantifier, err := p.parseQuantifier()
	if err != nil {
		return "", err
	}
	
	return "(" + inner + ")" + quantifier, nil
}

// parseCharacterClass 解析字符类
func (p *IRegexpParser) parseCharacterClass() (string, error) {
	p.pos++ // skip '['
	
	var result strings.Builder
	result.WriteString("[")
	
	if p.pos < len(p.pattern) && p.pattern[p.pos] == '^' {
		result.WriteString("^")
		p.pos++
	}
	
	for p.pos < len(p.pattern) && p.pattern[p.pos] != ']' {
		if p.pattern[p.pos] == '\\' {
			escape, err := p.parseEscape()
			if err != nil {
				return "", err
			}
			result.WriteString(escape)
		} else {
			result.WriteByte(p.pattern[p.pos])
			p.pos++
		}
	}
	
	if p.pos >= len(p.pattern) {
		return "", fmt.Errorf("unclosed character class")
	}
	
	result.WriteString("]")
	p.pos++ // skip ']'
	
	// 检查量词
	quantifier, err := p.parseQuantifier()
	if err != nil {
		return "", err
	}
	
	return result.String() + quantifier, nil
}

// parseEscape 解析转义序列
func (p *IRegexpParser) parseEscape() (string, error) {
	p.pos++ // skip '\\'
	
	if p.pos >= len(p.pattern) {
		return "", fmt.Errorf("unexpected end of pattern after escape")
	}
	
	ch := p.pattern[p.pos]
	p.pos++
	
	switch ch {
	case 'd', 'D', 'w', 'W', 's', 'S':
		// 检查量词
		quantifier, err := p.parseQuantifier()
		if err != nil {
			return "", err
		}
		return "\\" + string(ch) + quantifier, nil
	case 'p':
		return p.parseUnicodeProperty()
	case 'P':
		return p.parseUnicodeProperty()
	default:
		// 普通转义字符
		quantifier, err := p.parseQuantifier()
		if err != nil {
			return "", err
		}
		return "\\" + string(ch) + quantifier, nil
	}
}

// parseUnicodeProperty 解析Unicode属性
func (p *IRegexpParser) parseUnicodeProperty() (string, error) {
	if p.pos >= len(p.pattern) || p.pattern[p.pos] != '{' {
		return "", fmt.Errorf("expected '{' after \\p")
	}
	p.pos++ // skip '{'
	
	start := p.pos
	for p.pos < len(p.pattern) && p.pattern[p.pos] != '}' {
		p.pos++
	}
	
	if p.pos >= len(p.pattern) {
		return "", fmt.Errorf("unclosed Unicode property")
	}
	
	property := p.pattern[start:p.pos]
	p.pos++ // skip '}'
	
	// 检查量词
	quantifier, err := p.parseQuantifier()
	if err != nil {
		return "", err
	}
	
	return "\\p{" + property + "}" + quantifier, nil
}

// parseQuantifier 解析量词
func (p *IRegexpParser) parseQuantifier() (string, error) {
	if p.pos >= len(p.pattern) {
		return "", nil
	}
	
	ch := p.pattern[p.pos]
	
	switch ch {
	case '*', '+', '?':
		p.pos++
		return string(ch), nil
	case '{':
		return p.parseRangeQuantifier()
	default:
		return "", nil
	}
}

// parseRangeQuantifier 解析范围量词 {n,m}
func (p *IRegexpParser) parseRangeQuantifier() (string, error) {
	p.pos++ // skip '{'
	
	start := p.pos
	for p.pos < len(p.pattern) && p.pattern[p.pos] != '}' {
		p.pos++
	}
	
	if p.pos >= len(p.pattern) {
		return "", fmt.Errorf("unclosed range quantifier")
	}
	
	quantifier := p.pattern[start:p.pos]
	p.pos++ // skip '}'
	
	return "{" + quantifier + "}", nil
}
```

**Step 2: 添加I-Regexp测试**

创建 `iregexp_test.go`：

```go
package jsonpath

import (
	"regexp"
	"testing"
)

func TestIRegexpValidation(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		valid   bool
	}{
		{"Simple pattern", "^S.*$", true},
		{"Character class", "[abc]", true},
		{"Quantifier", "a+", true},
		{"Alternation", "a|b", true},
		{"Group", "(a)", true},
		{"Unicode property", "\\p{L}", true},
		{"Back reference", "(a)\\1", false},
		{"Lookahead", "(?=a)", false},
		{"Lookbehind", "(?<=a)", false},
		{"Non-capturing group", "(?:a)", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidIRegexp(tt.pattern)
			if result != tt.valid {
				t.Errorf("IsValidIRegexp(%q) = %v, want %v", tt.pattern, result, tt.valid)
			}
		})
	}
}

func TestIRegexpToGoRegexp(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		{"Simple pattern", "^S.*$", false},
		{"Invalid pattern", "(?=a)", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goRegexp, err := IRegexpToGoRegexp(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("IRegexpToGoRegexp(%q) error = %v, wantErr %v", tt.pattern, err, tt.wantErr)
				return
			}
			
			if !tt.wantErr {
				// 验证转换后的正则可以编译
				_, err := regexp.Compile(goRegexp)
				if err != nil {
					t.Errorf("Failed to compile converted regexp: %v", err)
				}
			}
		})
	}
}
```

**Step 3: 运行测试确认通过**

```bash
go test -run TestIRegexp -v
```

Expected: PASS

**Step 4: 运行全量测试确认无回归**

```bash
go test ./...
```

Expected: 所有测试通过

**Step 5: Commit**

```bash
git add iregexp.go iregexp_test.go
git commit -m "feat: implement I-Regexp parser and validator"
```

---

## Task 4: 运行RFC 9535测试套件验证Phase 3效果

**Files:**
- Modify: `rfc9535_test.go` (更新基线数据)

**Step 1: 运行RFC 9535测试套件**

```bash
go test -run TestRFC9535Suite -v
```

**Step 2: 对比基线数据**

```bash
cat testdata/rfc9535/baseline.txt
```

**Step 3: 更新基线数据**

将新数据追加到 `testdata/rfc9535/baseline.txt`

**Step 4: Commit**

```bash
git add testdata/rfc9535/baseline.txt
git commit -m "test: verify Phase 3 against RFC 9535 suite"
```

---

## 验证标准

- [ ] `Node{Location: "$['name']", Value: "test"}` 可创建
- [ ] `NodeList` JSON序列化输出 `[{"location":"$[0]","value":1}]`
- [ ] `LogicalTrue.String()` 返回 `"true"`
- [ ] `$` → `"$"`
- [ ] `["store","book","0"]` → `"$['store']['book'][0]"`
- [ ] 含单引号的键名正确转义：`$['it\'s']`
- [ ] 含反斜杠的键名正确转义：`$['back\\slash']`
- [ ] 空键名：`$['']`
- [ ] `"^S.*$"` 识别为合法I-Regexp
- [ ] `"(a)\\1"` 识别为非法I-Regexp（反向引用）
- [ ] `"(?=a)"` 识别为非法I-Regexp（前瞻）
- [ ] `"\\p{L}"` 识别为合法I-Regexp
- [ ] 转换后的正则在Go `regexp` 中可编译
- [ ] 通过率不低于Phase 2最终结果
- [ ] 新类型和模块未引入回归

---

## 执行选项

Plan complete and saved to `docs/plans/2026-05-06-phase-3-type-system.md`. Two execution options:

**1. Subagent-Driven (this session)** - I dispatch fresh subagent per task, review between tasks, fast iteration

**2. Parallel Session (separate)** - Open new session with executing-plans, batch execution with checkpoints

Which approach?
