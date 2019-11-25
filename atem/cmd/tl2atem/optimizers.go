package main

import (
	. "github.com/metaleap/atmo/atem"
)

func walk(expr Expr, visitor func(Expr) Expr) Expr {
	expr = visitor(expr)
	if call, ok := expr.(ExprCall); ok {
		call.Callee, call.Arg = walk(call.Callee, visitor), walk(call.Arg, visitor)
		expr = call
	}
	return expr
}

func optimize(prog Prog) (ret Prog, didModify bool) {
	ret = prog
	for again := true; again; {
		again = false
		for _, opt := range []func(Prog) (Prog, bool){
			optimize_dropUnused,
			optimize_inlineNullaries,
			optimize_rewriteIdCalls,
			optimize_saturateArgsIfPartialCall,
		} {
			if ret, again = opt(ret); again {
				didModify = true
				break
			}
		}
	}
	return
}

func optimize_dropUnused(prog Prog) (ret Prog, didModify bool) {
	ret = prog
	defrefs := make(map[int]bool, len(ret))
	for i := range ret {
		walk(ret[i].Body, func(expr Expr) Expr {
			if fnref, ok := expr.(ExprFuncRef); ok && fnref >= 0 {
				defrefs[int(fnref)] = true
			}
			return expr
		})
	}
	for i := int(StdFuncCons + 1); i < len(ret)-1; i++ {
		if hasrefs := defrefs[i]; !hasrefs {
			didModify, ret = true, append(ret[:i], ret[i+1:]...)
			for j := range ret {
				ret[j].Body = walk(ret[j].Body, func(expr Expr) Expr {
					if fnref, ok := expr.(ExprFuncRef); ok && int(fnref) > i {
						return ExprFuncRef(fnref - 1)
					}
					return expr
				})
			}
			break
		}
	}
	return
}

func optimize_inlineNullaries(prog Prog) (ret Prog, didModify bool) {
	ret = prog
	nullaries := make(map[int]bool)
	for i := int(StdFuncCons + 1); i < len(ret)-1; i++ {
		if 0 == len(ret[i].Args) {
			nullaries[i] = true
		}
	}
	for i := int(StdFuncCons + 1); i < len(ret); i++ {
		ret[i].Body = walk(ret[i].Body, func(expr Expr) Expr {
			if fnref, ok := expr.(ExprFuncRef); ok && fnref > StdFuncCons {
				if _, is := nullaries[int(fnref)]; is {
					didModify = true
					return ret[int(fnref)].Body
				}
			}
			return expr
		})
	}
	return
}

func optimize_saturateArgsIfPartialCall(prog Prog) (ret Prog, didModify bool) {
	ret = prog
	var doidx, dodiff int
	var iscomplex bool
	for i := int(StdFuncCons + 1); i < len(ret)-1; i++ {
		if fnref, numargs, numargscalls, numargrefs := dissectCall(ret[i].Body); fnref != nil {
			goalnum := 2
			if fr := *fnref; fr >= 0 {
				goalnum = len(ret[fr].Args)
			}
			if diff := goalnum - numargs; diff > 0 {
				doidx, dodiff, iscomplex = i, diff, numargscalls > 0 || numargrefs > 0
				break
			}
		}
	}
	if didModify = (dodiff > 0); didModify {
		if !iscomplex { // inline the partial (if simple) into calls before finally modifying the def
			for i := 0; i < len(ret); i++ {
				ret[i].Body = walk(ret[i].Body, func(expr Expr) Expr {
					if fnref, _, _, _ := dissectCall(expr); fnref != nil && int(*fnref) == doidx {
						return rewriteInnerMostCallee(expr.(ExprCall), ret[doidx].Body)
					}
					return expr
				})
			}
		}
		for j := 0; j < dodiff; j++ {
			ret[doidx].Args, ret[doidx].Body = append(ret[doidx].Args, 1), ExprCall{Callee: ret[doidx].Body, Arg: ExprArgRef(j)}
		}
	}
	return
}

func optimize_rewriteIdCalls(prog Prog) (ret Prog, didModify bool) {
	ret = prog
	for i := int(StdFuncCons + 1); i < len(ret); i++ {
		ret[i].Body = walk(ret[i].Body, func(expr Expr) Expr {
			if call, ok := expr.(ExprCall); ok {
				if fnref, ok := call.Callee.(ExprFuncRef); ok && fnref == 0 {
					didModify = true
					return call.Arg
				}
			}
			return expr
		})
	}
	return
}

func dissectCall(expr Expr) (fnRef *ExprFuncRef, numCallArgs int, numCallArgsThatAreCalls int, numArgRefs int) {
	for call, ok := expr.(ExprCall); ok; call, ok = call.Callee.(ExprCall) {
		numCallArgs++
		if _, isargcall := call.Arg.(ExprCall); isargcall {
			numCallArgsThatAreCalls++
		} else if _, isargref := call.Arg.(ExprArgRef); isargref {
			numArgRefs++
		}
		if fnref, okf := call.Callee.(ExprFuncRef); okf {
			fnRef = &fnref
		} else if _, isargref := call.Callee.(ExprArgRef); isargref {
			numArgRefs++
		}
	}
	return
}

func rewriteInnerMostCallee(expr ExprCall, rewriteWith Expr) Expr {
	if calleecall, ok := expr.Callee.(ExprCall); ok {
		expr.Callee = rewriteInnerMostCallee(calleecall, rewriteWith)
	} else {
		expr.Callee = rewriteWith
	}
	return expr
}
