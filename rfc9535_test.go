package jsonpath

import (
	"encoding/json"
	"os"
	"testing"
)

type RFC9535TestCase struct {
	Name            string          `json:"name"`
	Selector        string          `json:"selector"`
	Document        json.RawMessage `json:"document"`
	Result          json.RawMessage `json:"result"`
	Results         []json.RawMessage `json:"results"`
	ResultPaths     []string        `json:"result_paths"`
	ResultsPaths    [][]string      `json:"results_paths"`
	InvalidSelector bool            `json:"invalid_selector"`
	Tags            []string        `json:"tags"`
}

type RFC9535TestSuite struct {
	Description string            `json:"description"`
	Tests       []RFC9535TestCase `json:"tests"`
}

func loadRFC9535TestSuite(t *testing.T) *RFC9535TestSuite {
	t.Helper()

	path := "testdata/rfc9535/testdata.json"
	if _, err := os.Stat(path); err != nil {
		t.Skip("RFC 9535 test data not found, skipping tests")
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read test data: %v", err)
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
	resultsCount := 0
	invalidCount := 0
	for _, tc := range suite.Tests {
		if tc.InvalidSelector {
			invalidCount++
		} else if len(tc.Results) > 0 {
			resultsCount++
		} else {
			resultCount++
		}
	}
	t.Logf("  - %d with single result", resultCount)
	t.Logf("  - %d with multiple results", resultsCount)
	t.Logf("  - %d invalid selectors", invalidCount)
}
