package main

import (
	. "github.com/metaleap/atmo/atem"
	tl "github.com/metaleap/go-machines/toylam"
)

var defsDone = map[string]int{}

func compile(mainTopDefQName string) {
	compileTopDef(tl.StdRequiredDefs_id)
	compileTopDef(tl.StdRequiredDefs_true)
	compileTopDef(tl.StdRequiredDefs_false)
	compileTopDef(tl.StdRequiredDefs_listNil)
	compileTopDef(tl.StdRequiredDefs_listCons)

	idx := compileTopDef(mainTopDefQName)
	outProg = append(outProg, outProg[idx])
	outProg[idx] = FuncDef{Args: nil, Body: ExprFuncRef(len(outProg) - 1)}
}

func compileTopDef(name string) int {
	idx, done := defsDone[name]
	if !done {
		prefstd, topdef, locals := false, inProg.TopDefs[name], inProg.TopDefSepLocals[name]
		if topdef == nil { // degraded case usually occurs (or should) only for recursive self-refs
			prefstd, topdef, locals = true, inProg.TopDefs[tl.StdModuleName+"."+name], inProg.TopDefSepLocals[tl.StdModuleName+"."+name]
			if idx, done = defsDone[tl.StdModuleName+"."+name]; done {
				return idx
			}
		}
		if topdef == nil {
			panic(name)
		}

		idx = len(outProg)
		if outProg, defsDone[name] = append(outProg, FuncDef{}), idx; prefstd {
			defsDone[tl.StdModuleName+"."+name] = idx
		}

		result := &outProg[idx]
		if len(locals) > 0 {
			panic(topdef.LocInfo().LocStr() + locals[0].Name)
		}
		funcs, body := untangle(topdef)
		result.Args = make([]int, len(funcs))
		for i, f := range funcs {
			result.Args[i] = body.ReplaceName(f.ArgName, f.ArgName)
		}
		result.Body = compileExpr(body)
		outProg[idx] = *result // crucial as the slice could have been resized by now (even tho we give it an initial cap that makes this unlikely, this safeguard will work for any outrageous program size)
	}
	return idx
}

func compileExpr(expr tl.Expr) Expr {
	switch it := expr.(type) {
	case *tl.ExprLitNum:
		return ExprNumInt(it.NumVal)
	case *tl.ExprName:
		if it.IdxOrInstr < 0 {
			return ExprArgRef(it.IdxOrInstr)
		} else if it.IdxOrInstr > 0 {
			if opcode, ok := instr2op[tl.Instr(it.IdxOrInstr)]; ok {
				return ExprFuncRef(opcode)
			}
		} else {
			idx := compileTopDef(it.NameVal)
			return ExprFuncRef(idx)
		}
	case *tl.ExprCall:
		return ExprCall{Callee: compileExpr(it.Callee), Arg: compileExpr(it.CallArg)}
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
