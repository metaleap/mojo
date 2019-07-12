package main

import (
	"os"

	"github.com/go-forks/go-ps"
	"github.com/go-leap/str"
	"github.com/go-leap/sys"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/session"
	"github.com/metaleap/atmo/tooling/repl"
)

var (
	replMultiLineSuffix    = ",,,"
	replDirSession         = "."
	replDirCache           = atmosess.CtxDefaultCacheDirPath()
	replDirsAdditionalKits []string
)

func mainRepl() {
	atmo.Options.Sorts = true
	var repl atmorepl.Repl
	repl.IO.MultiLineSuffix, repl.Ctx.Dirs.Kits, repl.Ctx.Dirs.Cache = replMultiLineSuffix, replDirsAdditionalKits, replDirCache
	if err := repl.Ctx.Init(false, replDirSession); err == nil {
		usys.OnSigint(func() {
			repl.QuitNonDirectiveInitiated(true)
			// repl.Ctx.Dispose()
			os.Exit(0)
		})
		atmorepl.Ux.WelcomeMsgLines = []string{
			"Now you're in a read-eval-print loop (\"repl\").",
			"", "─ demands (like `:quit`) start with `:`, all", "  other inputs are eval'd as atmo expressions",
			"", "─ in the latter case, multi-line inputs are started", "  and finished respectively by a line ending in " + repl.IO.MultiLineSuffix,
			"", "─ for infos on --options,", "  quit and run `atmo help`",
		}
		if atmorepl.Ux.OldSchoolTty = (replRunsVia("login") == "login"); replRunsVia("rlwrap", "rlfe") == "" {
			atmorepl.Ux.WelcomeMsgLines = append(atmorepl.Ux.WelcomeMsgLines, "", "─ for sane input-editing, quit and run", "  via `rlwrap` or `rlfe` or equivalent")
		}
		if atmorepl.Ux.MoreLines > 0 {
			atmorepl.Ux.WelcomeMsgLines = append(atmorepl.Ux.WelcomeMsgLines, "", "─ every "+ustr.Plu(atmorepl.Ux.MoreLines, "line")+", further output is held back", "  until ‹enter›ing on the `"+ustr.Trim(string(atmorepl.Ux.MoreLinesPrompt))+"` prompt shown")
		}
		repl.Run(true, true)
		// repl.Ctx.Dispose()
	} else {
		println(err.Error())
	}
}

func replRunsVia(parentProcessNames ...string) string {
	defer func() { _ = recover() }() // go-ps not doing all the bounds-checks it could be doing

	for ppid := os.Getppid(); ppid != 0; {
		if proc, _ := ps.FindProcess(ppid); proc == nil {
			ppid = 0
		} else {
			ppid = proc.PPid()
			parentexename := proc.Executable()
			for _, ppn := range parentProcessNames {
				if ppn == parentexename {
					return ppn
				}
			}
		}
	}
	return ""
}
