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
}

func (me *Repl) Run() (err error) {
	me.IO.Stdin, me.IO.Stderr, me.IO.Stdout =
		ustd.IfNil(me.IO.Stdin, os.Stdin).(io.Reader), ustd.IfNil(me.IO.Stderr, os.Stderr).(io.Writer), ustd.IfNil(me.IO.Stdout, os.Stdout).(io.Writer)

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
					me.writeLn("...and we're done.")
					return
				} else {
					me.writeLn("unknown directive: `:" + directive + "` — try: ")
					me.writeLn("\t:q — quit")
				}
			default:
				if out, err := me.Ctx.ReadEvalPrint(readln); err != nil {
					println(err.Error())
				} else {
					me.writeLn(out.String())
				}
			}
		}
	}

	return
}

func (me *Repl) writeErr(s string) {
	_, _ = me.IO.Stderr.Write([]byte(s))
}

func (me *Repl) writeErrLn(s string) {
	b := append(make([]byte, 0, len(s)+1), s...)
	_, _ = me.IO.Stderr.Write(append(b, '\n'))
}

func (me *Repl) write(s string) {
	_, _ = me.IO.Stdout.Write([]byte(s))
}

func (me *Repl) writeLn(s string) {
	b := append(make([]byte, 0, len(s)+1), s...)
	_, _ = me.IO.Stdout.Write(append(b, '\n'))
}
