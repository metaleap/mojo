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

func (me *Ctx) refreshFacsForTopLevelDefs(defIdsBorn map[string]*Kit, defIdsDependantsOfNamesOfChange map[string]*Kit) (freshFactsErrs atmo.Errors) {
	done := make(map[string]bool, len(defIdsBorn)+len(defIdsDependantsOfNamesOfChange))
	for defid, kit := range defIdsBorn {
		me.refreshFacsForTopLevelDef(&ctxFacts{kit: kit, done: done, defTop: kit.lookups.tlDefsByID[defid]})
	}
	for defid, kit := range defIdsDependantsOfNamesOfChange {
		me.refreshFacsForTopLevelDef(&ctxFacts{kit: kit, done: done, defTop: kit.lookups.tlDefsByID[defid]})
	}

	return
}

func (me *Ctx) refreshFacsForTopLevelDef(ctx *ctxFacts) {
	if isdone, isdoing := ctx.done[ctx.defTop.Id]; isdone {
		return
	} else if isdoing {
		panic("TODO: handle circular dependencies aka recursion!")
	}
	ctx.done[ctx.defTop.Id] = false // marks as "doing", at the end `true` marks as "done"
	me.refreshFactsForDef(ctx, &ctx.defTop.AstDef, nil)
	ctx.done[ctx.defTop.Id] = true
}

func (me *Ctx) refreshFactsForDef(ctx *ctxFacts, node *atmoil.AstDef, ancestors []atmoil.IAstNode) {
	me.refreshFactsForExpr(ctx, node.Body, append(ancestors, node))
	node.Facts().AllFacts = node.Body.Facts().AllFacts
}

func (me *Ctx) refreshFactsForExpr(ctx *ctxFacts, node atmoil.IAstExpr, ancestors []atmoil.IAstNode) {
	facts := node.Facts()
	facts.AllFacts = nil
}
