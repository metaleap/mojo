package main

import (
	"os"
	"strconv"
	"strings"

	. "github.com/metaleap/atmo/atem"
)

const trace = false

var traceRootStep = &EvalStep{}

type EvalStep struct {
	Input    Expr
	Args     []Expr
	Result   Expr
	Again    bool
	NextArgs []Expr
	SubSteps []*EvalStep
}

func init() {
	if trace {
		OnEvalStep = onEvalStep
	}
}

func onEvalStep(input Expr, args []Expr) func(Expr, []Expr, bool) {
	root, idx := traceRootStep, len(traceRootStep.SubSteps)
	traceRootStep.SubSteps = append(traceRootStep.SubSteps, &EvalStep{Input: input, Args: args})
	traceRootStep = traceRootStep.SubSteps[idx]
	return func(result Expr, nextArgs []Expr, again bool) {
		traceRootStep.Result, traceRootStep.Again, traceRootStep.NextArgs = result, again, nextArgs
		identical := (traceRootStep.SubSteps == nil) && (traceRootStep.Result == traceRootStep.Input)
		if traceRootStep = root; identical {
			traceRootStep.SubSteps[idx] = nil
		}
	}
}

func writeTraceFile() {
	file, _ := os.Create("traced.txt")
	writeSteps(file, 0, traceRootStep.SubSteps)
	file.Sync()
	file.Close()
}

func writeSteps(to *os.File, level int, steps []*EvalStep) {
	for _, step := range steps {
		if step != nil {
			writeStep(to, level, step)
		}
	}
}

func writeStep(to *os.File, level int, step *EvalStep) {
	ind := strings.Repeat("  ", level)
	to.WriteString(ind + toStr(step.Input) + "\t\t\t\t\t\t\t\t\t\t\t\t\t\t")
	for i := len(step.Args) - 1; i >= 0; i-- {
		to.WriteString("\t\t" + toStr(step.Args[i]))
	}
	to.WriteString("\n")
	writeSteps(to, level+1, step.SubSteps)
	to.WriteString(ind + " = " + toStr(step.Result))
	if step.Again {
		to.WriteString("\t\t···")
		for i := len(step.NextArgs) - 1; i >= 0; i-- {
			to.WriteString("\t\t" + toStr(step.NextArgs[i]))
		}
	}
	to.WriteString("\n")
}

func toStr(expr Expr) (ret string) {
	if ret = "_"; expr != nil {
		ret = expr.JsonSrc()
		switch it := expr.(type) {
		case ExprFuncRef:
			if it < 0 {
				switch OpCode(it) {
				case OpAdd:
					ret = "ADD"
				case OpSub:
					ret = "SUB"
				case OpMul:
					ret = "MUL"
				case OpDiv:
					ret = "DIV"
				case OpMod:
					ret = "MOD"
				case OpEq:
					ret = "EQ"
				case OpGt:
					ret = "GT"
				case OpLt:
					ret = "LT"
				case OpPrt:
					ret = "PRT"
				default:
					ret = "ERR"
				}
			} else {
				ret = prog[it].Meta[0]
				ret = ret[strings.IndexByte(ret, ']')+1:]
				for _, s := range []string{"std.list.", "std.num.", "std.json.", "std."} {
					ret = strings.TrimPrefix(ret, s)
				}
				for _, s := range []string{"://", "//", ":"} {
					ret = strings.Replace(ret, s, "_", -1)
				}
			}
		case *ExprCall:
			if list := listOfExprs(it); list == nil {
				if ret = ""; it.IsClosure != 0 {
					ret += strconv.Itoa(it.IsClosure) + "#"
				}
				ret += "(" + toStr(it.Callee)
				for i := len(it.Args) - 1; i >= 0; i-- {
					ret += " " + toStr(it.Args[i])
				}
				ret += ")"
			} else if bytes := ListToBytes(list); bytes != nil {
				ret = strconv.Quote(string(bytes))
			} else {
				ret = "["
				for _, item := range list {
					ret += toStr(item) + ", "
				}
				ret += "]"
			}
		}
	}
	return
}

func listOfExprs(expr Expr) (ret []Expr) {
	ret = make([]Expr, 0, 1024)
	for ok, next := true, expr; ok; {
		ok = false
		if fnref, _ := next.(ExprFuncRef); fnref == StdFuncNil {
			break
		} else if call, okc := next.(*ExprCall); okc && len(call.Args) == 2 {
			if fnref, _ = call.Callee.(ExprFuncRef); fnref == StdFuncCons {
				CurEvalStepDepth++
				for i := len(call.Args) - 1; i > 0; i-- {
					ret = append(ret, call.Args[i])
				}
				ok, next = true, call.Args[0]
				CurEvalStepDepth--
			}
		}
		if !ok {
			ret = nil
		}
	}
	return
}
