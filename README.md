# Go JSONPath

[![Go Reference](https://pkg.go.dev/badge/github.com/davidhoo/jsonpath.svg)](https://pkg.go.dev/github.com/davidhoo/jsonpath)
[![Go Report Card](https://goreportcard.com/badge/github.com/davidhoo/jsonpath)](https://goreportcard.com/report/github.com/davidhoo/jsonpath)
[![Coverage Status](https://coveralls.io/repos/github/davidhoo/jsonpath/badge.svg?branch=main)](https://coveralls.io/github/davidhoo/jsonpath?branch=main)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

[中文文档](README_zh.md)

A complete Go implementation of JSONPath that fully complies with [RFC 9535](https://www.rfc-editor.org/rfc/rfc9535). Provides both a command-line tool and a Go library with support for all standard JSONPath features.

## Features

- **RFC 9535 Compliant** - Passes all 703 compliance tests
- **Complete JSONPath Support**
  - Root node access (`$`)
  - Child node access (`.key` or `['key']`)
  - Recursive descent (`..`)
  - Array indices (`[0]`, `[-1]`)
  - Array slices (`[start:end:step]`)
  - Array wildcards (`[*]`)
  - Multiple indices (`[1,2,3]`)
  - Multiple field names (`['name','age']`)
  - Filter expressions (`[?(@.price < 10)]`)
  - Existence tests (`[?@.name]`)
  - Function calls (`length()`, `count()`, `match()`, `search()`, `value()`)
- **Command Line Tool (`jp`)**
  - Beautiful colorized output
  - Syntax highlighting for JSON
  - File and stdin input support
  - Formatted and compact output options
  - Normalized Path output (`--path`)
- **Go Library**
  - Clean API with `NodeList` return type
  - Normalized Path for each result node
  - Type-safe operations
  - Comprehensive documentation

## Installation

### Homebrew (Recommended)

```bash
brew tap davidhoo/tap
brew install jsonpath
```

### Go Install

```bash
go install github.com/davidhoo/jsonpath/cmd/jp@latest
```

### Manual Installation

Download the appropriate binary for your platform from the [releases page](https://github.com/davidhoo/jsonpath/releases).

## Command Line Usage

```bash
jp [-p <jsonpath_expression>] [-f <json_file>] [-c] [--no-color] [--path]
```

Options:

| Flag | Description |
|------|-------------|
| `-p` | JSONPath expression (if not specified, output entire JSON) |
| `-f` | JSON file path (reads from stdin if not specified) |
| `-c` | Compact output (no formatting) |
| `--no-color` | Disable colored output |
| `--path` | Output Normalized Path with each result |
| `-h` | Show help information |
| `-v` | Show version information |

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

# Show Normalized Paths
echo '{"a":1,"b":2}' | jp --path '$.*'
# Output:
# $['a'] 1
# $['b'] 2
```

## Go Library Usage

### Basic Usage

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
    data := `{
        "store": {
            "book": [
                {"category": "reference", "author": "Nigel Rees", "title": "Sayings of the Century", "price": 8.95},
                {"category": "fiction", "author": "Evelyn Waugh", "title": "Sword of Honour", "price": 12.99}
            ]
        }
    }`

    var v interface{}
    if err := json.Unmarshal([]byte(data), &v); err != nil {
        log.Fatal(err)
    }

    result, err := jsonpath.Query(v, "$.store.book[?(@.price < 10)].title")
    if err != nil {
        log.Fatal(err)
    }

    for _, node := range result {
        fmt.Println(node.Value) // Sayings of the Century
    }
}
```

### Common Query Examples

```go
// Get all prices (recursive descent)
"$..price"

// Get books with specific price range
"$.store.book[?(@.price < 10)].title"

// Get all authors
"$.store.book[*].author"

// Get first/last book
"$.store.book[0]"
"$.store.book[-1]"

// Get first two books
"$.store.book[0:2]"

// Complex filter conditions
"$.store.book[?(@.price > 10 && @.category == 'fiction')]"

// Existence test
"$[?@.name]"

// Function calls (RFC 9535)
"$.store.book[?match(@.title, '^S.*')]"
"$.store.book[?search(@.title, 'Century')]"
"$[?count(@..*) > 5]"

// Non-standard extensions
"$.store.book[*].price.min()"
"$.store.book[*].price.max()"
"$.store.book[*].price.avg()"
"$.store.book[*].price.sum()"
```

### Result Handling

`Query()` returns a `NodeList` (slice of `Node`). Each `Node` has:
- `Location` - Normalized Path (e.g., `$['store']['book'][0]`)
- `Value` - The actual value

```go
for _, node := range result {
    fmt.Printf("Location: %s\n", node.Location)
    fmt.Printf("Value: %v\n", node.Value)
}

// Access first result
if len(result) > 0 {
    firstValue := result[0].Value
}
```

## RFC 9535 Compliance

This implementation fully complies with [RFC 9535](https://www.rfc-editor.org/rfc/rfc9535):

- **100% pass rate** on the official compliance test suite (703/703)
- All standard selectors (name, index, slice, wildcard, filter, recursive descent, union)
- All standard functions (`length`, `count`, `match`, `search`, `value`)
- I-Regexp pattern matching (RFC 9485)
- Normalized Path generation
- Three-valued logic in filter expressions

See [RFC9535_COMPLIANCE_REPORT.md](docs/RFC9535_COMPLIANCE_REPORT.md) for detailed compliance information.

## Non-Standard Extensions

The following functions are non-standard extensions for convenience:

| Function | Description |
|----------|-------------|
| `keys()` | Returns sorted keys of an object |
| `values()` | Returns values of an object |
| `min()` | Returns minimum value in an array |
| `max()` | Returns maximum value in an array |
| `avg()` | Returns average of numeric values |
| `sum()` | Returns sum of numeric values |
| `occurrences()` | Counts occurrences of a value in an array |

## Testing

```bash
# Run all tests
go test -v ./...

# Run tests with coverage
go test -v -coverprofile=coverage.out ./...

# Run benchmarks
go test -bench=. -benchmem ./...

# Run RFC 9535 compliance tests
go test -run TestCTS -v
```

## Contributing

Issues and Pull Requests are welcome!

## License

MIT License

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for release history.
