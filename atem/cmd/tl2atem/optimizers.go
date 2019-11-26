package main

import (
	. "github.com/metaleap/atmo/atem"
)

const never ExprNumInt = -1 << 31

var optNumRounds int

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
		again, optNumRounds = false, optNumRounds+1
		for _, opt := range []func(Prog) (Prog, bool){
			optimize_inlineSelectorCalls,
			optimize_dropUnused,
			optimize_inlineNullaries,
			optimize_saturateArgsIfPartialCall,
			optimize_argDropperCalls,
			optimize_inlineArgCallers,
			optimize_inlineArgsRearrangers,
			optimize_primOpPreCalcs,
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
	nullaries := make(map[int]int)
	for i := int(StdFuncCons + 1); i < len(ret)-1; i++ {
		if 0 == len(ret[i].Args) {
			nullaries[i] = 0
		}
	}
	for i := int(StdFuncCons + 1); i < len(ret); i++ {
		_ = walk(ret[i].Body, func(expr Expr) Expr {
			if fnref, ok := expr.(ExprFuncRef); ok && fnref > StdFuncCons {
				if numrefs, is := nullaries[int(fnref)]; is {
					nullaries[int(fnref)] = numrefs + 1
				}
			}
			return expr
		})
	}
	for i := int(StdFuncCons + 1); i < len(ret); i++ {
		ret[i].Body = walk(ret[i].Body, func(expr Expr) Expr {
			if fnref, ok := expr.(ExprFuncRef); ok && fnref > StdFuncCons {
				if numrefs, is := nullaries[int(fnref)]; is {
					if _, iscall := ret[fnref].Body.(ExprCall); (!iscall) || numrefs <= 1 {
						didModify = true
						return ret[int(fnref)].Body
					}
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
		if _, fnref, numargs, numargscalls, numargrefs, _ := dissectCall(ret[i].Body); fnref != nil {
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
					if _, fnref, _, _, _, _ := dissectCall(expr); fnref != nil && int(*fnref) == doidx {
						return rewriteInnerMostCallee(expr.(ExprCall), func(Expr) Expr { return ret[doidx].Body })
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

func optimize_argDropperCalls(prog Prog) (ret Prog, didModify bool) {
	ret = prog
	argdroppers := make(map[int][]int, 8)
	for i := 0; i < len(ret)-1; i++ {
		var argdrops []int
		for argidx, argusage := range ret[i].Args {
			if argusage == 0 {
				argdrops = append(argdrops, argidx)
			}
		}
		if len(argdrops) > 0 {
			argdroppers[i] = argdrops
		}
	}
	if len(argdroppers) > 0 {
		for i := int(StdFuncCons + 1); i < len(ret); i++ {
			ret[i].Body = walk(ret[i].Body, func(expr Expr) Expr {
				if _, fnref, numargs, _, _, _ := dissectCall(expr); fnref != nil {
					if argdrops := argdroppers[int(*fnref)]; len(argdrops) > 0 {
						for _, argidx := range argdrops {
							if argidx >= numargs {
								return expr
							}
						}
						return rewriteCallArgs(expr.(ExprCall), numargs, func(argidx int, argval Expr) Expr {
							return never
						}, argdrops)
					}
				}
				return expr
			})
		}
	}
	return
}

func optimize_inlineArgCallers(prog Prog) (ret Prog, didModify bool) {
	ret = prog
	argcallers := make(map[int]ExprArgRef, 8)
	for i := 0; i < len(ret)-1; i++ {
		if caller, _, numargs, numargscalls, numargrefs, _ := dissectCall(ret[i].Body); numargs > 0 && numargrefs == 1 && numargscalls == 0 {
			if argref, ok := caller.(ExprArgRef); ok {
				argcallers[i] = argref
			}
		}
	}
	if len(argcallers) > 0 {
		for i := int(StdFuncCons + 1); i < len(ret); i++ {
			ret[i].Body = walk(ret[i].Body, func(expr Expr) Expr {
				if _, fnref, numargs, _, _, allargs := dissectCall(expr); fnref != nil {
					if argref, ok := argcallers[int(*fnref)]; ok {
						if argref != 0 {
							panic("TODO")
						} else {
							call2inline := rewriteInnerMostCallee(ret[*fnref].Body.(ExprCall), func(Expr) Expr { return allargs[0] })
							expr = rewriteInnerMostCallee(expr.(ExprCall), func(Expr) Expr { return StdFuncId })
							expr = rewriteCallArgs(expr.(ExprCall), numargs, func(argidx int, argval Expr) Expr {
								if argidx == 0 {
									return call2inline
								}
								return argval
							}, []int{0})
							didModify = true
						}
					}
				}
				return expr
			})
		}
	}
	return
}

func optimize_inlineSelectorCalls(prog Prog) (ret Prog, didModify bool) {
	ret = prog
	selectors := make(map[int]ExprArgRef, 8)
	for i := 0; i < len(ret)-1; i++ {
		if argref, ok := ret[i].Body.(ExprArgRef); ok {
			selectors[i] = argref
		}
	}
	if len(selectors) > 0 {
		for i := int(StdFuncCons + 1); i < len(ret); i++ {
			ret[i].Body = walk(ret[i].Body, func(expr Expr) Expr {
				if _, fnref, numargs, _, _, allargs := dissectCall(expr); fnref != nil {
					if argref, ok := selectors[int(*fnref)]; ok && int(argref) < numargs && numargs == len(ret[*fnref].Args) {
						didModify = true
						return allargs[argref]
					}
				}
				return expr
			})
		}
	}
	return
}

func optimize_inlineArgsRearrangers(prog Prog) (ret Prog, didModify bool) {
	ret = prog
	rearrangers := make(map[int]bool, 8)
	for i := 0; i < len(ret)-1; i++ {
		if callee, _, _, numargcalls, numargrefs, allargs := dissectCall(ret[i].Body); numargcalls == 0 && numargrefs > 0 {
			maxarg, callexprs := ExprArgRef(-1), append([]Expr{callee}, allargs...)
			for _, expr := range callexprs {
				if argref, ok := expr.(ExprArgRef); ok {
					if argref > maxarg {
						maxarg = argref
					}
				}
			}
			if maxarg > -1 {
				rearrangers[i] = true
			}
		}
	}
	if len(rearrangers) > 0 {
		for i := int(StdFuncCons + 1); i < len(ret); i++ {
			ret[i].Body = walk(ret[i].Body, func(expr Expr) Expr {
				if _, fnref, numargs, _, _, allargs := dissectCall(expr); fnref != nil {
					if rearrangers[int(*fnref)] && numargs == len(ret[*fnref].Args) {
						counts := map[ExprArgRef]int{}
						nuexpr := rewriteCallArgs(ret[*fnref].Body.(ExprCall), -1, func(argidx int, argval Expr) Expr {
							if argref, ok := argval.(ExprArgRef); ok {
								counts[argref] = counts[argref] + 1
								return allargs[argref]
							}
							return argval
						}, nil)
						nuexpr = rewriteInnerMostCallee(nuexpr, func(callee Expr) Expr {
							if argref, ok := callee.(ExprArgRef); ok {
								counts[argref] = counts[argref] + 1
								return allargs[argref]
							}
							return callee
						})
						for argref, count := range counts {
							if _, iscall := allargs[argref].(ExprCall); iscall && count > 1 {
								return expr
							}
						}
						expr, didModify = nuexpr, true
					}
				}
				return expr
			})
		}
	}
	return
}

func optimize_primOpPreCalcs(prog Prog) (ret Prog, didModify bool) {
	ret = prog
	for i := int(StdFuncCons + 1); i < len(ret); i++ {
		ret[i].Body = walk(ret[i].Body, func(expr Expr) Expr {
			if _, fnref, numargs, _, _, allargs := dissectCall(expr); fnref != nil && *fnref < 0 {
				if opcode := OpCode(*fnref); numargs == 1 {
					if didModify = (opcode == OpAdd && eq(allargs[0], ExprNumInt(0))); didModify {
						return StdFuncId
					} else if didModify = (opcode == OpMul && eq(allargs[0], ExprNumInt(1))); didModify {
						return StdFuncId
					}
				} else if numargs == 2 {
					if didModify = (opcode == OpAdd && eq(allargs[0], ExprNumInt(0))); didModify {
						return allargs[1]
					} else if didModify = (opcode == OpAdd && eq(allargs[1], ExprNumInt(0))); didModify {
						return allargs[0]
					} else if didModify = (opcode == OpMul && eq(allargs[0], ExprNumInt(1))); didModify {
						return allargs[1]
					} else if didModify = (opcode == OpMul && eq(allargs[1], ExprNumInt(1))); didModify {
						return allargs[0]
					} else if didModify = (opcode == OpSub && eq(allargs[1], ExprNumInt(0))); didModify {
						return allargs[0]
					} else if didModify = (opcode == OpDiv && eq(allargs[1], ExprNumInt(1))); didModify {
						return allargs[0]
					} else if didModify = (opcode == OpMod && eq(allargs[1], ExprNumInt(1))); didModify {
						return ExprNumInt(0)
					} else if didModify = (opcode == OpEq && eq(allargs[0], allargs[1])); didModify {
						return StdFuncTrue
					}
				}
			}
			return expr
		})
	}
	return
}

func dissectCall(expr Expr) (innerMostCallee Expr, innerMostCalleeFnRef *ExprFuncRef, numCallArgs int, numCallArgsThatAreCalls int, numArgRefs int, allArgs []Expr) {
	for call, okc := expr.(ExprCall); okc; call, okc = call.Callee.(ExprCall) {
		innerMostCallee, numCallArgs, allArgs = call.Callee, numCallArgs+1, append([]Expr{call.Arg}, allArgs...)
		if _, isargcall := call.Arg.(ExprCall); isargcall {
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

func eq(expr Expr, cmp Expr) bool {
	switch it := expr.(type) {
	case ExprNumInt:
		that, ok := cmp.(ExprNumInt)
		return ok && it == that
	case ExprArgRef:
		that, ok := cmp.(ExprArgRef)
		return ok && it == that
	case ExprFuncRef:
		that, ok := cmp.(ExprFuncRef)
		return ok && it == that
	case ExprCall:
		that, ok := cmp.(ExprCall)
		return ok && eq(it.Callee, that.Callee) && eq(it.Arg, that.Arg)
	}
	return false
}

func rewriteCallArgs(callExpr ExprCall, numArgs int, rewriter func(int, Expr) Expr, argIdxs []int) ExprCall {
	if numArgs <= 0 {
		_, _, numArgs, _, _, _ = dissectCall(callExpr)
	}
	idx, rewrite := numArgs-1, len(argIdxs) == 0
	if !rewrite {
		for _, argidx := range argIdxs {
			if rewrite = (argidx == idx); rewrite {
				break
			}
		}
	}
	if rewrite {
		callExpr.Arg = rewriter(idx, callExpr.Arg)
	}
	if subcall, ok := callExpr.Callee.(ExprCall); ok && idx > 0 {
		callExpr.Callee = rewriteCallArgs(subcall, numArgs-1, rewriter, argIdxs)
	}
	return callExpr
}

func rewriteInnerMostCallee(expr ExprCall, rewriter func(Expr) Expr) ExprCall {
	if calleecall, ok := expr.Callee.(ExprCall); ok {
		expr.Callee = rewriteInnerMostCallee(calleecall, rewriter)
	} else {
		expr.Callee = rewriter(expr.Callee)
	}
	return expr
}
