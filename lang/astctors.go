package odlang

import (
	"github.com/go-leap/dev/lex"
)

func newAstComment(tokens udevlex.Tokens, at int) *AstComment {
	var this AstComment
	this.Tokens = tokens[at : at+1]
	return &this
}

func (me *AstDefBase) ensureArgsLen(l int) {
	if ol := len(me.Args); ol > l {
		me.Args = me.Args[:l]
	} else if ol < l {
		me.Args = make([]AstIdent, l)
	}
}

func (me *AstDefBase) newIdent(arg int, torig udevlex.Tokens, ttmp udevlex.Tokens, at int, mtc mapTokCmnts, mti mapTokOldIdxs) {
	this := &me.Name
	if arg > -1 {
		this = &me.Args[arg]
	}
	if mti != nil {
		at = mti[&ttmp[at]]
	}
	this.AstBaseTokens.Tokens = torig[at : at+1]
	if mtc != nil {
		for _, ci := range mtc[&torig[at]] {
			this.Comments = append(this.Comments, newAstComment(torig, ci))
		}
	}
}
