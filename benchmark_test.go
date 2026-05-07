package jsonpath

import "testing"

func BenchmarkQuery(b *testing.B) {
	data := map[string]interface{}{
		"store": map[string]interface{}{
			"book": []interface{}{
				map[string]interface{}{"title": "Book 1", "price": 10.99},
				map[string]interface{}{"title": "Book 2", "price": 15.99},
			},
		},
	}
	path := "$.store.book[?@.price > 10]"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Query(data, path)
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkQuerySimplePath(b *testing.B) {
	data := map[string]interface{}{
		"store": map[string]interface{}{
			"book": []interface{}{
				map[string]interface{}{"title": "Book 1", "price": 10.99},
				map[string]interface{}{"title": "Book 2", "price": 15.99},
			},
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Query(data, "$.store.book[*].title")
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkQueryRecursiveDescent(b *testing.B) {
	data := map[string]interface{}{
		"store": map[string]interface{}{
			"book": []interface{}{
				map[string]interface{}{"title": "Book 1", "price": 10.99},
				map[string]interface{}{"title": "Book 2", "price": 15.99},
			},
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Query(data, "$..price")
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkQuerySlice(b *testing.B) {
	data := map[string]interface{}{
		"store": map[string]interface{}{
			"book": []interface{}{
				map[string]interface{}{"title": "Book 1", "price": 10.99},
				map[string]interface{}{"title": "Book 2", "price": 15.99},
				map[string]interface{}{"title": "Book 3", "price": 20.99},
			},
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Query(data, "$.store.book[0:2]")
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkQueryComplexFilter(b *testing.B) {
	data := map[string]interface{}{
		"store": map[string]interface{}{
			"book": []interface{}{
				map[string]interface{}{"title": "Book 1", "price": 10.99, "category": "fiction"},
				map[string]interface{}{"title": "Book 2", "price": 15.99, "category": "reference"},
				map[string]interface{}{"title": "Book 3", "price": 20.99, "category": "fiction"},
			},
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Query(data, "$.store.book[?@.price > 10 && @.category == 'fiction']")
		if err != nil {
			b.Fatalf("Unexpected error: %v", err)
		}
	}
}
