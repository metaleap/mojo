package main

import (
	"github.com/metaleap/atmo/tooling/repl"
)

var (
	replMultiLineSuffix    = ",,,"
	replAdditionalLibsDirs []string
)

func mainRepl() {
	var repl atmorepl.Repl
	repl.IO.MultiLineSuffix = replMultiLineSuffix
	repl.Ctx.Dirs.Libs = replAdditionalLibsDirs

	if err := repl.Ctx.Init("."); err == nil {
		repl.Run(true)
		repl.Ctx.Dispose()
	} else {
		println(err.Error())
	}
}
