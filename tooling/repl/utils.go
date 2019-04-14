package atemrepl

import (
	"io"
	"os"

	"github.com/go-leap/std"
	"github.com/go-leap/str"
)

func (me *Repl) init() {
	me.quit, me.IO.Stdin, me.IO.Stderr, me.IO.Stdout =
		false, ustd.IfNil(me.IO.Stdin, os.Stdin).(io.Reader), ustd.IfNil(me.IO.Stderr, os.Stderr).(io.Writer), ustd.IfNil(me.IO.Stdout, os.Stdout).(io.Writer)
	me.IO.writeLns, me.IO.printLns, me.IO.write =
		ustd.WriteLines(me.IO.Stdout), ustd.WriteLines(me.IO.Stderr), func(s string, n int) {
			if n > 0 {
				_, _ = me.IO.Stdout.Write(ustr.Repeat(s, n))
			}
		}

	if me.KnownDirectives == nil {
		me.KnownDirectives = map[string]func(string){}
	}
	me.KnownDirectives["q · quit"] = me.DQuit
	me.KnownDirectives["h · help"] = me.DWelcome
}
