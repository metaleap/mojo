package main

import (
	"io/ioutil"
	"os"
	"runtime"
	"runtime/debug"
)

func main() {
	runtime.LockOSThread()
	debug.SetGCPercent(-1)

	input_src_file_bytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	toks := tokenize(input_src_file_bytes, false)
	assert(len(toks) != 0)
	toksCheckBrackets(toks)

	ast := parse(toks, input_src_file_bytes)
	assert(len(ast.defs) != 0)
}
