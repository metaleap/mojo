package atmorepl

import (
	"path/filepath"

	"github.com/go-leap/std"
	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang"
	"github.com/metaleap/atmo/session"
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
		":info + ‹def›                   ── infos on the specified def, having",
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
	kd("reload", me.DReload).Hidden = (atmosess.KitsWatchInterval > 0)
}

type directive struct {
	Desc   string
	Help   []string
	Run    func(string) bool
	Hidden bool
}

func (me *directive) Name() string { return ustr.Until(me.Desc, " ") }

type directives []directive

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
				me.IO.writeLns("Input `"+args+"` insufficient for command `:"+found.Name()+"`.", "", "Usage:")
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
		me.IO.writeLns("Unknown command `:"+name+"` — try: ", "")
		for i := range me.KnownDirectives {
			if !me.KnownDirectives[i].Hidden {
				me.IO.writeLns("    :" + me.KnownDirectives[i].Desc)
			}
		}
		me.IO.writeLns("", "(For usage details on a complex", "command, invoke it without args.)")
	}
}

func (me *Repl) DQuit(s string) bool {
	me.run.quit = true
	return true
}

func (me *Repl) DReload(string) bool {
	if nummods := me.Ctx.ReloadModifiedKitsUnlessAlreadyWatching(); nummods == 0 {
		me.IO.writeLns("No relevant modifications ── nothing to (re)load.")
	} else if nummods < 0 {
		me.IO.writeLns("No manual (re)load possible: already checking every " + atmosess.KitsWatchInterval.String() + ".")
	}
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
	me.IO.writeLns(ustr.Map(me.Ctx.Dirs.Kits, func(s string) string { return "─── " + s })...)
	me.Ctx.WithKnownKits(func(kits atmosess.Kits) {
		me.IO.writeLns("", "Found "+ustr.Plu(len(kits), "kit")+":")
		for _, kit := range kits {
			numerrs := len(kit.Errors())
			me.decoAddNotice(false, "", true, "["+ustr.If(kit.WasEverToBeLoaded, "×", "_")+"] "+kit.ImpPath+ustr.If(numerrs == 0, "", " ── "+ustr.Plu(numerrs, "issue")))
		}
	})
	me.IO.writeLns("", "Legend: [_] = unloaded, [×] = loaded or load attempted", "(To see kit details, use `:info ‹kit›`.)")
}

func (me *Repl) dListDefs(whatKit string) {
	me.Ctx.WithKit(whatKit, func(kit *atmosess.Kit) {
		if kit == nil {
			me.IO.writeLns("Unknown kit: `" + whatKit + "`, see known kits via `:list _`.")
		} else {
			me.kitEnsureLoaded(kit)
			me.IO.writeLns("LIST of defs in kit:    `"+kit.ImpPath+"`", "           found in:    "+kit.DirPath)
			kitsrcfiles, numdefs := kit.SrcFiles(), 0
			for i := range kitsrcfiles {
				sf := &kitsrcfiles[i]
				nd, _ := sf.CountTopLevelDefs(true)
				me.IO.writeLns("", filepath.Base(sf.SrcFilePath)+": "+ustr.Plu(nd, "top-level def"))
				for d := range sf.TopLevel {
					if tld := &sf.TopLevel[d]; !tld.HasErrors() {
						if def := tld.Ast.Def.Orig; def != nil {
							numdefs++
							pos := ustr.If(!def.Name.Tokens[0].Meta.Position.IsValid(), "",
								"(line "+ustr.Int(def.Name.Tokens[0].Meta.Position.Line)+")")
							me.decoAddNotice(false, "", true, ustr.Combine(def.Name.Val, " ─── ", pos))
						}
					}
				}
			}
			if me.IO.writeLns("", "Total: "+ustr.Plu(numdefs, "def")+" in "+ustr.Plu(len(kitsrcfiles), "`*"+atmo.SrcFileExt+"` source file")); numdefs > 0 {
				me.IO.writeLns("", "(To see more details, try also:", "`:info "+whatKit+"` or `:info "+whatKit+" ‹def›`.)")
			}
		}
	})
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
	me.Ctx.WithKit(whatKit, func(kit *atmosess.Kit) {
		if kit == nil {
			me.IO.writeLns("Unknown kit: `" + whatKit + "`, see known kits via `:list _`.")
		} else {
			me.kitEnsureLoaded(kit)
			me.IO.writeLns("INFO summary on kit:    `"+kit.ImpPath+"`", "           found in:    "+kit.DirPath)
			kitsrcfiles := kit.SrcFiles()
			me.IO.writeLns("", ustr.Plu(len(kitsrcfiles), "source file")+" in kit `"+whatKit+"`:")
			numlines, numlinesnet, numdefs, numdefsinternal := 0, 0, 0, 0
			for i := range kitsrcfiles {
				sf := &kitsrcfiles[i]
				nd, ndi := sf.CountTopLevelDefs(true)
				sloc := sf.CountNetLinesOfCode(true)
				numlines, numlinesnet, numdefs, numdefsinternal = numlines+sf.LastLoad.NumLines, numlinesnet+sloc, numdefs+nd, numdefsinternal+ndi
				me.decoAddNotice(false, "", true, filepath.Base(sf.SrcFilePath), ustr.Plu(sf.LastLoad.NumLines, "line")+" ("+ustr.Int(sloc)+" sloc), "+ustr.Plu(nd, "top-level def")+", "+ustr.Int(nd-ndi)+" exported")
			}
			me.IO.writeLns("Total:", "    "+ustr.Plu(numlines, "line")+" ("+ustr.Int(numlinesnet)+" sloc), "+ustr.Plu(numdefs, "top-level def")+", "+ustr.Int(numdefs-numdefsinternal)+" exported",
				"    (Counts exclude failed-to-parse defs, if any.)")

			if kiterrs := kit.Errors(); len(kiterrs) > 0 {
				me.IO.writeLns("", ustr.Plu(len(kiterrs), "issue")+" in kit `"+whatKit+"`:")
				for i := range kiterrs {
					me.decoMsgNotice(false, kiterrs[i].Error())
				}
			}
			me.IO.writeLns("", "", "(To see kit defs, use `:list "+whatKit+"`.)")
		}
	})
}

func (me *Repl) dInfoDef(whatKit string, whatName string) {
	me.IO.writeLns("Info on name: " + whatName + " in " + whatKit)
}

func (me *Repl) DSrcs(what string) bool {
	if whatkit, whatname := me.what2KitAndName(what); whatkit != "" && whatname != "" {
		me.Ctx.WithKnownKits(func(kits atmosess.Kits) {
			var kit *atmosess.Kit
			if searchloaded, searchall := whatkit == "_", whatkit == "*"; !(searchall || searchloaded) {
				kit = kits.ByImpPath(whatkit)
				me.kitEnsureLoaded(kit)
			} else {
				var finds []*atmosess.Kit
				for i := range kits {
					k := &kits[i]
					if searchall {
						me.kitEnsureLoaded(k)
					}
					if k.HasDefs(whatname) {
						finds = append(finds, k)
					}
				}
				if len(finds) == 1 {
					kit = finds[0]
				} else {
					if len(finds) > 1 {
						me.IO.writeLns("Defs named `" + whatname + "` were found in " + ustr.Int(len(finds)) + " currently-" + ustr.If(searchall, "known", "loaded") + " kits. Pick one:")
						for _, k := range finds {
							me.IO.writeLns("    :srcs " + k.ImpPath + " " + whatname)
						}
					} else {
						me.IO.writeLns("No defs named `" + whatname + "` were found in any currently-" + ustr.If(searchall, "known", "loaded") + " kits.")
					}
					return
				}
			}
			if kit == nil {
				me.IO.writeLns("Unknown kit: `" + whatkit + "`, see known kits via `:list _`.")
			} else {
				defs, ctxp := kit.Defs(whatname), atmolang.CtxPrint{OneIndentLevel: "    ", Fmt: &atmolang.PrintFmtPretty{},
					ApplStyle: atmolang.APPLSTYLE_SVO, BytesWriter: ustd.BytesWriter{Data: make([]byte, 0, 256)}, NoComments: true}
				me.IO.writeLns(ustr.Plu(len(defs), "def")+" named `"+whatname+"` found in kit `"+kit.ImpPath+ustr.If(len(defs) > 0, "`:", "`."), "", "")
				for _, def := range defs {
					me.decoAddNotice(false, "", true, def.TopLevel.SrcFile.SrcFilePath, "")
					def.TopLevel.Print(&ctxp)
					ctxp.WriteTo(me.IO.Stdout)
					ctxp.Reset()

					ctxp.CurTopLevel = nil
					me.decoAddNotice(false, "", true, "internal representation:", "")
					def.Print().(*atmolang.AstDef).Print(&ctxp)
					ctxp.WriteTo(me.IO.Stdout)
					ctxp.Reset()
				}
			}

		})
		return true
	}
	return false
}
