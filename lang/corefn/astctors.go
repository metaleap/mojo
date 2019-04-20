package atmocorefn

import (
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang"
)

func newIdentFrom(orig *atmolang.AstIdent) (ident IAstIdent, errs atmo.Errors) {
	// val opish tag affix
	switch {
	case orig.IsTag:
		var tag AstIdentTag
		ident, errs = &tag, tag.initFrom(orig)
	case orig.IsOpish:
	default:
	}
	return
}
