package main

import (
	"github.com/go-leap/sys"
	"github.com/metaleap/atem"
)

var (
	ctx     *atem.Ctx
	writeLn = usys.WriteLn
)

func main() {
	atcmd := usys.Arg(1)
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
