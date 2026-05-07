package jsonpath

import "testing"

func TestNormalizedPathGeneration(t *testing.T) {
	tests := []struct {
		name     string
		segments []interface{}
		expected string
	}{
		{"Root", []interface{}{}, "$"},
		{"Simple member", []interface{}{"store"}, "$['store']"},
		{"Nested members", []interface{}{"store", "book"}, "$['store']['book']"},
		{"Array index", []interface{}{"store", "book", 0}, "$['store']['book'][0]"},
		{"Member with single quote", []interface{}{"it's"}, "$['it\\'s']"},
		{"Member with backslash", []interface{}{"back\\slash"}, "$['back\\\\slash']"},
		{"Empty member name", []interface{}{""}, "$['']"},
		{"Member with control char", []interface{}{"tab\there"}, "$['tab\\u0009here']"},
		{"Numeric index", []interface{}{42}, "$[42]"},
		{"Negative index", []interface{}{-1}, "$[-1]"},
		{"Mixed segments", []interface{}{"a", 0, "b", 1}, "$['a'][0]['b'][1]"},
		{"Member with special chars", []interface{}{"key with spaces"}, "$['key with spaces']"},
		{"Member with unicode", []interface{}{"名前"}, "$['名前']"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateNormalizedPath(tt.segments)
			if result != tt.expected {
				t.Errorf("GenerateNormalizedPath(%v) = %q, want %q", tt.segments, result, tt.expected)
			}
		})
	}
}

func TestNormalizedPathGeneratorMethods(t *testing.T) {
	npg := NewNormalizedPathGenerator()
	if npg.String() != "$" {
		t.Errorf("NewNormalizedPathGenerator().String() = %q, want %q", npg.String(), "$")
	}

	npg.AddMember("test")
	if npg.String() != "$['test']" {
		t.Errorf("After AddMember(\"test\") = %q, want %q", npg.String(), "$['test']")
	}

	npg.AddIndex(0)
	if npg.String() != "$['test'][0]" {
		t.Errorf("After AddIndex(0) = %q, want %q", npg.String(), "$['test'][0]")
	}
}

func TestEscapeMemberName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"No escaping", "simple", "simple"},
		{"Single quote", "it's", "it\\'s"},
		{"Backslash", "back\\slash", "back\\\\slash"},
		{"Tab", "\t", "\\u0009"},
		{"Newline", "\n", "\\u000a"},
		{"Carriage return", "\r", "\\u000d"},
		{"Mixed", "a'b\\c", "a\\'b\\\\c"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeMemberName(tt.input)
			if result != tt.expected {
				t.Errorf("escapeMemberName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}