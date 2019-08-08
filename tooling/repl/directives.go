package atmorepl

import (
	"path/filepath"

	"github.com/go-leap/std"
	"github.com/go-leap/str"
	. "github.com/metaleap/atmo"
	. "github.com/metaleap/atmo/ast"
	. "github.com/metaleap/atmo/il"
	. "github.com/metaleap/atmo/session"
)

func (me *Repl) initEnsureDefaultDirectives() {
	kd := me.KnownDirectives.ensure
	kd("list ‹kit›", me.DList,
		":list ‹kit/import/path› ── list defs in the specified kit",
		":list _                 ── list all currently known kits",
	)
	kd("info ‹kit› [‹def›]", me.DInfo,
		":info ‹kit/import/path›         ── infos on the specified kit",
		":info ‹kit/import/path› ‹def›   ── infos on the specified def",
		":info * ‹def›                   ── infos on the specified def, having",
		"                                   searched all currently loaded kits",
		":info _ ‹def›                   ── infos on the specified def, having",
		"                                   searched all currently known kits",
	)
	kd("srcs ‹kit› ‹def›", me.DSrcs,
		":srcs ‹kit/import/path› ‹def›   ── sources for the specified def",
		":srcs _ ‹def›                   ── sources for the specified def, having",
		"                                   searched all currently loaded kits",
		":srcs * ‹def›                   ── sources for the specified def, having",
		"                                   searched all currently known kits",
	)
	kd("quit", me.DQuit)
	kd("intro", me.DIntro)
}

func (me *directive) Name() string { return ustr.Until(me.Desc, " ") }

func (me *directives) ensure(desc string, run func(string) bool, help ...string) (ret *directive) {
	if ret = me.By(desc); ret != nil {
		ret.Desc, ret.Run, ret.Help = desc, run, help
	} else {
		this := *me
		idx := len(this)
		this = append(this, directive{Desc: desc, Run: run, Help: help})
		ret = &this[idx]
		*me = this
	}
	return
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
	if name, args = ustr.Trim(name), ustr.Trim(args); name == "" {
		name, args = args, ""
	}
	var found *directive
	if name = ustr.Lo(name); len(name) > 0 {
		if found = me.KnownDirectives.By(name); found != nil {
			if !found.Run(args) {
				me.IO.writeLns("Input `"+args+"` insufficient for demand `:"+found.Name()+"`.", "", "Usage:")
				if len(found.Help) > 0 {
					me.IO.writeLns("")
					me.IO.writeLns(found.Help...)
				} else {
					me.IO.writeLns(found.Desc)
				}
			}
		}
	}
	if found == nil {
		me.IO.writeLns("Unknown demand `:"+name+"` — try: ", "")
		for i := range me.KnownDirectives {
			if !me.KnownDirectives[i].Hidden {
				me.IO.writeLns("    :" + me.KnownDirectives[i].Desc)
			}
		}
		me.IO.writeLns("", "(For usage details on a demand", "with params, run it without any.)")
	}
}

func (me *Repl) DQuit(s string) bool {
	me.run.quit = true
	return true
}

func (me *Repl) DList(what string) bool {
	if what == "" {
		return false
	}
	if what == "_" {
		me.dListKits()
	} else {
		me.dListDefs(what)
	}
	return true
}

func (me *Repl) dListKits() {
	me.IO.writeLns("LIST of kits from current search paths:")
	me.IO.writeLns(ustr.Map(me.Ctx.Dirs.KitsStashes, func(s string) string { return "─── " + s })...)
	kits := me.Ctx.Kits.All
	me.IO.writeLns("", "Found "+ustr.Plu(len(kits), "kit")+":")
	for _, kit := range kits {
		numerrs := len(kit.Errors(nil))
		me.decoAddNotice(false, "", true, "["+ustr.If(kit.WasEverToBeLoaded, "×", "_")+"] "+kit.ImpPath+ustr.If(numerrs == 0, "", " ── "+ustr.Plu(numerrs, "issue")))
	}
	me.IO.writeLns("", "Legend: [_] = unloaded, [×] = loaded or load attempted", "(To see kit details, use `:info ‹kit›`)")
}

func (me *Repl) dListDefs(whatKit string) {
	if kit := me.Ctx.KitByImpPath(whatKit); kit == nil {
		me.IO.writeLns("Unknown kit: `" + whatKit + "`, see known kits via `:list _`.")
	} else {
		me.Ctx.KitEnsureLoaded(kit)
		me.IO.writeLns("LIST of defs in kit:    `"+kit.ImpPath+"`", "           found in:    "+kit.DirPath)
		numdefs := 0
		for _, sf := range kit.SrcFiles {
			if nd, _ := sf.CountTopLevelDefs(true); nd > 0 {
				me.IO.writeLns("", ustr.If(sf.SrcFilePath == "", me.Ctx.Options.Scratchpad.FauxFileNameForErrorMessages, filepath.Base(sf.SrcFilePath))+": "+ustr.Plu(nd, "top-level def"))
				for d := range sf.TopLevel {
					if tld := &sf.TopLevel[d]; !tld.HasErrors() {
						if def := tld.Ast.Def.Orig; def != nil {
							numdefs++
							pos := "(line " + ustr.Int(def.Name.Tokens[0].OffPos(tld.PosOffsetLine(), tld.PosOffsetByte()).Ln1) + ")"
							me.decoAddNotice(false, "", true, ustr.Combine(ustr.If(tld.Ast.Def.IsUnexported, "_", "")+def.Name.Val, " ─── ", pos))
						}
					}
				}
			}
		}
		if me.IO.writeLns("", "Total: "+ustr.Plu(numdefs, "def")+" in "+ustr.Plu(len(kit.SrcFiles), "`*"+SrcFileExt+"` source file")); numdefs > 0 {
			me.IO.writeLns("", "(To see more details, try also:", "`:info "+whatKit+"` or `:info "+whatKit+" ‹def›`.)")
		}
	}
}

func (me *Repl) DIntro(string) bool {
	me.IO.writeLns(Ux.WelcomeMsgLines...)
	return true
}

func (me *Repl) what2KitAndName(what string) (whatKit string, whatName string) {
	whatKit, whatName = ustr.BreakOnFirstOrPref(what, " ")
	whatKit, whatName = ustr.Trim(whatKit), ustr.Trim(whatName)
	return
}

func (me *Repl) DInfo(what string) bool {
	if what == "" {
		return false
	}
	if whatkit, whatname := me.what2KitAndName(what); whatname == "" {
		me.dInfoKit(whatkit)
	} else {
		me.dInfoDef(whatkit, whatname)
	}
	return true
}

func (me *Repl) dInfoKit(whatKit string) {
	if kit := me.Ctx.KitByImpPath(whatKit); kit == nil {
		me.IO.writeLns("Unknown kit: `" + whatKit + "`, see known kits via `:list _`.")
	} else {
		me.Ctx.KitEnsureLoaded(kit)
		me.IO.writeLns("INFO summary on kit:    `"+kit.ImpPath+"`", "           found in:    "+kit.DirPath)
		me.IO.writeLns("", ustr.Plu(len(kit.SrcFiles), "source file")+" in kit `"+whatKit+"`:")
		numlines, numlinesnet, numdefs, numdefsinternal := 0, 0, 0, 0
		for _, sf := range kit.SrcFiles {
			nd, ndi := sf.CountTopLevelDefs(true)
			sloc := sf.CountNetLinesOfCode(true)
			numlines, numlinesnet, numdefs, numdefsinternal = numlines+sf.LastLoad.NumLines, numlinesnet+sloc, numdefs+nd, numdefsinternal+ndi
			me.decoAddNotice(false, "", true, ustr.If(sf.SrcFilePath == "", me.Ctx.Options.Scratchpad.FauxFileNameForErrorMessages, filepath.Base(sf.SrcFilePath)), ustr.Plu(sf.LastLoad.NumLines, "line")+" ("+ustr.Int(sloc)+" sloc), "+ustr.Plu(nd, "top-level def")+", "+ustr.Int(nd-ndi)+" exported")
		}
		me.IO.writeLns("Total:", "    "+ustr.Plu(numlines, "line")+" ("+ustr.Int(numlinesnet)+" sloc), "+ustr.Plu(numdefs, "top-level def")+", "+ustr.Int(numdefs-numdefsinternal)+" exported",
			"    (Counts exclude failed-to-parse defs, if any.)")

		if kiterrs := kit.Errors(nil); len(kiterrs) > 0 {
			me.IO.writeLns("", ustr.Plu(len(kiterrs), "issue")+" in kit `"+whatKit+"`:")
			for i := range kiterrs {
				me.decoMsgNotice(false, kiterrs[i].Error())
			}
		}
		me.IO.writeLns("", "", "(To see kit defs, use `:list "+whatKit+"`)")
	}
}

func (me *Repl) dInfoDef(whatKit string, whatName string) {
	me.withKitDefs(whatKit, whatName, "info", func(kit *Kit, def *IrDef) {
		me.IO.writeLns("TODO")
	})
}

func (me *Repl) DSrcs(what string) bool {
	if whatkit, whatname := me.what2KitAndName(what); whatkit != "" && whatname != "" {
		ctxp := CtxPrint{OneIndentLevel: "    ", Fmt: &PrintFmtPretty{},
			ApplStyle: APPLSTYLE_SVO, BytesWriter: ustd.BytesWriter{Data: make([]byte, 0, 256)}, NoComments: true}

		me.withKitDefs(whatkit, whatname, "srcs", func(kit *Kit, def *IrDef) {
			me.decoAddNotice(false, "", true, ustr.FirstOf(def.OrigTopChunk.SrcFile.SrcFilePath, me.Ctx.Options.Scratchpad.FauxFileNameForErrorMessages))
			ctxp.ApplStyle = def.OrigTopChunk.SrcFile.Options.ApplStyle
			def.OrigTopChunk.Print(&ctxp)
			ctxp.WriteTo(me.IO.Stdout)
			ctxp.Reset()
			if !def.HasErrors() {
				ir2lang := def.Print().(*AstDef)
				me.decoAddNotice(false, "", true, "internal representation:", "")
				ctxp.ApplStyle = APPLSTYLE_VSO
				ctxp.CurTopLevel = ir2lang
				ir2lang.Print(&ctxp)
				ctxp.WriteTo(me.IO.Stdout)
				ctxp.Reset()
				me.IO.writeLns("")
			}
		})
		return true
	}
	return false
}

func (me *Repl) withKitDefs(whatKit string, whatName string, cmdName string, on func(*Kit, *IrDef)) {
	kits := me.Ctx.Kits.All
	var kit *Kit
	if searchloadeds, searchall := (whatKit == "_"), (whatKit == "*"); !(searchall || searchloadeds) {
		if kit = kits.ByImpPath(whatKit); kit == nil && (whatKit == "." || whatKit == "·") {
			for i := range kits {
				if me.Ctx.FauxKitsHas(kits[i].DirPath) {
					kit = kits[i]
					break
				}
			}
		}
	} else {
		var finds Kits
		for _, k := range kits {
			if searchall {
				me.Ctx.KitEnsureLoaded(k)
			}
			if k.HasDefs(whatName) {
				finds = append(finds, k)
			}
		}
		if len(finds) == 1 {
			kit = finds[0]
		} else {
			if len(finds) > 1 {
				me.IO.writeLns("Defs named `" + whatName + "` were found in " + ustr.Int(len(finds)) + " currently-" + ustr.If(searchall, "known", "loaded") + " kits. Pick one:")
				for _, k := range finds {
					me.IO.writeLns("    :" + cmdName + " " + k.ImpPath + " " + whatName)
				}
			} else {
				me.IO.writeLns("No defs named `" + whatName + "` were found in any currently-" + ustr.If(searchall, "known", "loaded") + " kits.")
			}
			return
		}
	}
	if kit == nil {
		me.IO.writeLns("Unknown kit: `" + whatKit + "`, see known kits via `:list _`.")
	} else {
		me.Ctx.KitEnsureLoaded(kit)
		defs := kit.Defs(whatName, true)
		me.IO.writeLns(ustr.Plu(len(defs), "def")+" named `"+whatName+"` found in kit `"+kit.ImpPath+ustr.If(len(defs) > 0, "`:", "`."), "", "")
		for _, def := range defs {
			on(kit, def)
		}
	}
}
