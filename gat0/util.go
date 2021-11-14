package main

func ifStr(cond bool, ifTrue string, ifFalse string) string {
	if cond {
		return ifTrue
	}
	return ifFalse
}
