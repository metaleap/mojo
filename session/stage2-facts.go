package atmosess

import (
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/il"
	"github.com/metaleap/atmo/lang"
)

type ctxFacts struct {
	kit              *Kit
	defTop           *atmoil.AstDefTop
	nodesRefrDerived map[atmoil.IAstNode][]atmoil.IAstNode
}

func (me *Ctx) refreshFactsForTopLevelDefs(defIdsBorn map[string]*Kit, defIdsDependantsOfNamesOfChange map[string]*Kit) (freshFactsErrs atmo.Errors) {
	l := len(defIdsBorn) + len(defIdsDependantsOfNamesOfChange)
	done, inorder, allnodes := make(map[string]bool, l), make([]*ctxFacts, 0, l), make(map[atmoil.IAstNode][]atmoil.IAstNode, l*3)

	for _, m := range []map[string]*Kit{defIdsBorn, defIdsDependantsOfNamesOfChange} {
		for defid, kit := range m {
			ctx := &ctxFacts{kit: kit, defTop: kit.lookups.tlDefsByID[defid], nodesRefrDerived: allnodes}
			if me.refreshCoreFactsForTopLevelDef(ctx, done) {
				inorder = append(inorder, ctx)
			}
		}
	}

	for _, ctx := range inorder {
		me.refreshDerivedFactsForDef(ctx, &ctx.defTop.AstDef, nil)
	}

	return
}

func (me *Ctx) refreshCoreFactsForTopLevelDef(ctx *ctxFacts, done map[string]bool) bool {
	if isdone, isdoing := done[ctx.defTop.Id]; isdone {
		return false
	} else if isdoing {
		panic("TODO for refreshCoreFactsForTopLevelDef: handle circular dependencies aka recursion!")
	}
	done[ctx.defTop.Id] = false // marks as "doing", at the end `true` marks as "done"
	me.refreshCoreFactsForDef(ctx, &ctx.defTop.AstDef, make([]atmoil.IAstNode, 0))
	done[ctx.defTop.Id] = true
	return true
}

func (me *Ctx) refreshCoreFactsForDef(ctx *ctxFacts, node *atmoil.AstDef, ancestors []atmoil.IAstNode) {
	ctx.nodesRefrDerived[node] = ancestors
	facts := node.Facts()
	facts.Reset()

	if node.Arg != nil {
		node.Arg.Facts().Reset()
		facts.Core = &atmoil.AnnFactCallable{Arg: &atmoil.AnnFactRef{To: node.Arg}, Ret: &atmoil.AnnFactRef{To: node.Body}}
	} else {
		facts.Core = &atmoil.AnnFactRef{To: node.Body}
	}

	me.refreshCoreFactsForExpr(ctx, node.Body, append(ancestors, node))
}

func (me *Ctx) refreshCoreFactsForExpr(ctx *ctxFacts, node atmoil.IAstExpr, ancestors []atmoil.IAstNode) {
	ctx.nodesRefrDerived[node] = ancestors
	facts := node.Facts()
	facts.Reset()

	var ancestorswithnode []atmoil.IAstNode
	let := node.Let()
	if let != nil && len(let.Defs) > 0 {
		ancestorswithnode = append(ancestors, node)
		for i := range let.Defs {
			me.refreshCoreFactsForDef(ctx, &let.Defs[i], ancestorswithnode)
		}
	}

	switch n := node.(type) {
	case *atmoil.AstLitFloat:
		facts.Core = &atmoil.AnnFactLit{Value: n.Val, Str: n.Orig.(atmolang.IAstExprAtomic).String}
	case *atmoil.AstLitRune:
		facts.Core = &atmoil.AnnFactLit{Value: n.Val, Str: n.Orig.(atmolang.IAstExprAtomic).String}
	case *atmoil.AstLitStr:
		facts.Core = &atmoil.AnnFactLit{Value: n.Val, Str: n.Orig.(atmolang.IAstExprAtomic).String}
	case *atmoil.AstLitUint:
		facts.Core = &atmoil.AnnFactLit{Value: n.Val, Str: n.Orig.(atmolang.IAstExprAtomic).String}
	case *atmoil.AstIdentTag:
		facts.Core = &atmoil.AnnFactTag{Value: n.Val}
	case *atmoil.AstUndef:
		facts.Core = &atmoil.AnnFactUndef{}
	case *atmoil.AstIdentName:
		switch l := len(n.Anns.ResolvesTo); l {
		case 0:
			facts.Core = &atmoil.AnnFactUndef{}
		case 1:
			facts.Core = &atmoil.AnnFactRef{To: n.Anns.ResolvesTo[0]}
		default:
			cases := &atmoil.AnnFactAlts{Possibilities: make(atmoil.AnnFacts, len(n.Anns.ResolvesTo))}
			for i := range cases.Possibilities {
				cases.Possibilities[i] = &atmoil.AnnFactRef{To: n.Anns.ResolvesTo[i]}
			}
			facts.Core = cases
		}
	case *atmoil.AstAppl:
		if ancestorswithnode == nil {
			ancestorswithnode = append(ancestors, node)
		}
		facts.Core = &atmoil.AnnFactCall{Callee: &atmoil.AnnFactRef{To: n.AtomicCallee}, Arg: &atmoil.AnnFactRef{To: n.AtomicArg}}
		me.refreshCoreFactsForExpr(ctx, n.AtomicCallee, ancestorswithnode)
		me.refreshCoreFactsForExpr(ctx, n.AtomicArg, ancestorswithnode)
	default:
		panic(n)
	}
}

func (me *Ctx) refreshedDerivedFactsForRef(ctx *ctxFacts, node atmoil.IAstNode) atmoil.AnnFacts {
	if nodeancestors, ispartofrefresh := ctx.nodesRefrDerived[node]; ispartofrefresh && nodeancestors != nil {
		switch n := node.(type) {
		case *atmoil.AstDef:
			me.refreshDerivedFactsForDef(ctx, n, nodeancestors)
		case atmoil.IAstExpr:
			me.refreshDerivedFactsForExpr(ctx, n, nodeancestors)
		}
	}
	return node.Facts().Derived
}

func (me *Ctx) refreshDerivedFactsForDef(ctx *ctxFacts, node *atmoil.AstDef, ancestors []atmoil.IAstNode) {
	ctx.nodesRefrDerived[node] = nil
	facts := node.Facts()

	me.refreshDerivedFactsForExpr(ctx, node.Body, append(ancestors, node))
	switch fc := facts.Core.(type) {
	case *atmoil.AnnFactRef:
		facts.Derived.Add(me.refreshedDerivedFactsForRef(ctx, fc.To)...)
	case *atmoil.AnnFactCallable:
		facts.Derived.Add(me.refreshedDerivedFactsForRef(ctx, fc.Ret.To)...)
	}
}

func (me *Ctx) refreshDerivedFactsForExpr(ctx *ctxFacts, node atmoil.IAstExpr, ancestors []atmoil.IAstNode) {
	ctx.nodesRefrDerived[node] = nil
	facts := node.Facts()

	var ancestorswithnode []atmoil.IAstNode
	let := node.Let()
	if let != nil && len(let.Defs) > 0 {
		ancestorswithnode = append(ancestors, node)
		for i := range let.Defs {
			me.refreshDerivedFactsForDef(ctx, &let.Defs[i], ancestorswithnode)
		}
	}

	facts.Derived.Add(facts.Core)
	switch n := node.(type) {
	case *atmoil.AstIdentName:
		switch fc := facts.Core.(type) {
		case *atmoil.AnnFactRef:
			facts.Derived.Add(me.refreshedDerivedFactsForRef(ctx, fc.To)...)
		case *atmoil.AnnFactAlts:
			alts := &atmoil.AnnFactAlts{Possibilities: make(atmoil.AnnFacts, len(fc.Possibilities))}
			for i, fcp := range fc.Possibilities {
				fcr := fcp.(*atmoil.AnnFactRef)
				alts.Possibilities[i] = me.refreshedDerivedFactsForRef(ctx, fcr.To)
			}
			facts.Derived.Add(alts)
		}
	case *atmoil.AstAppl:
		// fc := facts.Core.(*atmoil.AnnFactCall)
		if ancestorswithnode == nil {
			ancestorswithnode = append(ancestors, node)
		}
		me.refreshDerivedFactsForExpr(ctx, n.AtomicCallee, ancestorswithnode)
		me.refreshDerivedFactsForExpr(ctx, n.AtomicArg, ancestorswithnode)
	}
}
