// A basic C-style language interpreter that should be used for nothing.
package deslang

import (
	"fmt"
	"io"
)

type errorHandler func(int, string, string)

type Interpreter struct {
	hadErr  bool // True if there as an error doing the process
	scanner *Scanner
	parser  *Parser
	env     *Environment
	out     io.Writer
}

func NewInterpreter(out io.Writer) *Interpreter {
	var interpreter Interpreter

	interpreter.scanner = NewScanner(interpreter.errh)
	interpreter.parser = NewParser(interpreter.errh)
	interpreter.env = NewEnvironment()
	interpreter.out = out

	return &interpreter
}

// Parser and Scanner will report any syntax errors by calling this method.
func (interpreter *Interpreter) errh(line int, where string, msg string) {
	interpreter.hadErr = true
	fmt.Fprintf(interpreter.out, "[line %d] Error %s: %s\n", line, where, msg)
}

// Scan, check for errors, parse, check for errors, interpret, print result or
// any runtime errors. Will return an error if something unexpected goes wrong
// while attempting to scan, e.g. an issue reading from src; this error has
// nothing to do with syntax or runtime errors. Syntax errors, parsing errors,
// runtime errors, and evaluated result are printed via 'out' Writer.
//
// Run should be called when parsing every new source of code. When running as a
// REPL, Run should be called on every new line.
func (interpreter *Interpreter) Run(src io.Reader) error {
	interpreter.hadErr = false

	tokens, err := interpreter.scanner.Scan(src)
	if err != nil && err != io.EOF {
		return err
	}

	if interpreter.hadErr {
		return nil
	}

	stmts := interpreter.parser.Parse(tokens)

	if interpreter.hadErr {
		return nil
	}

	for _, s := range stmts {
		err := s.Execute(interpreter.out, interpreter.env)
		if err != nil {
			fmt.Fprintln(interpreter.out, err.Error())
			return nil
		}
	}

	return nil
}
