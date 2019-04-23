package atmolang

import (
	"github.com/go-leap/dev/lex"
)

type AstBaseComments struct {
	Comments []AstComment
}

type AstBaseTokens struct {
	Tokens udevlex.Tokens
}

// to implement IAstNode
func (me *AstBaseTokens) BaseTokens() *AstBaseTokens { return me }

type IAstNode interface {
	print(*CtxPrint)
	BaseTokens() *AstBaseTokens
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
	Args       []AstDefArg
	Meta       []IAstExpr
	Body       IAstExpr
	IsTopLevel bool
}

type AstDefArg struct {
	NameOrConstVal IAstExprAtom
	Affix          IAstExpr
}

type IAstExpr interface {
	IAstNode
}

type IAstExprAtom interface {
	IAstExpr
	atomic()
}

type AstExprBase struct {
	AstBaseTokens
}

type AstExprAtomBase struct {
	AstExprBase
	AstBaseComments
}

func (*AstExprAtomBase) atomic() {}

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
