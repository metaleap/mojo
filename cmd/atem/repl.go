package main

import (
	"github.com/metaleap/atem"
	"github.com/metaleap/atem/tooling/repl"
)

func mainRepl() {
	writeLns("atem repl:",
		"— directives are prefixed with `:`",
		"— a line ending in `...` either begins\n  or ends a multi-line input",
		"— enter any atem definition or expression")

	var err error
	var repl atemrepl.Repl
	if repl.Ctx, err = atem.New("."); err == nil {
		err = repl.Run()
	}

	if err != nil {
		println(err.Error())
	}
}
