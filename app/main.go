package main

import (
	"bufio"
	"fmt"
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

	var reader *os.File
	isFile := len(os.Args) == 4
	if isFile {
		reader, err = os.Open(os.Args[3])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: open file: %v\n", err)
			os.Exit(2)
		}
		defer reader.Close()
	} else {
		reader = os.Stdin
	}

	scanner := bufio.NewScanner(reader)
	found := false
	for scanner.Scan() {
		line := scanner.Text()
		ok, matchErr := re.Match([]byte(line))
		if matchErr != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", matchErr)
			os.Exit(2)
		}
		if ok {
			fmt.Println(line)
			found = true
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "error: read input: %v\n", err)
		os.Exit(2)
	}

	if found {
		os.Exit(0)
	}
	os.Exit(1)
}
