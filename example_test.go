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
			name:     "filter by value",
			json:     `{"items": [{"id": 1}, {"id": 2}, {"id": 3}]}`,
			path:     `$.items[?(@.id==2)]`,
			expected: []interface{}{map[string]interface{}{"id": float64(2)}},
		},
		{
			name:     "filter by comparison",
			json:     `{"items": [{"id": 1}, {"id": 2}, {"id": 3}]}`,
			path:     `$.items[?(@.id>2)]`,
			expected: []interface{}{map[string]interface{}{"id": float64(3)}},
		},
		{
			name:     "filter with nested field",
			json:     `{"items": [{"user": {"age": 25}}, {"user": {"age": 30}}]}`,
			path:     `$.items[?(@.user.age>27)]`,
			expected: []interface{}{map[string]interface{}{"user": map[string]interface{}{"age": float64(30)}}},
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
