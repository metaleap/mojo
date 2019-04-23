package atmocorefn

import (
	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang"
)

func newIdentFrom(orig *atmolang.AstIdent) (ident IAstIdent, errs atmo.Errors) {
	if orig.IsTag || ustr.BeginsUpper(orig.Val) {
		var tag AstIdentTag
		ident, errs = &tag, tag.initFrom(orig)

	} else if orig.IsOpish {
		if orig.Val == "()" {
			var empar AstIdentEmptyParens
			ident = &empar
		} else {
			var op AstIdentOp
			ident, errs = &op, op.initFrom(orig)
		}

	} else if orig.Val[0] != '_' {
		var name AstIdentName
		ident, errs = &name, name.initFrom(orig)

	} else if ustr.IsRepeat(orig.Val) {
		var unsco AstIdentUnderscores
		ident, errs = &unsco, unsco.initFrom(orig)

	} else if orig.Val[1] != '_' {
		var idvar AstIdentVar
		ident, errs = &idvar, idvar.initFrom(orig)

	} else {
		errs.AddFrom(atmo.ErrCatNaming, &orig.Tokens[0], "not a valid identifier: begins with multiple underscores")
	}
	return
}
