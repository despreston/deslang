package deslang

import (
	"fmt"
	"io"
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

func toFloat(s string) float64 {
	// TODO: Handle err when converting string to int
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func fromFloat(f float64) string {
	return strconv.FormatFloat(f, 'E', -1, 64)
}

func isTruthy(lit BasicLit) bool {
	switch lit.Kind {
	case nilLit:
		return false
	case floatLit:
		return lit.Value != "0"
	case stringLit:
		return len(lit.Value) > 0
	case boolLit:
		return lit.Value == "true"
	default:
		return false
	}
}

// ----------------------------------------------------------------------------
// Expressions

type (
	Expr interface {
		Interpret(*Environment) (BasicLit, error)
	}

	Unary struct {
		Right Expr
		Op    Token
	}

	Binary struct {
		Left, Right Expr
		Op          Token
	}

	Assign struct {
		Name  Token
		Value Expr
	}

	// (X)
	Grouping struct {
		X Expr
	}

	Variable struct {
		Name Token
	}

	// and, or
	Logical struct {
		Left, Right Expr
		Op          Token
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

func (expr Unary) Interpret(env *Environment) (BasicLit, error) {
	var result BasicLit

	right, err := expr.Right.Interpret(env)
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

func (expr Binary) Interpret(env *Environment) (BasicLit, error) {
	var result BasicLit

	left, err := expr.Left.Interpret(env)
	if err != nil {
		return result, err
	}

	right, err := expr.Right.Interpret(env)
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

func (expr Assign) Interpret(env *Environment) (BasicLit, error) {
	var result BasicLit

	result, err := expr.Value.Interpret(env)
	if err != nil {
		return result, err
	}

	env.Assign(expr.Name, result)
	return result, nil
}

func (expr Grouping) Interpret(env *Environment) (BasicLit, error) {
	return expr.X.Interpret(env)
}

func (expr BasicLit) Interpret(env *Environment) (BasicLit, error) {
	return expr, nil
}

func (expr Variable) Interpret(env *Environment) (BasicLit, error) {
	return env.Get(expr.Name)
}

func (expr Logical) Interpret(env *Environment) (BasicLit, error) {
	left, err := expr.Left.Interpret(env)
	if err != nil {
		return BasicLit{}, nil
	}

	right, err := expr.Right.Interpret(env)
	if err != nil {
		return BasicLit{}, nil
	}

	if expr.Op.Type == _or {
		if isTruthy(left) {
			return left, nil
		}
	} else if !isTruthy(left) {
		return left, nil
	}

	return right, nil
}

// ----------------------------------------------------------------------------
// Statements

type (
	Stmt interface {
		Execute(io.Writer, *Environment) error
	}

	// Empty Stmt
	NilStmt struct{}

	ExprStmt struct {
		Expr Expr
	}

	PrintStmt struct {
		Expr Expr
	}

	VarStmt struct {
		Name Token
		Expr Expr
	}

	AssignStmt struct {
		Name Token
		Expr Expr
	}

	BlockStmt struct {
		Stmts []Stmt
	}

	IfStmt struct {
		Cond Expr
		Then Stmt
		Else Stmt
	}
)

// ----------------------------------------------------------------------------
// Executor methods

func (stmt ExprStmt) Execute(_ io.Writer, env *Environment) error {
	stmt.Expr.Interpret(env)
	return nil
}

func (stmt PrintStmt) Execute(w io.Writer, env *Environment) error {
	lit, err := stmt.Expr.Interpret(env)
	if err != nil {
		return err
	}

	fmt.Fprintln(w, lit.Value)
	return nil
}

func (stmt VarStmt) Execute(_ io.Writer, env *Environment) error {
	lit, err := stmt.Expr.Interpret(env)
	if err != nil {
		return err
	}

	env.Define(string(stmt.Name.Lexeme), lit)
	return nil
}

func (stmt AssignStmt) Execute(_ io.Writer, env *Environment) error {
	lit, err := stmt.Expr.Interpret(env)
	if err != nil {
		return err
	}

	env.Assign(stmt.Name, lit)
	return nil
}

func (stmt BlockStmt) Execute(w io.Writer, env *Environment) error {
	local := NewEnvironment(false)
	local.Enclosing = env
	return executeBlock(stmt.Stmts, w, local)
}

func (NilStmt) Execute(w io.Writer, env *Environment) error {
	return nil
}

func (stmt IfStmt) Execute(w io.Writer, env *Environment) error {
	condLit, err := stmt.Cond.Interpret(env)
	if err != nil {
		return err
	}

	if isTruthy(condLit) {
		return stmt.Then.Execute(w, env)
	} else {
		switch stmt.Else.(type) {
		case NilStmt:
		default:
			return stmt.Else.Execute(w, env)
		}
	}

	return nil
}

func executeBlock(stmts []Stmt, w io.Writer, env *Environment) error {
	for _, s := range stmts {
		err := s.Execute(w, env)
		if err != nil {
			return err
		}
	}

	return nil
}
