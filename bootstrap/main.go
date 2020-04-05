package main

/*

very low-level and unidiomatic code! because we want to have a most compact
code base to later transliterate into the initial language iteration that's
largely LLVM-IR-like (just somewhat more human-readable/writable), hence:

- program input and output remain as `type Str = []byte`, no use of `string`s
- no proper `error` handling, immediate `panic`s upon detecting a problem
- no 3rd-party imports whatsoever
- no stdlib imports for core processing (just program setup & I/O)
  (hence manual implementations like uintToStr, uintFromStr, join etc)
- no `append`s, instead use of usually-over-sized but non-growing arrays being
  sliced by means of ad-hoc inline "custom" / "manual" `len` tracking
- all `make` / `new` to be replaced by custom fixed-buffer allocation scheme
- no (non-empty) interface type decls, just empty `interface{}` with type switch
  where needed, for later transliterating into low-level tagged-union-ish style
- no methods or struct-embeds, no point because we won't have them in IR either

*/

import (
	"io/ioutil"
	"os"
	"runtime"
	"runtime/debug"
)

func main() {
	runtime.LockOSThread()
	debug.SetGCPercent(-1)
	runtime.GOMAXPROCS(1)
	debug.SetMaxThreads(6) // should be 1 really, but Go has other ideas.. anything under 6 runtime-panics, ridiculously.

	input_src_file_name := os.Args[1]
	input_src_file_bytes, err := ioutil.ReadFile(input_src_file_name)
	if err != nil {
		panic(err)
	}

	toks := tokenize(input_src_file_bytes)
	assert(len(toks) != 0)

	ast := parse(toks, input_src_file_bytes)
	assert(len(ast.defs) != 0)

	println(len(toks), "\t\t", len(ast.defs))
}
