package main

import (
	. "github.com/metaleap/atmo/atem"
)

var exprNever = exprTmp(123456789)

func init() { OpPrtDst = func([]byte) (int, error) { panic("caught in `tryEval`") } }

type exprTmp int

func (me exprTmp) JsonSrc() string { return "-0" }

type exprAppl struct {
	Callee Expr
	Arg    Expr
}

func (me exprAppl) JsonSrc() string { return "[" + me.Callee.JsonSrc() + ", " + me.Arg.JsonSrc() + "]" }

func convFrom(expr Expr) Expr {
	if call, _ := expr.(*ExprCall); call != nil {
		expr = convFrom(call.Callee)
		for i := len(call.Args) - 1; i > -1; i-- {
			expr = exprAppl{Callee: expr, Arg: convFrom(call.Args[i])}
		}
	}
	return expr
}

func convTo(expr Expr) Expr {
	if call, ok := expr.(exprAppl); ok {
		callee := convTo(call.Callee)
		if subcall, _ := callee.(*ExprCall); subcall != nil {
			subcall.Args = append([]Expr{convTo(call.Arg)}, subcall.Args...)
			return subcall
		} else {
			return &ExprCall{Callee: callee, Args: []Expr{convTo(call.Arg)}}
		}
	}
	return expr
}

func dissectCall(expr Expr) (innerMostCallee Expr, innerMostCalleeFnRef *ExprFuncRef, numCallArgs int, numCallArgsThatAreCalls int, numArgRefs int, allArgs []Expr) {
	for call, okc := expr.(exprAppl); okc; call, okc = call.Callee.(exprAppl) {
		innerMostCallee, numCallArgs, allArgs = call.Callee, numCallArgs+1, append([]Expr{call.Arg}, allArgs...)
		if _, isargcall := call.Arg.(exprAppl); isargcall {
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
		} else if call, isc := expr.(exprAppl); isc {
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
	for i := 0; i < len(prog) && !doesHaveNonCalleeOccurrences; i++ {
		_ = walk(prog[i].Body, scrut)
	}
	return
}

func eq(prog Prog, expr Expr, cmp Expr) bool {
	if t1, okt1 := expr.(exprTmp); okt1 {
		t2, okt2 := cmp.(exprTmp)
		return okt2 && t1 == t2
	} else if a1, oka1 := expr.(exprAppl); oka1 {
		a2, oka2 := cmp.(exprAppl)
		return oka2 && eq(prog, a1.Callee, a2.Callee) && eq(prog, a1.Arg, a2.Arg)
	}
	return prog.Eq(expr, cmp)
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
				argref = (-argref) - 1
				prog[i].Args[argref] = 1 + prog[i].Args[argref]
			}
			return expr
		})
	}
	return prog
}

func rewriteCallArgs(callExpr exprAppl, numCallArgs int, rewriter func(int, Expr) Expr, argIdxs []int) exprAppl {
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
	if subcall, ok := callExpr.Callee.(exprAppl); ok && idx > 0 {
		callExpr.Callee = rewriteCallArgs(subcall, numCallArgs-1, rewriter, argIdxs)
	}
	return callExpr
}

func rewriteInnerMostCallee(expr exprAppl, rewriter func(Expr) Expr) exprAppl {
	if calleecall, ok := expr.Callee.(exprAppl); ok {
		expr.Callee = rewriteInnerMostCallee(calleecall, rewriter)
	} else {
		expr.Callee = rewriter(expr.Callee)
	}
	return expr
}

func tryEval(prog Prog, expr Expr, preCheckForArgRefs bool) (ret Expr) {
	ret = expr
	if _, ok := expr.(exprAppl); ok {
		checkforargrefs := func() {
			_ = walk(ret, func(it Expr) Expr {
				if _, isargref := it.(ExprArgRef); isargref {
					panic(it) // `recover`ed, see below
				}
				return it
			})
		}
		defer func() {
			if recover() != nil {
				ret = expr
			}
		}()
		if preCheckForArgRefs {
			checkforargrefs()
		}
		ret = prog.Eval(convTo(ret))
		ret = convFrom(ret)
		checkforargrefs()
	}
	return
}

func walk(expr Expr, visitor func(Expr) Expr) Expr {
	if ret := visitor(expr); ret != nil {
		expr = ret
		if call, ok := expr.(exprAppl); ok {
			call.Callee, call.Arg = walk(call.Callee, visitor), walk(call.Arg, visitor)
			expr = call
		}
	}
	return expr
}
