package main

import (
	at "github.com/metaleap/atmo/atem"
	tl "github.com/metaleap/go-machines/toylam"
)

var defsDone = map[string]int{}

func compile() {
	compileTopDef(tl.StdRequiredDefs_id)
	compileTopDef(tl.StdRequiredDefs_true)
	compileTopDef(tl.StdRequiredDefs_false)
	compileTopDef(tl.StdRequiredDefs_listNil)
	compileTopDef(tl.StdRequiredDefs_listCons)
	println(inProg.TopDefs[mainTopDefQName].String())
}

func compileTopDef(name string) int {
	idx, done := defsDone[name]
	if !done {
		var result at.FuncDef
		topdef, locals := inProg.TopDefs[name], inProg.TopDefSepLocals[name]
		if len(locals) > 0 {
			panic(topdef.LocInfo().LocStr())
		}
		funcs, body := untangle(topdef)
		result.Args = make([]int, len(funcs))
		for i, f := range funcs {
			result.Args[i] = body.ReplaceName(f.ArgName, f.ArgName)
		}
		result.Body = compileExpr(body)

		idx = len(outProg)
		outProg, defsDone[name] = append(outProg, result), idx
	}
	return idx
}

func compileExpr(expr tl.Expr) at.Expr {
	switch it := expr.(type) {
	case *tl.ExprLitNum:
		return at.ExprNumInt(it.NumVal)
	case *tl.ExprName:
		if it.IdxOrInstr < 0 {
			return at.ExprArgRef(it.IdxOrInstr)
		} else if it.IdxOrInstr > 0 {
			// prim instr op code
		} else {
			// global ref
		}
	case *tl.ExprCall:
		return at.ExprCall{Callee: compileExpr(it.Callee), Arg: compileExpr(it.CallArg)}
	}
	panic(expr)
}

func untangle(expr tl.Expr) (outerFuncs []*tl.ExprFunc, innerMostBody tl.Expr) {
	innerMostBody = expr
	for fn, _ := expr.(*tl.ExprFunc); fn != nil; fn, _ = fn.Body.(*tl.ExprFunc) {
		innerMostBody, outerFuncs = fn.Body, append(outerFuncs, fn)
	}
	return
}
