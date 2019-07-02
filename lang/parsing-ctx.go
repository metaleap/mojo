package atmolang

import (
	"github.com/go-leap/dev/lex"
	"github.com/go-leap/str"
)

type (
	ctxTldParse struct {
		file        *AstFile
		parensLevel int
	}
)

func (me *ctxTldParse) parseExprIdent(toks udevlex.Tokens, emptySeps bool) *AstIdent {
	var this AstIdent
	if emptySeps {
		this.Tokens, this.Val, this.IsOpish, this.IsTag =
			toks, toks[0].Lexeme+toks[1].Lexeme, true, false
	} else {
		this.Tokens, this.Val, this.IsOpish, this.IsTag =
			toks[0:1], toks[0].Lexeme, toks[0].Kind == udevlex.TOKEN_OPISH, ustr.BeginsUpper(toks[0].Lexeme)
	}
	return &this
}

func (me *ctxTldParse) parseExprLitFloat(toks udevlex.Tokens) *AstExprLitFloat {
	var this AstExprLitFloat
	this.Tokens, this.Val = toks[0:1], toks[0].Val.(float64)
	return &this
}

func (me *ctxTldParse) parseExprLitUint(toks udevlex.Tokens) *AstExprLitUint {
	var this AstExprLitUint
	this.Tokens, this.Val = toks[0:1], toks[0].Val.(uint64)
	return &this
}

func (me *ctxTldParse) parseExprLitStr(toks udevlex.Tokens) *AstExprLitStr {
	var this AstExprLitStr
	this.Tokens, this.Val = toks[0:1], toks[0].Val.(string)
	return &this
}
