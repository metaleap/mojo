package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
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
		tl.InstrERR: -424242424242424242,
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
	ioutil.WriteFile(dstfilepath, []byte(outProg.String()), os.ModePerm)
}
