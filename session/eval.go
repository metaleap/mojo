package atmosess

import (
	"os"

	"github.com/metaleap/atmo/lang"
	"github.com/metaleap/atmo/lang/irfun"
)

func (me *Ctx) Eval(kit *Kit, src string) (tmp func(), errs []error) {
	expr, err := atmolang.LexAndParseExpr("‹repl›", []byte(src))
	if err != nil {
		errs = append(errs, err)
	} else {
		irx, errors := atmolang_irfun.ExprFrom(expr)
		for e := range errors {
			errs = append(errs, &errors[e])
		}
		if irx != nil {
			tmp = func() { atmolang.PrintTo(nil, irx.Print(), os.Stdout, false) }
		}
	}
	return
}
