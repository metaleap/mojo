package atemrepl

import (
	"bufio"
	"io"
	"os"

	"github.com/go-leap/std"
	"github.com/go-leap/str"
	"github.com/metaleap/atem"
)

type Repl struct {
	Ctx *atem.Ctx
	IO  struct {
		Stdin  io.Reader
		Stdout io.Writer
		Stderr io.Writer
	}
	writeLns, printLns func(...string)
}

func (me *Repl) Run() (err error) {
	me.IO.Stdin, me.IO.Stderr, me.IO.Stdout =
		ustd.IfNil(me.IO.Stdin, os.Stdin).(io.Reader), ustd.IfNil(me.IO.Stderr, os.Stderr).(io.Writer), ustd.IfNil(me.IO.Stdout, os.Stdout).(io.Writer)
	me.writeLns, me.printLns = ustd.WriteLines(me.IO.Stdout), ustd.WriteLines(me.IO.Stderr)

	multiln, repl := "", bufio.NewScanner(os.Stdin)
	for repl.Scan() {
		if readln := ustr.Trim(repl.Text()); readln != "" {
			if ustr.Suff(readln, "...") {
				if multiln == "" {
					multiln = readln[:len(readln)-len("...")] + "\n    "
					continue
				} else {
					readln, multiln = ustr.Trim(multiln+readln[:len(readln)-len("...")]), ""
				}
			}
			switch {
			case multiln != "":
				multiln += readln + "\n    "
			case readln[0] == ':':
				directive, _ := ustr.BreakOnFirstOrPref(readln[1:], " ")
				if directive == "q" {
					me.writeLns("...and we're done.")
					return
				} else {
					me.writeLns("unknown directive: `:"+directive+"` — try: ",
						"\t:q — quit")
				}
			default:
				if out, err := me.Ctx.ReadEvalPrint(readln); err != nil {
					me.printLns(err.Error())
				} else {
					me.writeLns(out.String())
				}
			}
		}
	}

	return
}
