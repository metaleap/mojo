package main

/*

very low-level and unidiomatic code! because we want to have a most compact
code base to later transliterate into the initial language iteration that's
largely LLVM-IR-like (just somewhat more human-readable/writable), hence:

- program input and output remain as `type Str = []byte`, no use of `string`s
- no proper `error` handling, immediate `panic`s upon detecting a problem
- no 3rd-party imports whatsoever
- no stdlib imports for core processing (just program setup & I/O)
  (hence manual implementations like uintToStr, uintFromStr, strEql etc)
- no `append`s, instead use of usually-over-sized but non-growing arrays being
  sliced by means of ad-hoc inline "custom" / "manual" `len` tracking
- all `make` / `new` to be replaced by custom fixed-buffer allocation scheme
- no (non-empty) interface type decls, just empty `interface{}` with type switch
  where needed, for later transliterating into low-level tagged-union-ish style
- no methods or struct-embeds, no point because we won't have them in IR either
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
	runtime.GOMAXPROCS(1)

	name_main := Str("main")
	input_src_file_bytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	toks := tokenize(input_src_file_bytes, false)
	assert(len(toks) != 0)
	toksCheckBrackets(toks)

	ast := parse(toks, input_src_file_bytes)
	assert(len(ast.defs) != 0)
	astHoistLocalDefsToTopDefs(&ast, name_main)

	ir := irFromAst(&ast, &AstExpr{kind: AstExprIdent(name_main)})
	irReduceDefs(&ir)

	ll_mod := llModuleFrom(&ir, name_main)
	llResolveHoleTypes(&ll_mod)
	llEmit(&ll_mod)
}
