package segments

import (
	"testing"
)

// mockSegment 是用于测试的 Segment 接口实现
type mockSegment struct {
	evaluateResult []interface{}
	evaluateError  error
	stringResult   string
}

func (m *mockSegment) Evaluate(value interface{}) ([]interface{}, error) {
	return m.evaluateResult, m.evaluateError
}

func (m *mockSegment) String() string {
	return m.stringResult
}

func TestSegmentInterface(t *testing.T) {
	tests := []struct {
		name           string
		segment        Segment
		evaluateInput  interface{}
		evaluateResult []interface{}
		evaluateError  error
		stringResult   string
	}{
		{
			name: "successful evaluation",
			segment: &mockSegment{
				evaluateResult: []interface{}{"test"},
				stringResult:   "mock",
			},
			evaluateInput:  "input",
			evaluateResult: []interface{}{"test"},
			stringResult:   "mock",
		},
		{
			name: "evaluation with error",
			segment: &mockSegment{
				evaluateError: nil,
				stringResult:  "error",
			},
			evaluateInput:  "input",
			evaluateResult: nil,
			stringResult:   "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 测试 Evaluate 方法
			result, err := tt.segment.Evaluate(tt.evaluateInput)
			if err != tt.evaluateError {
				t.Errorf("Evaluate() error = %v, want %v", err, tt.evaluateError)
			}
			if len(result) != len(tt.evaluateResult) {
				t.Errorf("Evaluate() result length = %v, want %v", len(result), len(tt.evaluateResult))
			}

			// 测试 String 方法
			if got := tt.segment.String(); got != tt.stringResult {
				t.Errorf("String() = %v, want %v", got, tt.stringResult)
			}
		})
	}
}
