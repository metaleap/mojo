package odlang

import (
	"github.com/go-leap/dev/lex"
)

func newAstComment(tokens udevlex.Tokens, at int) *AstComment {
	var this AstComment
	this.Tokens = tokens[at : at+1]
	return &this
}

func (me *AstDefBase) newIdent(torig udevlex.Tokens, ttmp udevlex.Tokens, at int, mtc mapTokCmnts, mti mapTokOldIdxs) {
	if mti != nil {
		at = mti[&ttmp[at]]
	}
	me.Name.AstBaseTokens.Tokens = torig[at : at+1]
	if mtc != nil {
		for _, ci := range mtc[&torig[at]] {
			me.Name.Comments = append(me.Name.Comments, newAstComment(torig, ci))
		}
	}
}
