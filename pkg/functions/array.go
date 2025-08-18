package functions

import (
	"fmt"

	"github.com/davidhoo/jsonpath/pkg/errors"
)

// LengthFunction 实现length函数
type LengthFunction struct{}

func (f *LengthFunction) Name() string {
	return "length"
}

func (f *LengthFunction) Call(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, errors.NewError(errors.ErrInvalidArgument, "length function requires exactly one argument", fmt.Sprintf("%v", args))
	}

	switch v := args[0].(type) {
	case []interface{}:
		return float64(len(v)), nil
	case string:
		return float64(len(v)), nil
	default:
		return nil, errors.NewError(errors.ErrInvalidArgument, "length function requires array or string argument", fmt.Sprintf("%v", v))
	}
}

// FirstFunction 实现first函数
type FirstFunction struct{}

func (f *FirstFunction) Name() string {
	return "first"
}

func (f *FirstFunction) Call(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, errors.NewError(errors.ErrInvalidArgument, "first function requires exactly one argument", fmt.Sprintf("%v", args))
	}

	arr, ok := args[0].([]interface{})
	if !ok {
		return nil, errors.NewError(errors.ErrInvalidArgument, "first function requires array argument", fmt.Sprintf("%v", args[0]))
	}

	if len(arr) == 0 {
		return nil, nil
	}

	return arr[0], nil
}

// LastFunction 实现last函数
type LastFunction struct{}

func (f *LastFunction) Name() string {
	return "last"
}

func (f *LastFunction) Call(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, errors.NewError(errors.ErrInvalidArgument, "last function requires exactly one argument", fmt.Sprintf("%v", args))
	}

	arr, ok := args[0].([]interface{})
	if !ok {
		return nil, errors.NewError(errors.ErrInvalidArgument, "last function requires array argument", fmt.Sprintf("%v", args[0]))
	}

	if len(arr) == 0 {
		return nil, nil
	}

	return arr[len(arr)-1], nil
}

func init() {
	Register("length", &LengthFunction{})
	Register("first", &FirstFunction{})
	Register("last", &LastFunction{})
}
