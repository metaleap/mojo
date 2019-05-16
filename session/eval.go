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
		if kit := me.Kits.all.byDirPath(me.Dirs.sess[0]); len(errs) == 0 && irx != nil {
			kit.lookups.namesInScopeAll.repopulateAstIdents(irx)
			if retdesc, err := me.inferExpr(kit, irx); err != nil {
				errs = append(errs, err)
			} else {
				str = retdesc.String()
			}
		}
	}
	return
}
