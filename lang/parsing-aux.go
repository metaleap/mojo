package atmolang

import (
	"github.com/go-leap/dev/lex"
	"github.com/go-leap/str"
)

type (
	tldParse struct {
		file            *AstFile
		curDef          *AstDef
		indentHint      int
		parensLevel     int
		atTopLevelStill bool
		b               AstBuilder
	}
)

func (me *tldParse) parseExprIdent(toks udevlex.Tokens) *AstIdent {
	var this AstIdent
	this.Tokens, this.Val, this.IsOpish, this.IsTag =
		toks[0:1], toks[0].Str, toks[0].Kind() == udevlex.TOKEN_OPISH, ustr.BeginsUpper(toks[0].Str)
	if this.Val == "()" {
		this.IsOpish = true
	}
	return &this
}

func (me *tldParse) parseExprLitFloat(toks udevlex.Tokens) *AstExprLitFloat {
	var this AstExprLitFloat
	this.Tokens, this.Val = toks[0:1], toks[0].Float
	return &this
}

func (me *tldParse) parseExprLitUint(toks udevlex.Tokens) *AstExprLitUint {
	var this AstExprLitUint
	this.Tokens, this.Val = toks[0:1], toks[0].Uint
	return &this
}

func (me *tldParse) parseExprLitRune(toks udevlex.Tokens) *AstExprLitRune {
	var this AstExprLitRune
	this.Tokens, this.Val = toks[0:1], toks[0].Rune()
	return &this
}

func (me *tldParse) parseExprLitStr(toks udevlex.Tokens) *AstExprLitStr {
	var this AstExprLitStr
	this.Tokens, this.Val = toks[0:1], toks[0].Str
	return &this
}
