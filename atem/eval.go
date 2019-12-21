package atem

import (
	"fmt"
	"os"
	"time"
)

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

// OpPrtDst is the output sink for all `OpPrt` primitive instructions.
// Must never be `nil` during any `Prog`s that do potentially invoke `OpPrt`.
var OpPrtDst = os.Stderr.Write

// Eval operates non-recursively via an internal call stack. Any stack entry
// beyond the "root" / "base" one (that at first holds `expr` and at the end
// the final result value) first holds some call's callee and the args. The
// former is first evaluated (down to a "callable": ExprFuncRef or a closure),
// next then only those args that are actually used. Then the "callable"'s body
// (or prim-op) is evaluated, consuming those freshly-obtained arg values.
//
// If not enough args are available, the result is a closure that does keep
// the already-evaluated args around for later completion. These will not be
// re-evaluated.
//
// The final result of `Eval` will be an `ExprNumInt`, an `ExprFuncRef` or
// such a closure value (an `*ExprCall` with `.IsClosure != 0`), the latter
// can be tested for linked-list-ness and extracted via `Prog.ListOfExprs`.
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
	// every new call stacks a new `level` on top of prior ones, when call is
	// done it's dropped. but there's always 1 root / base `level` for our `expr`.
	type level struct {
		stash     []Expr // args in reverse order, then callee
		pos       int    // begins at end of `stash` and counts down
		argsLevel int    // index in `levels` from where `ExprArgRef`s resolve

		numArgs    int  // initially 0, until `ExprFuncRef` from the callee resolves
		argsDone   bool // `true` after `numArgs` known and all (used) args in `stash` fully eval'd
		calleeDone bool // `true` after the above and having jumped back to callee in `stash`
	}

	levels, idxlevel, idxcallee, numargsdone := make([]level, 1, initialLevelsCap), 0, 0, 0
	levels[idxlevel].stash = []Expr{expr}
	cur := &levels[idxlevel]

again:
	numSteps++
	idxcallee = len(cur.stash) - 1
	if idxcallee > maxStash {
		maxStash = idxcallee
	}

	for cur.pos < 0 { // in a new `level`, we start at end of `stash` (callee) and then travel down the args
		if idxlevel == 0 {
			goto allDoneThusReturn // initial `expr` maximally reduced: return.
		} else { // jump back up to parent call `level`, dropping the current one
			parent := &levels[idxlevel-1]
			parent.stash[parent.pos] = cur.stash[idxcallee]               // store result there
			cur, levels, idxlevel = parent, levels[:idxlevel], idxlevel-1 // now we're in `parent`
			idxcallee = len(cur.stash) - 1
		}
	}

	switch it := cur.stash[cur.pos].(type) {

	case nil: // a will-be-discarded call-arg slot. was cleared when the callee resolved to a final callable
		cur.pos-- // we arrived here as our `pos` counts down, and keep going

	case ExprNumInt: // a no-further-reducable final value
		cur.pos-- // count down as well to travel further down the `stash`

	case ExprArgRef:
		lookuplevel := cur.argsLevel // most common case
		if cur.calleeDone {          // very rare case:
			lookuplevel = idxlevel // essentially callees that merely return one of their args as-is
		}
		lookupstash := levels[lookuplevel].stash
		cur.stash[cur.pos] = lookupstash[(len(lookupstash)-1)+int(it)]
		goto again // whatever we got, we want to further evaluate: no need for the final post-`switch` checks on `cur.pos` since it hasn't changed, can go at it again right away

	case *ExprCall:
		if it.IsClosure != 0 { // if so: a currently-no-further-reducable final value (closure)
			cur.pos--
		} else { // build up & add & enter the next `level`
			callee, callargs := it.Callee, append(make([]Expr, 0, 3+len(it.Args)), it.Args...)
			for sub, isc := callee.(*ExprCall); isc; sub, isc = callee.(*ExprCall) { // flatten to single call
				callee, callargs = sub.Callee, append(callargs, sub.Args...)
			}
			lookuplevel := cur.argsLevel // same logic as above in `case` of `ExprArgRef`:
			if cur.calleeDone {          // ...but this now occurs ~50/50
				lookuplevel = idxlevel
			}
			idxlevel, levels = idxlevel+1, append(levels, level{
				pos: len(callargs), stash: append(callargs, callee), argsLevel: lookuplevel})
			cur = &levels[idxlevel] // now enter the newly created `level`
			if idxlevel > maxLevels {
				maxLevels = idxlevel
			}
			goto again
		}

	case ExprFuncRef: // recall: if it<0 the `ExprFuncRef` refers to an `OpCode`
		if isfn := it > -1; isfn && me[it].mereAlias {
			cur.stash[cur.pos] = me[it].Body
			goto again
		} else if cur.calleeDone || cur.pos != idxcallee { // either not in callee position or else callee reduced to current `it`?
			cur.pos-- // then the `ExprFuncRef` is a mere currently-no-further-reducable value to just pass along / return / preserve for now
		} else /* we are in callee position */ if cur.numArgs == 0 { // then must determine this now, first!
			cur.numArgs = 2     // prim-op default
			allargsused := true // prim-op default
			if isfn {           // refers to actual func, not prim-op
				cur.numArgs, allargsused = len(me[it].Args), me[it].allArgsUsed
				// optional micro-optimization block: entered-into for approx. 25% - 35% of cases here
				if me[it].selector != 0 && len(cur.stash) > cur.numArgs {
					if me[it].selector < 0 {
						selected := cur.stash[idxcallee+me[it].selector]
						cur.stash = append(cur.stash[:idxcallee-cur.numArgs], selected)
					} else {
						call, _ := me[it].Body.(*ExprCall)
						argref, _ := call.Callee.(ExprArgRef)
						newtail := make([]Expr, 1+len(call.Args))
						newtail[len(call.Args)] = cur.stash[idxcallee+int(argref)]
						for i := range call.Args {
							argref, _ = call.Args[i].(ExprArgRef)
							newtail[i] = cur.stash[idxcallee+int(argref)]
						}
						cur.stash = append(cur.stash[:idxcallee-cur.numArgs], newtail...)
					}
					if numargsdone -= cur.numArgs; numargsdone < 0 {
						numargsdone = 0
					}
					cur.numArgs, cur.pos = 0, len(cur.stash)-1
					goto again
				}
			}
			if cur.numArgs == 0 { // no args means a global:
				call, _ := me[it].Body.(*ExprCall) // also means *ExprCall because others were caught at top of this `case`
				cur.stash = append(append(cur.stash[:idxcallee], call.Args...), call.Callee)
				if cur.pos = len(cur.stash) - 1; call.IsClosure == 0 {
					numargsdone = 0
				} else {
					numargsdone += len(call.Args)
				}
				goto again
			} else if !allargsused { // then ditch unused ones: by setting unused arg-slots in `stash` to `nil`
				until := idxcallee
				if cur.numArgs < idxcallee { // very rare (at *this* code-path point), around 0% - 0.1% of the time depending on program
					until = cur.numArgs
				}
				for i := numargsdone; i < until; i++ {
					if me[it].Args[i] == 0 { // unused? then clear args-slot:
						cur.stash[len(cur.stash)-(2+i)] = nil
					}
				}
			}
			cur.pos, numargsdone = cur.pos-(1+numargsdone), 0 // jump down to first arg that needs eval-ing, will then count down from there
		} else if len(cur.stash) > cur.numArgs { // we have all args eval'd, now comes the callee's body
			var result Expr
			if isfn { // substitution
				result = me[it].Body
			} else { // prim-op instruction code on left-hand-side and right-hand-side operands
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
			}
			cur.calleeDone, cur.stash[idxcallee] = true, result
			goto again
		} else {
			cur.pos-- // in callee position but callee was already done: the post-`switch` checks below handle that if we count one down
		}

	} // done type-switch on cur.stash[cur.pos]

	if idxcallee != 0 && cur.pos < idxcallee { // we're still arg-ful and below callee-position
		if cur.argsDone { // again: we are below callee position though all args were already eval'd:
			result := cur.stash[idxcallee]                 // so `calleeDone` or not (50/50), we're good to return
			if diff := cur.numArgs - idxcallee; diff < 1 { // the `calleeDone` (non-closure) case
				cur.stash = append(cur.stash[:len(cur.stash)-1-cur.numArgs], result)
			} else /* result is closure */ if ilp := idxlevel - 1; ilp > 0 && len(levels[ilp].stash) != 1 && levels[ilp].numArgs == 0 && levels[ilp].pos == len(levels[ilp].stash)-1 {
				// this block optional micro-optimization
				callargs := cur.stash[:idxcallee]
				idxlevel, numargsdone, cur, levels = ilp, len(callargs), &levels[ilp], levels[:idxlevel]
				cur.stash = append(append(cur.stash[:len(cur.stash)-1], callargs...), result)
				cur.pos = len(cur.stash) - 1
				goto again
			} else {
				result = &ExprCall{IsClosure: diff, Callee: result, Args: cur.stash[:idxcallee]}
				cur.stash[idxcallee] = result
				cur.stash = cur.stash[idxcallee:]
			}
			cur.calleeDone, cur.numArgs, cur.argsDone = false, 0, false
			if len(cur.stash) == 1 {
				cur.pos = -1
			} else {
				cur.pos = len(cur.stash) - 1
			}
		} else if cur.numArgs == 0 { // callee was not an `ExprFuncRef` so must be a closure:
			closure, _ := cur.stash[idxcallee].(*ExprCall) // ... so unroll it into current `stash` :
			cur.stash = append(append(cur.stash[:idxcallee], closure.Args...), closure.Callee)
			numargsdone, cur.pos = len(closure.Args), len(cur.stash)-1 // ... and start over at callee
		} else if cur.pos < 0 || cur.pos < idxcallee-cur.numArgs {
			// okay, all args needed were eval'd, jump back to callee for its callable's body's reduction
			cur.pos, cur.argsDone = idxcallee, true
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
