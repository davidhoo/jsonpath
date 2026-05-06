# 开发计划

**项目：** jsonpath  
**依据：** `docs/PRD.md`、`docs/rfc9535-compliance-impact-analysis.md`  
**原则：** 每步可独立验证，先写测试再实现，每完成一步即 commit

---

## 阶段划分

| 阶段 | 版本 | 内容 | 前置条件 |
|------|------|------|----------|
| Phase 0 | — | 测试基础设施 | 无 |
| Phase 1 | v2.1.0 | 行为修正（bug fix） | Phase 0 |
| Phase 2 | v2.1.0 | 新增语法 | Phase 1 |
| Phase 3 | v3.0.0 | 类型系统基础 | Phase 2 |
| Phase 4 | v3.0.0 | 内部重构 | Phase 3 |
| Phase 5 | v3.0.0 | 公共 API 变更 | Phase 4 |
| Phase 6 | v3.0.0 | 测试通过 + 发布 | Phase 5 |

---

## Phase 0：测试基础设施

**目标：** 集成 RFC 9535 官方测试套件，建立当前通过率基线。

### Step 0.1：获取并集成 RFC 9535 测试套件

**做什么：**
1. 从 RFC 9535 参考实现（如 `cburgmer/jsonpath-comparison` 或 IETF 仓库）获取测试数据 JSON 文件
2. 将测试数据放入 `testdata/rfc9535/` 目录
3. 编写测试数据解析器 `rfc9535_test.go`，能读取 JSON 并转为 Go 测试结构

**验证标准：**
- [ ] `testdata/rfc9535/` 目录存在且包含测试 JSON 文件
- [ ] `go test -run TestRFC9535Suite_Parse -v` 能成功解析所有测试用例，不 panic
- [ ] 解析出的用例数量 ≥ 300

---

### Step 0.2：编写测试运行器并建立基线

**做什么：**
1. 在 `rfc9535_test.go` 中实现通用测试运行器，逐个执行测试用例
2. 对每个用例调用 `Query()`，比较实际结果与期望结果
3. 记录通过/失败/跳过的数量，输出基线报告

**验证标准：**
- [ ] `go test -run TestRFC9535Suite -v 2>&1 | tail -20` 输出包含通过率统计
- [ ] 输出格式类似：`PASS: 120/324, FAIL: 180/324, SKIP: 24/324`
- [ ] 将基线数据写入 `testdata/rfc9535/baseline.txt`
- [ ] commit：`test: integrate RFC 9535 test suite and establish baseline`

---

## Phase 1：v2.1.0 行为修正

**目标：** 修复 4 个已知行为错误，不改变任何公共 API 签名。

### Step 1.1：修复 `length()` 字符串计算

**做什么：**
1. 在 `functions_test.go` 中添加测试：
   - `length("日本語")` 期望 `3`
   - `length("hello")` 期望 `5`
   - `length("café")` 期望 `4`
   - `length("𝄞")` 期望 `1`（4 字节 Unicode 字符）
2. 运行测试确认失败
3. 修改 `functions.go` 中 `length()` 对字符串的处理：`len(s)` → `utf8.RuneCountInString(s)`
4. 运行测试确认通过
5. 运行全量测试确认无回归

**验证标准：**
- [ ] `go test -run TestLength -v` 全部通过
- [ ] `go test ./...` 无失败
- [ ] commit：`fix: length() now counts Unicode runes instead of bytes`

---

### Step 1.2：修复 `&&`/`||` 运算符优先级

**做什么：**
1. 在 `parser_test.go` 或 `example_test.go` 中添加测试：
   - `[?(@.a==1||@.b==2&&@.c==3)]` 对 `{"a":1,"b":0,"c":0}` 期望匹配（`a==1` 为真，整个 `||` 为真）
   - `[?(@.a==0||@.b==2&&@.c==3)]` 对 `{"a":0,"b":2,"c":3}` 期望匹配
   - `[?(@.a==0||@.b==2&&@.c==0)]` 对 `{"a":0,"b":2,"c":0}` 期望不匹配
   - `[?((@.a==1||@.b==2)&&@.c==3)]` 括号覆盖优先级的用例
2. 运行测试确认失败
3. 修改 `parser.go` 中过滤表达式的解析逻辑，实现 `&&` 优先于 `||`
4. 运行测试确认通过

**验证标准：**
- [ ] `&&` 优先级高于 `||`
- [ ] 括号可覆盖默认优先级
- [ ] `!` 对复合表达式求值正确
- [ ] `go test ./...` 无失败
- [ ] commit：`fix: implement correct operator precedence (&& before ||)`

---

### Step 1.3：修复递归下降包含根节点

**做什么：**
1. 添加测试：
   - `{"name":"root","child":{"name":"child1"}}` 上 `$..name` 期望 `["root","child1"]`
   - `{"value":1}` 上 `$..value` 期望包含根节点的 `value`
2. 运行测试确认失败
3. 修改 `segments.go` 中 `recursiveSegment` 的 `evaluate` 方法，从根节点开始递归
4. 运行测试确认通过

**验证标准：**
- [ ] `$..name` 包含根节点匹配
- [ ] `$..*` 包含根节点自身
- [ ] `go test ./...` 无失败
- [ ] commit：`fix: recursive descent now includes root node`

---

### Step 1.4：修复选择器错误处理

**做什么：**
1. 添加测试：
   - `$.name[0]`（对字符串用索引）期望空结果，无 error
   - `$.count.foo`（对数字用字段名）期望空结果，无 error
   - `$.null_field.bar`（对 null 用选择器）期望空结果，无 error
   - `$.store[`（语法错误）期望返回 error
2. 运行测试确认失败
3. 修改 `segments.go` 各段的 `evaluate` 方法：类型不匹配时返回空 `[]interface{}{}` 而非 error
4. 运行测试确认通过

**验证标准：**
- [ ] 类型不匹配返回空结果，`err == nil`
- [ ] 语法错误仍返回 `err != nil`
- [ ] `go test ./...` 无失败
- [ ] commit：`fix: selectors return empty result on type mismatch instead of error`

---

### Step 1.5：运行 RFC 9535 测试套件验证 Phase 1 效果

**做什么：**
1. 重新运行 `go test -run TestRFC9535Suite -v`
2. 对比 Step 0.2 的基线数据

**验证标准：**
- [ ] 通过数高于基线
- [ ] 新通过的用例与 Phase 1 修复的 4 个问题相关
- [ ] 将新数据追加到 `testdata/rfc9535/baseline.txt`

---

## Phase 2：v2.1.0 新增语法

**目标：** 添加存在性测试和 `@` 单独引用，纯新增，不影响现有查询。

### Step 2.1：添加存在性测试 `[?@.name]`

**做什么：**
1. 在 `parser_test.go` 添加解析测试：
   - `[?@.name]` 能正确解析为存在性测试条件
   - `[?@.nested.field]` 支持嵌套路径
2. 在 `example_test.go` 添加集成测试：
   - 数据 `[{"name":"a","v":1},{"v":2},{"name":"b","v":3}]`，`$[?@.name]` 期望返回第 1、3 个元素
   - 数据 `[{"a":null},{"a":1}]`，`$[?@.a]` 期望只返回第 2 个元素（null 视为不存在）
3. 修改 `parser.go`：解析 `[?@.path]` 形式（无比较运算符）为存在性测试
4. 修改 `segments.go`：过滤器求值时，存在性测试检查字段是否存在且非 null

**验证标准：**
- [ ] `[?@.name]` 正确过滤
- [ ] `[?@.nested.field]` 支持嵌套
- [ ] null 值字段不匹配
- [ ] 现有 `[?@.name == "foo"]` 不受影响
- [ ] `go test ./...` 无失败
- [ ] commit：`feat: add existence test [?@.name]`

---

### Step 2.2：添加 `@` 单独引用

**做什么：**
1. 添加测试：
   - `[5,3,1,4,2]` 上 `$[?@>3]` 期望 `[5,4]`
   - `["a","b","c"]` 上 `$[?@=="b"]` 期望 `["b"]`
   - `[{"t":"a","v":1},{"t":"b","v":2}]` 上 `$[?@.t=="a"&&@.v>0]` 期望第 1 个元素
2. 修改 `parser.go`：允许 `@` 后直接跟比较运算符（不跟 `.field`）

**验证标准：**
- [ ] `[?@ > 3]` 正确过滤数值
- [ ] `[?@ == "b"]` 正确过滤字符串
- [ ] `@` 在复合表达式中正确引用当前节点
- [ ] 现有 `@.field` 语法不受影响
- [ ] `go test ./...` 无失败
- [ ] commit：`feat: support bare @ reference in filters`

---

### Step 2.3：v2.1.0 发布准备

**做什么：**
1. 运行 RFC 9535 测试套件，记录最终通过率
2. 更新 `CHANGELOG.md`：列出所有 bug fix 和新增语法
3. 更新 `README.md` / `README_zh.md`：标注已知非合规行为
4. 更新 `version.go`：版本号改为 `2.1.0`

**验证标准：**
- [ ] `go test ./...` 全部通过
- [ ] RFC 9535 通过率高于 Phase 0 基线
- [ ] CHANGELOG 包含所有变更
- [ ] README 标注了已知非合规行为
- [ ] `go build ./cmd/jp && ./jp --version` 输出 `v2.1.0`
- [ ] commit：`release: v2.1.0`

---

## Phase 3：v3.0.0 类型系统基础

**目标：** 引入 RFC 9535 核心类型，不影响现有代码（纯新增）。

### Step 3.1：定义核心类型

**做什么：**
1. 创建 `types_v3.go`（或在 `types.go` 中新增），定义：
   - `Node` struct（Location + Value）
   - `NodeList` type（`[]Node`）
   - `Nothing` struct
   - `LogicalType` int8（`LogicalNothing` / `LogicalFalse` / `LogicalTrue`）
2. 为 `NodeList` 实现 `json.Marshaler` 接口（序列化为 JSON 数组）
3. 为 `LogicalType` 实现 `String()` 方法
4. 为 `Nothing` 实现 `String()` 方法

**验证标准：**
- [ ] `Node{Location: "$['name']", Value: "test"}` 可创建
- [ ] `NodeList` JSON 序列化输出 `[{"location":"$[0]","value":1}]`
- [ ] `LogicalTrue.String()` 返回 `"true"`
- [ ] `go test ./...` 无失败（新类型不影响现有代码）
- [ ] commit：`feat: add Node, NodeList, Nothing, LogicalType types`

---

### Step 3.2：实现 Normalized Path 生成器

**做什么：**
1. 创建 `normalized_path.go`
2. 实现函数：给定当前路径段列表，生成 Normalized Path 字符串
3. 规则：
   - 根节点：`$`
   - 对象成员：`$['memberName']`（单引号包裹）
   - 数组元素：`$[0]`
   - 特殊字符转义：`'` → `\'`，`\` → `\\`，控制字符 → `\uXXXX`
4. 添加全面的测试用例

**验证标准：**
- [ ] `$` → `"$"`
- - `["store","book","0"]` → `"$['store']['book'][0]"`
- [ ] 含单引号的键名正确转义：`$['it\'s']`
- [ ] 含反斜杠的键名正确转义：`$['back\\slash']`
- [ ] 空键名：`$['']`
- [ ] `go test ./...` 无失败
- [ ] commit：`feat: implement Normalized Path generator`

---

### Step 3.3：实现 I-Regexp 解析器

**做什么：**
1. 创建 `iregexp.go`
2. 实现 I-Regexp 语法解析器，输出为 Go `regexp` 可执行的正则字符串
3. I-Regexp 核心规则（RFC 9485）：
   - 支持：`.`、字符类 `[abc]`、`[^abc]`、量词 `*`、`+`、`?`、`{n,m}`、分组 `(...)`、交替 `|`、锚点 `^`、`$`
   - 支持：`\d`、`\w`、`\s` 及其否定
   - 支持：Unicode 属性 `\p{L}`、`\p{N}` 等
   - 不支持：反向引用 `\1`、前瞻 `(?=)`、后顾 `(?<=)`、非捕获 `(?:)` 以外的特殊分组
4. 实现验证函数 `IsValidIRegexp(pattern string) bool`
5. 实现转换函数 `IRegexpToGoRegexp(pattern string) (string, error)`

**验证标准：**
- [ ] `"^S.*$"` 识别为合法 I-Regexp
- [ ] `"(a)\\1"` 识别为非法 I-Regexp（反向引用）
- [ ] `"(?=a)"` 识别为非法 I-Regexp（前瞻）
- [ ] `"\\p{L}"` 识别为合法 I-Regexp
- [ ] 转换后的正则在 Go `regexp` 中可编译
- [ ] `go test ./...` 无失败
- [ ] commit：`feat: implement I-Regexp parser and validator`

---

### Step 3.4：运行 RFC 9535 测试套件验证 Phase 3 效果

**验证标准：**
- [ ] 通过率不低于 Phase 2 最终结果
- [ ] 新类型和模块未引入回归

---

## Phase 4：v3.0.0 内部重构

**目标：** 将内部实现迁移到新类型系统，保持公共 API 暂不变。

### Step 4.1：重构 segment 接口

**做什么：**
1. 定义新的 segment 接口：
   ```go
   type segment interface {
       evaluate(node Node) (NodeList, error)
       String() string
   }
   ```
2. 暂时保留旧接口，新接口命名为 `segmentV3`
3. 逐个迁移段实现（先从最简单的 `wildcardSegment` 开始）

**验证标准：**
- [ ] `segmentV3` 接口定义存在
- [ ] `wildcardSegment` 实现新接口并通过测试
- [ ] 旧接口和实现不受影响
- [ ] `go test ./...` 无失败
- [ ] commit：`refactor: define v3 segment interface and migrate wildcardSegment`

---

### Step 4.2：逐个迁移剩余段实现

**按依赖顺序迁移（每迁移一个都需验证）：**

1. `nameSegment` — 字段访问
2. `indexSegment` — 数组索引
3. `sliceSegment` — 数组切片
4. `multiIndexSegment` — 多索引
5. `multiNameSegment` — 多字段名
6. `recursiveSegment` — 递归下降
7. `filterSegment` — 过滤器（最复杂，最后迁移）
8. `functionSegment` — 函数调用

**每个段迁移的验证标准：**
- [ ] 新接口实现通过单元测试
- [ ] 旧接口实现不受影响
- [ ] `go test ./...` 无失败
- [ ] 每个段单独 commit

---

### Step 4.3：重构过滤器系统

**做什么：**
1. 过滤器求值改为返回 `LogicalType`（三值逻辑）
2. 存在性测试返回 `LogicalTrue` / `LogicalFalse`
3. 比较运算返回 `LogicalTrue` / `LogicalFalse`
4. `&&` 运算：任一为 `LogicalFalse` 则 `LogicalFalse`；任一为 `LogicalNothing` 则 `LogicalNothing`；否则 `LogicalTrue`
5. `||` 运算：任一为 `LogicalTrue` 则 `LogicalTrue`；任一为 `LogicalNothing` 则 `LogicalNothing`；否则 `LogicalFalse`
6. `!` 运算：`LogicalTrue` ↔ `LogicalFalse`；`LogicalNothing` 保持 `LogicalNothing`

**验证标准：**
- [ ] 过滤器返回 `LogicalType`
- [ ] 三值逻辑真值表全部正确
- [ ] `go test ./...` 无失败
- [ ] commit：`refactor: filter system uses three-valued logic`

---

### Step 4.4：统一段实现为新接口

**做什么：**
1. 所有段已迁移完毕后，移除旧 `segment` 接口
2. 将 `segmentV3` 重命名为 `segment`
3. 更新所有调用点

**验证标准：**
- [ ] 旧接口完全移除
- [ ] 仅保留新接口
- [ ] `go test ./...` 无失败
- [ ] commit：`refactor: remove legacy segment interface`

---

### Step 4.5：实现 Node 贯穿查询管道

**做什么：**
1. `Query()` 内部构造根 `Node`（Location = `$`）
2. 每个段接收 `Node`，输出 `NodeList`
3. 子节点的 Location 由父 Location + 当前段的 Normalized Path 拼接

**验证标准：**
- [ ] 查询管道全程传递 `Node`（含 Location）
- [ ] `Query(data, "$.store.book[0]")` 的结果中每个 Node 的 Location 正确
- [ ] `go test ./...` 无失败
- [ ] commit：`refactor: query pipeline passes Node with Normalized Path`

---

## Phase 5：v3.0.0 公共 API 变更

**目标：** 变更公共 API 签名，实现 RFC 9535 标准函数。

### Step 5.1：变更 `Query()` 返回类型

**做什么：**
1. 修改 `jsonpath.go`：
   ```go
   func Query(data interface{}, path string) (NodeList, error)
   ```
2. 更新所有测试中的 `Query()` 调用
3. 更新 CLI 工具 `cmd/jp/main.go`

**验证标准：**
- [ ] `Query()` 返回 `NodeList`
- [ ] `result[0].Value` 获取原始值
- [ ] `result[0].Location` 获取 Normalized Path
- [ ] 空结果返回 `NodeList{}`（非 nil），`err == nil`
- [ ] `go test ./...` 全部通过
- [ ] `go build ./cmd/jp` 编译成功
- [ ] commit：`breaking: Query() returns NodeList instead of interface{}`

---

### Step 5.2：修正 `count()` 函数

**做什么：**
1. 将现有 `count(array, value)` 重命名为 `occurrences()`
2. 新增 RFC 标准 `count()`：接受 `NodeList`，返回节点数量
3. 更新函数注册和调用

**验证标准：**
- [ ] `count(@.items[*])` 返回节点数量
- [ ] `occurrences(@.items, "value")` 保留旧功能
- [ ] `go test ./...` 无失败
- [ ] commit：`breaking: rename count() to occurrences(), add RFC count()`

---

### Step 5.3：修正 `search()` 函数

**做什么：**
1. 将现有 `search(array, pattern)` 重命名为 `filterMatch()`
2. 新增 RFC 标准 `search(string, pattern)` → `LogicalType`
3. 使用 I-Regexp 匹配

**验证标准：**
- [ ] `search(@.name, "pattern")` 返回 `LogicalType`
- [ ] `filterMatch(@.items, "pattern")` 保留旧功能
- [ ] 使用 I-Regexp 语法
- [ ] `go test ./...` 无失败
- [ ] commit：`breaking: rename search() to filterMatch(), add RFC search()`

---

### Step 5.4：变更 `match()` 调用形式

**做什么：**
1. 修改 `parser.go`：将 `match()` 解析为标准函数调用（非方法式）
2. 修改 `functions.go`：`match()` 使用全串匹配 + I-Regexp
3. 移除旧方法式 `.match()` 的解析支持

**验证标准：**
- [ ] `match(@.name, 'S')` 只匹配恰好为 `"S"` 的字符串
- [ ] `match(@.name, 'S')` 不匹配 `"AliceSmith"`
- [ ] 使用 I-Regexp 语法
- [ ] 旧 `.match()` 语法解析失败
- [ ] `go test ./...` 无失败
- [ ] commit：`breaking: match() uses function syntax and full-string matching`

---

### Step 5.5：新增 `value()` 函数

**做什么：**
1. 在 `functions.go` 中实现 `value(nodelist) → ValueType`
2. 语义：恰好 1 个节点 → 返回值；否则 → 返回 Nothing

**验证标准：**
- [ ] `value(@.name)` 返回单个值
- [ ] 多节点时返回 Nothing
- [ ] 空节点列表时返回 Nothing
- [ ] `go test ./...` 无失败
- [ ] commit：`feat: add value() function`

---

### Step 5.6：更新 CLI 工具

**做什么：**
1. 更新 `cmd/jp/main.go` 适配新的 `NodeList` 返回类型
2. 默认输出每个结果的 Value（兼容旧行为）
3. 添加 `--path` 标志：输出 Normalized Path

**验证标准：**
- [ ] `echo '{"a":1}' | jp '$.a'` 输出 `1`（旧行为兼容）
- [ ] `echo '{"a":1}' | jp --path '$.a'` 输出 `$['a']` + 值
- [ ] `go build ./cmd/jp && go test ./cmd/jp/...` 通过
- [ ] commit：`feat: update CLI for NodeList output, add --path flag`

---

## Phase 6：测试通过 + 发布

**目标：** 达到 RFC 9535 测试套件 100% 通过，发布 v3.0.0。

### Step 6.1：运行 RFC 9535 测试套件，修复剩余失败

**做什么：**
1. 运行完整测试套件
2. 按失败类别分组（语法解析、选择器、过滤器、函数、边界情况）
3. 逐个修复，每修复一类 commit 一次

**验证标准：**
- [ ] RFC 9535 测试套件 100% 通过（324/324）
- [ ] `go test ./...` 无失败
- [ ] commit：`test: achieve 100% RFC 9535 compliance`

---

### Step 6.2：回归测试 + 性能测试

**做什么：**
1. 运行全量项目测试（包括 CLI 测试）
2. 编写基准测试 `BenchmarkQuery`，与 v2.0.2 对比
3. 如果性能下降 > 10%，优化热点路径

**验证标准：**
- [ ] `go test ./...` 全部通过
- [ ] `go test -bench=. -benchmem` 输出基准数据
- [ ] 性能不低于 v2.0.2 的 90%
- [ ] commit：`test: regression and benchmark tests`

---

### Step 6.3：文档更新

**做什么：**
1. 编写 `MIGRATION.md`：v2.x → v3.0 迁移指南
2. 更新 `CHANGELOG.md`：列出所有 breaking changes
3. 更新 `README.md` / `README_zh.md`：
   - 更新 API 示例
   - 标注 RFC 9535 合规状态
   - 标注扩展函数为非标准
4. 更新 GoDoc 注释

**验证标准：**
- [ ] `MIGRATION.md` 存在且包含所有 breaking changes 的迁移示例
- [ ] CHANGELOG 包含完整变更列表
- [ ] README 示例代码使用新 API
- [ ] commit：`docs: v3.0.0 migration guide and changelog`

---

### Step 6.4：发布 v3.0.0

**做什么：**
1. 更新 `version.go` 为 `3.0.0`
2. 打 tag `v3.0.0`
3. 创建 GitHub Release

**验证标准：**
- [ ] `go build ./cmd/jp && ./jp --version` 输出 `v3.0.0`
- [ ] `git tag v3.0.0` 存在
- [ ] GitHub Release 包含 Release Notes
- [ ] commit：`release: v3.0.0`

---

## 附录：Step 依赖关系

```
Phase 0
  ├─ Step 0.1 ─→ Step 0.2

Phase 1（可并行，但建议按序）
  ├─ Step 1.1 ─→ Step 1.5
  ├─ Step 1.2 ─→ Step 1.5
  ├─ Step 1.3 ─→ Step 1.5
  └─ Step 1.4 ─→ Step 1.5

Phase 2
  ├─ Step 2.1 ─→ Step 2.2 ─→ Step 2.3

Phase 3（可并行）
  ├─ Step 3.1 ─→ Step 3.2
  │            ─→ Step 3.3
  └─ Step 3.2 + Step 3.3 ─→ Step 3.4

Phase 4（严格顺序）
  └─ Step 4.1 ─→ Step 4.2 ─→ Step 4.3 ─→ Step 4.4 ─→ Step 4.5

Phase 5（Step 5.1 必须最先，其余可按序）
  └─ Step 5.1 ─→ Step 5.2 ─→ Step 5.3 ─→ Step 5.4 ─→ Step 5.5 ─→ Step 5.6

Phase 6（严格顺序）
  └─ Step 6.1 ─→ Step 6.2 ─→ Step 6.3 ─→ Step 6.4
```

---

## 附录：每个 Step 的 commit message 模板

| Step | Commit Message |
|------|---------------|
| 0.1 | `test: add RFC 9535 test suite data` |
| 0.2 | `test: integrate RFC 9535 test suite and establish baseline` |
| 1.1 | `fix: length() now counts Unicode runes instead of bytes` |
| 1.2 | `fix: implement correct operator precedence (&& before ||)` |
| 1.3 | `fix: recursive descent now includes root node` |
| 1.4 | `fix: selectors return empty result on type mismatch instead of error` |
| 1.5 | `test: verify Phase 1 fixes against RFC 9535 suite` |
| 2.1 | `feat: add existence test [?@.name]` |
| 2.2 | `feat: support bare @ reference in filters` |
| 2.3 | `release: v2.1.0` |
| 3.1 | `feat: add Node, NodeList, Nothing, LogicalType types` |
| 3.2 | `feat: implement Normalized Path generator` |
| 3.3 | `feat: implement I-Regexp parser and validator` |
| 3.4 | `test: verify Phase 3 against RFC 9535 suite` |
| 4.1 | `refactor: define v3 segment interface and migrate wildcardSegment` |
| 4.2 | `refactor: migrate remaining segment implementations` |
| 4.3 | `refactor: filter system uses three-valued logic` |
| 4.4 | `refactor: remove legacy segment interface` |
| 4.5 | `refactor: query pipeline passes Node with Normalized Path` |
| 5.1 | `breaking: Query() returns NodeList instead of interface{}` |
| 5.2 | `breaking: rename count() to occurrences(), add RFC count()` |
| 5.3 | `breaking: rename search() to filterMatch(), add RFC search()` |
| 5.4 | `breaking: match() uses function syntax and full-string matching` |
| 5.5 | `feat: add value() function` |
| 5.6 | `feat: update CLI for NodeList output, add --path flag` |
| 6.1 | `test: achieve 100% RFC 9535 compliance` |
| 6.2 | `test: regression and benchmark tests` |
| 6.3 | `docs: v3.0.0 migration guide and changelog` |
| 6.4 | `release: v3.0.0` |
