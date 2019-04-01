package odlang

import (
	"github.com/go-leap/dev/lex"
	"github.com/go-leap/str"
)

func newAstComment(tokens udevlex.Tokens, at int) *AstComment {
	var this AstComment
	this.Tokens = tokens[at : at+1]
	return &this
}

func (me *AstDefBase) newIdent(arg int, ttmp udevlex.Tokens, at int, mtc mapTokCmnts, mti mapTokOldIdxs) *Error {
	this, isarg := &me.Name, arg > -1
	if isarg {
		this = &me.Args[arg]
	}

	if k := ttmp[at].Kind(); (isarg && !ustr.BeginsLower(ttmp[at].Str)) ||
		(k != udevlex.TOKEN_IDENT && (k != udevlex.TOKEN_OTHER || isarg)) {
		return errAt(&ttmp[at], "not a valid "+ustr.If(isarg, ustr.If(me.IsDefType, "type-var", "argument"), "definition")+" name")
	}

	if mti != nil {
		at = mti[&ttmp[at]]
	}
	this.AstBaseTokens.Tokens = me.Tokens[at : at+1]
	if mtc != nil {
		for _, ci := range mtc[&me.Tokens[at]] {
			this.Comments = append(this.Comments, newAstComment(me.Tokens, ci))
		}
	}
	return nil
}
