package parser

import (
	"testing"

	"github.com/davidhoo/jsonpath/pkg/errors"
	"github.com/davidhoo/jsonpath/pkg/segments"
	"github.com/stretchr/testify/assert"
)

func TestParseBracketSegment(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    segments.Segment
		wantErr bool
		errType errors.ErrorType
		errMsg  string
	}{
		{
			name:    "empty bracket",
			input:   "[]",
			wantErr: true,
			errType: errors.ErrSyntax,
			errMsg:  "empty bracket expression",
		},
		{
			name:  "wildcard",
			input: "[*]",
			want:  &segments.WildcardSegment{},
		},
		{
			name:  "single index",
			input: "[0]",
			want:  segments.NewIndexSegment(0),
		},
		{
			name:  "single name",
			input: "[name]",
			want:  segments.NewNameSegment("name"),
		},
		{
			name:  "quoted name",
			input: "['name']",
			want:  segments.NewNameSegment("name"),
		},
		{
			name:  "double quoted name",
			input: `["name"]`,
			want:  segments.NewNameSegment("name"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseBracketSegment(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				if err != nil {
					assert.Equal(t, tt.errType, err.(*errors.Error).Type)
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want.String(), got.String())
		})
	}
}

func TestParseMultiIndexSegment(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    segments.Segment
		wantErr bool
		errType errors.ErrorType
		errMsg  string
	}{
		{
			name:  "multiple indices",
			input: "[0,1,2]",
			want:  segments.NewMultiIndexSegment([]int{0, 1, 2}),
		},
		{
			name:  "multiple names",
			input: "['a','b','c']",
			want:  segments.NewMultiNameSegment([]string{"a", "b", "c"}),
		},
		{
			name:  "multiple names without quotes",
			input: "[a,b,c]",
			want:  segments.NewMultiNameSegment([]string{"a", "b", "c"}),
		},
		{
			name:    "mixed index and name",
			input:   "[0,'a']",
			wantErr: true,
			errType: errors.ErrSyntax,
			errMsg:  "mixed index and name segments",
		},
		{
			name:    "empty segment",
			input:   "[,]",
			wantErr: true,
			errType: errors.ErrSyntax,
			errMsg:  "empty multi-index segment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseMultiIndexSegment(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				if err != nil {
					assert.Equal(t, tt.errType, err.(*errors.Error).Type)
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want.String(), got.String())
		})
	}
}

func TestParseSliceSegment(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    segments.Segment
		wantErr bool
		errType errors.ErrorType
		errMsg  string
	}{
		{
			name:  "simple slice",
			input: "[1:3]",
			want:  segments.NewSliceSegment(1, 3, 1),
		},
		{
			name:  "slice with step",
			input: "[1:3:2]",
			want:  segments.NewSliceSegment(1, 3, 2),
		},
		{
			name:  "slice from start",
			input: "[:3]",
			want:  segments.NewSliceSegment(0, 3, 1),
		},
		{
			name:  "slice to end",
			input: "[1:]",
			want:  segments.NewSliceSegment(1, -1, 1),
		},
		{
			name:    "invalid slice",
			input:   "[1:2:3:4]",
			wantErr: true,
			errType: errors.ErrSyntax,
			errMsg:  "invalid slice expression",
		},
		{
			name:    "invalid start index",
			input:   "[a:3]",
			wantErr: true,
			errType: errors.ErrSyntax,
			errMsg:  "invalid start index",
		},
		{
			name:    "invalid end index",
			input:   "[1:b]",
			wantErr: true,
			errType: errors.ErrSyntax,
			errMsg:  "invalid end index",
		},
		{
			name:    "invalid step",
			input:   "[1:3:0]",
			wantErr: true,
			errType: errors.ErrSyntax,
			errMsg:  "step cannot be zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSliceSegment(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				if err != nil {
					assert.Equal(t, tt.errType, err.(*errors.Error).Type)
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want.String(), got.String())
		})
	}
}

func TestParseFilterSegment(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    segments.Segment
		wantErr bool
		errType errors.ErrorType
		errMsg  string
	}{
		{
			name:  "simple filter",
			input: "[?(@.age > 18)]",
			want:  segments.NewFilterSegment("age > 18"),
		},
		{
			name:    "empty filter",
			input:   "[?]",
			wantErr: true,
			errType: errors.ErrSyntax,
			errMsg:  "empty filter expression",
		},
		{
			name:    "missing @",
			input:   "[?(age > 18)]",
			wantErr: true,
			errType: errors.ErrSyntax,
			errMsg:  "filter expression must start with @",
		},
		{
			name:    "empty after @",
			input:   "[?(@)]",
			wantErr: true,
			errType: errors.ErrSyntax,
			errMsg:  "empty filter expression after @",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFilterSegment(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				if err != nil {
					assert.Equal(t, tt.errType, err.(*errors.Error).Type)
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want.String(), got.String())
		})
	}
}

func TestParseFunctionCall(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    segments.Segment
		wantErr bool
		errType errors.ErrorType
		errMsg  string
	}{
		{
			name:  "simple function call",
			input: "[@.length()]",
			want:  segments.NewFunctionSegment("length", nil),
		},
		{
			name:  "function call with arguments",
			input: "[@.substring(0,3)]",
			want:  segments.NewFunctionSegment("substring", []string{"0", "3"}),
		},
		{
			name:    "empty function call",
			input:   "[@]",
			wantErr: true,
			errType: errors.ErrSyntax,
			errMsg:  "empty function call",
		},
		{
			name:    "missing opening parenthesis",
			input:   "[@.length)]",
			wantErr: true,
			errType: errors.ErrSyntax,
			errMsg:  "missing opening parenthesis",
		},
		{
			name:    "missing closing parenthesis",
			input:   "[@.length(]",
			wantErr: true,
			errType: errors.ErrSyntax,
			errMsg:  "missing closing parenthesis",
		},
		{
			name:    "extra characters after closing parenthesis",
			input:   "[@.length())extra]",
			wantErr: true,
			errType: errors.ErrSyntax,
			errMsg:  "extra characters after closing parenthesis",
		},
		{
			name:    "empty function name",
			input:   "[@()]",
			wantErr: true,
			errType: errors.ErrSyntax,
			errMsg:  "empty function name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFunctionCall(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				if err != nil {
					assert.Equal(t, tt.errType, err.(*errors.Error).Type)
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want.String(), got.String())
		})
	}
}
