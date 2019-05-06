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

type AstTopDefs []AstDefTop

func (me AstTopDefs) Len() int          { return len(me) }
func (me AstTopDefs) Swap(i int, j int) { me[i], me[j] = me[j], me[i] }
func (me AstTopDefs) Less(i int, j int) bool {
	dis, dat := &me[i].Orig.Tokens[0].Meta, &me[j].Orig.Tokens[0].Meta
	return (dis.Filename == dat.Filename && dis.Offset < dat.Offset) || dis.Filename < dat.Filename
}

func (me AstTopDefs) IndexByID(id string) int {
	for i := range me {
		if me[i].TopLevel.ID() == id {
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
			this[idx].TopLevel, this[idx].Orig = tlc, tlc.Ast.Def.Orig
		}
	}

	// gather all def names, they then come into scope for each def
	names, ndone := make([]astNameInScope, 0, len(this)), make(map[string]int, len(this))
	for i := range this {
		name, arity := &this[i].Orig.Name, len(this[i].Orig.Args) > 0
		if idx, ok := ndone[name.Val]; !ok {
			ndone[name.Val], names = len(names), append(names, astNameInScope{name.Val, name.Tokens[0].Meta.Position, arity})
		} else if n := &names[idx]; i >= newstartfrom && !(arity && n.arity) {
			this[i].Errors.AddNaming(&name.Tokens[0], "name `"+name.Val+"` already taken by "+n.pos.String())
		}
	}

	// populate new `Def`s from orig AST node
	for i := newstartfrom; i < len(this); i++ {
		def := &this[i]
		var let AstExprLetBase
		var ctxastinit ctxAstInit
		let.letPrefix, ctxastinit.defsScope, ctxastinit.curTopLevel, ctxastinit.namesInScope = ctxastinit.nextPrefix(), &let.letDefs, def.Orig, append(ctxastinit.namesInScope, names...)
		def.Errors.Add(def.initFrom(&ctxastinit, def.Orig))
		if len(let.letDefs) > 0 {
			def.Errors.Add(ctxastinit.addLetDefsToNode(def.Orig.Body, def.AstDef.Body, &let))
		}
		if len(def.Errors) > 0 {
			for e := range def.Errors {
				freshErrs = append(freshErrs, &def.Errors[e])
			}
		}
		atmo.SortMaybe(def.Errors)
	}
	atmo.SortMaybe(this)
	*me = this
	return
}
