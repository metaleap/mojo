package atmocorefn

import (
	"github.com/metaleap/atmo/lang"
)

type astDefs []AstDefBase

type AstDefs []AstDef

func (me AstDefs) ByID(id string) *AstDef {
	for i := range me {
		if me[i].TopLevel.ID() == id {
			return &me[i]
		}
	}
	return nil
}
func (me AstDefs) IndexByID(id string) int {
	for i := range me {
		if me[i].TopLevel.ID() == id {
			return i
		}
	}
	return -1
}

func (me *AstDefs) Reload(kitSrcFiles atmolang.AstFiles) {
	this, newdefs, oldunchangeddefidxs := *me, make([]*atmolang.AstFileTopLevelChunk, 0, 2), make([]int, 0, len(*me))

	// gather whats "new" (newly added or source-wise modified) and whats "old" (source-wise unchanged)
	for i := range kitSrcFiles {
		for j := range kitSrcFiles[i].TopLevel {
			if tl := &kitSrcFiles[i].TopLevel[j]; tl.Ast.Def.Orig != nil {
				if defidx := this.IndexByID(tl.ID()); defidx < 0 {
					newdefs = append(newdefs, tl)
				} else {
					oldunchangeddefidxs = append(oldunchangeddefidxs, defidx)
				}
			}
		}
	}

	// gather & drop what's gone
	this.removeAt(oldunchangeddefidxs...)

	// add what's new
	newstartfrom := len(this)
	for _, tlc := range newdefs {
		this = append(this, AstDef{TopLevel: tlc})
	}

	// populate new `Def`s from orig AST node
	for i := newstartfrom; i < len(this); i++ {
		this[i].initFrom(this[i].TopLevel.Ast.Def.Orig)
	}

	*me = this
}

func (me *AstDefs) removeAt(idxs ...int) {
	this := *me
	n, nume := 0, make(AstDefs, len(this)-len(idxs))
	for i := range this {
		var remove bool
		for _, idx := range idxs {
			if remove = (i == idx); remove {
				break
			}
		}
		if !remove {
			n, nume[n] = n+1, this[i]
		}
	}
	*me = nume
}
