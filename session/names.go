package atmosess

import (
	"github.com/go-leap/str"
	"github.com/metaleap/atmo/lang/irfun"
)

type namesInScope map[string][]atmolang_irfun.IAstNode

func (me namesInScope) add(k string, v ...atmolang_irfun.IAstNode) {
	me[k] = append(me[k], v...)
}

func (me *Ctx) kitsRepopulateIdentNamesInScope() {
	kitrepops := make(map[*Kit]bool, len(me.Kits.all))

	{ // FIRST: namesInScopeOwn
		for _, kit := range me.Kits.all {
			if len(kit.state.defsGoneIDsNames) > 0 || len(kit.state.defsNew) > 0 {
				kitrepops[kit], kit.lookups.namesInScopeOwn = true, make(namesInScope, len(kit.topLevel))
				for _, tld := range kit.topLevel {
					kit.lookups.namesInScopeOwn.add(tld.Name.Val, tld)
				}
			}
		}
	}
	{ // NEXT: namesInScopeExt
		for _, kit := range me.Kits.all {
			if len(kit.Imports) > 0 {
				var totaldefscount int
				var anychanges bool
				kimps := me.Kits.all.Where(func(k *Kit) (iskimp bool) {
					if iskimp = ustr.In(k.ImpPath, kit.Imports...); iskimp {
						totaldefscount, anychanges = totaldefscount+len(k.topLevel), anychanges || len(k.state.defsGoneIDsNames) > 0 || len(k.state.defsNew) > 0
					}
					return
				})
				if anychanges {
					kitrepops[kit], kit.lookups.namesInScopeExt = true, make(namesInScope, totaldefscount)
					for _, kimp := range kimps {
						for k, v := range kimp.lookups.namesInScopeOwn {
							nodes := make([]atmolang_irfun.IAstNode, len(v))
							for i, n := range v {
								nodes[i] = astNodeExt{IAstNode: n, kit: kimp.ImpPath}
							}
							kit.lookups.namesInScopeExt.add(k, nodes...)
						}
					}
				}
			}
		}
	}

	copyAndAdd := func(copySrc namesInScope, add interface{}) (copyDst namesInScope) {
		addarg, _ := add.(*atmolang_irfun.AstDefArg)
		adddefs, _ := add.(atmolang_irfun.AstDefs)
		var namesexpected []string
		switch {
		case addarg != nil:
			namesexpected = []string{addarg.AstIdentName.Val}
		case len(adddefs) > 0:
			namesexpected = make([]string, len(adddefs))
			for i := range adddefs {
				namesexpected[i] = adddefs[i].Name.Val
			}
		}
		copyDst = make(namesInScope, len(copySrc)+1)
		for k, v := range copySrc {
			if !ustr.In(k, namesexpected...) {
				copyDst[k] = v
			} else {
				copyDst.add(k, v...)
			}
		}
		switch {
		case addarg != nil:
			k, v := addarg.AstIdentName.Val, addarg
			copyDst.add(k, v)
		case len(adddefs) > 0:
			for i := range adddefs {
				k, v := adddefs[i].Name.Val, &adddefs[i]
				copyDst.add(k, v)
			}
		}
		return
	}

	var repopulateIdentNamesInScope func(atmolang_irfun.IAstNode, namesInScope)
	repopulateIdentNamesInScope = func(node atmolang_irfun.IAstNode, inScope namesInScope) {
		inscope := inScope
		if ldx, _ := node.(atmolang_irfun.IAstExprWithLetDefs); ldx != nil {
			lds := ldx.LetDefs()
			if len(lds) > 0 {
				inscope = copyAndAdd(inscope, lds)
			}
			for i := range lds {
				repopulateIdentNamesInScope(&lds[i], inscope)
			}
		}
		switch n := node.(type) {
		case *atmolang_irfun.AstDef:
			if n.Arg != nil {
				inscope = copyAndAdd(inscope, n.Arg)
			}
			repopulateIdentNamesInScope(n.Body, inscope)
		case *atmolang_irfun.AstAppl:
			repopulateIdentNamesInScope(n.AtomicCallee, inscope)
			repopulateIdentNamesInScope(n.AtomicArg, inscope)
		case *atmolang_irfun.AstCases:
			for i := range n.Ifs {
				repopulateIdentNamesInScope(n.Ifs[i], inscope)
				repopulateIdentNamesInScope(n.Thens[i], inscope)
			}
		case *atmolang_irfun.AstIdentName:
			n.NamesInScope = inscope
		}
	}

	for _, kit := range me.Kits.all {
		if kitrepops[kit] {
			kit.errs.badNames = nil
			combined := make(namesInScope, len(kit.lookups.namesInScopeExt)+len(kit.lookups.namesInScopeOwn))
			for k, v := range kit.lookups.namesInScopeOwn {
				nodes := make([]atmolang_irfun.IAstNode, len(v))
				copy(nodes, v)
				combined[k] = nodes
			}
			for k, v := range kit.lookups.namesInScopeExt {
				combined[k] = append(combined[k], v...)
			}
			for _, tld := range kit.topLevel {
				repopulateIdentNamesInScope(&tld.AstDef, combined)
			}
		}
	}
}
