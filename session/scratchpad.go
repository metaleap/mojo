package atmosess

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/il"
	"github.com/metaleap/atmo/lang"
)

func (me *Ctx) Eval(kit *Kit, maybeTopDefId string, src string) (ret IPreduced, errs atmo.Errors) {
	spfile := kit.SrcFiles.EnsureScratchPadFile()
	origsrc := spfile.Options.TmpAltSrc
	isdef, _, toks, err := atmolang.LexAndGuess("", []byte(src))
	if err != nil {
		return nil, errs.AddFrom(err)
	}

	var defname string
	if !isdef {
		defname = strconv.FormatInt(time.Now().UnixNano(), 16) + strconv.FormatInt(rand.Int63(), 16)
		src = "_" + defname + " :=\n " + src
	} else if def, e := atmolang.LexAndParseDefOrExpr(isdef, toks); e != nil {
		return nil, errs.AddFrom(e)
	} else {
		defname = def.(*atmolang.AstDef).Name.Val
		existingdefids := kit.lookups.tlDefIDsByName[defname]
		for _, olddefid := range existingdefids {
			if tlc := kit.SrcFiles.TopLevelChunkByDefId(olddefid); tlc != nil {
				if tlc.SrcFile.SrcFilePath != "" {
					errs.AddSess(ErrSess_EvalDefNameExists, "", "name `"+defname+"` already declared in "+tlc.SrcFile.SrcFilePath+":"+strconv.Itoa(1+tlc.PosOffsetLine())+":1 ─ try again with another name")
					return
				}
			}
		}
	}

	spfile.Options.TmpAltSrc = append(spfile.Options.TmpAltSrc, '\n', '\n')
	spfile.Options.TmpAltSrc = append(spfile.Options.TmpAltSrc, src...)
	me.catchUpOnFileMods(kit)

	restoreorigsrc := !isdef
	if spfile.HasErrors() { // refers ONLY to lex/parse errors
		return nil, spfile.Errors()
	} else {
		for _, tld := range kit.topLevelDefs {
			if tld.OrigTopLevelChunk.SrcFile.SrcFilePath == "" && len(tld.Errs.Stage0Init) > 0 {
				println("AHA", tld.Name.Val, "EHE", tld.OrigDef.Name.Val, "UHU", tld.OrigTopLevelChunk.Ast.Def.NameIfErr, "OHO", tld.Errs.Stage0Init[0].Error())
			}
		}
		defids := kit.lookups.tlDefIDsByName[defname]
		if len(defids) != 1 {
			panic(len(defids))
		}
		def := kit.lookups.tlDefsByID[defids[0]]

		identexpr := atmoil.Build.IdentName(defname)
		identexpr.Anns.Candidates = []atmoil.INode{def}
		ret = me.PreduceExpr(kit, defids[0], identexpr)
	}

	if restoreorigsrc {
		spfile.Options.TmpAltSrc = origsrc
		me.catchUpOnFileMods(kit)
	}

	return
}
