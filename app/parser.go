package main

import "fmt"

type groupIndex map[int]int

func buildGroupIndex(pat string) groupIndex {
	g := make(groupIndex)
	esc := false
	br := 0
	num := 0
	for i := 0; i < len(pat); i++ {
		c := pat[i]
		if esc {
			esc = false
			continue
		}

		switch c {
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
				num++
				g[i] = num
			}
		case ')':
			if br == 0 {
			}
		}
	}
	return g
}

func indexOfClosingBracket(pat string, open int) int {
	esc := false
	for i := open + 1; i < len(pat); i++ {
		if esc {
			esc = false
			continue
		}
		switch pat[i] {
		case '\\':
			esc = true
		case ']':
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

// splitTopLevelAlternation splits pat by '|' ONLY when not inside () or [] and not escaped.
func splitTopLevelAlternation(pat string) []string {
	var parts []string
	last := 0
	parenDepth := 0
	bracketDepth := 0
	esc := false

	for i := 0; i < len(pat); i++ {
		c := pat[i]

		if esc {
			esc = false
			continue
		}

		switch c {
		case '\\':
			esc = true
		case '[':
			// bracket depth is independent of paren depth
			bracketDepth++
		case ']':
			if bracketDepth > 0 {
				bracketDepth--
			}
		case '(':
			if bracketDepth == 0 {
				parenDepth++
			}
		case ')':
			if bracketDepth == 0 && parenDepth > 0 {
				parenDepth--
			}
		case '|':
			if parenDepth == 0 && bracketDepth == 0 {
				parts = append(parts, pat[last:i])
				last = i + 1
			}
		}
	}

	if len(parts) == 0 {
		return []string{pat}
	}
	parts = append(parts, pat[last:])
	return parts
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
		// backref \1 como átomo especial nesta etapa
		if pat[1] == '1' {
			return pat[:2], 2, nil
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

type altSeg struct {
	s   string
	rel int
}

func splitTopLevelAlternationWithPos(pat string) []altSeg {
	var segs []altSeg
	last := 0
	parenDepth := 0
	bracketDepth := 0
	esc := false

	for i := 0; i < len(pat); i++ {
		c := pat[i]

		if esc {
			esc = false
			continue
		}

		switch c {
		case '\\':
			esc = true
		case '[':
			bracketDepth++
		case ']':
			if bracketDepth > 0 {
				bracketDepth--
			}
		case '(':
			if bracketDepth == 0 {
				parenDepth++
			}
		case ')':
			if bracketDepth == 0 && parenDepth > 0 {
				parenDepth--
			}
		case '|':
			if parenDepth == 0 && bracketDepth == 0 {
				segs = append(segs, altSeg{s: pat[last:i], rel: last})
				last = i + 1
			}
		}
	}

	if len(segs) == 0 {
		return []altSeg{{s: pat, rel: 0}}
	}
	segs = append(segs, altSeg{s: pat[last:], rel: last})
	return segs
}
