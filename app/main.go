package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// Ensures gofmt doesn't remove the "bytes" import above (feel free to remove this!)
var _ = bytes.ContainsAny

// Usage: echo <input_text> | your_program.sh -E <pattern>
func main() {
	if len(os.Args) < 3 || os.Args[1] != "-E" {
		fmt.Fprintf(os.Stderr, "usage: mygrep -E <pattern>\n")
		os.Exit(2) // 1 means no lines were selected, >1 means error
	}

	pattern := os.Args[2]

	line, err := io.ReadAll(os.Stdin) // assume we're only dealing with a single line
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: read input text: %v\n", err)
		os.Exit(2)
	}

	ok, err := matchLine(line, pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	if !ok {
		os.Exit(1)
	}

	// default exit code is 0 which means success
}

func matchLine(line []byte, pattern string) (bool, error) {
	if pattern == "" {
		return false, fmt.Errorf("unsupported pattern: empty")
	}
	for start := 0; start <= len(line); start++ {
		ok, err := matchHere(line[start:], pattern)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

func matchHere(text []byte, pat string) (bool, error) {
	ti := 0
	pi := 0
	for pi < len(pat) {
		if ti >= len(text) {
			return false, nil
		}
		switch pat[pi] {
		case '\\':
			if pi+1 >= len(pat) {
				return false, fmt.Errorf("dangling escape at end of pattern")
			}
			switch pat[pi+1] {
			case 'd':
				b := text[ti]
				if b < '0' || b > '9' {
					return false, nil
				}
				ti++
				pi += 2
			case 'w':
				b := text[ti]
				if !((b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_') {
					return false, nil
				}
				ti++
				pi += 2
			default:
				if text[ti] != pat[pi+1] {
					return false, nil
				}
				ti++
				pi += 2
			}
		case '[':
			closing := indexOfClosingBracket(pat, pi)
			if closing == -1 {
				return false, fmt.Errorf("unterminated character class starting at %d", pi)
			}
			inner := pat[pi+1 : closing]
			neg := false
			if len(inner) > 0 && inner[0] == '^' {
				neg = true
				inner = inner[1:]
			}
			if inner == "" {
				return false, fmt.Errorf("empty character class []")
			}
			b := text[ti]
			in := bytes.ContainsAny([]byte{b}, inner)
			if neg {
				if in {
					return false, nil
				}
			} else {
				if !in {
					return false, nil
				}
			}
			ti++
			pi = closing + 1
		default:
			if text[ti] != pat[pi] {
				return false, nil
			}
			ti++
			pi++
		}
	}
	return true, nil
}

func indexOfClosingBracket(pat string, open int) int {
	for i := open + 1; i < len(pat); i++ {
		if pat[i] == ']' {
			return i
		}
	}
	return -1
}
