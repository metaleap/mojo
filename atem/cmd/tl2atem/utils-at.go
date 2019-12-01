package main

import (
	. "github.com/metaleap/atmo/atem"
)

type exprTmp int

func (me exprTmp) JsonSrc() string { return "-0" }

var exprNever = exprTmp(123456789)

func init() { OpPrtDst = func([]byte) (int, error) { panic("caught in `tryEval`") } }

func dissectCall(expr Expr) (innerMostCallee Expr, innerMostCalleeFnRef *ExprFuncRef, numCallArgs int, numCallArgsThatAreCalls int, numArgRefs int, allArgs []Expr) {
	for call, okc := expr.(ExprAppl); okc; call, okc = call.Callee.(ExprAppl) {
		innerMostCallee, numCallArgs, allArgs = call.Callee, numCallArgs+1, append([]Expr{call.Arg}, allArgs...)
		if _, isargcall := call.Arg.(ExprAppl); isargcall {
			numCallArgsThatAreCalls++
		} else if _, isargref := call.Arg.(ExprArgRef); isargref {
			numArgRefs++
		}
		if fnref, okf := call.Callee.(ExprFuncRef); okf {
			innerMostCalleeFnRef = &fnref
		} else if _, isargref := call.Callee.(ExprArgRef); isargref {
			numArgRefs++
		}
	}
	return
}

func doesHaveNonCalleeUses(prog Prog, fn ExprFuncRef) (doesHaveNonCalleeOccurrences bool) {
	var scrut func(Expr) Expr
	scrut = func(expr Expr) Expr {
		if doesHaveNonCalleeOccurrences {
			return nil
		} else if call, isc := expr.(ExprAppl); isc {
			if scrut(call.Arg); !doesHaveNonCalleeOccurrences {
				if _, isf := call.Callee.(ExprFuncRef); !isf {
					scrut(call.Callee)
				}
			}
			return nil
		} else if fnref, isf := expr.(ExprFuncRef); isf && fnref == fn {
			doesHaveNonCalleeOccurrences = true
		}
		return expr
	}
	for i := StdFuncCons + 1; int(i) < len(prog) && !doesHaveNonCalleeOccurrences; i++ {
		_ = walk(prog[i].Body, scrut)
	}
	return
}

func eq(prog Prog, expr Expr, cmp Expr) bool {
	if t1, ok1 := expr.(exprTmp); ok1 {
		t2, ok2 := expr.(exprTmp)
		return ok2 && t1 == t2
	}
	return prog.Eq(expr, cmp, false)
}

// some optimizers may drop certain arg uses while others may expect correct values in `FuncDef.Args`,
// so as a first step before a new round, we ensure they're all correct for that round.
func fixFuncDefArgsUsageNumbers(prog Prog) Prog {
	for i := range prog {
		for j := range prog[i].Args {
			prog[i].Args[j] = 0
		}
		_ = walk(prog[i].Body, func(expr Expr) Expr {
			if argref, ok := expr.(ExprArgRef); ok {
				if argref == 0 {
					println(prog[i].JsonSrc(false), len(prog[i].Args), "\t\t\t", argref, "\t>>>\t", (-argref)-1)
				}
				argref = (-argref) - 1
				prog[i].Args[argref] = 1 + prog[i].Args[argref]
			}
			return expr
		})
	}
	return prog
}

func rewriteCallArgs(callExpr ExprAppl, numCallArgs int, rewriter func(int, Expr) Expr, argIdxs []int) ExprAppl {
	if numCallArgs <= 0 {
		_, _, numCallArgs, _, _, _ = dissectCall(callExpr)
	}
	idx, rewrite := numCallArgs-1, len(argIdxs) == 0
	if !rewrite { // then rewrite = argIdxs.contains(idx) ... in Go:
		for _, argidx := range argIdxs {
			if rewrite = (argidx == idx); rewrite {
				break
			}
		}
	}
	if rewrite {
		callExpr.Arg = rewriter(idx, callExpr.Arg)
	}
	if subcall, ok := callExpr.Callee.(ExprAppl); ok && idx > 0 {
		callExpr.Callee = rewriteCallArgs(subcall, numCallArgs-1, rewriter, argIdxs)
	}
	return callExpr
}

func rewriteInnerMostCallee(expr ExprAppl, rewriter func(Expr) Expr) ExprAppl {
	if calleecall, ok := expr.Callee.(ExprAppl); ok {
		expr.Callee = rewriteInnerMostCallee(calleecall, rewriter)
	} else {
		expr.Callee = rewriter(expr.Callee)
	}
	return expr
}

func tryEval(prog Prog, expr Expr, checkForArgRefs bool) (ret Expr) {
	ret = expr
	if call, ok := expr.(ExprAppl); ok {
		caneval := true
		if checkForArgRefs {
			_ = walk(call, func(it Expr) Expr {
				_, isargref := it.(ExprArgRef)
				caneval = caneval && !isargref
				return it
			})
		}
		if caneval {
			defer func() {
				if recover() != nil {
					ret = expr
				}
			}()
			ret = walk(ret, func(it Expr) Expr {
				return prog.Eval(it, nil)
			})
		}
	}
	return
}

func walk(expr Expr, visitor func(Expr) Expr) Expr {
	if ret := visitor(expr); ret != nil {
		expr = ret
		if call, ok := expr.(ExprAppl); ok {
			call.Callee, call.Arg = walk(call.Callee, visitor), walk(call.Arg, visitor)
			expr = call
		}
	}
	return expr
}
