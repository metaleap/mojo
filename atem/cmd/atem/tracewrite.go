package main

import (
	"os"
	"strconv"
	"strings"

	. "github.com/metaleap/atmo/atem"
)

const traceToFile = true

var traceOutFile *os.File

func init() {
	if traceToFile {
		traceOutFile, _ = os.Create("traced.txt")
		pwd, _ := os.Getwd()
		traceOutFile.WriteString(pwd + "\n" + strings.Join(os.Args, " ") + "\n\n")

		// var fnstack []ExprFuncRef
		cache := map[string]string{}
		var tostring func(Prog, Expr) string
		tostring = func(prog Prog, expr Expr) (ret string) {
			jstr := "_"
			if expr != nil {
				jstr = expr.JsonSrc()
			}
			if ret = cache[jstr]; ret == "" {
				cache[jstr] = "‹∞›"
				if expr == nil {
					ret = "_"
				} else {
					switch it := expr.(type) {
					case ExprArgRef:
						ret = jstr // prog[fnstack[len(fnstack)-1]].Meta[(-it)-1]
					case ExprFuncRef:
						if it > -1 {
							ret = prog[it].Meta[0][strings.LastIndexByte(prog[it].Meta[0], ']')+1:]
							ret = strings.TrimPrefix(ret, "std.num.")
							ret = strings.TrimPrefix(ret, "std.list.")
							ret = strings.TrimPrefix(ret, "std.json.")
							ret = strings.TrimPrefix(ret, "std.")
						} else {
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
							case OpLt:
								ret = "LT"
							case OpGt:
								ret = "GT"
							case OpPrt:
								ret = "PRT"
							default:
								ret = "ERR"
							}
						}
					case *ExprCall:
						if list := prog.ListOfExprs(it, false); list == nil {
							s := "(" + tostring(prog, it.Callee)
							for i := len(it.Args) - 1; i > -1; i-- {
								s += " " + tostring(prog, it.Args[i])
							}
							ret = s + ")"
						} else if bytes := ListToBytes(list); bytes == nil {
							s := "["
							for _, item := range list {
								s += tostring(prog, item) + ", "
							}
							ret = s + "]"
						} else {
							ret = strconv.Quote(string(bytes))
						}
					}
					if ret == "" {
						ret = jstr
					}
				}
				cache[jstr] = ret
			}
			return ret
		}

		OnEvalStep = func(prog Prog, expr Expr, stack []Expr) {
			traceOutFile.WriteString(strings.Repeat("\t", CurEvalDepth))
			traceOutFile.WriteString("|" + strconv.Itoa(len(stack)) + "|\t\t")
			if fnref, okf := expr.(ExprFuncRef); okf {
				numargs := 2
				if fnref > -1 {
					numargs = len(prog[fnref].Args)
				}
				if len(stack) >= numargs {
					traceOutFile.WriteString("DE-STACK " + strconv.Itoa(numargs) + ":\t")
				} else {
					traceOutFile.WriteString("CLOSURE x" + strconv.Itoa(numargs) + ":\t")
				}
			} else if call, _ := expr.(*ExprCall); call != nil {
				traceOutFile.WriteString("EN-STACK " + strconv.Itoa(len(call.Args)) + ":\t")
			}
			traceOutFile.WriteString(tostring(prog, expr))
			traceOutFile.WriteString("\n")
		}
	}
}
