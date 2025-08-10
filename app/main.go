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
	return match(line, pattern)
}

func match(text []byte, pat string) (bool, error) {
	if len(pat) > 0 && pat[0] == '^' {
		return matchHere(text, pat[1:])
	}
	for i := 0; i <= len(text); i++ {
		ok, err := matchHere(text[i:], pat)
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
	if len(pat) == 0 {
		return true, nil
	}
	if pat == "$" {
		return len(text) == 0, nil
	}

	ok, nextText, nextPat, err := matchAtom(text, pat)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}
	return matchHere(nextText, nextPat)
}

func matchAtom(text []byte, pat string) (bool, []byte, string, error) {
	if len(text) == 0 {
		return false, text, pat, nil
	}
	switch pat[0] {
	case '\\':
		if len(pat) < 2 {
			return false, text, pat, fmt.Errorf("dangling escape at end of pattern")
		}
		switch pat[1] {
		case 'd':
			b := text[0]
			if b < '0' || b > '9' {
				return false, text, pat, nil
			}
			return true, text[1:], pat[2:], nil
		case 'w':
			b := text[0]
			if !((b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_') {
				return false, text, pat, nil
			}
			return true, text[1:], pat[2:], nil
		default:
			if text[0] != pat[1] {
				return false, text, pat, nil
			}
			return true, text[1:], pat[2:], nil
		}

	case '[':
		closing := indexOfClosingBracket(pat, 0)
		if closing == -1 {
			return false, text, pat, fmt.Errorf("unterminated character class")
		}
		inner := pat[1:closing]
		neg := false
		if len(inner) > 0 && inner[0] == '^' {
			neg = true
			inner = inner[1:]
		}
		if len(inner) == 0 {
			return false, text, pat, fmt.Errorf("empty character class []")
		}
		b := text[0]
		in := bytes.ContainsAny([]byte{b}, inner)
		if neg {
			if in {
				return false, text, pat, nil
			}
		} else {
			if !in {
				return false, text, pat, nil
			}
		}
		return true, text[1:], pat[closing+1:], nil

	default:
		if text[0] != pat[0] {
			return false, text, pat, nil
		}
		return true, text[1:], pat[1:], nil
	}
}

func indexOfClosingBracket(pat string, open int) int {
	for i := open + 1; i < len(pat); i++ {
		if pat[i] == ']' {
			return i
		}
	}
	return -1
}
