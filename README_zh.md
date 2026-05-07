# Go JSONPath

[![Go Reference](https://pkg.go.dev/badge/github.com/davidhoo/jsonpath.svg)](https://pkg.go.dev/github.com/davidhoo/jsonpath)
[![Go Report Card](https://goreportcard.com/badge/github.com/davidhoo/jsonpath)](https://goreportcard.com/report/github.com/davidhoo/jsonpath)
[![Coverage Status](https://coveralls.io/repos/github/davidhoo/jsonpath/badge.svg?branch=main)](https://coveralls.io/github/davidhoo/jsonpath?branch=main)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

[English](README.md)

完全符合 [RFC 9535](https://www.rfc-editor.org/rfc/rfc9535) 标准的 Go 语言 JSONPath 实现。提供命令行工具和 Go 库，支持所有标准的 JSONPath 功能。

## 特性

- **RFC 9535 完全合规** - 通过全部 703 项合规测试
- **完整的 JSONPath 支持**
  - 根节点访问（`$`）
  - 子节点访问（`.key` 或 `['key']`）
  - 递归查找（`..`）
  - 数组索引（`[0]`、`[-1]`）
  - 数组切片（`[start:end:step]`）
  - 数组通配符（`[*]`）
  - 多重索引（`[1,2,3]`）
  - 多字段名称（`['name','age']`）
  - 过滤表达式（`[?(@.price < 10)]`）
  - 存在性测试（`[?@.name]`）
  - 函数调用（`length()`、`count()`、`match()`、`search()`、`value()`）
- **命令行工具（`jp`）**
  - 精美的彩色输出
  - JSON 语法高亮
  - 支持文件和标准输入
  - 支持格式化和压缩输出
  - 规范化路径输出（`--path`）
- **Go 库**
  - 返回 `NodeList` 的简洁 API
  - 每个结果节点包含规范化路径
  - 类型安全的操作
  - 详细的文档说明

## 安装方式

### Homebrew 安装（推荐）

```bash
brew tap davidhoo/tap
brew install jsonpath
```

### Go 模块安装

```bash
go install github.com/davidhoo/jsonpath/cmd/jp@latest
```

### 二进制安装

从 [releases 页面](https://github.com/davidhoo/jsonpath/releases) 下载适合你平台的二进制文件。

## 命令行使用方法

```bash
jp [-p <jsonpath表达式>] [-f <json文件>] [-c] [--no-color] [--path]
```

### 命令行选项

| 选项 | 说明 |
|------|------|
| `-p` | JSONPath 表达式（如果不指定，输出整个 JSON） |
| `-f` | JSON 文件路径（如果不指定，从标准输入读取） |
| `-c` | 压缩输出（不格式化） |
| `--no-color` | 禁用彩色输出 |
| `--path` | 输出规范化路径和值 |
| `-h` | 显示帮助信息 |
| `-v` | 显示版本信息 |

### 命令行示例

```bash
# 输出整个 JSON，带语法高亮
jp -f data.json

# 查询特定路径
jp -f data.json -p '$.store.book[*].author'

# 带条件的过滤
jp -f data.json -p '$.store.book[?(@.price > 10)]'

# 从标准输入读取
echo '{"name": "John"}' | jp -p '$.name'

# 压缩输出
jp -f data.json -c

# 显示规范化路径
echo '{"a":1,"b":2}' | jp --path '$.*'
# 输出:
# $['a'] 1
# $['b'] 2
```

## Go 库使用方法

### 基本用法

```go
import "github.com/davidhoo/jsonpath"

// 查询 JSON 数据 - 返回 NodeList
result, err := jsonpath.Query(data, "$.store.book[*].author")
if err != nil {
    log.Fatal(err)
}

// 处理结果（NodeList）
for _, node := range result {
    fmt.Printf("位置: %s, 值: %v\n", node.Location, node.Value)
}
```

### 完整代码示例

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "github.com/davidhoo/jsonpath"
)

func main() {
    data := `{
        "store": {
            "book": [
                {"category": "reference", "author": "Nigel Rees", "title": "Sayings of the Century", "price": 8.95},
                {"category": "fiction", "author": "Evelyn Waugh", "title": "Sword of Honour", "price": 12.99}
            ]
        }
    }`

    var v interface{}
    if err := json.Unmarshal([]byte(data), &v); err != nil {
        log.Fatal(err)
    }

    result, err := jsonpath.Query(v, "$.store.book[?(@.price < 10)].title")
    if err != nil {
        log.Fatal(err)
    }

    for _, node := range result {
        fmt.Println(node.Value) // Sayings of the Century
    }
}
```

### 常用查询示例

```go
// 获取所有价格（递归查找）
"$..price"

// 获取特定价格范围的书籍
"$.store.book[?(@.price < 10)].title"

// 获取所有作者
"$.store.book[*].author"

// 获取第一/最后一本书
"$.store.book[0]"
"$.store.book[-1]"

// 获取前两本书
"$.store.book[0:2]"

// 复杂过滤条件
"$.store.book[?(@.price > 10 && @.category == 'fiction')]"

// 存在性测试
"$[?@.name]"

// 函数调用（RFC 9535）
"$.store.book[?match(@.title, '^S.*')]"
"$.store.book[?search(@.title, 'Century')]"
"$[?count(@..*) > 5]"

// 非标准扩展
"$.store.book[*].price.min()"
"$.store.book[*].price.max()"
"$.store.book[*].price.avg()"
"$.store.book[*].price.sum()"
```

### 结果处理

`Query()` 返回 `NodeList`（`Node` 切片）。每个 `Node` 包含：
- `Location` - 规范化路径（例如 `$['store']['book'][0]`）
- `Value` - 实际值

```go
for _, node := range result {
    fmt.Printf("位置: %s\n", node.Location)
    fmt.Printf("值: %v\n", node.Value)
}

// 访问第一个结果
if len(result) > 0 {
    firstValue := result[0].Value
}
```

## RFC 9535 合规性

本实现完全符合 [RFC 9535](https://www.rfc-editor.org/rfc/rfc9535) 标准：

- **100% 通过率** - 官方合规测试套件全部通过（703/703）
- 所有标准选择器（名称、索引、切片、通配符、过滤、递归查找、联合）
- 所有标准函数（`length`、`count`、`match`、`search`、`value`）
- I-Regexp 模式匹配（RFC 9485）
- 规范化路径生成
- 过滤表达式中的三值逻辑

详见 [RFC9535_COMPLIANCE_REPORT.md](docs/RFC9535_COMPLIANCE_REPORT.md)。

## 非标准扩展

以下函数是为方便使用而添加的非标准扩展：

| 函数 | 说明 |
|------|------|
| `keys()` | 返回对象的排序键列表 |
| `values()` | 返回对象的值列表 |
| `min()` | 返回数组中的最小值 |
| `max()` | 返回数组中的最大值 |
| `avg()` | 返回数值的平均值 |
| `sum()` | 返回数值的总和 |
| `occurrences()` | 统计值在数组中出现的次数 |

## 测试

```bash
# 运行所有测试
go test -v ./...

# 运行测试并生成覆盖率报告
go test -v -coverprofile=coverage.out ./...

# 运行基准测试
go test -bench=. -benchmem ./...

# 运行 RFC 9535 合规测试
go test -run TestCTS -v
```

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License

## 更新日志

详见 [CHANGELOG.md](CHANGELOG.md)。
