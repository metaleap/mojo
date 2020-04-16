package main

/*

very low-level and unidiomatic code! because we want to have a most compact
code base to later transliterate into our initial language iteration that'll
begin most limited and low-levelish at first:

- program input and output remain as `type Str = []byte`, no use of `string`s
- no proper `error` handling, immediate `panic`s upon detecting a problem
- no 3rd-party imports whatsoever
- no stdlib imports for core processing (just here in main.go for setup & I/O)
  (hence manual implementations like uintToStr, uintFromStr, strEql etc)
- no `append`s, instead use of usually-over-sized but non-growing arrays being
  sliced by means of ad-hoc inline "custom" / "manual" `len` tracking
- all `make` / `new` to be replaced by custom fixed-buffer allocation scheme
- no (non-empty) interface type decls, just empty/marker `interface{}`s to be
  switched on, for later transliterating into low-level tagged-union-ish style
- no methods or struct-embeds, no point because we won't have them in IL either
- naming / casing conventions follow WIP target lang rather than Go customs

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

	input_src_file_bytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	toks := tokenize(input_src_file_bytes, false)
	assert(len(toks) != 0)
	toksCheckBrackets(toks)

	ast := parse(toks, input_src_file_bytes)
	assert(len(ast.defs) != 0)
	astPopulateScopes(&ast)

	ir_hl := irHLFrom(&ast)
	irHLDump(&ir_hl)
}
