package atem

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
)

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
			lhs, rhs := me.Eval(stack[len(stack)-1], stack).(ExprNumInt), me.Eval(stack[len(stack)-2], stack).(ExprNumInt)
			stack = stack[:len(stack)-2]
			switch opcode := OpCode(it); opcode {
			case OpAdd:
				return lhs + rhs
			case OpSub:
				return lhs - rhs
			case OpMul:
				return lhs * rhs
			case OpDiv:
				return lhs / rhs
			case OpMod:
				return lhs % rhs
			case OpEq, OpGt, OpLt:
				if (opcode == OpEq && lhs == rhs) || (opcode == OpLt && lhs < rhs) || (opcode == OpGt && lhs > rhs) {
					it, numargs = 1, 2
				} else {
					it, numargs = 2, 2
				}
			default:
				panic(opcode)
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
