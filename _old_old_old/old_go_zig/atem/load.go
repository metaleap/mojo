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
// A note on `ExprCall`s, their `Args` orderings are on-load reversed from those
// being read in or emitted back out via `JsonSrc()`. Args in the JSON format are
// ordered in a common intuitive manner: `[callee, arg1, arg2, arg3]`, but an
// `ExprCall` created from this will have an `Args` slice of `[arg3, arg2, arg1]`
// throughout its lifetime. Still, its `JsonSrc()` emits the original ordering.
// If the callee is another `ExprCall`, expect a JSON source notation of eg.
// `[[callee, x, y, z], a, b, c]` to turn into a single `ExprCall` with `Args`
// of [c, b, a, z, y, x], it would be re-emitted as `[callee, x, y, z, a, b, c]`.
// `ExprCall.Args` and `FuncDef.Args` orderings are consistent in the JSON
// source code format (when loading or emitting), but not at run time.
//
// A note on `ExprArgRef`s: these take different forms in the JSON format and
// at runtime. In the former, two intuitive-to-emit styles are supported: if
// positive they denote 0-based indexing such that 0 refers to the `FuncDef`'s
// first arg, 1 to the second, 2 to the third etc; if negative, they're read
// with -1 referring to the `FuncDef`'s last arg, -2 to the one-before-last, -3 to
// the one-before-one-before-last etc. Both styles at load time are translated
// into a form expected at run time, where 0 turns into -2, 1 into -3, 2 into
// -4 etc, for marginally speedier call-stack accesses in the interpreter.
// `ExprArgRef.JsonSrc()` will restore the 0-based indexing form, however.
func LoadFromJson(src []byte) Prog {
	arr := make([][]any, 0, 512)
	if e := json.Unmarshal(src, &arr); e != nil {
		panic(e)
	}
	return loadFromJson(arr)
}

func loadFromJson(arr [][]interface{}) Prog {
	me := make(Prog, 0, len(arr))
	for _, it := range arr {
		arrargs := it[1].([]any)
		fd := FuncDef{Body: exprFromJson(it[2], int64(len(arrargs))), allArgsUsed: true, Meta: []string{}, Args: make([]int, len(arrargs))}
		if metarr, _ := it[0].([]any); len(metarr) > 0 {
			for _, mstr := range metarr {
				fd.Meta = append(fd.Meta, mstr.(string))
			}
		}
		for i, v := range arrargs {
			if fd.Args[i] = int(v.(float64)); 0 == fd.Args[i] {
				fd.allArgsUsed = false
			}
		}
		me = append(me, fd)
	}
	for i := range me {
		me.postLoadPreProcess(i)
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
				n += curFnNumArgs // now positive starting from zero (if it was correct to begin with)
			}
			if n < 0 || n >= curFnNumArgs {
				panic("LoadFromJson: encountered bad ExprArgRef of " + strconv.FormatInt(n, 10) + " inside a FuncDef with " + strconv.FormatInt(curFnNumArgs, 10) + " arg(s)")
			}
			return ExprArgRef(int(-(n + 2))) // rewrite arg-refs for later stack-access-from-tail-end: 0 -> -2, 1 -> -3, 2 -> -4, etc.. note: reverted again in ExprArgRef.JsonSrc()
		}
	case []any:
		if len(it) == 1 { // func-ref literal
			return ExprFuncRef(int(it[0].(float64)))
		}
		callee, args := exprFromJson(it[0], curFnNumArgs), make([]Expr, 0, len(it))
		for i := len(it) - 1; i > 0; i-- {
			arg := exprFromJson(it[i], curFnNumArgs)
			args = append(args, arg)
		}
		var ret *ExprCall
		if subcall, _ := callee.(*ExprCall); subcall == nil {
			ret = &ExprCall{Callee: callee, Args: args}
		} else {
			subcall.Args = append(args, subcall.Args...)
			ret = subcall
		}
		return ret
	}
	panic(from)
}

func (me Prog) postLoadPreProcess(funcIdx int) {
	fd := &me[funcIdx]
	if len(fd.Args) == 0 {
		_, iscall := fd.Body.(*ExprCall)
		fd.isMereAlias = !iscall
	} else if len(fd.Args) >= 2 { // check if selector and set so
		if argref, isa := fd.Body.(ExprArgRef); isa {
			fd.selector = int(argref)
		} else if call, isc := fd.Body.(*ExprCall); isc {
			if argref, isa = call.Callee.(ExprArgRef); isa && argref != -2 {
				for ia := range call.Args {
					if _, isa = call.Args[ia].(ExprArgRef); !isa {
						break
					}
				}
				if isa {
					fd.selector = (len(call.Args))
				}
			}
		}
	}
	fd.Body = me.detectAndMarkClosures(fd.Body)
}

func (me Prog) detectAndMarkClosures(expr Expr) Expr {
	for fnr, _ := expr.(ExprFuncRef); fnr > 0 && me[fnr].isMereAlias; fnr, _ = expr.(ExprFuncRef) {
		// main reason for this pre-reduction is to not need this check plus
		// potential reduction in the interpreter loop on every `ExprFuncRef`
		// occurrence. especially in pre-optimized input programs the check
		// would always be incurred but never "fire". to also allow
		// less-/non-optimized input programs, we pre-reduce this post-load.
		expr = me[fnr].Body
	}
	// now, detect and mark all recognizable "closures" as defined in the doc-comment for the `ExprCall.IsClosure` field
	if call, iscall := expr.(*ExprCall); iscall {
		call.Callee = me.detectAndMarkClosures(call.Callee)
		for i := range call.Args {
			call.Args[i] = me.detectAndMarkClosures(call.Args[i])
		}
		if f, _ := call.Callee.(ExprFuncRef); f != 0 {
			numargs := (2)
			if f > 0 {
				numargs = (len(me[f].Args))
			}
			diff := numargs - (len(call.Args))
			for i := 0; (diff > 0) && (i < len(call.Args)); i++ {
				_, isa := call.Args[i].(ExprArgRef)
				if c, isc := call.Args[i].(*ExprCall); isa || (isc && c.IsClosure == 0) {
					diff = 0
				}
			}
			if call.IsClosure = diff; diff < 0 {
				call.IsClosure = 0
			}
		}
	}
	return expr
}
