package functions

import (
	"fmt"
	"math"

	"github.com/davidhoo/jsonpath/pkg/errors"
)

// AbsFunction 实现绝对值函数
type AbsFunction struct{}

func (f *AbsFunction) Name() string {
	return "abs"
}

func (f *AbsFunction) Call(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, errors.NewError(errors.ErrInvalidArgument, "abs function requires exactly one argument", fmt.Sprintf("%v", args))
	}

	switch v := args[0].(type) {
	case float64:
		return math.Abs(v), nil
	case int:
		return math.Abs(float64(v)), nil
	case int64:
		return math.Abs(float64(v)), nil
	default:
		return nil, errors.NewError(errors.ErrInvalidArgument, "abs function requires numeric argument", fmt.Sprintf("%v", v))
	}
}

// CeilFunction 实现向上取整函数
type CeilFunction struct{}

func (f *CeilFunction) Name() string {
	return "ceil"
}

func (f *CeilFunction) Call(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, errors.NewError(errors.ErrInvalidArgument, "ceil function requires exactly one argument", fmt.Sprintf("%v", args))
	}

	switch v := args[0].(type) {
	case float64:
		return math.Ceil(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	default:
		return nil, errors.NewError(errors.ErrInvalidArgument, "ceil function requires numeric argument", fmt.Sprintf("%v", v))
	}
}

// FloorFunction 实现向下取整函数
type FloorFunction struct{}

func (f *FloorFunction) Name() string {
	return "floor"
}

func (f *FloorFunction) Call(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, errors.NewError(errors.ErrInvalidArgument, "floor function requires exactly one argument", fmt.Sprintf("%v", args))
	}

	switch v := args[0].(type) {
	case float64:
		return math.Floor(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	default:
		return nil, errors.NewError(errors.ErrInvalidArgument, "floor function requires numeric argument", fmt.Sprintf("%v", v))
	}
}

// RoundFunction 实现四舍五入函数
type RoundFunction struct{}

func (f *RoundFunction) Name() string {
	return "round"
}

func (f *RoundFunction) Call(args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, errors.NewError(errors.ErrInvalidArgument, "round function requires exactly one argument", fmt.Sprintf("%v", args))
	}

	switch v := args[0].(type) {
	case float64:
		return math.Round(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	default:
		return nil, errors.NewError(errors.ErrInvalidArgument, "round function requires numeric argument", fmt.Sprintf("%v", v))
	}
}

// MaxFunction 实现最大值函数
type MaxFunction struct{}

func (f *MaxFunction) Name() string {
	return "max"
}

func (f *MaxFunction) Call(args []interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, errors.NewError(errors.ErrInvalidArgument, "max function requires at least one argument", fmt.Sprintf("%v", args))
	}

	var max float64
	first := true

	for _, arg := range args {
		switch v := arg.(type) {
		case float64:
			if first || v > max {
				max = v
				first = false
			}
		case int:
			if first || float64(v) > max {
				max = float64(v)
				first = false
			}
		case int64:
			if first || float64(v) > max {
				max = float64(v)
				first = false
			}
		default:
			return nil, errors.NewError(errors.ErrInvalidArgument, "max function requires numeric arguments", fmt.Sprintf("%v", v))
		}
	}

	return max, nil
}

// MinFunction 实现最小值函数
type MinFunction struct{}

func (f *MinFunction) Name() string {
	return "min"
}

func (f *MinFunction) Call(args []interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, errors.NewError(errors.ErrInvalidArgument, "min function requires at least one argument", fmt.Sprintf("%v", args))
	}

	var min float64
	first := true

	for _, arg := range args {
		switch v := arg.(type) {
		case float64:
			if first || v < min {
				min = v
				first = false
			}
		case int:
			if first || float64(v) < min {
				min = float64(v)
				first = false
			}
		case int64:
			if first || float64(v) < min {
				min = float64(v)
				first = false
			}
		default:
			return nil, errors.NewError(errors.ErrInvalidArgument, "min function requires numeric arguments", fmt.Sprintf("%v", v))
		}
	}

	return min, nil
}

func init() {
	Register("abs", &AbsFunction{})
	Register("ceil", &CeilFunction{})
	Register("floor", &FloorFunction{})
	Register("round", &RoundFunction{})
	Register("max", &MaxFunction{})
	Register("min", &MinFunction{})
}
