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

func compileExpr(expr tl.Expr) Expr {
	switch it := expr.(type) {
	case *tl.ExprLitNum:
		return ExprNumInt(it.NumVal)
	case *tl.ExprCall:
		if optIsCallIdentity(it) {
			return compileExpr(it.CallArg)
		}
		return ExprCall{Callee: compileExpr(it.Callee), Arg: compileExpr(it.CallArg)}
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

		localnames, newlocals := map[string]int{}, map[string]*tl.ExprFunc{}
		for i, local := range locals {
			localnames[local.Name] = i
		}
		for i, local := range locals {
			locals[i].Expr = extractFuncs(local.Expr, localnames, newlocals)
		}
		body = extractFuncs(body, localnames, newlocals)
		for name, expr := range newlocals {
			locals = append(locals, tl.LocalDef{Name: name, Expr: expr})
		}
		for _, local := range locals {
			println(name + local.Name)
			inProg.TopDefs[name+local.Name] = local.Expr
		}

		result.Body = compileExpr(body)
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

func extractFuncs(expr tl.Expr, localNames map[string]int, gatherInto map[string]*tl.ExprFunc) tl.Expr {
	switch it := expr.(type) {
	case *tl.ExprCall:
		it.Callee, it.CallArg = extractFuncs(it.Callee, localNames, gatherInto), extractFuncs(it.CallArg, localNames, gatherInto)
	case *tl.ExprFunc:
		newlocalname := "//" + strconv.Itoa(len(gatherInto)) + "//" + it.ArgName
		it.Body = extractFuncs(it.Body, localNames, gatherInto)
		freevars := map[string]int{}
		expr = &tl.ExprName{Loc: it.Loc, NameVal: newlocalname}
		freeVars(it, localNames, freevars)
		for fvname, dbidx := range freevars {
			expr = &tl.ExprCall{Loc: it.Loc, Callee: expr, CallArg: &tl.ExprName{Loc: it.Loc, NameVal: fvname, IdxOrInstr: dbidx + 1}}
			it = &tl.ExprFunc{Loc: it.Loc, ArgName: fvname, Body: it}
		}
		gatherInto[newlocalname] = it
	}
	return expr
}

func freeVars(expr tl.Expr, localNames map[string]int, results map[string]int) {
	switch it := expr.(type) {
	case *tl.ExprCall:
		freeVars(it.Callee, localNames, results)
		freeVars(it.CallArg, localNames, results)
	case *tl.ExprFunc:
		if _, exists := localNames[it.ArgName]; exists {
			panic(it.ArgName)
		}
		localNames[it.ArgName] = -1
		freeVars(it.Body, localNames, results)
		delete(localNames, it.ArgName)
	case *tl.ExprName:
		if it.IdxOrInstr <= 0 {
			if _, exists := localNames[it.NameVal]; !exists {
				if _, exists = inProg.TopDefs[it.NameVal]; !exists {
					if _, exists = inProg.TopDefs[tl.StdModuleName+"."+it.NameVal]; !exists {
						if _, exists = results[it.NameVal]; exists {
							panic(it.NameVal)
						}
						results[it.NameVal] = it.IdxOrInstr
					}
				}
			}
		}
	}
}
