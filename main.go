package main

import (
	"fmt"
	"os"
)

// Usage: mygrep [-r] -E <pattern> [path ...]

// ...existing code...

func main() {
	args := parseArgs()
	re, err := Compile(args.Pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: invalid pattern: %v\n", err)
		os.Exit(2)
	}

	found := false
	paths := args.Paths
	if len(paths) == 0 {
		if args.Recursive {
			paths = []string{"."}
		} else {
			found = grepStdin(re)
			if found {
				os.Exit(0)
			}
			os.Exit(1)
		}
	}

	multiPrefix := len(paths) > 1 || args.Recursive
	for _, p := range paths {
		if args.Recursive {
			if grepRecursive(re, p, multiPrefix) {
				found = true
			}
		} else {
			if grepFile(re, p, multiPrefix) {
				found = true
			}
		}
	}
	if found {
		os.Exit(0)
	}
	os.Exit(1)
}
