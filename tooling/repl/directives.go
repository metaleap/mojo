package atmorepl

import (
	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
)

func (me *Repl) initEnsureDefaultDirectives() {
	kd := me.KnownDirectives.ensure
	kd("quit", me.DQuit)
	kd("info [\"libpath\"] [name]", me.DInfo)
	kd("list <libs | defs | \"libpath\">", me.DList)
	if atmo.LibWatchInterval == 0 {
		kd("reload\n      (reloads modified code in known libs)", me.DReload)
	}
}

type directive struct {
	Desc string
	Run  func(string) bool
}

func (me *directive) Name() string { return ustr.Until(me.Desc, " ") }

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

func (me *Repl) runDirective(name string, args string) {
	var found *directive
	if len(name) > 0 {
		if found = me.KnownDirectives.By(name); found != nil {
			if args = ustr.Trim(args); !found.Run(args) {
				me.IO.writeLns("directive `:"+found.Name()+"` does not understand `"+args+"`,", "as a reminder:", "   :"+found.Desc)
			}
		}
	}
	if found == nil {
		me.IO.writeLns("unknown directive `:" + name + "` — try: ")
		for i := range me.KnownDirectives {
			me.IO.writeLns("   :" + me.KnownDirectives[i].Desc)
		}
	}
}

func (me *Repl) DQuit(string) bool {
	me.run.quit = true
	return true
}

func (me *Repl) DReload(string) bool {
	me.Ctx.ReloadModifiedLibsUnlessAlreadyWatching()
	return true
}

func (me *Repl) DList(what string) bool {
	switch what {
	case "libs":
		me.dListLibs()
	default:
		return false
	}
	return true
}

func (me *Repl) dListLibs() {
	libs := me.Ctx.KnownLibs()
	me.IO.writeLns("", ustr.Int(len(libs))+" known libs:")
	for i := range libs {
		lib := &libs[i]
		numerrs := len(lib.Errs())
		me.decoAddNotice(true, "\""+lib.LibPath+"\""+ustr.If(numerrs == 0, "", " ── "+ustr.Int(numerrs)+" error(s)"))
	}
	me.IO.writeLns("", "were found in the following "+ustr.Int(len(me.Ctx.Dirs.Libs)), "currently active search paths:")
	me.IO.writeLns(ustr.Map(me.Ctx.Dirs.Libs, func(s string) string { return "─── " + s })...)
	me.IO.writeLns("", "for lib details use `:info \"<lib>\"`")
}

func (me *Repl) DInfo(what string) bool {
	if what == "" {
		me.dInfo()
	} else {
		var whatlib, whatname string
		if whatname = what; what[0] == '"' {
			whatlib, whatname = ustr.BreakOnFirstOrPref(what[1:], "\"")
		}
		if whatname == "" {
			me.dInfoLib(whatlib)
		} else {
			me.dInfoDef(whatlib, whatname)
		}
	}
	return true
}

func (me *Repl) dInfo() {
	me.IO.writeLns(
		"", "— repl directives begin with `:`,\n  any other inputs are eval'd",
		"", "— a line ending in "+me.IO.MultiLineSuffix+" begins\n  or ends a multi-line input",
		"", "— for proper line-editing, run via\n  `rlwrap` or `rlfe` or similar.",
		"",
	)
}

func (me *Repl) dInfoLib(whatLib string) {
	lib := me.Ctx.Lib(whatLib)
	if lib == nil {
		me.IO.writeLns("unknown lib: `" + whatLib + "`, see known libs via `:list libs`")
	} else {
		me.IO.writeLns("\""+lib.LibPath+"\"", lib.DirPath)

		if liberrs := lib.Errs(); len(liberrs) > 0 {
			me.IO.writeLns("", ustr.Int(len(liberrs))+" error/s:")
			for i := range liberrs {
				errmsg := liberrs[i].Error()
				if pos := ustr.Pos(errmsg, ": ["); pos > 0 && ustr.Has(errmsg[:pos], atmo.SrcFileExt+":") {
					me.decoAddNotice(true, errmsg[:pos], errmsg[pos+2:])
				} else {
					me.decoAddNotice(true, errmsg)
				}
			}
		}
	}
}

func (me *Repl) dInfoDef(whatLib string, whatName string) {
	me.IO.writeLns("Info on name: " + whatName + " in \"" + whatLib + "\"")
}
