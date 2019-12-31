package atem

// Eq is the implementation of the `OpEq` prim-op instruction code.
func Eq(expr Expr, cmp Expr) bool {
	if expr == cmp { // rare but can happen depending on program
		return true
	}
	switch it := expr.(type) {
	case ExprNumInt:
		that, ok := cmp.(ExprNumInt)
		return ok && it == that
	case *ExprCall:
		if that, ok := cmp.(*ExprCall); ok {
			if ok = (len(it.Args) == len(that.Args)) && Eq(it.Callee, that.Callee); ok {
				for i := range it.Args {
					if ok = Eq(it.Args[i], that.Args[i]); !ok {
						break
					}
				}
			}
			return ok
		}
	case ExprFuncRef:
		that, ok := cmp.(ExprFuncRef)
		return ok && it == that
	case ExprArgRef:
		that, ok := cmp.(ExprArgRef)
		return ok && it == that
	}
	return false
}

// ListOfExprs dissects the given `expr` into an `[]Expr` slice only if it is
// a closure resulting from `StdFuncCons` / `StdFuncNil` usage during `Eval`.
// The individual element `Expr`s are not themselves scrutinized however.
// The `ret` is `return`ed as `nil` if `expr` isn't a product of `StdFuncCons`
// / `StdFuncNil` usage; yet a non-`nil`, zero-`len` `ret` will result from a
// mere `StdFuncNil` construction, aka. "empty linked-list value" `Expr`.
//
// The result of `ListOfExprs` can be passed to `ListToBytes` to extract the
// `string` value represented by `expr`, if any.
func ListOfExprs(expr Expr) (ret []Expr) {
	ret = make([]Expr, 0, 1024)
	for ok, next := true, expr; ok; {
		ok = false
		if fnref, _ := next.(ExprFuncRef); fnref == StdFuncNil {
			break
		} else if call, okc := next.(*ExprCall); okc && len(call.Args) == 2 {
			if fnref, _ = call.Callee.(ExprFuncRef); fnref == StdFuncCons {
				ok, next, ret = true, call.Args[0], append(ret, call.Args[1])
			}
		}
		if !ok {
			ret = nil
		}
	}
	return
}

// ListToBytes examines the given `[]Expr`, as normally obtained via
// `ListOfExprs` and accumulates a `[]byte` slice as long as all elements
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

// ListOfExprsToString is a wrapper around the combined usage of `ListOfExprs`
// and `ListToBytes` to extract the List-closure-encoded `string` of an `Eval`
// result, if it is one. Otherwise, `expr.JsonSrc()` is returned for convenience.
func ListOfExprsToString(expr Expr) string {
	if maybenumlist := ListOfExprs(expr); maybenumlist != nil {
		if bytes := ListToBytes(maybenumlist); bytes != nil {
			return string(bytes)
		}
	}
	return expr.JsonSrc()
}

// ListFrom converts the specified byte string to a linked-list representing a text string during `Eval` (via `ExprCall`s of `StdFuncCons` and `StdFuncNil`).
func ListFrom(str []byte) (ret Expr) {
	ret = StdFuncNil
	for i := len(str) - 1; i > -1; i-- {
		ret = &ExprCall{IsClosure: 2, Callee: StdFuncCons, Args: []Expr{ret, ExprNumInt(str[i])}}
	}
	return
}

// ListsFrom creates from `strs` linked-lists via `ListFrom`, and returns a linked-list of those.
func ListsFrom(strs []string) (ret Expr) {
	ret = StdFuncNil
	for i := len(strs) - 1; i > -1; i-- {
		ret = &ExprCall{IsClosure: 2, Callee: StdFuncCons, Args: []Expr{ret, ListFrom([]byte(strs[i]))}}
	}
	return
}

func decodeProgForOpEval(expr Expr) (prog [][]interface{}) {
	if lprog := ListOfExprs(expr); lprog == nil {
		panic(expr)
	} else if len(lprog) != 0 {
		prog = make([][]interface{}, len(lprog))
		for i := range lprog {
			if lfunc := ListOfExprs(lprog[i]); lfunc == nil {
				panic(lfunc)
			} else {

			}
		}
	}

	return
}

func decodeExprForOpEval(expr Expr) interface{} {
	return nil
}
