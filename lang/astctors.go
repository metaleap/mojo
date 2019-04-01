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

func (me *AstDefBase) newIdent(arg int, ttmp udevlex.Tokens, at int, ctx *ctxParseDef) *Error {
	this, isarg := &me.Name, arg > -1
	if isarg {
		this = &me.Args[arg]
	}

	if k := ttmp[at].Kind(); (isarg && !ustr.BeginsLower(ttmp[at].Str)) ||
		(k != udevlex.TOKEN_IDENT && (k != udevlex.TOKEN_OTHER || isarg)) {
		return errAt(&ttmp[at], "not a valid "+ustr.If(isarg, ustr.If(me.IsDefType, "type-var", "argument"), "definition")+" name")
	}
	me.setTokenAndCommentsFor(&this.AstBase, ttmp, at, ctx)
	return nil
}

func (me *AstBaseTokens) newExprLitFloat(toks udevlex.Tokens, ctx *ctxParseDef) *AstExprLitFloat {
	var this AstExprLitFloat
	me.setTokenAndCommentsFor(&this.AstBase, toks, 0, ctx)
	return &this
}

func (me *AstBaseTokens) newExprLitUint(toks udevlex.Tokens, ctx *ctxParseDef) *AstExprLitUint {
	var this AstExprLitUint
	me.setTokenAndCommentsFor(&this.AstBase, toks, 0, ctx)
	return &this
}

func (me *AstBaseTokens) newExprLitRune(toks udevlex.Tokens, ctx *ctxParseDef) *AstExprLitRune {
	var this AstExprLitRune
	me.setTokenAndCommentsFor(&this.AstBase, toks, 0, ctx)
	return &this
}

func (me *AstBaseTokens) newExprLitStr(toks udevlex.Tokens, ctx *ctxParseDef) *AstExprLitStr {
	var this AstExprLitStr
	me.setTokenAndCommentsFor(&this.AstBase, toks, 0, ctx)
	return &this
}

func (me *AstBaseTokens) newExprIdent(toks udevlex.Tokens, ctx *ctxParseDef) *AstExprIdent {
	var this AstExprIdent
	me.setTokenAndCommentsFor(&this.AstBase, toks, 0, ctx)
	return &this
}

func (me *AstBaseTokens) setTokenAndCommentsFor(this *AstBase, ttmp udevlex.Tokens, at int, ctx *ctxParseDef) {
	at = ctx.mapTokOldIdxs[&ttmp[at]]
	this.Tokens = me.Tokens[at : at+1]
	for _, ci := range ctx.mapTokCmnts[&me.Tokens[at]] {
		this.Comments = append(this.Comments, newAstComment(me.Tokens, ci))
	}
}
