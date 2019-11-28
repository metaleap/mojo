package atem

// ListOfExprs dissects the given `expr` into an `[]Expr` slice only if it is
// a closure resulting from `StdFuncCons` / `StdFuncNil` usage during `Eval`.
// The individual element `Expr`s are not themselves scrutinized however.
// The `ret` is `return`ed as `nil` if `expr` isn't a product of `StdFuncCons`
// / `StdFuncNil` usage; yet a non-`nil`, zero-`len` `ret` will result from a
// mere `StdFuncNil` construction, aka. "empty linked-list value" `Expr`.
func (me Prog) ListOfExprs(expr Expr) (ret []Expr) {
	ret = make([]Expr, 0, 1024)
	for again, next := true, expr; again; {
		again = false
		if fouter, ok0 := next.(ExprFuncRef); ok0 && fouter == StdFuncNil { // clean end-of-list
			break
		} else if aouter, ok1 := next.(ExprAppl); ok1 {
			if ainner, ok2 := aouter.Callee.(ExprAppl); ok2 {
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

// ListToBytes examines the given `[]Expr`, as normally obtained via
// `Prog.ListOfExprs` and accumulates a `[]byte` slice as long as all elements
// in said list are `ExprNumInt` values in the range 0 - 255. If the input is
// `nil`, so will be `retNumListAsBytes`. If the input has a `len` of zero,
// so will `retNumListAsBytes`. If any of the input `Expr`s isn't an in-range
// `ExprNumInt`, then too will `retNumListAsBytes` be `nil`.
func ListToBytes(maybeNumList []Expr) (retNumListAsBytes []byte) {
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

// ListOfExprsToString is a wrapper around the combined usage of `Prog.ListOfExprs`
// and `ListToBytes` to extract the List-closure-encoded `string` of an `Eval`
// result, if it is one. Otherwise, `expr.JsonSrc()` is returned for convenience.
func (me Prog) ListOfExprsToString(expr Expr) string {
	if maybenumlist := me.ListOfExprs(expr); maybenumlist != nil {
		if bytes := ListToBytes(maybenumlist); bytes != nil {
			return string(bytes)
		}
	}
	return expr.JsonSrc()
}
