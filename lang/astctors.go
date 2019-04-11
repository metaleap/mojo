package odlang

import (
	"github.com/go-leap/dev/lex"
	"github.com/go-leap/str"
)

func newAstComment(tokens udevlex.Tokens, at int) *AstComment {
	var this AstComment
	this.Tokens = tokens[at : at+1]
	this.ContentText, this.IsSelfTerminating = this.Tokens[0].Str, this.Tokens[0].IsCommentSelfTerminating()
	return &this
}

func (me *AstDefBase) newIdent(ctx *ctxParse, arg int, ttmp udevlex.Tokens, at int) *Error {
	this, isarg := &me.Name, arg > -1
	if isarg {
		this = &me.Args[arg]
	}

	if s, k := ttmp[at].Str, ttmp[at].Kind(); (k != udevlex.TOKEN_IDENT && k != udevlex.TOKEN_OPISH) ||
		(me.IsDefType && isarg && (k == udevlex.TOKEN_OPISH || !ustr.BeginsLower(s))) {
		return errAt(&ttmp[at], ErrCatSyntax, "not a valid "+
			ustr.If(!isarg, "definition", ustr.If(me.IsDefType, "type-var", "argument"))+" name")
	}

	ctx.setTokenAndCommentsFor(&this.AstBaseTokens, &this.AstBaseComments, ttmp, at)
	this.Val, this.IsOpish = this.Tokens[0].Str, me.Tokens[0].Kind() == udevlex.TOKEN_OPISH
	return nil
}

func (me *ctxParse) newExprLitFloat(toks udevlex.Tokens) *AstExprLitFloat {
	var this AstExprLitFloat
	me.setTokenAndCommentsFor(&this.AstBaseTokens, &this.AstBaseComments, toks, 0)
	this.Val = this.Tokens[0].Float
	return &this
}

func (me *ctxParse) newExprLitUint(toks udevlex.Tokens) *AstExprLitUint {
	var this AstExprLitUint
	me.setTokenAndCommentsFor(&this.AstBaseTokens, &this.AstBaseComments, toks, 0)
	this.Val = this.Tokens[0].Uint
	return &this
}

func (me *ctxParse) newExprLitRune(toks udevlex.Tokens) *AstExprLitRune {
	var this AstExprLitRune
	me.setTokenAndCommentsFor(&this.AstBaseTokens, &this.AstBaseComments, toks, 0)
	this.Val = this.Tokens[0].Rune()
	return &this
}

func (me *ctxParse) newExprLitStr(toks udevlex.Tokens) *AstExprLitStr {
	var this AstExprLitStr
	me.setTokenAndCommentsFor(&this.AstBaseTokens, &this.AstBaseComments, toks, 0)
	this.Val = this.Tokens[0].Str
	return &this
}

func (me *ctxParse) newExprIdent(toks udevlex.Tokens) *AstIdent {
	var this AstIdent
	me.setTokenAndCommentsFor(&this.AstBaseTokens, &this.AstBaseComments, toks, 0)
	this.Val, this.IsOpish = this.Tokens[0].Str, this.Tokens[0].Kind() == udevlex.TOKEN_OPISH
	return &this
}

func (me *ctxParse) newTypeExprIdent(toks udevlex.Tokens) *AstTypeExprIdent {
	var this AstTypeExprIdent
	me.setTokenAndCommentsFor(&this.AstBaseTokens, &this.AstBaseComments, toks, 0)
	this.Val, this.IsOpish = this.Tokens[0].Str, this.Tokens[0].Kind() == udevlex.TOKEN_OPISH
	return &this
}

func (me *ctxParse) setTokenAndCommentsFor(tbase *AstBaseTokens, cbase *AstBaseComments, toks udevlex.Tokens, at int) {
	at = me.mto[&toks[at]]
	tld := &me.def.DefBase().AstBaseTokens
	tbase.Tokens = tld.Tokens[at : at+1]
	for _, ci := range me.mtc[&tld.Tokens[at]] {
		cbase.Comments = append(cbase.Comments, newAstComment(tld.Tokens, ci))
	}
}

func (me *ctxParse) setTokensFor(this *AstBaseTokens, toks udevlex.Tokens, untilTok *udevlex.Token) {
	if untilTok != nil {
		for i := range toks {
			if &toks[i] == untilTok {
				toks = toks[:i]
				break
			}
		}
	}
	ifirst, ilast := me.mto[&toks[0]], me.mto[&toks[len(toks)-1]]
	tld := &me.def.DefBase().AstBaseTokens
	this.Tokens = tld.Tokens[ifirst : ilast+1]
}
