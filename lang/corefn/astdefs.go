package atmocorefn

import (
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang"
)

type astDefs []AstDefBase

func (me astDefs) byName(name string) *AstDefBase {
	for i := range me {
		if me[i].Name.Val == name {
			return &me[i]
		}
	}
	return nil
}

func (me *astDefs) add(body IAstExpr) (def *AstDefBase) {
	this := *me
	idx := len(this)
	this = append(this, AstDefBase{Body: body})
	*me, def = this, &this[idx]
	return
}

func (me astDefs) index(name string) int {
	for i := range me {
		if me[i].Name.Val == name {
			return i
		}
	}
	return -1
}

func (me astDefs) Len() int          { return len(me) }
func (me astDefs) Swap(i int, j int) { me[i], me[j] = me[j], me[i] }
func (me astDefs) Less(i int, j int) bool {
	ni, nj := me[i].Name.Val, me[j].Name.Val
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
	return (dis.Filename == dat.Filename && dis.Offset < dat.Offset) || dis.Filename < dat.Filename
}

func (me AstDefs) IndexByID(id string) int {
	for i := range me {
		if len(me[i].TopLevels) == 1 && me[i].TopLevels[0].ID() == id {
			return i
		}
	}
	return -1
}

func (me *AstDefs) ReInitFrom(kitSrcFiles atmolang.AstFiles) (freshErrs []error) {
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
			this = append(this, AstDef{})
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
		var ctxastinit ctxAstInit
		var let AstLet
		ctxastinit.namesInScope, ctxastinit.defsScope = names, &let.Defs
		if errsinit := def.AstDefBase.initFrom(&ctxastinit, def.Orig); def.Errors.Add(errsinit) {
			for e := range errsinit {
				freshErrs = append(freshErrs, &errsinit[e])
			}
		}
		if len(let.Defs) > 0 {
			let.Body, def.AstDefBase.Body = def.AstDefBase.Body, &let
		}
		atmo.SortMaybe(def.Errors)
	}
	atmo.SortMaybe(this)
	*me = this
	return
}
