package atmosess

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/metaleap/atmo/il"
	"github.com/metaleap/atmo/lang"
)

func (me *Ctx) Eval(kit *Kit, maybeTopDefId string, src string) (ret IPreduced, errs []error) {
	spfile := kit.SrcFiles.EnsureScratchPadFile()
	origsrc := spfile.Options.TmpAltSrc
	guessisdef, _, _, err := atmolang.LexAndGuess("", []byte(src))
	if err != nil {
		errs = append(errs, err)
	} else {
		var tmpdefname string
		if !guessisdef {
			tmpdefname = "_" + strconv.FormatInt(time.Now().UnixNano(), 16) + strconv.FormatInt(rand.Int63(), 16)
			src = tmpdefname + " :=\n " + src
		}

		spfile.Options.TmpAltSrc = append(spfile.Options.TmpAltSrc, '\n', '\n')
		spfile.Options.TmpAltSrc = append(spfile.Options.TmpAltSrc, src...)
		me.catchUpOnFileMods(kit)

		restoreorigsrc := !guessisdef
		if spfile.HasErrors() {
			errs, restoreorigsrc = spfile.Errors(), true
		} else {
			tlc := spfile.TopLevelChunkAt(len(origsrc))
			ret = me.PreduceExpr(kit, tlc.Id(), atmoil.Build.IdentName(tlc.Ast.Def.Orig.Name.Val))
		}

		if restoreorigsrc {
			spfile.Options.TmpAltSrc = origsrc
			me.catchUpOnFileMods(kit)
		}
	}

	return
}
