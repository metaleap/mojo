package atem

import (
	"encoding/json"
	"strconv"
)

const (
	StdFuncId    ExprFuncRef = 0
	StdFuncTrue  ExprFuncRef = 1
	StdFuncFalse ExprFuncRef = 2
	StdFuncNil   ExprFuncRef = 3
	StdFuncCons  ExprFuncRef = 4
)

type (
	Prog    []FuncDef
	FuncDef struct {
		Args []int
		Body Expr
	}
	Expr        interface{ String() string }
	ExprNumInt  int
	ExprArgRef  int
	ExprFuncRef int
	ExprCall    struct {
		Callee Expr
		Arg    Expr
	}
	any = interface{} // just for a less-noisy JSON-unmarshaling part
)

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
		me = append(me, FuncDef{args, exprFromJson(it[1])})
	}
	return me
}

func exprFromJson(from any) Expr {
	switch it := from.(type) {
	case float64:
		return ExprNumInt(int(it))
	case string:
		if n, e := strconv.ParseInt(it, 10, 0); e != nil {
			panic(e)
		} else if n >= 0 { // rewrite arg-refs for later stack-access-from-tail-end: 0 -> -1, 1 -> -2, 2 -> -3
			return ExprArgRef(int(-(n + 1))) // only doing this for >=0s means compilers can emit whatever scheme is less hassle to emit, as long as consistent within a given func def.
		}
	case []any:
		if len(it) == 1 {
			return ExprFuncRef(int(it[0].(float64)))
		}
		expr := exprFromJson(it[0])
		for i := 1; i < len(it); i++ {
			expr = ExprCall{expr, exprFromJson(it[i])}
		}
		return expr
	case map[string]any: // allows for free-form annotations / comments / meta-data ...
		return exprFromJson(it[""]) // ... by digging into this field and ignoring all others
	}
	panic(from)
}

func (me ExprNumInt) String() string  { return strconv.Itoa(int(me)) }
func (me ExprArgRef) String() string  { return "\"" + strconv.Itoa(int(me)) + "\"" }
func (me ExprFuncRef) String() string { return "[" + strconv.Itoa(int(me)) + "]" }
func (me ExprCall) String() string    { return "[" + me.Callee.String() + ", " + me.Arg.String() + "]" }
func (me *FuncDef) String() string {
	outjson := "[ ["
	for i, a := range me.Args {
		if i > 0 {
			outjson += ","
		}
		outjson += strconv.Itoa(a)
	}
	return outjson + "], " + me.Body.String() + " ]"
}
func (me Prog) String() string {
	outjson := "[ "
	for i, def := range me {
		if i > 0 {
			outjson += ", "
		}
		outjson += def.String() + "\n"
	}
	return outjson + "]\n"
}
