package main

import (
	"strings"

	tl "github.com/metaleap/go-machines/toylam"
)

func dissectFunc(expr tl.Expr) (outerFuncs []*tl.ExprFunc, innerMostBody tl.Expr) {
	innerMostBody = expr
	for fn, _ := expr.(*tl.ExprFunc); fn != nil; fn, _ = fn.Body.(*tl.ExprFunc) {
		innerMostBody, outerFuncs = fn.Body, append(outerFuncs, fn)
	}
	return
}

func fullCopy(expr tl.Expr) tl.Expr {
	switch it := expr.(type) {
	case *tl.ExprLitNum:
		return &*it
	case *tl.ExprName:
		return &*it
	case *tl.ExprFunc:
		ret := *it
		ret.Body = fullCopy(ret.Body)
		return &ret
	case *tl.ExprCall:
		ret := *it
		ret.Callee, ret.CallArg = fullCopy(ret.Callee), fullCopy(ret.CallArg)
		return &ret
	}
	panic(expr)
}

func freeVars(expr tl.Expr, localNames map[string]string, results map[string]int) {
	switch it := expr.(type) {
	case *tl.ExprCall:
		freeVars(it.Callee, localNames, results)
		freeVars(it.CallArg, localNames, results)
	case *tl.ExprFunc:
		if _, exists := localNames[it.ArgName]; exists {
			panic(it.ArgName)
		}
		localNames[it.ArgName] = ""
		freeVars(it.Body, localNames, results)
		delete(localNames, it.ArgName)
	case *tl.ExprName:
		if it.IdxOrInstr <= 0 {
			if _, exists := localNames[it.NameVal]; !exists {
				if _, exists = inProg.TopDefs[it.NameVal]; !exists {
					if _, exists = inProg.TopDefs[tl.StdModuleName+"."+it.NameVal]; !exists {
						for tdname := range inProg.TopDefs {
							if exists = strings.HasSuffix(tdname, "."+it.NameVal) && strings.HasPrefix(tdname, tl.StdModuleName+"."); exists {
								break
							}
						}
						if !exists {
							if have, oops := results[it.NameVal]; oops && have != it.IdxOrInstr {
								panic(it.NameVal)
							}
							results[it.NameVal] = it.IdxOrInstr
						}
					}
				}
			}
		}
	}
}
