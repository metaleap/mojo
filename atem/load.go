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
// is a JSON string parseable into an integer, and `ExprCall` is a variable
// length (greater than 1) array of any of those possibilities.
// A `panic` occurs on any sort of error encountered from the input `src`.
//
// A note on `ExprCall`s, their `Args` orderings are reversed from the JSON
// one being read in or emitted back out via `JsonSrc()`. Args in the JSON
// format are like in any common notation: `[callee, arg1, arg2, arg3]`, but an
// `ExprCall` created from this will have an `Args` slice of `[arg3, arg2, arg1]`
// throughout its lifetime. Still, its `JsonSrc()` emits the original ordering.
// If the callee is another `ExprCall`, expect a JSON source notation of eg.
// `[[callee, x, y, z], a, b, c]` to turn into a single `ExprCall` with `Args`
// of [c, b, a, z, y, x], it would be re-emitted as `[callee, x, y, z, a, b, c]`.
// In any event, `ExprCall.Args` and `FuncDef.Args` orderings shall be consistent
// in the JSON source code format regardless of these run time re-orderings.
//
// A note on `ExprArgRef`s: these take different forms in the JSON format and
// at runtime. In the former, two intuitive-to-emit styles are supported: if
// positive they denote 0-based indexing such that 0 refers to the `FuncDef`'s
// first arg, 1 to the second, 2 to the third etc; if negative, they're read
// with -1 referring to the `FuncDef`'s last arg, -2 to the one-before-last, -3 to
// the one-before-one-before-last etc. Both styles at load time are translated
// into a form expected at run time, where 0 turns into -1, 1 into -2, 2 into
// -3 etc, allowing for smoother stack accesses in the interpreter.
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
		callee, args := exprFromJson(it[0], curFnNumArgs), make([]Expr, 0, len(it)-1)
		for i := len(it) - 1; i > 0; i-- {
			args = append(args, exprFromJson(it[i], curFnNumArgs))
		}
		if subcall, _ := callee.(*ExprCall); subcall == nil {
			return &ExprCall{Callee: callee, Args: args}
		} else {
			subcall.Args = append(args, subcall.Args...)
			return subcall
		}
	}
	panic(from)
}
