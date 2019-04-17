package atmorepl

import (
	"path/filepath"

	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
)

func (me *Repl) initEnsureDefaultDirectives() {
	kd := me.KnownDirectives.ensure
	kd("quit", me.DQuit)
	kd("info [\"libpath\"] [name]", me.DInfo)
	kd("list <libs | defs | \"libpath\">", me.DList)
	if atmo.LibWatchInterval == 0 {
		kd("reload", me.DReload) //\n      (reloads modified code in known libs)", me.DReload)
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
	if what == "" || ustr.Pref("libs", what) {
		me.dListLibs()
	} else if ustr.Pref("defs", what) {
		me.dListDefs("")
	} else {
		me.dListDefs(what)
	}
	return true
}

func (me *Repl) dListLibs() {
	me.IO.writeLns("From current search paths:")
	me.IO.writeLns(ustr.Map(me.Ctx.Dirs.Libs, func(s string) string { return "─── " + s })...)

	libs := me.Ctx.KnownLibs()
	me.IO.writeLns("", "found "+ustr.Int(len(libs))+" libs:")
	for i := range libs {
		lib := &libs[i]
		numerrs := len(lib.Errs())
		me.decoAddNotice(true, "\""+lib.LibPath+"\""+ustr.If(numerrs == 0, "", " ── "+ustr.Int(numerrs)+" error(s)"))
	}
	me.IO.writeLns("", "(to see lib details, use `:info \"<lib>\"`)")
}

func (me *Repl) dListDefs(whatLib string) {
	if whatLib != "" && whatLib[0] == '"' && len(whatLib) > 1 {
		whatLib = ustr.If(whatLib[len(whatLib)-1] != '"', whatLib[1:], whatLib[1:len(whatLib)-1])
	}
	if whatLib == "" {
		me.IO.writeLns("TODO: list all defs")
	} else if lib := me.Ctx.Lib(whatLib); lib == nil {
		me.IO.writeLns("unknown lib: `" + whatLib + "`, see known libs via `:list libs`")
	} else {
		me.IO.writeLns("", "\""+lib.LibPath+"\"", "    "+lib.DirPath)
		for i := range lib.SrcFiles {
			sf := &lib.SrcFiles[i]
			nd, _ := sf.CountTopLevelDefs()
			me.IO.writeLns("", ustr.Int(nd)+" top-level def(s) in "+filepath.Base(sf.SrcFilePath)+":")
			for d := range sf.TopLevel {
				if def := sf.TopLevel[d].Ast.Def; def != nil {
					pos := ustr.If(!def.Name.Tokens[0].Meta.Position.IsValid(), "",
						"(line "+ustr.Int(def.Name.Tokens[0].Meta.Position.Line)+")")
					me.decoAddNotice(true, ustr.Combine(def.Name.Val, " ─── ", pos))
				}
			}
		}

		// if liberrs := lib.Errs(); len(liberrs) > 0 {
		// 	me.IO.writeLns("", ustr.Int(len(liberrs))+" issue(s) in lib \""+whatLib+"\":")
		// 	for i := range liberrs {
		// 		errmsg := liberrs[i].Error()
		// 		if pos := ustr.Pos(errmsg, ": ["); pos > 0 && ustr.Has(errmsg[:pos], atmo.SrcFileExt+":") {
		// 			me.decoAddNotice(true, errmsg[:pos], errmsg[pos+2:])
		// 		} else {
		// 			me.decoAddNotice(true, errmsg)
		// 		}
		// 	}
		// }

	}
}

func (me *Repl) DInfo(what string) bool {
	if what == "" {
		me.dInfo()
	} else {
		var whatlib, whatname string
		if whatname = what; what[0] == '"' {
			whatlib, whatname = ustr.BreakOnFirstOrPref(what[1:], "\"")
		}
		if whatlib, whatname = ustr.Trim(whatlib), ustr.Trim(whatname); whatname == "" {
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
		me.IO.writeLns("\""+lib.LibPath+"\"", "    "+lib.DirPath)

		me.IO.writeLns("", ustr.Int(len(lib.SrcFiles))+" source file(s) in lib \""+whatLib+"\":")
		numlines, numlinesnet, numdefs, numdefsinternal := 0, 0, 0, 0
		for i := range lib.SrcFiles {
			sf := &lib.SrcFiles[i]
			nd, ndi := sf.CountTopLevelDefs()
			sloc := sf.CountNetLinesOfCode()
			numlines, numlinesnet, numdefs, numdefsinternal = numlines+sf.LastLoad.NumLines, numlinesnet+sloc, numdefs+nd, numdefsinternal+ndi
			me.decoAddNotice(true, filepath.Base(sf.SrcFilePath), ustr.Int(sf.LastLoad.NumLines)+" lines ("+ustr.Int(sloc)+" net), "+ustr.Int(nd)+" top-level defs, "+ustr.Int(nd-ndi)+" exported")
		}
		me.IO.writeLns("Total: "+ustr.Int(numlines)+" lines ("+ustr.Int(numlinesnet)+" net), "+ustr.Int(numdefs)+" top-level defs, "+ustr.Int(numdefs-numdefsinternal)+" exported",
			"    (counts exclude failed-to-parse code portions, if any)")

		if liberrs := lib.Errs(); len(liberrs) > 0 {
			me.IO.writeLns("", ustr.Int(len(liberrs))+" issue(s) in lib \""+whatLib+"\":")
			for i := range liberrs {
				errmsg := liberrs[i].Error()
				if pos := ustr.Pos(errmsg, ": ["); pos > 0 && ustr.Has(errmsg[:pos], atmo.SrcFileExt+":") {
					me.decoAddNotice(true, errmsg[:pos], errmsg[pos+2:])
				} else {
					me.decoAddNotice(true, errmsg)
				}
			}
		}

		me.IO.writeLns("", "", "(to see lib defs, use `:list \""+whatLib+"\"`)")
	}
}

func (me *Repl) dInfoDef(whatLib string, whatName string) {
	me.IO.writeLns("Info on name: " + whatName + " in \"" + whatLib + "\"")
}
