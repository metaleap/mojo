package atmosess

import (
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/il"
	"github.com/metaleap/atmo/lang"
)

func (me *Ctx) Eval(kit *Kit, maybeTopDefId string, src string) (ret IPreduced, errs atmo.Errors) {
	kit.SrcFiles.ByFilePath("")
	expr, err := atmolang.LexAndParseExpr(me.Options.Eval.FauxFileNameForErrorMessages, []byte(src))

	if err != nil {
		errs = append(errs, err)
	} else {
		expril, errsil := atmoil.ExprFrom(expr)
		if errs.Add(errsil); len(errs) == 0 && expril != nil {
			ret = me.PreduceExpr(kit, maybeTopDefId, expril)
		}
	}
	return
}
