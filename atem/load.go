package atem

import (
	"encoding/json"
	"strconv"
)

type any = interface{} // just for less-noisily-reading JSON-unmarshalings below

// LoadFromJson parses and decodes a JSON `src` into an atem `Prog`. The format is
// expected to be: `[ func, func, ... , func ]` where `func` means: ` [ args, body ]`
// where `args` is a numbers array and `body` is the reverse of each concrete
// `Expr` implementer's `String` method implementation, meaning: `ExprNumInt`
// is a JSON number, `ExprFuncRef` is a length-1 numbers array, `ExprArgRef`
// is a JSON string parseable into an integer, and `ExprAppl` is a variable
// length (greater than 1) array of any of those possibilities.
// A `panic` occurs on any sort of error encountered from the input `src`.
func LoadFromJson(src []byte) Prog {
	arr := make([][]any, 0, 128)
	if e := json.Unmarshal(src, &arr); e != nil {
		panic(e)
	}
	me := make(Prog, 0, len(arr))
	for _, it := range arr {
		arrargs, args := it[0].([]any), make([]int, 0, 8)
		for _, v := range arrargs {
			args = append(args, int(v.(float64)))
		}
		me = append(me, FuncDef{Args: args, Body: exprFromJson(it[1], int64(len(args)))})
	}
	return me
}

func exprFromJson(from any, curFnNumArgs int64) Expr {
	switch it := from.(type) {
	case float64: // number literal
		return ExprNumInt(int(it))
	case string: // arg-ref
		if n, e := strconv.ParseInt(it, 10, 0); e != nil {
			panic(e)
		} else {
			if n < 0 { // support for de-brujin indices if negative
				n = curFnNumArgs + n // now positive starting from zero (if it was correct to begin with)
			}
			return ExprArgRef(int(-(n + 1))) // rewrite arg-refs for later stack-access-from-tail-end: 0 -> -1, 1 -> -2, 2 -> -3, etc..
		}
	case []any:
		if len(it) == 1 { // either func-ref literal..
			return ExprFuncRef(int(it[0].(float64)))
		}
		expr := exprFromJson(it[0], curFnNumArgs) // ..or else, func call aka. application
		for i := 1; i < len(it); i++ {
			expr = ExprAppl{expr, exprFromJson(it[i], curFnNumArgs)}
		}
		return expr
	}
	panic(from)
}
