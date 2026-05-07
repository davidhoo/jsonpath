# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v3.0.0] - 2026-05-07

### Breaking Changes

- **`Query()` returns `NodeList`** instead of `interface{}`
  - Each result is a `Node` with `Location` (Normalized Path) and `Value`
  - See [MIGRATION.md](MIGRATION.md) for upgrade guide
- **`match()` syntax changed** from method-style to function-style
  - Before: `@.field.match('pattern')`
  - After: `match(@.field, 'pattern')`
  - Now uses full-string matching with I-Regexp
- **`search()` signature changed** to RFC 9535 semantics
  - Before: `search(array, pattern)` - filtered array
  - After: `search(string, pattern)` - returns boolean
- **`count()` signature changed** to RFC 9535 semantics
  - Before: `count(array, value)` - counted value occurrences
  - After: `count(nodelist)` - counts nodes
  - Old behavior moved to `occurrences()`

### Added

- `value()` function - extracts single value from nodelist
- `occurrences()` function - counts value occurrences (non-standard)
- `Node` / `NodeList` types with Normalized Path support
- `Nothing` type - represents absence of value
- `LogicalType` - three-valued logic for filter expressions
- I-Regexp parser and validator (RFC 9485)
- Normalized Path generator (RFC 9535)
- Benchmark tests for performance testing
- CLI `--path` flag for Normalized Path output

### Changed

- Full RFC 9535 compliance (703/703 tests passing)
- Correct operator precedence (`&&` before `||`)
- Recursive descent (`..`) includes root node
- Selectors return empty result on type mismatch instead of error
- `length()` counts Unicode runes instead of bytes
- Filter expressions support existence tests (`[?@.name]`)
- Filter expressions support bare `@` reference
- Comparison with null/boolean returns false instead of error
- Zero step in slice returns empty result instead of error

### Fixed

- Nested filter expressions parsing (`$[?@[?@>1]]`)
- Function calls in filter expressions (`count()`, `length()`, etc.)
- Parenthesized expressions in filters
- Escaped quotes in filter strings
- Whitespace handling in various positions
- Unicode line separator matching (`\u2028`, `\u2029`)
- Index validation (leading zeros, out of range, etc.)
- Name selector validation (invalid characters)

## [v2.1.0] - 2026-05-06

### Added

- Existence test `[?@.name]` in filter expressions
- Bare `@` reference in filter expressions (`$[?@>3]`)
- RFC 9535 test suite integration (703 test cases)
- Baseline pass rate tracking

### Fixed

- `length()` now counts Unicode runes instead of bytes
- Operator precedence: `&&` now has higher precedence than `||`
- Recursive descent (`..`) now includes root node in results
- Selectors return empty result on type mismatch instead of error

## [v2.0.2] - 2026-05-06

### Fixed

- Stdin reading when no `-f` flag is specified (#6)
- Added `-f -` support for reading from stdin (#6)
- gosec security check failures (#5)

## [v2.0.1] - 2025-01-03

### Fixed

- Program waiting indefinitely when no input provided
- Added support for `--help` flag

### Changed

- Refactored main logic to `run()` function for better testability
- Improved error handling mechanism

## [v2.0.0] - 2024-12-16

### Changed

- Complete rewrite with RFC 9535 compliance
- API changes to align with RFC specification
- Updated function signatures for better usability
- Modified error types for more detailed error reporting

### Added

- Full implementation of JSONPath specification (RFC 9535)
- Enhanced error handling and reporting
- Improved performance and reliability
- Updated documentation with more examples

## [v1.0.4] - 2024-12-16

### Changed

- Centralized version management in `version.go`
- Updated `cmd/jp` to use centralized version
- Fixed UTF-8 encoding in Chinese comments

## [v1.0.3] - 2024-12-16

### Added

- Full support for logical operators (`&&`, `||`, `!`)
- Proper handling of complex filter conditions
- Support for De Morgan's Law in negated expressions
- New simplified `Query` function
- Examples demonstrating logical operators

### Changed

- Deprecated `Compile/Execute` in favor of `Query`
- Improved error handling and reporting

## [v1.0.2] - 2024-12-12

### Added

- Enhanced filter expressions with logical operators
- Beautiful syntax highlighting for JSON output
- Colorful command-line interface
- Better UTF-8 support for CJK characters

## [v1.0.1] - 2024-12-12

### Added

- Initial stable release
- Basic JSONPath query support
- Command-line tool implementation
- Core filter expression support
- Basic colorized output
- UTF-8 support
