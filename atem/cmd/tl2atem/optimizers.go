package main

import (
	. "github.com/metaleap/atmo/atem"
)

type exprNever struct{}

func (exprNever) JsonSrc() string { return "-0" }

var optNumRounds int

func walk(expr Expr, visitor func(Expr) Expr) Expr {
	expr = visitor(expr)
	if call, ok := expr.(ExprAppl); ok {
		call.Callee, call.Arg = walk(call.Callee, visitor), walk(call.Arg, visitor)
		expr = call
	}
	return expr
}

func optimize(src Prog) (ret Prog, didModify bool) {
	ret = src
	for again := true; again; {
		again, optNumRounds = false, optNumRounds+1
		for _, opt := range []func(Prog) (Prog, bool){
			optimize_inlineNullaries,
			optimize_ditchUnusedFuncDefs,
			optimize_inlineSelectorCalls,
			optimize_argDropperCalls,
			optimize_inlineArgCallers,
			optimize_inlineArgsRearrangers,
			optimize_primOpPreCalcs,
			optimize_callsToGeqOrLeq,
			optimize_minifyNeedlesslyElaborateBoolOpCalls,
		} {
			if ret, again = opt(ret); again {
				didModify = true
				break
			}
		}
	}
	return
}

// inliners or other optimizers may result in now-unused func-defs, here's a single routine that'll remove them
func optimize_ditchUnusedFuncDefs(src Prog) (ret Prog, didModify bool) {
	ret = src
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

// nullary func-defs get first `Eval`'d right here, then the result inlined at use sites
func optimize_inlineNullaries(src Prog) (ret Prog, didModify bool) {
	ret = src
	nullaries := make(map[int]bool)
	for i := int(StdFuncCons + 1); i < len(ret)-1; i++ {
		if 0 == len(ret[i].Args) {
			ret[i].Body = ret.Eval(ret[i].Body, nil)
			nullaries[i] = true
		}
	}
	for i := int(StdFuncCons + 1); i < len(ret); i++ {
		ret[i].Body = walk(ret[i].Body, func(expr Expr) Expr {
			if fnref, ok := expr.(ExprFuncRef); ok && fnref > StdFuncCons && nullaries[int(fnref)] {
				didModify = true
				return ret[int(fnref)].Body
			}
			return expr
		})
	}
	return
}

// in calls that provide known-to-be-discarded args, those are replaced with zero (neatly rendered as -0 in JSON output)
func optimize_argDropperCalls(src Prog) (ret Prog, didModify bool) {
	ret = src
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
						return rewriteCallArgs(expr.(ExprAppl), numargs, func(argidx int, argval Expr) Expr {
							if !eq(argval, exprNever{}) {
								argval, didModify = exprNever{}, true
							}
							return argval
						}, argdrops)
					}
				}
				return expr
			})
		}
	}
	return
}

func optimize_inlineArgCallers(src Prog) (ret Prog, didModify bool) {
	ret = src
	argcallers := make(map[int]ExprArgRef, 8)
	for i := 0; i < len(ret)-1; i++ {
		if callee, _, numargs, numargscalls, numargrefs, _ := dissectCall(ret[i].Body); numargs > 0 && numargrefs == 1 && numargscalls == 0 {
			if argref, ok := callee.(ExprArgRef); ok { // only 1 arg-ref in body and its the inner-most callee?
				argcallers[i] = (-argref) - 1
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
							call2inline := rewriteInnerMostCallee(ret[*fnref].Body.(ExprAppl), func(Expr) Expr { return allargs[0] })
							expr = rewriteInnerMostCallee(expr.(ExprAppl), func(Expr) Expr { return StdFuncId })
							expr = rewriteCallArgs(expr.(ExprAppl), numargs, func(argidx int, argval Expr) Expr {
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

func optimize_inlineSelectorCalls(src Prog) (ret Prog, didModify bool) {
	ret = src
	selectors := make(map[int]ExprArgRef, 8)
	for i := 0; i < len(ret)-1; i++ {
		if argref, ok := ret[i].Body.(ExprArgRef); ok {
			selectors[i] = (-argref) - 1
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

func optimize_inlineArgsRearrangers(src Prog) (ret Prog, didModify bool) {
	ret = src
	rearrangers := make(map[int]bool, 8)
	for i := 0; i < len(ret)-1; i++ {
		if callee, _, _, numargcalls, numargrefs, allargs := dissectCall(ret[i].Body); len(allargs) > 0 && numargcalls == 0 && numargrefs > 0 {
			callparts := append([]Expr{callee}, allargs...)
			for _, expr := range callparts {
				if _, ok := expr.(ExprArgRef); ok {
					rearrangers[i] = true
					break
				}
			}
		}
	}
	if len(rearrangers) == 0 {
		return
	}
	for i := int(StdFuncCons + 1); i < len(ret); i++ {
		ret[i].Body = walk(ret[i].Body, func(expr Expr) Expr {
			if _, fnref, numargs, _, _, allargs := dissectCall(expr); fnref != nil { // we have a call...
				if rearrangers[int(*fnref)] && numargs == len(ret[*fnref].Args) { // to a rearranger and with proper number of args
					counts := map[ExprArgRef]int{}
					nuexpr := rewriteCallArgs(ret[*fnref].Body.(ExprAppl), -1, func(argidx int, argval Expr) Expr {
						if argref, ok := argval.(ExprArgRef); ok {
							argref = (-argref) - 1
							counts[argref] = counts[argref] + 1
							return allargs[argref]
						}
						return argval
					}, nil)
					nuexpr = rewriteInnerMostCallee(nuexpr, func(callee Expr) Expr {
						if argref, ok := callee.(ExprArgRef); ok {
							argref = (-argref) - 1
							counts[argref] = counts[argref] + 1
							return allargs[argref]
						}
						return callee
					})
					for argref, count := range counts {
						if _, iscall := allargs[argref].(ExprAppl); iscall && count > 1 {
							return expr
						}
					}
					expr, didModify = nuexpr, true
				}
			}
			return expr
		})
	}
	return
}

// from `bexpr True False` to `bexpr`, from `(bexpr False True) foo bar` (aka `not`) to `bexpr bar foo`
func optimize_minifyNeedlesslyElaborateBoolOpCalls(src Prog) (ret Prog, didModify bool) {
	ret = src
	for i := int(StdFuncCons + 1); i < len(ret); i++ {
		ret[i].Body = walk(ret[i].Body, func(expr Expr) Expr {
			if _, fnref, numargs, _, _, allargs := dissectCall(expr); fnref != nil && (numargs == 4 || numargs == 6) {
				if opcode := OpCode(*fnref); opcode == OpEq || opcode == OpLt || opcode == OpGt {
					if fnl, _ := allargs[2].(ExprFuncRef); fnl == StdFuncTrue || fnl == StdFuncFalse {
						if fnr, _ := allargs[3].(ExprFuncRef); fnr == StdFuncTrue || fnr == StdFuncFalse {
							if numargs == 4 && fnl == StdFuncTrue && fnr == StdFuncFalse {
								didModify = true
								return ExprAppl{Callee: ExprAppl{Callee: *fnref, Arg: allargs[0]}, Arg: allargs[1]}
							} else if numargs == 6 && fnl == StdFuncFalse && fnr == StdFuncTrue {
								didModify = true
								return ExprAppl{Callee: ExprAppl{Callee: ExprAppl{Callee: ExprAppl{Callee: *fnref, Arg: allargs[0]}, Arg: allargs[1]}, Arg: allargs[5]}, Arg: allargs[4]}
							}
						}
					}
				}
			}
			return expr
		})
	}
	return
}

// ie. from foo>=1 to foo>0 etc.
func optimize_callsToGeqOrLeq(src Prog) (ret Prog, didModify bool) {
	ret = src
	geqsleqs := map[int]bool{}
	for i := int(StdFuncCons + 1); i < len(ret)-1; i++ {
		if _, fnr1, _, _, _, allargs1 := dissectCall(ret[i].Body); fnr1 != nil && len(allargs1) == 4 {
			if opcode1 := OpCode(*fnr1); opcode1 == OpEq || opcode1 == OpLt || opcode1 == OpGt {
				if fntrue, _ := allargs1[2].(ExprFuncRef); fntrue == StdFuncTrue {
					if _, fnr2, _, _, _, allargs2 := dissectCall(allargs1[3]); fnr2 != nil && len(allargs2) == 2 {
						if opcode2 := OpCode(*fnr2); opcode2 != opcode1 && (opcode2 == OpEq || opcode2 == OpLt || opcode2 == OpGt) && eq(allargs1[0], allargs2[0]) && eq(allargs1[1], allargs2[1]) && (opcode1 == OpEq || opcode2 == OpEq) {
							if isl, isg := (opcode1 == OpLt || opcode2 == OpLt), (opcode1 == OpGt || opcode2 == OpGt); isg || isl {
								geqsleqs[i] = isg
							}
						}
					}
				}
			}
		}
	}
	if len(geqsleqs) > 0 {
		for i := int(StdFuncCons + 1); i < len(ret); i++ {
			ret[i].Body = walk(ret[i].Body, func(expr Expr) Expr {
				if _, fnref, _, _, _, allargs := dissectCall(expr); fnref != nil && (len(allargs) == 1 || len(allargs) == 2) {
					numl, isnuml := allargs[0].(ExprNumInt)
					numr, isnumr := allargs[len(allargs)-1].(ExprNumInt)
					if isnuml || isnumr {
						if isgeq, orleq := geqsleqs[int(*fnref)]; orleq {
							if len(allargs) == 1 && isnuml {
								if didModify = true; isgeq {
									return ExprAppl{Callee: ExprFuncRef(OpGt), Arg: numl + 1}
								} else {
									return ExprAppl{Callee: ExprFuncRef(OpLt), Arg: numl - 1}
								}
							} else if len(allargs) == 2 {
								if isnuml {
									if didModify = true; isgeq {
										return ExprAppl{Callee: ExprAppl{Callee: ExprFuncRef(OpGt), Arg: numl + 1}, Arg: allargs[1]}
									} else {
										return ExprAppl{Callee: ExprAppl{Callee: ExprFuncRef(OpLt), Arg: numl - 1}, Arg: allargs[1]}
									}
								} else if isnumr {
									if didModify = true; isgeq {
										return ExprAppl{Callee: ExprAppl{Callee: ExprFuncRef(OpGt), Arg: allargs[0]}, Arg: numr - 1}
									} else {
										return ExprAppl{Callee: ExprAppl{Callee: ExprFuncRef(OpLt), Arg: allargs[0]}, Arg: numr + 1}
									}
								}
							}
						}
					}
				}
				return expr
			})
		}
	}
	return
}

// 0+foo, 1*foo, foo-0, foo/1 etc..
func optimize_primOpPreCalcs(src Prog) (ret Prog, didModify bool) {
	ret = src
	for i := int(StdFuncCons + 1); i < len(ret); i++ {
		ret[i].Body = walk(ret[i].Body, func(expr Expr) Expr {
			if _, fnref, numargs, _, _, allargs := dissectCall(expr); fnref != nil && *fnref < 0 {
				if opcode := OpCode(*fnref); numargs == 1 {
					if opcode == OpAdd && eq(allargs[0], ExprNumInt(0)) {
						didModify = true
						return StdFuncId
					} else if opcode == OpMul && eq(allargs[0], ExprNumInt(1)) {
						didModify = true
						return StdFuncId
					}
				} else if numargs == 2 {
					if opcode == OpAdd && eq(allargs[0], ExprNumInt(0)) {
						didModify = true
						return allargs[1]
					} else if opcode == OpAdd && eq(allargs[1], ExprNumInt(0)) {
						didModify = true
						return allargs[0]
					} else if opcode == OpMul && eq(allargs[0], ExprNumInt(1)) {
						didModify = true
						return allargs[1]
					} else if opcode == OpMul && eq(allargs[1], ExprNumInt(1)) {
						didModify = true
						return allargs[0]
					} else if opcode == OpSub && eq(allargs[1], ExprNumInt(0)) {
						didModify = true
						return allargs[0]
					} else if opcode == OpDiv && eq(allargs[1], ExprNumInt(1)) {
						didModify = true
						return allargs[0]
					} else if opcode == OpMod && eq(allargs[1], ExprNumInt(1)) {
						didModify = true
						return ExprNumInt(0)
					} else if opcode == OpEq && eq(allargs[0], allargs[1]) {
						didModify = true
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

func eq(expr Expr, cmp Expr) bool {
	switch it := expr.(type) {
	case exprNever:
		_, ok := cmp.(exprNever)
		return ok
	case ExprNumInt:
		that, ok := cmp.(ExprNumInt)
		return ok && it == that
	case ExprArgRef:
		that, ok := cmp.(ExprArgRef)
		return ok && (it == that || (it < 0 && that >= 0 && that == (-it)-1) || (it >= 0 && that < 0 && it == (-that)-1))
	case ExprFuncRef:
		that, ok := cmp.(ExprFuncRef)
		return ok && it == that
	case ExprAppl:
		that, ok := cmp.(ExprAppl)
		return ok && eq(it.Callee, that.Callee) && eq(it.Arg, that.Arg)
	}
	return false
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
