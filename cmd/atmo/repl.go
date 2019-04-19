package main

import (
	"os"
	"time"

	"github.com/go-forks/go-ps"
	"github.com/go-leap/str"
	"github.com/go-leap/sys"
	"github.com/metaleap/atmo/tooling/repl"
)

var (
	replCurDir               = "."
	replMultiLineSuffix      = ",,,"
	replAdditionalPacksDirs  []string
	replPacksWatchPauseAfter = 83 * time.Second
)

func mainRepl() {
	var repl atmorepl.Repl
	repl.IO.MultiLineSuffix = replMultiLineSuffix
	repl.Ctx.Dirs.Packs = replAdditionalPacksDirs
	repl.Ctx.OngoingPacksWatch.ShouldNow = func() bool {
		return replPacksWatchPauseAfter == 0 || time.Since(repl.IO.TimeLastInput) < replPacksWatchPauseAfter
	}
	if err := repl.Ctx.Init(replCurDir); err == nil {
		usys.OnSigint(func() {
			repl.QuitNonDirectiveInitiated(true)
			repl.Ctx.Dispose()
			os.Exit(0)
		})
		repl.Run(true,
			"", "This is a read-eval-print loop (repl).",
			"", "— repl directives start with `:`,", "  any other inputs are eval'd",
			"", "— a line ending in "+repl.IO.MultiLineSuffix+" introduces", "  or concludes a multi-line input",
			"", "- see optional flags via `atmo help`",
			"", ustr.If(replRunsVia("rlwrap", "rlfe") != "", "",
				"— for smooth line-editing, run this repl\n  via `rlwrap` or `rlfe` or equivalent.\n\n"),
		)
		repl.Ctx.Dispose()
	} else {
		println(err.Error())
	}
}

func replRunsVia(parentProcessNames ...string) (isRunVia string) {
	defer func() { _ = recover() }() // go-ps not doing all the bounds-checks it could be doing
	if ppid := os.Getppid(); ppid != 0 && len(parentProcessNames) > 0 {
		if p, _ := ps.FindProcess(ppid); p != nil {
			parentexename := p.Executable()
			for _, ppn := range parentProcessNames {
				if ppn == parentexename {
					return ppn
				}
			}
		}
	}
	return
}
