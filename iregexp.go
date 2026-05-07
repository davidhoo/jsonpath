package jsonpath

import (
	"fmt"
	"strconv"
	"strings"
)

// IRegexpParser parses I-Regexp syntax (RFC 9485).
type IRegexpParser struct {
	pattern string
	pos     int
}

// NewIRegexpParser creates a new I-Regexp parser.
func NewIRegexpParser(pattern string) *IRegexpParser {
	return &IRegexpParser{pattern: pattern, pos: 0}
}

// IsValidIRegexp checks if pattern is valid I-Regexp.
func IsValidIRegexp(pattern string) bool {
	parser := NewIRegexpParser(pattern)
	_, err := parser.Parse()
	return err == nil
}

// IRegexpToGoRegexp converts I-Regexp to Go regexp pattern.
func IRegexpToGoRegexp(pattern string) (string, error) {
	parser := NewIRegexpParser(pattern)
	return parser.Parse()
}

// Parse parses the I-Regexp pattern and returns the equivalent Go regexp.
func (p *IRegexpParser) Parse() (string, error) {
	if len(p.pattern) == 0 {
		return "", nil
	}
	result, err := p.parseAlternation()
	if err != nil {
		return "", err
	}
	if p.pos < len(p.pattern) {
		return "", fmt.Errorf("unexpected character at position %d: %c", p.pos, p.pattern[p.pos])
	}
	return result, nil
}

func (p *IRegexpParser) parseAlternation() (string, error) {
	left, err := p.parseSequence()
	if err != nil {
		return "", err
	}
	for p.pos < len(p.pattern) && p.pattern[p.pos] == '|' {
		p.pos++
		right, err := p.parseSequence()
		if err != nil {
			return "", err
		}
		left = left + "|" + right
	}
	return left, nil
}

func (p *IRegexpParser) parseSequence() (string, error) {
	var b strings.Builder
	for p.pos < len(p.pattern) {
		ch := p.pattern[p.pos]
		if ch == '|' || ch == ')' {
			break
		}
		atom, err := p.parseAtom()
		if err != nil {
			return "", err
		}
		b.WriteString(atom)
	}
	return b.String(), nil
}

func (p *IRegexpParser) parseAtom() (string, error) {
	if p.pos >= len(p.pattern) {
		return "", fmt.Errorf("unexpected end of pattern")
	}
	ch := p.pattern[p.pos]
	switch {
	case ch == '(':
		return p.parseGroup()
	case ch == '[':
		return p.parseCharacterClass()
	case ch == '\\':
		return p.parseEscape()
	case ch == '.':
		p.pos++
		q, err := p.parseQuantifier()
		if err != nil {
			return "", err
		}
		return "." + q, nil
	case ch == '^' || ch == '$':
		p.pos++
		return string(ch), nil
	case ch == '*' || ch == '+' || ch == '?':
		return "", fmt.Errorf("quantifier without operand at position %d", p.pos)
	case ch == '}' || ch == ']':
		return "", fmt.Errorf("unexpected '%c' at position %d", ch, p.pos)
	default:
		p.pos++
		q, err := p.parseQuantifier()
		if err != nil {
			return "", err
		}
		return regexpQuoteMeta(string(ch)) + q, nil
	}
}

func (p *IRegexpParser) parseGroup() (string, error) {
	p.pos++ // '('
	if p.pos >= len(p.pattern) {
		return "", fmt.Errorf("unexpected end of pattern in group")
	}
	if p.pattern[p.pos] == '?' {
		return "", fmt.Errorf("unsupported group type '(?' at position %d", p.pos-1)
	}
	inner, err := p.parseAlternation()
	if err != nil {
		return "", err
	}
	if p.pos >= len(p.pattern) || p.pattern[p.pos] != ')' {
		return "", fmt.Errorf("missing closing parenthesis")
	}
	p.pos++ // ')'
	q, err := p.parseQuantifier()
	if err != nil {
		return "", err
	}
	return "(" + inner + ")" + q, nil
}

func (p *IRegexpParser) parseCharacterClass() (string, error) {
	p.pos++ // '['
	var b strings.Builder
	b.WriteByte('[')
	if p.pos < len(p.pattern) && p.pattern[p.pos] == '^' {
		b.WriteByte('^')
		p.pos++
	}
	// Handle ] as first char in class (or after ^)
	if p.pos < len(p.pattern) && p.pattern[p.pos] == ']' {
		b.WriteByte(']')
		p.pos++
	}
	for p.pos < len(p.pattern) && p.pattern[p.pos] != ']' {
		if p.pattern[p.pos] == '\\' {
			p.pos++
			if p.pos >= len(p.pattern) {
				return "", fmt.Errorf("unexpected end of pattern in character class")
			}
			ch := p.pattern[p.pos]
			p.pos++
			switch ch {
			case 'd', 'D', 'w', 'W', 's', 'S':
				b.WriteByte('\\')
				b.WriteByte(ch)
			case 'p', 'P':
				if p.pos >= len(p.pattern) || p.pattern[p.pos] != '{' {
					return "", fmt.Errorf("expected '{' after \\%c", ch)
				}
				b.WriteByte('\\')
				b.WriteByte(ch)
				b.WriteByte('{')
				p.pos++ // '{'
				for p.pos < len(p.pattern) && p.pattern[p.pos] != '}' {
					b.WriteByte(p.pattern[p.pos])
					p.pos++
				}
				if p.pos >= len(p.pattern) {
					return "", fmt.Errorf("unclosed Unicode property in character class")
				}
				b.WriteByte('}')
				p.pos++ // '}'
			default:
				b.WriteByte('\\')
				b.WriteByte(ch)
			}
		} else {
			b.WriteByte(p.pattern[p.pos])
			p.pos++
		}
	}
	if p.pos >= len(p.pattern) {
		return "", fmt.Errorf("unclosed character class")
	}
	b.WriteByte(']')
	p.pos++ // ']'
	q, err := p.parseQuantifier()
	if err != nil {
		return "", err
	}
	return b.String() + q, nil
}

func (p *IRegexpParser) parseEscape() (string, error) {
	p.pos++ // '\\'
	if p.pos >= len(p.pattern) {
		return "", fmt.Errorf("unexpected end of pattern after escape")
	}
	ch := p.pattern[p.pos]
	p.pos++
	switch {
	case ch >= '1' && ch <= '9':
		return "", fmt.Errorf("backreference \\%c is not allowed in I-Regexp", ch)
	case ch == 'p' || ch == 'P':
		return p.parseUnicodeProperty(ch)
	case ch == 'd', ch == 'D', ch == 'w', ch == 'W', ch == 's', ch == 'S':
		q, err := p.parseQuantifier()
		if err != nil {
			return "", err
		}
		return "\\" + string(ch) + q, nil
	default:
		q, err := p.parseQuantifier()
		if err != nil {
			return "", err
		}
		return "\\" + string(ch) + q, nil
	}
}

func (p *IRegexpParser) parseUnicodeProperty(prefix byte) (string, error) {
	if p.pos >= len(p.pattern) || p.pattern[p.pos] != '{' {
		return "", fmt.Errorf("expected '{' after \\%c", prefix)
	}
	p.pos++ // '{'
	start := p.pos
	for p.pos < len(p.pattern) && p.pattern[p.pos] != '}' {
		p.pos++
	}
	if p.pos >= len(p.pattern) {
		return "", fmt.Errorf("unclosed Unicode property")
	}
	prop := p.pattern[start:p.pos]
	p.pos++ // '}'
	q, err := p.parseQuantifier()
	if err != nil {
		return "", err
	}
	return "\\" + string(prefix) + "{" + prop + "}" + q, nil
}

func (p *IRegexpParser) parseQuantifier() (string, error) {
	if p.pos >= len(p.pattern) {
		return "", nil
	}
	ch := p.pattern[p.pos]
	switch ch {
	case '*', '+', '?':
		p.pos++
		return string(ch), nil
	case '{':
		return p.parseRangeQuantifier()
	default:
		return "", nil
	}
}

func (p *IRegexpParser) parseRangeQuantifier() (string, error) {
	start := p.pos
	p.pos++ // '{'
	if p.pos >= len(p.pattern) || !isDigit(p.pattern[p.pos]) {
		return "", fmt.Errorf("invalid range quantifier at position %d", start)
	}
	nStr := ""
	for p.pos < len(p.pattern) && isDigit(p.pattern[p.pos]) {
		nStr += string(p.pattern[p.pos])
		p.pos++
	}
	n, _ := strconv.Atoi(nStr)
	if p.pos >= len(p.pattern) {
		return "", fmt.Errorf("unclosed range quantifier")
	}
	if p.pattern[p.pos] == '}' {
		p.pos++
		return fmt.Sprintf("{%d}", n), nil
	}
	if p.pattern[p.pos] != ',' {
		return "", fmt.Errorf("invalid range quantifier at position %d", start)
	}
	p.pos++ // ','
	if p.pos >= len(p.pattern) {
		return "", fmt.Errorf("unclosed range quantifier")
	}
	if p.pattern[p.pos] == '}' {
		p.pos++
		return fmt.Sprintf("{%d,}", n), nil
	}
	if !isDigit(p.pattern[p.pos]) {
		return "", fmt.Errorf("invalid range quantifier at position %d", start)
	}
	mStr := ""
	for p.pos < len(p.pattern) && isDigit(p.pattern[p.pos]) {
		mStr += string(p.pattern[p.pos])
		p.pos++
	}
	m, _ := strconv.Atoi(mStr)
	if p.pos >= len(p.pattern) || p.pattern[p.pos] != '}' {
		return "", fmt.Errorf("unclosed range quantifier")
	}
	p.pos++
	if m < n {
		return "", fmt.Errorf("invalid range quantifier: %d > %d", n, m)
	}
	return fmt.Sprintf("{%d,%d}", n, m), nil
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

// regexpQuoteMeta quotes only characters that are special in Go regexp
// but are plain literals in I-Regexp context.
func regexpQuoteMeta(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		ch := s[i]
		switch ch {
		case '.', '*', '+', '?', '(', ')', '[', ']', '{', '}', '\\', '^', '$', '|':
			b.WriteByte('\\')
			b.WriteByte(ch)
		default:
			b.WriteByte(ch)
		}
	}
	return b.String()
}
