package atmocorefn

import (
	"strconv"

	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang"
)

type IAstNode interface {
	Origin() atmolang.IAstNode
	Print() atmolang.IAstNode
}

type IAstExpr interface {
	IAstNode
	DynName() string
}

type IAstExprAtomic interface {
	IAstExpr
	__implements_IAstExprAtomic()
}

type IAstIdent interface {
	IAstExprAtomic
	String() string
}

type AstNodeBase struct {
}

type AstDefBase struct {
	AstNodeBase
	Orig *atmolang.AstDef

	Name IAstIdent
	Args []AstDefArg
	Body IAstExpr
}

type AstDef struct {
	AstDefBase
	Locals astDefs

	TopLevel *atmolang.AstFileTopLevelChunk
	Errs     atmo.Errors

	b     AstBuilder
	state struct {
		counter int
	}
}

func (me *AstDef) Origin() atmolang.IAstNode { return me.Orig }

type AstDefArg struct {
	AstIdentName
	coerceValue IAstExprAtomic

	Orig *atmolang.AstDefArg
}

type AstExprBase struct {
	AstNodeBase
}

type AstExprAtomBase struct {
	AstExprBase
}

func (*AstExprAtomBase) __implements_IAstExprAtomic() {}

type AstIdentBase struct {
	AstExprAtomBase
	Val string

	Orig *atmolang.AstIdent
}

func (me *AstIdentBase) Origin() atmolang.IAstNode { return me.Orig }
func (me *AstIdentBase) String() string            { return me.Val }
func (me *AstIdentBase) DynName() string           { return me.Val }

type AstIdentName struct {
	AstIdentBase
}

type AstIdentVar struct {
	AstIdentBase
}

type AstIdentTag struct {
	AstIdentBase
}

type AstIdentOp struct {
	AstIdentBase
}

type AstIdentEmptyParens struct {
	AstIdentBase
}

type AstIdentUnderscores struct {
	AstIdentBase
}

type AstLitBase struct {
	AstExprAtomBase
	Orig atmolang.IAstExprAtomic
}

func (me *AstLitBase) Origin() atmolang.IAstNode {
	return me.Orig
}

type AstLitRune struct {
	AstLitBase
	Val rune
}

func (me *AstLitRune) DynName() string { return strconv.QuoteRune(me.Val) }

type AstLitStr struct {
	AstLitBase
	Val string
}

func (me *AstLitStr) DynName() string { return strconv.Quote(me.Val) }

type AstLitUint struct {
	AstLitBase
	Val uint64
}

func (me *AstLitUint) DynName() string { return strconv.FormatUint(me.Val, 10) }

type AstLitFloat struct {
	AstLitBase
	Val float64
}

func (me *AstLitFloat) DynName() string { return strconv.FormatFloat(me.Val, 'g', -1, 64) }

func (me *AstIdentUnderscores) Num() int { return len(me.Val) }

type AstAppl struct {
	AstExprBase
	Orig   *atmolang.AstExprAppl
	Callee IAstIdent
	Arg    IAstExprAtomic
}

func (me *AstAppl) Origin() atmolang.IAstNode { return me.Orig }
func (me *AstAppl) DynName() string           { return me.Callee.DynName() + "_" + me.Arg.DynName() }

type AstCases struct {
	AstExprBase
	Orig  *atmolang.AstExprCases
	Ifs   [][]IAstExpr
	Thens []IAstExpr
}

func (me *AstCases) Origin() atmolang.IAstNode { return me.Orig }
func (me *AstCases) DynName() string           { panic(me.Origin) }
