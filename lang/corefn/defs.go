package atmocorefn

import (
	"github.com/metaleap/atmo/lang"
)

type Defs []Def

func (me Defs) ByID(id string) *Def {
	for i := range me {
		if me[i].Orig.ID() == id {
			return &me[i]
		}
	}
	return nil
}

func (me *Defs) Reload(packSrcFiles atmolang.AstFiles) {
	this, olddefs, newdefs := *me, map[*Def]bool{}, make([]*atmolang.AstFileTopLevelChunk, 0, 2)

	// gather whats "new" (newly added or source-wise modified) and whats "old" (source-wise unchanged)
	for i := range packSrcFiles {
		for j := range packSrcFiles[i].TopLevel {
			tl := &packSrcFiles[i].TopLevel[j]
			tlid := tl.ID()
			if def := this.ByID(tlid); def == nil {
				newdefs = append(newdefs, tl)
			} else {
				olddefs[def] = true
			}
		}
	}

	// gather & drop what's gone
	for i := 0; i < len(this); i++ {
		if !olddefs[&this[i]] {
			me.removeAt(i)
			i--
		}
	}

	// add what's new
	newstartfrom := len(this)
	for _, tlc := range newdefs {
		this = append(this, Def{Orig: tlc})
	}

	// populate new `Def`s from orig AST node
	for i := newstartfrom; i < len(this); i++ {
		this[i].populate()
	}

	*me = this
}

func (me *Defs) removeAt(idx int) {
	this := *me
	for i := idx; i < len(this)-1; i++ {
		this[i] = this[i+1]
	}
	this = this[:len(this)-1]
	*me = this
}
