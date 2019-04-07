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

type IAstDef interface {
	Base() *AstDefBase
	parseDefBody(*ctxParseTopLevelDef, udevlex.Tokens) *Error
}

type AstDefBase struct {
	AstBaseTokens
	Name AstIdent
	Args []AstIdent
	Meta []IAstExpr

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
	Expr IAstTypeExpr
	Tags []AstDefTypeTag
}

type AstDefTypeTag struct {
	Name AstIdent
	Expr IAstTypeExpr
}

type AstDefFunc struct {
	AstDefBase
	Body IAstExpr
}

type IAstExpr interface {
	ExprBase() *AstExprBase
	Description() string
}

type IAstIdent interface {
	IAstExpr
	Val() string
	IsOpish() bool
	BeginsUpper() bool
	BeginsLower() bool
}

type AstExprBase struct {
	AstBaseTokens
}

func (me *AstExprBase) ExprBase() *AstExprBase { return me }
func (me *AstExprBase) Description() string    { return "(unknown) expression" }
func (me *AstExprBase) IsOp(anyOf ...string) bool {
	if len(me.Tokens) == 1 && me.Tokens[0].Kind() == udevlex.TOKEN_OPISH {
		for i := range anyOf {
			if me.Tokens[0].Meta.Orig == anyOf[i] {
				return true
			}
		}
		return len(anyOf) == 0
	}
	return false
}

type AstExprAtomBase struct {
	AstExprBase
	AstBaseComments
}

type AstExprLitBase struct {
	AstExprAtomBase
}

type AstExprLitUint struct {
	AstExprLitBase
}

func (me *AstExprLitUint) Description() string { return "'unsigned-integer literal' expression" }
func (me *AstExprLitUint) Val() uint64         { return me.Tokens[0].Uint }

type AstExprLitFloat struct {
	AstExprLitBase
}

func (me *AstExprLitFloat) Description() string { return "'float literal' expression" }
func (me *AstExprLitFloat) Val() float64        { return me.Tokens[0].Float }

type AstExprLitRune struct {
	AstExprLitBase
}

func (me *AstExprLitRune) Description() string { return "'rune literal' expression" }
func (me *AstExprLitRune) Val() rune           { return me.Tokens[0].Rune() }

type AstExprLitStr struct {
	AstExprLitBase
}

func (me *AstExprLitStr) Description() string { return "'string literal' expression" }
func (me *AstExprLitStr) Val() string         { return me.Tokens[0].Str }

type AstIdent struct {
	AstExprAtomBase
}

func (me *AstIdent) Description() string { return "'ident' expression" }
func (me *AstIdent) Val() string         { return me.Tokens[0].Str }
func (me *AstIdent) IsOpish() bool       { return me.Tokens[0].Kind() == udevlex.TOKEN_OPISH }
func (me *AstIdent) BeginsUpper() bool   { return ustr.BeginsUpper(me.Tokens[0].Str) }
func (me *AstIdent) BeginsLower() bool   { return ustr.BeginsLower(me.Tokens[0].Str) }

type AstExprLet struct {
	AstExprBase
	Defs []IAstDef
	Body IAstExpr
}

func (me *AstExprLet) Description() string { return "'let' expression" }

type AstExprAppl struct {
	AstExprBase
	Callee IAstExpr
	Args   []IAstExpr
}

func (me *AstExprAppl) Description() string { return "'composite' expression" }

type AstExprCase struct {
	AstExprBase
	Scrutinee    IAstExpr
	Alts         []AstCaseAlt
	defaultIndex int
}

func (me *AstExprCase) Description() string { return "'case' expression" }

func (me *AstExprCase) Default() *AstCaseAlt {
	if me.defaultIndex < 0 {
		return nil
	}
	return &me.Alts[me.defaultIndex]
}

type AstCaseAlt struct {
	AstBaseTokens
	Conds       []IAstExpr
	Body        IAstExpr
	IsShortForm bool
}

type IAstTypeExpr interface {
	IAstExpr
	TypeExprBase() *AstTypeExprBase
}

type AstTypeExprBase struct {
	AstExprBase
	Meta []IAstExpr
}

func (me *AstTypeExprBase) TypeExprBase() *AstTypeExprBase { return me }

type AstTypeExprIdent struct {
	AstTypeExprBase
	AstBaseComments
}

func (me *AstTypeExprIdent) Description() string { return "'ident' type expression" }
func (me *AstTypeExprIdent) Val() string         { return me.Tokens[0].Str }
func (me *AstTypeExprIdent) IsOpish() bool       { return me.Tokens[0].Kind() == udevlex.TOKEN_OPISH }
func (me *AstTypeExprIdent) BeginsUpper() bool   { return ustr.BeginsUpper(me.Tokens[0].Str) }
func (me *AstTypeExprIdent) BeginsLower() bool   { return ustr.BeginsLower(me.Tokens[0].Str) }

type AstTypeExprAppl struct {
	AstTypeExprBase
	Callee IAstExpr
	Args   []IAstExpr
}

func (me *AstTypeExprAppl) Description() string { return "'composite' type expression" }

type AstTypeExprRec struct {
	AstTypeExprBase
	Names []AstIdent
	Exprs []IAstTypeExpr
}

func (me *AstTypeExprRec) Description() string { return "'record' type expression" }
