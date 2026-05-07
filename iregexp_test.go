package jsonpath

import (
	"regexp"
	"testing"
)

func TestIRegexpValidation(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		valid   bool
	}{
		{"Empty pattern", "", true},
		{"Simple pattern", "^S.*$", true},
		{"Character class", "[abc]", true},
		{"Negated class", "[^abc]", true},
		{"Quantifier +", "a+", true},
		{"Quantifier *", "a*", true},
		{"Quantifier ?", "a?", true},
		{"Range quantifier exact", "a{3}", true},
		{"Range quantifier min", "a{3,}", true},
		{"Range quantifier range", "a{3,5}", true},
		{"Alternation", "a|b", true},
		{"Group", "(a)", true},
		{"Nested groups", "((a))", true},
		{"Unicode property", "\\p{L}", true},
		{"Unicode property negated", "\\P{L}", true},
		{"Shorthand \\d", "\\d+", true},
		{"Shorthand \\w", "\\w+", true},
		{"Shorthand \\s", "\\s+", true},
		{"Shorthand \\D", "\\D+", true},
		{"Shorthand \\W", "\\W+", true},
		{"Shorthand \\S", "\\S+", true},
		{"Dot", ".", true},
		{"Anchors", "^abc$", true},
		{"Escaped metachar", "\\.", true},
		{"Complex pattern", "^[a-z]+\\d{2,4}$", true},
		{"Alternation with group", "(a|b)+", true},
		{"Class with shorthand", "[\\d\\w]", true},
		{"Class with unicode", "[\\p{L}]", true},
		{"Class with negated unicode", "[\\P{N}]", true},
		{"Range quantifier zero", "a{0}", true},
		{"Group with quantifier", "(ab){2,3}", true},
		{"Back reference \\1", "(a)\\1", false},
		{"Back reference \\9", "(a)\\9", false},
		{"Lookahead", "(?=a)", false},
		{"Negative lookahead", "(?!a)", false},
		{"Lookbehind", "(?<=a)", false},
		{"Negative lookbehind", "(?<!a)", false},
		{"Non-capturing group", "(?:a)", false},
		{"Named group", "(?<name>a)", false},
		{"Quantifier without operand star", "*", false},
		{"Quantifier without operand plus", "+", false},
		{"Quantifier without operand question", "?", false},
		{"Unclosed group", "(abc", false},
		{"Unmatched close paren", "abc)", false},
		{"Unclosed class", "[abc", false},
		{"Empty range quantifier", "a{}", false},
		{"Invalid range quantifier", "a{b}", false},
		{"Reversed range", "a{5,3}", false},
		{"Unclosed escape", "a\\", false},
		{"Unclosed unicode property", "\\p{L", false},
		{"Trailing backslash in class", "[a\\", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidIRegexp(tt.pattern)
			if result != tt.valid {
				t.Errorf("IsValidIRegexp(%q) = %v, want %v", tt.pattern, result, tt.valid)
			}
		})
	}
}

func TestIRegexpToGoRegexp(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		{"Simple pattern", "^S.*$", false},
		{"Character class", "[abc]", false},
		{"Quantifier", "a+", false},
		{"Alternation", "a|b", false},
		{"Group", "(a)", false},
		{"Unicode property", "\\p{L}", false},
		{"Shorthand", "\\d+", false},
		{"Range quantifier", "a{2,4}", false},
		{"Complex", "^[a-z]+\\d{2,4}$", false},
		{"Invalid lookahead", "(?=a)", true},
		{"Invalid non-capturing", "(?:a)", true},
		{"Invalid backreference", "(a)\\1", true},
		{"Invalid empty range", "a{}", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goRegexp, err := IRegexpToGoRegexp(tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Errorf("IRegexpToGoRegexp(%q) error = %v, wantErr %v", tt.pattern, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				_, err := regexp.Compile(goRegexp)
				if err != nil {
					t.Errorf("Failed to compile converted regexp %q: %v", goRegexp, err)
				}
			}
		})
	}
}

func TestIRegexpToGoRegexpMatching(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		input   string
		match   bool
	}{
		{"Simple match", "^S.*$", "String", true},
		{"Simple no match", "^S.*$", "hello", false},
		{"Character class", "[abc]", "b", true},
		{"Character class no match", "[abc]", "d", false},
		{"Quantifier", "a+", "aaa", true},
		{"Quantifier no match", "a+", "bbb", false},
		{"Alternation", "cat|dog", "dog", true},
		{"Alternation no match", "cat|dog", "fish", false},
		{"Group with quantifier", "(ab)+", "abab", true},
		{"Shorthand", "\\d+", "123", true},
		{"Shorthand no match", "\\d+", "abc", false},
		{"Dot", "a.c", "abc", true},
		{"Range quantifier", "a{2,3}", "aa", true},
		{"Range quantifier no match", "a{2,3}", "a", false},
		{"Unicode property", "\\p{L}+", "Hello", true},
		{"Unicode property no match", "\\p{L}+", "123", false},
		{"Anchors", "^abc$", "abc", true},
		{"Anchors no match", "^abc$", "abcd", false},
		{"Escaped dot", "a\\.b", "a.b", true},
		{"Escaped dot no match", "a\\.b", "axb", false},
		{"Negated class", "[^abc]", "d", true},
		{"Negated class no match", "[^abc]", "a", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goRegexp, err := IRegexpToGoRegexp(tt.pattern)
			if err != nil {
				t.Fatalf("IRegexpToGoRegexp(%q) error: %v", tt.pattern, err)
			}
			re, err := regexp.Compile(goRegexp)
			if err != nil {
				t.Fatalf("Failed to compile %q: %v", goRegexp, err)
			}
			got := re.MatchString(tt.input)
			if got != tt.match {
				t.Errorf("regexp %q.MatchString(%q) = %v, want %v", tt.pattern, tt.input, got, tt.match)
			}
		})
	}
}
