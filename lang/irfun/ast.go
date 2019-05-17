package atmolang_irfun

import (
	"github.com/go-leap/dev/lex"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang"
)

type IAstNode interface {
	Print() atmolang.IAstNode
	Origin() atmolang.IAstNode
	OrigToks() udevlex.Tokens
	EquivTo(IAstNode) bool
	IsDefWithArg() bool
	RefersTo(string) bool
}

type IAstExpr interface {
	IAstNode
	IsAtomic() bool
}

// IAstExprWithLetDefs is implemented by `AstExprLetBase` embedders,
// that is, by `AstIdentName`, `AstAppl` and `AstCases`.
type IAstExprWithLetDefs interface {
	astExprLetBase() *AstExprLetBase
	LetDef(string) *AstDef
	LetDefs() AstDefs
	Names() []string
}

type astNodeBase struct {
}

func (me *astNodeBase) IsDefWithArg() bool { return false }

type AstDef struct {
	astNodeBase
	Orig *atmolang.AstDef

	Name AstIdentName
	Arg  *AstDefArg
	Body IAstExpr
}

func (me *AstDef) IsDefWithArg() bool        { return me.Arg != nil }
func (me *AstDef) Origin() atmolang.IAstNode { return me.Orig }
func (me *AstDef) OrigToks() (toks udevlex.Tokens) {
	if me.Orig != nil {
		toks = me.Orig.Tokens
	} else if toks = me.Name.OrigToks(); len(toks) == 0 {
		if toks = me.Arg.OrigToks(); len(toks) == 0 {
			toks = me.Body.OrigToks()
		}
	}
	return
}
func (me *AstDef) RefersTo(name string) bool { return me.Body.RefersTo(name) }
func (me *AstDef) EquivTo(node IAstNode) bool {
	cmp, _ := node.(*AstDef)
	return cmp != nil && cmp.Name.EquivTo(&me.Name) && cmp.Body.EquivTo(me.Body) &&
		((me.Arg == nil) == (cmp.Arg == nil)) && ((me.Arg == nil) || me.Arg.AstIdentName.EquivTo(&cmp.Arg.AstIdentName))
}

type AstDefTop struct {
	AstDef

	ID       string
	TopLevel *atmolang.AstFileTopLevelChunk
	Errs     atmo.Errors
}

type AstDefArg struct {
	AstIdentName

	Orig *atmolang.AstDefArg
}

func (me *AstDefArg) OrigToks() (toks udevlex.Tokens) {
	if me.Orig != nil {
		toks = me.Orig.Tokens
	}
	return me.AstIdentName.OrigToks()
}
func (me *AstDefArg) Origin() atmolang.IAstNode { return me.Orig }

type AstExprBase struct {
	astNodeBase
}

func (*AstExprBase) IsAtomic() bool { return false }

type AstExprAtomBase struct {
	AstExprBase
}

func (me *AstExprAtomBase) IsAtomic() bool       { return true }
func (me *AstExprAtomBase) RefersTo(string) bool { return false }

type AstLitBase struct {
	AstExprAtomBase
	Orig atmolang.IAstExprAtomic
}

func (me *AstLitBase) Origin() atmolang.IAstNode { return me.Orig }
func (me *AstLitBase) OrigToks() (toks udevlex.Tokens) {
	if me.Orig != nil {
		toks = me.Orig.Toks()
	}
	return
}

type AstLitRune struct {
	AstLitBase
	Val rune
}

func (me *AstLitRune) EquivTo(node IAstNode) bool {
	cmp, _ := node.(*AstLitRune)
	return cmp != nil && cmp.Val == me.Val
}

type AstLitStr struct {
	AstLitBase
	Val string
}

func (me *AstLitStr) EquivTo(node IAstNode) bool {
	cmp, _ := node.(*AstLitStr)
	return cmp != nil && cmp.Val == me.Val
}

type AstLitUint struct {
	AstLitBase
	Val uint64
}

func (me *AstLitUint) EquivTo(node IAstNode) bool {
	cmp, _ := node.(*AstLitUint)
	return cmp != nil && cmp.Val == me.Val
}

type AstLitFloat struct {
	AstLitBase
	Val float64
}

func (me *AstLitFloat) EquivTo(node IAstNode) bool {
	cmp, _ := node.(*AstLitFloat)
	return cmp != nil && cmp.Val == me.Val
}

type AstExprLetBase struct {
	letOrig   *atmolang.AstExprLet
	letDefs   AstDefs
	letPrefix string
}

func (me *AstExprLetBase) astExprLetBase() *AstExprLetBase { return me }
func (me *AstExprLetBase) Names() (names []string) {
	names = make([]string, len(me.letDefs))
	for i := range me.letDefs {
		names[i] = me.letDefs[i].Name.Val
	}
	return
}
func (me *AstExprLetBase) LetDef(name string) *AstDef {
	for i := range me.letDefs {
		if me.letDefs[i].Name.Val == name {
			return &me.letDefs[i]
		}
	}
	return nil
}
func (me *AstExprLetBase) LetDefs() AstDefs { return me.letDefs }
func (me *AstExprLetBase) letDefsEquivTo(cmp *AstExprLetBase) bool {
	if len(me.letDefs) == len(cmp.letDefs) {
		for i := range me.letDefs {
			if !me.letDefs[i].EquivTo(&cmp.letDefs[i]) {
				return false
			}
		}
		return true
	}
	return false
}
func (me *AstExprLetBase) letDefsReferTo(name string) (refers bool) {
	for i := range me.letDefs {
		if refers = me.letDefs[i].RefersTo(name); refers {
			break
		}
	}
	return
}

type AstIdentBase struct {
	AstExprAtomBase
	Val string

	Orig *atmolang.AstIdent
}

func (me *AstIdentBase) Origin() atmolang.IAstNode { return me.Orig }
func (me *AstIdentBase) OrigToks() (toks udevlex.Tokens) {
	if me.Orig != nil {
		toks = me.Orig.Tokens
	}
	return
}

type AstIdentName struct {
	AstIdentBase
	AstExprLetBase

	// "always `nil`" as far as this pkg is concerned, ie. populated and consumed from outside
	NamesInScope map[string][]IAstNode
}

func (me *AstIdentName) Origin() atmolang.IAstNode {
	if me.letOrig != nil {
		return me.letOrig
	}
	return me.Orig
}
func (me *AstIdentName) OrigToks() (toks udevlex.Tokens) {
	if me.Orig != nil {
		toks = me.Orig.Tokens
	} else if me.letOrig != nil {
		toks = me.letOrig.Tokens
	}
	return
}
func (me *AstIdentName) RefersTo(name string) bool {
	return me.Val == name || me.letDefsReferTo(name)
}
func (me *AstIdentName) EquivTo(node IAstNode) bool {
	cmp, _ := node.(*AstIdentName)
	return cmp != nil && cmp.Val == me.Val && cmp.letDefsEquivTo(&me.AstExprLetBase)
}

type AstIdentVar struct {
	AstIdentBase
}

func (me *AstIdentVar) EquivTo(node IAstNode) bool {
	cmp, _ := node.(*AstIdentVar)
	return cmp != nil && cmp.Val == me.Val
}

type AstIdentTag struct {
	AstIdentBase
}

func (me *AstIdentTag) EquivTo(node IAstNode) bool {
	cmp, _ := node.(*AstIdentTag)
	return cmp != nil && cmp.Val == me.Val
}

type AstIdentEmptyParens struct {
	AstIdentBase
}

func (me *AstIdentEmptyParens) EquivTo(node IAstNode) bool {
	cmp, _ := node.(*AstIdentEmptyParens)
	return cmp != nil
}

type AstAppl struct {
	AstExprBase
	AstExprLetBase
	Orig         *atmolang.AstExprAppl
	AtomicCallee IAstExpr
	AtomicArg    IAstExpr
}

func (me *AstAppl) Origin() atmolang.IAstNode {
	if me.letOrig != nil {
		return me.letOrig
	}
	return me.Orig
}
func (me *AstAppl) OrigToks() (toks udevlex.Tokens) {
	if me.Orig != nil {
		toks = me.Orig.Tokens
	} else if toks = me.AtomicCallee.OrigToks(); len(toks) == 0 {
		if toks = me.AtomicArg.OrigToks(); len(toks) == 0 && me.letOrig != nil {
			toks = me.letOrig.Tokens
		}
	}
	return
}
func (me *AstAppl) EquivTo(node IAstNode) bool {
	cmp, _ := node.(*AstAppl)
	return cmp != nil && cmp.AtomicCallee.EquivTo(me.AtomicCallee) && cmp.AtomicArg.EquivTo(me.AtomicArg) && cmp.letDefsEquivTo(&me.AstExprLetBase)
}
func (me *AstAppl) RefersTo(name string) bool {
	return me.AtomicCallee.RefersTo(name) || me.AtomicArg.RefersTo(name) || me.letDefsReferTo(name)
}

type AstCases struct {
	AstExprBase
	AstExprLetBase
	Orig  *atmolang.AstExprCases
	Ifs   []IAstExpr
	Thens []IAstExpr
}

func (me *AstCases) Origin() atmolang.IAstNode {
	if me.letOrig != nil {
		return me.letOrig
	}
	return me.Orig
}
func (me *AstCases) OrigToks() (toks udevlex.Tokens) {
	if me.Orig != nil {
		toks = me.Orig.Tokens
	} else if me.letOrig != nil {
		toks = me.letOrig.Tokens
	} else {
		for i := range me.Ifs {
			if toks = me.Ifs[i].OrigToks(); len(toks) > 0 {
				break
			} else if toks = me.Thens[i].OrigToks(); len(toks) > 0 {
				break
			}
		}
	}
	return
}

func (me *AstCases) EquivTo(node IAstNode) bool {
	cmp, _ := node.(*AstCases)
	if cmp != nil && len(cmp.Ifs) == len(me.Ifs) && len(cmp.Thens) == len(me.Thens) {
		for i := range cmp.Ifs {
			if !cmp.Ifs[i].EquivTo(me.Ifs[i]) {
				return false
			}
		}
		for i := range cmp.Thens {
			if !cmp.Thens[i].EquivTo(me.Thens[i]) {
				return false
			}
		}
		return cmp.letDefsEquivTo(&me.AstExprLetBase)
	}
	return false
}
func (me *AstCases) RefersTo(name string) bool {
	for i := range me.Thens {
		if me.Thens[i].RefersTo(name) || me.Ifs[i].RefersTo(name) {
			return true
		}
	}
	return me.letDefsReferTo(name)
}

func (me *AstExprLetBase) walk(on func(IAstNode)) {
	for i := range me.letDefs {
		me.letDefs[i].Walk(on)
	}
}

func (me *AstIdentName) walk(on func(IAstNode)) {
	me.AstExprLetBase.walk(on)
	on(me)
}

func (me *AstAppl) walk(on func(IAstNode)) {
	me.AstExprLetBase.walk(on)
	on(me)
	type iwalk interface{ walk(func(IAstNode)) }
	if c, _ := me.AtomicCallee.(iwalk); c != nil {
		c.walk(on)
	} else {
		on(me.AtomicCallee)
	}
	if a, _ := me.AtomicArg.(iwalk); a != nil {
		a.walk(on)
	} else {
		on(me.AtomicArg)
	}
}

func (me *AstCases) walk(on func(IAstNode)) {
	me.AstExprLetBase.walk(on)
	on(me)
	type iwalk interface{ walk(func(IAstNode)) }
	for i := range me.Ifs {
		if c, _ := me.Ifs[i].(iwalk); c != nil {
			c.walk(on)
		} else {
			on(me.Ifs[i])
		}
		if t, _ := me.Thens[i].(iwalk); t != nil {
			t.walk(on)
		} else {
			on(me.Thens[i])
		}
	}
}

func (me *AstDef) Walk(on func(IAstNode)) { me.walk(on) }
func (me *AstDef) walk(on func(IAstNode)) {
	on(me)
	type iwalk interface{ walk(func(IAstNode)) }
	if b, _ := me.Body.(iwalk); b != nil {
		b.walk(on)
	} else {
		on(me.Body)
	}
}
