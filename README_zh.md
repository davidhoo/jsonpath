# Go JSONPath

[![Go Reference](https://pkg.go.dev/badge/github.com/davidhoo/jsonpath.svg)](https://pkg.go.dev/github.com/davidhoo/jsonpath)
[![Go Report Card](https://goreportcard.com/badge/github.com/davidhoo/jsonpath)](https://goreportcard.com/report/github.com/davidhoo/jsonpath)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

[English](README.md)

完全符合 [RFC 9535](https://www.rfc-editor.org/rfc/rfc9535) 标准的 Go 语言 JSONPath 实现。提供命令行工具和 Go 库，支持所有标准的 JSONPath 功能。

## 特性

- 完整实现 RFC 9535 标准
  - 根节点访问（`$`）
  - 子节点访问（`.key` 或 `['key']`）
  - 递归查找（`..`）
  - 数组索引（`[0]`、`[-1]`）
  - 数组切片（`[start:end:step]`）
  - 数组通配符（`[*]`）
  - 多重索引（`[1,2,3]`）
  - 过滤表达式（`[?(@.price < 10)]`）
- 命令行工具（`jp`）
  - 精美的彩色输出
  - JSON 语法高亮
  - 支持文件和标准输入
  - 支持格式化和压缩输出
  - 友好的错误提示
  - 完整的 UTF-8 支持，正确显示中文
- 作为 Go 库使用
  - 简洁的 API 设计
  - 类型安全的操作
  - 丰富的示例
  - 详细的文档说明

## 新特性

### v1.0.4

- 集中管理版本号
  - 添加 version.go 用于集中版本控制
  - 更新 cmd/jp 使用集中管理的版本号
  - 修复中文注释的 UTF-8 编码问题

### v1.0.3

- 增强的过滤表达式
  - 完整支持逻辑运算符（`&&`、`||`、`!`）
  - 正确处理复杂的过滤条件
  - 支持 De Morgan 定律的否定表达式
  - 改进的数值和字符串比较
  - 更好的错误提示
- 改进的 API 设计
  - 新的简化 `Query` 函数，使用更方便
  - 弃用 `Compile/Execute` 而改用 `Query`
  - 更好的错误处理和报告
- 更新的示例
  - 新增逻辑运算符示例
  - 更新代码以使用新的 `Query` 函数
  - 修复示例中的 UTF-8 编码问题

### v1.0.2

- 增强的过滤表达式
  - 完整支持逻辑运算符（`&&`、`||`、`!`）
  - 正确处理复杂的过滤条件
  - 支持 De Morgan 定律的否定表达式
  - 改进的数值和字符串比较
  - 更好的错误提示
- 增强的彩色输出
  - 精美的 JSON 语法高亮
  - 彩色的命令行界面
  - 提升嵌套结构的可读性
- 更好的 UTF-8 支持
  - 修复中文字符显示
  - 正确处理多字节字符

## 安装方式

### Homebrew 安装推荐）

```bash
# 添加 tap
brew tap davidhoo/tap

# 安装 jsonpath
brew install jsonpath
```

### Go 模块安装

```bash
go install github.com/davidhoo/jsonpath/cmd/jp@latest
```

### 二进制安装

从 [releases 页面](https://github.com/davidhoo/jsonpath/releases) 下载适合你平台的二进制文件。

## 命令行使用方法

### 命令行参数

```bash
jp [-p <jsonpath表达式>] [-f <json文件>] [-c]
```

### 命令行选项

- `-p` JSONPath 表达式（如果不指定，输出整个 JSON）
- `-f` JSON 文件路径（如果不指定，从标准输入读取）
- `-c` 压缩输出（不格式化）
- `--no-color` 禁用彩色输出
- `-h` 显示帮助信息
- `-v` 显示版本信息

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
```

## Go 库使用方法

### Go 库基本用法

```go
import "github.com/davidhoo/jsonpath"

// 查询 JSON 数据
result, err := jsonpath.Query(data, "$.store.book[*].author")
if err != nil {
    log.Fatal(err)
}

// 处理结果
authors, ok := result.([]interface{})
if !ok {
    log.Fatal("unexpected result type")
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

    // 执行 JSONPath 查询
    result, err := jsonpath.Query(v, "$.store.book[?(@.price < 10)].title")
    if err != nil {
        log.Fatal(err)
    }

    // 打印结果
    fmt.Printf("%v\n", result) // ["Sayings of the Century"]
}
```

### 常用查询示例

```go
// 获取所有价格（递归）
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

// 获取价格大于 10 且类别为 fiction 的书籍
"$.store.book[?(@.price > 10 && @.category == 'fiction')]"

// 获取所有非参考类书籍
"$.store.book[?(!(@.category == 'reference'))]"

// 获取价格大于 10 或作者为 Evelyn 的书籍
"$.store.book[?(@.price > 10 || @.author == 'Evelyn Waugh')]"

// 获取书籍数组的长度
"$.store.book.length()"

// 获取商店对象的所有键
"$.store.keys()"

// 获取商店对象的所有值
"$.store.values()"

// 获取最低书价
"$.store.book[*].price.min()"

// 获取最高书价
"$.store.book[*].price.max()"

// 获取平均书价
"$.store.book[*].price.avg()"

// 获取所有书籍的总价
"$.store.book[*].price.sum()"

// 统计小说类书籍数量
"$.store.book[*].category.count('fiction')"

// 使用正则表达式匹配书名
"$.store.book[?@.title.match('^S.*')]"

// 搜索以 S 开头的书名
"$.store.book[*].title.search('^S.*')"

// 函数链式调用
"$.store.book[?@.price > 10].title.length()"

// 复杂的��数链式调用
"$.store.book[?@.price > $.store.book[*].price.avg()].title"

// 组合搜索和过滤条件
"$.store.book[?@.title.match('^S.*') && @.price < 10].author"
```

### 结果处理方法

根据结果类型使用类型断言处理结果：

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

## 实现细节

1. RFC 9535 标准合规性
   - 支持所有标准操作符
   - 标准兼容的语法解析
   - 标准的结果格式化

2. 过滤器支持
   - 比较操作符：`<`、`>`、`<=`、`>=`、`==`、`!=`
   - 逻辑运算符：`&&`、`||`、`!`
   - 支持复杂的过滤条件
   - 支持数值和字符串比较
   - 使用 De Morgan 定律处理否定表达式
   - 支持带括号的嵌套条件

3. 结果处理
   - 数组操作返回数组结果
   - 单值访问返回原始类型
   - 类型安全的结果处理

4. 错误处理
   - 详细的错误信息
   - 语法错误提示
   - 运行时错误处理

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License
