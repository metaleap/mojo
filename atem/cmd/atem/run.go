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

func ToBytes(maybeNumList []Expr) (retNumListAsBytes []byte) {
	if maybeNumList != nil {
		retNumListAsBytes = make([]byte, 0, len(maybeNumList))
		for _, expr := range maybeNumList {
			if num, ok := expr.(ExprNumInt); ok && num > -1 && num < 256 {
				retNumListAsBytes = append(retNumListAsBytes, byte(num))
			} else {
				retNumListAsBytes = nil
				break
			}
		}
	}
	return
}

func List(prog Prog, expr Expr) (ret []Expr) {
	ret = make([]Expr, 0, 1024)
	for again, next := true, expr; again; {
		again = false
		if fouter, ok0 := next.(ExprFuncRef); ok0 && fouter == StdFuncNil { // clean end-of-list
			break
		} else if aouter, ok1 := next.(ExprCall); ok1 {
			if ainner, ok2 := aouter.Callee.(ExprCall); ok2 {
				if finner, ok3 := ainner.Callee.(ExprFuncRef); ok3 && finner == StdFuncCons {
					elem := prog.Eval(ainner.Arg, nil)
					again, next, ret = true, prog.Eval(aouter.Arg, nil), append(ret, elem)
				}
			}
		}
		if !again {
			ret = nil
		}
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
