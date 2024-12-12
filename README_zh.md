# Go JSONPath

[![Go Reference](https://pkg.go.dev/badge/github.com/davidhoo/jsonpath.svg)](https://pkg.go.dev/github.com/davidhoo/jsonpath)
[![Go Report Card](https://goreportcard.com/badge/github.com/davidhoo/jsonpath)](https://goreportcard.com/report/github.com/davidhoo/jsonpath)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

[English](README.md)

一个完整的 Go 语言 JSONPath 实现，完全遵循 [RFC 9535](https://www.rfc-editor.org/rfc/rfc9535) 规范。提供了命令行工具和 Go 程序库，支持所有 JSONPath 标准特性。

## 特性

- 完整实现 RFC 9535 规范
  - 根节点访问 (`$`)
  - 子节点访问 (`.key` 或 `['key']`)
  - 递归下降 (`..`)
  - 数组索引 (`[0]`, `[-1]`)
  - 数组切片 (`[start:end:step]`)
  - 数组通配符 (`[*]`)
  - 多索引选择 (`[1,2,3]`)
  - 过滤表达式 (`[?(@.price < 10)]`)
- 提供命令行工具 (`jp`)
  - 彩色输出，提升可读性
  - 支持从文件或标准输入读取
  - 支持格式化和压缩输出
  - 友好的错误提示
- 作为 Go 库使用
  - 简洁的 API 设计
  - 完整���类型安全
  - 丰富的示例代码
  - 详细的文档说明

## 安装

### Homebrew（推荐）

```bash
# 添加 tap
brew tap davidhoo/tap

# 安装 jsonpath
brew install jsonpath
```

### Go 安装

```bash
go install github.com/davidhoo/jsonpath/cmd/jp@latest
```

### 手动安装

从[发布页面](https://github.com/davidhoo/jsonpath/releases)下载适合您平台的二进制文件。

## 命令行使用

### 基本用法

```bash
jp [-p <jsonpath_expression>] [-f <json_file>] [-c]
```

参数说明：
- `-p` JSONPath 表达式（如果不指定，则输出完整的 JSON）
- `-f` JSON 文件路径（如果不指定，则从标准输入读取）
- `-c` 压缩输出（不格式化）
- `-h` 显示帮助信息
- `-v` 显示版本信息

### 示例

```bash
# 输出完整的 JSON
jp -f data.json

# 查询特定路径
jp -f data.json -p '$.store.book[*].author'

# 从标准输入读取
echo '{"name": "John"}' | jp -p '$.name'

# 压缩输出
jp -f data.json -c
```

## 在 Go 程序中使用

### 基本用法

```go
import "github.com/davidhoo/jsonpath"

// 编译 JSONPath 表达式
jp, err := jsonpath.Compile("$.store.book[*].author")
if err != nil {
    log.Fatal(err)
}

// 执行查询
result, err := jp.Execute(data)
if err != nil {
    log.Fatal(err)
}

// 处��结果
authors, ok := result.([]interface{})
if !ok {
    log.Fatal("结果类型不匹配")
}
```

### 完整示例

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "github.com/davidhoo/jsonpath"
)

func main() {
    // JSON 数据
    data := `{
        "store": {
            "book": [
                {
                    "category": "reference",
                    "author": "Nigel Rees",
                    "title": "Sayings of the Century",
                    "price": 8.95
                },
                {
                    "category": "fiction",
                    "author": "Evelyn Waugh",
                    "title": "Sword of Honour",
                    "price": 12.99
                }
            ]
        }
    }`

    // 解析 JSON
    var v interface{}
    if err := json.Unmarshal([]byte(data), &v); err != nil {
        log.Fatal(err)
    }

    // 编译并执行 JSONPath
    jp, err := jsonpath.Compile("$.store.book[?(@.price < 10)].title")
    if err != nil {
        log.Fatal(err)
    }

    result, err := jp.Execute(v)
    if err != nil {
        log.Fatal(err)
    }

    // 输出结果
    fmt.Printf("%v\n", result) // ["Sayings of the Century"]
}
```

### 常见查询示例

```go
// 获取所有价格（递归查找）
"$..price"

// 获取特定价格范围的书籍
"$.store.book[?(@.price < 10)].title"

// 获取所有作者
"$.store.book[*].author"

// 获取第一本书
"$.store.book[0]"

// 获取最后一本书
"$.store.book[-1]"

// 获取前两本书
"$.store.book[0:2]"

// 获取所有价格大于 10 的书籍
"$.store.book[?(@.price > 10)]"
```

### 结果处理

根查询结果的类型，需要进行相应的类型断言：

```go
// 单个值结果
if str, ok := result.(string); ok {
    // 处理字符串结果
}

// 数组结果
if arr, ok := result.([]interface{}); ok {
    for _, item := range arr {
        // 处理每个元素
    }
}

// 对象结果
if obj, ok := result.(map[string]interface{}); ok {
    // 处理对象
}
```

## 实现说明

1. 完全遵循 RFC 9535 规范
   - 支持所有标准操作符
   - 符合标准的语法解析
   - 标准的结果格式

2. 过滤器支持
   - 比较操作符: `<`, `>`, `<=`, `>=`, `==`, `!=`
   - 目前支持数值比较
   - 未来将支持字符串比较和逻辑操作符

3. 结果处理
   - 数组操作返回数组结果
   - 单个值访问返回原始类型
   - 支持类型安全的结果处理

4. 错误处理
   - 详细的错误信息
   - 语法错误提示
   - 运行时错误处理

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License 