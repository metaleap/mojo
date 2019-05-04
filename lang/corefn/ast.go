package atmocorefn

import (
	"strconv"

	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang"
)

type IAstNode interface {
	Origin() atmolang.IAstNode
	Print() atmolang.IAstNode

	renameIdents(map[string]string)
	refersTo(string) bool
}

type IAstExpr interface {
	IAstNode
	DynName() string
	IsAtomic() bool
}

type IAstExprAtomic interface {
	IAstExpr
	__implements_IAstExprAtomic()
}

type astNodeBase struct {
}

type AstDefBase struct {
	astNodeBase
	Orig *atmolang.AstDef

	Name AstIdentName
	Args []AstDefArg
	Body IAstExpr

	nameCoerceFunc IAstExpr
}

func (me *AstDefBase) Origin() atmolang.IAstNode { return me.Orig }
func (me *AstDefBase) refersTo(name string) bool { return me.Body.refersTo(name) }

type AstDef struct {
	AstDefBase

	TopLevel *atmolang.AstFileTopLevelChunk
	Errors   atmo.Errors
}

type AstDefArg struct {
	AstIdentName
	coerceValue IAstExprAtomic
	coerceFunc  IAstExpr

	Orig *atmolang.AstDefArg
}

type AstExprBase struct {
	astNodeBase
}

func (*AstExprBase) IsAtomic() bool { return false }

type AstExprAtomBase struct {
	AstExprBase
}

func (*AstExprAtomBase) IsAtomic() bool                    { return true }
func (me *AstExprAtomBase) renameIdents(map[string]string) {}
func (*AstExprAtomBase) __implements_IAstExprAtomic()      {}

type AstIdentBase struct {
	AstExprAtomBase
	Val string

	Orig *atmolang.AstIdent
}

func (me *AstIdentBase) refersTo(name string) bool { return name == me.Val }
func (me *AstIdentBase) Origin() atmolang.IAstNode { return me.Orig }
func (me *AstIdentBase) DynName() string           { return me.Val }

type AstIdentName struct {
	AstIdentBase
}

func (me *AstIdentName) renameIdents(ren map[string]string) {
	if nu, ok := ren[me.Val]; ok {
		me.Val = nu
	}
}

type AstIdentVar struct {
	AstIdentBase
}

func (me *AstIdentVar) DynName() string { return "˘" + me.Val }

type AstIdentTag struct {
	AstIdentBase
}

type AstIdentEmptyParens struct {
	AstIdentBase
}

type AstIdentPlaceholder struct {
	AstIdentBase
}

type AstLitBase struct {
	AstExprAtomBase
	Orig atmolang.IAstExprAtomic
}

func (me *AstLitBase) Origin() atmolang.IAstNode { return me.Orig }
func (me *AstLitBase) refersTo(string) bool      { return false }

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

func (me *AstIdentPlaceholder) Num() int { return len(me.Val) }

type AstAppl struct {
	AstExprBase
	Orig   *atmolang.AstExprAppl
	Callee IAstExprAtomic
	Arg    IAstExprAtomic
}

func (me *AstAppl) Origin() atmolang.IAstNode { return me.Orig }
func (me *AstAppl) DynName() string           { return me.Callee.DynName() + "─" + me.Arg.DynName() }
func (me *AstAppl) renameIdents(ren map[string]string) {
	me.Callee.renameIdents(ren)
	me.Arg.renameIdents(ren)
}
func (me *AstAppl) refersTo(name string) bool {
	return me.Callee.refersTo(name) || me.Arg.refersTo(name)
}

type AstLet struct {
	AstExprBase
	Orig   *atmolang.AstExprLet
	Body   IAstExpr
	Defs   astDefs
	prefix string
}

func (me *AstLet) Origin() atmolang.IAstNode { return me.Orig }
func (me *AstLet) DynName() string           { return me.prefix + "└" }
func (me *AstLet) renameIdents(ren map[string]string) {
	me.Body.renameIdents(ren)
	for i := range me.Defs {
		me.Defs[i].Body.renameIdents(ren)
	}
}
func (me *AstLet) refersTo(name string) (refers bool) {
	if refers = me.Body.refersTo(name); !refers {
		for i := range me.Defs {
			if refers = me.Defs[i].refersTo(name); refers {
				break
			}
		}
	}
	return
}

type AstCases struct {
	AstExprBase
	Orig  *atmolang.AstExprCases
	Ifs   []IAstExpr
	Thens []IAstExpr
}

func (me *AstCases) Origin() atmolang.IAstNode { return me.Orig }
func (me *AstCases) DynName() string           { panic(me.Orig) }
func (me *AstCases) renameIdents(ren map[string]string) {
	for i := range me.Thens {
		me.Thens[i].renameIdents(ren)
		me.Ifs[i].renameIdents(ren)
	}
}
func (me *AstCases) refersTo(name string) bool {
	for i := range me.Thens {
		if me.Thens[i].refersTo(name) {
			return true
		}
		if me.Ifs[i].refersTo(name) {
			return true
		}
	}
	return false
}
