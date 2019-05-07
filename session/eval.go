package atmosess

import (
	"os"

	"github.com/metaleap/atmo/lang"
)

func (me *Ctx) Eval(kit *Kit, src string) (errs []error) {
	expr, err := atmolang.LexAndParseExpr([]byte(src))
	if err != nil {
		errs = append(errs, err)
	} else {
		atmolang.PrintTo(nil, expr, os.Stdout, false)
	}
	return
}
