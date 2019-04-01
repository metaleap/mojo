package odlang

import (
	"github.com/go-leap/dev/lex"
	"github.com/go-leap/str"
)

type AstBaseComments struct {
	Comments []*AstComment
}

type AstBaseTokens struct {
	Tokens udevlex.Tokens
}

type AstTopLevel struct {
	AstBaseTokens
	AstBaseComments
	Def IAstDef
}

type AstComment struct {
	AstBaseTokens
}

func (me *AstComment) ContentText() string     { return me.Tokens[0].Str }
func (me *AstComment) IsSelfTerminating() bool { return me.Tokens[0].IsCommentSelfTerminating() }

type AstBase struct {
	AstBaseTokens
	AstBaseComments
}

type AstIdent struct {
	AstBase
}

func (me *AstIdent) BeginsLower() bool { return ustr.BeginsLower(me.Tokens[0].Str) }
func (me *AstIdent) BeginsUpper() bool { return ustr.BeginsUpper(me.Tokens[0].Str) }
func (me *AstIdent) Val() string       { return me.Tokens[0].Str }
func (me *AstIdent) IsOpish() bool     { return me.Tokens[0].Kind() == udevlex.TOKEN_OTHER }

type IAstDef interface {
	Base() *AstDefBase
}

type AstDefBase struct {
	AstBaseTokens
	Name AstIdent
	Args []AstIdent

	IsDefType bool
}

func (me *AstDefBase) Base() *AstDefBase { return me }

func (me *AstDefBase) ensureArgsLen(l int) {
	if ol := len(me.Args); ol > l {
		me.Args = me.Args[:l]
	} else if ol < l {
		me.Args = make([]AstIdent, l)
	}
}

type AstDefType struct {
	AstDefBase
}

type AstDefFunc struct {
	AstDefBase
	Body interface{}
}

type AstExprLitUint struct {
	AstBase
}

func (me *AstExprLitUint) Val() uint64 { return me.Tokens[0].Uint }

type AstExprLitFloat struct {
	AstBase
}

func (me *AstExprLitFloat) Val() float64 { return me.Tokens[0].Float }

type AstExprLitRune struct {
	AstBase
}

func (me *AstExprLitRune) Val() rune { return me.Tokens[0].Rune() }

type AstExprLitStr struct {
	AstBase
}

func (me *AstExprLitStr) Val() string { return me.Tokens[0].Str }

type AstExprIdent struct {
	AstIdent
}

type AstExprLet struct {
	Defs []interface{}
}
