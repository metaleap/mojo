package atemlang

import (
	"github.com/go-leap/dev/lex"
)

type AstBaseComments struct {
	Comments []AstComment
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
	Body       IAstExpr
	IsTopLevel bool
}

type IAstExpr interface {
	IAstNode
}

type AstExprBase struct {
	AstBaseTokens
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
	Val uint64
}

type AstExprLitFloat struct {
	AstExprLitBase
	Val float64
}

type AstExprLitRune struct {
	AstExprLitBase
	Val rune
}

type AstExprLitStr struct {
	AstExprLitBase
	Val string
}

type AstIdent struct {
	AstExprAtomBase
	Val     string
	Affix   string
	IsOpish bool
	IsTag   bool
}

type AstExprLet struct {
	AstExprBase
	Defs []AstDef
	Body IAstExpr
}

type AstExprAppl struct {
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

type AstCaseAlt struct {
	AstBaseTokens
	Conds []IAstExpr
	Body  IAstExpr
}
