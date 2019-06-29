package atmolang

import (
	"github.com/go-leap/dev/lex"
	"github.com/go-leap/str"
)

type (
	ctxTldParse struct {
		file              *AstFile
		indentHintForLet  int
		parensLevel       int
		exprWillBeDefBody udevlex.Tokens // nil-ness as falsy, non-nil truthy even if empty
	}
)

func (me *ctxTldParse) parseExprIdent(toks udevlex.Tokens, emptySeps bool) *AstIdent {
	var this AstIdent
	if emptySeps {
		this.Tokens, this.Val, this.IsOpish, this.IsTag =
			toks, toks[0].Meta.Orig+toks[1].Meta.Orig, true, false
	} else {
		this.Tokens, this.Val, this.IsOpish, this.IsTag =
			toks[0:1], toks[0].Str, toks[0].Kind == udevlex.TOKEN_OPISH, ustr.BeginsUpper(toks[0].Str)
	}
	return &this
}

func (me *ctxTldParse) parseExprLitFloat(toks udevlex.Tokens) *AstExprLitFloat {
	var this AstExprLitFloat
	this.Tokens, this.Val = toks[0:1], toks[0].Float
	return &this
}

func (me *ctxTldParse) parseExprLitUint(toks udevlex.Tokens) *AstExprLitUint {
	var this AstExprLitUint
	this.Tokens, this.Val = toks[0:1], toks[0].Uint
	return &this
}

func (me *ctxTldParse) parseExprLitStr(toks udevlex.Tokens) *AstExprLitStr {
	var this AstExprLitStr
	this.Tokens, this.Val = toks[0:1], toks[0].Str
	return &this
}
