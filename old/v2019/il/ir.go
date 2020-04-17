package atmoil

import (
	"github.com/go-leap/dev/lex"
	"github.com/go-leap/str"
	. "github.com/metaleap/atmo/old/v2019"
	. "github.com/metaleap/atmo/old/v2019/ast"
)

func (*irNodeBase) IsDef() *IrDef        { return nil }
func (me *irNodeBase) AstOrig() IAstNode { return me.Orig }

func (me *IrDef) findByOrig(self IIrNode, orig IAstNode, ok func(IIrNode) bool) (nodes []IIrNode) {
	if nodes = me.Ident.findByOrig(&me.Ident, orig, ok); len(nodes) != 0 {
		nodes = append(nodes, self)
	} else if nodes = me.Body.findByOrig(me.Body, orig, ok); len(nodes) != 0 {
		nodes = append(nodes, self)
	} else if orig == me.Orig && (ok == nil || ok(self)) {
		nodes = []IIrNode{self}
	}
	return
}
func (me *IrDef) IsDef() *IrDef                { return me }
func (me *IrDef) OrigDef() (origDef *AstDef)   { origDef, _ = me.Orig.(*AstDef); return }
func (me *IrDef) refsTo(name string) []IIrExpr { return me.Body.refsTo(name) }
func (me *IrDef) EquivTo(node IIrNode, ignoreNames bool) bool {
	cmp, _ := node.(*IrDef)
	return cmp != nil && (ignoreNames || cmp.Ident.Name == me.Ident.Name) &&
		me.Body.EquivTo(cmp.Body, ignoreNames)
}
func (me *IrDef) walk(ancestors []IIrNode, self IIrNode, on func([]IIrNode, IIrNode, ...IIrNode) bool) (keepGoing bool) {
	if keepGoing = on(ancestors, self, &me.Ident, me.Body); keepGoing {
		ancestors = append(ancestors, self)
		if keepGoing = me.Ident.walk(ancestors, &me.Ident, on); keepGoing {
			keepGoing = me.Body.walk(ancestors, me.Body, on)
		}
	}
	return
}
func (me *IrDef) HasErrors() bool {
	return len(me.Errs.Stage1AstToIr) != 0 || len(me.Errs.Stage2BadNames) != 0 || len(me.Errs.Stage3Preduce) != 0
}
func (me *IrDef) Errors() (errs Errors) {
	errs = make(Errors, 0, len(me.Errs.Stage1AstToIr)+len(me.Errs.Stage2BadNames)+len(me.Errs.Stage3Preduce))
	errs = append(append(append(errs, me.Errs.Stage1AstToIr...), me.Errs.Stage2BadNames...), me.Errs.Stage3Preduce...)
	return
}
func (me *IrDef) FindByOrig(orig IAstNode, ok func(IIrNode) bool) []IIrNode {
	return me.findByOrig(me, orig, ok)
}
func (me *IrDef) AncestorsOf(node IIrNode) (nodeAncestors []IIrNode) {
	if me != nil && node != me {
		me.Walk(func(curnodeancestors []IIrNode, curnode IIrNode, _ ...IIrNode) (keepgoing bool) {
			if keepgoing = (nodeAncestors == nil); curnode == node {
				keepgoing, nodeAncestors = false, curnodeancestors
			}
			return
		})
	}
	return
}
func (me *IrDef) AncestorsAndChildrenOf(node IIrNode) (nodeAncestors []IIrNode, nodeChildren []IIrNode) {
	me.Walk(func(curnodeancestors []IIrNode, curnode IIrNode, curnodechildren ...IIrNode) (keepGoing bool) {
		if curnode == node {
			nodeAncestors, nodeChildren = curnodeancestors, curnodechildren
		}
		return nodeAncestors == nil && nodeChildren == nil
	})
	return
}
func (me *IrDef) ArgOwnerAbs(arg *IrArg) *IrAbs {
	if arg.Ann.Parent == nil && me != nil {
		me.Walk(func(_ []IIrNode, node IIrNode, _ ...IIrNode) bool {
			if abs, is := node.(*IrAbs); is {
				abs.Arg.Ann.Parent = abs
			}
			return true
		})
	}
	return arg.Ann.Parent
}
func (me *IrDef) AstOrigToks(node IIrNode) (toks udevlex.Tokens) {
	if node == nil {
		node = me
	}
	if node == nil {
		return
	}
	_ = node.walk(nil, node, func(_ []IIrNode, cn IIrNode, _ ...IIrNode) bool {
		if len(toks) == 0 {
			if orig := cn.AstOrig(); orig != nil {
				toks = orig.Toks()
			}
		}
		return len(toks) == 0
	})
	if me != nil && len(toks) == 0 {
		nodeancestors := me.AncestorsOf(node)
		for i := len(nodeancestors) - 1; len(toks) == 0 && i >= 0; i-- {
			if orig := nodeancestors[i].AstOrig(); orig != nil {
				toks = orig.Toks()
			}
		}
	}
	return
}
func (me *IrDef) RefersToOrDefines(name string) (relatesTo bool) {
	relatesTo = me.Ident.Name == name || me.RefersTo(name)
	if !relatesTo {
		me.Walk(func(_ []IIrNode, node IIrNode, _ ...IIrNode) (keepgoing bool) {
			abs, is := node.(*IrAbs)
			relatesTo = relatesTo || (is && abs.Arg.Name == name)
			return !relatesTo
		})
	}
	return
}
func (me *IrDef) RefersTo(name string) (refersTo bool) {
	return me.refersTo[name] // we now store this on idents in ast-to-ir stage

	// as long as an IrDef exists, it represents the same original code snippet: so any given
	// RefersTo(foo) truth will hold throughout: so we cache instead of continuously re-searching
	// var known bool
	// if refersTo, known = me.refersTo[name]; !known {
	// 	refersTo = me.Body.RefersTo(name)
	// 	me.refersTo[name] = refersTo
	// }
	// return
}
func (me *IrDef) RefsTo(name string) (refs []IIrExpr) {
	for len(name) != 0 && name[0] == '_' {
		name = name[1:]
	}
	if len(name) != 0 {
		// leverage the bool cache already in place two ways, though we dont cache the occurrences
		// in detail (they're usually for editor or error-message scenarios, not hi-perf paths)
		if refersto, known := me.refersTo[name]; refersto || !known {
			if refs = me.refsTo(name); !known {
				me.refersTo[name] = (len(refs) != 0)
			}
		}
	}
	return
}
func (me *IrDef) Walk(whetherToKeepTraversing func(curNodeAncestors []IIrNode, curNode IIrNode, curNodeChildrenThatWillBeTraversedIfReturningTrue ...IIrNode) bool) {
	_ = me.walk(nil, me, whetherToKeepTraversing)
}
func (me *IrDef) FindAny(where func(IIrNode) bool) (firstMatchWithAncestorsPrepended []IIrNode) {
	me.Walk(func(ancestors []IIrNode, curnode IIrNode, children ...IIrNode) bool {
		if where(curnode) {
			firstMatchWithAncestorsPrepended = append(ancestors, curnode)
		}
		return firstMatchWithAncestorsPrepended == nil
	})
	return
}
func (me *IrDef) FindAll(where func(IIrNode) bool) (matchingNodesWithAncestorsPrepended [][]IIrNode) {
	me.Walk(func(curnodeancestors []IIrNode, curnode IIrNode, curnodechildren ...IIrNode) bool {
		if where(curnode) {
			matchingNodesWithAncestorsPrepended = append(matchingNodesWithAncestorsPrepended,
				append(curnodeancestors, curnode))
		}
		return true
	})
	return
}
func (me *IrDef) HasAnyOf(nodes ...IIrNode) bool {
	return nil != me.FindAny(func(node IIrNode) bool {
		for _, n := range nodes {
			if n == node {
				return true
			}
		}
		return false
	})
}
func (me *IrDef) HasIdentDecl(name string) bool {
	return 0 < len(me.FindAny(func(n IIrNode) bool {
		identdecl, ok := n.(*IrIdentDecl)
		return ok && identdecl.Name == name
	}))
}
func (me *IrDef) NamesInScopeAt(descendantNodeInQuestion IIrNode, knownGlobalsInScope AnnNamesInScope, excludeInternalIdents bool) (namesInScope AnnNamesInScope) {
	namesInScope = knownGlobalsInScope.copy()
	if descendantNodeInQuestion != me {
		nodepath := me.FindAny(func(n IIrNode) bool { return n == descendantNodeInQuestion })
		for _, n := range nodepath {
			if abs, isabs := n.(*IrAbs); isabs {
				if (!excludeInternalIdents) || !abs.Arg.IsInternal() {
					namesInScope.Add(abs.Arg.Name, &abs.Arg)
				}
			}
		}
	}
	return
}

func (me *IrArg) EquivTo(node IIrNode, ignoreNames bool) bool {
	cmp, _ := node.(*IrArg)
	return cmp != nil && (ignoreNames || me.Name == cmp.Name)
}
func (me *IrArg) findByOrig(_ IIrNode, orig IAstNode, ok func(IIrNode) bool) (nodes []IIrNode) {
	if nodes = me.IrIdentDecl.findByOrig(&me.IrIdentDecl, orig, ok); len(nodes) != 0 {
		nodes = append(nodes, me)
	} else if me.Orig == orig && (ok == nil || ok(me)) {
		nodes = []IIrNode{me}
	}
	return
}
func (me *IrArg) AstOrig() IAstNode {
	if me.Orig != nil {
		return me.Orig
	}
	return me.IrIdentDecl.AstOrig()
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

func (me *IrExprAtomBase) findByOrig(self IIrNode, orig IAstNode, ok func(IIrNode) bool) (nodes []IIrNode) {
	if (me.Orig == orig || (me.Orig != nil && orig.Toks().EqLenAndOffsets(me.Orig.Toks(), false))) &&
		(ok == nil || ok(self)) {
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

func (me *IrLitUint) EquivTo(node IIrNode, ignoreNames bool) bool {
	cmp, _ := node.(*IrLitUint)
	return cmp != nil && cmp.Val == me.Val
}
func (me *IrLitUint) refsTo(s string) []IIrExpr { return me.irLitBase.refsTo(me, s) }
func (me *IrLitUint) walk(ancestors []IIrNode, self IIrNode, on func([]IIrNode, IIrNode, ...IIrNode) bool) bool {
	return me.IrExprAtomBase.walk(ancestors, me, on)
}

func (me *IrLitFloat) EquivTo(node IIrNode, ignoreNames bool) bool {
	cmp, _ := node.(*IrLitFloat)
	return cmp != nil && cmp.Val == me.Val
}
func (me *IrLitFloat) refsTo(s string) []IIrExpr { return me.irLitBase.refsTo(me, s) }
func (me *IrLitFloat) walk(ancestors []IIrNode, self IIrNode, on func([]IIrNode, IIrNode, ...IIrNode) bool) bool {
	return me.IrExprAtomBase.walk(ancestors, me, on)
}

func (me *IrNonValue) EquivTo(node IIrNode, ignoreNames bool) bool {
	return false
}
func (me *IrNonValue) findByOrig(_ IIrNode, orig IAstNode, ok func(IIrNode) bool) []IIrNode {
	return me.IrExprAtomBase.findByOrig(me, orig, ok)
}
func (me *IrNonValue) walk(ancestors []IIrNode, self IIrNode, on func([]IIrNode, IIrNode, ...IIrNode) bool) bool {
	return me.IrExprAtomBase.walk(ancestors, me, on)
}

func (me *IrLitTag) EquivTo(node IIrNode, ignoreNames bool) bool {
	cmp, _ := node.(*IrLitTag)
	return cmp != nil && cmp.Val == me.Val
}
func (me *IrLitTag) refsTo(name string) (refs []IIrExpr) {
	if me.Val == name {
		refs = append(refs, me)
	}
	return
}

func (me *IrIdentBase) IsInternal() bool {
	return ustr.Pref(me.Name, "__")
}

func (me *IrIdentDecl) EquivTo(node IIrNode, ignoreNames bool) bool {
	cmp, _ := node.(*IrIdentDecl)
	return cmp != nil && (ignoreNames || cmp.Name == me.Name)
}
func (me *IrIdentDecl) findByOrig(_ IIrNode, orig IAstNode, ok func(IIrNode) bool) []IIrNode {
	return me.IrIdentBase.findByOrig(me, orig, ok)
}
func (me *IrIdentDecl) walk(ancestors []IIrNode, self IIrNode, on func([]IIrNode, IIrNode, ...IIrNode) bool) bool {
	return me.IrExprAtomBase.walk(ancestors, me, on)
}

func (me *IrIdentName) findByOrig(_ IIrNode, orig IAstNode, ok func(IIrNode) bool) []IIrNode {
	return me.IrIdentBase.findByOrig(me, orig, ok)
}
func (me *IrIdentName) RefersTo(name string) bool {
	return me.Name == name
}
func (me *IrIdentName) refsTo(name string) (refs []IIrExpr) {
	if me.Name == name {
		refs = append(refs, me)
	}
	return
}
func (me *IrIdentName) ResolvesTo(node IIrNode) bool {
	for _, cand := range me.Ann.Candidates {
		if cand == node {
			return true
		}
	}
	return false
}
func (me *IrIdentName) EquivTo(node IIrNode, ignoreNames bool) bool {
	cmp, _ := node.(*IrIdentName)
	return cmp != nil && ((ignoreNames && me.Ann.AbsIdx == cmp.Ann.AbsIdx) ||
		((!ignoreNames) && cmp.Name == me.Name))
}
func (me *IrIdentName) walk(ancestors []IIrNode, _ IIrNode, on func([]IIrNode, IIrNode, ...IIrNode) bool) bool {
	return me.IrExprAtomBase.walk(ancestors, me, on)
}

func (me *IrAppl) findByOrig(_ IIrNode, orig IAstNode, ok func(IIrNode) bool) (nodes []IIrNode) {
	if nodes = me.Callee.findByOrig(me.Callee, orig, ok); len(nodes) != 0 {
		nodes = append(nodes, me)
	} else if nodes = me.CallArg.findByOrig(me.CallArg, orig, ok); len(nodes) != 0 {
		nodes = append(nodes, me)
	} else if me.Orig == orig && (ok == nil || ok(me)) {
		nodes = []IIrNode{me}
	}
	return
}
func (me *IrAppl) EquivTo(node IIrNode, ignoreNames bool) bool {
	cmp, _ := node.(*IrAppl)
	return cmp != nil && cmp.Callee.EquivTo(me.Callee, ignoreNames) &&
		cmp.CallArg.EquivTo(me.CallArg, ignoreNames)
}
func (me *IrAppl) RefersTo(name string) bool {
	return me.Callee.RefersTo(name) || me.CallArg.RefersTo(name)
}
func (me *IrAppl) refsTo(name string) []IIrExpr {
	return append(me.Callee.refsTo(name), me.CallArg.refsTo(name)...)
}
func (me *IrAppl) walk(ancestors []IIrNode, _ IIrNode, on func([]IIrNode, IIrNode, ...IIrNode) bool) (keepGoing bool) {
	trav := make([]IIrNode, 0, 2)
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

func (me *IrAbs) EquivTo(node IIrNode, ignoreNames bool) bool {
	cmp, _ := node.(*IrAbs)
	return cmp != nil && me.Arg.EquivTo(&cmp.Arg, ignoreNames) &&
		me.Body.EquivTo(cmp.Body, ignoreNames)
}
func (me *IrAbs) RefersTo(name string) bool { return me.Body.RefersTo(name) }
func (me *IrAbs) findByOrig(self IIrNode, orig IAstNode, ok func(IIrNode) bool) (nodes []IIrNode) {
	if nodes = me.Body.findByOrig(me.Body, orig, ok); len(nodes) != 0 {
		nodes = append(nodes, self)
	} else if nodes = me.Arg.findByOrig(&me.Arg, orig, ok); len(nodes) != 0 {
		nodes = append(nodes, self)
	} else if orig == me.Orig && (ok == nil || ok(self)) {
		nodes = []IIrNode{self}
	}
	return
}
func (me *IrAbs) refsTo(name string) []IIrExpr { return me.Body.refsTo(name) }
func (me *IrAbs) walk(ancestors []IIrNode, self IIrNode, on func([]IIrNode, IIrNode, ...IIrNode) bool) (keepGoing bool) {
	if keepGoing = on(ancestors, self, &me.Arg, me.Body); keepGoing {
		ancestors = append(ancestors, self)
		if keepGoing = me.Arg.walk(ancestors, &me.Arg, on); keepGoing {
			keepGoing = keepGoing && me.Body.walk(ancestors, me.Body, on)
		}
	}
	return
}
