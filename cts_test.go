package jsonpath

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// CTS test suite structures
type ctsTest struct {
	Name            string           `json:"name"`
	Selector        string           `json:"selector"`
	Document        json.RawMessage  `json:"document"`
	Result          json.RawMessage  `json:"result"`
	Results         []json.RawMessage `json:"results"`
	ResultPaths     []string         `json:"result_paths"`
	InvalidSelector bool             `json:"invalid_selector"`
	Tags            []string         `json:"tags"`
}

type ctsSuite struct {
	Description string    `json:"description"`
	Tests       []ctsTest `json:"tests"`
}

func loadCTS(t *testing.T) *ctsSuite {
	t.Helper()
	data, err := os.ReadFile("cts.json")
	if err != nil {
		t.Skip("cts.json not found, skipping CTS tests")
	}
	var suite ctsSuite
	if err := json.Unmarshal(data, &suite); err != nil {
		t.Fatalf("Failed to parse cts.json: %v", err)
	}
	return &suite
}

// normalizeForCompare recursively normalizes JSON values for comparison
func normalizeForCompare(v interface{}) interface{} {
	switch val := v.(type) {
	case []interface{}:
		normalized := make([]interface{}, len(val))
		for i, elem := range val {
			normalized[i] = normalizeForCompare(elem)
		}
		return normalized
	case map[string]interface{}:
		normalized := make(map[string]interface{}, len(val))
		for k, v := range val {
			normalized[k] = normalizeForCompare(v)
		}
		return normalized
	case float64:
		// Normalize integer-valued floats
		if val == float64(int64(val)) {
			return val
		}
		return val
	default:
		return v
	}
}

// wrapInNodelist wraps the Query() result into a nodelist (array) format
// that matches the CTS expected output.
// Query() returns a NodeList ([]Node), each Node having Location and Value.
// CTS expects a flat array of values (not Node structs).
func wrapInNodelist(result interface{}) []interface{} {
	if result == nil {
		return []interface{}{}
	}
	switch v := result.(type) {
	case NodeList:
		out := make([]interface{}, len(v))
		for i, n := range v {
			out[i] = n.Value
		}
		return out
	case []interface{}:
		return v
	default:
		return []interface{}{result}
	}
}

// parseCTSResult parses the expected result from CTS JSON
func parseCTSResult(raw json.RawMessage) (interface{}, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	var result interface{}
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("failed to parse expected result: %v", err)
	}
	return normalizeForCompare(result), nil
}

// deepEqual compares two values for deep equality
func deepEqual(a, b interface{}) bool {
	return reflect.DeepEqual(normalizeForCompare(a), normalizeForCompare(b))
}

// resultMatchesAny checks if the actual nodelist matches any of the valid non-deterministic results
func resultMatchesAny(actualNodelist []interface{}, expectedResults []json.RawMessage) bool {
	for _, raw := range expectedResults {
		expected, err := parseCTSResult(raw)
		if err != nil {
			continue
		}
		expectedArr, ok := expected.([]interface{})
		if !ok {
			continue
		}
		if nodelistsEqual(actualNodelist, expectedArr) {
			return true
		}
	}
	return false
}

// nodelistsEqual compares two nodelists, handling non-deterministic object member order
func nodelistsEqual(actual, expected []interface{}) bool {
	if len(actual) != len(expected) {
		return false
	}
	for i := range actual {
		if !deepEqual(actual[i], expected[i]) {
			return false
		}
	}
	return true
}

func TestCTS(t *testing.T) {
	suite := loadCTS(t)

	var pass, fail, skip, invalidPass, invalidFail int
	var failures []string
	var failureCategories map[string][]string = make(map[string][]string)

	for _, tc := range suite.Tests {
		t.Run(tc.Name, func(t *testing.T) {
			// Handle invalid selector tests
			if tc.InvalidSelector {
				// Some invalid selector tests have no document
				if len(tc.Document) > 0 {
					var doc interface{}
					_ = json.Unmarshal(tc.Document, &doc)
					_, err := Query(doc, tc.Selector)
					if err != nil {
						invalidPass++
					} else {
						invalidFail++
						failures = append(failures, fmt.Sprintf("FAIL [invalid_not_caught] %s: selector %q should be invalid but got result", tc.Name, tc.Selector))
						cat := extractCategory(tc.Name)
						failureCategories[cat] = append(failureCategories[cat], tc.Selector)
					}
				} else {
					// No document - just test that the selector is rejected
					_, err := Query(map[string]interface{}{}, tc.Selector)
					if err != nil {
						invalidPass++
					} else {
						invalidFail++
						failures = append(failures, fmt.Sprintf("FAIL [invalid_not_caught] %s: selector %q should be invalid but got result", tc.Name, tc.Selector))
						cat := extractCategory(tc.Name)
						failureCategories[cat] = append(failureCategories[cat], tc.Selector)
					}
				}
				return
			}

			// Parse the input document
			var doc interface{}
			if err := json.Unmarshal(tc.Document, &doc); err != nil {
				t.Fatalf("Failed to parse document: %v", err)
			}

			// Run the query
			actual, err := Query(doc, tc.Selector)

			if err != nil {
				// Valid selector but our implementation errored
				fail++
				failures = append(failures, fmt.Sprintf("FAIL [error] %s: selector %q got error: %v", tc.Name, tc.Selector, err))
				cat := extractCategory(tc.Name)
				failureCategories[cat] = append(failureCategories[cat], tc.Selector)
				return
			}

			// Wrap actual result into nodelist format
			actualNodelist := wrapInNodelist(actual)

			// Check against expected result(s)
			if len(tc.Results) > 0 {
				// Non-deterministic result - must match at least one alternative
				if resultMatchesAny(actualNodelist, tc.Results) {
					pass++
				} else {
					fail++
					var expectedStrs []string
					for _, raw := range tc.Results {
						expectedStrs = append(expectedStrs, string(raw))
					}
					failures = append(failures, fmt.Sprintf("FAIL [mismatch] %s: selector %q\n  expected one of: %s\n  got:      %v", tc.Name, tc.Selector, strings.Join(expectedStrs, " | "), actualNodelist))
					cat := extractCategory(tc.Name)
					failureCategories[cat] = append(failureCategories[cat], tc.Selector)
				}
			} else if len(tc.Result) > 0 {
				// Deterministic result
				expected, parseErr := parseCTSResult(tc.Result)
				if parseErr != nil {
					skip++
					t.Logf("SKIP: failed to parse expected result: %v", parseErr)
					return
				}

				expectedArr, ok := expected.([]interface{})
				if !ok {
					skip++
					t.Logf("SKIP: expected result is not an array")
					return
				}

				if nodelistsEqual(actualNodelist, expectedArr) {
					pass++
				} else {
					fail++
					failures = append(failures, fmt.Sprintf("FAIL [mismatch] %s: selector %q\n  expected: %v\n  got:      %v", tc.Name, tc.Selector, expectedArr, actualNodelist))
					cat := extractCategory(tc.Name)
					failureCategories[cat] = append(failureCategories[cat], tc.Selector)
				}
			} else {
				skip++
			}
		})
	}

	// Print summary
	t.Logf("\n========== CTS RESULTS ==========")
	t.Logf("Valid selectors - Pass: %d, Fail: %d, Skip: %d", pass, fail, skip)
	t.Logf("Invalid selectors - Correctly rejected: %d, Not caught: %d", invalidPass, invalidFail)
	total := pass + fail + skip + invalidPass + invalidFail
	rate := float64(0)
	if pass+fail+invalidPass+invalidFail > 0 {
		rate = float64(pass+invalidPass) / float64(pass+fail+invalidPass+invalidFail) * 100
	}
	t.Logf("Total: %d/%d passed (%.1f%%)", pass+invalidPass, total, rate)

	if len(failureCategories) > 0 {
		t.Logf("\n========== FAILURES BY CATEGORY ==========")
		var cats []string
		for c := range failureCategories {
			cats = append(cats, c)
		}
		sort.Strings(cats)
		for _, cat := range cats {
			sels := failureCategories[cat]
			t.Logf("\n[%s] (%d failures)", cat, len(sels))
			limit := len(sels)
			if limit > 10 {
				limit = 10
			}
			for i := 0; i < limit; i++ {
				t.Logf("  %s", sels[i])
			}
			if len(sels) > 10 {
				t.Logf("  ... and %d more", len(sels)-10)
			}
		}
	}

	if len(failures) > 0 {
		t.Logf("\n========== DETAILED FAILURES (first 30) ==========")
		limit := len(failures)
		if limit > 30 {
			limit = 30
		}
		for i := 0; i < limit; i++ {
			t.Log(failures[i])
		}
		if len(failures) > 30 {
			t.Logf("... and %d more failures", len(failures)-30)
		}
	}
}

func extractCategory(name string) string {
	for _, prefix := range []string{"basic", "filter", "index", "name selector", "slice", "functions", "whitespace", "descendant"} {
		if strings.HasPrefix(name, prefix) {
			return prefix
		}
	}
	return "other"
}

// TestCTSSummary runs the CTS and produces a categorized summary
func TestCTSSummary(t *testing.T) {
	suite := loadCTS(t)

	type categoryStats struct {
		total   int
		pass    int
		fail    int
		skip    int
		invPass int
		invFail int
	}

	categories := make(map[string]*categoryStats)

	for _, tc := range suite.Tests {
		cat := extractCategory(tc.Name)

		stats, ok := categories[cat]
		if !ok {
			stats = &categoryStats{}
			categories[cat] = stats
		}
		stats.total++

		// Handle invalid selector tests
		if tc.InvalidSelector {
			if len(tc.Document) > 0 {
				var doc interface{}
				_ = json.Unmarshal(tc.Document, &doc)
				_, err := Query(doc, tc.Selector)
				if err != nil {
					stats.invPass++
				} else {
					stats.invFail++
				}
			} else {
				_, err := Query(map[string]interface{}{}, tc.Selector)
				if err != nil {
					stats.invPass++
				} else {
					stats.invFail++
				}
			}
			continue
		}

		var doc interface{}
		if err := json.Unmarshal(tc.Document, &doc); err != nil {
			stats.skip++
			continue
		}

		actual, err := Query(doc, tc.Selector)

		if err != nil {
			stats.fail++
			continue
		}

		actualNodelist := wrapInNodelist(actual)

		if len(tc.Results) > 0 {
			if resultMatchesAny(actualNodelist, tc.Results) {
				stats.pass++
			} else {
				stats.fail++
			}
		} else if len(tc.Result) > 0 {
			expected, parseErr := parseCTSResult(tc.Result)
			if parseErr != nil {
				stats.skip++
			} else if expectedArr, ok := expected.([]interface{}); ok && nodelistsEqual(actualNodelist, expectedArr) {
				stats.pass++
			} else {
				stats.fail++
			}
		} else {
			stats.skip++
		}
	}

	// Print categorized summary
	t.Logf("\n========== CTS SUMMARY BY CATEGORY ==========")
	t.Logf("%-20s %5s %5s %5s %5s %5s %5s %6s", "Category", "Total", "Pass", "Fail", "Skip", "InvOk", "InvFl", "Rate")

	var cats []string
	for c := range categories {
		cats = append(cats, c)
	}
	sort.Strings(cats)

	totalPass, totalFail, totalSkip, totalInvPass, totalInvFail := 0, 0, 0, 0, 0
	for _, c := range cats {
		s := categories[c]
		rate := float64(0)
		if s.pass+s.fail+s.invPass+s.invFail > 0 {
			rate = float64(s.pass+s.invPass) / float64(s.pass+s.fail+s.invPass+s.invFail) * 100
		}
		t.Logf("%-20s %5d %5d %5d %5d %5d %5d %5.1f%%", c, s.total, s.pass, s.fail, s.skip, s.invPass, s.invFail, rate)
		totalPass += s.pass
		totalFail += s.fail
		totalSkip += s.skip
		totalInvPass += s.invPass
		totalInvFail += s.invFail
	}

	totalRate := float64(0)
	if totalPass+totalFail+totalInvPass+totalInvFail > 0 {
		totalRate = float64(totalPass+totalInvPass) / float64(totalPass+totalFail+totalInvPass+totalInvFail) * 100
	}
	t.Logf("%-20s %5d %5d %5d %5d %5d %5d %5.1f%%", "TOTAL", len(suite.Tests), totalPass, totalFail, totalSkip, totalInvPass, totalInvFail, totalRate)
}
