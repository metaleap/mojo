package atem

import (
	"encoding/json"
	"strconv"
)

type any = interface{} // just for less-noisily-reading JSON-unmarshalings below

func LoadFromJson(src []byte) Prog {
	arr := make([][]any, 0, 128)
	if e := json.Unmarshal(src, &arr); e != nil {
		panic(e)
	}
	me := make(Prog, 0, len(arr))
	for _, it := range arr {
		arrargs, args := it[1].([]any), make([]int, 0, 8)
		for _, v := range arrargs {
			args = append(args, int(v.(float64)))
		}
		me = append(me, FuncDef{args, exprFromJson(it[2], int64(len(args))), it[0].(string)})
	}
	return me
}

func exprFromJson(from any, curFnNumArgs int64) Expr {
	switch it := from.(type) {
	case float64:
		return ExprNumInt(int(it))
	case string:
		if n, e := strconv.ParseInt(it, 10, 0); e != nil {
			panic(e)
		} else {
			if n < 0 {
				n = curFnNumArgs + n
			}
			return ExprArgRef(int(-(n + 1))) // rewrite arg-refs for later stack-access-from-tail-end: 0 -> -1, 1 -> -2, 2 -> -3
		}
	case []any:
		if len(it) == 1 {
			return ExprFuncRef(int(it[0].(float64)))
		}
		expr := exprFromJson(it[0], curFnNumArgs)
		for i := 1; i < len(it); i++ {
			expr = ExprCall{expr, exprFromJson(it[i], curFnNumArgs)}
		}
		return expr
	case map[string]any: // allows for free-form annotations / comments / meta-data ...
		return exprFromJson(it[""], curFnNumArgs) // ... by digging into this field and ignoring all others
	}
	panic(from)
}
