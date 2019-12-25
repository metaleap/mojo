package main

import (
	. "github.com/metaleap/atmo/atem"
)

func optimize(src Prog) Prog {
	for again := true; again; {
		again, src = false, fixFuncDefArgsUsageNumbers(src)
		for _, opt := range []func(Prog) (Prog, bool){
			optimize_ditchUnusedFuncDefs,
			optimize_ditchDuplicateDefs,
			optimize_inlineNaryFuncAliases,
			optimize_inlineCallsToArgRefFuncs,
			optimize_argDropperCalls,
			optimize_inlineArgCallers,
			optimize_inlineArgsRearrangers,
			optimize_primOpPreCalcs,
			optimize_callsToGeqOrLeq,
			optimize_minifyNeedlesslyElaborateBoolOpCalls,
			optimize_inlineOnceCalleds,
			optimize_inlineEverSameArgs,
			optimize_preEvalArgRefLessCalls,
			optimize_inlineNullaries,
		} {
			if src, again = opt(src); again {
				break
			}
		}
	}
	return src
}

// inliners or other optimizers may result in now-unused func-defs, here's a single routine that'll remove them.
// it removes at most one at a time, fixing up all references, then returning with `didModify` of `true`, ensuring another call to find the next one
func optimize_ditchUnusedFuncDefs(src Prog) (ret Prog, didModify bool) {
	ret = src
	defrefs := make(map[int]bool, len(ret))
	for i := range ret {
		_ = walk(ret[i].Body, func(expr Expr) Expr {
			if fnref, ok := expr.(ExprFuncRef); ok && fnref >= 0 && int(fnref) != i {
				defrefs[int(fnref)] = true
			}
			return expr
		})
	}
	for i := int(StdFuncCons + 1); i < len(ret)-1 && !didModify; i++ {
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
		}
	}
	return
}

// lambda-lifting in complex programs may result in multiple structurually equivalent func-defs.
func optimize_ditchDuplicateDefs(src Prog) (ret Prog, didModify bool) {
	ret = src
	dupls := map[ExprFuncRef]ExprFuncRef{}
	for i := StdFuncId; int(i) < len(ret)-1; i++ {
		for j := i + 1; int(j) < len(ret)-1; j++ {
			if len(ret[i].Args) == len(ret[j].Args) && eq(ret[i].Body, ret[j].Body) && j > StdFuncCons {
				if _, have := dupls[j]; !have {
					dupls[j] = i
				}
			}
		}
	}
	if len(dupls) == 0 {
		return
	}
	for i := StdFuncCons + 1; int(i) < len(ret); i++ {
		ret[i].Body = walk(ret[i].Body, func(expr Expr) Expr {
			if fnref, ok := expr.(ExprFuncRef); ok {
				if orig, have := dupls[fnref]; have {
					didModify = true
					return orig
				}
			}
			return expr
		})
	}
	return
}

// lambda-lifting in complex programs may result in lots of func-defs that
// swallow-discard n args to return merely another func-def. refs to those are
// rewritten into calls of `StdFuncTrue` (nested if `n>1`) with said func-def
func optimize_inlineNaryFuncAliases(src Prog) (ret Prog, didModify bool) {
	ret = src
	aliasdefs := map[ExprFuncRef]int{}
	for i := StdFuncCons + 1; int(i) < len(ret)-1; i++ {
		_, okn := ret[i].Body.(ExprNumInt)
		if _, okf := ret[i].Body.(ExprFuncRef); okf || okn {
			aliasdefs[i] = 0
		}
	}
	if len(aliasdefs) == 0 {
		return
	}
	for i := StdFuncCons + 1; int(i) < len(ret); i++ {
		_ = walk(ret[i].Body, func(expr Expr) Expr {
			if fnref, ok := expr.(ExprFuncRef); ok {
				if numrefs, is := aliasdefs[fnref]; numrefs > 0 {
					delete(aliasdefs, fnref)
				} else if is {
					aliasdefs[fnref] = 1
				}
			}
			return expr
		})
	}
	if len(aliasdefs) == 0 {
		return
	}
	for i := StdFuncCons + 1; int(i) < len(ret); i++ {
		ret[i].Body = walk(ret[i].Body, func(expr Expr) Expr {
			if fnref, ok := expr.(ExprFuncRef); ok {
				if numrefs, is := aliasdefs[fnref]; is && numrefs == 1 {
					didModify, expr = true, ret[fnref].Body
					for i := 0; i < len(ret[fnref].Args); i++ {
						expr = exprAppl{Callee: StdFuncTrue, Arg: expr}
					}
				}
			}
			return expr
		})
	}
	return
}

// nullary func-defs get inlined at use sites if: that use site is the only
// reference to the def, or: the def's body is atomic, or: the def's body is a
// call with only atomic args _and_ all use sites are callee positions
func optimize_inlineNullaries(src Prog) (ret Prog, didModify bool) {
	type desc struct {
		numRefs                  int
		numRefsCallees           int
		isCall                   bool
		isCallWithOnlyAtomicArgs bool
	}
	ret = src
	descs := make(map[int]*desc)
	for i := int(StdFuncCons + 1); i < len(ret)-1; i++ {
		if 0 == len(ret[i].Args) {
			_, _, numargs, numargcalls, _, _ := dissectCall(ret[i].Body, nil)
			descs[i] = &desc{isCall: numargs != 0, isCallWithOnlyAtomicArgs: numargs != 0 && numargcalls == 0}
		}
	}
	if len(descs) == 0 {
		return
	}
	chk := func(expr Expr, isCallee bool) {
		if fnref, _ := expr.(ExprFuncRef); fnref > StdFuncCons {
			if desc, isnullary := descs[int(fnref)]; isnullary {
				if isCallee {
					desc.numRefsCallees++
				} else {
					desc.numRefs++
				}
			}
		}
	}
	for i := int(StdFuncCons + 1); i < len(ret); i++ {
		_ = walk(ret[i].Body, func(expr Expr) Expr {
			chk(expr, false)
			if callee, _, numargs, _, _, _ := dissectCall(expr, nil); numargs > 0 {
				chk(callee, true)
			}
			return expr
		})
	}
	for i, desc := range descs {
		if !(desc.numRefs == 1 || (!desc.isCall) ||
			(desc.isCallWithOnlyAtomicArgs && desc.numRefsCallees == desc.numRefs)) {
			delete(descs, i)
		}
	}
	if len(descs) == 0 {
		return
	}
	for i := int(StdFuncCons + 1); i < len(ret); i++ {
		ret[i].Body = walk(ret[i].Body, func(expr Expr) Expr {
			if fnref, _ := expr.(ExprFuncRef); fnref > StdFuncCons {
				if _, isnullary := descs[int(fnref)]; isnullary {
					retexpr := ret[int(fnref)].Body
					didModify = didModify || !eq(expr, retexpr)
					return retexpr
				}
			}
			return expr
		})
	}
	return
}

// in calls that provide known-to-be-discarded args, the latter are replaced with zero (rendered as -0 in JSON output)
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
	if len(argdroppers) == 0 {
		return
	}
	for i := int(StdFuncCons + 1); i < len(ret); i++ {
		ret[i].Body = walk(ret[i].Body, func(expr Expr) Expr {
			if _, fnref, numargs, _, _, _ := dissectCall(expr, nil); fnref != nil {
				if argdrops := argdroppers[int(*fnref)]; len(argdrops) > 0 {
					return rewriteCallArgs(expr.(exprAppl), numargs, func(argidx int, argval Expr) Expr {
						if !eq(argval, exprNever) {
							argval, didModify = exprNever, true
						}
						return argval
					}, argdrops)
				}
			}
			return expr
		})
	}
	return
}

// inlines a FuncDef into a call site if the latter is the only reference to the former, and supplies enough args, and no supplied non-atomic arg is used more than once in the orig def's body
func optimize_inlineOnceCalleds(src Prog) (ret Prog, didModify bool) {
	ret = src
	refs := make(map[ExprFuncRef]map[ExprFuncRef]int, 32)
	for i, l := StdFuncCons+1, ExprFuncRef(len(ret)); i < l; i++ {
		if refs[i] == nil {
			refs[i] = map[ExprFuncRef]int{}
		}
		_ = walk(ret[i].Body, func(expr Expr) Expr {
			if fnref, is := expr.(ExprFuncRef); is && fnref > StdFuncCons {
				m := refs[fnref]
				if m == nil {
					m = map[ExprFuncRef]int{}
				}
				m[i] = m[i] + 1
				refs[fnref] = m
			}
			return expr
		})
	}
	for fn, referencers := range refs {
		if 1 == len(referencers) { // fn referenced only in 1 FuncDef
			for referencer, numrefs := range referencers {
				if numrefs == 1 { // fn referenced only once in 1 FuncDef
					ret[referencer].Body = walk(ret[referencer].Body, func(expr Expr) Expr {
						if _, fnref, _, _, _, allargs := dissectCall(expr, nil); fnref != nil && *fnref == fn && len(allargs) == len(ret[fn].Args) {
							could := true
							for aidx, numargused := range ret[fn].Args {
								if numargused > 1 {
									if _, ok := allargs[aidx].(exprAppl); ok {
										could = false
										break
									}
								}
							}
							if could {
								didModify, expr = true, walk(walk(ret[fn].Body, func(it Expr) Expr {
									if argref, is := it.(ExprArgRef); is {
										return exprTmp(argref)
									}
									return it
								}), func(it Expr) Expr {
									if tmp, is := it.(exprTmp); is && tmp < 0 {
										return allargs[(-tmp)-1]
									}
									return it
								})
							}
						}
						return expr
					})
				}
			}
		}
	}
	return
}

// collects: FuncDefs with body being a call (with only atomic args) to one of its args, the callee being the only arg-ref in the body.
// rewrites: calls to the above
func optimize_inlineArgCallers(src Prog) (ret Prog, didModify bool) {
	ret = src
	argcallers := make(map[int]ExprArgRef, 8)
	for i := 0; i < len(ret)-1; i++ {
		if callee, _, numargs, numargscalls, numargrefs, _ := dissectCall(ret[i].Body, nil); numargs > 0 && numargrefs == 1 && numargscalls == 0 {
			if argref, ok := callee.(ExprArgRef); ok { // only 1 arg-ref in body and its the inner-most callee
				argcallers[i] = (-argref) - 1
			}
		}
	}
	if len(argcallers) == 0 {
		return
	}
	for i := int(StdFuncCons + 1); i < len(ret); i++ {
		ret[i].Body = walk(ret[i].Body, func(expr Expr) Expr {
			if _, fnref, numargs, _, _, allargs := dissectCall(expr, nil); fnref != nil {
				if argref, ok := argcallers[int(*fnref)]; ok && len(allargs) == len(ret[*fnref].Args) {
					if didModify = true; argref != 0 {
						panic("TODO")
					} else {
						call2inline := rewriteInnerMostCallee(ret[*fnref].Body.(exprAppl), func(Expr) Expr { return allargs[0] })
						expr = rewriteInnerMostCallee(expr.(exprAppl), func(Expr) Expr { return StdFuncId })
						expr = rewriteCallArgs(expr.(exprAppl), numargs, func(int, Expr) Expr {
							return call2inline
						}, []int{0})
					}
				}
			}
			return expr
		})
	}
	return
}

// gathers funcs that have arg-ref bodies, well-known examples are id / true / false / nil.
// calls to those (with sufficient args) get replaced by the respective call-arg.
func optimize_inlineCallsToArgRefFuncs(src Prog) (ret Prog, didModify bool) {
	ret = src
	selectors := make(map[int]ExprArgRef, 8)
	for i := 0; i < len(ret)-1; i++ {
		if argref, ok := ret[i].Body.(ExprArgRef); ok {
			selectors[i] = (-argref) - 1
		}
	}
	if len(selectors) == 0 {
		return
	}
	for i := int(StdFuncCons + 1); i < len(ret); i++ {
		ret[i].Body = walk(ret[i].Body, func(expr Expr) Expr {
			if _, fnref, _, _, _, allargs := dissectCall(expr, nil); fnref != nil {
				if argref, ok := selectors[int(*fnref)]; ok && len(allargs) == len(ret[*fnref].Args) {
					didModify = true
					return allargs[argref]
				}
			}
			return expr
		})
	}
	return
}

// collects: FuncDefs with call bodies of only atomic args with at least one arg-ref.
// rewrites: calls to the above with enough args, unless non-atomic args get used more than once
func optimize_inlineArgsRearrangers(src Prog) (ret Prog, didModify bool) {
	ret = src
	rearrangers := make(map[int]int, 8)
	for i := 0; i < len(ret)-1; i++ {
		if _, _, _, numargcalls, numargrefs, allargs := dissectCall(ret[i].Body, func(appl exprAppl) bool {
			_, _, _, numargcalls, numargrefs, _ := dissectCall(appl, nil)
			return numargcalls == 0 && numargrefs == 0
		}); len(allargs) > 0 && numargrefs > 0 && numargcalls == 0 {
			rearrangers[i] = len(allargs)
		}
	}
	if len(rearrangers) == 0 {
		return
	}
	for i := int(StdFuncCons + 1); i < len(ret); i++ {
		ret[i].Body = walk(ret[i].Body, func(expr Expr) Expr {
			if _, fnref, numargs, _, _, allargs := dissectCall(expr, nil); fnref != nil { // we have a call...
				if n := rearrangers[int(*fnref)]; n > 0 && numargs == len(ret[*fnref].Args) { // to a rearranger and with proper number of args
					counts := map[ExprArgRef]int{}
					// take orig body and replace all arg-refs with our call args:
					nuexpr := rewriteCallArgs(ret[*fnref].Body.(exprAppl), n, func(argidx int, argval Expr) Expr {
						if argref, ok := argval.(ExprArgRef); ok {
							argref = (-argref) - 1
							counts[argref] = counts[argref] + 1
							return allargs[argref]
						}
						return argval
					}, nil)
					// same for the callee in the orig body:
					nuexpr = rewriteInnerMostCallee(nuexpr, func(callee Expr) Expr {
						if argref, ok := callee.(ExprArgRef); ok {
							argref = (-argref) - 1
							counts[argref] = counts[argref] + 1
							return allargs[argref]
						}
						return callee
					})
					// abandon IF one of our call-args was another call and is used more than once:
					for argref, count := range counts {
						if _, iscall := allargs[argref].(exprAppl); iscall && count > 1 {
							return expr
						}
					}
					// else, done
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
			if _, fnref, numargs, _, _, allargs := dissectCall(expr, nil); fnref != nil {
				if numargs == 1 && *fnref == StdFuncFalse {
					didModify = true
					return StdFuncId
				} else if numargs == 2 && (*fnref == StdFuncTrue || *fnref == StdFuncNil) {
					didModify = true
					return allargs[0]
				} else if numargs == 4 || numargs == 6 {
					if opcode := OpCode(*fnref); opcode == OpEq || opcode == OpLt || opcode == OpGt {
						if fnl, _ := allargs[2].(ExprFuncRef); fnl == StdFuncTrue || fnl == StdFuncFalse {
							if fnr, _ := allargs[3].(ExprFuncRef); fnr == StdFuncTrue || fnr == StdFuncFalse {
								if numargs == 4 && fnl == StdFuncTrue && fnr == StdFuncFalse {
									didModify = true
									return exprAppl{Callee: exprAppl{Callee: *fnref, Arg: allargs[0]}, Arg: allargs[1]}
								} else if numargs == 6 && fnl == StdFuncFalse && fnr == StdFuncTrue {
									didModify = true
									return exprAppl{Callee: exprAppl{Callee: exprAppl{Callee: exprAppl{Callee: *fnref, Arg: allargs[0]}, Arg: allargs[1]}, Arg: allargs[5]}, Arg: allargs[4]}
								}
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

// ie. from foo>=1 to foo>0, 2<=foo to 1<foo etc.
func optimize_callsToGeqOrLeq(src Prog) (ret Prog, didModify bool) {
	ret = src
	geqsleqs := map[int]bool{} // gathers any global gEQ / lEQ defs that or-combine LT/GT with EQ, if such exist
	for i := int(StdFuncCons + 1); i < len(ret)-1; i++ {
		if _, fnr1, _, _, _, allargs1 := dissectCall(ret[i].Body, nil); fnr1 != nil && len(allargs1) == 4 {
			if opcode1 := OpCode(*fnr1); opcode1 == OpEq || opcode1 == OpLt || opcode1 == OpGt {
				if fntrue, _ := allargs1[2].(ExprFuncRef); fntrue == StdFuncTrue {
					if _, fnr2, _, _, _, allargs2 := dissectCall(allargs1[3], nil); fnr2 != nil && len(allargs2) == 2 {
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
	if len(geqsleqs) == 0 {
		return
	}
	for i := int(StdFuncCons + 1); i < len(ret); i++ {
		ret[i].Body = walk(ret[i].Body, func(expr Expr) Expr {
			if _, fnref, _, _, _, allargs := dissectCall(expr, nil); fnref != nil && (len(allargs) == 1 || len(allargs) == 2) {
				numl, isnuml := allargs[0].(ExprNumInt)
				numr, isnumr := allargs[len(allargs)-1].(ExprNumInt)
				if isnuml || isnumr {
					if isgeq, orleq := geqsleqs[int(*fnref)]; orleq {
						if len(allargs) == 1 && isnuml {
							if didModify = true; isgeq {
								return exprAppl{Callee: ExprFuncRef(OpGt), Arg: numl + 1}
							} else {
								return exprAppl{Callee: ExprFuncRef(OpLt), Arg: numl - 1}
							}
						} else if len(allargs) == 2 {
							if isnuml {
								if didModify = true; isgeq {
									return exprAppl{Callee: exprAppl{Callee: ExprFuncRef(OpGt), Arg: numl + 1}, Arg: allargs[1]}
								} else {
									return exprAppl{Callee: exprAppl{Callee: ExprFuncRef(OpLt), Arg: numl - 1}, Arg: allargs[1]}
								}
							} else if isnumr {
								if didModify = true; isgeq {
									return exprAppl{Callee: exprAppl{Callee: ExprFuncRef(OpGt), Arg: allargs[0]}, Arg: numr - 1}
								} else {
									return exprAppl{Callee: exprAppl{Callee: ExprFuncRef(OpLt), Arg: allargs[0]}, Arg: numr + 1}
								}
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

// 0+foo, 1*foo, foo-0, foo/1 etc..
func optimize_primOpPreCalcs(src Prog) (ret Prog, didModify bool) {
	ret = src
	for i := int(StdFuncCons + 1); i < len(ret); i++ {
		ret[i].Body = walk(ret[i].Body, func(expr Expr) Expr {
			if _, fnref, numargs, _, _, allargs := dissectCall(expr, nil); fnref != nil && *fnref < 0 {
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

// removes an arg from a FuncDef if no non-call uses of the FuncDef exist and all its callers supply said arg, and with the exact same (non-argref-containing) expression
func optimize_inlineEverSameArgs(src Prog) (ret Prog, didModify bool) {
	ret = src
	allappls, min, max := map[ExprFuncRef][]Expr{}, map[ExprFuncRef]int{}, map[ExprFuncRef]int{}
	for i := StdFuncCons + 1; int(i) < len(ret)-1; i++ {
		if !doesHaveNonCalleeUses(ret, i) {
			allappls[i] = nil
		}
	}
	for i := range allappls {
		max[i], min[i] = 0, len(ret[i].Args)
		var chk func(Expr) Expr
		chk = func(expr Expr) Expr {
			if _, fnref, _, _, _, allargs := dissectCall(expr, nil); fnref != nil && *fnref == i {
				if len(allargs) < min[i] {
					min[i] = len(allargs)
				}
				if len(allargs) > max[i] {
					max[i] = len(allargs)
				}
				if len(allargs) <= len(ret[i].Args) {
					allappls[i] = append(allappls[i], expr)
				}
				for _, arg := range allargs {
					_ = chk(arg)
				}
				return nil
			}
			return expr
		}
		for j := StdFuncCons + 1; int(j) < len(ret); j++ {
			_ = walk(ret[j].Body, chk)
		}
	}
	for i, appls := range allappls {
		if len(appls) > 0 {
			for aidx := 0; aidx < min[i]; aidx++ {
				var argval Expr
				for _, appl := range appls {
					if _, _, _, _, _, allargs := dissectCall(appl, nil); len(allargs) > aidx {
						_, hasargref := allargs[aidx].(ExprArgRef)
						_ = walk(allargs[aidx], func(expr Expr) Expr {
							if hasargref {
								return nil
							}
							_, hasargref = expr.(ExprArgRef)
							return expr
						})
						if argval == nil && !hasargref {
							argval = allargs[aidx]
						} else if hasargref || !eq(argval, allargs[aidx]) {
							argval = exprTmp(-987654321)
						}
					}
				}
				if tmp, _ := argval.(exprTmp); argval != nil && tmp != -987654321 {
					for j := StdFuncCons + 1; int(j) < len(ret); j++ {
						ret[j].Body = walk(ret[j].Body, func(expr Expr) Expr {
							if _, fnref, _, _, _, allargs := dissectCall(expr, nil); fnref != nil && *fnref == i && len(allargs) == 1+aidx {
								return expr.(exprAppl).Callee
							}
							return expr
						})
					}
					ret[i].Args = append(ret[i].Args[:aidx], ret[i].Args[1+aidx:]...)
					ret[i].Meta = append(ret[i].Meta[:1+aidx], ret[i].Meta[2+aidx:]...)
					ret[i].Body = walk(ret[i].Body, func(expr Expr) Expr {
						if argref, is := expr.(ExprArgRef); is {
							if idx := int(-argref) - 1; idx == aidx {
								return argval
							}
						}
						return expr
					})
					ret[i].Body = walk(ret[i].Body, func(expr Expr) Expr {
						if argref, is := expr.(ExprArgRef); is {
							if idx := int(-argref) - 1; idx > aidx {
								return 1 + argref
							}
						}
						return expr
					})
					didModify = true
					return
				}
			}
		}
	}
	return
}

// rewrites all calls with no arg-refs and no `OpPrt`s with their `Eval` result
func optimize_preEvalArgRefLessCalls(src Prog) (ret Prog, didModify bool) {
	ret = src
	conv := make([]FuncDef, len(ret))
	copy(conv, ret)
	for i := range conv {
		conv[i].Body = convTo(conv[i].Body)
	}
	for i := int(StdFuncCons + 1); i < len(conv); i++ {
		var didmodify bool
		conv[i].Body = walkInPostOrder(conv[i].Body, func(expr Expr) Expr {
			if evald := tryEvalArgRefLessCall(conv, expr, true); !eq(evald, expr) {
				expr, didModify, didmodify = evald, true, true
			}
			return expr
		})
		if didmodify {
			ret[i].Body = convFrom(conv[i].Body)
		}
	}
	return
}
