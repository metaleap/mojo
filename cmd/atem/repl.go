package main

import (
	"github.com/metaleap/atem/tooling/repl"
)

func mainRepl() {
	var err error
	var repl atemrepl.Repl
	repl.IO.MultiLineSuffix = ",,,"

	if err = repl.Ctx.Init("."); err == nil {
		err = repl.Run(true)
	}
	if err != nil {
		println(err.Error())
	}
}
