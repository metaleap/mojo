package atmoil

import (
	"github.com/go-leap/dev/lex"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang"
)

type IIrNode interface {
	Facts() *AnnFactAll
	Print() atmolang.IAstNode
	Origin() atmolang.IAstNode
	origToks() udevlex.Tokens
	EquivTo(IIrNode) bool
	findByOrig(IIrNode, atmolang.IAstNode) []IIrNode
	IsDef() *IrDef
	IsDefWithArg() bool
	Let() *IrExprLetBase
	RefersTo(string) bool
	refsTo(string) []IIrExpr
	walk(ancestors []IIrNode, self IIrNode, on func([]IIrNode, IIrNode, ...IIrNode) bool)
}

type IIrExpr interface {
	IIrNode
	IsAtomic() bool
	exprBase() *IrExprBase
}

type irNodeBase struct {
	facts AnnFactAll
}

func (me *irNodeBase) Facts() *AnnFactAll { return &me.facts }
func (*irNodeBase) Let() *IrExprLetBase   { return nil }
func (*irNodeBase) IsDef() *IrDef         { return nil }
func (*irNodeBase) IsDefWithArg() bool    { return false }

type IrDef struct {
	irNodeBase
	OrigDef *atmolang.AstDef

	Name IrIdentDecl
	Arg  *IrDefArg
	Body IIrExpr
}

func (me *IrDef) findByOrig(self IIrNode, orig atmolang.IAstNode) (nodes []IIrNode) {
	if orig == me.OrigDef {
		nodes = []IIrNode{self}
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
func (me *IrDef) IsDef() *IrDef             { return me }
func (me *IrDef) IsDefWithArg() bool        { return me.Arg != nil }
func (me *IrDef) Origin() atmolang.IAstNode { return me.OrigDef }
func (me *IrDef) origToks() (toks udevlex.Tokens) {
	if me.OrigDef != nil && me.OrigDef.Tokens != nil {
		toks = me.OrigDef.Tokens
	} else if toks = me.Name.origToks(); len(toks) == 0 {
		if toks = me.Body.origToks(); len(toks) == 0 && me.Arg != nil {
			toks = me.Arg.origToks()
		}
	}
	return
}
func (me *IrDef) RefersTo(name string) bool    { return me.Body.RefersTo(name) }
func (me *IrDef) refsTo(name string) []IIrExpr { return me.Body.refsTo(name) }
func (me *IrDef) EquivTo(node IIrNode) bool {
	cmp, _ := node.(*IrDef)
	return cmp != nil && cmp.Name.Val == me.Name.Val && me.Body.EquivTo(cmp.Body) &&
		((me.Arg == nil) == (cmp.Arg == nil)) && ((me.Arg == nil) || me.Arg.EquivTo(cmp.Arg))
}
func (me *IrDef) walk(ancestors []IIrNode, self IIrNode, on func([]IIrNode, IIrNode, ...IIrNode) bool) {
	if on(ancestors, self, &me.Name, me.Arg, me.Body) {
		ancestors = append(ancestors, self)
		me.Name.walk(ancestors, &me.Name, on)
		if me.Arg != nil {
			me.Arg.walk(ancestors, me.Arg, on)
		}
		me.Body.walk(ancestors, me.Body, on)
	}
}

type IrDefTop struct {
	IrDef

	Id                string
	OrigTopLevelChunk *atmolang.SrcTopChunk
	Errs              struct {
		Stage0Init     atmo.Errors
		Stage1BadNames atmo.Errors
	}

	refersTo map[string]bool
}

func (me *IrDefTop) Errors() (errs atmo.Errors) {
	errs = make(atmo.Errors, 0, len(me.Errs.Stage0Init)+len(me.Errs.Stage1BadNames))
	errs = append(append(errs, me.Errs.Stage0Init...), me.Errs.Stage1BadNames...)
	return
}
func (me *IrDefTop) FindByOrig(orig atmolang.IAstNode) []IIrNode {
	return me.IrDef.findByOrig(me, orig)
}
func (me *IrDefTop) FindDescendants(traverseIntoMatchesToo bool, max int, pred func(IIrNode) bool) (paths [][]IIrNode) {
	me.Walk(func(curnodeancestors []IIrNode, curnode IIrNode, curnodedescendants ...IIrNode) bool {
		if pred(curnode) {
			paths = append(paths, append(curnodeancestors, curnode))
			return traverseIntoMatchesToo
		}
		return max <= 0 || len(paths) < max
	})
	return
}
func (me *IrDefTop) OrigToks(node IIrNode) (toks udevlex.Tokens) {
	if toks = node.origToks(); len(toks) == 0 {
		if paths := me.FindDescendants(false, 1, func(n IIrNode) bool { return n == node }); len(paths) == 1 {
			for i := len(paths[0]) - 1; i >= 0; i-- {
				if toks = paths[0][i].origToks(); len(toks) > 0 {
					break
				}
			}
		}
	}
	return
}
func (me *IrDefTop) RefersTo(name string) (refersTo bool) {
	// as long as an IrDefTop exists, it represents the same original code snippet: so any given
	// RefersTo(foo) truth will hold throughout: so we cache instead of continuously re-searching
	var known bool
	if refersTo, known = me.refersTo[name]; !known {
		refersTo = me.IrDef.RefersTo(name)
		me.refersTo[name] = refersTo
	}
	return
}
func (me *IrDefTop) RefsTo(name string) (refs []IIrExpr) {
	for len(name) > 0 && name[0] == '_' {
		name = name[1:]
	}
	if len(name) > 0 {
		// leverage the bool cache already in place two ways, though we dont cache the occurrences
		// in detail (they're usually for editor or error-message scenarios, not hi-perf paths)
		if refersto, known := me.refersTo[name]; refersto || !known {
			if refs = me.IrDef.refsTo(name); !known {
				me.refersTo[name] = (len(refs) > 0)
			}
		}
	}
	return
}
func (me *IrDefTop) Walk(shouldTraverse func(curNodeAncestors []IIrNode, curNode IIrNode, curNodeDescendantsThatWillBeTraversedIfReturningTrue ...IIrNode) bool) {
	me.walk(nil, me, shouldTraverse)
}

type IrDefArg struct {
	IrIdentDecl
	Orig *atmolang.AstDefArg
}

func (me *IrDefArg) EquivTo(node IIrNode) bool {
	cmp, _ := node.(*IrDefArg)
	return ((me == nil) == (cmp == nil)) && (me == nil || me.Val == cmp.Val)
}
func (me *IrDefArg) findByOrig(_ IIrNode, orig atmolang.IAstNode) (nodes []IIrNode) {
	if me.Orig == orig {
		nodes = []IIrNode{me}
	} else if nodes = me.IrIdentDecl.findByOrig(&me.IrIdentDecl, orig); len(nodes) > 0 {
		nodes = append(nodes, me)
	}
	return
}
func (me *IrDefArg) origToks() udevlex.Tokens {
	if me.Orig != nil && me.Orig.Tokens != nil {
		return me.Orig.Tokens
	}
	return me.IrIdentDecl.origToks()
}
func (me *IrDefArg) Origin() atmolang.IAstNode {
	if me.Orig != nil {
		return me.Orig
	}
	return me.IrIdentDecl.Origin()
}
func (me *IrDefArg) walk(ancestors []IIrNode, self IIrNode, on func([]IIrNode, IIrNode, ...IIrNode) bool) {
	if on(ancestors, me, &me.IrIdentDecl) {
		ancestors = append(ancestors, me)
		me.IrIdentDecl.walk(ancestors, &me.IrIdentDecl, on)
	}
}

type IrExprBase struct {
	irNodeBase

	// some `IIrExpr`s' own `Orig` fields or `INode.Origin()` implementations might
	// point to (on-the-fly dynamically computed in-memory) desugared nodes, this
	// one always points to the "real origin" node (might be identical or not)
	Orig atmolang.IAstExpr
}

func (me *IrExprBase) exprBase() *IrExprBase     { return me }
func (*IrExprBase) IsAtomic() bool               { return false }
func (me *IrExprBase) Origin() atmolang.IAstNode { return me.Orig }
func (me *IrExprBase) origToks() udevlex.Tokens {
	if me.Orig != nil {
		return me.Orig.Toks()
	}
	return nil
}

type IrExprAtomBase struct {
	IrExprBase
}

func (me *IrExprAtomBase) findByOrig(self IIrNode, orig atmolang.IAstNode) (nodes []IIrNode) {
	if me.Orig == orig {
		nodes = []IIrNode{self}
	}
	return
}
func (me *IrExprAtomBase) IsAtomic() bool          { return true }
func (me *IrExprAtomBase) RefersTo(string) bool    { return false }
func (me *IrExprAtomBase) refsTo(string) []IIrExpr { return nil }
func (me *IrExprAtomBase) walk(ancestors []IIrNode, self IIrNode, on func([]IIrNode, IIrNode, ...IIrNode) bool) {
	_ = on(ancestors, self)
}

type irLitBase struct {
	IrExprAtomBase
}

func (me *irLitBase) refsTo(self IIrExpr, s string) []IIrExpr {
	if atom, _ := me.Orig.(atmolang.IAstExprAtomic); atom != nil && atom.String() == s {
		return []IIrExpr{self}
	}
	return nil
}

type IrLitRune struct {
	irLitBase
	Val rune
}

func (me *IrLitRune) EquivTo(node IIrNode) bool {
	cmp, _ := node.(*IrLitRune)
	return cmp != nil && cmp.Val == me.Val
}
func (me *IrLitRune) refsTo(s string) []IIrExpr { return me.irLitBase.refsTo(me, s) }
func (me *IrLitRune) walk(ancestors []IIrNode, self IIrNode, on func([]IIrNode, IIrNode, ...IIrNode) bool) {
	me.IrExprAtomBase.walk(ancestors, me, on)
}

type IrLitStr struct {
	irLitBase
	Val string
}

func (me *IrLitStr) EquivTo(node IIrNode) bool {
	cmp, _ := node.(*IrLitStr)
	return cmp != nil && cmp.Val == me.Val
}
func (me *IrLitStr) refsTo(s string) []IIrExpr { return me.irLitBase.refsTo(me, s) }
func (me *IrLitStr) walk(ancestors []IIrNode, self IIrNode, on func([]IIrNode, IIrNode, ...IIrNode) bool) {
	me.IrExprAtomBase.walk(ancestors, me, on)
}

type IrLitUint struct {
	irLitBase
	Val uint64
}

func (me *IrLitUint) EquivTo(node IIrNode) bool {
	cmp, _ := node.(*IrLitUint)
	return cmp != nil && cmp.Val == me.Val
}
func (me *IrLitUint) refsTo(s string) []IIrExpr { return me.irLitBase.refsTo(me, s) }
func (me *IrLitUint) walk(ancestors []IIrNode, self IIrNode, on func([]IIrNode, IIrNode, ...IIrNode) bool) {
	me.IrExprAtomBase.walk(ancestors, me, on)
}

type IrLitFloat struct {
	irLitBase
	Val float64
}

func (me *IrLitFloat) EquivTo(node IIrNode) bool {
	cmp, _ := node.(*IrLitFloat)
	return cmp != nil && cmp.Val == me.Val
}
func (me *IrLitFloat) refsTo(s string) []IIrExpr { return me.irLitBase.refsTo(me, s) }
func (me *IrLitFloat) walk(ancestors []IIrNode, self IIrNode, on func([]IIrNode, IIrNode, ...IIrNode) bool) {
	me.IrExprAtomBase.walk(ancestors, me, on)
}

type IrExprLetBase struct {
	Defs      IrDefs
	letOrig   *atmolang.AstExprLet
	letPrefix string

	Anns struct {
		// like `IrIdentName.Anns.Candidates`, contains the following `INode` types:
		// *atmoil.IrDef, *atmoil.IrDefArg, *atmoil.IrDefTop, atmosess.IrDefRef
		NamesInScope AnnNamesInScope
	}
}

func (me *IrExprLetBase) findByOrig(self IIrNode, orig atmolang.IAstNode) (nodes []IIrNode) {
	if me.letOrig == orig {
		nodes = []IIrNode{self}
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
func (me *IrExprLetBase) letDefsEquivTo(cmp *IrExprLetBase) bool {
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
func (me *IrExprLetBase) letDefsReferTo(name string) (refers bool) {
	for i := range me.Defs {
		if refers = me.Defs[i].RefersTo(name); refers {
			break
		}
	}
	return
}
func (me *IrExprLetBase) letDefsRefsTo(name string) (refs []IIrExpr) {
	for i := range me.Defs {
		refs = append(refs, me.Defs[i].refsTo(name)...)
	}
	return
}

type IrIdentBase struct {
	IrExprAtomBase
	Val string
}

func (me *IrIdentBase) findByOrig(self IIrNode, orig atmolang.IAstNode) (nodes []IIrNode) {
	if nodes = me.IrExprAtomBase.findByOrig(self, orig); len(nodes) == 0 {
		if orig.Toks().EqLenAndOffsets(me.IrExprBase.origToks(), false) {
			nodes = []IIrNode{self}
		}
	}
	return
}

type IrSpecial struct {
	IrExprAtomBase
	OneOf struct {
		InvalidToken         bool
		Undefined            bool
		DefArgfulButBodyless bool
	}
}

func (me *IrSpecial) EquivTo(node IIrNode) bool {
	cmp, _ := node.(*IrSpecial)
	return (cmp == nil) == (me == nil) && ((me == nil) || me.OneOf == cmp.OneOf)
}
func (me *IrSpecial) findByOrig(_ IIrNode, orig atmolang.IAstNode) (nodes []IIrNode) {
	return me.IrExprAtomBase.findByOrig(me, orig)
}
func (me *IrSpecial) walk(ancestors []IIrNode, self IIrNode, on func([]IIrNode, IIrNode, ...IIrNode) bool) {
	me.IrExprAtomBase.walk(ancestors, me, on)
}

type IrIdentTag struct {
	IrIdentBase
}

func (me *IrIdentTag) EquivTo(node IIrNode) bool {
	cmp, _ := node.(*IrIdentTag)
	return cmp != nil && cmp.Val == me.Val
}
func (me *IrIdentTag) refsTo(name string) (refs []IIrExpr) {
	if me.Val == name {
		refs = append(refs, me)
	}
	return
}
func (me *IrIdentTag) findByOrig(_ IIrNode, orig atmolang.IAstNode) (nodes []IIrNode) {
	return me.IrIdentBase.findByOrig(me, orig)
}
func (me *IrIdentTag) walk(ancestors []IIrNode, self IIrNode, on func([]IIrNode, IIrNode, ...IIrNode) bool) {
	me.IrExprAtomBase.walk(ancestors, me, on)
}

type IrIdentDecl struct {
	IrIdentBase
}

func (me *IrIdentDecl) EquivTo(node IIrNode) bool {
	cmp, _ := node.(*IrIdentDecl)
	return cmp != nil && cmp.Val == me.Val
}
func (me *IrIdentDecl) findByOrig(_ IIrNode, orig atmolang.IAstNode) (nodes []IIrNode) {
	return me.IrIdentBase.findByOrig(me, orig)
}
func (me *IrIdentDecl) walk(ancestors []IIrNode, self IIrNode, on func([]IIrNode, IIrNode, ...IIrNode) bool) {
	me.IrExprAtomBase.walk(ancestors, me, on)
}

type IrIdentName struct {
	IrIdentBase
	IrExprLetBase

	Anns struct {
		// like `IrExprLetBase.Anns.NamesInScope`, contains the following `IIrNode` types:
		// *atmoil.IrDef, *atmoil.IrDefArg, *atmoil.IrDefTop, atmosess.IrDefRef
		Candidates []IIrNode
	}
}

func (me *IrIdentName) Let() *IrExprLetBase { return &me.IrExprLetBase }
func (me *IrIdentName) Origin() atmolang.IAstNode {
	if me.letOrig != nil {
		return me.letOrig
	}
	return me.Orig
}
func (me *IrIdentName) origToks() (toks udevlex.Tokens) {
	if me.letOrig != nil && me.letOrig.Tokens != nil {
		return me.letOrig.Tokens
	}
	return me.IrExprBase.origToks()
}
func (me *IrIdentName) RefersTo(name string) bool {
	return me.Val == name || me.letDefsReferTo(name)
}
func (me *IrIdentName) refsTo(name string) (refs []IIrExpr) {
	if refs = me.letDefsRefsTo(name); me.Val == name {
		refs = append(refs, me)
	}
	return
}
func (me *IrIdentName) EquivTo(node IIrNode) bool {
	cmp, _ := node.(*IrIdentName)
	return cmp != nil && cmp.Val == me.Val && cmp.letDefsEquivTo(&me.IrExprLetBase)
}
func (me *IrIdentName) findByOrig(_ IIrNode, orig atmolang.IAstNode) (nodes []IIrNode) {
	if nodes = me.IrIdentBase.findByOrig(me, orig); len(nodes) == 0 {
		if orig.Toks().EqLenAndOffsets(me.origToks(), false) {
			nodes = []IIrNode{me}
		} else {
			nodes = me.IrExprLetBase.findByOrig(me, orig)
		}
	}
	return
}
func (me *IrIdentName) walk(ancestors []IIrNode, _ IIrNode, on func([]IIrNode, IIrNode, ...IIrNode) bool) {
	trav := make([]IIrNode, len(me.Defs))
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

type IrAppl struct {
	IrExprBase
	IrExprLetBase
	Orig         *atmolang.AstExprAppl
	AtomicCallee IIrExpr
	AtomicArg    IIrExpr
}

func (me *IrAppl) findByOrig(_ IIrNode, orig atmolang.IAstNode) (nodes []IIrNode) {
	if me.Orig == orig {
		nodes = []IIrNode{me}
	} else {
		if nodes = me.AtomicCallee.findByOrig(me.AtomicCallee, orig); len(nodes) > 0 {
			nodes = append(nodes, me)
		} else if nodes = me.AtomicArg.findByOrig(me.AtomicArg, orig); len(nodes) > 0 {
			nodes = append(nodes, me)
		} else {
			nodes = me.IrExprLetBase.findByOrig(me, orig)
		}
	}
	return
}
func (me *IrAppl) Origin() atmolang.IAstNode {
	if me.letOrig != nil {
		return me.letOrig
	} else if me.Orig != nil {
		return me.Orig
	}
	return me.IrExprBase.Orig
}
func (me *IrAppl) origToks() (toks udevlex.Tokens) {
	if me.letOrig != nil && me.letOrig.Tokens != nil {
		toks = me.letOrig.Tokens
	} else if me.Orig != nil && me.Orig.Tokens != nil {
		toks = me.Orig.Tokens
	} else if toks = me.AtomicCallee.origToks(); len(toks) == 0 {
		toks = me.AtomicArg.origToks()
	}
	return
}
func (me *IrAppl) EquivTo(node IIrNode) bool {
	cmp, _ := node.(*IrAppl)
	return cmp != nil && cmp.AtomicCallee.EquivTo(me.AtomicCallee) && cmp.AtomicArg.EquivTo(me.AtomicArg) && cmp.letDefsEquivTo(&me.IrExprLetBase)
}
func (me *IrAppl) Let() *IrExprLetBase { return &me.IrExprLetBase }
func (me *IrAppl) RefersTo(name string) bool {
	return me.AtomicCallee.RefersTo(name) || me.AtomicArg.RefersTo(name) || me.letDefsReferTo(name)
}
func (me *IrAppl) refsTo(name string) []IIrExpr {
	return append(me.AtomicCallee.refsTo(name), append(me.AtomicArg.refsTo(name), me.letDefsRefsTo(name)...)...)
}
func (me *IrAppl) walk(ancestors []IIrNode, _ IIrNode, on func([]IIrNode, IIrNode, ...IIrNode) bool) {
	trav := make([]IIrNode, len(me.Defs), 2+len(me.Defs))
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
