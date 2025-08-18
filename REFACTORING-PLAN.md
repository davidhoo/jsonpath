# JSONPath 项目重构计划

## 项目背景

JSONPath 项目当前存在几个文件过大的问题，降低了代码的可维护性和可读性。本重构计划旨在通过拆分大文件、重组项目结构，使代码组织更加科学合理。

## 重构目标

1. 解决单个文件过大问题（parser.go, functions.go, segments.go 等）
2. 按功能职责拆分代码到适当的包和文件中
3. 保持功能完整性和向后兼容性
4. 提高代码可维护性和可读性
5. 优化测试结构
6. 每次创建一个新的分支，完成一个阶段任务。每个阶段任务都运行单元测试和功能测试，保证重构不影响功能。

## 重构计划和任务清单

### 阶段一：项目结构重组

- [x] 1.1 创建新的目录结构
  - [x] 创建 `pkg` 目录用于存放核心功能包
  - [x] 创建 `internal` 目录用于存放内部助手功能
  - [x] 保留 `cmd` 和 `examples` 目录

- [x] 1.2 创建核心子包目录
  - [x] 创建 `pkg/parser` 用于解析 JSONPath 表达式
  - [x] 创建 `pkg/segments` 用于实现各种段
  - [x] 创建 `pkg/functions` 用于实现 JSONPath 函数
  - [x] 创建 `pkg/eval` 用于实现执行引擎
  - [x] 创建 `pkg/compare` 用于比较和过滤功能
  - [x] 创建 `pkg/errors` 用于错误处理

### 阶段二：核心接口和类型迁移

- [x] 2.1 创建和迁移基础接口和类型
  - [x] 创建 `pkg/segments/segment.go` 定义 segment 接口
  - [x] 创建 `pkg/functions/function.go` 定义 Function 接口
  - [x] 创建 `pkg/errors/errors.go` 定义错误类型和常量

### 阶段三：解析器模块重构

- [x] 3.1 拆分 parser.go
  - [x] 创建 `pkg/parser/parser.go` - 核心解析逻辑和入口函数
  - [x] 创建 `pkg/parser/filter.go` - 过滤表达式解析
  - [x] 创建 `pkg/parser/bracket.go` - 括号表达式解析
  - [x] 创建 `pkg/parser/function.go` - 函数表达式解析
  - [x] 创建 `pkg/parser/utils.go` - 辅助功能和工具

- [x] 3.2 重构解析器测试
  - [x] 创建 `pkg/parser/parser_test.go`
  - [x] 创建 `pkg/parser/filter_test.go`
  - [x] 创建 `pkg/parser/bracket_test.go`
  - [x] 创建 `pkg/parser/function_test.go`

### 阶段四：segments 模块重构

- [ ] 4.1 拆分 segments.go
  - [x] 创建 `pkg/segments/name.go` - 名称段实现
  - [x] 创建 `pkg/segments/index.go` - 索引段实现
  - [x] 创建 `pkg/segments/wildcard.go` - 通配符段实现
  - [x] 创建 `pkg/segments/recursive.go` - 递归段实现
  - [ ] 创建 `pkg/segments/slice.go` - 切片段实现
  - [ ] 创建 `pkg/segments/filter.go` - 过滤段实现
  - [ ] 创建 `pkg/segments/function.go` - 函数段实现

- [ ] 4.2 重构段测试
  - [x] 创建相应的测试文件，针对不同类型的段

### 阶段五：函数模块重构

- [x] 5.1 拆分 functions.go
  - [x] 创建 `pkg/functions/registry.go` - 函数注册和查找
  - [x] 创建 `pkg/functions/array.go` - 数组相关函数
  - [x] 创建 `pkg/functions/string.go` - 字符串处理函数
  - [x] 创建 `pkg/functions/math.go` - 数学和聚合函数
  - [x] 创建 `pkg/functions/object.go` - 对象操作函数
  - [x] 创建 `pkg/functions/type.go` - 类型相关函数

- [x] 5.2 重构函数测试
  - [x] 创建相应的测试文件，针对不同类别的函数

### 阶段六：命令行工具优化

- [x] 6.1 拆分 cmd/jp/main.go
  - [x] 保留 `cmd/jp/main.go` 作为入口点
  - [x] 创建 `cmd/jp/formatter.go` - 输出格式化
  - [x] 创建 `cmd/jp/cli.go` - 命令行参数处理
  - [x] 创建 `cmd/jp/processor.go` - 处理逻辑

- [x] 6.2 重构命令行工具测试
  - [x] 创建相应的测试文件

### 阶段七：集成与兼容性

- [x] 7.1 创建根包兼容层
  - [x] 更新 `jsonpath.go` 作为兼容层，保持现有公共 API
  - [x] 验证向后兼容性

- [x] 7.2 更新测试与示例
  - [x] 确保所有现有测试在新结构下通过
  - [x] 更新示例以反映新结构

- [x] 7.3 更新文档
  - [x] 更新 README.md 和其他文档
  - [x] 添加新的包级文档

### 阶段八：性能测试与优化

- [ ] 8.1 添加性能基准测试
  - [ ] 创建基准测试对比重构前后的性能
  - [ ] 识别并解决任何性能回退

- [ ] 8.2 优化关键路径
  - [ ] 优化解析器性能（如有必要）
  - [ ] 优化执行引擎性能（如有必要）

## 当前进展

1. 已完成阶段一、二、三、五、六、七的所有任务
2. 正在进行阶段四：segments 模块重构
3. 待开始阶段八：性能测试与优化

## 下一步计划

1. 开始阶段四的 segments 模块重构
   - 优先处理基础段类型（name.go, index.go）
   - 然后是复杂段类型（wildcard.go, recursive.go）
   - 最后是特殊段类型（slice.go, filter.go, function.go）

2. 为每个段类型创建对应的测试文件
   - 确保测试覆盖所有功能
   - 包括边界情况和错误处理

3. 完成重构后进行全面测试
   - 运行所有单元测试
   - 运行性能基准测试
   - 验证向后兼容性

## 实施注意事项

1. 每个阶段结束后运行完整测试套件确保功能正确性
2. 保持公共 API 向后兼容
3. 阶段之间可能会有交叉和依赖，需要灵活调整顺序
4. 优先处理核心功能，然后是辅助功能
5. 每次修改保持小的提交范围，便于回滚或审查

## 预期收益

1. 代码模块化，职责更明确
2. 降低单个文件复杂度，提高可维护性
3. 使新功能开发更加灵活
4. 提高代码复用率
5. 测试更有针对性
6. 更好的包级文档组织 