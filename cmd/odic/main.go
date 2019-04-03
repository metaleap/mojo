package main

import (
	"os"
	"runtime"
	"runtime/debug"

	"github.com/metaleap/odic"
)

var ctx *od.Ctx

func writeLn(s string) { _, _ = os.Stdout.WriteString(s + "\n") }

func main() {
	debug.SetGCPercent(-1) // GC off
	runtime.GOMAXPROCS(1)  // no thread scheduling

	if len(os.Args) == 1 {
		os.Args = append(os.Args, "")
	}

	var err error
	ctx, err = od.New(".", true)
	if err != nil {
		panic(err)
	}

	odcmd := os.Args[1]
	switch odcmd {
	case "help", "version", "run":
		writeLn("command " + odcmd + " recognized but not yet implemented")
	case "repl":
		mainRepl()
	case "fmt":
		mainFmt()
	default:
		writeLn("Usage:")
		writeLn("\todic repl")
		writeLn("\todic fmt")
		writeLn("\todic run")
		writeLn("\todic help")
		writeLn("\todic version")
	}
}
