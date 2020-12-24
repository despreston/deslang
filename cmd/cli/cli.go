package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/despreston/deslang"
	"io"
	"os"
)

func main() {
	arglen := len(os.Args)
	var err error

	switch {
	case arglen > 2:
		fmt.Println("Usage: deslang [script]")
		os.Exit(64)
	case arglen == 2:
		err = runFile(os.Args[1])
	default:
		err = runPrompt()
	}

	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}

func runFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return deslang.Run(f, os.Stdout)
}

// Reads stdin until the first '\n' then calls 'run' with the input.
func runPrompt() error {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("deslang> ")

		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		deslang.Run(bytes.NewReader(line), os.Stdout)
	}
}
