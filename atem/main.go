package main

import (
	"bufio"
	"io/ioutil"
	"os"
)

func main() {
	src, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	prog := LoadFromJson(src)

	outexpr := prog.eval(ExprAppl{
		Callee: ExprAppl{
			Callee: ExprFnRef(len(prog) - 1), // `main` is always last by convention
			Arg:    ListsFrom(os.Args[2:]),   // first `main` param is process args list
		},
		Arg: ListsFrom(os.Environ()), // second `main` param is env-vars: array of "FOO=Bar" strings
	}, make([]Expr, 0, 128))
	outlist := prog.List(outexpr) // forces lazy thunks

	if outbytes := ToBytes(outlist); outbytes != nil { // by convention we expect a byte-array return from `main`
		os.Stdout.Write(append(outbytes, 10))
	} else if outlist == nil || !prog.probeIfStdinReaderAndIfSoHandleOnceOrForever(outlist) {
		println("EXPR:\t" + outexpr.String() + "\n")
	}
}

func (me Prog) probeIfStdinReaderAndIfSoHandleOnceOrForever(retList []Expr) bool {
	if len(retList) == 3 {
		if fnhandler, okf := retList[0].(ExprFnRef); okf {
			if sepchar, oks := retList[1].(ExprNum); oks {
				if appl, oka := retList[2].(ExprAppl); oka {
					if strintro := ToBytes(me.List(appl)); strintro != nil {

						handlenextinput := func(prevstate Expr, input []byte) (nextstate Expr) {
							retexpr := me.eval(ExprAppl{Callee: ExprAppl{Callee: fnhandler, Arg: prevstate}, Arg: ListFrom(input)}, make([]Expr, 0, 128))
							if retlist := me.List(retexpr); len(retlist) == 2 {
								nextstate = retlist[0]
								os.Stdout.Write(ToBytes(me.List(retlist[1])))
							} else {
								panic(retexpr.String())
							}
							return
						}

						if os.Stdout.Write(strintro); sepchar == 0 {
							if allinputatonce, err := ioutil.ReadAll(os.Stdin); err != nil {
								panic(err)
							} else {
								_ = handlenextinput(StdFuncId, allinputatonce)
							}
						} else {
							reader := bufio.NewScanner(os.Stdin)
							if sepchar != '\n' {
								reader.Split(stdinReadSplitterBy(byte(sepchar)))
							}
							for state := Expr(StdFuncNil); reader.Scan(); {
								state = handlenextinput(state, reader.Bytes())
								if fn, _ := state.(ExprFnRef); fn == StdFuncId {
									break
								}
							}
							if err := reader.Err(); err != nil {
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
