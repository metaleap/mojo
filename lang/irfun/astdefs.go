package atmolang_irfun

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

type AstTopDefs []*AstDefTop

func (me AstTopDefs) Len() int          { return len(me) }
func (me AstTopDefs) Swap(i int, j int) { me[i], me[j] = me[j], me[i] }
func (me AstTopDefs) Less(i int, j int) bool {
	dis, dat := &me[i].Orig.Tokens[0].Meta, &me[j].Orig.Tokens[0].Meta
	return (dis.Filename == dat.Filename && dis.Offset < dat.Offset) || dis.Filename < dat.Filename
}

func (me AstTopDefs) IndexByID(id string) int {
	for i := range me {
		if me[i].Id == id {
			return i
		}
	}
	return -1
}

func (me *AstTopDefs) ReInitFrom(kitSrcFiles atmolang.AstFiles) (droppedTopLevelDefIDsAndNames map[string]string, newTopLevelDefIDsAndNames map[string]string, freshErrs []error) {
	this, newdefs, oldunchangeddefidxs := *me, make([]*atmolang.AstFileTopLevelChunk, 0, 2), make(map[int]bool, len(*me))

	// gather whats "new" (newly added or source-wise modified) and whats "old" (source-wise unchanged)
	for i := range kitSrcFiles {
		for j := range kitSrcFiles[i].TopLevel {
			if tl := &kitSrcFiles[i].TopLevel[j]; tl.Ast.Def.Orig != nil && !tl.HasErrors() {
				if defidx := this.IndexByID(tl.ID()); defidx < 0 {
					newdefs = append(newdefs, tl)
				} else {
					oldunchangeddefidxs[defidx], this[defidx].TopLevel, this[defidx].Orig = true, tl, tl.Ast.Def.Orig
				}
			}
		}
	}

	if len(oldunchangeddefidxs) > 0 && len(oldunchangeddefidxs) < len(this) { // gather & drop what's gone
		dels := make(map[*atmolang.AstDef]bool, len(this)-len(oldunchangeddefidxs))
		droppedTopLevelDefIDsAndNames = make(map[string]string, len(this)-len(oldunchangeddefidxs))
		for i := range this {
			if !oldunchangeddefidxs[i] {
				dels[this[i].Orig] = true
				droppedTopLevelDefIDsAndNames[this[i].Id] = this[i].Name.Val
			}
		}
		thisnew := make(AstTopDefs, 0, len(this)) // nasty way to delete but they're all nasty
		for i := range this {
			if !dels[this[i].Orig] {
				thisnew = append(thisnew, this[i])
			}
		}
		this = thisnew
	}

	// add what's new
	newstartfrom := len(this)
	newTopLevelDefIDsAndNames = make(map[string]string, len(newdefs))
	for _, tlc := range newdefs {
		if !tlc.HasErrors() {
			def := &AstDefTop{TopLevel: tlc, Id: tlc.ID(), refersTo: make(map[string]bool)}
			def.Orig = tlc.Ast.Def.Orig
			this = append(this, def)
			newTopLevelDefIDsAndNames[def.Id] = def.Orig.Name.Val
		}
	}
	// populate new `Def`s from orig AST node
	for i := newstartfrom; i < len(this); i++ {
		def := this[i]
		var let AstExprLetBase
		var ctxastinit ctxAstInit
		let.letPrefix, ctxastinit.defsScope, ctxastinit.curTopLevelDef = ctxastinit.nextPrefix(), &let.letDefs, def
		def.Errs.Add(def.initFrom(&ctxastinit, def.Orig))
		if len(let.letDefs) > 0 {
			def.Errs.Add(ctxastinit.addLetDefsToNode(def.Orig.Body, def.Body, &let))
		}
		if len(def.Errs) > 0 {
			freshErrs = append(freshErrs, def.Errs.Errors()...)
		}
		atmo.SortMaybe(def.Errs)
	}
	atmo.SortMaybe(this)
	*me = this
	return
}
