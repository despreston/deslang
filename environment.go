package deslang

import (
	"errors"
)

// lookup table for declared variables
type Environment struct {
	values    map[string]BasicLit
	Enclosing *Environment // parent scope
	global    bool         // global scope
}

func NewEnvironment(global bool) *Environment {
	return &Environment{
		values: make(map[string]BasicLit),
		global: global,
	}
}

func (env *Environment) Assign(tok Token, val BasicLit) error {
	s := string(tok.Lexeme)
	if _, has := env.values[s]; has {
		env.values[s] = val
		return nil
	}

	// As long as it hasn't reached the global scope, recursively check the chain
	// of environments.
	if !env.global {
		err := env.Enclosing.Assign(tok, val)
		if err != nil {
			return err
		}
	}

	return errors.New("Undefined variable '" + s + "'.")
}

func (env *Environment) Define(name string, lit BasicLit) {
	env.values[name] = lit
}

func (env *Environment) Get(tok Token) (BasicLit, error) {
	var lit BasicLit

	// Recursively look for the variable.
	lit, has := env.values[string(tok.Lexeme)]
	if !has {
		if env.global {
			return lit, errors.New("Undefined variable '" + string(tok.Lexeme) + "'.")
		}
		return env.Enclosing.Get(tok)
	}

	return lit, nil
}
