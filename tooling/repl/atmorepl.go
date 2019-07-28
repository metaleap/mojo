package atmorepl

import (
	"bufio"
	"io"
	"os"
	"time"

	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/session"
)

var (
	Ux struct {
		AnimsEnabled    bool
		MoreLines       int
		MoreLinesPrompt []byte
		WelcomeMsgLines []string
		OldSchoolTty    bool
		WelcomeMsgShow  bool
	}
)

type Repl struct {
	Ctx             atmosess.Ctx
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
	}
}

func init() {
	Ux.AnimsEnabled, Ux.WelcomeMsgShow, Ux.MoreLinesPrompt = true, true, []byte("     ¶¶¶")
}

func (me *Repl) Run(loadSessDirFauxKit bool, loadKitsByImpPaths ...string) {
	me.init()
	me.decoCtxMsgsIfAny(true)
	if Ux.WelcomeMsgShow {
		me.decoWelcomeMsgAnim()
	}

	me.Ctx.KitsEnsureLoaded(loadSessDirFauxKit, append([]string{ /*atmo.NameAutoKit*/ },
		ustr.Sans(loadKitsByImpPaths, atmo.NameAutoKit)...)...)

	me.decoInputStart(false, false)
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
				me.decoInputStart(false, false)
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
			if me.runDirective(ustr.BreakOnFirstOrPref(ustr.Trim(inputln[1:]), " ")); !me.run.quit {
				me.IO.writeLns("", "")
				me.decoInputStart(false, false)
			}

		// else, input to be EVAL'd now
		default:
			var caretpos int
			kit := me.Ctx.KitByImpPath("")
			preduced, errs := me.Ctx.ScratchpadEntry(kit, "", inputln)
			if len(errs) > 0 && (!ustr.Has(inputln, "\n")) {
				for _, err := range errs {
					if err0pos := err.Pos(); err0pos != nil && err0pos.Ln1 == 1 {
						caretpos = err0pos.Col1 + numleadingspaces
						break
					}
				}
			}
			me.decoInputDoneBut(false, false, caretpos)
			if caretpos == 0 {
				me.IO.writeLns("")
			} else {
				me.IO.write("╔", 1)
				me.IO.write("═", caretpos-1)
				me.IO.writeLns("╝")
			}
			for _, e := range errs {
				me.decoMsgNotice(false, e.Error())
			}
			if me.run.multiLnInputHadLeadingTabs {
				me.decoAddNotice(false, "", false, "multi-line input had leading tabs, note", "that repl auto-indent is based on spaces")
			}
			if preduced != nil {
				me.IO.writeLns(preduced.SummaryCompact())
			}

			me.IO.writeLns("", "")
			me.decoInputStart(false, false)
		}
	}
	if !me.run.quit {
		me.QuitNonDirectiveInitiated(true)
	}
}

func (me *Repl) QuitNonDirectiveInitiated(anim bool) {
	if me.run.quit = true; anim {
		me.decoTypingAnim(" :quit   \n", 42*time.Millisecond)
		me.decoInputDone(false)
		me.IO.writeLns("")
	}
}
