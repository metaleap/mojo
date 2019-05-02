package atmocorefn

import (
	"sort"

	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang"
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

func (me astDefs) index(name string) int {
	for i := range me {
		if me[i].Name.String() == name {
			return i
		}
	}
	return -1
}

func (me astDefs) Len() int          { return len(me) }
func (me astDefs) Swap(i int, j int) { me[i], me[j] = me[j], me[i] }
func (me astDefs) Less(i int, j int) bool {
	ni, nj := me[i].Name.String(), me[j].Name.String()
	if me[i].refersTo(nj) {
		return true
	}
	if me[j].refersTo(ni) {
		return false
	}
	return ni < nj
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
		def := &this[i]
		def.Orig, def.Ensure = def.TopLevel.Ast.Def.Orig, func() (errs atmo.Errors) {
			def.Ensure = func() atmo.Errors { return nil }
			const caplocals = 6
			def.Locals = make(astDefs, 0, caplocals)
			errs = def.AstDefBase.initFrom(def, def.Orig)
			def.Errors.Add(errs)
			sort.Sort(def.Errors)
			sort.Sort(def.Locals)
			if len(def.Locals) > caplocals {
				println("LOCALDEFS", len(def.Locals))
			}
			return
		}
	}

	sort.Sort(this)
	names, ndone := make([]string, 0, len(this)), make(map[string]bool, len(this))
	for i := range this {
		if name := this[i].Orig.Name.Val; !ndone[name] {
			ndone[name], names = true, append(names, name)
		}
	}
	for i := range this {
		this[i].state.namesInScope = names
	}
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
