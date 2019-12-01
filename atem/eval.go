package atem

import (
	"os"
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
	// Less-than test between 2 `Expr`s, result is `StdFuncTrue` or `StdFuncFalse`
	OpLt OpCode = -7
	// Greater-than test between 2 `Expr`s, result is `StdFuncTrue` or `StdFuncFalse`
	OpGt OpCode = -8
	// Writes both `Expr`s (the first one a string-ish `StdFuncCons`tructed linked-list of `ExprNumInt`s) to `OpPrtDst`, result is the right-hand-side `Expr` of the 2 input `Expr` operands
	OpPrt OpCode = -42
)

// OpPrtDst is the output destination for all `OpPrt` primitive instructions.
// Must never be `nil` during any `Prog`s that do potentially invoke `OpPrt`.
var OpPrtDst = os.Stderr.Write

// Eval operates thusly:
//
// - encountering an `ExprAppl`, its `Arg` is `append`ed to the `stack` and
// its `Callee` is then `Eval`'d;
//
// - encountering an `ExprFuncRef`, the `stack` is checked for having the
// proper minimum required `len` with regard to the referenced `FuncDef`'s
// number of `Args`. If okay, the pertinent number of args is taken (and
// removed) from the `stack` and the referenced `FuncDef`'s `Body`, rewritten
// with all inner `ExprArgRef`s (including those inside `ExprAppl`s) resolved
// to the `stack` entries, is `Eval`'d (with the appropriately reduced `stack`);
//
// - encountering  any other `Expr` type, it is merely returned.
//
// Corner cases for the `ExprFuncRef` situation: if the `stack` has too small
// a `len`, an `ExprAppl` representing the partial application is returned;
// if the `ExprFuncRef` is negative and thus referring to a primitive-instruction
// `OpCode`, the expected minimum required `len` for the `stack` is 2 and if
// this is met, the primitive instruction is carried out, its `Expr` result
// then being `Eval`'d with the reduced-by-2 `stack`. Unknown op-codes `panic`
// with a `[3]Expr` of first the `OpCode`-referencing `ExprFuncRef` followed
// by both its operands.
func (me Prog) Eval(expr Expr, stack []Expr) Expr {
	switch it := expr.(type) {
	case ExprAppl:
		return me.Eval(it.Callee, append(stack, it.Arg))
	case ExprFuncRef:
		numargs, isopcode := 2, (it < 0)
		if !isopcode {
			numargs = len(me[it].Args)
		}
		if len(stack) < numargs { // not enough args on stack: a partial-application aka closure
			for i := len(stack) - 1; i >= 0; i-- {
				expr = ExprAppl{expr, stack[i]} // OPT: can keep ref to top-app instead of recreating it from stack
			}
			return expr
		} else if isopcode {
			lhs, rhs := me.Eval(stack[len(stack)-1], nil), me.Eval(stack[len(stack)-2], nil)
			switch opcode := OpCode(it); opcode {
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
				if expr = StdFuncFalse; me.Eq(lhs, rhs, true) {
					expr = StdFuncTrue
				}
			case OpGt, OpLt:
				lt, l, r := (opcode == OpLt), lhs.(ExprNumInt), rhs.(ExprNumInt)
				if expr = StdFuncFalse; (lt && l < r) || ((!lt) && l > r) {
					expr = StdFuncTrue
				}
			case OpPrt:
				expr = rhs
				_, _ = OpPrtDst(append(append(append(ListToBytes(me.ListOfExprs(lhs)), '\t'), me.ListOfExprsToString(rhs)...), '\n'))
			default:
				panic([3]Expr{it, lhs, rhs})
			}
			return me.Eval(expr, stack[:len(stack)-2])
		} else {
			return me.Eval(exprRewrittenWithArgRefsResolvedToStackEntries(me[it].Body, stack), stack[:len(stack)-numargs])
		}
	}
	return expr
}

func exprRewrittenWithArgRefsResolvedToStackEntries(expr Expr, stack []Expr) Expr {
	switch it := expr.(type) {
	case ExprAppl: // OPT: selector funcs!
		return ExprAppl{exprRewrittenWithArgRefsResolvedToStackEntries(it.Callee, stack), exprRewrittenWithArgRefsResolvedToStackEntries(it.Arg, stack)}
	case ExprArgRef:
		return stack[len(stack)+int(it)]
	}
	return expr
}
