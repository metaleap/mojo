package atmosess

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/il"
	"github.com/metaleap/atmo/lang"
)

func (me *Ctx) ScratchpadEntry(kit *Kit, maybeTopDefId string, src string) (ret IPreduced, errs atmo.Errors) {
	spfile := kit.SrcFiles.EnsureScratchpadFile()
	origsrc, tmpaltsrc := spfile.Options.TmpAltSrc, spfile.Options.TmpAltSrc
	var restoreorigsrc bool
	defer func() {
		if restoreorigsrc {
			spfile.Options.TmpAltSrc = origsrc
			me.catchUpOnFileMods(kit)
		}
	}()

	isdef, _, toks, err := atmolang.LexAndGuess("", []byte(src))
	if err != nil {
		errs.Add(err)
		return
	}

	var defname string
	if !isdef { // just an expression: add temp def `_randomname := ‹input›` then eval that name
		defname = strconv.FormatInt(time.Now().UnixNano(), 16) + strconv.FormatInt(rand.Int63(), 16)
		src = "_" + defname + " :=\n " + src
	} else if defnode, e := atmolang.LexAndParseDefOrExpr(isdef, toks); e != nil {
		errs.Add(e)
		return
	} else { // a full def to add to (or update in) the scratch-pad
		def := defnode.(*atmolang.AstDef)
		defname = def.Name.Val
		var alreadyinscratchpad, alreadyinsrcfile *atmolang.SrcTopChunk
		for _, t := range kit.topLevelDefs {
			if t.OrigTopLevelChunk != nil && (t.Name.Val == defname || (t.OrigDef != nil && t.OrigDef.Name.Val == defname)) {
				if t.OrigTopLevelChunk.SrcFile.SrcFilePath == "" {
					alreadyinscratchpad = t.OrigTopLevelChunk
				} else {
					alreadyinsrcfile = t.OrigTopLevelChunk
					break
				}
			}
		}
		if alreadyinsrcfile != nil {
			errs.AddSess(ErrSess_EvalDefNameExists, me.Options.Eval.FauxFileNameForErrorMessages, "name `"+defname+"` already declared in "+alreadyinsrcfile.SrcFile.SrcFilePath+":"+strconv.Itoa(1+alreadyinsrcfile.PosOffsetLine())+":1 ─ try again with another name")
			return
		} else if alreadyinscratchpad != nil { // overwrite prev def by slicing out the old one
			boff := alreadyinscratchpad.PosOffsetByte()
			pref, suff := tmpaltsrc[:boff], tmpaltsrc[boff+len(alreadyinscratchpad.Src):]
			tmpaltsrc = append(pref, suff...)
		}
	}

	restoreorigsrc = !isdef
	spfile.Options.TmpAltSrc = append(append([]byte(src), '\n', '\n'), tmpaltsrc...)
	me.catchUpOnFileMods(kit)

	if spfile.HasErrors() { // refers ONLY to lex/parse errors
		restoreorigsrc, errs = true, spfile.Errors()
		return
	} else {
		defs := kit.topLevelDefs.ByName(defname)
		if len(defs) != 1 { // shouldn't happen based on above safeguards, else we'll have to look carefully where we missed some clean-up/sanity-check
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
