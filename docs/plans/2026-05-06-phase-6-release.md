# Phase 6: 测试通过 + 发布实施计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 达到RFC 9535测试套件100%通过，发布v3.0.0

**Architecture:** 运行完整测试套件，修复剩余失败，进行回归测试和性能测试，更新文档，发布v3.0.0

**Tech Stack:** Go, testing, benchmarking, documentation, release management

---

## Task 1: 运行RFC 9535测试套件，修复剩余失败

**Files:**
- Modify: 各种源文件（根据失败测试）

**Step 1: 运行完整测试套件**

```bash
go test -run TestRFC9535Suite -v
```

**Step 2: 按失败类别分组**

分析失败的测试用例，按类别分组：
- 语法解析失败
- 选择器失败
- 过滤器失败
- 函数失败
- 边界情况失败

**Step 3: 逐个修复失败类别**

对于每个失败类别：
1. 分析失败原因
2. 修改相应代码
3. 运行测试确认修复
4. Commit修复

**Step 4: 重复直到100%通过**

继续修复直到所有324个测试用例通过。

**Step 5: Commit**

```bash
git add .
git commit -m "test: achieve 100% RFC 9535 compliance"
```

---

## Task 2: 回归测试 + 性能测试

**Files:**
- Create: `benchmark_test.go`
- Modify: `go.mod` (如果需要)

**Step 1: 运行全量项目测试**

```bash
go test ./...
```

Expected: 所有测试通过

**Step 2: 编写基准测试**

创建 `benchmark_test.go`：

```go
package jsonpath

import (
	"testing"
)

func BenchmarkQuery(b *testing.B) {
	data := map[string]interface{}{
		"store": map[string]interface{}{
			"book": []interface{}{
				map[string]interface{}{
					"title": "Book 1",
					"price": 10.99,
				},
				map[string]interface{}{
					"title": "Book 2",
					"price": 15.99,
				},
			},
		},
	}
	
	path := "$.store.book[?@.price > 10]"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Query(data, path)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkQueryComplex(b *testing.B) {
	data := map[string]interface{}{
		"users": []interface{}{
			map[string]interface{}{
				"name": "Alice",
				"age":  30,
				"address": map[string]interface{}{
					"city": "New York",
				},
			},
			map[string]interface{}{
				"name": "Bob",
				"age":  25,
				"address": map[string]interface{}{
					"city": "San Francisco",
				},
			},
		},
	}
	
	path := "$.users[?@.age > 20 && @.address.city == 'New York']"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Query(data, path)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}
```

**Step 3: 运行基准测试**

```bash
go test -bench=. -benchmem
```

Expected: 输出基准数据

**Step 4: 如果性能下降 > 10%，优化热点路径**

分析性能瓶颈，优化关键路径。

**Step 5: Commit**

```bash
git add benchmark_test.go
git commit -m "test: regression and benchmark tests"
```

---

## Task 3: 文档更新

**Files:**
- Create: `MIGRATION.md`
- Modify: `CHANGELOG.md`
- Modify: `README.md`
- Modify: `README_zh.md`

**Step 1: 编写迁移指南**

创建 `MIGRATION.md`：

```markdown
# 迁移指南：v2.x → v3.0.0

## 概述

v3.0.0 是一个主要版本更新，包含破坏性变更。本指南帮助您从 v2.x 迁移到 v3.0.0。

## 主要变更

### 1. Query() 返回类型变更

**v2.x:**
```go
result, err := jsonpath.Query(data, path)
values := result.([]interface{})
```

**v3.0.0:**
```go
result, err := jsonpath.Query(data, path)
nodeList := result.(jsonpath.NodeList)
values := make([]interface{}, len(nodeList))
for i, node := range nodeList {
    values[i] = node.Value
}
```

### 2. count() 函数变更

**v2.x:**
```go
count(@.items, "value")  // 计算值出现次数
```

**v3.0.0:**
```go
count(@.items[*])  // 计算节点数量
occurrences(@.items, "value")  // 旧功能重命名
```

### 3. search() 函数变更

**v2.x:**
```go
search(@.items, "pattern")  // 按正则过滤数组
```

**v3.0.0:**
```go
search(@.name, "pattern")  // 字符串子串搜索
filterMatch(@.items, "pattern")  // 旧功能重命名
```

### 4. match() 函数变更

**v2.x:**
```go
?@.name.match('^S.*')  // 方法式调用，部分匹配
```

**v3.0.0:**
```go
?match(@.name, '^S.*$')  // 函数式调用，全串匹配
```

## 新增功能

### 1. Normalized Path

每个查询结果节点都包含 Normalized Path：

```go
result, _ := jsonpath.Query(data, "$.store.book[0]")
node := result.(jsonpath.NodeList)[0]
fmt.Println(node.Location)  // $['store']['book'][0]
```

### 2. value() 函数

```go
value(@.name)  // 返回单节点值
```

### 3. CLI --path 标志

```bash
echo '{"a":1}' | jp --path '$.a'
# 输出: $['a'] 1
```

## 迁移步骤

1. 更新依赖到 v3.0.0
2. 修改 `Query()` 调用以使用 `NodeList`
3. 更新 `count()` 和 `search()` 的使用
4. 修改 `match()` 为函数式调用
5. 运行测试确保行为正确
```

**Step 2: 更新CHANGELOG.md**

在 `CHANGELOG.md` 中添加v3.0.0版本记录：

```markdown
## [3.0.0] - 2026-05-06

### Breaking Changes
- breaking: Query() returns NodeList instead of interface{}
- breaking: rename count() to occurrences(), add RFC count()
- breaking: rename search() to filterMatch(), add RFC search()
- breaking: match() uses function syntax and full-string matching

### New Features
- feat: add Node, NodeList, Nothing, LogicalType types
- feat: implement Normalized Path generator
- feat: implement I-Regexp parser and validator
- feat: add value() function
- feat: update CLI for NodeList output, add --path flag

### Improvements
- refactor: define v3 segment interface and migrate wildcardSegment
- refactor: migrate remaining segment implementations
- refactor: filter system uses three-valued logic
- refactor: remove legacy segment interface
- refactor: query pipeline passes Node with Normalized Path
- test: achieve 100% RFC 9535 compliance
- test: regression and benchmark tests

### Bug Fixes (from v2.1.0)
- fix: length() now counts Unicode runes instead of bytes
- fix: implement correct operator precedence (&& before ||)
- fix: recursive descent now includes root node
- fix: selectors return empty result on type mismatch instead of error
```

**Step 3: 更新README.md和README_zh.md**

更新README：
- 更新API示例
- 标注RFC 9535合规状态
- 标注扩展函数为非标准

**Step 4: Commit**

```bash
git add MIGRATION.md CHANGELOG.md README.md README_zh.md
git commit -m "docs: v3.0.0 migration guide and changelog"
```

---

## Task 4: 发布v3.0.0

**Files:**
- Modify: `version.go`

**Step 1: 更新版本号**

修改 `version.go`：

```go
package jsonpath

// Version is the current version of jsonpath
const Version = "3.0.0"

// VersionWithPrefix returns the version with v prefix
func VersionWithPrefix() string {
	return "v" + Version
}
```

**Step 2: 运行全量测试确认无回归**

```bash
go test ./...
```

**Step 3: 构建并验证版本**

```bash
go build ./cmd/jp && ./jp --version
```

Expected: 输出 `v3.0.0`

**Step 4: 打tag**

```bash
git tag v3.0.0
```

**Step 5: 创建GitHub Release**

```bash
gh release create v3.0.0 --title "v3.0.0" --notes-file CHANGELOG.md
```

**Step 6: Commit**

```bash
git add version.go
git commit -m "release: v3.0.0"
```

---

## 验证标准

- [ ] RFC 9535测试套件100%通过（324/324）
- [ ] `go test ./...` 全部通过
- [ ] `go test -bench=. -benchmem` 输出基准数据
- [ ] 性能不低于v2.0.2的90%
- [ ] `MIGRATION.md` 存在且包含所有breaking changes的迁移示例
- [ ] CHANGELOG包含完整变更列表
- [ ] README示例代码使用新API
- [ ] `go build ./cmd/jp && ./jp --version` 输出 `v3.0.0`
- [ ] `git tag v3.0.0` 存在
- [ ] GitHub Release包含Release Notes

---

## 执行选项

Plan complete and saved to `docs/plans/2026-05-06-phase-6-release.md`. Two execution options:

**1. Subagent-Driven (this session)** - I dispatch fresh subagent per task, review between tasks, fast iteration

**2. Parallel Session (separate)** - Open new session with executing-plans, batch execution with checkpoints

Which approach?
