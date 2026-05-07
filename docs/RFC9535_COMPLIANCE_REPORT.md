# RFC 9535 Compliance Assessment Report

Date: 2026-05-06

Based on a line-by-line audit of all source files (`parser.go`, `segments.go`, `functions.go`, `jsonpath.go`, `types.go`, `errors.go`, `stringCompare.go`) and both README files against [RFC 9535](https://www.rfc-editor.org/rfc/rfc9535).

---

## 1. README Examples Compliance

| # | Example | RFC 9535 Status | Issue |
|---|---------|----------------|-------|
| 1 | `$.store.book[*].author` | COMPLIANT | |
| 2 | `$.store.book[?(@.price > 10)]` | COMPLIANT | `paren-expr` is valid syntax per Table 12 |
| 3 | `$.store.book[0]` / `[-1]` | COMPLIANT | |
| 4 | `$.store.book[0:2]` | COMPLIANT | |
| 5 | `$..price` | DEFECTIVE | Root node itself is missed by recursive descent |
| 6 | `$.store.book[?(@.price > 10 && @.category == 'fiction')]` | COMPLIANT syntax | Implementation has `&&`/`||` precedence bug |
| 7 | `$.store.book[?(!(@.category == 'reference'))]` | COMPLIANT | |
| 8 | `$.store.book.length()` | COMPLIANT | Implementation counts bytes, not characters |
| 9 | `$.store.keys()` | NON-STANDARD | Extension function, should be labeled |
| 10 | `$.store.values()` | NON-STANDARD | Extension function, should be labeled |
| 11 | `$.store.book[*].price.min()` | NON-STANDARD | Extension function, should be labeled |
| 12 | `$.store.book[*].price.max()` | NON-STANDARD | Extension function, should be labeled |
| 13 | `$.store.book[*].price.avg()` | NON-STANDARD | Extension function, should be labeled |
| 14 | `$.store.book[*].price.sum()` | NON-STANDARD | Extension function, should be labeled |
| 15 | `$.store.book[*].category.count('fiction')` | WRONG SIGNATURE | RFC `count(NodesType)->ValueType`, not this usage |
| 16 | `$.store.book[?@.title.match('^S.*')]` | SYNTAX DIFFERENCE | RFC syntax is `?match(@.title, '^S.*')`; method-chain is non-standard |
| 17 | `$.store.book[*].title.search('^S.*')` | WRONG SIGNATURE | RFC `search(ValueType,ValueType)->LogicalType` (string+regex->boolean), not array filter |
| 18 | `$.store.book[*]['author','price']` | COMPLIANT | |
| 19 | `$.store.book[?@.price > $.store.book[*].price.avg()]` | NOT SUPPORTED | Implementation does not support filter-query comparisons |

**Summary**: 10 compliant, 3 defective/implementation bugs, 6 non-standard/wrong.

---

## 2. Code Implementation Compliance

### A. Selectors and Segments

| RFC Section | Requirement | Current Status | Severity |
|-------------|-------------|----------------|----------|
| 2.3.1 | Name selector: non-object returns empty nodelist | Returns error | HIGH |
| 2.3.1 | Name selector: missing field returns empty nodelist | Returns error | HIGH |
| 2.3.1 | String escape sequences `\b \f \n \r \t \" \' \/ \\ \uXXXX` | Not implemented | HIGH |
| 2.3.2 | Wildcard on primitive returns empty nodelist | Returns error | HIGH |
| 2.3.3 | Index on non-array returns empty nodelist | Returns error | HIGH |
| 2.3.3 | Index: no leading zeros (`[07]` invalid) | Not validated; `07` parsed as `7` | LOW |
| 2.3.4 | Slice on non-array returns empty nodelist | Returns error | HIGH |
| 2.3.4 | Slice `step=0` returns empty nodelist | Returns error | HIGH |
| 2.3.4 | Slice defaults: `[:]` vs explicit `[0:0]` ambiguity | Uses `start=0, end=0` as sentinel, ambiguous with explicit zero | MEDIUM |
| 2.3.6 | `..` must be followed by a selector | Bare `..` is accepted | LOW |
| 2.3.6 | `..` recursive descent includes root node itself | Only collects children and descendants, root is missed | HIGH |
| 2.3.4 | Mixed selectors in brackets `[sel, sel]` | Partially compliant | MEDIUM |

### B. Filter Expressions

| RFC Section | Requirement | Current Status | Severity |
|-------------|-------------|----------------|----------|
| 2.3.5 | Existence test `[?@.name]` (no comparison operator) | Not supported; "no valid operator found" error | HIGH |
| 2.3.5 | Filter-query comparisons `[?@.a == $.root]` | Not supported; only field vs literal | HIGH |
| 2.3.5 | `&&` has higher precedence than `\|\|` | Evaluated strictly left-to-right, no precedence | HIGH |
| 2.3.5 | Filter evaluates ALL array elements (including primitives) | Only evaluates `map[string]interface{}` elements; skips primitives | HIGH |
| 2.3.5 | Filter field references support full sub-queries | Only supports simple dot-path splitting; no `@.items[0]`, `@['key']` | HIGH |
| 2.3.5.1 | Cross-type comparison returns `false` | Returns error | MEDIUM |
| 2.3.5.1 | Boolean ordered comparison (`<`) returns `false` | Returns error | MEDIUM |
| 2.3.5.1 | Nothing in comparisons: `nothing == nothing` -> true | No Nothing concept exists | MEDIUM |
| 2.3.5 | `!` logical NOT (general unary prefix) | Uses De Morgan's law transformation; fragile for complex expressions | MEDIUM |
| 2.3.5 | Parenthesized expressions | Partially compliant; fragile nesting | MEDIUM |

### C. Required Functions (all 5 have issues)

| Function | RFC Requirement | Current Implementation | Severity |
|----------|----------------|----------------------|----------|
| `length(ValueType)->ValueType` | String: count **characters** | Counts **bytes** (`len(str)`) | HIGH |
| `length(ValueType)->ValueType` | Other types: return Nothing | Returns error | HIGH |
| `count(NodesType)->ValueType` | Takes nodelist, returns node count | Takes `(array, value)`, counts occurrences -- **completely different signature** | CRITICAL |
| `match(ValueType,ValueType)->LogicalType` | Uses I-Regexp (RFC 9485) | Uses Go `regexp` (superset, not fully compatible) | MEDIUM |
| `search(ValueType,ValueType)->LogicalType` | Takes string+regex, returns boolean | Takes array+regex, returns matching elements -- **completely different signature** | CRITICAL |
| `value(NodesType)->ValueType` | Takes nodelist, returns single node value | **Not implemented at all** | CRITICAL |

### D. Data Model and Type System

| RFC Section | Requirement | Current Status | Severity |
|-------------|-------------|----------------|----------|
| 2.6 | Node type = location + value | No Node type; returns bare values | HIGH |
| 2.6 | Nothing type (distinct from null) | No Nothing concept | MEDIUM |
| 2.4 | ValueType / LogicalType / NodesType distinction | All use `interface{}` | MEDIUM |
| 2.7 | Normalized Path generation | Not implemented | HIGH |
| 2.1 | Query returns Nodelist | Returns unwrapped single value or `[]interface{}` | HIGH |

### E. Other Requirements

| RFC Section | Requirement | Current Status | Severity |
|-------------|-------------|----------------|----------|
| 2.3.3 | Integer range `[-(2^53)+1, (2^53)-1]` | Not validated | LOW |
| 2.6.1 | Array order preservation | COMPLIANT | -- |
| 2.6.1 | Duplicate node preservation | COMPLIANT | -- |
| 2.6.1 | Object member ordering unspecified | COMPLIANT (Go map iteration is non-deterministic) | -- |

---

## 3. Summary Statistics

| Category | Compliant | Partially Compliant | Non-Compliant |
|----------|-----------|-------------------|---------------|
| Selectors and Segments | 3 | 2 | 8 |
| Filter Expressions | 1 | 3 | 6 |
| Required Functions | 0 | 1 | 4 |
| Data Model | 0 | 1 | 4 |
| Other | 2 | 2 | 2 |
| **Total** | **6** | **9** | **24** |

---

## 4. Architectural Root Causes

Three fundamental architectural gaps prevent incremental RFC compliance:

### 4.1 Missing Node Type

RFC 9535's core data model is `Node = (location, value)`. All segments return `Nodelist` (ordered list of Nodes). The current implementation returns bare `interface{}` values, losing location information. This blocks:
- Normalized Path generation (§2.7)
- `count(NodesType)` function (§2.4.4)
- `value(NodesType)` function (§2.4.7)
- Proper nodelist semantics throughout

### 4.2 Error vs Empty Nodelist Confusion

RFC 9535 distinguishes:
- **Invalid query** (syntax/semantic error) -> MUST raise error
- **Structural mismatch** (valid query, data doesn't match) -> empty nodelist, NOT an error

The current implementation uses Go `error` for both cases, making it impossible for callers to distinguish "bad query" from "no match". This affects: name on non-object, index on non-array, wildcard on primitive, slice on non-array, missing fields.

### 4.3 Function Signature Divergence

`count` and `search` have completely different signatures from RFC 9535. Fixing them would break existing users. Options:
- Rename current implementations to non-standard names (e.g., `occurrences`, `filterMatch`)
- Add RFC-compliant versions alongside (risk of confusion)
- Major version bump with breaking changes

---

## 5. README Declaration Issues

Current README (line 10):
> "A complete Go implementation of JSONPath that **fully complies** with RFC 9535"

Current README (line 14):
> "Complete RFC 9535 Implementation"

These declarations are not accurate. Recommended revision:
> "A Go implementation of JSONPath based on RFC 9535, with extensions"

The Features section should clearly separate:
- RFC 9535 standard features
- Extension features (keys, values, min, max, avg, sum, custom count, custom search)

---

## 6. Compliant Items (for reference)

These items are correctly implemented:
- `$` root identifier (parser.go:17-21)
- Comparison operators `== != < <= > >=` (parser.go:437)
- Out-of-range index returns empty nodelist (segments.go:190-191)
- Filter on primitive returns empty nodelist (segments.go:467-501)
- Array order preservation (segments.go:213, 435-439)
- Duplicate node preservation (jsonpath.go:26-36)
- Negative index support (segments.go:183-195)
- `match()` function basic operation (functions.go:488-591)
- `length()` on arrays and objects (functions.go:182-203)
- Child segment bracket and dot notation (parser.go:66-134)
- Descendant segment basic recursion (segments.go:234-276)
- Slice basic parsing and evaluation (parser.go:787-826, segments.go:306-445)
- Filter `&&`, `||`, `!` basic operation (parser.go:276-393, segments.go:517-534)
- Parenthesized filter expressions (parser.go:340-360)
- `$` and `@` identifier recognition (parser.go:17-21, 398-404)
