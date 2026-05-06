package jsonpath

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"testing"
)

// TestCTSFailureAnalysis produces a detailed analysis of all CTS failures
func TestCTSFailureAnalysis(t *testing.T) {
	suite := loadCTS(t)

	type failureInfo struct {
		name     string
		selector string
		reason   string
		category string
	}

	var failures []failureInfo

	for _, tc := range suite.Tests {
		cat := extractCategory(tc.Name)

		if tc.InvalidSelector {
			var doc interface{}
			if len(tc.Document) > 0 {
				_ = json.Unmarshal(tc.Document, &doc)
			} else {
				doc = map[string]interface{}{}
			}
			_, err := Query(doc, tc.Selector)
			if err == nil {
				failures = append(failures, failureInfo{
					name:     tc.Name,
					selector: tc.Selector,
					reason:   "INVALID_NOT_CAUGHT",
					category: cat,
				})
			}
			continue
		}

		var doc interface{}
		if err := json.Unmarshal(tc.Document, &doc); err != nil {
			continue
		}

		actual, err := Query(doc, tc.Selector)
		if err != nil {
			reason := fmt.Sprintf("ERROR: %v", err)
			failures = append(failures, failureInfo{
				name:     tc.Name,
				selector: tc.Selector,
				reason:   reason,
				category: cat,
			})
			continue
		}

		actualNodelist := wrapInNodelist(actual)

		matched := false
		if len(tc.Results) > 0 {
			matched = resultMatchesAny(actualNodelist, tc.Results)
		} else if len(tc.Result) > 0 {
			expected, parseErr := parseCTSResult(tc.Result)
			if parseErr == nil {
				if expectedArr, ok := expected.([]interface{}); ok {
					matched = nodelistsEqual(actualNodelist, expectedArr)
				}
			}
		}

		if !matched {
			var expectedStr string
			if len(tc.Results) > 0 {
				expectedStr = "one of: " + strings.Join(func() []string {
					var s []string
					for _, r := range tc.Results {
						s = append(s, string(r))
					}
					return s
				}(), " | ")
			} else {
				expectedStr = string(tc.Result)
			}
			reason := fmt.Sprintf("MISMATCH: expected %s, got %v", expectedStr, actualNodelist)
			failures = append(failures, failureInfo{
				name:     tc.Name,
				selector: tc.Selector,
				reason:   reason,
				category: cat,
			})
		}
	}

	// Classify failure root causes
	type rootCause struct {
		name        string
		description string
		count       int
		selectors   []string
	}

	causes := make(map[string]*rootCause)

	classify := func(f failureInfo) string {
		sel := f.selector
		switch {
		// Root identifier issues
		case sel == "$" && strings.Contains(f.reason, "MISMATCH"):
			return "ROOT_NODELIST_FORMAT"

		// Missing field should return empty, not error
		case strings.Contains(f.reason, "field") && strings.Contains(f.reason, "not found"):
			return "MISSING_FIELD_RETURNS_ERROR"

		// Type mismatch errors - accessing wrong type
		case strings.Contains(f.reason, "value is not an object"):
			return "WRONG_TYPE_ACCESS_OBJECT"
		case strings.Contains(f.reason, "value is not an array"):
			return "WRONG_TYPE_ACCESS_ARRAY"
		case strings.Contains(f.reason, "value is neither array nor object"):
			return "WRONG_TYPE_ACCESS_WILDCARD"

		// Mixed selector types not supported
		case strings.Contains(f.reason, "cannot mix numeric indices"):
			return "MIXED_SELECTOR_TYPES"

		// Multi-name on non-object
		case strings.Contains(f.reason, "multi-name can only be applied to object"):
			return "MULTI_NAME_ON_ARRAY"

		// Filter existence not supported
		case strings.Contains(f.reason, "no valid operator found in condition"):
			return "FILTER_EXISTENCE_UNSUPPORTED"

		// Filter syntax issues
		case strings.Contains(f.reason, "invalid filter syntax"):
			return "FILTER_SYNTAX"

		// Comparison operator issues
		case strings.Contains(f.reason, "invalid operator"):
			return "FILTER_OPERATOR_TYPE_MISMATCH"

		// Invalid selector not caught
		case strings.Contains(f.reason, "INVALID_NOT_CAUGHT"):
			return "INVALID_NOT_CAUGHT"

		// Descendant wildcard issues (typically after collecting primitives)
		case strings.Contains(sel, "..") && (strings.Contains(sel, "*") || strings.Contains(sel, "[*]")):
			return "DESCENDANT_WILDCARD_ON_PRIMITIVES"

		// Filter with functions in CTS syntax
		case strings.Contains(sel, "count(") || strings.Contains(sel, "length(") || strings.Contains(sel, "match(") || strings.Contains(sel, "search(") || strings.Contains(sel, "value("):
			return "FILTER_FUNCTION_UNSUPPORTED"

		// Whitespace handling
		case strings.HasPrefix(f.category, "whitespace"):
			return "WHITESPACE_HANDLING"

		// Double-quoted name selectors
		case strings.Contains(sel, `["`) && !strings.Contains(f.reason, "INVALID_NOT_CAUGHT"):
			return "DOUBLE_QUOTED_NAME"

		// Name selector with special chars/escapes
		case (strings.Contains(sel, `$['`) || strings.Contains(sel, `["`)) && !strings.Contains(f.reason, "INVALID_NOT_CAUGHT"):
			return "NAME_SELECTOR_SPECIAL_CHARS"

		default:
			return "OTHER"
		}
	}

	for _, f := range failures {
		cause := classify(f)
		if _, ok := causes[cause]; !ok {
			desc := cause
			switch cause {
			case "ROOT_NODELIST_FORMAT":
				desc = "$ should return nodelist [root], not unwrapped root"
			case "MISSING_FIELD_RETURNS_ERROR":
				desc = "Missing field should return empty nodelist, not error"
			case "WRONG_TYPE_ACCESS_OBJECT":
				desc = "Name access on non-object should return empty, not error"
			case "WRONG_TYPE_ACCESS_ARRAY":
				desc = "Index access on non-array should return empty, not error"
			case "WRONG_TYPE_ACCESS_WILDCARD":
				desc = "Wildcard on primitive (after descendant) should return empty, not error"
			case "MIXED_SELECTOR_TYPES":
				desc = "Mixed selector types (index+wildcard, index+slice, etc.) in one bracket not supported"
			case "MULTI_NAME_ON_ARRAY":
				desc = "Multi-name selector on array should return empty, not error"
			case "FILTER_EXISTENCE_UNSUPPORTED":
				desc = "Filter existence test ($[?@.a]) not supported - no comparison operator required"
			case "FILTER_SYNTAX":
				desc = "Filter syntax not supported (absolute paths $, etc.)"
			case "FILTER_OPERATOR_TYPE_MISMATCH":
				desc = "Comparison between different types should return false, not error"
			case "INVALID_NOT_CAUGHT":
				desc = "Invalid selector not rejected by parser"
			case "DESCENDANT_WILDCARD_ON_PRIMITIVES":
				desc = "Descendant wildcard hits primitive values and errors"
			case "FILTER_FUNCTION_UNSUPPORTED":
				desc = "Filter function call syntax not supported in CTS format"
			case "WHITESPACE_HANDLING":
				desc = "Whitespace handling differs from RFC 9535 spec"
			case "DOUBLE_QUOTED_NAME":
				desc = "Double-quoted name selector issues"
			case "NAME_SELECTOR_SPECIAL_CHARS":
				desc = "Name selector with special characters/escapes"
			}
			causes[cause] = &rootCause{name: cause, description: desc}
		}
		causes[cause].count++
		if len(causes[cause].selectors) < 5 {
			causes[cause].selectors = append(causes[cause].selectors, f.selector)
		}
	}

	// Sort by count descending
	var causeKeys []string
	for k := range causes {
		causeKeys = append(causeKeys, k)
	}
	sort.Slice(causeKeys, func(i, j int) bool {
		return causes[causeKeys[i]].count > causes[causeKeys[j]].count
	})

	t.Logf("\n========== FAILURE ROOT CAUSE ANALYSIS ==========")
	t.Logf("Total failures: %d out of 703 tests\n", len(failures))
	t.Logf("%-35s %5s  %s", "Root Cause", "Count", "Description")
	t.Logf("%-35s %5s  %s", "----------", "-----", "-----------")
	for _, k := range causeKeys {
		c := causes[k]
		t.Logf("%-35s %5d  %s", c.name, c.count, c.description)
	}

	t.Logf("\n========== EXAMPLES PER ROOT CAUSE ==========")
	for _, k := range causeKeys {
		c := causes[k]
		t.Logf("\n[%s] (%d failures)", c.name, c.count)
		for _, sel := range c.selectors {
			t.Logf("  %s", sel)
		}
	}
}
