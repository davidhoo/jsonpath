# RFC 9535 Test Suite Data

This directory contains test data for validating JSONPath implementations against RFC 9535.

## Files

### testdata.json

The official JSONPath Compliance Test Suite (CTS) from the
[jsonpath-standard/jsonpath-compliance-test-suite](https://github.com/jsonpath-standard/jsonpath-compliance-test-suite)
repository.

- **Format**: JSON
- **Tests**: 703 test cases
- **Source**: Official RFC 9535 compliance test suite
- **Description**: Auto-generated test suite covering all JSONPath expressions defined in RFC 9535

#### Test Structure

Each test case in `testdata.json` has the following structure:

```json
{
  "name": "basic, root",
  "selector": "$",
  "document": ["first", "second"],
  "result": [["first", "second"]],
  "result_paths": ["$"]
}
```

- `name`: Test case description
- `selector`: JSONPath expression to test
- `document`: Input JSON document
- `result`: Expected output (array of matching values)
- `result_paths`: Expected normalized paths
- `invalid_selector`: (optional) If true, selector should be rejected

### rfc9535.txt

The full text of RFC 9535 ("JSONPath: Query Expressions for JSON") for reference.

## Usage

To run the compliance tests against your JSONPath implementation:

```go
// Load test data
data, _ := os.ReadFile("testdata/rfc9535/testdata.json")
var suite struct {
    Tests []struct {
        Name           string      `json:"name"`
        Selector       string      `json:"selector"`
        Document       interface{} `json:"document"`
        Result         interface{} `json:"result"`
        ResultPaths    []string    `json:"result_paths"`
        InvalidSelector bool       `json:"invalid_selector"`
    } `json:"tests"`
}
json.Unmarshal(data, &suite)

// Run each test
for _, test := range suite.Tests {
    // Your test logic here
}
```

## Updating

To update the test suite to the latest version:

```bash
curl -L "https://raw.githubusercontent.com/jsonpath-standard/jsonpath-compliance-test-suite/main/cts.json" \
  -o testdata/rfc9535/testdata.json
```

## References

- [RFC 9535](https://www.rfc-editor.org/rfc/rfc9535.html) - JSONPath: Query Expressions for JSON
- [JSONPath Compliance Test Suite](https://github.com/jsonpath-standard/jsonpath-compliance-test-suite) - Official test suite repository
