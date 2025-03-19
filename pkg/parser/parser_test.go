package parser

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		want    []string
		wantErr bool
	}{
		{
			name: "empty path",
			path: "",
			want: nil,
		},
		{
			name: "root only",
			path: "$",
			want: nil,
		},
		{
			name: "simple dot notation",
			path: "$.name",
			want: []string{"name"},
		},
		{
			name: "recursive descent",
			path: "$..name",
			want: []string{"..", "name"},
		},
		{
			name: "bracket notation",
			path: "$['name']",
			want: []string{"name"},
		},
		{
			name: "wildcard",
			path: "$.*",
			want: []string{"*"},
		},
		{
			name:    "missing $ prefix",
			path:    "name",
			wantErr: true,
		},
		{
			name:    "unclosed bracket",
			path:    "$[name",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("Parse() got %v segments, want %v", len(got), len(tt.want))
				return
			}
			for i, seg := range got {
				if seg.String() != tt.want[i] {
					t.Errorf("Parse() segment[%d] = %v, want %v", i, seg.String(), tt.want[i])
				}
			}
		})
	}
}
