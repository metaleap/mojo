package atmocorefn

import (
	"github.com/metaleap/atmo/lang"
	"sort"
)

type astDefs []AstDefBase

func (me astDefs) byName(name string) *AstDefBase {
	for i := range me {
		if me[i].Name.String() == name {
			return &me[i]
		}
	}
	return nil
}

type AstDefs []AstDef

func (me AstDefs) Len() int          { return len(me) }
func (me AstDefs) Swap(i int, j int) { me[i], me[j] = me[j], me[i] }
func (me AstDefs) Less(i int, j int) bool {
	dis, dat := &me[i].Orig.Tokens[0].Meta, &me[j].Orig.Tokens[0].Meta
	return (dis.Filename == dat.Filename && dis.Line < dat.Line) || dis.Filename < dat.Filename
}

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
			if tl := &kitSrcFiles[i].TopLevel[j]; tl.Ast.Def.Orig != nil && !tl.HasErrors() {
				if defidx := this.IndexByID(tl.ID()); defidx < 0 {
					newdefs = append(newdefs, tl)
				} else {
					oldunchangeddefidxs = append(oldunchangeddefidxs, defidx)
				}
			}
		}
	}

	// gather & drop what's gone
	this.removeAllExcept(oldunchangeddefidxs)

	// add what's new
	newstartfrom := len(this)
	for _, tlc := range newdefs {
		if !tlc.HasErrors() {
			this = append(this, AstDef{TopLevel: tlc})
		}
	}

	// populate new `Def`s from orig AST node
	for i := newstartfrom; i < len(this); i++ {
		this[i].initFrom(this[i].TopLevel.Ast.Def.Orig)
	}

	sort.Sort(this)
	*me = this
}

func (me *AstDefs) removeAllExcept(keepIdxs []int) {
	if this := *me; len(keepIdxs) == 0 {
		*me = this[:0]
	} else {
		nume := make(AstDefs, 0, len(keepIdxs))
		for _, idx := range keepIdxs {
			nume = append(nume, this[idx])
		}
		*me = nume
	}
}
