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

func (me Prog) Eval(expr Expr, stack []Expr) Expr {
	switch it := expr.(type) {
	case ExprCall:
		return me.Eval(it.Callee, append(stack, it.Arg))
	case ExprFuncRef:
		numargs, isopcode := 2, (it < 0)
		if !isopcode {
			numargs = len(me[it].Args)
		}
		if len(stack) < numargs { // not enough args on stack: a partial-application aka closure
			for i := len(stack) - 1; i >= 0; i-- {
				expr = ExprCall{expr, stack[i]} // OPT: can keep ref to top-app instead of recreating it from stack
			}
			return expr
		} else if isopcode {
			lhs, rhs := me.Eval(stack[len(stack)-1], stack), me.Eval(stack[len(stack)-2], stack)
			stack = stack[:len(stack)-2]
			switch opcode := OpCode(it); opcode {
			case OpAdd:
				return lhs.(ExprNumInt) + rhs.(ExprNumInt)
			case OpSub:
				return lhs.(ExprNumInt) - rhs.(ExprNumInt)
			case OpMul:
				return lhs.(ExprNumInt) * rhs.(ExprNumInt)
			case OpDiv:
				return lhs.(ExprNumInt) / rhs.(ExprNumInt)
			case OpMod:
				return lhs.(ExprNumInt) % rhs.(ExprNumInt)
			case OpEq, OpGt, OpLt:
				if l, r := lhs.(ExprNumInt), rhs.(ExprNumInt); (opcode == OpEq && l == r) || (opcode == OpLt && l < r) || (opcode == OpGt && l > r) {
					it, numargs = 1, 2
				} else {
					it, numargs = 2, 2
				}
			case OpPrt:
				OpPrtDst.Write(Bytes(me.List(lhs)))
				return rhs
			default:
				panic([2]Expr{lhs, rhs})
			}
		}
		return me.Eval(argRefsResolvedToCurrentStackEntries(me[it].Body, stack), stack[:len(stack)-numargs])
	}
	return expr
}

func argRefsResolvedToCurrentStackEntries(expr Expr, stack []Expr) Expr {
	switch it := expr.(type) {
	case ExprCall: // OPT: selector funcs!
		return ExprCall{argRefsResolvedToCurrentStackEntries(it.Callee, stack), argRefsResolvedToCurrentStackEntries(it.Arg, stack)}
	case ExprArgRef:
		return stack[len(stack)+int(it)]
	}
	return expr
}
