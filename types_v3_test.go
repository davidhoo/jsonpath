package jsonpath

import (
	"encoding/json"
	"testing"
)

func TestNodeCreation(t *testing.T) {
	n := Node{Location: "$.store.book[0]", Value: map[string]interface{}{"title": "Sayings of the Century"}}
	if n.Location != "$.store.book[0]" {
		t.Errorf("expected location $.store.book[0], got %s", n.Location)
	}
	if n.Value == nil {
		t.Error("expected non-nil value")
	}
}

func TestNodeListJSON(t *testing.T) {
	nl := NodeList{
		{Location: "$[0]", Value: 1},
		{Location: "$[1]", Value: "two"},
	}
	b, err := json.Marshal(nl)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	var out []Node
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if len(out) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(out))
	}
	if out[0].Location != "$[0]" {
		t.Errorf("expected location $[0], got %s", out[0].Location)
	}
}

func TestLogicalTypeString(t *testing.T) {
	tests := []struct {
		lt   LogicalType
		want string
	}{
		{LogicalNothing, "nothing"},
		{LogicalFalse, "false"},
		{LogicalTrue, "true"},
		{LogicalType(99), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.lt.String(); got != tt.want {
			t.Errorf("LogicalType(%d).String() = %q, want %q", tt.lt, got, tt.want)
		}
	}
}

func TestNothingString(t *testing.T) {
	n := Nothing{}
	if got := n.String(); got != "Nothing" {
		t.Errorf("Nothing.String() = %q, want %q", got, "Nothing")
	}
}
