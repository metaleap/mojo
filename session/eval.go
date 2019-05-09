package atmosess

import (
	"github.com/metaleap/atmo/lang"
	"github.com/metaleap/atmo/lang/irfun"
)

func (me *Ctx) Eval(kit *Kit, src string) (str string, errs []error) {
	expr, err := atmolang.LexAndParseExpr("‹repl›", []byte(src))
	if err != nil {
		errs = append(errs, err)
	} else {
		irx, errors := atmolang_irfun.ExprFrom(expr)
		for e := range errors {
			errs = append(errs, &errors[e])
		}
		if len(errs) == 0 && irx != nil {
			if _, retdesc, err := me.reduceExpr(me.Kits.all.byDirPath(me.Dirs.sess[0]), irx); err != nil {
				errs = append(errs, err)
			} else {
				str = retdesc.String()
			}
		}
	}
	return
}
