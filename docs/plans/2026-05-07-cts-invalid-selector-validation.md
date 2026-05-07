# CTS Invalid Selector Validation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Fix 8 CTS test failures where invalid selectors are not being rejected by the parser.

**Architecture:** Add targeted validation in `parser.go` at three points: (1) `createDotSegment` to validate dot-member-name format, (2) `parseBracketSegment` to reject empty brackets and `@`/`$` outside filters, and (3) `parseRecursive` to reject bare `$..`.

**Tech Stack:** Go, standard library

---

## Current State

- CTS pass rate: 58.2% (318 pass, 138 fail, 91/247 invalid correctly rejected)
- Target selectors: `$ `, `$.&`, `$.1`, `$[0 2]`, `$[]`, `$..`, `$[@.a]`, `$[$.a]`

## Root Cause Analysis

| Selector | Current behavior | Fix location |
|----------|-----------------|--------------|
| `$ ` | space becomes `nameSegment{" "}` via `parseRegular` | `createDotSegment` - validate name format |
| `$.&` | `&` becomes `nameSegment{"&"}` via `createDotSegment` | `createDotSegment` - validate name format |
| `$.1` | `1` becomes `nameSegment{"1"}` via `createDotSegment` | `createDotSegment` - validate name format |
| `$[0 2]` | `0 2` becomes `nameSegment{"0 2"}` via `parseIndexOrName` | `parseBracketSegment` - reject invalid content |
| `$[]` | empty becomes `nameSegment{""}` via `parseIndexOrName` | `parseBracketSegment` - reject empty |
| `$..` | bare recursive returns `[recursiveSegment]` | `parseRecursive` - reject empty path |
| `$[@.a]` | `@.a` becomes `nameSegment{"@.a"}` via `parseIndexOrName` | `parseBracketSegment` - reject `@` outside filter |
| `$[$.a]` | `$.a` becomes `nameSegment{"$.a"}` via `parseIndexOrName` | `parseBracketSegment` - reject `$` outside filter |

---

## Task 1: Validate dot-member-name format in `createDotSegment`

**Files:**
- Modify: `parser.go:298-303`

**Step 1: Write the failing test**

Add to `parser_test.go` a test that `parse()` rejects the selectors `$ `, `$.&`, `$.1`:

```go
func TestInvalidDotMemberName(t *testing.T) {
    invalid := []string{"$ ", "$.&", "$.1"}
    for _, sel := range invalid {
        t.Run(sel, func(t *testing.T) {
            _, err := parse(sel)
            if err == nil {
                t.Errorf("parse(%q) should return error", sel)
            }
        })
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test -run TestInvalidDotMemberName -v`
Expected: FAIL - all three return nil error

**Step 3: Write minimal implementation**

Add a `isValidDotMemberName` function and update `createDotSegment`:

```go
func isValidDotMemberName(name string) bool {
    if name == "" {
        return false
    }
    for i, r := range name {
        if i == 0 {
            if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_') {
                return false
            }
        } else {
            if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
                return false
            }
        }
    }
    return true
}

func createDotSegment(name string) (segment, error) {
    if name == "*" {
        return &wildcardSegment{}, nil
    }
    if !isValidDotMemberName(name) {
        return nil, NewError(ErrSyntax, fmt.Sprintf("invalid dot-member-name: %s", name), name)
    }
    return &nameSegment{name: name}, nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test -run TestInvalidDotMemberName -v`
Expected: PASS

**Step 5: Run full test suite**

Run: `go test ./... -count=1`
Expected: PASS - no regressions

---

## Task 2: Reject empty bracket and `@`/`$` outside filter in `parseBracketSegment`

**Files:**
- Modify: `parser.go:306-334` (`parseBracketSegment`)

**Step 1: Write the failing test**

Add to `parser_test.go`:

```go
func TestInvalidBracketSelectors(t *testing.T) {
    invalid := []string{"$[]", "$[@.a]", "$[$.a]", "$[0 2]"}
    for _, sel := range invalid {
        t.Run(sel, func(t *testing.T) {
            _, err := parse(sel)
            if err == nil {
                t.Errorf("parse(%q) should return error", sel)
            }
        })
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test -run TestInvalidBracketSelectors -v`
Expected: FAIL

**Step 3: Write minimal implementation**

In `parseBracketSegment`, add validation after `content = strings.TrimSpace(content)`:

```go
func parseBracketSegment(content string) (segment, error) {
    content = strings.TrimSpace(content)

    // Empty bracket selector is invalid
    if content == "" {
        return nil, NewError(ErrSyntax, "empty bracket selector", "$[]")
    }

    // @ and $ are only valid inside filter expressions
    if strings.HasPrefix(content, "@") || strings.HasPrefix(content, "$") {
        return nil, NewError(ErrSyntax, fmt.Sprintf("@ or $ used outside filter selector: %s", content), content)
    }

    // ... rest of existing code unchanged
}
```

For `$[0 2]`: the content is `"0 2"`. It's not empty and doesn't start with `@`/`$`, so it falls through to `parseIndexOrName("0 2")`. We need additional validation there. In `parseIndexOrName`, if content isn't a valid integer and isn't quoted, reject it:

```go
func parseIndexOrName(content string) (segment, error) {
    if strings.HasSuffix(content, ")") {
        return parseFunctionCall(content)
    }

    if idx, err := strconv.Atoi(content); err == nil {
        return &indexSegment{index: idx}, nil
    }

    if strings.HasPrefix(content, "'") && strings.HasSuffix(content, "'") && len(content) > 1 {
        return &nameSegment{name: content[1 : len(content)-1]}, nil
    }

    if strings.HasPrefix(content, "\"") && strings.HasSuffix(content, "\"") && len(content) > 1 {
        return &nameSegment{name: content[1 : len(content)-1]}, nil
    }

    // Unquoted non-integer content in brackets is invalid
    return nil, NewError(ErrSyntax, fmt.Sprintf("invalid bracket selector: %s", content), content)
}
```

Wait - this would break unquoted names in brackets like `$[name]`. Let me check if that's used...

Actually, looking at `parseMultiIndexSegment` at line 1142-1143, unquoted names ARE handled there. And `parseIndexOrName` is the fallback for single bracket content. We need to be careful.

The safest approach: reject content that contains whitespace (which is invalid in any single selector), and reject bare `@`/`$` prefixed content. For `$[0 2]`, the space makes it invalid.

Actually, re-reading RFC 9535 more carefully:
- In brackets, valid selectors are: `index-selector` (integer), `name-selector` (quoted string), `slice-selector` (int:int[:int]), `wildcard-selector` (`*`), `filter-selector` (`?...`)
- An unquoted name like `$[name]` is NOT valid per RFC 9535

But the existing tests may rely on unquoted names in brackets. Let me keep the current behavior for unquoted names and only fix the specific CTS failures:

1. `$[]` - empty → reject
2. `$[@.a]` - starts with `@` → reject
3. `$[$.a]` - starts with `$` → reject
4. `$[0 2]` - contains space and isn't valid → reject in `parseIndexOrName`

For `$[0 2]` specifically, in `parseIndexOrName`:
- `strconv.Atoi("0 2")` fails
- Not quoted
- Falls through to `return &nameSegment{name: content}`

The fix for `$[0 2]`: reject content that contains whitespace in `parseIndexOrName`. Or validate more broadly that bracket content is valid.

Actually, the cleanest fix: in `parseBracketSegment`, after the existing checks and before falling through to `parseIndexOrName`, validate that the content looks like a valid selector. Content with spaces that isn't a filter, multi-index, or slice is invalid.

Let me revise: add a check in `parseBracketSegment` for content containing spaces (after all the specific checks have been done and it's about to fall through to `parseIndexOrName`):

```go
// Before the final return parseIndexOrName(content):
if strings.ContainsAny(content, " \t\n\r") {
    return nil, NewError(ErrSyntax, fmt.Sprintf("invalid bracket selector: %s", content), content)
}
return parseIndexOrName(content)
```

**Step 4: Run test to verify it passes**

Run: `go test -run TestInvalidBracketSelectors -v`
Expected: PASS

**Step 5: Run full test suite**

Run: `go test ./... -count=1`
Expected: PASS

---

## Task 3: Reject bare `$..` recursive descent

**Files:**
- Modify: `parser.go:190-212` (`parseRecursive`)

**Step 1: Write the failing test**

```go
func TestBareRecursiveDescent(t *testing.T) {
    _, err := parse("$..")
    if err == nil {
        t.Error("parse(\"$..\") should return error")
    }
}
```

**Step 2: Run test to verify it fails**

Run: `go test -run TestBareRecursiveDescent -v`
Expected: FAIL

**Step 3: Write minimal implementation**

In `parseRecursive`, change the empty path check to return an error:

```go
func parseRecursive(path string) ([]segment, error) {
    var segments []segment
    segments = append(segments, &recursiveSegment{})

    if path == "" {
        return nil, NewError(ErrSyntax, "bare recursive descent segment", "$..")
    }

    path = strings.TrimPrefix(path, ".")

    if path == "" {
        return nil, NewError(ErrSyntax, "bare recursive descent segment", "$..")
    }

    remainingSegments, err := parseRegular(path)
    if err != nil {
        return nil, err
    }
    segments = append(segments, remainingSegments...)

    return segments, nil
}
```

**Step 4: Update existing test**

The existing `TestParseRecursive` has a test case "empty path" that expects `wantLen: 1` (success). This needs to change to expect an error:

```go
{
    name:        "empty path",
    path:        "",
    wantErr:     true,
    errType:     ErrSyntax,
    errContains: "bare recursive descent",
},
```

**Step 5: Run tests**

Run: `go test -run "TestBareRecursiveDescent|TestParseRecursive" -v`
Expected: PASS

**Step 6: Run full test suite**

Run: `go test ./... -count=1`
Expected: PASS

---

## Task 4: Run CTS and verify improvements

**Step 1: Run CTS tests**

Run: `go test -run TestCTS -v -count=1 2>&1 | grep -E "(CTS RESULTS|Invalid selectors|TOTAL)"`

Expected: Invalid selectors correctly rejected increases from 91 to 99 (8 new rejections), total pass rate increases.

**Step 2: Run full test suite**

Run: `go test ./... -count=1`
Expected: All tests pass, no regressions

**Step 3: Commit**

```bash
git add parser.go parser_test.go
git commit -fix: reject 8 invalid selector patterns in parser
```
