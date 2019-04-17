package main

import (
	"os"
	"time"

	"github.com/go-leap/std"
	"github.com/go-leap/str"
	"github.com/go-leap/sys"
	"github.com/metaleap/atmo"
)

var (
	writeLns = ustd.WriteLines(os.Stdout)
)

func main() {
	ustd.FlagsAddShortNames = true
	replAdditionalLibsDirs = ustd.FlagOfStrings("repl-libsdirs", replAdditionalLibsDirs, string(os.PathListSeparator),
		"    format: text (1 or more libs search-paths joined by `"+string(os.PathListSeparator)+"`)\n    will be used in addition to those in $"+atmo.EnvVarLibDirs)
	replMultiLineSuffix = ustd.FlagOfString("repl-multiline-suffix", replMultiLineSuffix,
		"    format: text; sets the line suffix that begins or ends a multi-line input")
	atmo.LibsWatchInterval = ustd.FlagOfDuration("repl-libswatch-interval", 123*time.Millisecond,
		"    format: time-duration; sets how often to check all known atmo libs-dirs for\n    file-modifications to reload accordingly. Disable with a zero duration.")
	replLibsWatchPauseAfter = ustd.FlagOfDuration("repl-libswatch-pauseafter", replLibsWatchPauseAfter,
		"    format: time-duration; sets after how long (since the last line input)\n    libs-watching is paused (to be resumed on the next line input)")

	atcmd, showinfousage, showinfoargs := usys.Arg(1), false, false
	switch atcmd {
	case "version", "run":
		writeLns("command " + atcmd + " recognized but not yet implemented")
	case "help":
		showinfoargs, showinfousage = true, true
	case "repl":
		mainRepl()
	case "fmt":
		mainFmt()
	default:
		showinfousage = true
	}

	if showinfousage {
		writeLns("", "Usage:", "",
			"  atmo repl",
			"  atmo fmt",
			"  atmo run",
			"  atmo help",
			"  atmo version", "")
	}

	if f := ustd.Flags; showinfoargs {
		writeLns("", "Optional flags:", "")
		for i := range f {
			writeLns("  --"+f[i].Name+" or --"+f[i].ShortName()+" (default: "+ustr.If(f[i].Default != "", f[i].Default, "<empty>")+")", f[i].Desc, "")
		}
	}
}
