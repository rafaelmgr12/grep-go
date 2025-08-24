package main

import "bytes"

func matchLine(line []byte, pattern string) (bool, error) {
	idx := buildGroupIndex(pattern)
	st := newEnv()
	ok, _, err := match(line, pattern, 0, pattern, idx, st)
	return ok, err
}

func match(text []byte, pat string, baseIdx int, full string, gi groupIndex, e *env) (bool, int, error) {
	if len(pat) > 0 && pat[0] == '^' {
		return matchHere(text, pat[1:], baseIdx+1, full, gi, e)
	}
	for i := 0; i <= len(text); i++ {
		st := cloneEnv(e)
		ok, cons, err := matchHere(text[i:], pat, baseIdx, full, gi, &st)
		if err != nil {
			return false, 0, err
		}
		if ok {
			return true, i + cons, nil
		}
	}
	return false, 0, nil
}

func matchHere(text []byte, pat string, baseIdx int, full string, gi groupIndex, e *env) (bool, int, error) {
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
			st := cloneEnv(e)
			ok, cons, err := matchHere(text, s.s, baseIdx+s.rel, full, gi, &st)
			if err != nil {
				return false, 0, err
			}
			if ok {
				*e = st
				return true, cons, nil
			}
		}
		return false, 0, nil
	}

	atom, atomEnd, err := nextAtom(pat)
	if err != nil {
		return false, 0, err
	}

	// '?'
	if atomEnd < len(pat) && pat[atomEnd] == '?' {
		st1 := cloneEnv(e)
		if ok1, n1 := matchAtomOnce(text, atom, baseIdx, full, gi, &st1); ok1 {
			if ok, cons, err := matchHere(text[n1:], pat[atomEnd+1:], baseIdx+atomEnd+1, full, gi, &st1); ok {
				*e = st1
				return true, n1 + cons, nil
			} else if err != nil {
				return false, 0, err
			}
		}
		st0 := cloneEnv(e)
		if ok, cons, err := matchHere(text, pat[atomEnd+1:], baseIdx+atomEnd+1, full, gi, &st0); ok {
			*e = st0
			return true, cons, nil
		} else if err != nil {
			return false, 0, err
		}
		return false, 0, nil
	}

	// '+'
	if atomEnd < len(pat) && pat[atomEnd] == '+' {
		st1 := cloneEnv(e)
		ok1, n1 := matchAtomOnce(text, atom, baseIdx, full, gi, &st1)
		if !ok1 || n1 == 0 {
			return false, 0, nil
		}
		type step struct {
			off int
			st  env
		}
		steps := []step{{off: n1, st: st1}}
		i := n1
		stAcc := st1
		for {
			stNext := cloneEnv(&stAcc)
			okMore, nMore := matchAtomOnce(text[i:], atom, baseIdx, full, gi, &stNext)
			if !okMore || nMore == 0 {
				break
			}
			i += nMore
			steps = append(steps, step{off: i, st: stNext})
			stAcc = stNext
		}
		for k := len(steps) - 1; k >= 0; k-- {
			stK := steps[k].st
			ok, cons, err := matchHere(text[steps[k].off:], pat[atomEnd+1:], baseIdx+atomEnd+1, full, gi, &stK)
			if err != nil {
				return false, 0, err
			}
			if ok {
				*e = stK
				return true, steps[k].off + cons, nil
			}
		}
		return false, 0, nil
	}

	// Special handling for capturing groups
	if atom[0] == '(' {
		inner := atom[1 : len(atom)-1]
		grpNum := gi[baseIdx]
		ok, cons, err := matchGroup(text, inner, baseIdx+1, full, gi, grpNum, e, pat[atomEnd:], baseIdx+atomEnd)
		if err != nil {
			return false, 0, err
		}
		if ok {
			return true, cons, nil
		}
		return false, 0, nil
	}

	ok, n := matchAtomOnce(text, atom, baseIdx, full, gi, e)
	if !ok {
		return false, 0, nil
	}
	ok2, cons, err := matchHere(text[n:], pat[atomEnd:], baseIdx+atomEnd, full, gi, e)
	if err != nil {
		return false, 0, err
	}
	if ok2 {
		return true, n + cons, nil
	}
	return false, 0, nil
}

func matchAtomOnce(text []byte, atom string, baseIdx int, full string, gi groupIndex, e *env) (bool, int) {
	if len(atom) == 0 {
		return false, 0
	}
	if len(text) == 0 {
		return false, 0
	}
	if atom[0] == '(' {
		inner := atom[1 : len(atom)-1]
		grpNum := gi[baseIdx]
		ok, cons := matchGroupOld(text, inner, baseIdx+1, full, gi, grpNum, e) // Use old version for nested if needed, but since special handled above, this shouldn't be called for groups
		return ok, cons
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

// Old matchGroup for if called, but should not be for top-level groups
func matchGroupOld(text []byte, pat string, baseIdx int, full string, gi groupIndex, grpNum int, e *env) (bool, int) {
	st := cloneEnv(e)
	ok, cons, err := matchHere(text, pat, baseIdx, full, gi, &st)
	if err != nil || !ok {
		return false, 0
	}
	tmp := make([]byte, cons)
	copy(tmp, text[:cons])
	st.groups[grpNum] = tmp
	*e = st
	return true, cons
}

// New matchGroup with backtracking support
func matchGroup(text []byte, pat string, baseIdx int, full string, gi groupIndex, grpNum int, e *env, postPat string, postBaseIdx int) (bool, int, error) {
	for i := len(text); i >= 0; i-- {
		st := cloneEnv(e)
		ok, cons, err := matchHere(text[:i], pat, baseIdx, full, gi, &st)
		if err != nil {
			continue
		}
		if !ok || cons != i {
			continue
		}
		tmp := make([]byte, i)
		copy(tmp, text[:i])
		st.groups[grpNum] = tmp
		st2 := cloneEnv(&st)
		ok2, cons2, err2 := matchHere(text[i:], postPat, postBaseIdx, full, gi, &st2)
		if err2 != nil {
			return false, 0, err2
		}
		if ok2 {
			*e = st2
			return true, i + cons2, nil
		}
	}
	return false, 0, nil
}
