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

	idx, meta := compileTopDef(mainTopDefQName), []string{mainTopDefQName}
	outProg = append(outProg, outProg[idx])
	outProg[idx] = FuncDef{Meta: meta, Body: ExprFuncRef(len(outProg) - 1)}

	for again := false; again; {
		outProg, again = optimize(outProg)
	}
}

func compileExpr(expr tl.Expr, curFunc *FuncDef, curFuncsArgs []*tl.ExprFunc) Expr {
	switch it := expr.(type) {
	case *tl.ExprLitNum:
		return ExprNumInt(it.NumVal)
	case *tl.ExprCall:
		return ExprAppl{Callee: compileExpr(it.Callee, curFunc, curFuncsArgs), Arg: compileExpr(it.CallArg, curFunc, curFuncsArgs)}
	case *tl.ExprName:
		if it.IdxOrInstr < 0 {
			argidx := len(curFuncsArgs) + it.IdxOrInstr
			return -ExprArgRef(1 + argidx)
		} else if it.IdxOrInstr > 0 {
			if opcode, ok := instr2op[tl.Instr(it.IdxOrInstr)]; ok {
				return ExprFuncRef(opcode)
			}
		} else {
			for i, farg := range curFuncsArgs {
				if farg.ArgName == it.NameVal {
					return -ExprArgRef(1 + i)
				}
			}
			idx := compileTopDef(it.NameVal)
			return ExprFuncRef(idx)
		}
	case *tl.ExprFunc:
		globalname, globalbody := "//"+curFunc.Meta[0]+"/lam/"+it.ArgName+strconv.Itoa(len(inProg.TopDefs)), it
		freevars, expr := map[string]int{}, tl.Expr(&tl.ExprName{NameVal: globalname})
		freeVars(globalbody, map[string]string{}, freevars)
		for fvname := range freevars {
			globalbody = &tl.ExprFunc{ArgName: fvname, Body: globalbody}
		}
		fargs, _ := dissectFunc(globalbody)
		for _, farg := range fargs[:len(freevars)] {
			expr = &tl.ExprCall{Callee: expr, CallArg: &tl.ExprName{NameVal: farg.ArgName}}
		}
		inProg.TopDefs[globalname] = globalbody
		return compileExpr(expr, curFunc, curFuncsArgs)
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
		outProg, defsDone[name] = append(outProg, FuncDef{Meta: []string{"[" + strconv.Itoa(idx) + "]" + name}}), idx

		result := &outProg[idx]
		curfuncsargs, body := dissectFunc(topdef)
		result.Args = make([]int, len(curfuncsargs))
		for i, f := range curfuncsargs {
			result.Meta = append(result.Meta, f.ArgName)
			result.Args[i] = body.ReplaceName(f.ArgName, f.ArgName) // just counts occurrences
			for _, local := range locals {
				if numrefs := local.Expr.ReplaceName(f.ArgName, f.ArgName); numrefs > 0 {
					result.Args[i] += numrefs
				}
			}
		}

		localnames := map[string]string{}
		for i, local := range locals {
			globalname := "//" + name + "/local/" + local.Name + strconv.Itoa(i)
			localnames[local.Name] = globalname
		}
		for i, local := range locals {
			globalname, globalbody := localnames[local.Name], local.Expr
			freevars, expr := map[string]int{}, tl.Expr(&tl.ExprName{NameVal: globalname})
			freeVars(local.Expr, localnames, freevars)
			for fvname := range freevars {
				globalbody = &tl.ExprFunc{ArgName: fvname, Body: globalbody}
			}
			fargs, _ := dissectFunc(globalbody)
			for _, farg := range fargs[:len(freevars)] {
				expr = &tl.ExprCall{Callee: expr, CallArg: &tl.ExprName{NameVal: farg.ArgName}}
			}

			inProg.TopDefs[globalname] = globalbody
			for j, exprcopy := 0, fullCopy(expr); j <= i; j++ {
				if gname := localnames[locals[j].Name]; 0 < locals[j].Expr.ReplaceName(local.Name, local.Name) {
					exprcopy, inProg.TopDefs[gname] = fullCopy(expr), inProg.TopDefs[gname].RewriteName(local.Name, exprcopy)
				}
			}
			body = body.RewriteName(local.Name, expr)
		}

		result.Body = compileExpr(body, result, curfuncsargs)
		outProg[idx] = *result // crucial as the slice could have been resized by now (even tho we give it an initial cap that makes this unlikely, this safeguard will work for any outrageous program size)
	}
	return idx
}
