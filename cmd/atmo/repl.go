package main

import (
	"os"
	"time"

	"github.com/go-forks/go-ps"
	"github.com/go-leap/str"
	"github.com/go-leap/sys"
	"github.com/metaleap/atmo/load"
	"github.com/metaleap/atmo/tooling/repl"
)

var (
	replMultiLineSuffix      = ",,,"
	replDirSession           = "."
	replDirCache             = atmoload.CtxDefaultCacheDirPath()
	replDirsAdditionalPacks  []string
	replPacksWatchPauseAfter = 83 * time.Second
)

func mainRepl() {
	var repl atmorepl.Repl
	repl.IO.MultiLineSuffix, repl.Ctx.Dirs.Packs, repl.Ctx.Dirs.Cache, repl.Ctx.Dirs.Session = replMultiLineSuffix, replDirsAdditionalPacks, replDirCache, replDirSession
	repl.Ctx.OngoingPacksWatch.ShouldNow = func() bool {
		return replPacksWatchPauseAfter == 0 || time.Since(repl.IO.TimeLastInput) < replPacksWatchPauseAfter
	}
	if err := repl.Ctx.Init(); err == nil {
		usys.OnSigint(func() {
			repl.QuitNonDirectiveInitiated(true)
			repl.Ctx.Dispose()
			os.Exit(0)
		})
		repl.Run(true,
			"", "This is a read-eval-print loop (repl).",
			"", "— repl commands start with `:`, any other", "  inputs are eval'd as atmo expressions",
			"", "— in case of the latter, a line ending in "+repl.IO.MultiLineSuffix, "  introduces or concludes a multi-line input",
			"", "- to see --flags, quit and run `atmo help`",
			ustr.If(replRunsVia("rlwrap", "rlfe") != "", "",
				"\n— for smooth line-editing, run the repl\n  via `rlwrap` or `rlfe` or equivalent\n"),
			ustr.If(atmorepl.StdoutUx.MoreLines <= 0, "",
				"— every "+ustr.Plu(atmorepl.StdoutUx.MoreLines, "line")+", further output is held back\n  until ‹enter›ing on the `"+ustr.Trim(string(atmorepl.StdoutUx.MoreLinesPrompt))+"` prompt shown\n\n"),
		)
		repl.Ctx.Dispose()
	} else {
		println(err.Error())
	}
}

func replRunsVia(parentProcessNames ...string) string {
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
	return ""
}
