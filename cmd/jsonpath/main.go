package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"

	"github.com/davidhoo/jsonpath"
	"github.com/fatih/color"
)

var (
	// 定义颜色输出
	titleStyle   = color.New(color.FgHiCyan, color.Bold)
	errorStyle   = color.New(color.FgHiRed, color.Bold)
	successStyle = color.New(color.FgHiGreen)
	exampleStyle = color.New(color.FgYellow)
	optionStyle  = color.New(color.FgHiMagenta)

	// JSON 元素颜色
	braceStyle   = color.New(color.FgMagenta)   // {} 大括号，使用紫色
	bracketStyle = color.New(color.FgYellow)    // [] 方括号，使用黄色
	keyStyle     = color.New(color.FgHiMagenta) // 键名，使用亮紫色
	stringStyle  = color.New(color.FgHiGreen)   // 字符串值，使用亮绿色
	numberStyle  = color.New(color.FgHiBlue)    // 数字值，使用亮蓝色
	boolStyle    = color.New(color.FgHiYellow)  // 布尔值，使用亮黄色
	nullStyle    = color.New(color.FgHiRed)     // null 值，使用亮红色
	commaStyle   = color.New(color.FgRed)       // 逗号，使用红色
	colonStyle   = color.New(color.FgWhite)     // 冒号，使用白色

	// 字符串引号颜色
	keyQuoteStyle   = color.New(color.FgMagenta) // 键名的引号，使用紫色
	valueQuoteStyle = color.New(color.FgGreen)   // 值的引号，使用绿色
)

const version = "1.0.0"

func printUsage() {
	titleStyle.Printf("jp - JSONPath Command Line Tool v%s\n", version)
	fmt.Println("\nUsage:")
	optionStyle.Println("  jp [-p <jsonpath_expression>] [-f <json_file>] [-c] [-h] [-v]")

	fmt.Println("\nOptions:")
	optionStyle.Println("  -f  JSON file path (if not specified, read from stdin)")
	optionStyle.Println("  -p  JSONPath expression (if not specified, output entire JSON)")
	optionStyle.Println("  -c  Compact output (no pretty print)")
	optionStyle.Println("  -h  Show this help")
	optionStyle.Println("  -v  Show version")

	fmt.Println("\nExamples:")
	exampleStyle.Println("  # Output entire JSON")
	exampleStyle.Println("  jp -f data.json")
	exampleStyle.Println("\n  # Query specific path")
	exampleStyle.Println("  jp -f data.json -p '$.store.book[*].author'")
	exampleStyle.Println("\n  # Query from stdin")
	exampleStyle.Println("  echo '{\"name\": \"John\"}' | jp -p '$.name'")
	exampleStyle.Println("\n  # Compact output")
	exampleStyle.Println("  jp -f data.json -c")

	fmt.Println("\nSupported JSONPath Features:")
	successStyle.Println("  - Root node:          $")
	successStyle.Println("  - Child node:         .key or ['key']")
	successStyle.Println("  - Recursive descent:   ..")
	successStyle.Println("  - Array index:        [0]")
	successStyle.Println("  - Array slice:        [start:end:step]")
	successStyle.Println("  - Array wildcard:     [*]")
	successStyle.Println("  - Multiple indexes:   [1,2,3]")
	successStyle.Println("  - Filter expression:  [?(@.price < 10)]")

	fmt.Println("\nMore info: https://github.com/davidhoo/jsonpath")
}

func printError(format string, a ...interface{}) {
	errorStyle.Fprintf(os.Stderr, "Error: "+format+"\n", a...)
}

func colorizeJSON(data interface{}, compact bool) (string, error) {
	var output []byte
	var err error

	if compact {
		output, err = json.Marshal(data)
		if err != nil {
			return "", err
		}
		return colorizeJSONString(string(output)), nil
	}

	output, err = json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}

	return colorizeJSONString(string(output)), nil
}

func colorizeJSONString(jsonStr string) string {
	var result strings.Builder
	inString := false
	inKey := false
	var prev rune
	runes := []rune(jsonStr)

	for i := 0; i < len(runes); i++ {
		c := runes[i]
		switch c {
		case '"':
			if prev != '\\' {
				if !inString {
					// 开始一个字符串
					inString = true
					// 检查是否是键名
					inKey = len(strings.TrimSpace(result.String())) > 0 && strings.HasSuffix(strings.TrimSpace(result.String()), ":")
					if inKey {
						result.WriteString(keyQuoteStyle.Sprint("\""))
					} else {
						result.WriteString(valueQuoteStyle.Sprint("\""))
					}
				} else {
					// 结束一个字符串
					inString = false
					if inKey {
						result.WriteString(keyQuoteStyle.Sprint("\""))
					} else {
						result.WriteString(valueQuoteStyle.Sprint("\""))
					}
					inKey = false
				}
			} else {
				result.WriteRune(c)
			}
		case '{', '}':
			if !inString {
				result.WriteString(braceStyle.Sprint(string(c)))
			} else {
				result.WriteRune(c)
			}
		case '[', ']':
			if !inString {
				result.WriteString(bracketStyle.Sprint(string(c)))
			} else {
				result.WriteRune(c)
			}
		case ',':
			if !inString {
				result.WriteString(commaStyle.Sprint(string(c)))
			} else {
				result.WriteRune(c)
			}
		case ':':
			if !inString {
				result.WriteString(colonStyle.Sprint(string(c)))
			} else {
				result.WriteRune(c)
			}
		default:
			if !inString {
				// 处理布尔值和 null
				rest := string(runes[i:])
				switch {
				case strings.HasPrefix(rest, "true"):
					result.WriteString(boolStyle.Sprint("true"))
					i += 3 // 跳过剩余字符
				case strings.HasPrefix(rest, "false"):
					result.WriteString(boolStyle.Sprint("false"))
					i += 4 // 跳过剩余字符
				case strings.HasPrefix(rest, "null"):
					result.WriteString(nullStyle.Sprint("null"))
					i += 3 // 跳过剩余字符
				case unicode.IsDigit(c) || c == '-' || c == '.':
					// 处理数字
					numStr := string(c)
					j := i + 1
					for j < len(runes) {
						next := runes[j]
						if unicode.IsDigit(next) || next == '.' || next == 'e' || next == 'E' || next == '-' || next == '+' {
							numStr += string(next)
							j++
						} else {
							break
						}
					}
					if _, err := fmt.Sscanf(numStr, "%f"); err == nil {
						result.WriteString(numberStyle.Sprint(numStr))
						i = j - 1 // 更新索引
					} else {
						result.WriteRune(c)
					}
				default:
					result.WriteRune(c)
				}
			} else {
				if inKey {
					result.WriteString(keyStyle.Sprint(string(c)))
				} else {
					result.WriteString(stringStyle.Sprint(string(c)))
				}
			}
		}
		prev = c
	}
	return result.String()
}

func main() {
	// 定义命令行参数
	var (
		jsonFile = flag.String("f", "", "JSON file path")
		path     = flag.String("p", "", "JSONPath expression")
		compact  = flag.Bool("c", false, "Compact output")
		help     = flag.Bool("h", false, "Show help")
		ver      = flag.Bool("v", false, "Show version")
	)

	// 解析命令行参数
	flag.Parse()

	// 显示版本信息
	if *ver {
		fmt.Printf("jp version %s\n", version)
		return
	}

	// 显示帮助信息
	if *help {
		printUsage()
		return
	}

	// 读取 JSON 数据
	var data interface{}
	if *jsonFile != "" {
		// 从文件读取
		jsonData, err := os.ReadFile(*jsonFile)
		if err != nil {
			printError("reading file: %v", err)
			os.Exit(1)
		}

		if err := json.Unmarshal(jsonData, &data); err != nil {
			printError("parsing JSON: %v", err)
			os.Exit(1)
		}
	} else {
		// 从标准输入读取
		jsonData, err := io.ReadAll(os.Stdin)
		if err != nil {
			printError("reading from stdin: %v", err)
			os.Exit(1)
		}

		if err := json.Unmarshal(jsonData, &data); err != nil {
			printError("parsing JSON: %v", err)
			os.Exit(1)
		}
	}

	// 如果没有提供 JSONPath 表达式，输出完整的 JSON
	if *path == "" {
		output, err := colorizeJSON(data, *compact)
		if err != nil {
			printError("formatting output: %v", err)
			os.Exit(1)
		}
		fmt.Println(output)
		return
	}

	// 编译并执行 JSONPath 表达式
	jp, err := jsonpath.Compile(*path)
	if err != nil {
		printError("compiling JSONPath: %v", err)
		os.Exit(1)
	}

	result, err := jp.Execute(data)
	if err != nil {
		printError("executing JSONPath: %v", err)
		os.Exit(1)
	}

	// 格式化并输出结果
	output, err := colorizeJSON(result, *compact)
	if err != nil {
		printError("formatting output: %v", err)
		os.Exit(1)
	}

	fmt.Println(output)
}
