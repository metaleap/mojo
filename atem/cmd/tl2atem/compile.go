package main

import (
	"fmt"
	"strconv"
	"strings"

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

	for again := true; again; optNumRounds = 0 {
		outProg, again = optimize(outProg)
		println("OPT:", optNumRounds, "round(s)")
	}
}

func compileExpr(expr tl.Expr, funcsArgs []*tl.ExprFunc) Expr {
	switch it := expr.(type) {
	case *tl.ExprLitNum:
		return ExprNumInt(it.NumVal)
	case *tl.ExprCall:
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
	case *tl.ExprFunc:
		globalname, globalbody := "//"+strconv.Itoa(len(inProg.TopDefs))+"//"+it.ArgName, it
		freevars, expr := map[string]int{}, tl.Expr(&tl.ExprName{NameVal: globalname})
		freeVars(globalbody, map[string]string{}, freevars)
		for fvname := range freevars {
			globalbody = &tl.ExprFunc{ArgName: fvname, Body: globalbody}
		}
		fargs, _ := flattenFunc(globalbody)
		for _, farg := range fargs[:len(freevars)] {
			expr = &tl.ExprCall{Callee: expr, CallArg: &tl.ExprName{NameVal: farg.ArgName}}
		}
		inProg.TopDefs[globalname] = globalbody
		return compileExpr(expr, funcsArgs)
	}
	panic(fmt.Sprintf("%T\t%v", expr, expr))
}

func compileTopDef(name string) int {
	idx, done := defsDone[name]
	if !done {
		prefstd, topdef, locals := "", inProg.TopDefs[name], inProg.TopDefSepLocals[name]
		if topdef == nil {
			if prefstd, topdef = tl.StdModuleName+".", inProg.TopDefs[tl.StdModuleName+"."+name]; topdef == nil {
				for k, v := range inProg.TopDefs {
					if strings.HasSuffix(k, "."+name) && strings.HasPrefix(k, tl.StdModuleName+".") {
						prefstd, topdef = k[:len(k)-len(name)], v
						break
					}
				}
			}
			if idx, done = defsDone[prefstd+name]; done {
				return idx
			}
			locals = inProg.TopDefSepLocals[prefstd+name]
		}
		if topdef == nil {
			panic(name)
		}
		idx, name = len(outProg), prefstd+name
		outProg, defsDone[name] = append(outProg, FuncDef{}), idx

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
			globalname, globalbody := localnames[local.Name], local.Expr
			freevars, expr := map[string]int{}, tl.Expr(&tl.ExprName{NameVal: globalname})
			freeVars(local.Expr, localnames, freevars)
			for fvname := range freevars {
				globalbody = &tl.ExprFunc{ArgName: fvname, Body: globalbody}
			}
			fargs, _ := flattenFunc(globalbody)
			for _, farg := range fargs[:len(freevars)] {
				expr = &tl.ExprCall{Callee: expr, CallArg: &tl.ExprName{NameVal: farg.ArgName}}
			}

			inProg.TopDefs[globalname] = globalbody
			for j, exprcopy := 0, fullCopy(expr); j <= i; j++ {
				if gname := localnames[locals[j].Name]; 0 < locals[j].Expr.ReplaceName(local.Name, local.Name) {
					exprcopy, inProg.TopDefs[gname] = fullCopy(expr), inProg.TopDefs[gname].RewriteName(local.Name, exprcopy)
				}
			}
			if 0 < body.ReplaceName(local.Name, local.Name) {
				body = body.RewriteName(local.Name, expr)
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

func fullCopy(expr tl.Expr) tl.Expr {
	switch it := expr.(type) {
	case *tl.ExprLitNum:
		return &*it
	case *tl.ExprName:
		return &*it
	case *tl.ExprFunc:
		ret := *it
		ret.Body = fullCopy(ret.Body)
		return &ret
	case *tl.ExprCall:
		ret := *it
		ret.Callee, ret.CallArg = fullCopy(ret.Callee), fullCopy(ret.CallArg)
		return &ret
	}
	panic(expr)
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
