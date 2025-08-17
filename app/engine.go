package main

import "bytes"

func matchLine(line []byte, pattern string) (bool, error) {
	st := &env{}
	ok, _, err := match(line, pattern, st)
	return ok, err
}

func match(text []byte, pat string, e *env) (bool, int, error) {
	if len(pat) > 0 && pat[0] == '^' {
		return matchHere(text, pat[1:], e)
	}
	for i := 0; i <= len(text); i++ {
		st := cloneEnv(e)
		ok, cons, err := matchHere(text[i:], pat, &st)
		if err != nil {
			return false, 0, err
		}
		if ok {
			return true, i + cons, nil
		}
	}
	return false, 0, nil
}

func matchHere(text []byte, pat string, e *env) (bool, int, error) {
	// Handle end anchor specially if pat ends with '$' but is not just "$"
	if len(pat) > 0 && pat[len(pat)-1] == '$' && pat != "$" {
		ok, cons, err := matchHere(text, pat[:len(pat)-1], e)
		if err != nil {
			return false, 0, err
		}
		if ok && cons == len(text) {
			return true, cons, nil
		}
		return false, 0, nil
	}

	// base cases
	if pat == "" {
		return true, 0, nil
	}
	if pat == "$" {
		if len(text) == 0 {
			return true, 0, nil
		}
		return false, 0, nil
	}

	// Alternation at top-level for this slice (must be BEFORE atom parsing)
	if alts := splitTopLevelAlternation(pat); len(alts) > 1 {
		for _, alt := range alts {
			st := cloneEnv(e)
			ok, cons, err := matchHere(text, alt, &st)
			if err != nil {
				return false, 0, err
			}
			if ok {
				*e = st // propaga captura do ramo que venceu
				return true, cons, nil
			}
		}
		return false, 0, nil
	}

	atom, atomEnd, err := nextAtom(pat)
	if err != nil {
		return false, 0, err
	}

	// '?' quantifier (0 or 1)
	if atomEnd < len(pat) && pat[atomEnd] == '?' {
		// tenta 1x
		st1 := cloneEnv(e)
		ok1, n1 := matchAtomOnce(text, atom, &st1)
		if ok1 {
			ok2, cons2, err := matchHere(text[n1:], pat[atomEnd+1:], &st1)
			if err != nil {
				return false, 0, err
			}
			if ok2 {
				*e = st1
				return true, n1 + cons2, nil
			}
		}
		// tenta 0x
		st0 := cloneEnv(e)
		ok0, cons0, err := matchHere(text, pat[atomEnd+1:], &st0)
		if err != nil {
			return false, 0, err
		}
		if ok0 {
			*e = st0
			return true, cons0, nil
		}
		return false, 0, nil
	}

	// '+' quantifier (1 or more), greedy with backtracking
	if atomEnd < len(pat) && pat[atomEnd] == '+' {
		st1 := cloneEnv(e)
		ok1, n1 := matchAtomOnce(text, atom, &st1)
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
			okMore, nMore := matchAtomOnce(text[i:], atom, &stNext) // CORRETO: text[i:]
			if !okMore || nMore == 0 {
				break
			}
			i += nMore
			steps = append(steps, step{off: i, st: stNext})
			stAcc = stNext
		}

		// backtrack do maior para o menor
		for k := len(steps) - 1; k >= 0; k-- {
			stK := steps[k].st
			ok, consRest, err := matchHere(text[steps[k].off:], pat[atomEnd+1:], &stK)
			if err != nil {
				return false, 0, err
			}
			if ok {
				*e = stK
				return true, steps[k].off + consRest, nil
			}
		}

		return false, 0, nil
	}

	// Single occurrence
	ok, n := matchAtomOnce(text, atom, e)
	if !ok {
		return false, 0, nil
	}
	ok2, cons2, err := matchHere(text[n:], pat[atomEnd:], e)
	if err != nil {
		return false, 0, err
	}
	if ok2 {
		return true, n + cons2, nil
	}
	return false, 0, nil
}

func matchAtomOnce(text []byte, atom string, e *env) (bool, int) {
	if len(atom) == 0 {
		return false, 0
	}
	// Be careful: some atoms (like '$') are handled in matchHere, not here.
	if len(text) == 0 {
		// only zero-width atoms could match empty text; we não suportamos aqui
		// (evita loops; backref vazia é tratada como falha)
		return false, 0
	}

	// Group (...) — match the inner pattern and determine consumption
	if atom[0] == '(' {
		inner := atom[1 : len(atom)-1]
		ok, n, err := matchGroup(text, inner, e)
		if err != nil {
			return false, 0
		}
		return ok, n
	}

	switch atom[0] {
	case '\\':
		if len(atom) < 2 {
			return false, 0
		}
		switch atom[1] {
		case '1':
			// backref: precisa já ter grupo 1 capturado e não-vazio
			if !e.capSet || len(e.capVal) == 0 {
				return false, 0
			}
			if len(text) < len(e.capVal) {
				return false, 0
			}
			if bytes.Equal(text[:len(e.capVal)], e.capVal) {
				return true, len(e.capVal)
			}
			return false, 0

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
			// escape literal: \X
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
// Returns (matched, consumed, error) where consumed is how much text was used
func matchGroup(text []byte, pat string, e *env) (bool, int, error) {
	// Greedy: maior prefixo primeiro
	for i := len(text); i >= 0; i-- {
		st := cloneEnv(e)
		ok, cons, err := matchHere(text[:i], pat, &st)
		if err != nil {
			return false, 0, err
		}
		if ok && cons == i {
			// define captura do grupo 1 se ainda não definida
			if !st.capSet {
				st.capSet = true
				tmp := make([]byte, i)
				copy(tmp, text[:i]) // copia para não depender do slice original
				st.capVal = tmp
			}
			*e = st
			return true, i, nil
		}
	}
	return false, 0, nil
}
