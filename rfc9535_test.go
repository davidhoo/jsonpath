package jsonpath

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type RFC9535TestCase struct {
	Name     string      `json:"name"`
	Selector string      `json:"selector"`
	Document interface{} `json:"document"`
	Expected interface{} `json:"result"`
	Invalid  bool        `json:"invalid,omitempty"`
}

type RFC9535TestSuite struct {
	Tests []RFC9535TestCase `json:"tests"`
}

func loadRFC9535TestSuite(t *testing.T) *RFC9535TestSuite {
	t.Helper()

	paths := []string{
		"testdata/rfc9535/testdata.json",
		filepath.Join("testdata", "rfc9535", "regression_suite.yaml"),
	}

	var data []byte
	var err error
	for _, path := range paths {
		if _, err = os.Stat(path); err == nil {
			data, err = os.ReadFile(path)
			if err == nil {
				break
			}
		}
	}
	if data == nil {
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
