package atmolang

import (
	"github.com/go-leap/dev/lex"
	"github.com/go-leap/str"
)

type (
	tldParse struct {
		file            *AstFile
		curDef          *AstDef
		mto             map[*udevlex.Token]int   // maps comments-stripped Tokens to orig Tokens
		mtc             map[*udevlex.Token][]int // maps comments-stripped Tokens to comment Tokens in orig
		indentHint      int
		parensLevel     int
		atTopLevelStill bool
		b               AstBuilder
	}
)

func (me *tldParse) parseExprIdent(toks udevlex.Tokens) *AstIdent {
	var this AstIdent
	me.setTokenFor(&this.AstBaseTokens, toks, 0)
	this.Val, this.IsOpish, this.IsTag =
		this.Tokens[0].Str, this.Tokens[0].Kind() == udevlex.TOKEN_OPISH, ustr.BeginsUpper(this.Tokens[0].Str)
	if this.Val == "()" {
		this.IsOpish = true
	}
	return &this
}

func (me *tldParse) parseExprLitFloat(toks udevlex.Tokens) *AstExprLitFloat {
	var this AstExprLitFloat
	me.setTokenFor(&this.AstBaseTokens, toks, 0)
	this.Val = this.Tokens[0].Float
	return &this
}

func (me *tldParse) parseExprLitUint(toks udevlex.Tokens) *AstExprLitUint {
	var this AstExprLitUint
	me.setTokenFor(&this.AstBaseTokens, toks, 0)
	this.Val = this.Tokens[0].Uint
	return &this
}

func (me *tldParse) parseExprLitRune(toks udevlex.Tokens) *AstExprLitRune {
	var this AstExprLitRune
	me.setTokenFor(&this.AstBaseTokens, toks, 0)
	this.Val = this.Tokens[0].Rune()
	return &this
}

func (me *tldParse) parseExprLitStr(toks udevlex.Tokens) *AstExprLitStr {
	var this AstExprLitStr
	me.setTokenFor(&this.AstBaseTokens, toks, 0)
	this.Val = this.Tokens[0].Str
	return &this
}

// previously setTokenAndCommentsFor
func (me *tldParse) setTokenFor(tbase *AstBaseTokens /*cbase *AstBaseComments,*/, toks udevlex.Tokens, at int) {
	at = me.mto[&toks[at]]
	tld := &me.curDef.AstBaseTokens
	tbase.Tokens = tld.Tokens[at : at+1]
	// cidxs := me.mtc[&tld.Tokens[at]]
	// cbase.Comments = make([]AstComment, len(cidxs))
	// for i, cidx := range cidxs {
	// 	cbase.Comments[i].initFrom(tld.Tokens, cidx)
	// }
}

func (me *tldParse) setTokensFor(this *AstBaseTokens, toks udevlex.Tokens) {
	ifirst, ilast := me.mto[&toks[0]], me.mto[&toks[len(toks)-1]]
	tld := &me.curDef.AstBaseTokens
	this.Tokens = tld.Tokens[ifirst : ilast+1]
}
