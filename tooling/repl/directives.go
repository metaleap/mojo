package atemrepl

import (
	"github.com/go-leap/str"
)

func (me *Repl) initEnsureDefaultDirectives() {
	kd := me.KnownDirectives.ensure
	kd("q · quit", me.DQuit)
	kd("h · help", me.DWelcomeMsg)
	kd("l · list <libs|defs>", me.DList)
	kd("i · info on [\"libpath\"] [name]", me.DInfo)
}

type directive struct {
	Desc string
	Run  func(string) bool
}

type directives []directive

func (me *directives) ensure(desc string, run func(string) bool) {
	if found := me.By(desc[0]); found != nil {
		found.Desc, found.Run = desc, run
	} else {
		*me = append(*me, directive{Desc: desc, Run: run})
	}
}

func (me directives) By(letter byte) *directive {
	for i := range me {
		if me[i].Desc[0] == letter {
			return &me[i]
		}
	}
	return nil
}

func (me *Repl) DQuit(string) bool {
	me.run.quit = true
	return true
}

func (me *Repl) DWelcomeMsg(string) bool {
	me.IO.writeLns(
		"", "— repl directives begin with `:`,\n  any other inputs are eval'd",
		"", "— a line ending in "+me.IO.MultiLineSuffix+" begins\n  or ends a multi-line input",
		"", "— for proper line-editing, run via\n  `rlwrap` or `rlfe` or similar.",
		"",
	)
	return true
}

func (me *Repl) DList(what string) bool {
	switch what {
	case "libs":
		libs := me.Ctx.KnownLibs()
		for i := range libs {
			errstr, lib := "", &libs[i]
			if numerrs := len(lib.Errs()); numerrs > 0 {
				errstr = ustr.Int(numerrs) + " error(s)"
			}
			me.decoAddNotice(true, ustr.Combine("\""+lib.LibPath+"\"", " ══!══ ", errstr), lib.DirPath)
		}
	default:
		return false
	}
	return true
}

func (me *Repl) DInfo(what string) bool {
	var whatlib, whatname string
	if whatname = what; what[0] == '"' {
		whatlib, whatname = ustr.BreakOnFirstOrPref(what[1:], "\"")
	}
	if whatname == "" {
		lib := me.Ctx.Lib(whatlib)
		if lib == nil {
			me.IO.writeLns("unknown lib: `" + whatlib + "`, see known libs via `:l libs`")
		} else {
			me.IO.writeLns("\""+lib.LibPath+"\"", lib.DirPath)
			for _, e := range lib.Errs() {
				me.IO.writeLns("══!══ " + e.Error())
			}
		}
		return true
	} else {

	}
	return false
}
