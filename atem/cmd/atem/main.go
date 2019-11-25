package main

import (
	"bufio"
	"io/ioutil"
	"os"

	. "github.com/metaleap/atmo/atem"
)

func main() {
	src, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	prog := LoadFromJson(src)

	outexpr := prog.Eval(ExprCall{
		Callee: ExprCall{
			Callee: ExprFuncRef(len(prog) - 1), // `main` is always last by convention
			Arg:    ListsFrom(os.Args[2:]),     // first `main` param: list of all process args following `atem inputfile`
		},
		Arg: ListsFrom(os.Environ()), // second `main` param: list of all env-vars (list of "FOO=Bar" strings)
	}, make([]Expr, 0, 128))
	outlist := prog.List(outexpr) // forces lazy thunks

	if outbytes := Bytes(outlist); outbytes != nil { // by convention we expect a byte-array return from `main`
		os.Stdout.Write(append(outbytes, 10))
	} else if outlist == nil || !probeIfStdinReaderAndIfSoHandleOnceOrForever(prog, outlist) {
		println("?!EXPR:\t" + outexpr.String() + "\n")
	}
}

func probeIfStdinReaderAndIfSoHandleOnceOrForever(prog Prog, retList []Expr) bool {
	if len(retList) == 4 {
		if fnhandler, okf := retList[0].(ExprFuncRef); okf {
			if sepchar, oks := retList[1].(ExprNumInt); oks {
				if initialoutputlist, oka := retList[3].(ExprCall); oka {
					if initialoutput := Bytes(prog.List(initialoutputlist)); initialoutput != nil {
						initialstate, handlenextinput := retList[2], func(prevstate Expr, input []byte) (nextstate Expr) {
							retexpr := prog.Eval(ExprCall{Callee: ExprCall{Callee: fnhandler, Arg: prevstate}, Arg: ListFrom(input)}, make([]Expr, 0, 128))
							if retlist := prog.List(retexpr); len(retlist) == 2 {
								if outlist := prog.List(retlist[1]); outlist != nil {
									nextstate = retlist[0]
									os.Stdout.Write(Bytes(outlist))
								}
							}
							if nextstate == nil {
								panic(retexpr.String())
							}
							return
						}

						if os.Stdout.Write(initialoutput); sepchar == 0 {
							if allinputatonce, err := ioutil.ReadAll(os.Stdin); err != nil {
								panic(err)
							} else {
								_ = handlenextinput(initialstate, allinputatonce)
							}
						} else {
							stdin := bufio.NewScanner(os.Stdin)
							if sepchar != '\n' {
								stdin.Split(stdinReadSplitterBy(byte(sepchar)))
							}
							for state := initialstate; stdin.Scan(); {
								state = handlenextinput(state, stdin.Bytes())
								if fn, ok := state.(ExprFuncRef); ok && fn == StdFuncId {
									break // `ok` above required because `StdFuncId` is 0
								}
							}
							if err := stdin.Err(); err != nil {
								panic(err)
							}
						}
						return true
					}
				}
			}
		}
	}
	return false
}