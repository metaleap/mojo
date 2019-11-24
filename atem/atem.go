package atem

import (
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
)

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
	return outjson + "],\t\t" + me.Body.String() + " ]"
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
