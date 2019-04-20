package atmocorefn

import (
	"github.com/metaleap/atmo/lang"
)

func newIdentFrom(orig *atmolang.AstIdent) (ident IAstIdent) {
	// val opish tag affix
	switch {
	case orig.IsTag:
		// var tag AstIdentTag
		// tag.initFrom(orig)
		// return &tag
	case orig.IsOpish:
	default:
	}
	return
}
