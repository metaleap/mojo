package atmorepl

import (
	"io"
	"os"

	"github.com/go-leap/std"
	"github.com/go-leap/str"
)

const (
	multiLnMinIndent = 2
)

var (
	sepLine = ustr.Times("─", 41)
)

func (me *Repl) init() {
	me.run.quit, me.run.indent, me.run.multiLnInputHadLeadingTabs = false, 0, false

	me.IO.Stdin, me.IO.Stderr, me.IO.Stdout =
		ustd.IfNil(me.IO.Stdin, os.Stdin).(io.Reader), ustd.IfNil(me.IO.Stderr, os.Stderr).(io.Writer), ustd.IfNil(me.IO.Stdout, os.Stdout).(io.Writer)
	me.IO.writeLns, me.IO.printLns, me.IO.write =
		ustd.WriteLines(me.IO.Stdout), ustd.WriteLines(me.IO.Stderr), func(s string, n int) {
			if n > 0 {
				_, _ = me.IO.Stdout.Write(ustr.Repeat(s, n))
			}
		}

	me.initEnsureDefaultDirectives()
}

func (me *Repl) decoInputStart() {
	me.run.multiLnInputHadLeadingTabs = false
	me.IO.writeLns("┌" + sepLine)
	me.decoInputAddLine()
}

func (me *Repl) decoInputDone() {
	me.IO.writeLns("└" + sepLine)
}

func (me *Repl) decoInputAddLine() {
	me.IO.write("│", 1)
	me.IO.write(" ", me.run.indent)
}

func (me *Repl) decoAddNotice(compact bool, noticeLines ...string) {
	for i := range noticeLines {
		if i == 0 {
			noticeLines[i] = "├── " + noticeLines[i]
		} else {
			noticeLines[i] = "    " + noticeLines[i]
		}
	}
	if !compact {
		noticeLines = append(noticeLines, "", "")
	}
	me.IO.writeLns(noticeLines...)
}

func trimAndCountPrefixRunes(s string) (trimmed string, count int, numtabs int) {
	for _, r := range s {
		if r == ' ' {
			count++
		} else if r == '\t' {
			count++
			numtabs++
		} else {
			break
		}
	}
	if trimmed = s; count > 0 {
		trimmed = s[count:]
	}
	return
}
