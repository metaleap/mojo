package atmorepl

import (
	"path/filepath"

	"github.com/go-leap/str"
	"github.com/metaleap/atmo/load"
)

func (me *Repl) initEnsureDefaultDirectives() {
	kd := me.KnownDirectives.ensure
	kd("list ‹pack›", me.DList,
		":list ‹pack/import/path›    ── list defs in the specified pack",
		":list _                     ── list all currently known packs",
	)
	kd("info ‹pack› [‹name›]", me.DInfo,
		":info ‹pack/import/path›        ── infos on the specified pack",
		":info ‹pack/import/path› ‹def›  ── infos on the specified def",
		":info _ ‹def›                   ── infos on the specified def,",
		"                                   having searched all currently known packs",
	)
	kd("srcs ‹pack› ‹name›", me.DSrcs,
		":srcs ‹pack/import/path› ‹def›  ── sources for the specified def",
		":srcs _ ‹def›                   ── sources for the specified def,",
		"                                   having searched all currently known packs",
	)
	kd("quit", me.DQuit)
	kd("intro", me.DIntro)
	if atmoload.PacksWatchInterval == 0 {
		kd("reload", me.DReload)
	}
}

type directive struct {
	Desc string
	Help []string
	Run  func(string) bool
}

func (me *directive) Name() string { return ustr.Until(me.Desc, " ") }

type directives []directive

func (me *directives) ensure(desc string, run func(string) bool, help ...string) {
	if found := me.By(desc); found != nil {
		found.Desc, found.Run = desc, run
	} else {
		*me = append(*me, directive{Desc: desc, Run: run, Help: help})
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
	if name = ustr.Lo(name); len(name) > 0 {
		if found = me.KnownDirectives.By(name); found != nil {
			if args = ustr.Trim(args); !found.Run(args) {
				if args != "" {
					me.IO.writeLns("Input `"+args+"` refused by command `:"+found.Name()+"`.", "")
				} else {
					me.IO.writeLns("Input needed for command `:"+found.Name()+"`.", "")
				}
				me.IO.writeLns("Usage:")
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
		me.IO.writeLns("unknown command `:"+name+"` — try: ", "")
		for i := range me.KnownDirectives {
			me.IO.writeLns("    :" + me.KnownDirectives[i].Desc)
		}
		me.IO.writeLns("", "(further help for a given complex command", "will display when invoking it without args)")
	}
}

func (me *Repl) DQuit(s string) bool {
	me.run.quit = true
	return true
}

func (me *Repl) DReload(string) bool {
	if nummods := me.Ctx.ReloadModifiedPacksUnlessAlreadyWatching(); nummods == 0 {
		me.IO.writeLns("No relevant modifications noted ── nothing to (re)load.")
	}
	return true
}

func (me *Repl) DList(what string) bool {
	if what == "" {
		return false
	}
	if what == "_" {
		me.dListPacks()
	} else {
		me.dListDefs(what)
	}
	return true
}

func (me *Repl) dListPacks() {
	me.IO.writeLns("From current search paths:")
	me.IO.writeLns(ustr.Map(me.Ctx.Dirs.Packs, func(s string) string { return "─── " + s })...)
	me.Ctx.WithKnownPacks(func(packs []atmoload.Pack) {
		me.IO.writeLns("", "found "+ustr.Plu(len(packs), "pack")+":")
		for _, pack := range packs {
			numerrs := len(pack.Errs())
			me.decoAddNotice(false, "", true, pack.ImpPath+ustr.If(numerrs == 0, "", " ── "+ustr.Plu(numerrs, "error")))
		}
	})
	me.IO.writeLns("", "(to see pack details, use `:info ‹pack›`)")
}

func (me *Repl) dListDefs(whatPack string) {
	me.Ctx.WithPack(whatPack, true, func(pack *atmoload.Pack) {
		if pack == nil {
			me.IO.writeLns("unknown pack: `" + whatPack + "`, see known packs via `:list _`")
		} else {
			me.IO.writeLns("", pack.ImpPath, "    "+pack.DirPath)
			packsrcfiles, numdefs := pack.SrcFiles(), 0
			for i := range packsrcfiles {
				sf := &packsrcfiles[i]
				nd, _ := sf.CountTopLevelDefs()
				me.IO.writeLns("", filepath.Base(sf.SrcFilePath)+": "+ustr.Plu(nd, "top-level def"))
				for d := range sf.TopLevel {
					if def := sf.TopLevel[d].Ast.Def.Orig; def != nil {
						numdefs++
						pos := ustr.If(!def.Name.Tokens[0].Meta.Position.IsValid(), "",
							"(line "+ustr.Int(def.Name.Tokens[0].Meta.Position.Line)+")")
						me.decoAddNotice(false, "", true, ustr.Combine(def.Name.Val, " ─── ", pos))
					}
				}
			}
			if me.IO.writeLns("", "Total: "+ustr.Plu(numdefs, "def")+" in "+ustr.Plu(len(packsrcfiles), ".at source file")); numdefs > 0 {
				me.IO.writeLns("", "(To see more details, try also:", "`:info "+whatPack+"` or `:info "+whatPack+" ‹def›`.)")
			}
		}
	})
}

func (me *Repl) DIntro(string) bool {
	me.IO.writeLns(me.run.welcomeMsgLines...)
	return true
}

func (me *Repl) what2PackAndName(what string) (whatPack string, whatName string) {
	whatPack, whatName = ustr.BreakOnFirstOrPref(what, " ")
	whatPack, whatName = ustr.Trim(whatPack), ustr.Trim(whatName)
	return
}

func (me *Repl) DInfo(what string) bool {
	if what == "" {
		return false
	}
	if whatpack, whatname := me.what2PackAndName(what); whatname == "" {
		me.dInfoPack(whatpack)
	} else {
		me.dInfoDef(whatpack, whatname)
	}
	return true
}

func (me *Repl) dInfoPack(whatPack string) {
	me.Ctx.WithPack(whatPack, true, func(pack *atmoload.Pack) {
		if pack == nil {
			me.IO.writeLns("unknown pack: `" + whatPack + "`, see known packs via `:list _`")
		} else {
			me.IO.writeLns(pack.ImpPath, "    "+pack.DirPath)
			packsrcfiles := pack.SrcFiles()
			me.IO.writeLns("", ustr.Plu(len(packsrcfiles), "source file")+" in pack "+whatPack+":")
			numlines, numlinesnet, numdefs, numdefsinternal := 0, 0, 0, 0
			for i := range packsrcfiles {
				sf := &packsrcfiles[i]
				nd, ndi := sf.CountTopLevelDefs()
				sloc := sf.CountNetLinesOfCode()
				numlines, numlinesnet, numdefs, numdefsinternal = numlines+sf.LastLoad.NumLines, numlinesnet+sloc, numdefs+nd, numdefsinternal+ndi
				me.decoAddNotice(false, "", true, filepath.Base(sf.SrcFilePath), ustr.Plu(sf.LastLoad.NumLines, "line")+" ("+ustr.Int(sloc)+" sloc), "+ustr.Plu(nd, "top-level def")+", "+ustr.Int(nd-ndi)+" exported")
			}
			me.IO.writeLns("Total:", "    "+ustr.Plu(numlines, "line")+" ("+ustr.Int(numlinesnet)+" sloc), "+ustr.Plu(numdefs, "top-level def")+", "+ustr.Int(numdefs-numdefsinternal)+" exported",
				"    (counts exclude failed-to-parse code portions, if any)")

			if packerrs := pack.Errs(); len(packerrs) > 0 {
				me.IO.writeLns("", ustr.Plu(len(packerrs), "issue")+" in pack "+whatPack+":")
				for i := range packerrs {
					me.decoMsgNotice(packerrs[i].Error())
				}
			}
			me.IO.writeLns("", "", "(to see pack defs, use `:list "+whatPack+"`)")
		}
	})
}

func (me *Repl) dInfoDef(whatPack string, whatName string) {
	me.IO.writeLns("Info on name: " + whatName + " in " + whatPack)
}

func (me *Repl) DSrcs(what string) bool {
	if whatpack, whatname := me.what2PackAndName(what); whatpack != "" && whatname != "" {
		me.Ctx.WithPack(whatpack, true, func(pack *atmoload.Pack) {
			if pack == nil {
				me.IO.writeLns("unknown pack: `" + whatpack + "`, see known packs via `:list _`")
			} else {
				defs := pack.Defs(whatname)
				me.IO.writeLns(ustr.Plu(len(defs), "def") + " found")
			}
		})
		return true
	}
	return false
}
