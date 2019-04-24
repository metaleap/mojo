package atmolang

import (
	"github.com/go-leap/dev/lex"
	"github.com/go-leap/str"
)

type (
	ctxParseTld struct {
		file            *AstFile
		curDef          *AstDef
		mto             map[*udevlex.Token]int   // maps comments-stripped Tokens to orig Tokens
		mtc             map[*udevlex.Token][]int // maps comments-stripped Tokens to comment Tokens in orig
		indentHint      int
		parensLevel     int
		atTopLevelStill bool
	}
)

func (me *ctxParseTld) newExprIdent(toks udevlex.Tokens) *AstIdent {
	var this AstIdent
	me.setTokenAndCommentsFor(&this.AstBaseTokens, &this.AstBaseComments, toks, 0)
	this.Val, this.IsOpish, this.IsTag =
		this.Tokens[0].Str, this.Tokens[0].Kind() == udevlex.TOKEN_OPISH, ustr.BeginsUpper(this.Tokens[0].Str)
	if this.Val == "()" {
		this.IsOpish = true
	}
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
	cidxs := me.mtc[&tld.Tokens[at]]
	cbase.Comments = make([]AstComment, len(cidxs))
	for i, cidx := range cidxs {
		cbase.Comments[i].initFrom(tld.Tokens, cidx)
	}
}

func (me *ctxParseTld) getTokensFor(from *AstBaseTokens, until *AstBaseTokens) udevlex.Tokens {
	ifirst, ilast := me.mto[from.Tokens.First(nil)], me.mto[until.Tokens.Last(nil)]
	tld := &me.curDef.AstBaseTokens
	return tld.Tokens[ifirst : ilast+1]
}

func (me *ctxParseTld) setTokensFor(this *AstBaseTokens, toks udevlex.Tokens) {
	ifirst, ilast := me.mto[&toks[0]], me.mto[&toks[len(toks)-1]]
	tld := &me.curDef.AstBaseTokens
	this.Tokens = tld.Tokens[ifirst : ilast+1]
}
