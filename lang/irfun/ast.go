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
	find(IAstNode, atmolang.IAstNode) []IAstNode
	IsDef() *AstDef
	IsDefWithArg() bool
	Let() *AstExprLetBase
	RefersTo(string) bool
	RefsTo(string) []*AstIdentName
}

type IAstExpr interface {
	IAstNode
	IsAtomic() bool
	astExprBase() *AstExprBase
}

type astNodeBase struct {
}

func (*astNodeBase) Let() *AstExprLetBase { return nil }
func (*astNodeBase) IsDef() *AstDef       { return nil }
func (*astNodeBase) IsDefWithArg() bool   { return false }

type AstDef struct {
	astNodeBase
	OrigDef *atmolang.AstDef

	Name AstIdentName
	Arg  *AstDefArg
	Body IAstExpr
}

func (me *AstDef) find(self IAstNode, orig atmolang.IAstNode) (nodes []IAstNode) {
	if orig == me.OrigDef {
		nodes = []IAstNode{self}
	} else {
		if nodes = me.Name.find(&me.Name, orig); len(nodes) > 0 {
			nodes = append(nodes, self)
		} else if nodes = me.Body.find(me.Body, orig); len(nodes) > 0 {
			nodes = append(nodes, self)
		} else if me.Arg != nil {
			if nodes = me.Arg.find(me.Arg, orig); len(nodes) > 0 {
				nodes = append(nodes, self)
			}
		}
	}
	return
}
func (me *AstDef) IsDef() *AstDef            { return me }
func (me *AstDef) IsDefWithArg() bool        { return me.Arg != nil }
func (me *AstDef) Origin() atmolang.IAstNode { return me.OrigDef }
func (me *AstDef) OrigToks() (toks udevlex.Tokens) {
	if me.OrigDef != nil && me.OrigDef.Tokens != nil {
		toks = me.OrigDef.Tokens
	} else if toks = me.Body.OrigToks(); len(toks) == 0 {
		if toks = me.Name.OrigToks(); len(toks) == 0 {
			toks = me.Arg.OrigToks()
		}
	}
	return
}
func (me *AstDef) RefersTo(name string) bool          { return me.Body.RefersTo(name) }
func (me *AstDef) RefsTo(name string) []*AstIdentName { return me.Body.RefsTo(name) }
func (me *AstDef) EquivTo(node IAstNode) bool {
	cmp, _ := node.(*AstDef)
	return cmp != nil && cmp.Name.EquivTo(&me.Name) && cmp.Body.EquivTo(me.Body) &&
		((me.Arg == nil) == (cmp.Arg == nil)) && ((me.Arg == nil) || me.Arg.AstIdentName.EquivTo(&cmp.Arg.AstIdentName))
}

type AstDefTop struct {
	AstDef

	Id                string
	OrigTopLevelChunk *atmolang.AstFileTopLevelChunk
	Errs              atmo.Errors

	refersTo map[string]bool
}

func (me *AstDefTop) Find(orig atmolang.IAstNode) []IAstNode { return me.AstDef.find(me, orig) }
func (me *AstDefTop) RefersTo(name string) (refersTo bool) {
	// as long as an AstDefTop exists, it represents the same original code snippet: so any given
	// RefersTo(foo) truth will hold throughout: so we cache instead of continuously re-searching
	var known bool
	if refersTo, known = me.refersTo[name]; !known {
		refersTo = me.AstDef.RefersTo(name)
		me.refersTo[name] = refersTo
	}
	return
}
func (me *AstDefTop) RefsTo(name string) (refs []*AstIdentName) {
	// leverage the bool cache already in place two ways, though we dont cache the occurrences
	// in detail (they're usually for editor or error-message scenarios, not hi-perf paths)
	if refersto, known := me.refersTo[name]; refersto || !known {
		if refs = me.AstDef.RefsTo(name); !known {
			me.refersTo[name] = (len(refs) > 0)
		}
	}
	return
}

type AstDefArg struct {
	AstIdentName

	Orig *atmolang.AstDefArg
}

func (me *AstDefArg) find(_ IAstNode, orig atmolang.IAstNode) (nodes []IAstNode) {
	if me.Orig == orig {
		nodes = []IAstNode{me}
	} else if nodes = me.AstIdentName.find(&me.AstIdentName, orig); len(nodes) > 0 {
		nodes = append(nodes, me)
	}
	return
}
func (me *AstDefArg) OrigToks() udevlex.Tokens {
	if me.Orig != nil && me.Orig.Tokens != nil {
		return me.Orig.Tokens
	}
	return me.AstIdentName.OrigToks()
}
func (me *AstDefArg) Origin() atmolang.IAstNode {
	if me.Orig != nil {
		return me.Orig
	}
	return me.AstIdentName.Origin()
}

type AstExprBase struct {
	astNodeBase

	// some `IAstExpr`s' own `Orig` fields or `IAstNode.Origin()` implementations might
	// point to (on-the-fly dynamically computed in-memory) desugared nodes, this
	// one always points to the "real origin" node (might be identical or not)
	Orig atmolang.IAstExpr
}

func (me *AstExprBase) astExprBase() *AstExprBase { return me }
func (*AstExprBase) IsAtomic() bool               { return false }
func (me *AstExprBase) Origin() atmolang.IAstNode { return me.Orig }
func (me *AstExprBase) OrigToks() udevlex.Tokens {
	if me.Orig != nil {
		return me.Orig.Toks()
	}
	return nil
}

type AstExprAtomBase struct {
	AstExprBase
}

func (me *AstExprAtomBase) find(self IAstNode, orig atmolang.IAstNode) (nodes []IAstNode) {
	if self.Origin() == orig {
		nodes = []IAstNode{self}
	}
	return
}
func (me *AstExprAtomBase) IsAtomic() bool                { return true }
func (me *AstExprAtomBase) RefersTo(string) bool          { return false }
func (me *AstExprAtomBase) RefsTo(string) []*AstIdentName { return nil }

type AstLitBase struct {
	AstExprAtomBase
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

type AstLitUndef struct {
	AstLitBase
}

func (me *AstLitUndef) EquivTo(node IAstNode) bool {
	cmp, _ := node.(*AstLitUndef)
	return cmp != nil
}

type AstExprLetBase struct {
	Defs      AstDefs
	letOrig   *atmolang.AstExprLet
	letPrefix string

	Anns struct {
		NamesInScope AnnNamesInScope
	}
}

func (me *AstExprLetBase) astExprLetBase() *AstExprLetBase { return me }
func (me *AstExprLetBase) find(self IAstNode, orig atmolang.IAstNode) (nodes []IAstNode) {
	if me.letOrig == orig {
		nodes = []IAstNode{self}
	} else {
		for i := range me.Defs {
			if nodes = me.Defs[i].find(&me.Defs[i], orig); len(nodes) > 0 {
				nodes = append(nodes, self)
				break
			}
		}
	}
	return
}
func (me *AstExprLetBase) letDefsEquivTo(cmp *AstExprLetBase) bool {
	if len(me.Defs) == len(cmp.Defs) {
		for i := range me.Defs {
			if !me.Defs[i].EquivTo(&cmp.Defs[i]) {
				return false
			}
		}
		return true
	}
	return false
}
func (me *AstExprLetBase) letDefsReferTo(name string) (refers bool) {
	for i := range me.Defs {
		if refers = me.Defs[i].RefersTo(name); refers {
			break
		}
	}
	return
}
func (me *AstExprLetBase) letDefsRefsTo(name string) (refs []*AstIdentName) {
	for i := range me.Defs {
		refs = append(refs, me.Defs[i].RefsTo(name)...)
	}
	return
}

type AstIdentBase struct {
	AstExprAtomBase
	Val string
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

type AstIdentName struct {
	AstIdentBase
	AstExprLetBase

	Anns struct {
		ResolvesTo []IAstNode
	}
}

func (me *AstIdentName) Let() *AstExprLetBase { return &me.AstExprLetBase }
func (me *AstIdentName) Origin() atmolang.IAstNode {
	if me.letOrig != nil {
		return me.letOrig
	}
	return me.Orig
}
func (me *AstIdentName) OrigToks() (toks udevlex.Tokens) {
	if me.letOrig != nil && me.letOrig.Tokens != nil {
		return me.letOrig.Tokens
	}
	return me.AstExprBase.OrigToks()
}
func (me *AstIdentName) RefersTo(name string) bool {
	return me.Val == name || me.letDefsReferTo(name)
}
func (me *AstIdentName) RefsTo(name string) (refs []*AstIdentName) {
	if refs = me.letDefsRefsTo(name); me.Val == name {
		refs = append(refs, me)
	}
	return
}
func (me *AstIdentName) EquivTo(node IAstNode) bool {
	cmp, _ := node.(*AstIdentName)
	return cmp != nil && cmp.Val == me.Val && cmp.letDefsEquivTo(&me.AstExprLetBase)
}
func (me *AstIdentName) find(_ IAstNode, orig atmolang.IAstNode) (nodes []IAstNode) {
	if nodes = me.AstExprAtomBase.find(me, orig); len(nodes) == 0 {
		if orig.Toks().EqLenAndOffsets(me.OrigToks(), false) || orig.Toks().EqLenAndOffsets(me.AstExprBase.OrigToks(), false) { // *AstIdentName gets copied sometimes because it's not a pointer in AstDef, bounding-offsets checking is ok because callers ensure they're in the right srcfile
			nodes = []IAstNode{me}
		} else {
			nodes = me.AstExprLetBase.find(me, orig)
		}
	}
	return
}

type AstAppl struct {
	AstExprBase
	AstExprLetBase
	Orig         *atmolang.AstExprAppl
	AtomicCallee IAstExpr
	AtomicArg    IAstExpr
}

func (me *AstAppl) find(_ IAstNode, orig atmolang.IAstNode) (nodes []IAstNode) {
	if me.Orig == orig {
		nodes = []IAstNode{me}
	} else {
		if nodes = me.AtomicCallee.find(me.AtomicCallee, orig); len(nodes) > 0 {
			nodes = append(nodes, me)
		} else if nodes = me.AtomicArg.find(me.AtomicArg, orig); len(nodes) > 0 {
			nodes = append(nodes, me)
		} else {
			nodes = me.AstExprLetBase.find(me, orig)
		}
	}
	return
}
func (me *AstAppl) Origin() atmolang.IAstNode {
	if me.letOrig != nil {
		return me.letOrig
	} else if me.Orig != nil {
		return me.Orig
	}
	return me.AstExprBase.Orig
}
func (me *AstAppl) OrigToks() (toks udevlex.Tokens) {
	if me.letOrig != nil && me.letOrig.Tokens != nil {
		toks = me.letOrig.Tokens
	} else if me.Orig != nil && me.Orig.Tokens != nil {
		toks = me.Orig.Tokens
	} else if toks = me.AtomicCallee.OrigToks(); len(toks) == 0 {
		toks = me.AtomicArg.OrigToks()
	}
	return
}
func (me *AstAppl) EquivTo(node IAstNode) bool {
	cmp, _ := node.(*AstAppl)
	return cmp != nil && cmp.AtomicCallee.EquivTo(me.AtomicCallee) && cmp.AtomicArg.EquivTo(me.AtomicArg) && cmp.letDefsEquivTo(&me.AstExprLetBase)
}
func (me *AstAppl) Let() *AstExprLetBase { return &me.AstExprLetBase }
func (me *AstAppl) RefersTo(name string) bool {
	return me.AtomicCallee.RefersTo(name) || me.AtomicArg.RefersTo(name) || me.letDefsReferTo(name)
}
func (me *AstAppl) RefsTo(name string) []*AstIdentName {
	return append(me.AtomicCallee.RefsTo(name), append(me.AtomicArg.RefsTo(name), me.letDefsRefsTo(name)...)...)
}
