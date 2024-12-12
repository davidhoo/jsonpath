package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/davidhoo/jsonpath"
	"github.com/fatih/color"
)

// 定义颜色
var (
	errorColor   = color.New(color.FgRed).SprintFunc()
	successColor = color.New(color.FgGreen).SprintFunc()
	pathColor    = color.New(color.FgCyan).SprintFunc()
)

// 命令行参数
type config struct {
	path    string
	file    string
	compact bool
	noColor bool
	indent  string
	help    bool
	version bool
}

// 版本信息
const version = "1.0.1"

// 帮助信息
const usage = `Usage: jp [options] <jsonpath>

Options:
  -f, --file       JSON file path (if not specified, read from stdin)
  -c, --compact    Output compact JSON instead of pretty-printed
  -n, --no-color   Disable colored output
  -i, --indent     Indentation string for pretty-printing (default: "  ")
  -h, --help       Show this help message
  -v, --version    Show version information

Examples:
  jp -f data.json '$.store.book[0].title'     # Get the title of the first book from file
  jp -f data.json '$.store.book[*].author'    # Get all book authors from file
  jp '$.store.book[?(@.price < 10)].title'    # Get titles of books cheaper than 10
  jp '$.store..price'                         # Get all prices in the store
  cat data.json | jp '$.store.book[0]'        # Read JSON from stdin

For more information and examples, visit:
https://github.com/davidhoo/jsonpath`

// 主函数
func main() {
	// 解析命令行参数
	cfg := parseFlags()

	// 处理特殊命令
	if handleSpecialCommands(cfg) {
		return
	}

	// 读取输入
	input, err := readInput(cfg.file)
	if err != nil {
		exitWithError("Error reading input: %v", err)
	}

	// 执行 JSONPath 查询
	result, err := executeQuery(cfg.path, input)
	if err != nil {
		exitWithError("Error executing query: %v", err)
	}

	// 输出结果
	if err := outputResult(result, cfg); err != nil {
		exitWithError("Error outputting result: %v", err)
	}
}

// 解析命令行参数
func parseFlags() *config {
	cfg := &config{}

	flag.StringVar(&cfg.file, "f", "", "JSON file path")
	flag.StringVar(&cfg.file, "file", "", "JSON file path")
	flag.BoolVar(&cfg.compact, "c", false, "Output compact JSON")
	flag.BoolVar(&cfg.compact, "compact", false, "Output compact JSON")
	flag.BoolVar(&cfg.noColor, "n", false, "Disable colored output")
	flag.BoolVar(&cfg.noColor, "no-color", false, "Disable colored output")
	flag.StringVar(&cfg.indent, "i", "  ", "Indentation string")
	flag.StringVar(&cfg.indent, "indent", "  ", "Indentation string")
	flag.BoolVar(&cfg.help, "h", false, "Show help message")
	flag.BoolVar(&cfg.help, "help", false, "Show help message")
	flag.BoolVar(&cfg.version, "v", false, "Show version")
	flag.BoolVar(&cfg.version, "version", false, "Show version")

	flag.Usage = func() {
		fmt.Println(usage)
	}

	flag.Parse()

	// 获取 JSONPath 表达式
	if flag.NArg() > 0 {
		cfg.path = flag.Arg(0)
	}

	return cfg
}

// 处理特殊命令（帮助和版本信息）
func handleSpecialCommands(cfg *config) bool {
	if cfg.help {
		fmt.Println(usage)
		return true
	}

	if cfg.version {
		fmt.Printf("jp version %s\n", version)
		return true
	}

	if cfg.path == "" {
		fmt.Fprintln(os.Stderr, errorColor("Error: JSONPath expression is required"))
		fmt.Println(usage)
		os.Exit(1)
	}

	return false
}

// 读取输入
func readInput(filePath string) (interface{}, error) {
	var bytes []byte
	var err error

	if filePath != "" {
		// 从文件读取
		bytes, err = os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("reading file: %w", err)
		}
	} else {
		// 从标准输入读取
		bytes, err = io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("reading stdin: %w", err)
		}
	}

	var data interface{}
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}

	return data, nil
}

// 执行 JSONPath 查询
func executeQuery(path string, data interface{}) (interface{}, error) {
	jp, err := jsonpath.Compile(path)
	if err != nil {
		return nil, fmt.Errorf("compiling JSONPath: %w", err)
	}

	result, err := jp.Execute(data)
	if err != nil {
		return nil, fmt.Errorf("executing JSONPath: %w", err)
	}

	return result, nil
}

// 输出结果
func outputResult(result interface{}, cfg *config) error {
	var output []byte
	var err error

	if cfg.compact {
		output, err = json.Marshal(result)
	} else {
		output, err = json.MarshalIndent(result, "", cfg.indent)
	}

	if err != nil {
		return fmt.Errorf("marshaling result: %w", err)
	}

	if cfg.noColor {
		fmt.Println(string(output))
	} else {
		fmt.Println(successColor(string(output)))
	}

	return nil
}

// 输出错误并退出
func exitWithError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, errorColor(format)+"\n", args...)
	os.Exit(1)
}
