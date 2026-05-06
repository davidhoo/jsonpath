# Phase 1: v2.1.0 行为修正实施计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 修复4个已知行为错误，不改变任何公共API签名

**Architecture:** 通过测试驱动开发修复length()函数、运算符优先级、递归下降和选择器错误处理

**Tech Stack:** Go, testing, Unicode, operator precedence

---

## Task 1: 修复length()字符串计算

**Files:**
- Modify: `functions_test.go`
- Modify: `functions.go`

**Step 1: 添加Unicode字符串测试**

在 `functions_test.go` 中添加测试：

```go
func TestLengthUnicode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"Japanese", "日本語", 3},
		{"ASCII", "hello", 5},
		{"Accented", "café", 4},
		{"4-byte Unicode", "𝄞", 1},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := length(tt.input)
			if result != tt.expected {
				t.Errorf("length(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}
```

**Step 2: 运行测试确认失败**

```bash
go test -run TestLengthUnicode -v
```

Expected: FAIL - length函数按字节计数而非字符

**Step 3: 修改length()函数**

在 `functions.go` 中找到length函数，修改字符串处理：

```go
// 找到类似这样的代码：
// case string:
//     return len(s)

// 修改为：
case string:
    return utf8.RuneCountInString(s)
```

需要添加导入：

```go
import "unicode/utf8"
```

**Step 4: 运行测试确认通过**

```bash
go test -run TestLengthUnicode -v
```

Expected: PASS

**Step 5: 运行全量测试确认无回归**

```bash
go test ./...
```

Expected: 所有测试通过

**Step 6: Commit**

```bash
git add functions_test.go functions.go
git commit -m "fix: length() now counts Unicode runes instead of bytes"
```

---

## Task 2: 修复&&/||运算符优先级

**Files:**
- Modify: `parser_test.go` 或 `example_test.go`
- Modify: `parser.go`

**Step 1: 添加优先级测试**

在测试文件中添加：

```go
func TestOperatorPrecedence(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		document interface{}
		expected []interface{}
	}{
		{
			name: "AND higher than OR",
			path: `[?(@.a==1||@.b==2&&@.c==3)]`,
			document: map[string]interface{}{
				"a": 1, "b": 0, "c": 0,
			},
			expected: []interface{}{map[string]interface{}{"a": 1, "b": 0, "c": 0}},
		},
		{
			name: "OR with AND",
			path: `[?(@.a==0||@.b==2&&@.c==3)]`,
			document: map[string]interface{}{
				"a": 0, "b": 2, "c": 3,
			},
			expected: []interface{}{map[string]interface{}{"a": 0, "b": 2, "c": 3}},
		},
		{
			name: "AND with OR",
			path: `[?(@.a==0||@.b==2&&@.c==0)]`,
			document: map[string]interface{}{
				"a": 0, "b": 2, "c": 0,
			},
			expected: []interface{}{},
		},
		{
			name: "Parentheses override precedence",
			path: `[?((@.a==1||@.b==2)&&@.c==3)]`,
			document: map[string]interface{}{
				"a": 1, "b": 0, "c": 3,
			},
			expected: []interface{}{map[string]interface{}{"a": 1, "b": 0, "c": 3}},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Query(tt.document, tt.path)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			
			expectedJSON, _ := json.Marshal(tt.expected)
			resultJSON, _ := json.Marshal(result)
			
			if string(expectedJSON) != string(resultJSON) {
				t.Errorf("Expected %s, got %s", expectedJSON, resultJSON)
			}
		})
	}
}
```

**Step 2: 运行测试确认失败**

```bash
go test -run TestOperatorPrecedence -v
```

Expected: FAIL - 运算符优先级错误

**Step 3: 修改解析器实现正确的优先级**

在 `parser.go` 中修改过滤表达式的解析逻辑，实现 `&&` 优先于 `||`。

具体修改取决于当前实现，可能需要：
- 修改解析器语法
- 实现运算符优先级解析
- 添加括号支持

**Step 4: 运行测试确认通过**

```bash
go test -run TestOperatorPrecedence -v
```

Expected: PASS

**Step 5: 运行全量测试确认无回归**

```bash
go test ./...
```

Expected: 所有测试通过

**Step 6: Commit**

```bash
git add parser_test.go parser.go
git commit -m "fix: implement correct operator precedence (&& before ||)"
```

---

## Task 3: 修复递归下降包含根节点

**Files:**
- Modify: `segments_test.go` 或 `example_test.go`
- Modify: `segments.go`

**Step 1: 添加递归下降测试**

在测试文件中添加：

```go
func TestRecursiveDescentIncludesRoot(t *testing.T) {
	tests := []struct {
		name     string
		document interface{}
		path     string
		expected []interface{}
	}{
		{
			name: "Recursive descent with name",
			document: map[string]interface{}{
				"name": "root",
				"child": map[string]interface{}{
					"name": "child1",
				},
			},
			path:     `$..name`,
			expected: []interface{}{"root", "child1"},
		},
		{
			name: "Recursive descent with value",
			document: map[string]interface{}{
				"value": 1,
			},
			path:     `$..value`,
			expected: []interface{}{1},
		},
		{
			name: "Recursive descent wildcard",
			document: map[string]interface{}{
				"a": 1,
				"b": 2,
			},
			path:     `$..*`,
			expected: []interface{}{1, 2},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Query(tt.document, tt.path)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			
			expectedJSON, _ := json.Marshal(tt.expected)
			resultJSON, _ := json.Marshal(result)
			
			if string(expectedJSON) != string(resultJSON) {
				t.Errorf("Expected %s, got %s", expectedJSON, resultJSON)
			}
		})
	}
}
```

**Step 2: 运行测试确认失败**

```bash
go test -run TestRecursiveDescentIncludesRoot -v
```

Expected: FAIL - 递归下降不包含根节点

**Step 3: 修改递归下降实现**

在 `segments.go` 中找到 `recursiveSegment` 的 `evaluate` 方法，修改为从根节点开始递归。

**Step 4: 运行测试确认通过**

```bash
go test -run TestRecursiveDescentIncludesRoot -v
```

Expected: PASS

**Step 5: 运行全量测试确认无回归**

```bash
go test ./...
```

Expected: 所有测试通过

**Step 6: Commit**

```bash
git add segments_test.go segments.go
git commit -m "fix: recursive descent now includes root node"
```

---

## Task 4: 修复选择器错误处理

**Files:**
- Modify: `segments_test.go` 或 `errors_test.go`
- Modify: `segments.go`

**Step 1: 添加选择器错误处理测试**

在测试文件中添加：

```go
func TestSelectorErrorHandling(t *testing.T) {
	tests := []struct {
		name     string
		document interface{}
		path     string
		expected []interface{}
		wantErr  bool
	}{
		{
			name: "Index on string",
			document: map[string]interface{}{
				"name": "hello",
			},
			path:     `$.name[0]`,
			expected: []interface{}{},
			wantErr:  false,
		},
		{
			name: "Field on number",
			document: map[string]interface{}{
				"count": 42,
			},
			path:     `$.count.foo`,
			expected: []interface{}{},
			wantErr:  false,
		},
		{
			name: "Selector on null",
			document: map[string]interface{}{
				"null_field": nil,
			},
			path:     `$.null_field.bar`,
			expected: []interface{}{},
			wantErr:  false,
		},
		{
			name: "Syntax error",
			document: map[string]interface{}{
				"store": map[string]interface{}{},
			},
			path:    `$.store[`,
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Query(tt.document, tt.path)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
				return
			}
			
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			
			expectedJSON, _ := json.Marshal(tt.expected)
			resultJSON, _ := json.Marshal(result)
			
			if string(expectedJSON) != string(resultJSON) {
				t.Errorf("Expected %s, got %s", expectedJSON, resultJSON)
			}
		})
	}
}
```

**Step 2: 运行测试确认失败**

```bash
go test -run TestSelectorErrorHandling -v
```

Expected: FAIL - 类型不匹配返回错误而非空结果

**Step 3: 修改选择器错误处理**

在 `segments.go` 中修改各段的 `evaluate` 方法：
- 类型不匹配时返回空 `[]interface{}{}` 而非 error
- 语法错误仍返回 error

**Step 4: 运行测试确认通过**

```bash
go test -run TestSelectorErrorHandling -v
```

Expected: PASS

**Step 5: 运行全量测试确认无回归**

```bash
go test ./...
```

Expected: 所有测试通过

**Step 6: Commit**

```bash
git add segments_test.go segments.go
git commit -m "fix: selectors return empty result on type mismatch instead of error"
```

---

## Task 5: 运行RFC 9535测试套件验证Phase 1效果

**Files:**
- Modify: `rfc9535_test.go` (更新基线数据)

**Step 1: 重新运行RFC 9535测试套件**

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
git commit -m "test: verify Phase 1 fixes against RFC 9535 suite"
```

---

## 验证标准

- [ ] `go test -run TestLengthUnicode -v` 全部通过
- [ ] `&&` 优先级高于 `||`
- [ ] 括号可覆盖默认优先级
- [ ] `$..name` 包含根节点匹配
- [ ] `$..*` 包含根节点自身
- [ ] 类型不匹配返回空结果，`err == nil`
- [ ] 语法错误仍返回 `err != nil`
- [ ] `go test ./...` 无失败
- [ ] 通过数高于Phase 0基线
- [ ] 新通过的用例与Phase 1修复的4个问题相关

---

## 执行选项

Plan complete and saved to `docs/plans/2026-05-06-phase-1-behavior-fixes.md`. Two execution options:

**1. Subagent-Driven (this session)** - I dispatch fresh subagent per task, review between tasks, fast iteration

**2. Parallel Session (separate)** - Open new session with executing-plans, batch execution with checkpoints

Which approach?
