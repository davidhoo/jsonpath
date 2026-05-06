# Phase 0: 测试基础设施实施计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 集成RFC 9535官方测试套件，建立当前通过率基线

**Architecture:** 从RFC 9535参考实现获取测试数据JSON文件，创建测试数据解析器和测试运行器，建立基线通过率

**Tech Stack:** Go, testing, JSON parsing

---

## Task 1: 获取RFC 9535测试套件数据

**Files:**
- Create: `testdata/rfc9535/` 目录
- Create: `testdata/rfc9535/testdata.json` (测试数据文件)

**Step 1: 创建测试数据目录**

```bash
mkdir -p testdata/rfc9535
```

**Step 2: 从RFC 9535参考实现获取测试数据**

从cburgmer/jsonpath-comparison项目获取测试数据：

```bash
curl -L https://github.com/cburgmer/jsonpath-comparison/raw/master/regression_suite/regression_suite.yaml -o testdata/rfc9535/regression_suite.yaml
```

或者从IETF RFC 9535仓库获取测试数据：

```bash
curl -L https://www.rfc-editor.org/rfc/rfc9535.txt -o testdata/rfc9535/rfc9535.txt
```

**Step 3: 验证测试数据文件存在**

```bash
ls -la testdata/rfc9535/
```

Expected: 目录存在且包含测试数据文件

**Step 4: Commit**

```bash
git add testdata/rfc9535/
git commit -m "test: add RFC 9535 test suite data"
```

---

## Task 2: 创建测试数据解析器

**Files:**
- Create: `rfc9535_test.go`
- Modify: `go.mod` (如果需要添加依赖)

**Step 1: 创建测试数据解析器结构**

在 `rfc9535_test.go` 中定义测试用例结构：

```go
package jsonpath

import (
	"encoding/json"
	"os"
	"testing"
)

// RFC9535TestCase 表示RFC 9535测试用例
type RFC9535TestCase struct {
	Name     string      `json:"name"`
	Selector string      `json:"selector"`
	Document interface{} `json:"document"`
	Expected interface{} `json:"expected"`
	Invalid  bool        `json:"invalid,omitempty"`
}

// RFC9535TestSuite 表示RFC 9535测试套件
type RFC9535TestSuite struct {
	Tests []RFC9535TestCase `json:"tests"`
}

// loadRFC9535TestSuite 加载RFC 9535测试套件
func loadRFC9535TestSuite(t *testing.T) *RFC9535TestSuite {
	t.Helper()
	
	// 尝试从不同来源加载测试数据
	testDataPaths := []string{
		"testdata/rfc9535/testdata.json",
		"testdata/rfc9535/regression_suite.yaml",
	}
	
	for _, path := range testDataPaths {
		if _, err := os.Stat(path); err == nil {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("Failed to read test data: %v", err)
			}
			
			var suite RFC9535TestSuite
			if err := json.Unmarshal(data, &suite); err != nil {
				t.Fatalf("Failed to parse test data: %v", err)
			}
			
			return &suite
		}
	}
	
	t.Skip("RFC 9535 test data not found, skipping tests")
	return nil
}
```

**Step 2: 验证解析器能加载测试数据**

```bash
go test -run TestRFC9535Suite_Parse -v
```

Expected: 测试被跳过（因为测试数据文件不存在）

**Step 3: Commit**

```bash
git add rfc9535_test.go
git commit -m "test: add RFC 9535 test data parser"
```

---

## Task 3: 创建测试运行器并建立基线

**Files:**
- Modify: `rfc9535_test.go`

**Step 1: 实现通用测试运行器**

在 `rfc9535_test.go` 中添加测试运行器：

```go
// TestRFC9535Suite 运行RFC 9535测试套件
func TestRFC9535Suite(t *testing.T) {
	suite := loadRFC9535TestSuite(t)
	if suite == nil {
		return
	}
	
	passCount := 0
	failCount := 0
	skipCount := 0
	
	for _, test := range suite.Tests {
		t.Run(test.Name, func(t *testing.T) {
			if test.Invalid {
				// 测试无效选择器应该返回错误
				_, err := Query(test.Document, test.Selector)
				if err == nil {
					t.Errorf("Expected error for invalid selector %q, got nil", test.Selector)
					failCount++
				} else {
					passCount++
				}
				return
			}
			
			// 执行查询
			result, err := Query(test.Document, test.Selector)
			if err != nil {
				t.Errorf("Unexpected error for selector %q: %v", test.Selector, err)
				failCount++
				return
			}
			
			// 比较结果
			expectedJSON, _ := json.Marshal(test.Expected)
			resultJSON, _ := json.Marshal(result)
			
			if string(expectedJSON) != string(resultJSON) {
				t.Errorf("Selector %q: expected %s, got %s", test.Selector, expectedJSON, resultJSON)
				failCount++
			} else {
				passCount++
			}
		})
	}
	
	// 输出基线报告
	t.Logf("RFC 9535 Test Suite Results:")
	t.Logf("PASS: %d/%d", passCount, len(suite.Tests))
	t.Logf("FAIL: %d/%d", failCount, len(suite.Tests))
	t.Logf("SKIP: %d/%d", skipCount, len(suite.Tests))
	
	// 将基线数据写入文件
	baselineData := fmt.Sprintf("PASS: %d/%d\nFAIL: %d/%d\nSKIP: %d/%d\n",
		passCount, len(suite.Tests),
		failCount, len(suite.Tests),
		skipCount, len(suite.Tests))
	
	if err := os.WriteFile("testdata/rfc9535/baseline.txt", []byte(baselineData), 0644); err != nil {
		t.Fatalf("Failed to write baseline data: %v", err)
	}
}
```

**Step 2: 添加必要的导入**

在文件顶部添加：

```go
import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)
```

**Step 3: 运行测试套件建立基线**

```bash
go test -run TestRFC9535Suite -v 2>&1 | tail -20
```

Expected: 输出包含通过率统计

**Step 4: 验证基线文件已创建**

```bash
cat testdata/rfc9535/baseline.txt
```

Expected: 基线数据文件存在且包含统计信息

**Step 5: Commit**

```bash
git add rfc9535_test.go testdata/rfc9535/baseline.txt
git commit -m "test: integrate RFC 9535 test suite and establish baseline"
```

---

## 验证标准

- [ ] `testdata/rfc9535/` 目录存在且包含测试JSON文件
- [ ] `go test -run TestRFC9535Suite_Parse -v` 能成功解析所有测试用例，不panic
- [ ] 解析出的用例数量 ≥ 300
- [ ] `go test -run TestRFC9535Suite -v 2>&1 | tail -20` 输出包含通过率统计
- [ ] 输出格式类似：`PASS: 120/324, FAIL: 180/324, SKIP: 24/324`
- [ ] 将基线数据写入 `testdata/rfc9535/baseline.txt`
- [ ] commit：`test: integrate RFC 9535 test suite and establish baseline`

---

## 执行选项

Plan complete and saved to `docs/plans/2026-05-06-phase-0-test-infrastructure.md`. Two execution options:

**1. Subagent-Driven (this session)** - I dispatch fresh subagent per task, review between tasks, fast iteration

**2. Parallel Session (separate)** - Open new session with executing-plans, batch execution with checkpoints

Which approach?
