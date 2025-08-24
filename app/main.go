package main

import (
	"bufio"
	"fmt"
	"os"
)

// Usage: mygrep -E <pattern> [file ...]
func main() {
	if len(os.Args) < 3 || os.Args[1] != "-E" {
		fmt.Fprintf(os.Stderr, "usage: mygrep -E <pattern> [file ...]\n")
		os.Exit(2) // 1 means no lines were selected, >1 means error
	}

	pattern := os.Args[2]
	re, err := Compile(pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: invalid pattern: %v\n", err)
		os.Exit(2)
	}

	files := os.Args[3:]
	numFiles := len(files)
	found := false

	if numFiles == 0 {
		scanner := bufio.NewScanner(os.Stdin)
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
	} else {
		for _, file := range files {
			reader, err := os.Open(file)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: open file %s: %v\n", file, err)
				os.Exit(2)
			}
			defer reader.Close()

			scanner := bufio.NewScanner(reader)
			for scanner.Scan() {
				line := scanner.Text()
				ok, matchErr := re.Match([]byte(line))
				if matchErr != nil {
					fmt.Fprintf(os.Stderr, "error: %v\n", matchErr)
					os.Exit(2)
				}
				if ok {
					if numFiles > 1 {
						fmt.Printf("%s:%s\n", file, line)
					} else {
						fmt.Println(line)
					}
					found = true
				}
			}
			if err := scanner.Err(); err != nil {
				fmt.Fprintf(os.Stderr, "error: read file %s: %v\n", file, err)
				os.Exit(2)
			}
		}
	}

	if found {
		os.Exit(0)
	}
	os.Exit(1)
}
