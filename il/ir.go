package atmoil

import (
	"github.com/go-leap/dev/lex"
	. "github.com/metaleap/atmo"
	. "github.com/metaleap/atmo/ast"
)

const requireAtomicCalleeAndCallArg = false

func (*irNodeBase) Let() *IrExprLetBase { return nil }
func (*irNodeBase) IsDef() *IrDef       { return nil }
func (*irNodeBase) IsExt() bool         { return false }
func (me *irNodeBase) Origin() IAstNode { return me.Orig }

func (me *IrDef) findByOrig(self IIrNode, orig IAstNode) (nodes []IIrNode) {
	if orig == me.OrigDef() {
		nodes = []IIrNode{self}
	} else if nodes = me.Name.findByOrig(&me.Name, orig); len(nodes) != 0 {
		nodes = append(nodes, self)
	} else if nodes = me.Body.findByOrig(me.Body, orig); len(nodes) != 0 {
		nodes = append(nodes, self)
	}
	return
}
func (me *IrDef) IsDef() *IrDef              { return me }
func (me *IrDef) IsLam() (ifSo *IrLam)       { ifSo, _ = me.Body.(*IrLam); return }
func (me *IrDef) OrigDef() (origDef *AstDef) { origDef, _ = me.Orig.(*AstDef); return }
func (me *IrDef) origToks() (toks udevlex.Tokens) {
	if orig := me.OrigDef(); orig != nil && orig.Tokens != nil {
		toks = orig.Tokens
	} else if toks = me.Name.origToks(); len(toks) == 0 {
		toks = me.Body.origToks()
	}
	return
}
func (me *IrDef) RefersTo(name string) bool    { return me.Body.RefersTo(name) }
func (me *IrDef) refsTo(name string) []IIrExpr { return me.Body.refsTo(name) }
func (me *IrDef) EquivTo(node IIrNode) bool {
	cmp, _ := node.(*IrDef)
	return cmp != nil && cmp.Name.Val == me.Name.Val && me.Body.EquivTo(cmp.Body)
}
func (me *IrDef) walk(ancestors []IIrNode, self IIrNode, on func([]IIrNode, IIrNode, ...IIrNode) bool) (keepGoing bool) {
	if keepGoing = on(ancestors, self, &me.Name, me.Body); keepGoing {
		ancestors = append(ancestors, self)
		if keepGoing = me.Name.walk(ancestors, &me.Name, on); keepGoing {
			keepGoing = me.Body.walk(ancestors, me.Body, on)
		}
	}
	return
}

func (me *IrDefTop) HasErrors() bool {
	return len(me.Errs.Stage1AstToIr) != 0 || len(me.Errs.Stage2BadNames) != 0 || len(me.Errs.Stage3Preduce) != 0
}
func (me *IrDefTop) Errors() (errs Errors) {
	errs = make(Errors, 0, len(me.Errs.Stage1AstToIr)+len(me.Errs.Stage2BadNames)+len(me.Errs.Stage3Preduce))
	errs = append(append(append(errs, me.Errs.Stage1AstToIr...), me.Errs.Stage2BadNames...), me.Errs.Stage3Preduce...)
	return
}
func (me *IrDefTop) FindByOrig(orig IAstNode) []IIrNode {
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
				if toks = paths[0][i].origToks(); len(toks) != 0 {
					break
				}
			}
		}
	}
	return
}
func (me *IrDefTop) ForAllLocalDefs(onLocalDef func(*IrDef) (done bool)) {
	me.Walk(func(curnodeancestors []IIrNode, curnode IIrNode, curnodedescendants ...IIrNode) bool {
		let := curnode.Let()
		var done bool
		if let != nil {
			for i := range let.Defs {
				done = done || onLocalDef(&let.Defs[i])
			}
		}
		return (!done) && (let != nil || curnode.IsDef() != nil)
	})
}
func (me *IrDefTop) RefersToOrDefines(name string) (relatesTo bool) {
	relatesTo = me.Name.Val == name || me.RefersTo(name)
	if !relatesTo {
		me.ForAllLocalDefs(func(localdef *IrDef) (done bool) {
			relatesTo = relatesTo || (localdef.Name.Val == name)
			return relatesTo
		})
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
	for len(name) != 0 && name[0] == '_' {
		name = name[1:]
	}
	if len(name) != 0 {
		// leverage the bool cache already in place two ways, though we dont cache the occurrences
		// in detail (they're usually for editor or error-message scenarios, not hi-perf paths)
		if refersto, known := me.refersTo[name]; refersto || !known {
			if refs = me.IrDef.refsTo(name); !known {
				me.refersTo[name] = (len(refs) != 0)
			}
		}
	}
	return
}
func (me *IrDefTop) Walk(shouldTraverse func(curNodeAncestors []IIrNode, curNode IIrNode, curNodeDescendantsThatWillBeTraversedIfReturningTrue ...IIrNode) bool) {
	_ = me.walk(nil, me, shouldTraverse)
}
func (me *IrDefTop) HasDef(maybeDef IIrNode) (defIfLocal *IrDef) {
	if maybedef, _ := maybeDef.(*IrDef); maybedef != nil {
		me.ForAllLocalDefs(func(defLocal *IrDef) bool {
			if defIfLocal == nil && defLocal == maybedef {
				defIfLocal = defLocal
			}
			return defIfLocal == nil
		})
	}
	return
}
func (me *IrDefTop) FindAny(where func(IIrNode) bool) (firstMatch []IIrNode) {
	me.Walk(func(ancestors []IIrNode, curnode IIrNode, descendants ...IIrNode) bool {
		if where(curnode) {
			firstMatch = append(ancestors, curnode)
		}
		return firstMatch == nil
	})
	return
}
func (me *IrDefTop) FindAll(where func(IIrNode) bool) (matches [][]IIrNode) {
	me.Walk(func(ancestors []IIrNode, curnode IIrNode, descendants ...IIrNode) bool {
		if where(curnode) {
			matches = append(matches, append(ancestors, curnode))
		}
		return true
	})
	return
}
func (me *IrDefTop) HasAnyOf(nodes ...IIrNode) bool {
	return nil != me.FindAny(func(node IIrNode) bool {
		for _, n := range nodes {
			if n == node {
				return true
			}
		}
		return false
	})
}
func (me *IrDefTop) HasIdentDecl(name string) bool {
	return 0 < len(me.FindAny(func(n IIrNode) bool {
		identdecl, ok := n.(*IrIdentDecl)
		return ok && identdecl.Val == name
	}))
}

func (me *IrArg) EquivTo(node IIrNode) bool {
	cmp, _ := node.(*IrArg)
	return ((me == nil) == (cmp == nil)) && (me == nil || me.Val == cmp.Val)
}
func (me *IrArg) findByOrig(_ IIrNode, orig IAstNode) (nodes []IIrNode) {
	if me.Orig == orig {
		nodes = []IIrNode{me}
	} else if nodes = me.IrIdentDecl.findByOrig(&me.IrIdentDecl, orig); len(nodes) != 0 {
		nodes = append(nodes, me)
	}
	return
}
func (me *IrArg) origToks() udevlex.Tokens {
	if me.Orig != nil {
		if toks := me.Orig.Toks(); len(toks) != 0 {
			return toks
		}
	}
	return me.IrIdentDecl.origToks()
}
func (me *IrArg) Origin() IAstNode {
	if me.Orig != nil {
		return me.Orig
	}
	return me.IrIdentDecl.Origin()
}
func (me *IrArg) walk(ancestors []IIrNode, self IIrNode, on func([]IIrNode, IIrNode, ...IIrNode) bool) (keepGoing bool) {
	if keepGoing = on(ancestors, me, &me.IrIdentDecl); keepGoing {
		ancestors = append(ancestors, me)
		keepGoing = me.IrIdentDecl.walk(ancestors, &me.IrIdentDecl, on)
	}
	return
}

func (me *IrExprBase) exprBase() *IrExprBase { return me }
func (*IrExprBase) IsAtomic() bool           { return false }
func (me *IrExprBase) origToks() udevlex.Tokens {
	if me.Orig != nil {
		return me.Orig.Toks()
	}
	return nil
}

func (me *IrExprAtomBase) findByOrig(self IIrNode, orig IAstNode) (nodes []IIrNode) {
	if me.Orig == orig {
		nodes = []IIrNode{self}
	} else if orig.Toks().EqLenAndOffsets(me.origToks(), false) {
		nodes = []IIrNode{self}
	}
	return
}
func (me *IrExprAtomBase) IsAtomic() bool          { return true }
func (me *IrExprAtomBase) RefersTo(string) bool    { return false }
func (me *IrExprAtomBase) refsTo(string) []IIrExpr { return nil }
func (me *IrExprAtomBase) walk(ancestors []IIrNode, self IIrNode, on func([]IIrNode, IIrNode, ...IIrNode) bool) bool {
	return on(ancestors, self)
}

func (me *irLitBase) refsTo(self IIrExpr, s string) []IIrExpr {
	if atom, _ := me.Orig.(IAstExprAtomic); atom != nil && atom.String() == s {
		return []IIrExpr{self}
	}
	return nil
}

func (me *IrLitUint) EquivTo(node IIrNode) bool {
	cmp, _ := node.(*IrLitUint)
	return cmp != nil && cmp.Val == me.Val
}
func (me *IrLitUint) refsTo(s string) []IIrExpr { return me.irLitBase.refsTo(me, s) }
func (me *IrLitUint) walk(ancestors []IIrNode, self IIrNode, on func([]IIrNode, IIrNode, ...IIrNode) bool) bool {
	return me.IrExprAtomBase.walk(ancestors, me, on)
}

func (me *IrLitFloat) EquivTo(node IIrNode) bool {
	cmp, _ := node.(*IrLitFloat)
	return cmp != nil && cmp.Val == me.Val
}
func (me *IrLitFloat) refsTo(s string) []IIrExpr { return me.irLitBase.refsTo(me, s) }
func (me *IrLitFloat) walk(ancestors []IIrNode, self IIrNode, on func([]IIrNode, IIrNode, ...IIrNode) bool) bool {
	return me.IrExprAtomBase.walk(ancestors, me, on)
}

func (me *IrExprLetBase) findByOrig(self IIrNode, orig IAstNode) (nodes []IIrNode) {
	if me.letOrig == orig {
		nodes = []IIrNode{self}
	} else {
		for i := range me.Defs {
			if nodes = me.Defs[i].findByOrig(&me.Defs[i], orig); len(nodes) != 0 {
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

func (me *IrIdentBase) findByOrig(self IIrNode, orig IAstNode) (nodes []IIrNode) {
	if nodes = me.IrExprAtomBase.findByOrig(self, orig); len(nodes) == 0 {
		if orig.Toks().EqLenAndOffsets(me.IrExprBase.origToks(), false) {
			nodes = []IIrNode{self}
		}
	}
	return
}

func (me *IrNonValue) EquivTo(node IIrNode) bool {
	cmp, _ := node.(*IrNonValue)
	return (cmp == nil) == (me == nil) && ((me == nil) || me.OneOf == cmp.OneOf)
}
func (me *IrNonValue) findByOrig(_ IIrNode, orig IAstNode) (nodes []IIrNode) {
	return me.IrExprAtomBase.findByOrig(me, orig)
}
func (me *IrNonValue) walk(ancestors []IIrNode, self IIrNode, on func([]IIrNode, IIrNode, ...IIrNode) bool) bool {
	return me.IrExprAtomBase.walk(ancestors, me, on)
}

func (me *IrLitTag) EquivTo(node IIrNode) bool {
	cmp, _ := node.(*IrLitTag)
	return cmp != nil && cmp.Val == me.Val
}
func (me *IrLitTag) refsTo(name string) (refs []IIrExpr) {
	if me.Val == name {
		refs = append(refs, me)
	}
	return
}

func (me *IrIdentDecl) EquivTo(node IIrNode) bool {
	cmp, _ := node.(*IrIdentDecl)
	return cmp != nil && cmp.Val == me.Val
}
func (me *IrIdentDecl) findByOrig(_ IIrNode, orig IAstNode) (nodes []IIrNode) {
	return me.IrIdentBase.findByOrig(me, orig)
}
func (me *IrIdentDecl) walk(ancestors []IIrNode, self IIrNode, on func([]IIrNode, IIrNode, ...IIrNode) bool) bool {
	return me.IrExprAtomBase.walk(ancestors, me, on)
}

func (me *IrIdentName) IsArgRef(maybeSpecificArg *IrArg) bool {
	anyargref := (maybeSpecificArg == nil)
	for _, cand := range me.Anns.Candidates {
		if arg, isargref := cand.(*IrArg); isargref && (anyargref || arg == maybeSpecificArg) {
			return true
		}
	}
	return false
}
func (me *IrIdentName) Let() *IrExprLetBase { return &me.IrExprLetBase }
func (me *IrIdentName) Origin() IAstNode {
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
func (me *IrIdentName) ResolvesTo(n IIrNode) bool {
	for _, cand := range me.Anns.Candidates {
		if cand == n {
			return true
		}
	}
	return false
}
func (me *IrIdentName) EquivTo(node IIrNode) bool {
	cmp, _ := node.(*IrIdentName)
	return cmp != nil && cmp.Val == me.Val && cmp.letDefsEquivTo(&me.IrExprLetBase)
}
func (me *IrIdentName) findByOrig(_ IIrNode, orig IAstNode) (nodes []IIrNode) {
	if nodes = me.IrIdentBase.findByOrig(me, orig); len(nodes) == 0 {
		if orig.Toks().EqLenAndOffsets(me.origToks(), false) {
			nodes = []IIrNode{me}
		} else {
			nodes = me.IrExprLetBase.findByOrig(me, orig)
		}
	}
	return
}
func (me *IrIdentName) walk(ancestors []IIrNode, _ IIrNode, on func([]IIrNode, IIrNode, ...IIrNode) bool) (keepGoing bool) {
	trav := make([]IIrNode, len(me.Defs))
	for i := range me.Defs {
		trav[i] = &me.Defs[i]
	}
	if keepGoing = on(ancestors, me, trav...); keepGoing {
		ancestors = append(ancestors, me)
		for i := range trav {
			if keepGoing = trav[i].walk(ancestors, trav[i], on); !keepGoing {
				break
			}
		}
	}
	return
}

func (me *IrAppl) findByOrig(_ IIrNode, orig IAstNode) (nodes []IIrNode) {
	if me.Orig == orig {
		nodes = []IIrNode{me}
	} else {
		if nodes = me.Callee.findByOrig(me.Callee, orig); len(nodes) != 0 {
			nodes = append(nodes, me)
		} else if nodes = me.CallArg.findByOrig(me.CallArg, orig); len(nodes) != 0 {
			nodes = append(nodes, me)
		} else {
			nodes = me.IrExprLetBase.findByOrig(me, orig)
		}
	}
	return
}
func (me *IrAppl) Origin() IAstNode {
	if me.letOrig != nil {
		return me.letOrig
	}
	return me.Orig
}
func (me *IrAppl) origToks() (toks udevlex.Tokens) {
	if me.letOrig != nil && len(me.letOrig.Tokens) != 0 {
		toks = me.letOrig.Tokens
	} else if toks = me.IrExprBase.origToks(); len(toks) == 0 {
		if toks = me.Callee.origToks(); len(toks) == 0 {
			toks = me.CallArg.origToks()
		}
	}
	return
}
func (me *IrAppl) EquivTo(node IIrNode) bool {
	cmp, _ := node.(*IrAppl)
	return cmp != nil && cmp.Callee.EquivTo(me.Callee) && cmp.CallArg.EquivTo(me.CallArg) && cmp.letDefsEquivTo(&me.IrExprLetBase)
}
func (me *IrAppl) Let() *IrExprLetBase { return &me.IrExprLetBase }
func (me *IrAppl) RefersTo(name string) bool {
	return me.Callee.RefersTo(name) || me.CallArg.RefersTo(name) || me.letDefsReferTo(name)
}
func (me *IrAppl) refsTo(name string) []IIrExpr {
	return append(me.Callee.refsTo(name), append(me.CallArg.refsTo(name), me.letDefsRefsTo(name)...)...)
}
func (me *IrAppl) walk(ancestors []IIrNode, _ IIrNode, on func([]IIrNode, IIrNode, ...IIrNode) bool) (keepGoing bool) {
	trav := make([]IIrNode, len(me.Defs), 2+len(me.Defs))
	for i := range me.Defs {
		trav[i] = &me.Defs[i]
	}
	trav = append(trav, me.Callee, me.CallArg)
	if keepGoing = on(ancestors, me, trav...); keepGoing {
		ancestors = append(ancestors, me)
		for i := range trav {
			if keepGoing = trav[i].walk(ancestors, trav[i], on); !keepGoing {
				break
			}
		}
	}
	return
}

func (me *IrLam) EquivTo(node IIrNode) bool {
	cmp, _ := node.(*IrLam)
	return cmp != nil && me.Arg.EquivTo(&cmp.Arg) && me.Body.EquivTo(cmp.Body)
}
func (me *IrLam) RefersTo(name string) bool { return me.Body.RefersTo(name) }
func (me *IrLam) findByOrig(self IIrNode, orig IAstNode) (nodes []IIrNode) {
	if orig == me.Orig {
		nodes = []IIrNode{self}
	} else if nodes = me.Body.findByOrig(me.Body, orig); len(nodes) != 0 {
		nodes = append(nodes, self)
	} else if nodes = me.Arg.findByOrig(&me.Arg, orig); len(nodes) != 0 {
		nodes = append(nodes, self)
	}
	return
}
func (me *IrLam) refsTo(name string) []IIrExpr { return me.Body.refsTo(name) }
func (me *IrLam) walk(ancestors []IIrNode, self IIrNode, on func([]IIrNode, IIrNode, ...IIrNode) bool) (keepGoing bool) {
	if keepGoing = on(ancestors, self, &me.Arg, me.Body); keepGoing {
		ancestors = append(ancestors, self)
		if keepGoing = me.Arg.walk(ancestors, &me.Arg, on); keepGoing {
			keepGoing = keepGoing && me.Body.walk(ancestors, me.Body, on)
		}
	}
	return
}
