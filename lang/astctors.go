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

func (me *AstDefBase) newIdent(ctx *ctxTopLevelDef, arg int, ttmp udevlex.Tokens, at int) *Error {
	this, isarg := &me.Name, arg > -1
	if isarg {
		this = &me.Args[arg]
	}

	if k := ttmp[at].Kind(); (isarg && !ustr.BeginsLower(ttmp[at].Str)) ||
		(k != udevlex.TOKEN_IDENT && (k != udevlex.TOKEN_OTHER || isarg)) {
		return errAt(&ttmp[at], "not a valid "+ustr.If(isarg, ustr.If(me.IsDefType, "type-var", "argument"), "definition")+" name")
	}
	ctx.setTokenAndCommentsFor(&this.AstBase, ttmp, at)
	return nil
}

func (me *ctxTopLevelDef) newExprLitFloat(toks udevlex.Tokens) *AstExprLitFloat {
	var this AstExprLitFloat
	me.setTokenAndCommentsFor(&this.AstBase, toks, 0)
	return &this
}

func (me *ctxTopLevelDef) newExprLitUint(toks udevlex.Tokens) *AstExprLitUint {
	var this AstExprLitUint
	me.setTokenAndCommentsFor(&this.AstBase, toks, 0)
	return &this
}

func (me *ctxTopLevelDef) newExprLitRune(toks udevlex.Tokens) *AstExprLitRune {
	var this AstExprLitRune
	me.setTokenAndCommentsFor(&this.AstBase, toks, 0)
	return &this
}

func (me *ctxTopLevelDef) newExprLitStr(toks udevlex.Tokens) *AstExprLitStr {
	var this AstExprLitStr
	me.setTokenAndCommentsFor(&this.AstBase, toks, 0)
	return &this
}

func (me *ctxTopLevelDef) newExprIdent(toks udevlex.Tokens) *AstExprIdent {
	var this AstExprIdent
	me.setTokenAndCommentsFor(&this.AstBase, toks, 0)
	return &this
}

func (me *ctxTopLevelDef) setTokenAndCommentsFor(this *AstBase, toks udevlex.Tokens, at int) {
	at = me.mto[&toks[at]]
	tld := &me.def.Base().AstBaseTokens
	this.Tokens = tld.Tokens[at : at+1]
	for _, ci := range me.mtc[&tld.Tokens[at]] {
		this.Comments = append(this.Comments, newAstComment(tld.Tokens, ci))
	}
}

func (me *ctxTopLevelDef) setTokensFor(this *AstBaseTokens, toks udevlex.Tokens, untilTok *udevlex.Token) {
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
