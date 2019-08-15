package atmoil

import (
	. "github.com/metaleap/atmo"
	. "github.com/metaleap/atmo/ast"
)

func (me IrDefs) Len() int          { return len(me) }
func (me IrDefs) Swap(i int, j int) { me[i], me[j] = me[j], me[i] }
func (me IrDefs) Less(i int, j int) bool {
	dis, dat := &me[i].AstOrigToks(nil).First1().Pos, &me[j].AstOrigToks(nil).First1().Pos
	return (dis.FilePath == dat.FilePath && me[i].OrigTopChunk.PosOffsetByte() < me[j].OrigTopChunk.PosOffsetByte()) || dis.FilePath < dat.FilePath
}

func (me IrDefs) ByName(name string, onlyFor *AstFile) (defs []*IrDef) {
	allfiles := (onlyFor == nil)
	for _, tld := range me {
		if allfiles || (tld.OrigTopChunk.SrcFile.SrcFilePath == onlyFor.SrcFilePath) {
			if orig := tld.OrigDef(); tld.Name.Val == name || (orig != nil && orig.Name.Val == name) {
				defs = append(defs, tld)
			}
		}
	}
	return
}

func (me IrDefs) IndexByID(id string) int {
	for i := range me {
		if me[i].Id == id {
			return i
		}
	}
	return -1
}

func (me *IrDefs) ReInitFrom(kitSrcFiles AstFiles) (droppedTopLevelDefIdsAndNames map[string]string, newTopLevelDefIdsAndNames map[string]string, freshErrs Errors) {
	this, newdefs, oldunchangeds := *me, make([]*AstFileChunk, 0, 2), make(map[int]Exist, len(*me))

	// gather what's "new" (newly added or source-wise modified) and what's "old" (source-wise unchanged)
	for i := range kitSrcFiles {
		for j := range kitSrcFiles[i].TopLevel {
			if tl := &kitSrcFiles[i].TopLevel[j]; tl.Ast.Def.Orig != nil && !tl.HasErrors() {
				if defidx := this.IndexByID(tl.Id()); defidx >= 0 {
					oldunchangeds[defidx], this[defidx].OrigTopChunk, this[defidx].Orig = Є, tl, tl.Ast.Def.Orig
				} else { // any source chunk with parse/lex errs doesn't exist for us anymore at this point
					newdefs = append(newdefs, tl)
				}
			} else if tl.Ast.Def.NameIfErr != "" {
				newdefs = append(newdefs, tl) // temporarily
			}
		}
	}

	if len(newdefs) == 0 && len(oldunchangeds) == len(this) {
		return
	}

	// drop what's gone
	if l := len(oldunchangeds); l < len(this) { // either some (l>0) or all (l==0) are gone, the latter will occur too seldomly in practice to optimize for
		thiswithout := make(IrDefs, 0, l+len(newdefs))
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
	if len(newdefs) != 0 {
		newTopLevelDefIdsAndNames = make(map[string]string, len(newdefs))
		for _, tlc := range newdefs {
			if tlc.Ast.Def.NameIfErr != "" {
				newTopLevelDefIdsAndNames[tlc.Id()] = tlc.Ast.Def.NameIfErr
			} else { // add the def skeleton
				orig, def := tlc.Ast.Def.Orig, &IrDef{OrigTopChunk: tlc, Id: tlc.Id(), refersTo: make(map[string]bool, 8)}
				this, newTopLevelDefIdsAndNames[def.Id] =
					append(this, def), tlc.Ast.Def.Orig.Name.Val
				// populate it
				ctxinit := ctxIrFromAst{curTopLevelDef: def, absIdx: -1}
				def.Errs.Stage1AstToIr.Add(def.initFrom(&ctxinit, orig)...)
				if len(def.Errs.Stage1AstToIr) != 0 {
					freshErrs.Add(def.Errs.Stage1AstToIr...)
				}
			}
		}
	}
	SortMaybe(this)
	*me = this
	return
}
