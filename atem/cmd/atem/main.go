// A simple executable form of the [atem reference interpreter](../../readme.md)
// lib. The first (and required) command arg is the `.json` source file for the
// `atem.Prog` to first `atem.LoadFromJson()` and then run. All further process
// args are passed on to the loaded source program's main `FuncDef`.
//
// Since there are no identifiers in `atem` programs, by (hereby decreed)
// convention the very last `FuncDef` in the `Prog` is expected to be the one
// to run (atem code emitters must ensure this if their outputs are to be run
// in here), and is expected to have a `FuncDef.Args` of `len` 2. The first
// one will be populated by this interpreter executable with the current process
// args (sans the current process executable name and the input `.json` source
// file path) as a linked-list of text strings, the second will be the current
// process environment variables as a linked-list of `NAME=Value` text strings.
//
// ## stdout, stderr, stdin
//
// The main `FuncDef` is by default expected to return a linked list of
// `atem.ExprNumInt`s in the range of 0 .. 255, if it does that is considered the
// text output to be written to `stdout` and so will it be done. Other returned
// `Expr`s will have their `.JsonSrc()` written to `stderr` instead. For source
// programs to force extra writes to `stderr` during their run, the `atem.OpPrt`
// op-code is to be used. For access to `stdin`, the main `FuncDef` must return
// a specific predefined linked-list meeting the following characteristics:
//
// - it has 4 elements, in order:
//   1. a valid `ExprFuncRef` (the "handler"),
//   2. an `ExprNumInt` in the range of 0 .. 255 (the "separator char"),
//   3. any `Expr` (the "initial state"),
//   4. and a text string linked list (the "initial output")
//
// - the "handler"-referenced `FuncDef` must take 2 args, the "previous state" (for the first call this will be the "initial state" mentioned above) and the "input" (a text string linked list of any length). It must always return a linked-list of length 2, with the first element being "next state" (will be passed as is into "handler" in the next upcoming call) and "output", a text string linked list of any length incl. zero to be immediately written to `stdout`. If "next state" is returned as an `ExprFuncRef` pointing to `StdFuncId` (aka. `0`), this indicates to cease further `stdin` reading and "handler" calling, essentially terminating the program.
//
// - the "separator char" is 0 to indicate to read in all `stdin` data at once until EOF before passing it all at once to "handler" in a single and final call. If it isn't 0, "handler" is called with fresh input whenever in incoming `stdin` data the "separator char" is next encountered, it's never included in the "handler" input. So to achieve a typical read-line functionality, one would use a "separator char" of `10` aka. `'\n'`.
//
// - the "initial state" is what gets passed in the first call to "handler". Subsequent "handler" calls will instead receive the previous call's returned "next state" as described above.
//
// - the "initial output", a text string linked list of any length incl. zero, will be written to `stdout` before the first read from `stdin` and the first call to "handler".
//
package main

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"os"
	"runtime/debug"
	"strconv"

	. "github.com/metaleap/atmo/atem"
)

func main() {
	debug.SetGCPercent(-1)
	src, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	prog := LoadFromJson(src)
	if numargs := len(prog[len(prog)-1].Args); 2 != numargs {
		panic("Your main FuncDef needs exactly 2 Args but has " + strconv.Itoa(numargs) + ": " + prog[len(prog)-1].JsonSrc(false))
	}
	defer func() {
		println("ND", NumDrops)
		if traceToFile {
			traceOutFile.Sync()
			traceOutFile.Close()
		}
		if thrown := recover(); thrown != nil {
			if err, ok := thrown.([3]Expr); !ok {
				panic(thrown)
			} else {
				os.Stderr.WriteString(prog.ListOfExprsToString(err[1]) + "\t" + prog.ListOfExprsToString(err[2]) + "\n")
			}
		}
	}()
	outexpr := prog.Eval(&ExprCall{ // we start!
		Callee: ExprFuncRef(len(prog) - 1), // `main` is always last by convention
		Args: []Expr{ListsFrom(os.Environ()), // second `main` param: `env`, a list of all env-vars (list of "FOO=Bar" strings)
			ListsFrom(os.Args[2:]), // first `main` param: `args`, a list of all process args following `atem inputfile`
		}})
	outlist := prog.ListOfExprs(outexpr) // forces lazy thunks

	if outbytes := ListToBytes(outlist); outbytes != nil { // by convention we expect a byte-array return from `main`
		os.Stdout.Write(append(outbytes, '\n'))
	} else if outlist == nil || !probeIfStdinReaderAndIfSoHandleOnceOrForever(prog, outlist) {
		os.Stderr.WriteString("RET-EXPR:\t" + outexpr.JsonSrc() + "\n")
	}
}

func probeIfStdinReaderAndIfSoHandleOnceOrForever(prog Prog, retList []Expr) bool {
	if len(retList) == 4 {
		if fnhandler, okf := retList[0].(ExprFuncRef); okf && fnhandler > StdFuncCons && int(fnhandler) < len(prog)-1 && len(prog[fnhandler].Args) == 2 {
			if sepchar, oks := retList[1].(ExprNumInt); oks && sepchar > -1 && sepchar < 256 {
				_, okc := retList[3].(*ExprCall)
				if okf, _ := retList[3].(ExprFuncRef); okc || okf == StdFuncNil {
					if initialoutput := ListToBytes(prog.ListOfExprs(retList[3])); initialoutput != nil {
						initialstate, handlenextinput := retList[2], func(prevstate Expr, input []byte) (nextstate Expr) {
							retexpr := prog.Eval(&ExprCall{Callee: fnhandler, Args: []Expr{ListFrom(input), prevstate}}) //  &ExprCall{Callee: fnhandler, Arg: prevstate}, Arg: ListFrom(input)})
							if retlist := prog.ListOfExprs(retexpr); len(retlist) == 2 {
								nextstate = retlist[0]
								if outlist := prog.ListOfExprs(retlist[1]); outlist != nil {
									os.Stdout.Write(ListToBytes(outlist))
								} else {
									os.Stderr.WriteString("RET-EXPR:\t" + retlist[1].JsonSrc() + "\n")
								}
							}
							if nextstate == nil {
								panic(len(retList))
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
