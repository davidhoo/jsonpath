package functions

import (
	"fmt"
	"strconv"

	"github.com/davidhoo/jsonpath/pkg/errors"
)

// ToStringFunction 实现转换为字符串函数
type ToStringFunction struct{}

func (f *ToStringFunction) Name() string {
	return "toString"
}

func (f *ToStringFunction) Call(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, errors.NewError(errors.ErrInvalidArgument, "toString function requires exactly one argument", fmt.Sprintf("%v", args))
	}

	switch v := args[0].(type) {
	case string:
		return v, nil
	case float64:
		return strconv.FormatFloat(v, 'g', -1, 64), nil
	case int:
		return strconv.Itoa(v), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case bool:
		return strconv.FormatBool(v), nil
	case nil:
		return "null", nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

// ToNumberFunction 实现转换为数字函数
type ToNumberFunction struct{}

func (f *ToNumberFunction) Name() string {
	return "toNumber"
}

func (f *ToNumberFunction) Call(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, errors.NewError(errors.ErrInvalidArgument, "toNumber function requires exactly one argument", fmt.Sprintf("%v", args))
	}

	switch v := args[0].(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case string:
		if num, err := strconv.ParseFloat(v, 64); err == nil {
			return num, nil
		}
		return nil, errors.NewError(errors.ErrInvalidArgument, "toNumber function requires valid numeric string", v)
	case bool:
		if v {
			return 1.0, nil
		}
		return 0.0, nil
	case nil:
		return 0.0, nil
	default:
		return nil, errors.NewError(errors.ErrInvalidArgument, "toNumber function requires string or numeric argument", fmt.Sprintf("%v", v))
	}
}

// ToBooleanFunction 实现转换为布尔值函数
type ToBooleanFunction struct{}

func (f *ToBooleanFunction) Name() string {
	return "toBoolean"
}

func (f *ToBooleanFunction) Call(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, errors.NewError(errors.ErrInvalidArgument, "toBoolean function requires exactly one argument", fmt.Sprintf("%v", args))
	}

	switch v := args[0].(type) {
	case bool:
		return v, nil
	case string:
		if v == "true" {
			return true, nil
		}
		if v == "false" {
			return false, nil
		}
		return nil, errors.NewError(errors.ErrInvalidArgument, "toBoolean function requires 'true' or 'false' string", v)
	case float64:
		return v != 0, nil
	case int:
		return v != 0, nil
	case int64:
		return v != 0, nil
	case nil:
		return false, nil
	default:
		return nil, errors.NewError(errors.ErrInvalidArgument, "toBoolean function requires string or numeric argument", fmt.Sprintf("%v", v))
	}
}

// IsStringFunction 实现判断是否为字符串函数
type IsStringFunction struct{}

func (f *IsStringFunction) Name() string {
	return "isString"
}

func (f *IsStringFunction) Call(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, errors.NewError(errors.ErrInvalidArgument, "isString function requires exactly one argument", fmt.Sprintf("%v", args))
	}

	_, ok := args[0].(string)
	return ok, nil
}

// IsNumberFunction 实现判断是否为数字函数
type IsNumberFunction struct{}

func (f *IsNumberFunction) Name() string {
	return "isNumber"
}

func (f *IsNumberFunction) Call(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, errors.NewError(errors.ErrInvalidArgument, "isNumber function requires exactly one argument", fmt.Sprintf("%v", args))
	}

	switch args[0].(type) {
	case float64, int, int64:
		return true, nil
	default:
		return false, nil
	}
}

// IsBooleanFunction 实现判断是否为布尔值函数
type IsBooleanFunction struct{}

func (f *IsBooleanFunction) Name() string {
	return "isBoolean"
}

func (f *IsBooleanFunction) Call(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, errors.NewError(errors.ErrInvalidArgument, "isBoolean function requires exactly one argument", fmt.Sprintf("%v", args))
	}

	_, ok := args[0].(bool)
	return ok, nil
}

// IsNullFunction 实现判断是否为空函数
type IsNullFunction struct{}

func (f *IsNullFunction) Name() string {
	return "isNull"
}

func (f *IsNullFunction) Call(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, errors.NewError(errors.ErrInvalidArgument, "isNull function requires exactly one argument", fmt.Sprintf("%v", args))
	}

	return args[0] == nil, nil
}

func init() {
	Register("toString", &ToStringFunction{})
	Register("toNumber", &ToNumberFunction{})
	Register("toBoolean", &ToBooleanFunction{})
	Register("isString", &IsStringFunction{})
	Register("isNumber", &IsNumberFunction{})
	Register("isBoolean", &IsBooleanFunction{})
	Register("isNull", &IsNullFunction{})
}
