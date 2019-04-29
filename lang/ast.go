package atmolang

import (
	"github.com/go-leap/dev/lex"
)

type IAstNode interface {
	print(*CtxPrint)
	Toks() udevlex.Tokens
}

type IAstComments interface {
	Comments() *astBaseComments
}

type IAstExpr interface {
	IAstNode
	IAstComments
}

type IAstExprAtomic interface {
	IAstExpr
	__implements_IAstExprAtomic()
}

type AstBaseTokens struct {
	Tokens udevlex.Tokens
}

func (me *AstBaseTokens) Toks() udevlex.Tokens { return me.Tokens }

type astBaseComments = struct {
	Leading  AstComments
	Trailing AstComments
}

type AstBaseComments struct {
	comments astBaseComments
}

func (me *AstBaseComments) Comments() *astBaseComments {
	return &me.comments
}

type AstTopLevel struct {
	AstBaseTokens
	AstBaseComments
	Def struct {
		Orig         *AstDef
		IsUnexported bool
	}
}

type AstComments []AstComment

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
	AstBaseTokens
	NameOrConstVal IAstExprAtomic
	Affix          IAstExpr
}

type AstBaseExpr struct {
	AstBaseTokens
	AstBaseComments
}

type AstBaseExprAtom struct {
	AstBaseExpr
}

func (*AstBaseExprAtom) __implements_IAstExprAtomic() {}

type AstBaseExprAtomLit struct {
	AstBaseExprAtom
}

type AstExprLitUint struct {
	AstBaseExprAtomLit
	Val uint64
}

type AstExprLitFloat struct {
	AstBaseExprAtomLit
	Val float64
}

type AstExprLitRune struct {
	AstBaseExprAtomLit
	Val rune
}

type AstExprLitStr struct {
	AstBaseExprAtomLit
	Val string
}

type AstIdent struct {
	AstBaseExprAtom
	Val     string
	IsOpish bool
	IsTag   bool
}

type AstExprLet struct {
	AstBaseExpr
	Defs []AstDef
	Body IAstExpr
}

type AstExprAppl struct {
	AstBaseExpr
	Callee IAstExpr
	Args   []IAstExpr
}

type AstExprCases struct {
	AstBaseExpr
	Scrutinee    IAstExpr
	Alts         []AstCase
	Desugared    *AstExprLet
	defaultIndex int
}

type AstCase struct {
	AstBaseTokens
	Conds []IAstExpr
	Body  IAstExpr
}

func (me *AstComments) initFrom(accumComments []udevlex.Tokens) {
	this := make(AstComments, len(accumComments))
	for i := range accumComments {
		this[i].initFrom(accumComments[i], 0)
	}
	*me = this
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

func (me *AstExprAppl) Claspish() (claspish bool) {
	if ident, ok := me.Callee.(*AstIdent); ok && ident.IsOpish {
		claspish = true
		for i := 0; claspish && i < len(me.Args); i++ {
			if i >= 2 && i%2 == 0 {
				if ident, ok = me.Args[i].(*AstIdent); !ok {
					claspish = false
				}
			} else if _, ok = me.Args[i].(IAstExprAtomic); !ok {
				claspish = false
			} else if ident, ok = me.Args[i].(*AstIdent); ok && ident.IsOpish {
				claspish = false
			}
		}
	}
	return
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
