// A simple executable form of the [atem reference interpreter](../../readme.md).
// The first (and required) command arg is the `.json` source file for the
// `atem.Prog` to `atem.LoadFromJson()`. All further command args are passed
// on to the loaded program's `main`.
//
// Since there are no identifiers in `atem` programs, by (hereby decreed)
// convention the very last `FuncDef` in the `Prog` is assumed to be the `main`
// to run (atem code emitters must ensure this if their outputs are to be run
// in here), and is expected to have a `FuncDef.Args` of `len` 2. The first
// one will be populated by this interpreter executable with the current process
// args (sans the current process executable name and the input `.json` source
// file path) as a linked-list of text strings, the second will be the current
// process environment variables as a linked-list of `NAME=Value` text strings.
package main

import (
	"bufio"
	"bytes"
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

	outexpr := prog.Eval(ExprAppl{
		Callee: ExprAppl{
			Callee: ExprFuncRef(len(prog) - 1), // `main` is always last by convention
			Arg:    ListsFrom(os.Args[2:]),     // first `main` param: list of all process args following `atem inputfile`
		},
		Arg: ListsFrom(os.Environ()), // second `main` param: list of all env-vars (list of "FOO=Bar" strings)
	}, make([]Expr, 0, 128))
	outlist := prog.ListOfExprs(outexpr) // forces lazy thunks

	if outbytes := ListToBytes(outlist); outbytes != nil { // by convention we expect a byte-array return from `main`
		os.Stdout.Write(append(outbytes, '\n'))
	} else if outlist == nil || !probeIfStdinReaderAndIfSoHandleOnceOrForever(prog, outlist) {
		println("RET-EXPR:\t" + outexpr.JsonSrc() + "\n")
	}
}

func probeIfStdinReaderAndIfSoHandleOnceOrForever(prog Prog, retList []Expr) bool {
	if len(retList) == 4 {
		if fnhandler, okf := retList[0].(ExprFuncRef); okf {
			if sepchar, oks := retList[1].(ExprNumInt); oks {
				if initialoutputlist, oka := retList[3].(ExprAppl); oka {
					if initialoutput := ListToBytes(prog.ListOfExprs(initialoutputlist)); initialoutput != nil {
						initialstate, handlenextinput := retList[2], func(prevstate Expr, input []byte) (nextstate Expr) {
							retexpr := prog.Eval(ExprAppl{Callee: ExprAppl{Callee: fnhandler, Arg: prevstate}, Arg: ListFrom(input)}, make([]Expr, 0, 128))
							if retlist := prog.ListOfExprs(retexpr); len(retlist) == 2 {
								if outlist := prog.ListOfExprs(retlist[1]); outlist != nil {
									nextstate = retlist[0]
									os.Stdout.Write(ListToBytes(outlist))
								}
							}
							if nextstate == nil {
								panic(retexpr.JsonSrc())
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

func stdinReadSplitterBy(sep byte) bufio.SplitFunc {
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if i := bytes.IndexByte(data, sep); i >= 0 {
			advance, token = i+1, data[0:i]
		} else if atEOF {
			advance, token = len(data), data
		}
		return
	}
}
