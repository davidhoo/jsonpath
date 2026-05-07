# Migration Guide: v2.x → v3.0.0

This guide covers all breaking changes in v3.0.0 and how to update your code.

## 1. `Query()` now returns `NodeList`

**v2.x:**
```go
result, err := jsonpath.Query(data, "$.store.book[*].title")
// result is interface{}, typically []interface{}
authors, ok := result.([]interface{})
```

**v3.0.0:**
```go
result, err := jsonpath.Query(data, "$.store.book[*].title")
// result is NodeList ([]Node)
for _, node := range result {
    fmt.Printf("Location: %s, Value: %v\n", node.Location, node.Value)
}
```

**Migration steps:**
- Replace `result.([]interface{})` with direct iteration over `NodeList`
- Access values via `node.Value` instead of direct element access
- Access normalized paths via `node.Location`

## 2. `match()` syntax change

**v2.x** (non-standard method-style):
```go
// Used as: $.store.book[?@.title.match('^S.*')]
```

**v3.0.0** (RFC 9535 function-style):
```go
// Used as: $.store.book[?match(@.title, '^S.*')]
```

**Migration steps:**
- Change `@.field.match('pattern')` to `match(@.field, 'pattern')`
- The function now takes two arguments: the string and the pattern

## 3. `search()` signature change

**v2.x** (non-standard method-style):
```go
// Used as: $.store.book[*].title.search('^S.*')
```

**v3.0.0** (RFC 9535 function-style):
```go
// Used as: $.store.book[?search(@.title, '^S.*')]
```

**Migration steps:**
- Change `@.field.search('pattern')` to `search(@.field, 'pattern')`
- Use inside filter expressions: `[?search(@.field, 'pattern')]`

## 4. `count()` signature change

**v2.x:**
```go
// count() counted occurrences of a value in an array
// $.store.book[*].category.count('fiction')
```

**v3.0.0:**
```go
// count() counts nodes in a nodelist (RFC 9535 compliant)
// $.store.book.count()
```

**Migration steps:**
- Replace `count('value')` with `occurrences(arr, 'value')` for value counting
- Use `count()` without arguments for nodelist counting

## 5. New `value()` function

v3.0.0 adds the RFC 9535 `value()` function to extract a single value from a nodelist:

```go
// Returns the single value if nodelist has exactly one element
result, err := jsonpath.Query(data, "$.store.book[0].title")
title := result[0].Value // "Book 1"
```

## 6. `Node` type

v3.0.0 introduces the `Node` type with `Location` and `Value` fields:

```go
type Node struct {
    Location string      // Normalized Path (e.g., "$['store']['book'][0]")
    Value    interface{} // The actual value
}
```

## Quick Reference

| v2.x | v3.0.0 |
|------|--------|
| `result.([]interface{})` | `for _, node := range result { node.Value }` |
| `@.field.match('p')` | `match(@.field, 'p')` |
| `@.field.search('p')` | `search(@.field, 'p')` |
| `count('value')` | `occurrences(arr, 'value')` |
| N/A | `value(nodelist)` (new) |
