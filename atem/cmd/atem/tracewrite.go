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
		OnEvalStep = func(prog Prog, expr Expr, stack []Expr) {
			traceOutFile.WriteString(strings.Repeat("\t", CurEvalDepth))
			switch it := expr.(type) {
			case ExprNumInt:
				traceOutFile.WriteString("INUM:\t" + strconv.Itoa(int(it)))
			case ExprArgRef:
				traceOutFile.WriteString("AREF:\t" + it.JsonSrc())
			case ExprFuncRef:
				numargs, name, bodysrc := 2, strconv.Itoa(int(it)), ""
				if it > -1 {
					numargs, name, bodysrc = len(prog[it].Args), prog[it].Meta[0], prog[it].Body.JsonSrc()
				} else {
					lhs, rhs := stack[len(stack)-1], stack[len(stack)-2]
					switch OpCode(it) {
					case OpAdd:
						name, bodysrc = "ADD", "{"+lhs.JsonSrc()+" + "+rhs.JsonSrc()+"}"
					case OpSub:
						name, bodysrc = "SUB", "{"+lhs.JsonSrc()+" - "+rhs.JsonSrc()+"}"
					case OpMul:
						name, bodysrc = "MUL", "{"+lhs.JsonSrc()+" * "+rhs.JsonSrc()+"}"
					case OpDiv:
						name, bodysrc = "DIV", "{"+lhs.JsonSrc()+" / "+rhs.JsonSrc()+"}"
					case OpMod:
						name, bodysrc = "MOD", "{"+lhs.JsonSrc()+" % "+rhs.JsonSrc()+"}"
					case OpEq:
						name, bodysrc = "EQ", "{"+lhs.JsonSrc()+" = "+rhs.JsonSrc()+"}"
					case OpGt:
						name, bodysrc = "GT", "{"+lhs.JsonSrc()+" > "+rhs.JsonSrc()+"}"
					case OpLt:
						name, bodysrc = "LT", "{"+lhs.JsonSrc()+" < "+rhs.JsonSrc()+"}"
					case OpPrt:
						name, bodysrc = "PRT", rhs.JsonSrc()
					}
				}
				traceOutFile.WriteString("FREF:\t" + name + "\tcurStackLen=" + strconv.Itoa(len(stack)) + "\tnumArgs=" + strconv.Itoa(numargs) + "\t\t" + bodysrc)
			case *ExprCall:
				traceOutFile.WriteString("CALL:" + "\tcurStackLen=" + strconv.Itoa(len(stack)) + "\tnumArgs=" + strconv.Itoa(len(it.Args)) + "\t\t" + it.JsonSrc())
			default:
				panic(it)
			}
			traceOutFile.WriteString("\n")
		}
	}
}
