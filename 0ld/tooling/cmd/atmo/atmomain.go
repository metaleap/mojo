package main

import (
	"os"

	"github.com/go-leap/std"
	"github.com/go-leap/str"
	"github.com/go-leap/sys"
	"github.com/metaleap/atmo/0ld"
	"github.com/metaleap/atmo/0ld/tooling/repl"
)

var (
	writeLns = ustd.WriteLines(os.Stdout)
)

func main() {
	ustd.Flags.AddShortNames, ustd.Flags.OnErr = true, func(_ string, _ string, e error) { panic(e) }
	replDirSession = ustd.FlagOfString("repl-dir-session", replDirSession,
		"    format: text (a dir path)\n    treats the given dir as a (\"faux\") kit even if not in a kitstash dir search path")
	replDirCache = ustd.FlagOfString("repl-dir-cache", replDirCache,
		"    format: text (a dir path), currently not yet in use")
	replDirsAdditionalKits = ustd.FlagOfStrings("repl-dirs-kits", replDirsAdditionalKits, string(os.PathListSeparator),
		"    format: text (1 or more kitstash dir search paths sep'd by `"+string(os.PathListSeparator)+"`)\n    will be used in addition to those in $"+atmo.EnvVarKitsDirs)
	replMultiLineSuffix = ustd.FlagOfString("repl-ux-longsuffix", replMultiLineSuffix,
		"    format: text; sets the line suffix that begins or ends a multi-line input")
	atmorepl.Ux.MoreLines = int(ustd.FlagOfUint("repl-ux-morelines", uint64(atmorepl.Ux.MoreLines),
		"    format: integral number >= 0;\n    enables `more`-like output page breaks every n lines"))
	atmorepl.Ux.AnimsEnabled = ustd.FlagOfBool("repl-ux-anims", atmorepl.Ux.AnimsEnabled,
		"    format: one of 1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False")
	atmorepl.Ux.WelcomeMsgShow = ustd.FlagOfBool("repl-ux-intro", atmorepl.Ux.WelcomeMsgShow,
		"    format: one of 1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False")

	atcmd, showinfousage, showinfoargs := usys.Arg(1), false, false
	switch atcmd {
	case "version", "run":
		writeLns("command `" + atcmd + "` recognized but not yet implemented")
	case "help":
		showinfoargs, showinfousage = true, true
	case "tinker", "repl":
		mainRepl()
	case "fmt":
		mainFmt()
	default:
		showinfousage = true
	}

	if showinfousage {
		writeLns("", "Usage:", "",
			"  atmo help     ─── infos on --options (wordy)",
			"  atmo version  ─── not yet implemented",
			"  atmo tinker   ─── live code&play -ground",
			"  atmo fmt      ─── not yet implemented",
			"  atmo run      ─── not yet implemented",
			"")
	}

	if f := ustd.Flags.Known; showinfoargs {
		writeLns("", "Optional flags:", "")
		for i := range f {
			if f[i].Desc != "" {
				writeLns("  --"+f[i].Name+" (or --"+f[i].ShortName()+") ─── default value: "+ustr.If(f[i].Default != "", f[i].Default, "‹empty›"), f[i].Desc, "")
			}
		}
	}
}
