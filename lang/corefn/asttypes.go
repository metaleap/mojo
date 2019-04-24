package atmocorefn

import (
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang"
)

type IAstNode interface {
	Origin() atmolang.IAstNode
}

type IAstExpr interface {
	IAstNode
}

type IAstAtom interface {
	IAstExpr
}

type IAstIdent interface {
	IAstAtom
}

type AstNodeBase struct {
}

type AstDef struct {
	AstNodeBase

	Name IAstIdent
	Args []AstDefArg
	Meta []IAstExpr
	Body IAstExpr

	Orig     *atmolang.AstDef
	TopLevel *atmolang.AstFileTopLevelChunk
	Errs     atmo.Errors
}

type AstDefArg struct {
	NameOrConstVal IAstAtom
	Affix          IAstExpr

	Orig *atmolang.AstDefArg
}

type AstExprBase struct {
	AstNodeBase
}

type AstAtomBase struct {
	AstExprBase
}

type AstIdentBase struct {
	AstAtomBase
	Val string

	Orig *atmolang.AstIdent
}

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
	AstAtomBase
	Orig atmolang.IAstExprAtomic
}

type AstLitRune struct {
	AstLitBase
	Val rune
}

type AstLitStr struct {
	AstLitBase
	Val string
}

type AstLitUint struct {
	AstLitBase
	Val uint64
}

type AstLitFloat struct {
	AstLitBase
	Val float64
}
