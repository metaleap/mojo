package atmosess

import (
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/il"
	"github.com/metaleap/atmo/lang"
)

type ctxFacts struct {
	// ctxSess          *Ctx
	defTop             *atmoil.AstDefTop
	kit                *Kit
	nodesWithAncestors map[atmoil.IAstNode][]atmoil.IAstNode
	allRefrCtxs        map[string]*ctxFacts
}

func (me *Ctx) refreshFactsForTopLevelDefs(defIdsBorn map[string]*Kit, defIdsDependantsOfNamesOfChange map[string]*Kit) (freshFactsErrs atmo.Errors) {
	l := len(defIdsBorn) + len(defIdsDependantsOfNamesOfChange)
	done, inorder, allnodes, allctxs := make(map[string]bool, l), make([]*ctxFacts, 0, l), make(map[atmoil.IAstNode][]atmoil.IAstNode, l*3), make(map[string]*ctxFacts, l)

	for _, m := range []map[string]*Kit{defIdsBorn, defIdsDependantsOfNamesOfChange} {
		for defid, kit := range m {
			ctx := &ctxFacts{ /*ctxSess: me,*/ defTop: kit.lookups.tlDefsByID[defid], kit: kit, nodesWithAncestors: allnodes, allRefrCtxs: allctxs}
			if allctxs[defid] = ctx; ctx.refreshCoreFactsForTopLevelDef(done) {
				inorder = append(inorder, ctx)
			}
		}
	}

	for _, ctx := range inorder {
		ctx.refreshDerivedFactsForDef(&ctx.defTop.AstDef)
	}

	return
}

func (me *ctxFacts) refreshCoreFactsForTopLevelDef(done map[string]bool) bool {
	if isdone, isdoing := done[me.defTop.Id]; isdone {
		return false
	} else if isdoing {
		panic("TODO for refreshCoreFactsForTopLevelDef: handle circular dependencies aka recursion!")
	}
	done[me.defTop.Id] = false // marks as "doing", at the end `true` marks as "done"
	me.refreshCoreFactsForDef(&me.defTop.AstDef, make([]atmoil.IAstNode, 0, 1))
	done[me.defTop.Id] = true
	return true
}

func (me *ctxFacts) refreshCoreFactsForDef(node *atmoil.AstDef, ancestors []atmoil.IAstNode) {
	me.nodesWithAncestors[node] = ancestors
	facts := node.Facts()
	facts.Reset()

	if node.Arg != nil {
		node.Arg.Facts().Reset()
		facts.Core = &atmoil.AnnFactCallable{Arg: &atmoil.AnnFactRef{To: node.Arg}, Ret: &atmoil.AnnFactRef{To: node.Body}}
	} else {
		facts.Core = &atmoil.AnnFactRef{To: node.Body}
	}

	me.refreshCoreFactsForExpr(node.Body, append(ancestors, node))
}

func (me *ctxFacts) refreshCoreFactsForExpr(node atmoil.IAstExpr, ancestors []atmoil.IAstNode) {
	me.nodesWithAncestors[node] = ancestors
	facts := node.Facts()
	facts.Reset()

	var ancestorswithnode []atmoil.IAstNode
	let := node.Let()
	if let != nil && len(let.Defs) > 0 {
		ancestorswithnode = append(ancestors, node)
		for i := range let.Defs {
			me.refreshCoreFactsForDef(&let.Defs[i], ancestorswithnode)
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
		me.refreshCoreFactsForExpr(n.AtomicCallee, ancestorswithnode)
		me.refreshCoreFactsForExpr(n.AtomicArg, ancestorswithnode)
	default:
		panic(n)
	}
}

func (me *ctxFacts) refreshedDerivedFactsForRef(node atmoil.IAstNode) atmoil.AnnFacts {
	ctx := me
	nodeancestors, ispartofrefresh := ctx.nodesWithAncestors[node]
	if !ispartofrefresh {
		switch n := node.(type) {
		case *atmoil.AstDefTop:
			node = &n.AstDef
			nodeancestors, ispartofrefresh = ctx.nodesWithAncestors[node]
		case AstDefRef:
			node = &n.AstDef
			if nodeancestors, ispartofrefresh = ctx.nodesWithAncestors[node]; !ispartofrefresh {
				ctx = ctx.allRefrCtxs[n.Id]
				nodeancestors, ispartofrefresh = ctx.nodesWithAncestors[node]
			}
		}
	}

	if ispartofrefresh && nodeancestors != nil {
		switch n := node.(type) {
		case *atmoil.AstDef:
			ctx.refreshDerivedFactsForDef(n)
		case atmoil.IAstExpr:
			ctx.refreshDerivedFactsForExpr(n)
		default:
			panic(n)
		}
	}
	return node.Facts().Derived
}

func (me *ctxFacts) refreshDerivedFactsForDef(node *atmoil.AstDef) {
	nodeancestors, ispartofrefresh := me.nodesWithAncestors[node]
	if (!ispartofrefresh) || nodeancestors == nil {
		return
	}
	facts := node.Facts()

	me.refreshDerivedFactsForExpr(node.Body)
	switch fc := facts.Core.(type) {
	case *atmoil.AnnFactRef:
		facts.Derived.Add(me.refreshedDerivedFactsForRef(fc.To))
	case *atmoil.AnnFactCallable:
		facts.Derived.Add(me.refreshedDerivedFactsForRef(fc.Ret.To))
	}

	me.nodesWithAncestors[node] = nil
}

func (me *ctxFacts) refreshDerivedFactsForExpr(node atmoil.IAstExpr) {
	nodeancestors, ispartofrefresh := me.nodesWithAncestors[node]
	if (!ispartofrefresh) || nodeancestors == nil {
		return
	}
	facts := node.Facts()

	let := node.Let()
	if let != nil && len(let.Defs) > 0 {
		for i := range let.Defs {
			me.refreshDerivedFactsForDef(&let.Defs[i])
		}
	}

	switch n := node.(type) {
	case *atmoil.AstIdentName:
		switch fc := facts.Core.(type) {
		case *atmoil.AnnFactRef:
			d := me.refreshedDerivedFactsForRef(fc.To)
			facts.Derived.Add(d)
		case *atmoil.AnnFactAlts:
			alts := &atmoil.AnnFactAlts{Possibilities: make(atmoil.AnnFacts, len(fc.Possibilities))}
			for i, fcp := range fc.Possibilities {
				fcr := fcp.(*atmoil.AnnFactRef)
				alts.Possibilities[i] = me.refreshedDerivedFactsForRef(fcr.To)
			}
			facts.Derived.Add(alts)
		}
	case *atmoil.AstAppl:
		// fc := facts.Core.(*atmoil.AnnFactCall)
		me.refreshDerivedFactsForExpr(n.AtomicCallee)
		me.refreshDerivedFactsForExpr(n.AtomicArg)
	default:
		facts.Derived.Add(facts.Core)
	}

	me.nodesWithAncestors[node] = nil
}
