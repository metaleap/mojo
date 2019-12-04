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
	for i := StdFuncCons + 1; int(i) < len(prog) && !doesHaveNonCalleeOccurrences; i++ {
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

func tryEval(prog Prog, expr Expr, checkForArgRefs bool) (ret Expr) {
	ret = expr
	if call, ok := expr.(exprAppl); ok {
		ok2try := true
		if checkForArgRefs {
			_ = walk(call, func(it Expr) Expr {
				_, isargref := it.(ExprArgRef)
				ok2try = ok2try && !isargref
				return it
			})
		}
		if ok2try {
			defer func() {
				if recover() != nil {
					ret = expr
				}
			}()
			var convto, convfrom func(Expr) Expr
			convto = func(expr Expr) Expr {
				if call, ok := expr.(exprAppl); ok {
					return &ExprCall{Callee: convto(call.Callee), Args: []Expr{convto(call.Arg)}}
				}
				return expr
			}
			convfrom = func(expr Expr) Expr {
				if call, _ := expr.(*ExprCall); call != nil {
					expr = convfrom(call.Callee)
					for i := len(call.Args) - 1; i > -1; i-- {
						expr = exprAppl{Callee: expr, Arg: convfrom(call.Args[i])}
					}
					// return exprAppl{Callee: convfrom(call.Callee), Arg: convfrom(call.Arg)}
				}
				return expr
			}
			var hasargref bool // since interpreter is graph-rewriting, instantiating func bodies without a fully filled stack: reject such bodies as a sensible pre-eval result. otherwise, for example list creations would turn into the full, tiny, useless body of Cons etc.
			ret = convfrom(prog.Eval(convto(ret)))
			_ = walk(ret, func(it Expr) Expr {
				_, isargref := it.(ExprArgRef)
				hasargref = hasargref || isargref
				return it
			})
			if hasargref {
				ret = expr
			}
		}
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
