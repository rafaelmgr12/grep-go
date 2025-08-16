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
	// default exit code 0
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
	// Alternation at top-level for this slice
	alts := splitAlternatives(pat)
	if len(alts) > 1 {
		for _, alt := range alts {
			ok, err := matchHere(text, alt)
			if err != nil {
				return false, err
			}
			if ok {
				return true, nil
			}
		}
		return false, nil
	}

	if len(pat) == 0 {
		return true, nil
	}
	if pat == "$" {
		return len(text) == 0, nil
	}

	atom, atomEnd, err := nextAtom(pat)
	if err != nil {
		return false, err
	}

	// '?' quantifier (0 or 1)
	if atomEnd < len(pat) && pat[atomEnd] == '?' {
		ok1, n1 := matchAtomOnce(text, atom)
		if ok1 {
			if ok, err := matchHere(text[n1:], pat[atomEnd+1:]); err != nil {
				return false, err
			} else if ok {
				return true, nil
			}
		}
		return matchHere(text, pat[atomEnd+1:])
	}

	// '+' quantifier (1 or more), greedy with backtracking
	if atomEnd < len(pat) && pat[atomEnd] == '+' {
		ok1, n1 := matchAtomOnce(text, atom)
		if !ok1 {
			return false, nil
		}
		i := n1
		for {
			okMore, nMore := matchAtomOnce(text[i:], atom)
			if !okMore {
				break
			}
			i += nMore
		}
		// backtrack the number of repetitions
		for consumed := i; consumed >= n1; consumed-- {
			ok, err := matchHere(text[consumed:], pat[atomEnd+1:])
			if err != nil {
				return false, err
			}
			if ok {
				return true, nil
			}
		}
		return false, nil
	}

	// Single occurrence
	ok, n := matchAtomOnce(text, atom)
	if !ok {
		return false, nil
	}
	return matchHere(text[n:], pat[atomEnd:])
}

func nextAtom(pat string) (string, int, error) {
	if len(pat) == 0 {
		return "", 0, fmt.Errorf("empty pattern in nextAtom")
	}
	switch pat[0] {
	case '(':
		closing := indexOfClosingParen(pat, 0)
		if closing == -1 {
			return "", 0, fmt.Errorf("unterminated group")
		}
		return pat[:closing+1], closing + 1, nil
	case '+':
		return "", 0, fmt.Errorf("leading '+' without a preceding atom")
	case '?':
		return "", 0, fmt.Errorf("leading '?' without a preceding atom")
	case '\\':
		if len(pat) < 2 {
			return "", 0, fmt.Errorf("dangling escape at end of pattern")
		}
		return pat[:2], 2, nil
	case '[':
		closing := indexOfClosingBracket(pat, 0)
		if closing == -1 {
			return "", 0, fmt.Errorf("unterminated character class")
		}
		if closing == 1 {
			return "", 0, fmt.Errorf("empty character class []")
		}
		return pat[:closing+1], closing + 1, nil
	default:
		return pat[:1], 1, nil
	}
}

func matchAtomOnce(text []byte, atom string) (bool, int) {
	if len(text) == 0 || len(atom) == 0 {
		return false, 0
	}

	// Group (...) — match the inner pattern and determine consumption
	if atom[0] == '(' {
		inner := atom[1 : len(atom)-1]
		return matchGroup(text, inner)
	}

	switch atom[0] {
	case '\\':
		if len(atom) < 2 {
			return false, 0
		}
		switch atom[1] {
		case 'd':
			b := text[0]
			if b >= '0' && b <= '9' {
				return true, 1
			}
			return false, 0
		case 'w':
			b := text[0]
			if (b >= 'a' && b <= 'z') ||
				(b >= 'A' && b <= 'Z') ||
				(b >= '0' && b <= '9') ||
				b == '_' {
				return true, 1
			}
			return false, 0
		default:
			if text[0] == atom[1] {
				return true, 1
			}
			return false, 0
		}

	case '[':
		inner := atom[1 : len(atom)-1]
		neg := false
		if len(inner) > 0 && inner[0] == '^' {
			neg = true
			inner = inner[1:]
		}
		if len(inner) == 0 {
			return false, 0
		}
		in := bytes.ContainsAny([]byte{text[0]}, inner)
		if neg {
			if !in {
				return true, 1
			}
			return false, 0
		}
		if in {
			return true, 1
		}
		return false, 0

	case '.':
		return true, 1

	default:
		if text[0] == atom[0] {
			return true, 1
		}
		return false, 0
	}
}

// matchGroup tries to match a group pattern against text
// Returns (matched, consumed) where consumed is how much text was used
func matchGroup(text []byte, pat string) (bool, int) {

	isWord := func(b byte) bool {
		return (b >= 'a' && b <= 'z') ||
			(b >= 'A' && b <= 'Z') ||
			(b >= '0' && b <= '9') ||
			b == '_'
	}

	for consumed := len(text); consumed >= 0; consumed-- {
		fullPat := pat + "$"
		if ok, err := matchHere(text[:consumed], fullPat); err == nil && ok {

			if consumed < len(text) && isWord(text[consumed]) {
				continue
			}
			return true, consumed
		}
	}
	return false, 0
}

func indexOfClosingBracket(pat string, open int) int {
	for i := open + 1; i < len(pat); i++ {
		if pat[i] == ']' {
			return i
		}
	}
	return -1
}

func indexOfClosingParen(pat string, open int) int {
	depth := 0
	esc := false
	br := 0 // bracket depth to ignore ')' inside [...]
	for i := open; i < len(pat); i++ {
		if esc {
			esc = false
			continue
		}
		switch pat[i] {
		case '\\':
			esc = true
		case '[':
			br++
		case ']':
			if br > 0 {
				br--
			}
		case '(':
			if br == 0 {
				depth++
			}
		case ')':
			if br == 0 {
				depth--
				if depth == 0 {
					return i
				}
			}
		}
	}
	return -1
}

func splitAlternatives(pat string) []string {
	var parts []string
	last := 0
	parenDepth := 0
	brDepth := 0
	esc := false

	for i := 0; i < len(pat); i++ {
		c := pat[i]
		if esc {
			esc = false
			continue
		}
		if c == '\\' {
			esc = true
			continue
		}
		switch c {
		case '[':
			brDepth++
		case ']':
			if brDepth > 0 {
				brDepth--
			}
		case '(':
			if brDepth == 0 {
				parenDepth++
			}
		case ')':
			if brDepth == 0 && parenDepth > 0 {
				parenDepth--
			}
		case '|':
			if parenDepth == 0 && brDepth == 0 {
				parts = append(parts, pat[last:i])
				last = i + 1
			}
		}
	}
	parts = append(parts, pat[last:])
	return parts
}
