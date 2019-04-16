package main

import (
	"os"

	"github.com/go-leap/std"
	"github.com/go-leap/sys"
	"github.com/metaleap/atmo"
)

var (
	writeLns = ustd.WriteLines(os.Stdout)
)

func main() {
	atmo.LibWatchInterval = ustd.FlagOfDuration("lib_watch_interval", atmo.LibWatchInterval,
		"    format: time-duration; sets how often to check all known atmo libs-dirs for\n    file-modifications to reload accordingly. Disable with a zero duration.")

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
			writeLns("  --"+f[i].Name+" (default: "+f[i].Default+")", f[i].Desc, "")
		}
	}
}
