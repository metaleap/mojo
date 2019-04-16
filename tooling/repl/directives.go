package atmorepl

import (
	"github.com/go-leap/str"
)

func (me *Repl) initEnsureDefaultDirectives() {
	kd := me.KnownDirectives.ensure
	kd("quit", me.DQuit)
	kd("info [\"libpath\"] [name]", me.DInfo)
	kd("list <libs|defs>", me.DList)
}

type directive struct {
	Desc string
	Run  func(string) bool
}

type directives []directive

func (me *directives) ensure(desc string, run func(string) bool) {
	if found := me.By(desc); found != nil {
		found.Desc, found.Run = desc, run
	} else {
		*me = append(*me, directive{Desc: desc, Run: run})
	}
}

func (me directives) By(name string) *directive {
	for i := range me {
		if ustr.Pref(me[i].Desc, name) {
			return &me[i]
		}
	}
	return nil
}

func (me *Repl) DQuit(string) bool {
	me.run.quit = true
	return true
}

func (me *Repl) DList(what string) bool {
	switch what {
	case "libs":
		libs := me.Ctx.KnownLibs()
		me.IO.writeLns("", ustr.Int(len(libs))+" known libs:")
		for i := range libs {
			lib := &libs[i]
			numerrs := len(lib.Errs())
			me.decoAddNotice(true, "\""+lib.LibPath+"\""+ustr.If(numerrs == 0, "", " ── "+ustr.Int(numerrs)+" error(s)"))
		}
		me.IO.writeLns("", "were found in the following "+ustr.Int(len(me.Ctx.Dirs.Libs)), "currently active search paths:")
		me.IO.writeLns(ustr.Map(me.Ctx.Dirs.Libs, func(s string) string { return "─── " + s })...)
		me.IO.writeLns("", "for lib details, type `:i \"<lib>\"`")
	default:
		return false
	}
	return true
}

func (me *Repl) DInfo(what string) bool {
	if what == "" {
		me.IO.writeLns(
			"", "— repl directives begin with `:`,\n  any other inputs are eval'd",
			"", "— a line ending in "+me.IO.MultiLineSuffix+" begins\n  or ends a multi-line input",
			"", "— for proper line-editing, run via\n  `rlwrap` or `rlfe` or similar.",
			"",
		)
		return true
	}

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
	} else {
		me.IO.writeLns("Info on name: " + whatname + " in \"" + whatlib + "\"")
	}
	return true
}
