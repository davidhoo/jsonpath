package jsonpath

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

type RFC9535TestCase struct {
	Name        string        `json:"name"`
	Selector    string        `json:"selector"`
	Document    interface{}   `json:"document"`
	Expected    interface{}   `json:"result"`
	Results     []interface{} `json:"results,omitempty"`
	ResultPaths interface{}   `json:"result_paths,omitempty"`
	Invalid     bool          `json:"invalid_selector,omitempty"`
}

type RFC9535TestSuite struct {
	Tests []RFC9535TestCase `json:"tests"`
}

func loadRFC9535TestSuite(t *testing.T) *RFC9535TestSuite {
	t.Helper()

	path := "testdata/rfc9535/testdata.json"

	data, err := os.ReadFile(path)
	if err != nil {
		t.Skip("RFC 9535 test data not found, skipping tests")
		return nil
	}

	var suite RFC9535TestSuite
	if err := json.Unmarshal(data, &suite); err != nil {
		t.Fatalf("Failed to parse test data: %v", err)
	}

	return &suite
}

func TestRFC9535Suite_Parse(t *testing.T) {
	suite := loadRFC9535TestSuite(t)

	if len(suite.Tests) == 0 {
		t.Fatal("Test suite has no test cases")
	}

	t.Logf("Loaded %d test cases from RFC 9535 test suite", len(suite.Tests))

	resultCount := 0
	invalidCount := 0
	for _, tc := range suite.Tests {
		if tc.Invalid {
			invalidCount++
		} else {
			resultCount++
		}
	}
	t.Logf("  - %d with result", resultCount)
	t.Logf("  - %d invalid selectors", invalidCount)
}

func TestRFC9535Suite(t *testing.T) {
	suite := loadRFC9535TestSuite(t)
	if suite == nil {
		return
	}

	passCount := 0
	failCount := 0
	skipCount := 0

	for _, test := range suite.Tests {
		t.Run(test.Name, func(t *testing.T) {
			if test.Invalid {
				_, err := Query(test.Document, test.Selector)
				if err == nil {
					t.Logf("Expected error for invalid selector %q, got nil", test.Selector)
					failCount++
				} else {
					passCount++
				}
				return
			}

			result, err := Query(test.Document, test.Selector)
			if err != nil {
				t.Logf("Unexpected error for selector %q: %v", test.Selector, err)
				failCount++
				return
			}

			// RFC 9535 expects results as a nodelist (array).
			// Query() returns single value for single results; normalize to array.
			normalizedResult := result
			if _, isArray := result.([]interface{}); !isArray {
				normalizedResult = []interface{}{result}
			}

			resultJSON, _ := json.Marshal(normalizedResult)

			// Check if test has multiple valid results (results field)
			if len(test.Results) > 0 {
				matched := false
				for _, validResult := range test.Results {
					validJSON, _ := json.Marshal(validResult)
					if string(resultJSON) == string(validJSON) {
						matched = true
						break
					}
				}
				if matched {
					passCount++
				} else {
					t.Logf("Selector %q: result %s matches none of the expected results", test.Selector, resultJSON)
					failCount++
				}
			} else {
				expectedJSON, _ := json.Marshal(test.Expected)
				if string(expectedJSON) != string(resultJSON) {
					t.Logf("Selector %q: expected %s, got %s", test.Selector, expectedJSON, resultJSON)
					failCount++
				} else {
					passCount++
				}
			}
		})
	}

	t.Logf("RFC 9535 Test Suite Results:")
	t.Logf("PASS: %d/%d", passCount, len(suite.Tests))
	t.Logf("FAIL: %d/%d", failCount, len(suite.Tests))
	t.Logf("SKIP: %d/%d", skipCount, len(suite.Tests))

	baselineData := fmt.Sprintf("PASS: %d/%d\nFAIL: %d/%d\nSKIP: %d/%d\n",
		passCount, len(suite.Tests),
		failCount, len(suite.Tests),
		skipCount, len(suite.Tests))

	if err := os.WriteFile("testdata/rfc9535/baseline.txt", []byte(baselineData), 0644); err != nil {
		t.Fatalf("Failed to write baseline data: %v", err)
	}
}
