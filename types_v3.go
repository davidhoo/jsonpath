package jsonpath

import "encoding/json"

// Node represents RFC 9535 node (location + value)
type Node struct {
	Location string      `json:"location"`
	Value    interface{} `json:"value"`
	Root     interface{} `json:"-"` // document root, not serialized
}

// NodeList represents a list of nodes
type NodeList []Node

// MarshalJSON implements json.Marshaler
func (nl NodeList) MarshalJSON() ([]byte, error) {
	return json.Marshal([]Node(nl))
}

// Nothing represents Nothing value (different from null)
type Nothing struct{}

func (n Nothing) String() string {
	return "Nothing"
}

// LogicalType represents three-valued logic
type LogicalType int8

const (
	LogicalNothing LogicalType = iota
	LogicalFalse
	LogicalTrue
)

func (lt LogicalType) String() string {
	switch lt {
	case LogicalNothing:
		return "nothing"
	case LogicalFalse:
		return "false"
	case LogicalTrue:
		return "true"
	default:
		return "unknown"
	}
}
