package atmosess

import (
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/il"
	"github.com/metaleap/atmo/lang"
)

func (me *Ctx) Eval(kit *Kit, src string) (ret IPreduced, errs atmo.Errors) {
	expr, err := atmolang.LexAndParseExpr("‹repl›", []byte(src))
	if err != nil {
		errs = append(errs, err)
	} else {
		expril, errsil := atmoil.ExprFrom(expr)
		if errs.Add(errsil); len(errs) == 0 && expril != nil {
			ret, _ = errs.AddVia(me.PreduceExpr(kit, expril)).(IPreduced)
		}
	}
	return
}
