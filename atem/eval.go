package main

func (me Prog) eval(expr Expr, stack []Expr) Expr {
	switch it := expr.(type) {
	case ExprAppl:
		return me.eval(it.Callee, append(stack, it.Arg))
	case ExprFnRef:
		numargs, isopcode := 2, (it < 0)
		if !isopcode {
			numargs = me[it].NumArgs
		}
		if len(stack) < numargs { // not enough args on stack: a partial-application aka closure
			for i := len(stack) - 1; i >= 0; i-- {
				expr = ExprAppl{expr, stack[i]}
			}
			return expr
		} else if isopcode {
			opcode := OpCode(it)
			lhs, rhs := me.eval(stack[len(stack)-1], stack).(ExprNum), me.eval(stack[len(stack)-2], stack).(ExprNum)
			stack = stack[:len(stack)-2]
			switch opcode {
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
				panic(stack)
			}
		}
		return me.eval(argRefsResolvedToCurrentStackEntries(me[it].Body, stack), stack[:len(stack)-numargs])
	}
	return expr
}

func argRefsResolvedToCurrentStackEntries(expr Expr, stack []Expr) Expr {
	switch it := expr.(type) {
	case ExprAppl:
		return ExprAppl{argRefsResolvedToCurrentStackEntries(it.Callee, stack), argRefsResolvedToCurrentStackEntries(it.Arg, stack)}
	case ExprArgRef:
		return stack[len(stack)+int(it)]
	}
	return expr
}
