package atem

import (
	"io"
	"os"
)

type OpCode int

const (
	OpAdd OpCode = -1
	OpSub OpCode = -2
	OpMul OpCode = -3
	OpDiv OpCode = -4
	OpMod OpCode = -5
	OpEq  OpCode = -6
	OpLt  OpCode = -7
	OpGt  OpCode = -8
	OpPrt OpCode = -42
)

var OpPrtDst io.Writer = os.Stderr

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
// this is met, the primitive instruction is carried out. For "boolish" prim ops,
// namely `OpEq`, `OpGt`, `OpLt` the resulting `StdFuncFalse` or `StdFuncTrue`
// is directly applied (as described above) with the remainder of the `stack`.
// For "integer" prim ops (`OpAdd`, `OpSub` etc.) the 2 operands are forced into
// `ExprNumInt`s and an `ExprNumInt` result will be returned. For `OpPrt`,
// the side-effect write of both operands to `OpPrtDst` is performed and the
// second operand is then returned. Other / unknown op-codes `panic` with a
// `[3]Expr` of first the `ExprFuncRef` followed by both its operands.
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
			lhs, rhs := me.Eval(stack[len(stack)-1], stack), me.Eval(stack[len(stack)-2], stack)
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
			case OpEq, OpGt, OpLt:
				if l, r := lhs.(ExprNumInt), rhs.(ExprNumInt); (opcode == OpEq && l == r) || (opcode == OpLt && l < r) || (opcode == OpGt && l > r) {
					expr = StdFuncTrue
				} else {
					expr = StdFuncFalse
				}
			case OpPrt:
				OpPrtDst.Write(append(append(ListToBytes(me.ListOfExprs(lhs)), '\t'), me.ListOfExprsToString(rhs)...))
				expr = rhs
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
