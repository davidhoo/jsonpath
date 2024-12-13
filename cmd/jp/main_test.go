package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"strings"
	"testing"
)

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantPath string
		wantFile string
		wantErr  bool
	}{
		{
			name:     "default flags",
			args:     []string{},
			wantPath: "",
			wantFile: "",
			wantErr:  false,
		},
		{
			name:     "with path",
			args:     []string{"-p", "$.store.book[*].author"},
			wantPath: "$.store.book[*].author",
			wantFile: "",
			wantErr:  false,
		},
		{
			name:     "with file",
			args:     []string{"-f", "test.json"},
			wantPath: "",
			wantFile: "test.json",
			wantErr:  false,
		},
		{
			name:     "with path and file",
			args:     []string{"-p", "$.store.book[*]", "-f", "test.json"},
			wantPath: "$.store.book[*]",
			wantFile: "test.json",
			wantErr:  false,
		},
		{
			name:     "invalid flag",
			args:     []string{"-x"},
			wantPath: "",
			wantFile: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			// Save and restore os.Args
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()

			os.Args = append([]string{"jp"}, tt.args...)
			path, file, err := ParseFlags()

			if (err != nil) != tt.wantErr {
				t.Errorf("parseFlags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if path != tt.wantPath {
					t.Errorf("parseFlags() path = %v, want %v", path, tt.wantPath)
				}
				if file != tt.wantFile {
					t.Errorf("parseFlags() file = %v, want %v", file, tt.wantFile)
				}
			}
		})
	}
}

func TestFormatJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		compress bool
		want     string
	}{
		{
			name:     "simple object",
			input:    map[string]interface{}{"name": "test"},
			compress: false,
			want:     "{\n  \"name\": \"test\"\n}",
		},
		{
			name:     "simple object compressed",
			input:    map[string]interface{}{"name": "test"},
			compress: true,
			want:     "{\"name\":\"test\"}",
		},
		{
			name:     "array",
			input:    []interface{}{"a", "b", "c"},
			compress: false,
			want:     "[\n  \"a\",\n  \"b\",\n  \"c\"\n]",
		},
		{
			name:     "array compressed",
			input:    []interface{}{"a", "b", "c"},
			compress: true,
			want:     "[\"a\",\"b\",\"c\"]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatJSON(tt.input, tt.compress)
			// Normalize line endings
			got = strings.ReplaceAll(got, "\r\n", "\n")
			if got != tt.want {
				t.Errorf("formatJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadInput(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "valid json",
			input:   "{\"name\":\"test\"}",
			want:    "{\"name\":\"test\"}",
			wantErr: false,
		},
		{
			name:    "invalid json",
			input:   "{invalid}",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary file
			tmpfile, err := os.CreateTemp("", "test*.json")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(tmpfile.Name())

			// Write test data
			if _, err := tmpfile.Write([]byte(tt.input)); err != nil {
				t.Fatal(err)
			}
			if err := tmpfile.Close(); err != nil {
				t.Fatal(err)
			}

			// Test reading from file
			got, err := readInput(tmpfile.Name())
			if (err != nil) != tt.wantErr {
				t.Errorf("readInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("readInput() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function to compare JSON objects
func jsonEqual(a, b interface{}) bool {
	aj, _ := json.Marshal(a)
	bj, _ := json.Marshal(b)
	return bytes.Equal(aj, bj)
}
