# Changelog

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