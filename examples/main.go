package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/davidhoo/jsonpath"
)

func main() {
	// 示例 JSON 数据
	jsonData := `{
		"store": {
			"book": [
				{
					"category": "reference",
					"author": "Nigel Rees",
					"title": "Sayings of the Century",
					"price": 8.95
				},
				{
					"category": "fiction",
					"author": "Evelyn Waugh",
					"title": "Sword of Honour",
					"price": 12.99
				}
			],
			"bicycle": {
				"color": "red",
				"price": 19.95
			}
		}
	}`

	// 解析 JSON 数据
	var data interface{}
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		log.Fatal(err)
	}

	// 示例 1: 获取所有作者
	example1, err := getAuthors(data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("所有作者:", example1)

	// 示例 2: 获取所有价格
	example2, err := getAllPrices(data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("所有价格:", example2)

	// 示例 3: 获取特定价格范围的书籍
	example3, err := getBooksInPriceRange(data, 10)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("价格小于 10 的书籍:", example3)

	// 示例 4: 使用通配符和过滤器
	example4, err := getSpecificBooks(data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("特定条件的书籍:", example4)

	// 示例 5: 使用复杂过滤条件
	example5, err := getComplexFilteredBooks(data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("复杂条件过滤的书籍:", example5)

	// 示例 6: 使用否定过滤条件
	example6, err := getNonReferenceBooks(data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("非参考类书籍:", example6)
}

// 获取所有作者
func getAuthors(data interface{}) ([]string, error) {
	result, err := jsonpath.Query(data, "$.store.book[*].author")
	if err != nil {
		return nil, fmt.Errorf("查询失败: %v", err)
	}

	authors, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("结果类型不匹配")
	}

	var names []string
	for _, author := range authors {
		if name, ok := author.(string); ok {
			names = append(names, name)
		}
	}

	return names, nil
}

// 获取所有价格
func getAllPrices(data interface{}) ([]float64, error) {
	result, err := jsonpath.Query(data, "$..price")
	if err != nil {
		return nil, err
	}

	prices, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("结果类型不匹配")
	}

	var priceList []float64
	for _, price := range prices {
		if p, ok := price.(float64); ok {
			priceList = append(priceList, p)
		}
	}

	return priceList, nil
}

// 获取特定价格范围的书籍
func getBooksInPriceRange(data interface{}, maxPrice float64) ([]string, error) {
	result, err := jsonpath.Query(data, fmt.Sprintf("$.store.book[?(@.price < %v)].title", maxPrice))
	if err != nil {
		return nil, err
	}

	titles, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("结果类型不匹配")
	}

	var bookTitles []string
	for _, title := range titles {
		if t, ok := title.(string); ok {
			bookTitles = append(bookTitles, t)
		}
	}

	return bookTitles, nil
}

// 使用通配符和过滤器的查询
func getSpecificBooks(data interface{}) ([]interface{}, error) {
	result, err := jsonpath.Query(data, "$.store.book[?(@.price > 10)]")
	if err != nil {
		return nil, err
	}

	books, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("结果类型不匹配")
	}

	return books, nil
}

// 使用复杂过滤条件的查询
func getComplexFilteredBooks(data interface{}) ([]interface{}, error) {
	result, err := jsonpath.Query(data, "$.store.book[?(@.price > 10 && @.category == 'fiction')]")
	if err != nil {
		return nil, err
	}

	books, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("结果类型不匹配")
	}

	return books, nil
}

// 使用否定过滤条件的查询
func getNonReferenceBooks(data interface{}) ([]interface{}, error) {
	result, err := jsonpath.Query(data, "$.store.book[?!(@.category == 'reference')]")
	if err != nil {
		return nil, err
	}

	books, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("结果类型不匹配")
	}

	return books, nil
}
