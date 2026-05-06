# Phase 4: v3.0.0 内部重构实施计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 将内部实现迁移到新类型系统，保持公共API暂不变

**Architecture:** 定义新的segment接口，逐个迁移段实现，重构过滤器系统为三值逻辑，实现Node贯穿查询管道

**Tech Stack:** Go, interfaces, refactoring, three-valued logic

---

## Task 1: 重构segment接口

**Files:**
- Modify: `segments.go`
- Create: `segments_v3.go`

**Step 1: 定义新的segment接口**

在 `segments_v3.go` 中定义：

```go
package jsonpath

// segmentV3 定义v3版本的segment接口
type segmentV3 interface {
	evaluate(node Node) (NodeList, error)
	String() string
}

// wildcardSegmentV3 通配符段的v3实现
type wildcardSegmentV3 struct{}

func (s *wildcardSegmentV3) evaluate(node Node) (NodeList, error) {
	var result NodeList
	
	switch v := node.Value.(type) {
	case map[string]interface{}:
		for key, value := range v {
			childNode := Node{
				Location: GenerateNormalizedPath([]interface{}{key}),
				Value:    value,
			}
			result = append(result, childNode)
		}
	case []interface{}:
		for i, value := range v {
			childNode := Node{
				Location: GenerateNormalizedPath([]interface{}{i}),
				Value:    value,
			}
			result = append(result, childNode)
		}
	}
	
	return result, nil
}

func (s *wildcardSegmentV3) String() string {
	return ".*"
}
```

**Step 2: 添加wildcardSegmentV3测试**

在 `segments_v3_test.go` 中添加：

```go
package jsonpath

import "testing"

func TestWildcardSegmentV3(t *testing.T) {
	tests := []struct {
		name     string
		node     Node
		expected int
	}{
		{
			name: "Object wildcard",
			node: Node{
				Location: "$",
				Value: map[string]interface{}{
					"a": 1,
					"b": 2,
				},
			},
			expected: 2,
		},
		{
			name: "Array wildcard",
			node: Node{
				Location: "$",
				Value:    []interface{}{1, 2, 3},
			},
			expected: 3,
		},
		{
			name: "Empty object",
			node: Node{
				Location: "$",
				Value:    map[string]interface{}{},
			},
			expected: 0,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			segment := &wildcardSegmentV3{}
			result, err := segment.evaluate(tt.node)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			
			if len(result) != tt.expected {
				t.Errorf("Expected %d results, got %d", tt.expected, len(result))
			}
		})
	}
}
```

**Step 3: 运行测试确认通过**

```bash
go test -run TestWildcardSegmentV3 -v
```

Expected: PASS

**Step 4: 运行全量测试确认无回归**

```bash
go test ./...
```

Expected: 所有测试通过

**Step 5: Commit**

```bash
git add segments_v3.go segments_v3_test.go
git commit -m "refactor: define v3 segment interface and migrate wildcardSegment"
```

---

## Task 2: 逐个迁移剩余段实现

**Files:**
- Modify: `segments_v3.go`
- Modify: `segments_v3_test.go`

**Step 1: 迁移nameSegment**

在 `segments_v3.go` 中添加：

```go
// nameSegmentV3 字段访问段的v3实现
type nameSegmentV3 struct {
	name string
}

func (s *nameSegmentV3) evaluate(node Node) (NodeList, error) {
	var result NodeList
	
	if obj, ok := node.Value.(map[string]interface{}); ok {
		if value, exists := obj[s.name]; exists {
			childNode := Node{
				Location: GenerateNormalizedPath([]interface{}{s.name}),
				Value:    value,
			}
			result = append(result, childNode)
		}
	}
	
	return result, nil
}

func (s *nameSegmentV3) String() string {
	return "['" + s.name + "']"
}
```

**Step 2: 添加nameSegmentV3测试**

```go
func TestNameSegmentV3(t *testing.T) {
	segment := &nameSegmentV3{name: "name"}
	node := Node{
		Location: "$",
		Value: map[string]interface{}{
			"name": "test",
			"age":  30,
		},
	}
	
	result, err := segment.evaluate(node)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	if len(result) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(result))
	}
	
	if result[0].Value != "test" {
		t.Errorf("Expected 'test', got %v", result[0].Value)
	}
}
```

**Step 3: 继续迁移其他段实现**

按照相同模式迁移：
- `indexSegmentV3` - 数组索引
- `sliceSegmentV3` - 数组切片
- `multiIndexSegmentV3` - 多索引
- `multiNameSegmentV3` - 多字段名
- `recursiveSegmentV3` - 递归下降
- `filterSegmentV3` - 过滤器
- `functionSegmentV3` - 函数调用

**Step 4: 每迁移一个段都运行测试**

```bash
go test -run TestNameSegmentV3 -v
go test -run TestIndexSegmentV3 -v
# ... 依此类推
```

**Step 5: 每迁移一个段都Commit**

```bash
git add segments_v3.go segments_v3_test.go
git commit -m "refactor: migrate nameSegment to v3 interface"
```

---

## Task 3: 重构过滤器系统

**Files:**
- Modify: `segments_v3.go`
- Modify: `segments_v3_test.go`

**Step 1: 实现三值逻辑过滤器**

在 `segments_v3.go` 中添加：

```go
// filterEvaluatorV3 过滤器求值器
type filterEvaluatorV3 struct {
	expression filterExpression
}

// filterExpression 过滤器表达式接口
type filterExpression interface {
	evaluate(node Node) LogicalType
}

// existenceExpression 存在性测试表达式
type existenceExpression struct {
	path string
}

func (e *existenceExpression) evaluate(node Node) LogicalType {
	// 检查字段是否存在且非null
	// 实现存在性测试逻辑
	return LogicalFalse
}

// comparisonExpression 比较表达式
type comparisonExpression struct {
	left     filterExpression
	operator string
	right    filterExpression
}

func (e *comparisonExpression) evaluate(node Node) LogicalType {
	left := e.left.evaluate(node)
	right := e.right.evaluate(node)
	
	switch e.operator {
	case "&&":
		if left == LogicalFalse || right == LogicalFalse {
			return LogicalFalse
		}
		if left == LogicalNothing || right == LogicalNothing {
			return LogicalNothing
		}
		return LogicalTrue
	case "||":
		if left == LogicalTrue || right == LogicalTrue {
			return LogicalTrue
		}
		if left == LogicalNothing || right == LogicalNothing {
			return LogicalNothing
		}
		return LogicalFalse
	case "!":
		if left == LogicalTrue {
			return LogicalFalse
		}
		if left == LogicalFalse {
			return LogicalTrue
		}
		return LogicalNothing
	default:
		return LogicalNothing
	}
}

// filterSegmentV3 过滤器段的v3实现
type filterSegmentV3 struct {
	expression filterExpression
}

func (s *filterSegmentV3) evaluate(node Node) (NodeList, error) {
	var result NodeList
	
	// 根据节点类型进行过滤
	switch v := node.Value.(type) {
	case []interface{}:
		for i, item := range v {
			itemNode := Node{
				Location: GenerateNormalizedPath([]interface{}{i}),
				Value:    item,
			}
			if s.expression.evaluate(itemNode) == LogicalTrue {
				result = append(result, itemNode)
			}
		}
	case map[string]interface{}:
		for key, value := range v {
			itemNode := Node{
				Location: GenerateNormalizedPath([]interface{}{key}),
				Value:    value,
			}
			if s.expression.evaluate(itemNode) == LogicalTrue {
				result = append(result, itemNode)
			}
		}
	}
	
	return result, nil
}

func (s *filterSegmentV3) String() string {
	return "[?...]"
}
```

**Step 2: 添加三值逻辑测试**

```go
func TestThreeValuedLogic(t *testing.T) {
	tests := []struct {
		name     string
		left     LogicalType
		operator string
		right    LogicalType
		expected LogicalType
	}{
		{"True AND True", LogicalTrue, "&&", LogicalTrue, LogicalTrue},
		{"True AND False", LogicalTrue, "&&", LogicalFalse, LogicalFalse},
		{"True AND Nothing", LogicalTrue, "&&", LogicalNothing, LogicalNothing},
		{"False AND False", LogicalFalse, "&&", LogicalFalse, LogicalFalse},
		{"True OR False", LogicalTrue, "||", LogicalFalse, LogicalTrue},
		{"False OR False", LogicalFalse, "||", LogicalFalse, LogicalFalse},
		{"False OR Nothing", LogicalFalse, "||", LogicalNothing, LogicalNothing},
		{"NOT True", LogicalTrue, "!", LogicalTrue, LogicalFalse},
		{"NOT False", LogicalFalse, "!", LogicalFalse, LogicalTrue},
		{"NOT Nothing", LogicalNothing, "!", LogicalNothing, LogicalNothing},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &comparisonExpression{
				left:     &literalExpression{value: tt.left},
				operator: tt.operator,
				right:    &literalExpression{value: tt.right},
			}
			
			result := expr.evaluate(Node{})
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// literalExpression 字面量表达式（用于测试）
type literalExpression struct {
	value LogicalType
}

func (e *literalExpression) evaluate(node Node) LogicalType {
	return e.value
}
```

**Step 3: 运行测试确认通过**

```bash
go test -run TestThreeValuedLogic -v
```

Expected: PASS

**Step 4: Commit**

```bash
git add segments_v3.go segments_v3_test.go
git commit -m "refactor: filter system uses three-valued logic"
```

---

## Task 4: 统一段实现为新接口

**Files:**
- Modify: `segments.go`
- Modify: `segments_v3.go`
- Modify: `segments_test.go`

**Step 1: 移除旧segment接口**

从 `segments.go` 中移除旧的segment接口定义。

**Step 2: 将segmentV3重命名为segment**

在 `segments_v3.go` 中：

```go
// 将segmentV3重命名为segment
type segment interface {
	evaluate(node Node) (NodeList, error)
	String() string
}
```

**Step 3: 更新所有调用点**

更新所有使用旧segment接口的地方。

**Step 4: 运行全量测试确认无回归**

```bash
go test ./...
```

Expected: 所有测试通过

**Step 5: Commit**

```bash
git add segments.go segments_v3.go segments_test.go
git commit -m "refactor: remove legacy segment interface"
```

---

## Task 5: 实现Node贯穿查询管道

**Files:**
- Modify: `jsonpath.go`
- Modify: `segments.go`

**Step 1: 修改Query()函数**

在 `jsonpath.go` 中修改：

```go
// Query 执行JSONPath查询
func Query(data interface{}, path string) (interface{}, error) {
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
	
	// 提取值（保持向后兼容）
	var values []interface{}
	for _, node := range currentNodes {
		values = append(values, node.Value)
	}
	
	return values, nil
}
```

**Step 2: 添加Node管道测试**

```go
func TestNodePipeline(t *testing.T) {
	data := map[string]interface{}{
		"store": map[string]interface{}{
			"book": []interface{}{
				map[string]interface{}{"title": "Book 1"},
				map[string]interface{}{"title": "Book 2"},
			},
		},
	}
	
	result, err := Query(data, "$.store.book[0]")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	values, ok := result.([]interface{})
	if !ok {
		t.Fatalf("Expected slice, got %T", result)
	}
	
	if len(values) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(values))
	}
	
	book, ok := values[0].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map, got %T", values[0])
	}
	
	if book["title"] != "Book 1" {
		t.Errorf("Expected 'Book 1', got %v", book["title"])
	}
}
```

**Step 3: 运行测试确认通过**

```bash
go test -run TestNodePipeline -v
```

Expected: PASS

**Step 4: 运行全量测试确认无回归**

```bash
go test ./...
```

Expected: 所有测试通过

**Step 5: Commit**

```bash
git add jsonpath.go segments.go
git commit -m "refactor: query pipeline passes Node with Normalized Path"
```

---

## 验证标准

- [ ] `segmentV3` 接口定义存在
- [ ] `wildcardSegment` 实现新接口并通过测试
- [ ] 旧接口和实现不受影响
- [ ] 过滤器返回 `LogicalType`
- [ ] 三值逻辑真值表全部正确
- [ ] 旧接口完全移除
- [ ] 仅保留新接口
- [ ] 查询管道全程传递 `Node`（含Location）
- [ ] `Query(data, "$.store.book[0]")` 的结果中每个Node的Location正确
- [ ] `go test ./...` 无失败

---

## 执行选项

Plan complete and saved to `docs/plans/2026-05-06-phase-4-internal-refactor.md`. Two execution options:

**1. Subagent-Driven (this session)** - I dispatch fresh subagent per task, review between tasks, fast iteration

**2. Parallel Session (separate)** - Open new session with executing-plans, batch execution with checkpoints

Which approach?
