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
	dis, dat := &me[i].OrigDef.Tokens[0].Meta, &me[j].OrigDef.Tokens[0].Meta
	return (dis.Pos.Filename == dat.Pos.Filename && me[i].OrigTopLevelChunk.PosOffsetByte() < me[j].OrigTopLevelChunk.PosOffsetByte()) || dis.Pos.Filename < dat.Pos.Filename
}

func (me AstTopDefs) ByName(name string) (defs []*AstDefTop) {
	for _, tld := range me {
		if tld.Name.Val == name {
			defs = append(defs, tld)
		}
	}
	return
}

func (me AstTopDefs) IndexByID(id string) int {
	for i := range me {
		if me[i].Id == id {
			return i
		}
	}
	return -1
}

func (me *AstTopDefs) ReInitFrom(kitSrcFiles atmolang.AstFiles) (droppedTopLevelDefIdsAndNames map[string]string, newTopLevelDefIdsAndNames map[string]string, freshErrs []error) {
	this, newdefs, oldunchangeds := *me, make([]*atmolang.SrcTopChunk, 0, 2), make(map[int]atmo.Exist, len(*me))

	// gather what's "new" (newly added or source-wise modified) and what's "old" (source-wise unchanged)
	for i := range kitSrcFiles {
		for j := range kitSrcFiles[i].TopLevel {
			if tl := &kitSrcFiles[i].TopLevel[j]; tl.Ast.Def.Orig != nil && !tl.HasErrors() {
				if defidx := this.IndexByID(tl.Id()); defidx >= 0 {
					oldunchangeds[defidx], this[defidx].OrigTopLevelChunk, this[defidx].OrigDef = atmo.Є, tl, tl.Ast.Def.Orig
				} else if !tl.HasErrors() { // any source chunk with parse/lex errs doesn't exist for us anymore at this point
					newdefs = append(newdefs, tl)
				}
			}
		}
	}

	if len(newdefs) == 0 && len(oldunchangeds) == len(this) {
		return
	}

	// drop what's gone
	if l := len(oldunchangeds); l < len(this) { // either some (l>0) or all (l==0) are gone, the latter will occur too seldomly in practice to optimize for
		thiswithout := make(AstTopDefs, 0, l+len(newdefs))
		droppedTopLevelDefIdsAndNames = make(map[string]string, len(this)-l)
		for i := range this {
			if _, oldunchanged := oldunchangeds[i]; oldunchanged {
				thiswithout = append(thiswithout, this[i])
			} else {
				droppedTopLevelDefIdsAndNames[this[i].Id] = this[i].Name.Val
			}
		}
		this = thiswithout
	}

	// add what's new
	if len(newdefs) > 0 {
		newTopLevelDefIdsAndNames = make(map[string]string, len(newdefs))
		for _, tlc := range newdefs {
			// add the def skeleton
			def := &AstDefTop{OrigTopLevelChunk: tlc, Id: tlc.Id(), refersTo: make(map[string]bool)}
			this, def.OrigDef, newTopLevelDefIdsAndNames[def.Id] =
				append(this, def), tlc.Ast.Def.Orig, tlc.Ast.Def.Orig.Name.Val
			// populate it
			var let AstExprLetBase
			var ctxastinit ctxAstInit
			let.letPrefix, ctxastinit.defsScope, ctxastinit.curTopLevelDef = ctxastinit.nextPrefix(), &let.Defs, def
			def.Errs.Stage0Init.Add(def.initFrom(&ctxastinit, def.OrigDef))
			if len(let.Defs) > 0 {
				def.Errs.Stage0Init.Add(ctxastinit.addLetDefsToNode(def.OrigDef.Body, def.Body, &let))
			}
			if len(def.Errs.Stage0Init) > 0 {
				freshErrs = append(freshErrs, def.Errs.Stage0Init.Errors()...)
			}
		}
	}
	atmo.SortMaybe(this)
	*me = this
	return
}
