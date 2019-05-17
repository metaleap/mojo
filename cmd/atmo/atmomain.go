package main

import (
	"os"
	"time"

	"github.com/go-leap/std"
	"github.com/go-leap/str"
	"github.com/go-leap/sys"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/session"
	"github.com/metaleap/atmo/tooling/repl"
)

var (
	writeLns = ustd.WriteLines(os.Stdout)
)

func main() {
	ustd.FlagsAddShortNames, ustd.FlagsOnErr = true, func(_ string, _ string, e error) { panic(e) }
	replMultiLineSuffix = ustd.FlagOfString("repl-ux-longsuffix", replMultiLineSuffix,
		"    format: text; sets the line suffix that begins or ends a multi-line input")
	replDirSession = ustd.FlagOfString("repl-dir-session", replDirSession,
		"    format: text (a dir path)\n    treats the given dir as a kit even if not in a kits-dir search path")
	replDirCache = ustd.FlagOfString("repl-dir-cache", replDirCache,
		"    format: text (a dir path)")
	replDirsAdditionalKits = ustd.FlagOfStrings("repl-dirs-kits", replDirsAdditionalKits, string(os.PathListSeparator),
		"    format: text (1 or more kits-dir search paths sep'd by `"+string(os.PathListSeparator)+"`)\n    will be used in addition to those in $"+atmo.EnvVarKitsDirs)
	atmorepl.Ux.KitsWatchInfoFlagName = "repl-kitswatch-interval"
	atmosess.KitsWatchInterval = ustd.FlagOfDuration(atmorepl.Ux.KitsWatchInfoFlagName, 1234*time.Millisecond,
		"    format: time-duration; sets how often to check all known atmo kits for\n    file-modifications to reload accordingly. Disable with a zero duration\n    (doing so will make available the `:reload` repl command).")
	replKitsWatchPauseAfter = ustd.FlagOfDuration("repl-kitswatch-pauseafter", replKitsWatchPauseAfter,
		"    format: time-duration; sets how soon after the most-recent line input\n    kits-watching will suspend (to be resumed on the next line input)")
	atmorepl.Ux.MoreLines = int(ustd.FlagOfUint("repl-ux-morelines", uint64(atmorepl.Ux.MoreLines),
		"    format: integral number >= 0;\n    enables `more`-like output page breaks every n lines"))
	atmorepl.Ux.AnimsEnabled = ustd.FlagOfBool("repl-ux-anims", atmorepl.Ux.AnimsEnabled,
		"    format: one of 1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False")

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
				writeLns("  --"+f[i].Name+" (or --"+f[i].ShortName()+") ─── default value: "+ustr.If(f[i].Default != "", f[i].Default, "‹empty›"), f[i].Desc, "")
			}
		}
	}
}
