# Go JSONPath

[![Go Reference](https://pkg.go.dev/badge/github.com/davidhoo/jsonpath.svg)](https://pkg.go.dev/github.com/davidhoo/jsonpath)
[![Go Report Card](https://goreportcard.com/badge/github.com/davidhoo/jsonpath)](https://goreportcard.com/report/github.com/davidhoo/jsonpath)
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

// Query JSON data
result, err := jsonpath.Query(data, "$.store.book[*].author")
if err != nil {
    log.Fatal(err)
}

// Handle result
authors, ok := result.([]interface{})
if !ok {
    log.Fatal("unexpected result type")
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

    // Execute JSONPath query
    result, err := jsonpath.Query(v, "$.store.book[?(@.price < 10)].title")
    if err != nil {
        log.Fatal(err)
    }

    // Print result
    fmt.Printf("%v\n", result) // ["Sayings of the Century"]
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

// Count fiction books
"$.store.book[*].category.count('fiction')"

// Match book titles with regex pattern
"$.store.book[?@.title.match('^S.*')]"

// Search for books with titles starting with 'S'
"$.store.book[*].title.search('^S.*')"

// Chain multiple functions
"$.store.book[?@.price > 10].title.length()"

// Complex function chaining
"$.store.book[?@.price > $.store.book[*].price.avg()].title"

// Combine search and filter
"$.store.book[?@.title.match('^S.*') && @.price < 10].author"
```

### Result Handling

Handle results according to their type using type assertions:

```go
// Single value result
if str, ok := result.(string); ok {
    // Handle string result
}

// Array result
if arr, ok := result.([]interface{}); ok {
    for _, item := range arr {
        // Handle each item
    }
}

// Object result
if obj, ok := result.(map[string]interface{}); ok {
    // Handle object
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
