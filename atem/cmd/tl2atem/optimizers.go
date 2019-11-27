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
			optimize_dropUnused,
			optimize_inlineSelectorCalls,
			optimize_inlineNullaries,
			optimize_argDropperCalls,
			optimize_inlineArgCallers,
			optimize_inlineArgsRearrangers,
			optimize_primOpPreCalcs,
			optimize_callsToGeqOrLeq,
			optimize_minifyNeedlesslyElaborateBoolOpCalls,
			optimize_tryPreReduceCalls,
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
						didModify = true
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

func optimize_minifyNeedlesslyElaborateBoolOpCalls(prog Prog) (ret Prog, didModify bool) {
	ret = prog
	for i := int(StdFuncCons + 1); i < len(ret); i++ {
		ret[i].Body = walk(ret[i].Body, func(expr Expr) Expr {
			if _, fnref, numargs, _, _, allargs := dissectCall(expr); fnref != nil && (numargs == 4 || numargs == 6) {
				if opcode := OpCode(*fnref); opcode == OpEq || opcode == OpLt || opcode == OpGt {
					if fnl, _ := allargs[2].(ExprFuncRef); fnl == StdFuncTrue || fnl == StdFuncFalse {
						if fnr, _ := allargs[3].(ExprFuncRef); fnr == StdFuncTrue || fnr == StdFuncFalse {
							if numargs == 4 && fnl == StdFuncTrue && fnr == StdFuncFalse {
								didModify = true
								return ExprCall{Callee: ExprCall{Callee: *fnref, Arg: allargs[0]}, Arg: allargs[1]}
							} else if numargs == 6 && fnl == StdFuncFalse && fnr == StdFuncTrue {
								didModify = true
								return ExprCall{Callee: ExprCall{Callee: ExprCall{Callee: ExprCall{Callee: *fnref, Arg: allargs[0]}, Arg: allargs[1]}, Arg: allargs[5]}, Arg: allargs[4]}
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

func optimize_tryPreReduceCalls(prog Prog) (ret Prog, didModify bool) {
	ret = prog
	var progwithproperargsformat Prog
	for i := int(StdFuncCons + 1); i < len(ret); i++ {
		ret[i].Body = walk(ret[i].Body, func(expr Expr) Expr {
			if _, fnref, _, _, numargrefs, allargs := dissectCall(expr); fnref != nil && numargrefs == 0 {
				_ = walk(expr, func(it Expr) Expr {
					if _, isargref := it.(ExprArgRef); isargref {
						numargrefs++
					}
					return it
				})
				if argsneeded := 2; numargrefs == 0 {
					if *fnref >= 0 {
						argsneeded = len(ret[*fnref].Args)
					}
					if len(allargs) == argsneeded {
						retval := func() Expr {
							defer func() { _ = recover() }()
							if progwithproperargsformat == nil { // kinda Ouch approach, but we'll hit it only ever so rarely really.. this whole thing more for completeness' sake and the occasional "write-time readability gain"
								progwithproperargsformat = LoadFromJson([]byte(ret.String()))
							}
							return progwithproperargsformat.Eval(expr, nil)
						}()
						if _, isnonatomic := retval.(ExprCall); retval != nil && !isnonatomic {
							didModify = true
							return retval
						}
					}
				}
			}
			return expr
		})
	}
	return
}

func optimize_callsToGeqOrLeq(prog Prog) (ret Prog, didModify bool) {
	ret = prog
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
									return ExprCall{Callee: ExprFuncRef(OpGt), Arg: numl + 1}
								} else {
									return ExprCall{Callee: ExprFuncRef(OpLt), Arg: numl - 1}
								}
							} else if len(allargs) == 2 {
								if isnuml {
									if didModify = true; isgeq {
										return ExprCall{Callee: ExprCall{Callee: ExprFuncRef(OpGt), Arg: numl + 1}, Arg: allargs[1]}
									} else {
										return ExprCall{Callee: ExprCall{Callee: ExprFuncRef(OpLt), Arg: numl - 1}, Arg: allargs[1]}
									}
								} else if isnumr {
									if didModify = true; isgeq {
										return ExprCall{Callee: ExprCall{Callee: ExprFuncRef(OpGt), Arg: allargs[0]}, Arg: numr - 1}
									} else {
										return ExprCall{Callee: ExprCall{Callee: ExprFuncRef(OpLt), Arg: allargs[0]}, Arg: numr + 1}
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

func optimize_primOpPreCalcs(prog Prog) (ret Prog, didModify bool) {
	ret = prog
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
