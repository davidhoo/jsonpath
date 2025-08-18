package functions

import (
	"fmt"
	"strings"

	"github.com/davidhoo/jsonpath/pkg/errors"
)

// ConcatFunction 实现字符串连接函数
type ConcatFunction struct{}

func (f *ConcatFunction) Name() string {
	return "concat"
}

func (f *ConcatFunction) Call(args []interface{}) (interface{}, error) {
	if len(args) < 2 {
		return nil, errors.NewError(errors.ErrInvalidArgument, "concat function requires at least two arguments", fmt.Sprintf("%v", args))
	}

	var result strings.Builder
	for _, arg := range args {
		switch v := arg.(type) {
		case string:
			result.WriteString(v)
		case float64:
			result.WriteString(fmt.Sprintf("%g", v))
		case bool:
			result.WriteString(fmt.Sprintf("%v", v))
		case nil:
			result.WriteString("null")
		case int:
			result.WriteString(fmt.Sprintf("%d", v))
		case int64:
			result.WriteString(fmt.Sprintf("%d", v))
		default:
			return nil, errors.NewError(errors.ErrInvalidArgument, "concat function requires arguments that can be converted to string", fmt.Sprintf("%v", v))
		}
	}
	return result.String(), nil
}

// ContainsFunction 实现字符串包含函数
type ContainsFunction struct{}

func (f *ContainsFunction) Name() string {
	return "contains"
}

func (f *ContainsFunction) Call(args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, errors.NewError(errors.ErrInvalidArgument, "contains function requires exactly two arguments", fmt.Sprintf("%v", args))
	}

	str, ok := args[0].(string)
	if !ok {
		return nil, errors.NewError(errors.ErrInvalidArgument, "contains function requires string as first argument", fmt.Sprintf("%v", args[0]))
	}

	substr, ok := args[1].(string)
	if !ok {
		return nil, errors.NewError(errors.ErrInvalidArgument, "contains function requires string as second argument", fmt.Sprintf("%v", args[1]))
	}

	return strings.Contains(str, substr), nil
}

// StartsWithFunction 实现字符串前缀函数
type StartsWithFunction struct{}

func (f *StartsWithFunction) Name() string {
	return "startsWith"
}

func (f *StartsWithFunction) Call(args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, errors.NewError(errors.ErrInvalidArgument, "startsWith function requires exactly two arguments", fmt.Sprintf("%v", args))
	}

	str, ok := args[0].(string)
	if !ok {
		return nil, errors.NewError(errors.ErrInvalidArgument, "startsWith function requires string as first argument", fmt.Sprintf("%v", args[0]))
	}

	prefix, ok := args[1].(string)
	if !ok {
		return nil, errors.NewError(errors.ErrInvalidArgument, "startsWith function requires string as second argument", fmt.Sprintf("%v", args[1]))
	}

	return strings.HasPrefix(str, prefix), nil
}

// EndsWithFunction 实现字符串后缀函数
type EndsWithFunction struct{}

func (f *EndsWithFunction) Name() string {
	return "endsWith"
}

func (f *EndsWithFunction) Call(args []interface{}) (interface{}, error) {
	if len(args) != 2 {
		return nil, errors.NewError(errors.ErrInvalidArgument, "endsWith function requires exactly two arguments", fmt.Sprintf("%v", args))
	}

	str, ok := args[0].(string)
	if !ok {
		return nil, errors.NewError(errors.ErrInvalidArgument, "endsWith function requires string as first argument", fmt.Sprintf("%v", args[0]))
	}

	suffix, ok := args[1].(string)
	if !ok {
		return nil, errors.NewError(errors.ErrInvalidArgument, "endsWith function requires string as second argument", fmt.Sprintf("%v", args[1]))
	}

	return strings.HasSuffix(str, suffix), nil
}

func init() {
	Register("concat", &ConcatFunction{})
	Register("contains", &ContainsFunction{})
	Register("startsWith", &StartsWithFunction{})
	Register("endsWith", &EndsWithFunction{})
}
