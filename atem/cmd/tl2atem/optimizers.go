package main

import (
	. "github.com/metaleap/atmo/atem"
	"strconv"
)

func init() { OpPrtDst = func([]byte) (int, error) { panic("caught in `tryEval`") } }

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

func optimize(src Prog) (ret Prog, didModify bool) {
	ret = src
	for again := true; again; {
		ret = fixFuncDefArgsUsageNumbers(ret)
		for _, opt := range []func(Prog) (Prog, bool){
			optimize_inlineNullaries,
			// optimize_ditchUnusedFuncDefs,
			optimize_inlineSelectorCalls,
			optimize_argDropperCalls,
			optimize_inlineArgCallers,
			optimize_inlineArgsRearrangers,
			optimize_primOpPreCalcs,
			optimize_callsToGeqOrLeq,
			optimize_minifyNeedlesslyElaborateBoolOpCalls,
			optimize_inlineOnceCalleds,
			optimize_inlineEverSameArgs,
			optimize_preEvals,
		} {
			if ret, again = opt(ret); again {
				didModify = true
				break
			}
		}
	}
	return
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

// inliners or other optimizers may result in now-unused func-defs, here's a single routine that'll remove them.
// it removes at most one at a time, fixing up all references, then returning with `didModify` of `true`, ensuring another call to find the next one
func optimize_ditchUnusedFuncDefs(src Prog) (ret Prog, didModify bool) {
	ret = src
	defrefs := make(map[int]bool, len(ret))
	for i := range ret {
		walk(ret[i].Body, func(expr Expr) Expr {
			if fnref, ok := expr.(ExprFuncRef); ok && fnref >= 0 && int(fnref) != i {
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

// nullary func-defs get first `tryEval`'d right here, then the result inlined at use sites if:
// that site is the only reference to the def, or the def's eval'd body is atomic or a call with only atomic args
func optimize_inlineNullaries(src Prog) (ret Prog, didModify bool) {
	ret = src
	nullaries, caninlinealways := make(map[int]int), make(map[int]bool)
	for i := int(StdFuncCons + 1); i < len(ret)-1; i++ {
		if 0 == len(ret[i].Args) {
			if body := tryEval(ret, ret[i].Body, false); !eq(body, ret[i].Body) {
				didModify, ret[i].Body = true, body
			}
			_, _, numargs, numargcalls, _, _ := dissectCall(ret[i].Body)
			nullaries[i], caninlinealways[i] = 0, numargs == 0 || numargcalls == 0
		}
	}
	if len(nullaries) == 0 {
		return
	}
	for i := int(StdFuncCons + 1); i < len(ret); i++ {
		_ = walk(ret[i].Body, func(expr Expr) Expr {
			if fnref, _ := expr.(ExprFuncRef); fnref > StdFuncCons {
				if numrefs, isnullary := nullaries[int(fnref)]; isnullary {
					nullaries[int(fnref)] = 1 + numrefs
				}
			}
			return expr
		})
	}
	for i := int(StdFuncCons + 1); i < len(ret); i++ {
		ret[i].Body = walk(ret[i].Body, func(expr Expr) Expr {
			if fnref, _ := expr.(ExprFuncRef); fnref > StdFuncCons {
				if numrefs, isnullary := nullaries[int(fnref)]; isnullary {
					if retexpr := ret[int(fnref)].Body; numrefs == 1 || caninlinealways[int(fnref)] {
						didModify = didModify || !eq(expr, retexpr)
						return retexpr
					}
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
			if _, fnref, numargs, _, _, _ := dissectCall(expr); fnref != nil {
				if argdrops := argdroppers[int(*fnref)]; len(argdrops) > 0 {
					return rewriteCallArgs(expr.(ExprAppl), numargs, func(argidx int, argval Expr) Expr {
						if !eq(argval, exprTmp(-123456789)) {
							argval, didModify = exprTmp(-123456789), true
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

// removes args from FuncDefs if all callers supply the exact same (non-argref-containing) expression
func optimize_inlineEverSameArgs(src Prog) (ret Prog, didModify bool) {
	ret = src
	allappls, num := map[ExprFuncRef][]Expr{}, map[ExprFuncRef]int{}
	for i := StdFuncCons + 1; int(i) < len(ret)-1; i++ {
		for j := StdFuncCons + 1; int(j) < len(ret); j++ {
			_ = walk(ret[j].Body, func(expr Expr) Expr {
				if fnr, _ := expr.(ExprFuncRef); fnr == i {
					num[i] = 1 + num[i]
				} else if _, fnref, _, _, _, _ := dissectCall(expr); fnref != nil && *fnref == i {
					allappls[i] = append(allappls[i], expr)
				}
				return expr
			})
		}
	}
	for i, appls := range allappls {
		if num[i] > len(appls) {
			delete(num, i)
			delete(allappls, i)
		}
	}
	for i := range num {
		allappls[i], num[i] = nil, len(ret[i].Args)
		var chk func(Expr) Expr
		chk = func(expr Expr) Expr {
			if _, fnref, _, _, _, allargs := dissectCall(expr); fnref != nil && *fnref == i {
				if len(allargs) < num[i] {
					num[i] = len(allargs)
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
		for aidx := 0; aidx < num[i]; aidx++ {
			var argval Expr
			for _, appl := range appls {
				if _, _, _, _, _, allargs := dissectCall(appl); len(allargs) > aidx {
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
				println("all calls to", ret[i].Meta[0], "have arg", aidx, "as", argval.JsonSrc())
			}
		}
	}
	return
}

// inlines a FuncDef into a call site if the latter is the only reference to the former, and provides enough args, and no arg is used more than once in the orig def's body
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
		could := (1 == len(referencers))
		if could {
			for _, argused := range ret[fn].Args {
				if could = (argused <= 1); !could {
					break
				}
			}
		}
		if could {
			for referencer, numrefs := range referencers {
				if numrefs == 1 {
					ret[referencer].Body = walk(ret[referencer].Body, func(expr Expr) Expr {
						if _, fnref, _, _, _, allargs := dissectCall(expr); fnref != nil && *fnref == fn && len(allargs) == len(ret[fn].Args) {
							didModify, expr = true, walk(walk(ret[fn].Body, func(it Expr) Expr {
								if argref, is := it.(ExprArgRef); is {
									return exprTmp(argref)
								}
								return it
							}), func(it Expr) Expr {
								if tmp, is := it.(exprTmp); is && tmp != -123456789 && tmp != -987654321 {
									return allargs[(-tmp)-1]
								}
								return it
							})
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
		if callee, _, numargs, numargscalls, numargrefs, _ := dissectCall(ret[i].Body); numargs > 0 && numargrefs == 1 && numargscalls == 0 {
			if argref, ok := callee.(ExprArgRef); ok { // only 1 arg-ref in body and its the inner-most callee?
				argcallers[i] = (-argref) - 1
			}
		}
	}
	if len(argcallers) == 0 {
		return
	}
	for i := int(StdFuncCons + 1); i < len(ret); i++ {
		ret[i].Body = walk(ret[i].Body, func(expr Expr) Expr {
			if _, fnref, numargs, _, _, allargs := dissectCall(expr); fnref != nil {
				if argref, ok := argcallers[int(*fnref)]; ok {
					if didModify = true; argref != 0 {
						panic("TODO")
					} else {
						call2inline := rewriteInnerMostCallee(ret[*fnref].Body.(ExprAppl), func(Expr) Expr { return allargs[0] })
						expr = rewriteInnerMostCallee(expr.(ExprAppl), func(Expr) Expr { return StdFuncId })
						expr = rewriteCallArgs(expr.(ExprAppl), numargs, func(argidx int, argval Expr) Expr {
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

// "selectors" have arg-ref bodies, well-known examples are id / true / false / nil.
// calls to those (that do provide enough args) get replaced by the respective call-arg.
func optimize_inlineSelectorCalls(src Prog) (ret Prog, didModify bool) {
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
			if _, fnref, numargs, _, _, allargs := dissectCall(expr); fnref != nil {
				if argref, ok := selectors[int(*fnref)]; ok && int(argref) < numargs && numargs == len(ret[*fnref].Args) {
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
		if _, _, _, numargcalls, numargrefs, allargs := dissectCall(ret[i].Body); len(allargs) > 0 && numargcalls == 0 && numargrefs > 0 {
			rearrangers[i] = len(allargs)
		}
	}
	if len(rearrangers) == 0 {
		return
	}
	for i := int(StdFuncCons + 1); i < len(ret); i++ {
		ret[i].Body = walk(ret[i].Body, func(expr Expr) Expr {
			if _, fnref, numargs, _, _, allargs := dissectCall(expr); fnref != nil { // we have a call...
				if n := rearrangers[int(*fnref)]; n > 0 && numargs == len(ret[*fnref].Args) { // to a rearranger and with proper number of args
					counts := map[ExprArgRef]int{}
					// take orig body and replace all arg-refs with our call args:
					nuexpr := rewriteCallArgs(ret[*fnref].Body.(ExprAppl), n, func(argidx int, argval Expr) Expr {
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
						if _, iscall := allargs[argref].(ExprAppl); iscall && count > 1 {
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

// ie. from foo>=1 to foo>0, 2<=foo to 1<foo etc.
func optimize_callsToGeqOrLeq(src Prog) (ret Prog, didModify bool) {
	ret = src
	geqsleqs := map[int]bool{} // gathers any global gEQ / lEQ defs that or-combine LT/GT with EQ, if such exist
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
	if len(geqsleqs) == 0 {
		return
	}
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

// rewrites all occurrences of calls with no arg-refs and no side-effects (eg. OpPrt) with their `Eval` result
func optimize_preEvals(src Prog) (ret Prog, didModify bool) {
	ret = src
	for i := int(StdFuncCons + 1); i < len(ret); i++ {
		ret[i].Body = walk(ret[i].Body, func(expr Expr) Expr {
			if evald := tryEval(ret, expr, true); !eq(evald, expr) {
				expr, didModify = evald, true
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
	if expr == cmp {
		return true
	}
	switch it := expr.(type) {
	case exprTmp:
		that, ok := cmp.(exprTmp)
		return ok && it == that
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

func tryEval(prog Prog, expr Expr, checkForArgRefs bool) (ret Expr) {
	ret = expr
	if call, ok := expr.(ExprAppl); ok {
		caneval := true
		if checkForArgRefs {
			_ = walk(call, func(it Expr) Expr {
				_, isargref := it.(ExprArgRef)
				caneval = caneval && (!isargref)
				return it
			})
		}
		if caneval {
			defer func() {
				if recover() != nil {
					ret = expr
				}
			}()
			ret = walk(ret, func(it Expr) Expr { return prog.Eval(it, nil) })
		}
	}
	return
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

type exprTmp int

func (me exprTmp) JsonSrc() string { return strconv.Itoa(int(me)) }
