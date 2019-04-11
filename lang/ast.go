package atemlang

import (
	"github.com/go-leap/dev/lex"
)

type AstBaseComments struct {
	Comments []*AstComment
}

type AstBaseTokens struct {
	Tokens udevlex.Tokens
}

type IAstNode interface {
	print(*CtxPrint)
}

type AstTopLevel struct {
	AstBaseTokens
	AstBaseComments
	Def             *AstDef
	DefIsUnexported bool
}

type AstComment struct {
	AstBaseTokens
	ContentText       string
	IsSelfTerminating bool
}

type AstDef struct {
	AstBaseTokens
	Name       AstIdent
	Args       []AstIdent
	Meta       []IAstExpr
	IsTopLevel bool
	Body       IAstExpr
}

type IAstExpr interface {
	IAstNode
	ExprBase() *AstExprBase
	Description() string
}

type AstExprBase struct {
	AstBaseTokens
}

func (me *AstExprBase) ExprBase() *AstExprBase { return me }

type AstExprAtomBase struct {
	AstExprBase
	AstBaseComments
}

type AstExprLitBase struct {
	AstExprAtomBase
}

type AstExprLitUint struct {
	AstExprLitBase
	Val uint64
}

func (me *AstExprLitUint) Description() string { return "'unsigned-integer literal' expression" }

type AstExprLitFloat struct {
	AstExprLitBase
	Val float64
}

func (me *AstExprLitFloat) Description() string { return "'float literal' expression" }

type AstExprLitRune struct {
	AstExprLitBase
	Val rune
}

func (me *AstExprLitRune) Description() string { return "'rune literal' expression" }

type AstExprLitStr struct {
	AstExprLitBase
	Val string
}

func (me *AstExprLitStr) Description() string { return "'string literal' expression" }

type AstIdent struct {
	AstExprAtomBase
	Val     string
	IsOpish bool
	IsTag   bool
}

func (me *AstIdent) Description() string { return "'ident' expression" }

type AstExprLet struct {
	AstExprBase
	Defs []AstDef
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
	Conds []IAstExpr
	Body  IAstExpr
}
