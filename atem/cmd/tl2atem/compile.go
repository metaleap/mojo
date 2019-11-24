package main

import (
	"fmt"
	"strconv"

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

func compileExpr(expr tl.Expr, funcsArgs []*tl.ExprFunc) Expr {
	switch it := expr.(type) {
	case *tl.ExprLitNum:
		return ExprNumInt(it.NumVal)
	case *tl.ExprCall:
		if optIsCallIdentity(it) {
			return compileExpr(it.CallArg, funcsArgs)
		}
		return ExprCall{Callee: compileExpr(it.Callee, funcsArgs), Arg: compileExpr(it.CallArg, funcsArgs)}
	case *tl.ExprName:
		if it.IdxOrInstr < 0 {
			return ExprArgRef(it.IdxOrInstr)
		} else if it.IdxOrInstr > 0 {
			if opcode, ok := instr2op[tl.Instr(it.IdxOrInstr)]; ok {
				return ExprFuncRef(opcode)
			}
		} else {
			for i, farg := range funcsArgs {
				if farg.ArgName == it.NameVal {
					return ExprArgRef(i)
				}
			}
			idx := compileTopDef(it.NameVal)
			return ExprFuncRef(idx)
		}
	}
	panic(fmt.Sprintf("%T\t%v", expr, expr))
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
		} else if optIsNameIdentity(topdef) {
			defsDone[name], defsDone[tl.StdModuleName+"."+name] = 0, 0
			return 0
		}
		idx = len(outProg)
		if outProg, defsDone[name] = append(outProg, FuncDef{}), idx; prefstd {
			name = tl.StdModuleName + "." + name
			defsDone[name] = idx
		}

		result := &outProg[idx]
		funcsargs, body := flattenFunc(topdef)
		result.Args = make([]int, len(funcsargs))
		for i, f := range funcsargs {
			result.Args[i] = body.ReplaceName(f.ArgName, f.ArgName) // just counts occurrences
		}

		localnames := map[string]string{}
		for i, local := range locals {
			globalname := name + "//" + local.Name + "//" + strconv.Itoa(i)
			localnames[local.Name] = globalname
		}
		for i, local := range locals {
			globalname := localnames[local.Name]
			inProg.TopDefs[globalname] = local.Expr
			freevars := map[string]int{}
			if freeVars(local.Expr, localnames, freevars); len(freevars) > 0 {
				println(name + "\t\t" + local.Name)
				for k := range freevars {
					println("\t\t" + k)
				}
				panic(len(freevars))
			}
			for j := 0; j <= i; j++ {
				if gname := localnames[locals[j].Name]; 0 < locals[j].Expr.ReplaceName(local.Name, local.Name) {
					inProg.TopDefs[gname] = inProg.TopDefs[gname].RewriteName(local.Name, &tl.ExprName{NameVal: globalname})
				}
			}
			if 0 < body.ReplaceName(local.Name, local.Name) {
				body = body.RewriteName(local.Name, &tl.ExprName{NameVal: globalname})
			}
		}

		result.Body = compileExpr(body, funcsargs)
		outProg[idx] = *result // crucial as the slice could have been resized by now (even tho we give it an initial cap that makes this unlikely, this safeguard will work for any outrageous program size)
	}
	return idx
}

func flattenFunc(expr tl.Expr) (outerFuncs []*tl.ExprFunc, innerMostBody tl.Expr) {
	innerMostBody = expr
	for fn, _ := expr.(*tl.ExprFunc); fn != nil; fn, _ = fn.Body.(*tl.ExprFunc) {
		innerMostBody, outerFuncs = fn.Body, append(outerFuncs, fn)
	}
	return
}

func freeVars(expr tl.Expr, localNames map[string]string, results map[string]int) {
	switch it := expr.(type) {
	case *tl.ExprCall:
		freeVars(it.Callee, localNames, results)
		freeVars(it.CallArg, localNames, results)
	case *tl.ExprFunc:
		if _, exists := localNames[it.ArgName]; exists {
			panic(it.ArgName)
		}
		localNames[it.ArgName] = ""
		freeVars(it.Body, localNames, results)
		delete(localNames, it.ArgName)
	case *tl.ExprName:
		if it.IdxOrInstr <= 0 {
			if _, exists := localNames[it.NameVal]; !exists {
				if _, exists = inProg.TopDefs[it.NameVal]; !exists {
					if _, exists = inProg.TopDefs[tl.StdModuleName+"."+it.NameVal]; !exists {
						if have, exists := results[it.NameVal]; exists && have != it.IdxOrInstr {
							panic(it.NameVal)
						}
						results[it.NameVal] = it.IdxOrInstr
					}
				}
			}
		}
	}
}
