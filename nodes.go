package deslang

import (
	"fmt"
	"strconv"
)

type litKind uint

const (
	nilLit litKind = iota
	floatLit
	stringLit
	boolLit
)

var types = map[litKind]string{
	nilLit:    "nil",
	floatLit:  "float",
	stringLit: "string",
	boolLit:   "boolean",
}

// ----------------------------------------------------------------------------
// Expressions

type (
	Expr interface {
		Interpret() (BasicLit, error)
	}

	Unary struct {
		Right Expr
		Op    Token
	}

	Binary struct {
		Left, Right Expr
		Op          Token
	}

	// (X)
	Grouping struct {
		X Expr
	}

	// Primitive value. It's stored as a string and the Kind field is used to
	// figure out typecasting. It would probably be faster to split this into a
	// few structs based on type but this way is simpler.
	BasicLit struct {
		Value string
		Kind  litKind
	}
)

// ----------------------------------------------------------------------------
// Interpretation methods

func (expr Unary) Interpret() (BasicLit, error) {
	var result BasicLit

	right, err := expr.Right.Interpret()
	if err != nil {
		return result, err
	}

	switch expr.Op.Type {
	case _bang:
		result.Kind = boolLit
		if right.Kind == boolLit && right.Value == "true" {
			result.Value = "false"
		} else {
			result.Value = "true"
		}
	case _minus:
		result.Kind = floatLit
		result.Value = fromFloat(-toFloat(right.Value))
	}

	return result, nil
}

func (expr Binary) Interpret() (BasicLit, error) {
	var result BasicLit

	left, err := expr.Left.Interpret()
	if err != nil {
		return result, err
	}

	right, err := expr.Right.Interpret()
	if err != nil {
		return result, err
	}

	// Type check
	if left.Kind != right.Kind {
		err := fmt.Errorf(
			"Invalid operation. Mismatched types %s and %s",
			types[left.Kind],
			types[right.Kind],
		)
		return result, err
	}

	switch expr.Op.Type {
	case _plus:
		if left.Kind == stringLit && right.Kind == stringLit {
			result.Value = left.Value + right.Value
			result.Kind = stringLit
		} else {
			result.Value = fromFloat(toFloat(left.Value) + toFloat(right.Value))
			result.Kind = floatLit
		}
	case _minus:
		result.Value = fromFloat(toFloat(left.Value) - toFloat(right.Value))
		result.Kind = floatLit
	case _slash:
		result.Value = fromFloat(toFloat(left.Value) / toFloat(right.Value))
		result.Kind = floatLit
	case _star:
		result.Value = fromFloat(toFloat(left.Value) * toFloat(right.Value))
		result.Kind = floatLit
	case _greater:
		result.Kind = boolLit
		if toFloat(left.Value) > toFloat(right.Value) {
			result.Value = "true"
		} else {
			result.Value = "false"
		}
	case _greater_equal:
		result.Kind = boolLit
		if toFloat(left.Value) >= toFloat(right.Value) {
			result.Value = "true"
		} else {
			result.Value = "false"
		}
	case _less:
		result.Kind = boolLit
		if toFloat(left.Value) < toFloat(right.Value) {
			result.Value = "true"
		} else {
			result.Value = "false"
		}
	case _less_equal:
		result.Kind = boolLit
		if toFloat(left.Value) <= toFloat(right.Value) {
			result.Value = "true"
		} else {
			result.Value = "false"
		}
	case _bang_equal:
		result.Kind = boolLit
		if left.Value == right.Value {
			result.Value = "false"
		} else {
			result.Value = "true"
		}
	case _equal_equal:
		result.Kind = boolLit
		if left.Value == right.Value {
			result.Value = "true"
		} else {
			result.Value = "false"
		}
	default:
		result = BasicLit{Value: "", Kind: nilLit}
	}

	return result, nil
}

func (expr Grouping) Interpret() (BasicLit, error) {
	return expr.X.Interpret()
}

func (expr BasicLit) Interpret() (BasicLit, error) {
	return expr, nil
}

func toFloat(s string) float64 {
	// TODO: Handle err when converting string to int
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func fromFloat(f float64) string {
	return strconv.FormatFloat(f, 'E', -1, 64)
}
