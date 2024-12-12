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

var (
	// 定义颜色输出
	titleStyle   = color.New(color.FgHiCyan, color.Bold)
	errorStyle   = color.New(color.FgHiRed, color.Bold)
	successStyle = color.New(color.FgHiGreen)
	exampleStyle = color.New(color.FgYellow)
	optionStyle  = color.New(color.FgHiMagenta)
)

const version = "1.0.0"

func printUsage() {
	titleStyle.Printf("jp - JSONPath Command Line Tool v%s\n", version)
	fmt.Println("\nUsage:")
	optionStyle.Println("  jp -p <jsonpath_expression> [-f <json_file>] [-c] [-h] [-v]")
	fmt.Println("  jp -f <json_file> [-c]     # Output entire JSON if -p is not provided")

	fmt.Println("\nOptions:")
	optionStyle.Println("  -f  JSON file path (if not specified, read from stdin)")
	optionStyle.Println("  -p  JSONPath expression (optional, outputs entire JSON if not provided)")
	optionStyle.Println("  -c  Compact output (no pretty print)")
	optionStyle.Println("  -h  Show this help")
	optionStyle.Println("  -v  Show version")

	fmt.Println("\nExamples:")
	exampleStyle.Println("  # Query from file")
	exampleStyle.Println("  jp -f data.json -p '$.store.book[*].author'")
	exampleStyle.Println("\n  # Output entire JSON")
	exampleStyle.Println("  jp -f data.json")
	exampleStyle.Println("\n  # Query from stdin")
	exampleStyle.Println("  echo '{\"name\": \"John\"}' | jp -p '$.name'")
	exampleStyle.Println("\n  # Compact output")
	exampleStyle.Println("  jp -f data.json -p '$.store.book[0]' -c")

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

func formatJSON(data interface{}, compact bool) (string, error) {
	var output []byte
	var err error

	if compact {
		output, err = json.Marshal(data)
	} else {
		output, err = json.MarshalIndent(data, "", "  ")
	}

	if err != nil {
		return "", err
	}

	// 为不同类型的值添加不同的颜色
	var colored string
	switch v := data.(type) {
	case []interface{}:
		colored = successStyle.Sprintf("%s", string(output))
	case string:
		colored = color.HiGreenString("%q", v)
	case float64:
		colored = color.HiCyanString("%g", v)
	case bool:
		colored = color.HiMagentaString("%v", v)
	case nil:
		colored = color.HiRedString("null")
	default:
		colored = string(output)
	}

	return colored, nil
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
		output, err := formatJSON(data, *compact)
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
	output, err := formatJSON(result, *compact)
	if err != nil {
		printError("formatting output: %v", err)
		os.Exit(1)
	}

	fmt.Println(output)
}
