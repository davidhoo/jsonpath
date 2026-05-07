package jsonpath

import (
	"fmt"
	"strings"
	"unicode"
)

// NormalizedPathGenerator generates RFC 9535 Normalized Paths
type NormalizedPathGenerator struct {
	segments []string
}

func NewNormalizedPathGenerator() *NormalizedPathGenerator {
	return &NormalizedPathGenerator{segments: []string{"$"}}
}

func (npg *NormalizedPathGenerator) AddMember(name string) *NormalizedPathGenerator {
	escaped := escapeMemberName(name)
	npg.segments = append(npg.segments, fmt.Sprintf("['%s']", escaped))
	return npg
}

func (npg *NormalizedPathGenerator) AddIndex(index int) *NormalizedPathGenerator {
	npg.segments = append(npg.segments, fmt.Sprintf("[%d]", index))
	return npg
}

func (npg *NormalizedPathGenerator) String() string {
	return strings.Join(npg.segments, "")
}

func escapeMemberName(name string) string {
	var result strings.Builder
	for _, r := range name {
		switch {
		case r == '\'':
			result.WriteString("\\'")
		case r == '\\':
			result.WriteString("\\\\")
		case unicode.IsControl(r):
			result.WriteString(fmt.Sprintf("\\u%04x", r))
		default:
			result.WriteRune(r)
		}
	}
	return result.String()
}

// GenerateNormalizedPath generates from path segments
func GenerateNormalizedPath(segments []interface{}) string {
	npg := NewNormalizedPathGenerator()
	for _, seg := range segments {
		switch s := seg.(type) {
		case string:
			npg.AddMember(s)
		case int:
			npg.AddIndex(s)
		}
	}
	return npg.String()
}