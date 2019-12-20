package atem

import (
	"fmt"
	"os"
	"time"
)

var OnEvalStep = onEvalStepNoOp

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

// Eval operates non-recursively via an internal call stack. A stack entry
// holds the callee and the args. The former is first evaluated down to a
// "callable" (ExprFuncRef or a closure), then only those args that are
// actually used. Then the "callable"'s body (or prim-op) is evaluated with
// all evaluated args reachable.
//
// If not enough args are available, the result is a closure that does keep
// the already-evaluated args around for later completion. These will not be
// re-evaluated.
func (me Prog) Eval(expr Expr) Expr {
	maxLevels, maxStash, numSteps = 0, 0, 0
	t := time.Now().UnixNano()
	ret := me.eval(expr, 32*1024)
	t = time.Now().UnixNano() - t
	println(fmt.Sprintf("%T", ret), time.Duration(t).String(), "\t\t\t", maxLevels, maxStash, numSteps, "\t\t", count1, count2, count3, count4)
	return ret
}

var maxLevels int
var maxStash int
var numSteps int

func (me Prog) eval(expr Expr, initialLevelsCap int) Expr {
	// every call stacks a new `level` on top of lower ones, when call is done it's dropped.
	// but there's always 1 root / base `level` for our `expr`
	type level struct {
		stash     []Expr // args in reverse order, then callee
		pos       int    // begins at end of `stash` and counts down
		argsLevel int

		numArgs    int
		argsDone   bool
		calleeDone bool
	}

	levels := make([]level, 1, initialLevelsCap)
	levels[0].stash = append(make([]Expr, 0, 32), expr)
	var numargsdone int
	var idxcurlevel int
	var idxlast int
	cur := &levels[idxcurlevel]

again:
	numSteps++
	idxlast = len(cur.stash) - 1
	if idxlast > maxStash {
		maxStash = idxlast
	}

	for cur.pos < 0 {
		if cur.argsDone { // set near the end of the `again` loop. now with all (used) args eval'd we jump back to also-already-eval'd-to-callable callee
			cur.pos = idxlast
			// } else if idxlast != 0 {
			// 	count1++
			// 	cur.argsDone, cur.calleeDone, cur.numArgs, cur.pos, cur.stash =
			// 		false, false, 0, 0, []Expr{&ExprCall{Callee: cur.stash[idxlast], Args: cur.stash[:idxlast]}}
		} else if idxcurlevel == 0 {
			goto allDoneThusReturn // initial `expr` maximally reduced: return.
		} else { // jump back up to prior level, dropping the current one
			prev := &levels[idxcurlevel-1]
			prev.stash[prev.pos] = cur.stash[idxlast]
			cur, levels, idxcurlevel = prev, levels[:idxcurlevel], idxcurlevel-1
			idxlast = len(cur.stash) - 1
		}
	}

	switch it := cur.stash[cur.pos].(type) {

	case nil: // a will-be-discarded call-arg slot. we arrive here as our `pos` counts down, and continue:
		cur.pos--

	case ExprNumInt: // a no-further-reducable final value, count down to "next" slot
		cur.pos--

	case ExprArgRef:
		stash := levels[cur.argsLevel].stash
		if cur.calleeDone {
			stash = levels[idxcurlevel].stash
		}
		cur.stash[cur.pos] = stash[(len(stash)-1)+int(it)]
		if _, isargref := cur.stash[cur.pos].(ExprArgRef); isargref {
			cur.pos--
		} else {
			goto again
		}

	case *ExprCall:
		if it.IsClosure != 0 {
			cur.pos-- // we have a no-further-reducable final value (closure)
		} else { // build up & add & enter the next `level`
			callee, callargs := it.Callee, append(make([]Expr, 0, 3+len(it.Args)), it.Args...)
			for sub, isc := callee.(*ExprCall); isc; sub, isc = callee.(*ExprCall) {
				callee, callargs = sub.Callee, append(callargs, sub.Args...)
			}
			argslevel := cur.argsLevel
			if cur.calleeDone {
				argslevel = idxcurlevel
			}
			idxcurlevel, levels = idxcurlevel+1, append(levels, level{pos: len(callargs), stash: append(callargs, callee), argsLevel: argslevel})
			cur = &levels[idxcurlevel]
			if idxcurlevel > maxLevels {
				maxLevels = idxcurlevel
			}
			goto again
		}

	case ExprFuncRef:
		if it > 0 && len(me[it].Args) == 0 {
			cur.stash[cur.pos] = me[it].Body
			goto again
		} else if cur.calleeDone || idxlast == 0 || cur.pos != idxlast {
			cur.pos--
		} else if cur.numArgs == 0 {
			cur.numArgs = 2     // prim-op default
			allargsused := true // prim-op default
			if it > -1 {        // refers to actual func, not prim-op
				cur.numArgs, allargsused = len(me[it].Args), me[it].allArgsUsed
				// optional micro-optimization block: activates with approx. 25% - 35% of cases here
				if me[it].selector != 0 && len(cur.stash) > cur.numArgs {
					if me[it].selector < 0 {
						selected := cur.stash[idxlast+me[it].selector]
						cur.stash = append(cur.stash[:idxlast-cur.numArgs], selected)
					} else {
						call, _ := me[it].Body.(*ExprCall)
						argref, _ := call.Callee.(ExprArgRef)
						newtail := make([]Expr, 1+len(call.Args))
						newtail[len(call.Args)] = cur.stash[idxlast+int(argref)]
						for i := range call.Args {
							argref, _ = call.Args[i].(ExprArgRef)
							newtail[i] = cur.stash[idxlast+int(argref)]
						}
						cur.stash = append(cur.stash[:idxlast-cur.numArgs], newtail...)
					}
					if numargsdone -= cur.numArgs; numargsdone < 0 {
						numargsdone = 0
					}
					cur.numArgs, cur.pos = 0, len(cur.stash)-1
					goto again
				}
			}
			if !allargsused { // then ditch unused ones
				until := idxlast
				if cur.numArgs < idxlast { // very rare but happens 0% - 0.1% of the time depending on program
					until = cur.numArgs
				}
				for i := numargsdone; i < until; i++ {
					if me[it].Args[i] == 0 {
						cur.stash[len(cur.stash)-(2+i)] = nil
					}
				}
			}
			numargsdone, cur.pos = 0, cur.pos-(1+numargsdone)
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
			cur.calleeDone, cur.stash[idxlast] = true, result
			goto again
		} else {
			cur.pos--
		}
	}

	if idxlast != 0 && cur.pos < idxlast {
		if cur.argsDone {
			result := cur.stash[idxlast]
			if diff := cur.numArgs - idxlast; diff < 1 {
				cur.stash = append(cur.stash[:len(cur.stash)-1-cur.numArgs], result)
				// } else if idxcurlevel > 1 && len(levels[idxcurlevel-1].stash) != 1 && levels[idxcurlevel-1].pos == len(levels[idxcurlevel-1].stash)-1 {
				// 	callee, callargs := result, cur.stash[:idxlast]
				// 	idxcurlevel--
				// 	cur, levels = &levels[idxcurlevel], levels[:len(levels)-1]
				// 	cur.stash = append(append(cur.stash[:len(cur.stash)-1], callargs...), callee)
				// 	cur.numArgs, cur.calleeDone, cur.pos, cur.argsDone = 0, false, len(cur.stash)-1, false
				// 	goto again
			} else {
				result = &ExprCall{IsClosure: diff, Callee: result, Args: cur.stash[:idxlast]}
				cur.stash[idxlast] = result
				cur.stash = cur.stash[idxlast:]
			}
			cur.calleeDone, cur.numArgs, cur.argsDone = false, 0, false
			if len(cur.stash) == 1 {
				cur.pos = -1
			} else {
				cur.pos = len(cur.stash) - 1
			}
		} else if cur.numArgs == 0 {
			closure, _ := cur.stash[idxlast].(*ExprCall)
			cur.stash = append(append(cur.stash[:idxlast], closure.Args...), closure.Callee)
			numargsdone, cur.pos = len(closure.Args), len(cur.stash)-1
		} else if cur.pos < 0 || cur.pos < idxlast-cur.numArgs {
			cur.pos, cur.argsDone = -1, true
		}
	}
	goto again

allDoneThusReturn:
	return levels[0].stash[0]
}

var count1 int
var count2 int
var count3 int
var count4 int

func onEvalStepNoOp(Expr, []Expr) func(Expr, []Expr, bool) { return onEvalStepDoneNoOp }

func onEvalStepDoneNoOp(Expr, []Expr, bool) {}
