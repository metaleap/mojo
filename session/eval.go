package atmosess

import (
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang"
	"github.com/metaleap/atmo/lang/irfun"
)

func (me *Ctx) Eval(kit *Kit, src string) (str string, errs atmo.Errors) {
	expr, err := atmolang.LexAndParseExpr("‹repl›", []byte(src))
	if err != nil {
		errs = append(errs, err)
	} else {
		irx, errsir := atmolang_irfun.ExprFrom(expr)
		if errs.Add(errsir); len(errs) == 0 && irx != nil {
			kit.lookups.namesInScopeAll.RepopulateAstDefsAndIdentsFor(irx)
			// if retdesc, err := me.inferFactsForExpr(kit, irx); err != nil {
			// 	errs = append(errs, err)
			// } else {
			// 	str = retdesc.String()
			// }
		}
	}
	return
}
