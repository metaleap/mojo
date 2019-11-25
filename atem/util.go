package atem

func (me Prog) List(expr Expr) (ret []Expr) {
	ret = make([]Expr, 0, 1024)
	for again, next := true, expr; again; {
		again = false
		if fouter, ok0 := next.(ExprFuncRef); ok0 && fouter == StdFuncNil { // clean end-of-list
			break
		} else if aouter, ok1 := next.(ExprCall); ok1 {
			if ainner, ok2 := aouter.Callee.(ExprCall); ok2 {
				if finner, ok3 := ainner.Callee.(ExprFuncRef); ok3 && finner == StdFuncCons {
					elem := me.Eval(ainner.Arg, nil)
					again, next, ret = true, me.Eval(aouter.Arg, nil), append(ret, elem)
				}
			}
		}
		if !again {
			ret = nil
		}
	}
	return
}

func Bytes(maybeNumList []Expr) (retNumListAsBytes []byte) {
	if maybeNumList != nil {
		retNumListAsBytes = make([]byte, 0, len(maybeNumList))
		for _, expr := range maybeNumList {
			if num, ok := expr.(ExprNumInt); ok && num > -1 && num < 256 {
				retNumListAsBytes = append(retNumListAsBytes, byte(num))
			} else {
				retNumListAsBytes = nil
				break
			}
		}
	}
	return
}
