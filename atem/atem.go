package main

import (
	"encoding/json"
	"strconv"
)

type OpCode int

const (
	OpAdd OpCode = -1
	OpSub OpCode = -2
	OpMul OpCode = -3
	OpDiv OpCode = -4
	OpMod OpCode = -5
	OpEq  OpCode = -6
	OpLt  OpCode = -7
	OpGt  OpCode = -8
)

type Prog []FuncDef

type FuncDef struct {
	NumArgs int
	Body    Expr
}

type Expr interface{ String() string }

func (me ExprNum) String() string    { return strconv.Itoa(int(me)) }
func (me ExprArgRef) String() string { return "\"" + strconv.Itoa(int(me)) + "\"" }
func (me ExprFnRef) String() string  { return "[" + strconv.Itoa(int(me)) + "]" }
func (me ExprAppl) String() string   { return "[" + me.Callee.String() + ", " + me.Arg.String() + "]" }

type ExprNum int

type ExprArgRef int

type ExprFnRef int

type ExprAppl struct {
	Callee Expr
	Arg    Expr
}

type any = interface{}

func LoadFromJson(src []byte) Prog {
	arr := make([][]any, 0, 128)
	if e := json.Unmarshal(src, &arr); e != nil {
		panic(e)
	}
	me := make(Prog, 0, len(arr))
	for _, it := range arr {
		me = append(me, FuncDef{int(it[0].(float64)), exprFromJson(it[1])})
	}
	return me
}

func exprFromJson(from any) Expr {
	switch it := from.(type) {
	case float64:
		return ExprNum(int(it))
	case string:
		if n, e := strconv.ParseInt(it, 10, 0); e != nil {
			panic(e)
		} else { // rewrite arg-refs for later stack-access-from-tail-end: 0 -> -1, 1 -> -2, 2 -> -3
			return ExprArgRef(int(-(n + 1)))
		}
	case []any:
		if len(it) == 1 {
			return ExprFnRef(int(it[0].(float64)))
		}
		expr := exprFromJson(it[0])
		for i := 1; i < len(it); i++ {
			expr = ExprAppl{expr, exprFromJson(it[i])}
		}
		return expr
	case map[string]any: // allows for free-form annotations / comments / meta-data ...
		return exprFromJson(it[""]) // ... by digging into this field and ignoring all others
	}
	panic(from)
}
