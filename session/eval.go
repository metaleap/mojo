package atmosess

import (
	"errors"

	"github.com/metaleap/atmo/lang"
	"github.com/metaleap/atmo/lang/irfun"
)

func (me *Ctx) Eval(kit *Kit, src string, maybeFauxKitDir string) (str string, errs []error) {
	expr, err := atmolang.LexAndParseExpr("‹repl›", []byte(src))
	if err != nil {
		errs = append(errs, err)
	} else {
		irx, errsir := atmolang_irfun.ExprFrom(expr)
		errs = append(errs, errsir.Errors()...)
		if maybeFauxKitDir == "" {
			if fauxkitdirs := me.FauxKitDirPaths(); len(fauxkitdirs) > 0 {
				maybeFauxKitDir = fauxkitdirs[0]
			}
		}
		if kit := me.Kits.all.byDirPath(maybeFauxKitDir); kit == nil {
			errs = append(errs, errors.New("bad faux-kit dir path `"+maybeFauxKitDir+"`"))
		} else if len(errs) == 0 && irx != nil {
			kit.lookups.namesInScopeAll.repopulateAstIdents(irx)
			// if retdesc, err := me.inferFactsForExpr(kit, irx); err != nil {
			// 	errs = append(errs, err)
			// } else {
			// 	str = retdesc.String()
			// }
		}
	}
	return
}
