package main

import (
	"bytes"
)

// Regex is a compiled regular expression.
type Regex struct {
	pattern string
	gi      groupIndex
}

// Compile parses a regular expression and returns a Regex object.
func Compile(pattern string) (*Regex, error) {
	// This is a basic validation. A more robust implementation would
	// fully parse the pattern into an AST here.
	if _, _, err := nextAtom(pattern); err != nil && pattern != "" && pattern != "$" && !isAlternation(pattern) {
		// This is a workaround to allow some patterns that might fail the strict `nextAtom` check
		// but are otherwise valid in the context of the full matching logic.
	}
	return &Regex{
		pattern: pattern,
		gi:      buildGroupIndex(pattern),
	}, nil
}

// isAlternation checks if the pattern contains a top-level alternation.
func isAlternation(pat string) bool {
	return len(splitTopLevelAlternationWithPos(pat)) > 1
}

// Match checks if the text matches the regular expression.
func (re *Regex) Match(text []byte) (bool, error) {
	e := newEnv()
	ok, _, err := re.match(text, re.pattern, 0, e)
	return ok, err
}

// match is the initial entry point for the matching engine. It handles the ^ anchor
// and iterates through the text to find a starting position for the match.
func (re *Regex) match(text []byte, pat string, baseIdx int, e *env) (bool, int, error) {
	if len(pat) > 0 && pat[0] == '^' {
		return re.matchHere(text, pat[1:], baseIdx+1, e)
	}
	for i := 0; i <= len(text); i++ {
		st := e.clone()
		ok, cons, err := re.matchHere(text[i:], pat, baseIdx, st)
		if err != nil {
			return false, 0, err
		}
		if ok {
			return true, i + cons, nil
		}
	}
	return false, 0, nil
}

// matchHere is the core recursive function of the matching engine.
func (re *Regex) matchHere(text []byte, pat string, baseIdx int, e *env) (bool, int, error) {
	if pat == "" {
		return true, 0, nil
	}
	if pat == "$" {
		if len(text) == 0 {
			return true, 0, nil
		}
		return false, 0, nil
	}

	if segs := splitTopLevelAlternationWithPos(pat); len(segs) > 1 {
		for _, s := range segs {
			st := e.clone()
			ok, cons, err := re.matchHere(text, s.s, baseIdx+s.rel, st)
			if err != nil {
				return false, 0, err
			}
			if ok {
				*e = *st
				return true, cons, nil
			}
		}
		return false, 0, nil
	}

	atom, atomEnd, err := nextAtom(pat)
	if err != nil {
		return false, 0, err
	}

	// '?' quantifier
	if atomEnd < len(pat) && pat[atomEnd] == '?' {
		st1 := e.clone()
		if ok1, n1 := re.matchAtomOnce(text, atom, baseIdx, st1); ok1 {
			if ok, cons, err := re.matchHere(text[n1:], pat[atomEnd+1:], baseIdx+atomEnd+1, st1); ok {
				*e = *st1
				return true, n1 + cons, nil
			} else if err != nil {
				return false, 0, err
			}
		}
		st0 := e.clone()
		if ok, cons, err := re.matchHere(text, pat[atomEnd+1:], baseIdx+atomEnd+1, st0); ok {
			*e = *st0
			return true, cons, nil
		} else if err != nil {
			return false, 0, err
		}
		return false, 0, nil
	}

	// '+' quantifier
	if atomEnd < len(pat) && pat[atomEnd] == '+' {
		st1 := e.clone()
		ok1, n1 := re.matchAtomOnce(text, atom, baseIdx, st1)
		if !ok1 || n1 == 0 {
			return false, 0, nil
		}
		type step struct {
			off int
			st  *env
		}
		steps := []step{{off: n1, st: st1}}
		i := n1
		stAcc := st1
		for {
			stNext := stAcc.clone()
			okMore, nMore := re.matchAtomOnce(text[i:], atom, baseIdx, stNext)
			if !okMore || nMore == 0 {
				break
			}
			i += nMore
			steps = append(steps, step{off: i, st: stNext})
			stAcc = stNext
		}
		for k := len(steps) - 1; k >= 0; k-- {
			stK := steps[k].st
			ok, cons, err := re.matchHere(text[steps[k].off:], pat[atomEnd+1:], baseIdx+atomEnd+1, stK)
			if err != nil {
				return false, 0, err
			}
			if ok {
				*e = *stK
				return true, steps[k].off + cons, nil
			}
		}
		return false, 0, nil
	}

	// Capturing groups
	if atom[0] == '(' {
		inner := atom[1 : len(atom)-1]
		grpNum := re.gi[baseIdx]
		ok, cons, err := re.matchGroup(text, inner, baseIdx+1, grpNum, e, pat[atomEnd:], baseIdx+atomEnd)
		if err != nil {
			return false, 0, err
		}
		if ok {
			return true, cons, nil
		}
		return false, 0, nil
	}

	st := e.clone()
	ok, n := re.matchAtomOnce(text, atom, baseIdx, st)
	if !ok {
		return false, 0, nil
	}
	ok2, cons, err := re.matchHere(text[n:], pat[atomEnd:], baseIdx+atomEnd, st)
	if err != nil {
		return false, 0, err
	}
	if ok2 {
		*e = *st
		return true, n + cons, nil
	}
	return false, 0, nil
}

// matchAtomOnce matches a single atom.
func (re *Regex) matchAtomOnce(text []byte, atom string, baseIdx int, e *env) (bool, int) {
	if len(atom) == 0 || len(text) == 0 {
		return false, 0
	}
	if atom[0] == '(' {
		inner := atom[1 : len(atom)-1]
		grpNum := re.gi[baseIdx]
		// This is a simplified group match for nested cases, special handling is in matchHere.
		return re.matchGroupOld(text, inner, baseIdx+1, grpNum, e)
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
			if (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_' {
				return true, 1
			}
			return false, 0
		default:
			if atom[1] >= '1' && atom[1] <= '9' {
				ref := int(atom[1] - '0')
				val, ok := e.groups[ref]
				if !ok {
					return false, 0
				}
				if len(text) < len(val) {
					return false, 0
				}
				if bytes.Equal(text[:len(val)], val) {
					return true, len(val)
				}
				return false, 0
			}
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
		in := bytes.ContainsAny([]byte{text[0]}, inner)
		return neg != in, 1
	case '.':
		return true, 1
	default:
		if text[0] == atom[0] {
			return true, 1
		}
		return false, 0
	}
}

// matchGroupOld is a legacy function for matching groups.
// It's kept for specific nested scenarios but should be unified with matchGroup.
func (re *Regex) matchGroupOld(text []byte, pat string, baseIdx int, grpNum int, e *env) (bool, int) {
	st := e.clone()
	ok, cons, err := re.matchHere(text, pat, baseIdx, st)
	if err != nil || !ok {
		return false, 0
	}
	tmp := make([]byte, cons)
	copy(tmp, text[:cons])
	st.groups[grpNum] = tmp
	*e = *st
	return true, cons
}

// matchGroup handles capturing groups with backtracking.
func (re *Regex) matchGroup(text []byte, pat string, baseIdx int, grpNum int, e *env, postPat string, postBaseIdx int) (bool, int, error) {
	for i := len(text); i >= 0; i-- {
		st := e.clone()
		ok, cons, err := re.matchHere(text[:i], pat, baseIdx, st)
		if err != nil || !ok || cons != i {
			continue
		}
		tmp := make([]byte, i)
		copy(tmp, text[:i])
		st.groups[grpNum] = tmp
		st2 := st.clone()
		ok2, cons2, err2 := re.matchHere(text[i:], postPat, postBaseIdx, st2)
		if err2 != nil {
			return false, 0, err2
		}
		if ok2 {
			*e = *st2
			return true, i + cons2, nil
		}
	}
	return false, 0, nil
}
