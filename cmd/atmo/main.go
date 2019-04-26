package main

import (
	"os"
	"time"

	"github.com/go-leap/std"
	"github.com/go-leap/str"
	"github.com/go-leap/sys"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/load"
	"github.com/metaleap/atmo/tooling/repl"
)

var (
	writeLns = ustd.WriteLines(os.Stdout)
)

func main() {
	ustd.FlagsAddShortNames = true
	replMultiLineSuffix = ustd.FlagOfString("repl-multiline-suffix", replMultiLineSuffix,
		"    format: text; sets the line suffix that begins or ends a multi-line input")
	replDirsAdditionalPacks = ustd.FlagOfStrings("repl-dirs-packs", replDirsAdditionalPacks, string(os.PathListSeparator),
		"    format: text (1 or more packs dir paths joined by `"+string(os.PathListSeparator)+"`)\n    will be used in addition to those in $"+atmo.EnvVarPacksDirs)
	replDirSession = ustd.FlagOfString("repl-dir-session", replDirSession,
		"    format: text (1 dir path)")
	replDirCache = ustd.FlagOfString("repl-dir-cache", replDirCache,
		"    format: text (1 dir path)")
	atmoload.PacksWatchInterval = ustd.FlagOfDuration("repl-packswatch-interval", 123*time.Millisecond,
		"    format: time-duration; sets how often to check all known atmo packs for\n    file-modifications to reload accordingly. Disable with a zero duration\n    (doing so will make available the `:reload` repl command).")
	replPacksWatchPauseAfter = ustd.FlagOfDuration("repl-packswatch-pauseafter", replPacksWatchPauseAfter,
		"    format: time-duration; sets how soon (since the most-recent line input)\n    packs-watching will pause (to be resumed on the next line input)")
	atmorepl.AnimsDisabled = ustd.FlagOfBool("repl-anims-disabled", atmorepl.AnimsDisabled,
		"")

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
			if f[i].Desc != "" {
				writeLns("  --"+f[i].Name+" or --"+f[i].ShortName()+" (default: "+ustr.If(f[i].Default != "", f[i].Default, "<empty>")+")", f[i].Desc, "")
			}
		}
	}
}
