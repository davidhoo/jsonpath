# Phase 5: v3.0.0 公共API变更实施计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 变更公共API签名，实现RFC 9535标准函数

**Architecture:** 修改Query()返回类型，修正count()、search()、match()函数，新增value()函数，更新CLI工具

**Tech Stack:** Go, API design, function implementation, CLI tools

---

## Task 1: 变更Query()返回类型

**Files:**
- Modify: `jsonpath.go`
- Modify: `jsonpath_test.go`
- Modify: `cmd/jp/main.go`

**Step 1: 修改Query()函数签名**

在 `jsonpath.go` 中修改：

```go
// Query 执行JSONPath查询
func Query(data interface{}, path string) (NodeList, error) {
	// 解析路径
	segments, err := Parse(path)
	if err != nil {
		return nil, err
	}
	
	// 创建根节点
	rootNode := Node{
		Location: "$",
		Value:    data,
	}
	
	// 执行查询管道
	currentNodes := NodeList{rootNode}
	for _, seg := range segments {
		var nextNodes NodeList
		for _, node := range currentNodes {
			results, err := seg.evaluate(node)
			if err != nil {
				return nil, err
			}
			nextNodes = append(nextNodes, results...)
		}
		currentNodes = nextNodes
	}
	
	return currentNodes, nil
}
```

**Step 2: 更新测试中的Query()调用**

在 `jsonpath_test.go` 中更新所有测试：

```go
func TestQueryReturnsNodeList(t *testing.T) {
	data := map[string]interface{}{
		"name": "test",
		"age":  30,
	}
	
	result, err := Query(data, "$.name")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	nodeList, ok := result.(NodeList)
	if !ok {
		t.Fatalf("Expected NodeList, got %T", result)
	}
	
	if len(nodeList) != 1 {
		t.Fatalf("Expected 1 node, got %d", len(nodeList))
	}
	
	if nodeList[0].Value != "test" {
		t.Errorf("Expected 'test', got %v", nodeList[0].Value)
	}
	
	if nodeList[0].Location != "$['name']" {
		t.Errorf("Expected $['name'], got %s", nodeList[0].Location)
	}
}
```

**Step 3: 更新CLI工具**

在 `cmd/jp/main.go` 中更新：

```go
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/yourusername/jsonpath"
)

func main() {
	// 解析命令行参数
	path := os.Args[1]
	data, err := readInput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}
	
	// 执行查询
	result, err := jsonpath.Query(data, path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing query: %v\n", err)
		os.Exit(1)
	}
	
	// 输出结果
	nodeList, ok := result.(jsonpath.NodeList)
	if !ok {
		fmt.Fprintf(os.Stderr, "Unexpected result type: %T\n", result)
		os.Exit(1)
	}
	
	// 默认输出值（兼容旧行为）
	for _, node := range nodeList {
		jsonData, err := json.Marshal(node.Value)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshaling result: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(jsonData))
	}
}

func readInput() (interface{}, error) {
	var data interface{}
	decoder := json.NewDecoder(os.Stdin)
	if err := decoder.Decode(&data); err != nil {
		return nil, err
	}
	return data, nil
}
```

**Step 4: 运行测试确认通过**

```bash
go test -run TestQueryReturnsNodeList -v
```

Expected: PASS

**Step 5: 构建CLI工具**

```bash
go build ./cmd/jp
```

Expected: 编译成功

**Step 6: Commit**

```bash
git add jsonpath.go jsonpath_test.go cmd/jp/main.go
git commit -m "breaking: Query() returns NodeList instead of interface{}"
```

---

## Task 2: 修正count()函数

**Files:**
- Modify: `functions.go`
- Modify: `functions_test.go`

**Step 1: 将现有count()重命名为occurrences()**

在 `functions.go` 中：

```go
// occurrences 计算值出现次数（非标准扩展）
func occurrences(args ...interface{}) interface{} {
	// 保留原有count()的实现
	// ...
}

// count 计算节点数量（RFC标准）
func count(args ...interface{}) interface{} {
	if len(args) != 1 {
		return Nothing{}
	}
	
	nodeList, ok := args[0].(NodeList)
	if !ok {
		return Nothing{}
	}
	
	return float64(len(nodeList))
}
```

**Step 2: 更新函数注册**

在函数注册表中更新：

```go
var functions = map[string]Function{
	"length":      {fn: length,ReturnType: "number"},
	"count":       {fn: count, ReturnType: "number"},
	"occurrences": {fn: occurrences, ReturnType: "number"},
	// ... 其他函数
}
```

**Step 3: 添加测试**

在 `functions_test.go` 中添加：

```go
func TestCountFunction(t *testing.T) {
	tests := []struct {
		name     string
		args     []interface{}
		expected interface{}
	}{
		{
			name:     "Count node list",
			args:     []interface{}{NodeList{{}, {}, {}}},
			expected: float64(3),
		},
		{
			name:     "Count empty node list",
			args:     []interface{}{NodeList{}},
			expected: float64(0),
		},
		{
			name:     "Count non-node list",
			args:     []interface{}{"not a node list"},
			expected: Nothing{},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := count(tt.args...)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestOccurrencesFunction(t *testing.T) {
	tests := []struct {
		name     string
		args     []interface{}
		expected interface{}
	}{
		{
			name:     "Count occurrences",
			args:     []interface{}{[]interface{}{"a", "b", "a", "c"}, "a"},
			expected: float64(2),
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := occurrences(tt.args...)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
```

**Step 4: 运行测试确认通过**

```bash
go test -run TestCountFunction -v
go test -run TestOccurrencesFunction -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add functions.go functions_test.go
git commit -m "breaking: rename count() to occurrences(), add RFC count()"
```

---

## Task 3: 修正search()函数

**Files:**
- Modify: `functions.go`
- Modify: `functions_test.go`

**Step 1: 将现有search()重命名为filterMatch()**

在 `functions.go` 中：

```go
// filterMatch 按正则过滤数组（非标准扩展）
func filterMatch(args ...interface{}) interface{} {
	// 保留原有search()的实现
	// ...
}

// search 字符串子串正则搜索（RFC标准）
func search(args ...interface{}) interface{} {
	if len(args) != 2 {
		return LogicalNothing
	}
	
	str, ok := args[0].(string)
	if !ok {
		return LogicalNothing
	}
	
	pattern, ok := args[1].(string)
	if !ok {
		return LogicalNothing
	}
	
	// 使用I-Regexp匹配
	goRegexp, err := IRegexpToGoRegexp(pattern)
	if err != nil {
		return LogicalNothing
	}
	
	matched, err := regexp.MatchString(goRegexp, str)
	if err != nil {
		return LogicalNothing
	}
	
	if matched {
		return LogicalTrue
	}
	return LogicalFalse
}
```

**Step 2: 更新函数注册**

```go
var functions = map[string]Function{
	// ...
	"search":      {fn: search, ReturnType: "logical"},
	"filterMatch": {fn: filterMatch, ReturnType: "array"},
	// ...
}
```

**Step 3: 添加测试**

```go
func TestSearchFunction(t *testing.T) {
	tests := []struct {
		name     string
		args     []interface{}
		expected interface{}
	}{
		{
			name:     "Search matches substring",
			args:     []interface{}{"hello world", "world"},
			expected: LogicalTrue,
		},
		{
			name:     "Search no match",
			args:     []interface{}{"hello world", "xyz"},
			expected: LogicalFalse,
		},
		{
			name:     "Search invalid pattern",
			args:     []interface{}{"hello", "(?=a)"},
			expected: LogicalNothing,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := search(tt.args...)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
```

**Step 4: 运行测试确认通过**

```bash
go test -run TestSearchFunction -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add functions.go functions_test.go
git commit -m "breaking: rename search() to filterMatch(), add RFC search()"
```

---

## Task 4: 变更match()调用形式

**Files:**
- Modify: `parser.go`
- Modify: `functions.go`
- Modify: `functions_test.go`

**Step 1: 修改解析器支持函数式match()**

在 `parser.go` 中修改：

```go
// 解析match()函数调用
func parseMatchFunction(tokens []Token) (filterExpression, error) {
	// 解析 match(@.path, 'pattern') 形式
	// ...
}
```

**Step 2: 修改match()函数使用全串匹配**

在 `functions.go` 中：

```go
// match 全串匹配（RFC标准）
func match(args ...interface{}) interface{} {
	if len(args) != 2 {
		return LogicalNothing
	}
	
	str, ok := args[0].(string)
	if !ok {
		return LogicalNothing
	}
	
	pattern, ok := args[1].(string)
	if !ok {
		return LogicalNothing
	}
	
	// 使用I-Regexp全串匹配
	goRegexp, err := IRegexpToGoRegexp(pattern)
	if err != nil {
		return LogicalNothing
	}
	
	// 添加锚点确保全串匹配
	fullPattern := "^" + goRegexp + "$"
	matched, err := regexp.MatchString(fullPattern, str)
	if err != nil {
		return LogicalNothing
	}
	
	if matched {
		return LogicalTrue
	}
	return LogicalFalse
}
```

**Step 3: 添加测试**

```go
func TestMatchFunction(t *testing.T) {
	tests := []struct {
		name     string
		args     []interface{}
		expected interface{}
	}{
		{
			name:     "Exact match",
			args:     []interface{}{"S", "S"},
			expected: LogicalTrue,
		},
		{
			name:     "Partial match fails",
			args:     []interface{}{"AliceSmith", "S"},
			expected: LogicalFalse,
		},
		{
			name:     "Pattern match",
			args:     []interface{}{"Smith", "^S.*$"},
			expected: LogicalTrue,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := match(tt.args...)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
```

**Step 4: 运行测试确认通过**

```bash
go test -run TestMatchFunction -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add parser.go functions.go functions_test.go
git commit -m "breaking: match() uses function syntax and full-string matching"
```

---

## Task 5: 新增value()函数

**Files:**
- Modify: `functions.go`
- Modify: `functions_test.go`

**Step 1: 实现value()函数**

在 `functions.go` 中：

```go
// value 返回单节点值（RFC标准）
func value(args ...interface{}) interface{} {
	if len(args) != 1 {
		return Nothing{}
	}
	
	nodeList, ok := args[0].(NodeList)
	if !ok {
		return Nothing{}
	}
	
	if len(nodeList) != 1 {
		return Nothing{}
	}
	
	return nodeList[0].Value
}
```

**Step 2: 更新函数注册**

```go
var functions = map[string]Function{
	// ...
	"value": {fn: value, ReturnType: "value"},
	// ...
}
```

**Step 3: 添加测试**

```go
func TestValueFunction(t *testing.T) {
	tests := []struct {
		name     string
		args     []interface{}
		expected interface{}
	}{
		{
			name:     "Single node",
			args:     []interface{}{NodeList{{Value: "test"}}},
			expected: "test",
		},
		{
			name:     "Multiple nodes",
			args:     []interface{}{NodeList{{Value: "a"}, {Value: "b"}}},
			expected: Nothing{},
		},
		{
			name:     "Empty node list",
			args:     []interface{}{NodeList{}},
			expected: Nothing{},
		},
		{
			name:     "Non-node list",
			args:     []interface{}{"not a node list"},
			expected: Nothing{},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := value(tt.args...)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
```

**Step 4: 运行测试确认通过**

```bash
go test -run TestValueFunction -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add functions.go functions_test.go
git commit -m "feat: add value() function"
```

---

## Task 6: 更新CLI工具

**Files:**
- Modify: `cmd/jp/main.go`

**Step 1: 添加--path标志**

在 `cmd/jp/main.go` 中：

```go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/yourusername/jsonpath"
)

func main() {
	// 定义命令行标志
	pathFlag := flag.Bool("path", false, "输出Normalized Path")
	flag.Parse()
	
	// 获取查询路径
	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: jp [--path] <jsonpath>\n")
		os.Exit(1)
	}
	queryPath := flag.Arg(0)
	
	// 读取输入数据
	data, err := readInput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}
	
	// 执行查询
	result, err := jsonpath.Query(data, queryPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing query: %v\n", err)
		os.Exit(1)
	}
	
	// 输出结果
	nodeList, ok := result.(jsonpath.NodeList)
	if !ok {
		fmt.Fprintf(os.Stderr, "Unexpected result type: %T\n", result)
		os.Exit(1)
	}
	
	for _, node := range nodeList {
		if *pathFlag {
			// 输出Normalized Path + 值
			jsonData, err := json.Marshal(node.Value)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error marshaling result: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("%s %s\n", node.Location, string(jsonData))
		} else {
			// 默认输出值（兼容旧行为）
			jsonData, err := json.Marshal(node.Value)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error marshaling result: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(jsonData))
		}
	}
}

func readInput() (interface{}, error) {
	var data interface{}
	decoder := json.NewDecoder(os.Stdin)
	if err := decoder.Decode(&data); err != nil {
		return nil, err
	}
	return data, nil
}
```

**Step 2: 测试CLI工具**

```bash
echo '{"a":1}' | go run ./cmd/jp '$.a'
```

Expected: 输出 `1`

```bash
echo '{"a":1}' | go run ./cmd/jp --path '$.a'
```

Expected: 输出 `$['a'] 1`

**Step 3: 运行CLI测试**

```bash
go test ./cmd/jp/...
```

Expected: 所有测试通过

**Step 4: Commit**

```bash
git add cmd/jp/main.go
git commit -m "feat: update CLI for NodeList output, add --path flag"
```

---

## 验证标准

- [ ] `Query()` 返回 `NodeList`
- [ ] `result[0].Value` 获取原始值
- [ ] `result[0].Location` 获取Normalized Path
- [ ] 空结果返回 `NodeList{}`（非nil），`err == nil`
- [ ] `count(@.items[*])` 返回节点数量
- [ ] `occurrences(@.items, "value")` 保留旧功能
- [ ] `search(@.name, "pattern")` 返回 `LogicalType`
- [ ] `filterMatch(@.items, "pattern")` 保留旧功能
- [ ] `match(@.name, 'S')` 只匹配恰好为 `"S"` 的字符串
- [ ] `match(@.name, 'S')` 不匹配 `"AliceSmith"`
- [ ] `value(@.name)` 返回单个值
- [ ] 多节点时返回Nothing
- [ ] `echo '{"a":1}' | jp '$.a'` 输出 `1`
- [ ] `echo '{"a":1}' | jp --path '$.a'` 输出 `$['a']` + 值
- [ ] `go test ./...` 全部通过
- [ ] `go build ./cmd/jp` 编译成功

---

## 执行选项

Plan complete and saved to `docs/plans/2026-05-06-phase-5-public-api.md`. Two execution options:

**1. Subagent-Driven (this session)** - I dispatch fresh subagent per task, review between tasks, fast iteration

**2. Parallel Session (separate)** - Open new session with executing-plans, batch execution with checkpoints

Which approach?
