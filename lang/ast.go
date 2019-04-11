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

type IAstNode interface {
	print(*CtxPrint)
}

type AstTopLevel struct {
	AstBaseTokens
	AstBaseComments
	Def *AstDef
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

func (me *AstDef) ensureArgsLen(l int) {
	if ol := len(me.Args); ol > l {
		me.Args = me.Args[:l]
	} else if ol < l {
		me.Args = make([]AstIdent, l)
	}
}

type IAstExpr interface {
	IAstNode
	ExprBase() *AstExprBase
	Description() string
}

type IAstIdent interface {
	IAstExpr
	IsOpish() bool
	BeginsUpper() bool
	BeginsLower() bool
}

type AstExprBase struct {
	AstBaseTokens
}

func (me *AstExprBase) ExprBase() *AstExprBase { return me }
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
	astExprAtomBase
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

type astExprAtomBase struct {
	AstBaseComments
}

type astIdent struct {
	Val     string
	IsOpish bool
}

type AstIdent struct {
	AstExprAtomBase
	astIdent
}

func (me *AstIdent) Description() string { return "'ident' expression" }
func (me *AstIdent) BeginsUpper() bool   { return ustr.BeginsUpper(me.Val) }
func (me *AstIdent) BeginsLower() bool   { return ustr.BeginsLower(me.Val) }

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
	Conds       []IAstExpr
	Body        IAstExpr
	IsShortForm bool
}
