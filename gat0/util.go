package main

import (
	"strconv"
)

var itoa = strconv.Itoa

func ifStr(cond bool, ifTrue string, ifFalse string) string {
	if cond {
		return ifTrue
	}
	return ifFalse
}
