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

func (me *AstDefBase) newIdent(ctx *ctxParseTopLevelDef, arg int, ttmp udevlex.Tokens, at int) *Error {
	this, isarg := &me.Name, arg > -1
	if isarg {
		this = &me.Args[arg]
	}

	if s, k := ttmp[at].Str, ttmp[at].Kind(); (k != udevlex.TOKEN_IDENT && k != udevlex.TOKEN_OTHER) ||
		(me.IsDefType && isarg && (k == udevlex.TOKEN_OTHER || !ustr.BeginsLower(s))) {
		return errAt(&ttmp[at], "not a valid "+
			ustr.If(!isarg, "definition", ustr.If(me.IsDefType, "type-var", "argument"))+" name")
	}
	ctx.setTokenAndCommentsFor(&this.AstBaseTokens, &this.AstBaseComments, ttmp, at)
	return nil
}

func (me *ctxParseTopLevelDef) newExprLitFloat(toks udevlex.Tokens) *AstExprLitFloat {
	var this AstExprLitFloat
	me.setTokenAndCommentsFor(&this.AstBaseTokens, &this.AstBaseComments, toks, 0)
	return &this
}

func (me *ctxParseTopLevelDef) newExprLitUint(toks udevlex.Tokens) *AstExprLitUint {
	var this AstExprLitUint
	me.setTokenAndCommentsFor(&this.AstBaseTokens, &this.AstBaseComments, toks, 0)
	return &this
}

func (me *ctxParseTopLevelDef) newExprLitRune(toks udevlex.Tokens) *AstExprLitRune {
	var this AstExprLitRune
	me.setTokenAndCommentsFor(&this.AstBaseTokens, &this.AstBaseComments, toks, 0)
	return &this
}

func (me *ctxParseTopLevelDef) newExprLitStr(toks udevlex.Tokens) *AstExprLitStr {
	var this AstExprLitStr
	me.setTokenAndCommentsFor(&this.AstBaseTokens, &this.AstBaseComments, toks, 0)
	return &this
}

func (me *ctxParseTopLevelDef) newExprIdent(toks udevlex.Tokens) *AstIdent {
	var this AstIdent
	me.setTokenAndCommentsFor(&this.AstBaseTokens, &this.AstBaseComments, toks, 0)
	return &this
}

func (me *ctxParseTopLevelDef) newTypeExprIdent(toks udevlex.Tokens) *AstTypeExprIdent {
	var this AstTypeExprIdent
	me.setTokenAndCommentsFor(&this.AstBaseTokens, &this.AstBaseComments, toks, 0)
	return &this
}

func (me *ctxParseTopLevelDef) setTokenAndCommentsFor(tbase *AstBaseTokens, cbase *AstBaseComments, toks udevlex.Tokens, at int) {
	at = me.mto[&toks[at]]
	tld := &me.def.Base().AstBaseTokens
	tbase.Tokens = tld.Tokens[at : at+1]
	for _, ci := range me.mtc[&tld.Tokens[at]] {
		cbase.Comments = append(cbase.Comments, newAstComment(tld.Tokens, ci))
	}
}

func (me *ctxParseTopLevelDef) setTokensFor(this *AstBaseTokens, toks udevlex.Tokens, untilTok *udevlex.Token) {
	if untilTok != nil {
		for i := range toks {
			if &toks[i] == untilTok {
				toks = toks[:i]
				break
			}
		}
	}
	ifirst, ilast := me.mto[&toks[0]], me.mto[&toks[len(toks)-1]]
	tld := &me.def.Base().AstBaseTokens
	this.Tokens = tld.Tokens[ifirst : ilast+1]
}
