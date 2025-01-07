package jsonpath

import "testing"

func TestFilterCondition(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		operator string
		value    interface{}
	}{
		{
			name:     "simple field condition",
			field:    "name",
			operator: "==",
			value:    "John",
		},
		{
			name:     "nested field condition",
			field:    "user.name",
			operator: "!=",
			value:    "Jane",
		},
		{
			name:     "numeric comparison",
			field:    "age",
			operator: ">",
			value:    18,
		},
		{
			name:     "boolean condition",
			field:    "active",
			operator: "==",
			value:    true,
		},
		{
			name:     "null comparison",
			field:    "optional",
			operator: "==",
			value:    nil,
		},
		{
			name:     "array index condition",
			field:    "items[0].id",
			operator: "<",
			value:    100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := &filterCondition{
				field:    tt.field,
				operator: tt.operator,
				value:    tt.value,
			}

			if fc.field != tt.field {
				t.Errorf("filterCondition.field = %v, want %v", fc.field, tt.field)
			}
			if fc.operator != tt.operator {
				t.Errorf("filterCondition.operator = %v, want %v", fc.operator, tt.operator)
			}
			if fc.value != tt.value {
				t.Errorf("filterCondition.value = %v, want %v", fc.value, tt.value)
			}
		})
	}
}
