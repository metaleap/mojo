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

type IAstDef interface {
	Base() *AstDefBase
	parseDefBody(*ctxTopLevelDef, udevlex.Tokens) *Error
}

type AstDefBase struct {
	AstBaseTokens
	Name AstExprIdent
	Args []AstExprIdent

	IsDefType bool
}

func (me *AstDefBase) Base() *AstDefBase { return me }

func (me *AstDefBase) ensureArgsLen(l int) {
	if ol := len(me.Args); ol > l {
		me.Args = me.Args[:l]
	} else if ol < l {
		me.Args = make([]AstExprIdent, l)
	}
}

type AstDefType struct {
	AstDefBase
}

type AstDefFunc struct {
	AstDefBase
	Body IAstExpr
}

type IAstExpr interface {
	Base() *AstExprBase
}

type AstExprBase struct {
	AstBase
}

func (me *AstExprBase) Base() *AstExprBase { return me }

type AstExprLitUint struct {
	AstExprBase
}

func (me *AstExprLitUint) Val() uint64 { return me.Tokens[0].Uint }

type AstExprLitFloat struct {
	AstExprBase
}

func (me *AstExprLitFloat) Val() float64 { return me.Tokens[0].Float }

type AstExprLitRune struct {
	AstExprBase
}

func (me *AstExprLitRune) Val() rune { return me.Tokens[0].Rune() }

type AstExprLitStr struct {
	AstExprBase
}

func (me *AstExprLitStr) Val() string { return me.Tokens[0].Str }

type AstExprIdent struct {
	AstExprBase
}

func (me *AstExprIdent) Val() string       { return me.Tokens[0].Str }
func (me *AstExprIdent) IsOpish() bool     { return me.Tokens[0].Kind() == udevlex.TOKEN_OTHER }
func (me *AstExprIdent) BeginsUpper() bool { return ustr.BeginsUpper(me.Tokens[0].Str) }
func (me *AstExprIdent) BeginsLower() bool { return ustr.BeginsLower(me.Tokens[0].Str) }

type AstExprLet struct {
	AstExprBase
	Defs []IAstDef
	Body IAstExpr
}

type AstExprCall struct {
	AstExprBase
	Callee IAstExpr
	Args   []IAstExpr
}

type AstExprCase struct {
	AstExprBase
	Scrutinee    IAstExpr
	Alts         []AstCaseAlt
	defaultIndex int
}

func (me *AstExprCase) Default() *AstCaseAlt {
	if me.defaultIndex < 0 {
		return nil
	}
	return &me.Alts[me.defaultIndex]
}

type AstCaseAlt struct {
	AstBase
	Cond IAstExpr
	Body IAstExpr
}

type AstExprCtor struct {
	AstExprBase
	Name AstExprIdent
	Args []IAstExpr
}
