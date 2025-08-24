package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// Usage: echo <input_text> | your_program.sh -E <pattern>
func main() {
	if len(os.Args) < 3 || len(os.Args) > 4 || os.Args[1] != "-E" {
		fmt.Fprintf(os.Stderr, "usage: mygrep -E <pattern> [file]\n")
		os.Exit(2) // 1 means no lines were selected, >1 means error
	}

	pattern := os.Args[2]
	re, err := Compile(pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: invalid pattern: %v\n", err)
		os.Exit(2)
	}

	var input []byte
	var readErr error
	isFile := len(os.Args) == 4
	if isFile {
		input, readErr = os.ReadFile(os.Args[3])
	} else {
		input, readErr = io.ReadAll(os.Stdin)
	}
	if readErr != nil {
		fmt.Fprintf(os.Stderr, "error: read input: %v\n", readErr)
		os.Exit(2)
	}

	line := bytes.TrimSuffix(input, []byte("\n"))
	ok, matchErr := re.Match(line)
	if matchErr != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", matchErr)
		os.Exit(2)
	}

	if ok {
		if isFile {
			os.Stdout.Write(input)
		}
		os.Exit(0)
	}
	os.Exit(1)
}
