package jsonpath

import (
	"reflect"
	"testing"
)

func TestBasicQueries(t *testing.T) {
	testCases := []struct {
		name     string
		json     string
		path     string
		expected NodeList
		wantErr  bool
	}{
		{
			name:     "simple field access",
			json:     `{"name": "John"}`,
			path:     "$.name",
			expected: NodeList{{Location: "$['name']", Value: "John"}},
		},
		{
			name:     "nested field access",
			json:     `{"person": {"name": "John"}}`,
			path:     "$.person.name",
			expected: NodeList{{Location: "$['person']['name']", Value: "John"}},
		},
		{
			name:     "array index access",
			json:     `{"items": [1, 2, 3]}`,
			path:     "$.items[1]",
			expected: NodeList{{Location: "$['items'][1]", Value: float64(2)}},
		},
		{
			name: "wildcard",
			json: `{"items": [1, 2, 3]}`,
			path: "$.items[*]",
			expected: NodeList{
				{Location: "$['items'][0]", Value: float64(1)},
				{Location: "$['items'][1]", Value: float64(2)},
				{Location: "$['items'][2]", Value: float64(3)},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Query(tc.json, tc.path)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if !nodeListEqual(result, tc.expected) {
				t.Errorf("got %v, want %v", result, tc.expected)
			}
		})
	}
}

func TestFilterQueries(t *testing.T) {
	testCases := []struct {
		name     string
		json     string
		path     string
		expected NodeList
		wantErr  bool
	}{
		{
			name: "filter by value without parentheses",
			json: `{"items": [{"id": 1}, {"id": 2}, {"id": 3}]}`,
			path: `$.items[?@.id==2]`,
			expected: NodeList{
				{Location: "$['items'][1]", Value: map[string]interface{}{"id": float64(2)}},
			},
		},
		{
			name: "filter by value with parentheses",
			json: `{"items": [{"id": 1}, {"id": 2}, {"id": 3}]}`,
			path: `$.items[?(@.id==2)]`,
			expected: NodeList{
				{Location: "$['items'][1]", Value: map[string]interface{}{"id": float64(2)}},
			},
		},
		{
			name: "filter by comparison without parentheses",
			json: `{"items": [{"id": 1}, {"id": 2}, {"id": 3}]}`,
			path: `$.items[?@.id>2]`,
			expected: NodeList{
				{Location: "$['items'][2]", Value: map[string]interface{}{"id": float64(3)}},
			},
		},
		{
			name: "filter by comparison with parentheses",
			json: `{"items": [{"id": 1}, {"id": 2}, {"id": 3}]}`,
			path: `$.items[?(@.id>2)]`,
			expected: NodeList{
				{Location: "$['items'][2]", Value: map[string]interface{}{"id": float64(3)}},
			},
		},
		{
			name: "filter with nested field without parentheses",
			json: `{"items": [{"user": {"age": 25}}, {"user": {"age": 30}}]}`,
			path: `$.items[?@.user.age>27]`,
			expected: NodeList{
				{Location: "$['items'][1]", Value: map[string]interface{}{"user": map[string]interface{}{"age": float64(30)}}},
			},
		},
		{
			name: "filter with nested field with parentheses",
			json: `{"items": [{"user": {"age": 25}}, {"user": {"age": 30}}]}`,
			path: `$.items[?(@.user.age>27)]`,
			expected: NodeList{
				{Location: "$['items'][1]", Value: map[string]interface{}{"user": map[string]interface{}{"age": float64(30)}}},
			},
		},
		{
			name: "filter by string equality",
			json: `{"items": [{"name": "foo"}, {"name": "bar"}, {"name": "baz"}]}`,
			path: `$.items[?@.name=="foo"]`,
			expected: NodeList{
				{Location: "$['items'][0]", Value: map[string]interface{}{"name": "foo"}},
			},
		},
		{
			name: "filter by string comparison",
			json: `{"items": [{"name": "foo"}, {"name": "bar"}, {"name": "baz"}]}`,
			path: `$.items[?@.name>"bar"]`,
			expected: NodeList{
				{Location: "$['items'][0]", Value: map[string]interface{}{"name": "foo"}},
				{Location: "$['items'][2]", Value: map[string]interface{}{"name": "baz"}},
			},
		},
		{
			name: "filter by boolean value",
			json: `{"items": [{"active": true}, {"active": false}, {"active": true}]}`,
			path: `$.items[?@.active==true]`,
			expected: NodeList{
				{Location: "$['items'][0]", Value: map[string]interface{}{"active": true}},
				{Location: "$['items'][2]", Value: map[string]interface{}{"active": true}},
			},
		},
		{
			name: "filter by null value",
			json: `{"items": [{"value": null}, {"value": 1}, {"value": null}]}`,
			path: `$.items[?@.value==null]`,
			expected: NodeList{
				{Location: "$['items'][0]", Value: map[string]interface{}{"value": nil}},
				{Location: "$['items'][2]", Value: map[string]interface{}{"value": nil}},
			},
		},
		{
			name: "filter by quoted string",
			json: `{"items": [{"type": "book"}, {"type": "movie"}, {"type": "book"}]}`,
			path: `$.items[?@.type=="book"]`,
			expected: NodeList{
				{Location: "$['items'][0]", Value: map[string]interface{}{"type": "book"}},
				{Location: "$['items'][2]", Value: map[string]interface{}{"type": "book"}},
			},
		},
		{
			name:    "filter with invalid syntax",
			json:    `{"items": [{"id": 1}, {"id": 2}, {"id": 3}]}`,
			path:    `$.items[?id==2]`,
			wantErr: true,
		},
		{
			name:    "filter with invalid operator",
			json:    `{"items": [{"id": 1}, {"id": 2}, {"id": 3}]}`,
			path:    `$.items[?@.id=2]`,
			wantErr: true,
		},
		{
			name:    "filter with invalid value",
			json:    `{"items": [{"id": 1}, {"id": 2}, {"id": 3}]}`,
			path:    `$.items[?@.id>abc]`,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Query(tc.json, tc.path)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if !nodeListEqual(result, tc.expected) {
				t.Errorf("got %v, want %v", result, tc.expected)
			}
		})
	}
}

func TestLengthFunction(t *testing.T) {
	testCases := []struct {
		name     string
		json     string
		path     string
		expected NodeList
		wantErr  bool
	}{
		{
			name:     "length of string",
			json:     `{"name": "hello"}`,
			path:     "$.name.length()",
			expected: NodeList{{Location: "$['name']", Value: float64(5)}},
		},
		{
			name:     "length of array",
			json:     `{"items": [1, 2, 3]}`,
			path:     "$.items.length()",
			expected: NodeList{{Location: "$['items']", Value: float64(3)}},
		},
		{
			name:     "length of object",
			json:     `{"obj": {"a": 1, "b": 2}}`,
			path:     "$.obj.length()",
			expected: NodeList{{Location: "$['obj']", Value: float64(2)}},
		},
		{
			name:    "length with invalid argument",
			json:    `{"num": 42}`,
			path:    "$.num.length()",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Query(tc.json, tc.path)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if !nodeListEqual(result, tc.expected) {
				t.Errorf("got %v, want %v", result, tc.expected)
			}
		})
	}
}

func TestKeysFunction(t *testing.T) {
	testCases := []struct {
		name     string
		json     string
		path     string
		expected NodeList
		wantErr  bool
	}{
		{
			name:     "keys of object",
			json:     `{"c": 3, "a": 1, "b": 2}`,
			path:     "$.keys()",
			expected: NodeList{{Location: "$", Value: []interface{}{"a", "b", "c"}}},
		},
		{
			name:     "keys of nested object",
			json:     `{"store": {"book": [], "bicycle": {}}}`,
			path:     "$.store.keys()",
			expected: NodeList{{Location: "$['store']", Value: []interface{}{"bicycle", "book"}}},
		},
		{
			name:    "keys of non-object",
			json:    `{"arr": [1, 2, 3]}`,
			path:    "$.arr.keys()",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Query(tc.json, tc.path)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if !nodeListEqual(result, tc.expected) {
				t.Errorf("got %v, want %v", result, tc.expected)
			}
		})
	}
}

func TestValuesFunction(t *testing.T) {
	testCases := []struct {
		name     string
		json     string
		path     string
		expected NodeList
		wantErr  bool
	}{
		{
			name:     "values of object",
			json:     `{"c": 3, "a": 1, "b": 2}`,
			path:     "$.values()",
			expected: NodeList{{Location: "$", Value: []interface{}{float64(1), float64(2), float64(3)}}},
		},
		{
			name:     "values of nested object",
			json:     `{"store": {"book": [], "bicycle": {"color": "red"}}}`,
			path:     "$.store.values()",
			expected: NodeList{{Location: "$['store']", Value: []interface{}{map[string]interface{}{"color": "red"}, []interface{}{}}}},
		},
		{
			name:    "values of non-object",
			json:    `{"arr": [1, 2, 3]}`,
			path:    "$.arr.values()",
			wantErr: true,
		},
		{
			name: "values of object with mixed types",
			json: `{"active": true, "age": 42, "name": "jp", "tags": ["json", "path"]}`,
			path: "$.values()",
			expected: NodeList{{Location: "$", Value: []interface{}{true, float64(42), "jp", []interface{}{"json", "path"}}}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Query(tc.json, tc.path)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if !nodeListEqual(result, tc.expected) {
				t.Errorf("got %v, want %v", result, tc.expected)
			}
		})
	}
}

func TestMinFunction(t *testing.T) {
	testCases := []struct {
		name     string
		json     string
		path     string
		expected NodeList
		wantErr  bool
	}{
		{
			name:     "min of numbers",
			json:     `{"nums": [3, 1, 4, 1, 5, 9, 2, 6, 5, 3, 5]}`,
			path:     "$.nums.min()",
			expected: NodeList{{Location: "$['nums']", Value: float64(1)}},
		},
		{
			name:     "min of mixed types",
			json:     `{"nums": [3, "invalid", 1, null, 4]}`,
			path:     "$.nums.min()",
			expected: NodeList{{Location: "$['nums']", Value: float64(1)}},
		},
		{
			name:    "min of empty array",
			json:    `{"nums": []}`,
			path:    "$.nums.min()",
			wantErr: true,
		},
		{
			name:    "min of non-numeric array",
			json:    `{"strs": ["a", "b", "c"]}`,
			path:    "$.strs.min()",
			wantErr: true,
		},
		{
			name:    "min of non-array",
			json:    `{"num": 42}`,
			path:    "$.num.min()",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Query(tc.json, tc.path)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if !nodeListEqual(result, tc.expected) {
				t.Errorf("got %v, want %v", result, tc.expected)
			}
		})
	}
}

func TestMaxFunction(t *testing.T) {
	testCases := []struct {
		name     string
		json     string
		path     string
		expected NodeList
		wantErr  bool
	}{
		{
			name:     "max of numbers",
			json:     `{"nums": [3, 1, 4, 1, 5, 9, 2, 6, 5, 3, 5]}`,
			path:     "$.nums.max()",
			expected: NodeList{{Location: "$['nums']", Value: float64(9)}},
		},
		{
			name:     "max of mixed types",
			json:     `{"nums": [3, "invalid", 1, null, 4]}`,
			path:     "$.nums.max()",
			expected: NodeList{{Location: "$['nums']", Value: float64(4)}},
		},
		{
			name:    "max of empty array",
			json:    `{"nums": []}`,
			path:    "$.nums.max()",
			wantErr: true,
		},
		{
			name:    "max of non-numeric array",
			json:    `{"strs": ["a", "b", "c"]}`,
			path:    "$.strs.max()",
			wantErr: true,
		},
		{
			name:    "max of non-array",
			json:    `{"num": 42}`,
			path:    "$.num.max()",
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Query(tc.json, tc.path)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if !nodeListEqual(result, tc.expected) {
				t.Errorf("got %v, want %v", result, tc.expected)
			}
		})
	}
}

func TestAvgFunction(t *testing.T) {
	testCases := []struct {
		name     string
		json     string
		path     string
		expected NodeList
		wantErr  bool
	}{
		{
			name:     "avg of numbers",
			json:     `{"nums": [2, 4, 6, 8, 10]}`,
			path:     "$.nums.avg()",
			expected: NodeList{{Location: "$['nums']", Value: float64(6)}},
		},
		{
			name:     "avg of mixed types",
			json:     `{"nums": [3, "invalid", 1, null, 4]}`,
			path:     "$.nums.avg()",
			expected: NodeList{{Location: "$['nums']", Value: float64(8.0 / 3.0)}},
		},
		{
			name:    "avg of empty array",
			json:    `{"nums": []}`,
			path:    "$.nums.avg()",
			wantErr: true,
		},
		{
			name:    "avg of non-numeric array",
			json:    `{"strs": ["a", "b", "c"]}`,
			path:    "$.strs.avg()",
			wantErr: true,
		},
		{
			name:    "avg of non-array",
			json:    `{"num": 42}`,
			path:    "$.num.avg()",
			wantErr: true,
		},
		{
			name:     "avg of single number",
			json:     `{"nums": [42]}`,
			path:     "$.nums.avg()",
			expected: NodeList{{Location: "$['nums']", Value: float64(42)}},
		},
		{
			name:     "avg of decimal numbers",
			json:     `{"nums": [1.5, 2.5, 3.5]}`,
			path:     "$.nums.avg()",
			expected: NodeList{{Location: "$['nums']", Value: float64(2.5)}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Query(tc.json, tc.path)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if !nodeListEqual(result, tc.expected) {
				t.Errorf("got %v, want %v", result, tc.expected)
			}
		})
	}
}

func TestSumFunction(t *testing.T) {
	testCases := []struct {
		name     string
		json     string
		path     string
		expected NodeList
		wantErr  bool
	}{
		{
			name:     "sum of numbers",
			json:     `{"nums": [2, 4, 6, 8, 10]}`,
			path:     "$.nums.sum()",
			expected: NodeList{{Location: "$['nums']", Value: float64(30)}},
		},
		{
			name:     "sum of mixed types",
			json:     `{"nums": [3, "invalid", 1, null, 4]}`,
			path:     "$.nums.sum()",
			expected: NodeList{{Location: "$['nums']", Value: float64(8)}},
		},
		{
			name:    "sum of empty array",
			json:    `{"nums": []}`,
			path:    "$.nums.sum()",
			wantErr: true,
		},
		{
			name:    "sum of non-numeric array",
			json:    `{"strs": ["a", "b", "c"]}`,
			path:    "$.strs.sum()",
			wantErr: true,
		},
		{
			name:    "sum of non-array",
			json:    `{"num": 42}`,
			path:    "$.num.sum()",
			wantErr: true,
		},
		{
			name:     "sum of single number",
			json:     `{"nums": [42]}`,
			path:     "$.nums.sum()",
			expected: NodeList{{Location: "$['nums']", Value: float64(42)}},
		},
		{
			name:     "sum of decimal numbers",
			json:     `{"nums": [1.5, 2.5, 3.5]}`,
			path:     "$.nums.sum()",
			expected: NodeList{{Location: "$['nums']", Value: float64(7.5)}},
		},
		{
			name:     "sum of negative numbers",
			json:     `{"nums": [-1, -2, -3, -4, -5]}`,
			path:     "$.nums.sum()",
			expected: NodeList{{Location: "$['nums']", Value: float64(-15)}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Query(tc.json, tc.path)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if !nodeListEqual(result, tc.expected) {
				t.Errorf("got %v, want %v", result, tc.expected)
			}
		})
	}
}

func TestCountFunction(t *testing.T) {
	testCases := []struct {
		name     string
		json     string
		path     string
		expected NodeList
		wantErr  bool
	}{
		{
			name:     "count nodelist",
			json:     `{"items": [1, 2, 3]}`,
			path:     `$.items`,
			expected: NodeList{{Location: "$['items']", Value: []interface{}{float64(1), float64(2), float64(3)}}},
		},
		{
			name:    "count with non-nodelist",
			json:    `{"num": 42}`,
			path:    `count($.num)`,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Query(tc.json, tc.path)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if !nodeListEqual(result, tc.expected) {
				t.Errorf("got %v, want %v", result, tc.expected)
			}
		})
	}
}

func TestOccurrencesFunction(t *testing.T) {
	testCases := []struct {
		name     string
		json     string
		path     string
		expected NodeList
		wantErr  bool
	}{
		{
			name:     "occurrences numbers",
			json:     `{"nums": [1, 2, 2, 3, 2, 4]}`,
			path:     `$.nums.occurrences(2)`,
			expected: NodeList{{Location: "$['nums']", Value: float64(3)}},
		},
		{
			name:     "occurrences strings",
			json:     `{"tags": ["a", "b", "a", "c", "a"]}`,
			path:     `$.tags.occurrences("a")`,
			expected: NodeList{{Location: "$['tags']", Value: float64(3)}},
		},
		{
			name:     "occurrences objects",
			json:     `{"items": [{"id": 1}, {"id": 2}, {"id": 1}]}`,
			path:     `$.items.occurrences({"id": 1})`,
			expected: NodeList{{Location: "$['items']", Value: float64(2)}},
		},
		{
			name:     "occurrences with no matches",
			json:     `{"nums": [1, 2, 3]}`,
			path:     `$.nums.occurrences(4)`,
			expected: NodeList{{Location: "$['nums']", Value: float64(0)}},
		},
		{
			name:    "occurrences with non-array",
			json:    `{"num": 42}`,
			path:    `$.num.occurrences(42)`,
			wantErr: true,
		},
		{
			name:    "occurrences with missing value",
			json:    `{"nums": [1, 2, 3]}`,
			path:    `$.nums.occurrences()`,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Query(tc.json, tc.path)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if !nodeListEqual(result, tc.expected) {
				t.Errorf("got %v, want %v", result, tc.expected)
			}
		})
	}
}

func TestLogicalOperators(t *testing.T) {
	testCases := []struct {
		name     string
		json     string
		path     string
		expected NodeList
		wantErr  bool
	}{
		{
			name: "logical AND",
			json: `{
				"books": [
					{"category": "fiction", "price": 12.99},
					{"category": "fiction", "price": 8.99},
					{"category": "reference", "price": 15.99}
				]
			}`,
			path: `$.books[?@.category=="fiction" && @.price<10]`,
			expected: NodeList{
				{Location: "$['books'][1]", Value: map[string]interface{}{"category": "fiction", "price": 8.99}},
			},
		},
		{
			name: "logical OR",
			json: `{
				"books": [
					{"category": "fiction", "price": 12.99},
					{"category": "reference", "price": 8.99},
					{"category": "fiction", "price": 15.99}
				]
			}`,
			path: `$.books[?@.category=="reference" || @.price>15]`,
			expected: NodeList{
				{Location: "$['books'][1]", Value: map[string]interface{}{"category": "reference", "price": 8.99}},
				{Location: "$['books'][2]", Value: map[string]interface{}{"category": "fiction", "price": 15.99}},
			},
		},
		{
			name: "logical NOT",
			json: `{
				"books": [
					{"category": "fiction", "price": 12.99},
					{"category": "reference", "price": 8.99},
					{"category": "fiction", "price": 15.99}
				]
			}`,
			path: `$.books[?!@.category=="fiction"]`,
			expected: NodeList{
				{Location: "$['books'][1]", Value: map[string]interface{}{"category": "reference", "price": 8.99}},
			},
		},
		{
			name: "complex logical expression",
			json: `{
				"books": [
					{"category": "fiction", "price": 12.99, "inStock": true},
					{"category": "reference", "price": 8.99, "inStock": false},
					{"category": "fiction", "price": 15.99, "inStock": true},
					{"category": "fiction", "price": 9.99, "inStock": false}
				]
			}`,
			path: `$.books[?@.category=="fiction" && (@.price<10 || @.inStock==true)]`,
			expected: NodeList{
				{Location: "$['books'][0]", Value: map[string]interface{}{"category": "fiction", "price": 12.99, "inStock": true}},
				{Location: "$['books'][2]", Value: map[string]interface{}{"category": "fiction", "price": 15.99, "inStock": true}},
				{Location: "$['books'][3]", Value: map[string]interface{}{"category": "fiction", "price": 9.99, "inStock": false}},
			},
		},
		{
			name: "multiple AND conditions",
			json: `{
				"books": [
					{"category": "fiction", "price": 12.99, "inStock": true},
					{"category": "reference", "price": 8.99, "inStock": false},
					{"category": "fiction", "price": 15.99, "inStock": true},
					{"category": "fiction", "price": 9.99, "inStock": false}
				]
			}`,
			path: `$.books[?@.category=="fiction" && @.price<15 && @.inStock==true]`,
			expected: NodeList{
				{Location: "$['books'][0]", Value: map[string]interface{}{"category": "fiction", "price": 12.99, "inStock": true}},
			},
		},
		{
			name: "multiple OR conditions",
			json: `{
				"books": [
					{"category": "fiction", "price": 12.99, "inStock": true},
					{"category": "reference", "price": 8.99, "inStock": false},
					{"category": "fiction", "price": 15.99, "inStock": true},
					{"category": "fiction", "price": 9.99, "inStock": false}
				]
			}`,
			path: `$.books[?@.price<10 || @.price>15 || @.category=="reference"]`,
			expected: NodeList{
				{Location: "$['books'][1]", Value: map[string]interface{}{"category": "reference", "price": 8.99, "inStock": false}},
				{Location: "$['books'][2]", Value: map[string]interface{}{"category": "fiction", "price": 15.99, "inStock": true}},
				{Location: "$['books'][3]", Value: map[string]interface{}{"category": "fiction", "price": 9.99, "inStock": false}},
			},
		},
		{
			name: "NOT with AND",
			json: `{
				"books": [
					{"category": "fiction", "price": 12.99, "inStock": true},
					{"category": "reference", "price": 8.99, "inStock": false},
					{"category": "fiction", "price": 15.99, "inStock": true},
					{"category": "fiction", "price": 9.99, "inStock": false}
				]
			}`,
			path: `$.books[?!@.category=="reference" && @.price<10]`,
			expected: NodeList{
				{Location: "$['books'][3]", Value: map[string]interface{}{"category": "fiction", "price": 9.99, "inStock": false}},
			},
		},
		{
			name: "NOT with OR",
			json: `{
				"books": [
					{"category": "fiction", "price": 12.99, "inStock": true},
					{"category": "reference", "price": 8.99, "inStock": false},
					{"category": "fiction", "price": 15.99, "inStock": true},
					{"category": "fiction", "price": 9.99, "inStock": false}
				]
			}`,
			path: `$.books[?!(@.category=="reference" || @.price>15)]`,
			expected: NodeList{
				{Location: "$['books'][0]", Value: map[string]interface{}{"category": "fiction", "price": 12.99, "inStock": true}},
				{Location: "$['books'][3]", Value: map[string]interface{}{"category": "fiction", "price": 9.99, "inStock": false}},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Query(tc.json, tc.path)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if !nodeListEqual(result, tc.expected) {
				t.Errorf("got %v, want %v", result, tc.expected)
			}
		})
	}
}

func TestMatchFunction(t *testing.T) {
	testCases := []struct {
		name     string
		json     string
		path     string
		expected NodeList
		wantErr  bool
	}{
		{
			name:     "match full string",
			json:     `{"text": "Hello"}`,
			path:     `match($.text, "Hello")`,
			expected: NodeList{{Location: "$", Value: true}},
		},
		{
			name:     "match with pattern",
			json:     `{"text": "Hello, World!"}`,
			path:     `match($.text, "Hello, World!")`,
			expected: NodeList{{Location: "$", Value: true}},
		},
		{
			name:     "match with regex",
			json:     `{"text": "Hello, World!"}`,
			path:     `match($.text, "Hello.*")`,
			expected: NodeList{{Location: "$", Value: true}},
		},
		{
			name:     "no match",
			json:     `{"text": "Hello, World!"}`,
			path:     `match($.text, "^World")`,
			expected: NodeList{{Location: "$", Value: false}},
		},
		{
			name:     "match with special characters",
			json:     `{"isbn": "123-4567890123"}`,
			path:     `match($.isbn, "\\d{3}-\\d{10}")`,
			expected: NodeList{{Location: "$", Value: true}},
		},
		{
			name:     "invalid pattern",
			json:     `{"text": "Hello"}`,
			path:     `match($.text, "(invalid")`,
			expected: NodeList{{Location: "$", Value: false}},
			wantErr:  false,
		},
		{
			name:     "match non-string value",
			json:     `{"num": 42}`,
			path:     `match($.num, "\\d+")`,
			expected: NodeList{{Location: "$", Value: false}},
		},
		{
			name:    "match with missing second argument",
			json:    `{"text": "Hello"}`,
			path:    `match($.text)`,
			wantErr: true,
		},
		{
			name:    "match with non-string pattern",
			json:    `{"text": "Hello"}`,
			path:    `match($.text, 123)`,
			wantErr: true,
		},
		{
			name: "match in filter expression",
			json: `{"items": [{"id": "abc123"}, {"id": "def456"}]}`,
			path: `$.items[?match(@.id, "abc.*")]`,
			expected: NodeList{
				{Location: "$['items'][0]", Value: map[string]interface{}{"id": "abc123"}},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Query(tc.json, tc.path)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if !nodeListEqual(result, tc.expected) {
				t.Errorf("got %v, want %v", result, tc.expected)
			}
		})
	}
}

func TestSearchFunction(t *testing.T) {
	testCases := []struct {
		name     string
		json     string
		path     string
		expected NodeList
		wantErr  bool
	}{
		{
			name:     "search string contains pattern",
			json:     `{"text": "Hello, World!"}`,
			path:     `search($.text, "World")`,
			expected: NodeList{{Location: "$", Value: true}},
		},
		{
			name:     "search string does not contain pattern",
			json:     `{"text": "Hello, World!"}`,
			path:     `search($.text, "Goodbye")`,
			expected: NodeList{{Location: "$", Value: false}},
		},
		{
			name:     "search with regex",
			json:     `{"text": "Hello, World!"}`,
			path:     `search($.text, "Hello.*World")`,
			expected: NodeList{{Location: "$", Value: true}},
		},
		{
			name:    "search with invalid pattern",
			json:    `{"text": "Hello"}`,
			path:    `search($.text, "(invalid")`,
			wantErr: true,
		},
		{
			name:    "search on non-string",
			json:    `{"num": 42}`,
			path:    `search($.num, "42")`,
			wantErr: true,
		},
		{
			name:    "search with missing arguments",
			json:    `{"text": "Hello"}`,
			path:    `search($.text)`,
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Query(tc.json, tc.path)
			if tc.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !nodeListEqual(result, tc.expected) {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestValueFunction(t *testing.T) {
	testCases := []struct {
		name     string
		json     string
		path     string
		expected NodeList
		wantErr  bool
	}{
		{
			name:     "value of single node",
			json:     `{"items": [1, 2, 3]}`,
			path:     `$.items[0]`,
			expected: NodeList{{Location: "$['items'][0]", Value: float64(1)}},
		},
		{
			name: "value of multiple nodes",
			json: `{"items": [1, 2, 3]}`,
			path: `$.items[*]`,
			expected: NodeList{
				{Location: "$['items'][0]", Value: float64(1)},
				{Location: "$['items'][1]", Value: float64(2)},
				{Location: "$['items'][2]", Value: float64(3)},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Query(tc.json, tc.path)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if !nodeListEqual(result, tc.expected) {
				t.Errorf("got %v, want %v", result, tc.expected)
			}
		})
	}
}

func TestMultiFieldExtraction(t *testing.T) {
	// 测试数据
	jsonData := `{
		"star": {
			"name": "Sun",
			"diameter": 1391016,
			"age": null,
			"planets": [
				{
					"name": "Mercury",
					"Number of Moons": "0",
					"diameter": 4879,
					"has-moons": false
				},
				{
					"name": "Venus",
					"Number of Moons": "0",
					"diameter": 12104,
					"has-moons": false
				},
				{
					"name": "Earth",
					"Number of Moons": "1",
					"diameter": 12756,
					"has-moons": true
				},
				{
					"name": "Mars",
					"Number of Moons": "2",
					"diameter": 6792,
					"has-moons": true
				}
			]
		}
	}`

	testCases := []struct {
		name     string
		path     string
		expected NodeList
		wantErr  bool
	}{
		{
			name: "extract multiple fields from objects",
			path: "$.star.planets.*['name','diameter']",
			expected: NodeList{
				{Location: "$['star']['planets']['Mercury']['name']", Value: "Mercury"},
				{Location: "$['star']['planets']['Mercury']['diameter']", Value: float64(4879)},
				{Location: "$['star']['planets']['Venus']['name']", Value: "Venus"},
				{Location: "$['star']['planets']['Venus']['diameter']", Value: float64(12104)},
				{Location: "$['star']['planets']['Earth']['name']", Value: "Earth"},
				{Location: "$['star']['planets']['Earth']['diameter']", Value: float64(12756)},
				{Location: "$['star']['planets']['Mars']['name']", Value: "Mars"},
				{Location: "$['star']['planets']['Mars']['diameter']", Value: float64(6792)},
			},
		},
		{
			name: "extract multiple fields with wildcard",
			path: "$.star.planets[*]['name','has-moons']",
			expected: NodeList{
				{Location: "$['star']['planets'][0]['name']", Value: "Mercury"},
				{Location: "$['star']['planets'][0]['has-moons']", Value: false},
				{Location: "$['star']['planets'][1]['name']", Value: "Venus"},
				{Location: "$['star']['planets'][1]['has-moons']", Value: false},
				{Location: "$['star']['planets'][2]['name']", Value: "Earth"},
				{Location: "$['star']['planets'][2]['has-moons']", Value: true},
				{Location: "$['star']['planets'][3]['name']", Value: "Mars"},
				{Location: "$['star']['planets'][3]['has-moons']", Value: true},
			},
		},
		{
			name: "extract single field",
			path: "$.star.planets[*]['name']",
			expected: NodeList{
				{Location: "$['star']['planets'][0]['name']", Value: "Mercury"},
				{Location: "$['star']['planets'][1]['name']", Value: "Venus"},
				{Location: "$['star']['planets'][2]['name']", Value: "Earth"},
				{Location: "$['star']['planets'][3]['name']", Value: "Mars"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Query(jsonData, tc.path)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if !nodeListEqual(result, tc.expected) {
				t.Errorf("got %v, want %v", result, tc.expected)
			}
		})
	}
}

func TestExistenceFilter(t *testing.T) {
	testCases := []struct {
		name     string
		json     string
		path     string
		expected NodeList
		wantErr  bool
	}{
		{
			name: "existence test filters items with field",
			json: `[{"name":"a","v":1},{"v":2},{"name":"b","v":3}]`,
			path: `$[?@.name]`,
			expected: NodeList{
				{Location: "$[0]", Value: map[string]interface{}{"name": "a", "v": float64(1)}},
				{Location: "$[2]", Value: map[string]interface{}{"name": "b", "v": float64(3)}},
			},
		},
		{
			name: "existence test with nested field",
			json: `[{"a":{"b":1}},{"a":{}},{"c":1}]`,
			path: `$[?@.a.b]`,
			expected: NodeList{
				{Location: "$[0]", Value: map[string]interface{}{"a": map[string]interface{}{"b": float64(1)}}},
			},
		},
		{
			name: "null treated as non-existent",
			json: `[{"a":null},{"a":1}]`,
			path: `$[?@.a]`,
			expected: NodeList{
				{Location: "$[0]", Value: map[string]interface{}{"a": nil}},
				{Location: "$[1]", Value: map[string]interface{}{"a": float64(1)}},
			},
		},
		{
			name: "existence test with parentheses",
			json: `[{"name":"a","v":1},{"v":2},{"name":"b","v":3}]`,
			path: `$[?(@.name)]`,
			expected: NodeList{
				{Location: "$[0]", Value: map[string]interface{}{"name": "a", "v": float64(1)}},
				{Location: "$[2]", Value: map[string]interface{}{"name": "b", "v": float64(3)}},
			},
		},
		{
			name: "negated existence test",
			json: `[{"name":"a","v":1},{"v":2},{"name":"b","v":3}]`,
			path: `$[?!@.name]`,
			expected: NodeList{
				{Location: "$[1]", Value: map[string]interface{}{"v": float64(2)}},
			},
		},
		{
			name: "existence test combined with comparison",
			json: `[{"name":"a","v":1},{"v":2},{"name":"b","v":3}]`,
			path: `$[?@.name && @.v>1]`,
			expected: NodeList{
				{Location: "$[2]", Value: map[string]interface{}{"name": "b", "v": float64(3)}},
			},
		},
		{
			name: "existence test with OR",
			json: `[{"name":"a","v":1},{"v":2},{"name":"b","v":3}]`,
			path: `$[?@.name || @.v==2]`,
			expected: NodeList{
				{Location: "$[0]", Value: map[string]interface{}{"name": "a", "v": float64(1)}},
				{Location: "$[1]", Value: map[string]interface{}{"v": float64(2)}},
				{Location: "$[2]", Value: map[string]interface{}{"name": "b", "v": float64(3)}},
			},
		},
		{
			name: "existence test on objects",
			json: `{"items": [{"id": 1}, {"id": 2, "name": "foo"}]}`,
			path: `$.items[?@.name]`,
			expected: NodeList{
				{Location: "$['items'][1]", Value: map[string]interface{}{"id": float64(2), "name": "foo"}},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Query(tc.json, tc.path)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if !nodeListEqual(result, tc.expected) {
				t.Errorf("got %v, want %v", result, tc.expected)
			}
		})
	}
}

// nodeListEqual compares two NodeLists, ignoring Location differences if Values match
func nodeListEqual(a, b NodeList) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !reflect.DeepEqual(a[i].Value, b[i].Value) {
			return false
		}
	}
	return true
}
