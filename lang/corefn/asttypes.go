package atmocorefn

import (
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang"
)

type IAstNode interface {
	FromOrig() atmolang.IAstNode
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

type AstDef struct {
	AstNodeBase

	Name IAstIdent

	Orig     *atmolang.AstDef
	TopLevel *atmolang.AstFileTopLevelChunk
	Errs     atmo.Errors
}

func (me *AstDef) FromOrig() atmolang.IAstNode { return me.Orig }

type AstNodeBase struct {
}

type AstExprBase struct {
	AstNodeBase
}

type AstAtomBase struct {
	AstExprBase
}

type AstIdentBase struct {
	AstAtomBase
	Name string
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
}

type AstLitRune struct {
	AstLitBase
}

type AstLitStr struct {
	AstLitBase
}

type AstLitUint struct {
	AstLitBase
}

type AstLitFloat struct {
	AstLitBase
}
