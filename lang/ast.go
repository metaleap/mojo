package atmolang

import (
	"github.com/go-leap/dev/lex"
)

type IAstNode interface {
	print(*CtxPrint)
	BaseTokens() *AstBaseTokens
}

type IAstExpr interface {
	IAstNode
	IsAtomic() bool
}

type IAstExprAtomic interface {
	IAstExpr
	__implements_IAstExprAtom()
}

type AstBaseComments struct {
	Comments []AstComment
}

type AstBaseTokens struct {
	Tokens udevlex.Tokens
}

// to implement IAstNode
func (me *AstBaseTokens) BaseTokens() *AstBaseTokens { return me }

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
	NameOrConstVal IAstExprAtomic
	Affix          IAstExpr
}
type AstExprBase struct {
	AstBaseTokens
}

func (*AstExprBase) IsAtomic() bool { return false }

type AstExprAtomBase struct {
	AstExprBase
	AstBaseComments
}

func (*AstExprAtomBase) IsAtomic() bool             { return true }
func (*AstExprAtomBase) __implements_IAstExprAtom() {}

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

func (me *AstComment) initFrom(tokens udevlex.Tokens, at int) {
	me.Tokens = tokens[at : at+1]
	me.ContentText, me.IsSelfTerminating = me.Tokens[0].Str, me.Tokens[0].IsCommentSelfTerminating()
}

func (me *AstExprCase) Default() *AstCaseAlt {
	if me.defaultIndex < 0 {
		return nil
	}
	return &me.Alts[me.defaultIndex]
}

func (me *AstExprCase) removeAltAt(idx int) {
	for i := idx; i < len(me.Alts)-1; i++ {
		me.Alts[i] = me.Alts[i+1]
	}
	me.Alts = me.Alts[:len(me.Alts)-1]
}
