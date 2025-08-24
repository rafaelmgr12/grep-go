package main

type env struct {
	groups map[int][]byte
}

func newEnv() *env {
	return &env{groups: make(map[int][]byte)}
}

func (e *env) clone() *env {
	if e == nil {
		return newEnv()
	}
	cl := newEnv()
	for k, v := range e.groups {
		if v == nil {
			cl.groups[k] = nil
			continue
		}
		cp := make([]byte, len(v))
		copy(cp, v)
		cl.groups[k] = cp
	}
	return cl
}
