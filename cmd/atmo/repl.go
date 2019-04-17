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
		repl.Run(true,
			"", "This is a read-eval-print loop (repl).",
			"", "— repl directives begin with `:`,", "  any other inputs are eval'd",
			"", "— a line ending in "+repl.IO.MultiLineSuffix+" starts", "  or concludes a multi-line input",
			"", "— for proper line-editing, run this repl", "  via `rlwrap` or `rlfe` or equivalent.",
			"",
		)
		repl.Ctx.Dispose()
	} else {
		println(err.Error())
	}
}
