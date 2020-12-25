package deslang

import (
	"errors"
)

// lookup table for declared variables
type Environment struct {
	values map[string]BasicLit
}

func NewEnvironment() *Environment {
	return &Environment{
		values: make(map[string]BasicLit),
	}
}

func (env *Environment) Define(name string, lit BasicLit) {
	env.values[name] = lit
}

func (env *Environment) Get(tok Token) (BasicLit, error) {
	var lit BasicLit

	lit, has := env.values[string(tok.Lexeme)]
	if !has {
		return lit, errors.New("Undefined variable '" + string(tok.Lexeme) + "'.")
	}

	return lit, nil
}
