package main

import (
	"os"

	"github.com/metaleap/mojo"
)

var ctx mo.ICtx

func writeLn(s string) { _, _ = os.Stdout.WriteString(s + "\n") }

func main() {
	if len(os.Args) == 1 {
		os.Args = append(os.Args, "")
	}

	moctx, err := mo.New(".")
	if err != nil {
		panic(err)
	}
	ctx = moctx

	mocmd := os.Args[1]
	switch mocmd {
	case "help", "version", "run":
		writeLn("command " + mocmd + " recognized but not yet implemented")
	case "repl":
		mainRepl()
	default:
		writeLn("Usage:")
		writeLn("\tmojo repl")
		writeLn("\tmojo run")
		writeLn("\tmojo help")
		writeLn("\tmojo version")
		writeLn("\ndefaulting to `repl`:\n\n")
		mainRepl()
	}
}
