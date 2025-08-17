package main

type env struct {
	capSet bool
	capVal []byte
}

func cloneEnv(e *env) env {
	if e == nil {
		return env{}
	}
	cl := env{capSet: e.capSet}
	if e.capVal != nil {
		cl.capVal = make([]byte, len(e.capVal))
		copy(cl.capVal, e.capVal)
	}
	return cl
}
