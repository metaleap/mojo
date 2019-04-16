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

	warns, err := repl.Ctx.Init(".")
	for _, e := range warns {
		println(e.Error())
	}
	if err == nil {
		repl.Run(true)
		repl.Ctx.Dispose()
	}
	if err != nil {
		println(err.Error())
	}
}
