package atmocorefn

import (
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang"
)

type AstDefs []AstDef

func (me AstDefs) byName(name string) *AstDef {
	for i := range me {
		if me[i].Name.Val == name {
			return &me[i]
		}
	}
	return nil
}

func (me *AstDefs) add(body IAstExpr) (def *AstDef) {
	this := *me
	idx := len(this)
	this = append(this, AstDef{Body: body})
	*me, def = this, &this[idx]
	return
}

func (me AstDefs) index(name string) int {
	for i := range me {
		if me[i].Name.Val == name {
			return i
		}
	}
	return -1
}

type AstTopDefs []AstDefTop

func (me AstTopDefs) Len() int          { return len(me) }
func (me AstTopDefs) Swap(i int, j int) { me[i], me[j] = me[j], me[i] }
func (me AstTopDefs) Less(i int, j int) bool {
	dis, dat := &me[i].Orig.Tokens[0].Meta, &me[j].Orig.Tokens[0].Meta
	return (dis.Filename == dat.Filename && dis.Offset < dat.Offset) || dis.Filename < dat.Filename
}

func (me AstTopDefs) IndexByID(id string) int {
	for i := range me {
		if len(me[i].TopLevels) == 1 && me[i].TopLevels[0].ID() == id {
			return i
		}
	}
	return -1
}

func (me *AstTopDefs) ReInitFrom(kitSrcFiles atmolang.AstFiles) (freshErrs []error) {
	this, newdefs, oldunchangeddefidxs := *me, make([]*atmolang.AstFileTopLevelChunk, 0, 2), make(map[int]bool, len(*me))

	// gather whats "new" (newly added or source-wise modified) and whats "old" (source-wise unchanged)
	for i := range kitSrcFiles {
		for j := range kitSrcFiles[i].TopLevel {
			if tl := &kitSrcFiles[i].TopLevel[j]; tl.Ast.Def.Orig != nil && !tl.HasErrors() {
				if defidx := this.IndexByID(tl.ID()); defidx < 0 {
					newdefs = append(newdefs, tl)
				} else {
					oldunchangeddefidxs[defidx] = true
				}
			}
		}
	}

	if len(oldunchangeddefidxs) > 0 { // gather & drop what's gone
		dels := make(map[*atmolang.AstDef]bool, len(this)-len(oldunchangeddefidxs))
		for i := range this {
			if !oldunchangeddefidxs[i] {
				dels[this[i].Orig] = true
			}
		}
		for i := range this {
			if dels[this[i].Orig] {
				for j := i; j < len(this)-1; j++ {
					this[j] = this[j+1]
				}
			}
		}
		this = this[:len(this)-len(dels)]
	}

	// add what's new
	newstartfrom := len(this)
	for _, tlc := range newdefs {
		if !tlc.HasErrors() {
			idx := len(this)
			this = append(this, AstDefTop{})
			this[idx].TopLevels, this[idx].Orig = append(this[idx].TopLevels, tlc), tlc.Ast.Def.Orig
		}
	}

	names, ndone := make([]*atmolang.AstIdent, 0, len(this)), make(map[string]bool, len(this))
	for i := range this {
		if name := &this[i].Orig.Name; !ndone[name.Val] {
			ndone[name.Val], names = true, append(names, name)
		}
	}

	// populate new `Def`s from orig AST node
	for i := newstartfrom; i < len(this); i++ {
		def := &this[i]
		var let AstExprLetBase
		var ctxastinit ctxAstInit
		ctxastinit.namesInScope, ctxastinit.defsScope, ctxastinit.curTopLevel = names, &let.letDefs, def.Orig
		errs := def.AstDef.initFrom(&ctxastinit, def.Orig)
		if len(let.letDefs) > 0 {
			errs.Add(ctxastinit.addLetDefsToNode(def.Orig.Body, def.AstDef.Body, &let))
		}
		if def.Errors.Add(errs) {
			for e := range errs {
				freshErrs = append(freshErrs, &errs[e])
			}
		}
		atmo.SortMaybe(def.Errors)
	}
	atmo.SortMaybe(this)
	*me = this
	return
}
