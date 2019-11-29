package atem

import (
	"encoding/json"
	"strconv"
)

type any = interface{} // just for less-noisily-reading JSON-unmarshalings below

// LoadFromJson parses and decodes a JSON `src` into an atem `Prog`. The format is
// expected to be: `[ func, func, ... , func ]` where `func` means: ` [ args, body ]`
// where `args` is a numbers array and `body` is the reverse of each concrete
// `Expr` implementer's `JsonSrc` method implementation, meaning: `ExprNumInt`
// is a JSON number, `ExprFuncRef` is a length-1 numbers array, `ExprArgRef`
// is a JSON string parseable into an integer, and `ExprAppl` is a variable
// length (greater than 1) array of any of those possibilities.
// A `panic` occurs on any sort of error encountered from the input `src`.
//
// A note on `ExprArgRef`s: these take different forms in the JSON format and
// at runtime. In the former, two intuitive-to-emit styles are supported: if
// positive they denote 0-based indexing such that 0 refers to the `FuncDef`'s
// first arg, 1 to the second, 2 to the third etc; if negative, they're translated
// into this just-mentioned positive format by treating them as De Brujin indices,
// with -1 referring to the `FuncDef`'s last arg, -2 to the one-before-last, -3 to
// the one-before-one-before-last etc. However at parse time, they're turned
// into a form expected at run time, where 0 turns into -1, 1 into -2, 2 into
// -3 etc, allowing for faster stack accesses in the interpreter.
// `ExprArgRef.JsonSrc()` will restore the 0-based indexing form, however.
func LoadFromJson(src []byte) Prog {
	arr := make([][]any, 0, 128)
	if e := json.Unmarshal(src, &arr); e != nil {
		panic(e)
	}
	me := make(Prog, 0, len(arr))
	for _, it := range arr {
		meta, arrargs, args := []string{}, it[1].([]any), make([]int, 0, 0)
		if metarr, _ := it[0].([]any); len(metarr) > 0 {
			for _, mstr := range metarr {
				meta = append(meta, mstr.(string))
			}
		}
		for _, v := range arrargs {
			args = append(args, int(v.(float64)))
		}
		me = append(me, FuncDef{Args: args, Body: exprFromJson(it[2], int64(len(args))), Meta: meta})
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
			if n < 0 || n >= curFnNumArgs {
				panic("LoadFromJson: encountered bad ExprArgRef of " + strconv.FormatInt(n, 10) + " inside a FuncDef with " + strconv.FormatInt(curFnNumArgs, 10) + " arg(s)")
			}
			return ExprArgRef(int(-(n + 1))) // rewrite arg-refs for later stack-access-from-tail-end: 0 -> -1, 1 -> -2, 2 -> -3, etc.. note: reverted again in ExprArgRef.JsonSrc()
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
