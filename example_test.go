package jsonpath

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestBasicQueries(t *testing.T) {
	testCases := []struct {
		name     string
		json     string
		path     string
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "simple field access",
			json:     `{"name": "John"}`,
			path:     "$.name",
			expected: "John",
		},
		{
			name:     "nested field access",
			json:     `{"person": {"name": "John"}}`,
			path:     "$.person.name",
			expected: "John",
		},
		{
			name:     "array index access",
			json:     `{"items": [1, 2, 3]}`,
			path:     "$.items[1]",
			expected: float64(2),
		},
		{
			name:     "wildcard",
			json:     `{"items": [1, 2, 3]}`,
			path:     "$.items[*]",
			expected: []interface{}{float64(1), float64(2), float64(3)},
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
			if !reflect.DeepEqual(result, tc.expected) {
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
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "filter by value without parentheses",
			json:     `{"items": [{"id": 1}, {"id": 2}, {"id": 3}]}`,
			path:     `$.items[?@.id==2]`,
			expected: []interface{}{map[string]interface{}{"id": float64(2)}},
		},
		{
			name:     "filter by value with parentheses",
			json:     `{"items": [{"id": 1}, {"id": 2}, {"id": 3}]}`,
			path:     `$.items[?(@.id==2)]`,
			expected: []interface{}{map[string]interface{}{"id": float64(2)}},
		},
		{
			name:     "filter by comparison without parentheses",
			json:     `{"items": [{"id": 1}, {"id": 2}, {"id": 3}]}`,
			path:     `$.items[?@.id>2]`,
			expected: []interface{}{map[string]interface{}{"id": float64(3)}},
		},
		{
			name:     "filter by comparison with parentheses",
			json:     `{"items": [{"id": 1}, {"id": 2}, {"id": 3}]}`,
			path:     `$.items[?(@.id>2)]`,
			expected: []interface{}{map[string]interface{}{"id": float64(3)}},
		},
		{
			name:     "filter with nested field without parentheses",
			json:     `{"items": [{"user": {"age": 25}}, {"user": {"age": 30}}]}`,
			path:     `$.items[?@.user.age>27]`,
			expected: []interface{}{map[string]interface{}{"user": map[string]interface{}{"age": float64(30)}}},
		},
		{
			name:     "filter with nested field with parentheses",
			json:     `{"items": [{"user": {"age": 25}}, {"user": {"age": 30}}]}`,
			path:     `$.items[?(@.user.age>27)]`,
			expected: []interface{}{map[string]interface{}{"user": map[string]interface{}{"age": float64(30)}}},
		},
		{
			name:     "filter by string equality",
			json:     `{"items": [{"name": "foo"}, {"name": "bar"}, {"name": "baz"}]}`,
			path:     `$.items[?@.name=="foo"]`,
			expected: []interface{}{map[string]interface{}{"name": "foo"}},
		},
		{
			name:     "filter by string comparison",
			json:     `{"items": [{"name": "foo"}, {"name": "bar"}, {"name": "baz"}]}`,
			path:     `$.items[?@.name>"bar"]`,
			expected: []interface{}{map[string]interface{}{"name": "foo"}, map[string]interface{}{"name": "baz"}},
		},
		{
			name:     "filter by boolean value",
			json:     `{"items": [{"active": true}, {"active": false}, {"active": true}]}`,
			path:     `$.items[?@.active==true]`,
			expected: []interface{}{map[string]interface{}{"active": true}, map[string]interface{}{"active": true}},
		},
		{
			name:     "filter by null value",
			json:     `{"items": [{"value": null}, {"value": 1}, {"value": null}]}`,
			path:     `$.items[?@.value==null]`,
			expected: []interface{}{map[string]interface{}{"value": nil}, map[string]interface{}{"value": nil}},
		},
		{
			name:     "filter by quoted string",
			json:     `{"items": [{"type": "book"}, {"type": "movie"}, {"type": "book"}]}`,
			path:     `$.items[?@.type=="book"]`,
			expected: []interface{}{map[string]interface{}{"type": "book"}, map[string]interface{}{"type": "book"}},
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
			if !reflect.DeepEqual(result, tc.expected) {
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
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "length of string",
			json:     `{"name": "hello"}`,
			path:     "$.name.length()",
			expected: float64(5),
		},
		{
			name:     "length of array",
			json:     `{"items": [1, 2, 3]}`,
			path:     "$.items.length()",
			expected: float64(3),
		},
		{
			name:     "length of object",
			json:     `{"obj": {"a": 1, "b": 2}}`,
			path:     "$.obj.length()",
			expected: float64(2),
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
			if !reflect.DeepEqual(result, tc.expected) {
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
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "keys of object",
			json:     `{"c": 3, "a": 1, "b": 2}`,
			path:     "$.keys()",
			expected: []interface{}{"a", "b", "c"},
		},
		{
			name:     "keys of nested object",
			json:     `{"store": {"book": [], "bicycle": {}}}`,
			path:     "$.store.keys()",
			expected: []interface{}{"bicycle", "book"},
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
			if !reflect.DeepEqual(result, tc.expected) {
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
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "values of object",
			json:     `{"c": 3, "a": 1, "b": 2}`,
			path:     "$.values()",
			expected: []interface{}{float64(1), float64(2), float64(3)},
		},
		{
			name:     "values of nested object",
			json:     `{"store": {"book": [], "bicycle": {"color": "red"}}}`,
			path:     "$.store.values()",
			expected: []interface{}{map[string]interface{}{"color": "red"}, []interface{}{}},
		},
		{
			name:    "values of non-object",
			json:    `{"arr": [1, 2, 3]}`,
			path:    "$.arr.values()",
			wantErr: true,
		},
		{
			name:     "values of object with mixed types",
			json:     `{"active": true, "age": 42, "name": "jp", "tags": ["json", "path"]}`,
			path:     "$.values()",
			expected: []interface{}{true, float64(42), "jp", []interface{}{"json", "path"}},
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
			if !reflect.DeepEqual(result, tc.expected) {
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
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "min of numbers",
			json:     `{"nums": [3, 1, 4, 1, 5, 9, 2, 6, 5, 3, 5]}`,
			path:     "$.nums.min()",
			expected: float64(1),
		},
		{
			name:     "min of mixed types",
			json:     `{"nums": [3, "invalid", 1, null, 4]}`,
			path:     "$.nums.min()",
			expected: float64(1),
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
			if !reflect.DeepEqual(result, tc.expected) {
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
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "max of numbers",
			json:     `{"nums": [3, 1, 4, 1, 5, 9, 2, 6, 5, 3, 5]}`,
			path:     "$.nums.max()",
			expected: float64(9),
		},
		{
			name:     "max of mixed types",
			json:     `{"nums": [3, "invalid", 1, null, 4]}`,
			path:     "$.nums.max()",
			expected: float64(4),
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
			if !reflect.DeepEqual(result, tc.expected) {
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
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "avg of numbers",
			json:     `{"nums": [2, 4, 6, 8, 10]}`,
			path:     "$.nums.avg()",
			expected: float64(6),
		},
		{
			name:     "avg of mixed types",
			json:     `{"nums": [3, "invalid", 1, null, 4]}`,
			path:     "$.nums.avg()",
			expected: float64(8.0 / 3.0),
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
			expected: float64(42),
		},
		{
			name:     "avg of decimal numbers",
			json:     `{"nums": [1.5, 2.5, 3.5]}`,
			path:     "$.nums.avg()",
			expected: float64(2.5),
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
			if !reflect.DeepEqual(result, tc.expected) {
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
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "sum of numbers",
			json:     `{"nums": [2, 4, 6, 8, 10]}`,
			path:     "$.nums.sum()",
			expected: float64(30),
		},
		{
			name:     "sum of mixed types",
			json:     `{"nums": [3, "invalid", 1, null, 4]}`,
			path:     "$.nums.sum()",
			expected: float64(8),
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
			expected: float64(42),
		},
		{
			name:     "sum of decimal numbers",
			json:     `{"nums": [1.5, 2.5, 3.5]}`,
			path:     "$.nums.sum()",
			expected: float64(7.5),
		},
		{
			name:     "sum of negative numbers",
			json:     `{"nums": [-1, -2, -3, -4, -5]}`,
			path:     "$.nums.sum()",
			expected: float64(-15),
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
			if !reflect.DeepEqual(result, tc.expected) {
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
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "count numbers",
			json:     `{"nums": [1, 2, 2, 3, 2, 4]}`,
			path:     "$.nums.count(2)",
			expected: float64(3),
		},
		{
			name:     "count strings",
			json:     `{"tags": ["a", "b", "a", "c", "a"]}`,
			path:     `$.tags.count("a")`,
			expected: float64(3),
		},
		{
			name:     "count objects",
			json:     `{"items": [{"id": 1}, {"id": 2}, {"id": 1}]}`,
			path:     `$.items.count({"id": 1})`,
			expected: float64(2),
		},
		{
			name:     "count with no matches",
			json:     `{"nums": [1, 2, 3]}`,
			path:     "$.nums.count(4)",
			expected: float64(0),
		},
		{
			name:    "count with non-array",
			json:    `{"num": 42}`,
			path:    "$.num.count(42)",
			wantErr: true,
		},
		{
			name:    "count with missing value",
			json:    `{"nums": [1, 2, 3]}`,
			path:    "$.nums.count()",
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
			if !reflect.DeepEqual(result, tc.expected) {
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
		expected interface{}
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
			path:     `$.books[?@.category=="fiction" && @.price<10]`,
			expected: []interface{}{map[string]interface{}{"category": "fiction", "price": 8.99}},
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
			expected: []interface{}{
				map[string]interface{}{"category": "reference", "price": 8.99},
				map[string]interface{}{"category": "fiction", "price": 15.99},
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
			path:     `$.books[?!@.category=="fiction"]`,
			expected: []interface{}{map[string]interface{}{"category": "reference", "price": 8.99}},
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
			expected: []interface{}{
				map[string]interface{}{"category": "fiction", "price": 12.99, "inStock": true},
				map[string]interface{}{"category": "fiction", "price": 15.99, "inStock": true},
				map[string]interface{}{"category": "fiction", "price": 9.99, "inStock": false},
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
			path:     `$.books[?@.category=="fiction" && @.price<15 && @.inStock==true]`,
			expected: []interface{}{map[string]interface{}{"category": "fiction", "price": 12.99, "inStock": true}},
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
			expected: []interface{}{
				map[string]interface{}{"category": "reference", "price": 8.99, "inStock": false},
				map[string]interface{}{"category": "fiction", "price": 15.99, "inStock": true},
				map[string]interface{}{"category": "fiction", "price": 9.99, "inStock": false},
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
			expected: []interface{}{
				map[string]interface{}{"category": "fiction", "price": 9.99, "inStock": false},
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
			expected: []interface{}{
				map[string]interface{}{"category": "fiction", "price": 12.99, "inStock": true},
				map[string]interface{}{"category": "fiction", "price": 9.99, "inStock": false},
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
			if !reflect.DeepEqual(result, tc.expected) {
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
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "match string with pattern",
			json:     `{"text": "Hello, World!"}`,
			path:     `$.text.match("^Hello")`,
			expected: true,
		},
		{
			name:     "match with case insensitive pattern",
			json:     `{"text": "Hello, World!"}`,
			path:     `$.text.match("(?i)world")`,
			expected: true,
		},
		{
			name:     "no match",
			json:     `{"text": "Hello, World!"}`,
			path:     `$.text.match("^World")`,
			expected: false,
		},
		{
			name:     "match with special characters",
			json:     `{"isbn": "123-4567890123"}`,
			path:     `$.isbn.match("^\\d{3}-\\d{10}$")`,
			expected: true,
		},
		{
			name:     "invalid pattern",
			json:     `{"text": "Hello"}`,
			path:     `$.text.match("(invalid")`,
			expected: false,
			wantErr:  false,
		},
		{
			name:     "match non-string value",
			json:     `{"num": 42}`,
			path:     `$.num.match("\\d+")`,
			expected: false,
		},
		{
			name:    "match with missing second argument",
			json:    `{"text": "Hello"}`,
			path:    `$.text.match()`,
			wantErr: true,
		},
		{
			name:    "match with non-string pattern",
			json:    `{"text": "Hello"}`,
			path:    `$.text.match(123)`,
			wantErr: true,
		},
		{
			name:     "match in filter expression",
			json:     `{"items": [{"id": "abc123"}, {"id": "def456"}]}`,
			path:     `$.items[?@.id.match("^abc")]`,
			expected: []interface{}{map[string]interface{}{"id": "abc123"}},
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
			if !reflect.DeepEqual(result, tc.expected) {
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
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "search strings in array",
			json:     `{"items": ["apple", "banana", "apricot", "grape"]}`,
			path:     `$.items.search("^ap")`,
			expected: []interface{}{"apple", "apricot"},
		},
		{
			name:     "search with case insensitive pattern",
			json:     `{"items": ["Apple", "APRICOT", "banana", "grape"]}`,
			path:     `$.items.search("(?i)^ap")`,
			expected: []interface{}{"Apple", "APRICOT"},
		},
		{
			name:     "search with no matches",
			json:     `{"items": ["apple", "banana", "grape"]}`,
			path:     `$.items.search("^x")`,
			expected: []interface{}{},
		},
		{
			name:     "search in array with non-string elements",
			json:     `{"items": ["apple", 42, "apricot", true]}`,
			path:     `$.items.search("^ap")`,
			expected: []interface{}{"apple", "apricot"},
		},
		{
			name:     "search with special characters",
			json:     `{"items": ["123-456", "abc-def", "789-012"]}`,
			path:     `$.items.search("^\\d{3}-\\d{3}$")`,
			expected: []interface{}{"123-456", "789-012"},
		},
		{
			name:    "search with invalid pattern",
			json:    `{"items": ["apple", "banana"]}`,
			path:    `$.items.search("(invalid")`,
			wantErr: true,
		},
		{
			name:    "search with missing pattern",
			json:    `{"items": ["apple", "banana"]}`,
			path:    `$.items.search()`,
			wantErr: true,
		},
		{
			name:    "search with non-string pattern",
			json:    `{"items": ["apple", "banana"]}`,
			path:    `$.items.search(123)`,
			wantErr: true,
		},
		{
			name:    "search on non-array",
			json:    `{"text": "apple"}`,
			path:    `$.text.search("^ap")`,
			wantErr: true,
		},
		{
			name:     "search in filter expression",
			json:     `{"fruits": [{"name": "apple"}, {"name": "banana"}, {"name": "apricot"}]}`,
			path:     `$.fruits[?@.name.match("^ap")].name`,
			expected: []interface{}{"apple", "apricot"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var data interface{}
			if err := json.Unmarshal([]byte(tc.json), &data); err != nil {
				t.Fatalf("failed to parse JSON: %v", err)
			}

			result, err := Query(data, tc.path)
			if tc.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}
