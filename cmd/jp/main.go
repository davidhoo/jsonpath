package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/davidhoo/jsonpath"
	"github.com/fatih/color"
)

// 定义颜色函数
var (
	// 错误和成功提示颜色
	errorColor   = color.New(color.FgRed, color.Bold).SprintfFunc()
	successColor = color.New(color.FgGreen).SprintFunc()
	pathColor    = color.New(color.FgCyan).SprintFunc()

	// JSON 元素颜色
	braceColor      = color.New(color.FgMagenta).SprintFunc() // {} 大括号
	bracketColor    = color.New(color.FgYellow).SprintFunc()  // [] 方括号
	commaColor      = color.New(color.FgWhite).SprintFunc()   // 逗号
	colonColor      = color.New(color.FgWhite).SprintFunc()   // 冒号
	keyColor        = color.New(color.FgCyan).SprintFunc()    // 键名
	stringColor     = color.New(color.FgGreen).SprintFunc()   // 字符串值
	numberColor     = color.New(color.FgBlue).SprintFunc()    // 数字值
	booleanColor    = color.New(color.FgYellow).SprintFunc()  // 布尔值
	nullColor       = color.New(color.FgRed).SprintFunc()     // null 值
	keyQuoteColor   = color.New(color.FgMagenta).SprintFunc() // 键名的引号
	valueQuoteColor = color.New(color.FgGreen).SprintFunc()   // 值的引号
)

// 命令行参数
type config struct {
	path    string
	expr    string
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
const usage = `Usage: jp [options] [jsonpath]

Options:
  -p, --path       JSONPath expression (optional, output full JSON if not specified)
  -f, --file       JSON file path (if not specified, read from stdin)
  -c, --compact    Output compact JSON instead of pretty-printed
  -n, --no-color   Disable colored output
  -i, --indent     Indentation string for pretty-printing (default: "  ")
  -h, --help       Show this help message
  -v, --version    Show version information

Examples:
  jp -f data.json                               # Output entire JSON from file
  jp -f data.json -p '$.store.book[0].title'    # Get the title of the first book
  jp -f data.json '$.store.book[*].author'      # Get all book authors
  jp -p '$.store.book[?(@.price < 10)].title'   # Get titles of books cheaper than 10
  jp '$.store..price'                           # Get all prices in the store
  cat data.json | jp                            # Output entire JSON from stdin
  cat data.json | jp '$.store.book[0]'          # Get first book from stdin

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

	flag.StringVar(&cfg.expr, "p", "", "JSONPath expression")
	flag.StringVar(&cfg.expr, "path", "", "JSONPath expression")
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
	if cfg.expr != "" {
		cfg.path = cfg.expr
	} else if flag.NArg() > 0 {
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

	return false
}

// 读取输入
func readInput(filePath string) (interface{}, error) {
	var reader io.Reader

	if filePath != "" {
		// 从文件读取
		file, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("opening file: %w", err)
		}
		defer file.Close()
		reader = file
	} else {
		// 从标准输入读取
		reader = os.Stdin
	}

	// 使用 decoder 直接解码 JSON
	var data interface{}
	decoder := json.NewDecoder(reader)
	decoder.UseNumber()
	if err := decoder.Decode(&data); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}

	return data, nil
}

// 执行 JSONPath 查询
func executeQuery(path string, data interface{}) (interface{}, error) {
	// 如果没有指定 JSONPath 表达式，返回完整数据
	if path == "" {
		return data, nil
	}

	// 编译并执行 JSONPath 表达式
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
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", cfg.indent)
	enc.SetEscapeHTML(false)

	if err := enc.Encode(result); err != nil {
		return fmt.Errorf("encoding JSON: %w", err)
	}

	output := strings.TrimSuffix(buf.String(), "\n")
	if cfg.noColor {
		fmt.Println(output)
	} else {
		fmt.Println(colorizeJSON(output))
	}

	return nil
}

// 给 JSON 添加颜色
func colorizeJSON(jsonStr string) string {
	var result strings.Builder
	inString := false
	inKey := false
	var prev rune

	for i := 0; i < len(jsonStr); {
		r, size := utf8.DecodeRuneInString(jsonStr[i:])
		rest := jsonStr[i:]

		switch r {
		case '"':
			if prev != '\\' {
				if !inString {
					// 开始一个字符串
					inString = true
					if inKey {
						result.WriteString(keyQuoteColor("\""))
					} else {
						result.WriteString(valueQuoteColor("\""))
					}
				} else {
					// 结束一个字符串
					inString = false
					if inKey {
						result.WriteString(keyQuoteColor("\""))
						inKey = false
					} else {
						result.WriteString(valueQuoteColor("\""))
					}
				}
			} else {
				result.WriteRune(r)
			}

		case '{', '}':
			if !inString {
				result.WriteString(braceColor(string(r)))
			} else {
				result.WriteRune(r)
			}

		case '[', ']':
			if !inString {
				result.WriteString(bracketColor(string(r)))
			} else {
				result.WriteRune(r)
			}

		case ',':
			if !inString {
				result.WriteString(commaColor(string(r)))
			} else {
				result.WriteRune(r)
			}

		case ':':
			if !inString {
				result.WriteString(colonColor(string(r)))
				inKey = false
			} else {
				result.WriteRune(r)
			}

		default:
			if !inString {
				// 检查布尔值和 null
				switch {
				case strings.HasPrefix(rest, "true"):
					result.WriteString(booleanColor("true"))
					i += len("true") - 1
				case strings.HasPrefix(rest, "false"):
					result.WriteString(booleanColor("false"))
					i += len("false") - 1
				case strings.HasPrefix(rest, "null"):
					result.WriteString(nullColor("null"))
					i += len("null") - 1
				default:
					// 检查数字
					if unicode.IsDigit(r) || r == '-' || r == '.' {
						numStr := string(r)
						j := i + size
						for j < len(jsonStr) {
							next, nextSize := utf8.DecodeRuneInString(jsonStr[j:])
							if unicode.IsDigit(next) || next == '.' || next == 'e' || next == 'E' || next == '-' || next == '+' {
								numStr += string(next)
								j += nextSize
							} else {
								break
							}
						}
						if _, err := strconv.ParseFloat(numStr, 64); err == nil {
							result.WriteString(numberColor(numStr))
							i = j - 1
						} else {
							result.WriteRune(r)
						}
					} else {
						result.WriteRune(r)
					}
				}
			} else {
				if inKey {
					result.WriteString(keyColor(string(r)))
				} else {
					result.WriteString(stringColor(string(r)))
				}
			}
		}

		prev = r
		i += size

		// 检查是否是键名
		if !inString && len(strings.TrimSpace(result.String())) > 0 {
			inKey = strings.HasSuffix(strings.TrimSpace(result.String()), ":")
		}
	}

	return result.String()
}

// 输出错误并退出
func exitWithError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, errorColor(format+"\n", args...))
	os.Exit(1)
}
