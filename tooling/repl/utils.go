package atmorepl

import (
	"io"
	"os"
	"time"

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

	me.IO.TimeLastInput, me.IO.Stdin, me.IO.Stderr, me.IO.Stdout =
		time.Now(), ustd.IfNil(me.IO.Stdin, os.Stdin).(io.Reader), ustd.IfNil(me.IO.Stderr, os.Stderr).(io.Writer), ustd.IfNil(me.IO.Stdout, os.Stdout).(io.Writer)
	me.IO.writeLns, me.IO.printLns, me.IO.write =
		ustd.WriteLines(me.IO.Stdout), ustd.WriteLines(me.IO.Stderr), func(s string, n int) {
			if n > 0 {
				_, _ = me.IO.Stdout.Write(ustr.Repeat(s, n))
			}
		}

	me.initEnsureDefaultDirectives()
}

func (me *Repl) decoInputStart(altStyle bool) {
	me.decoCtxMsgsIfAny(false)
	me.run.multiLnInputHadLeadingTabs = false
	me.IO.writeLns(ustr.If(altStyle, "╔", "┌") + sepLine)
	me.decoInputBeginLine(altStyle, "")
}

func (me *Repl) decoInputDone(altStyle bool) {
	me.IO.writeLns(ustr.If(altStyle, "╚", "└") + sepLine)
	me.decoCtxMsgsIfAny(false)
}

func (me *Repl) decoInputBeginLine(altStyle bool, andThen string) {
	me.IO.write(ustr.If(altStyle, "║", "│"), 1)
	if me.IO.write(" ", me.run.indent); len(andThen) > 0 {
		me.IO.write(andThen, 1)
	}
}

func (me *Repl) decoAddNotice(altStyle bool, altPrefix string, compact bool, noticeLines ...string) {
	for i := range noticeLines {
		if i == 0 {
			noticeLines[i] = ustr.If(altPrefix != "", altPrefix, ustr.If(altStyle, "╠══ ", "├── ")) + noticeLines[i]
		} else {
			noticeLines[i] = "    " + noticeLines[i]
		}
	}
	if !compact {
		noticeLines = append(noticeLines, "", "")
	}
	me.IO.writeLns(noticeLines...)
}

func (me *Repl) decoCtxMsgsIfAny(initial bool) {
	if msgs := me.Ctx.Messages(true); len(msgs) > 0 {
		for i := range msgs {
			if me.decoAddNotice(true, "▓▒░ ", true, msgs[i].Time.Format("15:04:05"), msgs[i].Text); !initial {
				time.Sleep(42 * time.Millisecond) // this is to easier notice they're there
			}
		}
		me.IO.writeLns("", "")
	}
}

func (me *Repl) decoTypingAnim(s string, speed time.Duration) {
	for _, r := range s {
		time.Sleep(speed)
		me.IO.write(string(r), 1)
	}
}

func (me *Repl) decoWelcomeMsgAnim() {
	me.IO.writeLns("")
	me.decoInputStart(false)
	time.Sleep(345 * time.Millisecond)
	me.decoTypingAnim(":info\n", 234*time.Millisecond)
	me.decoInputDone(false)
	me.DInfo("")
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
