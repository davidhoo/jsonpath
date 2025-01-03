package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/davidhoo/jsonpath"
	"github.com/fatih/color"
)

func init() {
	// 在测试环境中强制启用颜色输出
	color.NoColor = false
	os.Setenv("FORCE_COLOR", "true")
}

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
			name:     "with -h flag",
			args:     []string{"-h"},
			wantPath: "",
			wantFile: "",
			wantErr:  true,
		},
		{
			name:     "with --help flag",
			args:     []string{"--help"},
			wantPath: "",
			wantFile: "",
			wantErr:  true,
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
			if !tt.wantErr {
				var gotObj, wantObj interface{}
				if err := json.Unmarshal([]byte(got), &gotObj); err != nil {
					t.Fatal(err)
				}
				if err := json.Unmarshal([]byte(tt.want), &wantObj); err != nil {
					t.Fatal(err)
				}
				if !jsonEqual(gotObj, wantObj) {
					t.Errorf("readInput() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestNoInputShowsHelp(t *testing.T) {
	// 保存原始的 os.Args 和 os.Stderr
	oldArgs := os.Args
	oldStderr := os.Stderr
	defer func() {
		os.Args = oldArgs
		os.Stderr = oldStderr
	}()

	// 创建管道来捕获输出
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = w

	// 设置空参数
	os.Args = []string{"jp"}

	// 在一个 goroutine 中运行程序
	exit := make(chan bool)
	go func() {
		err := run()
		if err != nil {
			t.Error(err)
		}
		w.Close()
		exit <- true
	}()

	// 读取输出
	var output strings.Builder
	io.Copy(&output, r)
	<-exit

	// 检查输出是否包含帮助信息的关键部分
	if !strings.Contains(output.String(), "A JSONPath processor that fully complies with RFC 9535") {
		t.Error("Help message not shown when no input provided")
	}
}

func TestPrintVersion(t *testing.T) {
	// 捕获标准输出
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// 调用函数
	printVersion()

	// 恢复标准输出
	w.Close()
	os.Stdout = old

	// 读取捕获的输出
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// 验证输出包含版本信息
	version := jsonpath.VersionWithPrefix()
	if !strings.Contains(output, version) {
		t.Errorf("printVersion() output does not contain version %q", version)
	}
}

func TestOutputResult(t *testing.T) {
	tests := []struct {
		name    string
		result  interface{}
		cfg     *config
		want    string
		wantErr bool
	}{
		{
			name:   "string result without color",
			result: "test string",
			cfg:    &config{noColor: true},
			want:   "test string\n",
		},
		{
			name:   "string result with color",
			result: "test string",
			cfg:    &config{noColor: false},
			want:   "\x1b[32mtest string\x1b[0m\n",
		},
		{
			name:   "object result compact without color",
			result: map[string]interface{}{"key": "value"},
			cfg:    &config{noColor: true, compact: true},
			want:   "{\"key\":\"value\"}\n",
		},
		{
			name:   "object result pretty without color",
			result: map[string]interface{}{"key": "value"},
			cfg:    &config{noColor: true, compact: false, indent: "  "},
			want:   "{\n  \"key\": \"value\"\n}\n",
		},
		{
			name:   "array result compact without color",
			result: []interface{}{1, "two", true},
			cfg:    &config{noColor: true, compact: true},
			want:   "[1,\"two\",true]\n",
		},
		{
			name:   "array result pretty without color",
			result: []interface{}{1, "two", true},
			cfg:    &config{noColor: true, compact: false, indent: "  "},
			want:   "[\n  1,\n  \"two\",\n  true\n]\n",
		},
		{
			name:    "invalid result",
			result:  make(chan int),
			cfg:     &config{noColor: true},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 捕获标准输出
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// 调用函数
			err := outputResult(tt.result, tt.cfg)

			// 恢复标准输出
			w.Close()
			os.Stdout = old

			// 读取捕获的输出
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			// 验证错误
			if (err != nil) != tt.wantErr {
				t.Errorf("outputResult() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// 验证输出
			if !tt.wantErr && output != tt.want {
				t.Errorf("outputResult() output = %q, want %q", output, tt.want)
			}
		})
	}
}

func TestColorizeJSON(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty object",
			input: "{}",
			want:  "\x1b[90m{\x1b[0m\x1b[90m}\x1b[0m",
		},
		{
			name:  "empty array",
			input: "[]",
			want:  "\x1b[90m[\x1b[0m\x1b[90m]\x1b[0m",
		},
		{
			name:  "simple string",
			input: `"hello"`,
			want:  "\x1b[32m\"\x1b[0m\x1b[32mhello\x1b[0m\x1b[32m\"\x1b[0m",
		},
		{
			name:  "simple number",
			input: "42",
			want:  "\x1b[34m42\x1b[0m",
		},
		{
			name:  "simple boolean",
			input: "true",
			want:  "\x1b[33mtrue\x1b[0m",
		},
		{
			name:  "simple null",
			input: "null",
			want:  "\x1b[31mnull\x1b[0m",
		},
		{
			name:  "object with string",
			input: `{"key":"value"}`,
			want:  "\x1b[90m{\x1b[0m\x1b[36m\"\x1b[0m\x1b[36mkey\x1b[0m\x1b[36m\"\x1b[0m\x1b[90m:\x1b[0m\x1b[32m\"\x1b[0m\x1b[32mvalue\x1b[0m\x1b[32m\"\x1b[0m\x1b[90m}\x1b[0m",
		},
		{
			name:  "object with number",
			input: `{"key":42}`,
			want:  "\x1b[90m{\x1b[0m\x1b[36m\"\x1b[0m\x1b[36mkey\x1b[0m\x1b[36m\"\x1b[0m\x1b[90m:\x1b[0m\x1b[34m42\x1b[0m\x1b[90m}\x1b[0m",
		},
		{
			name:  "object with boolean",
			input: `{"key":true}`,
			want:  "\x1b[90m{\x1b[0m\x1b[36m\"\x1b[0m\x1b[36mkey\x1b[0m\x1b[36m\"\x1b[0m\x1b[90m:\x1b[0m\x1b[33mtrue\x1b[0m\x1b[90m}\x1b[0m",
		},
		{
			name:  "object with null",
			input: `{"key":null}`,
			want:  "\x1b[90m{\x1b[0m\x1b[36m\"\x1b[0m\x1b[36mkey\x1b[0m\x1b[36m\"\x1b[0m\x1b[90m:\x1b[0m\x1b[31mnull\x1b[0m\x1b[90m}\x1b[0m",
		},
		{
			name:  "array with mixed types",
			input: `[1,"two",true,null]`,
			want:  "\x1b[90m[\x1b[0m\x1b[34m1\x1b[0m\x1b[90m,\x1b[0m\x1b[32m\"\x1b[0m\x1b[32mtwo\x1b[0m\x1b[32m\"\x1b[0m\x1b[90m,\x1b[0m\x1b[33mtrue\x1b[0m\x1b[90m,\x1b[0m\x1b[31mnull\x1b[0m\x1b[90m]\x1b[0m",
		},
		{
			name:  "nested object",
			input: `{"outer":{"inner":"value"}}`,
			want:  "\x1b[90m{\x1b[0m\x1b[36m\"\x1b[0m\x1b[36mouter\x1b[0m\x1b[36m\"\x1b[0m\x1b[90m:\x1b[0m\x1b[90m{\x1b[0m\x1b[36m\"\x1b[0m\x1b[36minner\x1b[0m\x1b[36m\"\x1b[0m\x1b[90m:\x1b[0m\x1b[32m\"\x1b[0m\x1b[32mvalue\x1b[0m\x1b[32m\"\x1b[0m\x1b[90m}\x1b[0m\x1b[90m}\x1b[0m",
		},
		{
			name:  "escaped quotes in string",
			input: `{"key":"value\"with\"quotes"}`,
			want:  "\x1b[90m{\x1b[0m\x1b[36m\"\x1b[0m\x1b[36mkey\x1b[0m\x1b[36m\"\x1b[0m\x1b[90m:\x1b[0m\x1b[32m\"\x1b[0m\x1b[32mvalue\\\"with\\\"quotes\x1b[0m\x1b[32m\"\x1b[0m\x1b[90m}\x1b[0m",
		},
		{
			name:  "scientific notation",
			input: `{"key":1.23e-4}`,
			want:  "\x1b[90m{\x1b[0m\x1b[36m\"\x1b[0m\x1b[36mkey\x1b[0m\x1b[36m\"\x1b[0m\x1b[90m:\x1b[0m\x1b[34m1.23e-4\x1b[0m\x1b[90m}\x1b[0m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := colorizeJSON(tt.input)
			if got != tt.want {
				t.Errorf("colorizeJSON() = %q, want %q", got, tt.want)
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

func TestRun(t *testing.T) {
	// 保存原始的标准输入输出和参数
	oldStdin := os.Stdin
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	oldArgs := os.Args
	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
		os.Stderr = oldStderr
		os.Args = oldArgs
	}()

	tests := []struct {
		name     string
		args     []string
		input    string
		wantOut  string
		wantErr  bool
		setupEnv func()
	}{
		{
			name:    "no args shows help",
			args:    []string{"jp"},
			wantOut: "A JSONPath processor that fully complies with RFC 9535",
			wantErr: false,
		},
		{
			name:    "version flag",
			args:    []string{"jp", "-v"},
			wantOut: "jp version v2.0.1",
			wantErr: false,
		},
		{
			name:    "help flag",
			args:    []string{"jp", "-h"},
			wantOut: "A JSONPath processor that fully complies with RFC 9535",
			wantErr: false,
		},
		{
			name:    "invalid flag",
			args:    []string{"jp", "-x"},
			wantErr: true,
		},
		{
			name:    "invalid JSON input",
			args:    []string{"jp"},
			input:   "invalid json",
			wantErr: true,
		},
		{
			name:    "valid JSON without query",
			args:    []string{"jp"},
			input:   `{"name":"test"}`,
			wantOut: `{`,
			wantErr: false,
		},
		{
			name:    "valid JSON with query",
			args:    []string{"jp", "-p", "$.name"},
			input:   `{"name":"test"}`,
			wantOut: "test",
			wantErr: false,
		},
		{
			name:    "invalid query",
			args:    []string{"jp", "-p", "$[invalid]"},
			input:   `{"name":"test"}`,
			wantErr: true,
		},
		{
			name:    "compact output",
			args:    []string{"jp", "-c"},
			input:   `{"name":"test"}`,
			wantOut: `{"name":"test"}`,
			wantErr: false,
		},
		{
			name:    "no color output",
			args:    []string{"jp", "--no-color"},
			input:   `{"name":"test"}`,
			wantOut: `{`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建临时文件用于捕获输出
			outFile, err := os.CreateTemp("", "out")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(outFile.Name())
			os.Stdout = outFile

			// 创建临时文件用于捕获错误输出
			errFile, err := os.CreateTemp("", "err")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(errFile.Name())
			os.Stderr = errFile

			// 如果有输入，创建一个临时文件作为标准输入
			if tt.input != "" {
				inFile, err := os.CreateTemp("", "in")
				if err != nil {
					t.Fatal(err)
				}
				defer os.Remove(inFile.Name())
				if _, err := inFile.WriteString(tt.input); err != nil {
					t.Fatal(err)
				}
				if _, err := inFile.Seek(0, 0); err != nil {
					t.Fatal(err)
				}
				os.Stdin = inFile
			}

			// 设置命令行参数
			os.Args = tt.args

			// 如果有环境设置函数，执行它
			if tt.setupEnv != nil {
				tt.setupEnv()
			}

			// 运行测试
			err = run()

			// 检查错误
			if (err != nil) != tt.wantErr {
				t.Errorf("run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// 如果期望输出不为空，检查输出
			if tt.wantOut != "" {
				var output string
				if tt.name == "no args shows help" || tt.name == "help flag" {
					if _, err := errFile.Seek(0, 0); err != nil {
						t.Fatal(err)
					}
					errOutput, err := io.ReadAll(errFile)
					if err != nil {
						t.Fatal(err)
					}
					output = string(errOutput)
				} else {
					if _, err := outFile.Seek(0, 0); err != nil {
						t.Fatal(err)
					}
					outOutput, err := io.ReadAll(outFile)
					if err != nil {
						t.Fatal(err)
					}
					output = string(outOutput)
				}

				// 移除颜色代码
				output = removeANSIEscapeCodes(output)
				// 移除换行符
				output = strings.TrimSpace(output)

				if !strings.Contains(output, tt.wantOut) {
					t.Errorf("run() output = %v, want %v", output, tt.wantOut)
				}
			}
		})
	}
}

// removeANSIEscapeCodes 移除 ANSI 转义序列（颜色代码）
func removeANSIEscapeCodes(s string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	return re.ReplaceAllString(s, "")
}
