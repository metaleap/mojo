package atem

import (
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
	CurEvalStepDepth, maxDepth = 0, 0
	fnNumCalls = make(map[ExprFuncRef]int, len(me))
	t := time.Now().UnixNano()
	ret := me.eval(expr, nil)
	t = time.Now().UnixNano() - t
	println(time.Duration(t).String(), "\t\t\t", maxDepth, "\t\t", Count1, Count2, Count3, Count4)
	for fnr, num := range fnNumCalls {
		println(num, "\tx\t", me[fnr].Meta[0])
	}
	return ret
}

var CurEvalStepDepth int
var maxDepth int
var fnNumCalls map[ExprFuncRef]int

func (me Prog) eval(expr Expr, curFnArgs []Expr) Expr {
	if CurEvalStepDepth++; CurEvalStepDepth > maxDepth {
		maxDepth = CurEvalStepDepth
	}
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
							expr = &ExprCall{allArgsDone: true, Callee: expr, Args: fnargs[len(fnargs)-me[fnref].selector.numArgs:]}
						}
						again, curFnArgs = true, nil
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
	CurEvalStepDepth--
	return expr
}

var Count1 int
var Count2 int
var Count3 int
var Count4 int

func onEvalStepNoOp(Expr, []Expr) func(Expr, []Expr, bool) { return onEvalStepDoneNoOp }

func onEvalStepDoneNoOp(Expr, []Expr, bool) {}
