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

const version = "1.0.1"

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

// 处理字符串字符
func handleStringChar(c rune, inString bool, inKey bool, prev rune) (string, bool, bool) {
	if prev != '\\' {
		if !inString {
			// 开始一个字符串
			inString = true
			if inKey {
				return keyQuoteStyle.Sprint("\""), inString, inKey
			}
			return valueQuoteStyle.Sprint("\""), inString, inKey
		}
		// 结束一个字符串
		inString = false
		if inKey {
			return keyQuoteStyle.Sprint("\""), inString, false
		}
		return valueQuoteStyle.Sprint("\""), inString, false
	}
	return string(c), inString, inKey
}

// 处理结构字符
func handleStructureChar(c rune, inString bool) string {
	if !inString {
		switch c {
		case '{', '}':
			return braceStyle.Sprint(string(c))
		case '[', ']':
			return bracketStyle.Sprint(string(c))
		case ',':
			return commaStyle.Sprint(string(c))
		case ':':
			return colonStyle.Sprint(string(c))
		}
	}
	return string(c)
}

// 处理布尔值和 null
func handleLiteralPrefix(rest string, inString bool) (string, int) {
	if !inString {
		switch {
		case strings.HasPrefix(rest, "true"):
			return boolStyle.Sprint("true"), 3
		case strings.HasPrefix(rest, "false"):
			return boolStyle.Sprint("false"), 4
		case strings.HasPrefix(rest, "null"):
			return nullStyle.Sprint("null"), 3
		}
	}
	return "", -1
}

// 处理数字
func handleNumber(c rune, rest string, inString bool) (string, int) {
	if !inString && (unicode.IsDigit(c) || c == '-' || c == '.') {
		numStr := string(c)
		i := 0
		for j, next := range rest {
			if unicode.IsDigit(next) || next == '.' || next == 'e' || next == 'E' || next == '-' || next == '+' {
				numStr += string(next)
				i = j
			} else {
				break
			}
		}
		if _, err := fmt.Sscanf(numStr, "%f"); err == nil {
			return numberStyle.Sprint(numStr), i
		}
	}
	return "", -1
}

// 处理字符串内容
func handleStringContent(c rune, inKey bool) string {
	if inKey {
		return keyStyle.Sprint(string(c))
	}
	return stringStyle.Sprint(string(c))
}

func colorizeJSONString(jsonStr string) string {
	var result strings.Builder
	inString := false
	inKey := false
	var prev rune
	runes := []rune(jsonStr)

	for i := 0; i < len(runes); i++ {
		c := runes[i]
		rest := string(runes[i:])

		switch c {
		case '"':
			str, newInString, newInKey := handleStringChar(c, inString, inKey, prev)
			result.WriteString(str)
			inString = newInString
			inKey = newInKey

		case '{', '}', '[', ']', ',', ':':
			result.WriteString(handleStructureChar(c, inString))

		default:
			if !inString {
				// 检查布尔值和 null
				if str, skip := handleLiteralPrefix(rest, inString); skip >= 0 {
					result.WriteString(str)
					i += skip
					continue
				}

				// 检查数字
				if str, skip := handleNumber(c, rest[1:], inString); skip >= 0 {
					result.WriteString(str)
					i += skip
					continue
				}

				result.WriteRune(c)
			} else {
				result.WriteString(handleStringContent(c, inKey))
			}
		}

		// 更新前一个字符
		prev = c

		// 检查是否是键名
		if !inString && len(strings.TrimSpace(result.String())) > 0 {
			inKey = strings.HasSuffix(strings.TrimSpace(result.String()), ":")
		}
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
