package main

import (
	"os"

	"github.com/go-leap/std"
	"github.com/go-leap/sys"
)

var (
	writeLns = ustd.WriteLines(os.Stdout)
)

func main() {
	atcmd := usys.Arg(1)
	switch atcmd {
	case "help", "version", "run":
		writeLns("command " + atcmd + " recognized but not yet implemented")
	case "repl":
		mainRepl()
	case "fmt":
		mainFmt()
	default:
		writeLns("Usage:",
			"\tatem repl",
			"\tatem fmt",
			"\tatem run",
			"\tatem help",
			"\tatem version")
	}
}
