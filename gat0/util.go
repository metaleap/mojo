package main

import (
	"strconv"
)

var itoa = strconv.Itoa
var strQuote = strconv.Quote

func ifStr(cond bool, ifTrue string, ifFalse string) string {
	if cond {
		return ifTrue
	}
	return ifFalse
}
