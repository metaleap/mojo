package atem

import (
	"fmt"
	"os"
	"time"
)

var OnEvalStep = onEvalStepNoOp

// selectors: funcs with 2+ args and a body of:
// - either only an argref
// - or a call with argref callee and *only* argref call-args, callee > "0"

// consider case of call: "someListArg" caseNil caseCons
//   - for one, in general: to have this call's args discarded or eval'd, need to eval callee first
//     - status quo: only argref, no eval'ing
//     - okay, but argref means our caller eval'd (also had to, tho)
//   - mapping someListArg to known-selector-fnref we could avoid
//     eval'ing first its closure value (as in, eval-the-callee) then again its full body,
//     already knowing mere selection occurs

// OpCode denotes a "primitive instruction", eg. one that is hardcoded in the
// interpreter and invoked when encountering a call to a negative `ExprFuncRef`
// with at least 2 operands on the current `Eval` stack. All `OpCode`-denoted
// primitive instructions consume always exactly 2 operands from said stack.
type OpCode int

const (
	// Addition of 2 `ExprNumInt`s, result 1 `ExprNumInt`
	OpAdd OpCode = -1
	// Subtraction of 2 `ExprNumInt`s, result 1 `ExprNumInt`
	OpSub OpCode = -2
	// Multiplication of 2 `ExprNumInt`s, result 1 `ExprNumInt`
	OpMul OpCode = -3
	// Division of 2 `ExprNumInt`s, result 1 `ExprNumInt`
	OpDiv OpCode = -4
	// Modulo of 2 `ExprNumInt`s, result 1 `ExprNumInt`
	OpMod OpCode = -5
	// Equality test between 2 `Expr`s, result is `StdFuncTrue` or `StdFuncFalse`
	OpEq OpCode = -6
	// Less-than test between 2 `ExprNumInt`s, result is `StdFuncTrue` or `StdFuncFalse`
	OpLt OpCode = -7
	// Greater-than test between 2 `ExprNumInt`s, result is `StdFuncTrue` or `StdFuncFalse`
	OpGt OpCode = -8
	// Writes both `Expr`s (the first one a string-ish `StdFuncCons`tructed linked-list of `ExprNumInt`s) to `OpPrtDst`, result is the right-hand-side `Expr` of the 2 input `Expr` operands
	OpPrt OpCode = -42
)

// OpPrtDst is the output destination for all `OpPrt` primitive instructions.
// Must never be `nil` during any `Prog`s that do potentially invoke `OpPrt`.
var OpPrtDst = os.Stderr.Write

// Eval operates thusly, keeping an internal `stack` of `[]Expr`:
//
// - encountering an `ExprCall`, its `Args` are `append`ed to the `stack` and
// its `Callee` is then `Eval`'d;
//
// - encountering an `ExprFuncRef`, the `stack` is checked for having the
// proper minimum required `len` with regard to the referenced `FuncDef`'s
// number of `Args`. If okay, the pertinent number of args is taken (and
// removed) from the `stack` and the referenced `FuncDef`'s `Body`, rewritten
// with all inner `ExprArgRef`s (including those inside `ExprCall`s) resolved
// to the `stack` entries, is `Eval`'d (with the appropriately reduced `stack`);
//
// - encountering  any other `Expr` type, it is merely returned if the `stack`
// is empty, else a `panic` with the `stack` signals that from the input
// `expr` a non-callable value ended up as the callee of an `ExprCall`.
//
// Corner cases for the `ExprFuncRef` situation: if the `stack` has too small a
// `len`, either an `ExprCall` representing the partial-application closure is
// returned, or just the `ExprFuncRef` in case of a totally empty `stack`;
// if the `ExprFuncRef` is negative and thus referring to a primitive-instruction
// `OpCode`, 2 is the expected minimum required `len` for the `stack` and if
// this is met, the primitive instruction is carried out, its `Expr` result
// then being `Eval`'d with the reduced-by-2 `stack`. Unknown op-codes `panic`
// with a `[3]Expr` of first the `OpCode`-referencing `ExprFuncRef` followed
// by both its operands.
func (me Prog) Eval(expr Expr) Expr {
	// id := func(it Expr) Expr { return &ExprCall{Callee: StdFuncId, Args: []Expr{it}} }
	// me[len(me)-1] = FuncDef{Meta: []string{"main", "args", "env"}, allArgsUsed: false, hasArgRefs: true,
	// 	Args: []int{0, 0},
	// 	Body: &ExprCall{Callee: ExprFuncRef(6), Args: []Expr{ListFrom([]byte("!?")), ExprFuncRef(OpAdd)}},
	// }

	maxLevels, maxStash, numSteps = 0, 0, 0
	fnNumCalls = make(map[ExprFuncRef]int, len(me))
	t := time.Now().UnixNano()
	ret := me.eval2(expr)
	t = time.Now().UnixNano() - t
	println(fmt.Sprintf("%T", ret), time.Duration(t).String(), "\t\t\t", maxLevels, maxStash, numSteps, "\t\t", Count1, Count2, Count3, Count4)
	for fnr, num := range fnNumCalls {
		println(num, "\tx\t", me[fnr].Meta[0])
	}
	return ret
}

var maxLevels int
var maxStash int
var numSteps int
var fnNumCalls = map[ExprFuncRef]int{}

func (me Prog) eval2(expr Expr) Expr {
	// every call stacks a new `level` on top of lower ones, when call is done it's dropped.
	// but there's always 1 root / base `level` for our `expr`
	type level struct {
		stash      []Expr // args in reverse order, then callee
		pos        int    // begins at end of `stash` and counts down
		numArgs    int
		argsLevel  int
		argsDone   bool
		calleeDone bool
	}
	levels := make([]level, 1, 1024)
	levels[0].stash = append(make([]Expr, 0, 32), expr)

	for ; numSteps < 12345; numSteps++ {
		print("\n\t")
		cur := &levels[len(levels)-1]
		for i := len(cur.stash) - 1; i >= 0; i-- {
			str := "_"
			if cur.stash[i] != nil {
				str = cur.stash[i].JsonSrc()
			}
			print(str + "\t\t")
		}
		println("\nlevel", len(levels)-1, "stash", len(cur.stash), "\tpos", cur.pos, "\t\targs", cur.argsDone, cur.numArgs, "@", cur.argsLevel)
		if len(levels) > maxLevels {
			maxLevels = len(levels)
		}
		if len(cur.stash) > maxStash {
			maxStash = len(cur.stash)
		}

		for cur.pos < 0 {
			if cur.argsDone { // set at end-of-loop. now with all (used) args eval'd we jump back to callee
				cur.pos = len(cur.stash) - 1
			} else if len(cur.stash) != 1 {
				cur.argsDone, cur.calleeDone, cur.numArgs, cur.pos, cur.stash =
					false, false, 0, 0, []Expr{&ExprCall{Callee: cur.stash[len(cur.stash)-1], Args: append([]Expr{}, cur.stash[:len(cur.stash)-1]...)}}
			} else if len(levels) == 1 { // initial `expr` maximally reduced: return.
				goto allDoneThusReturn
			} else { // jump back up to prior level, dropping the current one
				prev := &levels[len(levels)-2]
				prev.stash[prev.pos] = cur.stash[len(cur.stash)-1]
				println("\tBACK FROM", len(levels)-1, "TO", len(levels)-2, "AT", prev.pos, "NOW HAS", prev.stash[prev.pos].JsonSrc())
				cur, levels = prev, levels[:len(levels)-1]
			}
		}

		if cur.stash[cur.pos] != nil {
			println(cur.pos, fmt.Sprintf("NOW\t%T\t\t%s", cur.stash[cur.pos], cur.stash[cur.pos].JsonSrc()), "\t\t\t", cur.stash[len(cur.stash)-1].JsonSrc())
		}

		switch it := cur.stash[cur.pos].(type) {

		// a will-be-discarded call-arg slot. we arrive here as our `pos` counts down, and continue:
		case nil:
			cur.pos--

		// a no-further-reducable final value, count down to "next" slot
		case ExprNumInt:
			cur.pos--

		case ExprArgRef:
			stash := levels[cur.argsLevel].stash
			if cur.calleeDone {
				stash = levels[len(levels)-1].stash
			}
			cur.stash[cur.pos] = stash[(len(stash)-1)+int(it)]
			if _, isargref := cur.stash[cur.pos].(ExprArgRef); isargref {
				cur.pos--
			}

		case *ExprCall:
			if it.allArgsDone && it.IsClosure != 0 {
				cur.pos-- // we have a no-further-reducable final value (closure)
			} else { // build up & add & enter the next `level`
				callee, callargs := it.Callee, append([]Expr{}, it.Args...)
				for sub, isc := callee.(*ExprCall); isc; sub, isc = callee.(*ExprCall) {
					callee, callargs = sub.Callee, append(callargs, sub.Args...)
				}
				argslevel := cur.argsLevel
				if cur.calleeDone {
					argslevel = len(levels) - 1
				}
				levels = append(levels, level{pos: len(callargs), stash: append(callargs, callee), argsLevel: argslevel})
				println("\tNEXTLEV, RET TO:", cur.pos)
				continue
			}

		case ExprFuncRef:
			if it > 0 && len(me[it].Args) == 0 {
				cur.stash[cur.pos] = me[it].Body
			} else if (!cur.calleeDone) && cur.pos == len(cur.stash)-1 && len(cur.stash) != 1 {
				if !cur.argsDone {
					cur.numArgs = 2
					allargsused := true
					if it > -1 {
						cur.numArgs, allargsused = len(me[it].Args), me[it].allArgsUsed
					}
					if !allargsused {
						for i, idx := 0, len(cur.stash)-2; idx > -1 && i < cur.numArgs; i, idx = i+1, idx-1 {
							if me[it].Args[i] == 0 {
								cur.stash[idx] = nil
							}
						}
					}
					cur.pos--
				} else if len(cur.stash) > cur.numArgs {
					var result Expr
					if it < 0 {
						lhs, rhs := cur.stash[len(cur.stash)-2], cur.stash[len(cur.stash)-3]
						switch OpCode(it) {
						case OpAdd:
							result = lhs.(ExprNumInt) + rhs.(ExprNumInt)
						case OpSub:
							result = lhs.(ExprNumInt) - rhs.(ExprNumInt)
						case OpMul:
							result = lhs.(ExprNumInt) * rhs.(ExprNumInt)
						case OpDiv:
							result = lhs.(ExprNumInt) / rhs.(ExprNumInt)
						case OpMod:
							result = lhs.(ExprNumInt) % rhs.(ExprNumInt)
						case OpGt:
							if result = StdFuncFalse; lhs.(ExprNumInt) > rhs.(ExprNumInt) {
								result = StdFuncTrue
							}
						case OpLt:
							if result = StdFuncFalse; lhs.(ExprNumInt) < rhs.(ExprNumInt) {
								result = StdFuncTrue
							}
						case OpEq:
							if result = StdFuncFalse; me.Eq(lhs, rhs) {
								result = StdFuncTrue
							}
						case OpPrt:
							result = rhs
							_, _ = OpPrtDst(append(append(append(ListToBytes(me.ListOfExprs(lhs)), '\t'), me.ListOfExprsToString(rhs)...), '\n'))
						default:
							panic([3]Expr{it, lhs, rhs})
						}
					} else {
						result = me[it].Body
					}
					cur.calleeDone, cur.stash[len(cur.stash)-1] = true, result
				} else {
					cur.pos--
				}
			} else {
				cur.pos--
			}
		}
		if cur.pos >= 0 && cur.stash[cur.pos] != nil {
			println("\tTHEN\t", cur.pos, fmt.Sprintf("%T", cur.stash[cur.pos]), "\t", cur.stash[cur.pos].JsonSrc(), "\t\t\t", cur.stash[len(cur.stash)-1].JsonSrc())
		}

		if len(cur.stash) != 1 && cur.pos < (len(cur.stash)-1) {
			if cur.argsDone {
				result := cur.stash[len(cur.stash)-1]
				if diff := cur.numArgs - (len(cur.stash) - 1); diff > 0 {
					result = &ExprCall{allArgsDone: true, IsClosure: diff, Callee: result, Args: cur.stash[:len(cur.stash)-1]}
					cur.stash = []Expr{result}
					println("\tB1.C", len(cur.stash))
				} else {
					cur.stash = append(append([]Expr{}, cur.stash[:len(cur.stash)-1-cur.numArgs]...), result)
					println("\tB1.T", len(cur.stash))
				}
				cur.calleeDone, cur.argsDone, cur.numArgs = false, false, 0
				if len(cur.stash) == 1 {
					cur.pos = -1
				} else {
					cur.pos = len(cur.stash) - 1
				}
			} else if cur.numArgs == 0 {
				if closure, iscl := cur.stash[len(cur.stash)-1].(*ExprCall); iscl {
					cur.argsDone, cur.calleeDone, cur.numArgs, cur.pos, cur.stash =
						false, false, 0, 0, []Expr{&ExprCall{Callee: closure, Args: append([]Expr{}, cur.stash[:len(cur.stash)-1]...)}}
					continue
				} else { // callee evaluated to non-callable
					panic(cur.stash[len(cur.stash)-1])
				}
			} else if cur.pos < 0 || cur.pos < (len(cur.stash)-(1+cur.numArgs)) {
				println("\tB2.A")
				cur.pos, cur.argsDone = -1, true
			}
		}
		println(numSteps, "\tFINALLY\t", cur.pos, cur.argsDone, cur.calleeDone, "\n")
	}
allDoneThusReturn:
	return levels[0].stash[0]
}

func (me Prog) eval(expr Expr, curFnArgs []Expr) Expr {
	for again := true; again; {
		ondone := OnEvalStep(expr, curFnArgs)
		again = false

		switch it := expr.(type) {
		case ExprArgRef:
			expr = curFnArgs[len(curFnArgs)+int(it)]
		case ExprFuncRef:
			if it > StdFuncCons && len(me[it].Args) == 0 {
				fnNumCalls[it] = 1 + fnNumCalls[it]
				again, expr = true, me[it].Body
			}
		case *ExprCall:
			if it.IsClosure == 0 || !it.allArgsDone { // for ADT-heavy progs, no-op case covers between 1/3 to 3/4 of *ExprCall cases
				numargsdone, callee, callargs := 0, me.eval(it.Callee, curFnArgs), it.Args
				if it.allArgsDone {
					numargsdone = len(it.Args)
				}
				for sub, isc := callee.(*ExprCall); isc; sub, isc = callee.(*ExprCall) {
					callee = me.eval(sub.Callee, curFnArgs)
					if sub.allArgsDone && numargsdone == len(callargs) {
						numargsdone += len(sub.Args)
					}
					callargs = append(callargs, sub.Args...)
				}
				numargs, fnref := 2, callee.(ExprFuncRef)
				isop := fnref < 0
				allargsused := isop
				if !isop {
					fnNumCalls[fnref] = 1 + fnNumCalls[fnref]
					numargs, allargsused = len(me[fnref].Args), me[fnref].allArgsUsed
				}
				var nextargs []Expr
				var nextargsdone bool
				var closure int
				if diff := len(callargs) - numargs; diff < 0 {
					closure = -diff
				} else if diff > 0 { // usually 1 or 2
					nextargsdone, nextargs = numargsdone >= diff, make([]Expr, diff)
					copy(nextargs, callargs[:diff])
					if callargs = callargs[diff:]; numargsdone <= diff {
						numargsdone = 0
					} else {
						numargsdone -= diff
					}
				}
				fnargs := callargs
				if numargsdone < len(fnargs) {
					fnargs = make([]Expr, len(callargs))
					tmp := numargs - closure - 1
					for i := range fnargs {
						if idx := tmp - i; allargsused || me[fnref].Args[idx] != 0 {
							if numargsdone > i {
								fnargs[i] = callargs[i]
							} else {
								fnargs[i] = me.eval(callargs[i], curFnArgs)
							}
						}
					}
				}
				if closure != 0 {
					expr = &ExprCall{allArgsDone: true, IsClosure: closure, Callee: fnref, Args: fnargs}
				} else {
					if isop {
						lhs, rhs := fnargs[1], fnargs[0]
						switch OpCode(fnref) {
						case OpAdd:
							expr = lhs.(ExprNumInt) + rhs.(ExprNumInt)
						case OpSub:
							expr = lhs.(ExprNumInt) - rhs.(ExprNumInt)
						case OpMul:
							expr = lhs.(ExprNumInt) * rhs.(ExprNumInt)
						case OpDiv:
							expr = lhs.(ExprNumInt) / rhs.(ExprNumInt)
						case OpMod:
							expr = lhs.(ExprNumInt) % rhs.(ExprNumInt)
						case OpEq:
							if expr = StdFuncFalse; me.Eq(lhs, rhs) {
								expr = StdFuncTrue
							}
						case OpLt:
							if expr = StdFuncFalse; lhs.(ExprNumInt) < rhs.(ExprNumInt) {
								expr = StdFuncTrue
							}
						case OpGt:
							if expr = StdFuncFalse; lhs.(ExprNumInt) > rhs.(ExprNumInt) {
								expr = StdFuncTrue
							}
						case OpPrt:
							expr = rhs
							_, _ = OpPrtDst(append(append(append(ListToBytes(me.ListOfExprs(lhs)), '\t'), me.ListOfExprsToString(rhs)...), '\n'))
						default:
							panic([3]Expr{it, lhs, rhs})
						}
					} else if nextargs != nil {
						expr = me.eval(me[fnref].Body, fnargs)
					} else if me[fnref].selector.of != 0 {
						if expr = fnargs[len(fnargs)+int(me[fnref].selector.of)]; me[fnref].selector.numArgs > 0 {
							again, curFnArgs, expr = true, nil, &ExprCall{allArgsDone: true, Callee: expr, Args: fnargs[len(fnargs)-me[fnref].selector.numArgs:]}
						}
					} else {
						again, expr, curFnArgs = true, me[fnref].Body, fnargs
					}
					if nextargs != nil {
						if fnr, _ := expr.(ExprFuncRef); fnr > 0 && me[fnr].selector.of != 0 && me[fnr].selector.numArgs == 0 && len(nextargs) >= len(me[fnr].Args) {
							again, expr = true, nextargs[len(nextargs)+int(me[fnr].selector.of)]
							nextargs = nextargs[:len(nextargs)-len(me[fnr].Args)]
						}
					}
					if len(nextargs) > 0 {
						again, expr = true, &ExprCall{allArgsDone: nextargsdone, Callee: expr, Args: nextargs}
					}
				}
			}
		}
		ondone(expr, curFnArgs, again)
	}
	return expr
}

var Count1 int
var Count2 int
var Count3 int
var Count4 int

func onEvalStepNoOp(Expr, []Expr) func(Expr, []Expr, bool) { return onEvalStepDoneNoOp }

func onEvalStepDoneNoOp(Expr, []Expr, bool) {}
