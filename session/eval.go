package atmosess

import (
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/il"
	"github.com/metaleap/atmo/lang"
)

func (me *Ctx) Eval(kit *Kit, src string) (str string, errs atmo.Errors) {
	return "", atmo.Errors{atmo.ErrTodo(nil, "TO-DO")}
	expr, err := atmolang.LexAndParseExpr("‹repl›", []byte(src))
	if err != nil {
		errs = append(errs, err)
	} else {
		irx, errsir := atmoil.ExprFrom(expr)
		if errs.Add(errsir); len(errs) == 0 && irx != nil {
			// kit.lookups.namesInScopeAll.RepopulateDefsAndIdentsFor(nil, irx, )
			// if retdesc, err := me.inferFactsForExpr(kit, irx); err != nil {
			// 	errs = append(errs, err)
			// } else {
			// 	str = retdesc.String()
			// }
		}
	}
	return
}
