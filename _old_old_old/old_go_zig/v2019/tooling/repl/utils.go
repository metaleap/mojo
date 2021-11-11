package atmorepl

import (
	"io"
	"os"
	"time"

	"github.com/go-leap/std"
	"github.com/go-leap/str"
	. "github.com/metaleap/atmo/old/v2019"
)

const (
	multiLnMinIndent = 2
)

var (
	sepLine = ustr.Times("─", 77)
)

func (me *Repl) init() {
	me.run.quit, me.run.indent, me.run.multiLnInputHadLeadingTabs = false, 0, false

	me.IO.TimeLastInput, me.IO.Stdin, me.IO.Stderr, me.IO.Stdout =
		time.Now(), ustd.IfNil(me.IO.Stdin, os.Stdin).(io.Reader), ustd.IfNil(me.IO.Stderr, os.Stderr).(io.Writer), ustd.IfNil(me.IO.Stdout, os.Stdout).(io.Writer)

	if Ux.MoreLines > 0 {
		orig := me.IO.Stdout
		more := ustd.Writer{Writer: orig}
		more.On.Byte, more.On.AfterEveryNth, more.On.ButDontCountImmediateRepeats, more.On.Do =
			'\n', Ux.MoreLines, true, func(int) bool {
				orig.Write(Ux.MoreLinesPrompt)
				_, _ = me.IO.Stdin.Read([]byte{0})
				return false
			}
		me.IO.Stdout = &more
	}
	me.uxMore(false) // initially

	me.IO.writeLns, me.IO.printLns, me.IO.write =
		ustd.WriteLines(me.IO.Stdout), ustd.WriteLines(me.IO.Stderr), func(s string, n int) {
			if n > 0 {
				_, _ = me.IO.Stdout.Write(ustr.Repeat(s, n))
			}
		}

	me.Ctx.Options.BgMsgs.IncludeLiveKitsErrs = true
	me.Ctx.Options.Scratchpad.FauxFileNameForErrorMessages = "‹repl›"
	me.initEnsureDefaultDirectives()
}
func (me *Repl) decoInputStart(altStyle bool, msgsSummaryOnly bool) {
	time.Sleep(1 * time.Millisecond)
	if !msgsSummaryOnly {
		me.decoCtxMsgsIfAny(false)
	} else if msgcount := me.Ctx.BackgroundMessagesCount(); msgcount > 3 {
		me.decoAddNotice(true, "", true, ustr.Int(msgcount)+" new notices while processing your input, show via any input:", "", "")
	}
	me.run.multiLnInputHadLeadingTabs = false
	me.uxMore(false)
	me.IO.writeLns(ustr.If(altStyle, "╔", "┌") + sepLine)
	me.decoInputBeginLine(altStyle, "")
}

func (me *Repl) decoInputDone(altStyle bool) {
	me.decoInputDoneBut(altStyle, true, 0)
}

func (me *Repl) decoInputDoneBut(altStyle bool, showMsgsIfAny bool, caretPos int) {
	sepline := sepLine
	if l := len("─"); caretPos > 0 {
		caretPos = caretPos * l
		for caretPos > len(sepline) {
			sepline = sepline + sepline
		}
		sepline = sepline[l:caretPos] + "╬" + sepline[caretPos:]
	}
	me.IO.writeLns(ustr.If(altStyle, "╚", "└") + sepline)
	me.uxMore(true)
	if me.Ctx.CatchUpOnFileMods(); showMsgsIfAny {
		me.decoCtxMsgsIfAny(false)
	}
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

func (me *Repl) decoMsgNotice(bg bool, lines ...string) {
	for i := 0; i < len(lines); i++ {
		ln := lines[i]
		if pos := ustr.Pos(ln, ": ["); pos > 0 && ustr.Has(ln[:pos], SrcFileExt+":") {
			prefix, suffix := lines[:i], lines[i+1:]
			i, lines = i+1, append(append(prefix, ln[:pos], ln[pos+2:]), suffix...)
		}
	}
	me.decoAddNotice(false, ustr.If(bg,
		ustr.If(Ux.OldSchoolTty, "≡‹≡ ", "░▒▓ "),
		ustr.If(Ux.OldSchoolTty, "≡›≡ ", "▓▒░ "),
	), true, lines...)
}

func (me *Repl) decoCtxMsgsIfAny(initial bool) {
	if msgs := me.Ctx.BackgroundMessages(true); len(msgs) > 0 {
		me.IO.writeLns("", "")
		if !initial {
			if me.IO.writeLns("Also noticed in the meantime:"); Ux.AnimsEnabled {
				time.Sleep(567 * time.Millisecond)
			}
		}
		for i := range msgs {
			msg := &msgs[i]
			if lines := msg.Lines; len(lines) > 0 {
				lines[0] = msg.Time.Format("15:04:05") + ustr.If(msg.Issue, " ══════ ", " ────── ") + lines[0]
				me.decoMsgNotice(true, lines...)
				if (!initial) && Ux.AnimsEnabled {
					time.Sleep(42 * time.Millisecond) // this is to easier notice they're there
				}
			}
		}
		me.IO.writeLns("", "")
	}
}

func (me *Repl) decoTypingAnim(s string, speed time.Duration) {
	for _, r := range s {
		if Ux.AnimsEnabled {
			time.Sleep(speed)
		}
		me.IO.write(string(r), 1)
	}
}

func (me *Repl) decoWelcomeMsgAnim() {
	me.IO.writeLns("")
	if me.decoInputStart(false, false); Ux.AnimsEnabled {
		time.Sleep(456 * time.Millisecond)
	}
	me.decoTypingAnim(":intro\n", 123*time.Millisecond)
	me.decoInputDone(false)
	me.uxMore(false)
	me.IO.writeLns("")
	me.DIntro("")
	me.IO.writeLns("", "")
}

func (me *Repl) uxMore(restartIfTrueElseSuspend bool) {
	if w, _ := me.IO.Stdout.(*ustd.Writer); w != nil {
		if restartIfTrueElseSuspend {
			w.RestartOnDo()
		} else {
			w.SuspendOnDo()
		}
	}
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
