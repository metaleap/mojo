package main

/*

very low-level (thus unidiomatic) code because we want to have a code base we can
transliterate into the initial language iteration that's largely LLVM-IR-like
(just somewhat more human-readable/writable), hence:

- program input and output remain as `type Str = []byte`, not `string`s
- no `append`s, instead use of usually-over-sized but non-growing buffer-ish
  arrays (with ad-hoc inline custom `len` tracking)
- all `make` / `new` to be replaced by custom fixed-buffer allocator scheme
- no (non-empty) interface type decls, just empty `interface{}` with type
  switches, for later transliterating into pseudo-tagged-union-ish style
- no methods or struct-embeds, no big point when losing them anyway in IR

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

	toks := tokenize(input_src_file_bytes, false)
	assert(len(toks) != 0)

	ast := parse(toks, input_src_file_bytes)
	assert(len(ast.defs) != 0)

	println(len(toks), len(ast.defs))
}
