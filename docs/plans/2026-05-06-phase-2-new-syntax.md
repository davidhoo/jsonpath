# Phase 2: v2.1.0 新增语法实施计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 添加存在性测试和@单独引用，纯新增，不影响现有查询

**Architecture:** 通过测试驱动开发添加存在性测试语法和@单独引用支持

**Tech Stack:** Go, testing, parser extensions

---

## Task 1: 添加存在性测试[?@.name]

**Files:**
- Modify: `parser_test.go`
- Modify: `example_test.go`
- Modify: `parser.go`
- Modify: `segments.go`

**Step 1: 添加解析测试**

在 `parser_test.go` 中添加：

```go
func TestExistenceTestParsing(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantErr  bool
	}{
		{
			name: "Simple existence test",
			path: `[?@.name]`,
			wantErr: false,
		},
		{
			name: "Nested existence test",
			path: `[?@.nested.field]`,
			wantErr: false,
		},
		{
			name: "Existence test with comparison",
			path: `[?@.name == "foo"]`,
			wantErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}
```

**Step 2: 添加集成测试**

在 `example_test.go` 中添加：

```go
func TestExistenceTestIntegration(t *testing.T) {
	tests := []struct {
		name     string
		document interface{}
		path     string
		expected []interface{}
	}{
		{
			name: "Existence test filters objects with field",
			document: []interface{}{
				map[string]interface{}{"name": "a", "v": 1},
				map[string]interface{}{"v": 2},
				map[string]interface{}{"name": "b", "v": 3},
			},
			path:     `$[?@.name]`,
			expected: []interface{}{
				map[string]interface{}{"name": "a", "v": 1},
				map[string]interface{}{"name": "b", "v": 3},
			},
		},
		{
			name: "Existence test excludes null values",
			document: []interface{}{
				map[string]interface{}{"a": nil},
				map[string]interface{}{"a": 1},
			},
			path:     `$[?@.a]`,
			expected: []interface{}{
				map[string]interface{}{"a": 1},
			},
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

**Step 3: 运行测试确认失败**

```bash
go test -run TestExistenceTest -v
```

Expected: FAIL - 存在性测试语法未实现

**Step 4: 修改解析器支持存在性测试**

在 `parser.go` 中修改过滤表达式解析逻辑：
- 解析 `[?@.path]` 形式（无比较运算符）为存在性测试
- 修改 `segments.go`：过滤器求值时，存在性测试检查字段是否存在且非 null

**Step 5: 运行测试确认通过**

```bash
go test -run TestExistenceTest -v
```

Expected: PASS

**Step 6: 运行全量测试确认无回归**

```bash
go test ./...
```

Expected: 所有测试通过

**Step 7: Commit**

```bash
git add parser_test.go example_test.go parser.go segments.go
git commit -m "feat: add existence test [?@.name]"
```

---

## Task 2: 添加@单独引用

**Files:**
- Modify: `parser_test.go`
- Modify: `example_test.go`
- Modify: `parser.go`

**Step 1: 添加@单独引用测试**

在测试文件中添加：

```go
func TestBareAtReference(t *testing.T) {
	tests := []struct {
		name     string
		document interface{}
		path     string
		expected []interface{}
	}{
		{
			name:     "Bare @ with numeric comparison",
			document: []interface{}{5, 3, 1, 4, 2},
			path:     `$[?@>3]`,
			expected: []interface{}{5, 4},
		},
		{
			name:     "Bare @ with string comparison",
			document: []interface{}{"a", "b", "c"},
			path:     `$[?@=="b"]`,
			expected: []interface{}{"b"},
		},
		{
			name: "Bare @ in compound expression",
			document: []interface{}{
				map[string]interface{}{"t": "a", "v": 1},
				map[string]interface{}{"t": "b", "v": 2},
			},
			path: `$[?@.t=="a"&&@.v>0]`,
			expected: []interface{}{
				map[string]interface{}{"t": "a", "v": 1},
			},
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
go test -run TestBareAtReference -v
```

Expected: FAIL - @单独引用语法未实现

**Step 3: 修改解析器支持@单独引用**

在 `parser.go` 中修改：
- 允许 `@` 后直接跟比较运算符（不跟 `.field`）
- 确保 `@` 在复合表达式中正确引用当前节点

**Step 4: 运行测试确认通过**

```bash
go test -run TestBareAtReference -v
```

Expected: PASS

**Step 5: 运行全量测试确认无回归**

```bash
go test ./...
```

Expected: 所有测试通过

**Step 6: Commit**

```bash
git add parser_test.go example_test.go parser.go
git commit -m "feat: support bare @ reference in filters"
```

---

## Task 3: v2.1.0发布准备

**Files:**
- Modify: `CHANGELOG.md`
- Modify: `README.md`
- Modify: `README_zh.md`
- Modify: `version.go`

**Step 1: 运行RFC 9535测试套件记录最终通过率**

```bash
go test -run TestRFC9535Suite -v
```

**Step 2: 更新CHANGELOG.md**

在 `CHANGELOG.md` 中添加v2.1.0版本记录：

```markdown
## [2.1.0] - 2026-05-06

### Bug Fixes
- fix: length() now counts Unicode runes instead of bytes
- fix: implement correct operator precedence (&& before ||)
- fix: recursive descent now includes root node
- fix: selectors return empty result on type mismatch instead of error

### New Features
- feat: add existence test [?@.name]
- feat: support bare @ reference in filters

### Improvements
- test: integrate RFC 9535 test suite and establish baseline
- test: verify Phase 1 fixes against RFC 9535 suite
```

**Step 3: 更新README.md和README_zh.md**

在README中标注已知的非合规行为。

**Step 4: 更新版本号**

修改 `version.go`：

```go
package jsonpath

// Version is the current version of jsonpath
const Version = "2.1.0"

// VersionWithPrefix returns the version with v prefix
func VersionWithPrefix() string {
	return "v" + Version
}
```

**Step 5: 运行全量测试确认无回归**

```bash
go test ./...
```

**Step 6: 构建并验证版本**

```bash
go build ./cmd/jp && ./jp --version
```

Expected: 输出 `v2.1.0`

**Step 7: Commit**

```bash
git add CHANGELOG.md README.md README_zh.md version.go
git commit -m "release: v2.1.0"
```

---

## 验证标准

- [ ] `[?@.name]` 正确过滤
- [ ] `[?@.nested.field]` 支持嵌套
- [ ] null值字段不匹配
- [ ] 现有 `[?@.name == "foo"]` 不受影响
- [ ] `[?@ > 3]` 正确过滤数值
- [ ] `[?@ == "b"]` 正确过滤字符串
- [ ] `@` 在复合表达式中正确引用当前节点
- [ ] 现有 `@.field` 语法不受影响
- [ ] `go test ./...` 全部通过
- [ ] RFC 9535通过率高于Phase 0基线
- [ ] CHANGELOG包含所有变更
- [ ] README标注了已知非合规行为
- [ ] `go build ./cmd/jp && ./jp --version` 输出 `v2.1.0`

---

## 执行选项

Plan complete and saved to `docs/plans/2026-05-06-phase-2-new-syntax.md`. Two execution options:

**1. Subagent-Driven (this session)** - I dispatch fresh subagent per task, review between tasks, fast iteration

**2. Parallel Session (separate)** - Open new session with executing-plans, batch execution with checkpoints

Which approach?
