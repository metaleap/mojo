package atmoil

import (
	"github.com/go-leap/dev/lex"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang"
)

type IAstNode interface {
	Facts() *AnnFactAll
	Print() atmolang.IAstNode
	Origin() atmolang.IAstNode
	origToks() udevlex.Tokens
	EquivTo(IAstNode) bool
	findByOrig(IAstNode, atmolang.IAstNode) []IAstNode
	IsDef() *AstDef
	IsDefWithArg() bool
	Let() *AstExprLetBase
	RefersTo(string) bool
	refsTo(string) []IAstExpr
	walk(ancestors []IAstNode, self IAstNode, on func([]IAstNode, IAstNode, ...IAstNode) bool)
}

type IAstExpr interface {
	IAstNode
	IsAtomic() bool
	astExprBase() *AstExprBase
}

type astNodeBase struct {
	facts AnnFactAll
}

func (me *astNodeBase) Facts() *AnnFactAll { return &me.facts }
func (*astNodeBase) Let() *AstExprLetBase  { return nil }
func (*astNodeBase) IsDef() *AstDef        { return nil }
func (*astNodeBase) IsDefWithArg() bool    { return false }

type AstDef struct {
	astNodeBase
	OrigDef *atmolang.AstDef

	Name AstIdentDecl
	Arg  *AstDefArg
	Body IAstExpr
}

func (me *AstDef) findByOrig(self IAstNode, orig atmolang.IAstNode) (nodes []IAstNode) {
	if orig == me.OrigDef {
		nodes = []IAstNode{self}
	} else if nodes = me.Name.findByOrig(&me.Name, orig); len(nodes) > 0 {
		nodes = append(nodes, self)
	} else if nodes = me.Body.findByOrig(me.Body, orig); len(nodes) > 0 {
		nodes = append(nodes, self)
	} else if me.Arg != nil {
		if nodes = me.Arg.findByOrig(me.Arg, orig); len(nodes) > 0 {
			nodes = append(nodes, self)
		}
	}
	return
}
func (me *AstDef) IsDef() *AstDef            { return me }
func (me *AstDef) IsDefWithArg() bool        { return me.Arg != nil }
func (me *AstDef) Origin() atmolang.IAstNode { return me.OrigDef }
func (me *AstDef) origToks() (toks udevlex.Tokens) {
	if me.OrigDef != nil && me.OrigDef.Tokens != nil {
		toks = me.OrigDef.Tokens
	} else if toks = me.Name.origToks(); len(toks) == 0 {
		if me.Body != nil {
			toks = me.Body.origToks()
		}
		if len(toks) == 0 && me.Arg != nil {
			toks = me.Arg.origToks()
		}
	}
	return
}
func (me *AstDef) RefersTo(name string) bool     { return me.Body.RefersTo(name) }
func (me *AstDef) refsTo(name string) []IAstExpr { return me.Body.refsTo(name) }
func (me *AstDef) EquivTo(node IAstNode) bool {
	cmp, _ := node.(*AstDef)
	return cmp != nil && cmp.Name.Val == me.Name.Val && cmp.Body.EquivTo(me.Body) &&
		((me.Arg == nil) == (cmp.Arg == nil)) && ((me.Arg == nil) || me.Arg.EquivTo(cmp.Arg))
}
func (me *AstDef) walk(ancestors []IAstNode, self IAstNode, on func([]IAstNode, IAstNode, ...IAstNode) bool) {
	if on(ancestors, self, &me.Name, me.Arg, me.Body) {
		ancestors = append(ancestors, self)
		me.Name.walk(ancestors, &me.Name, on)
		if me.Arg != nil {
			me.Arg.walk(ancestors, me.Arg, on)
		}
		if me.Body != nil {
			me.Body.walk(ancestors, me.Body, on)
		}
	}
}

type AstDefTop struct {
	AstDef

	Id                string
	OrigTopLevelChunk *atmolang.SrcTopChunk
	Errs              struct {
		Stage0Init     atmo.Errors
		Stage1BadNames atmo.Errors
	}

	refersTo map[string]bool
}

func (me *AstDefTop) Errors() (errs atmo.Errors) {
	errs = make(atmo.Errors, 0, len(me.Errs.Stage0Init)+len(me.Errs.Stage1BadNames))
	errs = append(append(errs, me.Errs.Stage0Init...), me.Errs.Stage1BadNames...)
	return
}
func (me *AstDefTop) FindByOrig(orig atmolang.IAstNode) []IAstNode {
	return me.AstDef.findByOrig(me, orig)
}
func (me *AstDefTop) FindDescendants(traverseIntoMatchesToo bool, max int, pred func(IAstNode) bool) (paths [][]IAstNode) {
	me.Walk(func(curnodeancestors []IAstNode, curnode IAstNode, curnodedescendants ...IAstNode) bool {
		if pred(curnode) {
			paths = append(paths, append(curnodeancestors, curnode))
			return traverseIntoMatchesToo
		}
		return max <= 0 || len(paths) < max
	})
	return
}
func (me *AstDefTop) OrigToks(node IAstNode) (toks udevlex.Tokens) {
	if toks = node.origToks(); len(toks) == 0 {
		if paths := me.FindDescendants(false, 1, func(n IAstNode) bool { return n == node }); len(paths) == 1 {
			for i := len(paths[0]) - 1; i >= 0; i-- {
				if toks = paths[0][i].origToks(); len(toks) > 0 {
					break
				}
			}
		}
	}
	return
}
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
func (me *AstDefTop) RefsTo(name string) (refs []IAstExpr) {
	for len(name) > 0 && name[0] == '_' {
		name = name[1:]
	}
	if len(name) > 0 {
		// leverage the bool cache already in place two ways, though we dont cache the occurrences
		// in detail (they're usually for editor or error-message scenarios, not hi-perf paths)
		if refersto, known := me.refersTo[name]; refersto || !known {
			if refs = me.AstDef.refsTo(name); !known {
				me.refersTo[name] = (len(refs) > 0)
			}
		}
	}
	return
}
func (me *AstDefTop) Walk(shouldTraverse func(curNodeAncestors []IAstNode, curNode IAstNode, curNodeDescendantsThatWillBeTraversedIfReturningTrue ...IAstNode) bool) {
	me.walk(nil, me, shouldTraverse)
}

type AstDefArg struct {
	AstIdentDecl
	Orig *atmolang.AstDefArg
}

func (me *AstDefArg) EquivTo(node IAstNode) bool {
	cmp, _ := node.(*AstDefArg)
	return ((me == nil) == (cmp == nil)) && (me == nil || me.Val == cmp.Val)
}
func (me *AstDefArg) findByOrig(_ IAstNode, orig atmolang.IAstNode) (nodes []IAstNode) {
	if me.Orig == orig {
		nodes = []IAstNode{me}
	} else if nodes = me.AstIdentDecl.findByOrig(&me.AstIdentDecl, orig); len(nodes) > 0 {
		nodes = append(nodes, me)
	}
	return
}
func (me *AstDefArg) origToks() udevlex.Tokens {
	if me.Orig != nil && me.Orig.Tokens != nil {
		return me.Orig.Tokens
	}
	return me.AstIdentDecl.origToks()
}
func (me *AstDefArg) Origin() atmolang.IAstNode {
	if me.Orig != nil {
		return me.Orig
	}
	return me.AstIdentDecl.Origin()
}
func (me *AstDefArg) walk(ancestors []IAstNode, self IAstNode, on func([]IAstNode, IAstNode, ...IAstNode) bool) {
	if on(ancestors, me, &me.AstIdentDecl) {
		ancestors = append(ancestors, me)
		me.AstIdentDecl.walk(ancestors, &me.AstIdentDecl, on)
	}
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
func (me *AstExprBase) origToks() udevlex.Tokens {
	if me.Orig != nil {
		return me.Orig.Toks()
	}
	return nil
}

type AstExprAtomBase struct {
	AstExprBase
}

func (me *AstExprAtomBase) findByOrig(self IAstNode, orig atmolang.IAstNode) (nodes []IAstNode) {
	if me.Orig == orig {
		nodes = []IAstNode{self}
	}
	return
}
func (me *AstExprAtomBase) IsAtomic() bool           { return true }
func (me *AstExprAtomBase) RefersTo(string) bool     { return false }
func (me *AstExprAtomBase) refsTo(string) []IAstExpr { return nil }
func (me *AstExprAtomBase) walk(ancestors []IAstNode, self IAstNode, on func([]IAstNode, IAstNode, ...IAstNode) bool) {
	_ = on(ancestors, self)
}

type AstLitBase struct {
	AstExprAtomBase
}

func (me *AstLitBase) refsTo(self IAstExpr, s string) []IAstExpr {
	if atom, _ := me.Orig.(atmolang.IAstExprAtomic); atom != nil && atom.String() == s {
		return []IAstExpr{self}
	}
	return nil
}

type AstLitRune struct {
	AstLitBase
	Val rune
}

func (me *AstLitRune) EquivTo(node IAstNode) bool {
	cmp, _ := node.(*AstLitRune)
	return cmp != nil && cmp.Val == me.Val
}
func (me *AstLitRune) refsTo(s string) []IAstExpr { return me.AstLitBase.refsTo(me, s) }
func (me *AstLitRune) walk(ancestors []IAstNode, self IAstNode, on func([]IAstNode, IAstNode, ...IAstNode) bool) {
	me.AstExprAtomBase.walk(ancestors, me, on)
}

type AstLitStr struct {
	AstLitBase
	Val string
}

func (me *AstLitStr) EquivTo(node IAstNode) bool {
	cmp, _ := node.(*AstLitStr)
	return cmp != nil && cmp.Val == me.Val
}
func (me *AstLitStr) refsTo(s string) []IAstExpr { return me.AstLitBase.refsTo(me, s) }
func (me *AstLitStr) walk(ancestors []IAstNode, self IAstNode, on func([]IAstNode, IAstNode, ...IAstNode) bool) {
	me.AstExprAtomBase.walk(ancestors, me, on)
}

type AstLitUint struct {
	AstLitBase
	Val uint64
}

func (me *AstLitUint) EquivTo(node IAstNode) bool {
	cmp, _ := node.(*AstLitUint)
	return cmp != nil && cmp.Val == me.Val
}
func (me *AstLitUint) refsTo(s string) []IAstExpr { return me.AstLitBase.refsTo(me, s) }
func (me *AstLitUint) walk(ancestors []IAstNode, self IAstNode, on func([]IAstNode, IAstNode, ...IAstNode) bool) {
	me.AstExprAtomBase.walk(ancestors, me, on)
}

type AstLitFloat struct {
	AstLitBase
	Val float64
}

func (me *AstLitFloat) EquivTo(node IAstNode) bool {
	cmp, _ := node.(*AstLitFloat)
	return cmp != nil && cmp.Val == me.Val
}
func (me *AstLitFloat) refsTo(s string) []IAstExpr { return me.AstLitBase.refsTo(me, s) }
func (me *AstLitFloat) walk(ancestors []IAstNode, self IAstNode, on func([]IAstNode, IAstNode, ...IAstNode) bool) {
	me.AstExprAtomBase.walk(ancestors, me, on)
}

type AstExprLetBase struct {
	Defs      AstDefs
	letOrig   *atmolang.AstExprLet
	letPrefix string

	Anns struct {
		// like `AstIdentName.Anns.ResolvesTo`, contains the following `IAstNode` types:
		// *atmoil.AstDef, *atmoil.AstDefArg, *atmoil.AstDefTop, atmosess.AstDefRef
		NamesInScope AnnNamesInScope
	}
}

func (me *AstExprLetBase) astExprLetBase() *AstExprLetBase { return me }
func (me *AstExprLetBase) findByOrig(self IAstNode, orig atmolang.IAstNode) (nodes []IAstNode) {
	if me.letOrig == orig {
		nodes = []IAstNode{self}
	} else {
		for i := range me.Defs {
			if nodes = me.Defs[i].findByOrig(&me.Defs[i], orig); len(nodes) > 0 {
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
func (me *AstExprLetBase) letDefsRefsTo(name string) (refs []IAstExpr) {
	for i := range me.Defs {
		refs = append(refs, me.Defs[i].refsTo(name)...)
	}
	return
}

type AstIdentBase struct {
	AstExprAtomBase
	Val string
}

func (me *AstIdentBase) findByOrig(self IAstNode, orig atmolang.IAstNode) (nodes []IAstNode) {
	if nodes = me.AstExprAtomBase.findByOrig(self, orig); len(nodes) == 0 {
		if orig.Toks().EqLenAndOffsets(me.AstExprBase.origToks(), false) {
			nodes = []IAstNode{self}
		}
	}
	return
}

type AstUndef struct {
	AstExprAtomBase
	FromInvalidToken bool
}

func (me *AstUndef) EquivTo(node IAstNode) bool {
	cmp, _ := node.(*AstUndef)
	return (cmp == nil) == (me == nil)
}
func (me *AstUndef) findByOrig(_ IAstNode, orig atmolang.IAstNode) (nodes []IAstNode) {
	return me.AstExprAtomBase.findByOrig(me, orig)
}
func (me *AstUndef) walk(ancestors []IAstNode, self IAstNode, on func([]IAstNode, IAstNode, ...IAstNode) bool) {
	me.AstExprAtomBase.walk(ancestors, me, on)
}

type AstIdentTag struct {
	AstIdentBase
}

func (me *AstIdentTag) EquivTo(node IAstNode) bool {
	cmp, _ := node.(*AstIdentTag)
	return cmp != nil && cmp.Val == me.Val
}
func (me *AstIdentTag) refsTo(name string) (refs []IAstExpr) {
	if me.Val == name {
		refs = append(refs, me)
	}
	return
}
func (me *AstIdentTag) findByOrig(_ IAstNode, orig atmolang.IAstNode) (nodes []IAstNode) {
	return me.AstIdentBase.findByOrig(me, orig)
}
func (me *AstIdentTag) walk(ancestors []IAstNode, self IAstNode, on func([]IAstNode, IAstNode, ...IAstNode) bool) {
	me.AstExprAtomBase.walk(ancestors, me, on)
}

type AstIdentDecl struct {
	AstIdentBase
}

func (me *AstIdentDecl) EquivTo(node IAstNode) bool {
	cmp, _ := node.(*AstIdentDecl)
	return cmp != nil && cmp.Val == me.Val
}
func (me *AstIdentDecl) findByOrig(_ IAstNode, orig atmolang.IAstNode) (nodes []IAstNode) {
	return me.AstIdentBase.findByOrig(me, orig)
}
func (me *AstIdentDecl) walk(ancestors []IAstNode, self IAstNode, on func([]IAstNode, IAstNode, ...IAstNode) bool) {
	me.AstExprAtomBase.walk(ancestors, me, on)
}

type AstIdentName struct {
	AstIdentBase
	AstExprLetBase

	Anns struct {
		// like `AstExprLetBase.Anns.NamesInScope`, contains the following `IAstNode` types:
		// *atmoil.AstDef, *atmoil.AstDefArg, *atmoil.AstDefTop, atmosess.AstDefRef
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
func (me *AstIdentName) origToks() (toks udevlex.Tokens) {
	if me.letOrig != nil && me.letOrig.Tokens != nil {
		return me.letOrig.Tokens
	}
	return me.AstExprBase.origToks()
}
func (me *AstIdentName) RefersTo(name string) bool {
	return me.Val == name || me.letDefsReferTo(name)
}
func (me *AstIdentName) refsTo(name string) (refs []IAstExpr) {
	if refs = me.letDefsRefsTo(name); me.Val == name {
		refs = append(refs, me)
	}
	return
}
func (me *AstIdentName) EquivTo(node IAstNode) bool {
	cmp, _ := node.(*AstIdentName)
	return cmp != nil && cmp.Val == me.Val && cmp.letDefsEquivTo(&me.AstExprLetBase)
}
func (me *AstIdentName) findByOrig(_ IAstNode, orig atmolang.IAstNode) (nodes []IAstNode) {
	if nodes = me.AstIdentBase.findByOrig(me, orig); len(nodes) == 0 {
		if orig.Toks().EqLenAndOffsets(me.origToks(), false) {
			nodes = []IAstNode{me}
		} else {
			nodes = me.AstExprLetBase.findByOrig(me, orig)
		}
	}
	return
}
func (me *AstIdentName) walk(ancestors []IAstNode, _ IAstNode, on func([]IAstNode, IAstNode, ...IAstNode) bool) {
	trav := make([]IAstNode, len(me.Defs))
	for i := range me.Defs {
		trav[i] = &me.Defs[i]
	}
	if on(ancestors, me, trav...) {
		ancestors = append(ancestors, me)
		for i := range trav {
			trav[i].walk(ancestors, trav[i], on)
		}
	}
}

type AstAppl struct {
	AstExprBase
	AstExprLetBase
	Orig         *atmolang.AstExprAppl
	AtomicCallee IAstExpr
	AtomicArg    IAstExpr
}

func (me *AstAppl) findByOrig(_ IAstNode, orig atmolang.IAstNode) (nodes []IAstNode) {
	if me.Orig == orig {
		nodes = []IAstNode{me}
	} else {
		if nodes = me.AtomicCallee.findByOrig(me.AtomicCallee, orig); len(nodes) > 0 {
			nodes = append(nodes, me)
		} else if nodes = me.AtomicArg.findByOrig(me.AtomicArg, orig); len(nodes) > 0 {
			nodes = append(nodes, me)
		} else {
			nodes = me.AstExprLetBase.findByOrig(me, orig)
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
func (me *AstAppl) origToks() (toks udevlex.Tokens) {
	if me.letOrig != nil && me.letOrig.Tokens != nil {
		toks = me.letOrig.Tokens
	} else if me.Orig != nil && me.Orig.Tokens != nil {
		toks = me.Orig.Tokens
	} else if toks = me.AtomicCallee.origToks(); len(toks) == 0 {
		toks = me.AtomicArg.origToks()
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
func (me *AstAppl) refsTo(name string) []IAstExpr {
	return append(me.AtomicCallee.refsTo(name), append(me.AtomicArg.refsTo(name), me.letDefsRefsTo(name)...)...)
}
func (me *AstAppl) walk(ancestors []IAstNode, _ IAstNode, on func([]IAstNode, IAstNode, ...IAstNode) bool) {
	trav := make([]IAstNode, len(me.Defs), 2+len(me.Defs))
	for i := range me.Defs {
		trav[i] = &me.Defs[i]
	}
	trav = append(trav, me.AtomicCallee, me.AtomicArg)
	if on(ancestors, me, trav...) {
		ancestors = append(ancestors, me)
		for i := range trav {
			trav[i].walk(ancestors, trav[i], on)
		}
	}
}
