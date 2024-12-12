package jsonpath_test

import (
	"encoding/json"
	"testing"

	"github.com/davidhoo/jsonpath"
)

func TestJSONPath(t *testing.T) {
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
			]
		}
	}`

	var data interface{}
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		path string
		want interface{}
	}{
		{"$.store.book[0].author", "Nigel Rees"},
		{"$.store.book[*].author", []interface{}{"Nigel Rees", "Evelyn Waugh"}},
		{"$.store.book[?(@.price < 10)].title", []interface{}{"Sayings of the Century"}},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			jp, err := jsonpath.Compile(tt.path)
			if err != nil {
				t.Fatal(err)
			}

			result, err := jp.Execute(data)
			if err != nil {
				t.Fatal(err)
			}

			resultJSON, _ := json.Marshal(result)
			wantJSON, _ := json.Marshal(tt.want)
			if string(resultJSON) != string(wantJSON) {
				t.Errorf("got %v, want %v", string(resultJSON), string(wantJSON))
			}
		})
	}
}

func TestJSONPathExtended(t *testing.T) {
	jsonData := `{
		"store": {
			"book": [
				{
					"category": "reference",
					"author": "Nigel Rees",
					"title": "Sayings of the Century",
					"price": 8.95,
					"reviews": [
						{"rating": 4, "text": "Good book"},
						{"rating": 5, "text": "Excellent"}
					]
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

	var data interface{}
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		path string
		want interface{}
	}{
		{"$..price", []interface{}{8.95, 12.99, 19.95}},
		{"$..rating", []interface{}{4.0, 5.0}},
		{"$.store.book[0,1].author", []interface{}{"Nigel Rees", "Evelyn Waugh"}},
		{"$.store.book[0:2].title", []interface{}{"Sayings of the Century", "Sword of Honour"}},
		{"$.store.book[:1].title", []interface{}{"Sayings of the Century"}},
		{"$.store.book[::2].title", []interface{}{"Sayings of the Century"}},
		{"$.store.book[::-1].title", []interface{}{"Sword of Honour", "Sayings of the Century"}},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			jp, err := jsonpath.Compile(tt.path)
			if err != nil {
				t.Fatal(err)
			}

			result, err := jp.Execute(data)
			if err != nil {
				t.Fatal(err)
			}

			resultJSON, _ := json.Marshal(result)
			wantJSON, _ := json.Marshal(tt.want)
			if string(resultJSON) != string(wantJSON) {
				t.Errorf("got %v, want %v", string(resultJSON), string(wantJSON))
			}
		})
	}
}
