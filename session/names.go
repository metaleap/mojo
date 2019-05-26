package atmosess

import (
	"github.com/go-leap/str"
	"github.com/metaleap/atmo"
	"github.com/metaleap/atmo/lang/irfun"
)

type astDefRef struct {
	*atmolang_irfun.AstDefTop
	kit string
}

type namesInScope map[string][]atmolang_irfun.IAstNode

func (me namesInScope) add(k string, v ...atmolang_irfun.IAstNode) {
	me[k] = append(me[k], v...)
}

func (me *Ctx) kitsRepopulateIdentNamesInScope() {
	kitrepops := make(map[*Kit]atmo.Empty, len(me.Kits.all))

	{ // FIRST: namesInScopeOwn
		for _, kit := range me.Kits.all {
			if len(kit.state.defsGoneIdsNames) > 0 || len(kit.state.defsBornIdsNames) > 0 {
				kitrepops[kit], kit.lookups.namesInScopeAll, kit.lookups.namesInScopeOwn = atmo.Exists, nil, make(namesInScope, len(kit.topLevelDefs))
				for _, tld := range kit.topLevelDefs {
					kit.lookups.namesInScopeOwn.add(tld.Name.Val, tld)
				}
			}
		}
	}
	{ // NEXT: namesInScopeExt (extra loop because potentially utilizing above namesInScopeOwn adds)
		for _, kit := range me.Kits.all {
			if len(kit.Imports) > 0 {
				var totaldefscount int
				var anychangesinkimps bool
				kimps := me.Kits.all.Where(func(k *Kit) (iskimp bool) {
					if iskimp = ustr.In(k.ImpPath, kit.Imports...); iskimp {
						totaldefscount, anychangesinkimps = totaldefscount+len(k.topLevelDefs), anychangesinkimps || len(k.state.defsGoneIdsNames) > 0 || len(k.state.defsBornIdsNames) > 0
					}
					return
				})
				if anychangesinkimps {
					kitrepops[kit], kit.lookups.namesInScopeAll, kit.lookups.namesInScopeExt = atmo.Exists, nil, make(namesInScope, totaldefscount)
					for _, kimp := range kimps {
						for k, v := range kimp.lookups.namesInScopeOwn {
							nodes := make([]atmolang_irfun.IAstNode, len(v))
							for i, n := range v {
								nodes[i] = astDefRef{kit: kimp.ImpPath,
									AstDefTop: n.(*atmolang_irfun.AstDefTop) /* ok to panic here bc should-never-happen-else-its-a-bug */}
							}
							kit.lookups.namesInScopeExt.add(k, nodes...)
						}
					}
				}
			}
		}
	}

	for kit := range kitrepops {
		kit.lookups.namesInScopeAll = make(namesInScope, len(kit.lookups.namesInScopeExt)+len(kit.lookups.namesInScopeOwn))
		for k, v := range kit.lookups.namesInScopeOwn {
			nodes := make([]atmolang_irfun.IAstNode, len(v))
			copy(nodes, v)
			kit.lookups.namesInScopeAll[k] = nodes
		}
		for k, v := range kit.lookups.namesInScopeExt {
			kit.lookups.namesInScopeAll[k] = append(kit.lookups.namesInScopeAll[k], v...)
		}
		for _, tld := range kit.topLevelDefs {
			kit.lookups.namesInScopeAll.repopulateAstIdentsFor(&tld.AstDef)
		}
	}
}

func (me namesInScope) repopulateAstIdentsFor(node atmolang_irfun.IAstNode) {
	inscope := me
	if ldx, _ := node.(atmolang_irfun.IAstExprWithLetDefs); ldx != nil {
		lds := ldx.LetDefs()
		if len(lds) > 0 {
			inscope = inscope.copyAndAdd(lds)
		}
		for i := range lds {
			inscope.repopulateAstIdentsFor(&lds[i])
		}
	}
	switch n := node.(type) {
	case *atmolang_irfun.AstDef:
		if n.Arg != nil {
			inscope = inscope.copyAndAdd(n.Arg)
		}
		inscope.repopulateAstIdentsFor(n.Body)
	case *atmolang_irfun.AstAppl:
		inscope.repopulateAstIdentsFor(n.AtomicCallee)
		inscope.repopulateAstIdentsFor(n.AtomicArg)
	// case *atmolang_irfun.AstCases:
	// 	for i := range n.Ifs {
	// 		inscope.repopulateAstIdentsFor(n.Ifs[i])
	// 		inscope.repopulateAstIdentsFor(n.Thens[i])
	// 	}
	case *atmolang_irfun.AstIdentName:
		n.NamesInScope = inscope
	}
}

func (me namesInScope) copyAndAdd(add interface{}) (namesInScopeCopy namesInScope) {
	addarg, _ := add.(*atmolang_irfun.AstDefArg)
	adddefs, _ := add.(atmolang_irfun.AstDefs)
	var namestoadd []string
	switch {
	case addarg != nil:
		namestoadd = []string{addarg.AstIdentName.Val}
	case len(adddefs) > 0:
		namestoadd = make([]string, len(adddefs))
		for i := range adddefs {
			namestoadd[i] = adddefs[i].Name.Val
		}
	}
	namesInScopeCopy = make(namesInScope, len(me)+len(namestoadd))
	// copy old names:
	for k, v := range me {
		if !ustr.In(k, namestoadd...) {
			namesInScopeCopy[k] = v // safe to keep existing slice as-is
		} else {
			namesInScopeCopy.add(k, v...) // effectively copy existing slice
		}
	}
	// add new names:
	switch {
	case addarg != nil:
		k, v := addarg.AstIdentName.Val, addarg
		namesInScopeCopy.add(k, v)
	case len(adddefs) > 0:
		for i := range adddefs {
			k, v := adddefs[i].Name.Val, &adddefs[i]
			namesInScopeCopy.add(k, v)
		}
	}
	return
}
