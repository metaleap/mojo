package atemrepl

import (
	"bufio"
	"io"
	"os"

	"github.com/go-leap/str"
	"github.com/metaleap/atem"
)

type Repl struct {
	Ctx atem.Ctx
	IO  struct {
		Stdin           io.Reader
		Stdout          io.Writer
		Stderr          io.Writer
		MultiLineSuffix string

		write              func(string, int)
		writeLns, printLns func(...string)
	}

	KnownDirectives directives

	quit bool
}

func (me *Repl) Run(showWelcomeMsg bool) (err error) {
	if me.init(); showWelcomeMsg {
		me.DWelcomeMsg("")
	}
	const mlmi = /* multiline minindent */ 2
	multiln, indent, inputhadleadingtabs, sepline := "", 0, false, ustr.Times("─", 41)
	decoinputaddline := func() {
		me.IO.write("│", 1)
		me.IO.write(" ", indent)
	}
	decoinputstart, decoinputdone, decoaddnotice := func() {
		inputhadleadingtabs = false
		me.IO.writeLns("┌" + sepline)
		decoinputaddline()
	}, func() {
		me.IO.writeLns("└" + sepline)
	}, func(notice string) {
		me.IO.writeLns("", "├── "+notice, "")
	}
	decoinputstart()
	for readln := bufio.NewScanner(os.Stdin); (!me.quit) && readln.Scan(); {
		inputln, numleadingspaces, numleadingtabs := trimAndCountPrefixRunes(readln.Text())
		if numleadingtabs > 0 {
			inputhadleadingtabs = len(multiln) > 0
		}
		if inputln == "" {
			if indent > mlmi {
				if indent -= mlmi; indent%2 != 0 {
					indent++
				}
			}
			decoinputaddline()
		} else {
			if neat := (multiln == "" && ustr.Suff(inputln, " :=")); neat || ustr.Suff(inputln, me.IO.MultiLineSuffix) {
				if multiln == "" {
					if inputln[0] != ':' {
						if indent, multiln = mlmi, inputln[:len(inputln)-len(me.IO.MultiLineSuffix)]+"\n  "; neat {
							multiln = inputln + "\n  "
						}
						decoinputaddline()
						continue
					}
				} else if multiln, indent, inputln = "", 0, ustr.Trim(multiln+inputln[:len(inputln)-len(me.IO.MultiLineSuffix)]); inputln == "" {
					decoinputdone()
					decoinputstart()
					continue
				}
			}
			switch {
			case multiln != "":
				indent += numleadingspaces
				multiln += ustr.Times(" ", numleadingspaces) + inputln + "\n" + ustr.Times(" ", indent)
				decoinputaddline()
				continue
			case inputln[0] == ':':
				decoinputdone()
				dletter, dargs := ustr.BreakOnFirstOrPref(inputln[1:], " ")
				var found *directive
				if len(dletter) > 0 {
					if found = me.KnownDirectives.By(dletter[0]); found != nil {
						found.Run(dargs)
					}
				}
				if found == nil {
					me.IO.writeLns("unknown directive `:" + dletter + "` — try: ")
					for i := range me.KnownDirectives {
						me.IO.writeLns("\t:" + me.KnownDirectives[i].Desc)
					}
				}
				if !me.quit {
					decoinputstart()
				}
			default:
				decoinputdone()
				if inputhadleadingtabs {
					decoaddnotice("input had leading tabs, use spaces for indent")
				}
				if out, err := me.Ctx.ReadEvalPrint(inputln); err != nil {
					me.IO.printLns(err.Error())
				} else {
					me.IO.writeLns(out.String())
				}
				decoinputstart()
			}
		}
	}
	return
}
