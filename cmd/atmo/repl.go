package main

import (
	"github.com/metaleap/atmo/tooling/repl"
)

func mainRepl() {
	var err error
	var repl atmorepl.Repl
	repl.IO.MultiLineSuffix = ",,,"

	if err = repl.Ctx.Init("."); err == nil {
		repl.Run(true)
		repl.Ctx.Dispose()
	}
	if err != nil {
		println(err.Error())
	}
}
