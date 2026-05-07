# Changelog

## [v3.0.0] - 2026-05-07

### Breaking Changes
- **`Query()` returns `NodeList`** - `Query()` now returns `NodeList` (a slice of `Node` with `Location` and `Value`) instead of `interface{}`
- **`match()` syntax change** - Changed from method-style `@.field.match('pattern')` to RFC 9535 function-style `match(@.field, 'pattern')`
- **`search()` signature change** - Changed from method-style `@.field.search('pattern')` to RFC 9535 function-style `search(@.field, 'pattern')`
- **`count()` signature change** - Now follows RFC 9535 semantics: counts nodes in a nodelist instead of counting value occurrences

### New Features
- **RFC 9535 `value()` function** - Extracts a single value from a nodelist
- **`occurrences()` function** - Non-standard extension to count value occurrences in an array (replaces old `count()` behavior)
- **`Node` type** - New type with `Location` (Normalized Path) and `Value` fields
- **`NodeList` type** - New type representing a list of nodes returned by `Query()`
- **`LogicalType`** - Three-valued logic type for RFC 9535 compliance
- **`Nothing` type** - Represents Nothing value (different from null)
- **I-Regexp parser** - Full implementation of RFC 9535 I-Regexp specification
- **Normalized Path generator** - Generates RFC 9535 Normalized Paths
- **Benchmark tests** - Added performance benchmarks for Query()

### Improvements
- Full RFC 9535 compliance for filter expressions
- Correct operator precedence (`&&` before `||`)
- Recursive descent (`..`) now includes root node
- Selectors return empty result on type mismatch instead of error
- `length()` counts Unicode runes instead of bytes
- Existence test `[?@.name]` support
- Bare `@` reference in filters

## [v2.1.0] - 2026-05-06

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

## [v2.0.2] - 2026-05-06

### Fixed
- 修复了无 `-f` 参数时 `jp` 无法从 stdin 读取数据的问题 (#6)
- 支持 `-f -` 从 stdin 读取数据 (#6)
- 修复了 gosec 安全检查失败的问题 (#5)

## [v2.0.1] - 2024-01-17

### Fixed
- 修复了无输入时程序一直等待的问题
- 添加了对 `--help` 标志的支持

### Changed
- 重构了主要逻辑到 `run()` 函数，提高代码可测试性
- 改进了错误处理机制

### Chore
- 将 `.cursorrules` 从版本控制中移除
- 添加了新的测试用例 