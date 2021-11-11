package atem

import (
	"fmt"
	"time"
)

// Eval reduces `expr` to an `ExprNumInt`, an `ExprFuncRef` or a closure value
// (an `*ExprCall` with `.IsClosure > 0`, see field description there), the
// latter can be tested for linked-list-ness and extracted via `ListOfExprs`.
//
// The evaluator is akin to a tree-walking interpreter of the input `Prog` but
// given the nature of the `atem` intermediate-representation language, that
// amounts to a sort of register machine. A call stack is kept so that `Eval`
// never needs to recursively call itself. Any stack entry beyond the "root" /
// "base" one (that at first holds `expr` and at the end the final result value)
// represents a call: it at first holds both said call's callee and its args.
// The former is evaluated first (only down to a "callable": `ExprFuncRef` or
// closure), next then only those args are evaluated that are actually needed.
// Finally, the "callable"'s body (or prim-op) is evaluated further, consuming
// those freshly-obtained arg values while producing the call's result value.
// (If in a call not enough args are supplied to the callee, the result is a
// closure that does keep its fully-evaluated args around for later completion.)
//
// The `big` arg fine-tunes how much call-stack memory to pre-allocate at once
// beforehand. If `true`, this will be to the tune of ~2 MB, else under 10 KB.
// Put simply, `true` is for full-program running, `false` is for smallish
// "drive-by" / "side-car" expression evaluation attempts in the context of a
// given `Prog` such as in REPLs, optimizers, compilers or similar tooling.
func (me Prog) Eval(expr Expr, big bool) Expr {
	maxFrames, maxStash, numSteps = 0, 0, 0
	capframes := 64
	if big {
		capframes = 32 * 1024
	}
	ret, t := me.eval(expr, capframes)
	t = time.Now().UnixNano() - t
	if big {
		println(fmt.Sprintf("%T", ret), time.Duration(t).String(), "\t\t\t", maxFrames, maxStash, numSteps, "\t\t", count1, count2, count3, count4)
	}
	return ret
}

var maxFrames int
var maxStash int
var numSteps int
var count1 int
var count2 int
var count3 int
var count4 int

func (me Prog) eval(expr Expr, initialFramesCap int) (Expr, int64) {
	// every new call stacks a new `frame` on top of prior ones, when call is
	// done it's dropped. but there's always 1 root / base `frame` for our `expr`.
	type frame struct {
		stash     []Expr // args (could be too many or too few) in reverse order, then callee
		pos       int    // begins at end of `stash` and counts down
		argsFrame int    // index in `frames` from where `ExprArgRef`s resolve

		numArgs    int  // initially 0, until resolving callee to `ExprFuncRef`
		argsDone   bool // `true` after `numArgs` known and all needed args in `stash` fully eval'd
		calleeDone bool // `true` after the above and having jumped back to callee in `stash`
	}

	frames, idxframe, idxcallee, numargsdone := make([]frame, 1, initialFramesCap), 0, 0, 0
	frames[idxframe].stash = []Expr{expr}
	cur, starttime := &frames[idxframe], time.Now().UnixNano()

restep:
	numSteps++
	idxcallee = len(cur.stash) - 1
	if (len(cur.stash)) > maxStash {
		maxStash = len(cur.stash)
	}

	for cur.pos < 0 { // in a new `frame`, we start at end of `stash` (callee) and then travel down the args until below 0
		if idxframe == 0 {
			goto allDoneThusReturn // initial `expr` maximally reduced: return.
		} else { // jump back up to parent call `frame`, dropping the current one
			parent := &frames[idxframe-1]
			parent.stash[parent.pos] = cur.stash[idxcallee]               // store result there
			cur, frames, idxframe = parent, frames[:idxframe], idxframe-1 // now we're in `parent`
			idxcallee = len(cur.stash) - 1
		}
	}

	switch it := cur.stash[cur.pos].(type) {

	case nil: // a will-be-discarded call-arg slot. was cleared when the callee resolved to a final callable
		cur.pos-- // we arrived here as our `pos` counts down, and keep going

	case ExprNumInt: // a no-further-reducable final value
		cur.pos-- // count down as well to travel further down the `stash`

	case ExprArgRef:
		lookupstash := frames[cur.argsFrame].stash // most common case
		if cur.calleeDone {                        // very rare case:
			lookupstash = frames[idxframe].stash // essentially callees that merely return one of their args as-is
		}
		cur.stash[cur.pos] = lookupstash[len(lookupstash)+int(it)]
		if cur.pos == idxcallee { // in callee pos? then:
			goto restep // whatever we got, we want to further evaluate: no need for the final post-`switch` checks on `cur.pos` since it hasn't changed, can go at it again right away
		} else {
			cur.pos-- // no need to further evaluate because arg-values from `lookupstash` are already "maximally reduced for not-in-callee-pos purposes"
		}

	case *ExprCall:
		if it.IsClosure != 0 { // if so: a currently-no-further-reducable final value (closure)
			cur.pos--
		} else { // build up & add & enter the next `frame`
			callee, callargs := it.Callee, append(make([]Expr, 0, 3+len(it.Args)), it.Args...)
			for sub, isc := callee.(*ExprCall); isc; sub, isc = callee.(*ExprCall) { // flatten to single call
				callee, callargs = sub.Callee, append(callargs, sub.Args...)
			}
			lookupframe := cur.argsFrame // same logic as above in `case` of `ExprArgRef`:
			if cur.calleeDone {          // ...but this now occurs ~50/50
				lookupframe = idxframe
			}
			idxframe, frames = idxframe+1, append(frames, frame{
				pos: len(callargs), stash: append(callargs, callee), argsFrame: lookupframe})
			cur = &frames[idxframe] // now enter the newly created `frame`
			if idxframe > maxFrames {
				maxFrames = idxframe
			}
			goto restep
		}

	case ExprFuncRef: // recall: if it<0 the `ExprFuncRef` refers to an `OpCode`
		// note, scenario of `me[it].isMereAlias && 0 == len(me[it].Args)` will never occur thanks to load-time `Prog.postLoadPreProcess` call.
		if cur.calleeDone || cur.pos != idxcallee { // either not in callee position or else callee reduced to current `it`?
			cur.pos-- // then the `ExprFuncRef` is a mere currently-no-further-reducable value to just pass along / return / preserve for now
		} else /* we are in callee position */ if isfn := it > -1; cur.numArgs == 0 { // then must determine this now, first!
			cur.numArgs = 2 // prim-op default
			if isfn {       // refers to actual func, not prim-op
				cur.numArgs = len(me[it].Args)
				// optional micro-optimization block: entered-into for approx. 25% - 35% of cases here
				if me[it].selector != 0 && len(cur.stash) > cur.numArgs {
					if me[it].selector < 0 {
						selected := cur.stash[len(cur.stash)+me[it].selector]
						cur.stash = append(cur.stash[:idxcallee-cur.numArgs], selected)
					} else {
						call, _ := me[it].Body.(*ExprCall)
						argref, _ := call.Callee.(ExprArgRef)
						newtail := make([]Expr, 1+len(call.Args))
						newtail[len(call.Args)] = cur.stash[len(cur.stash)+int(argref)]
						for i := range call.Args {
							argref, _ = call.Args[i].(ExprArgRef)
							newtail[i] = cur.stash[len(cur.stash)+int(argref)]
						}
						cur.stash = append(cur.stash[:idxcallee-cur.numArgs], newtail...)
					}
					if numargsdone -= cur.numArgs; numargsdone < 0 {
						numargsdone = 0
					}
					cur.numArgs, cur.pos = 0, len(cur.stash)-1
					goto restep
				}
				if cur.numArgs == 0 { // no args means a shared global constant:
					call, _ := me[it].Body.(*ExprCall) // also means *ExprCall because others were caught during post-load pre-processing
					cur.stash = append(append(cur.stash[:idxcallee], call.Args...), call.Callee)
					if cur.pos = len(cur.stash) - 1; call.IsClosure == 0 {
						numargsdone = 0
					} else {
						numargsdone += len(call.Args)
					}
					goto restep
				} else if !me[it].allArgsUsed { // then ditch unused ones: by setting unused arg-slots in `stash` to `nil`
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
			}
			cur.pos, numargsdone = cur.pos-(1+numargsdone), 0 // jump down to first arg that needs eval-ing, will then count down from there
		} else if len(cur.stash) > cur.numArgs { // with all args eval'd, now comes the callee's body
			var result Expr
			if isfn { // substitution
				result = me[it].Body
			} else { // prim-op instruction code: consume left-hand-side and right-hand-side operands
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
					if result = StdFuncFalse; Eq(lhs, rhs) {
						result = StdFuncTrue
					}
				case OpPrt:
					result = rhs
					_, _ = OpPrtDst(append(append(append(ListToBytes(ListOfExprs(lhs)), '\t'), ListOfExprsToString(rhs)...), '\n'))
				case OpEval:
					prog, jsonprog, jsonexpr := me, decodeJsonishProgForOpEval(lhs), decodeJsonishExprForOpEval(rhs)
					if jsonprog != nil {
						prog = loadFromJson(jsonprog)
					}
					result, _ = prog.eval(exprFromJson(jsonexpr, 0), 128)
				default:
					panic([3]Expr{it, lhs, rhs})
				}
			}
			cur.calleeDone, cur.stash[idxcallee] = true, result
			goto restep // whatever we got here now, reduce it further until no longer reducible
		} else {
			cur.pos-- // in callee position but callee was already done: the post-`switch` checks below handle that if counting one down here, so we do
		}
	} // done type-switch on cur.stash[cur.pos]

	if idxcallee != 0 && cur.pos < idxcallee { // still arg-ful and below callee-position
		if cur.argsDone { // below callee position. have all args already eval'd previously? so return then, `calleeDone` or not (50/50)
			if result, diff := cur.stash[idxcallee], cur.numArgs-idxcallee; diff < 1 { // the `calleeDone` (non-closure) case:
				cur.stash = append(cur.stash[:(len(cur.stash)-1)-cur.numArgs], result) // if extraneous args were around, then len(cur.stash) > 1 now still, so our `frame` is not done yet
			} else /* result is closure */ if ilp := idxframe - 1; ilp > 0 && frames[ilp].numArgs == 0 && len(frames[ilp].stash) != 1 && frames[ilp].pos == len(frames[ilp].stash)-1 {
				// this block optional micro-optimization: unroll into parent's `stash` instead of alloc'ing a new `ExprCall`
				callee, callargs := result, cur.stash[:idxcallee]
				cur, idxframe, numargsdone, frames = &frames[ilp], ilp, len(callargs), frames[:idxframe]
				cur.stash = append(append(cur.stash[:len(cur.stash)-1], callargs...), callee)
				cur.pos = len(cur.stash) - 1
				goto restep
			} else { // still closure case
				result = &ExprCall{IsClosure: diff, Callee: result, Args: cur.stash[:idxcallee]}
				cur.stash[idxcallee] = result
				cur.stash = cur.stash[idxcallee:] // now 1 == len(cur.stash)
			}
			cur.calleeDone, cur.numArgs, cur.argsDone = false, 0, false
			if len(cur.stash) == 1 { // is this `frame` done now?
				cur.pos = -1 // caught at top of next `restep` iteration to push result back up to caller
			} else { // we had extra args so another round of this (now smaller) `frame`
				cur.pos = len(cur.stash) - 1
			}
		} else if cur.numArgs == 0 { // callee was not an `ExprFuncRef` so must be a closure:
			closure := cur.stash[idxcallee].(*ExprCall) // ... so unroll it into current `stash` :
			cur.stash = append(append(cur.stash[:idxcallee], closure.Args...), closure.Callee)
			numargsdone, cur.pos = len(closure.Args), len(cur.stash)-1 // ... and start over at callee
		} else if cur.pos < 0 || cur.pos < idxcallee-cur.numArgs { // all args needed were eval'd:
			cur.pos, cur.argsDone = idxcallee, true // note it down, and jump back to callee for eval'ing
		}
	}
	goto restep

allDoneThusReturn:
	return frames[0].stash[0], starttime
}
