package atemlang

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

func (me *AstDef) newIdent(ctx *ctxParseTld, arg int, ttmp udevlex.Tokens, at int) (err *Error) {
	tok, this, isarg := &ttmp[at], &me.Name, arg > -1
	if isarg {
		this = &me.Args[arg]
	}

	namedesc := ustr.If(isarg, "argument", "definition")
	if s, k := tok.Meta.Orig, tok.Kind(); k != udevlex.TOKEN_IDENT && k != udevlex.TOKEN_OPISH {
		err = errSyntax(tok, "not a valid "+namedesc+" name: "+s)
	} else if _s := ustr.If(me.IsTopLevel && s[0] == '_', s[1:], s); (!isarg) && k != udevlex.TOKEN_OPISH && !ustr.BeginsLower(_s) {
		err = errSyntax(tok, "not a valid "+namedesc+" name: "+_s+" should begin with lower-case letter")
	} else if isarg && s[0] == '_' && len(s) > 1 {
		err = errSyntax(tok, "not a valid "+namedesc+" name: "+s+" shouldn't begin with underscore")
	} else if k == udevlex.TOKEN_OPISH && len(s) == 1 {
		err = errSyntax(tok, "not a valid "+namedesc+" name: `"+s+"` needs to be 2 or more characters")
	} else if tok.IsOpishAndAnyOneOf(langReservedOps...) {
		err = errSyntax(tok, "not a valid "+namedesc+" name: `"+s+"` is reserved and cannot be overloaded")
	} else {
		this.Val, this.IsOpish, this.IsTag = s, k == udevlex.TOKEN_OPISH, isarg && ustr.BeginsUpper(s)
		ctx.setTokenAndCommentsFor(&this.AstBaseTokens, &this.AstBaseComments, ttmp, at)
	}
	return
}

func (me *ctxParseTld) newExprIdent(toks udevlex.Tokens) *AstIdent {
	var this AstIdent
	me.setTokenAndCommentsFor(&this.AstBaseTokens, &this.AstBaseComments, toks, 0)
	this.Val, this.IsOpish, this.IsTag =
		this.Tokens[0].Str, this.Tokens[0].Kind() == udevlex.TOKEN_OPISH, ustr.BeginsUpper(this.Tokens[0].Str)
	return &this
}

func (me *ctxParseTld) newExprLitFloat(toks udevlex.Tokens) *AstExprLitFloat {
	var this AstExprLitFloat
	me.setTokenAndCommentsFor(&this.AstBaseTokens, &this.AstBaseComments, toks, 0)
	this.Val = this.Tokens[0].Float
	return &this
}

func (me *ctxParseTld) newExprLitUint(toks udevlex.Tokens) *AstExprLitUint {
	var this AstExprLitUint
	me.setTokenAndCommentsFor(&this.AstBaseTokens, &this.AstBaseComments, toks, 0)
	this.Val = this.Tokens[0].Uint
	return &this
}

func (me *ctxParseTld) newExprLitRune(toks udevlex.Tokens) *AstExprLitRune {
	var this AstExprLitRune
	me.setTokenAndCommentsFor(&this.AstBaseTokens, &this.AstBaseComments, toks, 0)
	this.Val = this.Tokens[0].Rune()
	return &this
}

func (me *ctxParseTld) newExprLitStr(toks udevlex.Tokens) *AstExprLitStr {
	var this AstExprLitStr
	me.setTokenAndCommentsFor(&this.AstBaseTokens, &this.AstBaseComments, toks, 0)
	this.Val = this.Tokens[0].Str
	return &this
}

func (me *ctxParseTld) setTokenAndCommentsFor(tbase *AstBaseTokens, cbase *AstBaseComments, toks udevlex.Tokens, at int) {
	at = me.mto[&toks[at]]
	tld := &me.curDef.AstBaseTokens
	tbase.Tokens = tld.Tokens[at : at+1]
	for _, ci := range me.mtc[&tld.Tokens[at]] {
		cbase.Comments = append(cbase.Comments, newAstComment(tld.Tokens, ci))
	}
}

func (me *ctxParseTld) setTokensFor(this *AstBaseTokens, toks udevlex.Tokens, untilTok *udevlex.Token) {
	if untilTok != nil {
		for i := range toks {
			if &toks[i] == untilTok {
				toks = toks[:i]
				break
			}
		}
	}
	ifirst, ilast := me.mto[&toks[0]], me.mto[&toks[len(toks)-1]]
	tld := &me.curDef.AstBaseTokens
	this.Tokens = tld.Tokens[ifirst : ilast+1]
}
