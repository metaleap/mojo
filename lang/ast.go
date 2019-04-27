package atmolang

import (
	"github.com/go-leap/dev/lex"
)

type IAstNode interface {
	print(*CtxPrint)
	Toks() udevlex.Tokens
}

type IAstExpr interface {
	IAstNode
	__implements_IAstExpr()
}

type IAstExprAtomic interface {
	IAstExpr
	__implements_IAstExprAtomic()
}

type AstBaseTokens struct {
	Tokens udevlex.Tokens
}

func (me *AstBaseTokens) Toks() udevlex.Tokens { return me.Tokens }

type AstTopLevel struct {
	AstBaseTokens
	Comments []AstComment
	Def      struct {
		Orig         *AstDef
		IsUnexported bool
	}
}

type AstComment struct {
	AstBaseTokens
	ContentText       string
	IsSelfTerminating bool
}

type AstDef struct {
	AstBaseTokens
	Name       AstIdent
	NameAffix  IAstExpr
	Args       []AstDefArg
	Meta       []IAstExpr
	Body       IAstExpr
	IsTopLevel bool
}

type AstDefArg struct {
	NameOrConstVal IAstExprAtomic
	Affix          IAstExpr
}

type AstExprBase struct {
	AstBaseTokens
}

func (*AstExprBase) __implements_IAstExpr() {}

type AstExprAtomBase struct {
	AstExprBase
}

func (*AstExprAtomBase) __implements_IAstExprAtomic() {}

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

type AstExprCases struct {
	AstExprBase
	Scrutinee    IAstExpr
	Alts         []AstCase
	Desugared    *AstExprLet
	defaultIndex int
}

type AstCase struct {
	Conds []IAstExpr
	Body  IAstExpr
}

func (me *AstComment) initFrom(tokens udevlex.Tokens, at int) {
	me.Tokens = tokens[at : at+1]
	me.ContentText, me.IsSelfTerminating = me.Tokens[0].Str, me.Tokens[0].IsCommentSelfTerminating()
}

func (me *AstExprCases) Default() *AstCase {
	if me.defaultIndex < 0 {
		return nil
	}
	return &me.Alts[me.defaultIndex]
}

func (me *AstExprCases) removeAltAt(idx int) {
	for i := idx; i < len(me.Alts)-1; i++ {
		me.Alts[i] = me.Alts[i+1]
	}
	me.Alts = me.Alts[:len(me.Alts)-1]
}

func (me *AstExprAppl) ToUnary() (unary *AstExprAppl) {
	/*
		callee arg0 arg1 arg2
		(callee arg0) arg1 arg2
		((callee arg0) arg1) arg2
	*/
	if unary = me; len(me.Args) > 1 {
		appl := *me
		for len(appl.Args) > 1 {
			appl.Callee = &AstExprAppl{Callee: appl.Callee, Args: appl.Args[:1]}
			appl.Args = appl.Args[1:]
		}
		unary = &appl
	}
	return
}
