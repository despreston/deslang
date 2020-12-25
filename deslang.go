// A basic C-style language interpreter that should be used for nothing.
package deslang

import (
	"fmt"
	"io"
)

type errorHandler func(int, string, string)

// Scan, check for errors, parse, check for errors, interpret, print result or
// any runtime errors. Will return an error if something unexpected goes wrong
// while attempting to scan, e.g. an issue reading from src; this error has
// nothing to do with syntax or runtime errors. Syntax errors, parsing errors,
// runtime errors, and evaluated result are printed via 'out' Writer.
//
// Run should be called when parsing every new source of code. When running as a
// REPL, Run should be called on every new line.
func Run(src io.Reader, out io.Writer) error {
	var hadErr bool

	errh := func(line int, where string, msg string) {
		hadErr = true
		fmt.Fprintf(out, "[line %d] Error %s: %s\n", line, where, msg)
	}

	tokens, err := NewScanner(src, errh).Scan()
	if err != nil && err != io.EOF {
		return err
	}

	if hadErr {
		return nil
	}

	stmts := NewParser(tokens, errh).Parse()

	if hadErr {
		return nil
	}

	for _, s := range stmts {
		err := s.Execute(out)
		if err != nil {
			fmt.Fprintln(out, err.Error())
			return nil
		}
	}

	return nil
}
