package main

import (
	"time"

	"github.com/go-leap/sys"
	"github.com/metaleap/atmo/tooling/repl"
)

var (
	replMultiLineSuffix      = ",,,"
	replAdditionalPacksDirs  []string
	replPacksWatchPauseAfter = 83 * time.Second
)

func mainRepl() {
	var repl atmorepl.Repl
	repl.IO.MultiLineSuffix = replMultiLineSuffix
	repl.Ctx.Dirs.Packs = replAdditionalPacksDirs
	repl.Ctx.PacksWatch.Should = func() bool {
		return replPacksWatchPauseAfter == 0 || time.Since(repl.IO.TimeLastInput) < replPacksWatchPauseAfter
	}
	if err := repl.Ctx.Init("."); err == nil {
		usys.OnSigint(repl.Quit)
		repl.Run(true,
			"", "This is a read-eval-print loop (repl).",
			"", "— repl directives start with `:`,", "  any other inputs are eval'd",
			"", "— a line ending in "+repl.IO.MultiLineSuffix+" introduces", "  or concludes a multi-line input",
			"", "— for smooth line-editing, run this repl", "  via `rlwrap` or `rlfe` or equivalent.",
			"",
		)
		repl.Ctx.Dispose()
	} else {
		println(err.Error())
	}
}
