package main

import (
	"bufio"
	"bytes"

	. "github.com/metaleap/atmo/atem"
)

func ListFrom(str []byte) (ret Expr) {
	ret = StdFuncNil
	for i := len(str) - 1; i > -1; i-- {
		ret = ExprCall{Callee: ExprCall{Callee: StdFuncCons, Arg: ExprNumInt(str[i])}, Arg: ret}
	}
	return
}

func ListsFrom(strs []string) (ret Expr) {
	ret = StdFuncNil
	for i := len(strs) - 1; i > -1; i-- {
		ret = ExprCall{Callee: ExprCall{Callee: StdFuncCons, Arg: ListFrom([]byte(strs[i]))}, Arg: ret}
	}
	return
}

func stdinReadSplitterBy(sep byte) bufio.SplitFunc {
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if i := bytes.IndexByte(data, sep); i >= 0 {
			advance, token = i+1, data[0:i]
		} else if atEOF {
			advance, token = len(data), data
		}
		return
	}
}
