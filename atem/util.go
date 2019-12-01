package atem

// Eq is the fallback for `OpEq` calls with 2 operands that aren't both `ExprNumInt`s.
func (me Prog) Eq(expr Expr, cmp Expr, evalAppls bool) bool {
	switch it := expr.(type) {
	case ExprNumInt:
		that, ok := cmp.(ExprNumInt)
		return ok && it == that
	case ExprAppl:
		if that, ok := cmp.(ExprAppl); ok {
			if evalAppls {
				return me.Eq(me.eval(it.Callee, nil), me.eval(that.Callee, nil), true) && me.Eq(me.eval(it.Arg, nil), me.eval(that.Arg, nil), true)
			} else {
				return me.Eq(it.Callee, that.Callee, false) && me.Eq(it.Arg, that.Arg, false)
			}
		}
	case ExprArgRef:
		that, ok := cmp.(ExprArgRef)
		return ok && (it == that || (it < 0 && that >= 0 && that == (-it)-1) || (it >= 0 && that < 0 && it == (-that)-1))
	case ExprFuncRef:
		that, ok := cmp.(ExprFuncRef)
		return ok && it == that
	}
	return expr == cmp
}

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
					elem := me.eval(ainner.Arg, nil)
					again, next, ret = true, me.eval(aouter.Arg, nil), append(ret, elem)
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

// ListFrom converts the specified byte string to a linked-list representing a text string during `Eval` (via `ExprAppl`s of `StdFuncCons` and `StdFuncNil`).
func ListFrom(str []byte) (ret Expr) {
	ret = StdFuncNil
	for i := len(str) - 1; i > -1; i-- {
		ret = ExprAppl{Callee: ExprAppl{Callee: StdFuncCons, Arg: ExprNumInt(str[i])}, Arg: ret}
	}
	return
}

// ListsFrom creates from `strs` linked-lists via `ListFrom`, and returns a linked-list of those.
func ListsFrom(strs []string) (ret Expr) {
	ret = StdFuncNil
	for i := len(strs) - 1; i > -1; i-- {
		ret = ExprAppl{Callee: ExprAppl{Callee: StdFuncCons, Arg: ListFrom([]byte(strs[i]))}, Arg: ret}
	}
	return
}
