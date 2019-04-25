package atmorepl

import (
	"path/filepath"

	"github.com/go-leap/str"
	"github.com/metaleap/atmo/load"
)

func (me *Repl) initEnsureDefaultDirectives() {
	kd := me.KnownDirectives.ensure
	kd("list <packs | defs | \"pack/import/path\">", me.DList)
	kd("info [\"pack/import/path\"] [name]", me.DInfo)
	kd("quit", me.DQuit)
	if atmoload.PacksWatchInterval == 0 {
		kd("reload", me.DReload) //\n      (reloads modified code in known packs)", me.DReload)
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
		if n, a := ustr.BreakOnFirstOrSuff(name, "\""); n != "" && args == "" {
			me.runDirective(n, "\""+a)
			return
		}

		me.IO.writeLns("unknown directive `:" + name + "` — try: ")
		for i := range me.KnownDirectives {
			me.IO.writeLns("   :" + me.KnownDirectives[i].Desc)
		}
	}
}

func (me *Repl) DQuit(s string) bool {
	me.run.quit = true
	return true
}

func (me *Repl) DReload(string) bool {
	me.Ctx.ReloadModifiedPacksUnlessAlreadyWatching()
	return true
}
func (me *Repl) DList(what string) bool {
	if what == "" || ustr.Pref("packs", what) {
		me.dListPacks()
	} else if ustr.Pref("defs", what) {
		me.dListDefs("")
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
			me.decoAddNotice(false, "", true, "\""+pack.ImpPath+"\""+ustr.If(numerrs == 0, "", " ── "+ustr.Plu(numerrs, "error")))
		}
	})
	me.IO.writeLns("", "(to see pack details, use `:info \"<pack/import/path>\"`)")
}

func (me *Repl) dListDefs(whatPack string) {
	if whatPack != "" && whatPack[0] == '"' && len(whatPack) > 1 {
		whatPack = ustr.If(whatPack[len(whatPack)-1] != '"', whatPack[1:], whatPack[1:len(whatPack)-1])
	}
	if whatPack == "" {
		me.IO.writeLns("TODO: list ALL defs")

	} else {
		me.Ctx.WithPack(whatPack, true, func(pack *atmoload.Pack) {
			if pack == nil {
				me.IO.writeLns("unknown pack: `" + whatPack + "`, see known packs via `:list packs`")
			} else {
				me.IO.writeLns("", "\""+pack.ImpPath+"\"", "    "+pack.DirPath)
				packsrcfiles := pack.SrcFiles()
				for i := range packsrcfiles {
					sf := &packsrcfiles[i]
					nd, _ := sf.CountTopLevelDefs()
					me.IO.writeLns("", filepath.Base(sf.SrcFilePath)+": "+ustr.Plu(nd, "top-level def"))
					for d := range sf.TopLevel {
						if def := sf.TopLevel[d].Ast.Def.Orig; def != nil {
							pos := ustr.If(!def.Name.Tokens[0].Meta.Position.IsValid(), "",
								"(line "+ustr.Int(def.Name.Tokens[0].Meta.Position.Line)+")")
							me.decoAddNotice(false, "", true, ustr.Combine(def.Name.Val, " ─── ", pos))
						}
					}
				}
			}
		})
	}
}

func (me *Repl) DInfo(what string) bool {
	if what == "" {
		me.dInfo()
	} else {
		var whatpack, whatname string
		if whatname = what; what[0] == '"' {
			whatpack, whatname = ustr.BreakOnFirstOrPref(what[1:], "\"")
		}
		if whatpack, whatname = ustr.Trim(whatpack), ustr.Trim(whatname); whatname == "" {
			me.dInfoPack(whatpack)
		} else {
			me.dInfoDef(whatpack, whatname)
		}
	}
	return true
}

func (me *Repl) dInfo() {
	me.IO.writeLns(me.run.welcomeMsgLines...)
}

func (me *Repl) dInfoPack(whatPack string) {
	me.Ctx.WithPack(whatPack, true, func(pack *atmoload.Pack) {
		if pack == nil {
			me.IO.writeLns("unknown pack: `" + whatPack + "`, see known packs via `:list packs`")
		} else {
			me.IO.writeLns("\""+pack.ImpPath+"\"", "    "+pack.DirPath)
			packsrcfiles := pack.SrcFiles()
			me.IO.writeLns("", ustr.Plu(len(packsrcfiles), "source file")+" in pack \""+whatPack+"\":")
			numlines, numlinesnet, numdefs, numdefsinternal := 0, 0, 0, 0
			for i := range packsrcfiles {
				sf := &packsrcfiles[i]
				nd, ndi := sf.CountTopLevelDefs()
				sloc := sf.CountNetLinesOfCode()
				numlines, numlinesnet, numdefs, numdefsinternal = numlines+sf.LastLoad.NumLines, numlinesnet+sloc, numdefs+nd, numdefsinternal+ndi
				me.decoAddNotice(false, "", true, filepath.Base(sf.SrcFilePath), ustr.Plu(sf.LastLoad.NumLines, "line")+" ("+ustr.Int(sloc)+" sloc), "+ustr.Plu(nd, "top-level def")+", "+ustr.Int(nd-ndi)+" exported")
			}
			me.IO.writeLns("Total: "+ustr.Plu(numlines, "line")+" ("+ustr.Int(numlinesnet)+" sloc), "+ustr.Plu(numdefs, "top-level def")+", "+ustr.Int(numdefs-numdefsinternal)+" exported",
				"    (counts exclude failed-to-parse code portions, if any)")

			if packerrs := pack.Errs(); len(packerrs) > 0 {
				me.IO.writeLns("", ustr.Plu(len(packerrs), "issue")+" in pack \""+whatPack+"\":")
				for i := range packerrs {
					me.decoMsgNotice(packerrs[i].Error())
				}
			}
			me.IO.writeLns("", "", "(to see pack defs, use `:list \""+whatPack+"\"`)")
		}
	})
}

func (me *Repl) dInfoDef(whatPack string, whatName string) {
	me.IO.writeLns("Info on name: " + whatName + " in \"" + whatPack + "\"")
}
