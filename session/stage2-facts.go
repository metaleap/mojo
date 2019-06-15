package atmosess

import (
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/il"
)

type ctxFacts struct {
	kit    *Kit
	done   map[string]bool
	defTop *atmoil.AstDefTop
}

func (me *Ctx) refreshFactsForTopLevelDefs(defIdsBorn map[string]*Kit, defIdsDependantsOfNamesOfChange map[string]*Kit) (freshFactsErrs atmo.Errors) {
	done := make(map[string]bool, len(defIdsBorn)+len(defIdsDependantsOfNamesOfChange))
	inorder := make([]*ctxFacts, 0, len(defIdsBorn)+len(defIdsDependantsOfNamesOfChange))

	for _, m := range []map[string]*Kit{defIdsBorn, defIdsDependantsOfNamesOfChange} {
		for defid, kit := range m {
			ctx := &ctxFacts{kit: kit, done: done, defTop: kit.lookups.tlDefsByID[defid]}
			if me.refreshCoreFactsForTopLevelDef(ctx) {
				inorder, ctx.done = append(inorder, ctx), nil
			}
		}
	}

	for _, ctx := range inorder {
		me.refreshDerivedFactsForTopLevelDef(ctx)
	}

	return
}

func (me *Ctx) refreshCoreFactsForTopLevelDef(ctx *ctxFacts) bool {
	if isdone, isdoing := ctx.done[ctx.defTop.Id]; isdone {
		return false
	} else if isdoing {
		panic("TODO for refreshCoreFactsForTopLevelDef: handle circular dependencies aka recursion!")
	}
	ctx.done[ctx.defTop.Id] = false // marks as "doing", at the end `true` marks as "done"
	me.refreshCoreFactsForDef(ctx, &ctx.defTop.AstDef, nil)
	ctx.done[ctx.defTop.Id] = true
	return true
}

func (me *Ctx) refreshCoreFactsForDef(ctx *ctxFacts, node *atmoil.AstDef, ancestors []atmoil.IAstNode) {
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
		facts.Core = &atmoil.AnnFactLit{Value: n.Val}
	case *atmoil.AstLitRune:
		facts.Core = &atmoil.AnnFactLit{Value: n.Val}
	case *atmoil.AstLitStr:
		facts.Core = &atmoil.AnnFactLit{Value: n.Val}
	case *atmoil.AstLitUint:
		facts.Core = &atmoil.AnnFactLit{Value: n.Val}
	case *atmoil.AstIdentTag:
		facts.Core = &atmoil.AnnFactTag{Value: n.Val}
	case *atmoil.AstIdentBad:
		facts.Core = &atmoil.AnnFactBad{}
	case *atmoil.AstIdentName:
		switch l := len(n.Anns.ResolvesTo); l {
		case 0:
			facts.Core = &atmoil.AnnFactBad{}
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

func (me *Ctx) refreshDerivedFactsForTopLevelDef(ctx *ctxFacts) {
	// me.refreshDerivedFactsForDef(ctx, &ctx.defTop.AstDef, nil)
}
