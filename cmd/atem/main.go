package main

import (
	"os"
	"runtime"
	"runtime/debug"

	"github.com/metaleap/atem"
)

var ctx *atem.Ctx

func writeLn(s string) { _, _ = os.Stdout.WriteString(s + "\n") }

func main() {
	debug.SetGCPercent(-1) // GC off
	runtime.GOMAXPROCS(1)  // no thread scheduling

	if len(os.Args) == 1 {
		os.Args = append(os.Args, "")
	}

	var err error
	ctx, err = atem.New(".", true)
	if err != nil {
		panic(err)
	}

	atcmd := os.Args[1]
	switch atcmd {
	case "help", "version", "run":
		writeLn("command " + atcmd + " recognized but not yet implemented")
	case "repl":
		mainRepl()
	case "fmt":
		mainFmt()
	default:
		writeLn("Usage:")
		writeLn("\tatem repl")
		writeLn("\tatem fmt")
		writeLn("\tatem run")
		writeLn("\tatem help")
		writeLn("\tatem version")
	}
}
