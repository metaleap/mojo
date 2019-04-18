package atmorepl

import (
	"bufio"
	"io"
	"os"
	"time"

	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
)

type Repl struct {
	Ctx             atmo.Ctx
	KnownDirectives directives
	IO              struct {
		Stdin           io.Reader
		Stdout          io.Writer
		Stderr          io.Writer
		MultiLineSuffix string
		TimeLastInput   time.Time

		write              func(string, int)
		writeLns, printLns func(...string)
	}

	// current mutable state during a `Run` loop
	run struct {
		quit                       bool
		indent                     int
		multiLnInputHadLeadingTabs bool
		welcomeMsgLines            []string
	}
}

func (me *Repl) Run(showWelcomeMsg bool, welcomeMsgLines ...string) {
	me.init()
	if me.decoCtxMsgsIfAny(true); showWelcomeMsg {
		me.run.welcomeMsgLines = welcomeMsgLines
		me.decoWelcomeMsgAnim()
	}
	me.decoInputStart(false)
	for multiln, readln := "", bufio.NewScanner(os.Stdin); (!me.run.quit) && readln.Scan() && (!me.run.quit); {
		me.IO.TimeLastInput = time.Now()
		inputln, numleadingspaces, numleadingtabs := trimAndCountPrefixRunes(readln.Text())
		me.run.multiLnInputHadLeadingTabs = me.run.multiLnInputHadLeadingTabs || (len(multiln) > 0 && numleadingtabs > 0)

		if inputln == "" {
			if me.run.indent > multiLnMinIndent {
				if me.run.indent -= multiLnMinIndent; me.run.indent%2 != 0 {
					me.run.indent++
				}
			}
			me.decoInputBeginLine(false, "")
			continue
		}

		if ismultiln, isdefbegin := ustr.Suff(inputln, me.IO.MultiLineSuffix), (multiln == "" && ustr.Suff(inputln, ":=")); isdefbegin || ismultiln {
			if multiln == "" {
				if inputln[0] != ':' {
					if me.run.indent, multiln = multiLnMinIndent, inputln[:len(inputln)-len(me.IO.MultiLineSuffix)]+"\n  "; isdefbegin {
						multiln = inputln + "\n  "
					}
					me.decoInputBeginLine(false, "")
					continue
				}
			} else if multiln, me.run.indent, inputln = "", 0, ustr.Trim(multiln+inputln[:len(inputln)-len(me.IO.MultiLineSuffix)]); inputln == "" {
				me.decoInputDone(false)
				me.decoInputStart(false)
				continue
			}
		}

		switch {

		// just another line to add to current multi-line input?
		case multiln != "":
			me.run.indent += numleadingspaces
			multiln += ustr.Times(" ", numleadingspaces) + inputln + "\n" + ustr.Times(" ", me.run.indent)
			me.decoInputBeginLine(false, "")
			continue

		// else, a directive?
		case inputln[0] == ':':
			me.decoInputDone(false)
			me.IO.writeLns("")
			if me.runDirective(ustr.BreakOnFirstOrPref(inputln[1:], " ")); !me.run.quit {
				me.IO.writeLns("", "")
				me.decoInputStart(false)
			}
		// else, input to be EVAL'd now
		default:
			me.decoInputDone(false)
			if me.run.multiLnInputHadLeadingTabs {
				me.decoAddNotice(false, "", false, "multi-line input had leading tabs,note", "that repl auto-indent is based on spaces")
			}
			if out, err := me.Ctx.ReadEvalPrint(inputln); err != nil {
				me.IO.printLns(err.Error())
			} else {
				me.IO.writeLns(out.String())
			}
			me.decoInputStart(false)
		}
	}
}

func (me *Repl) Quit() {
	me.run.quit = true
	me.decoTypingAnim(" :quit   \n", 42*time.Millisecond)
	me.decoInputDone(false)
	me.IO.writeLns("")
	os.Exit(0)
}
