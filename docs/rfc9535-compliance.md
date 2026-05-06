# RFC 9535 合规性分析

**项目**: github.com/davidhoo/jsonpath v2.0.1
**分析日期**: 2026-05-06
**标准**: RFC 9535 — JSONPath: Query Expressions for JSON (February 2024)

---

## 一、已实现的功能

### 1. 基本语法

| RFC 9535 特性 | 项目实现 | 位置 |
|---|---|---|
| 根标识符 `$` | ✅ | `parser.go:11` |
| 子段 `.name` | ✅ | `segments.go:21` (nameSegment) |
| 括号表示法 `['key']` / `["key"]` | ✅ | `parser.go:145` |
| 递归下降 `..` | ✅ | `segments.go:234` (recursiveSegment) |
| 通配符 `.*` / `[*]` | ✅ | `segments.go:208` (wildcardSegment) |

### 2. 选择器

| RFC 9535 选择器 | 项目实现 | 位置 |
|---|---|---|
| 名称选择器 `'name'` | ✅ | nameSegment |
| 通配符选择器 `*` | ✅ | wildcardSegment |
| 索引选择器 `[0]` / `[-1]` | ✅ | `segments.go:178` (indexSegment) |
| 数组切片 `[start:end:step]` | ✅ | `segments.go:278` (sliceSegment) |
| 过滤器选择器 `[?(...)]` | ✅ | `segments.go:459` (filterSegment) |

### 3. 段类型

| RFC 9535 段 | 项目实现 | 位置 |
|---|---|---|
| 子段 `[sel1,sel2]` | ✅ | multiIndexSegment / multiNameSegment |
| 后代段 `..name` / `..*` | ✅ | recursiveSegment |

### 4. 过滤器表达式

| RFC 9535 功能 | 项目实现 | 位置 |
|---|---|---|
| 比较运算符 `==`, `!=`, `<`, `>`, `<=`, `>=` | ✅ | `parser.go:437`, `parser.go:489` |
| 逻辑与 `&&` | ✅ | `parser.go:938` |
| 逻辑或 `\|\|` | ✅ | `parser.go:938` |
| 逻辑非 `!` | ✅ | `parser.go:278-317` |
| 括号分组 | ✅ | `parser.go:298-302` |
| 当前节点 `@` | ✅ | filterCondition.field |
| 嵌套字段路径 `@.a.b.c` | ✅ | `parser.go:622` (getFieldValue) |
| 字符串字面量 `'...'` / `"..."` | ✅ | `parser.go:998` |
| 数字字面量 | ✅ | `parser.go:1003` |
| 布尔字面量 `true` / `false` | ✅ | `parser.go:990` |
| 空值字面量 `null` | ✅ | `parser.go:985` |
| De Morgan 定律转换 | ✅ | `parser.go:183` |

### 5. 标准函数逐项合规状态

| RFC 9535 函数 | 参数类型 | 当前实现 | 位置 | 合规状态 |
|---|---|---|---|---|
| `length()` — 字符串 | 应返回 Unicode 标量值数 | 返回字节数 `len(str)` | `functions.go:193` | ❌ 不合规 |
| `length()` — 数组 | 应返回元素数量 | 返回 `len(arr)` | `functions.go:189` | ✅ 合规 |
| `length()` — 对象 | 应返回成员数量 | 返回 `len(obj)` | `functions.go:199` | ✅ 合规 |
| `length()` — null/数字/布尔 | 应返回 Nothing | 返回 error | `functions.go:202` | ❌ 不合规 |
| `count()` | `NodesType → ValueType`（1个参数：节点列表） | `(array, value) → number`（2个参数：计数特定值） | `functions.go:263` | ❌ 签名完全不同 |
| `value()` | `NodesType → ValueType` | 未实现 | — | ❌ 未实现 |
| `match()` | `ValueType, ValueType → LogicalType` | 方法式调用 `.match()`，使用 Go regexp 部分匹配 | `functions.go:488` | ❌ 调用形式和语义均不同 |
| `search()` | `ValueType, ValueType → LogicalType` | 数组过滤，返回匹配元素数组 | `functions.go:593` | ❌ 签名和语义完全不同 |

---

## 二、未实现的功能

### 1. 核心缺失

| RFC 9535 功能 | 说明 |
|---|---|
| **`value()` 函数** | 标准要求的 5 个内置函数之一。将 NodesType 转为 ValueType，提取单节点值。当参数为空或多个节点时返回 Nothing。 |
| **过滤器中的存在性测试** | RFC 9535 允许 `[?@.isbn]` 作为存在性测试（字段存在即为真），当前实现的过滤器必须包含比较运算符。 |
| **过滤器中的 `$` (根引用)** | 标准允许 `[?@.price > $.threshold]`，在过滤器中引用根节点。当前不支持。 |
| **过滤器中的函数表达式** | 标准允许 `[?length(@.title) > 10]`、`[?count(@..item) > 0]`，函数作为过滤器条件的一部分。当前 `match` 是通过 `.match()` 语法作为运算符实现的，不是标准函数调用形式。 |
| **规范化路径 (Normalized Paths)** | RFC 9535 Section 2.7 要求每个节点有唯一的规范化路径标识，如 `$['store']['book'][0]`。项目未实现。 |
| **I-Regexp (RFC 9485)** | `match()` 和 `search()` 应使用 I-Regexp 语法（标准化的正则子集），当前直接使用 Go `regexp`。 |

### 2. 过滤器操作数类型缺失

| RFC 9535 功能 | 说明 |
|---|---|
| **过滤器中的 Singular Query** | 标准允许 `@.a`、`@['key']`、`@[0]` 作为比较操作数，不仅限于字面量。当前实现 `@.field` 只支持作为左侧操作数。 |
| **过滤器中的 Relative Query** | `@..key` 在过滤器中作为操作数（非 singular query），需要通过 `value()` 或 `count()` 函数转换。 |

---

## 三、与标准不符的实现

### 1. `length()` 函数 — 字符串计字节而非字符

| 方面 | RFC 9535 | 当前实现 |
|---|---|---|
| 字符串 | Unicode 标量值数量 | `len(str)` 返回 UTF-8 字节数 |
| 数组 | 元素数量 | `len(arr)` ✅ |
| 对象 | 成员数量 | `len(obj)` ✅ |
| null/数字/布尔 | Nothing | error |

**示例**: `length("日本語")` 应返回 3，当前返回 9（3 个字符 × 3 字节/字符）。应使用 `utf8.RuneCountInString(str)` 代替 `len(str)`。

### 2. `match()` 函数 — 部分匹配 vs 全串匹配（严重）

| 方面 | RFC 9535 | 当前实现 |
|---|---|---|
| **调用形式** | `match(@.field, "pattern")` — 标准函数调用，2个参数 | `@.field.match("pattern")` — 方法式调用 |
| **匹配方式** | **全串匹配** — pattern 必须匹配整个字符串 | `re.MatchString()` — **部分匹配**，匹配字符串中任意位置即可 |
| **返回类型** | `LogicalType`（用于过滤器逻辑） | `bool`（Go 布尔值） |
| **正则语法** | 必须使用 I-Regexp (RFC 9485) | 直接使用 Go `regexp` |
| **过滤器语法** | `$[?match(@.author, ".*")]` | `$[?@.author.match(".*")]` |

**关键问题**: 这是比 I-Regexp 兼容性更严重的语义差异。`[?match(@.name, "S")]` 在当前实现中会匹配任何包含 "S" 的字符串（如 "AliceSmith"），而 RFC 要求只匹配恰好为 "S" 的字符串。

RFC 9535 明确区分：
- `match()` = 全串匹配（pattern 必须匹配整个字符串）
- `search()` = 子串搜索（等价于 `.*pattern.*`）

当前实现中 `match()` 使用 `MatchString()` 实际上是搜索语义，等同于 RFC 的 `search()` 而非 `match()`。

代码证据 (`segments.go:556`):
```go
re, err := regexp.Compile(pattern)
return re.MatchString(str), nil
```

### 3. `search()` 函数 — 语义完全不同

| 方面 | RFC 9535 | 当前实现 |
|---|---|---|
| **调用形式** | `search(@.field, "pattern")` — 2个参数 | `$.items.search("pattern")` — 2个参数但语义不同 |
| **返回类型** | `LogicalType`（布尔值） | `[]interface{}`（返回匹配元素数组） |
| **功能** | 检查字符串中是否包含匹配子串 | 过滤数组中匹配正则的元素 |
| **正则语法** | I-Regexp | Go regexp |

### 4. `count()` 函数 — 签名完全不同

| 方面 | RFC 9535 | 当前实现 |
|---|---|---|
| **签名** | `NodesType → ValueType`（1个参数：节点列表） | `(array, value) → number`（2个参数） |
| **功能** | 计算节点列表中的节点数量 | 计算数组中某值出现的次数 |
| **示例** | `count(@..item)` → 返回匹配节点数 | `count([1,2,1], 1)` → 返回 2 |

### 5. 过滤器存在性测试 — 不支持

RFC 9535 定义了过滤器中的存在性测试（filter test expression）：

```
$[?@.isbn]       // 存在 isbn 字段即为真
$[?@..color]     // 递归查找到 color 即为真
```

当前实现要求过滤器条件必须包含比较运算符，不支持纯存在性测试。

### 6. 函数返回值类型系统 — 未实现

RFC 9535 定义了严格的函数类型系统：

| 类型 | 说明 | 当前实现 |
|---|---|---|
| `LogicalType` | 逻辑真/假，与 JSON `true`/`false` 不同 | ❌ 未区分，使用 Go `bool` |
| `ValueType` | JSON 值或 Nothing | ❌ 未实现 Nothing 概念 |
| `NodesType` | 节点列表 | ❌ 未区分节点列表和普通数组 |

标准中 `Nothing` 是一个重要概念：字段不存在时返回 `Nothing`，与 `null` 不同。当前实现将字段不存在等同于 `nil`。

### 7. 函数扩展注册机制 — 不符合标准

RFC 9535 定义了函数扩展的注册机制（Section 2.4.9），要求声明函数的参数类型和返回类型。当前的 `Function` 接口过于宽松：

```go
type Function interface {
    Call(args []interface{}) (interface{}, error)
    Name() string
}
```

没有类型约束声明，无法在调用前验证参数类型。

---

## 四、Error vs 空 Nodelist 问题（系统性缺陷）

RFC 9535 严格区分两种情况：
- **无效查询**（语法/语义错误）→ 必须 raise error
- **结构不匹配**（有效查询，数据不匹配）→ 返回空 nodelist，不是 error

当前实现使用 Go `error` 处理这两种情况，调用者无法区分"查询语法错误"和"没有匹配结果"。

| 场景 | RFC 9535 | 当前实现 | 位置 |
|---|---|---|---|
| Name selector 应用于非对象 | 空 nodelist | `fmt.Errorf("value is not an object")` | `segments.go:163` |
| Name selector 字段不存在 | 空 nodelist | `fmt.Errorf("field %s not found")` | `segments.go:168` |
| Wildcard 应用于原始类型 | 空 nodelist | `fmt.Errorf("value is neither array nor object")` | `segments.go:218` |
| Index 应用于非数组 | 空 nodelist | `fmt.Errorf("value is not an array")` | `segments.go:186` |
| Slice 应用于非数组 | 空 nodelist | `fmt.Errorf("value is not an array")` | `segments.go:421` |
| Multi-index 应用于非数组 | 空 nodelist | `fmt.Errorf("multi-index can only be applied to array")` | `segments.go:602` |
| 跨类型比较 | 返回 `false` | `fmt.Errorf("cannot compare %T and %T")` | `parser.go` (compareValues) |
| Boolean 有序比较 | 返回 `false` | error | `parser.go` (compareValues) |

---

## 五、其他实现缺陷

### 1. 递归下降未包含根节点

RFC 9535 Section 2.5.2 规定：后代段访问 D₁（输入节点本身）及其所有后代 D₂, ..., Dₙ，然后对每个节点应用子段选择器。

当前 `recursiveSegment` (`segments.go:237-276`) 只收集子节点和后代，不包含根节点本身。这意味着 `$..price` 如果根对象本身有 `price` 字段，会被遗漏。

### 2. `&&`/`||` 优先级错误

RFC 9535 Table 10 定义运算符优先级（从高到低）：

| 优先级 | 运算符 |
|---|---|
| 5 | 括号 `(...)`、函数调用 |
| 4 | 逻辑非 `!` |
| 3 | 比较 `== != < <= > >=` |
| 2 | 逻辑与 `&&` |
| 1 | 逻辑或 `\|\|` |

当前 `splitLogicalOperators` (`parser.go:900`) 按出现顺序从左到右分割，`evaluateConditions` (`segments.go:504`) 严格从左到右求值，不区分优先级。

**示例**: `@.a == 1 || @.b == 2 && @.c == 3`
- RFC 9535: `@.a == 1 || (@.b == 2 && @.c == 3)` — `&&` 优先
- 当前实现: `(@.a == 1 || @.b == 2) && @.c == 3` — 从左到右

### 3. 过滤器只评估 `map[string]interface{}` 元素

`filterSegment.evaluate` (`segments.go:486-497`) 在遍历数组时，只处理 `map[string]interface{}` 类型的元素：

```go
for _, item := range arr {
    if m, ok := item.(map[string]interface{}); ok {
        result, err := s.evaluateConditions(m)
```

RFC 9535 要求过滤器评估**所有**数组元素，包括原始类型。例如 `[1, 5, 10][?@ > 5]` 应返回 `[10]`。

### 4. `@` 单独引用不支持

RFC 9535 允许 `[?@ > 5]` 对当前节点本身进行判断。当前 `getFieldValue` (`parser.go:622`) 将字段名 `"@"` 作为 map 的键查找，完全错误。`@` 单独使用时应该直接返回当前节点的值。

### 5. 字符串转义序列未实现

RFC 9535 要求名称选择器和字符串字面量支持转义序列：`\b \f \n \r \t \" \' \/ \\ \uXXXX`。

当前 `parseFilterValue` (`parser.go:998`) 和 `parseMultiIndexSegment` 只做简单的引号剥离，不处理转义序列。`'\n'` 会被解析为字面的 `\` + `n`，而非换行符。

### 6. 过滤器字段引用不支持完整子查询

`getFieldValue` (`parser.go:622`) 只支持简单的点分隔路径 `@.user.address.city`，不支持：
- `@.items[0]` — 数组索引
- `@['key']` — 括号表示法
- `@..key` — 递归下降

### 7. 递归下降后对原始类型应用选择器会报错

`$..price` 遍历到字符串、数字等原始值时，后续的 `nameSegment.evaluate` 返回 `fmt.Errorf("value is not an object")`，导致整个查询失败。RFC 要求不匹配的节点应被跳过（返回空 nodelist），不应报错。

### 8. Slice `start=0`/`end=0` 哨兵值歧义 — `[:]` 与 `[0:0]` 不可区分

`parseSliceSegment` (`parser.go:793`) 用 `start: 0, end: 0` 作为"未指定"的默认值。这导致用户显式写 `[0:0]` 时，解析器无法区分"未指定"和"显式指定索引 0"两种意图。

在 `normalizeRange` (`segments.go:306`) 中：

```go
start = s.start
if start == 0 {
    if step > 0 { start = 0 }
    else { start = length - 1 }
}
```

具体影响：
- `[:]`（未指定）和 `[0:0]`（显式）在解析后结构相同，语义却应不同
- 对于负步长，`[0:3:-1]` 的 `start=0` 会被误判为"未指定"而替换为 `length-1`，但用户实际要求从索引 0 开始
- `end=0` 存在同样的问题：`[3:0]` 的 `end=0` 被当作"未指定"处理为 `length`

### 9. 索引选择器允许前导零 — 应在解析阶段拒绝

RFC 9535 ABNF 定义 `index-selector = int`，其中 `int = "0" / (["-"] DIGIT1 *DIGIT)`，`DIGIT1 = %x31-39`（即 1-9）。这意味着 `[07]` 违反 ABNF 语法，是 **well-formedness violation**（格式错误），应在**解析阶段**就拒绝，而非运行时默默接受。

当前使用 `strconv.Atoi` 解析，`07` 被当作十进制 `7` 接受，没有报错。

### 10. 裸 `..` 被接受

RFC 9535 ABNF 定义 `descendant-segment = ".." (bracketed-selection / wildcard-selector / member-name-shorthand)`，并明确说明 `..` 本身不是有效段。当前 `parseRecursive` (`segments.go`) 接受裸 `$..` 且不报错。

---

## 六、扩展功能（超出 RFC 9535）

| 函数 | 签名 | 说明 |
|---|---|---|
| `keys()` | `object → array` | 返回对象键的排序数组 |
| `values()` | `object → array` | 返回对象值的排序数组 |
| `min()` | `array → number` | 返回数组最小值 |
| `max()` | `array → number` | 返回数组最大值 |
| `avg()` | `array → number` | 返回数组平均值 |
| `sum()` | `array → number` | 返回数组总和 |
| `count(array, value)` | 计数特定值 | 非标准签名，标准 `count` 接受 NodesType |

---

## 七、架构级根因分析

三个根本性架构缺陷阻碍了渐进式 RFC 合规：

### 7.1 缺少 Node 类型

RFC 9535 的核心数据模型是 `Node = (location, value)`。所有段返回 `Nodelist`（有序 Node 列表）。当前实现返回裸 `interface{}` 值，丢失了位置信息。这阻塞了：
- 规范化路径生成（§2.7）
- `count(NodesType)` 函数（§2.4.5）
- `value(NodesType)` 函数（§2.4.8）
- 正确的 nodelist 语义

### 7.2 Error vs 空 Nodelist 混淆

RFC 9535 区分"无效查询"和"结构不匹配"，当前都用 Go `error`。这影响所有选择器在类型不匹配时的行为。

### 7.3 函数签名分歧

`count` 和 `search` 的签名与 RFC 9535 完全不同。修复会破坏现有用户。选项：
- 重命名为非标准名称（如 `occurrences`、`filterMatch`）
- 旁边添加 RFC 合规版本（易混淆）
- 主版本升级做破坏性变更

---

## 八、对 `RFC9535_COMPLIANCE_REPORT.md` 的勘误

项目根目录下已有的 `RFC9535_COMPLIANCE_REPORT.md` 存在以下不准确之处：

### 8.1 `length()` 对数组/对象的行为 — 报告遗漏合规点

报告函数表中 `length()` 只列了两行：String（按字符计 vs 按字节计）和 "Other types: return Nothing"（返回 error）。数组（Array）和对象（Object）未单独列出。

RFC 9535 Section 2.4.4 明确规定：

> If the argument value is an object, the result is the number of members in the object.
> If the argument value is an array, the result is the number of elements in the array.

当前实现 `float64(len(arr))` 和 `float64(len(obj))` 均符合标准。报告的 "Other types" 本意可能指 null/number/boolean，但措辞有歧义，且章节标题 "all 5 have issues" 给读者造成 `length()` 全面不合规的印象。应补充明确标注 Array 和 Object 是合规的。

### 8.2 `match()` 函数 — 遗漏关键语义差异

报告只提到 I-Regexp vs Go regexp 的语法差异，**遗漏了更根本的语义差异**：
- RFC 9535 `match()` 要求**全串匹配**
- 当前 `re.MatchString()` 是**部分匹配**

这意味着 `[?match(@.name, "S")]` 在当前实现中匹配任何包含 "S" 的字符串，而 RFC 只匹配恰好为 "S" 的字符串。这个差异比 I-Regexp 兼容性问题严重得多，严重度应为 CRITICAL 而非 MEDIUM。

### 8.3 `search()` 函数 — 遗漏全串 vs 子串差异

报告只提到签名不同，未指出 RFC `search()` 是字符串级别的子串搜索操作，当前实现是数组级别的过滤操作，语义完全不同。

### 8.4 Slice `step=0` 严重度

报告标记为 HIGH。RFC 规定 `step=0` 应返回空 nodelist 而非报错。当前实现在解析阶段拒绝。这不会产生错误数据，只是错误处理方式不同，MEDIUM 可能更准确。

### 8.5 报告完全遗漏的问题

| 问题 | 说明 |
|---|---|
| `@` 单独引用不支持 | `[?@ > 5]` 对纯数值数组过滤，`getFieldValue` 会在 map 中查找 `"@"` 键 |
| Slice `end=0` 哨兵值歧义 | `[0:0]` 的 `end=0` 被当作"未指定"处理（已在第五章第 8 点详细说明） |
| 递归下降后原始类型报错 | `$..price` 遍历到原始值时 `nameSegment` 返回 error 导致查询失败 |

---

## 九、README 示例合规性分析

README 中列出的 19 个示例的 RFC 9535 合规性：

| # | 示例 | 状态 | 问题 |
|---|------|------|------|
| 1 | `$.store.book[*].author` | ✅ 合规 | |
| 2 | `$.store.book[?(@.price > 10)]` | ✅ 合规 | |
| 3 | `$.store.book[0]` / `[-1]` | ✅ 合规 | |
| 4 | `$.store.book[0:2]` | ✅ 合规 | |
| 5 | `$..price` | ⚠️ 有缺陷 | 递归下降未包含根节点本身 |
| 6 | `$.store.book[?(@.price > 10 && @.category == 'fiction')]` | ✅ 合规 | 仅使用 `&&`，优先级问题不在此例体现 |
| 7 | `$.store.book[?(!(@.category == 'reference'))]` | ✅ 合规 | |
| 8 | `$.store.book.length()` | ⚠️ 有缺陷 | `length()` 计字节数非字符数 |
| 9 | `$.store.keys()` | ➕ 非标准 | 扩展函数，应标注为非标准 |
| 10 | `$.store.values()` | ➕ 非标准 | 扩展函数，应标注为非标准 |
| 11 | `$.store.book[*].price.min()` | ➕ 非标准 | 扩展函数，应标注为非标准 |
| 12 | `$.store.book[*].price.max()` | ➕ 非标准 | 扩展函数，应标注为非标准 |
| 13 | `$.store.book[*].price.avg()` | ➕ 非标准 | 扩展函数，应标注为非标准 |
| 14 | `$.store.book[*].price.sum()` | ➕ 非标准 | 扩展函数，应标注为非标准 |
| 15 | `$.store.book[*].category.count('fiction')` | ❌ 签名错误 | RFC `count(NodesType)->ValueType`，非 `(array, value)` |
| 16 | `$.store.book[?@.title.match('^S.*')]` | ❌ 不支持标准语法 | 实现**不支持** RFC 标准函数调用语法 `match(@.title, '^S.*')`，仅支持方法链 `.match()` |
| 17 | `$.store.book[*].title.search('^S.*')` | ❌ 签名错误 | RFC `search(ValueType,ValueType)->LogicalType`（字符串+正则→布尔），非数组过滤 |
| 18 | `$.store.book[*]['author','price']` | ✅ 合规 | |
| 19 | `$.store.book[?@.price > $.store.book[*].price.avg()]` | ❌ 不支持 | 过滤器不支持 `$` 根引用查询 |

**统计**: 8 合规，3 有缺陷，3 非标准（应标注），4 不合规/签名错误/不支持

---

## 十、总结

| 类别 | 数量 | 说明 |
|---|---|---|
| ✅ 完全符合标准 | 7 | `$` 根标识符、比较运算符、负索引、越界索引返回空、数组顺序保持、括号点表示法、基本递归/切片/过滤 |
| ⚠️ 部分合规 | 9 | `length()`(字符串/其他类型有问题)、`match()`(I-Regexp/全串匹配)、递归下降(缺根节点)、`&&`/`||`(优先级)、过滤器(只处理 map)、字段引用(只支持点路径)、De Morgan(复杂表达式脆弱)、slice(哨兵值歧义)、bare `..`(接受) |
| ❌ 不合规 | 24 | 见上述各章详细列表 |
| ➕ 超出标准的扩展 | 6 | `keys`/`values`/`min`/`max`/`avg`/`sum` |

### 最大的合规性差距

1. **函数调用语法**：标准使用 `func(args)` 形式，项目使用 `.func(args)` 方法式调用
2. **`match()` 部分/全串匹配**：`MatchString()` 是搜索语义，RFC `match()` 要求全串匹配 — 这是比 I-Regexp 兼容性更根本的差异
3. **类型系统**：`LogicalType`/`ValueType`/`NodesType` + `Nothing` 概念完全缺失
4. **Error vs 空 Nodelist**：系统性地将结构不匹配当作错误处理
5. **`&&`/`||` 优先级**：从左到右求值，`&&` 不优先于 `||`
6. **I-Regexp**：正则表达式未遵循 RFC 9485 规范
