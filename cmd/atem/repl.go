package main

import (
	"github.com/metaleap/atem/tooling/repl"
)

func mainRepl() {
	var err error
	var repl atemrepl.Repl
	repl.IO.MultiLineSuffix = ",,,"

	if err = repl.Ctx.Init("."); err == nil {
		writeLns(
			"", "— repl directives begin with `:`,\n  all other inputs are eval'd",
			"", "— a line ending in "+repl.IO.MultiLineSuffix+" either begins\n  or ends a multi-line input",
			"", "— for line-editing, consider using\n  `rlwrap` or some equivalent",
			"",
		)
		err = repl.Run()
	}
	if err != nil {
		println(err.Error())
	}
}
