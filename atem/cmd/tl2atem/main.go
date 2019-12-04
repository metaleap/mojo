package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	. "github.com/metaleap/atmo/atem"
	tl "github.com/metaleap/go-machines/toylam"
)

var (
	inProg   tl.Prog
	outProg  = make(Prog, 0, 1024)
	instr2op = map[tl.Instr]OpCode{
		tl.InstrADD: OpAdd,
		tl.InstrDIV: OpDiv,
		tl.InstrEQ:  OpEq,
		tl.InstrGT:  OpGt,
		tl.InstrLT:  OpLt,
		tl.InstrMOD: OpMod,
		tl.InstrMUL: OpMul,
		tl.InstrSUB: OpSub,
		tl.InstrMSG: OpPrt,
		tl.InstrERR: -1010101,
	}
)

func main() {
	srcfilepath, dstdirpath := os.Args[1], os.Args[2]
	if err := os.MkdirAll(dstdirpath, os.ModePerm); err != nil {
		panic(err)
	}
	srcdirpath := filepath.Dir(srcfilepath)
	files, err := ioutil.ReadDir(srcdirpath)
	if err != nil {
		panic(err)
	}
	modules := make(map[string][]byte, len(files))
	for _, file := range files {
		if curfilepath := filepath.Join(srcdirpath, file.Name()); !file.IsDir() {
			if idxdot := strings.LastIndexByte(file.Name(), '.'); (curfilepath == srcfilepath) || (idxdot > 0 && file.Name()[idxdot:] == ".tl") {
				if src, err := ioutil.ReadFile(curfilepath); err == nil {
					modules[file.Name()[:idxdot]] = src
				} else {
					panic(err)
				}
			}
		}
	}
	srcfilename, srcfileext := filepath.Base(srcfilepath), filepath.Ext(srcfilepath)
	maintopdefqname := srcfilename[:len(srcfilename)-len(srcfileext)] + ".main"
	dstfilepath := filepath.Join(dstdirpath,
		maintopdefqname[:len(maintopdefqname)-len(".main")]+".json")

	inProg.ParseModules(modules, tl.ParseOpts{KeepNameRefs: true, KeepOpRefs: true, KeepRec: true, KeepSepLocals: true})
	compile(maintopdefqname)

	// outProg[2] = FuncDef{Meta: []string{"stdFalse", "discard"}, Args: []int{0}, Body: StdFuncId} // TODO! temporary until selectors recognized for Eval
	outProg = fixFuncDefArgsUsageNumbers(outProg)
	prefixNameMetasWithIdxs()
	ioutil.WriteFile(dstfilepath+".non-opt", []byte(outProg.JsonSrc(false)), os.ModePerm)
	println("Compilation done, optimizing...")

	outProg = optimize(outProg)
	prefixNameMetasWithIdxs()
	for i := range outProg {
		var walker func(expr Expr) Expr
		walker = func(expr Expr) Expr {
			if call, _ := expr.(*ExprCall); call != nil {
				if subcall, _ := call.Callee.(*ExprCall); subcall != nil {
					panic("convTo got buggy")
				}
				call.Callee = walker(call.Callee)
				for i := range call.Args {
					call.Args[i] = walker(call.Args[i])
				}
				if fnref, ok := call.Callee.(ExprFuncRef); ok && fnref > 0 && len(outProg[fnref].Args) > len(call.Args) {
					call.Curried = len(outProg[fnref].Args) - len(call.Args)
				}
			}
			return expr
		}
		outProg[i].Body = walk(convTo(outProg[i].Body), walker)
	}
	ioutil.WriteFile(dstfilepath, []byte(outProg.JsonSrc(true)), os.ModePerm)
	println("...done.")
}

func prefixNameMetasWithIdxs() {
	for i := 0; i < len(outProg)-1; i++ {
		pos := strings.IndexByte(outProg[i].Meta[0], ']')
		outProg[i].Meta[0] = "[" + strconv.Itoa(i) + "]" + outProg[i].Meta[0][pos+1:]
	}
}
