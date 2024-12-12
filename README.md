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
  - Colorized output for readability
  - File and stdin input support
  - Formatted and compact output options
  - User-friendly error messages
- Go Library
  - Clean API design
  - Type-safe operations
  - Rich examples
  - Comprehensive documentation

## Installation

### Command Line Tool

```bash
go install github.com/davidhoo/jsonpath/cmd/jp@latest
```

Or build from source:

```bash
git clone https://github.com/davidhoo/jsonpath.git
cd jsonpath
go build -o jp cmd/jp/main.go
```

### Go Library

```bash
go get github.com/davidhoo/jsonpath
```

## Command Line Usage

### Basic Usage

```bash
jp -p <jsonpath_expression> [-f <json_file>] [-c]
```

Options:
- `-p` JSONPath expression (required)
- `-f` JSON file path (reads from stdin if not specified)
- `-c` Compact output (no formatting)
- `-h` Show help information
- `-v` Show version information

### Examples

```bash
# Read from file
jp -f data.json -p '$.store.book[*].author'

# Read from stdin
echo '{"name": "John"}' | jp -p '$.name'

# Compact output
jp -f data.json -p '$.store.book[0]' -c
```

## Go Library Usage

### Basic Usage

```go
import "github.com/davidhoo/jsonpath"

// Compile JSONPath expression
jp, err := jsonpath.Compile("$.store.book[*].author")
if err != nil {
    log.Fatal(err)
}

// Execute query
result, err := jp.Execute(data)
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

    // Compile and execute JSONPath
    jp, err := jsonpath.Compile("$.store.book[?(@.price < 10)].title")
    if err != nil {
        log.Fatal(err)
    }

    result, err := jp.Execute(v)
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

// Get books within price range
"$.store.book[?(@.price < 10)].title"

// Get all authors
"$.store.book[*].author"

// Get first book
"$.store.book[0]"

// Get last book
"$.store.book[-1]"

// Get first two books
"$.store.book[0:2]"

// Get all books with price > 10
"$.store.book[?(@.price > 10)]"
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
   - Currently supports numeric comparisons
   - Future support for string comparisons and logical operators

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