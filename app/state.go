package main

type env struct {
	groups map[int][]byte
}

func newEnv() *env {
	return &env{groups: make(map[int][]byte)}
}

func cloneEnv(e *env) env {
	if e == nil {
		return env{}
	}
	cl := env{groups: make(map[int][]byte)}
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
