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

var (
	// 颜色定义
	keyColor        = color.New(color.FgWhite).SprintFunc()
	keyQuoteColor   = color.New(color.FgWhite).SprintFunc()
	stringColor     = color.New(color.FgGreen).SprintFunc()
	valueQuoteColor = color.New(color.FgGreen).SprintFunc()
	numberColor     = color.New(color.FgBlue).SprintFunc()
	booleanColor    = color.New(color.FgYellow).SprintFunc()
	nullColor       = color.New(color.FgRed).SprintFunc()
	braceColor      = color.New(color.FgHiBlack).SprintFunc()
	bracketColor    = color.New(color.FgHiBlack).SprintFunc()
	commaColor      = color.New(color.FgHiBlack).SprintFunc()
	colonColor      = color.New(color.FgHiBlack).SprintFunc()

	// 命令行帮助颜色
	cmdColor     = color.New(color.FgCyan).SprintFunc()
	flagColor    = color.New(color.FgYellow).SprintFunc()
	descColor    = color.New(color.FgWhite).SprintFunc()
	exampleColor = color.New(color.FgGreen).SprintFunc()
	versionColor = color.New(color.FgMagenta).SprintFunc()
	errorColor   = color.New(color.FgRed, color.Bold).SprintFunc()
)

// 配置
type config struct {
	path    string
	file    string
	compact bool
	noColor bool
	indent  string
}

func main() {
	cfg := &config{
		indent: "  ",
	}

	// 自定义 usage
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n\n", descColor("A JSONPath processor that fully complies with RFC 9535"))
		fmt.Fprintf(os.Stderr, "%s\n", descColor("Usage:"))
		fmt.Fprintf(os.Stderr, "  %s %s %s\n\n",
			cmdColor("jp"),
			flagColor("[-p <jsonpath>]"),
			flagColor("[-f <file>]"),
		)
		fmt.Fprintf(os.Stderr, "%s\n", descColor("Options:"))
		fmt.Fprintf(os.Stderr, "  %s  %s %s\n",
			flagColor("-p"),
			descColor("JSONPath expression"),
			descColor("(if not specified, output entire JSON)"),
		)
		fmt.Fprintf(os.Stderr, "  %s  %s %s\n",
			flagColor("-f"),
			descColor("JSON file path"),
			descColor("(reads from stdin if not specified)"),
		)
		fmt.Fprintf(os.Stderr, "  %s  %s\n",
			flagColor("-c"),
			descColor("Compact output (no formatting)"),
		)
		fmt.Fprintf(os.Stderr, "  %s  %s\n",
			flagColor("--no-color"),
			descColor("Disable colored output"),
		)
		fmt.Fprintf(os.Stderr, "  %s  %s\n",
			flagColor("-h"),
			descColor("Show this help"),
		)
		fmt.Fprintf(os.Stderr, "  %s  %s\n\n",
			flagColor("-v"),
			descColor("Show version"),
		)
		fmt.Fprintf(os.Stderr, "%s\n", descColor("Examples:"))
		fmt.Fprintf(os.Stderr, "  %s\n", exampleColor("# Output entire JSON with colors"))
		fmt.Fprintf(os.Stderr, "  %s %s\n\n",
			cmdColor("jp"),
			flagColor("-f data.json"),
		)
		fmt.Fprintf(os.Stderr, "  %s\n", exampleColor("# Query specific path"))
		fmt.Fprintf(os.Stderr, "  %s %s %s\n\n",
			cmdColor("jp"),
			flagColor("-f data.json"),
			flagColor("-p '$.store.book[*].author'"),
		)
		fmt.Fprintf(os.Stderr, "  %s\n", exampleColor("# Filter by condition"))
		fmt.Fprintf(os.Stderr, "  %s %s %s\n\n",
			cmdColor("jp"),
			flagColor("-f data.json"),
			flagColor("-p '$.store.book[?(@.price > 10)]'"),
		)
		fmt.Fprintf(os.Stderr, "  %s\n", exampleColor("# Read from stdin"))
		fmt.Fprintf(os.Stderr, "  %s %s %s\n\n",
			exampleColor("echo '{\"name\":\"jp\"}' |"),
			cmdColor("jp"),
			flagColor("-p '$.name'"),
		)
		fmt.Fprintf(os.Stderr, "%s %s\n",
			descColor("Version:"),
			versionColor("v1.0.2"),
		)
	}

	// 解析命令行参数
	flag.StringVar(&cfg.path, "p", "", "JSONPath expression")
	flag.StringVar(&cfg.file, "f", "", "JSON file path")
	flag.BoolVar(&cfg.compact, "c", false, "Compact output")
	flag.BoolVar(&cfg.noColor, "no-color", false, "Disable colored output")
	version := flag.Bool("v", false, "Show version")
	flag.Parse()

	// 显示版本信息
	if *version {
		fmt.Printf("%s %s\n",
			descColor("jp version"),
			versionColor("v1.0.2"),
		)
		os.Exit(0)
	}

	// 处理特殊命令
	handleSpecialCommands(cfg)

	// 读取输入
	data, err := readInput(cfg.file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// 如果指定了 JSONPath 表达式，执行查询
	if cfg.path != "" {
		// 编译 JSONPath 表达式
		jp, err := jsonpath.Compile(cfg.path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", errorColor("error compiling JSONPath"), err)
			os.Exit(1)
		}

		// 执行查询
		result, err := jp.Execute(data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", errorColor("error executing query"), err)
			os.Exit(1)
		}

		data = result
	}

	// 输出结果
	if err := outputResult(data, cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// 处理特殊命令
func handleSpecialCommands(cfg *config) {
	// 如果没有参数，显示帮助
	if len(os.Args) == 1 {
		flag.Usage()
		os.Exit(0)
	}
}

// 读取输入
func readInput(filePath string) (interface{}, error) {
	var reader io.Reader

	if filePath != "" {
		// 从文件读取
		file, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", errorColor("error opening file"), err)
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
		return nil, fmt.Errorf("%s: %w", errorColor("error parsing JSON"), err)
	}

	return data, nil
}

// 输出结果
func outputResult(result interface{}, cfg *config) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", cfg.indent)
	enc.SetEscapeHTML(false)

	if err := enc.Encode(result); err != nil {
		return fmt.Errorf("%s: %w", errorColor("error encoding JSON"), err)
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
