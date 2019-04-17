package main

import (
	"github.com/metaleap/atmo/tooling/repl"
	"time"
)

var (
	replMultiLineSuffix     = ",,,"
	replAdditionalLibsDirs  []string
	replLibsWatchPauseAfter = 83 * time.Second
)

func mainRepl() {
	var repl atmorepl.Repl
	repl.IO.MultiLineSuffix = replMultiLineSuffix
	repl.Ctx.Dirs.Libs = replAdditionalLibsDirs
	repl.Ctx.LibsWatch.Should = func() bool {
		return replLibsWatchPauseAfter == 0 || time.Since(repl.IO.TimeLastInput) < replLibsWatchPauseAfter
	}

	if err := repl.Ctx.Init("."); err == nil {
		repl.Run(true)
		repl.Ctx.Dispose()
	} else {
		println(err.Error())
	}
}
