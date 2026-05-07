# Go JSONPath

[![Go Reference](https://pkg.go.dev/badge/github.com/davidhoo/jsonpath.svg)](https://pkg.go.dev/github.com/davidhoo/jsonpath)
[![Go Report Card](https://goreportcard.com/badge/github.com/davidhoo/jsonpath)](https://goreportcard.com/report/github.com/davidhoo/jsonpath)
[![Coverage Status](https://coveralls.io/repos/github/davidhoo/jsonpath/badge.svg?branch=main)](https://coveralls.io/github/davidhoo/jsonpath?branch=main)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

[中文文档](README_zh.md)

A complete Go implementation of JSONPath that fully complies with [RFC 9535](https://www.rfc-editor.org/rfc/rfc9535). Provides both a command-line tool and a Go library with support for all standard JSONPath features.

## Features

- Complete RFC 9535 Implementation
  - Root node access (`$`)
  - Child node access (`.key` or `['key']`)
  - Recursive descent (`..`)
  - Array indices (`[0]`, `[-1]`)
  - Array slices (`[start:end:step]`)
  - Array wildcards (`[*]`)
  - Multiple indices (`[1,2,3]`)
  - Multiple field names (`['name','age']`)
  - Filter expressions (`[?(@.price < 10)]`)
- Command Line Tool (`jp`)
  - Beautiful colorized output
  - Syntax highlighting for JSON
  - File and stdin input support
  - Formatted and compact output options
  - User-friendly error messages
  - UTF-8 support with proper CJK display
- Go Library
  - Clean API design
  - Type-safe operations
  - Rich examples
  - Comprehensive documentation

## What's New

### v3.0.0
- **Breaking Changes**
  - `Query()` now returns `NodeList` (slice of `Node` with `Location` and `Value`) instead of `interface{}`
  - `match()` syntax changed: `@.field.match('pattern')` → `match(@.field, 'pattern')`
  - `search()` syntax changed: `@.field.search('pattern')` → `search(@.field, 'pattern')`
  - `count()` now follows RFC 9535 semantics (counts nodelist nodes)
- **New Features**
  - RFC 9535 `value()` function
  - `occurrences()` function for value counting
  - `Node`/`NodeList` types with Normalized Path support
  - I-Regexp parser and validator
  - Benchmark tests
- **RFC 9535 Compliance**: Full compliance with RFC 9535 specification
- See [MIGRATION.md](MIGRATION.md) for upgrade guide

### v2.0.2
- Bug fixes
  - Fixed stdin reading when no `-f` flag is specified (#6)
  - Added `-f -` support for reading from stdin (#6)
  - Fixed gosec security check failures (#5)

### v2.0.1
- Bug fixes and improvements
  - Fixed minor issues from the initial 2.0.0 release
  - Improved error handling and reporting
  - Performance optimizations
- Documentation updates
  - Updated examples to reflect latest API changes
  - Enhanced documentation clarity
  - Fixed typos and inconsistencies

### v2.1.0
- Bug Fixes
  - `length()` now counts Unicode runes instead of bytes
  - Implement correct operator precedence (`&&` before `||`)
  - Recursive descent (`..`) now includes root node
  - Selectors return empty result on type mismatch instead of error
- New Features
  - Existence test `[?@.name]`
  - Bare `@` reference in filters
- Improvements
  - Integrated RFC 9535 test suite and established baseline
  - Verified Phase 1 fixes against RFC 9535 suite

### RFC 9535 Compliance
This implementation fully complies with [RFC 9535](https://www.rfc-editor.org/rfc/rfc9535), including:
- All standard selectors (name, index, slice, wildcard, filter, recursive descent, union)
- All standard functions (`length`, `count`, `match`, `search`, `value`)
- I-Regexp pattern matching
- Normalized Path generation
- Three-valued logic in filter expressions

### v2.0.0
- Complete rewrite with RFC 9535 compliance
  - Full implementation of JSONPath specification (RFC 9535)
  - Enhanced error handling and reporting
  - Improved performance and reliability
  - Better support for various JSONPath expressions
- Breaking Changes
  - API changes to align with RFC specification
  - Updated function signatures for better usability
  - Modified error types for more detailed error reporting
- Documentation
  - Updated documentation to reflect RFC compliance
  - Added more examples and use cases
  - Improved API documentation

### v1.0.4
- Centralize version management
  - Add version.go for centralized version control
  - Update cmd/jp to use centralized version
  - Fix UTF-8 encoding in Chinese comments

### v1.0.3
- Enhanced filter expressions
  - Full support for logical operators (`&&`, `||`, `!`)
  - Proper handling of complex filter conditions
  - Support for De Morgan's Law in negated expressions
  - Improved numeric and string comparisons
  - Better error messages
- Improved API design
  - New simplified `Query` function for easier usage
  - Deprecated `Compile/Execute` in favor of `Query`
  - Better error handling and reporting
- Updated examples
  - New examples demonstrating logical operators
  - Updated code to use new `Query` function
  - Fixed UTF-8 encoding issues in examples

### v1.0.2
- Enhanced filter expressions
  - Full support for logical operators (`&&`, `||`, `!`)
  - Proper handling of complex filter conditions
  - Support for De Morgan's Law in negated expressions
  - Improved numeric and string comparisons
  - Better error messages
- Enhanced colorized output
  - Beautiful syntax highlighting for JSON
  - Colorful command-line interface
  - Improved readability for nested structures
- Better UTF-8 support
  - Fixed CJK character display
  - Proper handling of multi-byte characters

### v1.0.1
- Initial stable release
  - Basic JSONPath query support
  - Command-line tool implementation
  - Core filter expression support
  - Basic colorized output
  - Initial documentation
  - Basic error handling
  - UTF-8 support

## Installation

### Homebrew (Recommended)

```bash
# Add tap
brew tap davidhoo/tap

# Install jsonpath
brew install jsonpath
```

### Go Install

```bash
go install github.com/davidhoo/jsonpath/cmd/jp@latest
```

### Manual Installation

Download the appropriate binary for your platform from the [releases page](https://github.com/davidhoo/jsonpath/releases).

## Command Line Usage

### CLI Basic Usage

```bash
jp [-p <jsonpath_expression>] [-f <json_file>] [-c]
```

Options:

- `-p` JSONPath expression (if not specified, output entire JSON)
- `-f` JSON file path (reads from stdin if not specified)
- `-c` Compact output (no formatting)
- `--no-color` Disable colored output
- `-h` Show help information
- `-v` Show version information

### Examples

```bash
# Output entire JSON with syntax highlighting
jp -f data.json

# Query specific path
jp -f data.json -p '$.store.book[*].author'

# Filter with conditions
jp -f data.json -p '$.store.book[?(@.price > 10)]'

# Read from stdin
echo '{"name": "John"}' | jp -p '$.name'

# Compact output
jp -f data.json -c
```

## Go Library Usage

### Library Basic Usage

```go
import "github.com/davidhoo/jsonpath"

// Query JSON data - returns NodeList
result, err := jsonpath.Query(data, "$.store.book[*].author")
if err != nil {
    log.Fatal(err)
}

// Handle result (NodeList)
for _, node := range result {
    fmt.Printf("Location: %s, Value: %v\n", node.Location, node.Value)
}
```

### Complete Example

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "github.com/davidhoo/jsonpath"
)

func main() {
    // JSON data
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

    // Parse JSON
    var v interface{}
    if err := json.Unmarshal([]byte(data), &v); err != nil {
        log.Fatal(err)
    }

    // Execute JSONPath query - returns NodeList
    result, err := jsonpath.Query(v, "$.store.book[?(@.price < 10)].title")
    if err != nil {
        log.Fatal(err)
    }

    // Print result
    for _, node := range result {
        fmt.Println(node.Value) // Sayings of the Century
    }
}
```

### Common Query Examples

```go
// Get all prices (recursive)
"$..price"

// Get books with specific price range
"$.store.book[?(@.price < 10)].title"

// Get all authors
"$.store.book[*].author"

// Get first book
"$.store.book[0]"

// Get last book
"$.store.book[-1]"

// Get first two books
"$.store.book[0:2]"

// Get books with price > 10 and category == fiction
"$.store.book[?(@.price > 10 && @.category == 'fiction')]"

// Get all non-reference books
"$.store.book[?(!(@.category == 'reference'))]"

// Get books with price > 10 or author containing 'Evelyn'
"$.store.book[?(@.price > 10 || @.author == 'Evelyn Waugh')]"

// Get length of book array
"$.store.book.length()"

// Get all keys of store object
"$.store.keys()"

// Get all values of store object
"$.store.values()"

// Get minimum book price
"$.store.book[*].price.min()"

// Get maximum book price
"$.store.book[*].price.max()"

// Get average book price
"$.store.book[*].price.avg()"

// Get total price of all books
"$.store.book[*].price.sum()"

// Count fiction books (RFC 9535 count())
"$.store.book.count()"

// Count occurrences of a value (non-standard extension)
"$.store.book[*].category.occurrences('fiction')"

// Match book titles with regex pattern (RFC 9535 function-style)
"$.store.book[?match(@.title, '^S.*')]"

// Search for books with titles starting with 'S' (RFC 9535 function-style)
"$.store.book[?search(@.title, '^S.*')]"

// Extract multiple fields from objects
"$.store.book[*]['author','price']"

// Extract multiple fields with wildcard
"$.store.book[*]['title','category']"

// Chain multiple functions
"$.store.book[?@.price > 10].title.length()"

// Complex function chaining
"$.store.book[?@.price > $.store.book[*].price.avg()].title"

// Combine search and filter (RFC 9535 function-style)
"$.store.book[?match(@.title, '^S.*') && @.price < 10].author"
```

### Result Handling

`Query()` returns a `NodeList` (slice of `Node`). Each `Node` has a `Location` (Normalized Path) and a `Value`:

```go
// Iterate over results
for _, node := range result {
    fmt.Printf("Location: %s\n", node.Location)  // e.g., "$['store']['book'][0]"
    fmt.Printf("Value: %v\n", node.Value)         // the actual value
}

// Access the first result's value
if len(result) > 0 {
    firstValue := result[0].Value
}

// Type-assert individual values
for _, node := range result {
    if str, ok := node.Value.(string); ok {
        // Handle string value
    }
}
```

## Implementation Details

1. RFC 9535 Compliance
   - Support for all standard operators
   - Standard-compliant syntax parsing
   - Standard result formatting

2. Filter Support
   - Comparison operators: `<`, `>`, `<=`, `>=`, `==`, `!=`
   - Logical operators: `&&`, `||`, `!`
   - Support for complex filter conditions
   - Support for numeric and string comparisons
   - Proper handling of negated expressions using De Morgan's Law
   - Nested filter conditions with parentheses

3. Result Handling
   - Array operations return array results
   - Single value access returns original type
   - Type-safe result handling

4. Error Handling
   - Detailed error messages
   - Syntax error reporting
   - Runtime error handling

## Contributing

Issues and Pull Requests are welcome!

## License

MIT License

## Testing

### Running Tests

```bash
# Run all tests
go test -v ./...

# Run tests with coverage
go test -v -coverprofile=coverage.out ./...

# View coverage report in browser
go tool cover -html=coverage.out

# Run benchmarks
go test -bench=. -benchmem ./...
```

### Test Organization

The test suite is organized into several categories:

- Unit Tests: Testing individual functions and components
- Integration Tests: Testing interactions between components
- Benchmark Tests: Performance testing of critical paths
- Example Tests: Documenting usage through examples

### Test Coverage

Current test coverage metrics:
- Overall Coverage: 90%+
- Core Package: 90%+
- Command Line Tool: 90%+

### Writing Tests

When contributing tests, please follow these guidelines:

1. Use table-driven tests for similar test cases
2. Provide clear test case descriptions
3. Test both success and error cases
4. Include edge cases and boundary conditions
5. Use test suites for related functionality
6. Add benchmarks for performance-critical code

Example of a well-structured test:

```go
func TestParseJSONPath(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected interface{}
        wantErr  bool
    }{
        {
            name:     "simple path",
            input:    "$.store.book",
            expected: []string{"store", "book"},
            wantErr:  false,
        },
        // Add more test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := ParseJSONPath(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ParseJSONPath() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(result, tt.expected) {
                t.Errorf("ParseJSONPath() = %v, want %v", result, tt.expected)
            }
        })
    }
}
```
