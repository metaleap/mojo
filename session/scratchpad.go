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

	var restoreorigsrc bool
	defer func() {
		if restoreorigsrc {
			spfile.Options.TmpAltSrc = origsrc
			me.catchUpOnFileMods(kit)
		}
	}()

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

	restoreorigsrc = !isdef
	spfile.Options.TmpAltSrc = append(spfile.Options.TmpAltSrc, '\n', '\n')
	spfile.Options.TmpAltSrc = append(spfile.Options.TmpAltSrc, src...)
	me.catchUpOnFileMods(kit)

	if spfile.HasErrors() { // refers ONLY to lex/parse errors
		restoreorigsrc = true
		return nil, spfile.Errors()
	} else {
		defs := kit.topLevelDefs.ByName(defname)
		if len(defs) != 1 { // shouldn't happen based on above safeguards
			restoreorigsrc = true
			panic(len(defs))
		}
		if errs = defs[0].Errors(); len(errs) > 0 {
			restoreorigsrc = true
			return
		}
		identexpr := atmoil.Build.IdentName(defname)
		identexpr.Anns.Candidates = []atmoil.INode{defs[0]}
		ret = me.PreduceExpr(kit, defs[0].Id, identexpr)
	}

	return
}
