package atmosess

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/il"
	"github.com/metaleap/atmo/lang"
)

func (me *Kit) ensureScratchpadFile() (pretendFile *atmolang.AstFile) {
	if pretendFile = me.SrcFiles.ByFilePath(""); pretendFile == nil {
		pretendFile = &atmolang.AstFile{}
		me.SrcFiles = append(me.SrcFiles, pretendFile)
		me.ScratchpadClear() // must be after the append
	}
	return
}

func (me *Kit) ScratchpadView() []byte {
	return me.ensureScratchpadFile().Options.TmpAltSrc
}

func (me *Kit) ScratchpadClear() {
	me.ensureScratchpadFile().Options.TmpAltSrc = make([]byte, 0, 128) // what matters is that it mustn't be `nil` for scratchpad purposes
}

func (me *Ctx) ScratchpadEntry(kit *Kit, maybeTopDefId string, src string) (ret atmoil.IPreduced, errs atmo.Errors) {
	if src = ustr.Trim(src); len(src) == 0 {
		return
	}
	spfile, bgmsgs := kit.ensureScratchpadFile(), me.Options.BgMsgs.IncludeLiveKitsErrs
	origsrc, tmpaltsrc := spfile.Options.TmpAltSrc, spfile.Options.TmpAltSrc
	isdef, _, toks, err := atmolang.LexAndGuess("", []byte(src))
	var restoreorigsrc bool
	var prefixlength int
	defer func() {
		if !isdef {
			for _, e := range errs {
				if pos := e.Pos(); pos != nil && (pos.FilePath == "" || pos.FilePath == me.Options.Scratchpad.FauxFileNameForErrorMessages) {
					e.UpdatePosOffsets(nil)
					pos.Ln1, pos.Col1, pos.Off0 = pos.Ln1-1, pos.Col1-1, pos.Off0-prefixlength
				}
			}
		}
		if restoreorigsrc {
			spfile.Options.TmpAltSrc = origsrc
			me.catchUpOnFileMods(kit)
		}
		me.Options.BgMsgs.IncludeLiveKitsErrs = bgmsgs
	}()

	if err != nil {
		errs.Add(err)
		return
	}
	src += "\n\n"

	var defname string
	if !isdef { // entry is an expr: add temp def `eval‹RandomNoise› := ‹input›` then eval that name
		defname = "eval" + strconv.FormatInt(time.Now().UnixNano(), 16) + strconv.FormatInt(rand.Int63(), 16)
		prefix := "_" + defname + " :=\n "
		src, prefixlength = prefix+src, len(prefix)
	} else if defnode, e := atmolang.LexAndParseDefOrExpr(isdef, toks); e != nil {
		errs.Add(e)
		return
	} else { // a full def to add to (or update in) the scratch-pad
		def := defnode.(*atmolang.AstDef)
		defname = def.Name.Val
		var alreadyinscratchpad *atmolang.SrcTopChunk
		for _, t := range kit.topLevelDefs {
			if t.OrigTopChunk != nil && (t.Name.Val == defname || (t.OrigDef != nil && t.OrigDef.Name.Val == defname)) {
				if t.OrigTopChunk.SrcFile.SrcFilePath == "" {
					alreadyinscratchpad = t.OrigTopChunk
					break
				}
			}
		}
		if alreadyinscratchpad != nil { // overwrite prev def by slicing out the old one
			boff := alreadyinscratchpad.PosOffsetByte()
			pref, suff := tmpaltsrc[:boff], tmpaltsrc[boff+len(alreadyinscratchpad.Src):]
			tmpaltsrc = append(append(make([]byte, 0, len(pref)+len(suff)), pref...), suff...)
			if ident, _ := def.Body.(*atmolang.AstIdent); ident != nil && ident.IsPlaceholder() {
				src = ""
			}
		}
	}

	restoreorigsrc = !isdef
	spfile.Options.TmpAltSrc = append([]byte(src), tmpaltsrc...)
	me.catchUpOnFileMods(kit)
	me.Options.BgMsgs.IncludeLiveKitsErrs = false

	if spfile.HasErrors() { // refers ONLY to lex/parse errors
		restoreorigsrc, errs = true, spfile.Errors()
		return
	} else if src == "" {
		me.bgMsg(false, "def removed from scratchpad: "+defname)
	} else {
		defs := kit.topLevelDefs.ByName(defname, spfile)
		if len(defs) != 1 { // shouldn't happen based on above safeguards, else we'll have to look carefully where we missed some clean-up/sanity-check
			restoreorigsrc = true
			panic(len(defs))
		}
		if errs = defs[0].Errors(); len(errs) > 0 {
			restoreorigsrc = true
			return
		}
		if len(tmpaltsrc) != len(origsrc) {
			me.bgMsg(false, "def modified in scratchpad: "+defname)
		} else if isdef {
			me.bgMsg(false, "def added to scratchpad: "+defname)
		}
		identexpr := atmoil.Build.IdentName(defname)
		identexpr.Anns.Candidates = []atmoil.INode{defs[0]}
		ret, errs = me.PreduceExpr(kit, identexpr)
	}

	return
}
