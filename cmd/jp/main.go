package main

import (
	"bytes"
	"encoding/json"
	"errors"
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
	errHelp    = errors.New("help requested")
	errVersion = errors.New("version requested")
)

var (
	help    bool
	version bool
	cfg     config
)

// config holds command line configuration
type config struct {
	path    string
	file    string
	compact bool
	noColor bool
	indent  string
}

// printHelp prints the help message
func printHelp() {
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
		flagColor("-p '$.store.book[?@.price > 10]'"),
	)
	fmt.Fprintf(os.Stderr, "  %s\n", exampleColor("# Filter with multiple conditions"))
	fmt.Fprintf(os.Stderr, "  %s %s %s\n\n",
		cmdColor("jp"),
		flagColor("-f data.json"),
		flagColor("-p '$.store.book[?@.price > 10 && @.category == \"fiction\"]'"),
	)
	fmt.Fprintf(os.Stderr, "  %s\n", exampleColor("# Filter with logical OR"))
	fmt.Fprintf(os.Stderr, "  %s %s %s\n\n",
		cmdColor("jp"),
		flagColor("-f data.json"),
		flagColor("-p '$.store.book[?@.price < 10 || @.category == \"fiction\"]'"),
	)
	fmt.Fprintf(os.Stderr, "  %s\n", exampleColor("# Filter with NOT"))
	fmt.Fprintf(os.Stderr, "  %s %s %s\n\n",
		cmdColor("jp"),
		flagColor("-f data.json"),
		flagColor("-p '$.store.book[?!(@.category == \"reference\")]'"),
	)
	fmt.Fprintf(os.Stderr, "  %s\n", exampleColor("# Filter with complex conditions"))
	fmt.Fprintf(os.Stderr, "  %s %s %s\n\n",
		cmdColor("jp"),
		flagColor("-f data.json"),
		flagColor("-p '$.store.book[?(@.price > 10 && (@.category == \"fiction\" || @.author == \"Evelyn Waugh\"))]'"),
	)
	fmt.Fprintf(os.Stderr, "  %s\n", exampleColor("# Use length function"))
	fmt.Fprintf(os.Stderr, "  %s %s %s\n\n",
		cmdColor("jp"),
		flagColor("-f data.json"),
		flagColor("-p '$.store.book.length()'"),
	)
	fmt.Fprintf(os.Stderr, "  %s\n", exampleColor("# Get object keys"))
	fmt.Fprintf(os.Stderr, "  %s %s %s\n\n",
		cmdColor("jp"),
		flagColor("-f data.json"),
		flagColor("-p '$.store.keys()'"),
	)
	fmt.Fprintf(os.Stderr, "  %s\n", exampleColor("# Get object values"))
	fmt.Fprintf(os.Stderr, "  %s %s %s\n\n",
		cmdColor("jp"),
		flagColor("-f data.json"),
		flagColor("-p '$.store.values()'"),
	)
	fmt.Fprintf(os.Stderr, "  %s\n", exampleColor("# Get minimum price"))
	fmt.Fprintf(os.Stderr, "  %s %s %s\n\n",
		cmdColor("jp"),
		flagColor("-f data.json"),
		flagColor("-p '$.store.book[*].price.min()'"),
	)
	fmt.Fprintf(os.Stderr, "  %s\n", exampleColor("# Count occurrences of a value"))
	fmt.Fprintf(os.Stderr, "  %s %s %s\n\n",
		cmdColor("jp"),
		flagColor("-f data.json"),
		flagColor("-p '$.store.book[*].category.count(\"fiction\")'"),
	)
	fmt.Fprintf(os.Stderr, "  %s\n", exampleColor("# Read from stdin"))
	fmt.Fprintf(os.Stderr, "  %s %s %s\n\n",
		exampleColor("echo '{\"name\":\"jp\"}' |"),
		cmdColor("jp"),
		flagColor("-p '$.name'"),
	)
	fmt.Fprintf(os.Stderr, "%s\n", descColor("Functions:"))
	fmt.Fprintf(os.Stderr, "  %s  %s\n",
		flagColor("length()"),
		descColor("Returns the length of a string, array, or object"),
	)
	fmt.Fprintf(os.Stderr, "  %s  %s\n",
		flagColor("keys()"),
		descColor("Returns an array of the object's property names in alphabetical order"),
	)
	fmt.Fprintf(os.Stderr, "  %s  %s\n",
		flagColor("values()"),
		descColor("Returns an array of the object's property values in key order"),
	)
	fmt.Fprintf(os.Stderr, "  %s  %s\n",
		flagColor("min()"),
		descColor("Returns the minimum value in a numeric array (ignores non-numeric values)"),
	)
	fmt.Fprintf(os.Stderr, "  %s  %s\n",
		flagColor("max()"),
		descColor("Returns the maximum value in a numeric array (ignores non-numeric values)"),
	)
	fmt.Fprintf(os.Stderr, "  %s  %s\n",
		flagColor("avg()"),
		descColor("Returns the average value in a numeric array (ignores non-numeric values)"),
	)
	fmt.Fprintf(os.Stderr, "  %s  %s\n",
		flagColor("count(value)"),
		descColor("Returns the number of occurrences of a value in an array"),
	)
	fmt.Fprintf(os.Stderr, "\n%s\n", descColor("Filter Syntax:"))
	fmt.Fprintf(os.Stderr, "  %s  %s\n",
		flagColor("[?@.field > value]"),
		descColor("Filter by comparison (>, <, >=, <=, ==, !=)"),
	)
	fmt.Fprintf(os.Stderr, "  %s  %s\n",
		flagColor("[?@.field1 == value1 && @.field2 != value2]"),
		descColor("Filter with logical AND (&&)"),
	)
	fmt.Fprintf(os.Stderr, "  %s  %s\n",
		flagColor("[?@.field1 == value1 || @.field2 == value2]"),
		descColor("Filter with logical OR (||)"),
	)
	fmt.Fprintf(os.Stderr, "  %s  %s\n",
		flagColor("[?!(@.field == value)]"),
		descColor("Filter with logical NOT (!)"),
	)
	fmt.Fprintf(os.Stderr, "  %s  %s\n",
		flagColor("[?!(@.field1 == value1 && @.field2 == value2)]"),
		descColor("Complex NOT expression (applies De Morgan's Law)"),
	)
	fmt.Fprintf(os.Stderr, "  %s  %s\n",
		flagColor("[?(@.field1 > 10 && (@.field2 == 'value' || @.field3 != null))]"),
		descColor("Nested conditions with parentheses"),
	)
	fmt.Fprintf(os.Stderr, "\n%s\n", descColor("Filter Examples:"))
	fmt.Fprintf(os.Stderr, "  %s  %s\n",
		flagColor("[?@.price > 10]"),
		descColor("$.store.book[?@.price > 10]                # Books over $10"),
	)
	fmt.Fprintf(os.Stderr, "  %s  %s\n",
		flagColor("[?@.category == 'fiction']"),
		descColor("$.store.book[?@.category == 'fiction']     # Fiction books"),
	)
	fmt.Fprintf(os.Stderr, "  %s  %s\n",
		flagColor("[?@.price > 10 && @.category == 'fiction']"),
		descColor("# Fiction books over $10"),
	)
	fmt.Fprintf(os.Stderr, "  %s  %s\n",
		flagColor("[?!(@.category == 'reference')]"),
		descColor("# Books that are not reference books"),
	)
	fmt.Fprintf(os.Stderr, "  %s  %s\n",
		flagColor("[?@.price < 10 || @.category == 'fiction']"),
		descColor("# Books under $10 or fiction books"),
	)
	fmt.Fprintf(os.Stderr, "\n%s\n", descColor("Function Examples:"))
	fmt.Fprintf(os.Stderr, "  %s  %s\n",
		flagColor("min()"),
		descColor("$.numbers.min()                    # [3, 1, 4] -> 1"),
	)
	fmt.Fprintf(os.Stderr, "  %s  %s\n",
		flagColor("min()"),
		descColor("$.mixed.min()                      # [3, \"a\", 1, null] -> 1"),
	)
	fmt.Fprintf(os.Stderr, "  %s  %s\n",
		flagColor("min()"),
		descColor("$.store.book[*].price.min()        # Get minimum price"),
	)
	fmt.Fprintf(os.Stderr, "  %s  %s\n",
		flagColor("max()"),
		descColor("$.numbers.max()                    # [3, 1, 4] -> 4"),
	)
	fmt.Fprintf(os.Stderr, "  %s  %s\n",
		flagColor("max()"),
		descColor("$.mixed.max()                      # [3, \"a\", 1, null] -> 3"),
	)
	fmt.Fprintf(os.Stderr, "  %s  %s\n",
		flagColor("max()"),
		descColor("$.store.book[*].price.max()        # Get maximum price"),
	)
	fmt.Fprintf(os.Stderr, "  %s  %s\n",
		flagColor("avg()"),
		descColor("$.numbers.avg()                    # [2, 4, 6] -> 4"),
	)
	fmt.Fprintf(os.Stderr, "  %s  %s\n",
		flagColor("avg()"),
		descColor("$.mixed.avg()                      # [3, \"a\", 1, null] -> 2"),
	)
	fmt.Fprintf(os.Stderr, "  %s  %s\n",
		flagColor("avg()"),
		descColor("$.store.book[*].price.avg()        # Get average price"),
	)
	fmt.Fprintf(os.Stderr, "  %s  %s\n",
		flagColor("count(number)"),
		descColor("$.numbers.count(2)"),
	)
	fmt.Fprintf(os.Stderr, "  %s  %s\n",
		flagColor("count(string)"),
		descColor("$.tags.count(\"a\")"),
	)
	fmt.Fprintf(os.Stderr, "  %s  %s\n",
		flagColor("count(object)"),
		descColor("$.items.count({\"id\": 1})"),
	)
	fmt.Fprintf(os.Stderr, "\n%s %s\n",
		descColor("Version:"),
		versionColor(jsonpath.VersionWithPrefix()),
	)
}

// printVersion prints the version information
func printVersion() {
	fmt.Printf("%s %s\n",
		descColor("jp version"),
		versionColor(jsonpath.VersionWithPrefix()),
	)
}

// 颜色定义
var (
	keyColor        = color.New(color.FgCyan).SprintFunc()
	keyQuoteColor   = color.New(color.FgCyan).SprintFunc()
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

// ParseFlags parses command line flags and returns path and file
func ParseFlags() (string, string, error) {
	flagSet := flag.NewFlagSet("jp", flag.ContinueOnError)
	flagSet.StringVar(&cfg.path, "p", "", "JSONPath expression")
	flagSet.StringVar(&cfg.file, "f", "", "JSON file path")
	flagSet.BoolVar(&cfg.compact, "c", false, "Compact output")
	flagSet.BoolVar(&cfg.noColor, "no-color", false, "Disable colored output")
	flagSet.BoolVar(&help, "h", false, "Show help")
	flagSet.BoolVar(&version, "v", false, "Show version")

	if err := flagSet.Parse(os.Args[1:]); err != nil {
		return "", "", err
	}

	if err := handleSpecialCommands(); err != nil {
		return "", "", err
	}

	return cfg.path, cfg.file, nil
}

// handleSpecialCommands handles special commands like -h and -v
func handleSpecialCommands() error {
	if help {
		printHelp()
		return errHelp
	}
	if version {
		printVersion()
		return errVersion
	}
	return nil
}

// readInput reads JSON input from file or stdin
func readInput(file string) (string, error) {
	var input []byte
	var err error

	if file == "" {
		input, err = io.ReadAll(os.Stdin)
	} else {
		input, err = os.ReadFile(file)
	}
	if err != nil {
		return "", fmt.Errorf("%s: %v", errorColor("error reading input"), err)
	}

	// Validate JSON
	var data interface{}
	if err := json.Unmarshal(input, &data); err != nil {
		return "", fmt.Errorf("%s: %v", errorColor("invalid JSON"), err)
	}

	return string(input), nil
}

// 输出结果
func outputResult(result interface{}, cfg *config) error {
	// 如果结果是字符串，直接输出
	if str, ok := result.(string); ok {
		if cfg.noColor {
			fmt.Println(str)
		} else {
			fmt.Println(stringColor(str))
		}
		return nil
	}

	var output string
	if cfg.compact {
		// 压缩输出
		jsonBytes, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("%s: %v", errorColor("error encoding JSON"), err)
		}
		output = string(jsonBytes)
	} else {
		// 格式化输出
		jsonBytes, err := json.MarshalIndent(result, "", cfg.indent)
		if err != nil {
			return fmt.Errorf("%s: %v", errorColor("error encoding JSON"), err)
		}
		output = string(jsonBytes)
	}

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
						result.WriteString(keyQuoteColor(`"`))
					} else {
						result.WriteString(valueQuoteColor(`"`))
					}
				} else {
					// 结束一个字符串
					inString = false
					if inKey {
						result.WriteString(keyQuoteColor(`"`))
						inKey = false
					} else {
						result.WriteString(valueQuoteColor(`"`))
					}
				}
			} else {
				result.WriteRune(r)
			}

		case '{', '}':
			if !inString {
				result.WriteString(braceColor(string(r)))
				if r == '{' {
					inKey = true
				}
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
				inKey = true
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
	}

	return result.String()
}

// FormatJSON formats JSON data with optional compression
func FormatJSON(data interface{}, compress bool) string {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if !compress {
		enc.SetIndent("", "  ")
	}
	enc.SetEscapeHTML(false)

	if err := enc.Encode(data); err != nil {
		return fmt.Sprintf("%s: %v", errorColor("error encoding JSON"), err)
	}

	return strings.TrimSpace(buf.String())
}

func main() {
	var err error
	cfg.path, cfg.file, err = ParseFlags()
	if err == errHelp || err == errVersion {
		os.Exit(0)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// 读取输入
	data, err := readInput(cfg.file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// 如果 JSONPath 表达式被指定，执行查询
	if cfg.path != "" {
		// 执行 JSONPath 查询
		result, err := jsonpath.Query(data, cfg.path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", errorColor("error executing query"), err)
			os.Exit(1)
		}

		// 如果结果是字符串，直接输出
		if str, ok := result.(string); ok {
			if cfg.noColor {
				fmt.Println(str)
			} else {
				fmt.Println(stringColor(str))
			}
			return
		}

		// 设置缩进
		if !cfg.compact {
			cfg.indent = "  "
		}

		// 输出结果
		if err := outputResult(result, &cfg); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	} else {
		// 解析原始 JSON 数据
		var jsonData interface{}
		if err := json.Unmarshal([]byte(data), &jsonData); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", errorColor("error parsing JSON"), err)
			os.Exit(1)
		}

		// 设置缩进
		if !cfg.compact {
			cfg.indent = "  "
		}

		// 输出结果
		if err := outputResult(jsonData, &cfg); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}
